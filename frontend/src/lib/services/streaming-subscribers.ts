import { api } from '$lib/api';

export interface Subscriber {
	id: number;
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	email_verified: boolean;
	plan_id?: number;
	plan_name?: string;
	plan_price?: number;
	plan_currency?: string;
	subscription_id?: number;
	subscription_status?: string;
	current_period_start?: string;
	current_period_end?: string;
	stripe_customer_id?: string;
	stripe_subscription_id?: string;
	last_login?: string;
	created_at: string;
	updated_at: string;
}

export interface SubscriberStats {
	total_subscribers: number;
	active_subscribers: number;
	trialing_subscribers: number;
	past_due_subscribers: number;
	canceled_subscribers: number;
	monthly_revenue: number;
	annual_revenue: number;
	average_revenue_per_user: number;
	churn_rate: number;
}

export interface SubscriberFilters {
	plan_id?: number;
	status?: string;
	search?: string;
	email_verified?: boolean;
	date_range?: {
		start: string;
		end: string;
	};
}

export interface SubscribersResponse {
	subscribers: Subscriber[];
	pagination: {
		limit: number;
		offset: number;
	};
}

export class StreamingSubscriberService {
	/**
	 * Get all subscribers with optional filters and pagination
	 */
	static async getSubscribers(params?: {
		limit?: number;
		offset?: number;
		filters?: SubscriberFilters;
	}): Promise<SubscribersResponse> {
		try {
			const queryParams = new URLSearchParams();
			
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());
			
			if (params?.filters) {
				if (params.filters.plan_id) queryParams.append('plan_id', params.filters.plan_id.toString());
				if (params.filters.status) queryParams.append('status', params.filters.status);
				if (params.filters.search) queryParams.append('search', params.filters.search);
				if (params.filters.email_verified !== undefined) queryParams.append('email_verified', params.filters.email_verified.toString());
				if (params.filters.date_range) {
					queryParams.append('start_date', params.filters.date_range.start);
					queryParams.append('end_date', params.filters.date_range.end);
				}
			}

			const response = await api.get(`/admin/subscribers?${queryParams}`);
			
			if (response.data) {
				return response.data as SubscribersResponse;
			} else {
				throw new Error(response.error || 'Failed to load subscribers');
			}
		} catch (error) {
			console.error('Error fetching subscribers:', error);
			throw error;
		}
	}

	/**
	 * Get a single subscriber by ID
	 */
	static async getSubscriber(id: number): Promise<Subscriber> {
		try {
			const response = await api.get(`/admin/subscribers/${id}`);
			
			if (response.data) {
				return (response.data as { subscriber: Subscriber }).subscriber;
			} else {
				throw new Error(response.error || 'Failed to load subscriber');
			}
		} catch (error) {
			console.error('Error fetching subscriber:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber count with optional filters
	 */
	static async getSubscriberCount(filters?: SubscriberFilters): Promise<number> {
		try {
			const queryParams = new URLSearchParams();
			
			if (filters?.plan_id) queryParams.append('plan_id', filters.plan_id.toString());
			if (filters?.status) queryParams.append('status', filters.status);
			if (filters?.search) queryParams.append('search', filters.search);
			if (filters?.email_verified !== undefined) queryParams.append('email_verified', filters.email_verified.toString());
			if (filters?.date_range) {
				queryParams.append('start_date', filters.date_range.start);
				queryParams.append('end_date', filters.date_range.end);
			}

			const response = await api.get(`/admin/subscribers/count?${queryParams}`);
			
			if (response.data) {
				return (response.data as { count: number }).count;
			} else {
				throw new Error(response.error || 'Failed to get subscriber count');
			}
		} catch (error) {
			console.error('Error fetching subscriber count:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber statistics
	 */
	static async getSubscriberStats(): Promise<SubscriberStats> {
		try {
			const response = await api.get('/admin/subscribers/stats');
			
			if (response.data) {
				return response.data as SubscriberStats;
			} else {
				throw new Error(response.error || 'Failed to get subscriber stats');
			}
		} catch (error) {
			console.error('Error fetching subscriber stats:', error);
			throw error;
		}
	}

	/**
	 * Get subscribers by plan
	 */
	static async getSubscribersByPlan(planId: number, params?: {
		limit?: number;
		offset?: number;
	}): Promise<SubscribersResponse> {
		try {
			const queryParams = new URLSearchParams();
			
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());

			const response = await api.get(`/admin/subscribers/plan/${planId}?${queryParams}`);
			
			if (response.data) {
				return response.data as SubscribersResponse;
			} else {
				throw new Error(response.error || 'Failed to get subscribers by plan');
			}
		} catch (error) {
			console.error('Error fetching subscribers by plan:', error);
			throw error;
		}
	}

	/**
	 * Get subscribers by status
	 */
	static async getSubscribersByStatus(status: string, params?: {
		limit?: number;
		offset?: number;
	}): Promise<SubscribersResponse> {
		try {
			const queryParams = new URLSearchParams();
			
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());

			const response = await api.get(`/admin/subscribers/status/${status}?${queryParams}`);
			
			if (response.data) {
				return response.data as SubscribersResponse;
			} else {
				throw new Error(response.error || 'Failed to get subscribers by status');
			}
		} catch (error) {
			console.error('Error fetching subscribers by status:', error);
			throw error;
		}
	}

	/**
	 * Search subscribers
	 */
	static async searchSubscribers(searchTerm: string, params?: {
		limit?: number;
		offset?: number;
	}): Promise<SubscribersResponse> {
		try {
			const queryParams = new URLSearchParams();
			
			queryParams.append('q', searchTerm);
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());

			const response = await api.get(`/admin/subscribers/search?${queryParams}`);
			
			if (response.data) {
				return response.data as SubscribersResponse;
			} else {
				throw new Error(response.error || 'Failed to search subscribers');
			}
		} catch (error) {
			console.error('Error searching subscribers:', error);
			throw error;
		}
	}

	/**
	 * Format subscriber name
	 */
	static formatSubscriberName(subscriber: Subscriber): string {
		return `${subscriber.first_name} ${subscriber.last_name}`.trim() || subscriber.email;
	}

	/**
	 * Get subscriber status badge info
	 */
	static getSubscriberStatusInfo(status?: string): { text: string; class: string; color: string } {
		switch (status) {
			case 'active':
				return { text: 'Active', class: 'bg-green-100 text-green-800', color: 'green' };
			case 'trialing':
				return { text: 'Trialing', class: 'bg-blue-100 text-blue-800', color: 'blue' };
			case 'past_due':
				return { text: 'Past Due', class: 'bg-yellow-100 text-yellow-800', color: 'yellow' };
			case 'canceled':
				return { text: 'Canceled', class: 'bg-red-100 text-red-800', color: 'red' };
			case 'incomplete':
				return { text: 'Incomplete', class: 'bg-gray-100 text-gray-800', color: 'gray' };
			default:
				return { text: 'Unknown', class: 'bg-gray-100 text-gray-800', color: 'gray' };
		}
	}

	/**
	 * Format currency
	 */
	static formatCurrency(amount?: number, currency: string = 'USD'): string {
		if (amount === undefined || amount === null) return 'N/A';
		
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(amount);
	}

	/**
	 * Format date
	 */
	static formatDate(dateString?: string): string {
		if (!dateString) return 'N/A';
		
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	/**
	 * Check if subscription is active
	 */
	static isSubscriptionActive(subscriber: Subscriber): boolean {
		return subscriber.subscription_status === 'active' || subscriber.subscription_status === 'trialing';
	}

	/**
	 * Check if subscription is past due
	 */
	static isSubscriptionPastDue(subscriber: Subscriber): boolean {
		return subscriber.subscription_status === 'past_due';
	}

	/**
	 * Check if subscription is canceled
	 */
	static isSubscriptionCanceled(subscriber: Subscriber): boolean {
		return subscriber.subscription_status === 'canceled';
	}
} 