// BOME Frontend API Client
// Centralized HTTP client for all backend communication

import { apiCache, cacheKeys, cacheInvalidation } from '$lib/utils/cache';
import { browser } from '$app/environment';

interface ApiResponse<T> {
	data?: T;
	error?: string;
	message?: string;
}

interface PaginatedResponse<T> {
	data: T[];
	pagination: {
		current_page: number;
		per_page: number;
		total: number;
		total_pages: number;
	};
}

interface LoginRequest {
	email: string;
	password: string;
}

interface LoginResponse {
	token: string;
	user: {
		id: number;
		email: string;
		role: string;
		full_name: string;
	};
}

interface RetryConfig {
	maxRetries: number;
	baseDelay: number;
	maxDelay: number;
	retryableStatuses: number[];
}

interface TokenInfo {
	token: string;
	expiresAt: number;
	refreshToken?: string;
}

interface RequestOptions {
	useCache?: boolean;
	cacheTTL?: number;
	invalidateCache?: boolean;
}

export class ApiClient {
	private baseURL: string;
	private tokenInfo: TokenInfo | null = null;
	private refreshPromise: Promise<boolean> | null = null;
	private retryConfig: RetryConfig = {
		maxRetries: 3,
		baseDelay: 1000, // 1 second
		maxDelay: 10000, // 10 seconds
		retryableStatuses: [408, 429, 500, 502, 503, 504]
	};

	constructor() {
		// Environment-specific API endpoints
		this.baseURL = this.getBaseURL();
		this.tokenInfo = this.getStoredTokenInfo();
	}

	private getBaseURL(): string {
		if (typeof window === 'undefined') return 'http://localhost:8080'; // SSR fallback
		
		// In development, use the proxy setup - just return empty string so requests go to same origin
		const env = import.meta.env;
		if (env.DEV) {
			return ''; // Use proxy, requests will go to /api/* and be proxied to backend
		}
		
		// In production, use environment variable or default
		return env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';
	}

	private getStoredTokenInfo(): TokenInfo | null {
		if (typeof window === 'undefined') return null;
		
		const stored = localStorage.getItem('bome_token_info');
		if (!stored) return null;
		
		try {
			return JSON.parse(stored);
		} catch {
			return null;
		}
	}

	// Decode JWT token to get expiration time
	private decodeJWT(token: string): { exp?: number } {
		try {
			const payload = token.split('.')[1];
			const decoded = JSON.parse(atob(payload));
			return decoded;
		} catch {
			return {};
		}
	}

	// Check if token is expired or will expire soon (within 5 minutes)
	private isTokenExpired(): boolean {
		if (!this.tokenInfo) return true;
		
		const now = Date.now() / 1000;
		const buffer = 5 * 60; // 5 minutes buffer
		
		return this.tokenInfo.expiresAt <= (now + buffer);
	}

	// Get current token (for backward compatibility)
	get token(): string | null {
		return this.tokenInfo?.token || null;
	}

	setToken(token: string, refreshToken?: string) {
		const decoded = this.decodeJWT(token);
		const expiresAt = decoded.exp || (Date.now() / 1000) + (24 * 60 * 60); // Default 24 hours
		
		this.tokenInfo = {
			token,
			expiresAt,
			refreshToken
		};
		
		if (typeof window !== 'undefined') {
			localStorage.setItem('bome_token_info', JSON.stringify(this.tokenInfo));
		}
	}

	removeToken() {
		this.tokenInfo = null;
		if (typeof window !== 'undefined') {
			localStorage.removeItem('bome_token_info');
		}
		// Clear user-related cache on logout
		apiCache.invalidatePattern(/^user:/);
		apiCache.invalidatePattern(/^dashboard:/);
	}

	// Refresh token if needed
	private async refreshTokenIfNeeded(): Promise<boolean> {
		if (!this.isTokenExpired()) return true;
		
		// If refresh is already in progress, wait for it
		if (this.refreshPromise) {
			return this.refreshPromise;
		}
		
		// Start refresh process
		this.refreshPromise = this.performTokenRefresh();
		const result = await this.refreshPromise;
		this.refreshPromise = null;
		
		return result;
	}

