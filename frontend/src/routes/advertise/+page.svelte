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
	let currentStep = 1;
	let advertiserAccount: AdvertiserAccount | null = null;
	let selectedPackage: AdvertiserPackage | null = null;

	// Mock packages data
	const packages: AdvertiserPackage[] = [
		{
			id: 1,
			name: 'Starter',
			description: 'Perfect for small businesses getting started with advertising',
			price: 299,
			billing_cycle: 'monthly',
			features: [],
			limits: {
				max_campaigns: 3,
				max_ads_per_campaign: 5,
				max_monthly_impressions: 50000,
				max_file_size_mb: 10,
				max_storage_gb: 1,
				allowed_ad_types: ['banner', 'large'],
				allowed_placements: ['articles-sidebar', 'videos-sidebar'],
				priority_boost: 1,
				analytics_retention_days: 30,
				support_level: 'basic'
			},
			is_active: true,
			is_featured: false,
			sort_order: 1,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		},
		{
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
		},
		{
			id: 3,
			name: 'Enterprise',
			description: 'Full-scale advertising solution for large organizations',
			price: 1299,
			billing_cycle: 'monthly',
			features: [],
			limits: {
				max_campaigns: 50,
				max_ads_per_campaign: 50,
				max_monthly_impressions: 1000000,
				max_file_size_mb: 100,
				max_storage_gb: 25,
				allowed_ad_types: ['banner', 'large', 'small', 'video'],
				allowed_placements: ['articles-header', 'articles-mid', 'articles-sidebar', 'articles-footer', 'videos-header', 'videos-mid', 'videos-sidebar', 'videos-footer', 'events-header', 'events-mid', 'events-footer'],
				priority_boost: 5,
				analytics_retention_days: 365,
				support_level: 'premium'
			},
			is_active: true,
			is_featured: false,
			sort_order: 3,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString()
		}
	];

	// Business registration form
	let businessForm = {
		company_name: '',
		business_email: '',
		contact_name: '',
		contact_phone: '',
		business_address: '',
		tax_id: '',
		website: '',
		industry: ''
	};

	let formErrors: Record<string, string> = {};
	let submitting = false;

	onMount(async () => {
		auth.subscribe((state) => {
			user = state.user;
			isAuthenticated = state.isAuthenticated;
		});

		if (!isAuthenticated) {
			goto('/login?redirect=/advertise');
			return;
		}

		// Check if user already has an advertiser account
		await checkAdvertiserStatus();
		loading = false;
	});

	async function checkAdvertiserStatus() {
		// Mock check - in production, this would be an API call
		// For now, assume user doesn't have an advertiser account
		advertiserAccount = null;
	}

	function validateBusinessForm() {
		formErrors = {};
		
		if (!businessForm.company_name.trim()) {
			formErrors.company_name = 'Company name is required';
		}
		
		if (!businessForm.business_email.trim()) {
			formErrors.business_email = 'Business email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(businessForm.business_email)) {
			formErrors.business_email = 'Please enter a valid email address';
		}
		
		if (!businessForm.contact_name.trim()) {
			formErrors.contact_name = 'Contact name is required';
		}
		
		if (!businessForm.contact_phone.trim()) {
			formErrors.contact_phone = 'Contact phone is required';
		}
		
		if (!businessForm.industry.trim()) {
			formErrors.industry = 'Industry is required';
		}

		return Object.keys(formErrors).length === 0;
	}

	async function submitBusinessRegistration() {
		if (!validateBusinessForm()) return;
		
		submitting = true;
		try {
			// Mock API call - in production, this would register the business
			await new Promise(resolve => setTimeout(resolve, 2000));
			
			// Mock success response
			advertiserAccount = {
				id: Math.floor(Math.random() * 1000),
				user_id: user.id,
				...businessForm,
				status: 'pending',
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};
			
			currentStep = 2;
		} catch (error) {
			console.error('Registration failed:', error);
		} finally {
			submitting = false;
		}
	}

	function selectPackage(pkg: AdvertiserPackage) {
		selectedPackage = pkg;
		currentStep = 3;
	}

	function goToStep(step: number) {
		if (step <= currentStep || (step === 2 && advertiserAccount)) {
			currentStep = step;
		}
	}

	function handleCampaignCreated(event: CustomEvent) {
		const campaignData = event.detail;
		console.log('Campaign created successfully:', campaignData);
		
		// Show success message and redirect to next step or completion
		alert('Campaign created successfully! Your campaign has been submitted for approval. You will receive an email notification once it has been reviewed.');
		
		// In a real implementation, you might redirect to a success page or dashboard
		// goto('/advertise/success');
	}

	function handleGoBack() {
		currentStep = 2;
	}

	function formatPrice(price: number) {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(price);
	}

	function formatNumber(num: number) {
		return new Intl.NumberFormat('en-US').format(num);
	}
