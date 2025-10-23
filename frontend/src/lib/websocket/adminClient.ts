/**
 * Admin WebSocket Client for Real-Time Updates
 * 
 * Provides real-time updates for the admin dashboard including:
 * - New subscriber notifications
 * - Subscription changes
 * - Payment events
 * - KPI updates
 * 
 * Uses Svelte 5 runes for reactivity
 */

import { writable } from 'svelte/store';

// Event types matching backend
export const EVENT_TYPES = {
	SUBSCRIBER_CREATED: 'subscriber.created',
	SUBSCRIBER_UPDATED: 'subscriber.updated',
	SUBSCRIPTION_CREATED: 'subscription.created',
	SUBSCRIPTION_UPDATED: 'subscription.updated',
	SUBSCRIPTION_CANCELED: 'subscription.canceled',
	PAYMENT_RECEIVED: 'payment.received',
	PAYMENT_FAILED: 'payment.failed',
	KPI_UPDATE: 'kpi.update',
	PLAN_CREATED: 'plan.created',
	PLAN_UPDATED: 'plan.updated',
	PLAN_DELETED: 'plan.deleted',
	SYSTEM_ALERT: 'system.alert',
	CONNECTED: 'connected',
	PING: 'ping',
	PONG: 'pong'
} as const;

export type EventType = typeof EVENT_TYPES[keyof typeof EVENT_TYPES];

export interface AdminEvent {
	type: EventType;
	timestamp: string;
	data: Record<string, any>;
	message?: string;
}

export interface WebSocketStats {
	active_connections: number;
	total_connections: number;
	total_messages: number;
	total_broadcasts: number;
	uptime_seconds: number;
	broadcast_queue_len: number;
}

class AdminWebSocketClient {
	private ws: WebSocket | null = null;
	private reconnectTimeout = 3000; // Start with 3 seconds
	private maxReconnectTimeout = 30000; // Max 30 seconds
	private reconnectAttempts = 0;
	private maxReconnectAttempts = 10;
	private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
	private pingTimer: ReturnType<typeof setInterval> | null = null;
	private shouldReconnect = true;
	private token: string | null = null;

	// Stores for reactive updates (Svelte stores for compatibility)
	public connected = writable(false);
	public events = writable<AdminEvent[]>([]);
	public latestEvent = writable<AdminEvent | null>(null);
	public kpis = writable<any>(null);
	public connectionError = writable<string | null>(null);

	// Event counters
	private eventCounts: Record<string, number> = {};

	/**
	 * Connect to the WebSocket server
	 * @param token - JWT access token for authentication
	 */
	connect(token: string) {
		if (!token) {
			console.error('❌ AdminWS: No token provided');
			this.connectionError.set('No authentication token provided');
			return;
		}

		this.token = token;
		this.shouldReconnect = true;

		// Determine WebSocket URL based on current location
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		const host = window.location.hostname;
		const port = import.meta.env.VITE_API_PORT || '8080';
		
		// Use /api/v1/ws/admin endpoint
		const wsUrl = `${protocol}//${host}:${port}/api/v1/ws/admin`;

		console.log(`🌐 AdminWS: Connecting to ${wsUrl}...`);

		try {
			// Create WebSocket connection
			// Note: We pass the token via query param OR via first message
			// Backend expects Authorization header from middleware, so we'll send it in first message
			this.ws = new WebSocket(wsUrl);

			// Connection opened
			this.ws.onopen = () => {
				console.log('✅ AdminWS: Connected!');
				this.connected.set(true);
				this.connectionError.set(null);
				this.reconnectAttempts = 0;
				this.reconnectTimeout = 3000; // Reset timeout

				// Send authentication token (if backend expects it via message)
				// this.send({ type: 'auth', token });

				// Start ping/pong keepalive
				this.startPingTimer();
			};

			// Message received
			this.ws.onmessage = (event) => {
				try {
					const adminEvent: AdminEvent = JSON.parse(event.data);
					this.handleEvent(adminEvent);
				} catch (error) {
					console.error('❌ AdminWS: Failed to parse message:', error);
				}
			};

			// Connection closed
			this.ws.onclose = (event) => {
				console.log(`❌ AdminWS: Disconnected (code: ${event.code}, reason: ${event.reason})`);
				this.connected.set(false);
				this.stopPingTimer();

				if (this.shouldReconnect) {
					this.attemptReconnect();
				}
			};

			// Connection error
			this.ws.onerror = (error) => {
				console.error('❌ AdminWS: Error:', error);
				this.connectionError.set('WebSocket connection error');
			};

		} catch (error) {
			console.error('❌ AdminWS: Failed to create WebSocket:', error);
			this.connectionError.set('Failed to create WebSocket connection');
			this.attemptReconnect();
		}
	}

	/**
	 * Disconnect from the WebSocket server
	 */
	disconnect() {
		console.log('🔌 AdminWS: Disconnecting...');
		this.shouldReconnect = false;
		this.stopPingTimer();

		if (this.reconnectTimer) {
			clearTimeout(this.reconnectTimer);
			this.reconnectTimer = null;
		}

		if (this.ws) {
			this.ws.close();
			this.ws = null;
		}

		this.connected.set(false);
	}

