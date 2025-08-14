<script>
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';

	// Reactive variables
	let isLoading = true;
	let subscribers = [];
	let error = null;
	let selectedSubscriber = null;
	let showCustomerModal = false;
	let showRefundModal = false;
	let showCommunicationModal = false;
	let isSubmitting = false;


	// Search and filters
	let searchTerm = '';
	let statusFilter = 'all';
	let planFilter = 'all';

	// Pagination
	let currentPage = 1;
	let itemsPerPage = 20;
	let totalItems = 0;

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

			const response = await fetch(`/api/admin/streaming/customers?${params}`);
			
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

			const response = await fetch(`/api/admin/streaming/subscriptions/${selectedCustomer.subscription.id}/refund`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
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

			const response = await fetch(`/api/admin/streaming/customers/${selectedCustomer.id}/communication`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
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
			const response = await fetch(`/api/admin/streaming/subscriptions/${customer.subscription.id}`, {
				method: 'DELETE',
				headers: {
					'Content-Type': 'application/json'
				},
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

		<!-- Customers Table -->
		<div class="bg-white border border-gray-200 rounded-lg overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Customer
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Subscription
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Status
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Next Billing
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Actions
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each customers as customer (customer.id)}
							<tr class="hover:bg-gray-50" in:fly={{ y: 20, duration: 300 }}>
								<td class="px-6 py-4 whitespace-nowrap">
									<div class="flex items-center">
										<div class="h-10 w-10 bg-blue-600 rounded-full flex items-center justify-center">
											<span class="text-white text-sm font-medium">
												{customer.name?.charAt(0).toUpperCase() || 'U'}
											</span>
										</div>
										<div class="ml-4">
											<div class="text-sm font-medium text-gray-900">{customer.name}</div>
											<div class="text-sm text-gray-500">{customer.email}</div>
										</div>
									</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									{#if customer.subscription}
										<div>
											<div class="text-sm font-medium text-gray-900">
												{customer.subscription.plan_name}
											</div>
											<div class="text-sm text-gray-500">
												{formatCurrency(customer.subscription.price, customer.subscription.currency)}
											</div>
										</div>
									{:else}
										<span class="text-sm text-gray-500">No subscription</span>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									{#if customer.subscription}
										<div class="flex items-center">
											<svelte:component this={getStatusIcon(customer.subscription.status)} class="h-4 w-4 mr-2" />
											<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadge(customer.subscription.status).class}">
												{getStatusBadge(customer.subscription.status).text}
											</span>
										</div>
									{:else}
										<span class="text-sm text-gray-500">-</span>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{#if customer.subscription?.current_period_end}
										{formatDate(customer.subscription.current_period_end)}
									{:else}
										-
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
									<div class="flex items-center space-x-2">
										<button
											class="text-blue-600 hover:text-blue-900"
											on:click={() => openCustomerModal(customer)}
										>
											<EyeIcon class="h-4 w-4" />
										</button>
										{#if customer.subscription}
											<button
												class="text-green-600 hover:text-green-900"
												on:click={() => openRefundModal(customer)}
											>
												<CurrencyDollarIcon class="h-4 w-4" />
											</button>
											<button
												class="text-red-600 hover:text-red-900"
												on:click={() => cancelSubscription(customer)}
											>
												<XMarkIcon class="h-4 w-4" />
											</button>
										{/if}
										<button
											class="text-purple-600 hover:text-purple-900"
											on:click={() => openCommunicationModal(customer)}
										>
											<ChatBubbleLeftIcon class="h-4 w-4" />
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="bg-white px-4 py-3 border-t border-gray-200 sm:px-6">
					<div class="flex items-center justify-between">
						<div class="flex-1 flex justify-between sm:hidden">
							<button
								class="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
								disabled={currentPage === 1}
								on:click={() => currentPage > 1 && (currentPage--, loadCustomers())}
							>
								Previous
							</button>
							<button
								class="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50"
								disabled={currentPage === totalPages}
								on:click={() => currentPage < totalPages && (currentPage++, loadCustomers())}
							>
								Next
							</button>
						</div>
						<div class="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
							<div>
								<p class="text-sm text-gray-700">
									Showing <span class="font-medium">{((currentPage - 1) * itemsPerPage) + 1}</span> to <span class="font-medium">{Math.min(currentPage * itemsPerPage, totalItems)}</span> of <span class="font-medium">{totalItems}</span> results
								</p>
							</div>
							<div>
								<nav class="relative z-0 inline-flex rounded-md shadow-sm -space-x-px">
									<button
										class="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
										disabled={currentPage === 1}
										on:click={() => currentPage > 1 && (currentPage--, loadCustomers())}
									>
										Previous
									</button>
									{#each Array.from({ length: totalPages }, (_, i) => i + 1) as page}
										<button
											class="relative inline-flex items-center px-4 py-2 border text-sm font-medium {currentPage === page ? 'z-10 bg-blue-50 border-blue-500 text-blue-600' : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'}"
											on:click={() => currentPage !== page && (currentPage = page, loadCustomers())}
										>
											{page}
										</button>
									{/each}
									<button
										class="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50"
										disabled={currentPage === totalPages}
										on:click={() => currentPage < totalPages && (currentPage++, loadCustomers())}
									>
										Next
									</button>
								</nav>
							</div>
						</div>
					</div>
				</div>
			{/if}
		</div>

		{#if customers.length === 0}
			<div class="text-center py-12">
				<UsersIcon class="h-12 w-12 text-gray-400 mx-auto mb-4" />
				<h3 class="text-lg font-medium text-gray-900 mb-2">No customers found</h3>
				<p class="text-gray-600">Try adjusting your search or filter criteria.</p>
			</div>
		{/if}
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
