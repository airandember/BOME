<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
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

		// Check if user has advertiser role
		if ($auth.user?.role !== 'advertiser' && $auth.user?.role !== 'admin') {
			goto('/advertise'); // Redirect to advertiser sign-up page
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
		// For mock users, create mock advertiser account data
		if ($auth.token?.startsWith('mock-advertiser-token-')) {
			// Create mock advertiser account for testing
			advertiserAccount = {
				id: 1,
				user_id: $auth.user?.id || 3,
				company_name: 'Mock Business Company',
				business_email: $auth.user?.email || 'business@bome.test',
				contact_name: `${$auth.user?.firstName} ${$auth.user?.lastName}`,
				contact_phone: '+1 (555) 123-4567',
				business_address: '123 Business St, Business City, BC 12345',
				website: 'https://mockbusiness.com',
				industry: 'Education & Research',
				status: 'approved',
				verification_notes: 'Mock account for testing',
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};

			// Create mock campaigns
			campaigns = [
				{
					id: 1,
					advertiser_id: 1,
					name: 'Book of Mormon Research Campaign',
					description: 'Promoting educational content about Book of Mormon evidence',
					status: 'active',
					start_date: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
					end_date: new Date(Date.now() + 60 * 24 * 60 * 60 * 1000).toISOString(),
					budget: 5000,
					spent_amount: 1250.75,
					target_audience: 'Religious scholars, researchers, students',
					billing_type: 'monthly',
					billing_rate: 2.50,
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				},
				{
					id: 2,
					advertiser_id: 1,
					name: 'Historical Evidence Promotion',
					description: 'Showcasing archaeological and historical evidence',
					status: 'pending',
					start_date: new Date(Date.now() + 7 * 24 * 60 * 60 * 1000).toISOString(),
					end_date: new Date(Date.now() + 90 * 24 * 60 * 60 * 1000).toISOString(),
					budget: 3000,
					spent_amount: 0,
					target_audience: 'History enthusiasts, academics',
					billing_type: 'weekly',
					billing_rate: 1.75,
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				}
			];

			// Set mock analytics
			analytics = {
				totalImpressions: 45230,
				totalClicks: 1876,
				totalSpent: 1250.75,
				activeCampaigns: campaigns.filter(c => c.status === 'active').length
			};

			return;
		}

		// Check if user has advertiser account (for real API)
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

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function formatNumber(num: number): string {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function getStatusClass(status: string): string {
		switch (status) {
			case 'active':
				return 'status-success';
			case 'pending':
				return 'status-pending';
			case 'paused':
				return 'status-secondary';
			case 'rejected':
				return 'status-error';
			default:
				return 'status-secondary';
		}
	}
</script>

<svelte:head>
	<title>Advertiser Dashboard - BOME</title>
</svelte:head>

<Navigation />

<div class="page-container">
	<div class="content-wrapper">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
			</div>
		{:else if error}
			<div class="error-card">
				<div class="error-icon">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10" />
						<path d="M15 9l-6 6" />
						<path d="M9 9l6 6" />
					</svg>
				</div>
				<div class="error-content">
					<h3>Error loading advertiser data</h3>
					<p>{error}</p>
				</div>
			</div>
		{:else if !advertiserAccount}
			<!-- No advertiser account setup -->
			<div class="welcome-section">
				<div class="welcome-content">
					<div class="welcome-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M7 8h10M7 12h4m1 8l-4-4H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-3l-4 4z" />
						</svg>
					</div>
					<h1>Start Advertising on BOME</h1>
					<p>Reach our engaged audience of Book of Mormon researchers and enthusiasts. Create your advertiser account to start running targeted campaigns.</p>
					
					<div class="welcome-features">
						<div class="feature-item">
							<div class="feature-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									<path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
								</svg>
							</div>
							<div class="feature-content">
								<h3>Targeted Audience</h3>
								<p>Reach engaged viewers interested in Book of Mormon evidence and research</p>
							</div>
						</div>
						<div class="feature-item">
							<div class="feature-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
								</svg>
							</div>
							<div class="feature-content">
								<h3>Detailed Analytics</h3>
								<p>Track impressions, clicks, and ROI with comprehensive reporting</p>
							</div>
						</div>
						<div class="feature-item">
							<div class="feature-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M12 2L2 7l10 5 10-5-10-5z" />
									<path d="M2 17l10 5 10-5" />
									<path d="M2 12l10 5 10-5" />
								</svg>
							</div>
							<div class="feature-content">
								<h3>Multiple Placements</h3>
								<p>Choose from 21 strategic placement locations across the platform</p>
							</div>
						</div>
					</div>

					<button class="btn btn-primary btn-lg" on:click={handleCreateAccount}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
						</svg>
						Create Advertiser Account
					</button>
				</div>
			</div>
		{:else}
			<!-- Advertiser dashboard -->
			<div class="dashboard-header">
				<div class="header-content">
					<div class="header-info">
						<h1>Advertiser Dashboard</h1>
						<p>Welcome back, {advertiserAccount.company_name}</p>
					</div>
					<div class="header-actions">
						<button class="btn btn-primary" on:click={handleCreateCampaign}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
							</svg>
							New Campaign
						</button>
					</div>
				</div>
			</div>

			<!-- Account Status -->
			{#if advertiserAccount.status !== 'approved'}
				<div class="status-alert">
					<div class="alert-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="alert-content">
						<h3>
							Account {advertiserAccount.status === 'pending' ? 'Pending Approval' : 'Needs Attention'}
						</h3>
						<div class="alert-description">
							{#if advertiserAccount.status === 'pending'}
								<p>Your advertiser account is currently under review. You'll be able to create campaigns once approved.</p>
							{:else if advertiserAccount.status === 'rejected'}
								<p>Your account application was rejected. Please contact support for more information.</p>
								{#if advertiserAccount.verification_notes}
									<p><strong>Notes:</strong> {advertiserAccount.verification_notes}</p>
								{/if}
							{/if}
						</div>
					</div>
				</div>
			{/if}

			<!-- Analytics Overview -->
			<div class="analytics-section">
				<h2>Performance Overview</h2>
				<div class="analytics-grid">
					<div class="analytics-card">
						<div class="analytics-icon impressions">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								<path d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
							</svg>
						</div>
						<div class="analytics-data">
							<h3>{formatNumber(analytics.totalImpressions)}</h3>
							<p>Total Impressions</p>
						</div>
					</div>

					<div class="analytics-card">
						<div class="analytics-icon clicks">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M15 15l-2 5L9 9l11 4-5 2zm0 0l5 5M7.188 2.239l.777 2.897M5.136 7.965l-2.898-.777M13.95 4.05l-2.122 2.122m-5.657 5.656l-2.12 2.122" />
							</svg>
						</div>
						<div class="analytics-data">
							<h3>{formatNumber(analytics.totalClicks)}</h3>
							<p>Total Clicks</p>
						</div>
					</div>

					<div class="analytics-card">
						<div class="analytics-icon revenue">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
							</svg>
						</div>
						<div class="analytics-data">
							<h3>{formatCurrency(analytics.totalSpent)}</h3>
							<p>Total Spent</p>
						</div>
					</div>

					<div class="analytics-card">
						<div class="analytics-icon campaigns">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
						<div class="analytics-data">
							<h3>{analytics.activeCampaigns}</h3>
							<p>Active Campaigns</p>
						</div>
					</div>
				</div>
			</div>

			<!-- Quick Actions -->
			<div class="actions-section">
				<h2>Quick Actions</h2>
				<div class="actions-grid">
					<a href="/advertiser/campaigns" class="action-card">
						<div class="action-icon campaigns">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
						<div class="action-content">
							<h3>Manage Campaigns</h3>
							<p>View, edit, and monitor your advertising campaigns</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6" />
							</svg>
						</div>
					</a>

					<a href="/advertiser/analytics" class="action-card">
						<div class="action-icon analytics">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
						</div>
						<div class="action-content">
							<h3>View Analytics</h3>
							<p>Detailed performance metrics and insights for your ads</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6" />
							</svg>
						</div>
					</a>

					<a href="/advertiser/account" class="action-card">
						<div class="action-icon settings">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
								<path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
							</svg>
						</div>
						<div class="action-content">
							<h3>Account Settings</h3>
							<p>Update your business information and billing details</p>
						</div>
						<div class="action-arrow">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M9 18l6-6-6-6" />
							</svg>
						</div>
					</a>
				</div>
			</div>

			<!-- Recent Campaigns -->
			{#if campaigns.length > 0}
				<div class="campaigns-section">
					<div class="section-header">
						<div class="header-info">
							<h2>Recent Campaigns</h2>
							<p>Your latest advertising campaigns and their status</p>
						</div>
						<div class="header-actions">
							<a href="/advertiser/campaigns" class="btn btn-outline">View All</a>
						</div>
					</div>
					
					<div class="campaigns-table">
						<div class="table-header">
							<div class="header-cell">Campaign</div>
							<div class="header-cell">Status</div>
							<div class="header-cell">Budget</div>
							<div class="header-cell">Spent</div>
							<div class="header-cell">Actions</div>
						</div>
						{#each campaigns.slice(0, 5) as campaign}
							<div class="table-row">
								<div class="table-cell campaign-info">
									<div class="campaign-details">
										<h4>{campaign.name}</h4>
										<p>{campaign.description || 'No description'}</p>
									</div>
								</div>
								<div class="table-cell">
									<span class="status-badge {getStatusClass(campaign.status)}">
										{campaign.status}
									</span>
								</div>
								<div class="table-cell">{formatCurrency(campaign.budget || 0)}</div>
								<div class="table-cell">{formatCurrency(campaign.spent_amount || 0)}</div>
								<div class="table-cell">
									<a href="/advertiser/campaigns/{campaign.id}" class="btn btn-sm btn-outline">
										View
									</a>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		{/if}
	</div>
</div>

<Footer />

<style>
	.page-container {
		min-height: 100vh;
		background: var(--bg-primary);
		padding: var(--space-2xl) 0;
	}

	.content-wrapper {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 var(--space-lg);
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.loading-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-card {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-2xl);
		border: 1px solid var(--error);
		display: flex;
		align-items: flex-start;
		gap: var(--space-lg);
	}

	.error-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		background: var(--error);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		flex-shrink: 0;
	}

	.error-icon svg {
		width: 24px;
		height: 24px;
	}

	.error-content h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.error-content p {
		color: var(--text-secondary);
		margin: 0;
	}

	/* Welcome Section */
	.welcome-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-4xl);
		text-align: center;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.welcome-content {
		max-width: 800px;
		margin: 0 auto;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2xl);
	}

	.welcome-icon {
		width: 80px;
		height: 80px;
		border-radius: var(--radius-xl);
		background: var(--primary-gradient);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		box-shadow: var(--shadow-lg);
	}

	.welcome-icon svg {
		width: 40px;
		height: 40px;
	}

	.welcome-content h1 {
		font-size: var(--text-4xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
		font-family: var(--font-display);
	}

	.welcome-content > p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
		line-height: 1.6;
		margin: 0;
	}

	.welcome-features {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: var(--space-xl);
		width: 100%;
	}

	.feature-item {
		display: flex;
		align-items: flex-start;
		gap: var(--space-lg);
		text-align: left;
	}

	.feature-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass-dark);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--primary);
		flex-shrink: 0;
	}

	.feature-icon svg {
		width: 24px;
		height: 24px;
	}

	.feature-content h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.feature-content p {
		color: var(--text-secondary);
		margin: 0;
	}

	/* Dashboard Header */
	.dashboard-header {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-2xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-lg);
	}

	.header-info h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
		font-family: var(--font-display);
	}

	.header-info p {
		color: var(--text-secondary);
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
	}

	/* Status Alert */
	.status-alert {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid var(--warning);
		display: flex;
		align-items: flex-start;
		gap: var(--space-lg);
	}

	.alert-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		background: var(--warning);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		flex-shrink: 0;
	}

	.alert-icon svg {
		width: 24px;
		height: 24px;
	}

	.alert-content h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-md) 0;
	}

	.alert-description p {
		color: var(--text-secondary);
		margin: 0 0 var(--space-sm) 0;
	}

	.alert-description p:last-child {
		margin-bottom: 0;
	}

	/* Analytics Section */
	.analytics-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-2xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.analytics-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xl) 0;
	}

	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-lg);
	}

	.analytics-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.05);
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.analytics-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		flex-shrink: 0;
	}

	.analytics-icon.impressions { background: var(--primary-gradient); }
	.analytics-icon.clicks { background: var(--secondary-gradient); }
	.analytics-icon.revenue { background: var(--success); }
	.analytics-icon.campaigns { background: var(--info); }

	.analytics-icon svg {
		width: 24px;
		height: 24px;
	}

	.analytics-data h3 {
		font-size: var(--text-xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.analytics-data p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	/* Actions Section */
	.actions-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-2xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.actions-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xl) 0;
	}

	.actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
		gap: var(--space-lg);
	}

	.action-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.05);
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		text-decoration: none;
		transition: all var(--transition-normal);
		position: relative;
		overflow: hidden;
	}

	.action-card:hover {
		background: var(--bg-glass);
		border-color: var(--primary);
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.action-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		flex-shrink: 0;
		transition: transform var(--transition-normal);
	}

	.action-icon.campaigns { background: var(--primary-gradient); }
	.action-icon.analytics { background: var(--success); }
	.action-icon.settings { background: var(--secondary-gradient); }

	.action-card:hover .action-icon {
		transform: scale(1.1);
	}

	.action-icon svg {
		width: 24px;
		height: 24px;
	}

	.action-content {
		flex: 1;
		min-width: 0;
	}

	.action-content h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.action-content p {
		color: var(--text-secondary);
		margin: 0;
		font-size: var(--text-sm);
	}

	.action-arrow {
		width: 24px;
		height: 24px;
		color: var(--text-secondary);
		flex-shrink: 0;
		transition: all var(--transition-normal);
	}

	.action-card:hover .action-arrow {
		color: var(--primary);
		transform: translateX(4px);
	}

	.action-arrow svg {
		width: 100%;
		height: 100%;
	}

	/* Campaigns Section */
	.campaigns-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-2xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl);
		gap: var(--space-lg);
	}

	.section-header .header-info h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.section-header .header-info p {
		color: var(--text-secondary);
		margin: 0;
	}

	.campaigns-table {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		overflow: hidden;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: rgba(255, 255, 255, 0.02);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.header-cell {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr;
		gap: var(--space-md);
		padding: var(--space-lg);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		transition: background-color var(--transition-normal);
	}

	.table-row:hover {
		background: rgba(255, 255, 255, 0.02);
	}

	.table-row:last-child {
		border-bottom: none;
	}

	.table-cell {
		display: flex;
		align-items: center;
		font-size: var(--text-sm);
		color: var(--text-primary);
	}

	.campaign-info {
		flex-direction: column;
		align-items: flex-start;
		gap: var(--space-xs);
	}

	.campaign-details h4 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.campaign-details p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-success {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-pending {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.status-secondary {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}

	.status-error {
		background: var(--error-light);
		color: var(--error-dark);
	}

	/* Responsive Design */
	@media (max-width: 1024px) {
		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.actions-grid {
			grid-template-columns: 1fr;
		}

		.analytics-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 768px) {
		.page-container {
			padding: var(--space-lg) 0;
		}

		.content-wrapper {
			padding: 0 var(--space-md);
			gap: var(--space-xl);
		}

		.welcome-section {
			padding: var(--space-2xl);
		}

		.welcome-content h1 {
			font-size: var(--text-3xl);
		}

		.welcome-features {
			grid-template-columns: 1fr;
		}

		.analytics-grid {
			grid-template-columns: 1fr;
		}

		.table-header,
		.table-row {
			grid-template-columns: 1fr;
			gap: var(--space-sm);
		}

		.table-header {
			display: none;
		}

		.table-cell {
			padding: var(--space-sm) 0;
			border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		}

		.table-cell:last-child {
			border-bottom: none;
		}

		.section-header {
			flex-direction: column;
			align-items: stretch;
		}
	}

	@media (max-width: 480px) {
		.dashboard-header,
		.analytics-section,
		.actions-section,
		.campaigns-section {
			padding: var(--space-lg);
		}

		.action-card {
			padding: var(--space-lg);
		}
	}
</style> 