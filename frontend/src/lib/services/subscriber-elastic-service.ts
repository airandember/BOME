// Unified Subscriber Elastic Service for Frontend
import { apiRequest } from '$lib/auth';

export interface UnifiedSubscriber {
	// User data
	id: number;
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	email_verified: boolean;
	is_active: boolean;
	created_at: string;
	last_login?: string;
	
	// Stripe customer data
	stripe_customer_id?: string;
	stripe_customer_ids: string[];
	
	// Subscription data
	subscription_id?: string;
	plan_name?: string;
	plan_type: string; // "premium", "basic", "none"
	plan_status: string; // "active", "trialing", "expired", etc.
	plan_price: number;
	plan_currency: string;
	plan_interval: string; // "monthly", "yearly"
	plan_start_date?: string;
	billing_period_start?: string;
	billing_period_end?: string;
	days_until_expiry?: number;
	
	// Access control
	has_active_plan: boolean;
	has_video_access: boolean;
	manual_access_granted: boolean;
	
	// Business intelligence
	mrr_contribution: number;
	arr_contribution: number;
	ltv_estimate: number;
	account_age_days: number;
	
	// Legacy status
	plan_legacy_status: string; // "legacy", "current", "unknown"
	
	// Computed fields
	full_name: string;
	is_expiring_soon: boolean;
}

export interface SubscriberStats {
	total_subscribers: number;
	active_plans: number;
	video_access: number;
	manual_access: number;
	expiring_soon: number;
	multiple_stripe_customers: number;
	total_mrr: number;
	total_arr: number;
	plan_types: {
		premium: number;
		basic: number;
		none: number;
	};
	legacy_status: {
		current: number;
		legacy: number;
		unknown: number;
	};
}

export interface DiagnosticData {
	summary: {
		total_subscribers: number;
		issues_found: number;
		manual_overrides: number;
	};
	issues: {
		multiple_stripe_customers: {
			count: number;
			description: string;
			subscribers: UnifiedSubscriber[];
		};
		active_plan_no_access: {
			count: number;
			description: string;
			subscribers: UnifiedSubscriber[];
		};
	};
	manual_overrides: {
		count: number;
		description: string;
		subscribers: UnifiedSubscriber[];
	};
	statistics: SubscriberStats;
}

class SubscriberElasticService {
	private baseUrl = '/admin/subscriber-elastic-v2'; // Updated to v2!

