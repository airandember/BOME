import { apiRequest } from '$lib/auth';

export interface PublicSubscriptionPlan {
	id: string;
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'month' | 'year' | 'week' | 'day';
	interval_count: number;
	stripe_price_id?: string;
	features: string[];
	is_active: boolean;
	promotion_end_date?: string;
	promotion_start_date?: string;
	created_at: string;
	updated_at: string;
	popular?: boolean;
	sub_type: string; // stnd = standard plan, prmo = promotional plan
}

export interface PublicSubscriptionOffer {
	id: string;
	off_name: string;
	off_description: string;
	off_discount_type: 'percentage' | 'fixed';
	off_discount_value: number;
	off_max_uses?: number;
	off_current_uses: number;
	off_auto_apply: boolean;
	plan_id: number;
	item_id: number;
	offer_start_date?: string;
	off_end_date?: string;
	is_active: boolean;
	off_target?: string;
	off_priority: number;
	created_at: string;
	updated_at: string;
}

export interface SubscriptionData {
	standard_plans: PublicSubscriptionPlan[];
	promotional_plans: PublicSubscriptionPlan[];
	offers: PublicSubscriptionOffer[];
}

// Public service for fetching subscription plans without authentication
export const publicPlansService = {
	// Get all subscription data (plans + offers) in one call
	getAllSubscriptionData: async (): Promise<SubscriptionData> => {
		try {
			const response = await fetch('/api/v1/subscription-plans/all');
			const data = await response.json();
			
			if (data.status === 'success' && data.data) {
				return {
					standard_plans: data.data.standard_plans || [],
					promotional_plans: data.data.promotional_plans || [],
					offers: data.data.offers || []
				};
			}
			
			throw new Error(data.message || 'Failed to fetch subscription data');
		} catch (error) {
			console.error('Error fetching all subscription data:', error);
			throw error;
		}
	},

	// Get all active plans (standard and promotional)
	getAllActivePlans: async (): Promise<PublicSubscriptionPlan[]> => {
		try {
			const [standardResponse, promotionalResponse] = await Promise.allSettled([
				fetch('/api/v1/subscription-plans/active'),
				fetch('/api/v1/subscription-plans/promoted')
			]);

			const allPlans: PublicSubscriptionPlan[] = [];

			// Handle standard plans
			if (standardResponse.status === 'fulfilled') {
				const data = await standardResponse.value.json();
				if (data.data?.subscription_plans) {
					allPlans.push(...data.data.subscription_plans);
				}
			}

			// Handle promotional plans
			if (promotionalResponse.status === 'fulfilled') {
				const data = await promotionalResponse.value.json();
				if (data.data?.subscription_plans) {
					allPlans.push(...data.data.subscription_plans);
				}
			}

			return allPlans;
		} catch (error) {
			console.error('Error fetching public plans:', error);
			throw error;
		}
	},

	// Get only standard plans
	getStandardPlans: async (): Promise<PublicSubscriptionPlan[]> => {
		try {
			const response = await fetch('/api/v1/subscription-plans/active');
			const data = await response.json();
			return data.data?.subscription_plans || [];
		} catch (error) {
			console.error('Error fetching standard plans:', error);
			throw error;
		}
	},

	// Get only promotional plans
	getPromotionalPlans: async (): Promise<PublicSubscriptionPlan[]> => {
		try {
			const response = await fetch('/api/v1/subscription-plans/promoted');
			const data = await response.json();
			return data.data?.subscription_plans || [];
		} catch (error) {
			console.error('Error fetching promotional plans:', error);
			throw error;
		}
	}
}; 