<script>
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { fade, fly } from 'svelte/transition';
	import StripeOnlyTable from '../subscribers/customers/StripeOnlyTable.svelte';
	import SyncedCustomersTable from '../subscribers/customers/SyncedCustomersTable.svelte';
	import LocalOnlyTable from '../subscribers/customers/LocalOnlyTable.svelte';

	// Reactive variables
	let isLoading = true;
	let customers = [];
	let error = null;
	let selectedCustomer = null;
	let showCustomerModal = false;
	let showRefundModal = false;
	let showCommunicationModal = false;
	let isSubmitting = false;

	// Props from parent component
	export let summary = { total_customers: 0, total_subscriptions: 0 };
	export let stripeData = { customers: [], subscriptions: [] };

	// Separate customer categories
	$: stripeOnlyCustomers = stripeData.customers.filter(customer => 
		customer.source === 'stripe' && !customer.localId
	);
	$: syncedCustomers = stripeData.customers.filter(customer => 
		customer.source === 'stripe' && customer.localId
	);
	$: localOnlyCustomers = stripeData.customers.filter(customer => 
		customer.source === 'local' && !customer.stripeId
	);

	// State for syncing operations
	let syncingCustomers = new Set();
	let bulkCreatingUsers = false;

	// Event handlers for table actions
	function handleCreateUser(customer) {
		console.log('Creating user for customer:', customer);
		// Add the customer ID to syncing set
		syncingCustomers = new Set([...syncingCustomers, customer.id]);
		
		// Simulate API call (replace with actual implementation)
		setTimeout(() => {
			syncingCustomers = new Set([...syncingCustomers].filter(id => id !== customer.id));
			console.log('User created successfully');
		}, 2000);
	}

	function handleCreateAllUsers() {
		console.log('Creating all users...');
		bulkCreatingUsers = true;
		
		// Simulate API call (replace with actual implementation)
		setTimeout(() => {
			bulkCreatingUsers = false;
			console.log('All users created successfully');
		}, 3000);
	}


	// Search and filters
	let searchTerm = '';
	let statusFilter = 'all';
	let planFilter = 'all';

	// Communication form
	let communicationData = {
		subject: '',
		message: '',
		type: 'email' // email, sms, in_app
	};

	// Refund form
	let refundData = {
		amount: 0,
		reason: '',
		notes: ''
	};

	// DataTable event handlers
	function handleRowAction(event) {
		const { item: customer, action } = event.detail;
		switch (action) {
			case 'view':
				openCustomerModal(customer);
				break;
			case 'edit':
				openCustomerModal(customer);
				break;
		}
	}

	// Transform customers data for DataTable
	$: customersForTable = customers.map(customer => ({
		...customer,
		plan_name: customer.subscription?.plan_name || 'No subscription',
		plan_price: customer.subscription ? formatCurrency(customer.subscription.price, customer.subscription.currency) : '-',
		status: customer.subscription?.status || 'inactive',
		next_billing: customer.subscription?.current_period_end || null
	}));

	// Load customers
	async function loadCustomers() {
		try {
			isLoading = true;
			error = null;

			const params = new URLSearchParams({
				page: currentPage.toString(),
				limit: itemsPerPage.toString(),
				search: searchTerm,
				status: statusFilter,
				plan: planFilter
			});

			const response = await apiRequest(`/admin/streaming/customers?${params}`);
			
			if (response.ok) {
				const data = await response.json();
				customers = data.customers || [];
				totalItems = data.total || 0;
			} else {
				throw new Error('Failed to load customers');
			}
		} catch (err) {
			console.error('Error loading customers:', err);
			error = err.message;
		} finally {
			isLoading = false;
		}
	}

	// Search customers
	function handleSearch() {
		currentPage = 1;
		loadCustomers();
	}

	// Filter customers
	function handleFilter() {
		currentPage = 1;
		loadCustomers();
	}

	// Open customer details modal
	function openCustomerModal(customer) {
		selectedCustomer = customer;
		showCustomerModal = true;
	}

	// Open refund modal
	function openRefundModal(customer) {
		selectedCustomer = customer;
		refundData.amount = customer.subscription?.price || 0;
		showRefundModal = true;
	}

	// Open communication modal
	function openCommunicationModal(customer) {
		selectedCustomer = customer;
		showCommunicationModal = true;
	}

	// Process refund
	async function processRefund() {
		try {
			isSubmitting = true;

			const response = await apiRequest(`/admin/streaming/subscriptions/${selectedCustomer.subscription.id}/refund`, {
				method: 'POST',
				body: JSON.stringify(refundData)
			});

			if (response.ok) {
				showRefundModal = false;
				selectedCustomer = null;
				await loadCustomers();
			} else {
				const data = await response.json();
				throw new Error(data.error || 'Failed to process refund');
			}
		} catch (err) {
			console.error('Error processing refund:', err);
			error = err.message;
		} finally {
			isSubmitting = false;
		}
	}

	// Send communication
	async function sendCommunication() {
		try {
			isSubmitting = true;

			const response = await apiRequest(`/admin/streaming/customers/${selectedCustomer.id}/communication`, {
				method: 'POST',
				body: JSON.stringify(communicationData)
			});

			if (response.ok) {
				showCommunicationModal = false;
				selectedCustomer = null;
				communicationData = { subject: '', message: '', type: 'email' };
			} else {
				const data = await response.json();
				throw new Error(data.error || 'Failed to send communication');
			}
		} catch (err) {
			console.error('Error sending communication:', err);
			error = err.message;
		} finally {
			isSubmitting = false;
		}
	}

	// Cancel subscription
	async function cancelSubscription(customer) {
		if (!confirm(`Are you sure you want to cancel ${customer.name}'s subscription?`)) {
			return;
		}

		try {
		const response = await apiRequest(`/admin/streaming/subscriptions/${customer.subscription.id}`, {
			method: 'DELETE',
				body: JSON.stringify({ reason: 'Admin cancellation' })
			});

			if (response.ok) {
				await loadCustomers();
			} else {
				const data = await response.json();
				throw new Error(data.error || 'Failed to cancel subscription');
			}
		} catch (err) {
			console.error('Error cancelling subscription:', err);
			error = err.message;
		}
	}

	// Format currency
	function formatCurrency(amount, currency = 'USD') {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency
		}).format(amount);
	}

	// Format date
	function formatDate(dateString) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	// Get status badge
	function getStatusBadge(status) {
		const statusConfig = {
			active: { text: 'Active', class: 'bg-green-100 text-green-800' },
			cancelled: { text: 'Cancelled', class: 'bg-red-100 text-red-800' },
			past_due: { text: 'Past Due', class: 'bg-yellow-100 text-yellow-800' },
			unpaid: { text: 'Unpaid', class: 'bg-gray-100 text-gray-800' },
			trialing: { text: 'Trial', class: 'bg-blue-100 text-blue-800' }
		};
		return statusConfig[status] || { text: status, class: 'bg-gray-100 text-gray-800' };
	}

	// Get subscription status icon
	function getStatusIcon(status) {
		const iconConfig = {
			active: CheckCircleIcon,
			cancelled: XCircleIcon,
			past_due: ExclamationTriangleIcon,
			unpaid: XCircleIcon,
			trialing: ClockIcon
		};
		return iconConfig[status] || ClockIcon;
	}

	// Calculate total pages
	$: totalPages = Math.ceil(totalItems / itemsPerPage);

	// Initialize component
	onMount(() => {
		loadCustomers();
	});
