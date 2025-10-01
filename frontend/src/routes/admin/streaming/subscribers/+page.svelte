<script lang="ts">
	import { onMount } from 'svelte';
	import { StreamingSubscriberService, type Subscriber, type NonSubscriber, type SubscriberFilters, type NonSubscriberFilters } from '$lib/services/streaming-subscribers';
	import { subscriberCache } from '$lib/cache/subscriber-cache';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import SubscriberFiltersComponent from './SubscriberFilters.svelte';
	import SubscriberTable from './SubscriberTable.svelte';
	import NonSubscriberTable from './NonSubscriberTable.svelte';
	import SubscriberPagination from './SubscriberPagination.svelte';
	import SubscriberEditModal from './SubscriberEditModal.svelte';
	import EnhancedSubscribersPage from './EnhancedSubscribersPage.svelte';
	import { auth, apiRequest } from '$lib/auth';
	import SendOfferModal from './SendOfferModal.svelte';
	import type { SubscriptionOffer } from '$lib/services/subscription-offers';
	import { SubscriptionOfferService } from '$lib/services/subscription-offers';
	import StripeCustomers from './customers/+page.svelte';

	// State
	let activeTab: 'subscribers' | 'stripe-subs' = $state('subscribers');
	let loading = $state(false);
	
	// Separate loading states for each tab
	let subscribersLoading = $state(false);
	let nonSubscribersLoading = $state(false);
	let stripeSubsLoading = $state(false);
	
	// Track which tabs have been loaded
	let subscribersLoaded = $state(false);
	let nonSubscribersLoaded = $state(false);
	let stripeSubsLoaded = $state(false);
	
	// All data arrays (source of truth)
	let allSubscribers = $state<Subscriber[]>([]);
	let allNonSubscribers = $state<NonSubscriber[]>([]);
	
	// Display arrays (what gets shown in tables)
	let displayedSubscribers = $state<Subscriber[]>([]);
	let displayedNonSubscribers = $state<NonSubscriber[]>([]);
	
	let subscriberCount = $state(0);
	let nonSubscriberCount = $state(0);
	let animationDirection: 'left' | 'right' = $state('right');
	let isAnimating = $state(false);
	let isTransitioning = $state(false);

	// Roles data for normalized names
	let roles: any[] = $state([]);
	let rolesLoading = $state(true);

	// Subscription plans for edit modal
	let subscriptionPlans: any[] = $state([]);

	// Stripe data from database
	let stripeCustomers = $state<any[]>([]);
	let stripeSubscriptions = $state<any[]>([]);

	// Edit modal state
	let showEditModal = $state(false);
	let selectedSubscriber: Subscriber | null = $state(null);

	// Pagination
	let currentPage = $state(1);
	let itemsPerPage = $state(50);
	let totalPages = $state(1);

	// Separate filters for each tab
	let subscriberSearchTerm = $state('');
	let subscriberEmailVerifiedFilter: boolean | undefined = $state(undefined);
	let subscriberRoleFilter = $state('');
	let subscriberPlanFilter = $state('');
	let subscriberLastLoginFilter = $state('');
	let subscriberCreatedDateFilter = $state('');

	let nonSubscriberSearchTerm = $state('');
	let nonSubscriberEmailVerifiedFilter: boolean | undefined = $state(undefined);
	let nonSubscriberRoleFilter = $state('');
	let nonSubscriberLastLoginFilter = $state('');
	let nonSubscriberCreatedDateFilter = $state('');
	let nonSubscriberHasSubbedFilter: boolean | undefined = $state(undefined);

	// Selection state
	let selectedSubscribers: Set<number> = $state(new Set());
	let selectedNonSubscribers: Set<number> = $state(new Set());
	let selectAllSubscribers = $state(false);
	let selectAllNonSubscribers = $state(false);

	// Simplified computed properties
	let currentSelectedItems = $derived(activeTab === 'subscribers' ? selectedSubscribers : selectedNonSubscribers);
	let currentSelectAll = $derived(activeTab === 'subscribers' ? selectAllSubscribers : selectAllNonSubscribers);
	let currentDisplayedItems = $derived(activeTab === 'subscribers' ? displayedSubscribers : displayedNonSubscribers);
	let currentDisplayedCount = $derived(activeTab === 'subscribers' ? displayedSubscribers.length : displayedNonSubscribers.length);
	let currentSelectedCount = $derived(activeTab === 'subscribers' ? selectedSubscribers.size : selectedNonSubscribers.size);

	// Reactive paginated data - updates automatically when display arrays or currentPage changes
	let paginatedSubscribers = $state<Subscriber[]>([]);
	let paginatedNonSubscribers = $state<NonSubscriber[]>([]);

	// Update paginated data when dependencies change
	$effect(() => {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		paginatedSubscribers = displayedSubscribers.slice(startIndex, endIndex);
	});

	$effect(() => {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		paginatedNonSubscribers = displayedNonSubscribers.slice(startIndex, endIndex);
	});

// Load roles data
async function fetchRoles() {
	try {
		rolesLoading = true;
		
		// Get the auth token from the auth store
		const token = $auth.token;
		if (!token) {
			console.error('No auth token found');
			return;
		}
		
				const response = await apiRequest('/admin/rolesAndDepartments');
		
		if (response.ok) {
			const data = await response.json();
			roles = data.data?.roles || [];
			console.log('FETCHED roles from API:', roles);
		} else {
			console.error('Failed to fetch roles:', response.status, response.statusText);
		}
	} catch (error) {
		console.error('Error fetching roles:', error);
	} finally {
		rolesLoading = false;
	}
}

// Load subscription plans for edit modal
async function fetchSubscriptionPlans() {
	try {
		const response = await apiRequest('/admin/subscription-plans/', {
			method: 'GET'
		});
		
		if (response.ok) {
			const data = await response.json();
			subscriptionPlans = data || [];
		}
	} catch (error) {
		console.error('Error fetching subscription plans:', error);
	}
}

