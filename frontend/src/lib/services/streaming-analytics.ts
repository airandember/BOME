import { api } from '$lib/api';

export interface AnalyticsPeriod {
	period: '7d' | '30d' | '90d' | '1y';
	metric: 'revenue' | 'subscriptions' | 'churn';
}

export interface RevenueData {
	date: string;
	amount: number;
	currency: string;
	subscriptions: number;
}

export interface SubscriptionData {
	date: string;
	new_subscriptions: number;
	cancelled_subscriptions: number;
	active_subscriptions: number;
}

export interface ChurnData {
	date: string;
	churn_rate: number;
	cancelled_count: number;
	total_subscribers: number;
}

export interface AnalyticsOverview {
	revenue_trend: RevenueData[];
	subscription_trend: SubscriptionData[];
	churn_trend: ChurnData[];
	summary: {
		total_revenue: number;
		total_subscriptions: number;
		avg_churn_rate: number;
		growth_rate: number;
	};
}

export interface CustomerMetrics {
	total_customers: number;
	active_customers: number;
	new_customers_this_month: number;
	customer_lifetime_value: number;
	avg_subscription_duration: number;
}

export interface PlanMetrics {
	plan_name: string;
	subscribers: number;
	revenue: number;
	churn_rate: number;
	popularity_rank: number;
}

export class StreamingAnalyticsService {
	/**
	 * Get analytics overview data
	 */
	static async getAnalyticsOverview(params: AnalyticsPeriod): Promise<AnalyticsOverview> {
		try {
			const queryParams = new URLSearchParams({
				period: params.period,
				metric: params.metric
			});

			const response = await api.get(`/api/admin/streaming/analytics/overview?${queryParams}`);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to load analytics data');
			}
		} catch (error) {
			console.error('Error fetching analytics overview:', error);
			throw error;
		}
	}

	/**
	 * Get customer metrics
	 */
	static async getCustomerMetrics(): Promise<CustomerMetrics> {
		try {
			const response = await api.get('/api/admin/streaming/analytics/customers');
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to load customer metrics');
			}
		} catch (error) {
			console.error('Error fetching customer metrics:', error);
			throw error;
		}
	}

	/**
	 * Get plan performance metrics
	 */
	static async getPlanMetrics(): Promise<PlanMetrics[]> {
		try {
			const response = await api.get('/api/admin/streaming/analytics/plans');
			
			if (response.data) {
				return response.data.plans;
			} else {
				throw new Error(response.error || 'Failed to load plan metrics');
			}
		} catch (error) {
			console.error('Error fetching plan metrics:', error);
			throw error;
		}
	}

	/**
	 * Get revenue breakdown by period
	 */
	static async getRevenueBreakdown(period: '7d' | '30d' | '90d' | '1y'): Promise<RevenueData[]> {
		try {
			const response = await api.get(`/api/admin/streaming/analytics/revenue?period=${period}`);
			
			if (response.data) {
				return response.data.revenue_data;
			} else {
				throw new Error(response.error || 'Failed to load revenue breakdown');
			}
		} catch (error) {
			console.error('Error fetching revenue breakdown:', error);
			throw error;
		}
	}

	/**
	 * Get subscription growth data
	 */
	static async getSubscriptionGrowth(period: '7d' | '30d' | '90d' | '1y'): Promise<SubscriptionData[]> {
		try {
			const response = await api.get(`/api/admin/streaming/analytics/subscriptions?period=${period}`);
			
			if (response.data) {
				return response.data.subscription_data;
			} else {
				throw new Error(response.error || 'Failed to load subscription growth data');
			}
		} catch (error) {
			console.error('Error fetching subscription growth:', error);
			throw error;
		}
	}

	/**
	 * Get churn analysis data
	 */
	static async getChurnAnalysis(period: '7d' | '30d' | '90d' | '1y'): Promise<ChurnData[]> {
		try {
			const response = await api.get(`/api/admin/streaming/analytics/churn?period=${period}`);
			
			if (response.data) {
				return response.data.churn_data;
			} else {
				throw new Error(response.error || 'Failed to load churn analysis');
			}
		} catch (error) {
			console.error('Error fetching churn analysis:', error);
			throw error;
		}
	}

	/**
	 * Export analytics report
	 */
	static async exportAnalyticsReport(params: {
		period: '7d' | '30d' | '90d' | '1y';
		format: 'csv' | 'excel' | 'pdf';
		include_charts?: boolean;
	}): Promise<Blob> {
		try {
			const queryParams = new URLSearchParams({
				period: params.period,
				format: params.format,
				include_charts: params.include_charts?.toString() || 'false'
			});

			const response = await api.get(`/api/admin/streaming/analytics/export?${queryParams}`, {
				responseType: 'blob'
			});
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to export analytics report');
			}
		} catch (error) {
			console.error('Error exporting analytics report:', error);
			throw error;
		}
	}

	/**
	 * Get real-time metrics
	 */
	static async getRealTimeMetrics(): Promise<{
		active_users: number;
		revenue_today: number;
		new_subscriptions_today: number;
		cancellations_today: number;
	}> {
		try {
			const response = await api.get('/api/admin/streaming/analytics/realtime');
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to load real-time metrics');
			}
		} catch (error) {
			console.error('Error fetching real-time metrics:', error);
			throw error;
		}
	}
} 