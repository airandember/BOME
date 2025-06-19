<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdvertiserAccount, AdCampaign } from '$lib/types/advertising';
	
	let advertiserAccounts: AdvertiserAccount[] = [];
	let campaigns: AdCampaign[] = [];
	let loading = true;
	let error: string | null = null;
	let activeTab: 'advertisers' | 'campaigns' = 'advertisers';

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		// Check if user is admin
		const user = $auth.user;
		if (!user || user.role !== 'admin') {
			goto('/');
			return;
		}

		try {
			await loadData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	});

	async function loadData() {
		if (activeTab === 'advertisers') {
			await loadAdvertisers();
		} else {
			await loadCampaigns();
		}
	}

	async function loadAdvertisers() {
		// For now, we'll use mock data since the backend endpoint is not implemented
		// TODO: Replace with actual API call when backend is implemented
		advertiserAccounts = [
			{
				id: 1,
				user_id: 2,
				company_name: 'Example Corp',
				business_email: 'contact@example.com',
				contact_name: 'John Doe',
				contact_phone: '(555) 123-4567',
				business_address: '123 Main St, City, State 12345',
				tax_id: '12-3456789',
				website: 'https://example.com',
				industry: 'technology',
				status: 'pending',
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			}
		];
	}

	async function loadCampaigns() {
		// For now, we'll use mock data since the backend endpoint is not implemented
		// TODO: Replace with actual API call when backend is implemented
		campaigns = [
			{
				id: 1,
				advertiser_id: 1,
				name: 'Q4 Product Launch',
				description: 'Promoting our new product line',
				status: 'pending',
				start_date: '2024-01-01',
				end_date: '2024-03-31',
				budget: 5000,
				spent_amount: 0,
				target_audience: 'Tech enthusiasts aged 25-45',
				billing_type: 'monthly',
				billing_rate: 1666.67,
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			}
		];
	}

	async function approveAdvertiser(advertiserId: number) {
		try {
			const response = await fetch(`/api/v1/admin/ads/advertisers/${advertiserId}/approve`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify({ notes: 'Account approved by admin' })
			});

			if (response.ok) {
				// Update local state
				advertiserAccounts = advertiserAccounts.map(acc => 
					acc.id === advertiserId ? { ...acc, status: 'approved' } : acc
				);
			} else {
				const data = await response.json();
				error = data.error || 'Failed to approve advertiser';
			}
		} catch (err) {
			error = 'Network error occurred';
		}
	}

	async function rejectAdvertiser(advertiserId: number) {
		const reason = prompt('Please provide a reason for rejection:');
		if (!reason) return;

		try {
			const response = await fetch(`/api/v1/admin/ads/advertisers/${advertiserId}/reject`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify({ notes: reason })
			});

			if (response.ok) {
				// Update local state
				advertiserAccounts = advertiserAccounts.map(acc => 
					acc.id === advertiserId ? { ...acc, status: 'rejected', verification_notes: reason } : acc
				);
			} else {
				const data = await response.json();
				error = data.error || 'Failed to reject advertiser';
			}
		} catch (err) {
			error = 'Network error occurred';
		}
	}

	async function approveCampaign(campaignId: number) {
		try {
			const response = await fetch(`/api/v1/admin/ads/campaigns/${campaignId}/approve`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify({ notes: 'Campaign approved by admin' })
			});

			if (response.ok) {
				// Update local state
				campaigns = campaigns.map(campaign => 
					campaign.id === campaignId ? { ...campaign, status: 'approved' } : campaign
				);
			} else {
				const data = await response.json();
				error = data.error || 'Failed to approve campaign';
			}
		} catch (err) {
			error = 'Network error occurred';
		}
	}

	async function rejectCampaign(campaignId: number) {
		const reason = prompt('Please provide a reason for rejection:');
		if (!reason) return;

		try {
			const response = await fetch(`/api/v1/admin/ads/campaigns/${campaignId}/reject`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify({ notes: reason })
			});

			if (response.ok) {
				// Update local state
				campaigns = campaigns.map(campaign => 
					campaign.id === campaignId ? { ...campaign, status: 'rejected', approval_notes: reason } : campaign
				);
			} else {
				const data = await response.json();
				error = data.error || 'Failed to reject campaign';
			}
		} catch (err) {
			error = 'Network error occurred';
		}
	}

	function switchTab(tab: 'advertisers' | 'campaigns') {
		activeTab = tab;
		loadData();
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'approved':
				return 'bg-green-100 text-green-800';
			case 'pending':
				return 'bg-yellow-100 text-yellow-800';
			case 'rejected':
				return 'bg-red-100 text-red-800';
			case 'active':
				return 'bg-blue-100 text-blue-800';
			case 'paused':
				return 'bg-gray-100 text-gray-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString();
	}
</script>

