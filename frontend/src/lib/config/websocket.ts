export const WS_CONFIG = {
    ENDPOINTS: {
        ANALYTICS: '/api/v1/ws',
    },
    METRICS: {
        REALTIME: 'realtime_metrics',
    },
    MESSAGE_TYPES: {
        SUBSCRIBE: 'subscribe',
        METRICS_UPDATE: 'metrics_update',
    },
    RECONNECT: {
        MAX_ATTEMPTS: 5,
        BASE_DELAY: 1000,
        MAX_DELAY: 30000,
    }
} as const;

export const getWebSocketUrl = (endpoint: string, token: string): string => {
    // Use environment variable if available, otherwise construct from current location
    const wsBaseUrl = import.meta.env.VITE_WS_URL;
    
    if (wsBaseUrl) {
        // Environment variable is set (production)
        return `${wsBaseUrl}${endpoint}?token=${encodeURIComponent(token)}`;
    } else {
        // Development fallback
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const backendHost = window.location.host;
        return `${protocol}//${backendHost}${endpoint}?token=${encodeURIComponent(token)}`;
    }
}; 