// Load Stripe customers from database (all customers, no limit)
async function loadStripeCustomers() {
	try {
		console.log('🔄 Loading ALL Stripe customers from database...');
		let allCustomers: any[] = [];
		let offset = 0;
		const limit = 1000; // Batch size for API calls
		let hasMore = true;

		while (hasMore) {
			const response = await apiRequest(`/admin/streaming/stripe/database/customers?limit=${limit}&offset=${offset}&include_subscriptions=true`);
			
			if (response.ok) {
				const data = await response.json();
				const customers = data.customers || [];
				allCustomers = [...allCustomers, ...customers];
				
				console.log(`✅ Loaded batch: ${customers.length} customers (offset: ${offset}, total so far: ${allCustomers.length})`);
				
				// Check if we have more data
				hasMore = customers.length === limit;
				offset += limit;
			} else {
				console.error('❌ Failed to load Stripe customers batch:', response.status);
				hasMore = false;
			}
		}

		console.log('✅ Loaded ALL Stripe customers from database:', allCustomers.length);
		return allCustomers;
	} catch (error) {
		console.error('❌ Error loading Stripe customers:', error);
		return [];
	}
}

// Load Stripe subscriptions from database
async function loadStripeSubscriptions() {
	try {
		console.log('🔄 Loading Stripe subscriptions from database...');
		const response = await apiRequest('/admin/streaming/stripe/database/subscriptions'); // No limit
		
		if (response.ok) {
			const data = await response.json();
			const subscriptions = data.subscriptions || [];
			console.log('✅ Loaded Stripe subscriptions from database:', subscriptions.length);
			return subscriptions;
		} else {
			console.error('❌ Failed to load Stripe subscriptions:', response.status);
			return [];
		}
	} catch (error) {
		console.error('❌ Error loading Stripe subscriptions:', error);
		return [];
	}
}


// Load data on mount
onMount(async () => {
	await fetchRoles();
	await fetchSubscriptionPlans();
	// Load the active tab's data immediately
	await loadTabData(activeTab);
});

// Load data for a specific tab
async function loadTabData(tab: 'subscribers' | 'non-subscribers' | 'stripe-subs') {
	console.log(`🔄 Loading data for tab: ${tab}`);
	
	switch (tab) {
		case 'subscribers':
			await loadSubscribersData();
			break;
		case 'non-subscribers':
			await loadNonSubscribersData();
			break;
		case 'stripe-subs':
			await loadStripeSubsData();
			break;
	}
}

// Load subscribers data
async function loadSubscribersData() {
	if (subscribersLoaded) {
		console.log('✅ Subscribers already loaded, skipping...');
		return;
	}
	
	subscribersLoading = true;
	try {
		console.log('🔄 Loading subscribers...');
		
		const subscribersResponse = await subscriberCache.getSubscribers(1, 0, {}); // Get all subscribers
		allSubscribers = subscribersResponse.subscribers || [];
		subscriberCount = subscribersResponse.total_count || allSubscribers.length;
		
		console.log('✅ Loaded subscribers:', allSubscribers.length);
		
		// Initialize display array
		displayedSubscribers = [...allSubscribers];
		
		// Apply filters if any are set
		if (activeTab === 'subscribers') {
			debounceApplyFilters();
		}
		
		subscribersLoaded = true;
	} catch (error) {
		console.error('❌ Error loading subscribers:', error);
		showToast('Failed to load subscribers', 'error');
		allSubscribers = [];
		displayedSubscribers = [];
		subscriberCount = 0;
	} finally {
		subscribersLoading = false;
	}
}

// Load non-subscribers data
async function loadNonSubscribersData() {
	if (nonSubscribersLoaded) {
		console.log('✅ Non-subscribers already loaded, skipping...');
		return;
	}
	
	nonSubscribersLoading = true;
	try {
		console.log('🔄 Loading non-subscribers...');
		
		const nonSubscribersResponse = await StreamingSubscriberService.getNonSubscribers({
			limit: 0, // Get all non-subscribers (no limit)
			offset: 0
		});
		allNonSubscribers = nonSubscribersResponse.non_subscribers || [];
		nonSubscriberCount = allNonSubscribers.length;
		
		console.log('✅ Loaded non-subscribers:', allNonSubscribers.length);
		
		// Initialize display array
		displayedNonSubscribers = [...allNonSubscribers];
		
		// Apply filters if any are set
		if (activeTab === 'non-subscribers') {
			debounceApplyFilters();
		}
		
		nonSubscribersLoaded = true;
	} catch (error) {
		console.error('❌ Error loading non-subscribers:', error);
		showToast('Failed to load non-subscribers', 'error');
		allNonSubscribers = [];
		displayedNonSubscribers = [];
		nonSubscriberCount = 0;
	} finally {
		nonSubscribersLoading = false;
	}
}

