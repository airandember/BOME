// Frontend Security Enhancement System
import { browser } from '$app/environment';

// Rate limiting store
interface RateLimitEntry {
	count: number;
	resetTime: number;
	blocked: boolean;
}

class RateLimiter {
	private limits = new Map<string, RateLimitEntry>();
	private defaultLimit = 100; // requests per window
	private defaultWindow = 60 * 1000; // 1 minute

	// Check if action is rate limited
	isRateLimited(key: string, limit?: number, window?: number): boolean {
		if (!browser) return false;

		const now = Date.now();
		const currentLimit = limit || this.defaultLimit;
		const currentWindow = window || this.defaultWindow;
		
		const entry = this.limits.get(key);
		
		if (!entry) {
			// First request
			this.limits.set(key, {
				count: 1,
				resetTime: now + currentWindow,
				blocked: false
			});
			return false;
		}

		// Reset if window has passed
		if (now > entry.resetTime) {
			entry.count = 1;
			entry.resetTime = now + currentWindow;
			entry.blocked = false;
			return false;
		}

		// Increment counter
		entry.count++;

		// Check if limit exceeded
		if (entry.count > currentLimit) {
			entry.blocked = true;
			return true;
		}

		return false;
	}

	// Get remaining requests
	getRemainingRequests(key: string, limit?: number): number {
		const entry = this.limits.get(key);
		if (!entry) return limit || this.defaultLimit;
		
		const currentLimit = limit || this.defaultLimit;
		return Math.max(0, currentLimit - entry.count);
	}

	// Get time until reset
	getResetTime(key: string): number {
		const entry = this.limits.get(key);
		if (!entry) return 0;
		
		return Math.max(0, entry.resetTime - Date.now());
	}

	// Clear rate limit for key
	clearLimit(key: string): void {
		this.limits.delete(key);
	}

	// Clear all rate limits
	clearAll(): void {
		this.limits.clear();
	}
}

export const rateLimiter = new RateLimiter();

// Input validation and sanitization
export class InputValidator {
	// Email validation
	static isValidEmail(email: string): boolean {
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		return emailRegex.test(email.trim());
	}

