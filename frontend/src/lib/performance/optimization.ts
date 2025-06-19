// Performance Optimization Utilities
import { writable } from 'svelte/store';

// Performance Configuration
export const PERFORMANCE_CONFIG = {
	LAZY_LOAD_THRESHOLD: 100, // pixels
	IMAGE_CACHE_SIZE: 50, // number of images
	VIDEO_PRELOAD_SIZE: 3, // number of videos
	DEBOUNCE_DELAY: 300, // milliseconds
	THROTTLE_DELAY: 100, // milliseconds
	INTERSECTION_THRESHOLD: 0.1 // 10% visibility
};

// Image Optimization
export class ImageOptimizer {
	private static cache = new Map<string, HTMLImageElement>();
	private static loadingImages = new Set<string>();

	static async optimizeImage(src: string, options: {
		width?: number;
		height?: number;
		quality?: number;
		format?: 'webp' | 'jpeg' | 'png';
	} = {}): Promise<string> {
		const { width, height, quality = 80, format = 'webp' } = options;
		
		// Create optimized URL with parameters
		const params = new URLSearchParams();
		if (width) params.append('w', width.toString());
		if (height) params.append('h', height.toString());
		params.append('q', quality.toString());
		params.append('f', format);

		return `${src}?${params.toString()}`;
	}

	static async preloadImage(src: string): Promise<HTMLImageElement> {
		if (this.cache.has(src)) {
			return this.cache.get(src)!;
		}

		if (this.loadingImages.has(src)) {
			// Wait for existing load to complete
			return new Promise((resolve) => {
				const checkCache = () => {
					if (this.cache.has(src)) {
						resolve(this.cache.get(src)!);
					} else {
						setTimeout(checkCache, 10);
					}
				};
				checkCache();
			});
		}

		this.loadingImages.add(src);

		return new Promise((resolve, reject) => {
			const img = new Image();
			img.onload = () => {
				this.cache.set(src, img);
				this.loadingImages.delete(src);
				
				// Manage cache size
				if (this.cache.size > PERFORMANCE_CONFIG.IMAGE_CACHE_SIZE) {
					const firstKey = this.cache.keys().next().value;
					this.cache.delete(firstKey);
				}
				
				resolve(img);
			};
			img.onerror = () => {
				this.loadingImages.delete(src);
				reject(new Error(`Failed to load image: ${src}`));
			};
			img.src = src;
		});
	}

	static clearCache(): void {
		this.cache.clear();
		this.loadingImages.clear();
	}

	static getCacheSize(): number {
		return this.cache.size;
	}
}

// Lazy Loading Manager
export class LazyLoader {
	private static observer: IntersectionObserver | null = null;
	private static targets = new WeakMap<Element, () => void>();

	static initialize(): void {
		if (!this.observer && typeof window !== 'undefined') {
			this.observer = new IntersectionObserver(
				(entries) => {
					entries.forEach((entry) => {
						if (entry.isIntersecting) {
							const callback = this.targets.get(entry.target);
							if (callback) {
								callback();
								this.observer?.unobserve(entry.target);
								this.targets.delete(entry.target);
							}
						}
					});
				},
				{
					rootMargin: `${PERFORMANCE_CONFIG.LAZY_LOAD_THRESHOLD}px`,
					threshold: PERFORMANCE_CONFIG.INTERSECTION_THRESHOLD
				}
			);
		}
	}

	static observe(element: Element, callback: () => void): void {
		this.initialize();
		if (this.observer) {
			this.targets.set(element, callback);
			this.observer.observe(element);
		}
	}

	static unobserve(element: Element): void {
		if (this.observer) {
			this.observer.unobserve(element);
			this.targets.delete(element);
		}
	}

	static disconnect(): void {
		if (this.observer) {
			this.observer.disconnect();
			this.observer = null;
			this.targets = new WeakMap();
		}
	}
}

// Video Optimization
export class VideoOptimizer {
	private static preloadedVideos = new Map<string, HTMLVideoElement>();
	private static loadingVideos = new Set<string>();

