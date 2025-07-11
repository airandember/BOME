// BOME Frontend API Client
// Centralized HTTP client for all backend communication

import { browser } from '$app/environment';
import { goto } from '$app/navigation';

// Import the centralized token storage
import { SecureTokenStorage } from '$lib/auth';

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
	private tokenInfo: TokenInfo | null = null;

	constructor() {
		this.baseUrl = browser ? (import.meta.env.VITE_API_URL || '/api/v1') : '';
		this.loadTokenFromStorage();
	}

	private loadTokenFromStorage(): void {
		if (!browser) return;

		try {
			const token = SecureTokenStorage.getAccessToken();
			if (token) {
				this.tokenInfo = {
					token,
					expiresAt: Date.now() + (4 * 60 * 60 * 1000), // 4 hours
					refreshToken: SecureTokenStorage.getRefreshToken() || undefined
				};
			}
		} catch (error) {
			console.error('Failed to load token from storage:', error);
			this.tokenInfo = null;
		}
	}

	private saveTokenToStorage(): void {
		if (!browser || !this.tokenInfo) return;

		try {
			if (this.tokenInfo.refreshToken) {
				SecureTokenStorage.storeTokens(this.tokenInfo.token, this.tokenInfo.refreshToken);
			}
		} catch (error) {
			console.error('Failed to save token to storage:', error);
		}
	}

	get token(): string | null {
		if (!this.tokenInfo) return null;
		
		// Check if token is expired
		if (Date.now() >= this.tokenInfo.expiresAt) {
			this.tokenInfo = null;
			SecureTokenStorage.clearTokens();
			return null;
		}
		
		return this.tokenInfo.token;
	}

	private async request<T>(
		endpoint: string,
		options: RequestInit = {}
	): Promise<ApiResponse<T>> {
		try {
			const url = `${this.baseUrl}${endpoint}`;
			const token = this.token;
			
			const config: RequestInit = {
				headers: {
					'Content-Type': 'application/json',
					...(token && { Authorization: `Bearer ${token}` }),
					...options.headers,
				},
				...options,
			};

			const response = await fetch(url, config);
			
			// Handle 401 - token expired
			if (response.status === 401 && token) {
				this.tokenInfo = null;
				SecureTokenStorage.clearTokens();
				
				// Redirect to login for auth endpoints
				if (browser && !endpoint.includes('/auth/')) {
					goto('/login');
				}
				
				return { error: 'Authentication required' };
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
			this.tokenInfo = {
				token: response.data.token,
				expiresAt: Date.now() + (4 * 60 * 60 * 1000), // 4 hours
			};
			this.saveTokenToStorage();
		}

		return response;
	}

	async logout(): Promise<ApiResponse<void>> {
		const response = await this.request<void>('/auth/logout', {
			method: 'POST',
		});

		// Clear tokens regardless of response
		this.tokenInfo = null;
		SecureTokenStorage.clearTokens();

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

	async put<T>(endpoint: string, data?: any): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined,
		});
	}

	async delete<T>(endpoint: string): Promise<ApiResponse<T>> {
		return this.request<T>(endpoint, {
			method: 'DELETE',
		});
	}
}

export const apiClient = new ApiClient(); 