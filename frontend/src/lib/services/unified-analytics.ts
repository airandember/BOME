import { apiRequest } from '$lib/auth';

export interface UnifiedAnalytics {
	// Core metrics
	total_users: number;
	total_videos: number;
	total_views: number;
	total_revenue: number;
	active_subscriptions: number;
	
	// Video-specific metrics
	video_stats: {
		total_videos: number;
		synced_videos: number;
		needs_attention: number;
		total_views: number;
		videos_by_status: Record<string, number>;
		videos_by_sync_status: Record<string, number>;
		pending_conflicts: number;
	};
	
	// View analytics
	view_analytics: {
		total_views: number;
		views_today: number;
		views_week: number;
		growth_rate: number;
	};
	
	// Subscriber metrics
	subscriber_metrics: {
		total_subscribers: number;
		active_subscriptions: number;
		monthly_revenue: number;
		churn_rate: number;
	};
	
	// Real-time metrics
	real_time: {
		active_users: number;
		recent_activity: Array<{
			type: string;
			message: string;
			time: string;
			user?: string;
		}>;
	};
}

export interface AnalyticsOptions {
	period?: '24h' | '7d' | '30d' | '90d';
	include_realtime?: boolean;
	include_activity?: boolean;
}

class UnifiedAnalyticsService {
	private static instance: UnifiedAnalyticsService;
	private cache = new Map<string, { data: any; timestamp: number }>();
	private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

	private constructor() {}

	static getInstance(): UnifiedAnalyticsService {
		if (!UnifiedAnalyticsService.instance) {
			UnifiedAnalyticsService.instance = new UnifiedAnalyticsService();
		}
		return UnifiedAnalyticsService.instance;
	}

	/**
	 * Get comprehensive analytics data in a single API call
	 * This consolidates all the separate analytics endpoints into one efficient call
	 */
	async getUnifiedAnalytics(options: AnalyticsOptions = {}): Promise<UnifiedAnalytics> {
		const cacheKey = `unified_analytics_${JSON.stringify(options)}`;
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const response = await apiRequest('/admin/analytics/unified', {
				method: 'POST',
				body: JSON.stringify(options)
			});

			if (!response.ok) {
				throw new Error('Failed to fetch unified analytics');
			}

			const data = await response.json();
			
			this.cache.set(cacheKey, {
				data: data.data,
				timestamp: Date.now()
			});

			return data.data;
		} catch (error) {
			console.error('Failed to fetch unified analytics:', error);
			throw error;
		}
	}

	/**
	 * Get basic stats (for simple dashboards that don't need full analytics)
	 */
	async getBasicStats(): Promise<{
		total_users: number;
		total_videos: number;
		total_views: number;
		total_revenue: number;
	}> {
		const cacheKey = 'basic_stats';
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const response = await apiRequest('/admin/analytics/basic');
			
			if (!response.ok) {
				throw new Error('Failed to fetch basic stats');
			}

			const data = await response.json();
			
			this.cache.set(cacheKey, {
				data: data.data,
				timestamp: Date.now()
			});

			return data.data;
		} catch (error) {
			console.error('Failed to fetch basic stats:', error);
			throw error;
		}
	}

	/**
	 * Get video-specific analytics (for streaming dashboard)
	 */
	async getVideoAnalytics(): Promise<{
		video_stats: UnifiedAnalytics['video_stats'];
		view_analytics: UnifiedAnalytics['view_analytics'];
	}> {
		const cacheKey = 'video_analytics';
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const response = await apiRequest('/admin/analytics/video');
			
			if (!response.ok) {
				throw new Error('Failed to fetch video analytics');
			}

			const data = await response.json();
			
			this.cache.set(cacheKey, {
				data: data.data,
				timestamp: Date.now()
			});

			return data.data;
		} catch (error) {
			console.error('Failed to fetch video analytics:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber metrics (for streaming dashboard)
	 */
	async getSubscriberMetrics(): Promise<{
		total_subscribers: number;
		active_subscriptions: number;
		monthly_revenue: number;
		churn_rate: number;
	}> {
		const cacheKey = 'subscriber_metrics';
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const data = await this.getUnifiedAnalytics();
			
			this.cache.set(cacheKey, {
				data: data.subscriber_metrics,
				timestamp: Date.now()
			});

			return data.subscriber_metrics;
		} catch (error) {
			console.error('Failed to fetch subscriber metrics:', error);
			return {
				total_subscribers: 0,
				active_subscriptions: 0,
				monthly_revenue: 0,
				churn_rate: 0
			};
		}
	}

	/**
	 * Get dashboard stats (for admin dashboard)
	 */
	async getDashboardStats(): Promise<{
		success: boolean;
		data: UnifiedAnalytics;
	}> {
		try {
			const data = await this.getUnifiedAnalytics({
				period: '7d',
				include_realtime: true,
				include_activity: true
			});
			
			return {
				success: true,
				data
			};
		} catch (error) {
			console.error('Failed to fetch dashboard stats:', error);
			return {
				success: false,
				data: {
					total_users: 0,
					total_videos: 0,
					total_views: 0,
					total_revenue: 0,
					active_subscriptions: 0,
					video_stats: {
						total_videos: 0,
						synced_videos: 0,
						needs_attention: 0,
						total_views: 0,
						videos_by_status: {},
						videos_by_sync_status: {},
						pending_conflicts: 0
					},
					view_analytics: {
						total_views: 0,
						views_today: 0,
						views_week: 0,
						growth_rate: 0
					},
					subscriber_metrics: {
						total_subscribers: 0,
						active_subscriptions: 0,
						monthly_revenue: 0,
						churn_rate: 0
					},
					real_time: {
						active_users: 0,
						recent_activity: []
					}
				}
			};
		}
	}

	/**
	 * Clear cache (useful for testing or when data needs to be refreshed)
	 */
	clearCache(): void {
		this.cache.clear();
	}

	/**
	 * Clear specific cache entry
	 */
	clearCacheEntry(key: string): void {
		this.cache.delete(key);
	}
}

// Export singleton instance
export const unifiedAnalyticsService = UnifiedAnalyticsService.getInstance(); 
