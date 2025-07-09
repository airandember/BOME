import { browser } from '$app/environment';
import { apiRequest } from '$lib/auth';
import type { AdminAnalytics } from '$lib/types/advertising';
import { WS_CONFIG, getWebSocketUrl } from '$lib/config/websocket';

interface AnalyticsEvent {
    type: string;
    timestamp: Date;
    user_id?: string;
    data: Record<string, any>;
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

export interface RealTimeMetrics {
    active_users: number;
    current_active_users: number;
    page_views_last_minute: number;
    current_streams: number;
    server_load: number;
    bandwidth_usage: string;
    recent_signups: number;
    recent_subscriptions: number;
    error_rate: number;
    response_time: number;
    events_last_minute: any[];
    live_events: any[];
    top_content_now: any[];
}

export interface AnalyticsResponse {
    metadata: {
        last_updated: string;
        version: string;
    };
    real_time: RealTimeMetrics;
    system_health: SystemHealth;
    users: {
        total: number;
        new_today: number;
        new_week: number;
        new_month: number;
        active_today: number;
        growth_rate: number;
        churn_rate: number;
        retention_rate: number;
        daily_active: Record<string, number>;
        weekly_active: Record<string, number>;
        monthly_active: Record<string, number>;
    };
    videos: {
        total: number;
        published: number;
        pending: number;
        draft: number;
        total_views: number;
        total_likes: number;
        total_comments: number;
        total_shares: number;
        avg_rating: number;
        views: Record<string, any>;
        engagement: Record<string, any>;
        completion_rates: Record<string, any>;
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
        cancelled: number;
        revenue_today: number;
        revenue_week: number;
        revenue_month: number;
        revenue_year: number;
        mrr: number;
        arr: number;
        ltv: number;
        plans: Array<{
            name: string;
            count: number;
            revenue: number;
        }>;
        history: Record<string, any>;
    };
    engagement: {
        avg_watch_time: string;
        completion_rate: number;
        like_ratio: number;
        comment_rate: number;
        share_count: number;
        bounce_rate: number;
        session_duration: string;
        pages_per_session: number;
        daily_stats: Record<string, any>;
        hourly_stats: Record<string, any>;
    };
    geographic: {
        top_countries: Array<{
            country: string;
            users: number;
            percentage: number;
        }>;
        top_states: Array<{
            state: string;
            users: number;
            percentage: number;
        }>;
        daily_distribution: Record<string, any>;
    };
    devices: {
        desktop: {
            users: number;
            percentage: number;
            avg_session: string;
        };
        mobile: {
            users: number;
            percentage: number;
            avg_session: string;
        };
        tablet: {
            users: number;
            percentage: number;
            avg_session: string;
        };
        browsers: Array<{
            name: string;
            users: number;
            percentage: number;
        }>;
    };
    time_series: {
        users: Array<{
            date: string;
            new_users: number;
            active_users: number;
            returning_users: number;
        }>;
        revenue: Array<{
            date: string;
            revenue: number;
            subscriptions: number;
            upgrades: number;
        }>;
        engagement: Array<{
            date: string;
            views: number;
            likes: number;
            comments: number;
            shares: number;
        }>;
    };
    conversion: {
        funnel: Array<{
            stage: string;
            count: number;
            conversion: number;
        }>;
        cohort_analysis: Array<{
            cohort: string;
            users: number;
            retention_30d: number;
            retention_90d: number;
        }>;
        daily_conversion: Record<string, any>;
    };
    events: any[];
    page_views: Record<string, any>;
    user_interactions: Record<string, any>;
}

export class AnalyticsService {
    private static instance: AnalyticsService;
    private cache: Map<string, { data: any; timestamp: number }> = new Map();
    private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes
    private eventQueue: AnalyticsEvent[] = [];
    private flushInterval: number = 5000; // 5 seconds
    private maxQueueSize: number = 50;
    private isProcessing: boolean = false;
    private sessionStartTime: number = Date.now();
    private lastPageViewTime: number = Date.now();
    private pageViewCount: number = 0;
    private ws: WebSocket | null = null;
    private wsReconnectTimeout: number | null = null;
    private wsSubscriptions: Set<string> = new Set();
    private wsReconnectAttempts: number = 0;
    private maxReconnectAttempts: number = 5;
    private realTimeInterval: number | null = null;
    private isProduction: boolean;

