import { api } from '$lib/api';

export interface Customer {
	id: string;
	email: string;
	name: string;
	status: 'active' | 'cancelled' | 'past_due' | 'unpaid' | 'trialing';
	subscription?: {
		id: string;
		plan_name: string;
		price: number;
		currency: string;
		interval: string;
		status: string;
		current_period_start: string;
		current_period_end: string;
		cancel_at_period_end: boolean;
	};
	created_at: string;
	updated_at: string;
	last_login?: string;
	total_spent: number;
	subscription_count: number;
}

export interface CustomerFilters {
	page?: number;
	limit?: number;
	search?: string;
	status?: 'active' | 'cancelled' | 'past_due' | 'unpaid' | 'trialing' | 'all';
	plan?: string;
	date_from?: string;
	date_to?: string;
}

export interface CustomersResponse {
	customers: Customer[];
	total: number;
	page: number;
	limit: number;
}

export interface RefundData {
	amount: number;
	reason: string;
	notes?: string;
}

export interface CommunicationData {
	subject: string;
	message: string;
	type: 'email' | 'sms' | 'in_app';
}

export class StreamingCustomerService {
	/**
	 * Get all customers with optional filters and pagination
	 */
	static async getCustomers(filters?: CustomerFilters): Promise<CustomersResponse> {
		try {
			const queryParams = new URLSearchParams();
			
			if (filters?.page) queryParams.append('page', filters.page.toString());
			if (filters?.limit) queryParams.append('limit', filters.limit.toString());
			if (filters?.search) queryParams.append('search', filters.search);
			if (filters?.status) queryParams.append('status', filters.status);
			if (filters?.plan) queryParams.append('plan', filters.plan);
			if (filters?.date_from) queryParams.append('date_from', filters.date_from);
			if (filters?.date_to) queryParams.append('date_to', filters.date_to);

			const response = await api.get(`/admin/streaming/customers?${queryParams}`);
			
			if (response.data) {
				return response.data as CustomersResponse;
			} else {
				throw new Error(response.error || 'Failed to load customers');
			}
		} catch (error) {
			console.error('Error fetching customers:', error);
			throw error;
		}
	}

	/**
	 * Get a single customer by ID
	 */
	static async getCustomer(id: number): Promise<Customer> {
		try {
			const response = await api.get(`/admin/streaming/customers/${id}`);
			
			if (response.data) {
				return (response.data as { customer: Customer }).customer;
			} else {
				throw new Error(response.error || 'Failed to load customer');
			}
		} catch (error) {
			console.error('Error fetching customer:', error);
			throw error;
		}
	}

	/**
	 * Update customer information
	 */
	static async updateCustomer(id: number, data: Partial<Customer>): Promise<Customer> {
		try {
			const response = await api.put(`/admin/streaming/customers/${id}`, data);
			
			if (response.data) {
				return (response.data as { customer: Customer }).customer;
			} else {
				throw new Error(response.error || 'Failed to update customer');
			}
		} catch (error) {
			console.error('Error updating customer:', error);
			throw error;
		}
	}

