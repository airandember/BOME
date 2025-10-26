// BOME Frontend API Client
// Centralized HTTP client for all backend communication

import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { get } from 'svelte/store';
import { config } from '$lib/config';

// Import the auth store for token management
import { auth } from '$lib/auth';

export interface ApiResponse<T> {
	data?: T;
	error?: string;
	message?: string;
}

export interface LoginResponse {
	user: {
		id: number;
		email: string;
		role: string;
		full_name: string;
	};
	token: string;
}

export interface TokenInfo {
	token: string;
	expiresAt: number;
	refreshToken?: string;
}

class ApiClient {
	private baseUrl: string;

	constructor() {
		this.baseUrl = browser ? config.apiBaseUrl : '';
	}

	get token(): string | null {
		if (!browser) return null;
		
		// Get token from auth store
		const authState = get(auth);
		if (!authState.isAuthenticated || !authState.token) {
			return null;
		}
		
		return authState.token;
	}

	private async request<T>(
		endpoint: string,
		options: RequestInit = {}
	): Promise<ApiResponse<T>> {
		try {
			const url = `${this.baseUrl}${endpoint}`;
			const token = this.token;
			
			console.log('ApiClient: Making request to:', url);
			console.log('ApiClient: Token available:', !!token);
			console.log('ApiClient: Token value:', token ? `${token.substring(0, 20)}...` : 'null');
			
			const config: RequestInit = {
				headers: {
					'Content-Type': 'application/json',
					...(token && { Authorization: `Bearer ${token}` }),
					...options.headers,
				},
				...options,
			};

			console.log('ApiClient: Request headers:', config.headers);
			const response = await fetch(url, config);
			
		// Handle 401 - token expired
		if (response.status === 401 && token) {
			console.warn('🔴 [API] 401 Unauthorized for endpoint:', endpoint);
			console.warn('🔴 [API] Response status:', response.status);
			
			// Clear tokens regardless of response
			// The auth store will handle clearing tokens on 401
			
			// Redirect to login for non-auth endpoints
			if (browser && !endpoint.includes('/auth/')) {
				console.warn('🔴 [API] Redirecting to login due to 401');
				goto('/auth/login?expired=true');
			}
			
			return { error: 'Authentication required' };
		}
			
			// Check if response is JSON
			const contentType = response.headers.get('content-type');
			if (!contentType || !contentType.includes('application/json')) {
				// Handle non-JSON responses (like HTML error pages)
				const text = await response.text();
				console.error('Non-JSON response received:', {
					status: response.status,
					contentType,
					url,
					text: text.substring(0, 200) // Log first 200 chars
				});
				
				return {
					error: `Expected JSON response but got ${contentType || 'unknown content type'} (HTTP ${response.status})`
				};
			}
			
			const data = await response.json();
			
			if (!response.ok) {
				return {
					error: data.error || data.message || `HTTP ${response.status}`,
					message: data.message
				};
			}
			
			return { data };
		} catch (error) {
			console.error('API request failed:', error);
			return {
				error: error instanceof Error ? error.message : 'Network error'
			};
		}
	}

	async login(credentials: { email: string; password: string }): Promise<ApiResponse<LoginResponse>> {
		const response = await this.request<LoginResponse>('/auth/login', {
			method: 'POST',
			body: JSON.stringify(credentials),
		});

		if (response.data) {
			// The auth store will handle saving tokens
		}

		return response;
	}

	async logout(): Promise<ApiResponse<void>> {
		const response = await this.request<void>('/auth/logout', {
			method: 'POST',
		});

		// Clear tokens regardless of response
		// The auth store will handle clearing tokens on 401

		return response;
	}

	async get<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint);
	}

	async post<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'POST',
			body: data ? JSON.stringify(data) : undefined,
		});
	}

	async postWithHeaders<T>(endpoint: string, data?: any, customHeaders?: Record<string, string>): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'POST',
			body: data ? JSON.stringify(data) : undefined,
			headers: customHeaders,
		});
	}

	async put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined,
		});
	}

	async putWithHeaders<T>(endpoint: string, data?: any, customHeaders?: Record<string, string>): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined,
			headers: customHeaders,
		});
	}

	async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'DELETE',
		});
	}
}

export const apiClient = new ApiClient(); 
