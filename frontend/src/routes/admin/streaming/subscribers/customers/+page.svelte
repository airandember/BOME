<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import CustomerSyncPanel from './components/CustomerSyncPanel.svelte';
	import SimpleTable from './SimpleTable.svelte';
	import { StreamingSubscriberService } from '$lib/services/streaming-subscribers';
	import { showToast } from '$lib/toast';
	import StripeResyncModal from '$lib/components/StripeResyncModal.svelte';

	// State variables
	let loading = $state(true);
	let error = $state('');
	
	// Data arrays for different sync states
	let stripeOnlyCustomers = $state<any[]>([]);
	let syncedCustomers = $state<any[]>([]);
	let localOnlyUsers = $state<any[]>([]);
 // New: Multiple Stripe IDs for same email
	
	// Sync state
	let syncingCustomers = $state(new Set<string>());
	let bulkCreatingUsers = $state(false);
	
	// Stats
	let totalCount = $state(0);
	let syncedCount = $state(0);
	let localOnlyCount = $state(0);
	let stripeOnlyCount = $state(0);

	// Resync modal state
	let showResyncModal = $state(false);


	// Pagination for Stripe Only customers
	let stripeOnlyCurrentPage = $state(1);
	let stripeOnlyItemsPerPage = $state(50);
	let stripeOnlyTotalPages = $state(1);
	let stripeOnlyDisplayed = $state<any[]>([]);
	
	// All stripe only customers (unpaginated)
	let allStripeOnlyCustomers = $state<any[]>([]);

	// Pagination and search for Synced customers
	// All synced customers (no pagination needed - handled by SimpleTable)
	let allSyncedCustomers = $state<any[]>([]);

	// Accept data from parent - NO API calls needed!
	const { summary: parentSummary = null, stripeData = null } = $props<{ 
		summary?: any, 
		stripeData?: any 
	}>();

	// Local state variables that were missing
	let summary = $state<any>(null);
	let customers = $state<any[]>([]);

	onMount(async () => {
		if (parentSummary && stripeData) {
			// Use pre-loaded data from parent
			summary = parentSummary;
			customers = stripeData.customers || [];
			
			// Load local users to compare against Stripe customers
			const usersRes = await apiRequest('/admin/users?limit=3000');
			let localUsers: any[] = [];
			if (usersRes.ok) {
				const usersData = await usersRes.json();
				localUsers = usersData.users || [];
				console.log('✅ Loaded local users for matching:', localUsers.length);
			} else {
				console.warn('⚠️ Users endpoint not available (404), using subscriber data as fallback');
				
				// Also fetch subscribers to get subscription plan information
				const subscribersData = await StreamingSubscriberService.getSubscribers({ limit: 1000 });
				const subscribers = subscribersData.subscribers || [];
				console.log('✅ Loaded subscribers for plan data:', subscribers.length);

				// Use subscribers as the base data source
				localUsers = subscribers.map((sub: any) => ({
					ID: sub.id,
					id: sub.id,
					Email: sub.email,
					email: sub.email,
					FirstName: sub.first_name,
					first_name: sub.first_name,
					LastName: sub.last_name,
					last_name: sub.last_name,
					Role: sub.role,
					role: sub.role,
					StripeCustomerID: sub.stripe_customer_id,
					stripe_customer_id: sub.stripe_customer_id,
					CreatedAt: sub.created_at,
					created_at: sub.created_at,
					plan_name: sub.plan_name,
					stripe_price_id: sub.stripe_price_id,
					subscription_id: sub.subscription_id
				}));
			}

			// Create customer lists using the pre-loaded Stripe data
			createCustomerLists(stripeData.customers || [], localUsers);
			loading = false;
			console.log('✅ Customers: Using pre-loaded data from parent');
		} else {
			// Fallback to loading data directly (shouldn't happen in normal flow)
			console.warn('⚠️ No data from parent, loading directly...');
			await loadAllData();
		}
	});

	// Load both local users and Stripe customers from database
	async function loadAllData() {
		try {
			loading = true;
			error = '';

			console.log('🔄 Loading Stripe customers from database and checking against local users...');

			// First, fetch all Stripe customers from database (primary source)
			const stripeRes = await apiRequest('/admin/streaming/stripe/database/customers?limit=1000&include_subscriptions=true');
			let allStripeCustomers: any[] = [];
			if (stripeRes.ok) {
				const stripeData = await stripeRes.json();
				allStripeCustomers = stripeData.customers || [];
				console.log('✅ Loaded Stripe customers from database:', allStripeCustomers.length);
			
				// Debug: Show first customer structure
				if (allStripeCustomers.length > 0) {
					console.log('🔍 Sample Stripe customer from database:', {
						stripe_id: allStripeCustomers[0].stripe_id,
						email: allStripeCustomers[0].email,
						name: allStripeCustomers[0].name,
						subscriptions: allStripeCustomers[0].subscriptions?.length || 0
					});
				} else {
					console.warn('⚠️ No customers found in stripe_customers database table');
					console.warn('💡 You may need to run a Stripe sync first to populate the database');
				}
			} else {
				console.error('❌ Failed to load Stripe customers from database');
				console.error('Response status:', stripeRes.status);
				try {
					const errorText = await stripeRes.text();
					console.error('Response text:', errorText);
				} catch (e) {
					console.error('Could not read response text');
				}
				allStripeCustomers = [];
			}

			// Then, fetch all local users to check against
			const usersRes = await apiRequest('/admin/users?limit=3000');
			let localUsers: any[] = [];
			if (usersRes.ok) {
				const usersData = await usersRes.json();
				localUsers = usersData.users || [];
				console.log('✅ Loaded local users for matching:', localUsers.length);
				
				// 🔍 DEBUG: Check if users have stripe_customer_ids arrays
				const usersWithArrays = localUsers.filter(user => 
					Array.isArray(user.StripeCustomerIDs || user.stripe_customer_ids) && 
					(user.StripeCustomerIDs || user.stripe_customer_ids).length > 0
				).length;
				console.log(`🔍 Users with stripe_customer_ids arrays: ${usersWithArrays}/${localUsers.length}`);
			} else {
				console.warn('⚠️ Users endpoint not available (404), using subscriber data as fallback');
				console.warn('Response status:', usersRes.status);
			}

			// Also fetch subscribers to get subscription plan information
			const subscribersData = await StreamingSubscriberService.getSubscribers({ limit: 1000 });
			const subscribers = subscribersData.subscribers || [];
			console.log('✅ Loaded subscribers for plan data:', subscribers.length);

			// If users endpoint failed, use subscribers as the base data source
			if (localUsers.length === 0 && subscribers.length > 0) {
				console.log('🔄 Using subscribers as primary data source');
				localUsers = subscribers.map((sub: any) => ({
					ID: sub.id,
					id: sub.id,
					Email: sub.email,
					email: sub.email,
					FirstName: sub.first_name,
					first_name: sub.first_name,
					LastName: sub.last_name,
					last_name: sub.last_name,
					Role: sub.role,
					role: sub.role,
					StripeCustomerID: sub.stripe_customer_id,
					stripe_customer_id: sub.stripe_customer_id,
					// 🔥 CRITICAL: Include the stripe_customer_ids array!
					StripeCustomerIDs: sub.stripe_customer_ids || [],
					stripe_customer_ids: sub.stripe_customer_ids || [],
					CreatedAt: sub.created_at,
					created_at: sub.created_at,
					plan_name: sub.plan_name,
					stripe_price_id: sub.stripe_price_id,
					subscription_id: sub.subscription_id
				}));
			} else {
				// Merge user data with subscriber plan data
				const subscribersByEmail = new Map();
				subscribers.forEach((sub: any) => {
					subscribersByEmail.set(sub.email.toLowerCase(), sub);
				});

				// Enhance local users with subscription plan data
				localUsers = localUsers.map((user: any) => {
					const email = (user.Email || user.email || '').toLowerCase();
					const subscriber = subscribersByEmail.get(email);
					if (subscriber) {
						return {
							...user,
							plan_name: subscriber.plan_name,
							stripe_price_id: subscriber.stripe_price_id,
							subscription_id: subscriber.subscription_id
						};
					}
					return user;
				});
			}

			// Create separate customer lists
			createCustomerLists(allStripeCustomers, localUsers);

		} catch (err) {
			error = 'Failed to load data';
			console.error('❌ Error loading data:', err);
		} finally {
			loading = false;
		}
	}

	// Create separate customer lists for different sync states
	function createCustomerLists(allStripeCustomers: any[], localUsers: any[]) {
		console.log('🔄 Creating customer lists...');
		console.log(`📊 Input data: ${allStripeCustomers.length} Stripe customers, ${localUsers.length} local users`);
		
		// 🔍 DEBUG: Sample the first few users to see their structure
		console.log('🔍 Sample local users (first 3):');
		localUsers.slice(0, 3).forEach((user, index) => {
			const primaryStripeId = user.StripeCustomerID || user.stripe_customer_id;
			console.log(`User ${index + 1}:`, {
				email: user.Email || user.email,
				primary_stripe_id_raw: primaryStripeId,
				primary_stripe_id_type: typeof primaryStripeId,
				primary_stripe_id_valid: primaryStripeId?.Valid,
				primary_stripe_id_string: primaryStripeId?.String,
				stripe_ids_array: user.StripeCustomerIDs || user.stripe_customer_ids,
				array_type: typeof (user.StripeCustomerIDs || user.stripe_customer_ids),
				array_length: Array.isArray(user.StripeCustomerIDs || user.stripe_customer_ids) ? (user.StripeCustomerIDs || user.stripe_customer_ids).length : 'not array'
			});
		});
		
		console.log('🔍 Sample Stripe customers (first 3):');
		allStripeCustomers.slice(0, 3).forEach((customer, index) => {
			console.log(`Stripe Customer ${index + 1}:`, {
				email: customer.email || customer.Email,
				stripe_id: customer.stripe_id || customer.ID
			});
		});
		
		let totalStripeIdsMapped = 0;

		// Create maps for fast lookup
		const usersByEmail = new Map();
		const usersByStripeId = new Map();
		const stripeCustomersByEmail = new Map();
		const stripeCustomersById = new Map();
		
		// Map local users
		let usersWithArrays = 0;
		let usersWithPrimary = 0;
		localUsers.forEach((user: any, index) => {
			const email = (user.Email || user.email || '').toLowerCase();
			if (email) {
				usersByEmail.set(email, user);
			}
			
			// Map primary Stripe ID - handle sql.NullString object
			let primaryStripeId = user.StripeCustomerID || user.stripe_customer_id;
			
			// Handle sql.NullString object from backend
			if (primaryStripeId && typeof primaryStripeId === 'object') {
				if (primaryStripeId.Valid && primaryStripeId.String) {
					primaryStripeId = primaryStripeId.String;
				} else {
					primaryStripeId = null;
				}
			}
			
			if (primaryStripeId && typeof primaryStripeId === 'string') {
				usersByStripeId.set(primaryStripeId, user);
				totalStripeIdsMapped++;
				usersWithPrimary++;
				
				// Debug first few mappings
				if (index < 3) {
					console.log(`🔗 Mapped primary Stripe ID: ${primaryStripeId} -> ${email}`);
				}
			}
			
			// 🔥 CRITICAL: Also map ALL Stripe IDs from the array
			const allStripeIds = user.StripeCustomerIDs || user.stripe_customer_ids || [];
			if (Array.isArray(allStripeIds) && allStripeIds.length > 0) {
				usersWithArrays++;
				allStripeIds.forEach((stripeId: string) => {
					if (stripeId && stripeId.trim()) {
						usersByStripeId.set(stripeId.trim(), user);
						totalStripeIdsMapped++;
						
						// Debug first few array mappings
						if (index < 3) {
							console.log(`🔗 Mapped array Stripe ID: ${stripeId.trim()} -> ${email}`);
						}
					}
				});
			}
		});
		
		console.log(`📊 Mapping summary: ${usersWithPrimary} users with primary IDs, ${usersWithArrays} users with ID arrays`);
		
		// 🔍 DEBUG: Sample what's actually in our Maps
		console.log('🔍 Sample usersByStripeId entries (first 5):');
		let count = 0;
		for (const [stripeId, user] of usersByStripeId) {
			if (count < 5) {
				console.log(`  ${stripeId} -> ${user.Email || user.email}`);
				count++;
			} else {
				break;
			}
		}

		// Map Stripe customers (from database format)
		allStripeCustomers.forEach((customer: any) => {
			const email = (customer.email || customer.Email)?.toLowerCase();
			if (email) {
				stripeCustomersByEmail.set(email, customer);
			}
			// Use stripe_id from database instead of ID from API
			const stripeId = customer.stripe_id || customer.ID;
			if (stripeId) {
				stripeCustomersById.set(stripeId, customer);
			}
		});

		// 1. Stripe Only Customers - exist in Stripe_customers database but not locally
		let filteredOut = 0;
		let keptAsStripeOnly = 0;
		
		allStripeOnlyCustomers = allStripeCustomers.filter((stripeCustomer: any) => {
			const email = (stripeCustomer.email || stripeCustomer.Email)?.toLowerCase();
			const stripeId = stripeCustomer.stripe_id || stripeCustomer.ID;
			
			// Check if this Stripe customer has a local user
			const localUserByStripeId = usersByStripeId.get(stripeId);
			const localUserByEmail = usersByEmail.get(email);
			const hasLocalUser = !!(localUserByStripeId || localUserByEmail);
			
			// Debug logging for first few customers
			if (allStripeCustomers.indexOf(stripeCustomer) < 10) {
				console.log(`🔍 Customer ${stripeId} (${email}):`, {
					hasLocalUser,
					foundByStripeId: !!localUserByStripeId,
					foundByEmail: !!localUserByEmail,
					localUserEmail: localUserByStripeId?.Email || localUserByEmail?.Email || 'none'
				});
			}
			
			if (hasLocalUser) {
				filteredOut++;
				return false; // Filter out - user exists locally
			} else {
				keptAsStripeOnly++;
				return true; // Keep as Stripe-only
			}
		}).map((customer: any) => ({
			id: `stripe_${customer.stripe_id || customer.ID}`,
			source: 'stripe',
			name: customer.name || customer.Name || 'Unnamed Customer',
			email: customer.email || customer.Email,
			localId: null,
			role: null,
			planName: null,
			stripePriceId: null,
			stripeId: customer.stripe_id || customer.ID,
			stripeCreatedAt: customer.created_at || customer.CreatedAt,
			stripeMetadata: customer.metadata || customer.Metadata,
			createdAt: customer.created_at || customer.CreatedAt,
			stripeCustomerId: customer.stripe_id || customer.ID,
			subscriptions: customer.subscriptions || [],
			syncStatus: 'stripe_only'
		}));

		// Set up pagination for Stripe Only customers
		stripeOnlyCount = allStripeOnlyCustomers.length;
		stripeOnlyTotalPages = Math.ceil(stripeOnlyCount / stripeOnlyItemsPerPage);
		stripeOnlyCurrentPage = 1; // Reset to first page
		updateStripeOnlyPagination();
		
		console.log(`🔍 Filtering results:`, {
			totalStripeIdsMapped,
			filteredOut,
			keptAsStripeOnly,
			finalStripeOnlyCount: stripeOnlyCount,
			usersByStripeIdSize: usersByStripeId.size,
			usersByEmailSize: usersByEmail.size
		});
		
		// 🔍 DEBUG: Check if we're finding matches for the first few Stripe customers
		//console.log('🔍 Testing first 5 Stripe customers for matches:');
		//allStripeCustomers.slice(0, 5).forEach((customer, index) => {
		//	const email = (customer.email || customer.Email)?.toLowerCase();
		//	const stripeId = customer.stripe_id || customer.ID;
		//	const foundByStripeId = usersByStripeId.get(stripeId);
		//	const foundByEmail = usersByEmail.get(email);
		//	console.log(`Customer ${index + 1} (${stripeId}):`, {
		//		email,
		//		foundByStripeId: !!foundByStripeId,
		//		foundByEmail: !!foundByEmail,
		//		shouldBeFiltered: !!(foundByStripeId || foundByEmail)
		//	});
		//});

		// 2. Synced Customers - exist in both Stripe and locally
		syncedCustomers = allStripeCustomers.filter((stripeCustomer: any) => {
			const email = (stripeCustomer.email || stripeCustomer.Email)?.toLowerCase();
			const stripeId = stripeCustomer.stripe_id || stripeCustomer.ID;
			
			// Check if this Stripe customer has a local user
			const localUser = usersByStripeId.get(stripeId) || usersByEmail.get(email);
			return !!localUser; // Only include if local user found
		}).map((stripeCustomer: any) => {
			const email = (stripeCustomer.email || stripeCustomer.Email)?.toLowerCase();
			const stripeId = stripeCustomer.stripe_id || stripeCustomer.ID;
			const localUser = usersByStripeId.get(stripeId) || usersByEmail.get(email);
			
			return {
				id: `hybrid_${stripeId}`,
				source: 'hybrid',
				name: stripeCustomer.name || stripeCustomer.Name || `${localUser.FirstName || localUser.first_name || ''} ${localUser.LastName || localUser.last_name || ''}`.trim(),
				email: stripeCustomer.email || stripeCustomer.Email,
				localId: localUser.ID || localUser.id,
				role: localUser.Role || localUser.role,
				planName: localUser.plan_name || null,
				stripePriceId: localUser.stripe_price_id || null,
				stripeId: stripeId,
				stripeCreatedAt: stripeCustomer.created_at || stripeCustomer.CreatedAt,
				stripeMetadata: stripeCustomer.metadata || stripeCustomer.Metadata,
				createdAt: localUser.CreatedAt || localUser.created_at,
				stripeCustomerId: localUser.StripeCustomerID || localUser.stripe_customer_id,
				subscriptions: stripeCustomer.subscriptions || [],
				syncStatus: 'synced'
			};
		});

		// Store all synced customers (pagination handled by SimpleTable)
		allSyncedCustomers = [...syncedCustomers];

		// 3. Local Only Users - exist locally but not in Stripe
		localOnlyUsers = localUsers.filter((localUser: any) => {
			const email = (localUser.Email || localUser.email || '').toLowerCase();
			const stripeId = localUser.StripeCustomerID || localUser.stripe_customer_id;
			
			// Check if this local user has a Stripe customer
			const stripeCustomer = stripeCustomersById.get(stripeId) || stripeCustomersByEmail.get(email);
			return !stripeCustomer; // Only include if no Stripe customer found
		}).map((user: any) => ({
			id: `local_${user.ID || user.id}`,
			source: 'local',
			name: `${user.FirstName || user.first_name || ''} ${user.LastName || user.last_name || ''}`.trim(),
			email: user.Email || user.email,
			localId: user.ID || user.id,
			role: user.Role || user.role,
			planName: user.plan_name || null,
			stripePriceId: user.stripe_price_id || null,
			stripeId: null,
			stripeCreatedAt: null,
			stripeMetadata: null,
			createdAt: user.CreatedAt || user.created_at,
			stripeCustomerId: user.StripeCustomerID || user.stripe_customer_id,
			syncStatus: 'local_only'
		}));



		// Calculate stats
		totalCount = allStripeCustomers.length + localOnlyUsers.length;
		syncedCount = syncedCustomers.length;
		localOnlyCount = localOnlyUsers.length;
		stripeOnlyCount = stripeOnlyCustomers.length;


		console.log('✅ Customer lists created:', {
			total: totalCount,
			synced: syncedCount,
			localOnly: localOnlyCount,
			stripeOnly: stripeOnlyCount
		});
		
		console.log('📋 Breakdown:');
		console.log(`  • ${allStripeCustomers.length} customers from stripe_customers table`);
		console.log(`  • ${localUsers.length} local users`);
		console.log(`  • ${syncedCustomers.length} synced (in both Stripe DB and local)`);
		console.log(`  • ${stripeOnlyCustomers.length} Stripe-only (in Stripe DB but not local)`);
		console.log(`  • ${localOnlyUsers.length} local-only (local but not in Stripe DB)`);

	}

	// Create user directly from Stripe customer data (inline)
	async function createUserFromStripe(customer: any) {
		if (!customer.stripeId) {
			showToast('Invalid Stripe customer ID', 'error');
			return;
		}

		if (syncingCustomers.has(customer.id)) return;

		try {
			// Svelte 5 immutable Set update
			syncingCustomers = new Set([...syncingCustomers, customer.id]);

			// Parse name into first and last name with better fallbacks
			let firstName = '';
			let lastName = '';
			
			if (customer.name && customer.name.trim()) {
				const cleanName = customer.name.trim();
				const nameParts = cleanName.split(' ').filter((part: any) => part.length > 0);
				
				if (nameParts.length === 1) {
					firstName = nameParts[0];
					lastName = 'Unknown'; // Provide fallback
				} else if (nameParts.length >= 2) {
					firstName = nameParts[0];
					lastName = nameParts.slice(1).join(' ');
				}
			} else {
				// If no name provided, try to extract from email
				const emailParts = customer.email.split('@')[0];
				const emailName = emailParts.replace(/[._-]/g, ' ');
				const emailNameParts = emailName.split(' ').filter((part: any) => part.length > 0);
				
				if (emailNameParts.length >= 2) {
					firstName = emailNameParts[0];
					lastName = emailNameParts.slice(1).join(' ');
				} else {
					firstName = emailNameParts[0] || 'User';
					lastName = 'Unknown';
				}
			}
			
			// Clean up names - remove numbers and special characters
			firstName = firstName.replace(/[0-9._@-]/g, '').trim() || 'User';
			lastName = lastName.replace(/[0-9._@-]/g, '').trim() || 'Unknown';
			
			// Capitalize names
			firstName = firstName.charAt(0).toUpperCase() + firstName.slice(1).toLowerCase();
			lastName = lastName.charAt(0).toUpperCase() + lastName.slice(1).toLowerCase();

			// Create user with Stripe customer data
			const userData = {
				first_name: firstName,
				last_name: lastName,
				email: customer.email,
				role: 'user', // Default role
				role_id: 'user', // Required role_id for database foreign key
				stripe_customer_id: customer.stripeId,
				email_verified: false,
				is_active: true,
				has_subbed: false
			};

			const response = await apiRequest('/admin/users', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(userData)
			});

			if (response.ok) {
				const responseData = await response.json();
				showToast(`User created successfully! Temporary password: ${responseData.temporary_password}`, 'success');
				
				// Immediately update state - no refresh needed!
				moveCustomerFromStripeOnlyToSynced(customer, responseData.user?.id);
			} else {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to create user');
			}
		} catch (error: any) {
			console.error('Failed to create user from Stripe:', error);
			showToast(error.message || 'Failed to create user', 'error');
		} finally {
			// Svelte 5 immutable Set update
			syncingCustomers = new Set([...syncingCustomers].filter(id => id !== customer.id));
		}
	}

	// Create all users from Stripe-only customers (bulk action)
	async function createAllUsersFromStripe() {
		if (allStripeOnlyCustomers.length === 0) {
			showToast('No Stripe-only customers to create', 'info');
			return;
		}

		if (bulkCreatingUsers) return;

		try {
			bulkCreatingUsers = true;

			// Work with a snapshot of customers to avoid mutation during iteration
			const customersToCreate = [...allStripeOnlyCustomers];
			showToast(`🚀 Creating ${customersToCreate.length} users via bulk API...`, 'info');

			// Prepare bulk user data
			const usersToCreate = customersToCreate.map(customer => {
				// Parse name into first and last name with better fallbacks
				let firstName = '';
				let lastName = '';
				
				if (customer.name && customer.name.trim()) {
					const cleanName = customer.name.trim();
					const nameParts = cleanName.split(' ').filter((part: any) => part.length > 0);
					
					if (nameParts.length === 1) {
						firstName = nameParts[0];
						lastName = 'Unknown'; // Provide fallback
					} else if (nameParts.length >= 2) {
						firstName = nameParts[0];
						lastName = nameParts.slice(1).join(' ');
					}
				} else {
					// If no name provided, try to extract from email
					const emailParts = customer.email.split('@')[0];
					const emailName = emailParts.replace(/[._-]/g, ' ');
					const emailNameParts = emailName.split(' ').filter((part: any) => part.length > 0);
					
					if (emailNameParts.length >= 2) {
						firstName = emailNameParts[0];
						lastName = emailNameParts.slice(1).join(' ');
					} else {
						firstName = emailNameParts[0] || 'User';
						lastName = 'Unknown';
					}
				}
				
				// Clean up names - remove numbers and special characters
				firstName = firstName.replace(/[0-9._@-]/g, '').trim() || 'User';
				lastName = lastName.replace(/[0-9._@-]/g, '').trim() || 'Unknown';
				
				// Capitalize names
				firstName = firstName.charAt(0).toUpperCase() + firstName.slice(1).toLowerCase();
				lastName = lastName.charAt(0).toUpperCase() + lastName.slice(1).toLowerCase();

				return {
					first_name: firstName,
					last_name: lastName,
					email: customer.email,
					role: 'user', // Default role
					role_id: 'user', // Required role_id for database foreign key
					stripe_customer_id: customer.stripeId,
					email_verified: false,
					is_active: true,
					has_subbed: false
				};
			});

			// Make single bulk API call instead of thousands of individual calls
			const response = await apiRequest('/admin/users/bulk', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ users: usersToCreate })
			});

			if (response.ok) {
				const responseData = await response.json();
				
				console.log(`✅ Bulk creation completed: ${responseData.total_created}/${responseData.total_requested} users created`);
				
				// Update frontend state for successfully created users
				if (responseData.created_users && responseData.created_users.length > 0) {
					for (const createdUser of responseData.created_users) {
						// Find the corresponding customer and move to synced
						const customer = customersToCreate.find(c => c.email === createdUser.email);
						if (customer) {
							moveCustomerFromStripeOnlyToSynced(customer, createdUser.id);
						}
					}
				}

				// Show success message
				if (responseData.total_created > 0) {
					showToast(`✅ Successfully processed ${responseData.total_created} users!`, 'success');
				}

				// Show success details for added Stripe IDs
				if (responseData.successes && responseData.successes.length > 0) {
					const addedStripeIDs = responseData.successes.filter((success: string) => success.includes('Added Stripe ID'));
					if (addedStripeIDs.length > 0) {
						showToast(`🔗 ${addedStripeIDs.length} Stripe IDs added to existing users!`, 'success');
						console.log('🔗 Added Stripe IDs:', addedStripeIDs);
					}
				}

				// 🔄 CRITICAL: Refresh data from database to get latest state
				console.log('🔄 Refreshing data from database after bulk creation...');
				await refreshData();
				console.log('✅ Data refreshed. New Stripe Only count:', allStripeOnlyCustomers.length);

				// Show error details if any failed
				if (responseData.total_failed > 0) {
					// Categorize errors by actual issue type
					const emailErrors = responseData.errors.filter((error: string) => 
						error.includes('email is required') || error.includes('invalid email')
					);
					const nameErrors = responseData.errors.filter((error: string) => 
						error.includes('name is required') || error.includes('invalid name')
					);
					const duplicateErrors = responseData.errors.filter((error: string) => 
						error.includes('already exists with this Stripe ID')
					);
					const databaseErrors = responseData.errors.filter((error: string) => 
						error.includes('Failed to create user') || error.includes('database error')
					);
					const otherErrors = responseData.errors.filter((error: string) => 
						!emailErrors.includes(error) && 
						!nameErrors.includes(error) && 
						!duplicateErrors.includes(error) && 
						!databaseErrors.includes(error)
					);
					
					// Show specific error categories
					if (emailErrors.length > 0) {
						showToast(`📧 ${emailErrors.length} users failed due to email issues.`, 'error');
						console.error('📧 Email validation errors:', emailErrors);
					}
					
					if (nameErrors.length > 0) {
						showToast(`👤 ${nameErrors.length} users failed due to name validation.`, 'error');
						console.error('👤 Name validation errors:', nameErrors);
					}
					
					if (duplicateErrors.length > 0) {
						showToast(`ℹ️ ${duplicateErrors.length} users already exist with same Stripe ID (skipped).`, 'info');
						console.log('📋 Duplicate users (same Stripe ID):', duplicateErrors.length);
					}
					
					if (databaseErrors.length > 0) {
						showToast(`💾 ${databaseErrors.length} users failed due to database issues.`, 'error');
						console.error('💾 Database errors:', databaseErrors);
					}
					
					if (otherErrors.length > 0) {
						showToast(`⚠️ ${otherErrors.length} users failed due to other issues.`, 'warning');
						console.warn('⚠️ Other errors:', otherErrors);
					}
					
					// Full error list for debugging
					console.warn('📊 Bulk user creation summary:', {
						total_failed: responseData.total_failed,
						email_errors: emailErrors.length,
						name_errors: nameErrors.length,
						duplicates: duplicateErrors.length,
						database_errors: databaseErrors.length,
						other_errors: otherErrors.length,
						all_errors: responseData.errors
					});
				}

			} else {
				const errorData = await response.json();
				console.error('Bulk user creation failed:', errorData);
				showToast(`❌ Bulk user creation failed: ${errorData.error || 'Unknown error'}`, 'error');
			}

		} catch (error: any) {
			console.error('Failed to create users in bulk:', error);
			showToast('Failed to create users in bulk', 'error');
		} finally {
			bulkCreatingUsers = false;
		}
	}

	// Immediately move customer from Stripe Only to Synced (immutable update)
	function moveCustomerFromStripeOnlyToSynced(customer: any, localId?: number) {
		console.log('🔄 Moving customer from Stripe Only to Synced:', customer.email);
		
		// Remove from Stripe Only list (immutable)
		allStripeOnlyCustomers = allStripeOnlyCustomers.filter(c => c.id !== customer.id);
		
		// Create synced customer with proper local ID
		const syncedCustomer = {
			...customer,
			syncStatus: 'synced',
			source: 'hybrid',
			id: `hybrid_${customer.stripeId}`,
			localId: localId || 'pending',
			role: 'user'
		};
		
		// Add to synced list (immutable)
		syncedCustomers = [syncedCustomer, ...syncedCustomers];
		
		// Update counts immediately
		stripeOnlyCount = allStripeOnlyCustomers.length;
		syncedCount = syncedCustomers.length;
		
		// Update pagination
		stripeOnlyTotalPages = Math.ceil(stripeOnlyCount / stripeOnlyItemsPerPage);
		updateStripeOnlyPagination();
		
		console.log('✅ Customer moved successfully. New counts:', {
			stripeOnly: stripeOnlyCount,
			synced: syncedCount
		});
	}

	// Update Stripe Only pagination
	function updateStripeOnlyPagination() {
		const startIndex = (stripeOnlyCurrentPage - 1) * stripeOnlyItemsPerPage;
		const endIndex = startIndex + stripeOnlyItemsPerPage;
		stripeOnlyDisplayed = allStripeOnlyCustomers.slice(startIndex, endIndex);
		stripeOnlyCustomers = stripeOnlyDisplayed; // For backward compatibility with existing components
		
		console.log(`📄 Stripe Only pagination: Page ${stripeOnlyCurrentPage}/${stripeOnlyTotalPages}, showing ${stripeOnlyDisplayed.length} of ${allStripeOnlyCustomers.length} customers`);
	}

	// Pagination navigation functions
	function goToStripeOnlyPage(page: number) {
		if (page >= 1 && page <= stripeOnlyTotalPages) {
			stripeOnlyCurrentPage = page;
			updateStripeOnlyPagination();
		}
	}

	function nextStripeOnlyPage() {
		if (stripeOnlyCurrentPage < stripeOnlyTotalPages) {
			stripeOnlyCurrentPage++;
			updateStripeOnlyPagination();
		}
	}

	function prevStripeOnlyPage() {
		if (stripeOnlyCurrentPage > 1) {
			stripeOnlyCurrentPage--;
			updateStripeOnlyPagination();
		}
	}



	// Format date
	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Refresh data
	async function refreshData() {
		if (parentSummary && stripeData) {
			// If we have parent data, we should refresh from parent
			// For now, just reload the current data
			showToast('Please refresh from the main subscribers page', 'info');
		} else {
			await loadAllData();
			showToast('Data refreshed', 'success');
		}
	}

	// Event handlers for table components
	function handleCreateUser(event: CustomEvent) {
		const customer = event.detail;
		createUserFromStripe(customer);
	}

	function handleCreateAllUsers() {
		createAllUsersFromStripe();
	}

	function handleSyncToStripe(event: CustomEvent) {
		const customer = event.detail;
		// TODO: Implement individual sync to Stripe
		showToast('Sync to Stripe functionality coming soon!', 'info');
	}

	function handleSyncAllToStripe() {
		// TODO: Implement bulk sync to Stripe
		showToast('Sync All to Stripe functionality coming soon!', 'info');
	}

	// Handle resync modal actions
	function openResyncModal() {
		showResyncModal = true;
	}

	function closeResyncModal() {
		showResyncModal = false;
	}

	function handleResyncSuccess() {
		// Refresh data after successful resync
		showToast('Resync completed! Refreshing data...', 'success');
		refreshData();
	}
