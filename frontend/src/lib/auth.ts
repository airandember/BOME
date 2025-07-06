import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { page } from '$app/stores';
import type { Role } from './types/roles';

export interface User {
	id: number;
	email: string;
	role: string;
	first_name: string;
	last_name: string;
	email_verified: boolean;
}

export interface AuthTokens {
	access_token: string;
	refresh_token: string;
	expires_in: number;
	token_type: string;
}

export interface LoginCredentials {
	email: string;
	password: string;
}

export interface RegisterData {
	email: string;
	password: string;
	first_name: string;
	last_name: string;
}

export interface AuthError {
	message: string;
	code?: string;
}

// Configuration
const API_BASE_URL = browser ? (import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1') : '';
const TOKEN_STORAGE_KEY = 'bome_auth_tokens';
const USER_STORAGE_KEY = 'bome_user_data';

// Stores
export const authTokens = writable<AuthTokens | null>(null);
export const currentUser = writable<User | null>(null);
export const isAuthenticated = derived(
	[authTokens, currentUser], 
	([$tokens, $user]) => !!$tokens && !!$user
);
export const isLoading = writable(false);
export const authError = writable<AuthError | null>(null);

// Token refresh scheduling
let refreshTimeout: number | null = null;

// Secure token storage using httpOnly cookies (fallback to localStorage for development)
class SecureTokenStorage {
	private static readonly TOKEN_KEY = 'bome_auth_token';
	private static readonly REFRESH_KEY = 'bome_refresh_token';
	
	static storeTokens(accessToken: string, refreshToken: string): void {
		if (browser) {
			// In production, use httpOnly cookies
			// For development, use localStorage with additional security
			try {
				// Store with expiration
				const tokenData = {
					token: accessToken,
					refresh: refreshToken,
					expires: Date.now() + (15 * 60 * 1000), // 15 minutes
					created: Date.now()
				};
				
				localStorage.setItem(this.TOKEN_KEY, JSON.stringify(tokenData));
				
				// Set a flag to indicate secure storage
				localStorage.setItem('bome_secure_storage', 'true');
			} catch (error) {
				console.error('Failed to store tokens securely:', error);
			}
		}
	}
	
	static getAccessToken(): string | null {
		if (!browser) return null;
		
		try {
			const stored = localStorage.getItem(this.TOKEN_KEY);
			if (!stored) return null;
			
			const tokenData = JSON.parse(stored);
			
			// Check if token is expired
			if (Date.now() > tokenData.expires) {
				this.clearTokens();
				return null;
			}
			
			return tokenData.token;
		} catch {
			return null;
		}
	}
	
	static getRefreshToken(): string | null {
		if (!browser) return null;
		
		try {
			const stored = localStorage.getItem(this.TOKEN_KEY);
			if (!stored) return null;
			
			const tokenData = JSON.parse(stored);
			return tokenData.refresh;
		} catch {
			return null;
		}
	}
	
	static clearTokens(): void {
		if (browser) {
			localStorage.removeItem(this.TOKEN_KEY);
			localStorage.removeItem('bome_secure_storage');
		}
	}
	
	static isSecure(): boolean {
		if (!browser) return false;
		return localStorage.getItem('bome_secure_storage') === 'true';
	}
}

// Create the auth store with methods
function createAuthStore() {
	const { subscribe, set, update } = writable({
		isAuthenticated: false,
		user: null as User | null,
		token: null as string | null,
		loading: false,
		error: null as string | null
	});

	return {
		subscribe,
		set,
		update,
		
		// Authentication methods
		async login(email: string, password: string) {
			try {
				console.log('Auth: Starting login process for:', email);
				update(state => ({ ...state, loading: true, error: null }));
				isLoading.set(true);
				authError.set(null);
				
				const response = await apiRequest('/auth/login', {
					method: 'POST',
					body: JSON.stringify({ email, password }),
				});
				
				console.log('Auth: Login response status:', response.status);
				
				if (!response.ok) {
					const error = await response.json();
					console.log('Auth: Login failed with error:', error);
					throw new Error(error.error || 'Login failed');
				}
				
				const data = await response.json();
				console.log('Auth: Login response data:', data);
				
				const tokens: AuthTokens = {
					access_token: data.access_token,
					refresh_token: data.refresh_token,
					expires_in: data.expires_in,
					token_type: data.token_type
				};
				
				const user: User = {
					id: data.user.id,
					email: data.user.email,
					role: data.user.role,
					first_name: data.user.first_name,
					last_name: data.user.last_name,
					email_verified: data.user.email_verified
				};
				
				console.log('Auth: Parsed tokens and user:', { tokens, user });
				storeAuthData(tokens, user);
				console.log('Auth: Login successful, returning result');
				return { success: true, user };
			} catch (error) {
				console.error('Login error:', error);
				const errorMessage = error instanceof Error ? error.message : 'Login failed';
				authError.set({
					message: errorMessage,
					code: 'LOGIN_FAILED'
				});
				update(state => ({ ...state, loading: false, error: errorMessage }));
				return { success: false, error: errorMessage };
			} finally {
				isLoading.set(false);
			}
		},

		async register(data: RegisterData) {
			try {
				update(state => ({ ...state, loading: true, error: null }));
				isLoading.set(true);
				authError.set(null);
				
				const response = await apiRequest('/auth/register', {
					method: 'POST',
					body: JSON.stringify(data),
				});
				
				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Registration failed');
				}
				
				return { success: true };
			} catch (error) {
				console.error('Registration error:', error);
				const errorMessage = error instanceof Error ? error.message : 'Registration failed';
				authError.set({
					message: errorMessage,
					code: 'REGISTRATION_FAILED'
				});
				update(state => ({ ...state, loading: false, error: errorMessage }));
				return { success: false, error: errorMessage };
			} finally {
				isLoading.set(false);
			}
		},

		async logout() {
			try {
				const refreshToken = SecureTokenStorage.getRefreshToken();
				if (refreshToken) {
					await apiRequest('/auth/logout', {
						method: 'POST',
						body: JSON.stringify({ 
							refresh_token: refreshToken,
							all_devices: false // Can be made configurable
						}),
					}).catch(() => {
						// Ignore errors on logout
					});
				}
			} catch (error) {
				console.error('Logout error:', error);
			} finally {
				SecureTokenStorage.clearTokens();
				clearAuthData();
				if (browser) {
					goto('/login');
				}
			}
		},

		async forgotPassword(email: string) {
			try {
				update(state => ({ ...state, loading: true, error: null }));
				isLoading.set(true);
				authError.set(null);
				
				const response = await apiRequest('/auth/forgot-password', {
					method: 'POST',
					body: JSON.stringify({ email }),
				});
				
				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Failed to send reset email');
				}
				
				return { success: true };
			} catch (error) {
				console.error('Forgot password error:', error);
				const errorMessage = error instanceof Error ? error.message : 'Failed to send reset email';
				authError.set({
					message: errorMessage,
					code: 'FORGOT_PASSWORD_FAILED'
				});
				update(state => ({ ...state, loading: false, error: errorMessage }));
				return { success: false, error: errorMessage };
			} finally {
				isLoading.set(false);
			}
		},

		async resetPassword(token: string, password: string) {
			try {
				update(state => ({ ...state, loading: true, error: null }));
				isLoading.set(true);
				authError.set(null);
				
				const response = await apiRequest('/auth/reset-password', {
					method: 'POST',
					body: JSON.stringify({ token, password }),
				});
				
				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Password reset failed');
				}
				
				return { success: true };
			} catch (error) {
				console.error('Reset password error:', error);
				const errorMessage = error instanceof Error ? error.message : 'Password reset failed';
				authError.set({
					message: errorMessage,
					code: 'RESET_PASSWORD_FAILED'
				});
				update(state => ({ ...state, loading: false, error: errorMessage }));
				return { success: false, error: errorMessage };
			} finally {
				isLoading.set(false);
			}
		},

		async verifyEmail(token: string) {
			try {
				update(state => ({ ...state, loading: true, error: null }));
				isLoading.set(true);
				authError.set(null);
				
				const response = await apiRequest('/auth/verify-email', {
					method: 'POST',
					body: JSON.stringify({ token }),
				});
				
				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Email verification failed');
				}
				
				// Update user's email verification status
				const user = getCurrentUser();
				if (user) {
					user.email_verified = true;
					currentUser.set(user);
					update(state => ({
						...state,
						user,
						loading: false
					}));
					if (browser) {
						localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
					}
				}
				
				return { success: true };
			} catch (error) {
				console.error('Email verification error:', error);
				const errorMessage = error instanceof Error ? error.message : 'Email verification failed';
				authError.set({
					message: errorMessage,
					code: 'EMAIL_VERIFICATION_FAILED'
				});
				update(state => ({ ...state, loading: false, error: errorMessage }));
				return { success: false, error: errorMessage };
			} finally {
				isLoading.set(false);
			}
		},

		async changePassword(currentPassword: string, newPassword: string) {
			try {
				update(state => ({ ...state, loading: true, error: null }));
				isLoading.set(true);
				authError.set(null);
				
				const response = await apiRequest('/auth/change-password', {
					method: 'POST',
					body: JSON.stringify({ current_password: currentPassword, new_password: newPassword }),
				});
				
				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Password change failed');
				}
				
				return { success: true };
			} catch (error) {
				console.error('Change password error:', error);
				const errorMessage = error instanceof Error ? error.message : 'Password change failed';
				authError.set({
					message: errorMessage,
					code: 'CHANGE_PASSWORD_FAILED'
				});
				update(state => ({ ...state, loading: false, error: errorMessage }));
				return { success: false, error: errorMessage };
			} finally {
				isLoading.set(false);
			}
		},

		clearError() {
			authError.set(null);
			update(state => ({ ...state, error: null }));
		}
	};
}

