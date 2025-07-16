// Video optimization and intelligent preloading system
import { browser } from '$app/environment';

// Types
interface VideoMetadata {
  id: string;
  title: string;
  duration: number;
  thumbnailUrl: string;
  playbackUrl: string;
  iframeSrc: string;
  cached: boolean;
  cacheTime: number;
}

interface CachedVideoData {
  data: any;
  timestamp: number;
  accessCount: number;
}

interface PerformanceMetrics {
  cacheHits: number;
  cacheMisses: number;
  preloadSuccess: number;
  preloadFailures: number;
  averageLoadTime: number;
  networkCondition: string;
}

interface NetworkCondition {
  type: string;
  speed: number;
  quality: 'slow' | 'medium' | 'fast';
}

// Video optimization service with intelligent preloading and caching
export class VideoOptimizationService {
  private preloadQueue: Set<string> = new Set();
  private processingQueue: boolean = false;
  private cache: Map<string, CachedVideoData> = new Map();
  private performanceMetrics: PerformanceMetrics = {
    cacheHits: 0,
    cacheMisses: 0,
    preloadSuccess: 0,
    preloadFailures: 0,
    averageLoadTime: 0,
    networkCondition: 'unknown'
  };
  
  // Reduced concurrency and added rate limiting
  private readonly MAX_CONCURRENT_REQUESTS = 2; // Reduced from 5
  private readonly BATCH_SIZE = 3; // Reduced from 5
  private readonly RATE_LIMIT_DELAY = 1000; // 1 second between batches
  private readonly PRELOAD_DELAY = 500; // 500ms delay before processing
  private lastRequestTime = 0;
  private networkCondition: NetworkCondition = { type: 'unknown', speed: 0, quality: 'medium' };

  constructor() {
    if (browser) {
      this.detectNetworkCondition();
      this.startCacheCleanup();
    }
  }

  // Detect network condition
  private detectNetworkCondition(): void {
    if (browser && 'connection' in navigator) {
      const connection = (navigator as any).connection;
      if (connection) {
        this.networkCondition = {
          type: connection.effectiveType || 'unknown',
          speed: connection.downlink || 0,
          quality: connection.effectiveType === '4g' ? 'fast' : 
                   connection.effectiveType === '3g' ? 'medium' : 'slow'
        };
        this.performanceMetrics.networkCondition = this.networkCondition.type;
      }
    }
  }

  // Start cache cleanup
  private startCacheCleanup(): void {
    if (browser) {
      setInterval(() => {
        this.cleanupExpiredCache();
      }, 5 * 60 * 1000); // Every 5 minutes
    }
  }

  // Clean up expired cache entries
  private cleanupExpiredCache(): void {
    const now = Date.now();
    const maxAge = 10 * 60 * 1000; // 10 minutes

    for (const [key, value] of this.cache.entries()) {
      if (now - value.timestamp > maxAge) {
        this.cache.delete(key);
      }
    }
  }

  // Add rate limiting check
  private async rateLimitCheck(): Promise<void> {
    const now = Date.now();
    const timeSinceLastRequest = now - this.lastRequestTime;
    
    if (timeSinceLastRequest < this.RATE_LIMIT_DELAY) {
      const waitTime = this.RATE_LIMIT_DELAY - timeSinceLastRequest;
      console.log(`🚦 Rate limiting: waiting ${waitTime}ms`);
      await new Promise(resolve => setTimeout(resolve, waitTime));
    }
    
    this.lastRequestTime = Date.now();
  }

  // Enhanced queue processing with better rate limiting
  private async processPreloadQueue(): Promise<void> {
    if (this.processingQueue || this.preloadQueue.size === 0) return;
    
    this.processingQueue = true;
    console.log(`📦 Processing preload queue: ${this.preloadQueue.size} items`);
    
    try {
      // Convert to array and take only a small batch
      const queueArray = Array.from(this.preloadQueue);
      const batch = queueArray.slice(0, this.BATCH_SIZE);
      
      // Remove processed items from queue
      batch.forEach(videoId => this.preloadQueue.delete(videoId));
      
      // Apply rate limiting
      await this.rateLimitCheck();
      
      // Process batch with controlled concurrency
      const results = await Promise.allSettled(
        batch.map(videoId => this.preloadVideoMetadata(videoId))
      );
      
      // Log results
      const successful = results.filter(r => r.status === 'fulfilled').length;
      const failed = results.filter(r => r.status === 'rejected').length;
      
      console.log(`📊 Batch complete: ${successful} successful, ${failed} failed`);
      
      // If there are more items, schedule next batch with delay
      if (this.preloadQueue.size > 0) {
        console.log(`⏳ Scheduling next batch in ${this.RATE_LIMIT_DELAY}ms`);
        setTimeout(() => this.processPreloadQueue(), this.RATE_LIMIT_DELAY);
      }
      
    } catch (error) {
      console.error('❌ Error processing preload queue:', error);
    } finally {
      this.processingQueue = false;
    }
  }

