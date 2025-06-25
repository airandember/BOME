import { browser } from '$app/environment';
import { api } from '$lib/auth';
import type { AdminAnalytics } from '$lib/types/advertising';

interface AnalyticsEvent {
    type: string;
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

class AnalyticsService {
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
    private readonly MAX_RECONNECT_ATTEMPTS = 10;
    private readonly INITIAL_RECONNECT_DELAY = 1000; // 1 second
    private readonly MAX_RECONNECT_DELAY = 30000; // 30 seconds

    private constructor() {
        if (browser) {
            // Start periodic flush
            setInterval(() => this.flushEvents(), this.flushInterval);

            // Track page views
            this.trackPageView();

            // Track session start
            this.trackSessionStart();

            // Add event listeners for user interactions
            document.addEventListener('click', (e) => this.handleUserInteraction(e));
            document.addEventListener('scroll', this.debounce(() => this.handleScroll(), 500));

            // Track session end on page unload
            window.addEventListener('beforeunload', () => this.trackSessionEnd());

            // Track device and browser info
            this.trackDeviceInfo();
        }
        this.initializeWebSocket();
        this.setupEventListeners();
        this.startEventProcessing();
        window.addEventListener('beforeunload', () => this.cleanup());
    }

    public static getInstance(): AnalyticsService {
        if (!AnalyticsService.instance) {
            AnalyticsService.instance = new AnalyticsService();
        }
        return AnalyticsService.instance;
    }