<svelte:head>
	<title>Advertisement Management - BOME Admin</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		<div class="mb-8">
			<h1 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
				Advertisement Management
			</h1>
			<p class="mt-1 text-sm text-gray-500">
				Manage advertiser accounts and campaign approvals
			</p>
		</div>

		{#if error}
			<div class="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<h3 class="text-sm font-medium text-red-800">Error</h3>
						<div class="mt-2 text-sm text-red-700">
							<p>{error}</p>
						</div>
					</div>
				</div>
			</div>
		{/if}

		<!-- Tab Navigation -->
		<div class="mb-6">
			<nav class="flex space-x-8" aria-label="Tabs">
				<button
					on:click={() => switchTab('advertisers')}
					class="whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'advertisers' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Advertiser Accounts
				</button>
				<button
					on:click={() => switchTab('campaigns')}
					class="whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm {activeTab === 'campaigns' ? 'border-blue-500 text-blue-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}"
				>
					Campaign Approvals
				</button>
			</nav>
		</div>

		{#if loading}
			<div class="flex justify-center items-center h-64">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		{:else if activeTab === 'advertisers'}
			<!-- Advertiser Accounts -->
			<div class="bg-white shadow overflow-hidden sm:rounded-md">
				<div class="px-4 py-5 sm:px-6 border-b border-gray-200">
					<h3 class="text-lg leading-6 font-medium text-gray-900">Advertiser Accounts</h3>
					<p class="mt-1 max-w-2xl text-sm text-gray-500">Review and approve advertiser account applications</p>
				</div>

				{#if advertiserAccounts.length === 0}
					<div class="text-center py-12">
						<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
						</svg>
						<h3 class="mt-2 text-sm font-medium text-gray-900">No advertiser accounts</h3>
						<p class="mt-1 text-sm text-gray-500">No advertiser accounts to review at this time.</p>
					</div>
				{:else}
					<ul class="divide-y divide-gray-200">
						{#each advertiserAccounts as account}
							<li class="px-4 py-4 sm:px-6">
								<div class="flex items-center justify-between">
									<div class="flex-1">
										<div class="flex items-center">
											<div class="flex-shrink-0">
												<div class="h-10 w-10 rounded-full bg-blue-100 flex items-center justify-center">
													<svg class="h-6 w-6 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
													</svg>
												</div>
											</div>
											<div class="ml-4">
												<div class="flex items-center">
													<p class="text-sm font-medium text-gray-900">{account.company_name}</p>
													<span class="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusColor(account.status)}">
														{account.status}
													</span>
												</div>
												<div class="mt-1 text-sm text-gray-500">
													<p>{account.contact_name} • {account.business_email}</p>
													<p>{account.industry || 'No industry specified'} • Applied {formatDate(account.created_at)}</p>
												</div>
											</div>
										</div>
									</div>
									<div class="flex space-x-2">
										{#if account.status === 'pending'}
											<button
												on:click={() => approveAdvertiser(account.id)}
												class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
											>
												Approve
											</button>
											<button
												on:click={() => rejectAdvertiser(account.id)}
												class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
											>
												Reject
											</button>
										{:else}
											<span class="text-sm text-gray-500">
												{account.status === 'approved' ? 'Approved' : 'Rejected'}
											</span>
										{/if}
									</div>
								</div>
								{#if account.verification_notes}
									<div class="mt-2 text-sm text-gray-600">
										<strong>Notes:</strong> {account.verification_notes}
									</div>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{:else}
			<!-- Campaign Approvals -->
			<div class="bg-white shadow overflow-hidden sm:rounded-md">
				<div class="px-4 py-5 sm:px-6 border-b border-gray-200">
					<h3 class="text-lg leading-6 font-medium text-gray-900">Campaign Approvals</h3>
					<p class="mt-1 max-w-2xl text-sm text-gray-500">Review and approve advertising campaigns</p>
				</div>

				{#if campaigns.length === 0}
					<div class="text-center py-12">
						<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
						</svg>
						<h3 class="mt-2 text-sm font-medium text-gray-900">No campaigns</h3>
						<p class="mt-1 text-sm text-gray-500">No campaigns pending approval at this time.</p>
					</div>
				{:else}
					<ul class="divide-y divide-gray-200">
						{#each campaigns as campaign}
							<li class="px-4 py-4 sm:px-6">
								<div class="flex items-center justify-between">
									<div class="flex-1">
										<div class="flex items-center">
											<div class="flex-shrink-0">
												<div class="h-10 w-10 rounded-full bg-purple-100 flex items-center justify-center">
													<svg class="h-6 w-6 text-purple-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
													</svg>
												</div>
											</div>
											<div class="ml-4">
												<div class="flex items-center">
													<p class="text-sm font-medium text-gray-900">{campaign.name}</p>
													<span class="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusColor(campaign.status)}">
														{campaign.status}
													</span>
												</div>
												<div class="mt-1 text-sm text-gray-500">
													<p>{campaign.description || 'No description'}</p>
													<p>Budget: {formatCurrency(campaign.budget)} • {formatDate(campaign.start_date)} - {campaign.end_date ? formatDate(campaign.end_date) : 'Ongoing'}</p>
												</div>
											</div>
										</div>
									</div>
									<div class="flex space-x-2">
										{#if campaign.status === 'pending'}
											<button
												on:click={() => approveCampaign(campaign.id)}
												class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
											>
												Approve
											</button>
											<button
												on:click={() => rejectCampaign(campaign.id)}
												class="inline-flex items-center px-3 py-2 border border-transparent text-sm leading-4 font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
											>
												Reject
											</button>
										{:else}
											<span class="text-sm text-gray-500">
												{campaign.status === 'approved' ? 'Approved' : 'Rejected'}
											</span>
										{/if}
									</div>
								</div>
								{#if campaign.approval_notes}
									<div class="mt-2 text-sm text-gray-600">
										<strong>Notes:</strong> {campaign.approval_notes}
									</div>
								{/if}
							</li>
						{/each}
					</ul>
				{/if}
			</div>
		{/if}
	</div>
</div> 