// Create the auth store
export const auth = createAuthStore();

// Initialize auth state from storage
if (browser) {
	initializeAuth();
}

// Test function to check backend connectivity
export async function testBackendConnectivity() {
	try {
		console.log('Auth: Testing backend connectivity...');
		// Health endpoint is at root, not under /api/v1
		const healthUrl = API_BASE_URL.replace('/api/v1', '') + '/health';
		console.log('Auth: Health check URL:', healthUrl);
		const response = await fetch(healthUrl);
		console.log('Auth: Backend health check response:', { status: response.status, ok: response.ok });
		return response.ok;
	} catch (error) {
		console.error('Auth: Backend connectivity test failed:', error);
		return false;
	}
}

export async function initializeAuth() {
	try {
		console.log('Auth: Starting initialization...');
		const storedTokens = localStorage.getItem(TOKEN_STORAGE_KEY);
		const storedUser = localStorage.getItem(USER_STORAGE_KEY);
		
		console.log('Auth: Stored tokens:', storedTokens ? 'Found' : 'Not found');
		console.log('Auth: Stored user:', storedUser ? 'Found' : 'Not found');
		
		if (storedTokens && storedUser) {
			const tokens: AuthTokens = JSON.parse(storedTokens);
			const user: User = JSON.parse(storedUser);
			
			console.log('Auth: Parsed tokens:', tokens);
			console.log('Auth: Parsed user:', user);
			
			// Check if tokens are still valid (basic check)
			if (isTokenValid(tokens)) {
				console.log('Auth: Tokens are valid, setting auth state');
				authTokens.set(tokens);
				currentUser.set(user);
				auth.set({
					isAuthenticated: true,
					user,
					token: tokens.access_token,
					loading: false,
					error: null
				});
				
				// Schedule token refresh
				scheduleTokenRefresh(tokens);
				console.log('Auth: Auth state set successfully');
			} else {
				console.log('Auth: Tokens are invalid, clearing auth data');
				// Tokens expired, clear storage
				clearAuthData();
			}
		} else {
			console.log('Auth: No stored tokens or user found');
		}
		
		// Add a small delay to ensure state is properly set
		await new Promise(resolve => setTimeout(resolve, 50));
	} catch (error) {
		console.error('Failed to initialize auth from storage:', error);
		clearAuthData();
	}
}

