import { writable, type Writable } from 'svelte/store';
import { browser } from '$app/environment';

export interface AnalyticsData {
	metadata: {
		last_updated: string;
		version: string;
	};
	real_time: {
		active_users: number;
		current_streams: number;
		server_load: number;
		bandwidth_usage: string;
		recent_signups: number;
		recent_subscriptions: number;
		error_rate: number;
		response_time: number;
		live_events?: Array<{
			time: string;
			event: string;
			details: string;
		}>;
		top_content_now?: Array<{
			title: string;
			viewers: number;
		}>;
	};
	users: {
		total: number;
		new_today: number;
		new_week: number;
		new_month: number;
		active_today: number;
		growth_rate: number;
	};
	videos: {
		total: number;
		published: number;
		pending: number;
		draft: number;
		total_views: number;
		total_likes: number;
		avg_rating: number;
		top_categories: Array<{
			name: string;
			count: number;
			views: number;
		}>;
	};
	subscriptions: {
		active: number;
		new_today: number;
		new_week: number;
		new_month: number;
		revenue_today: number;
		revenue_week: number;
		revenue_month: number;
		mrr: number;
		arr: number;
	};
	activity: Array<{
		id: number;
		user_id: number;
		action: string;
		details: string;
		created_at: string;
	}>;
	period: string;
}

export interface SystemHealth {
	uptime: string;
	response_time: string;
	error_rate: string;
	storage_used: string;
	bandwidth_used: string;
	cdn_hits: string;
	database_size: string;
	active_sessions: number;
	last_write: string;
	total_events_tracked: number;
}

export interface MonitoringData {
	metrics: {
		cpu_usage: number;
		memory_usage: number;
		disk_usage: number;
		network_in: string;
		network_out: string;
		uptime: string;
		load_average: number[];
	};
	events: Array<{
		id: number;
		event_type: string;
		subsite: string;
		endpoint: string;
		status: string;
		response_time: number;
		payload_size: number;
		status_code?: number;
		error_message?: string;
		retry_count: number;
		created_at: string;
	}>;
	health: {
		streaming: {
			status: string;
			response_time: number;
			error_rate: number;
			active_connections: number;
			last_check: string;
		};
		articles: {
			status: string;
			response_time: number;
			error_rate: number;
			active_connections: number;
			last_check: string;
		};
		expo: {
			status: string;
			response_time: number;
			error_rate: number;
			active_connections: number;
			last_check: string;
		};
	};
	alerts: Array<{
		id: number;
		severity: string;
		title: string;
		message: string;
		subsite?: string;
		acknowledged: boolean;
		acknowledged_by?: number;
		acknowledged_at?: string;
		created_at: string;
	}>;
}

export interface CrossSubsiteAnalytics {
	stats: {
		[key: string]: {
			users: number;
			content: number;
			views: number;
			revenue: number;
			engagement_rate: number;
		};
	};
	timeframe: string;
	subsite: string;
}

export interface WebhookAnalytics {
	total_events: number;
	success_rate: number;
	avg_response_time: number;
	events_by_subsite: { [key: string]: number };
	events_by_type: { [key: string]: number };
	recent_failures: Array<{
		timestamp: string;
		event_type: string;
		subsite: string;
		error: string;
	}>;
}

class AnalyticsService {
	private analyticsStore: Writable<AnalyticsData | null> = writable(null);
	private systemHealthStore: Writable<SystemHealth | null> = writable(null);
	private monitoringStore: Writable<MonitoringData | null> = writable(null);
	private crossSubsiteStore: Writable<CrossSubsiteAnalytics | null> = writable(null);
	private webhookAnalyticsStore: Writable<WebhookAnalytics | null> = writable(null);
	private loadingStore: Writable<boolean> = writable(false);
	private errorStore: Writable<string | null> = writable(null);

	constructor() {
		if (browser) {
			// Initialize real-time updates
			this.startRealTimeUpdates();
		}
	}

	// Getters for stores
	get analytics() {
		return this.analyticsStore;
	}

	get systemHealth() {
		return this.systemHealthStore;
	}

	get monitoring() {
		return this.monitoringStore;
	}

	get crossSubsite() {
		return this.crossSubsiteStore;
	}

	get webhookAnalytics() {
		return this.webhookAnalyticsStore;
	}

	get loading() {
		return this.loadingStore;
	}

	get error() {
		return this.errorStore;
	}

