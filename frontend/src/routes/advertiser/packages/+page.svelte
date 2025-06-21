<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import type { AdvertiserPackage, AdvertiserSubscription } from '$lib/types/advertising';
	
	let packages: AdvertiserPackage[] = [];
	let currentSubscription: AdvertiserSubscription | null = null;
	let loading = true;
	let error = '';
	let selectedPackage: AdvertiserPackage | null = null;
	let showUpgradeModal = false;
	let upgrading = false;
	
	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}
		
		await Promise.all([
			loadPackages(),
			loadCurrentSubscription()
		]);
		loading = false;
	});
	
	async function loadPackages() {
		try {
			// Mock data - replace with actual API call
			packages = [
				{
					id: 1,
					name: 'Starter',
					description: 'Perfect for small businesses getting started with advertising',
					price: 29.99,
					billing_cycle: 'monthly',
					features: [
						{ id: 1, package_id: 1, name: 'Basic Analytics', description: 'Essential performance metrics', is_included: true },
						{ id: 2, package_id: 1, name: 'Email Support', description: 'Standard email support', is_included: true },
						{ id: 3, package_id: 1, name: 'Campaign Templates', description: 'Pre-built campaign templates', is_included: true },
						{ id: 4, package_id: 1, name: 'Priority Support', description: '24/7 priority support', is_included: false },
						{ id: 5, package_id: 1, name: 'Advanced Analytics', description: 'Detailed performance insights', is_included: false }
					],
					limits: {
						max_campaigns: 3,
						max_ads_per_campaign: 5,
						max_monthly_impressions: 10000,
						max_file_size_mb: 5,
						max_storage_gb: 1,
						allowed_ad_types: ['banner', 'large', 'small'],
						allowed_placements: ['header', 'sidebar', 'footer'],
						priority_boost: 0,
						analytics_retention_days: 30,
						support_level: 'basic'
					},
					is_active: true,
					is_featured: false,
					sort_order: 1,
					created_at: '2024-01-01T00:00:00Z',
					updated_at: '2024-01-01T00:00:00Z'
				},
				{
					id: 2,
					name: 'Professional',
					description: 'Ideal for growing businesses with advanced advertising needs',
					price: 79.99,
					billing_cycle: 'monthly',
					features: [
						{ id: 6, package_id: 2, name: 'Advanced Analytics', description: 'Detailed performance insights', is_included: true },
						{ id: 7, package_id: 2, name: 'Priority Support', description: '24/7 priority support', is_included: true },
						{ id: 8, package_id: 2, name: 'A/B Testing', description: 'Campaign optimization tools', is_included: true },
						{ id: 9, package_id: 2, name: 'Custom Targeting', description: 'Advanced audience targeting', is_included: true },
						{ id: 10, package_id: 2, name: 'Video Ads', description: 'Video advertisement support', is_included: true },
						{ id: 11, package_id: 2, name: 'White-label Reports', description: 'Branded reporting', is_included: false }
					],
					limits: {
						max_campaigns: 10,
						max_ads_per_campaign: 15,
						max_monthly_impressions: 50000,
						max_file_size_mb: 25,
						max_storage_gb: 5,
						allowed_ad_types: ['banner', 'large', 'small', 'video'],
						allowed_placements: ['header', 'sidebar', 'footer', 'content', 'video_overlay'],
						priority_boost: 2,
						analytics_retention_days: 90,
						support_level: 'priority'
					},
					is_active: true,
					is_featured: true,
					sort_order: 2,
					created_at: '2024-01-01T00:00:00Z',
					updated_at: '2024-01-01T00:00:00Z'
				},
				{
					id: 3,
					name: 'Enterprise',
					description: 'Complete advertising solution for large organizations',
					price: 199.99,
					billing_cycle: 'monthly',
					features: [
						{ id: 12, package_id: 3, name: 'Premium Analytics', description: 'Complete analytics suite', is_included: true },
						{ id: 13, package_id: 3, name: 'Dedicated Support', description: 'Dedicated account manager', is_included: true },
						{ id: 14, package_id: 3, name: 'White-label Reports', description: 'Branded reporting', is_included: true },
						{ id: 15, package_id: 3, name: 'API Access', description: 'Full API integration', is_included: true },
						{ id: 16, package_id: 3, name: 'Custom Integration', description: 'Custom development support', is_included: true },
						{ id: 17, package_id: 3, name: 'Interactive Ads', description: 'Advanced interactive formats', is_included: true }
					],
					limits: {
						max_campaigns: -1, // Unlimited
						max_ads_per_campaign: -1, // Unlimited
						max_monthly_impressions: -1, // Unlimited
						max_file_size_mb: 100,
						max_storage_gb: 50,
						allowed_ad_types: ['banner', 'large', 'small', 'video', 'interactive'],
						allowed_placements: ['header', 'sidebar', 'footer', 'content', 'video_overlay', 'between_videos'],
						priority_boost: 5,
						analytics_retention_days: 365,
						support_level: 'premium'
					},
					is_active: true,
					is_featured: false,
					sort_order: 3,
					created_at: '2024-01-01T00:00:00Z',
					updated_at: '2024-01-01T00:00:00Z'
				}
			];
		} catch (err) {
			error = 'Failed to load packages';
			console.error('Error loading packages:', err);
		}
	}
	
	async function loadCurrentSubscription() {
		try {
			// Mock data - replace with actual API call
			currentSubscription = {
				id: 1,
				advertiser_id: 1,
				package_id: 1,
				stripe_subscription_id: 'sub_1234567890',
				status: 'active',
				current_period_start: '2024-06-01T00:00:00Z',
				current_period_end: '2024-07-01T00:00:00Z',
				cancel_at_period_end: false,
				usage_stats: {
					campaigns_used: 2,
					ads_used: 8,
					storage_used_gb: 0.3,
					monthly_impressions: 3420,
					monthly_clicks: 156
				},
				created_at: '2024-06-01T00:00:00Z',
				updated_at: '2024-06-01T00:00:00Z'
			};
		} catch (err) {
			console.error('Error loading subscription:', err);
		}
	}
	
	function handleSelectPackage(pkg: AdvertiserPackage) {
		if (currentSubscription && currentSubscription.package_id === pkg.id) {
			showToast('You are already subscribed to this package', 'info');
			return;
		}
		
		selectedPackage = pkg;
		showUpgradeModal = true;
	}
	
	async function handleUpgrade() {
		if (!selectedPackage) return;
		
		upgrading = true;
		
		try {
			// Mock API call - replace with actual implementation
			await new Promise(resolve => setTimeout(resolve, 2000));
			
			showToast(`Successfully upgraded to ${selectedPackage.name}!`, 'success');
			showUpgradeModal = false;
			selectedPackage = null;
			
			// Reload subscription data
			await loadCurrentSubscription();
			
		} catch (err) {
			showToast('Failed to upgrade package', 'error');
			console.error('Error upgrading package:', err);
		} finally {
			upgrading = false;
		}
	}
	
	function formatPrice(price: number, cycle: string): string {
		const formatted = new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(price);
		
		return `${formatted}/${cycle}`;
	}
	
	function formatLimit(value: number): string {
		if (value === -1) return 'Unlimited';
		if (value >= 1000000) return `${(value / 1000000).toFixed(1)}M`;
		if (value >= 1000) return `${(value / 1000).toFixed(1)}K`;
		return value.toString();
	}
	
	function getCurrentPackage(): AdvertiserPackage | null {
		if (!currentSubscription) return null;
		return packages.find(pkg => pkg.id === currentSubscription!.package_id) || null;
	}
	
	function getUsagePercentage(used: number, limit: number): number {
		if (limit === -1) return 0; // Unlimited
		return Math.min((used / limit) * 100, 100);
	}
	
	function getUsageColor(percentage: number): string {
		if (percentage >= 90) return 'var(--error)';
		if (percentage >= 75) return 'var(--warning)';
		return 'var(--success)';
	}