function isTokenValid(tokens: AuthTokens): boolean {
	// This is a basic check - in production you might want to decode the JWT
	// and check the actual expiration time
	const isValid = !!tokens.access_token && !!tokens.refresh_token;
	console.log('Auth: Token validation check:', {
		hasAccessToken: !!tokens.access_token,
		hasRefreshToken: !!tokens.refresh_token,
		isValid
	});
	return isValid;
}

function clearAuthData() {
	if (browser) {
		localStorage.removeItem(TOKEN_STORAGE_KEY);
		localStorage.removeItem(USER_STORAGE_KEY);
	}
	authTokens.set(null);
	currentUser.set(null);
	auth.set({
		isAuthenticated: false,
		user: null,
		token: null,
		loading: false,
		error: null
	});
	authError.set(null);
}

function storeAuthData(tokens: AuthTokens, user: User) {
	console.log('Auth: Storing auth data:', { tokens, user });
	if (browser) {
		localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(tokens));
		localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
		console.log('Auth: Data stored in localStorage');
	}
	authTokens.set(tokens);
	currentUser.set(user);
	auth.set({
		isAuthenticated: true,
		user,
		token: tokens.access_token,
		loading: false,
		error: null
	});
	
	// Schedule token refresh
	scheduleTokenRefresh(tokens);
	console.log('Auth: Auth state updated in stores');
}

function scheduleTokenRefresh(tokens: AuthTokens) {
	if (refreshTimeout) {
		clearTimeout(refreshTimeout);
	}
	
	// Refresh 1 minute before expiration
	const refreshTime = (tokens.expires_in - 60) * 1000;
	
	if (refreshTime > 0) {
		refreshTimeout = setTimeout(() => {
			refreshTokens();
		}, refreshTime);
	}
}