</script>

<svelte:head>
	<title>Customer Management - Streaming Admin</title>
</svelte:head>

{#if isLoading}
	<div class="flex items-center justify-center py-12">
		<div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
	</div>
{:else if error}
	<div class="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
		<div class="flex items-center">
			<ExclamationTriangleIcon class="h-5 w-5 text-red-400 mr-2" />
			<p class="text-red-800">{error}</p>
		</div>
	</div>
{:else}
	<div class="space-y-6" in:fade={{ duration: 300 }}>
		<!-- Header -->
		<div class="flex justify-between items-center">
			<div>
				<h1 class="text-2xl font-bold text-gray-900">Customer Management</h1>
				<p class="text-gray-600">Manage customer subscriptions and support</p>
			</div>
			<div class="flex items-center space-x-4">
				<span class="text-sm text-gray-600">
					{totalItems} customers total
				</span>
			</div>
		</div>

		<!-- Search and Filters -->
		<div class="bg-white border border-gray-200 rounded-lg p-4">
			<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
				<!-- Search -->
				<div class="relative">
					<MagnifyingGlassIcon class="h-5 w-5 text-gray-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
					<input
						type="text"
						bind:value={searchTerm}
						placeholder="Search customers..."
						class="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						on:keydown={(e) => e.key === 'Enter' && handleSearch()}
					/>
				</div>

				<!-- Status Filter -->
				<div>
					<select
						bind:value={statusFilter}
						on:change={handleFilter}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="all">All Statuses</option>
						<option value="active">Active</option>
						<option value="cancelled">Cancelled</option>
						<option value="past_due">Past Due</option>
						<option value="unpaid">Unpaid</option>
						<option value="trialing">Trial</option>
					</select>
				</div>

				<!-- Plan Filter -->
				<div>
					<select
						bind:value={planFilter}
						on:change={handleFilter}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="all">All Plans</option>
						<option value="monthly">Monthly</option>
						<option value="annual">Annual</option>
						<option value="premium">Premium</option>
					</select>
				</div>

				<!-- Search Button -->
				<button
					class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors flex items-center justify-center space-x-2"
					on:click={handleSearch}
				>
					<MagnifyingGlassIcon class="h-4 w-4" />
					<span>Search</span>
				</button>
			</div>
		</div>

		<!-- Individual Customer Tables -->
		<div class="customer-tables">
			<!-- Stripe Only Customers -->
			<StripeOnlyTable 
				customers={stripeOnlyCustomers}
				{syncingCustomers}
				{bulkCreatingUsers}
				oncreateUser={handleCreateUser}
				oncreateAllUsers={handleCreateAllUsers}
			/>

			<!-- Synced Customers -->
			<SyncedCustomersTable 
				customers={syncedCustomers}
			/>

			<!-- Local Only Users -->
			<LocalOnlyTable 
				customers={localOnlyCustomers}
			/>
		</div>
	</div>
{/if}

<!-- Customer Details Modal -->
{#if showCustomerModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" in:fade={{ duration: 200 }}>
		<div class="bg-white rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto" in:fly={{ y: 20, duration: 200 }}>
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-xl font-semibold text-gray-900">Customer Details</h2>
				<button
					class="text-gray-400 hover:text-gray-600"
					on:click={() => showCustomerModal = false}
				>
					<XMarkIcon class="h-6 w-6" />
				</button>
			</div>

			{#if selectedCustomer}
				<div class="space-y-6">
					<!-- Customer Info -->
					<div>
						<h3 class="text-lg font-medium text-gray-900 mb-4">Customer Information</h3>
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700">Name</label>
								<p class="text-sm text-gray-900">{selectedCustomer.name}</p>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700">Email</label>
								<p class="text-sm text-gray-900">{selectedCustomer.email}</p>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700">Joined</label>
								<p class="text-sm text-gray-900">{formatDate(selectedCustomer.created_at)}</p>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700">Last Login</label>
								<p class="text-sm text-gray-900">{selectedCustomer.last_login ? formatDate(selectedCustomer.last_login) : 'Never'}</p>
							</div>
						</div>
					</div>

					<!-- Subscription Info -->
					{#if selectedCustomer.subscription}
						<div>
							<h3 class="text-lg font-medium text-gray-900 mb-4">Subscription Information</h3>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
								<div>
									<label class="block text-sm font-medium text-gray-700">Plan</label>
									<p class="text-sm text-gray-900">{selectedCustomer.subscription.plan_name}</p>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700">Status</label>
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadge(selectedCustomer.subscription.status).class}">
										{getStatusBadge(selectedCustomer.subscription.status).text}
									</span>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700">Price</label>
									<p class="text-sm text-gray-900">{formatCurrency(selectedCustomer.subscription.price, selectedCustomer.subscription.currency)}</p>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700">Next Billing</label>
									<p class="text-sm text-gray-900">{formatDate(selectedCustomer.subscription.current_period_end)}</p>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700">Created</label>
									<p class="text-sm text-gray-900">{formatDate(selectedCustomer.subscription.created_at)}</p>
								</div>
								<div>
									<label class="block text-sm font-medium text-gray-700">Stripe ID</label>
									<p class="text-sm text-gray-900 font-mono">{selectedCustomer.subscription.stripe_subscription_id}</p>
								</div>
							</div>
						</div>
					{:else}
						<div>
							<h3 class="text-lg font-medium text-gray-900 mb-4">Subscription Information</h3>
							<p class="text-sm text-gray-500">No active subscription</p>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	</div>
{/if}

<!-- Refund Modal -->
{#if showRefundModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" in:fade={{ duration: 200 }}>
		<div class="bg-white rounded-lg p-6 w-full max-w-md" in:fly={{ y: 20, duration: 200 }}>
			<div class="flex items-center mb-4">
				<CurrencyDollarIcon class="h-6 w-6 text-green-600 mr-3" />
				<h2 class="text-xl font-semibold text-gray-900">Process Refund</h2>
			</div>
			
			<form on:submit|preventDefault={processRefund} class="space-y-4">
				<div>
					<label for="refund-amount" class="block text-sm font-medium text-gray-700 mb-2">Refund Amount</label>
					<input
						id="refund-amount"
						type="number"
						step="0.01"
						min="0"
						bind:value={refundData.amount}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						required
					/>
				</div>

				<div>
					<label for="refund-reason" class="block text-sm font-medium text-gray-700 mb-2">Reason</label>
					<select
						id="refund-reason"
						bind:value={refundData.reason}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						required
					>
						<option value="">Select a reason</option>
						<option value="customer_request">Customer Request</option>
						<option value="service_issue">Service Issue</option>
						<option value="billing_error">Billing Error</option>
						<option value="duplicate_charge">Duplicate Charge</option>
						<option value="other">Other</option>
					</select>
				</div>

				<div>
					<label for="refund-notes" class="block text-sm font-medium text-gray-700 mb-2">Notes</label>
					<textarea
						id="refund-notes"
						bind:value={refundData.notes}
						rows="3"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						placeholder="Additional notes..."
					></textarea>
				</div>

				<div class="flex justify-end space-x-3 pt-4">
					<button
						type="button"
						class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
						on:click={() => showRefundModal = false}
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 transition-colors disabled:opacity-50"
					>
						{isSubmitting ? 'Processing...' : 'Process Refund'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Communication Modal -->
{#if showCommunicationModal}
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50" in:fade={{ duration: 200 }}>
		<div class="bg-white rounded-lg p-6 w-full max-w-md" in:fly={{ y: 20, duration: 200 }}>
			<div class="flex items-center mb-4">
				<ChatBubbleLeftIcon class="h-6 w-6 text-purple-600 mr-3" />
				<h2 class="text-xl font-semibold text-gray-900">Send Communication</h2>
			</div>
			
			<form on:submit|preventDefault={sendCommunication} class="space-y-4">
				<div>
					<label for="communication-type" class="block text-sm font-medium text-gray-700 mb-2">Type</label>
					<select
						id="communication-type"
						bind:value={communicationData.type}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="email">Email</option>
						<option value="sms">SMS</option>
						<option value="in_app">In-App Message</option>
					</select>
				</div>

				<div>
					<label for="communication-subject" class="block text-sm font-medium text-gray-700 mb-2">Subject</label>
					<input
						id="communication-subject"
						type="text"
						bind:value={communicationData.subject}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						required
					/>
				</div>

				<div>
					<label for="communication-message" class="block text-sm font-medium text-gray-700 mb-2">Message</label>
					<textarea
						id="communication-message"
						bind:value={communicationData.message}
						rows="4"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
						required
					></textarea>
				</div>

				<div class="flex justify-end space-x-3 pt-4">
					<button
						type="button"
						class="px-4 py-2 text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors"
						on:click={() => showCommunicationModal = false}
					>
						Cancel
					</button>
					<button
						type="submit"
						disabled={isSubmitting}
						class="px-4 py-2 bg-purple-600 text-white rounded-md hover:bg-purple-700 transition-colors disabled:opacity-50"
					>
						{isSubmitting ? 'Sending...' : 'Send Message'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if} 

<style>
	.customer-tables {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	@media (max-width: 768px) {
		.customer-tables {
			gap: 1rem;
		}
	}
</style>
