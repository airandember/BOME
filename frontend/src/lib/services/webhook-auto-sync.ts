import { subscriberStoreActions } from '$lib/stores/subscribers-store';
import { showToast } from '$lib/toast';

export interface WebhookEvent {
	type: string;
	data: {
		object: any;
		previous_attributes?: any;
	};
	created: number;
	livemode: boolean;
}

export class StripeWebhookAutoSync {
	private static instance: StripeWebhookAutoSync;
	private eventSource: EventSource | null = null;
	private isConnected = false;
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 5;
	private reconnectDelay = 1000; // Start with 1 second
	
	private constructor() {}
	
	static getInstance(): StripeWebhookAutoSync {
		if (!StripeWebhookAutoSync.instance) {
			StripeWebhookAutoSync.instance = new StripeWebhookAutoSync();
		}
		return StripeWebhookAutoSync.instance;
	}
	
	/**
	 * Start listening for Stripe webhook events via Server-Sent Events
	 */
	startListening() {
		if (this.isConnected) {
			console.log('🔗 Webhook auto-sync already connected');
			return;
		}
		
		try {
			// Connect to the webhook event stream endpoint
			this.eventSource = new EventSource('/api/v1/webhooks/stripe/events-stream', {
				withCredentials: true
			});
			
			this.eventSource.onopen = () => {
				console.log('✅ Webhook auto-sync connected');
				this.isConnected = true;
				this.reconnectAttempts = 0;
				this.reconnectDelay = 1000;
				
				showToast({
					type: 'success',
					message: 'Real-time sync connected',
					duration: 3000
				});
			};
			
			this.eventSource.onmessage = (event) => {
				try {
					const webhookEvent: WebhookEvent = JSON.parse(event.data);
					this.handleWebhookEvent(webhookEvent);
				} catch (error) {
					console.error('❌ Failed to parse webhook event:', error);
				}
			};
			
			this.eventSource.onerror = (error) => {
				console.error('❌ Webhook auto-sync connection error:', error);
				this.isConnected = false;
				
				if (this.reconnectAttempts < this.maxReconnectAttempts) {
					this.scheduleReconnect();
				} else {
					console.error('❌ Max reconnection attempts reached');
					showToast({
						type: 'error',
						message: 'Real-time sync disconnected. Please refresh the page.',
						duration: 10000
					});
				}
			};
			
		} catch (error) {
			console.error('❌ Failed to start webhook auto-sync:', error);
		}
	}
	
	/**
	 * Stop listening for webhook events
	 */
	stopListening() {
		if (this.eventSource) {
			this.eventSource.close();
			this.eventSource = null;
		}
		this.isConnected = false;
		console.log('🔌 Webhook auto-sync disconnected');
	}
	
	/**
	 * Schedule a reconnection attempt
	 */
	private scheduleReconnect() {
		this.reconnectAttempts++;
		
		console.log(`🔄 Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts}) in ${this.reconnectDelay}ms...`);
		
		setTimeout(() => {
			this.stopListening();
			this.startListening();
		}, this.reconnectDelay);
		
		// Exponential backoff
		this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000); // Max 30 seconds
	}
	
	/**
	 * Handle incoming webhook events and update the subscriber store
	 */
	private async handleWebhookEvent(event: WebhookEvent) {
		console.log('🎣 Received webhook event:', event.type, event.data);
		
		switch (event.type) {
			case 'customer.created':
			case 'customer.updated':
				await this.handleCustomerEvent(event);
				break;
				
			case 'customer.subscription.created':
			case 'customer.subscription.updated':
			case 'customer.subscription.deleted':
				await this.handleSubscriptionEvent(event);
				break;
				
			case 'invoice.payment_succeeded':
			case 'invoice.payment_failed':
				await this.handlePaymentEvent(event);
				break;
				
			case 'product.created':
			case 'product.updated':
			case 'product.deleted':
				await this.handleProductEvent(event);
				break;
				
			case 'price.created':
			case 'price.updated':
			case 'price.deleted':
				await this.handlePriceEvent(event);
				break;
				
			default:
				console.log('ℹ️ Unhandled webhook event type:', event.type);
		}
	}
	
	/**
	 * Handle customer-related webhook events
	 */
	private async handleCustomerEvent(event: WebhookEvent) {
		const customer = event.data.object;
		
		try {
			// Refresh the subscriber data to get updated customer info
			await subscriberStoreActions.refresh();
			
			showToast({
				type: 'info',
				message: `Customer ${customer.email || customer.id} updated`,
				duration: 5000
			});
		} catch (error) {
			console.error('❌ Failed to handle customer event:', error);
		}
	}
	
	/**
	 * Handle subscription-related webhook events
	 */
	private async handleSubscriptionEvent(event: WebhookEvent) {
		const subscription = event.data.object;
		
		try {
			// Refresh the subscriber data to get updated subscription info
			await subscriberStoreActions.refresh();
			
			let message = '';
			switch (event.type) {
				case 'customer.subscription.created':
					message = 'New subscription created';
					break;
				case 'customer.subscription.updated':
					message = `Subscription ${subscription.id} updated`;
					break;
				case 'customer.subscription.deleted':
					message = `Subscription ${subscription.id} cancelled`;
					break;
			}
			
			showToast({
				type: event.type.includes('deleted') ? 'warning' : 'success',
				message,
				duration: 5000
			});
		} catch (error) {
			console.error('❌ Failed to handle subscription event:', error);
		}
	}
	
	/**
	 * Handle payment-related webhook events
	 */
	private async handlePaymentEvent(event: WebhookEvent) {
		const invoice = event.data.object;
		
		try {
			// Refresh the subscriber data to get updated payment status
			await subscriberStoreActions.refresh();
			
			const message = event.type === 'invoice.payment_succeeded' 
				? 'Payment successful' 
				: 'Payment failed';
			
			showToast({
				type: event.type === 'invoice.payment_succeeded' ? 'success' : 'error',
				message,
				duration: 5000
			});
		} catch (error) {
			console.error('❌ Failed to handle payment event:', error);
		}
	}
	
	/**
	 * Handle product-related webhook events
	 */
	private async handleProductEvent(event: WebhookEvent) {
		const product = event.data.object;
		
		try {
			// Refresh the subscriber data to get updated product info
			await subscriberStoreActions.refresh();
			
			showToast({
				type: 'info',
				message: `Product ${product.name || product.id} updated`,
				duration: 5000
			});
		} catch (error) {
			console.error('❌ Failed to handle product event:', error);
		}
	}
	
	/**
	 * Handle price-related webhook events
	 */
	private async handlePriceEvent(event: WebhookEvent) {
		const price = event.data.object;
		
		try {
			// Refresh the subscriber data to get updated pricing info
			await subscriberStoreActions.refresh();
			
			showToast({
				type: 'info',
				message: `Price ${price.id} updated`,
				duration: 5000
			});
		} catch (error) {
			console.error('❌ Failed to handle price event:', error);
		}
	}
	
	/**
	 * Get connection status
	 */
	getConnectionStatus() {
		return {
			isConnected: this.isConnected,
			reconnectAttempts: this.reconnectAttempts,
			maxReconnectAttempts: this.maxReconnectAttempts
		};
	}
}
