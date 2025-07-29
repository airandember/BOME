import { api } from '$lib/api';

export interface SubscriptionPlan {
	id: string;
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'month' | 'year' | 'week' | 'day';
	interval_count: number;
	stripe_price_id: string | null;
	features: string[];
	is_active: boolean;
	promotion_end_date: string | null;
	promotion_start_date: string | null;
	plan_change_history: string[]; // Renamed from promotion_history
	promotion_metadata: Record<string, any>; // New field for promotion analytics
	sub_type: string; // stnd = standard plan, prmo = promotional plan
	created_at: string;
	updated_at: string;
	is_deleted: boolean;
}

export interface CreateSubscriptionPlanData {
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'month' | 'year' | 'week' | 'day';
	interval_count: number;
	stripe_price_id: string;
	features: string[];
	is_active: boolean;
	promotion_start_date: string | null;
	promotion_end_date: string | null;
	sub_type: 'stnd' | 'prmo';
}

export interface UpdateSubscriptionPlanData extends Partial<CreateSubscriptionPlanData> {
	id: string;
	sub_type?: 'stnd' | 'prmo'; // Override to make sub_type optional for updates
}

export class StreamingSubscriptionService {
	private static readonly BASE_PATH = '/admin/subscription-plans';

	/**
	 * Get all subscription plans
	 */
	static async getAll(): Promise<SubscriptionPlan[]> {
		try {
			console.log('StreamingSubscriptionService: Making request to:', this.BASE_PATH);
			const response = await api.get(this.BASE_PATH);

			console.log('StreamingSubscriptionService: Response received:', response);
			return response.data as SubscriptionPlan[];
		} catch (error) {	
			console.error('StreamingSubscriptionService: Error fetching subscription plans:', error);
			throw error;
		}
	}

	/**
	 * Get a single subscription plan by ID
	 */
	static async getById(id: string): Promise<SubscriptionPlan> {
		try {
			const response = await api.get(`${this.BASE_PATH}/${id}`);
			return response.data as SubscriptionPlan;
		} catch (error) {
			console.error('Error fetching subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Create a new subscription plan
	 */
	static async create(data: CreateSubscriptionPlanData): Promise<SubscriptionPlan> {
		try {
			const response = await api.post(this.BASE_PATH, data);
			return response.data as SubscriptionPlan;
		} catch (error) {
			console.error('Error creating subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Update an existing subscription plan
	 */
	static async update(data: UpdateSubscriptionPlanData): Promise<SubscriptionPlan> {
		try {
			const response = await api.put(`${this.BASE_PATH}/${data.id}`, data);
			return response.data as SubscriptionPlan;
		} catch (error) {
			console.error('Error updating subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Soft delete a subscription plan (creates a timestamp in DB "deleted_at" column and "is_deleted" boolean to true)
	 */
	static async delete(id: string): Promise<void> {
		try {
			await api.delete(`${this.BASE_PATH}/${id}`);
		} catch (error) {
			console.error('Error deleting subscription plan:', error);
			throw error;
		}
	}

	/**
	 * Toggle subscription plan active status
	 */
	static async toggleStatus(id: string, isActive: boolean): Promise<SubscriptionPlan> {
		try {
			const response = await api.put(`${this.BASE_PATH}/${id}/status`, { is_active: isActive });
			return response.data as SubscriptionPlan;
		} catch (error) {
			console.error('Error toggling subscription plan status:', error);
			throw error;
		}
	}

	/**
	 * Toggle promotion status (changes sub_type between 'stnd' and 'prmo')
	 */
	static async togglePromotion(id: string, isPromoted: boolean): Promise<SubscriptionPlan> {
		try {
			const subType = isPromoted ? 'prmo' : 'stnd';
			const response = await api.put(`${this.BASE_PATH}/${id}/promotion`, { sub_type: subType });
			return response.data as SubscriptionPlan;
		} catch (error) {
			console.error('Error toggling promotion status:', error);
			throw error;
		}
	}
} 