</script>

<svelte:head>
	<title>Advertising Packages - BOME Advertiser</title>
</svelte:head>

<div class="packages-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Advertising Packages</h1>
				<p>Choose the perfect package for your advertising needs</p>
			</div>
		</div>
	</div>
	
	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading packages...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={() => location.reload()}>
				Try Again
			</button>
		</div>
	{:else}
		<!-- Current Subscription -->
		{#if currentSubscription}
			{@const currentPackage = getCurrentPackage()}
			{#if currentPackage}
				<div class="current-subscription glass">
					<div class="subscription-header">
						<div class="subscription-info">
							<h2>Current Plan: {currentPackage.name}</h2>
							<p class="subscription-status status-{currentSubscription.status}">
								{currentSubscription.status.toUpperCase()}
							</p>
						</div>
						<div class="subscription-price">
							<span class="price">{formatPrice(currentPackage.price, currentPackage.billing_cycle)}</span>
						</div>
					</div>
					
					<div class="usage-stats">
						<h3>Usage Statistics</h3>
						<div class="stats-grid">
							<div class="stat-item">
								<label>Campaigns</label>
								<div class="usage-bar">
									<div class="usage-fill" style="width: {getUsagePercentage(currentSubscription.usage_stats.campaigns_used, currentPackage.limits.max_campaigns)}%; background: {getUsageColor(getUsagePercentage(currentSubscription.usage_stats.campaigns_used, currentPackage.limits.max_campaigns))}"></div>
								</div>
								<span class="usage-text">
									{currentSubscription.usage_stats.campaigns_used} / {formatLimit(currentPackage.limits.max_campaigns)}
								</span>
							</div>
							
							<div class="stat-item">
								<label>Advertisements</label>
								<div class="usage-bar">
									<div class="usage-fill" style="width: {getUsagePercentage(currentSubscription.usage_stats.ads_used, currentPackage.limits.max_ads_per_campaign * currentPackage.limits.max_campaigns)}%; background: {getUsageColor(getUsagePercentage(currentSubscription.usage_stats.ads_used, currentPackage.limits.max_ads_per_campaign * currentPackage.limits.max_campaigns))}"></div>
								</div>
								<span class="usage-text">
									{currentSubscription.usage_stats.ads_used} / {formatLimit(currentPackage.limits.max_ads_per_campaign * currentPackage.limits.max_campaigns)}
								</span>
							</div>
							
							<div class="stat-item">
								<label>Storage</label>
								<div class="usage-bar">
									<div class="usage-fill" style="width: {getUsagePercentage(currentSubscription.usage_stats.storage_used_gb, currentPackage.limits.max_storage_gb)}%; background: {getUsageColor(getUsagePercentage(currentSubscription.usage_stats.storage_used_gb, currentPackage.limits.max_storage_gb))}"></div>
								</div>
								<span class="usage-text">
									{currentSubscription.usage_stats.storage_used_gb.toFixed(1)} GB / {formatLimit(currentPackage.limits.max_storage_gb)} GB
								</span>
							</div>
							
							<div class="stat-item">
								<label>Monthly Impressions</label>
								<div class="usage-bar">
									<div class="usage-fill" style="width: {getUsagePercentage(currentSubscription.usage_stats.monthly_impressions, currentPackage.limits.max_monthly_impressions)}%; background: {getUsageColor(getUsagePercentage(currentSubscription.usage_stats.monthly_impressions, currentPackage.limits.max_monthly_impressions))}"></div>
								</div>
								<span class="usage-text">
									{formatLimit(currentSubscription.usage_stats.monthly_impressions)} / {formatLimit(currentPackage.limits.max_monthly_impressions)}
								</span>
							</div>
						</div>
					</div>
				</div>
			{/if}
		{/if}
		
		<!-- Available Packages -->
		<div class="packages-section">
			<h2>Available Packages</h2>
			<div class="packages-grid">
				{#each packages as pkg}
					<div class="package-card glass" class:featured={pkg.is_featured} class:current={currentSubscription?.package_id === pkg.id}>
						{#if pkg.is_featured}
							<div class="featured-badge">Most Popular</div>
						{/if}
						
						{#if currentSubscription?.package_id === pkg.id}
							<div class="current-badge">Current Plan</div>
						{/if}
						
						<div class="package-header">
							<h3>{pkg.name}</h3>
							<p class="package-description">{pkg.description}</p>
							<div class="package-price">
								<span class="price">{formatPrice(pkg.price, pkg.billing_cycle)}</span>
							</div>
						</div>
						
						<div class="package-features">
							<h4>Features</h4>
							<ul>
								{#each pkg.features as feature}
									<li class:included={feature.is_included} class:not-included={!feature.is_included}>
										<span class="feature-icon">
											{#if feature.is_included}
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<polyline points="20,6 9,17 4,12"></polyline>
												</svg>
											{:else}
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<line x1="18" y1="6" x2="6" y2="18"></line>
													<line x1="6" y1="6" x2="18" y2="18"></line>
												</svg>
											{/if}
										</span>
										<div class="feature-text">
											<span class="feature-name">{feature.name}</span>
											<span class="feature-description">{feature.description}</span>
										</div>
									</li>
								{/each}
							</ul>
						</div>
						
						<div class="package-limits">
							<h4>Limits</h4>
							<div class="limits-grid">
								<div class="limit-item">
									<label>Campaigns</label>
									<span>{formatLimit(pkg.limits.max_campaigns)}</span>
								</div>
								<div class="limit-item">
									<label>Ads per Campaign</label>
									<span>{formatLimit(pkg.limits.max_ads_per_campaign)}</span>
								</div>
								<div class="limit-item">
									<label>Monthly Impressions</label>
									<span>{formatLimit(pkg.limits.max_monthly_impressions)}</span>
								</div>
								<div class="limit-item">
									<label>File Size</label>
									<span>{pkg.limits.max_file_size_mb} MB</span>
								</div>
								<div class="limit-item">
									<label>Storage</label>
									<span>{pkg.limits.max_storage_gb} GB</span>
								</div>
								<div class="limit-item">
									<label>Support Level</label>
									<span class="support-level support-{pkg.limits.support_level}">
										{pkg.limits.support_level.toUpperCase()}
									</span>
								</div>
							</div>
						</div>
						
						<div class="package-actions">
							{#if currentSubscription?.package_id === pkg.id}
								<button class="btn btn-outline" disabled>
									Current Plan
								</button>
							{:else}
								<button class="btn btn-primary" on:click={() => handleSelectPackage(pkg)}>
									{currentSubscription ? 'Upgrade' : 'Select'} Plan
								</button>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Upgrade Modal -->
{#if showUpgradeModal && selectedPackage}
	<div class="modal-overlay" on:click={() => showUpgradeModal = false}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Upgrade to {selectedPackage.name}</h2>
				<button class="modal-close" on:click={() => showUpgradeModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-body">
				<div class="upgrade-summary">
					<p>You are about to upgrade to the <strong>{selectedPackage.name}</strong> package.</p>
					
					<div class="pricing-info">
						<div class="price-comparison">
							{#if currentSubscription}
								{@const currentPackage = getCurrentPackage()}
								{#if currentPackage}
									<div class="price-item">
										<label>Current Plan</label>
										<span>{formatPrice(currentPackage.price, currentPackage.billing_cycle)}</span>
									</div>
								{/if}
							{/if}
							
							<div class="price-item new-price">
								<label>New Plan</label>
								<span>{formatPrice(selectedPackage.price, selectedPackage.billing_cycle)}</span>
							</div>
						</div>
					</div>
					
					<div class="upgrade-benefits">
						<h4>What you'll get:</h4>
						<ul>
							<li>Up to {formatLimit(selectedPackage.limits.max_campaigns)} campaigns</li>
							<li>Up to {formatLimit(selectedPackage.limits.max_ads_per_campaign)} ads per campaign</li>
							<li>{formatLimit(selectedPackage.limits.max_monthly_impressions)} monthly impressions</li>
							<li>{selectedPackage.limits.max_storage_gb} GB storage</li>
							<li>{selectedPackage.limits.support_level.toUpperCase()} support level</li>
						</ul>
					</div>
					
					<div class="billing-notice">
						<p class="text-sm text-secondary">
							Your billing cycle will be adjusted and you'll be charged the prorated amount for the remaining period.
						</p>
					</div>
				</div>
			</div>
			
			<div class="modal-footer">
				<button class="btn btn-outline" on:click={() => showUpgradeModal = false} disabled={upgrading}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={handleUpgrade} disabled={upgrading}>
					{#if upgrading}
						<div class="loading-spinner small"></div>
						Upgrading...
					{:else}
						Confirm Upgrade
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.packages-page {
		padding: var(--space-xl);
		background: var(--bg-secondary);
		min-height: 100vh;
	}
	
	.page-header {
		margin-bottom: var(--space-2xl);
	}
	
	.header-content {
		text-align: center;
	}
	
	.header-text h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}
	
	.header-text p {
		color: var(--text-secondary);
		font-size: var(--text-lg);
		margin: 0;
	}
	
	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
	}
	
	.loading-spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--border-color);
		border-top: 3px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}
	
	.loading-spinner.small {
		width: 16px;
		height: 16px;
		border-width: 2px;
	}
	
	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
	
	.error-message {
		color: var(--error);
		font-size: var(--text-lg);
	}
	
	/* Current Subscription */
	.current-subscription {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-2xl);
	}
	
	.subscription-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl);
	}
	
	.subscription-info h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}
	
	.subscription-status {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.status-active {
		background: var(--success-bg);
		color: var(--success-text);
	}
	
	.subscription-price .price {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--primary);
	}
	
	.usage-stats h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}
	
	.stat-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}
	
	.stat-item label {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
	}
	
	.usage-bar {
		width: 100%;
		height: 8px;
		background: var(--bg-secondary);
		border-radius: var(--radius-full);
		overflow: hidden;
	}
	
	.usage-fill {
		height: 100%;
		transition: width var(--transition-fast);
		border-radius: var(--radius-full);
	}
	
	.usage-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
	}
	
	/* Packages */
	.packages-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
		text-align: center;
	}
	
	.packages-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
		gap: var(--space-xl);
		max-width: 1200px;
		margin: 0 auto;
	}
	
	.package-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		position: relative;
		transition: all var(--transition-fast);
		display: flex;
		flex-direction: column;
	}
	
	.package-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}
	
	.package-card.featured {
		border-color: var(--primary);
		box-shadow: 0 0 0 1px var(--primary);
	}
	
	.package-card.current {
		border-color: var(--success);
		background: var(--success-bg);
	}
	
	.featured-badge,
	.current-badge {
		position: absolute;
		top: -12px;
		left: 50%;
		transform: translateX(-50%);
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.featured-badge {
		background: var(--primary);
		color: var(--white);
	}
	
	.current-badge {
		background: var(--success);
		color: var(--white);
	}
	
	.package-header {
		text-align: center;
		margin-bottom: var(--space-xl);
	}
	
	.package-header h3 {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}
	
	.package-description {
		color: var(--text-secondary);
		margin: 0 0 var(--space-lg) 0;
		line-height: 1.5;
	}
	
	.package-price .price {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--primary);
	}
	
	.package-features,
	.package-limits {
		margin-bottom: var(--space-xl);
	}
	
	.package-features h4,
	.package-limits h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.package-features ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}
	
	.package-features li {
		display: flex;
		align-items: flex-start;
		gap: var(--space-sm);
		margin-bottom: var(--space-md);
	}
	
	.feature-icon {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
		margin-top: 2px;
	}
	
	.feature-icon svg {
		width: 100%;
		height: 100%;
	}
	
	.package-features li.included .feature-icon {
		color: var(--success);
	}
	
	.package-features li.not-included .feature-icon {
		color: var(--text-secondary);
	}
	
	.package-features li.not-included {
		opacity: 0.6;
	}
	
	.feature-text {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}
	
	.feature-name {
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.feature-description {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}
	
	.limits-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--space-md);
	}
	
	.limit-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}
	
	.limit-item label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}
	
	.limit-item span {
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.support-level {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.support-basic {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}
	
	.support-priority {
		background: var(--warning-bg);
		color: var(--warning-text);
	}
	
	.support-premium {
		background: var(--primary-bg);
		color: var(--primary-text);
	}
	
	.package-actions {
		margin-top: auto;
		display: flex;
		justify-content: center;
	}
	
	/* Modal */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
	}
	
	.modal-content {
		background: var(--bg-primary);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		max-width: 500px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}
	
	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-xl);
		border-bottom: 1px solid var(--border-color);
	}
	
	.modal-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}
	
	.modal-close {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-md);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: var(--text-secondary);
	}
	
	.modal-close:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}
	
	.modal-close svg {
		width: 18px;
		height: 18px;
	}
	
	.modal-body {
		padding: var(--space-xl);
	}
	
	.upgrade-summary p {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-primary);
	}
	
	.pricing-info {
		margin-bottom: var(--space-lg);
	}
	
	.price-comparison {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}
	
	.price-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-md);
	}
	
	.price-item.new-price {
		background: var(--primary-bg);
		border: 1px solid var(--primary);
	}
	
	.price-item label {
		font-weight: 600;
		color: var(--text-secondary);
	}
	
	.price-item span {
		font-weight: 700;
		color: var(--text-primary);
	}
	
	.upgrade-benefits h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-md) 0;
	}
	
	.upgrade-benefits ul {
		list-style: none;
		padding: 0;
		margin: 0 0 var(--space-lg) 0;
	}
	
	.upgrade-benefits li {
		padding: var(--space-sm) 0;
		color: var(--text-primary);
		position: relative;
		padding-left: var(--space-lg);
	}
	
	.upgrade-benefits li::before {
		content: '✓';
		position: absolute;
		left: 0;
		color: var(--success);
		font-weight: 600;
	}
	
	.billing-notice {
		padding: var(--space-md);
		background: var(--info-bg);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--info);
	}
	
	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-xl);
		border-top: 1px solid var(--border-color);
	}
	
	@media (max-width: 768px) {
		.packages-page {
			padding: var(--space-lg);
		}
		
		.packages-grid {
			grid-template-columns: 1fr;
		}
		
		.subscription-header {
			flex-direction: column;
			gap: var(--space-md);
		}
		
		.stats-grid {
			grid-template-columns: 1fr;
		}
		
		.limits-grid {
			grid-template-columns: 1fr;
		}
		
		.modal-content {
			width: 95%;
		}
		
		.price-comparison {
			gap: var(--space-sm);
		}
	}
</style> 