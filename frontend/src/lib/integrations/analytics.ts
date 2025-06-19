// Google Analytics and Monitoring Integration
export interface AnalyticsConfig {
	googleAnalyticsId: string;
	customDimensions?: Record<string, string>;
	enableDebug?: boolean;
}

export interface AnalyticsEvent {
	action: string;
	category: string;
	label?: string;
	value?: number;
	customDimensions?: Record<string, string>;
}

export interface MonitoringConfig {
	alertEmail: string;
	webhookUrl?: string;
	thresholds: {
		errorRate: number;
		responseTime: number;
		memoryUsage: number;
		diskUsage: number;
	};
}

export interface AlertData {
	type: 'error' | 'warning' | 'info';
	title: string;
	message: string;
	timestamp: string;
	metadata?: Record<string, any>;
}

// Google Analytics Service
class GoogleAnalyticsService {
	private config: AnalyticsConfig;
	private initialized = false;

	constructor() {
		this.config = {
			googleAnalyticsId: import.meta.env.VITE_GA_MEASUREMENT_ID || '',
			enableDebug: import.meta.env.VITE_GA_DEBUG === 'true'
		};
	}

	async initialize(): Promise<void> {
		if (this.initialized || !this.config.googleAnalyticsId) {
			return;
		}

		try {
			// Load Google Analytics script
			const script = document.createElement('script');
			script.async = true;
			script.src = `https://www.googletagmanager.com/gtag/js?id=${this.config.googleAnalyticsId}`;
			document.head.appendChild(script);

			// Initialize gtag
			(window as any).dataLayer = (window as any).dataLayer || [];
			(window as any).gtag = function() {
				(window as any).dataLayer.push(arguments);
			};

			(window as any).gtag('js', new Date());
			(window as any).gtag('config', this.config.googleAnalyticsId, {
				debug_mode: this.config.enableDebug
			});

			this.initialized = true;
			console.log('Google Analytics initialized');
		} catch (error) {
			console.error('Failed to initialize Google Analytics:', error);
		}
	}

	trackEvent(event: AnalyticsEvent): void {
		if (!this.initialized) {
			console.warn('Google Analytics not initialized');
			return;
		}

		try {
			const eventData: any = {
				event_category: event.category,
				event_label: event.label,
				value: event.value
			};

			// Add custom dimensions
			if (event.customDimensions) {
				Object.entries(event.customDimensions).forEach(([key, value]) => {
					eventData[`custom_parameter_${key}`] = value;
				});
			}

			(window as any).gtag('event', event.action, eventData);

			if (this.config.enableDebug) {
				console.log('Analytics event tracked:', event);
			}
		} catch (error) {
			console.error('Failed to track analytics event:', error);
		}
	}

	trackPageView(path: string, title?: string): void {
		if (!this.initialized) {
			console.warn('Google Analytics not initialized');
			return;
		}

		try {
			(window as any).gtag('config', this.config.googleAnalyticsId, {
				page_path: path,
				page_title: title
			});

			if (this.config.enableDebug) {
				console.log('Page view tracked:', { path, title });
			}
		} catch (error) {
			console.error('Failed to track page view:', error);
		}
	}

	trackUserProperty(property: string, value: string): void {
		if (!this.initialized) {
			console.warn('Google Analytics not initialized');
			return;
		}

		try {
			(window as any).gtag('config', this.config.googleAnalyticsId, {
				user_properties: {
					[property]: value
				}
			});

			if (this.config.enableDebug) {
				console.log('User property tracked:', { property, value });
			}
		} catch (error) {
			console.error('Failed to track user property:', error);
		}
	}

	trackPurchase(transactionId: string, value: number, currency: string = 'USD', items?: any[]): void {
		if (!this.initialized) {
			console.warn('Google Analytics not initialized');
			return;
		}

		try {
			(window as any).gtag('event', 'purchase', {
				transaction_id: transactionId,
				value: value,
				currency: currency,
				items: items || []
			});

			if (this.config.enableDebug) {
				console.log('Purchase tracked:', { transactionId, value, currency, items });
			}
		} catch (error) {
			console.error('Failed to track purchase:', error);
		}
	}
}

// Monitoring and Alerting Service
class MonitoringService {
	private config: MonitoringConfig;
	private metrics: Map<string, number> = new Map();
	private alerts: AlertData[] = [];

