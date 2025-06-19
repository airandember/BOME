import { writable } from 'svelte/store';
import { showToast } from '$lib/toast';

// API Configuration
export const API_CONFIG = {
	BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1',
	TIMEOUT: 30000,
	RETRY_ATTEMPTS: 3,
	RETRY_DELAY: 1000,
	CACHE_TTL: 5 * 60 * 1000, // 5 minutes
};

// API Response Types
export interface ApiResponse<T = any> {
	success: boolean;
	data?: T;
	error?: string;
	message?: string;
	meta?: {
		page?: number;
		limit?: number;
		total?: number;
		totalPages?: number;
	};
}

// Custom Error Class
export class ApiError extends Error {
	public code: string;
	public details?: any;
	public timestamp: string;

	constructor({ code, message, details, timestamp }: {
		code: string;
		message: string;
		details?: any;
		timestamp: string;
	}) {
		super(message);
		this.name = 'ApiError';
		this.code = code;
		this.details = details;
		this.timestamp = timestamp;
	}
}

// Cache Management
class ApiCache {
	private cache = new Map<string, { data: any; timestamp: number; ttl: number }>();

	set(key: string, data: any, ttl: number = API_CONFIG.CACHE_TTL): void {
		this.cache.set(key, {
			data,
			timestamp: Date.now(),
			ttl
		});
	}

	get(key: string): any | null {
		const item = this.cache.get(key);
		if (!item) return null;

		if (Date.now() - item.timestamp > item.ttl) {
			this.cache.delete(key);
			return null;
		}

		return item.data;
	}

	clear(): void {
		this.cache.clear();
	}

	delete(key: string): void {
		this.cache.delete(key);
	}
}

// Request Queue for Retry Logic
class RequestQueue {
	private queue: Array<() => Promise<any>> = [];
	private processing = false;

	async add<T>(request: () => Promise<T>): Promise<T> {
		return new Promise((resolve, reject) => {
			this.queue.push(async () => {
				try {
					const result = await request();
					resolve(result);
				} catch (error) {
					reject(error);
				}
			});

			if (!this.processing) {
				this.process();
			}
		});
	}

	private async process(): Promise<void> {
		this.processing = true;

		while (this.queue.length > 0) {
			const request = this.queue.shift();
			if (request) {
				await request();
			}
		}

		this.processing = false;
	}
}

// API Client Class
class ApiClient {
	private cache = new ApiCache();
	private requestQueue = new RequestQueue();
	private abortControllers = new Map<string, AbortController>();

	// Connection Status Store
	public connectionStatus = writable<'connected' | 'disconnected' | 'reconnecting'>('connected');

	private async makeRequest<T>(
		url: string,
		options: RequestInit = {},
		useCache: boolean = true,
		cacheTtl?: number
	): Promise<ApiResponse<T>> {
		const fullUrl = `${API_CONFIG.BASE_URL}${url}`;
		const cacheKey = `${options.method || 'GET'}:${fullUrl}:${JSON.stringify(options.body || {})}`;

		// Check cache for GET requests
		if (useCache && (!options.method || options.method === 'GET')) {
			const cachedData = this.cache.get(cacheKey);
			if (cachedData) {
				return cachedData;
			}
		}

		// Create abort controller for request cancellation
		const abortController = new AbortController();
		this.abortControllers.set(cacheKey, abortController);

		const requestOptions: RequestInit = {
			...options,
			signal: abortController.signal,
			headers: {
				'Content-Type': 'application/json',
				...this.getAuthHeaders(),
				...options.headers,
			},
		};

		try {
			const response = await this.executeWithRetry(
				() => fetch(fullUrl, requestOptions),
				API_CONFIG.RETRY_ATTEMPTS
			);

			const data = await response.json();

			if (!response.ok) {
				throw new ApiError({
					code: response.status.toString(),
					message: data.error || data.message || 'Request failed',
					details: data,
					timestamp: new Date().toISOString()
				});
			}

			const result: ApiResponse<T> = {
				success: true,
				data: data.data || data,
				message: data.message,
				meta: data.meta
			};

			// Cache successful GET requests
			if (useCache && (!options.method || options.method === 'GET')) {
				this.cache.set(cacheKey, result, cacheTtl);
			}

			this.connectionStatus.set('connected');
			return result;

		} catch (error) {
			if (error instanceof Error && error.name === 'AbortError') {
				throw new ApiError({
					code: 'ABORTED',
					message: 'Request was cancelled',
					timestamp: new Date().toISOString()
				});
			}

			if (error instanceof ApiError) {
				throw error;
			}

			// Network error
			this.connectionStatus.set('disconnected');
			const errorMessage = error instanceof Error ? error.message : 'Network request failed';
			throw new ApiError({
				code: 'NETWORK_ERROR',
				message: errorMessage,
				details: error,
				timestamp: new Date().toISOString()
			});
		} finally {
			this.abortControllers.delete(cacheKey);
		}
	}

	private async executeWithRetry<T>(
		request: () => Promise<T>,
		maxAttempts: number
	): Promise<T> {
		let lastError: Error;

		for (let attempt = 1; attempt <= maxAttempts; attempt++) {
			try {
				return await request();
			} catch (error) {
				lastError = error as Error;

				if (attempt === maxAttempts) {
					break;
				}

				// Don't retry for certain error types
				const errorStatus = (error as any)?.status;
				if (errorStatus >= 400 && errorStatus < 500) {
					break;
				}

				// Wait before retry with exponential backoff
				const delay = API_CONFIG.RETRY_DELAY * Math.pow(2, attempt - 1);
				await new Promise(resolve => setTimeout(resolve, delay));

				this.connectionStatus.set('reconnecting');
			}
		}

		throw lastError!;
	}

