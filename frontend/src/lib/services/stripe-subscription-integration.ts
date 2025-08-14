import { apiRequest } from '$lib/auth';
import type { SubscriptionPlan } from './streaming-subscriptions';

export interface StripeIntegrationStatus {
	has_stripe_product: boolean;
	has_stripe_price: boolean;
	stripe_product_id?: string;
	stripe_price_id?: string;
	sync_status: 'synced' | 'partial' | 'not_synced';
	last_synced?: string;
}

export interface CreatePlanWithStripeRequest {
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'month' | 'year' | 'week' | 'day';
	interval_count: number;
	features: string[];
	is_active: boolean;
	sub_type: string;
	auto_create_stripe: boolean; // New field
	promotion_start_date?: string;
	promotion_end_date?: string;
}

export class StripeSubscriptionIntegrationService {
	// Create plan with optional Stripe integration
	static async createPlanWithStripe(data: CreatePlanWithStripeRequest): Promise<SubscriptionPlan> {
		const response = await apiRequest('/admin/streaming/subscription-plans/stripe', {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify(data)
		});

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.error || 'Failed to create plan with Stripe integration');
		}

		return response.json();
	}

	// Sync existing plan with Stripe
	static async syncPlanWithStripe(planId: string): Promise<SubscriptionPlan> {
		const response = await apiRequest(`/admin/streaming/subscription-plans/stripe/${planId}/sync`, {
			method: 'POST'
		});

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.error || 'Failed to sync plan with Stripe');
		}

		const result = await response.json();
		return result.plan;
	}

	// Get Stripe integration status for a plan
	static async getStripeStatus(planId: string): Promise<StripeIntegrationStatus> {
		const response = await apiRequest(`/admin/streaming/subscription-plans/stripe/${planId}/stripe-status`);

		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.error || 'Failed to get Stripe status');
		}

		return response.json();
	}

	// Get all Stripe products (from existing Stripe dashboard)
	static async getStripeProducts() {
		const response = await apiRequest('/admin/streaming/stripe/summary');
		if (!response.ok) {
			throw new Error('Failed to get Stripe products');
		}
		
		const data = await response.json();
		return data.summary?.products || [];
	}

	// Check if Stripe is enabled/connected
	static async checkStripeConnection(): Promise<boolean> {
		try {
			const response = await apiRequest('/admin/streaming/stripe/summary');
			if (!response.ok) return false;
			
			const data = await response.json();
			return data.summary?.enabled || false;
		} catch {
			return false;
		}
	}
} 