    private async flushEvents() {
        if (this.eventQueue.length === 0 || this.isProcessing) {
            return;
        }

        this.isProcessing = true;
        const events = [...this.eventQueue];
        this.eventQueue = [];

        const maxRetries = 3;
        let retryCount = 0;
        let retryDelay = 1000; // Start with 1 second delay

        const tryFlush = async (): Promise<boolean> => {
            try {
                const response = await fetch('/api/v1/admin/dashboard/analytics/batch', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${localStorage.getItem('token')}`
                    },
                    body: JSON.stringify(events)
                });

                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }

                return true;
            } catch (error) {
                console.error('Failed to flush analytics events:', error);
                retryCount++;
                if (retryCount < maxRetries) {
                    await new Promise(resolve => setTimeout(resolve, retryDelay));
                    retryDelay *= 2; // Exponential backoff
                    return tryFlush();
                }
                // Store failed events for later retry
                this.storeFailedEvents(events);
                return false;
            }
        };

        await tryFlush();
        this.isProcessing = false;
    }

    private async retryFailedEvents() {
        try {
            const failedEvents = JSON.parse(localStorage.getItem('failedAnalyticsEvents') || '[]');
            if (failedEvents.length === 0) return;

            localStorage.removeItem('failedAnalyticsEvents');
            this.eventQueue.unshift(...failedEvents);
        } catch (error) {
            console.error('Failed to retry failed events:', error);
        }
    }

    private trackPageView() {
        const now = Date.now();
        const timeOnPreviousPage = now - this.lastPageViewTime;
        this.lastPageViewTime = now;
        this.pageViewCount++;

        this.queueEvent({
            type: 'page_view',
            data: {
                path: window.location.pathname,
                referrer: document.referrer,
                title: document.title,
                time_on_previous_page: timeOnPreviousPage,
                page_view_count: this.pageViewCount
            }
        });
    }

    private trackSessionStart() {
        this.queueEvent({
            type: 'session_start',
            data: {
                start_time: this.sessionStartTime,
                user_agent: navigator.userAgent,
                language: navigator.language,
                screen_size: `${window.screen.width}x${window.screen.height}`,
                viewport_size: `${window.innerWidth}x${window.innerHeight}`
            }
        });
    }

    private trackSessionEnd() {
        const sessionDuration = Date.now() - this.sessionStartTime;
        this.queueEvent({
            type: 'session_end',
            data: {
                duration: sessionDuration,
                pages_viewed: this.pageViewCount
            }
        });
    }

    private trackDeviceInfo() {
        const userAgent = navigator.userAgent;
        let deviceType = 'desktop';
        if (/Mobile|Android|iPhone/i.test(userAgent)) {
            deviceType = 'mobile';
        } else if (/iPad|Tablet/i.test(userAgent)) {
            deviceType = 'tablet';
        }

        this.queueEvent({
            type: 'device_info',
            data: {
                device_type: deviceType,
                browser: this.getBrowserInfo(),
                os: this.getOSInfo(),
                screen_resolution: `${window.screen.width}x${window.screen.height}`,
                pixel_ratio: window.devicePixelRatio,
                is_touch: 'ontouchstart' in window
            }
        });
    }

    private handleUserInteraction(e: MouseEvent) {
        const target = e.target as HTMLElement;
        if (!target) return;

        // Track button clicks
        if (target.tagName === 'BUTTON' || target.closest('button')) {
            this.queueEvent({
                type: 'button_click',
                data: {
                    button_text: target.textContent?.trim(),
                    button_id: target.id,
                    path: window.location.pathname
                }
            });
        }

        // Track link clicks
        if (target.tagName === 'A' || target.closest('a')) {
            this.queueEvent({
                type: 'link_click',
                data: {
                    href: (target as HTMLAnchorElement).href,
                    text: target.textContent?.trim(),
                    path: window.location.pathname
                }
            });
        }
    }

    private handleScroll() {
        const scrollDepth = Math.round((window.scrollY + window.innerHeight) / document.documentElement.scrollHeight * 100);
        this.queueEvent({
            type: 'scroll',
            data: {
                depth_percentage: scrollDepth,
                path: window.location.pathname
            }
        });
    }

    public trackEvent(type: string, data: Record<string, any>) {
        this.queueEvent({ type, data });
    }

    public trackVideoEvent(videoId: string, action: string, data: Record<string, any> = {}) {
        this.queueEvent({
            type: 'video_interaction',
            data: {
                video_id: videoId,
                action,
                ...data
            }
        });
    }

    public trackSubscriptionEvent(action: string, data: Record<string, any> = {}) {
        this.queueEvent({
            type: 'subscription',
            data: {
                action,
                timestamp: Date.now(),
                ...data
            }
        });
    }

    public trackError(error: Error, context: Record<string, any> = {}) {
        this.queueEvent({
            type: 'error',
            data: {
                message: error.message,
                stack: error.stack,
                ...context,
                path: window.location.pathname
            }
        });
    }

    private validateEvent(event: AnalyticsEvent): boolean {
        // Validate event type
        if (!event.type || typeof event.type !== 'string' || event.type.trim() === '') {
            console.error('Invalid event type:', event);
            return false;
        }

        // Validate data object
        if (!event.data || typeof event.data !== 'object') {
            console.error('Invalid event data:', event);
            return false;
        }

        // Validate specific event types
        switch (event.type) {
            case 'page_view':
                if (!event.data.path || typeof event.data.path !== 'string') {
                    console.error('Invalid page_view event data:', event);
                    return false;
                }
                break;

            case 'video_interaction':
                if (!event.data.video_id || !event.data.action) {
                    console.error('Invalid video_interaction event data:', event);
                    return false;
                }
                break;

            case 'subscription':
                if (!event.data.action || typeof event.data.action !== 'string') {
                    console.error('Invalid subscription event data:', event);
                    return false;
                }
                break;

            case 'error':
                if (!event.data.message) {
                    console.error('Invalid error event data:', event);
                    return false;
                }
                break;

            case 'device_info':
                if (!event.data.device_type || !event.data.browser) {
                    console.error('Invalid device_info event data:', event);
                    return false;
                }
                break;

            case 'session_start':
            case 'session_end':
                if (!event.data.timestamp) {
                    console.error(`Invalid ${event.type} event data:`, event);
                    return false;
                }
                break;
        }

        // Sanitize data values
        Object.keys(event.data).forEach(key => {
            const value = event.data[key];
            if (typeof value === 'string') {
                event.data[key] = this.sanitizeString(value);
            }
        });

        return true;
    }

    private sanitizeString(str: string): string {
        // Remove any HTML/script tags
        str = str.replace(/<[^>]*>/g, '');
        // Remove any potential script injection
        str = str.replace(/javascript:/gi, '');
        str = str.replace(/on\w+=/gi, '');
        // Trim whitespace
        str = str.trim();
        return str;
    }

    private queueEvent(event: AnalyticsEvent) {
        // Add user ID if available
        const userId = this.getUserId();
        if (userId) {
            event.user_id = userId;
        }

        // Add timestamp and session ID
        event.data = {
            ...event.data,
            timestamp: Date.now(),
            session_id: this.getSessionId()
        };

        // Validate event before queueing
        if (!this.validateEvent(event)) {
            console.error('Event validation failed:', event);
            return;
        }

        this.eventQueue.push(event);

        // Flush if queue is full
        if (this.eventQueue.length >= this.maxQueueSize) {
            this.flushEvents();
        }
    }

    private getBrowserInfo(): string {
        const ua = navigator.userAgent;
        if (ua.includes('Firefox')) return 'Firefox';
        if (ua.includes('Chrome')) return 'Chrome';
        if (ua.includes('Safari')) return 'Safari';
        if (ua.includes('Edge')) return 'Edge';
        if (ua.includes('Opera')) return 'Opera';
        return 'Other';
    }

    private getOSInfo(): string {
        const ua = navigator.userAgent;
        if (ua.includes('Windows')) return 'Windows';
        if (ua.includes('Mac')) return 'MacOS';
        if (ua.includes('Linux')) return 'Linux';
        if (ua.includes('Android')) return 'Android';
        if (ua.includes('iOS')) return 'iOS';
        return 'Other';
    }

    private getUserId(): string | undefined {
        // Implementation depends on your auth system
        return undefined;
    }

    private getSessionId(): string {
        // Implementation depends on your session management
        return 'session-' + this.sessionStartTime;
    }

    private debounce(func: Function, wait: number) {
        let timeout: ReturnType<typeof setTimeout>;
        return function executedFunction(...args: any[]) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    private getCacheKey(endpoint: string, params?: Record<string, string>): string {
        return `${endpoint}${params ? JSON.stringify(params) : ''}`;
    }

    private isDataCached(key: string): boolean {
        const cached = this.cache.get(key);
        if (!cached) return false;
        return Date.now() - cached.timestamp < this.CACHE_DURATION;
    }

    private getCachedData(key: string): any {
        return this.cache.get(key)?.data;
    }

    private setCachedData(key: string, data: any): void {
        this.cache.set(key, { data, timestamp: Date.now() });
    }

    public async getAnalytics(period: string = '7d', useCache: boolean = true): Promise<AnalyticsResponse> {
        const cacheKey = this.getCacheKey('/api/v1/admin/dashboard/analytics', { period });

        if (useCache && this.isDataCached(cacheKey)) {
            return this.getCachedData(cacheKey);
        }

        try {
            const response = await api.get(`/api/v1/admin/dashboard/analytics?period=${period}`);
            const data = response.data;

            if (useCache) {
                this.setCachedData(cacheKey, data);
            }

            return data;
        } catch (error) {
            console.error('Failed to fetch analytics:', error);
            throw error;
        }
    }

    public async getRealTimeMetrics(): Promise<RealTimeMetrics> {
        try {
            const response = await api.get('/api/v1/admin/dashboard/analytics/realtime');
            return response.data;
        } catch (error) {
            console.error('Failed to fetch real-time metrics:', error);
            return this.getDefaultRealTimeMetrics();
        }
    }

    private getDefaultRealTimeMetrics(): RealTimeMetrics {
        return {
            active_users: 0,
            current_active_users: 0,
            page_views_last_minute: 0,
            current_streams: 0,
            server_load: 0,
            bandwidth_usage: 'N/A',
            recent_signups: 0,
            recent_subscriptions: 0,
            error_rate: 0,
            response_time: 0,
            events_last_minute: [],
            live_events: [],
            top_content_now: []
        };
    }

    public async exportAnalytics(format: 'csv' | 'json', period: string = '7d'): Promise<Blob> {
        try {
            const response = await fetch(`/api/v1/admin/dashboard/analytics/export?format=${format}&period=${period}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    'Accept': format === 'csv' ? 'text/csv' : 'application/json'
                }
            });
            
            if (!response.ok) {
                throw new Error('Failed to export analytics');
            }
            
            return await response.blob();
        } catch (error) {
            console.error('Failed to export analytics:', error);
            throw error;
        }
    }