// Load Stripe subs data
async function loadStripeSubsData() {
	if (stripeSubsLoaded) {
		console.log('✅ Stripe subs already loaded, skipping...');
		return;
	}
	
	stripeSubsLoading = true;
	try {
		console.log('🔄 Loading Stripe subs data...');
		
		// Load Stripe data from database in parallel
		const [customers, subscriptions] = await Promise.all([
			loadStripeCustomers(),
			loadStripeSubscriptions()
		]);
		
		stripeCustomers = customers;
		stripeSubscriptions = subscriptions;
		
		console.log('✅ Loaded Stripe subs data:', {
			customers: stripeCustomers.length,
			subscriptions: stripeSubscriptions.length
		});
		
		stripeSubsLoaded = true;
	} catch (error) {
		console.error('❌ Error loading Stripe subs data:', error);
		showToast('Failed to load Stripe data', 'error');
		stripeCustomers = [];
		stripeSubscriptions = [];
	} finally {
		stripeSubsLoading = false;
	}
}
	
	// Apply filters to the loaded data - optimized for real-time performance
	function applyFilters() {
		const currentFilters = getCurrentFilters();
		
		// Check if any filters are applied
		const hasFilters = currentFilters.searchTerm || 
			currentFilters.emailVerifiedFilter !== undefined || 
			currentFilters.roleFilter || 
			currentFilters.planFilter ||
			(currentFilters.lastLoginFilter && currentFilters.lastLoginFilter.trim() !== '') || 
			(currentFilters.createdDateFilter && currentFilters.createdDateFilter.trim() !== '') ||
			currentFilters.hasSubbedFilter !== undefined;
		
		console.log('Applying filters:', {
			activeTab,
			hasFilters,
			filters: currentFilters,
			lastLoginFilter: currentFilters.lastLoginFilter,
			createdDateFilter: currentFilters.createdDateFilter
		});
		
		if (activeTab === 'subscribers') {
			if (!hasFilters) {
				// No filters applied - show all subscribers (instant)
				displayedSubscribers = [...allSubscribers];
			} else {
				// Apply filters efficiently
				displayedSubscribers = allSubscribers.filter(subscriber => {
					// Search filter (most common, check first)
					if (currentFilters.searchTerm) {
						const search = currentFilters.searchTerm.toLowerCase();
						const matchesSearch = subscriber.email.toLowerCase().includes(search) ||
							subscriber.first_name.toLowerCase().includes(search) ||
							subscriber.last_name.toLowerCase().includes(search);
						if (!matchesSearch) return false;
					}
					
					// Email verified filter (simple boolean check)
					if (currentFilters.emailVerifiedFilter !== undefined && 
						subscriber.email_verified !== currentFilters.emailVerifiedFilter) {
						return false;
					}
					
					// Role filter (only if roles are loaded)
					if (currentFilters.roleFilter && roles.length > 0) {
						const userRole = roles.find(r => r.id === subscriber.role);
						const userRoleName = userRole ? userRole.name : subscriber.role;
						if (userRoleName !== currentFilters.roleFilter) {
							return false;
						}
					}
					
					// Plan filter (only for subscribers)
					if (currentFilters.planFilter && subscriber.plan_name !== currentFilters.planFilter) {
						return false;
					}
					
					// Last login filter
					if (currentFilters.lastLoginFilter) {
						console.log('🔍 Applying last login filter to subscriber:', subscriber.email, 'Filter:', currentFilters.lastLoginFilter);
						const matches = matchesLastLoginFilter(subscriber, currentFilters.lastLoginFilter);
						if (!matches) {
							console.log('❌ Subscriber filtered out by last login:', {
								subscriber: subscriber.email,
								lastLogin: subscriber.last_login,
								filter: currentFilters.lastLoginFilter,
								matches
							});
							return false;
						}
						console.log('✅ Subscriber passed last login filter:', subscriber.email);
					}
					
					// Created date filter
					if (currentFilters.createdDateFilter) {
						const matches = matchesCreatedDateFilter(subscriber, currentFilters.createdDateFilter);
						if (!matches) {
							console.log('❌ Subscriber filtered out by created date:', {
								subscriber: subscriber.email,
								createdAt: subscriber.created_at,
								filter: currentFilters.createdDateFilter,
								matches
							});
							return false;
						}
						console.log('✅ Subscriber passed created date filter:', subscriber.email);
					}
					
					return true;
				});
			}
			
			// Update pagination
			totalPages = Math.ceil(displayedSubscribers.length / itemsPerPage);
			if (currentPage > totalPages) {
				currentPage = 1;
			}
		} else {
			if (!hasFilters) {
				// No filters applied - show all non-subscribers (instant)
				displayedNonSubscribers = [...allNonSubscribers];
			} else {
				// Apply filters efficiently
				displayedNonSubscribers = allNonSubscribers.filter(nonSubscriber => {
					// Search filter (most common, check first)
					if (currentFilters.searchTerm) {
						const search = currentFilters.searchTerm.toLowerCase();
						const matchesSearch = nonSubscriber.email.toLowerCase().includes(search) ||
							nonSubscriber.first_name.toLowerCase().includes(search) ||
							nonSubscriber.last_name.toLowerCase().includes(search);
						if (!matchesSearch) return false;
					}
					
					// Email verified filter (simple boolean check)
					if (currentFilters.emailVerifiedFilter !== undefined && 
						nonSubscriber.email_verified !== currentFilters.emailVerifiedFilter) {
						return false;
					}
					
					// Role filter (only if roles are loaded)
					if (currentFilters.roleFilter && roles.length > 0) {
						const userRole = roles.find(r => r.id === nonSubscriber.role);
						const userRoleName = userRole ? userRole.name : "No Name";
						if (userRoleName !== currentFilters.roleFilter) {
							return false;
						}
					}
					
					// Last login filter
					if (currentFilters.lastLoginFilter) {
						console.log('🔍 Applying last login filter to non-subscriber:', nonSubscriber.email, 'Filter:', currentFilters.lastLoginFilter);
						const matches = matchesLastLoginFilter(nonSubscriber, currentFilters.lastLoginFilter);
						if (!matches) {
							console.log('❌ Non-subscriber filtered out by last login:', {
								nonSubscriber: nonSubscriber.email,
								lastLogin: nonSubscriber.last_login,
								filter: currentFilters.lastLoginFilter,
								matches
							});
							return false;
						}
						console.log('✅ Non-subscriber passed last login filter:', nonSubscriber.email);
					}
					
					// Created date filter
					if (currentFilters.createdDateFilter) {
						console.log('🔍 Applying created date filter to non-subscriber:', nonSubscriber.email, 'Filter:', currentFilters.createdDateFilter);
						const matches = matchesCreatedDateFilter(nonSubscriber, currentFilters.createdDateFilter);
						if (!matches) {
							console.log('❌ Non-subscriber filtered out by created date:', {
								nonSubscriber: nonSubscriber.email,
								createdAt: nonSubscriber.created_at,
								filter: currentFilters.createdDateFilter,
								matches
							});
							return false;
						}
						console.log('✅ Non-subscriber passed created date filter:', nonSubscriber.email);
					}
					
					// Has Subbed filter
					if (currentFilters.hasSubbedFilter !== undefined) {
						if (nonSubscriber.has_subscription_history !== currentFilters.hasSubbedFilter) {
							console.log('❌ Non-subscriber filtered out by has subbed filter:', {
								nonSubscriber: nonSubscriber.email,
								hasSubscriptionHistory: nonSubscriber.has_subscription_history,
								filter: currentFilters.hasSubbedFilter
							});
							return false;
						}
						console.log('✅ Non-subscriber passed has subbed filter:', nonSubscriber.email);
					}
					
					return true;
				});
			}
			
			// Update pagination
			totalPages = Math.ceil(displayedNonSubscribers.length / itemsPerPage);
			if (currentPage > totalPages) {
				currentPage = 1;
			}
		}
	}

	// Helper function to check if item matches last login filter - updated for date ranges
	function matchesLastLoginFilter(item: Subscriber | NonSubscriber, filter: string): boolean {
		console.log('🔍 matchesLastLoginFilter called:', { 
			itemEmail: item.email, 
			lastLogin: item.last_login, 
			filter 
		});
		
		if (!filter || filter.trim() === '') {
			console.log('✅ No filter provided, returning true');
			return true;
		}
		
		// Handle the new date range format (startDate|endDate)
		if (filter.includes('|')) {
			const [startDateStr, endDateStr] = filter.split('|');
			console.log('📅 Date range filter detected:', { startDateStr, endDateStr });
			
			// If both dates are empty, return true (no filter)
			if ((!startDateStr || startDateStr.trim() === '') && (!endDateStr || endDateStr.trim() === '')) {
				console.log('✅ Both dates empty, returning true');
				return true;
			}
			
			try {
				let startDate: Date | null = null;
				let endDate: Date | null = null;
				
				// Parse start date if provided - set to 00:00:00
				if (startDateStr && startDateStr.trim() !== '') {
					startDate = new Date(startDateStr + 'T00:00:00');
					if (isNaN(startDate.getTime())) {
						console.error('❌ Invalid start date:', startDateStr);
						return true;
					}
					console.log('📅 Parsed start date:', startDate);
				}
				
				// Parse end date if provided - set to 23:59:59
				if (endDateStr && endDateStr.trim() !== '') {
					endDate = new Date(endDateStr + 'T23:59:59');
					if (isNaN(endDate.getTime())) {
						console.error('❌ Invalid end date:', endDateStr);
						return true;
					}
					console.log('📅 Parsed end date:', endDate);
				}
				
				if (!item.last_login) {
					console.log('❌ No last login for item:', item.email);
					return false;
				}
				
				const lastLogin = new Date(item.last_login);
				if (isNaN(lastLogin.getTime())) {
					console.error('❌ Invalid last_login date:', item.last_login);
					return false;
				}
				
				console.log('📅 Item last login:', lastLogin);
				
				// Compare dates - check if lastLogin is within the range (inclusive)
				let result = true;
				
				if (startDate) {
					const startComparison = lastLogin >= startDate;
					result = result && startComparison;
					console.log('🔍 Start date comparison:', { lastLogin, startDate, startComparison, result });
				}
				
				if (endDate) {
					const endComparison = lastLogin <= endDate;
					result = result && endComparison;
					console.log('🔍 End date comparison:', { lastLogin, endDate, endComparison, result });
				}
				
				console.log('✅ Final result for', item.email, ':', result);
				return result;
			} catch (error) {
				console.error('❌ Error parsing last login date range:', error);
				return true;
			}
		}
		
		// Handle legacy filter values for backward compatibility
		if (!item.last_login) {
			const legacyResult = filter === 'never';
			console.log('🔍 Legacy filter (no last login):', { filter, result: legacyResult });
			return legacyResult;
		}
		
		const lastLogin = new Date(item.last_login);
		const now = new Date();
		
		switch (filter) {
			case 'today':
				const todayResult = lastLogin.toDateString() === now.toDateString();
				console.log('🔍 Today filter:', { lastLogin, now, result: todayResult });
				return todayResult;
			case 'week':
				const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
				const weekResult = lastLogin >= weekAgo;
				console.log('🔍 Week filter:', { lastLogin, weekAgo, result: weekResult });
				return weekResult;
			case 'month':
				const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
				const monthResult = lastLogin >= monthAgo;
				console.log('🔍 Month filter:', { lastLogin, monthAgo, result: monthResult });
				return monthResult;
			default:
				console.log('🔍 Default case, returning true');
				return true;
		}
	}

	// Helper function to check if item matches created date filter - updated for date ranges
	function matchesCreatedDateFilter(item: Subscriber | NonSubscriber, filter: string): boolean {
		if (!filter || filter.trim() === '') return true;
		
		// Handle the new date range format (startDate|endDate)
		if (filter.includes('|')) {
			const [startDateStr, endDateStr] = filter.split('|');
			
			// If both dates are empty, return true (no filter)
			if ((!startDateStr || startDateStr.trim() === '') && (!endDateStr || endDateStr.trim() === '')) {
				return true;
			}
			
			try {
				let startDate: Date | null = null;
				let endDate: Date | null = null;
				
				// Parse start date if provided - set to 00:00:00
				if (startDateStr && startDateStr.trim() !== '') {
					startDate = new Date(startDateStr + 'T00:00:00');
					if (isNaN(startDate.getTime())) {
						return true;
					}
				}
				
				// Parse end date if provided - set to 23:59:59
				if (endDateStr && endDateStr.trim() !== '') {
					endDate = new Date(endDateStr + 'T23:59:59');
					if (isNaN(endDate.getTime())) {
						return true;
					}
				}
				
				const created = new Date(item.created_at);
				if (isNaN(created.getTime())) {
					return false;
				}
				
				// Compare dates - check if created date is within the range (inclusive)
				let result = true;
				
				if (startDate) {
					result = result && created >= startDate;
				}
				
				if (endDate) {
					result = result && created <= endDate;
				}
				
				return result;
			} catch (error) {
				return true;
			}
		}
		
		// Handle legacy filter values for backward compatibility
		const created = new Date(item.created_at);
		const now = new Date();
		
		switch (filter) {
			case 'today':
				return created.toDateString() === now.toDateString();
			case 'week':
				const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
				return created >= weekAgo;
			case 'month':
				const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
				return created >= monthAgo;
			case 'year':
				const yearAgo = new Date(now.getTime() - 365 * 24 * 60 * 60 * 1000);
				return created >= yearAgo;
			default:
				return true;
		}
	}

	// Get paginated data from display arrays
	function getPaginatedData() {
		const startIndex = (currentPage - 1) * itemsPerPage;
		const endIndex = startIndex + itemsPerPage;
		
		if (activeTab === 'subscribers') {
			return displayedSubscribers.slice(startIndex, endIndex);
		} else {
			return displayedNonSubscribers.slice(startIndex, endIndex);
		}
	}

	// Debounced filter application to prevent infinite loops
	let filterTimeout: NodeJS.Timeout | null = null;
	
	function debounceApplyFilters() {
		if (filterTimeout) {
			clearTimeout(filterTimeout);
		}
		filterTimeout = setTimeout(() => {
			applyFilters();
		}, 100); // 100ms debounce
	}

	// Apply filters when data is loaded for the first time
	$effect(() => {
		if (activeTab === 'subscribers' && allSubscribers.length > 0 && subscribersLoaded) {
			// Only apply on initial load, not on every filter change
			const hasAnyFilters = subscriberSearchTerm || subscriberEmailVerifiedFilter !== undefined || 
				subscriberRoleFilter || subscriberPlanFilter || subscriberLastLoginFilter || subscriberCreatedDateFilter;
			if (!hasAnyFilters) {
				displayedSubscribers = [...allSubscribers];
			}
		}
	});

	$effect(() => {
		if (activeTab === 'non-subscribers' && allNonSubscribers.length > 0 && nonSubscribersLoaded) {
			// Only apply on initial load, not on every filter change
			const hasAnyFilters = nonSubscriberSearchTerm || nonSubscriberEmailVerifiedFilter !== undefined || 
				nonSubscriberRoleFilter || nonSubscriberLastLoginFilter || nonSubscriberCreatedDateFilter || 
				nonSubscriberHasSubbedFilter !== undefined;
			if (!hasAnyFilters) {
				displayedNonSubscribers = [...allNonSubscribers];
			}
		}
	});

	// Keep only the direct callback function
	function handleSelectItemDirect(event: CustomEvent<{ itemId: number; checked: boolean }>) {
		const { itemId, checked } = event.detail;
		console.log('handleSelectItemDirect called:', { itemId, checked, activeTab });
		
		if (itemId === -1) {
			// Header checkbox clicked - toggle all
			if (checked) {
				// Select all displayed items - create new Set for reactivity
				if (activeTab === 'subscribers') {
					selectedSubscribers = new Set(currentDisplayedItems.map(item => item.id));
					selectAllSubscribers = true;
				} else {
					selectedNonSubscribers = new Set(currentDisplayedItems.map(item => item.id));
					selectAllNonSubscribers = true;
				}
			} else {
				// Deselect all - create new empty Set for reactivity
				if (activeTab === 'subscribers') {
					selectedSubscribers = new Set();
					selectAllSubscribers = false;
				} else {
					selectedNonSubscribers = new Set();
					selectAllNonSubscribers = false;
				}
			}
		} else {
			// Individual item checkbox clicked - create new Set for reactivity
			if (activeTab === 'subscribers') {
				const newSelected = new Set(selectedSubscribers);
				if (checked) {
					newSelected.add(itemId);
				} else {
					newSelected.delete(itemId);
				}
				selectedSubscribers = newSelected;
				selectAllSubscribers = newSelected.size === currentDisplayedItems.length;
			} else {
				const newSelected = new Set(selectedNonSubscribers);
				if (checked) {
					newSelected.add(itemId);
				} else {
					newSelected.delete(itemId);
				}
				selectedNonSubscribers = newSelected;
				selectAllNonSubscribers = newSelected.size === currentDisplayedItems.length;
			}
		}
		
		console.log('Selection state updated:', {
			activeTab,
			selectedCount: activeTab === 'subscribers' ? selectedSubscribers.size : selectedNonSubscribers.size,
			displayedCount: currentDisplayedItems.length,
			selectAll: activeTab === 'subscribers' ? selectAllSubscribers : selectAllNonSubscribers
		});
	}

	// Separate function for NonSubscriberTable callback props
	function handleNonSubscriberSelectItem(itemId: number, checked: boolean) {
		console.log('handleNonSubscriberSelectItem called:', { itemId, checked });
		
		if (itemId === -1) {
			// Header checkbox clicked - toggle all
			if (checked) {
				selectedNonSubscribers = new Set(currentDisplayedItems.map(item => item.id));
				selectAllNonSubscribers = true;
			} else {
				selectedNonSubscribers = new Set();
				selectAllNonSubscribers = false;
			}
		} else {
			// Individual item checkbox clicked
			const newSelected = new Set(selectedNonSubscribers);
			if (checked) {
				newSelected.add(itemId);
			} else {
				newSelected.delete(itemId);
			}
			selectedNonSubscribers = newSelected;
			selectAllNonSubscribers = newSelected.size === currentDisplayedItems.length;
		}
		
		console.log('Non-subscriber selection state updated:', {
			selectedCount: selectedNonSubscribers.size,
			displayedCount: currentDisplayedItems.length,
			selectAll: selectAllNonSubscribers
		});
	}

	// Modal state
	let showSendOfferModal = $state(false);
	let offers: SubscriptionOffer[] = $state([]);
	let offersLoading = false;

	// Load offers for the modal
	async function loadOffers() {
		try {
			offersLoading = true;
			offers = await SubscriptionOfferService.getForSending();
			console.log('Loaded offers for sending:', offers);
		} catch (error) {
			console.error('Error loading offers:', error);
			showToast('Failed to load offers', 'error');
		} finally {
			offersLoading = false;
		}
	}

	// Update handleSendOffer to open modal
	function handleSendOffer() {
		loadOffers();
		showSendOfferModal = true;
	}

	function handleSendOfferCancel() {
		showSendOfferModal = false;
	}

	async function handleSendOfferSubmit(offerId: number, emails: string[]) {
		try {
			const result = await SubscriptionOfferService.sendOffer(offerId, emails);
			showToast(`Offer sent to ${result.emails_sent} users!`, 'success');
			showSendOfferModal = false;
			clearSelection(); // Clear selections after sending
		} catch (error) {
			console.error('Error sending offer:', error);
			showToast('Failed to send offer', 'error');
		}
	}

	function clearSelection() {
		selectedSubscribers = new Set();
		selectedNonSubscribers = new Set();
		selectAllSubscribers = false;
		selectAllNonSubscribers = false;
		console.log('Cleared all selections');
	}

	// The selection state is now managed directly in the handler

	// Handle tab change
	function handleTabChange(event: CustomEvent<{ tab: 'subscribers' | 'non-subscribers' }>) {
		const { tab } = event.detail;
		
		// Determine animation direction
		if (activeTab === 'subscribers' && tab === 'non-subscribers') {
			animationDirection = 'right';
		} else if (activeTab === 'non-subscribers' && tab === 'subscribers') {
			animationDirection = 'left';
		}
		
		activeTab = tab;
		currentPage = 1;
		isTransitioning = true;
		isAnimating = true;
		
		// Apply filters immediately
		debounceApplyFilters();
		
		// End transition after brief period
		setTimeout(() => {
			isTransitioning = false;
		}, 300);
		
		// End animation after content loads
		setTimeout(() => {
			isAnimating = false;
		}, 600);
	}


	// Update the changeTab function to handle the new tab
	async function changeTab(tab: 'subscribers' | 'non-subscribers' | 'stripe-subs') {
		if (tab === activeTab) return;
		
		// Clear selections when switching tabs
		selectedSubscribers.clear();
		selectedNonSubscribers.clear();
		selectAllSubscribers = false;
		selectAllNonSubscribers = false;
		
		// Reset pagination when switching tabs
		currentPage = 1;
		
		// Set animation direction based on tab order
		if (activeTab === 'subscribers' && tab === 'non-subscribers') {
			animationDirection = 'right';
		} else if (activeTab === 'non-subscribers' && tab === 'subscribers') {
			animationDirection = 'left';
		} else if (tab === 'stripe-subs') {
			// New tab - determine direction based on current tab
			animationDirection = activeTab === 'subscribers' ? 'right' : 'left';
		}
		
		activeTab = tab;
		
		// Load data for the new tab immediately
		await loadTabData(tab);
	}

	// Handle search
	function handleSearch(term: string) {
		if (activeTab === 'subscribers') {
			subscriberSearchTerm = term;
		} else {
			nonSubscriberSearchTerm = term;
		}
		currentPage = 1;
		debounceApplyFilters();
	}

	// Handle filter changes
	function handleFilterChange(type: 'emailVerified' | 'role' | 'plan' | 'lastLogin' | 'createdDate' | 'hasSubbed', value: any) {
		console.log('🎯 Filter change received:', { type, value });
		
		if (activeTab === 'subscribers') {
			switch (type) {
				case 'emailVerified':
					subscriberEmailVerifiedFilter = value;
					break;
				case 'role':
					subscriberRoleFilter = value;
					break;
				case 'plan':
					subscriberPlanFilter = value;
					break;
				case 'lastLogin':
					console.log('📅 Setting subscriber last login filter:', value);
					subscriberLastLoginFilter = value;
					break;
				case 'createdDate':
					console.log('📅 Setting subscriber created date filter:', value);
					subscriberCreatedDateFilter = value;
					break;
			}
		} else {
			switch (type) {
				case 'emailVerified':
					nonSubscriberEmailVerifiedFilter = value;
					break;
				case 'role':
					nonSubscriberRoleFilter = value;
					break;
				case 'lastLogin':
					console.log('📅 Setting non-subscriber last login filter:', value);
					nonSubscriberLastLoginFilter = value;
					break;
				case 'createdDate':
					console.log('📅 Setting non-subscriber created date filter:', value);
					nonSubscriberCreatedDateFilter = value;
					break;
				case 'hasSubbed':
					console.log('📅 Setting non-subscriber has subbed filter:', value);
					nonSubscriberHasSubbedFilter = value;
					break;
			}
		}
		currentPage = 1;
		debounceApplyFilters();
	}

	// Handle clear all filters
	function handleClearAllFilters() {
		if (activeTab === 'subscribers') {
			subscriberSearchTerm = '';
			subscriberEmailVerifiedFilter = undefined;
			subscriberRoleFilter = '';
			subscriberPlanFilter = '';
			subscriberLastLoginFilter = '';
			subscriberCreatedDateFilter = '';
		} else {
			nonSubscriberSearchTerm = '';
			nonSubscriberEmailVerifiedFilter = undefined;
			nonSubscriberRoleFilter = '';
			nonSubscriberLastLoginFilter = '';
			nonSubscriberCreatedDateFilter = '';
			nonSubscriberHasSubbedFilter = undefined;
		}
		currentPage = 1;
		debounceApplyFilters();
	}

	// Handle page change
	function handlePageChange(event: CustomEvent<{ page: number }>) {
		currentPage = event.detail.page;
	}

	// Helper to get current filters for active tab
	function getCurrentFilters() {
		if (activeTab === 'subscribers') {
			return {
				searchTerm: subscriberSearchTerm,
				emailVerifiedFilter: subscriberEmailVerifiedFilter,
				roleFilter: subscriberRoleFilter,
				planFilter: subscriberPlanFilter,
				lastLoginFilter: subscriberLastLoginFilter,
				createdDateFilter: subscriberCreatedDateFilter,
			};
		} else {
			return {
				searchTerm: nonSubscriberSearchTerm,
				emailVerifiedFilter: nonSubscriberEmailVerifiedFilter,
				roleFilter: nonSubscriberRoleFilter,
				lastLoginFilter: nonSubscriberLastLoginFilter,
				createdDateFilter: nonSubscriberCreatedDateFilter,
				hasSubbedFilter: nonSubscriberHasSubbedFilter,
			};
		}
	}

	// Handle non-subscriber update from edit modal
	async function handleNonSubscriberUpdate(updatedNonSubscriber: NonSubscriber) {
		try {
			// Update the non-subscriber in our local arrays
			allNonSubscribers = allNonSubscribers.map(ns => 
				ns.id === updatedNonSubscriber.id ? updatedNonSubscriber : ns
			);
			displayedNonSubscribers = displayedNonSubscribers.map(ns => 
				ns.id === updatedNonSubscriber.id ? updatedNonSubscriber : ns
			);
			
			showToast('Non-subscriber updated successfully', 'success');
		} catch (error) {
			console.error('Error updating non-subscriber:', error);
			showToast('Failed to update non-subscriber', 'error');
		}
	}

	// Bulk actions
	async function handleBulkSuspend() {
		if (currentSelectedCount === 0) {
			showToast('Please select subscribers to suspend', 'warning');
			return;
		}

		try {
			const selectedIds = Array.from(currentSelectedItems);
			const result = await StreamingSubscriberService.bulkSuspendSubscribers(selectedIds);
			
			if (result.successful && result.successful.length > 0) {
				showToast(`Suspended ${result.successful.length} subscribers`, 'success');
				clearSelection();
				await loadTabData(activeTab); // Refresh current tab data
			}
			
			if (result.failed && result.failed.length > 0) {
				showToast(`Failed to suspend ${result.failed.length} subscribers`, 'error');
			}
		} catch (error) {
			console.error('Error bulk suspending subscribers:', error);
			showToast('Failed to suspend subscribers', 'error');
		}
	}

	async function handleBulkActivate() {
		if (currentSelectedCount === 0) {
			showToast('Please select subscribers to activate', 'warning');
			return;
		}

		try {
			const selectedIds = Array.from(currentSelectedItems);
			const result = await StreamingSubscriberService.bulkActivateSubscribers(selectedIds);
			
			if (result.successful && result.successful.length > 0) {
				showToast(`Activated ${result.successful.length} subscribers`, 'success');
				clearSelection();
				await loadTabData(activeTab); // Refresh current tab data
			}
			
			if (result.failed && result.failed.length > 0) {
				showToast(`Failed to activate ${result.failed.length} subscribers`, 'error');
			}
		} catch (error) {
			console.error('Error bulk activating subscribers:', error);
			showToast('Failed to activate subscribers', 'error');
		}
	}

	async function handleBulkChangePlan() {
		if (currentSelectedCount === 0) {
			showToast('Please select subscribers to change plan', 'warning');
			return;
		}

		// TODO: Implement plan selection modal
		showToast('Bulk plan change feature coming soon', 'info');
	}

	async function handleExport() {
		try {
			let blob: Blob;
			let filename: string;
			
			if (activeTab === 'subscribers') {
				blob = await StreamingSubscriberService.exportSubscribers('csv');
				filename = `subscribers-${new Date().toISOString().split('T')[0]}.csv`;
			} else {
				blob = await StreamingSubscriberService.exportNonSubscribers('csv');
				filename = `non-subscribers-${new Date().toISOString().split('T')[0]}.csv`;
			}
			
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = filename;
			document.body.appendChild(a);
			a.click();
			window.URL.revokeObjectURL(url);
			document.body.removeChild(a);
			
			const tabName = activeTab === 'subscribers' ? 'Subscribers' : 'Non-subscribers';
			showToast(`${tabName} exported successfully`, 'success');
		} catch (error) {
			console.error('Error exporting data:', error);
			const tabName = activeTab === 'subscribers' ? 'subscribers' : 'non-subscribers';
			showToast(`Failed to export ${tabName}`, 'error');
		}
	}

	// Handle edit subscriber event from table
	function handleEditSubscriber(event: CustomEvent<{ subscriber: Subscriber }>) {
		const { subscriber } = event.detail;
		selectedSubscriber = subscriber;
		showEditModal = true;
	}

	// Handle subscriber update from edit modal
	async function handleSubscriberUpdate(updatedSubscriber: Subscriber) {
		try {
			// Update the subscriber in our local arrays
			allSubscribers = allSubscribers.map(s => 
				s.id === updatedSubscriber.id ? updatedSubscriber : s
			);
			displayedSubscribers = displayedSubscribers.map(s => 
				s.id === updatedSubscriber.id ? updatedSubscriber : s
			);
			
			showToast('Subscriber updated successfully', 'success');
			showEditModal = false;
			selectedSubscriber = null;
		} catch (error) {
			console.error('Error updating subscriber:', error);
			showToast('Failed to update subscriber', 'error');
		}
	}

	// Add new state variables for Stripe customers tab
	let stripeOnlyCustomers = $state<any[]>([]);
	let syncedCustomers = $state<any[]>([]);
	let localOnlyUsers = $state<any[]>([]);
	let syncingCustomers = $state(new Set<string>());
	let bulkCreatingUsers = $state(false);

	// Add stats for Stripe tab
	let totalCount = $state(0);
	let syncedCount = $state(0);
	let localOnlyCount = $state(0);
	let stripeOnlyCount = $state(0);



