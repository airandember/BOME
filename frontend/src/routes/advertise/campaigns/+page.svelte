<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import CampaignCreator from '$lib/components/advertiser/CampaignCreator.svelte';
	import type { AdvertiserAccount, AdvertiserPackage } from '$lib/types/advertising';

	let user: any = null;
	let isAuthenticated = false;
	let loading = true;
	let advertiserAccount: AdvertiserAccount | null = null;
	let selectedPackage: AdvertiserPackage | null = null;
	let hasAccess = false;

	onMount(async () => {
		auth.subscribe((state) => {
			user = state.user;
			isAuthenticated = state.isAuthenticated;
		});

		if (!isAuthenticated) {
			goto('/login?redirect=/advertise/campaigns');
			return;
		}

		// Check if user has an approved advertiser account
		await checkAdvertiserAccess();
		loading = false;
	});

	async function checkAdvertiserAccess() {
		try {
			// Mock API call to check advertiser status
			// In production, this would be: const response = await fetch('/api/v1/advertisers/me');
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			// Mock approved advertiser account for testing
			// In production, this would come from the API response
			advertiserAccount = {
				id: 1,
				user_id: user?.id || 1,
				company_name: 'Test Company',
				business_email: 'business@testcompany.com',
				contact_name: 'John Doe',
				contact_phone: '+1 (555) 123-4567',
				business_address: '123 Main St, City, State 12345',
				tax_id: '12-3456789',
				website: 'https://www.testcompany.com',
				industry: 'technology',
				status: 'approved', // This is the key - must be 'approved'
				approved_by: 1,
				approved_at: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(), // Approved yesterday
				created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			};

			// Mock selected package - in production, this would come from the advertiser's subscription
			selectedPackage = {
				id: 2,
				name: 'Professional',
				description: 'Ideal for growing businesses with comprehensive advertising needs',
				price: 599,
				billing_cycle: 'monthly',
				features: [],
				limits: {
					max_campaigns: 10,
					max_ads_per_campaign: 15,
					max_monthly_impressions: 200000,
					max_file_size_mb: 25,
					max_storage_gb: 5,
					allowed_ad_types: ['banner', 'large', 'small'],
					allowed_placements: ['articles-header', 'articles-sidebar', 'videos-header', 'videos-sidebar', 'events-header'],
					priority_boost: 2,
					analytics_retention_days: 90,
					support_level: 'priority'
				},
				is_active: true,
				is_featured: true,
				sort_order: 2,
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};

			// Check if advertiser is approved
			hasAccess = advertiserAccount.status === 'approved';
		} catch (error) {
			console.error('Failed to check advertiser access:', error);
			hasAccess = false;
		}
	}

	function handleCampaignCreated(event: CustomEvent) {
		const campaignData = event.detail;
		console.log('Campaign created successfully:', campaignData);
		
		// Show success message and redirect to advertiser dashboard
		alert('Campaign created successfully! Your campaign has been submitted for approval. You will receive an email notification once it has been reviewed.');
		
		// In a real implementation, you might redirect to the advertiser dashboard
		goto('/advertise/dashboard');
	}

	function handleGoBack() {
		goto('/advertise/dashboard');
	}
</script>

<svelte:head>
	<title>Create Campaign - BOME Advertiser Portal</title>
	<meta name="description" content="Create and manage your advertising campaigns on the Book of Mormon Evidence platform." />
</svelte:head>

