<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdCampaign, AdvertiserAccount } from '$lib/types/advertising';
	
	let campaigns: AdCampaign[] = [];
	let advertiserAccount: AdvertiserAccount | null = null;
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
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
		// Check advertiser account
		const accountResponse = await fetch('/api/v1/advertiser/account', {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (!accountResponse.ok) {
			throw new Error('No advertiser account found');
		}

		const accountData = await accountResponse.json();
		advertiserAccount = accountData.data;

		if (advertiserAccount?.status !== 'approved') {
			return; // Don't load campaigns if account not approved
		}

		// Load campaigns
		const campaignsResponse = await fetch('/api/v1/advertiser/campaigns', {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (campaignsResponse.ok) {
			const campaignsData = await campaignsResponse.json();
			campaigns = campaignsData.data || [];
		}
	}

	function handleCreateCampaign() {
		goto('/advertiser/campaigns/new');
	}

	function handleEditCampaign(campaignId: number) {
		goto(`/advertiser/campaigns/${campaignId}/edit`);
	}

	function handleViewAnalytics(campaignId: number) {
		goto(`/advertiser/campaigns/${campaignId}/analytics`);
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'approved':
			case 'active':
				return 'text-green-600 bg-green-50';
			case 'pending':
				return 'text-yellow-600 bg-yellow-50';
			case 'paused':
				return 'text-blue-600 bg-blue-50';
			case 'completed':
				return 'text-gray-600 bg-gray-50';
			case 'rejected':
				return 'text-red-600 bg-red-50';
			default:
				return 'text-gray-600 bg-gray-50';
		}
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString();
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}
</script>

<svelte:head>
	<title>My Campaigns - BOME Advertiser</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		{#if loading}
			<div class="flex justify-center items-center h-64">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		{:else if error}
			<div class="bg-red-50 border border-red-200 rounded-lg p-6">
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
		{:else if !advertiserAccount}
			<div class="text-center">
				<h1 class="text-2xl font-bold text-gray-900 mb-4">No Advertiser Account</h1>
				<p class="text-gray-600 mb-6">You need to create an advertiser account first.</p>
				<a href="/advertiser/setup" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700">
					Create Advertiser Account
				</a>
			</div>
		{:else if advertiserAccount.status !== 'approved'}
			<div class="text-center">
				<h1 class="text-2xl font-bold text-gray-900 mb-4">Account Pending Approval</h1>
				<p class="text-gray-600 mb-6">Your advertiser account is still being reviewed. You'll be able to create campaigns once it's approved.</p>
				<a href="/advertiser" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700">
					Back to Dashboard
				</a>
			</div>
		{:else}
			<!-- Campaigns list -->
			<div class="mb-8">
				<div class="md:flex md:items-center md:justify-between">
					<div class="flex-1 min-w-0">
						<h1 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
							My Campaigns
						</h1>
						<p class="mt-1 text-sm text-gray-500">
							Manage your advertising campaigns
						</p>
					</div>
					<div class="mt-4 flex md:mt-0 md:ml-4">
						<button 
							on:click={handleCreateCampaign}
							class="inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						>
							<svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
							</svg>
							New Campaign
						</button>
					</div>
				</div>
			</div>

			{#if campaigns.length === 0}
				<div class="text-center py-12">
					<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
					</svg>
					<h3 class="mt-2 text-sm font-medium text-gray-900">No campaigns</h3>
					<p class="mt-1 text-sm text-gray-500">Get started by creating your first advertising campaign.</p>
					<div class="mt-6">
						<button 
							on:click={handleCreateCampaign}
							class="inline-flex items-center px-4 py-2 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						>
							<svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
							</svg>
							New Campaign
						</button>
					</div>
				</div>
			{:else}
				<div class="bg-white shadow overflow-hidden sm:rounded-md">
					<ul class="divide-y divide-gray-200">
						{#each campaigns as campaign}
							<li>
								<div class="px-4 py-4 sm:px-6">
									<div class="flex items-center justify-between">
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
													<p class="text-sm font-medium text-gray-900 truncate">
														{campaign.name}
													</p>
													<span class="ml-2 inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusColor(campaign.status)}">
														{campaign.status}
													</span>
												</div>
												<div class="mt-2 flex items-center text-sm text-gray-500">
													<p class="truncate">
														{campaign.description || 'No description'}
													</p>
												</div>
											</div>
										</div>
										<div class="flex items-center space-x-4">
											<div class="text-right">
												<p class="text-sm font-medium text-gray-900">
													{formatCurrency(campaign.budget || 0)}
												</p>
												<p class="text-sm text-gray-500">
													Budget
												</p>
											</div>
											<div class="text-right">
												<p class="text-sm font-medium text-gray-900">
													{formatCurrency(campaign.spent_amount || 0)}
												</p>
												<p class="text-sm text-gray-500">
													Spent
												</p>
											</div>
											<div class="text-right">
												<p class="text-sm font-medium text-gray-900">
													{formatDate(campaign.start_date)}
												</p>
												<p class="text-sm text-gray-500">
													Start Date
												</p>
											</div>
											<div class="flex space-x-2">
												<button
													on:click={() => handleViewAnalytics(campaign.id)}
													class="inline-flex items-center p-2 border border-transparent rounded-full shadow-sm text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500"
													title="View Analytics"
												>
													<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
													</svg>
												</button>
												<button
													on:click={() => handleEditCampaign(campaign.id)}
													class="inline-flex items-center p-2 border border-transparent rounded-full shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
													title="Edit Campaign"
												>
													<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
														<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
													</svg>
												</button>
											</div>
										</div>
									</div>
								</div>
							</li>
						{/each}
					</ul>
				</div>
			{/if}
		{/if}
	</div>
</div> 