</script>

<div class="customers-page">
	<div class="page-header">
		<div class="header-content">
			<h1>👥 Customer Sync Dashboard</h1>
			<p>Compare local subscribers with Stripe customers from database and manage synchronization</p>
		</div>
		<div class="header-stats">
			<div class="stat-card">
				<span class="stat-value">{totalCount}</span>
				<span class="stat-label">Total Customers</span>
			</div>
			<div class="stat-card synced">
				<span class="stat-value">{syncedCount}</span>
				<span class="stat-label">Synced</span>
			</div>
			<div class="stat-card local">
				<span class="stat-value">{localOnlyCount}</span>
				<span class="stat-label">Local Only</span>
			</div>
			<div class="stat-card stripe">
				<span class="stat-value">{allStripeOnlyCustomers.length}</span>
				<span class="stat-label">Stripe Only</span>
			</div>
		</div>
		
		<div class="header-actions">
			<button class="btn btn-secondary" onclick={refreshData}>
				🔄 Refresh Data
			</button>
			
			<button class="btn btn-primary" onclick={openResyncModal}>
				🔄 Resync Database
			</button>
		</div>
	</div>

	<!-- Customer Sync Panel 
	<CustomerSyncPanel customerId={null} customerEmail="" />-->

	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading customer data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<h3>Error Loading Data</h3>
			<p>{error}</p>
			<button class="btn btn-primary" onclick={refreshData}>Retry</button>
		</div>
	{:else if localOnlyUsers.length === 0 && stripeOnlyCustomers.length === 0 && syncedCustomers.length === 0}
		<div class="empty-state">
			<div class="empty-icon">👥</div>
			<h3>No Customers Found</h3>
			<p>No local subscribers or Stripe customers found in the database.</p>
			<p class="empty-hint">💡 If you have Stripe customers, you may need to run a sync first to populate the database tables.</p>
			<button class="btn btn-primary" onclick={() => window.location.href = '/admin/streaming/stripe/setup'}>
				Go to Stripe Setup
			</button>
		</div>
	{:else}
		<div class="customers-table-container">
			<!-- Debug information 
			<div style="background: #f0f0f0; padding: 1rem; margin: 1rem 0; border-radius: 0.5rem;">
				<h3>🔧 Debug Info:</h3>
				<p><strong>Stripe Only:</strong> {stripeOnlyCustomers.length}</p>
				<p><strong>Synced:</strong> {syncedCustomers.length}</p>
				<p><strong>Local Only:</strong> {localOnlyUsers.length}</p>
				<p><strong>Loading:</strong> {loading}</p>
				<p><strong>Error:</strong> {error || 'None'}</p>
			</div>-->

			<!-- Stripe Only Customers Table -->
			<SimpleTable
				title ="💳 Stripe Only Customers ({allStripeOnlyCustomers.length} total)"
				customers={stripeOnlyCustomers}
				tableType="stripe-only"
				initiallyExpanded={true}
				{syncingCustomers}
				{bulkCreatingUsers}
				oncreateUser={handleCreateUser}
				oncreateAllUsers={handleCreateAllUsers}
			/>
			<!-- Synced Customers Table with Built-in Search and Pagination -->
			<SimpleTable
				title="🔗 Synced Customers"
				customers={allSyncedCustomers}
				tableType="synced"
				initiallyExpanded={true}
				enableSearch={true}
				enablePagination={true}
				itemsPerPage={50}
				searchPlaceholder="Search by name, email, or Stripe ID..."
			/>
			
			<!-- Local Only Users Table -->
			<SimpleTable
				title="🏠 Local Only Users ({localOnlyUsers.length})"
				customers={localOnlyUsers}
				tableType="local-only"
				initiallyExpanded={false}
				onsyncToStripe={handleSyncToStripe}
				onsyncAllToStripe={handleSyncAllToStripe}
			/>
			
	
		</div>
	{/if}
