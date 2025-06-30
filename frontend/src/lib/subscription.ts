import { apiRequest } from './auth';

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

export interface Refund {
	id: string;
	amount: number;
	currency: string;
	status: 'succeeded' | 'pending' | 'failed' | 'canceled';
	reason: 'duplicate' | 'fraudulent' | 'requested_by_customer';
	paymentIntentId: string;
	chargeId?: string;
	createdAt: string;
	receiptNumber?: string;
	failureReason?: string;
}

// Subscription service
export const subscriptionService = {
	// Get available subscription plans
	getPlans: async () => {
		const response = await apiRequest('/subscriptions/plans');
		return response.json();
	},

	// Get current user's subscription
	getCurrentSubscription: async () => {
		const response = await apiRequest('/subscriptions/current');
		return response.json();
	},

	// Create a new subscription
	createSubscription: async (planId: string, paymentMethodId: string) => {
		const response = await apiRequest('/subscriptions', {
			method: 'POST',
			body: JSON.stringify({
				planId,
				paymentMethodId
			})
		});
		return response.json();
	},

	// Cancel subscription
	cancelSubscription: async (subscriptionId: string, cancelAtPeriodEnd: boolean = true) => {
		const response = await apiRequest(`/subscriptions/${subscriptionId}/cancel`, {
			method: 'POST',
			body: JSON.stringify({
				cancelAtPeriodEnd
			})
		});
		return response.json();
	},

	// Reactivate subscription
	reactivateSubscription: async (subscriptionId: string) => {
		const response = await apiRequest(`/subscriptions/${subscriptionId}/reactivate`, {
			method: 'POST',
			body: JSON.stringify({})
		});
		return response.json();
	},

	// Update subscription (change plan)
	updateSubscription: async (subscriptionId: string, planId: string) => {
		const response = await apiRequest(`/subscriptions/${subscriptionId}`, {
			method: 'PUT',
			body: JSON.stringify({
				planId
			})
		});
		return response.json();
	},

	// Get payment methods
	getPaymentMethods: async () => {
		const response = await apiRequest('/payment-methods');
		return response.json();
	},

	// Add payment method
	addPaymentMethod: async (paymentMethodData: any) => {
		const response = await apiRequest('/payment-methods', {
			method: 'POST',
			body: JSON.stringify(paymentMethodData)
		});
		return response.json();
	},

	// Remove payment method
	removePaymentMethod: async (paymentMethodId: string) => {
		const response = await apiRequest(`/payment-methods/${paymentMethodId}`, {
			method: 'DELETE'
		});
		return response.json();
	},

	// Set default payment method
	setDefaultPaymentMethod: async (paymentMethodId: string) => {
		const response = await apiRequest(`/payment-methods/${paymentMethodId}/default`, {
			method: 'POST',
			body: JSON.stringify({})
		});
		return response.json();
	},

	// Get billing history
	getBillingHistory: async (page = 1, limit = 20) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});
		
		const response = await apiRequest(`/billing/history?${params.toString()}`);
		return response.json();
	},

	// Get specific invoice
	getInvoice: async (invoiceId: string) => {
		const response = await apiRequest(`/billing/invoices/${invoiceId}`);
		return response.json();
	},

	// Download invoice
	downloadInvoice: async (invoiceId: string) => {
		const response = await apiRequest(`/billing/invoices/${invoiceId}/download`);
		return response.json();
	},

	// Create checkout session for Stripe
	createCheckoutSession: async (planId: string, successUrl: string, cancelUrl: string) => {
		const response = await apiRequest('/subscriptions/checkout', {
			method: 'POST',
			body: JSON.stringify({
				planId,
				successUrl,
				cancelUrl
			})
		});
		return response.json();
	},

	// Create customer portal session
	createCustomerPortalSession: async (returnUrl: string) => {
		const response = await apiRequest('/subscriptions/portal', {
			method: 'POST',
			body: JSON.stringify({
				returnUrl
			})
		});
		return response.json();
	},

	// Get customer refunds
	getRefunds: async (limit = 20) => {
		const params = new URLSearchParams({
			limit: limit.toString()
		});
		
		const response = await apiRequest(`/billing/refunds?${params.toString()}`);
		return response.json();
	},

	// Get specific refund
	getRefund: async (refundId: string) => {
		const response = await apiRequest(`/billing/refunds/${refundId}`);
		return response.json();
	},

	// Create refund
	createRefund: async (paymentIntentId: string, amount?: number, reason?: string) => {
		const response = await apiRequest('/billing/refunds', {
			method: 'POST',
			body: JSON.stringify({
				payment_intent_id: paymentIntentId,
				amount: amount,
				reason: reason || 'requested_by_customer'
			})
		});
		return response.json();
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