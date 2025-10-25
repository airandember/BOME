import { writable, derived } from 'svelte/store';
import type { EnhancedSubscriber } from '$lib/types/enhanced-subscriber';
import { api } from '$lib/api';
import { subscriberElasticService, type UnifiedSubscriber } from '$lib/services/subscriber-elastic-service';

// Types
export interface SubscriberFilters {
	search?: string;
	hasActivePlan?: boolean;
	hasVideoAccess?: boolean;
	planType?: 'premium' | 'basic' | 'none';
	emailVerified?: boolean;
	role?: string;
	isExpiringSoon?: boolean;
	minMRR?: number;
	maxMRR?: number;
	minARR?: number;
	maxARR?: number;
}

export interface SubscriberKPIs {
	total_subscribers: number;
	active_subscribers: number;
	video_access_users: number;
	total_mrr: number;
	total_arr?: number; // Optional for now
	avg_days_to_expiry: number;
	churn_risk_count: number;
	premium_users: number;
	basic_users: number;
	manual_access_users: number;
	monthly_subscribers?: number; // Optional for now
	yearly_subscribers?: number; // Optional for now
}

// Store State
interface SubscriberStoreState {
	subscribers: EnhancedSubscriber[];
	kpis: SubscriberKPIs | null;
	loading: boolean;
	error: string | null;
	lastUpdated: Date | null;
	filters: SubscriberFilters;
}

// Initial State
const initialState: SubscriberStoreState = {
	subscribers: [],
	kpis: null,
	loading: false,
	error: null,
	lastUpdated: null,
	filters: {}
};

// Create Stores
export const subscriberStore = writable<SubscriberStoreState>(initialState);

// Derived Stores
export const filteredSubscribers = derived(
	subscriberStore,
	($store) => {
		if (!$store.subscribers.length) return [];
		
		let filtered = $store.subscribers;
		const filters = $store.filters;
		
		// Search filter
		if (filters.search?.trim()) {
			const searchTerm = filters.search.toLowerCase();
			filtered = filtered.filter(sub =>
				sub.email.toLowerCase().includes(searchTerm) ||
				sub.first_name?.toLowerCase().includes(searchTerm) ||
				sub.last_name?.toLowerCase().includes(searchTerm) ||
				sub.plan_name?.toLowerCase().includes(searchTerm)
			);
		}
		
		// Active plan filter
		if (filters.hasActivePlan !== undefined) {
			filtered = filtered.filter(sub => sub.has_active_plan === filters.hasActivePlan);
		}
		
		// Video access filter
		if (filters.hasVideoAccess !== undefined) {
			filtered = filtered.filter(sub => sub.has_video_access === filters.hasVideoAccess);
		}
		
		// Plan type filter
		if (filters.planType) {
			filtered = filtered.filter(sub => sub.plan_type === filters.planType);
		}
		
		// Email verified filter
		if (filters.emailVerified !== undefined) {
			filtered = filtered.filter(sub => sub.email_verified === filters.emailVerified);
		}
		
		// Role filter
		if (filters.role) {
			filtered = filtered.filter(sub => sub.role === filters.role);
		}
		
		// Expiring soon filter
		if (filters.isExpiringSoon !== undefined) {
			filtered = filtered.filter(sub => sub.is_expiring_soon === filters.isExpiringSoon);
		}
		
		// MRR range filters
		if (filters.minMRR !== undefined) {
			filtered = filtered.filter(sub => sub.mrr_contribution >= filters.minMRR!);
		}
		
		if (filters.maxMRR !== undefined) {
			filtered = filtered.filter(sub => sub.mrr_contribution <= filters.maxMRR!);
		}
		
		// ARR range filters (ARR = MRR × 12)
		if (filters.minARR !== undefined) {
			filtered = filtered.filter(sub => (Number(sub.mrr_contribution) || 0) * 12 >= filters.minARR!);
		}
		
		if (filters.maxARR !== undefined) {
			filtered = filtered.filter(sub => (Number(sub.mrr_contribution) || 0) * 12 <= filters.maxARR!);
		}
		
		return filtered;
	}
);

// Computed KPIs from filtered data
export const computedKPIs = derived(
	[subscriberStore, filteredSubscribers],
	([$store, $filtered]) => {
		console.log('🔍 computedKPIs - store.subscribers.length:', $store.subscribers.length);
		
		if (!$store.subscribers.length) {
			console.log('🔍 computedKPIs - returning null (no subscribers)');
			return null;
		}
		
		const allSubs = $store.subscribers;
		console.log('🔍 allSubs sample:', allSubs[0]);
		
		// SIMPLE LINEAR CALCULATIONS
		const totalSubscribers = allSubs.length;
		const activeSubscribers = allSubs.filter(sub => sub.has_active_plan === true).length;
		const videoAccessUsers = allSubs.filter(sub => sub.has_video_access === true).length;
		const churnRiskCount = allSubs.filter(sub => sub.is_expiring_soon === true).length;
		
		// Simple MRR calculation
		let totalMRR = 0;
		for (const sub of allSubs) {
			const mrr = Number(sub.mrr_contribution) || 0;
			if (!isNaN(mrr)) {
				totalMRR += mrr;
			}
		}
		
		// Simple ARR = MRR * 12
		const totalARR = totalMRR * 12;
		
		console.log('🔍 SIMPLE CALC - totalSubscribers:', totalSubscribers);
		console.log('🔍 SIMPLE CALC - totalMRR:', totalMRR);
		console.log('🔍 SIMPLE CALC - totalARR:', totalARR);
		
		return {
			total_subscribers: totalSubscribers,
			active_subscribers: activeSubscribers,
			video_access_users: videoAccessUsers,
			total_mrr: totalMRR,
			total_arr: totalARR,
			avg_days_to_expiry: 0,
			churn_risk_count: churnRiskCount,
			premium_users: 0,
			basic_users: 0,
			manual_access_users: 0,
			monthly_subscribers: 0,
			yearly_subscribers: 0,
		};
	}
);

