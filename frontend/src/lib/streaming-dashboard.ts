import { apiRequest } from '$lib/auth';

export interface DashboardData {
	active_users: number;
	recent_activity: Activity[];
	view_analytics: ViewAnalytics;
	subscriber_metrics: SubscriberMetrics;
	video_stats: VideoStats;
}

export interface Activity {
	type: string;
	user_id: number;
	action: string;
	details: string;
	created_at: string;
}

export interface ViewAnalytics {
	total_views: number;
	views_today: number;
	views_week: number;
	growth_rate: number;
}

export interface SubscriberMetrics {
	total_subscribers: number;
	active_subscriptions: number;
	monthly_revenue: number;
	churn_rate: number;
}

export interface VideoStats {
	total_videos: number;
	synced_videos: number;
	needs_attention: number;
	total_views: number;
}

export interface ActiveUsersData {
	active_users: number;
	trend: Array<{
		hour: string;
		active_users: number;
	}>;
}

export interface ViewAnalyticsPeriod {
	period: string;
	total_views: number;
	daily_views: Array<{
		date: string;
		views: number;
	}>;
}

class StreamingDashboardService {
	private baseUrl = '/admin/streaming';

	async getDashboardData(): Promise<DashboardData> {
		const response = await apiRequest(`${this.baseUrl}/dashboard`);
		
		if (!response.ok) {
			throw new Error('Failed to fetch dashboard data');
		}

		const data = await response.json();
		return data.data;
	}

	async getActiveUsers(): Promise<ActiveUsersData> {
		const response = await apiRequest(`${this.baseUrl}/active-users`);
		
		if (!response.ok) {
			throw new Error('Failed to fetch active users data');
		}

		const data = await response.json();
		return data.data;
	}

	async getRecentActivity(limit: number = 20): Promise<Activity[]> {
		const response = await apiRequest(`${this.baseUrl}/recent-activity?limit=${limit}`);
		
		if (!response.ok) {
			throw new Error('Failed to fetch recent activity');
		}

		const data = await response.json();
		return data.data;
	}

	async getViewAnalytics(period: string = '7d'): Promise<ViewAnalyticsPeriod> {
		const response = await apiRequest(`${this.baseUrl}/view-analytics?period=${period}`);
		
		if (!response.ok) {
			throw new Error('Failed to fetch view analytics');
		}

		const data = await response.json();
		return data.data;
	}

	async getSubscriberMetrics(): Promise<SubscriberMetrics> {
		const response = await apiRequest(`${this.baseUrl}/subscriber-metrics`);
		
		if (!response.ok) {
			throw new Error('Failed to fetch subscriber metrics');
		}

		const data = await response.json();
		return data.data;
	}
}

export const streamingDashboardService = new StreamingDashboardService(); 