	constructor() {
		this.config = {
			alertEmail: import.meta.env.VITE_ALERT_EMAIL || '',
			webhookUrl: import.meta.env.VITE_WEBHOOK_URL || '',
			thresholds: {
				errorRate: parseFloat(import.meta.env.VITE_ERROR_RATE_THRESHOLD) || 5, // 5%
				responseTime: parseFloat(import.meta.env.VITE_RESPONSE_TIME_THRESHOLD) || 2000, // 2 seconds
				memoryUsage: parseFloat(import.meta.env.VITE_MEMORY_USAGE_THRESHOLD) || 80, // 80%
				diskUsage: parseFloat(import.meta.env.VITE_DISK_USAGE_THRESHOLD) || 85 // 85%
			}
		};
	}

	recordMetric(name: string, value: number): void {
		this.metrics.set(name, value);
		this.checkThresholds(name, value);
	}

	private checkThresholds(metricName: string, value: number): void {
		const threshold = this.config.thresholds[metricName as keyof typeof this.config.thresholds];
		
		if (threshold && value > threshold) {
			this.createAlert('warning', `${metricName} threshold exceeded`, 
				`${metricName} is ${value}, which exceeds the threshold of ${threshold}`);
		}
	}

	createAlert(type: AlertData['type'], title: string, message: string, metadata?: Record<string, any>): void {
		const alert: AlertData = {
			type,
			title,
			message,
			timestamp: new Date().toISOString(),
			metadata
		};

		this.alerts.push(alert);
		this.sendAlert(alert);

		// Keep only last 100 alerts
		if (this.alerts.length > 100) {
			this.alerts = this.alerts.slice(-100);
		}
	}

	private async sendAlert(alert: AlertData): Promise<void> {
		try {
			// Send email alert
			if (this.config.alertEmail) {
				await this.sendEmailAlert(alert);
			}

			// Send webhook alert
			if (this.config.webhookUrl) {
				await this.sendWebhookAlert(alert);
			}

			console.log('Alert sent:', alert);
		} catch (error) {
			console.error('Failed to send alert:', error);
		}
	}

	private async sendEmailAlert(alert: AlertData): Promise<void> {
		// Mock email sending - replace with actual email service
		const emailData = {
			to: this.config.alertEmail,
			subject: `[BOME Alert] ${alert.title}`,
			body: `
				Alert Type: ${alert.type.toUpperCase()}
				Title: ${alert.title}
				Message: ${alert.message}
				Timestamp: ${alert.timestamp}
				
				${alert.metadata ? `Metadata: ${JSON.stringify(alert.metadata, null, 2)}` : ''}
			`
		};

		console.log('Email alert would be sent:', emailData);
	}