</script>

<svelte:head>
	<title>Advertise with Book of Mormon Evidence - BOME</title>
	<meta name="description" content="Partner with BOME to reach our engaged audience of Book of Mormon enthusiasts and researchers. Choose from flexible advertising packages designed for businesses of all sizes." />
</svelte:head>

<Navigation />

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading advertiser portal...</p>
	</div>
{:else}
	<div class="advertise-page">
		<!-- Hero Section -->
		<div class="hero-section">
			<div class="hero-content">
				<h1>Advertise with Book of Mormon Evidence</h1>
				<p>Reach thousands of engaged readers, researchers, and enthusiasts in the Book of Mormon community. Our platform offers targeted advertising opportunities with detailed analytics and flexible packages.</p>
				
				<div class="hero-stats">
					<div class="stat">
						<div class="stat-number">50K+</div>
						<div class="stat-label">Monthly Visitors</div>
					</div>
					<div class="stat">
						<div class="stat-number">25K+</div>
						<div class="stat-label">Engaged Users</div>
					</div>
					<div class="stat">
						<div class="stat-number">95%</div>
						<div class="stat-label">Audience Retention</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Progress Steps -->
		<div class="steps-container">
			<div class="steps">
				<div class="step" class:active={currentStep >= 1} class:completed={currentStep > 1}>
					<div class="step-number">1</div>
					<div class="step-label">Business Registration</div>
				</div>
				<div class="step-line" class:completed={currentStep > 1}></div>
				<div class="step" class:active={currentStep >= 2} class:completed={currentStep > 2}>
					<div class="step-number">2</div>
					<div class="step-label">Choose Package</div>
				</div>
				<div class="step-line" class:completed={currentStep > 2}></div>
				<div class="step" class:active={currentStep >= 3}>
					<div class="step-number">3</div>
					<div class="step-label">Create Campaign</div>
				</div>
			</div>
		</div>

		<!-- Step Content -->
		<div class="step-content">
			{#if currentStep === 1}
				<!-- Business Registration Step -->
				<div class="registration-section glass">
					<h2>Register Your Business</h2>
					<p>Tell us about your business to get started with advertising on BOME. All applications are reviewed to ensure quality and relevance to our community.</p>
					
					<form on:submit|preventDefault={submitBusinessRegistration} class="business-form">
						<div class="form-row">
							<div class="form-group">
								<label for="company_name">Company Name *</label>
								<input
									type="text"
									id="company_name"
									bind:value={businessForm.company_name}
									class:error={formErrors.company_name}
									placeholder="Your Company Name"
									required
								/>
								{#if formErrors.company_name}
									<span class="error-message">{formErrors.company_name}</span>
								{/if}
							</div>
							
							<div class="form-group">
								<label for="business_email">Business Email *</label>
								<input
									type="email"
									id="business_email"
									bind:value={businessForm.business_email}
									class:error={formErrors.business_email}
									placeholder="business@company.com"
									required
								/>
								{#if formErrors.business_email}
									<span class="error-message">{formErrors.business_email}</span>
								{/if}
							</div>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label for="contact_name">Contact Name *</label>
								<input
									type="text"
									id="contact_name"
									bind:value={businessForm.contact_name}
									class:error={formErrors.contact_name}
									placeholder="Primary Contact Person"
									required
								/>
								{#if formErrors.contact_name}
									<span class="error-message">{formErrors.contact_name}</span>
								{/if}
							</div>
							
							<div class="form-group">
								<label for="contact_phone">Contact Phone *</label>
								<input
									type="tel"
									id="contact_phone"
									bind:value={businessForm.contact_phone}
									class:error={formErrors.contact_phone}
									placeholder="+1 (555) 123-4567"
									required
								/>
								{#if formErrors.contact_phone}
									<span class="error-message">{formErrors.contact_phone}</span>
								{/if}
							</div>
						</div>

						<div class="form-group">
							<label for="business_address">Business Address</label>
							<textarea
								id="business_address"
								bind:value={businessForm.business_address}
								placeholder="Your business address"
								rows="3"
							></textarea>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label for="website">Website</label>
								<input
									type="url"
									id="website"
									bind:value={businessForm.website}
									placeholder="https://yourwebsite.com"
								/>
							</div>
							
							<div class="form-group">
								<label for="tax_id">Tax ID (Optional)</label>
								<input
									type="text"
									id="tax_id"
									bind:value={businessForm.tax_id}
									placeholder="12-3456789"
								/>
							</div>
						</div>

						<div class="form-group">
							<label for="industry">Industry *</label>
							<select id="industry" bind:value={businessForm.industry} class:error={formErrors.industry} required>
								<option value="">Select your industry</option>
								<option value="education">Education</option>
								<option value="religious">Religious Organizations</option>
								<option value="publishing">Publishing & Media</option>
								<option value="technology">Technology</option>
								<option value="retail">Retail & E-commerce</option>
								<option value="consulting">Consulting & Services</option>
								<option value="nonprofit">Non-profit</option>
								<option value="other">Other</option>
							</select>
							{#if formErrors.industry}
								<span class="error-message">{formErrors.industry}</span>
							{/if}
						</div>

						<div class="form-actions">
							<button type="submit" class="submit-btn" disabled={submitting}>
								{#if submitting}
									<LoadingSpinner size="small" color="white" />
									Submitting...
								{:else}
									Register Business
								{/if}
							</button>
						</div>
					</form>
				</div>

			{:else if currentStep === 2}
				<!-- Package Selection Step -->
				<div class="packages-section">
					<h2>Choose Your Advertising Package</h2>
					<p>Select the package that best fits your advertising needs and budget. You can upgrade or downgrade at any time.</p>
					
					<div class="packages-grid">
						{#each packages as pkg}
							<div class="package-card glass" class:featured={pkg.is_featured}>
								{#if pkg.is_featured}
									<div class="featured-badge">Most Popular</div>
								{/if}
								
								<div class="package-header">
									<h3>{pkg.name}</h3>
									<div class="package-price">
										{formatPrice(pkg.price)}
										<span class="billing-cycle">/{pkg.billing_cycle}</span>
									</div>
								</div>
								
								<p class="package-description">{pkg.description}</p>
								
								<div class="package-features">
									<div class="feature">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										Up to {pkg.limits.max_campaigns} campaigns
									</div>
									<div class="feature">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										{pkg.limits.max_ads_per_campaign} ads per campaign
									</div>
									<div class="feature">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										{formatNumber(pkg.limits.max_monthly_impressions)} monthly impressions
									</div>
									<div class="feature">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										{pkg.limits.max_storage_gb}GB storage
									</div>
									<div class="feature">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										{pkg.limits.support_level} support
									</div>
									<div class="feature">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="20,6 9,17 4,12"></polyline>
										</svg>
										{pkg.limits.analytics_retention_days} days analytics retention
									</div>
								</div>
								
								<button class="select-package-btn" on:click={() => selectPackage(pkg)}>
									Select {pkg.name}
								</button>
							</div>
						{/each}
					</div>
				</div>

			{:else if currentStep === 3}
				<!-- Campaign Creation Step -->
				<CampaignCreator 
					{selectedPackage} 
					{advertiserAccount}
					on:campaignCreated={handleCampaignCreated}
					on:goBack={handleGoBack}
				/>
			{/if}
		</div>

		<!-- Ad Specifications Section -->
		<div class="specifications-section glass">
			<h2>Advertisement Specifications</h2>
			<p>Ensure your advertisements meet our technical requirements for optimal display across all placements.</p>
			
			<div class="specs-grid">
				<div class="spec-card">
					<h3>Banner Ads (728x90px)</h3>
					<ul>
						<li>Recommended for header and footer placements</li>
						<li>Maximum file size: 2MB</li>
						<li>Formats: JPG, PNG, GIF, WebP</li>
						<li>Base rate: $80-100/week</li>
					</ul>
				</div>
				
				<div class="spec-card">
					<h3>Large Rectangle (300x250px)</h3>
					<ul>
						<li>Premium sidebar and content placements</li>
						<li>Maximum file size: 2MB</li>
						<li>Formats: JPG, PNG, GIF, WebP</li>
						<li>Base rate: $150-200/week</li>
					</ul>
				</div>
				
				<div class="spec-card">
					<h3>Small Rectangle (300x125px)</h3>
					<ul>
						<li>Compact sidebar placement option</li>
						<li>Maximum file size: 1MB</li>
						<li>Formats: JPG, PNG, WebP</li>
						<li>Base rate: $75/week</li>
					</ul>
				</div>
			</div>
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

	.advertise-page {
		min-height: 100vh;
		padding: var(--space-xl) var(--space-lg);
		max-width: 1200px;
		margin: 0 auto;
	}

	.hero-section {
		text-align: center;
		padding: var(--space-3xl) 0;
		margin-bottom: var(--space-3xl);
	}

	.hero-content h1 {
		font-size: var(--text-4xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
		background-clip: text;
	}

	.hero-content p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
		max-width: 600px;
		margin: 0 auto var(--space-2xl);
		line-height: 1.6;
	}

	.hero-stats {
		display: flex;
		justify-content: center;
		gap: var(--space-3xl);
		margin-top: var(--space-2xl);
	}

	.stat {
		text-align: center;
	}

	.stat-number {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.steps-container {
		margin-bottom: var(--space-3xl);
	}

	.steps {
		display: flex;
		align-items: center;
		justify-content: center;
		max-width: 600px;
		margin: 0 auto;
	}

	.step {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-sm);
	}

	.step-number {
		width: 40px;
		height: 40px;
		border-radius: 50%;
		background: var(--bg-glass);
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		transition: all var(--transition-normal);
	}

	.step.active .step-number {
		background: var(--primary);
		color: var(--white);
	}

	.step.completed .step-number {
		background: var(--success);
		color: var(--white);
	}

	.step-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 500;
		white-space: nowrap;
	}

	.step.active .step-label {
		color: var(--text-primary);
	}

	.step-line {
		flex: 1;
		height: 2px;
		background: var(--bg-glass);
		margin: 0 var(--space-lg);
		transition: background-color var(--transition-normal);
	}

	.step-line.completed {
		background: var(--success);
	}

	.step-content {
		margin-bottom: var(--space-3xl);
	}

	.registration-section {
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
		margin-bottom: var(--space-2xl);
	}

	.registration-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.registration-section p {
		color: var(--text-secondary);
		margin-bottom: var(--space-2xl);
		line-height: 1.6;
	}

	.business-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-lg);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.form-group label {
		font-weight: 500;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.form-group input,
	.form-group textarea,
	.form-group select {
		padding: var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: all var(--transition-normal);
	}

	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.1);
	}

	.form-group input.error,
	.form-group select.error {
		border-color: var(--error);
	}

	.error-message {
		color: var(--error);
		font-size: var(--text-xs);
		margin-top: var(--space-xs);
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: var(--space-lg);
	}

	.submit-btn {
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		border: none;
		padding: var(--space-md) var(--space-xl);
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.submit-btn:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.submit-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.packages-section {
		margin-bottom: var(--space-2xl);
	}

	.packages-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		text-align: center;
		margin-bottom: var(--space-md);
	}

	.packages-section p {
		color: var(--text-secondary);
		text-align: center;
		margin-bottom: var(--space-2xl);
		line-height: 1.6;
	}

	.packages-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-xl);
		margin-bottom: var(--space-2xl);
	}

	.package-card {
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
		position: relative;
		border: 2px solid transparent;
		transition: all var(--transition-normal);
	}

	.package-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}

	.package-card.featured {
		border-color: var(--primary);
		transform: scale(1.05);
	}

	.featured-badge {
		position: absolute;
		top: -12px;
		left: 50%;
		transform: translateX(-50%);
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.package-header {
		text-align: center;
		margin-bottom: var(--space-lg);
	}

	.package-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.package-price {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--primary);
	}

	.billing-cycle {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 400;
	}

	.package-description {
		color: var(--text-secondary);
		text-align: center;
		margin-bottom: var(--space-lg);
		line-height: 1.5;
	}

	.package-features {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		margin-bottom: var(--space-xl);
	}

	.feature {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.feature svg {
		width: 16px;
		height: 16px;
		color: var(--success);
		flex-shrink: 0;
	}

	.select-package-btn {
		width: 100%;
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		border: none;
		padding: var(--space-md);
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.select-package-btn:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.campaign-info {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-2xl);
	}

	.selected-package-summary h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.package-limits {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.limit-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm) 0;
		border-bottom: 1px solid var(--border-color);
	}

	.limit-label {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.limit-value {
		color: var(--text-primary);
		font-weight: 600;
		font-size: var(--text-sm);
	}

	.next-steps h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.next-steps ol {
		padding-left: var(--space-lg);
		margin-bottom: var(--space-xl);
	}

	.next-steps li {
		color: var(--text-secondary);
		margin-bottom: var(--space-sm);
		line-height: 1.5;
	}

	.action-buttons {
		display: flex;
		gap: var(--space-md);
	}

	.back-btn {
		background: var(--bg-glass);
		color: var(--text-primary);
		border: 1px solid var(--border-color);
		padding: var(--space-md) var(--space-lg);
		border-radius: var(--radius-md);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.back-btn:hover {
		background: var(--bg-glass-dark);
	}

	.proceed-btn {
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		border: none;
		padding: var(--space-md) var(--space-lg);
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		flex: 1;
	}

	.proceed-btn:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.specifications-section {
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
	}

	.specifications-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		text-align: center;
		margin-bottom: var(--space-md);
	}

	.specifications-section p {
		color: var(--text-secondary);
		text-align: center;
		margin-bottom: var(--space-2xl);
		line-height: 1.6;
	}

	.specs-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: var(--space-xl);
	}

	.spec-card {
		background: var(--bg-glass);
		padding: var(--space-xl);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border-color);
	}

	.spec-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.spec-card ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.spec-card li {
		color: var(--text-secondary);
		margin-bottom: var(--space-sm);
		padding-left: var(--space-lg);
		position: relative;
		line-height: 1.5;
	}

	.spec-card li::before {
		content: '•';
		color: var(--primary);
		position: absolute;
		left: 0;
		font-weight: 600;
	}

	@media (max-width: 1024px) {
		.campaign-info {
			grid-template-columns: 1fr;
		}
		
		.hero-stats {
			gap: var(--space-xl);
		}
	}

	@media (max-width: 768px) {
		.advertise-page {
			padding: var(--space-lg) var(--space-md);
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.hero-stats {
			flex-direction: column;
			gap: var(--space-lg);
		}

		.steps {
			flex-direction: column;
			gap: var(--space-lg);
		}

		.step-line {
			width: 2px;
			height: 40px;
			margin: 0;
		}

		.action-buttons {
			flex-direction: column;
		}

		.packages-grid {
			grid-template-columns: 1fr;
		}

		.package-card.featured {
			transform: none;
		}

		.hero-content h1 {
			font-size: var(--text-3xl);
		}
	}
</style> 