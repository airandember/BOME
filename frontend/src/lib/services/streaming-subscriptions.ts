import { apiRequest } from '$lib/auth';

export interface StripeProduct {
	id: number;
	stripe_id: string;
	name: string;
	description: string | null;
	active: boolean;
	available: boolean;
	created_at: string;
	updated_at?: string;
	metadata?: Record<string, string>;
	url?: string;
	images?: string[];
	package_dimensions?: {
		height?: number;
		length?: number;
		weight?: number;
		width?: number;
	};
	shippable?: boolean;
	statement_descriptor?: string;
	tax_code?: string;
	unit_label?: string;
	livemode?: boolean;
	// Pricing information from stripe_prices table
	price?: number; // unit_amount in cents
	price_id?: string; // Stripe price ID
	currency?: string; // Price currency (e.g., 'usd')
	recurring_interval?: string; // Billing interval (e.g., 'month', 'year')
}

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
	stripe_product_id: string | null; // Link to Stripe product
	features: string[];
	is_active: boolean;
	promotion_end_date: string | null;
	promotion_start_date: string | null;
	plan_change_history: Record<string, any>[]; // Array of history event objects
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
	stripe_product_id?: string | null; // Optional Stripe product ID
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
	private static readonly BASE_PATH = '/admin/subscription-plans/';

	/**
	 * Get all subscription plans
	 */
	static async getAll(): Promise<SubscriptionPlan[]> {
		try {
			console.log('StreamingSubscriptionService: Making request to:', this.BASE_PATH);
			const response = await apiRequest(this.BASE_PATH);

			if (!response.ok) {
				throw new Error(`Failed to fetch subscription plans: ${response.status}`);
			}

			const data = await response.json();
			console.log('StreamingSubscriptionService: Response received:', data);
			return data as SubscriptionPlan[];
		} catch (error) {	
			console.error('StreamingSubscriptionService: Error fetching subscription plans:', error);
			throw error;
		}
	}

	/**
	 * Get all subscription plans including available Stripe products as potential plans
	 */
	static async getAllWithStripeProducts(): Promise<SubscriptionPlan[]> {
		try {
			// Get existing subscription plans
			const existingPlans = await this.getAll();
			
			// Get available Stripe products
			const stripeProducts = await this.getAvailableStripeProducts();
			
			// Convert Stripe products to subscription plan format
			const stripeAsPlans: SubscriptionPlan[] = stripeProducts.map(product => ({
				id: `stripe_${product.stripe_id}`, // Prefix to distinguish from real plans
				name: product.name,
				description: product.description || 'Imported from Stripe',
				short_desc: product.description?.substring(0, 100) || 'Stripe Product',
				price: product.price ? product.price / 100 : 0, // Convert from cents to dollars
				currency: product.currency || 'usd',
				interval: product.recurring_interval || 'month' as const,
				interval_count: 1,
				stripe_price_id: product.price_id || null,
				stripe_product_id: product.stripe_id,
				features: [],
				is_active: product.active && product.available,
				promotion_end_date: null,
				promotion_start_date: null,
				plan_change_history: [],
				promotion_metadata: {},
				sub_type: 'stnd',
				created_at: product.created_at,
				updated_at: product.created_at,
				is_deleted: false
			}));

			// Combine existing plans with Stripe products
			// Filter out Stripe products that already have corresponding plans
			const existingStripeProductIds = new Set(
				existingPlans
					.filter(plan => plan.stripe_product_id)
					.map(plan => plan.stripe_product_id)
			);

			const newStripeProducts = stripeAsPlans.filter(
				stripePlan => !existingStripeProductIds.has(stripePlan.stripe_product_id)
			);

			console.log('StreamingSubscriptionService: Combined plans:', {
				existing: existingPlans.length,
				stripeProducts: newStripeProducts.length,
				total: existingPlans.length + newStripeProducts.length
			});

			return [...existingPlans, ...newStripeProducts];
		} catch (error) {
			console.error('StreamingSubscriptionService: Error fetching plans with Stripe products:', error);
			// Fallback to just existing plans if Stripe fetch fails
			return await this.getAll();
		}
	}

	/**
	 * Get a single subscription plan by ID
	 */
	static async getById(id: string): Promise<SubscriptionPlan> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch subscription plan: ${response.status}`);
			}

			const data = await response.json();
			return data as SubscriptionPlan;
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
			// Get user data directly from localStorage
			const userDataRaw = typeof window !== 'undefined' ? localStorage.getItem('bome_user_data') : null;
			const headers: Record<string, string> = {};
			
			console.log('StreamingSubscriptionService: Raw user data from localStorage:', userDataRaw);
			if (userDataRaw) {
				headers['X-User-Data'] = userDataRaw;
				console.log('StreamingSubscriptionService: Setting X-User-Data header:', userDataRaw);
			} else {
				console.log('StreamingSubscriptionService: No user data found in localStorage');
			}
			
			const response = await apiRequest(this.BASE_PATH, {
				method: 'POST',
				body: JSON.stringify(data),
				headers
			});

			if (!response.ok) {
				throw new Error(`Failed to create subscription plan: ${response.status}`);
			}

			const result = await response.json();
			return result as SubscriptionPlan;
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
			// Get user data directly from localStorage
			const userDataRaw = typeof window !== 'undefined' ? localStorage.getItem('bome_user_data') : null;
			const headers: Record<string, string> = {};
			
			console.log('StreamingSubscriptionService: Raw user data from localStorage:', userDataRaw);
			if (userDataRaw) {
				headers['X-User-Data'] = userDataRaw;
				console.log('StreamingSubscriptionService: Setting X-User-Data header:', userDataRaw);
			} else {
				console.log('StreamingSubscriptionService: No user data found in localStorage');
			}
			
			const response = await apiRequest(`${this.BASE_PATH}${data.id}`, {
				method: 'PUT',
				body: JSON.stringify(data),
				headers
			});

			if (!response.ok) {
				throw new Error(`Failed to update subscription plan: ${response.status}`);
			}

			const result = await response.json();
			return result as SubscriptionPlan;
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
			await apiRequest(`${this.BASE_PATH}${id}`, { method: 'DELETE' });
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
			// Get user data directly from localStorage
			const userDataRaw = typeof window !== 'undefined' ? localStorage.getItem('bome_user_data') : null;
			const headers: Record<string, string> = {};
			
			if (userDataRaw) {
				headers['X-User-Data'] = userDataRaw;
			}
			
			const response = await apiRequest(`${this.BASE_PATH}${id}/status`, {
				method: 'PUT',
				body: JSON.stringify({ is_active: isActive }),
				headers
			});

			if (!response.ok) {
				throw new Error(`Failed to toggle subscription plan status: ${response.status}`);
			}

			const result = await response.json();
			return result as SubscriptionPlan;
		} catch (error) {
			console.error('Error toggling subscription plan status:', error);
			throw error;
		}
	}

	/**
	 * Toggle promotion status of a subscription plan
	 */
	static async togglePromotion(id: string, isPromoted: boolean): Promise<SubscriptionPlan> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}/promotion`, {
				method: 'PUT',
				body: JSON.stringify({ is_promoted: isPromoted })
			});

			if (!response.ok) {
				throw new Error(`Failed to toggle promotion status: ${response.status}`);
			}

			const result = await response.json();
			return result as SubscriptionPlan;
		} catch (error) {
			console.error('Error toggling promotion status:', error);
			throw error;
		}
	}

	/**
	 * Get plan history for analytics
	 */
	static async getPlanHistory(id: string): Promise<Record<string, any>[]> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}${id}/history`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch plan history: ${response.status}`);
			}

			const data = await response.json();
			return (data as any).history as Record<string, any>[];
		} catch (error) {
			console.error('Error fetching plan history:', error);
			throw error;
		}
	}

	/**
	 * Get history statistics for analytics
	 */
	static async getHistoryStats(): Promise<Record<string, any>> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}/analytics/history-stats`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch history stats: ${response.status}`);
			}

			const data = await response.json();
			return (data as any).stats as Record<string, any>;
		} catch (error) {
			console.error('Error fetching history stats:', error);
			throw error;
		}
	}

	/**
	 * Get history events by type
	 */
	static async getHistoryByType(eventType: string, limit: number = 100): Promise<Record<string, any>[]> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}/analytics/history-by-type/${eventType}?limit=${limit}`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch history by type: ${response.status}`);
			}

			const data = await response.json();
			return (data as any).events as Record<string, any>[];
		} catch (error) {
			console.error('Error fetching history by type:', error);
			throw error;
		}
	}

	/**
	 * Get history events by user
	 */
	static async getHistoryByUser(userID: string, limit: number = 100): Promise<Record<string, any>[]> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}/analytics/history-by-user/${userID}?limit=${limit}`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch history by user: ${response.status}`);
			}

			const data = await response.json();
			return (data as any).events as Record<string, any>[];
		} catch (error) {
			console.error('Error fetching history by user:', error);
			throw error;
		}
	}

	/**
	 * Get history events by date range
	 */
	static async getHistoryByDateRange(startDate: string, endDate: string, limit: number = 100): Promise<Record<string, any>[]> {
		try {
			const response = await apiRequest(`${this.BASE_PATH}/analytics/history-by-date-range?start_date=${startDate}&end_date=${endDate}&limit=${limit}`);
			
			if (!response.ok) {
				throw new Error(`Failed to fetch history by date range: ${response.status}`);
			}

			const data = await response.json();
			return (data as any).events as Record<string, any>[];
		} catch (error) {
			console.error('Error fetching history by date range:', error);
			throw error;
		}
	}

	/**
	 * Get available Stripe products
	 */
	static async getAvailableStripeProducts(): Promise<StripeProduct[]> {
		try {
			const response = await apiRequest('/admin/streaming/stripe/products/available');
			
			if (!response.ok) {
				throw new Error(`Failed to fetch available Stripe products: ${response.status}`);
			}

			const data = await response.json();
			return data.products || [];
		} catch (error) {
			console.error('Error fetching available Stripe products:', error);
			throw error;
		}
	}

	/**
	 * Update Stripe product availability
	 */
	static async updateStripeProductAvailability(stripeProductId: string, available: boolean): Promise<void> {
		try {
			const response = await apiRequest(`/admin/streaming/stripe/products/${stripeProductId}/availability`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ available }),
			});
			
			if (!response.ok) {
				throw new Error(`Failed to update Stripe product availability: ${response.status}`);
			}
		} catch (error) {
			console.error('Error updating Stripe product availability:', error);
			throw error;
		}
	}

	/**
	 * Import selected Stripe products as subscription plans
	 */
	static async importStripeProductsAsPlans(stripeProductIds: string[]): Promise<{imported_count: number, skipped_count: number, total_count: number, message: string}> {
		try {
			const response = await apiRequest('/admin/streaming/stripe/products/import-as-plans', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ stripe_product_ids: stripeProductIds })
			});

			if (!response.ok) {
				throw new Error(`Failed to import products as plans: ${response.status}`);
			}

			return await response.json();
		} catch (error) {
			console.error('Error importing Stripe products as plans:', error);
			throw error;
		}
	}
} 