	private async performTokenRefresh(): Promise<boolean> {
		if (!this.tokenInfo?.refreshToken) {
			console.warn('No refresh token available, user needs to login again');
			this.removeToken();
			// Dispatch custom event for logout
			if (typeof window !== 'undefined') {
				window.dispatchEvent(new CustomEvent('auth:token-expired'));
			}
			return false;
		}

		try {
			// Build URL using same logic as main request method
			const url = this.baseURL ? `${this.baseURL}/api/v1/auth/refresh` : `/api/v1/auth/refresh`;
			
			const response = await fetch(url, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${this.tokenInfo.refreshToken}`
				}
			});

			if (!response.ok) {
				throw new Error('Token refresh failed');
			}

			const data = await response.json();
			
			if (data.token) {
				this.setToken(data.token, data.refreshToken || this.tokenInfo.refreshToken);
				console.log('Token refreshed successfully');
				return true;
			}

			throw new Error('No token in refresh response');
		} catch (error) {
			console.error('Token refresh failed:', error);
			this.removeToken();
			
			// Dispatch custom event for logout
			if (typeof window !== 'undefined') {
				window.dispatchEvent(new CustomEvent('auth:token-expired'));
			}
			
			return false;
		}
	}

	// Exponential backoff delay calculation
	private calculateDelay(attempt: number): number {
		const delay = this.retryConfig.baseDelay * Math.pow(2, attempt);
		// Add jitter to prevent thundering herd
		const jitter = Math.random() * 0.1 * delay;
		return Math.min(delay + jitter, this.retryConfig.maxDelay);
	}

	// Sleep utility for retry delays
	private sleep(ms: number): Promise<void> {
		return new Promise(resolve => setTimeout(resolve, ms));
	}

	// Check if status code is retryable
	private isRetryableError(status: number): boolean {
		return this.retryConfig.retryableStatuses.includes(status);
	}

	// Check if error is a network error (no response)
	private isNetworkError(error: any): boolean {
		return error instanceof TypeError && error.message.includes('fetch');
	}

	private async requestWithRetry<T>(
		endpoint: string, 
		options: RequestInit = {},
		attempt: number = 0,
		requestOptions: RequestOptions = {}
	): Promise<ApiResponse<T>> {
		// Check cache first if enabled
		if (requestOptions.useCache && options.method === 'GET') {
			const cached = apiCache.get(endpoint);
			if (cached) {
				return { data: cached };
			}
		}

		// Refresh token if needed (except for auth endpoints)
		if (!endpoint.startsWith('/auth/') && !await this.refreshTokenIfNeeded()) {
			return { error: 'Authentication required' };
		}

		// Build URL - if baseURL is empty (dev mode), just use the endpoint (proxy will handle it)
		const url = this.baseURL ? `${this.baseURL}/api/v1${endpoint}` : `/api/v1${endpoint}`;
		
		const headers: Record<string, string> = {
			'Content-Type': 'application/json',
		};

		// Add existing headers
		if (options.headers) {
			Object.assign(headers, options.headers);
		}

		if (this.tokenInfo?.token) {
			headers['Authorization'] = `Bearer ${this.tokenInfo.token}`;
		}

		const config: RequestInit = {
			...options,
			headers,
			// Set reasonable timeout
			signal: AbortSignal.timeout(30000), // 30 seconds
		};

		try {
			const response = await fetch(url, config);
			
			if (!response.ok) {
				// Handle 401 Unauthorized - try token refresh once
				if (response.status === 401 && attempt === 0 && !endpoint.startsWith('/auth/')) {
					console.log('Received 401, attempting token refresh...');
					if (await this.performTokenRefresh()) {
						// Retry the request with new token
						return this.requestWithRetry(endpoint, options, attempt + 1, requestOptions);
					}
				}

				// Check if we should retry
				if (attempt < this.retryConfig.maxRetries && this.isRetryableError(response.status)) {
					const delay = this.calculateDelay(attempt);
					console.warn(`API request failed (${response.status}), retrying in ${delay}ms... (attempt ${attempt + 1}/${this.retryConfig.maxRetries})`);
					await this.sleep(delay);
					return this.requestWithRetry(endpoint, options, attempt + 1, requestOptions);
				}

				const errorData = await response.json().catch(() => ({}));
				throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
			}

			const data = await response.json();
			
			// Cache successful GET requests
			if (requestOptions.useCache && options.method === 'GET' && data) {
				apiCache.set(endpoint, data.data || data, requestOptions.cacheTTL);
			}

			// Invalidate cache if requested
			if (requestOptions.invalidateCache) {
				this.invalidateRelatedCache(endpoint, options.method);
			}

			return { data };
		} catch (error) {
			// Check if we should retry on network errors
			if (attempt < this.retryConfig.maxRetries && this.isNetworkError(error)) {
				const delay = this.calculateDelay(attempt);
				console.warn(`Network error, retrying in ${delay}ms... (attempt ${attempt + 1}/${this.retryConfig.maxRetries})`);
				await this.sleep(delay);
				return this.requestWithRetry(endpoint, options, attempt + 1, requestOptions);
			}

			console.error(`API Error [${endpoint}]:`, error);
			return { 
				error: error instanceof Error ? error.message : 'Unknown error occurred' 
			};
		}
	}

	private invalidateRelatedCache(endpoint: string, method?: string) {
		// Invalidate cache based on endpoint and method
		if (endpoint.includes('/videos')) {
			cacheInvalidation.videos();
		} else if (endpoint.includes('/admin')) {
			cacheInvalidation.admin();
		} else if (endpoint.includes('/users') || endpoint.includes('/dashboard')) {
			// Extract user ID if possible and invalidate user cache
			const userIdMatch = endpoint.match(/\/users\/(\d+)/);
			if (userIdMatch) {
				cacheInvalidation.user(parseInt(userIdMatch[1]));
			}
		}
	}

	private async request<T>(
		endpoint: string, 
		options: RequestInit = {},
		requestOptions: RequestOptions = {}
	): Promise<ApiResponse<T>> {
		return this.requestWithRetry(endpoint, options, 0, requestOptions);
	}

	// Authentication endpoints
	async login(credentials: LoginRequest): Promise<ApiResponse<LoginResponse>> {
		const response = await this.request<LoginResponse>('/auth/login', {
			method: 'POST',
			body: JSON.stringify(credentials),
		}, { invalidateCache: true });

		if (response.data?.token) {
			// Note: In a real implementation, the backend should return refreshToken
			this.setToken(response.data.token, response.data.token); // Using same token as refresh for mock
		}

		return response;
	}

	async logout(): Promise<ApiResponse<{ message: string }>> {
		const response = await this.request<{ message: string }>('/auth/logout', {
			method: 'POST',
		}, { invalidateCache: true });

		this.removeToken();
		return response;
	}

	async register(userData: any): Promise<ApiResponse<{ message: string }>> {
		return this.request<{ message: string }>('/auth/register', {
			method: 'POST',
			body: JSON.stringify(userData),
		});
	}

	// Video endpoints with caching
	async getVideos(params?: {
		page?: number;
		limit?: number;
		category?: string;
		search?: string;
	}): Promise<ApiResponse<PaginatedResponse<any>>> {
		const searchParams = new URLSearchParams();
		if (params?.page) searchParams.set('page', params.page.toString());
		if (params?.limit) searchParams.set('limit', params.limit.toString());
		if (params?.category) searchParams.set('category', params.category);
		if (params?.search) searchParams.set('search', params.search);

		const queryString = searchParams.toString();
		const endpoint = `/videos${queryString ? `?${queryString}` : ''}`;
		
		return this.request<PaginatedResponse<any>>(endpoint, {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	async getVideo(id: number): Promise<ApiResponse<{ video: any }>> {
		return this.request<{ video: any }>(`/videos/${id}`, {}, {
			useCache: true,
			cacheTTL: 10 * 60 * 1000 // 10 minutes
		});
	}

	async getVideoComments(id: number): Promise<ApiResponse<{ comments: any[] }>> {
		return this.request<{ comments: any[] }>(`/videos/${id}/comments`, {}, {
			useCache: true,
			cacheTTL: 2 * 60 * 1000 // 2 minutes
		});
	}

	async getVideoCategories(): Promise<ApiResponse<{ categories: any[] }>> {
		return this.request<{ categories: any[] }>('/videos/categories', {}, {
			useCache: true,
			cacheTTL: 30 * 60 * 1000 // 30 minutes
		});
	}

	// Admin endpoints with caching
	async getAdminAnalytics(): Promise<ApiResponse<{ analytics: any }>> {
		return this.request<{ analytics: any }>('/admin/analytics', {}, {
			useCache: true,
			cacheTTL: 2 * 60 * 1000 // 2 minutes
		});
	}

	async getAdminUsers(): Promise<ApiResponse<{ users: any[], total: number }>> {
		return this.request<{ users: any[], total: number }>('/admin/users', {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	async getAdminVideos(): Promise<ApiResponse<any>> {
		return this.request<any>('/admin/videos', {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	// Advertisement endpoints with caching
	async getAdvertisers(status?: string): Promise<ApiResponse<{ advertisers: any[], total: number }>> {
		const endpoint = status ? `/admin/advertisers?status=${status}` : '/admin/advertisers';
		return this.request<{ advertisers: any[], total: number }>(endpoint, {}, {
			useCache: true,
			cacheTTL: 3 * 60 * 1000 // 3 minutes
		});
	}

	async getAdvertiser(id: number): Promise<ApiResponse<{ advertiser: any }>> {
		return this.request<{ advertiser: any }>(`/admin/advertisers/${id}`, {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	async getCampaigns(params?: {
		advertiser_id?: number;
		status?: string;
	}): Promise<ApiResponse<{ campaigns: any[], total: number }>> {
		const searchParams = new URLSearchParams();
		if (params?.advertiser_id) searchParams.set('advertiser_id', params.advertiser_id.toString());
		if (params?.status) searchParams.set('status', params.status);

		const queryString = searchParams.toString();
		const endpoint = `/admin/campaigns${queryString ? `?${queryString}` : ''}`;
		
		return this.request<{ campaigns: any[], total: number }>(endpoint, {}, {
			useCache: true,
			cacheTTL: 3 * 60 * 1000 // 3 minutes
		});
	}

	async getCampaign(id: number): Promise<ApiResponse<{ campaign: any }>> {
		return this.request<{ campaign: any }>(`/admin/campaigns/${id}`, {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	// Dashboard endpoint with caching
	async getDashboard(): Promise<ApiResponse<{ data: any }>> {
		return this.request<{ data: any }>('/dashboard', {}, {
			useCache: true,
			cacheTTL: 2 * 60 * 1000 // 2 minutes
		});
	}

	// Article endpoints with caching
	async getArticles(params?: {
		page?: number;
		limit?: number;
		category?: string;
		search?: string;
		featured?: boolean;
	}): Promise<ApiResponse<PaginatedResponse<any>>> {
		const searchParams = new URLSearchParams();
		if (params?.page) searchParams.set('page', params.page.toString());
		if (params?.limit) searchParams.set('limit', params.limit.toString());
		if (params?.category) searchParams.set('category', params.category);
		if (params?.search) searchParams.set('search', params.search);
		if (params?.featured) searchParams.set('featured', 'true');

		const queryString = searchParams.toString();
		const endpoint = `/articles${queryString ? `?${queryString}` : ''}`;
		
		return this.request<PaginatedResponse<any>>(endpoint, {}, {
			useCache: true,
			cacheTTL: 10 * 60 * 1000 // 10 minutes
		});
	}

	async getArticle(slug: string): Promise<ApiResponse<{ article: any }>> {
		return this.request<{ article: any }>(`/articles/${slug}`, {}, {
			useCache: true,
			cacheTTL: 30 * 60 * 1000 // 30 minutes
		});
	}

	async getArticleCategories(): Promise<ApiResponse<{ categories: any[] }>> {
		return this.request<{ categories: any[] }>('/articles/categories', {}, {
			useCache: true,
			cacheTTL: 60 * 60 * 1000 // 1 hour
		});
	}

	async getAuthors(): Promise<ApiResponse<{ authors: any[] }>> {
		return this.request<{ authors: any[] }>('/authors', {}, {
			useCache: true,
			cacheTTL: 30 * 60 * 1000 // 30 minutes
		});
	}

	// Role endpoints with caching
	async getRoles(): Promise<ApiResponse<{ roles: any[] }>> {
		return this.request<{ roles: any[] }>('/roles', {}, {
			useCache: true,
			cacheTTL: 15 * 60 * 1000 // 15 minutes
		});
	}

	async getPermissions(): Promise<ApiResponse<{ permissions: any[] }>> {
		return this.request<{ permissions: any[] }>('/permissions', {}, {
			useCache: true,
			cacheTTL: 30 * 60 * 1000 // 30 minutes
		});
	}

	async getRoleAnalytics(): Promise<ApiResponse<{ analytics: any }>> {
		return this.request<{ analytics: any }>('/roles/analytics', {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	async getUsersWithRoles(): Promise<ApiResponse<{ users: any[] }>> {
		return this.request<{ users: any[] }>('/users/roles', {}, {
			useCache: true,
			cacheTTL: 5 * 60 * 1000 // 5 minutes
		});
	}

	// Cache management methods
	getCacheStats() {
		return apiCache.getStats();
	}

	clearCache() {
		apiCache.clear();
	}

	invalidateCache(pattern: string | RegExp) {
		return apiCache.invalidatePattern(pattern);
	}
}

// Create singleton instance
export const apiClient = new ApiClient();

// Export types for external use
export type { ApiResponse, PaginatedResponse, LoginRequest, LoginResponse }; 