    private constructor() {
        this.isProduction = !browser || window.location.hostname === 'localhost' || window.location.hostname === '127.0.0.1';
        if (browser) {
            this.startPeriodicFlush();
            this.trackSessionStart();
            this.initializeWebSocket();
        }
    }

    public static getInstance(): AnalyticsService {
        if (!AnalyticsService.instance) {
            AnalyticsService.instance = new AnalyticsService();
        }
        return AnalyticsService.instance;
    }

    private initializeWebSocket() {
        if (!browser || this.ws) return;

        const tokens = JSON.parse(localStorage.getItem('bome_auth_tokens') || 'null');
        if (!tokens?.access_token) {
            console.debug('No auth token found, skipping WebSocket connection');
            return;
        }

        try {
            this.ws = new WebSocket(getWebSocketUrl(WS_CONFIG.ENDPOINTS.ANALYTICS, tokens.access_token));

            this.ws.onopen = () => {
                console.debug('Analytics WebSocket connection established');
                this.wsReconnectAttempts = 0;
                this.ws?.send(JSON.stringify({
                    type: WS_CONFIG.MESSAGE_TYPES.SUBSCRIBE,
                    metrics: [WS_CONFIG.METRICS.REALTIME]
                }));
            };

            this.ws.onclose = () => {
                console.debug('Analytics WebSocket connection closed');
                this.ws = null;
                // Only attempt to reconnect if we're still on a page that needs analytics
                if (document.visibilityState === 'visible') {
                    this.scheduleReconnect();
                }
            };

            this.ws.onerror = (error) => {
                console.debug('Analytics WebSocket error:', error);
                // Don't log the full error in production
                if (!this.isProduction) {
                    console.error('WebSocket error details:', error);
                }
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('Failed to parse WebSocket message:', error);
                }
            };
        } catch (error) {
            console.debug('Failed to initialize WebSocket connection');
            this.scheduleReconnect();
        }
    }

    private scheduleReconnect() {
        if (this.wsReconnectAttempts >= this.maxReconnectAttempts) {
            console.log('Max WebSocket reconnection attempts reached');
            return;
        }

        if (this.wsReconnectTimeout) {
            clearTimeout(this.wsReconnectTimeout);
        }

        const delay = Math.min(1000 * Math.pow(2, this.wsReconnectAttempts), 30000);
        this.wsReconnectAttempts++;

        this.wsReconnectTimeout = window.setTimeout(() => {
            this.initializeWebSocket();
        }, delay);
    }

    private resubscribeToMetrics() {
        this.wsSubscriptions.forEach(subscription => {
            this.ws?.send(JSON.stringify({ action: 'subscribe', type: subscription }));
        });
    }

    private handleWebSocketMessage(data: any) {
        // Handle real-time updates from WebSocket
        if (data.type === 'metrics_update') {
            // Update cached data or trigger events
            this.invalidateCache('realtime');
        }
    }

    private async makeAuthenticatedRequest(endpoint: string, options: RequestInit = {}): Promise<Response> {
        try {
            // Remove any double base URLs
            endpoint = endpoint.replace(/^http[s]?:\/\/[^/]+\/api\/v1/, '');
            
            // Make the request using the auth module
            const response = await apiRequest(endpoint, options);
            
            if (!response.ok) {
                console.error(` Analytics API request failed:`, response.statusText);
                throw new Error(`Analytics API request failed: ${response.statusText}`);
            }
            
            return response;
        } catch (error) {
            console.error(` Analytics API request failed:`, error);
            throw error;
        }
    }

    public async getAnalytics(period: string = '7d'): Promise<AnalyticsResponse> {
        const cacheKey = `analytics_${period}`;
        const cached = this.getCachedData(cacheKey);
        if (cached) return cached;

        try {
            const response = await this.makeAuthenticatedRequest(
                `/admin/dashboard/analytics?period=${period}`
            );
            const data = await response.json();
            this.setCachedData(cacheKey, data);
            return data;
        } catch (error) {
            console.error('Failed to fetch analytics:', error);
            throw error;
        }
    }