</div>

<!-- Stripe Resync Modal -->
<StripeResyncModal 
	isOpen={showResyncModal} 
	onClose={closeResyncModal} 
	onSuccess={handleResyncSuccess} 
/>

<style>
	.customers-page {
		padding: var(--space-lg, 1.5rem);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl, 2rem);
		flex-wrap: wrap;
		gap: var(--space-lg, 1.5rem);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs, 0.5rem) 0;
		color: var(--text, #111827);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted, #6b7280);
		font-size: 1.1rem;
	}

	.header-stats {
		display: flex;
		gap: var(--space-md, 1rem);
		flex-wrap: wrap;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md, 1rem);
		background: var(--surface, white);
		border-radius: var(--radius-lg, 0.5rem);
		border: 1px solid var(--border, #e5e7eb);
		min-width: 100px;
	}

	.stat-card.synced {
		border-color: #059669;
		background: #f0fdf4;
	}

	.stat-card.local {
		border-color: #2563eb;
		background: #eff6ff;
	}

	.stat-card.stripe {
		border-color: #d97706;
		background: #fffbeb;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary, #2563eb);
		margin-bottom: var(--space-xs, 0.5rem);
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted, #6b7280);
		text-align: center;
	}

	.header-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: var(--space-md, 1rem);
	}

	.btn {
		padding: var(--space-sm, 0.75rem) var(--space-lg, 1.5rem);
		border: none;
		border-radius: var(--radius-md, 0.375rem);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs, 0.5rem);
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #1d4ed8;
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.loading-container, .error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl, 2rem);
		text-align: center;
		min-height: 400px;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl, 2rem);
		text-align: center;
		min-height: 400px;
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg, 1.5rem);
		opacity: 0.5;
	}

	.empty-hint {
		color: var(--text-muted, #6b7280);
		font-style: italic;
		margin: var(--space-md, 1rem) 0;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: var(--space-md, 1rem);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.customers-table-container {
		overflow-x: auto;
		border-radius: var(--radius-lg, 0.5rem);
		border: 1px solid var(--border, #e5e7eb);
		box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
		margin-top: var(--space-lg, 1.5rem);
	}

	@media (max-width: 768px) {
		.page-header {
			flex-direction: column;
			align-items: center;
			text-align: center;
		}

			.header-stats {
		justify-content: center;
	}

	/* Conflict Section Styles */
	.conflict-section {
		background: #fef3c7;
		border: 1px solid #f59e0b;
		border-radius: var(--radius-lg, 0.5rem);
		padding: var(--space-lg, 1.5rem);
		margin-top: var(--space-lg, 1.5rem);
	}

	.conflict-header h3 {
		margin: 0 0 var(--space-xs, 0.5rem) 0;
		color: #92400e;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.conflict-header p {
		margin: 0 0 var(--space-lg, 1.5rem) 0;
		color: #92400e;
		font-size: 0.875rem;
	}

	.conflict-item {
		background: white;
		border: 1px solid #f59e0b;
		border-radius: var(--radius-md, 0.375rem);
		padding: var(--space-md, 1rem);
		margin-bottom: var(--space-md, 1rem);
	}

	.conflict-item:last-child {
		margin-bottom: 0;
	}

	.conflict-email {
		margin-bottom: var(--space-md, 1rem);
		padding-bottom: var(--space-sm, 0.75rem);
		border-bottom: 1px solid #fde68a;
		font-size: 1.1rem;
		color: #92400e;
	}

	.conflict-details {
		display: flex;
		gap: var(--space-lg, 1.5rem);
		margin-bottom: var(--space-md, 1rem);
		flex-wrap: wrap;
	}

	.conflict-customer {
		flex: 1;
		min-width: 250px;
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: var(--radius-sm, 0.25rem);
		padding: var(--space-sm, 0.75rem);
	}

	.conflict-source {
		font-weight: 600;
		margin-bottom: var(--space-xs, 0.5rem);
		color: #374151;
	}

	.conflict-info div {
		margin-bottom: var(--space-xs, 0.5rem);
		font-size: 0.875rem;
		color: #6b7280;
	}

	.conflict-info div:last-child {
		margin-bottom: 0;
	}

	.conflict-actions {
		text-align: right;
	}

	.btn-sm {
		padding: var(--space-xs, 0.5rem) var(--space-sm, 0.75rem);
		font-size: 0.875rem;
	}


}
</style> 
