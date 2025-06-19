// Stripe Payment Processing Integration
export interface StripeConfig {
	publishableKey: string;
	secretKey: string;
	webhookSecret: string;
}

export interface PaymentIntent {
	id: string;
	amount: number;
	currency: string;
	status: 'requires_payment_method' | 'requires_confirmation' | 'requires_action' | 'processing' | 'requires_capture' | 'canceled' | 'succeeded';
	clientSecret: string;
	metadata?: Record<string, string>;
}

export interface Subscription {
	id: string;
	customerId: string;
	status: 'incomplete' | 'incomplete_expired' | 'trialing' | 'active' | 'past_due' | 'canceled' | 'unpaid';
	currentPeriodStart: number;
	currentPeriodEnd: number;
	cancelAtPeriodEnd: boolean;
	items: SubscriptionItem[];
	metadata?: Record<string, string>;
}

export interface SubscriptionItem {
	id: string;
	priceId: string;
	quantity: number;
	price: Price;
}

export interface Price {
	id: string;
	productId: string;
	unitAmount: number;
	currency: string;
	interval?: 'day' | 'week' | 'month' | 'year';
	intervalCount?: number;
	type: 'one_time' | 'recurring';
}

export interface Customer {
	id: string;
	email: string;
	name?: string;
	phone?: string;
	address?: {
		line1?: string;
		line2?: string;
		city?: string;
		state?: string;
		postalCode?: string;
		country?: string;
	};
	metadata?: Record<string, string>;
}

export interface Invoice {
	id: string;
	customerId: string;
	subscriptionId?: string;
	status: 'draft' | 'open' | 'paid' | 'uncollectible' | 'void';
	total: number;
	subtotal: number;
	tax?: number;
	amountPaid: number;
	amountRemaining: number;
	currency: string;
	dueDate?: number;
	paidAt?: number;
	hostedInvoiceUrl?: string;
	invoicePdf?: string;
}

class StripeService {
	private config: StripeConfig;
	private stripe: any; // Stripe instance

	constructor() {
		this.config = {
			publishableKey: import.meta.env.VITE_STRIPE_PUBLISHABLE_KEY || '',
			secretKey: import.meta.env.VITE_STRIPE_SECRET_KEY || '',
			webhookSecret: import.meta.env.VITE_STRIPE_WEBHOOK_SECRET || ''
		};
	}

	async initializeStripe(): Promise<void> {
		if (typeof window !== 'undefined' && !this.stripe) {
			// Dynamically import Stripe.js
			const { loadStripe } = await import('@stripe/stripe-js');
			this.stripe = await loadStripe(this.config.publishableKey);
		}
	}

	private getHeaders(): Record<string, string> {
		return {
			'Authorization': `Bearer ${this.config.secretKey}`,
			'Content-Type': 'application/x-www-form-urlencoded'
		};
	}

