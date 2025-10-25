import { api } from '$lib/api';
import { subscriberElasticService, type UnifiedSubscriber } from '$lib/services/subscriber-elastic-service';
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
   * Get subscribers with smart caching using UNIFIED ELASTIC SERVICE
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
    
    // Fetch fresh data using UNIFIED ELASTIC SERVICE
    console.log('🔄 Fetching fresh subscriber data using UNIFIED ELASTIC SERVICE');
    try {
      // Get ALL subscribers from elastic service
      const unifiedSubscribers = await subscriberElasticService.getAllSubscribers();
      
      // Convert UnifiedSubscriber to EnhancedSubscriber format
      const allEnhancedSubscribers: EnhancedSubscriber[] = unifiedSubscribers.map(sub => ({
        id: sub.id,
        email: sub.email,
        first_name: sub.first_name,
        last_name: sub.last_name,
        role: sub.role,
        email_verified: sub.email_verified,
        is_active: sub.is_active,
        created_at: sub.created_at,
        last_login: sub.last_login,
        stripe_customer_id: sub.stripe_customer_id,
        stripe_customer_ids: sub.stripe_customer_ids,
        subscription_id: sub.subscription_id,
        plan_name: sub.plan_name || 'No Plan',
        plan_type: sub.plan_type,
        plan_status: sub.plan_status,
        plan_price: sub.plan_price,
        plan_currency: sub.plan_currency,
        plan_interval: sub.plan_interval,
        plan_start_date: sub.plan_start_date,
        billing_period_start: sub.billing_period_start,
        billing_period_end: sub.billing_period_end,
        days_until_expiry: sub.days_until_expiry,
        has_active_plan: sub.has_active_plan,
        has_video_access: sub.has_video_access,
        manual_access_granted: sub.manual_access_granted,
        mrr_contribution: sub.mrr_contribution,
        arr_contribution: sub.arr_contribution,
        ltv_estimate: sub.ltv_estimate,
        account_age_days: sub.account_age_days,
        plan_legacy_status: sub.plan_legacy_status,
        full_name: sub.full_name,
        is_expiring_soon: sub.is_expiring_soon,
        // Additional fields for compatibility
        is_high_value: sub.ltv_estimate > 1000,
        subscription_duration_days: sub.account_age_days,
        updated_at: sub.created_at // Use created_at as fallback
      }));
      
      // Apply filters
      let filteredSubscribers = allEnhancedSubscribers;
      if (Object.keys(filters).length > 0) {
        filteredSubscribers = this.applyFilters(allEnhancedSubscribers, filters);
      }
      
      // Apply pagination
      const startIndex = (page - 1) * limit;
      const endIndex = startIndex + limit;
      const paginatedSubscribers = filteredSubscribers.slice(startIndex, endIndex);
      
      // Create response in expected format
      const response: SubscriberResponse = {
        subscribers: paginatedSubscribers,
        total_count: filteredSubscribers.length,
        page: page,
        limit: limit,
        total_pages: Math.ceil(filteredSubscribers.length / limit),
        kpis: await this.generateKPIsFromSubscribers(allEnhancedSubscribers)
      };
      
      // Cache the response
      this.cache.set(cacheKey, response);
      this.lastFetch.set(cacheKey, now);
      
      // Also cache KPIs separately
      if (response.kpis) {
        this.kpisCache = response.kpis;
        this.kpisLastFetch = now;
      }
      
      console.log(`✅ Loaded ${paginatedSubscribers.length} subscribers using UNIFIED ELASTIC SERVICE (filtered from ${allEnhancedSubscribers.length} total)`);
      return response;
      
    } catch (error) {
      console.error('Failed to fetch subscriber data from elastic service:', error);
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
   * Apply filters to subscribers array
   */
  private applyFilters(subscribers: EnhancedSubscriber[], filters: SubscriberFilters): EnhancedSubscriber[] {
    return subscribers.filter(subscriber => {
      // Active plan filter
      if (filters.has_active_plan !== undefined) {
        if (filters.has_active_plan && !subscriber.has_active_plan) return false;
        if (!filters.has_active_plan && subscriber.has_active_plan) return false;
      }
      
      // Video access filter
      if (filters.has_video_access !== undefined) {
        if (filters.has_video_access && !subscriber.has_video_access) return false;
        if (!filters.has_video_access && subscriber.has_video_access) return false;
      }
      
      // Plan type filter
      if (filters.plan_type && subscriber.plan_type !== filters.plan_type) return false;
      
      // Expiring soon filter
      if (filters.is_expiring_soon !== undefined) {
        if (filters.is_expiring_soon && !subscriber.is_expiring_soon) return false;
        if (!filters.is_expiring_soon && subscriber.is_expiring_soon) return false;
      }
      
      // Video access source filter
      if (filters.video_access_source) {
        if (filters.video_access_source === 'manual' && !subscriber.manual_access_granted) return false;
        if (filters.video_access_source === 'subscription' && !subscriber.has_active_plan) return false;
      }
      
      // Search filter
      if (filters.search) {
        const searchTerm = filters.search.toLowerCase();
        const searchableText = [
          subscriber.email,
          subscriber.first_name,
          subscriber.last_name,
          subscriber.full_name,
          subscriber.plan_name
        ].join(' ').toLowerCase();
        
        if (!searchableText.includes(searchTerm)) return false;
      }
      
      return true;
    });
  }
  
  /**
   * Generate KPIs from subscribers array
   */
  private async generateKPIsFromSubscribers(subscribers: EnhancedSubscriber[]): Promise<SubscriberKPIs> {
    const totalSubscribers = subscribers.length;
    const activeSubscribers = subscribers.filter(s => s.has_active_plan).length;
    const videoAccessSubscribers = subscribers.filter(s => s.has_video_access).length;
    const expiringSoonSubscribers = subscribers.filter(s => s.is_expiring_soon).length;
    
    const totalMRR = subscribers.reduce((sum, s) => sum + (s.mrr_contribution || 0), 0);
    const totalARR = subscribers.reduce((sum, s) => sum + (s.arr_contribution || 0), 0);
    
    return {
      total_subscribers: totalSubscribers,
      active_subscribers: activeSubscribers,
      video_access_subscribers: videoAccessSubscribers,
      expiring_soon_subscribers: expiringSoonSubscribers,
      total_mrr: totalMRR,
      total_arr: totalARR,
      avg_mrr_per_subscriber: totalSubscribers > 0 ? totalMRR / totalSubscribers : 0,
      avg_arr_per_subscriber: totalSubscribers > 0 ? totalARR / totalSubscribers : 0
    };
  }
  
  /**
   * Serialize filters for API call (legacy method - kept for compatibility)
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