    public async getRealTimeMetrics(): Promise<RealTimeMetrics> {
        const cacheKey = 'realtime';
        const cached = this.getCachedData(cacheKey, 30000); // 30 second cache for real-time data
        if (cached) return cached;

        try {
            const response = await this.makeAuthenticatedRequest(
                '/admin/dashboard/analytics/realtime'
            );
            const data = await response.json();
            this.setCachedData(cacheKey, data);
            return data;
        } catch (error) {
            console.error('Failed to fetch real-time metrics:', error);
            throw error;
        }
    }

    public async getSystemHealth(): Promise<SystemHealth> {
        const cacheKey = 'system_health';
        const cached = this.getCachedData(cacheKey);
        if (cached) return cached;

        try {
            const response = await this.makeAuthenticatedRequest(
                '/admin/dashboard/analytics/system-health'
            );
            const data = await response.json();
            this.setCachedData(cacheKey, data);
            return data;
        } catch (error) {
            console.error('Failed to fetch system health:', error);
            throw error;
        }
    }

    public async exportAnalytics(format: 'csv' | 'json' = 'csv', period: string = '7d'): Promise<void> {
        try {
            const response = await this.makeAuthenticatedRequest(
                `/admin/dashboard/analytics/export?format=${format}&period=${period}`
            );

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `analytics_export.${format}`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Failed to export analytics:', error);
            throw new Error('Failed to export analytics');
        }
    }

    public async trackEvent(type: string, data: Record<string, any> = {}, userId?: string): Promise<void> {
        if (!browser) return;

        const event: AnalyticsEvent = {
            type,
            timestamp: new Date(),
            user_id: userId,
            data,
        };

        this.eventQueue.push(event);

        if (this.eventQueue.length >= this.maxQueueSize) {
            await this.flushEvents();
        }
    }

    private async flushEvents(): Promise<void> {
        if (this.isProcessing || this.eventQueue.length === 0) return;

        this.isProcessing = true;
        const events = [...this.eventQueue];
        this.eventQueue = [];

        try {
            await this.makeAuthenticatedRequest('/admin/dashboard/analytics/batch', {
                method: 'POST',
                body: JSON.stringify(events),
            });
        } catch (error) {
            console.error('Failed to flush analytics events:', error);
            // Re-queue events on failure
            this.eventQueue.unshift(...events);
        } finally {
            this.isProcessing = false;
        }
    }

    private startPeriodicFlush(): void {
        if (!browser) return;

        setInterval(() => {
            this.flushEvents();
        }, this.flushInterval);
    }

    private trackSessionStart(): void {
        this.trackEvent('session_start', {
            user_agent: navigator.userAgent,
            screen_resolution: `${screen.width}x${screen.height}`,
            referrer: document.referrer,
        });
    }

    private getCachedData(key: string, customDuration?: number): any {
        const cached = this.cache.get(key);
        if (!cached) return null;

        const duration = customDuration || this.CACHE_DURATION;
        if (Date.now() - cached.timestamp > duration) {
            this.cache.delete(key);
            return null;
        }

        return cached.data;
    }

    private setCachedData(key: string, data: any): void {
        this.cache.set(key, { data, timestamp: Date.now() });
    }

    private invalidateCache(key?: string): void {
        if (key) {
            this.cache.delete(key);
        } else {
            this.cache.clear();
        }
    }

    public trackPageView(path: string): void {
        this.trackEvent('page_view', {
            path,
            timestamp: Date.now(),
            session_duration: Date.now() - this.sessionStartTime,
        });
        this.pageViewCount++;
        this.lastPageViewTime = Date.now();
    }

    public async trackVideoEvent(videoId: string, action: string, data: Record<string, any> = {}): Promise<void> {
        await this.trackEvent('video', {
            video_id: videoId,
            action,
            ...data
        });
    }

    public destroy(): void {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        if (this.wsReconnectTimeout) {
            clearTimeout(this.wsReconnectTimeout);
            this.wsReconnectTimeout = null;
        }

        if (this.realTimeInterval) {
            clearInterval(this.realTimeInterval);
            this.realTimeInterval = null;
        }

        this.flushEvents();
    }
}

// Export the singleton instance
export const analytics = AnalyticsService.getInstance();