	private getAuthHeaders(): Record<string, string> {
		const token = localStorage.getItem('token');
		return token ? { Authorization: `Bearer ${token}` } : {};
	}

	// Public API Methods
	async get<T>(url: string, useCache: boolean = true, cacheTtl?: number): Promise<ApiResponse<T>> {
		return this.makeRequest<T>(url, { method: 'GET' }, useCache, cacheTtl);
	}

	async post<T>(url: string, data?: any): Promise<ApiResponse<T>> {
		return this.makeRequest<T>(url, {
			method: 'POST',
			body: data ? JSON.stringify(data) : undefined
		}, false);
	}

	async put<T>(url: string, data?: any): Promise<ApiResponse<T>> {
		return this.makeRequest<T>(url, {
			method: 'PUT',
			body: data ? JSON.stringify(data) : undefined
		}, false);
	}

	async patch<T>(url: string, data?: any): Promise<ApiResponse<T>> {
		return this.makeRequest<T>(url, {
			method: 'PATCH',
			body: data ? JSON.stringify(data) : undefined
		}, false);
	}

	async delete<T>(url: string): Promise<ApiResponse<T>> {
		return this.makeRequest<T>(url, { method: 'DELETE' }, false);
	}

	// File Upload
	async upload<T>(url: string, file: File, onProgress?: (progress: number) => void): Promise<ApiResponse<T>> {
		const formData = new FormData();
		formData.append('file', file);

		const xhr = new XMLHttpRequest();
		
		return new Promise((resolve, reject) => {
			xhr.upload.addEventListener('progress', (e) => {
				if (e.lengthComputable && onProgress) {
					const progress = (e.loaded / e.total) * 100;
					onProgress(progress);
				}
			});

			xhr.addEventListener('load', () => {
				try {
					const response = JSON.parse(xhr.responseText);
					if (xhr.status >= 200 && xhr.status < 300) {
						resolve({
							success: true,
							data: response.data || response
						});
					} else {
						reject(new ApiError({
							code: xhr.status.toString(),
							message: response.error || 'Upload failed',
							timestamp: new Date().toISOString()
						}));
					}
				} catch (error) {
					reject(new ApiError({
						code: 'PARSE_ERROR',
						message: 'Failed to parse response',
						timestamp: new Date().toISOString()
					}));
				}
			});

			xhr.addEventListener('error', () => {
				reject(new ApiError({
					code: 'UPLOAD_ERROR',
					message: 'Upload failed',
					timestamp: new Date().toISOString()
				}));
			});

			xhr.open('POST', `${API_CONFIG.BASE_URL}${url}`);
			
			// Add auth headers
			const authHeaders = this.getAuthHeaders();
			Object.entries(authHeaders).forEach(([key, value]) => {
				xhr.setRequestHeader(key, value);
			});

			xhr.send(formData);
		});
	}

	// Cache Management
	clearCache(): void {
		this.cache.clear();
	}

	deleteCacheKey(key: string): void {
		this.cache.delete(key);
	}

	// Request Cancellation
	cancelRequest(url: string, method: string = 'GET'): void {
		const cacheKey = `${method}:${API_CONFIG.BASE_URL}${url}`;
		const controller = this.abortControllers.get(cacheKey);
		if (controller) {
			controller.abort();
		}
	}

	cancelAllRequests(): void {
		this.abortControllers.forEach(controller => controller.abort());
		this.abortControllers.clear();
	}
}

// WebSocket Manager
class WebSocketManager {
	private ws: WebSocket | null = null;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000;
	private listeners = new Map<string, Set<(data: any) => void>>();

	public connectionStatus = writable<'connected' | 'disconnected' | 'connecting'>('disconnected');

	connect(url?: string): void {
		const wsUrl = url || import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';
		
		this.connectionStatus.set('connecting');
		this.ws = new WebSocket(wsUrl);

		this.ws.onopen = () => {
			console.log('WebSocket connected');
			this.connectionStatus.set('connected');
			this.reconnectAttempts = 0;
		};

		this.ws.onmessage = (event) => {
			try {
				const data = JSON.parse(event.data);
				const { type, payload } = data;

				const typeListeners = this.listeners.get(type);
				if (typeListeners) {
					typeListeners.forEach(callback => callback(payload));
				}
			} catch (error) {
				console.error('Failed to parse WebSocket message:', error);
			}
		};

		this.ws.onclose = () => {
			console.log('WebSocket disconnected');
			this.connectionStatus.set('disconnected');
			this.attemptReconnect();
		};

		this.ws.onerror = (error) => {
			console.error('WebSocket error:', error);
			this.connectionStatus.set('disconnected');
		};
	}

	private attemptReconnect(): void {
		if (this.reconnectAttempts < this.maxReconnectAttempts) {
			this.reconnectAttempts++;
			const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
			
			setTimeout(() => {
				console.log(`Attempting WebSocket reconnection (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
				this.connect();
			}, delay);
		}
	}

	subscribe(type: string, callback: (data: any) => void): () => void {
		if (!this.listeners.has(type)) {
			this.listeners.set(type, new Set());
		}

		this.listeners.get(type)!.add(callback);

		// Return unsubscribe function
		return () => {
			const typeListeners = this.listeners.get(type);
			if (typeListeners) {
				typeListeners.delete(callback);
				if (typeListeners.size === 0) {
					this.listeners.delete(type);
				}
			}
		};
	}

	send(type: string, payload: any): void {
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify({ type, payload }));
		} else {
			console.warn('WebSocket is not connected');
		}
	}

	disconnect(): void {
		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}
		this.listeners.clear();
	}
}

// Export instances
export const apiClient = new ApiClient();
export const wsManager = new WebSocketManager(); 