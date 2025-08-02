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
	item_id?: number;
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
	private static readonly BASE_PATH = '/admin/subscription-offers/';

	// Get all offers
	static async getAll(): Promise<SubscriptionOffer[]> {
		try {
			console.log('SubscriptionOfferService: Starting getAll request');
			console.log('SubscriptionOfferService: Making request to:', this.BASE_PATH);
			
			const response = await apiRequest(this.BASE_PATH);

			console.log('SubscriptionOfferService: Response received:', {
				status: response.status,
				statusText: response.statusText,
				ok: response.ok,
				url: response.url
			});

			if (!response.ok) {
				const errorText = await response.text();
				console.error('SubscriptionOfferService: Error response body:', errorText);
				throw new Error(`Failed to fetch subscription offers: ${response.status} - ${errorText}`);
			}

			const data = await response.json();
			console.log('SubscriptionOfferService: Successfully retrieved offers:', data);
			return data as SubscriptionOffer[];
		} catch (error) {
			console.error('SubscriptionOfferService: Error getting subscription offers:', error);
			throw error;
		}
	}

	// Get offer by ID
	static async getById(id: string): Promise<SubscriptionOffer> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch subscription offer: ${response.status}`);
			}

			const data = await response.json();
			return data as SubscriptionOffer;
		} catch (error) {
			console.error('Error getting subscription offer:', error);
			throw error;
		}
	}

	// Create new offer
	static async create(data: CreateSubscriptionOfferData): Promise<SubscriptionOffer> {
		try {
			console.log('SubscriptionOfferService: Starting create offer request');
			console.log('SubscriptionOfferService: Request data:', data);
			console.log('SubscriptionOfferService: Target URL:', this.BASE_PATH);
			
			// Get user data from localStorage
			const userData = localStorage.getItem('bome_user_data');
			console.log('SubscriptionOfferService: Raw user data from localStorage:', userData);
			
			const headers: Record<string, string> = {
				'Content-Type': 'application/json'
			};
			
			if (userData) {
				headers['X-User-Data'] = userData;
				console.log('SubscriptionOfferService: Setting X-User-Data header:', userData);
			} else {
				console.log('SubscriptionOfferService: No user data found in localStorage');
			}

			console.log('SubscriptionOfferService: Making API request with headers:', headers);
			console.log('SubscriptionOfferService: Request body:', JSON.stringify(data));

			const response = await apiRequest(this.BASE_PATH, {
				method: 'POST',
				headers,
				body: JSON.stringify(data)
			});

			console.log('SubscriptionOfferService: Response received:', {
				status: response.status,
				statusText: response.statusText,
				ok: response.ok,
				url: response.url
			});

			if (!response.ok) {
				const errorText = await response.text();
				console.error('SubscriptionOfferService: Error response body:', errorText);
				throw new Error(`Failed to create subscription offer: ${response.status} - ${errorText}`);
			}

			const result = await response.json();
			console.log('SubscriptionOfferService: Successfully created offer:', result);
			return result as SubscriptionOffer;
		} catch (error) {
			console.error('SubscriptionOfferService: Error creating subscription offer:', error);
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

			if (!response.ok) {
				throw new Error(`Failed to update subscription offer: ${response.status}`);
			}

			const result = await response.json();
			return result as SubscriptionOffer;
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

			const response = await apiRequest(`${this.BASE_PATH}${id}`, {
				method: 'DELETE',
				headers
			});

			if (!response.ok) {
				throw new Error(`Failed to delete subscription offer: ${response.status}`);
			}
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

			if (!response.ok) {
				throw new Error(`Failed to toggle subscription offer status: ${response.status}`);
			}

			const result = await response.json();
			return result as SubscriptionOffer;
		} catch (error) {
			console.error('Error toggling subscription offer status:', error);
			throw error;
		}
	}

	// Get offer history
	static async getOfferHistory(id: string): Promise<Record<string, any>[]> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}/history`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch offer history: ${response.status}`);
			}

			const data = await response.json();
			return data as Record<string, any>[];
		} catch (error) {
			console.error('Error getting offer history:', error);
			throw error;
		}
	}
} 