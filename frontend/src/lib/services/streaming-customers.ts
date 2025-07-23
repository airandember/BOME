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

			const response = await api.get(`/api/admin/streaming/customers?${queryParams}`);
			
			if (response.data) {
				return response.data;
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
	static async getCustomer(id: string): Promise<Customer> {
		try {
			const response = await api.get(`/api/admin/streaming/customers/${id}`);
			
			if (response.data) {
				return response.data.customer;
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
	static async updateCustomer(id: string, data: Partial<Customer>): Promise<Customer> {
		try {
			const response = await api.put(`/api/admin/streaming/customers/${id}`, data);
			
			if (response.data) {
				return response.data.customer;
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
	static async cancelSubscription(customerId: string, cancelAtPeriodEnd: boolean = true): Promise<void> {
		try {
			const response = await api.post(`/api/admin/streaming/customers/${customerId}/cancel`, {
				cancel_at_period_end: cancelAtPeriodEnd
			});
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to cancel subscription');
			}
		} catch (error) {
			console.error('Error cancelling subscription:', error);
			throw error;
		}
	}

	/**
	 * Reactivate customer subscription
	 */
	static async reactivateSubscription(customerId: string): Promise<void> {
		try {
			const response = await api.post(`/api/admin/streaming/customers/${customerId}/reactivate`);
			
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
	static async processRefund(customerId: string, refundData: RefundData): Promise<void> {
		try {
			const response = await api.post(`/api/admin/streaming/customers/${customerId}/refund`, refundData);
			
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
	static async sendCommunication(customerId: string, communicationData: CommunicationData): Promise<void> {
		try {
			const response = await api.post(`/api/admin/streaming/customers/${customerId}/communication`, communicationData);
			
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
	static async getCustomerSubscriptionHistory(customerId: string): Promise<{
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
			const response = await api.get(`/api/admin/streaming/customers/${customerId}/subscription-history`);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to load subscription history');
			}
		} catch (error) {
			console.error('Error fetching subscription history:', error);
			throw error;
		}
	}

	/**
	 * Get customer payment history
	 */
	static async getCustomerPaymentHistory(customerId: string): Promise<{
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
			const response = await api.get(`/api/admin/streaming/customers/${customerId}/payment-history`);
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to load payment history');
			}
		} catch (error) {
			console.error('Error fetching payment history:', error);
			throw error;
		}
	}

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
			const response = await api.get('/api/admin/streaming/customers/stats');
			
			if (response.data) {
				return response.data;
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
	static async exportCustomers(filters?: CustomerFilters): Promise<Blob> {
		try {
			const queryParams = new URLSearchParams();
			
			if (filters?.status) queryParams.append('status', filters.status);
			if (filters?.plan) queryParams.append('plan', filters.plan);
			if (filters?.date_from) queryParams.append('date_from', filters.date_from);
			if (filters?.date_to) queryParams.append('date_to', filters.date_to);

			const response = await api.get(`/api/admin/streaming/customers/export?${queryParams}`, {
				responseType: 'blob'
			});
			
			if (response.data) {
				return response.data;
			} else {
				throw new Error(response.error || 'Failed to export customers');
			}
		} catch (error) {
			console.error('Error exporting customers:', error);
			throw error;
		}
	}
} 