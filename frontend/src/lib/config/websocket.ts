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
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const backendHost = window.location.hostname === 'localhost' ? 'localhost:8080' : window.location.host;
    return `${protocol}//${backendHost}${endpoint}?token=${encodeURIComponent(token)}`;
}; 