/**
 * User Subscription Service
 * Phase 7.2 - User-controlled subscription management
 */

import { apiClient } from '$lib/api/client';

export interface UserSubscription {
	id: string;
	stripe_customer_id: string;
	plan_name: string;
	status: string;
	price: number; // in cents
	currency: string;
	interval: string; // 'month' | 'year'
	current_period_start: string;
	current_period_end: string;
	days_until_renewal: number;
	cancel_at_period_end: boolean;
	canceled_at?: string;
	created_at: string;
	is_lifetime: boolean;
	is_primary: boolean;
}

export interface SubscriptionCounts {
	active: number;
	trialing: number;
	canceled: number;
	past_due: number;
	unpaid: number;
	total: number;
}

export interface UserSubscriptionsResponse {
	active_subscriptions: UserSubscription[];
	canceled_subscriptions: UserSubscription[];
	subscription_count: SubscriptionCounts;
	has_multiple_active: boolean;
	video_access: boolean;
}

export interface CancelMultipleRequest {
	subscription_ids: string[];
	keep_subscription_id?: string;
	reason?: string;
}

export interface CanceledSubscriptionInfo {
	id: string;
	status: string;
	cancel_at_period_end: boolean;
	canceled_at: string;
	ends_on: string;
}

export interface CancelMultipleResponse {
	success: boolean;
	message: string;
	canceled_subscriptions: CanceledSubscriptionInfo[];
	kept_subscription?: UserSubscription;
}

export interface CancelSingleResponse {
	success: boolean;
	message: string;
	subscription: CanceledSubscriptionInfo;
}

export class UserSubscriptionService {
	/**
	 * Get all subscriptions for the current user
	 */
	static async getSubscriptions(): Promise<UserSubscriptionsResponse> {
		const response = await apiClient.get<{ success: boolean; subscriptions: UserSubscriptionsResponse }>(
			'/user/subscriptions'
		);

		if (response.error) {
			throw new Error(response.error);
		}

		return response.data!.subscriptions;
	}

	/**
	 * Get subscription history for the current user
	 */
	static async getSubscriptionHistory(): Promise<UserSubscription[]> {
		const response = await apiClient.get<{ success: boolean; history: UserSubscription[] }>(
			'/user/subscriptions/history'
		);

		if (response.error) {
			throw new Error(response.error);
		}

		return response.data!.history;
	}

	/**
	 * Cancel multiple subscriptions (bulk)
	 */
	static async cancelMultipleSubscriptions(
		request: CancelMultipleRequest
	): Promise<CancelMultipleResponse> {
		const response = await apiClient.post<CancelMultipleResponse>(
			'/user/subscriptions/cancel-multiple',
			request
		);

		if (response.error) {
			throw new Error(response.error);
		}

		return response.data!;
	}

	/**
	 * Cancel a single subscription
	 */
	static async cancelSingleSubscription(
		subscriptionId: string,
		reason?: string
	): Promise<CancelSingleResponse> {
		const response = await apiClient.post<CancelSingleResponse>(
			`/user/subscriptions/${subscriptionId}/cancel`,
			{ reason }
		);

		if (response.error) {
			throw new Error(response.error);
		}

		return response.data!;
	}

	/**
	 * Format price from cents to dollars
	 */
	static formatPrice(cents: number, currency: string = 'USD'): string {
		const dollars = cents / 100;
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase()
		}).format(dollars);
	}

	/**
	 * Format interval display
	 */
	static formatInterval(interval: string): string {
		const intervalMap: Record<string, string> = {
			month: 'Monthly',
			year: 'Yearly',
			week: 'Weekly',
			day: 'Daily'
		};
		return intervalMap[interval] || interval;
	}

	/**
	 * Format date
	 */
	static formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	/**
	 * Get status badge color
	 */
	static getStatusColor(status: string): string {
		const statusColors: Record<string, string> = {
			active: 'green',
			trialing: 'blue',
			canceled: 'gray',
			past_due: 'orange',
			unpaid: 'red',
			incomplete: 'yellow'
		};
		return statusColors[status] || 'gray';
	}

	/**
	 * Get days left color (for styling)
	 */
	static getDaysLeftColor(days: number): string {
		if (days < 0) return 'red'; // Expired
		if (days < 7) return 'orange'; // Expiring soon
		if (days < 30) return 'yellow'; // Within a month
		return 'green'; // Good standing
	}
}

