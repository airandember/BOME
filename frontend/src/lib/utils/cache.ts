// Advanced Caching System with TTL, LRU, and Invalidation
import { browser } from '$app/environment';

interface CacheEntry<T> {
	data: T;
	timestamp: number;
	ttl: number;
	accessCount: number;
	lastAccessed: number;
}

interface CacheConfig {
	maxSize: number;
	defaultTTL: number;
	storagePrefix: string;
	persistToStorage: boolean;
}

export class AdvancedCache<T = any> {
	private cache = new Map<string, CacheEntry<T>>();
	private config: CacheConfig;
	private cleanupInterval: number | null = null;

	constructor(config: Partial<CacheConfig> = {}) {
		this.config = {
			maxSize: 100,
			defaultTTL: 5 * 60 * 1000, // 5 minutes
			storagePrefix: 'bome_cache_',
			persistToStorage: true,
			...config
		};

		// Initialize from localStorage if available
		if (browser && this.config.persistToStorage) {
			this.loadFromStorage();
		}

		// Start cleanup interval
		this.startCleanup();
	}

	// Set cache entry with optional TTL
	set(key: string, data: T, ttl?: number): void {
		const now = Date.now();
		const entry: CacheEntry<T> = {
			data,
			timestamp: now,
			ttl: ttl || this.config.defaultTTL,
			accessCount: 0,
			lastAccessed: now
		};

		// Remove oldest entries if cache is full
		if (this.cache.size >= this.config.maxSize) {
			this.evictLRU();
		}

		this.cache.set(key, entry);

		// Persist to storage
		if (browser && this.config.persistToStorage) {
			this.saveToStorage(key, entry);
		}
	}

	// Get cache entry
	get(key: string): T | null {
		const entry = this.cache.get(key);
		if (!entry) return null;

		const now = Date.now();

		// Check if expired
		if (now - entry.timestamp > entry.ttl) {
			this.delete(key);
			return null;
		}

		// Update access stats
		entry.accessCount++;
		entry.lastAccessed = now;

		return entry.data;
	}

	// Check if key exists and is not expired
	has(key: string): boolean {
		return this.get(key) !== null;
	}

	// Delete cache entry
	delete(key: string): boolean {
		const deleted = this.cache.delete(key);
		
		if (browser && this.config.persistToStorage) {
			localStorage.removeItem(this.config.storagePrefix + key);
		}

		return deleted;
	}

	// Clear all cache
	clear(): void {
		this.cache.clear();
		
		if (browser && this.config.persistToStorage) {
			// Remove all cache entries from localStorage
			for (let i = localStorage.length - 1; i >= 0; i--) {
				const key = localStorage.key(i);
				if (key?.startsWith(this.config.storagePrefix)) {
					localStorage.removeItem(key);
				}
			}
		}
	}

	// Get cache statistics
	getStats(): {
		size: number;
		maxSize: number;
		hitRate: number;
		entries: Array<{
			key: string;
			size: number;
			age: number;
			accessCount: number;
			lastAccessed: number;
		}>;
	} {
		const now = Date.now();
		const entries = Array.from(this.cache.entries()).map(([key, entry]) => ({
			key,
			size: JSON.stringify(entry.data).length,
			age: now - entry.timestamp,
			accessCount: entry.accessCount,
			lastAccessed: entry.lastAccessed
		}));

		return {
			size: this.cache.size,
			maxSize: this.config.maxSize,
			hitRate: this.calculateHitRate(),
			entries
		};
	}

	// Invalidate cache entries by pattern
	invalidatePattern(pattern: string | RegExp): number {
		let count = 0;
		const regex = typeof pattern === 'string' ? new RegExp(pattern) : pattern;

		for (const key of this.cache.keys()) {
			if (regex.test(key)) {
				this.delete(key);
				count++;
			}
		}

		return count;
	}

	// Refresh cache entry (useful for background updates)
	async refresh<R>(
		key: string, 
		fetcher: () => Promise<R>, 
		ttl?: number
	): Promise<R> {
		try {
			const data = await fetcher();
			this.set(key, data as unknown as T, ttl);
			return data;
		} catch (error) {
			// Return cached data if available, even if expired
			const entry = this.cache.get(key);
			if (entry) {
				console.warn(`Cache refresh failed for ${key}, returning stale data:`, error);
				return entry.data as unknown as R;
			}
			throw error;
		}
	}

	// Get or fetch pattern (cache-aside pattern)
	async getOrFetch<R>(
		key: string,
		fetcher: () => Promise<R>,
		ttl?: number
	): Promise<R> {
		// Try cache first
		const cached = this.get(key);
		if (cached !== null) {
			return cached as unknown as R;
		}

		// Fetch and cache
		const data = await fetcher();
		this.set(key, data as unknown as T, ttl);
		return data;
	}