    private getDefaultSystemHealth(): SystemHealth {
        return {
            uptime: 'N/A',
            response_time: 'N/A',
            error_rate: 'N/A',
            storage_used: 'N/A',
            bandwidth_used: 'N/A',
            cdn_hits: 'N/A',
            database_size: 'N/A',
            active_sessions: 0,
            last_write: 'N/A',
            total_events_tracked: 0
        };
    }

    public clearCache(): void {
        this.cache.clear();
    }

    private initializeWebSocket() {
        if (!browser || this.ws) return;

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host;
        this.ws = new WebSocket(`${protocol}//${host}/api/v1/admin/dashboard/analytics/ws`);

        this.ws.onopen = () => {
            console.log('Analytics WebSocket connection established');
            this.wsReconnectAttempts = 0;
            this.resubscribeToMetrics();
        };

        this.ws.onclose = () => {
            console.log('Analytics WebSocket connection closed');
            this.ws = null;
            this.attemptReconnect();
        };

        this.ws.onerror = (error) => {
            console.error('Analytics WebSocket error:', error);
            this.ws?.close();
        };

        this.ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                this.handleWebSocketUpdate(data);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };
    }

    private attemptReconnect() {
        if (this.wsReconnectTimeout !== null) {
            window.clearTimeout(this.wsReconnectTimeout);
            this.wsReconnectTimeout = null;
        }

        if (this.wsReconnectAttempts >= this.MAX_RECONNECT_ATTEMPTS) {
            console.error('Max WebSocket reconnection attempts reached');
            return;
        }

        const delay = Math.min(
            this.INITIAL_RECONNECT_DELAY * Math.pow(2, this.wsReconnectAttempts),
            this.MAX_RECONNECT_DELAY
        );

        this.wsReconnectTimeout = window.setTimeout(() => {
            this.wsReconnectTimeout = null;
            this.wsReconnectAttempts++;
            console.log(`Attempting WebSocket reconnection (${this.wsReconnectAttempts}/${this.MAX_RECONNECT_ATTEMPTS})`);
            this.initializeWebSocket();
        }, delay);
    }

    private handleWebSocketUpdate(update: any) {
        switch (update.type) {
            case 'analytics_update':
                // Update real-time metrics in the UI
                this.updateRealTimeMetrics(update.data);
                break;
            case 'system_health':
                // Update system health metrics
                this.updateSystemHealth(update.data);
                break;
            case 'error':
                console.error('Analytics error:', update.message);
                break;
        }
    }

    private updateRealTimeMetrics(data: RealTimeMetrics) {
        // Emit event for UI components to update
        const event = new CustomEvent('analytics-update', { detail: data });
        window.dispatchEvent(event);
    }

    private updateSystemHealth(data: SystemHealth) {
        // Emit event for UI components to update
        const event = new CustomEvent('system-health-update', { detail: data });
        window.dispatchEvent(event);
    }

    private resubscribeToMetrics() {
        if (this.ws?.readyState === WebSocket.OPEN && this.wsSubscriptions.size > 0) {
            this.ws.send(JSON.stringify({
                type: 'subscribe',
                metrics: Array.from(this.wsSubscriptions)
            }));
        }
    }

    public subscribeToMetrics(metrics: string[]) {
        metrics.forEach(metric => this.wsSubscriptions.add(metric));
        if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'subscribe',
                metrics
            }));
        }
    }

    public unsubscribeFromMetrics(metrics: string[]) {
        metrics.forEach(metric => this.wsSubscriptions.delete(metric));
        if (this.ws?.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify({
                type: 'unsubscribe',
                metrics
            }));
        }
    }

    private cleanup() {
        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }
        if (this.wsReconnectTimeout !== null) {
            window.clearTimeout(this.wsReconnectTimeout);
            this.wsReconnectTimeout = null;
        }
        this.wsReconnectAttempts = 0;
    }

    private setupEventListeners() {
        window.addEventListener('click', this.handleUserInteraction.bind(this));
        window.addEventListener('scroll', this.handleScroll.bind(this));
        window.addEventListener('visibilitychange', () => {
            if (document.hidden) {
                this.trackSessionEnd();
            } else {
                this.trackSessionStart();
            }
        });
    }

    private startEventProcessing() {
        setInterval(() => {
            if (this.eventQueue.length > 0) {
                this.flushEvents();
            }
        }, this.flushInterval);
    }

    private storeFailedEvents(events: AnalyticsEvent[]) {
        try {
            const failedEvents = JSON.parse(localStorage.getItem('failedAnalyticsEvents') || '[]');
            failedEvents.push(...events);
            localStorage.setItem('failedAnalyticsEvents', JSON.stringify(failedEvents.slice(-1000))); // Keep last 1000 events
        } catch (error) {
            console.error('Failed to persist failed events:', error);
        }
    }
}

export const analyticsService = AnalyticsService.getInstance();