import { api } from '$lib/api';
import type { EnhancedSubscriber, SubscriberFilters, SubscriberResponse, SubscriberKPIs } from '$lib/types/enhanced-subscriber';

/**
 * Smart caching layer for subscriber data
 * Reduces API calls and improves performance across streaming tabs
 */
export class SubscriberCache {
  private cache = new Map<string, SubscriberResponse>();
  private lastFetch = new Map<string, number>();
  private kpisCache: SubscriberKPIs | null = null;
  private kpisLastFetch = 0;
  
  // Cache TTL: 5 minutes for data, 2 minutes for KPIs
  private readonly DATA_CACHE_TTL = 5 * 60 * 1000;
  private readonly KPIS_CACHE_TTL = 2 * 60 * 1000;
  
  /**
   * Get subscribers with smart caching
   */
  async getSubscribers(
    page = 1, 
    limit = 50, 
    filters: SubscriberFilters = {}
  ): Promise<SubscriberResponse> {
    const cacheKey = this.generateCacheKey({ page, limit, ...filters });
    const now = Date.now();
    
    // Check if we have valid cached data
    if (this.cache.has(cacheKey)) {
      const lastFetch = this.lastFetch.get(cacheKey) || 0;
      if ((now - lastFetch) < this.DATA_CACHE_TTL) {
        console.log('📦 Using cached subscriber data');
        return this.cache.get(cacheKey)!;
      }
    }
    
    // Fetch fresh data
    console.log('🔄 Fetching fresh subscriber data');
    try {
      // Build query parameters
      const params = new URLSearchParams({
        page: page.toString(),
        limit: limit.toString(),
        ...this.serializeFilters(filters)
      });
      
      const response = await api.get(`/admin/subscribers/enhanced?${params}`);
      
      // Check if response is an error (handle multiple error formats)
      if (response && typeof response === 'object') {
        if ('error' in response) {
          throw new Error(response.error as string);
        }
        // Handle middleware auth errors that might have different format
        if ('message' in response && typeof response.message === 'string' && response.message.includes('Authorization')) {
          throw new Error(response.message as string);
        }
        // Handle case where response is null or undefined
        if (response === null || response === undefined) {
          throw new Error('No response received from server');
        }
      }
      
      // Log the actual response for debugging
      console.log('Enhanced subscriber response:', response);
      
      // Handle both enhanced format and old format wrapped in 'data'
      if (response && typeof response === 'object') {
        let actualData: any;
        
        // The API client wraps responses in { data: ... }, so check for that first
        if ('data' in response && response.data && typeof response.data === 'object') {
          actualData = response.data;
        }
        // Check if it's the enhanced format (direct) - fallback
        else if ('subscribers' in response) {
          actualData = response;
        }
        
        if (actualData && 'subscribers' in actualData) {
          const data = actualData as SubscriberResponse;
          
          // Deduplicate subscribers by ID to prevent Svelte each_key_duplicate errors
          if (data.subscribers && Array.isArray(data.subscribers)) {
            const seen = new Set();
            data.subscribers = data.subscribers.filter(subscriber => {
              if (seen.has(subscriber.id)) {
                console.warn(`Duplicate subscriber found with ID: ${subscriber.id}`);
                return false;
              }
              seen.add(subscriber.id);
              return true;
            });
          }
          
          // Cache the response
          this.cache.set(cacheKey, data);
          this.lastFetch.set(cacheKey, now);
          
          // Also cache KPIs separately
          if (data.kpis) {
            this.kpisCache = data.kpis;
            this.kpisLastFetch = now;
          }
          
          return data;
        }
      }
      
      throw new Error('Invalid response format');
    } catch (error) {
      console.error('Failed to fetch subscriber data:', error);
      throw error;
    }
  }
  
  /**
   * Get KPIs with separate caching (more frequent updates)
   */
  async getKPIs(): Promise<SubscriberKPIs> {
    const now = Date.now();
    
    // Check if we have valid cached KPIs
    if (this.kpisCache && (now - this.kpisLastFetch) < this.KPIS_CACHE_TTL) {
      console.log('📊 Using cached KPIs');
      return this.kpisCache;
    }
    
    // Fetch fresh KPIs
    console.log('🔄 Fetching fresh KPIs');
    try {
      const response = await api.get('/admin/subscribers/kpis');

      // Log the actual response for debugging
      console.log('Enhanced subscriber KPIs response:', response);

      // Check if response is an error (handle multiple error formats)
      if (response && typeof response === 'object') {
        if ('error' in response) {
          throw new Error(response.error as string);
        }
        // Handle middleware auth errors that might have different format
        if ('message' in response && typeof response.message === 'string' && response.message.includes('Authorization')) {
          throw new Error(response.message as string);
        }
      }

      if (response && typeof response === 'object') {
        let actualKpis: any;
        
        // The API client wraps responses in { data: ... }, so check for that first
        if ('data' in response && response.data && typeof response.data === 'object') {
          actualKpis = response.data;
        }
        // Check if it's the enhanced format (direct KPIs) - fallback
        else if ('total_subscribers' in response || 'active_subscribers' in response) {
          actualKpis = response;
        }
        
        if (actualKpis) {
          const kpis = actualKpis as SubscriberKPIs;
          this.kpisCache = kpis;
          this.kpisLastFetch = now;
          return kpis;
        }
      }
      
      throw new Error('Invalid KPIs response format');
    } catch (error) {
      console.error('Failed to fetch KPIs:', error);
      throw error;
    }
  }
  
  /**
   * Invalidate cache when data changes
   */
  invalidate(pattern?: string): void {
    if (pattern) {
      // Invalidate specific cache entries matching pattern
      for (const key of this.cache.keys()) {
        if (key.includes(pattern)) {
          this.cache.delete(key);
          this.lastFetch.delete(key);
        }
      }
    } else {
      // Invalidate all cache
      this.cache.clear();
      this.lastFetch.clear();
      this.kpisCache = null;
      this.kpisLastFetch = 0;
    }
    console.log('🗑️ Cache invalidated');
  }
  
  /**
   * Preload common data combinations
   */
  async preloadCommonViews(): Promise<void> {
    const commonFilters = [
      { has_active_plan: true },
      { has_video_access: true },
      { plan_type: 'premium' as const },
      { is_expiring_soon: true },
      { video_access_source: 'manual' as const }
    ];
    
    // Preload in background without blocking
    Promise.all(
      commonFilters.map(filter => 
        this.getSubscribers(1, 50, filter).catch(console.error)
      )
    );
  }
  
  /**
   * Get cache statistics for debugging
   */
  getCacheStats(): { entries: number; size: string; hitRate: number } {
    const entries = this.cache.size;
    const size = `${Math.round(JSON.stringify([...this.cache.values()]).length / 1024)}KB`;
    
    // Simple hit rate calculation (would need more sophisticated tracking in production)
    const hitRate = entries > 0 ? 0.85 : 0; // Placeholder
    
    return { entries, size, hitRate };
  }
  
  /**
   * Generate cache key from parameters
   */
  private generateCacheKey(params: any): string {
    return JSON.stringify(params, Object.keys(params).sort());
  }
  
  /**
   * Serialize filters for API call
   */
  private serializeFilters(filters: SubscriberFilters): Record<string, any> {
    const serialized: Record<string, any> = {};
    
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        serialized[key] = value;
      }
    });
    
    return serialized;
  }
}

// Singleton instance for global use
export const subscriberCache = new SubscriberCache();

// Preload common views on initialization
if (typeof window !== 'undefined') {
  subscriberCache.preloadCommonViews();
}
