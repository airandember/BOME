import { api } from './auth';

export interface SubscriptionPlan {
	id: string;
	name: string;
	description: string;
	price: number;
	currency: string;
	interval: 'month' | 'year';
	features: string[];
	popular?: boolean;
}

export interface Subscription {
	id: string;
	planId: string;
	status: 'active' | 'canceled' | 'past_due' | 'unpaid';
	currentPeriodStart: string;
	currentPeriodEnd: string;
	cancelAtPeriodEnd: boolean;
	createdAt: string;
	updatedAt: string;
}

export interface PaymentMethod {
	id: string;
	type: 'card';
	card: {
		brand: string;
		last4: string;
		expMonth: number;
		expYear: number;
	};
}

export interface Invoice {
	id: string;
	amount: number;
	currency: string;
	status: 'paid' | 'open' | 'void' | 'uncollectible';
	createdAt: string;
	dueDate: string;
	periodStart: string;
	periodEnd: string;
	downloadUrl?: string;
}

// Subscription service
export const subscriptionService = {
	// Get available subscription plans
	getPlans: async () => {
		return api.get('/api/v1/subscriptions/plans');
	},

	// Get current user's subscription
	getCurrentSubscription: async () => {
		return api.get('/api/v1/subscriptions/current');
	},

	// Create a new subscription
	createSubscription: async (planId: string, paymentMethodId: string) => {
		return api.post('/api/v1/subscriptions', {
			planId,
			paymentMethodId
		});
	},

	// Cancel subscription
	cancelSubscription: async (subscriptionId: string, cancelAtPeriodEnd: boolean = true) => {
		return api.post(`/api/v1/subscriptions/${subscriptionId}/cancel`, {
			cancelAtPeriodEnd
		});
	},

	// Reactivate subscription
	reactivateSubscription: async (subscriptionId: string) => {
		return api.post(`/api/v1/subscriptions/${subscriptionId}/reactivate`);
	},

	// Update subscription (change plan)
	updateSubscription: async (subscriptionId: string, planId: string) => {
		return api.put(`/api/v1/subscriptions/${subscriptionId}`, {
			planId
		});
	},

	// Get payment methods
	getPaymentMethods: async () => {
		return api.get('/api/v1/payment-methods');
	},

	// Add payment method
	addPaymentMethod: async (paymentMethodData: any) => {
		return api.post('/api/v1/payment-methods', paymentMethodData);
	},

	// Remove payment method
	removePaymentMethod: async (paymentMethodId: string) => {
		return api.delete(`/api/v1/payment-methods/${paymentMethodId}`, {});
	},

	// Set default payment method
	setDefaultPaymentMethod: async (paymentMethodId: string) => {
		return api.post(`/api/v1/payment-methods/${paymentMethodId}/default`);
	},

	// Get billing history
	getBillingHistory: async (page = 1, limit = 20) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});
		
		return api.get(`/api/v1/billing/history?${params.toString()}`);
	},

	// Get specific invoice
	getInvoice: async (invoiceId: string) => {
		return api.get(`/api/v1/billing/invoices/${invoiceId}`);
	},

	// Download invoice
	downloadInvoice: async (invoiceId: string) => {
		return api.get(`/api/v1/billing/invoices/${invoiceId}/download`);
	},

	// Create checkout session for Stripe
	createCheckoutSession: async (planId: string, successUrl: string, cancelUrl: string) => {
		return api.post('/api/v1/subscriptions/checkout', {
			planId,
			successUrl,
			cancelUrl
		});
	},

	// Create customer portal session
	createCustomerPortalSession: async (returnUrl: string) => {
		return api.post('/api/v1/subscriptions/portal', {
			returnUrl
		});
	}
};

// Subscription utilities
export const subscriptionUtils = {
	// Format price for display
	formatPrice: (amount: number, currency: string = 'USD'): string => {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase()
		}).format(amount / 100); // Convert from cents
	},

	// Get plan price per month
	getMonthlyPrice: (plan: SubscriptionPlan): number => {
		if (plan.interval === 'month') {
			return plan.price;
		}
		return Math.round(plan.price / 12);
	},

	// Check if subscription is active
	isSubscriptionActive: (subscription: Subscription): boolean => {
		return subscription.status === 'active';
	},

	// Check if subscription will be canceled
	isSubscriptionCanceling: (subscription: Subscription): boolean => {
		return subscription.cancelAtPeriodEnd;
	},

	// Get subscription status display text
	getStatusText: (status: string): string => {
		switch (status) {
			case 'active':
				return 'Active';
			case 'canceled':
				return 'Canceled';
			case 'past_due':
				return 'Past Due';
			case 'unpaid':
				return 'Unpaid';
			default:
				return 'Unknown';
		}
	},

	// Get subscription status color
	getStatusColor: (status: string): string => {
		switch (status) {
			case 'active':
				return 'var(--success-bg)';
			case 'canceled':
				return 'var(--text-secondary)';
			case 'past_due':
				return 'var(--warning-bg)';
			case 'unpaid':
				return 'var(--error-bg)';
			default:
				return 'var(--text-secondary)';
		}
	},

	// Format date for display
	formatDate: (dateString: string): string => {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}
}; 