	/**
	 * Cancel customer subscription
	 */
	static async cancelCustomerSubscription(customerId: number, reason?: string): Promise<void> {
		try {
			const response = await api.post(`/admin/streaming/customers/${customerId}/cancel`, {
				reason: reason || 'Admin cancellation'
			});
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to cancel subscription');
			}
		} catch (error) {
			console.error('Error canceling subscription:', error);
			throw error;
		}
	}

	/**
	 * Reactivate customer subscription
	 */
	static async reactivateCustomerSubscription(customerId: number): Promise<void> {
		try {
			const response = await api.post(`/admin/streaming/customers/${customerId}/reactivate`);
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to reactivate subscription');
			}
		} catch (error) {
			console.error('Error reactivating subscription:', error);
			throw error;
		}
	}

	/**
	 * Process refund for customer
	 */
	static async processRefund(customerId: number, refundData: RefundData): Promise<void> {
		try {
			const response = await api.post(`/admin/streaming/customers/${customerId}/refund`, refundData);
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to process refund');
			}
		} catch (error) {
			console.error('Error processing refund:', error);
			throw error;
		}
	}

	/**
	 * Send communication to customer
	 */
	static async sendCommunication(customerId: number, communicationData: CommunicationData): Promise<void> {
		try {
			const response = await api.post(`/admin/streaming/customers/${customerId}/communication`, communicationData);
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to send communication');
			}
		} catch (error) {
			console.error('Error sending communication:', error);
			throw error;
		}
	}

	/**
	 * Get customer subscription history
	 */
	static async getCustomerSubscriptionHistory(customerId: number): Promise<{
		subscriptions: Array<{
			id: string;
			plan_name: string;
			status: string;
			start_date: string;
			end_date?: string;
			amount: number;
			currency: string;
		}>;
	}> {
		try {
			const response = await api.get(`/admin/streaming/customers/${customerId}/subscription-history`);
			
			if (response.data) {
				return {
					subscriptions: (response.data as { history: Array<{
						id: string;
						plan_name: string;
						status: string;
						start_date: string;
					end_date?: string;
					amount: number;
					currency: string;
				}> }).history,
				};
			} else {
				throw new Error(response.error || 'Failed to load subscription history');
			}
		} catch (error) {
			console.error('Error fetching subscription history:', error);
			throw error;
		}
	}

	/**
	 * Get customer payment history - This we'll have to coordinate with Stripe API capabilities
	 
	static async getCustomerPaymentHistory(customerId: number): Promise<{
		payments: Array<{
			id: string;
			amount: number;
			currency: string;
			status: string;
			date: string;
			description: string;
		}>;
	}> {
		try {
			const response = await api.get(`/admin/streaming/customers/${customerId}/payment-history`);
			
			if (response.data) {
				return {
					payments: (response.data as Array<{
						id: string;
						amount: number;
						currency: string;
						status: string;
					date: string;
						description: string;
					}>).payments,
				};
			} else {
				throw new Error(response.error || 'Failed to load payment history');
			}
		} catch (error) {
			console.error('Error fetching payment history:', error);
			throw error;
		}
	}*/

	/**
	 * Get customer statistics
	 */
	static async getCustomerStats(): Promise<{
		total_customers: number;
		active_customers: number;
		cancelled_customers: number;
		trial_customers: number;
		revenue_this_month: number;
		avg_customer_lifetime_value: number;
	}> {
		try {
			const response = await api.get('/admin/streaming/customers/stats');
			
			if (response.data) {
				return response.data as {
					total_customers: number;
					active_customers: number;
					cancelled_customers: number;
					trial_customers: number;
					revenue_this_month: number;
					avg_customer_lifetime_value: number;
				};
			} else {
				throw new Error(response.error || 'Failed to load customer stats');
			}
		} catch (error) {
			console.error('Error fetching customer stats:', error);
			throw error;
		}
	}

	/**
	 * Export customers data
	 */
	static async exportCustomers(params: {
		format: 'csv' | 'json' | 'excel';
		filters?: CustomerFilters;
	}): Promise<Blob> {
		try {
			const queryParams = new URLSearchParams();
			queryParams.append('format', params.format);
			
			if (params.filters) {
				if (params.filters.status) queryParams.append('status', params.filters.status);
				if (params.filters.search) queryParams.append('search', params.filters.search);
				if (params.filters?.date_from) queryParams.append('date_from', params.filters.date_from);
				if (params.filters?.date_to) queryParams.append('date_to', params.filters.date_to);
			}

			const response = await api.get(`/admin/streaming/customers/export?${queryParams}`);
			
			if (response.data) {
				return response.data as Blob;
			} else {
				throw new Error('Failed to export customers');
			}
		} catch (error) {
			console.error('Error exporting customers:', error);
			throw error;
		}
	}
} 
