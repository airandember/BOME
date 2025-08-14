import { apiRequest } from '$lib/auth';

// Types for Stripe offers integration
export interface StripeOfferIntegrationStatus {
	has_stripe_coupon: boolean;
	has_stripe_promotion_code: boolean;
	sync_status: 'synced' | 'partial' | 'not_synced';
	stripe_coupon_id?: string;
	stripe_promotion_code_id?: string;
	last_synced?: string;
}

export interface CreateOfferWithStripeRequest {
	plan_id: number;
	item_id?: string;
	off_discount_type: string;
	off_discount_value: number;
	offer_start_date?: string;
	off_end_date?: string;
	is_active: boolean;
	off_description?: string;
	off_name: string;
	off_code?: string;
	off_max_uses?: number;
	off_current_uses: number;
	off_terms_conditions?: string;
	off_target?: string;
	off_priority: number;
	off_auto_apply: boolean;
	auto_create_stripe: boolean;
}

export interface StripeOfferResponse {
	id: number;
	plan_id: number;
	item_id?: string;
	off_discount_type: string;
	off_discount_value: number;
	offer_start_date?: string;
	off_end_date?: string;
	is_active: boolean;
	off_description?: string;
	off_created_at: string;
	off_updated_at: string;
	off_name: string;
	off_code?: string;
	off_max_uses?: number;
	off_current_uses: number;
	off_terms_conditions?: string;
	off_target?: string;
	off_priority: number;
	off_auto_apply: boolean;
	stripe_coupon_id?: string;
	stripe_promotion_code_id?: string;
	offer_history: Record<string, any>[];
}

export class StripeOfferIntegrationService {
	private static baseUrl = '/admin/streaming/subscription-offers/stripe';

	/**
	 * Create a new offer with optional Stripe integration
	 */
	static async createOfferWithStripe(offerData: CreateOfferWithStripeRequest): Promise<StripeOfferResponse> {
		const response = await apiRequest(`${this.baseUrl}/`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify(offerData),
		});

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to create offer with Stripe');
		}

		return response.json();
	}

	/**
	 * Sync an existing offer with Stripe
	 */
	static async syncOfferWithStripe(offerId: string): Promise<{ message: string; offer: StripeOfferResponse }> {
		const response = await apiRequest(`${this.baseUrl}/${offerId}/sync`, {
			method: 'POST',
		});

		if (!response.ok) {
			const error = await response.json();
			
			// Extract more detailed error information
			let errorMessage = 'Failed to sync offer with Stripe';
			
			if (error.error) {
				errorMessage = error.error;
			} else if (error.message) {
				errorMessage = error.message;
			} else if (error.detail) {
				errorMessage = error.detail;
			}
			
			// Create a more informative error object
			const enhancedError = new Error(errorMessage);
			(enhancedError as any).stripeError = error;
			(enhancedError as any).statusCode = response.status;
			
			throw enhancedError;
		}

		return response.json();
	}

	/**
	 * Get Stripe integration status for an offer
	 */
	static async getStripeStatus(offerId: string): Promise<StripeOfferIntegrationStatus> {
		const response = await apiRequest(`${this.baseUrl}/${offerId}/stripe-status`);

		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || 'Failed to get Stripe status');
		}

		return response.json();
	}

	/**
	 * Check if Stripe connection is working
	 */
	static async checkStripeConnection(): Promise<{ enabled: boolean; environment?: string }> {
		try {
			const response = await apiRequest('/admin/streaming/stripe/summary');
			
			if (!response.ok) {
				return { enabled: false };
			}

			const data = await response.json();
			return {
				enabled: data.enabled || false,
				environment: data.environment
			};
		} catch (error) {
			console.error('Error checking Stripe connection:', error);
			return { enabled: false };
		}
	}
} 
