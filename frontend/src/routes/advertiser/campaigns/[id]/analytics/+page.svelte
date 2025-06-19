<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/auth';
	import type { AdCampaign, CampaignAnalytics, AdAnalytics } from '$lib/types/advertising';
	
	let campaign: AdCampaign | null = null;
	let analytics: CampaignAnalytics | null = null;
	let dailyAnalytics: AdAnalytics[] = [];
	let loading = true;
	let error: string | null = null;
	
	let dateRange = {
		start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0], // 30 days ago
		end: new Date().toISOString().split('T')[0] // today
	};

	$: campaignId = parseInt($page.params.id);

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
		await Promise.all([
			loadCampaign(),
			loadAnalytics()
		]);
	}

	async function loadCampaign() {
		const response = await fetch(`/api/v1/advertiser/campaigns/${campaignId}`, {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (response.ok) {
			const data = await response.json();
			campaign = data.data;
		} else {
			throw new Error('Failed to load campaign');
		}
	}

	async function loadAnalytics() {
		const response = await fetch(`/api/v1/advertiser/campaigns/${campaignId}/analytics?start_date=${dateRange.start}&end_date=${dateRange.end}`, {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (response.ok) {
			const data = await response.json();
			analytics = data.data;
		} else {
			// Use mock data for demonstration
			analytics = {
				total_impressions: 15420,
				total_clicks: 342,
				total_unique_views: 12890,
				total_revenue: 856.30,
				ctr: 2.22
			};
		}

		// Load daily analytics (mock data for now)
		dailyAnalytics = generateMockDailyAnalytics();
	}

	function generateMockDailyAnalytics(): AdAnalytics[] {
		const days = [];
		const startDate = new Date(dateRange.start);
		const endDate = new Date(dateRange.end);
		
		for (let d = new Date(startDate); d <= endDate; d.setDate(d.getDate() + 1)) {
			days.push({
				id: Math.floor(Math.random() * 10000),
				ad_id: 1,
				date: d.toISOString().split('T')[0],
				impressions: Math.floor(Math.random() * 1000) + 100,
				clicks: Math.floor(Math.random() * 50) + 5,
				unique_views: Math.floor(Math.random() * 800) + 80,
				view_duration: Math.floor(Math.random() * 300) + 30,
				revenue: Math.random() * 50 + 10,
				created_at: d.toISOString(),
				updated_at: d.toISOString()
			});
		}
		return days.reverse(); // Most recent first
	}

	async function updateDateRange() {
		loading = true;
		try {
			await loadAnalytics();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to update analytics';
		} finally {
			loading = false;
		}
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function formatNumber(num: number): string {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function formatPercentage(num: number): string {
		return `${num.toFixed(2)}%`;
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString();
	}

	function calculateCTR(clicks: number, impressions: number): number {
		return impressions > 0 ? (clicks / impressions) * 100 : 0;
	}
</script>

<svelte:head>
	<title>Campaign Analytics - {campaign?.name || 'Loading...'}</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
		{#if loading && !campaign}
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
		{:else if campaign}
			<!-- Header -->
			<div class="mb-8">
				<nav class="flex mb-4" aria-label="Breadcrumb">
					<ol class="flex items-center space-x-4">
						<li>
							<div>
								<a href="/advertiser" class="text-gray-400 hover:text-gray-500">
									<svg class="flex-shrink-0 h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
										<path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
									</svg>
									<span class="sr-only">Dashboard</span>
								</a>
							</div>
						</li>
						<li>
							<div class="flex items-center">
								<svg class="flex-shrink-0 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
								</svg>
								<a href="/advertiser/campaigns" class="ml-4 text-sm font-medium text-gray-500 hover:text-gray-700">Campaigns</a>
							</div>
						</li>
						<li>
							<div class="flex items-center">
								<svg class="flex-shrink-0 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
								</svg>
								<span class="ml-4 text-sm font-medium text-gray-500" aria-current="page">Analytics</span>
							</div>
						</li>
					</ol>
				</nav>

				<div class="md:flex md:items-center md:justify-between">
					<div class="flex-1 min-w-0">
						<h1 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
							{campaign.name} - Analytics
						</h1>
						<p class="mt-1 text-sm text-gray-500">
							Performance metrics and insights for your campaign
						</p>
					</div>
				</div>
			</div>

			<!-- Date Range Selector -->
			<div class="mb-6 bg-white shadow rounded-lg p-4">
				<div class="flex items-center space-x-4">
					<div>
						<label for="start-date" class="block text-sm font-medium text-gray-700">Start Date</label>
						<input
							type="date"
							id="start-date"
							bind:value={dateRange.start}
							class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
						/>
					</div>
					<div>
						<label for="end-date" class="block text-sm font-medium text-gray-700">End Date</label>
						<input
							type="date"
							id="end-date"
							bind:value={dateRange.end}
							class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
						/>
					</div>
					<div class="flex items-end">
						<button
							on:click={updateDateRange}
							class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						>
							Update
						</button>
					</div>
				</div>
			</div>

			{#if analytics}
				<!-- Key Metrics -->
				<div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
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
										<dd class="text-lg font-medium text-gray-900">{formatNumber(analytics.total_impressions)}</dd>
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
										<dd class="text-lg font-medium text-gray-900">{formatNumber(analytics.total_clicks)}</dd>
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
										<dt class="text-sm font-medium text-gray-500 truncate">Click-Through Rate</dt>
										<dd class="text-lg font-medium text-gray-900">{formatPercentage(analytics.ctr)}</dd>
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
										<dt class="text-sm font-medium text-gray-500 truncate">Revenue Generated</dt>
										<dd class="text-lg font-medium text-gray-900">{formatCurrency(analytics.total_revenue)}</dd>
									</dl>
								</div>
							</div>
						</div>
					</div>
				</div>

				<!-- Daily Performance Table -->
				<div class="bg-white shadow overflow-hidden sm:rounded-md">
					<div class="px-4 py-5 sm:px-6 border-b border-gray-200">
						<h3 class="text-lg leading-6 font-medium text-gray-900">Daily Performance</h3>
						<p class="mt-1 max-w-2xl text-sm text-gray-500">Detailed day-by-day analytics for the selected period</p>
					</div>

					<div class="overflow-x-auto">
						<table class="min-w-full divide-y divide-gray-200">
							<thead class="bg-gray-50">
								<tr>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Date</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Impressions</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Clicks</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CTR</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Unique Views</th>
									<th scope="col" class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Revenue</th>
								</tr>
							</thead>
							<tbody class="bg-white divide-y divide-gray-200">
								{#each dailyAnalytics.slice(0, 10) as day}
									<tr>
										<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
											{formatDate(day.date)}
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatNumber(day.impressions)}
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatNumber(day.clicks)}
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatPercentage(calculateCTR(day.clicks, day.impressions))}
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatNumber(day.unique_views)}
										</td>
										<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
											{formatCurrency(day.revenue)}
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div> 