	// Password strength validation
	static validatePassword(password: string): {
		isValid: boolean;
		score: number;
		feedback: string[];
	} {
		const feedback: string[] = [];
		let score = 0;

		if (password.length < 8) {
			feedback.push('Password must be at least 8 characters long');
		} else {
			score += 1;
		}

		if (!/[a-z]/.test(password)) {
			feedback.push('Password must contain at least one lowercase letter');
		} else {
			score += 1;
		}

		if (!/[A-Z]/.test(password)) {
			feedback.push('Password must contain at least one uppercase letter');
		} else {
			score += 1;
		}

		if (!/\d/.test(password)) {
			feedback.push('Password must contain at least one number');
		} else {
			score += 1;
		}

		if (!/[!@#$%^&*(),.?":{}|<>]/.test(password)) {
			feedback.push('Password must contain at least one special character');
		} else {
			score += 1;
		}

		return {
			isValid: score >= 4,
			score,
			feedback
		};
	}

	// URL validation
	static isValidURL(url: string): boolean {
		try {
			new URL(url);
			return true;
		} catch {
			return false;
		}
	}

	// Phone number validation (basic)
	static isValidPhone(phone: string): boolean {
		const phoneRegex = /^\+?[\d\s\-\(\)]{10,}$/;
		return phoneRegex.test(phone.trim());
	}

	// Sanitize HTML input (basic XSS prevention)
	static sanitizeHTML(input: string): string {
		const div = document.createElement('div');
		div.textContent = input;
		return div.innerHTML;
	}

	// Sanitize for SQL-like injection attempts
	static sanitizeQuery(query: string): string {
		return query
			.replace(/['"\\;]/g, '') // Remove quotes and semicolons
			.replace(/\b(SELECT|INSERT|UPDATE|DELETE|DROP|CREATE|ALTER|EXEC|UNION|SCRIPT)\b/gi, '') // Remove SQL keywords
			.trim();
	}

	// Validate file upload
	static validateFile(file: File, options: {
		maxSize?: number; // in bytes
		allowedTypes?: string[];
		allowedExtensions?: string[];
	} = {}): {
		isValid: boolean;
		errors: string[];
	} {
		const errors: string[] = [];
		const maxSize = options.maxSize || 10 * 1024 * 1024; // 10MB default
		const allowedTypes = options.allowedTypes || ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
		const allowedExtensions = options.allowedExtensions || ['.jpg', '.jpeg', '.png', '.gif', '.webp'];

		// Check file size
		if (file.size > maxSize) {
			errors.push(`File size must be less than ${Math.round(maxSize / 1024 / 1024)}MB`);
		}

		// Check file type
		if (!allowedTypes.includes(file.type)) {
			errors.push(`File type ${file.type} is not allowed`);
		}

		// Check file extension
		const extension = '.' + file.name.split('.').pop()?.toLowerCase();
		if (!allowedExtensions.includes(extension)) {
			errors.push(`File extension ${extension} is not allowed`);
		}

		return {
			isValid: errors.length === 0,
			errors
		};
	}

	// Validate JSON input
	static validateJSON(jsonString: string): {
		isValid: boolean;
		data?: any;
		error?: string;
	} {
		try {
			const data = JSON.parse(jsonString);
			return { isValid: true, data };
		} catch (error) {
			return { 
				isValid: false, 
				error: error instanceof Error ? error.message : 'Invalid JSON' 
			};
		}
	}

	// Generic field validation
	static validateField(value: any, rules: {
		required?: boolean;
		minLength?: number;
		maxLength?: number;
		pattern?: RegExp;
		type?: 'string' | 'number' | 'email' | 'url' | 'phone';
		custom?: (value: any) => boolean | string;
	}): {
		isValid: boolean;
		errors: string[];
	} {
		const errors: string[] = [];

		// Required check
		if (rules.required && (!value || (typeof value === 'string' && value.trim() === ''))) {
			errors.push('This field is required');
			return { isValid: false, errors };
		}

		// Skip other validations if value is empty and not required
		if (!value && !rules.required) {
			return { isValid: true, errors: [] };
		}

		const stringValue = String(value);

		// Length checks
		if (rules.minLength && stringValue.length < rules.minLength) {
			errors.push(`Must be at least ${rules.minLength} characters long`);
		}

		if (rules.maxLength && stringValue.length > rules.maxLength) {
			errors.push(`Must be no more than ${rules.maxLength} characters long`);
		}

		// Pattern check
		if (rules.pattern && !rules.pattern.test(stringValue)) {
			errors.push('Invalid format');
		}

		// Type-specific validation
		switch (rules.type) {
			case 'email':
				if (!this.isValidEmail(stringValue)) {
					errors.push('Invalid email address');
				}
				break;
			case 'url':
				if (!this.isValidURL(stringValue)) {
					errors.push('Invalid URL');
				}
				break;
			case 'phone':
				if (!this.isValidPhone(stringValue)) {
					errors.push('Invalid phone number');
				}
				break;
			case 'number':
				if (isNaN(Number(value))) {
					errors.push('Must be a valid number');
				}
				break;
		}

		// Custom validation
		if (rules.custom) {
			const customResult = rules.custom(value);
			if (customResult !== true) {
				errors.push(typeof customResult === 'string' ? customResult : 'Custom validation failed');
			}
		}

		return {
			isValid: errors.length === 0,
			errors
		};
	}
}

// CSRF Protection
export class CSRFProtection {
	private static token: string | null = null;

	// Generate CSRF token
	static generateToken(): string {
		const array = new Uint8Array(32);
		crypto.getRandomValues(array);
		const token = Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
		
		if (browser) {
			localStorage.setItem('csrf_token', token);
			this.token = token;
		}
		
		return token;
	}

	// Get current token
	static getToken(): string | null {
		if (!browser) return null;
		
		if (!this.token) {
			this.token = localStorage.getItem('csrf_token');
		}
		
		return this.token;
	}

	// Validate token
	static validateToken(token: string): boolean {
		const storedToken = this.getToken();
		return storedToken === token;
	}

	// Add CSRF token to headers
	static addToHeaders(headers: Record<string, string> = {}): Record<string, string> {
		const token = this.getToken();
		if (token) {
			headers['X-CSRF-Token'] = token;
		}
		return headers;
	}
}

// Content Security Policy helpers
export class CSPHelper {
	// Generate nonce for inline scripts
	static generateNonce(): string {
		const array = new Uint8Array(16);
		crypto.getRandomValues(array);
		return btoa(String.fromCharCode(...array));
	}

	// Validate external URL against CSP
	static isAllowedURL(url: string, allowedDomains: string[] = []): boolean {
		try {
			const urlObj = new URL(url);
			const domain = urlObj.hostname;
			
			// Default allowed domains
			const defaultAllowed = [
				'bome.org',
				'api.bome.org',
				'cdn.bome.org',
				'bunnycdn.com',
				'b-cdn.net'
			];
			
			const allAllowed = [...defaultAllowed, ...allowedDomains];
			
			return allAllowed.some(allowed => 
				domain === allowed || domain.endsWith('.' + allowed)
			);
		} catch {
			return false;
		}
	}
}

// Security monitoring and logging
export class SecurityMonitor {
	private static violations: Array<{
		type: string;
		details: any;
		timestamp: number;
		userAgent: string;
		url: string;
	}> = [];

	// Log security violation
	static logViolation(type: string, details: any = {}): void {
		if (!browser) return;

		const violation = {
			type,
			details,
			timestamp: Date.now(),
			userAgent: navigator.userAgent,
			url: window.location.href
		};

		this.violations.push(violation);
		
		// Keep only last 100 violations
		if (this.violations.length > 100) {
			this.violations.shift();
		}

		// Log to console in development
		if (import.meta.env.DEV) {
			console.warn('Security violation:', violation);
		}

		// In production, you might want to send this to your backend
		// this.reportViolation(violation);
	}

	// Get violation history
	static getViolations(): typeof SecurityMonitor.violations {
		return [...this.violations];
	}

	// Clear violation history
	static clearViolations(): void {
		this.violations = [];
	}

	// Report violation to backend (placeholder)
	private static async reportViolation(violation: any): Promise<void> {
		try {
			// In a real implementation, send to your security endpoint
			// await fetch('/api/security/violation', {
			//     method: 'POST',
			//     headers: { 'Content-Type': 'application/json' },
			//     body: JSON.stringify(violation)
			// });
		} catch (error) {
			console.error('Failed to report security violation:', error);
		}
	}
}

// Secure form submission helper
export class SecureForm {
	static async submitForm(
		url: string,
		data: Record<string, any>,
		options: {
			method?: 'POST' | 'PUT' | 'PATCH';
			rateLimitKey?: string;
			rateLimitWindow?: number;
			rateLimitCount?: number;
			validateCSRF?: boolean;
		} = {}
	): Promise<{
		success: boolean;
		data?: any;
		error?: string;
		rateLimited?: boolean;
	}> {
		const {
			method = 'POST',
			rateLimitKey = url,
			rateLimitWindow = 60000, // 1 minute
			rateLimitCount = 10,
			validateCSRF = true
		} = options;

		// Check rate limiting
		if (rateLimiter.isRateLimited(rateLimitKey, rateLimitCount, rateLimitWindow)) {
			SecurityMonitor.logViolation('rate_limit_exceeded', {
				url,
				rateLimitKey,
				remainingTime: rateLimiter.getResetTime(rateLimitKey)
			});
			
			return {
				success: false,
				error: 'Rate limit exceeded. Please try again later.',
				rateLimited: true
			};
		}

		// Prepare headers
		const headers: Record<string, string> = {
			'Content-Type': 'application/json'
		};

		// Add CSRF protection
		if (validateCSRF) {
			CSRFProtection.addToHeaders(headers);
		}

		// Sanitize data
		const sanitizedData: Record<string, any> = {};
		for (const [key, value] of Object.entries(data)) {
			if (typeof value === 'string') {
				sanitizedData[key] = InputValidator.sanitizeHTML(value);
			} else {
				sanitizedData[key] = value;
			}
		}

		try {
			const response = await fetch(url, {
				method,
				headers,
				body: JSON.stringify(sanitizedData)
			});

			if (!response.ok) {
				throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			}

			const responseData = await response.json();
			return {
				success: true,
				data: responseData
			};
		} catch (error) {
			SecurityMonitor.logViolation('form_submission_error', {
				url,
				error: error instanceof Error ? error.message : 'Unknown error'
			});

			return {
				success: false,
				error: error instanceof Error ? error.message : 'Form submission failed'
			};
		}
	}
}

// Initialize security features
export function initializeSecurity(): void {
	if (!browser) return;

	// Generate initial CSRF token
	if (!CSRFProtection.getToken()) {
		CSRFProtection.generateToken();
	}

	// Set up CSP violation reporting
	document.addEventListener('securitypolicyviolation', (event) => {
		SecurityMonitor.logViolation('csp_violation', {
			blockedURI: event.blockedURI,
			violatedDirective: event.violatedDirective,
			originalPolicy: event.originalPolicy
		});
	});

	// Monitor for suspicious activity
	let clickCount = 0;
	let clickTimer: number;

	document.addEventListener('click', () => {
		clickCount++;
		
		if (clickTimer) {
			clearTimeout(clickTimer);
		}
		
		clickTimer = window.setTimeout(() => {
			if (clickCount > 50) { // Suspicious rapid clicking
				SecurityMonitor.logViolation('suspicious_activity', {
					type: 'rapid_clicking',
					count: clickCount
				});
			}
			clickCount = 0;
		}, 1000);
	});

	console.log('🔒 Security features initialized');
} 