import { onMount } from 'svelte';
import { goto } from '$app/navigation';
import { auth, initializeAuth, isAdmin } from '$lib/auth';
import { api } from '$lib/api';
import { showToast } from '$lib/toast';

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
	private static readonly BASE_PATH = '/admin/streaming/analytics';

	/**
	 * Get analytics overview data
	 */
	static async getAnalyticsOverview(params: AnalyticsPeriod): Promise<AnalyticsOverview> {
		try {
			const queryParams = new URLSearchParams();
			queryParams.append('period', params.period);
			queryParams.append('metric', params.metric);

			const response = await api.get(`${this.BASE_PATH}/overview?${queryParams}`);
			
			if (response.data) {
				return response.data as AnalyticsOverview;
			} else {
				throw new Error(response.error || 'Failed to load analytics data');
			}
		} catch (error) {
			console.error('Error fetching analytics overview:', error);
			throw error;
		}
	}

	/**
	 * Get customer analytics
	 */
	static async getCustomerMetrics(): Promise<CustomerMetrics> {
		try {
			const response = await api.get(`${this.BASE_PATH}/customers`);
			
			if (response.data) {
				return response.data as CustomerMetrics;
			} else {
				throw new Error(response.error || 'Failed to load customer metrics');
			}
		} catch (error) {
			console.error('Error fetching customer metrics:', error);
			throw error;
		}
	}

	/**
	 * Get plan analytics
	 */
	static async getPlanMetrics(): Promise<PlanMetrics[]> {
		try {
			const response = await api.get(`${this.BASE_PATH}/plans`);
			
			if (response.data) {
				return (response.data as { plans: PlanMetrics[] }).plans;
			} else {
				throw new Error(response.error || 'Failed to load plan metrics');
			}
		} catch (error) {
			console.error('Error fetching plan metrics:', error);
			throw error;
		}
	}

	/**
	 * Get revenue analytics
	 */
	static async getRevenueBreakdown(period: '7d' | '30d' | '90d' | '1y'): Promise<RevenueData[]> {
		try {
			const queryParams = new URLSearchParams();
			queryParams.append('period', period);

			const response = await api.get(`${this.BASE_PATH}/revenue?${queryParams}`);
			
			if (response.data) {
				return (response.data as { revenue_data: RevenueData[] }).revenue_data;
			} else {
				throw new Error(response.error || 'Failed to load revenue breakdown');
			}
		} catch (error) {
			console.error('Error fetching revenue breakdown:', error);
			throw error;
		}
	}

	/**
	 * Get subscription analytics
	 */
	static async getSubscriptionGrowth(period: '7d' | '30d' | '90d' | '1y'): Promise<SubscriptionData[]> {
		try {
			const queryParams = new URLSearchParams();
			queryParams.append('period', period);

			const response = await api.get(`${this.BASE_PATH}/subscriptions?${queryParams}`);
			
			if (response.data) {
				return (response.data as { subscription_data: SubscriptionData[] }).subscription_data;
			} else {
				throw new Error(response.error || 'Failed to load subscription growth');
			}
		} catch (error) {
			console.error('Error fetching subscription growth:', error);
			throw error;
		}
	}

	/**
	 * Get churn analytics
	 */
	static async getChurnAnalysis(period: '7d' | '30d' | '90d' | '1y'): Promise<ChurnData[]> {
		try {
			const queryParams = new URLSearchParams();
			queryParams.append('period', period);

			const response = await api.get(`${this.BASE_PATH}/churn?${queryParams}`);
			
			if (response.data) {
				return (response.data as { churn_data: ChurnData[] }).churn_data;
			} else {
				throw new Error(response.error || 'Failed to load churn analysis');
			}
		} catch (error) {
			console.error('Error fetching churn analysis:', error);
			throw error;
		}
	}

	/**
	 * Export analytics data
	 */
	static async exportAnalyticsData(params: {
		format: 'csv' | 'json' | 'excel';
		period: string;
		metrics?: string[];
	}): Promise<Blob> {
		try {
			const queryParams = new URLSearchParams();
			queryParams.append('format', params.format);
			queryParams.append('period', params.period);
			
			if (params.metrics && params.metrics.length > 0) {
				params.metrics.forEach(metric => {
					queryParams.append('metrics', metric);
				});
			}

			const response = await api.get(`${this.BASE_PATH}/export?${queryParams}`);
			
			if (response.data) {
				return response.data as Blob;
			} else {
				throw new Error('Failed to export analytics data');
			}
		} catch (error) {
			console.error('Error exporting analytics data:', error);
			throw error;
		}
	}

	/**
	 * Get real-time analytics
	 */
	static async getRealTimeAnalytics(): Promise<{
		active_users: number;
		revenue_today: number;
		new_subscriptions_today: number;
		cancellations_today: number;
	}> {
		try {
			const response = await api.get(`${this.BASE_PATH}/realtime`);
			
			if (response.data) {
				return response.data as {
					active_users: number;
					revenue_today: number;
					new_subscriptions_today: number;
					cancellations_today: number;
				};
			} else {
				throw new Error(response.error || 'Failed to load real-time metrics');
			}
		} catch (error) {
			console.error('Error fetching real-time analytics:', error);
			throw error;
		}
	}
} 