	// Batch operations
	setMany(entries: Array<{ key: string; data: T; ttl?: number }>): void {
		entries.forEach(({ key, data, ttl }) => {
			this.set(key, data, ttl);
		});
	}

	getMany(keys: string[]): Array<{ key: string; data: T | null }> {
		return keys.map(key => ({
			key,
			data: this.get(key)
		}));
	}

	// Private methods
	private evictLRU(): void {
		let oldestKey: string | null = null;
		let oldestTime = Date.now();

		for (const [key, entry] of this.cache.entries()) {
			if (entry.lastAccessed < oldestTime) {
				oldestTime = entry.lastAccessed;
				oldestKey = key;
			}
		}

		if (oldestKey) {
			this.delete(oldestKey);
		}
	}

	private calculateHitRate(): number {
		let totalAccess = 0;
		for (const entry of this.cache.values()) {
			totalAccess += entry.accessCount;
		}
		return totalAccess > 0 ? (this.cache.size / totalAccess) * 100 : 0;
	}

	private loadFromStorage(): void {
		try {
			for (let i = 0; i < localStorage.length; i++) {
				const key = localStorage.key(i);
				if (key?.startsWith(this.config.storagePrefix)) {
					const cacheKey = key.replace(this.config.storagePrefix, '');
					const stored = localStorage.getItem(key);
					if (stored) {
						const entry: CacheEntry<T> = JSON.parse(stored);
						// Only load if not expired
						if (Date.now() - entry.timestamp < entry.ttl) {
							this.cache.set(cacheKey, entry);
						} else {
							localStorage.removeItem(key);
						}
					}
				}
			}
		} catch (error) {
			console.warn('Failed to load cache from storage:', error);
		}
	}

	private saveToStorage(key: string, entry: CacheEntry<T>): void {
		try {
			localStorage.setItem(
				this.config.storagePrefix + key,
				JSON.stringify(entry)
			);
		} catch (error) {
			console.warn('Failed to save cache to storage:', error);
		}
	}

	private startCleanup(): void {
		if (this.cleanupInterval) return;
		if (!browser) return; // Skip cleanup during SSR

		this.cleanupInterval = window.setInterval(() => {
			this.cleanup();
		}, 60000); // Cleanup every minute
	}

	private cleanup(): void {
		const now = Date.now();
		const toDelete: string[] = [];

		for (const [key, entry] of this.cache.entries()) {
			if (now - entry.timestamp > entry.ttl) {
				toDelete.push(key);
			}
		}

		toDelete.forEach(key => this.delete(key));
	}

	// Cleanup on destruction
	destroy(): void {
		if (this.cleanupInterval) {
			clearInterval(this.cleanupInterval);
			this.cleanupInterval = null;
		}
	}
}

// Create global cache instances
export const apiCache = new AdvancedCache({
	maxSize: 200,
	defaultTTL: 5 * 60 * 1000, // 5 minutes
	storagePrefix: 'bome_api_',
	persistToStorage: true
});

export const imageCache = new AdvancedCache({
	maxSize: 500,
	defaultTTL: 30 * 60 * 1000, // 30 minutes
	storagePrefix: 'bome_img_',
	persistToStorage: false // Images are too large for localStorage
});

export const userCache = new AdvancedCache({
	maxSize: 50,
	defaultTTL: 15 * 60 * 1000, // 15 minutes
	storagePrefix: 'bome_user_',
	persistToStorage: true
});

// Cache key generators
export const cacheKeys = {
	videos: (page: number, category?: string, search?: string) => 
		`videos:${page}:${category || 'all'}:${search || 'none'}`,
	video: (id: number) => `video:${id}`,
	videoComments: (id: number) => `video_comments:${id}`,
	categories: () => 'categories',
	user: (id: number) => `user:${id}`,
	dashboard: (userId: number) => `dashboard:${userId}`,
	adminAnalytics: () => 'admin_analytics',
	advertisers: (status?: string) => `advertisers:${status || 'all'}`,
	campaigns: (advertiserId?: number, status?: string) => 
		`campaigns:${advertiserId || 'all'}:${status || 'all'}`
};

// Cache invalidation helpers
export const cacheInvalidation = {
	// Invalidate all video-related cache when videos change
	videos: () => {
		apiCache.invalidatePattern(/^videos:/);
		apiCache.invalidatePattern(/^video:/);
		apiCache.delete(cacheKeys.categories());
	},
	
	// Invalidate user-related cache
	user: (userId: number) => {
		apiCache.delete(cacheKeys.user(userId));
		apiCache.delete(cacheKeys.dashboard(userId));
		userCache.delete(userId.toString());
	},
	
	// Invalidate admin cache
	admin: () => {
		apiCache.invalidatePattern(/^admin_/);
		apiCache.invalidatePattern(/^advertisers:/);
		apiCache.invalidatePattern(/^campaigns:/);
	}
}; 