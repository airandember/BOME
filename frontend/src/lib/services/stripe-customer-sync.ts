import { apiRequest } from '$lib/auth';

export interface CustomerSyncResult {
	customer_id: number;
	stripe_id: string;
	action: string;
	message: string;
	error?: string;
	last_sync_at: string;
}

export interface CustomerSyncStats {
	total_processed: number;
	created: number;
	updated: number;
	synced: number;
	errors: number;
	duration: string;
}

export class StripeCustomerSyncService {
	/**
	 * Sync a local customer to Stripe
	 */
	static async syncCustomerToStripe(customerId: number): Promise<CustomerSyncResult> {
		const response = await apiRequest(`/admin/streaming/stripe/customers/${customerId}/sync-to-stripe`, {
			method: 'POST'
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to sync customer to Stripe');
		}

		const data = await response.json();
		return data.result;
	}

	/**
	 * Sync a Stripe customer to local database
	 */
	static async syncCustomerFromStripe(stripeCustomerId: string): Promise<CustomerSyncResult> {
		const response = await apiRequest('/admin/streaming/stripe/customers/sync-from-stripe', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				stripe_customer_id: stripeCustomerId
			})
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to sync customer from Stripe');
		}

		const data = await response.json();
		return data.result;
	}

	/**
	 * Bulk sync multiple customers to Stripe
	 */
	static async bulkSyncCustomersToStripe(customerIds: number[]): Promise<CustomerSyncStats> {
		const response = await apiRequest('/admin/streaming/stripe/customers/bulk-sync-to-stripe', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				customer_ids: customerIds
			})
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to bulk sync customers to Stripe');
		}

		const data = await response.json();
		return data.stats;
	}

	/**
	 * Bulk sync multiple customers from Stripe
	 */
	static async bulkSyncCustomersFromStripe(stripeCustomerIds: string[]): Promise<CustomerSyncStats> {
		const response = await apiRequest('/admin/streaming/stripe/customers/bulk-sync-from-stripe', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({
				stripe_customer_ids: stripeCustomerIds
			})
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to bulk sync customers from Stripe');
		}

		const data = await response.json();
		return data.stats;
	}

	/**
	 * Sync all customers in both directions
	 */
	static async syncAllCustomers(): Promise<CustomerSyncStats> {
		const response = await apiRequest('/admin/streaming/stripe/customers/sync-all', {
			method: 'POST'
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to sync all customers');
		}

		const data = await response.json();
		return data.stats;
	}

	/**
	 * Get sync status for a customer
	 */
	static async getSyncStatus(customerId: number): Promise<CustomerSyncResult> {
		const response = await apiRequest(`/admin/streaming/stripe/customers/${customerId}/sync-status`);

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to get sync status');
		}

		const data = await response.json();
		return data.result;
	}

	/**
	 * Get a human-readable action description
	 */
	static getActionDescription(action: string): string {
		switch (action) {
			case 'created':
				return 'Created';
			case 'updated':
				return 'Updated';
			case 'synced':
				return 'Synced';
			case 'not_synced':
				return 'Not Synced';
			case 'stripe_not_found':
				return 'Stripe Customer Not Found';
			case 'error':
				return 'Error';
			default:
				return action;
		}
	}

	/**
	 * Get action color for UI
	 */
	static getActionColor(action: string): string {
		switch (action) {
			case 'created':
			case 'updated':
			case 'synced':
				return 'var(--success)';
			case 'not_synced':
				return 'var(--warning)';
			case 'stripe_not_found':
			case 'error':
				return 'var(--error)';
			default:
				return 'var(--text-muted)';
		}
	}
} 