<Navigation />

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Checking advertiser access...</p>
	</div>
{:else if !hasAccess}
	<div class="access-denied">
		<div class="access-denied-content glass">
			<div class="access-icon">
				<svg width="64" height="64" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
					<path d="M12 9V13M12 17H12.01M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
				</svg>
			</div>
			
			<h1>Advertiser Approval Required</h1>
			<p>You need to be an approved advertiser to create campaigns. Please complete the advertiser registration process first.</p>
			
			{#if !advertiserAccount}
				<div class="access-actions">
					<a href="/advertise" class="register-btn">
						Register as Advertiser
					</a>
				</div>
			{:else if advertiserAccount.status === 'pending'}
				<div class="pending-status">
					<h3>Application Status: Pending Review</h3>
					<p>Your advertiser application is currently being reviewed. You'll receive an email notification once it's approved.</p>
					<div class="status-details">
						<div class="detail-item">
							<span class="label">Company:</span>
							<span class="value">{advertiserAccount.company_name}</span>
						</div>
						<div class="detail-item">
							<span class="label">Submitted:</span>
							<span class="value">{new Date(advertiserAccount.created_at).toLocaleDateString()}</span>
						</div>
					</div>
				</div>
			{:else if advertiserAccount.status === 'rejected'}
				<div class="rejected-status">
					<h3>Application Status: Rejected</h3>
					<p>Unfortunately, your advertiser application was not approved. Please contact our support team for more information.</p>
					<div class="access-actions">
						<a href="mailto:advertising@bome.org" class="contact-btn">
							Contact Support
						</a>
					</div>
				</div>
			{/if}
			
			<div class="back-action">
				<a href="/dashboard" class="back-link">← Return to Dashboard</a>
			</div>
		</div>
	</div>
{:else}
	<div class="campaign-page">
		<!-- Header Section -->
		<div class="campaign-header">
			<div class="header-content">
				<h1>Create New Campaign</h1>
				<p>Design and launch your advertising campaign to reach the BOME community.</p>
				
				<div class="account-info glass">
					<div class="account-details">
						<h3>{advertiserAccount?.company_name}</h3>
						<p>Package: {selectedPackage?.name} - ${selectedPackage?.price}/month</p>
					</div>
					<div class="account-status">
						<span class="status-badge approved">Approved</span>
					</div>
				</div>
			</div>
		</div>

		<!-- Campaign Creator -->
		<div class="campaign-content">
			<CampaignCreator 
				{selectedPackage} 
				{advertiserAccount}
				on:campaignCreated={handleCampaignCreated}
				on:goBack={handleGoBack}
			/>
		</div>
	</div>
{/if}

<Footer />

<style>
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: var(--space-lg);
	}

	.loading-container p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
	}

	.access-denied {
		min-height: 60vh;
		display: flex;
		align-items: center;
		justify-content: center;
		padding: var(--space-2xl) var(--space-lg);
	}

	.access-denied-content {
		max-width: 600px;
		padding: var(--space-3xl);
		border-radius: var(--radius-xl);
		text-align: center;
	}

	.access-icon {
		margin: 0 auto var(--space-xl);
		width: 64px;
		height: 64px;
		color: var(--warning);
	}

	.access-denied-content h1 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.access-denied-content > p {
		color: var(--text-secondary);
		margin-bottom: var(--space-2xl);
		line-height: 1.6;
	}

	.access-actions {
		margin-bottom: var(--space-2xl);
	}

	.register-btn, .contact-btn {
		display: inline-block;
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		text-decoration: none;
		padding: var(--space-md) var(--space-xl);
		border-radius: var(--radius-md);
		font-weight: 600;
		transition: all var(--transition-normal);
	}

	.register-btn:hover, .contact-btn:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.pending-status, .rejected-status {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
		margin-bottom: var(--space-2xl);
		text-align: left;
	}

	.pending-status h3, .rejected-status h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
		text-align: center;
	}

	.pending-status p, .rejected-status p {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
		text-align: center;
		line-height: 1.5;
	}

	.status-details {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.detail-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm) 0;
		border-bottom: 1px solid var(--border-color);
	}

	.detail-item:last-child {
		border-bottom: none;
	}

	.detail-item .label {
		color: var(--text-secondary);
		font-weight: 500;
	}

	.detail-item .value {
		color: var(--text-primary);
		font-weight: 600;
	}

	.back-action {
		border-top: 1px solid var(--border-color);
		padding-top: var(--space-lg);
	}

	.back-link {
		color: var(--primary);
		text-decoration: none;
		font-weight: 500;
	}

	.back-link:hover {
		text-decoration: underline;
	}

	.campaign-page {
		min-height: 80vh;
	}

	.campaign-header {
		background: linear-gradient(135deg, var(--primary-light) 0%, var(--secondary-light) 100%);
		padding: var(--space-3xl) var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.header-content {
		max-width: 1200px;
		margin: 0 auto;
	}

	.campaign-header h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
		text-align: center;
	}

	.campaign-header > p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
		text-align: center;
		margin-bottom: var(--space-2xl);
	}

	.account-info {
		max-width: 600px;
		margin: 0 auto;
		padding: var(--space-xl);
		border-radius: var(--radius-lg);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.account-details h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.account-details p {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.approved {
		background: rgba(var(--success-rgb), 0.1);
		color: var(--success);
	}

	.campaign-content {
		max-width: 1200px;
		margin: 0 auto;
		padding: 0 var(--space-lg) var(--space-3xl);
	}

	@media (max-width: 768px) {
		.campaign-header {
			padding: var(--space-2xl) var(--space-lg);
		}

		.campaign-header h1 {
			font-size: var(--text-2xl);
		}

		.account-info {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
		}

		.access-denied-content {
			padding: var(--space-2xl) var(--space-lg);
		}

		.status-details {
			gap: var(--space-sm);
		}
	}
</style> 