	/**
	 * Get all unified subscribers with complete data
	 */
	async getAllSubscribers(): Promise<UnifiedSubscriber[]> {
		try {
			const response = await apiRequest(`${this.baseUrl}/subscribers`);
			if (!response.ok) {
				throw new Error(`Failed to fetch subscribers: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data.subscribers || [];
		} catch (error) {
			console.error('Error fetching unified subscribers:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber by email
	 */
	async getSubscriberByEmail(email: string): Promise<UnifiedSubscriber | null> {
		try {
			const response = await apiRequest(`${this.baseUrl}/subscribers/email/${encodeURIComponent(email)}`);
			if (response.status === 404) {
				return null;
			}
			if (!response.ok) {
				throw new Error(`Failed to fetch subscriber: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data;
		} catch (error) {
			console.error('Error fetching subscriber by email:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber by ID
	 */
	async getSubscriberByID(id: number): Promise<UnifiedSubscriber | null> {
		try {
			const response = await apiRequest(`${this.baseUrl}/subscribers/id/${id}`);
			if (response.status === 404) {
				return null;
			}
			if (!response.ok) {
				throw new Error(`Failed to fetch subscriber: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data;
		} catch (error) {
			console.error('Error fetching subscriber by ID:', error);
			throw error;
		}
	}

	/**
	 * Get comprehensive diagnostic data
	 */
	async getDiagnosticData(): Promise<DiagnosticData> {
		try {
			const response = await apiRequest(`${this.baseUrl}/diagnose`);
			if (!response.ok) {
				throw new Error(`Failed to fetch diagnostic data: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data;
		} catch (error) {
			console.error('Error fetching diagnostic data:', error);
			throw error;
		}
	}

	/**
	 * Get subscribers with multiple Stripe customers (potential duplicates)
	 */
	async getSubscribersWithMultipleStripeCustomers(): Promise<UnifiedSubscriber[]> {
		try {
			const response = await apiRequest(`${this.baseUrl}/multiple-stripe-customers`);
			if (!response.ok) {
				throw new Error(`Failed to fetch multiple customers: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data.subscribers || [];
		} catch (error) {
			console.error('Error fetching multiple customers:', error);
			throw error;
		}
	}

	/**
	 * Get subscribers with active plans but no video access (potential bugs)
	 */
	async getSubscribersWithActivePlansButNoAccess(): Promise<UnifiedSubscriber[]> {
		try {
			const response = await apiRequest(`${this.baseUrl}/active-plan-no-access`);
			if (!response.ok) {
				throw new Error(`Failed to fetch no access subscribers: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data.subscribers || [];
		} catch (error) {
			console.error('Error fetching no access subscribers:', error);
			throw error;
		}
	}

	/**
	 * Get subscribers with manual video access (no active plan)
	 */
	async getSubscribersWithManualAccess(): Promise<UnifiedSubscriber[]> {
		try {
			const response = await apiRequest(`${this.baseUrl}/manual-access`);
			if (!response.ok) {
				throw new Error(`Failed to fetch manual access subscribers: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data.subscribers || [];
		} catch (error) {
			console.error('Error fetching manual access subscribers:', error);
			throw error;
		}
	}

	/**
	 * Get subscriber statistics
	 */
	async getSubscriberStats(): Promise<SubscriberStats> {
		try {
			const response = await apiRequest(`${this.baseUrl}/stats`);
			if (!response.ok) {
				throw new Error(`Failed to fetch subscriber stats: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.data;
		} catch (error) {
			console.error('Error fetching subscriber stats:', error);
			throw error;
		}
	}

	/**
	 * Update manual video access for a subscriber
	 */
	async updateManualVideoAccess(userID: number, hasAccess: boolean): Promise<void> {
		try {
			const response = await apiRequest(`${this.baseUrl}/subscribers/${userID}/manual-access`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ has_access: hasAccess }),
			});

			if (!response.ok) {
				throw new Error(`Failed to update manual access: ${response.statusText}`);
			}
		} catch (error) {
			console.error('Error updating manual video access:', error);
			throw error;
		}
	}

	/**
	 * Search subscribers with filters
	 */
	async searchSubscribers(filters: {
		search?: string;
		hasActivePlan?: boolean;
		hasVideoAccess?: boolean;
		planType?: string;
		emailVerified?: boolean;
		role?: string;
		isExpiringSoon?: boolean;
		minMRR?: number;
		maxMRR?: number;
	}): Promise<UnifiedSubscriber[]> {
		// For now, get all subscribers and filter client-side
		// In production, you'd want server-side filtering
		const allSubscribers = await this.getAllSubscribers();
		
		return allSubscribers.filter(subscriber => {
			// Search filter
			if (filters.search) {
				const search = filters.search.toLowerCase();
				const matchesSearch = 
					subscriber.email.toLowerCase().includes(search) ||
					subscriber.first_name.toLowerCase().includes(search) ||
					subscriber.last_name.toLowerCase().includes(search) ||
					subscriber.full_name.toLowerCase().includes(search);
				if (!matchesSearch) return false;
			}

			// Active plan filter
			if (filters.hasActivePlan !== undefined && subscriber.has_active_plan !== filters.hasActivePlan) {
				return false;
			}

			// Video access filter
			if (filters.hasVideoAccess !== undefined && subscriber.has_video_access !== filters.hasVideoAccess) {
				return false;
			}

			// Plan type filter
			if (filters.planType && subscriber.plan_type !== filters.planType) {
				return false;
			}

			// Email verified filter
			if (filters.emailVerified !== undefined && subscriber.email_verified !== filters.emailVerified) {
				return false;
			}

			// Role filter
			if (filters.role && subscriber.role !== filters.role) {
				return false;
			}

			// Expiring soon filter
			if (filters.isExpiringSoon !== undefined && subscriber.is_expiring_soon !== filters.isExpiringSoon) {
				return false;
			}

			// MRR range filter
			if (filters.minMRR !== undefined && subscriber.mrr_contribution < filters.minMRR) {
				return false;
			}
			if (filters.maxMRR !== undefined && subscriber.mrr_contribution > filters.maxMRR) {
				return false;
			}

			return true;
		});
	}
}

// Export singleton instance
export const subscriberElasticService = new SubscriberElasticService();
export default subscriberElasticService;
