// Configuration file for environment variables and app settings
// This centralizes all configuration and provides fallbacks

interface AppConfig {
	apiBaseUrl: string;
	wsUrl: string;
	appName: string;
	appVersion: string;
	environment: 'development' | 'production' | 'staging';
}

// Get environment variables with fallbacks - SSR compatible
const getEnvVar = (key: string, fallback: string): string => {
	// Try import.meta.env first (client-side and build-time)
	if (typeof import.meta !== 'undefined' && import.meta.env && import.meta.env[key]) {
		return import.meta.env[key];
	}
	
	// Try process.env (server-side and build-time)
	if (typeof process !== 'undefined' && process.env && process.env[key]) {
		return process.env[key];
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

// Validate required environment variables - SSR safe
const validateConfig = () => {
	const apiBaseUrl = getEnvVar('VITE_API_BASE_URL', '');
	if (!apiBaseUrl) {
		// During SSR build, we might not have access to env vars
		// Check if we're in a build context
		const isBuild = typeof process !== 'undefined' && process.env.NODE_ENV === 'production';
		
		if (isBuild) {
			console.warn('⚠️ VITE_API_BASE_URL not available during build - using placeholder');
			return 'https://placeholder-api.example.com/api/v1';
		} else {
			console.error('❌ VITE_API_BASE_URL environment variable is required but not set');
			throw new Error('Missing required environment variable: VITE_API_BASE_URL');
		}
	}
	return apiBaseUrl;
};

// Configuration object - relies on environment variables
export const config: AppConfig = {
	// Use environment variable - with build-time fallback
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
