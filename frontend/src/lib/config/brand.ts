/**
 * Brand Configuration
 * 
 * Central configuration for brand identity, colors, features, and contact info.
 * All values are loaded from environment variables to support multi-brand deployments.
 * 
 * Usage:
 *   import { brand } from '$lib/config/brand';
 *   <h1>{brand.name}</h1>
 *   <div style="--primary: {brand.colors.primary}">
 */

import { env } from '$env/dynamic/public';

export interface BrandConfig {
	// Brand Identity
	name: string;
	tagline: string;
	legalName: string;
	domain: string;
	
	// Visual Branding
	logo: {
		light: string;
		dark: string;
		favicon: string;
		ogImage: string;
	};
	
	colors: {
		primary: string;
		secondary: string;
		accent: string;
		gradient: {
			primary: string;
			secondary: string;
		};
	};
	
	// Features (what's enabled for this brand)
	features: {
		enableAdvertisements: boolean;
		enableLiveStreaming: boolean;
		enableSubscriptions: boolean;
		enableComments: boolean;
		enableArticles: boolean;
		enableEvents: boolean;
	};
	
	// Contact Information
	contact: {
		supportEmail: string;
		businessEmail: string;
		advertisingEmail: string;
		phone?: string;
		address?: string;
	};
	
	// Social Media
	social: {
		twitter?: string;
		facebook?: string;
		instagram?: string;
		youtube?: string;
		linkedin?: string;
	};
	
	// LocalStorage Keys (brand-specific to avoid conflicts)
	storage: {
		authData: string;
		userData: string;
		secureStorage: string;
		authTokens: string;
		tokenInfo: string;
	};
	
	// URLs
	urls: {
		website: string;
		api: string;
		cdn: string;
	};
}

/**
 * Brand Configuration - Loaded from Environment Variables
 */
export const brand: BrandConfig = {
	// Brand Identity
	name: env.PUBLIC_BRAND_NAME || 'BOME',
	tagline: env.PUBLIC_BRAND_TAGLINE || 'Book of Mormon Evidence | Modern Streaming Platform',
	legalName: env.PUBLIC_BRAND_LEGAL_NAME || 'Book of Mormon Evidence',
	domain: env.PUBLIC_BRAND_DOMAIN || 'bookofmormonevidence.org',
	
	// Visual Branding
	logo: {
		light: env.PUBLIC_LOGO_LIGHT || '/brands/bome/logo-light.svg',
		dark: env.PUBLIC_LOGO_DARK || '/brands/bome/logo-dark.svg',
		favicon: env.PUBLIC_FAVICON || '/favicon.ico',
		ogImage: env.PUBLIC_OG_IMAGE || '/og-image.png',
	},
	
	colors: {
		primary: env.PUBLIC_COLOR_PRIMARY || '#6366f1',
		secondary: env.PUBLIC_COLOR_SECONDARY || '#8b5cf6',
		accent: env.PUBLIC_COLOR_ACCENT || '#ec4899',
		gradient: {
			primary: env.PUBLIC_GRADIENT_PRIMARY || 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
			secondary: env.PUBLIC_GRADIENT_SECONDARY || 'linear-gradient(135deg, #8b5cf6 0%, #7c3aed 100%)',
		},
	},
	
	// Features
	features: {
		enableAdvertisements: env.PUBLIC_ENABLE_ADVERTISEMENTS !== 'false',
		enableLiveStreaming: env.PUBLIC_ENABLE_LIVE_STREAMING === 'true',
		enableSubscriptions: env.PUBLIC_ENABLE_SUBSCRIPTIONS !== 'false',
		enableComments: env.PUBLIC_ENABLE_COMMENTS !== 'false',
		enableArticles: env.PUBLIC_ENABLE_ARTICLES !== 'false',
		enableEvents: env.PUBLIC_ENABLE_EVENTS !== 'false',
	},
	
	// Contact Information
	contact: {
		supportEmail: env.PUBLIC_CONTACT_SUPPORT || 'support@bome.org',
		businessEmail: env.PUBLIC_CONTACT_BUSINESS || 'business@bome.org',
		advertisingEmail: env.PUBLIC_CONTACT_ADVERTISING || 'advertising@bome.org',
		phone: env.PUBLIC_CONTACT_PHONE,
		address: env.PUBLIC_CONTACT_ADDRESS,
	},
	
	// Social Media
	social: {
		twitter: env.PUBLIC_SOCIAL_TWITTER,
		facebook: env.PUBLIC_SOCIAL_FACEBOOK,
		instagram: env.PUBLIC_SOCIAL_INSTAGRAM,
		youtube: env.PUBLIC_SOCIAL_YOUTUBE,
		linkedin: env.PUBLIC_SOCIAL_LINKEDIN,
	},
	
	// LocalStorage Keys (brand-specific)
	storage: {
		authData: env.PUBLIC_STORAGE_PREFIX ? `${env.PUBLIC_STORAGE_PREFIX}_auth_data` : 'bome_auth_data',
		userData: env.PUBLIC_STORAGE_PREFIX ? `${env.PUBLIC_STORAGE_PREFIX}_user_data` : 'bome_user_data',
		secureStorage: env.PUBLIC_STORAGE_PREFIX ? `${env.PUBLIC_STORAGE_PREFIX}_secure_storage` : 'bome_secure_storage',
		authTokens: env.PUBLIC_STORAGE_PREFIX ? `${env.PUBLIC_STORAGE_PREFIX}_auth_tokens` : 'bome_auth_tokens',
		tokenInfo: env.PUBLIC_STORAGE_PREFIX ? `${env.PUBLIC_STORAGE_PREFIX}_token_info` : 'bome_token_info',
	},
	
	// URLs
	urls: {
		website: env.PUBLIC_WEBSITE_URL || 'https://bookofmormonevidence.org',
		api: env.PUBLIC_API_URL || 'https://watch.bookofmormonevidence.org/bome-backend/api/v1',
		cdn: env.PUBLIC_CDN_URL || 'https://cdn.bookofmormonevidence.org',
	},
};

/**
 * Helper function to get brand-specific localStorage key
 */
export function getBrandStorageKey(key: keyof BrandConfig['storage']): string {
	return brand.storage[key];
}

/**
 * Helper function to check if a feature is enabled
 */
export function isFeatureEnabled(feature: keyof BrandConfig['features']): boolean {
	return brand.features[feature];
}

/**
 * Helper function to get brand color
 */
export function getBrandColor(color: keyof BrandConfig['colors']): string {
	return brand.colors[color];
}

/**
 * Helper function to get contact email
 */
export function getContactEmail(type: keyof BrandConfig['contact']): string | undefined {
	return brand.contact[type];
}

/**
 * Export for debugging in console
 */
if (typeof window !== 'undefined') {
	(window as any).__BRAND_CONFIG__ = brand;
	console.log('🎨 Brand Config Loaded:', brand.name);
}

