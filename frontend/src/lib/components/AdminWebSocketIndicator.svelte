<script lang="ts">
/**
 * AdminWebSocketIndicator - Shows WebSocket connection status
 * 
 * Usage: Add this component to your admin layout to show connection status
 * and receive real-time notifications
 */

import { onMount, onDestroy } from 'svelte';
import { adminWS, EVENT_TYPES, type AdminEvent } from '$lib/websocket/adminClient';
import { auth } from '$lib/auth';

// Reactive state (Svelte 5 runes)
let connected = $state(false);
let latestEvent = $state<AdminEvent | null>(null);
let showToast = $state(false);
let toastMessage = $state('');
let toastType = $state<'success' | 'error' | 'warning'>('success');

// Subscribe to WebSocket stores
let unsubscribeConnected: (() => void) | null = null;
let unsubscribeEvent: (() => void) | null = null;

onMount(() => {
	// Subscribe to connection status
	unsubscribeConnected = adminWS.connected.subscribe(value => {
		connected = value;
	});

	// Subscribe to latest event
	unsubscribeEvent = adminWS.latestEvent.subscribe(event => {
		if (event) {
			latestEvent = event;
			
			// Show toast for important events
			if (event.message && shouldShowToast(event.type)) {
				showToastNotification(event.message, getToastType(event.type));
			}
		}
	});

	// Connect WebSocket if user is authenticated
	const token = $auth.tokens?.accessToken;
	if (token) {
		console.log('🌐 Connecting to admin WebSocket...');
		adminWS.connect(token);
		
		// Request notification permission
		adminWS.requestNotificationPermission();
	}
});

onDestroy(() => {
	// Unsubscribe from stores
	if (unsubscribeConnected) unsubscribeConnected();
	if (unsubscribeEvent) unsubscribeEvent();
	
	// Disconnect WebSocket
	adminWS.disconnect();
});

function shouldShowToast(eventType: string): boolean {
	// Only show toast for important events
	return [
		EVENT_TYPES.SUBSCRIBER_CREATED,
		EVENT_TYPES.PAYMENT_RECEIVED,
		EVENT_TYPES.PAYMENT_FAILED,
		EVENT_TYPES.SUBSCRIPTION_CANCELED,
		EVENT_TYPES.SYSTEM_ALERT
	].includes(eventType);
}

function getToastType(eventType: string): 'success' | 'error' | 'warning' {
	if (eventType === EVENT_TYPES.PAYMENT_FAILED) return 'error';
	if (eventType === EVENT_TYPES.SUBSCRIPTION_CANCELED || eventType === EVENT_TYPES.SYSTEM_ALERT) return 'warning';
	return 'success';
}

function showToastNotification(message: string, type: 'success' | 'error' | 'warning' = 'success') {
	toastMessage = message;
	toastType = type;
	showToast = true;
	
	// Auto-hide after 5 seconds
	setTimeout(() => {
		showToast = false;
	}, 5000);
}

function getEventIcon(eventType: string): string {
	switch (eventType) {
		case EVENT_TYPES.SUBSCRIBER_CREATED:
			return '👤';
		case EVENT_TYPES.SUBSCRIPTION_CREATED:
			return '💰';
		case EVENT_TYPES.PAYMENT_RECEIVED:
			return '💵';
		case EVENT_TYPES.PAYMENT_FAILED:
			return '⚠️';
		case EVENT_TYPES.KPI_UPDATE:
			return '📊';
		default:
			return '📨';
	}
}
</script>

<!-- Connection Status Indicator -->
<div class="websocket-indicator" class:connected={connected}>
	<div class="status-dot" class:pulse={connected}></div>
	<span class="status-text">
		{#if connected}
			LIVE
		{:else}
			Connecting...
		{/if}
	</span>
</div>

<!-- Toast Notification -->
{#if showToast}
	<div class="toast toast-{toastType}">
		<div class="toast-icon">
			{#if latestEvent}
				{getEventIcon(latestEvent.type)}
			{:else}
				📨
			{/if}
		</div>
		<div class="toast-content">
			<p class="toast-message">{toastMessage}</p>
			{#if latestEvent}
				<p class="toast-time">{new Date(latestEvent.timestamp).toLocaleTimeString()}</p>
			{/if}
		</div>
		<button class="toast-close" onclick={() => showToast = false}>×</button>
	</div>
{/if}

<style>
	.websocket-indicator {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		background: rgba(239, 68, 68, 0.1);
		border-radius: 9999px;
		font-size: 0.875rem;
		font-weight: 600;
		transition: all 0.3s ease;
	}

	.websocket-indicator.connected {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #ef4444;
		transition: all 0.3s ease;
	}

	.websocket-indicator.connected .status-dot {
		background: #10b981;
	}

	.status-dot.pulse {
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0%, 100% { 
			opacity: 1; 
			transform: scale(1);
		}
		50% { 
			opacity: 0.7; 
			transform: scale(1.1);
		}
	}

	.status-text {
		color: inherit;
	}

	/* Toast Notification */
	.toast {
		position: fixed;
		top: 1rem;
		right: 1rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem 1.5rem;
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
		z-index: 9999;
		animation: slideIn 0.3s ease-out;
		min-width: 300px;
		max-width: 400px;
	}

	@keyframes slideIn {
		from {
			transform: translateX(100%);
			opacity: 0;
		}
		to {
			transform: translateX(0);
			opacity: 1;
		}
	}

	.toast-success {
		border-left: 4px solid #10b981;
	}

	.toast-error {
		border-left: 4px solid #ef4444;
	}

	.toast-warning {
		border-left: 4px solid #f59e0b;
	}

	.toast-icon {
		font-size: 2rem;
		flex-shrink: 0;
	}

	.toast-content {
		flex: 1;
	}

	.toast-message {
		margin: 0;
		font-weight: 600;
		color: #1f2937;
		font-size: 0.875rem;
	}

	.toast-time {
		margin: 0.25rem 0 0 0;
		font-size: 0.75rem;
		color: #6b7280;
	}

	.toast-close {
		background: none;
		border: none;
		font-size: 1.5rem;
		color: #9ca3af;
		cursor: pointer;
		padding: 0;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: color 0.2s ease;
		flex-shrink: 0;
	}

	.toast-close:hover {
		color: #4b5563;
	}
</style>

