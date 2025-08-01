import { apiRequest } from '$lib/auth';

// Types
export interface SubscriptionOffer {
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
	offer_history: Record<string, any>[];
}

export interface CreateSubscriptionOfferData {
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
}

export interface UpdateSubscriptionOfferData {
	id: number;
	plan_id?: number;
	item_id?: string;
	off_discount_type?: string;
	off_discount_value?: number;
	offer_start_date?: string;
	off_end_date?: string;
	is_active?: boolean;
	off_description?: string;
	off_name?: string;
	off_code?: string;
	off_max_uses?: number;
	off_current_uses?: number;
	off_terms_conditions?: string;
	off_target?: string;
	off_priority?: number;
	off_auto_apply?: boolean;
}

// Service class
export class SubscriptionOfferService {
	private static readonly BASE_PATH = '/api/v1/admin/subscription-offers/';

	// Get all offers
	static async getAll(): Promise<SubscriptionOffer[]> {
		try {
			const response = await apiRequest(this.BASE_PATH.slice(0, -1)); // Remove trailing slash
			return response as unknown as SubscriptionOffer[];
		} catch (error) {
			console.error('Error getting subscription offers:', error);
			throw error;
		}
	}

	// Get offer by ID
	static async getById(id: string): Promise<SubscriptionOffer> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}`);
			return response as unknown as SubscriptionOffer;
		} catch (error) {
			console.error('Error getting subscription offer:', error);
			throw error;
		}
	}

	// Create new offer
	static async create(data: CreateSubscriptionOfferData): Promise<SubscriptionOffer> {
		try {
			// Get user data from localStorage
			const userData = localStorage.getItem('bome_user_data');
			console.log('SubscriptionOfferService: Raw user data from localStorage:', userData);
			
			const headers: Record<string, string> = {};
			if (userData) {
				headers['X-User-Data'] = userData;
				console.log('SubscriptionOfferService: Setting X-User-Data header:', userData);
			}

			const response = await apiRequest(this.BASE_PATH.slice(0, -1), {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					...headers
				},
				body: JSON.stringify(data)
			});
			return response as unknown as SubscriptionOffer;
		} catch (error) {
			console.error('Error creating subscription offer:', error);
			throw error;
		}
	}

	// Update offer
	static async update(data: UpdateSubscriptionOfferData): Promise<SubscriptionOffer> {
		try {
			// Get user data from localStorage
			const userData = localStorage.getItem('bome_user_data');
			console.log('SubscriptionOfferService: Raw user data from localStorage:', userData);
			
			const headers: Record<string, string> = {};
			if (userData) {
				headers['X-User-Data'] = userData;
				console.log('SubscriptionOfferService: Setting X-User-Data header:', userData);
			}

			const response = await apiRequest(`${this.BASE_PATH}${data.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					...headers
				},
				body: JSON.stringify(data)
			});
			return response as unknown as SubscriptionOffer;
		} catch (error) {
			console.error('Error updating subscription offer:', error);
			throw error;
		}
	}

	// Delete offer
	static async delete(id: string): Promise<void> {
		try {
			// Get user data from localStorage
			const userData = localStorage.getItem('bome_user_data');
			console.log('SubscriptionOfferService: Raw user data from localStorage:', userData);
			
			const headers: Record<string, string> = {};
			if (userData) {
				headers['X-User-Data'] = userData;
				console.log('SubscriptionOfferService: Setting X-User-Data header:', userData);
			}

			await apiRequest(`${this.BASE_PATH}${id}`, {
				method: 'DELETE',
				headers
			});
		} catch (error) {
			console.error('Error deleting subscription offer:', error);
			throw error;
		}
	}

	// Toggle offer status
	static async toggleStatus(id: string, isActive: boolean): Promise<SubscriptionOffer> {
		try {
			// Get user data from localStorage
			const userData = localStorage.getItem('bome_user_data');
			console.log('SubscriptionOfferService: Raw user data from localStorage:', userData);
			
			const headers: Record<string, string> = {};
			if (userData) {
				headers['X-User-Data'] = userData;
				console.log('SubscriptionOfferService: Setting X-User-Data header:', userData);
			}

			const response = await apiRequest(`${this.BASE_PATH}${id}/toggle`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					...headers
				}
			});
			return response as unknown as SubscriptionOffer;
		} catch (error) {
			console.error('Error toggling subscription offer status:', error);
			throw error;
		}
	}

	// Get offer history
	static async getOfferHistory(id: string): Promise<Record<string, any>[]> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}/history`);
			return response as unknown as Record<string, any>[];
		} catch (error) {
			console.error('Error getting offer history:', error);
			throw error;
		}
	}
} 