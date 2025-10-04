import { api } from '$lib/api';
import { apiRequest } from '$lib/auth';

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
	plan_interval?: string; // 'month', 'year', etc.
	plan_interval_count?: number; // 1, 12, etc.
	subscription_id?: number;
	sub_id?: number; // Alias for subscription_id
	subscription_status?: string;
	current_period_start?: string;
	current_period_end?: string;
	stripe_customer_id?: string;
	stripe_subscription_id?: string;
	last_login?: string;
	created_at: string;
	updated_at: string;
}

export interface NonSubscriber {
	id: number;
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	email_verified: boolean;
	sub_id?: number; // Will be null for non-subscribers
	has_subscription_history?: boolean;
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
	role?: string;
	last_login?: string;
	created_date?: string;
	date_range?: {
		start: string;
		end: string;
	};
	sub_id?: number;
	has_subscription_history?: boolean; // true = has history, false = no history, undefined = all
}

export interface NonSubscriberFilters {
	search?: string;
	email_verified?: boolean;
	role?: string;
	last_login?: string;
	created_date?: string;
	date_range?: {
		start: string;
		end: string;
	};
	has_subscription_history?: boolean; // true, false, or undefined for all
}

export interface SubscribersResponse {
	subscribers: Subscriber[];
	pagination: {
		limit: number;
		offset: number;
		total: number;
	};
}

export interface NonSubscribersResponse {
	non_subscribers: NonSubscriber[];
	pagination: {
		limit: number;
		offset: number;
		total: number;
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
			console.log('StreamingSubscriberService.getSubscribers called with params:', params);
			
			const queryParams = new URLSearchParams();
			
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());
			
			if (params?.filters) {
				if (params.filters.plan_id) queryParams.append('plan_id', params.filters.plan_id.toString());
				if (params.filters.status) queryParams.append('status', params.filters.status);
				if (params.filters.search) queryParams.append('search', params.filters.search);
				if (params.filters.email_verified !== undefined) queryParams.append('email_verified', params.filters.email_verified.toString());
				if (params.filters.role) {
					console.log('Adding role filter:', params.filters.role);
					queryParams.append('role', params.filters.role);
				}
				if (params.filters.last_login) {
					console.log('Adding last_login filter:', params.filters.last_login);
					queryParams.append('last_login', params.filters.last_login);
				}
				if (params.filters.created_date) {
					console.log('Adding created_date filter:', params.filters.created_date);
					queryParams.append('created_date', params.filters.created_date);
				}
			if (params.filters.date_range) {
				queryParams.append('start_date', params.filters.date_range.start);
				queryParams.append('end_date', params.filters.date_range.end);
			}
			if (params.filters.has_subscription_history !== undefined) {
				queryParams.append('has_subscription_history', params.filters.has_subscription_history.toString());
			}
		}

		const url = `/admin/subscribers/enhanced?${queryParams}`;
			console.log('Making API request to ENHANCED endpoint:', url);
			console.log('Query parameters being sent:', queryParams.toString());
			
			const response = await api.get(url);
			
			console.log('API response received:', response);
			
