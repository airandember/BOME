import { apiRequest } from '$lib/auth';

// Stripe Financial Types
export interface StripeCustomer {
	id: string;
	email: string;
	name: string;
	created: number;
	subscription?: StripeSubscription;
	payment_methods: StripePaymentMethod[];
	metadata: Record<string, string>;
}

export interface StripeSubscription {
	id: string;
	customer_id: string;
	status: 'active' | 'canceled' | 'incomplete' | 'incomplete_expired' | 'past_due' | 'trialing' | 'unpaid';
	plan_id: string;
	plan_name: string;
	plan_price: number;
	plan_interval: 'month' | 'year';
	current_period_start: number;
	current_period_end: number;
	canceled_at?: number;
	ended_at?: number;
	trial_start?: number;
	trial_end?: number;
	quantity: number;
	metadata: Record<string, string>;
}

export interface StripePaymentMethod {
	id: string;
	type: 'card' | 'bank_account' | 'sepa_debit';
	card?: {
		brand: string;
		last4: string;
		exp_month: number;
		exp_year: number;
	};
	bank_account?: {
		bank_name: string;
		last4: string;
		routing_number: string;
	};
}

export interface StripeInvoice {
	id: string;
	customer_id: string;
	subscription_id?: string;
	amount_paid: number;
	amount_due: number;
	currency: string;
	status: 'draft' | 'open' | 'paid' | 'uncollectible' | 'void';
	created: number;
	due_date?: number;
	paid_at?: number;
	lines: StripeInvoiceLine[];
}

export interface StripeInvoiceLine {
	id: string;
	description: string;
	amount: number;
	quantity: number;
	unit_amount: number;
	period_start?: number;
	period_end?: number;
}

export interface StripePayment {
	id: string;
	customer_id: string;
	amount: number;
	currency: string;
	status: 'succeeded' | 'pending' | 'failed' | 'canceled';
	created: number;
	payment_method: string;
	payment_intent?: string;
	refunded: boolean;
	refund_amount?: number;
	metadata: Record<string, string>;
}

export interface FinancialMetrics {
	// Revenue metrics
	total_revenue: number;
	monthly_recurring_revenue: number;
	annual_recurring_revenue: number;
	average_revenue_per_user: number;
	revenue_growth_rate: number;
	
	// Subscription metrics
	total_subscriptions: number;
	active_subscriptions: number;
	canceled_subscriptions: number;
	subscription_growth_rate: number;
	churn_rate: number;
	
	// Customer metrics
	total_customers: number;
	new_customers_this_month: number;
	customer_growth_rate: number;
	customer_lifetime_value: number;
	
	// Payment metrics
	payment_success_rate: number;
	average_payment_amount: number;
	failed_payments: number;
	refund_rate: number;
	
	// Plan distribution
	plan_distribution: Record<string, number>;
	revenue_by_plan: Record<string, number>;
}

export interface FinancialReport {
	period: '7d' | '30d' | '90d' | '1y' | 'all';
	start_date: string;
	end_date: string;
	metrics: FinancialMetrics;
	revenue_timeline: Array<{
		date: string;
		revenue: number;
		subscriptions: number;
		customers: number;
	}>;
	top_customers: Array<{
		customer: StripeCustomer;
		total_spent: number;
		subscription_count: number;
	}>;
	recent_transactions: StripePayment[];
}

export interface FinancialProjection {
	projection_period: '30d' | '90d' | '6m' | '1y';
	projected_revenue: number;
	projected_subscriptions: number;
	projected_customers: number;
	confidence_interval: {
		low: number;
		high: number;
	};
	assumptions: {
		growth_rate: number;
		churn_rate: number;
		price_changes: Record<string, number>;
	};
}

export interface StripeWebhookEvent {
	id: string;
	type: string;
	created: number;
	data: {
		object: any;
	};
	livemode: boolean;
}

class StripeFinancialService {
	private static instance: StripeFinancialService;
	private cache = new Map<string, { data: any; timestamp: number }>();
	private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