	// Payment Intents
	async createPaymentIntent(amount: number, currency: string = 'usd', metadata?: Record<string, string>): Promise<PaymentIntent> {
		try {
			const params = new URLSearchParams({
				amount: amount.toString(),
				currency: currency,
				automatic_payment_methods: JSON.stringify({ enabled: true })
			});

			if (metadata) {
				Object.entries(metadata).forEach(([key, value]) => {
					params.append(`metadata[${key}]`, value);
				});
			}

			const response = await fetch('https://api.stripe.com/v1/payment_intents', {
				method: 'POST',
				headers: this.getHeaders(),
				body: params
			});

			if (!response.ok) {
				throw new Error('Failed to create payment intent');
			}

			const data = await response.json();
			return this.mapPaymentIntent(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to create payment intent: ${errorMessage}`);
		}
	}

	async confirmPaymentIntent(paymentIntentId: string, paymentMethodId: string): Promise<PaymentIntent> {
		try {
			const params = new URLSearchParams({
				payment_method: paymentMethodId
			});

			const response = await fetch(`https://api.stripe.com/v1/payment_intents/${paymentIntentId}/confirm`, {
				method: 'POST',
				headers: this.getHeaders(),
				body: params
			});

			if (!response.ok) {
				throw new Error('Failed to confirm payment intent');
			}

			const data = await response.json();
			return this.mapPaymentIntent(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to confirm payment intent: ${errorMessage}`);
		}
	}

	// Customers
	async createCustomer(email: string, name?: string, metadata?: Record<string, string>): Promise<Customer> {
		try {
			const params = new URLSearchParams({
				email: email
			});

			if (name) params.append('name', name);

			if (metadata) {
				Object.entries(metadata).forEach(([key, value]) => {
					params.append(`metadata[${key}]`, value);
				});
			}

			const response = await fetch('https://api.stripe.com/v1/customers', {
				method: 'POST',
				headers: this.getHeaders(),
				body: params
			});

			if (!response.ok) {
				throw new Error('Failed to create customer');
			}

			const data = await response.json();
			return this.mapCustomer(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to create customer: ${errorMessage}`);
		}
	}

	async getCustomer(customerId: string): Promise<Customer> {
		try {
			const response = await fetch(`https://api.stripe.com/v1/customers/${customerId}`, {
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to get customer');
			}

			const data = await response.json();
			return this.mapCustomer(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get customer: ${errorMessage}`);
		}
	}

	// Subscriptions
	async createSubscription(customerId: string, priceId: string, metadata?: Record<string, string>): Promise<Subscription> {
		try {
			const params = new URLSearchParams({
				customer: customerId,
				'items[0][price]': priceId
			});

			if (metadata) {
				Object.entries(metadata).forEach(([key, value]) => {
					params.append(`metadata[${key}]`, value);
				});
			}

			const response = await fetch('https://api.stripe.com/v1/subscriptions', {
				method: 'POST',
				headers: this.getHeaders(),
				body: params
			});

			if (!response.ok) {
				throw new Error('Failed to create subscription');
			}

			const data = await response.json();
			return this.mapSubscription(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to create subscription: ${errorMessage}`);
		}
	}

	async getSubscription(subscriptionId: string): Promise<Subscription> {
		try {
			const response = await fetch(`https://api.stripe.com/v1/subscriptions/${subscriptionId}`, {
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to get subscription');
			}

			const data = await response.json();
			return this.mapSubscription(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get subscription: ${errorMessage}`);
		}
	}

	async cancelSubscription(subscriptionId: string, atPeriodEnd: boolean = true): Promise<Subscription> {
		try {
			const params = new URLSearchParams({
				cancel_at_period_end: atPeriodEnd.toString()
			});

			const response = await fetch(`https://api.stripe.com/v1/subscriptions/${subscriptionId}`, {
				method: 'POST',
				headers: this.getHeaders(),
				body: params
			});

			if (!response.ok) {
				throw new Error('Failed to cancel subscription');
			}

			const data = await response.json();
			return this.mapSubscription(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to cancel subscription: ${errorMessage}`);
		}
	}

	// Invoices
	async getInvoice(invoiceId: string): Promise<Invoice> {
		try {
			const response = await fetch(`https://api.stripe.com/v1/invoices/${invoiceId}`, {
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to get invoice');
			}

			const data = await response.json();
			return this.mapInvoice(data);
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get invoice: ${errorMessage}`);
		}
	}

	async getCustomerInvoices(customerId: string, limit: number = 10): Promise<Invoice[]> {
		try {
			const params = new URLSearchParams({
				customer: customerId,
				limit: limit.toString()
			});

			const response = await fetch(`https://api.stripe.com/v1/invoices?${params}`, {
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to get customer invoices');
			}

			const data = await response.json();
			return data.data.map((invoice: any) => this.mapInvoice(invoice));
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get customer invoices: ${errorMessage}`);
		}
	}

	// Webhook handling
	async verifyWebhookSignature(payload: string, signature: string): Promise<boolean> {
		try {
			// In a real implementation, use Stripe's webhook signature verification
			// This is a simplified version
			const expectedSignature = this.config.webhookSecret;
			return signature.includes(expectedSignature);
		} catch (error) {
			return false;
		}
	}

	// Frontend payment processing
	async processPayment(paymentMethodId: string, amount: number, currency: string = 'usd'): Promise<{ success: boolean; paymentIntent?: PaymentIntent; error?: string }> {
		try {
			await this.initializeStripe();

			if (!this.stripe) {
				throw new Error('Stripe not initialized');
			}

			// Create payment intent
			const paymentIntent = await this.createPaymentIntent(amount, currency);

			// Confirm payment
			const result = await this.stripe.confirmCardPayment(paymentIntent.clientSecret, {
				payment_method: paymentMethodId
			});

			if (result.error) {
				return {
					success: false,
					error: result.error.message
				};
			}

			return {
				success: true,
				paymentIntent: this.mapPaymentIntent(result.paymentIntent)
			};
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			return {
				success: false,
				error: errorMessage
			};
		}
	}

	// Helper mapping functions
	private mapPaymentIntent(data: any): PaymentIntent {
		return {
			id: data.id,
			amount: data.amount,
			currency: data.currency,
			status: data.status,
			clientSecret: data.client_secret,
			metadata: data.metadata
		};
	}

	private mapCustomer(data: any): Customer {
		return {
			id: data.id,
			email: data.email,
			name: data.name,
			phone: data.phone,
			address: data.address,
			metadata: data.metadata
		};
	}

	private mapSubscription(data: any): Subscription {
		return {
			id: data.id,
			customerId: data.customer,
			status: data.status,
			currentPeriodStart: data.current_period_start,
			currentPeriodEnd: data.current_period_end,
			cancelAtPeriodEnd: data.cancel_at_period_end,
			items: data.items.data.map((item: any) => ({
				id: item.id,
				priceId: item.price.id,
				quantity: item.quantity,
				price: {
					id: item.price.id,
					productId: item.price.product,
					unitAmount: item.price.unit_amount,
					currency: item.price.currency,
					interval: item.price.recurring?.interval,
					intervalCount: item.price.recurring?.interval_count,
					type: item.price.type
				}
			})),
			metadata: data.metadata
		};
	}

	private mapInvoice(data: any): Invoice {
		return {
			id: data.id,
			customerId: data.customer,
			subscriptionId: data.subscription,
			status: data.status,
			total: data.total,
			subtotal: data.subtotal,
			tax: data.tax,
			amountPaid: data.amount_paid,
			amountRemaining: data.amount_remaining,
			currency: data.currency,
			dueDate: data.due_date,
			paidAt: data.status_transitions?.paid_at,
			hostedInvoiceUrl: data.hosted_invoice_url,
			invoicePdf: data.invoice_pdf
		};
	}
}

export const stripeService = new StripeService(); 