</script>

<svelte:head>
	<title>Subscribers - BOME Admin</title>
</svelte:head>

<div class="subscribers-page">
	<!--<div class="page-header">
		<div class="header-content">
			<h1 class="page-title">Subscribers Management</h1>
			<p class="page-description">Manage subscribers and non-subscribers</p>
		</div>
	</div>-->

	<!-- Tab Navigation -->
	<div class="tab-navigation">
		<button 
			class="tab-button" 
			class:active={activeTab === 'subscribers'}
			onclick={() => changeTab('subscribers')}
		>
			<span class="tab-icon">👥</span>
			Subscribers ({displayedSubscribers.length})
		</button>
		<button 
			class="tab-button" 
			class:active={activeTab === 'stripe-subs'}
			onclick={() => changeTab('stripe-subs')}
		>
			<span class="tab-icon">🔗</span>
			Stripe Subs <!--({stripeOnlyCount + syncedCount})-->
		</button>
	</div>

	<!-- Tab Content -->
	<div class="tab-content">
	{#if activeTab === 'subscribers'}
		<!-- Enhanced Subscribers Tab with Email Verification Accordions -->
		<EnhancedSubscribersPage />
		{:else if activeTab === 'stripe-subs'}
			<!-- Stripe Subscribers Tab -->
			<div class="stripe-subs-section">
				{#if stripeSubsLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading Stripe customers and subscriptions...</p>
					</div>
				{:else}
					<!-- Pass the required props to the customers component -->
					<StripeCustomers 
						summary={{ 
							total_customers: stripeCustomers.length,
							total_subscriptions: stripeSubscriptions.length 
						}} 
						stripeData={{ 
							customers: stripeCustomers,
							subscriptions: stripeSubscriptions 
						}} 
					/>
				{/if}
			</div>
		{/if}
	</div>
</div>

<SendOfferModal 
	bind:isOpen={showSendOfferModal}
	selectedEmails={activeTab === 'subscribers' 
		? allSubscribers.filter(s => selectedSubscribers.has(s.id)).map(s => s.email)
		: allNonSubscribers.filter(ns => selectedNonSubscribers.has(ns.id)).map(ns => ns.email)
	}
	{offers}
	onSendOffer={handleSendOfferSubmit}
	onCancel={handleSendOfferCancel}
/>

{#if showEditModal && selectedSubscriber}
	<SubscriberEditModal 
		bind:isOpen={showEditModal}
		subscriber={selectedSubscriber}
		{subscriptionPlans}
		onSave={handleSubscriberUpdate}
		onCancel={() => { showEditModal = false; selectedSubscriber = null; }}
	/>
{/if}

<style>
	/* CSS Variables for child components */
	:global(:root) {
		--space-xs: 0.25rem;
		--space-sm: 0.5rem;
		--space-md: 1rem;
		--space-lg: 1.5rem;
		--space-xl: 2rem;
		--radius-sm: 0.25rem;
		--radius-md: 0.375rem;
		--radius-lg: 0.5rem;
		--text: #111827;
		--text-muted: #6b7280;
		--primary: #2563eb;
		--primary-hover: #1d4ed8;
		--surface: #ffffff;
		--bg-secondary: #f9fafb;
		--bg-hover: #f3f4f6;
		--border: #e5e7eb;
		--error: #dc2626;
		--success: #059669;
		--warning: #d97706;
		--font-mono: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
	}

	.subscribers-page {
		max-width: 100%;
		margin: 0 auto;
		padding: 0 0;
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		padding-bottom: 1rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.page-title {
		font-size: 2rem;
		font-weight: 700;
		color: #111827;
		margin: 0;
	}

	.page-description {
		color: #6b7280;
		margin: 0.5rem 0 0 0;
	}

	/* Tab Navigation */
	.tab-navigation {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 2rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		background: none;
		border: none;
		border-bottom: 2px solid transparent;
		color: #6b7280;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
	}

	.tab-button:hover {
		color: #374151;
		background-color: #f9fafb;
	}

	.tab-button.active {
		color: #2563eb;
		border-bottom-color: #2563eb;
	}

	.tab-icon {
		font-size: 1.1rem;
	}

	/* Tab Content */
	.tab-content {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		overflow: hidden;
		min-height: 50vh;
	}

	.subscribers-section,
	.stripe-subs-section {
		padding: 1.5rem;
	}

	.loading-container {
		text-align: center;
		padding: 3rem 0;
	}

	/* Selection Controls */
	.selection-actions {
		display: flex;
		flex-direction: row;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 1.5rem;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
	}

	.selection-button {
		height: 50px;
	}

	.selected-count {
		font-size: 0.875rem;
		color: #6b7280;
		font-weight: 500;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
	}

	.btn-outline {
		background: transparent;
		color: #2563eb;
		border: 1px solid #2563eb;
	}

	.btn-outline:hover {
		background: #2563eb;
		color: white;
	}

	.btn-warning {
		background: #d97706;
		color: white;
	}

	.btn-warning:hover {
		background: #b45309;
	}

	.btn-success {
		background: #059669;
		color: white;
	}

	.btn-success:hover {
		background: #047857;
	}

	.btn-info {
		background: #0891b2;
		color: white;
	}

	.btn-info:hover {
		background: #0e7490;
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
	}

	.btn-ghost {
		background: transparent;
		color: #6b7280;
		border: 1px solid transparent;
	}

	.btn-ghost:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.btn-error {
		background: #dc2626;
		color: white;
	}

	.btn-error:hover {
		background: #b91c1c;
	}

	.btn-danger {
		background: #dc2626;
		color: white;
	}

	.btn-danger:hover {
		background: #b91c1c;
	}

	.btn-small {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
	}

	.btn-lg {
		padding: 0.75rem 1.5rem;
		font-size: 1.125rem;
	}

	.btn-full {
		width: 100%;
		justify-content: center;
	}

	.btn-bottom {
		margin-top: auto;
	}

	.btn-disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-tag {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-tag:hover {
		background: #e5e7eb;
	}

	@media (max-width: 768px) {
		.subscribers-page {
			padding: 1rem;
		}

		.page-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		.tab-navigation {
			flex-wrap: wrap;
		}
	}
</style> 
