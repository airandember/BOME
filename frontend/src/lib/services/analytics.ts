import { browser } from '$app/environment';
import { apiRequest } from '$lib/auth';
import type { AdminAnalytics } from '$lib/types/advertising';
import { WS_CONFIG, getWebSocketUrl } from '$lib/config/websocket';
import { SecureTokenStorage } from '$lib/auth';

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
    private isProduction: boolean = false;
    
    // Enhanced retry and persistence properties
    private retryAttempts: number = 0;
    private maxRetryAttempts: number = 3;
    private baseRetryDelay: number = 1000; // 1 second
    private maxRetryDelay: number = 30000; // 30 seconds
    private localStorageKey: string = 'bome_analytics_queue';
    private isOnline: boolean = browser ? navigator.onLine : true;

    private constructor() {
        // Only initialize if we're in the browser
        if (!browser) {
            console.warn('AnalyticsService: Running in SSR mode, skipping initialization');
            return;
        }

        this.isProduction = import.meta.env.PROD;
        
        // Load persisted events from localStorage
        this.loadPersistedEvents();
        
        // Start periodic flush
        this.startPeriodicFlush();
        
        // Track session start
        this.trackSessionStart();
        
        // Handle visibility change
        this.handleVisibilityChange();
        
        // Handle online/offline events
        this.handleOnlineOffline();
        
        // Initialize WebSocket for real-time updates
        if (this.isProduction) {
            this.initializeWebSocket();
        }
    }

    private handleVisibilityChange(): void {
        if (typeof document !== 'undefined') {
            document.addEventListener('visibilitychange', () => {
                if (document.visibilityState === 'visible') {
                    this.flushEvents();
                }
            });
        }
    }

    private handleOnlineOffline(): void {
        if (typeof window !== 'undefined') {
            window.addEventListener('online', () => {
                this.isOnline = true;
                this.flushEvents();
            });
            
            window.addEventListener('offline', () => {
                this.isOnline = false;
            });
        }
    }

    // Load persisted events from localStorage
    private loadPersistedEvents(): void {
        if (!browser) return;
        
        try {
            const persisted = localStorage.getItem(this.localStorageKey);
            if (persisted) {
                const rawEvents = JSON.parse(persisted) as any[];
                if (Array.isArray(rawEvents) && rawEvents.length > 0) {
                    // Convert timestamps back to Date objects
                    const events = rawEvents.map(event => ({
                        ...event,
                        timestamp: new Date(event.timestamp)
                    })) as AnalyticsEvent[];
                    
                    this.eventQueue.push(...events);
                    console.log(`📊 Analytics: Loaded ${events.length} persisted events`);
                }
            }
        } catch (error) {
            console.error('Failed to load persisted analytics events:', error);
        }
    }

    // Persist events to localStorage
    private persistEvents(): void {
        if (!browser) return;
        
        try {
            localStorage.setItem(this.localStorageKey, JSON.stringify(this.eventQueue));
        } catch (error) {
            console.error('Failed to persist analytics events:', error);
        }
    }

    // Clear persisted events
    private clearPersistedEvents(): void {
        if (!browser) return;
        
        try {
            localStorage.removeItem(this.localStorageKey);
        } catch (error) {
            console.error('Failed to clear persisted analytics events:', error);
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

        const tokens = SecureTokenStorage.getTokens();
        if (!tokens?.access_token) {
            console.debug('No auth token found, skipping WebSocket connection');
            return;
        }

        try {
            const wsUrl = getWebSocketUrl(WS_CONFIG.ENDPOINTS.ANALYTICS, tokens.access_token);
            this.ws = new WebSocket(wsUrl);

            this.ws.onopen = () => {
                console.log('📊 Analytics: WebSocket connected');
                this.wsReconnectAttempts = 0;
                this.resubscribeToMetrics();
            };

            this.ws.onmessage = (event) => {
                try {
                    const data = JSON.parse(event.data);
                    this.handleWebSocketMessage(data);
                } catch (error) {
                    console.error('📊 Analytics: Failed to parse WebSocket message:', error);
                }
            };

            this.ws.onclose = (event) => {
                console.log(`📊 Analytics: WebSocket closed (code: ${event.code}, reason: ${event.reason})`);
                this.ws = null;
                
                // Implement exponential backoff reconnection
                if (this.wsReconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect();
                } else {
                    console.error('📊 Analytics: Max WebSocket reconnection attempts exceeded');
                }
            };

            this.ws.onerror = (error) => {
                console.error('📊 Analytics: WebSocket error:', error);
            };

        } catch (error) {
            console.error('📊 Analytics: Failed to initialize WebSocket:', error);
        }
    }

    private scheduleReconnect() {
        if (this.wsReconnectTimeout) {
            clearTimeout(this.wsReconnectTimeout);
        }

        this.wsReconnectAttempts++;
        
        // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (max)
        const baseDelay = 1000; // 1 second
        const maxDelay = 30000; // 30 seconds
        const delay = Math.min(baseDelay * Math.pow(2, this.wsReconnectAttempts - 1), maxDelay);
        
        console.log(`📊 Analytics: Scheduling WebSocket reconnection in ${delay}ms (attempt ${this.wsReconnectAttempts}/${this.maxReconnectAttempts})`);
        
        this.wsReconnectTimeout = window.setTimeout(() => {
            this.wsReconnectTimeout = null;
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
                `/admin/analytics?period=${period}`
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
        try {
            // Enhanced validation
            if (!this.isValidEventType(type)) {
                console.error('Invalid event type:', type);
                return;
            }

            // Sanitize data
            const sanitizedData = this.sanitizeEventData(data);

            const event: AnalyticsEvent = {
                type,
                timestamp: new Date(),
                user_id: userId,
                data: sanitizedData
            };

            // Validate event
            if (!this.validateEvent(event)) {
                console.error('Event validation failed:', event);
                return;
            }

            // Check rate limiting
            if (this.isRateLimited(event)) {
                console.warn('Rate limit exceeded for event:', type);
                return;
            }

            this.eventQueue.push(event);
            this.persistEvents();

            // Flush immediately if queue is full
            if (this.eventQueue.length >= this.maxQueueSize) {
                await this.flushEvents();
            }
        } catch (error) {
            console.error('Error tracking event:', error);
            this.logError('track_failed', error, { type, data });
        }
    }

    // Enhanced validation methods
    private isValidEventType(eventType: string): boolean {
        const validEventTypes = new Set([
            'page_view', 'video_view', 'video_like', 'video_comment', 'video_share',
            'user_signup', 'user_login', 'subscription_created', 'subscription_cancelled',
            'payment_processed', 'search_performed', 'session_start', 'session_end',
            'error_occurred', 'video_play', 'video_pause', 'video_seek', 'video_complete',
            'form_submit', 'button_click', 'link_click', 'scroll', 'time_on_page'
        ]);

        return validEventTypes.has(eventType);
    }

    private sanitizeEventData(data: Record<string, any>): Record<string, any> {
        const sanitized: Record<string, any> = {};

        for (const [key, value] of Object.entries(data)) {
            const sanitizedKey = this.sanitizeKey(key);
            sanitized[sanitizedKey] = this.sanitizeValue(value);
        }

        return sanitized;
    }

    private sanitizeKey(key: string): string {
        const dangerousChars = ['<', '>', '"', "'", '&', 'javascript:', 'onload=', 'onerror='];
        let sanitized = key;
        
        for (const char of dangerousChars) {
            sanitized = sanitized.replaceAll(char, '');
        }

        if (sanitized.length > 100) {
            sanitized = sanitized.substring(0, 100);
        }

        return sanitized;
    }

    private sanitizeValue(value: any, depth: number = 0): any {
        if (depth > 5) return '[truncated]';

        if (typeof value === 'string') {
            return this.sanitizeString(value);
        } else if (Array.isArray(value)) {
            if (value.length > 100) value = value.slice(0, 100);
            return value.map((item: any) => this.sanitizeValue(item, depth + 1));
        } else if (value && typeof value === 'object') {
            const sanitized: Record<string, any> = {};
            let count = 0;
            
            for (const [key, val] of Object.entries(value)) {
                if (count >= 50) break;
                sanitized[this.sanitizeKey(key)] = this.sanitizeValue(val, depth + 1);
                count++;
            }
            
            return sanitized;
        } else if (typeof value === 'number') {
            if (isNaN(value) || !isFinite(value) || value < -999999999 || value > 999999999) {
                return 0;
            }
            return value;
        }

        return value;
    }

    private sanitizeString(str: string): string {
        if (!str) return '';

        let sanitized = str.replace(/<[^>]*>/g, '');

        const dangerousPatterns = [
            '<script>', '</script>', 'javascript:', 'vbscript:',
            'onload=', 'onerror=', 'onclick=', 'onmouseover=',
            '<iframe>', '</iframe>', '<object>', '</object>',
            '<embed>', '</embed>', '<form>', '</form>',
            'union select', 'drop table', 'delete from',
            'insert into', 'update set', 'alter table',
            'exec(', 'eval(', 'system(', 'shell_exec(',
            'document.cookie', 'localStorage', 'sessionStorage',
            'window.location', 'history.pushState'
        ];

        for (const pattern of dangerousPatterns) {
            sanitized = sanitized.replaceAll(pattern, '');
        }

        sanitized = sanitized.replace(/[\x00-\x1F\x7F]/g, '');

        if (sanitized.length > 1000) {
            sanitized = sanitized.substring(0, 1000);
        }

        return sanitized;
    }

    private validateEvent(event: AnalyticsEvent): boolean {
        if (!event.type || !event.timestamp) {
            return false;
        }

        const timestamp = new Date(event.timestamp);
        if (isNaN(timestamp.getTime())) {
            return false;
        }

        const now = new Date();
        const futureLimit = new Date(now.getTime() + 60000);
        if (timestamp > futureLimit) {
            return false;
        }

        if (this.detectSuspiciousActivity(event)) {
            return false;
        }

        return true;
    }

    private detectSuspiciousActivity(event: AnalyticsEvent): boolean {
        const recentEvents = this.eventQueue.filter(e => 
            e.timestamp > new Date(Date.now() - 60000)
        );

        if (recentEvents.length > 100) {
            return true;
        }

        const userAgent = navigator.userAgent.toLowerCase();
        const suspiciousPatterns = [
            'bot', 'crawler', 'spider', 'scraper',
            'curl', 'wget', 'python', 'java',
            'sqlmap', 'nikto', 'nmap'
        ];

        for (const pattern of suspiciousPatterns) {
            if (userAgent.includes(pattern)) {
                return true;
            }
        }

        return false;
    }

    private isRateLimited(event: AnalyticsEvent): boolean {
        const now = Date.now();
        const windowStart = now - 60000;

        const recentEvents = this.eventQueue.filter(e => {
            // Handle both Date objects and string timestamps
            let eventTime: number;
            if (e.timestamp instanceof Date) {
                eventTime = e.timestamp.getTime();
            } else if (typeof e.timestamp === 'string') {
                eventTime = new Date(e.timestamp).getTime();
            } else {
                eventTime = e.timestamp as number;
            }
            
            return e.type === event.type && eventTime > windowStart;
        });

        return recentEvents.length >= 50;
    }

    private logError(errorType: string, error: any, context: Record<string, any> = {}) {
        const logEntry = {
            timestamp: new Date().toISOString(),
            error_type: errorType,
            component: 'analytics_service',
            error: error?.message || String(error),
            context,
            user_agent: navigator.userAgent,
            url: window.location.href
        };

        console.error('Analytics Error:', logEntry);
        this.sendErrorToBackend(logEntry);
    }

    private async sendErrorToBackend(logEntry: any) {
        try {
            await fetch('/api/analytics/errors', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${this.getAuthToken()}`
                },
                body: JSON.stringify(logEntry)
            });
        } catch (error) {
            console.error('Failed to send error to backend:', error);
        }
    }

    private async flushEvents(): Promise<void> {
        if (this.isProcessing || this.eventQueue.length === 0 || !this.isOnline) return;

        this.isProcessing = true;
        const events = [...this.eventQueue];
        this.eventQueue = [];
        this.persistEvents(); // Persist immediately

        try {
            // Transform events from frontend format to backend format
            const transformedEvents = events.map(event => this.transformEventForBackend(event));

            const response = await this.makeAuthenticatedRequest('/admin/analytics/batch', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(transformedEvents),
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                
                // Handle rate limiting
                if (response.status === 429) {
                    const retryAfter = errorData.retry_after || 60;
                    console.warn(`📊 Analytics: Rate limited, retrying in ${retryAfter} seconds`);
                    
                    // Re-queue events and wait
                    this.eventQueue.unshift(...events);
                    this.persistEvents();
                    
                    setTimeout(() => {
                        this.flushEvents();
                    }, retryAfter * 1000);
                    return;
                }
                
                // Handle validation errors
                if (response.status === 400 && errorData.validation_errors) {
                    console.error('📊 Analytics: Validation errors:', errorData.validation_errors);
                    
                    // Re-queue only valid events if any were processed
                    if (errorData.valid_events > 0) {
                        const validEvents = events.slice(0, errorData.valid_events);
                        this.eventQueue.unshift(...validEvents);
                        this.persistEvents();
                    }
                    return;
                }
                
                throw new Error(`HTTP ${response.status}: ${errorData.error || response.statusText}`);
            }

            const result = await response.json();
            console.log(`📊 Analytics: Successfully processed ${result.processed}/${result.total} events`);
            
            // Reset retry attempts on success
            this.retryAttempts = 0;
            
            // Clear persisted events on successful flush
            this.clearPersistedEvents();
            
        } catch (error) {
            console.error('📊 Analytics: Failed to flush events:', error);
            
            // Implement exponential backoff retry
            if (this.retryAttempts < this.maxRetryAttempts) {
                this.retryAttempts++;
                const delay = Math.min(
                    this.baseRetryDelay * Math.pow(2, this.retryAttempts - 1),
                    this.maxRetryDelay
                );
                
                console.log(`📊 Analytics: Retrying in ${delay}ms (attempt ${this.retryAttempts}/${this.maxRetryAttempts})`);
                
                // Re-queue events
                this.eventQueue.unshift(...events);
                this.persistEvents();
                
                // Schedule retry
                setTimeout(() => {
                    this.isProcessing = false;
                    this.flushEvents();
                }, delay);
                return;
            } else {
                // Max retries exceeded, re-queue events
                console.error('📊 Analytics: Max retry attempts exceeded, re-queueing events');
                this.eventQueue.unshift(...events);
                this.persistEvents();
                this.retryAttempts = 0; // Reset for next time
            }
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
        if (!browser) return;
        
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
        this.clearPersistedEvents(); // Clear persisted events on destroy
    }

    private getAuthToken(): string | null {
        const tokens = SecureTokenStorage.getTokens();
        return tokens?.access_token || null;
    }

    private generateSessionId(): string {
        // Generate a valid session ID that matches backend validation
        const timestamp = this.sessionStartTime;
        const random = Math.random().toString(36).substr(2, 9);
        return `session_${timestamp}_${random}`;
    }

    private transformEventForBackend(event: AnalyticsEvent): any {
        // Transform frontend event format to backend format
        const transformed: any = {
            event_type: event.type,
            timestamp: event.timestamp.toISOString(),
            session_id: this.generateSessionId(),
            subsite: 'streaming',
            user_agent: navigator.userAgent,
            ip_address: '' // Will be set by backend
        };

        // Add user_id if present
        if (event.user_id) {
            transformed.user_id = event.user_id;
        }

        // Add event-specific data
        if (event.data) {
            // For page_view events, ensure path starts with "/"
            if (event.type === 'page_view' && event.data.path) {
                const path = event.data.path;
                transformed.path = path.startsWith('/') ? path : `/${path}`;
            }
            
            // Merge all data into the transformed event
            Object.assign(transformed, event.data);
        }

        return transformed;
    }
}

// Export the singleton instance
export const analytics = AnalyticsService.getInstance();