			if (response.data) {
				console.log('Response data:', response.data);
				return response.data as SubscribersResponse;
			} else {
				console.error('No data in response:', response);
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
	 * Get subscribers by email verification status
	 */
	static async getSubscribersByEmailVerification(emailVerified: boolean, params?: {
		limit?: number;
		offset?: number;
		filters?: SubscriberFilters;
	}): Promise<SubscribersResponse> {
		try {
			console.log('StreamingSubscriberService.getSubscribersByEmailVerification called with:', { emailVerified, params });
			
			const queryParams = new URLSearchParams();
			
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());
			
			if (params?.filters) {
				if (params.filters.status) queryParams.append('status', params.filters.status);
				if (params.filters.search) queryParams.append('search', params.filters.search);
				if (params.filters.role) queryParams.append('role', params.filters.role);
				if (params.filters.last_login) queryParams.append('last_login', params.filters.last_login);
				if (params.filters.created_date) queryParams.append('created_date', params.filters.created_date);
				if (params.filters.date_range) {
					queryParams.append('start_date', params.filters.date_range.start);
					queryParams.append('end_date', params.filters.date_range.end);
				}
			}

			const endpoint = emailVerified ? '/admin/subscribers/verified' : '/admin/subscribers/unverified';
			const url = `${endpoint}?${queryParams}`;
			console.log('Making API request to:', url);
			
			const response = await api.get(url);
			
			console.log('API response received:', response);
			
			if (response.data) {
				const data = response.data as any; // Type assertion to fix TypeScript error
				// Transform the response to match expected format
				return {
					subscribers: data.subscribers || [],
					pagination: {
						limit: data.limit || params?.limit || 50,
						offset: data.offset || params?.offset || 0,
						total: data.total_count || 0
					}
				};
			} else {
				console.error('No data in response:', response);
				throw new Error(response.error || 'Failed to load subscribers');
			}
		} catch (error) {
			console.error('Error fetching subscribers by email verification:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber count by email verification status
	 */
	static async getSubscriberCountByEmailVerification(emailVerified: boolean, filters?: SubscriberFilters): Promise<number> {
		try {
			const queryParams = new URLSearchParams();
			
			if (filters?.status) queryParams.append('status', filters.status);
			if (filters?.search) queryParams.append('search', filters.search);
			if (filters?.role) queryParams.append('role', filters.role);
			if (filters?.last_login) queryParams.append('last_login', filters.last_login);
			if (filters?.created_date) queryParams.append('created_date', filters.created_date);
			if (filters?.date_range) {
				queryParams.append('start_date', filters.date_range.start);
				queryParams.append('end_date', filters.date_range.end);
			}

			const endpoint = emailVerified ? '/admin/subscribers/verified/count' : '/admin/subscribers/unverified/count';
			const response = await api.get(`${endpoint}?${queryParams}`);
			
			if (response.data) {
				const data = response.data as any; // Type assertion to fix TypeScript error
				return data.count || 0;
			} else {
				throw new Error(response.error || 'Failed to get subscriber count');
			}
		} catch (error) {
			console.error('Error fetching subscriber count by email verification:', error);
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
			if (filters?.has_subscription_history !== undefined) {
				queryParams.append('has_subscription_history', filters.has_subscription_history.toString());
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
			case 'inactive' :
				return { text: 'Inactive', class: 'bg-gray-100 text-gray-800', color: 'gray' };
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
	 * Calculate subscription start date (based on created_at)
	 */
	static calculateSubscriptionStartDate(subscriber: Subscriber): Date | null {
		if (!subscriber.created_at) return null;
		return new Date(subscriber.created_at);
	}

	/**
	 * Calculate subscription end date based on plan interval and start date
	 */
	static calculateSubscriptionEndDate(subscriber: Subscriber): Date | null {
		const startDate = this.calculateSubscriptionStartDate(subscriber);
		if (!startDate || !subscriber.plan_interval || !subscriber.plan_interval_count) {
			return null;
		}

		const endDate = new Date(startDate);
		const intervalCount = subscriber.plan_interval_count || 1;

		switch (subscriber.plan_interval.toLowerCase()) {
			case 'day':
				endDate.setDate(endDate.getDate() + intervalCount);
				break;
			case 'week':
				endDate.setDate(endDate.getDate() + (intervalCount * 7));
				break;
			case 'month':
				endDate.setMonth(endDate.getMonth() + intervalCount);
				break;
			case 'year':
				endDate.setFullYear(endDate.getFullYear() + intervalCount);
				break;
			default:
				// Default to monthly if interval is not recognized
				endDate.setMonth(endDate.getMonth() + intervalCount);
		}

		return endDate;
	}

	/**
	 * Format subscription dates for display
	 */
	static formatSubscriptionDates(subscriber: Subscriber): { startDate: string; endDate: string } {
		const startDate = this.calculateSubscriptionStartDate(subscriber);
		const endDate = this.calculateSubscriptionEndDate(subscriber);

		return {
			startDate: startDate ? this.formatDate(startDate.toISOString()) : 'N/A',
			endDate: endDate ? this.formatDate(endDate.toISOString()) : 'N/A'
		};
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

	/**
	 * Get all non-subscribers with optional filters and pagination
	 */
	static async getNonSubscribers(params?: {
		limit?: number;
		offset?: number;
		filters?: NonSubscriberFilters;
	}): Promise<NonSubscribersResponse> {
		try {
			console.log('StreamingSubscriberService.getNonSubscribers called with params:', params);
			
			const queryParams = new URLSearchParams();
			
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.offset) queryParams.append('offset', params.offset.toString());
			
			if (params?.filters) {
				if (params.filters.search) queryParams.append('search', params.filters.search);
				if (params.filters.email_verified !== undefined) queryParams.append('email_verified', params.filters.email_verified.toString());
				if (params.filters.role) queryParams.append('role', params.filters.role);
				if (params.filters.last_login) queryParams.append('last_login', params.filters.last_login);
				if (params.filters.created_date) queryParams.append('created_date', params.filters.created_date);
				if (params.filters.date_range) {
					queryParams.append('start_date', params.filters.date_range.start);
					queryParams.append('end_date', params.filters.date_range.end);
				}
				if (params.filters.has_subscription_history !== undefined) {
					queryParams.append('has_subscription_history', params.filters.has_subscription_history.toString());
				}
			}

			const url = `/admin/subscribers/non-subscribers?${queryParams}`;
			console.log('Making API request to:', url);

			const response = await api.get(url);
			
			console.log('API response received:', response);
			
			if (response.data) {
				console.log('Response data:', response.data);
				return response.data as NonSubscribersResponse;
			} else {
				console.error('No data in response:', response);
				throw new Error(response.error || 'Failed to load non-subscribers');
			}
		} catch (error) {
			console.error('Error fetching non-subscribers:', error);
			throw error;
		}
	}

	/**
	 * Get non-subscriber count with optional filters
	 */
	static async getNonSubscriberCount(filters?: NonSubscriberFilters): Promise<number> {
		try {
			const queryParams = new URLSearchParams();
			
			if (filters?.search) queryParams.append('search', filters.search);
			if (filters?.email_verified !== undefined) queryParams.append('email_verified', filters.email_verified.toString());
			if (filters?.role) queryParams.append('role', filters.role);
			if (filters?.last_login) queryParams.append('last_login', filters.last_login);
			if (filters?.created_date) queryParams.append('created_date', filters.created_date);
			if (filters?.date_range) {
				queryParams.append('start_date', filters.date_range.start);
				queryParams.append('end_date', filters.date_range.end);
			}
			if (filters?.has_subscription_history !== undefined) {
				queryParams.append('has_subscription_history', filters.has_subscription_history.toString());
			}

			const response = await api.get(`/admin/subscribers/non-subscribers/count?${queryParams}`);
			
			if (response.data) {
				return (response.data as { count: number }).count;
			} else {
				throw new Error(response.error || 'Failed to get non-subscriber count');
			}
		} catch (error) {
			console.error('Error fetching non-subscriber count:', error);
			throw error;
		}
	}

	/**
	 * Format non-subscriber name
	 */
	static formatNonSubscriberName(nonSubscriber: NonSubscriber): string {
		return `${nonSubscriber.first_name} ${nonSubscriber.last_name}`.trim() || nonSubscriber.email;
	}

	/**
	 * Update subscriber information
	 */
	static async updateSubscriber(id: number, updates: Partial<Subscriber>): Promise<Subscriber> {
		try {
			const response = await api.put(`/admin/subscribers/${id}`, updates);
			
			if (response.data) {
				return response.data as Subscriber;
			} else {
				throw new Error(response.error || 'Failed to update subscriber');
			}
		} catch (error) {
			console.error('Error updating subscriber:', error);
			throw error;
		}
	}

	/**
	 * Suspend a subscriber's account
	 */
	static async suspendSubscriber(id: number): Promise<Subscriber> {
		try {
			const response = await api.post(`/admin/subscribers/${id}/suspend`);
			
			if (response.data) {
				return (response.data as { subscriber: Subscriber }).subscriber;
			} else {
				throw new Error(response.error || 'Failed to suspend subscriber');
			}
		} catch (error) {
			console.error('Error suspending subscriber:', error);
			throw error;
		}
	}

	/**
	 * Activate a subscriber's account
	 */
	static async activateSubscriber(id: number): Promise<Subscriber> {
		try {
			const response = await api.post(`/admin/subscribers/${id}/activate`);
			
			if (response.data) {
				return (response.data as { subscriber: Subscriber }).subscriber;
			} else {
				throw new Error(response.error || 'Failed to activate subscriber');
			}
		} catch (error) {
			console.error('Error activating subscriber:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber history
	 */
	static async getSubscriberHistory(id: number): Promise<any[]> {
		try {
			const response = await api.get(`/admin/subscribers/${id}/history`);
			if (response.data) {
				return (response.data as { history: any[] }).history;
			} else {
				throw new Error(response.error || 'Failed to get subscriber history');
			}
		} catch (error) {
			console.error('Error getting subscriber history:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber history details from the new subscriber history service
	 */
	static async getSubscriberHistoryDetails(id: number): Promise<any> {
		try {
			const response = await api.get(`/admin/subscriber-history/${id}`);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to get subscriber history details');
			}
		} catch (error) {
			console.error('Error getting subscriber history details:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber history summary
	 */
	static async getSubscriberHistorySummary(id: number): Promise<any> {
		try {
			const response = await api.get(`/admin/subscriber-history/${id}/summary`);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to get subscriber history summary');
			}
		} catch (error) {
			console.error('Error getting subscriber history summary:', error);
			throw error;
		}
	}

	/**
	 * Add admin note to subscriber history
	 */
	static async addAdminNote(userId: number, note: { admin_id: number; admin_name: string; note: string; category: string }): Promise<any> {
		try {
			const response = await api.post(`/admin/subscriber-history/${userId}/notes`, note);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to add admin note');
			}
		} catch (error) {
			console.error('Error adding admin note:', error);
			throw error;
		}
	}

	/**
	 * Add system note to subscriber history
	 */
	static async addSystemNote(userId: number, note: { note: string; category: string }): Promise<any> {
		try {
			const response = await api.post(`/admin/subscriber-history/${userId}/system-notes`, note);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to add system note');
			}
		} catch (error) {
			console.error('Error adding system note:', error);
			throw error;
		}
	}

	/**
	 * Add user note to subscriber history
	 */
	static async addUserNote(userId: number, note: { note: string; category: string }): Promise<any> {
		try {
			const response = await api.post(`/admin/subscriber-history/${userId}/user-notes`, note);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to add user note');
			}
		} catch (error) {
			console.error('Error adding user note:', error);
			throw error;
		}
	}

	/**
	 * Export subscriber history
	 */
	static async exportSubscriberHistory(userId: number): Promise<Blob> {
		try {
			const response = await api.get(`/admin/subscriber-history/${userId}/export`);
			
			if (response.data) {
				// Convert the response data to a Blob
				const blob = new Blob([JSON.stringify(response.data)], { type: 'application/json' });
				return blob;
			} else {
				throw new Error(response.error || 'Failed to export subscriber history');
			}
		} catch (error) {
			console.error('Error exporting subscriber history:', error);
			throw error;
		}
	}

	/**
	 * Bulk suspend subscribers
	 */
	static async bulkSuspendSubscribers(userIDs: number[]): Promise<{ successful: Subscriber[]; failed: number[] }> {
		try {
			const response = await api.post('/admin/subscribers/bulk/suspend', userIDs);
			
			if (response.data) {
				const subscribers = (response.data as { subscribers: Subscriber[] }).subscribers || [];
				return {
					successful: subscribers,
					failed: []
				};
			} else {
				throw new Error(response.error || 'Failed to bulk suspend subscribers');
			}
		} catch (error) {
			console.error('Error bulk suspending subscribers:', error);
			throw error;
		}
	}

	/**
	 * Bulk activate subscribers
	 */
	static async bulkActivateSubscribers(userIDs: number[]): Promise<{ successful: Subscriber[]; failed: number[] }> {
		try {
			const response = await api.post('/admin/subscribers/bulk/activate', userIDs);
			
			if (response.data) {
				const subscribers = (response.data as { subscribers: Subscriber[] }).subscribers || [];
				return {
					successful: subscribers,
					failed: []
				};
			} else {
				throw new Error(response.error || 'Failed to bulk activate subscribers');
			}
		} catch (error) {
			console.error('Error bulk activating subscribers:', error);
			throw error;
		}
	}

	/**
	 * Bulk change plan for subscribers
	 */
	static async bulkChangePlan(planID: number, userIDs: number[]): Promise<{ successful: Subscriber[]; failed: number[] }> {
		try {
			const response = await api.post('/admin/subscribers/bulk/change-plan', { plan_id: planID, ids: userIDs });
			
			if (response.data) {
				const subscribers = (response.data as { subscribers: Subscriber[] }).subscribers || [];
				return {
					successful: subscribers,
					failed: []
				};
			} else {
				throw new Error(response.error || 'Failed to bulk change plan');
			}
		} catch (error) {
			console.error('Error bulk changing plan:', error);
			throw error;
		}
	}

	/**
	 * Export subscribers
	 */
	static async exportSubscribers(format: 'csv' | 'excel' = 'csv'): Promise<Blob> {
		try {
			// Get token from SecureTokenStorage
			const stored = localStorage.getItem('bome_auth_data');
			let token = '';
			if (stored) {
				const tokenData = JSON.parse(stored);
				token = tokenData.access_token || '';
			}
			
			const response = await apiRequest(`/admin/subscribers/export?format=${format}`, {
				method: 'GET',
			});
			
			if (!response.ok) {
				throw new Error(`Failed to export subscribers: ${response.status}`);
			}
			
			return await response.blob();
		} catch (error) {
			console.error('Error exporting subscribers:', error);
			throw error;
		}
	}

	/**
	 * Export non-subscribers
	 */
	static async exportNonSubscribers(format: 'csv' | 'excel' = 'csv'): Promise<Blob> {
		try {
			// Get token from SecureTokenStorage
			const stored = localStorage.getItem('bome_auth_data');
			let token = '';
			if (stored) {
				const tokenData = JSON.parse(stored);
				token = tokenData.access_token || '';
			}
			
			const response = await apiRequest(`/admin/subscribers/non-subscribers/export?format=${format}`, {
				method: 'GET'
			});
			
			if (!response.ok) {
				throw new Error(`Failed to export non-subscribers: ${response.status}`);
			}
			
			return await response.blob();
		} catch (error) {
			console.error('Error exporting non-subscribers:', error);
			throw error;
		}
	}
} 
