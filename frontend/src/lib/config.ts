// Configuration file for environment variables and app settings
// This centralizes all configuration and provides fallbacks

interface AppConfig {
	apiBaseUrl: string;
	wsUrl: string;
	appName: string;
	appVersion: string;
	environment: 'development' | 'production' | 'staging';
}

// Get environment variables with fallbacks
const getEnvVar = (key: string, fallback: string): string => {
	if (typeof import.meta !== 'undefined' && import.meta.env) {
		return import.meta.env[key] || fallback;
	}
	return fallback;
};

// Determine environment
const getEnvironment = (): 'development' | 'production' | 'staging' => {
	const env = getEnvVar('MODE', 'development');
	if (env === 'production') return 'production';
	if (env === 'staging') return 'staging';
	return 'development';
};

// Validate required environment variables
const validateConfig = () => {
	const apiBaseUrl = getEnvVar('VITE_API_BASE_URL', '');
	if (!apiBaseUrl) {
		console.error('❌ VITE_API_BASE_URL environment variable is required but not set');
		throw new Error('Missing required environment variable: VITE_API_BASE_URL');
	}
	return apiBaseUrl;
};

// Configuration object - relies on environment variables
export const config: AppConfig = {
	// Use environment variable - no hardcoded fallbacks
	apiBaseUrl: validateConfig(),
	wsUrl: getEnvVar('VITE_WS_URL', ''),
	appName: getEnvVar('VITE_APP_NAME', 'Book of Mormon Evidences'),
	appVersion: getEnvVar('VITE_APP_VERSION', '1.0.0'),
	environment: getEnvironment()
};

// Helper function to get the base URL without /api/v1
export const getApiBaseUrl = (): string => {
	return config.apiBaseUrl.replace('/api/v1', '');
};

// Helper function to check if we're in production
export const isProduction = (): boolean => {
	return config.environment === 'production';
};

// Helper function to check if we're in development
export const isDevelopment = (): boolean => {
	return config.environment === 'development';
};

// Export individual config values for convenience
export const { apiBaseUrl, wsUrl, appName, appVersion, environment } = config;

// Debug logging in development
if (isDevelopment() && typeof console !== 'undefined') {
	console.log('🔧 Config loaded:', {
		apiBaseUrl: config.apiBaseUrl,
		wsUrl: config.wsUrl,
		environment: config.environment
	});
}