	// Fetch analytics data
	async fetchAnalytics(period: string = '7d'): Promise<void> {
		this.loadingStore.set(true);
		this.errorStore.set(null);

		try {
			const response = await fetch(`/api/v1/admin/dashboard/analytics?period=${period}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: AnalyticsData = await response.json();
			this.analyticsStore.set(data);
		} catch (error) {
			console.error('Failed to fetch analytics:', error);
			this.errorStore.set(error instanceof Error ? error.message : 'Failed to fetch analytics');
		} finally {
			this.loadingStore.set(false);
		}
	}

	// Fetch real-time analytics
	async fetchRealTimeAnalytics(): Promise<void> {
		try {
			const response = await fetch('/api/v1/admin/dashboard/analytics/realtime', {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data = await response.json();
			// Update the real_time section of analytics store
			this.analyticsStore.update(current => {
				if (current) {
					return { ...current, real_time: data.real_time };
				}
				return current;
			});
		} catch (error) {
			console.error('Failed to fetch real-time analytics:', error);
		}
	}

	// Fetch system health
	async fetchSystemHealth(): Promise<void> {
		try {
			const response = await fetch('/api/v1/admin/dashboard/analytics/system-health', {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: SystemHealth = await response.json();
			this.systemHealthStore.set(data);
		} catch (error) {
			console.error('Failed to fetch system health:', error);
		}
	}

	// Fetch monitoring data
	async fetchMonitoringData(): Promise<void> {
		try {
			const response = await fetch('/api/v1/admin/monitoring', {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: MonitoringData = await response.json();
			this.monitoringStore.set(data);
		} catch (error) {
			console.error('Failed to fetch monitoring data:', error);
		}
	}

	// Fetch cross-subsite analytics
	async fetchCrossSubsiteAnalytics(timeframe: string = '24h', subsite: string = 'all'): Promise<void> {
		try {
			const response = await fetch(`/api/v1/admin/analytics/cross-subsite?timeframe=${timeframe}&subsite=${subsite}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data: CrossSubsiteAnalytics = await response.json();
			this.crossSubsiteStore.set(data);
		} catch (error) {
			console.error('Failed to fetch cross-subsite analytics:', error);
		}
	}

	// Fetch webhook analytics
	async fetchWebhookAnalytics(timeframe: string = '24h'): Promise<void> {
		try {
			const response = await fetch(`/api/v1/admin/analytics/webhooks?timeframe=${timeframe}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			const data = await response.json();
			this.webhookAnalyticsStore.set(data.analytics);
		} catch (error) {
			console.error('Failed to fetch webhook analytics:', error);
		}
	}

	// Acknowledge alert
	async acknowledgeAlert(alertId: number): Promise<boolean> {
		try {
			const response = await fetch(`/api/v1/admin/monitoring/alerts/${alertId}/acknowledge`, {
				method: 'POST',
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			// Refresh monitoring data after acknowledging
			await this.fetchMonitoringData();
			return true;
		} catch (error) {
			console.error('Failed to acknowledge alert:', error);
			return false;
		}
	}

	// Track analytics event
	async trackEvent(eventType: string, eventData: Record<string, any> = {}): Promise<void> {
		try {
			const payload = {
				event_type: eventType,
				session_id: this.getSessionId(),
				subsite: 'streaming', // Default subsite
				...eventData
			};

			await fetch('/api/v1/admin/dashboard/analytics/track', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify(payload)
			});
		} catch (error) {
			console.error('Failed to track event:', error);
		}
	}

	// Track batch events
	async trackBatchEvents(events: Array<{ event_type: string; [key: string]: any }>): Promise<void> {
		try {
			const payload = events.map(event => ({
				...event,
				session_id: this.getSessionId(),
				subsite: 'streaming'
			}));

			await fetch('/api/v1/admin/dashboard/analytics/batch', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify(payload)
			});
		} catch (error) {
			console.error('Failed to track batch events:', error);
		}
	}

	// Export analytics data
	async exportAnalytics(format: 'csv' | 'json' = 'csv', period: string = '7d'): Promise<void> {
		try {
			const response = await fetch(`/api/v1/admin/dashboard/analytics/export?format=${format}&period=${period}`, {
				credentials: 'include'
			});

			if (!response.ok) {
				throw new Error(`HTTP error! status: ${response.status}`);
			}

			if (format === 'csv') {
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `analytics_export_${new Date().toISOString().split('T')[0]}.csv`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);
			} else {
				const data = await response.json();
				const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = `analytics_export_${new Date().toISOString().split('T')[0]}.json`;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);
			}
		} catch (error) {
			console.error('Failed to export analytics:', error);
		}
	}

	// Start real-time updates
	private startRealTimeUpdates(): void {
		if (!browser) return;

		// Update real-time data every 30 seconds
		setInterval(() => {
			this.fetchRealTimeAnalytics();
		}, 30000);

		// Update system health every 2 minutes
		setInterval(() => {
			this.fetchSystemHealth();
		}, 120000);

		// Update monitoring data every 5 minutes
		setInterval(() => {
			this.fetchMonitoringData();
		}, 300000);
	}

	// Get session ID (simplified - in production this would come from session management)
	private getSessionId(): string {
		if (!browser) return '';
		
		let sessionId = sessionStorage.getItem('analytics_session_id');
		if (!sessionId) {
			sessionId = `sess_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
			sessionStorage.setItem('analytics_session_id', sessionId);
		}
		return sessionId;
	}

	// Clear stores
	clear(): void {
		this.analyticsStore.set(null);
		this.systemHealthStore.set(null);
		this.monitoringStore.set(null);
		this.crossSubsiteStore.set(null);
		this.webhookAnalyticsStore.set(null);
		this.errorStore.set(null);
	}
}

// Create singleton instance
export const analyticsService = new AnalyticsService(); 