	static async preloadVideo(src: string, options: {
		preload?: 'none' | 'metadata' | 'auto';
		muted?: boolean;
	} = {}): Promise<HTMLVideoElement> {
		const { preload = 'metadata', muted = true } = options;

		if (this.preloadedVideos.has(src)) {
			return this.preloadedVideos.get(src)!;
		}

		if (this.loadingVideos.has(src)) {
			return new Promise((resolve) => {
				const checkCache = () => {
					if (this.preloadedVideos.has(src)) {
						resolve(this.preloadedVideos.get(src)!);
					} else {
						setTimeout(checkCache, 10);
					}
				};
				checkCache();
			});
		}

		this.loadingVideos.add(src);

		return new Promise((resolve, reject) => {
			const video = document.createElement('video');
			video.preload = preload;
			video.muted = muted;
			video.playsInline = true;

			video.addEventListener('loadedmetadata', () => {
				this.preloadedVideos.set(src, video);
				this.loadingVideos.delete(src);
				
				// Manage cache size
				if (this.preloadedVideos.size > PERFORMANCE_CONFIG.VIDEO_PRELOAD_SIZE) {
					const firstKey = this.preloadedVideos.keys().next().value;
					const firstVideo = this.preloadedVideos.get(firstKey);
					if (firstVideo) {
						firstVideo.src = '';
						firstVideo.load();
					}
					this.preloadedVideos.delete(firstKey);
				}
				
				resolve(video);
			});

			video.addEventListener('error', () => {
				this.loadingVideos.delete(src);
				reject(new Error(`Failed to load video: ${src}`));
			});

			video.src = src;
		});
	}

	static optimizeVideoUrl(src: string, options: {
		quality?: 'auto' | '1080p' | '720p' | '480p' | '360p';
		format?: 'mp4' | 'webm' | 'hls';
		bitrate?: number;
	} = {}): string {
		const { quality = 'auto', format = 'mp4', bitrate } = options;
		
		const params = new URLSearchParams();
		params.append('quality', quality);
		params.append('format', format);
		if (bitrate) params.append('bitrate', bitrate.toString());

		return `${src}?${params.toString()}`;
	}

	static clearCache(): void {
		this.preloadedVideos.forEach((video) => {
			video.src = '';
			video.load();
		});
		this.preloadedVideos.clear();
		this.loadingVideos.clear();
	}
}

// Performance Utilities
export class PerformanceUtils {
	static debounce<T extends (...args: any[]) => void>(
		func: T,
		delay: number = PERFORMANCE_CONFIG.DEBOUNCE_DELAY
	): (...args: Parameters<T>) => void {
		let timeoutId: ReturnType<typeof setTimeout>;
		return (...args: Parameters<T>) => {
			clearTimeout(timeoutId);
			timeoutId = setTimeout(() => func(...args), delay);
		};
	}

	static throttle<T extends (...args: any[]) => void>(
		func: T,
		delay: number = PERFORMANCE_CONFIG.THROTTLE_DELAY
	): (...args: Parameters<T>) => void {
		let lastCall = 0;
		return (...args: Parameters<T>) => {
			const now = Date.now();
			if (now - lastCall >= delay) {
				lastCall = now;
				func(...args);
			}
		};
	}

	static async measurePerformance<T>(
		name: string,
		fn: () => Promise<T> | T
	): Promise<{ result: T; duration: number; memory?: number }> {
		const startTime = performance.now();
		const startMemory = this.getMemoryUsage();

		const result = await fn();

		const endTime = performance.now();
		const endMemory = this.getMemoryUsage();

		const duration = endTime - startTime;
		const memory = endMemory && startMemory ? endMemory - startMemory : undefined;

		console.log(`Performance [${name}]: ${duration.toFixed(2)}ms${memory ? `, ${memory}KB memory` : ''}`);

		return { result, duration, memory };
	}

	private static getMemoryUsage(): number | null {
		if ('memory' in performance) {
			return (performance as any).memory.usedJSHeapSize / 1024; // Convert to KB
		}
		return null;
	}

	static createVirtualScroller<T>(
		items: T[],
		itemHeight: number,
		containerHeight: number
	): {
		visibleItems: T[];
		startIndex: number;
		endIndex: number;
		totalHeight: number;
		offsetY: number;
	} {
		const totalHeight = items.length * itemHeight;
		const visibleCount = Math.ceil(containerHeight / itemHeight);
		const buffer = Math.max(5, Math.ceil(visibleCount * 0.5));
		
		return {
			visibleItems: items.slice(0, visibleCount + buffer * 2),
			startIndex: 0,
			endIndex: visibleCount + buffer * 2 - 1,
			totalHeight,
			offsetY: 0
		};
	}

