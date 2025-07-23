import { api } from '$lib/api';

export interface SubscriptionPlan {
	id: string;
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'monthly' | 'yearly' | 'weekly' | 'daily';
	interval_count: number;
	features: string[];
	is_active: boolean;
	is_promoted: boolean;
	promotion_end_date: string | null;
	sort_order: number;
	created_at: string;
	updated_at: string;
}

export interface CreateSubscriptionPlanData {
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'monthly' | 'yearly' | 'weekly' | 'daily';
	interval_count: number;
	features: string[];
	is_active: boolean;
	is_promoted: boolean;
	promotion_end_date: string | null;
	sort_order: number;
}

export interface UpdateSubscriptionPlanData extends Partial<CreateSubscriptionPlanData> {
	id: string;
}

export interface SubscriptionPlansResponse {
	subscription_plans: SubscriptionPlan[];
	total: number;
	page: number;
	limit: number;
}

export class StreamingSubscriptionService {
	/**
	 * Get all subscription plans with optional pagination and filters
	 */
	static async getSubscriptionPlans(params?: {
		page?: number;
		limit?: number;
		search?: string;
		status?: 'active' | 'inactive' | 'all';
	}): Promise<SubscriptionPlansResponse> {
		try {
			const queryParams = new URLSearchParams();
			if (params?.page) queryParams.append('page', params.page.toString());
			if (params?.limit) queryParams.append('limit', params.limit.toString());
			if (params?.search) queryParams.append('search', params.search);
			if (params?.status) queryParams.append('status', params.status);

			const response = await api.get(`/api/admin/subscription-plans?${queryParams}`);
			
			if (response.data) {
				return response.data as SubscriptionPlansResponse;
			} else {
				throw new Error(response.error || 'Failed to load subscription plans');
			}
		} catch (error) {
			console.error('Error fetching subscription plans:', error);
			throw error;
		}
	}

	/**
	 * Get a single subscription plan by ID
	 */
	static async getSubscriptionPlan(id: string): Promise<SubscriptionPlan> {
		try {
			const response = await api.get(`/api/admin/subscription-plans/${id}`);
			
			if (response.data) {
				return (response.data as { subscription_plan: SubscriptionPlan }).subscription_plan;
			} else {
				throw new Error(response.error || 'Failed to load subscription plan');
			}
		} catch (error) {
			console.error('Error fetching subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Create a new subscription plan
	 */
	static async createSubscriptionPlan(data: CreateSubscriptionPlanData): Promise<SubscriptionPlan> {
		try {
			const response = await api.post('/api/admin/subscription-plans', data);
			
			if (response.data) {
				return (response.data as { subscription_plan: SubscriptionPlan }).subscription_plan;
			} else {
				throw new Error(response.error || 'Failed to create subscription plan');
			}
		} catch (error) {
			console.error('Error creating subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Update an existing subscription plan
	 */
	static async updateSubscriptionPlan(data: UpdateSubscriptionPlanData): Promise<SubscriptionPlan> {
		try {
			const { id, ...updateData } = data;
			const response = await api.put(`/api/admin/subscription-plans/${id}`, updateData);
			
			if (response.data) {
				return (response.data as { subscription_plan: SubscriptionPlan }).subscription_plan;
			} else {
				throw new Error(response.error || 'Failed to update subscription plan');
			}
		} catch (error) {
			console.error('Error updating subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Delete a subscription plan
	 */
	static async deleteSubscriptionPlan(id: string): Promise<void> {
		try {
			const response = await api.delete(`/api/admin/subscription-plans/${id}`);
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to delete subscription plan');
			}
		} catch (error) {
			console.error('Error deleting subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Toggle subscription plan active status
	 */
	static async toggleSubscriptionPlanStatus(id: string, isActive: boolean): Promise<SubscriptionPlan> {
		try {
			const response = await api.put(`/api/admin/subscription-plans/${id}/status`, {
				is_active: isActive
			});
			
			if (response.data) {
				return (response.data as { subscription_plan: SubscriptionPlan }).subscription_plan;
			} else {
				throw new Error(response.error || 'Failed to update subscription plan status');
			}
		} catch (error) {
			console.error('Error updating subscription plan status:', error);
			throw error;
		}
	}

	/**
	 * Update subscription plan promotion status
	 */
	static async updatePromotionStatus(
		id: string, 
		isPromoted: boolean, 
		promotionEndDate?: string
	): Promise<SubscriptionPlan> {
		try {
			const response = await api.put(`/api/admin/subscription-plans/${id}/promotion`, {
				is_promoted: isPromoted,
				promotion_end_date: promotionEndDate
			});
			
			if (response.data) {
				return (response.data as { subscription_plan: SubscriptionPlan }).subscription_plan;
			} else {
				throw new Error(response.error || 'Failed to update promotion status');
			}
		} catch (error) {
			console.error('Error updating promotion status:', error);
			throw error;
		}
	}

	/**
	 * Reorder subscription plans
	 */
	static async reorderSubscriptionPlans(planIds: string[]): Promise<void> {
		try {
			const response = await api.post('/api/admin/subscription-plans/reorder', {
				plan_ids: planIds
			});
			
			if (!response.data) {
				throw new Error(response.error || 'Failed to reorder subscription plans');
			}
		} catch (error) {
			console.error('Error reordering subscription plans:', error);
			throw error;
		}
	}

	/**
	 * Get subscription plan statistics
	 */
	static async getSubscriptionPlanStats(): Promise<{
		total_plans: number;
		active_plans: number;
		promoted_plans: number;
		total_subscribers: number;
		revenue_this_month: number;
	}> {
		try {
			const response = await api.get('/api/admin/subscription-plans/stats');
			
			if (response.data) {
				return response.data as {
					total_plans: number;
					active_plans: number;
					promoted_plans: number;
					total_subscribers: number;
					revenue_this_month: number;
				};
			} else {
				throw new Error(response.error || 'Failed to load subscription plan stats');
			}
		} catch (error) {
			console.error('Error fetching subscription plan stats:', error);
			throw error;
		}
	}
} 