  // Enhanced preload with better error handling
  private async preloadVideoMetadata(videoId: string): Promise<void> {
    if (this.cache.has(videoId)) {
      console.log(`📋 Video ${videoId} already cached`);
      return;
    }

    try {
      console.log(`🔄 Preloading video metadata: ${videoId}`);
      
      const response = await fetch(`/api/v1/bunny-videos/${videoId}`, {
        method: 'GET',
        headers: {
          'Accept': 'application/json',
          'Cache-Control': 'no-cache'
        }
      });

      if (!response.ok) {
        if (response.status === 429) {
          console.warn(`⚠️ Rate limited for video ${videoId}, will retry later`);
          // Add back to queue for later retry
          this.preloadQueue.add(videoId);
          throw new Error(`Rate limited: ${response.status}`);
        }
        throw new Error(`Failed to fetch video metadata: ${response.status}`);
      }

      const videoData = await response.json();
      
      // Cache the data
      this.cache.set(videoId, {
        data: videoData,
        timestamp: Date.now(),
        accessCount: 0
      });

      this.performanceMetrics.preloadSuccess++;
      console.log(`✅ Successfully preloaded video: ${videoId}`);
      
    } catch (error) {
      this.performanceMetrics.preloadFailures++;
      console.warn(`❌ Video metadata preload failed: ${error}`);
      
      // Don't re-add to queue if it's a rate limit error (already handled above)
      if (error instanceof Error && !error.message.includes('Rate limited')) {
        // For other errors, we might want to retry later
        console.log(`🔄 Will retry ${videoId} later`);
      }
    }
  }

  // Optimized intersection observer with reduced sensitivity
  optimizeVideo(node: HTMLElement, videoId: string) {
    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach(entry => {
          // Only preload when element is closer to viewport
          if (entry.intersectionRatio > 0.1) { // Increased threshold
            console.log(`👁️ Video ${videoId} entering viewport, queuing for preload`);
            this.addToPreloadQueue(videoId);
          }
        });
      },
      {
        root: null,
        rootMargin: '200px', // Reduced from 300px
        threshold: [0.1, 0.5] // More specific thresholds
      }
    );

    observer.observe(node);

    return {
      destroy() {
        observer.disconnect();
      }
    };
  }

  // Throttled queue addition
  private addToPreloadQueue(videoId: string): void {
    if (this.preloadQueue.has(videoId)) return;
    
    this.preloadQueue.add(videoId);
    console.log(`📝 Added ${videoId} to preload queue (${this.preloadQueue.size} total)`);
    
    // Debounced processing with longer delay
    setTimeout(() => {
      if (!this.processingQueue) {
        this.processPreloadQueue();
      }
    }, this.PRELOAD_DELAY);
  }

  // Get cached video data
  getCachedVideo(videoId: string): any | null {
    const cached = this.cache.get(videoId);
    if (cached) {
      cached.accessCount++;
      this.performanceMetrics.cacheHits++;
      return cached.data;
    }
    this.performanceMetrics.cacheMisses++;
    return null;
  }

  // Get performance metrics
  getPerformanceMetrics(): PerformanceMetrics {
    return { ...this.performanceMetrics };
  }

  // Clear cache
  clearCache(): void {
    this.cache.clear();
    console.log('🧹 Video cache cleared');
  }
}

// Create singleton instance
let videoOptimizationService: VideoOptimizationService | null = null;

export function getVideoOptimizationService(): VideoOptimizationService {
  if (!videoOptimizationService) {
    videoOptimizationService = new VideoOptimizationService();
  }
  return videoOptimizationService;
}

// Svelte action for video optimization
export function optimizeVideo(node: HTMLElement, videoId: string) {
  const service = getVideoOptimizationService();
  return service.optimizeVideo(node, videoId);
}

// Preload video metadata
export async function preloadVideo(videoId: string): Promise<any> {
  const service = getVideoOptimizationService();
  return service.getCachedVideo(videoId);
}

// Get performance insights
export function getPerformanceMetrics(): PerformanceMetrics {
  const service = getVideoOptimizationService();
  return service.getPerformanceMetrics();
} 