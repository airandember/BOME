<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import CustomerSyncPanel from '../components/CustomerSyncPanel.svelte';
	import SimpleTable from './SimpleTable.svelte';
	import { StreamingSubscriberService } from '$lib/services/streaming-subscribers';
	import { showToast } from '$lib/toast';

	// State variables
	let loading = true;
	let error = '';
	
	// Data arrays for different sync states
	let stripeOnlyCustomers: any[] = [];
	let syncedCustomers: any[] = [];
	let localOnlyUsers: any[] = [];
	
	// Sync state
	let syncingCustomers = new Set<string>();
	let bulkCreatingUsers = false;
	
	// Stats
	let totalCount = 0;
	let syncedCount = 0;
	let localOnlyCount = 0;
	let stripeOnlyCount = 0;

	onMount(() => {
		// Load initial data
		loadAllData();

		// Add page visibility listener to refresh data when returning to this page
		// This ensures that after creating a user from the users page, the data updates
		const handleVisibilityChange = () => {
			if (!document.hidden) {
				console.log('🔄 Page became visible, refreshing customer data...');
				loadAllData();
			}
		};

		if (typeof document !== 'undefined') {
			document.addEventListener('visibilitychange', handleVisibilityChange);

			// Return cleanup function
			return () => {
				document.removeEventListener('visibilitychange', handleVisibilityChange);
			};
		}
	});

	// Load both local users and Stripe customers
	async function loadAllData() {
		try {
			loading = true;
			error = '';

			console.log('🔄 Loading Stripe customers and checking against local users...');

			// First, fetch all Stripe customers (primary source)
			const stripeRes = await apiRequest('/admin/streaming/stripe/summary');
			let allStripeCustomers: any[] = [];
			if (stripeRes.ok) {
				const stripeData = await stripeRes.json();
				allStripeCustomers = stripeData.summary?.customers || [];
				console.log('✅ Loaded Stripe customers:', allStripeCustomers.length);
			} else {
				console.error('❌ Failed to load Stripe customers');
				allStripeCustomers = [];
			}

			// Then, fetch all local users to check against
			const usersRes = await apiRequest('/admin/users?limit=1000');
			let localUsers: any[] = [];
			if (usersRes.ok) {
				const usersData = await usersRes.json();
				localUsers = usersData.users || [];
				console.log('✅ Loaded local users for matching:', localUsers.length);
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

		// Create maps for fast lookup
		const usersByEmail = new Map();
		const usersByStripeId = new Map();
		const stripeCustomersByEmail = new Map();
		const stripeCustomersById = new Map();
		
		// Map local users
		localUsers.forEach((user: any) => {
			const email = (user.Email || user.email || '').toLowerCase();
			if (email) {
				usersByEmail.set(email, user);
			}
			
			const stripeId = user.StripeCustomerID || user.stripe_customer_id;
			if (stripeId) {
				usersByStripeId.set(stripeId, user);
			}
		});

		// Map Stripe customers
		allStripeCustomers.forEach((customer: any) => {
			const email = customer.Email?.toLowerCase();
			if (email) {
				stripeCustomersByEmail.set(email, customer);
			}
			stripeCustomersById.set(customer.ID, customer);
		});

		// 1. Stripe Only Customers - exist in Stripe but not locally
		stripeOnlyCustomers = allStripeCustomers.filter((stripeCustomer: any) => {
			const email = stripeCustomer.Email?.toLowerCase();
			const stripeId = stripeCustomer.ID;
			
			// Check if this Stripe customer has a local user
			const localUser = usersByStripeId.get(stripeId) || usersByEmail.get(email);
			return !localUser; // Only include if no local user found
		}).map((customer: any) => ({
			id: `stripe_${customer.ID}`,
			source: 'stripe',
			name: customer.Name || 'Unnamed Customer',
			email: customer.Email,
			localId: null,
			role: null,
			planName: null,
			stripePriceId: null,
			stripeId: customer.ID,
			stripeCreatedAt: customer.CreatedAt,
			stripeMetadata: customer.Metadata,
			createdAt: customer.CreatedAt,
			stripeCustomerId: null,
			syncStatus: 'stripe_only'
		}));

		// 2. Synced Customers - exist in both Stripe and locally
		syncedCustomers = allStripeCustomers.filter((stripeCustomer: any) => {
			const email = stripeCustomer.Email?.toLowerCase();
			const stripeId = stripeCustomer.ID;
			
			// Check if this Stripe customer has a local user
			const localUser = usersByStripeId.get(stripeId) || usersByEmail.get(email);
			return !!localUser; // Only include if local user found
		}).map((stripeCustomer: any) => {
			const email = stripeCustomer.Email?.toLowerCase();
			const stripeId = stripeCustomer.ID;
			const localUser = usersByStripeId.get(stripeId) || usersByEmail.get(email);
			
			return {
				id: `hybrid_${stripeId}`,
				source: 'hybrid',
				name: stripeCustomer.Name || `${localUser.FirstName || localUser.first_name || ''} ${localUser.LastName || localUser.last_name || ''}`.trim(),
				email: stripeCustomer.Email,
				localId: localUser.ID || localUser.id,
				role: localUser.Role || localUser.role,
				planName: localUser.plan_name || null,
				stripePriceId: localUser.stripe_price_id || null,
				stripeId: stripeId,
				stripeCreatedAt: stripeCustomer.CreatedAt,
				stripeMetadata: stripeCustomer.Metadata,
				createdAt: localUser.CreatedAt || localUser.created_at,
				stripeCustomerId: localUser.StripeCustomerID || localUser.stripe_customer_id,
				syncStatus: 'synced'
			};
		});

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
	}

	// Create user directly from Stripe customer data (inline)
	async function createUserFromStripe(customer: any) {
		if (!customer.stripeId) {
			showToast('Invalid Stripe customer ID', 'error');
			return;
		}

		if (syncingCustomers.has(customer.id)) return;

		try {
			syncingCustomers.add(customer.id);
			syncingCustomers = new Set(syncingCustomers); // Trigger reactivity

			// Parse name into first and last name if possible
			let firstName = '';
			let lastName = '';
			if (customer.name) {
				const nameParts = customer.name.trim().split(' ');
				if (nameParts.length === 1) {
					firstName = nameParts[0];
				} else if (nameParts.length >= 2) {
					firstName = nameParts[0];
					lastName = nameParts.slice(1).join(' ');
				}
			}

			// Create user with Stripe customer data
			const userData = {
				first_name: firstName,
				last_name: lastName,
				email: customer.email,
				role: 'user', // Default role
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
				
				// Refresh data to show the updated status
				await loadAllData();
			} else {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to create user');
			}
		} catch (error: any) {
			console.error('Failed to create user from Stripe:', error);
			showToast(error.message || 'Failed to create user', 'error');
		} finally {
			syncingCustomers.delete(customer.id);
			syncingCustomers = new Set(syncingCustomers); // Trigger reactivity
		}
	}

	// Create all users from Stripe-only customers (bulk action)
	async function createAllUsersFromStripe() {
		if (stripeOnlyCustomers.length === 0) {
			showToast('No Stripe-only customers to create', 'info');
			return;
		}

		if (bulkCreatingUsers) return;

		try {
			bulkCreatingUsers = true;
			let successCount = 0;
			let errorCount = 0;
			const errors: string[] = [];

			showToast(`Creating ${stripeOnlyCustomers.length} users...`, 'info');

			// Create users one by one (could be optimized to batch API calls)
			for (const customer of stripeOnlyCustomers) {
				try {
					// Parse name into first and last name if possible
					let firstName = '';
					let lastName = '';
					if (customer.name) {
						const nameParts = customer.name.trim().split(' ');
						if (nameParts.length === 1) {
							firstName = nameParts[0];
						} else if (nameParts.length >= 2) {
							firstName = nameParts[0];
							lastName = nameParts.slice(1).join(' ');
						}
					}

					// Create user with Stripe customer data
					const userData = {
						first_name: firstName,
						last_name: lastName,
						email: customer.email,
						role: 'user', // Default role
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
						successCount++;
					} else {
						const errorData = await response.json();
						errors.push(`${customer.email}: ${errorData.error || 'Unknown error'}`);
						errorCount++;
					}
				} catch (error: any) {
					errors.push(`${customer.email}: ${error.message || 'Network error'}`);
					errorCount++;
				}
			}

			// Show results
			if (successCount > 0) {
				showToast(`✅ Successfully created ${successCount} users!`, 'success');
			}
			if (errorCount > 0) {
				showToast(`❌ Failed to create ${errorCount} users. Check console for details.`, 'error');
				console.error('Bulk user creation errors:', errors);
			}

			// Refresh data to show the updated status
			await loadAllData();

		} catch (error: any) {
			console.error('Failed to create users in bulk:', error);
			showToast('Failed to create users in bulk', 'error');
		} finally {
			bulkCreatingUsers = false;
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
		await loadAllData();
		showToast('Data refreshed', 'success');
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
</script>

<div class="customers-page">
	<div class="page-header">
		<div class="header-content">
			<h1>👥 Customer Sync Dashboard</h1>
			<p>Compare local subscribers with Stripe customers and manage synchronization</p>
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
				<span class="stat-value">{stripeOnlyCount}</span>
				<span class="stat-label">Stripe Only</span>
			</div>
		</div>
		
		<div class="header-actions">
			<button class="btn btn-secondary" on:click={refreshData}>
				🔄 Refresh Data
			</button>
		</div>
	</div>

	<!-- Customer Sync Panel -->
	<CustomerSyncPanel customerId={null} customerEmail="" />

	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading customer data...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<h3>Error Loading Data</h3>
			<p>{error}</p>
			<button class="btn btn-primary" on:click={refreshData}>Retry</button>
		</div>
	{:else if localOnlyUsers.length === 0 && stripeOnlyCustomers.length === 0 && syncedCustomers.length === 0}
		<div class="empty-state">
			<div class="empty-icon">👥</div>
			<h3>No Customers Found</h3>
			<p>No local subscribers or Stripe customers found.</p>
		</div>
	{:else}
		<div class="customers-table-container">
			<!-- Debug information -->
			<div style="background: #f0f0f0; padding: 1rem; margin: 1rem 0; border-radius: 0.5rem;">
				<h3>🔧 Debug Info:</h3>
				<p><strong>Stripe Only:</strong> {stripeOnlyCustomers.length}</p>
				<p><strong>Synced:</strong> {syncedCustomers.length}</p>
				<p><strong>Local Only:</strong> {localOnlyUsers.length}</p>
				<p><strong>Loading:</strong> {loading}</p>
				<p><strong>Error:</strong> {error || 'None'}</p>
			</div>

			<!-- Stripe Only Customers Table -->
			<SimpleTable
				title="💳 Stripe Only Customers"
				customers={stripeOnlyCustomers}
				tableType="stripe-only"
				{syncingCustomers}
				{bulkCreatingUsers}
				on:createUser={handleCreateUser}
				on:createAllUsers={handleCreateAllUsers}
			/>

			<!-- Synced Customers Table -->
			<SimpleTable
				title="🔗 Synced Customers"
				customers={syncedCustomers}
				tableType="synced"
			/>
			
			<!-- Local Only Users Table -->
			<SimpleTable
				title="🏠 Local Only Users"
				customers={localOnlyUsers}
				tableType="local-only"
				on:syncToStripe={handleSyncToStripe}
				on:syncAllToStripe={handleSyncAllToStripe}
			/>
		</div>
	{/if}
</div>

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
	}
</style> 