	private async sendWebhookAlert(alert: AlertData): Promise<void> {
		try {
			await fetch(this.config.webhookUrl!, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify(alert)
			});
		} catch (error) {
			console.error('Failed to send webhook alert:', error);
		}
	}

	getMetrics(): Record<string, number> {
		return Object.fromEntries(this.metrics);
	}

	getAlerts(): AlertData[] {
		return [...this.alerts];
	}

	clearAlerts(): void {
		this.alerts = [];
	}

	// System Health Monitoring
	async checkSystemHealth(): Promise<{
		status: 'healthy' | 'warning' | 'critical';
		checks: Record<string, { status: string; message: string; value?: number }>;
	}> {
		const checks: Record<string, { status: string; message: string; value?: number }> = {};

		// Check API health
		try {
			const apiStart = Date.now();
			const response = await fetch('/api/health', { method: 'GET' });
			const apiTime = Date.now() - apiStart;
			
			checks.api = {
				status: response.ok ? 'healthy' : 'unhealthy',
				message: response.ok ? 'API is responding' : 'API is not responding',
				value: apiTime
			};

			this.recordMetric('responseTime', apiTime);
		} catch (error) {
			checks.api = {
				status: 'unhealthy',
				message: 'API is unreachable'
			};
		}

		// Check memory usage
		if ('memory' in performance) {
			const memoryUsage = (performance as any).memory.usedJSHeapSize / (performance as any).memory.jsHeapSizeLimit * 100;
			checks.memory = {
				status: memoryUsage > this.config.thresholds.memoryUsage ? 'warning' : 'healthy',
				message: `Memory usage: ${memoryUsage.toFixed(1)}%`,
				value: memoryUsage
			};

			this.recordMetric('memoryUsage', memoryUsage);
		}

		// Check local storage usage
		try {
			const storageUsage = this.getLocalStorageUsage();
			checks.storage = {
				status: storageUsage > 80 ? 'warning' : 'healthy',
				message: `Local storage usage: ${storageUsage.toFixed(1)}%`,
				value: storageUsage
			};
		} catch (error) {
			checks.storage = {
				status: 'unknown',
				message: 'Could not check storage usage'
			};
		}

		// Determine overall status
		const hasUnhealthy = Object.values(checks).some(check => check.status === 'unhealthy');
		const hasWarning = Object.values(checks).some(check => check.status === 'warning');
		
		const status = hasUnhealthy ? 'critical' : hasWarning ? 'warning' : 'healthy';

		return { status, checks };
	}

	private getLocalStorageUsage(): number {
		let totalSize = 0;
		for (const key in localStorage) {
			if (localStorage.hasOwnProperty(key)) {
				totalSize += localStorage[key].length + key.length;
			}
		}
		
		// Approximate percentage (localStorage limit is usually 5-10MB)
		const approximateLimit = 5 * 1024 * 1024; // 5MB
		return (totalSize / approximateLimit) * 100;
	}

	// Performance monitoring
	startPerformanceMonitoring(): void {
		if (typeof window === 'undefined') return;

		// Monitor page load performance
		window.addEventListener('load', () => {
			if (performance.timing) {
				const loadTime = performance.timing.loadEventEnd - performance.timing.navigationStart;
				this.recordMetric('pageLoadTime', loadTime);
			}
		});

		// Monitor errors
		window.addEventListener('error', (event) => {
			this.createAlert('error', 'JavaScript Error', event.message, {
				filename: event.filename,
				lineno: event.lineno,
				colno: event.colno,
				stack: event.error?.stack
			});
		});

		// Monitor unhandled promise rejections
		window.addEventListener('unhandledrejection', (event) => {
			this.createAlert('error', 'Unhandled Promise Rejection', event.reason?.message || 'Unknown error', {
				reason: event.reason
			});
		});

		// Periodic health checks
		setInterval(() => {
			this.checkSystemHealth();
		}, 60000); // Every minute
	}
}

// Social Media Sharing Service
class SocialSharingService {
	static shareUrl(platform: 'facebook' | 'twitter' | 'linkedin' | 'reddit', url: string, title?: string, description?: string): void {
		const encodedUrl = encodeURIComponent(url);
		const encodedTitle = encodeURIComponent(title || '');
		const encodedDescription = encodeURIComponent(description || '');

		let shareUrl = '';

		switch (platform) {
			case 'facebook':
				shareUrl = `https://www.facebook.com/sharer/sharer.php?u=${encodedUrl}`;
				break;
			case 'twitter':
				shareUrl = `https://twitter.com/intent/tweet?url=${encodedUrl}&text=${encodedTitle}`;
				break;
			case 'linkedin':
				shareUrl = `https://www.linkedin.com/sharing/share-offsite/?url=${encodedUrl}`;
				break;
			case 'reddit':
				shareUrl = `https://reddit.com/submit?url=${encodedUrl}&title=${encodedTitle}`;
				break;
		}

		if (shareUrl) {
			window.open(shareUrl, '_blank', 'width=600,height=400');
		}
	}

	static async copyToClipboard(text: string): Promise<boolean> {
		try {
			if (navigator.clipboard) {
				await navigator.clipboard.writeText(text);
				return true;
			} else {
				// Fallback for older browsers
				const textArea = document.createElement('textarea');
				textArea.value = text;
				document.body.appendChild(textArea);
				textArea.select();
				document.execCommand('copy');
				document.body.removeChild(textArea);
				return true;
			}
		} catch (error) {
			console.error('Failed to copy to clipboard:', error);
			return false;
		}
	}

	static generateShareData(title: string, url: string, description?: string): {
		title: string;
		url: string;
		text?: string;
	} {
		return {
			title,
			url,
			text: description
		};
	}

	static async nativeShare(data: { title: string; url: string; text?: string }): Promise<boolean> {
		if (navigator.share) {
			try {
				await navigator.share(data);
				return true;
			} catch (error) {
				console.error('Native sharing failed:', error);
				return false;
			}
		}
		return false;
	}
}

// Export services
export const analyticsService = new GoogleAnalyticsService();
export const monitoringService = new MonitoringService();
export const socialSharingService = SocialSharingService; 