	static updateVirtualScroller<T>(
		items: T[],
		itemHeight: number,
		containerHeight: number,
		scrollTop: number
	): {
		visibleItems: T[];
		startIndex: number;
		endIndex: number;
		totalHeight: number;
		offsetY: number;
	} {
		const totalHeight = items.length * itemHeight;
		const visibleCount = Math.ceil(containerHeight / itemHeight);
		const buffer = Math.max(5, Math.ceil(visibleCount * 0.5));
		
		const startIndex = Math.max(0, Math.floor(scrollTop / itemHeight) - buffer);
		const endIndex = Math.min(items.length - 1, startIndex + visibleCount + buffer * 2);
		
		const visibleItems = items.slice(startIndex, endIndex + 1);
		const offsetY = startIndex * itemHeight;

		return {
			visibleItems,
			startIndex,
			endIndex,
			totalHeight,
			offsetY
		};
	}
}

// Bundle Optimization
export class BundleOptimizer {
	static async loadModule<T>(moduleFactory: () => Promise<T>): Promise<T> {
		try {
			return await moduleFactory();
		} catch (error) {
			console.error('Failed to load module:', error);
			throw error;
		}
	}

	static preloadRoute(routePath: string): void {
		if (typeof window !== 'undefined') {
			const link = document.createElement('link');
			link.rel = 'modulepreload';
			link.href = routePath;
			document.head.appendChild(link);
		}
	}

	static preloadRoutes(routePaths: string[]): void {
		routePaths.forEach(path => this.preloadRoute(path));
	}
}

// Cache Management
export class CacheManager {
	private static caches = new Map<string, Map<string, { data: any; timestamp: number; ttl: number }>>();

	static createCache(name: string): void {
		if (!this.caches.has(name)) {
			this.caches.set(name, new Map());
		}
	}

	static set(cacheName: string, key: string, data: any, ttl: number = 5 * 60 * 1000): void {
		this.createCache(cacheName);
		const cache = this.caches.get(cacheName)!;
		
		cache.set(key, {
			data,
			timestamp: Date.now(),
			ttl
		});
	}

	static get(cacheName: string, key: string): any | null {
		const cache = this.caches.get(cacheName);
		if (!cache) return null;

		const item = cache.get(key);
		if (!item) return null;

		if (Date.now() - item.timestamp > item.ttl) {
			cache.delete(key);
			return null;
		}

		return item.data;
	}

	static clear(cacheName?: string): void {
		if (cacheName) {
			const cache = this.caches.get(cacheName);
			if (cache) {
				cache.clear();
			}
		} else {
			this.caches.clear();
		}
	}

	static size(cacheName: string): number {
		const cache = this.caches.get(cacheName);
		return cache ? cache.size : 0;
	}

	static cleanup(): void {
		const now = Date.now();
		this.caches.forEach((cache) => {
			const keysToDelete: string[] = [];
			cache.forEach((item, key) => {
				if (now - item.timestamp > item.ttl) {
					keysToDelete.push(key);
				}
			});
			keysToDelete.forEach(key => cache.delete(key));
		});
	}
}

// Performance Monitoring
export class PerformanceMonitor {
	private static metrics = writable<{
		loadTime: number;
		renderTime: number;
		memoryUsage: number;
		cacheHitRate: number;
		apiResponseTime: number;
	}>({
		loadTime: 0,
		renderTime: 0,
		memoryUsage: 0,
		cacheHitRate: 0,
		apiResponseTime: 0
	});

	static getMetrics() {
		return this.metrics;
	}

	static updateMetric(metric: string, value: number): void {
		this.metrics.update(current => ({
			...current,
			[metric]: value
		}));
	}

	static startMeasurement(name: string): () => number {
		const startTime = performance.now();
		return () => {
			const duration = performance.now() - startTime;
			console.log(`${name}: ${duration.toFixed(2)}ms`);
			return duration;
		};
	}

	static measurePageLoad(): void {
		if (typeof window !== 'undefined') {
			window.addEventListener('load', () => {
				const loadTime = performance.timing.loadEventEnd - performance.timing.navigationStart;
				this.updateMetric('loadTime', loadTime);
			});
		}
	}

	static measureMemoryUsage(): void {
		if ('memory' in performance) {
			const memoryUsage = (performance as any).memory.usedJSHeapSize / 1024 / 1024; // MB
			this.updateMetric('memoryUsage', memoryUsage);
		}
	}
}

// Export all optimizers
export {
	ImageOptimizer,
	LazyLoader,
	VideoOptimizer,
	PerformanceUtils,
	BundleOptimizer,
	CacheManager,
	PerformanceMonitor
}; 