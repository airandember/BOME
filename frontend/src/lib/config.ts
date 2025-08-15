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

// Configuration object
export const config: AppConfig = {
	apiBaseUrl: getEnvVar('VITE_API_BASE_URL', 'http://localhost:8080/api/v1'),
	wsUrl: getEnvVar('VITE_WS_URL', 'ws://localhost:8080/ws'),
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
