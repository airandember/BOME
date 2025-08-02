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

// Public service for fetching subscription plans without authentication
export const publicPlansService = {
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