// Store Actions
export const subscriberStoreActions = {
	// Load all data once
	async loadSubscribers() {
		subscriberStore.update(state => ({ ...state, loading: true, error: null }));
		
		try {
			console.log('🔄 Loading all subscriber data using UNIFIED ELASTIC SERVICE...');
			console.log('🔍 About to call elastic service: subscriberElasticService.getAllSubscribers()');
			
			// Load ALL subscribers using the new unified elastic service
			const unifiedSubscribers = await subscriberElasticService.getAllSubscribers();
			
			console.log('🔍 Elastic service response received:', unifiedSubscribers);
			console.log('🔍 Response array length:', unifiedSubscribers.length);
			console.log('🔍 First subscriber:', unifiedSubscribers[0]);
			
			// Convert UnifiedSubscriber to EnhancedSubscriber format for compatibility
			const enhancedSubscribers: EnhancedSubscriber[] = unifiedSubscribers.map(sub => ({
				id: sub.id,
				email: sub.email,
				first_name: sub.first_name,
				last_name: sub.last_name,
				role: sub.role,
				email_verified: sub.email_verified,
				is_active: sub.is_active,
				created_at: sub.created_at,
				last_login: sub.last_login,
				stripe_customer_id: sub.stripe_customer_id,
				stripe_customer_ids: sub.stripe_customer_ids,
				subscription_id: sub.subscription_id,
				plan_name: sub.plan_name || 'No Plan',
				plan_type: sub.plan_type,
				plan_status: sub.plan_status,
				plan_price: sub.plan_price,
				plan_currency: sub.plan_currency,
				plan_interval: sub.plan_interval,
				plan_start_date: sub.plan_start_date,
				billing_period_start: sub.billing_period_start,
				billing_period_end: sub.billing_period_end,
				days_until_expiry: sub.days_until_expiry,
				has_active_plan: sub.has_active_plan,
				has_video_access: sub.has_video_access,
				manual_access_granted: sub.manual_access_granted,
				mrr_contribution: sub.mrr_contribution,
				arr_contribution: sub.arr_contribution,
				ltv_estimate: sub.ltv_estimate,
				account_age_days: sub.account_age_days,
				plan_legacy_status: sub.plan_legacy_status,
				full_name: sub.full_name,
				is_expiring_soon: sub.is_expiring_soon,
				// Additional fields for compatibility
				is_high_value: sub.ltv_estimate > 1000,
				subscription_duration_days: sub.account_age_days,
				updated_at: sub.created_at // Use created_at as fallback
			}));
			
			subscriberStore.update(state => ({
				...state,
				subscribers: enhancedSubscribers,
				loading: false,
				lastUpdated: new Date(),
				error: null
			}));
			
			console.log(`✅ Loaded ${enhancedSubscribers.length} subscribers using UNIFIED ELASTIC SERVICE`);
			console.log('🔍 Sample subscriber data:', enhancedSubscribers[0]);
		} catch (error) {
			console.error('❌ Failed to load subscribers:', error);
			subscriberStore.update(state => ({
				...state,
				loading: false,
				error: error instanceof Error ? error.message : 'Failed to load subscribers'
			}));
		}
	},
	
	// Update filters (triggers reactive filtering)
	setFilters(newFilters: Partial<SubscriberFilters>) {
		subscriberStore.update(state => ({
			...state,
			filters: { ...state.filters, ...newFilters }
		}));
	},
	
	// Clear all filters
	clearFilters() {
		subscriberStore.update(state => ({
			...state,
			filters: {}
		}));
	},
	
	// Add/update a subscriber (for real-time updates)
	updateSubscriber(updatedSubscriber: EnhancedSubscriber) {
		subscriberStore.update(state => {
			const index = state.subscribers.findIndex(sub => sub.id === updatedSubscriber.id);
			const newSubscribers = [...state.subscribers];
			
			if (index >= 0) {
				newSubscribers[index] = updatedSubscriber;
			} else {
				newSubscribers.push(updatedSubscriber);
			}
			
			return { ...state, subscribers: newSubscribers };
		});
	},
	
	// Remove a subscriber
	removeSubscriber(subscriberId: number) {
		subscriberStore.update(state => ({
			...state,
			subscribers: state.subscribers.filter(sub => sub.id !== subscriberId)
		}));
	},
	
	// Refresh data
	async refresh() {
		await this.loadSubscribers();
	}
};