	private constructor() {}

	static getInstance(): StripeFinancialService {
		if (!StripeFinancialService.instance) {
			StripeFinancialService.instance = new StripeFinancialService();
		}
		return StripeFinancialService.instance;
	}

	/**
	 * Get comprehensive financial dashboard data
	 */
	async getFinancialDashboard(): Promise<{
		success: boolean;
		data: {
			metrics: FinancialMetrics;
			recent_transactions: StripePayment[];
			top_customers: Array<{
				customer: StripeCustomer;
				total_spent: number;
			}>;
			revenue_chart: Array<{
				date: string;
				revenue: number;
				subscriptions: number;
			}>;
		};
	}> {
		const cacheKey = 'financial_dashboard';
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const response = await apiRequest('/admin/stripe/dashboard');
			
			if (!response.ok) {
				throw new Error('Failed to fetch financial dashboard');
			}

			const data = await response.json();
			
			this.cache.set(cacheKey, {
				data: data,
				timestamp: Date.now()
			});

			return data;
		} catch (error) {
			console.error('Failed to fetch financial dashboard:', error);
			return {
				success: false,
				data: {
					metrics: {
						total_revenue: 0,
						monthly_recurring_revenue: 0,
						annual_recurring_revenue: 0,
						average_revenue_per_user: 0,
						revenue_growth_rate: 0,
						total_subscriptions: 0,
						active_subscriptions: 0,
						canceled_subscriptions: 0,
						subscription_growth_rate: 0,
						churn_rate: 0,
						total_customers: 0,
						new_customers_this_month: 0,
						customer_growth_rate: 0,
						customer_lifetime_value: 0,
						payment_success_rate: 0,
						average_payment_amount: 0,
						failed_payments: 0,
						refund_rate: 0,
						plan_distribution: {},
						revenue_by_plan: {}
					},
					recent_transactions: [],
					top_customers: [],
					revenue_chart: []
				}
			};
		}
	}

	/**
	 * Get detailed financial report for a specific period
	 */
	async getFinancialReport(period: '7d' | '30d' | '90d' | '1y' | 'all'): Promise<{
		success: boolean;
		data: FinancialReport;
	}> {
		const cacheKey = `financial_report_${period}`;
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const response = await apiRequest(`/admin/stripe/report?period=${period}`);
			
			if (!response.ok) {
				throw new Error('Failed to fetch financial report');
			}

			const data = await response.json();
			
			this.cache.set(cacheKey, {
				data: data,
				timestamp: Date.now()
			});

			return data;
		} catch (error) {
			console.error('Failed to fetch financial report:', error);
			return {
				success: false,
				data: {
					period,
					start_date: new Date().toISOString(),
					end_date: new Date().toISOString(),
					metrics: {
						total_revenue: 0,
						monthly_recurring_revenue: 0,
						annual_recurring_revenue: 0,
						average_revenue_per_user: 0,
						revenue_growth_rate: 0,
						total_subscriptions: 0,
						active_subscriptions: 0,
						canceled_subscriptions: 0,
						subscription_growth_rate: 0,
						churn_rate: 0,
						total_customers: 0,
						new_customers_this_month: 0,
						customer_growth_rate: 0,
						customer_lifetime_value: 0,
						payment_success_rate: 0,
						average_payment_amount: 0,
						failed_payments: 0,
						refund_rate: 0,
						plan_distribution: {},
						revenue_by_plan: {}
					},
					revenue_timeline: [],
					top_customers: [],
					recent_transactions: []
				}
			};
		}
	}

	/**
	 * Get financial projections
	 */
	async getFinancialProjections(period: '30d' | '90d' | '6m' | '1y'): Promise<{
		success: boolean;
		data: FinancialProjection;
	}> {
		const cacheKey = `financial_projection_${period}`;
		const cached = this.cache.get(cacheKey);
		
		if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
			return cached.data;
		}

		try {
			const response = await apiRequest(`/admin/stripe/projections?period=${period}`);
			
			if (!response.ok) {
				throw new Error('Failed to fetch financial projections');
			}

			const data = await response.json();
			
			this.cache.set(cacheKey, {
				data: data,
				timestamp: Date.now()
			});

			return data;
		} catch (error) {
			console.error('Failed to fetch financial projections:', error);
			return {
				success: false,
				data: {
					projection_period: period,
					projected_revenue: 0,
					projected_subscriptions: 0,
					projected_customers: 0,
					confidence_interval: {
						low: 0,
						high: 0
					},
					assumptions: {
						growth_rate: 0,
						churn_rate: 0,
						price_changes: {}
					}
				}
			};
		}
	}

	/**
	 * Get all customers with their subscription details
	 */
	async getCustomers(params: {
		page?: number;
		limit?: number;
		search?: string;
		status?: string;
		sort_field?: string;
		sort_direction?: 'asc' | 'desc';
	}): Promise<{
		success: boolean;
		customers: StripeCustomer[];
		pagination: {
			current_page: number;
			per_page: number;
			total: number;
			total_pages: number;
		};
	}> {
		const searchParams = new URLSearchParams();
		if (params.page) searchParams.append('page', params.page.toString());
		if (params.limit) searchParams.append('limit', params.limit.toString());
		if (params.search) searchParams.append('search', params.search);
		if (params.status) searchParams.append('status', params.status);
		if (params.sort_field) searchParams.append('sort_field', params.sort_field);
		if (params.sort_direction) searchParams.append('sort_direction', params.sort_direction);

		try {
			const response = await apiRequest(`/admin/stripe/customers?${searchParams.toString()}`);
			
			if (!response.ok) {
				throw new Error('Failed to fetch customers');
			}

			return await response.json();
		} catch (error) {
			console.error('Failed to fetch customers:', error);
			return {
				success: false,
				customers: [],
				pagination: {
					current_page: 1,
					per_page: 20,
					total: 0,
					total_pages: 0
				}
			};
		}
	}

	/**
	 * Get customer details with full payment history
	 */
	async getCustomerDetails(customerId: string): Promise<{
		success: boolean;
		customer: StripeCustomer;
		subscriptions: StripeSubscription[];
		payments: StripePayment[];
		invoices: StripeInvoice[];
	}> {
		try {
			const response = await apiRequest(`/admin/stripe/customers/${customerId}`);
			
			if (!response.ok) {
				throw new Error('Failed to fetch customer details');
			}

			return await response.json();
		} catch (error) {
			console.error('Failed to fetch customer details:', error);
			return {
				success: false,
				customer: {} as StripeCustomer,
				subscriptions: [],
				payments: [],
				invoices: []
			};
		}
	}

	/**
	 * Get recent transactions
	 */
	async getRecentTransactions(limit: number = 50): Promise<{
		success: boolean;
		transactions: StripePayment[];
	}> {
		try {
			const response = await apiRequest(`/admin/stripe/transactions?limit=${limit}`);
			
			if (!response.ok) {
				throw new Error('Failed to fetch recent transactions');
			}

			return await response.json();
		} catch (error) {
			console.error('Failed to fetch recent transactions:', error);
			return {
				success: false,
				transactions: []
			};
		}
	}

	/**
	 * Export financial data
	 */
	async exportFinancialData(format: 'csv' | 'json' | 'pdf', period: string): Promise<{
		success: boolean;
		download_url: string;
	}> {
		try {
			const response = await apiRequest('/admin/stripe/export', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					format,
					period
				})
			});
			
			if (!response.ok) {
				throw new Error('Failed to export financial data');
			}

			return await response.json();
		} catch (error) {
			console.error('Failed to export financial data:', error);
			return {
				success: false,
				download_url: ''
			};
		}
	}

	/**
	 * Clear cache
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
export const stripeFinancialService = StripeFinancialService.getInstance(); 