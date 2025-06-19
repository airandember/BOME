<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdvertiserAccount, AdCampaign, DashboardAnalytics } from '$lib/types/advertising';
	
	let advertiserAccount: AdvertiserAccount | null = null;
	let campaigns: AdCampaign[] = [];
	let analytics: DashboardAnalytics = {
		totalImpressions: 0,
		totalClicks: 0,
		totalSpent: 0,
		activeCampaigns: 0
	};
	let loading = true;
	let error: string | null = null;

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			await loadAdvertiserData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	});

	async function loadAdvertiserData() {
		// Check if user has advertiser account
		try {
			const accountResponse = await fetch('/api/v1/advertiser/account', {
				headers: {
					'Authorization': `Bearer ${$auth.token}`
				}
			});

			if (accountResponse.ok) {
				const accountData = await accountResponse.json();
				advertiserAccount = accountData.data;
				
				// Load campaigns and analytics
				await loadCampaigns();
				await loadAnalytics();
			}
		} catch (err) {
			// User doesn't have advertiser account yet
			console.log('No advertiser account found');
		}
	}

	async function loadCampaigns() {
		const response = await fetch('/api/v1/advertiser/campaigns', {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (response.ok) {
			const data = await response.json();
			campaigns = data.data || [];
			analytics.activeCampaigns = campaigns.filter((c: AdCampaign) => c.status === 'active').length;
		}
	}

	async function loadAnalytics() {
		// Calculate totals from campaigns
		let totalImpressions = 0;
		let totalClicks = 0;
		let totalSpent = 0;

		for (const campaign of campaigns) {
			try {
				const response = await fetch(`/api/v1/advertiser/campaigns/${campaign.id}/analytics`, {
					headers: {
						'Authorization': `Bearer ${$auth.token}`
					}
				});

				if (response.ok) {
					const data = await response.json();
					totalImpressions += data.data?.total_impressions || 0;
					totalClicks += data.data?.total_clicks || 0;
					totalSpent += campaign.spent_amount || 0;
				}
			} catch (err) {
				console.error(`Failed to load analytics for campaign ${campaign.id}`);
			}
		}

		analytics = {
			...analytics,
			totalImpressions,
			totalClicks,
			totalSpent
		};
	}

	function handleCreateAccount() {
		goto('/advertiser/setup');
	}

	function handleCreateCampaign() {
		goto('/advertiser/campaigns/new');
	}
</script>

<svelte:head>
	<title>Advertiser Dashboard - BOME</title>
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
						<h3 class="text-sm font-medium text-red-800">Error loading advertiser data</h3>
						<div class="mt-2 text-sm text-red-700">
							<p>{error}</p>
						</div>
					</div>
				</div>
			</div>
		{:else if !advertiserAccount}
			<!-- No advertiser account setup -->
			<div class="text-center">
				<div class="mx-auto max-w-md">
					<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
					</svg>
					<h2 class="mt-2 text-lg font-medium text-gray-900">Start Advertising on BOME</h2>
					<p class="mt-1 text-sm text-gray-500">
						Create your advertiser account to start running campaigns and reaching our audience.
					</p>
				</div>
				<div class="mt-6">
					<button 
						on:click={handleCreateAccount}
						class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
					>
						<svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
						</svg>
						Create Advertiser Account
					</button>
				</div>
			</div>
		{:else}
			<!-- Advertiser dashboard -->
			<div class="mb-8">
				<div class="md:flex md:items-center md:justify-between">
					<div class="flex-1 min-w-0">
						<h1 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
							Advertiser Dashboard
						</h1>
						<p class="mt-1 text-sm text-gray-500">
							Welcome back, {advertiserAccount.company_name}
						</p>
					</div>
					<div class="mt-4 flex md:mt-0 md:ml-4">
						<button 
							on:click={handleCreateCampaign}
							class="ml-3 inline-flex items-center px-4 py-2 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						>
							<svg class="-ml-1 mr-2 h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
							</svg>
							New Campaign
						</button>
					</div>
				</div>
			</div>

			<!-- Account Status -->
			{#if advertiserAccount.status !== 'approved'}
				<div class="mb-8">
					<div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
						<div class="flex">
							<div class="flex-shrink-0">
								<svg class="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
								</svg>
							</div>
							<div class="ml-3">
								<h3 class="text-sm font-medium text-yellow-800">
									Account {advertiserAccount.status === 'pending' ? 'Pending Approval' : 'Needs Attention'}
								</h3>
								<div class="mt-2 text-sm text-yellow-700">
									{#if advertiserAccount.status === 'pending'}
										<p>Your advertiser account is currently under review. You'll be able to create campaigns once approved.</p>
									{:else if advertiserAccount.status === 'rejected'}
										<p>Your account application was rejected. Please contact support for more information.</p>
										{#if advertiserAccount.verification_notes}
											<p class="mt-1"><strong>Notes:</strong> {advertiserAccount.verification_notes}</p>
										{/if}
									{/if}
								</div>
							</div>
						</div>
					</div>
				</div>
			{/if}

			<!-- Analytics Overview -->
			<div class="mb-8">
				<h2 class="text-lg font-medium text-gray-900 mb-4">Performance Overview</h2>
				<div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
					<div class="bg-white overflow-hidden shadow rounded-lg">
						<div class="p-5">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
									</svg>
								</div>
								<div class="ml-5 w-0 flex-1">
									<dl>
										<dt class="text-sm font-medium text-gray-500 truncate">Total Impressions</dt>
										<dd class="text-lg font-medium text-gray-900">{analytics.totalImpressions.toLocaleString()}</dd>
									</dl>
								</div>
							</div>
						</div>
					</div>

					<div class="bg-white overflow-hidden shadow rounded-lg">
						<div class="p-5">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
									</svg>
								</div>
								<div class="ml-5 w-0 flex-1">
									<dl>
										<dt class="text-sm font-medium text-gray-500 truncate">Total Clicks</dt>
										<dd class="text-lg font-medium text-gray-900">{analytics.totalClicks.toLocaleString()}</dd>
									</dl>
								</div>
							</div>
						</div>
					</div>

					<div class="bg-white overflow-hidden shadow rounded-lg">
						<div class="p-5">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
									</svg>
								</div>
								<div class="ml-5 w-0 flex-1">
									<dl>
										<dt class="text-sm font-medium text-gray-500 truncate">Total Spent</dt>
										<dd class="text-lg font-medium text-gray-900">${analytics.totalSpent.toFixed(2)}</dd>
									</dl>
								</div>
							</div>
						</div>
					</div>

					<div class="bg-white overflow-hidden shadow rounded-lg">
						<div class="p-5">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<svg class="h-6 w-6 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
									</svg>
								</div>
								<div class="ml-5 w-0 flex-1">
									<dl>
										<dt class="text-sm font-medium text-gray-500 truncate">Active Campaigns</dt>
										<dd class="text-lg font-medium text-gray-900">{analytics.activeCampaigns}</dd>
									</dl>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Quick Actions -->
			<div class="mb-8">
				<h2 class="text-lg font-medium text-gray-900 mb-4">Quick Actions</h2>
				<div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
					<a href="/advertiser/campaigns" class="relative group bg-white p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-blue-500 rounded-lg shadow hover:shadow-md transition-shadow">
						<div>
							<span class="rounded-lg inline-flex p-3 bg-blue-50 text-blue-700 ring-4 ring-white">
								<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
								</svg>
							</span>
						</div>
						<div class="mt-8">
							<h3 class="text-lg font-medium">
								<span class="absolute inset-0" aria-hidden="true"></span>
								Manage Campaigns
							</h3>
							<p class="mt-2 text-sm text-gray-500">
								View, edit, and monitor your advertising campaigns.
							</p>
						</div>
						<span class="pointer-events-none absolute top-6 right-6 text-gray-300 group-hover:text-gray-400" aria-hidden="true">
							<svg class="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
								<path d="M20 4h1a1 1 0 00-1-1v1zm-1 12a1 1 0 102 0h-2zM8 3a1 1 0 000 2V3zM3.293 19.293a1 1 0 101.414 1.414l-1.414-1.414zM19 4v12h2V4h-2zm1-1H8v2h12V3zm-.707.293l-16 16 1.414 1.414 16-16-1.414-1.414z" />
							</svg>
						</span>
					</a>

					<a href="/advertiser/analytics" class="relative group bg-white p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-blue-500 rounded-lg shadow hover:shadow-md transition-shadow">
						<div>
							<span class="rounded-lg inline-flex p-3 bg-green-50 text-green-700 ring-4 ring-white">
								<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
								</svg>
							</span>
						</div>
						<div class="mt-8">
							<h3 class="text-lg font-medium">
								<span class="absolute inset-0" aria-hidden="true"></span>
								View Analytics
							</h3>
							<p class="mt-2 text-sm text-gray-500">
								Detailed performance metrics and insights for your ads.
							</p>
						</div>
						<span class="pointer-events-none absolute top-6 right-6 text-gray-300 group-hover:text-gray-400" aria-hidden="true">
							<svg class="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
								<path d="M20 4h1a1 1 0 00-1-1v1zm-1 12a1 1 0 102 0h-2zM8 3a1 1 0 000 2V3zM3.293 19.293a1 1 0 101.414 1.414l-1.414-1.414zM19 4v12h2V4h-2zm1-1H8v2h12V3zm-.707.293l-16 16 1.414 1.414 16-16-1.414-1.414z" />
							</svg>
						</span>
					</a>

					<a href="/advertiser/account" class="relative group bg-white p-6 focus-within:ring-2 focus-within:ring-inset focus-within:ring-blue-500 rounded-lg shadow hover:shadow-md transition-shadow">
						<div>
							<span class="rounded-lg inline-flex p-3 bg-purple-50 text-purple-700 ring-4 ring-white">
								<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								</svg>
							</span>
						</div>
						<div class="mt-8">
							<h3 class="text-lg font-medium">
								<span class="absolute inset-0" aria-hidden="true"></span>
								Account Settings
							</h3>
							<p class="mt-2 text-sm text-gray-500">
								Update your business information and billing details.
							</p>
						</div>
						<span class="pointer-events-none absolute top-6 right-6 text-gray-300 group-hover:text-gray-400" aria-hidden="true">
							<svg class="h-6 w-6" fill="currentColor" viewBox="0 0 24 24">
								<path d="M20 4h1a1 1 0 00-1-1v1zm-1 12a1 1 0 102 0h-2zM8 3a1 1 0 000 2V3zM3.293 19.293a1 1 0 101.414 1.414l-1.414-1.414zM19 4v12h2V4h-2zm1-1H8v2h12V3zm-.707.293l-16 16 1.414 1.414 16-16-1.414-1.414z" />
							</svg>
						</span>
					</a>
				</div>
			</div>

			<!-- Recent Campaigns -->
			{#if campaigns.length > 0}
				<div class="mb-8">
					<div class="sm:flex sm:items-center">
						<div class="sm:flex-auto">
							<h2 class="text-lg font-medium text-gray-900">Recent Campaigns</h2>
							<p class="mt-1 text-sm text-gray-500">Your latest advertising campaigns and their status.</p>
						</div>
						<div class="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
							<a href="/advertiser/campaigns" class="inline-flex items-center justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 sm:w-auto">
								View All
							</a>
						</div>
					</div>
					<div class="mt-6 flex flex-col">
						<div class="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
							<div class="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
								<div class="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
									<table class="min-w-full divide-y divide-gray-300">
										<thead class="bg-gray-50">
											<tr>
												<th scope="col" class="py-3.5 pl-4 pr-3 text-left text-xs font-medium uppercase tracking-wide text-gray-500 sm:pl-6">Campaign</th>
												<th scope="col" class="px-3 py-3.5 text-left text-xs font-medium uppercase tracking-wide text-gray-500">Status</th>
												<th scope="col" class="px-3 py-3.5 text-left text-xs font-medium uppercase tracking-wide text-gray-500">Budget</th>
												<th scope="col" class="px-3 py-3.5 text-left text-xs font-medium uppercase tracking-wide text-gray-500">Spent</th>
												<th scope="col" class="relative py-3.5 pl-3 pr-4 sm:pr-6">
													<span class="sr-only">Actions</span>
												</th>
											</tr>
										</thead>
										<tbody class="divide-y divide-gray-200 bg-white">
											{#each campaigns.slice(0, 5) as campaign}
												<tr>
													<td class="whitespace-nowrap py-4 pl-4 pr-3 text-sm sm:pl-6">
														<div class="flex items-center">
															<div>
																<div class="font-medium text-gray-900">{campaign.name}</div>
																<div class="text-gray-500">{campaign.description || 'No description'}</div>
															</div>
														</div>
													</td>
													<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
														<span class="inline-flex rounded-full px-2 text-xs font-semibold leading-5 
															{campaign.status === 'active' ? 'bg-green-100 text-green-800' : 
															 campaign.status === 'pending' ? 'bg-yellow-100 text-yellow-800' : 
															 campaign.status === 'paused' ? 'bg-gray-100 text-gray-800' : 
															 'bg-red-100 text-red-800'}">
															{campaign.status}
														</span>
													</td>
													<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
														${campaign.budget?.toFixed(2) || '0.00'}
													</td>
													<td class="whitespace-nowrap px-3 py-4 text-sm text-gray-500">
														${campaign.spent_amount?.toFixed(2) || '0.00'}
													</td>
													<td class="relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-6">
														<a href="/advertiser/campaigns/{campaign.id}" class="text-blue-600 hover:text-blue-900">
															View<span class="sr-only">, {campaign.name}</span>
														</a>
													</td>
												</tr>
											{/each}
										</tbody>
									</table>
								</div>
							</div>
						</div>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div> 