// API helper function
export async function apiRequest(endpoint: string, options: RequestInit = {}): Promise<Response> {
	const url = `${API_BASE_URL}${endpoint}`;
	console.log('Auth: Making API request to:', url);
	console.log('Auth: API_BASE_URL:', API_BASE_URL);
	
	const config: RequestInit = {
		headers: {
			'Content-Type': 'application/json',
			...options.headers,
		},
		...options,
	};
	
	// Add auth header if we have tokens
	const tokens = getCurrentTokens();
	if (tokens && !endpoint.includes('/auth/')) {
		config.headers = {
			...config.headers,
			'Authorization': `Bearer ${tokens.access_token}`,
		};
	}
	
	console.log('Auth: Request config:', { method: config.method, headers: config.headers, body: config.body });
	
	try {
		const response = await fetch(url, config);
		console.log('Auth: Response received:', { status: response.status, ok: response.ok, statusText: response.statusText });
		
		// Handle 401 - token expired
		if (response.status === 401 && tokens && !endpoint.includes('/auth/')) {
			// Try to refresh tokens
			const refreshed = await refreshTokens();
			if (refreshed) {
				// Retry the original request with new token
				config.headers = {
					...config.headers,
					'Authorization': `Bearer ${getCurrentTokens()?.access_token}`,
				};
				return fetch(url, config);
			} else {
				// Refresh failed, redirect to login
				await auth.logout();
				throw new Error('Authentication required');
			}
		}
		
		return response;
	} catch (error) {
		console.error('Auth: Network error during API request:', error);
		throw error;
	}
}

function getCurrentTokens(): AuthTokens | null {
	let tokens: AuthTokens | null = null;
	authTokens.subscribe(value => tokens = value)();
	return tokens;
}

function getCurrentUser(): User | null {
	let user: User | null = null;
	currentUser.subscribe(value => user = value)();
	return user;
}

export async function refreshTokens(): Promise<boolean> {
	try {
		const tokens = getCurrentTokens();
		if (!tokens) {
			return false;
		}
		
		const response = await apiRequest('/auth/refresh', {
			method: 'POST',
			body: JSON.stringify({ refresh_token: tokens.refresh_token }),
		});
		
		if (!response.ok) {
			return false;
		}
		
		const data = await response.json();
		const newTokens: AuthTokens = {
			access_token: data.access_token,
			refresh_token: data.refresh_token,
			expires_in: data.expires_in,
			token_type: data.token_type
		};
		
		// Update tokens in store
		authTokens.set(newTokens);
		
		// Update auth store
		auth.update(state => ({
			...state,
			token: newTokens.access_token
		}));
		
		// Update storage
		if (browser) {
			localStorage.setItem(TOKEN_STORAGE_KEY, JSON.stringify(newTokens));
		}
		
		// Schedule next refresh
		scheduleTokenRefresh(newTokens);
		
		return true;
	} catch (error) {
		console.error('Token refresh error:', error);
		return false;
	}
}

// Utility functions
export function requireAuth() {
	if (browser) {
		const user = getCurrentUser();
		if (!user) {
			goto('/login');
		}
	}
}

export function requireRole(allowedRoles: string[]) {
	if (browser) {
		const user = getCurrentUser();
		if (!user || !allowedRoles.includes(user.role)) {
			goto('/unauthorized');
		}
	}
}

export function requireEmailVerification() {
	if (browser) {
		const user = getCurrentUser();
		if (!user || !user.email_verified) {
			goto('/verify-email');
		}
	}
}

export function hasRole(role: string): boolean {
	const user = getCurrentUser();
	return user?.role === role;
}

export function isAdmin(): boolean {
	const user = getCurrentUser();
	if (!user) return false;
	
	// Admin roles include all roles with level 7+ (subsystem managers and above)
	const adminRoles = [
		'super_admin',           // Level 10: Super Administrator
		'system_admin',          // Level 9: System Administrator
		'content_manager',       // Level 8: Content Manager
		'articles_manager',      // Level 7: Articles Manager
		'youtube_manager',       // Level 7: YouTube Manager
		'streaming_manager',     // Level 7: Video Streaming Manager
		'events_manager',        // Level 7: Events Manager
		'advertisement_manager', // Level 7: Advertisement Manager
		'user_manager',          // Level 7: User Account Manager
		'analytics_manager',     // Level 7: Analytics Manager
		'financial_admin',       // Level 7: Financial Administrator
		'admin'                  // Legacy admin role
	];
	
	return adminRoles.includes(user.role);
}

export function isAdvertiser(): boolean {
	return hasRole('advertiser') || hasRole('admin');
}

export function isEmailVerified(): boolean {
	const user = getCurrentUser();
	return user?.email_verified || false;
} 