	/**
	 * Handle incoming WebSocket events
	 */
	private handleEvent(event: AdminEvent) {
		console.log(`📨 AdminWS: Event received - ${event.type}`, event.data);

		// Track event count
		this.eventCounts[event.type] = (this.eventCounts[event.type] || 0) + 1;

		// Update latest event
		this.latestEvent.set(event);

		// Add to events list (keep last 100)
		this.events.update(events => [event, ...events].slice(0, 100));

		// Handle specific event types
		switch (event.type) {
			case EVENT_TYPES.CONNECTED:
				console.log('🎉 AdminWS: Connection confirmed by server', event.data);
				break;

			case EVENT_TYPES.KPI_UPDATE:
				console.log('📊 AdminWS: KPI update received');
				this.kpis.set(event.data.kpis);
				break;

			case EVENT_TYPES.SUBSCRIBER_CREATED:
				console.log('👤 AdminWS: New subscriber!', event.data.subscriber);
				this.showNotification('New Subscriber!', event.message || 'A new user has signed up', 'success');
				break;

			case EVENT_TYPES.SUBSCRIBER_UPDATED:
				console.log('👤 AdminWS: Subscriber updated', event.data.subscriber);
				break;

			case EVENT_TYPES.SUBSCRIPTION_CREATED:
				console.log('💰 AdminWS: New subscription!', event.data.subscription);
				this.showNotification('New Subscription!', event.message || 'A user subscribed', 'success');
				break;

			case EVENT_TYPES.SUBSCRIPTION_UPDATED:
				console.log('💰 AdminWS: Subscription updated', event.data.subscription);
				break;

			case EVENT_TYPES.SUBSCRIPTION_CANCELED:
				console.log('❌ AdminWS: Subscription canceled', event.data.subscription);
				this.showNotification('Subscription Canceled', event.message || 'A subscription was canceled', 'warning');
				break;

			case EVENT_TYPES.PAYMENT_RECEIVED:
				console.log('💵 AdminWS: Payment received!', event.data.payment);
				this.showNotification('Payment Received!', event.message || 'Payment processed successfully', 'success');
				break;

			case EVENT_TYPES.PAYMENT_FAILED:
				console.log('⚠️ AdminWS: Payment failed', event.data.payment);
				this.showNotification('Payment Failed', event.message || 'A payment failed', 'error');
				break;

			case EVENT_TYPES.PLAN_CREATED:
				console.log('📦 AdminWS: Plan created', event.data.plan);
				break;

			case EVENT_TYPES.PLAN_UPDATED:
				console.log('📦 AdminWS: Plan updated', event.data.plan);
				break;

			case EVENT_TYPES.PLAN_DELETED:
				console.log('🗑️ AdminWS: Plan deleted', event.data.plan_id);
				break;

			case EVENT_TYPES.SYSTEM_ALERT:
				console.log('🚨 AdminWS: System alert', event.data);
				this.showNotification('System Alert', event.message || 'System alert received', 'warning');
				break;

			case EVENT_TYPES.PONG:
				// Pong received (keepalive response)
				break;

			default:
				console.log('❓ AdminWS: Unknown event type:', event.type);
		}
	}

	/**
	 * Send a message to the server
	 */
	private send(data: any) {
		if (this.ws && this.ws.readyState === WebSocket.OPEN) {
			this.ws.send(JSON.stringify(data));
		} else {
			console.warn('⚠️ AdminWS: Cannot send message, not connected');
		}
	}

	/**
	 * Start ping timer for keepalive
	 */
	private startPingTimer() {
		this.stopPingTimer();
		this.pingTimer = setInterval(() => {
			if (this.ws && this.ws.readyState === WebSocket.OPEN) {
				this.send({ type: EVENT_TYPES.PING });
			}
		}, 30000); // Ping every 30 seconds
	}

	/**
	 * Stop ping timer
	 */
	private stopPingTimer() {
		if (this.pingTimer) {
			clearInterval(this.pingTimer);
			this.pingTimer = null;
		}
	}

	/**
	 * Attempt to reconnect with exponential backoff
	 */
	private attemptReconnect() {
		if (this.reconnectAttempts >= this.maxReconnectAttempts) {
			console.error('❌ AdminWS: Max reconnection attempts reached');
			this.connectionError.set('Failed to reconnect after multiple attempts');
			return;
		}

		this.reconnectAttempts++;
		console.log(`🔄 AdminWS: Reconnecting... (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

		// Exponential backoff
		const timeout = Math.min(this.reconnectTimeout * Math.pow(1.5, this.reconnectAttempts), this.maxReconnectTimeout);

		this.reconnectTimer = setTimeout(() => {
			if (this.shouldReconnect && this.token) {
				this.connect(this.token);
			}
		}, timeout);
	}

	/**
	 * Show browser notification (if permitted)
	 */
	private showNotification(title: string, message: string, type: 'success' | 'error' | 'warning' = 'success') {
		// Check if browser notifications are supported and permitted
		if ('Notification' in window && Notification.permission === 'granted') {
			new Notification(title, {
				body: message,
				icon: '/favicon.png',
				badge: '/favicon.png'
			});
		}

		// Also emit custom event for in-app notifications
		window.dispatchEvent(new CustomEvent('admin-notification', {
			detail: { title, message, type }
		}));
	}

	/**
	 * Request browser notification permission
	 */
	async requestNotificationPermission(): Promise<boolean> {
		if (!('Notification' in window)) {
			console.log('❌ Browser does not support notifications');
			return false;
		}

		if (Notification.permission === 'granted') {
			return true;
		}

		if (Notification.permission !== 'denied') {
			const permission = await Notification.requestPermission();
			return permission === 'granted';
		}

		return false;
	}

	/**
	 * Get event statistics
	 */
	getEventStats(): Record<string, number> {
		return { ...this.eventCounts };
	}

	/**
	 * Clear event history
	 */
	clearEvents() {
		this.events.set([]);
		this.eventCounts = {};
	}
}

// Export singleton instance
export const adminWS = new AdminWebSocketClient();

// Export convenience functions
export function connectAdminWebSocket(token: string) {
	adminWS.connect(token);
}

export function disconnectAdminWebSocket() {
	adminWS.disconnect();
}

