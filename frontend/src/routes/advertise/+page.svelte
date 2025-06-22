<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { advertiserStore } from '$lib/stores/advertiser';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdvertiserAccount, AdvertiserPackage } from '$lib/types/advertising';

	let user: any = null;
	let isAuthenticated = false;
	let loading = true;
	let currentStep = 1;
	let advertiserAccount: AdvertiserAccount | null = null;
	let selectedPackage: AdvertiserPackage | null = null;
	let applicationSubmitted = false;

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
		try {
			// Check if user already has an advertiser account using the store
			const existingAccount = await advertiserStore.getByUserId(user.id);
			if (existingAccount) {
				advertiserAccount = existingAccount;
				// If account exists, redirect to dashboard with appropriate tab
				goto('/dashboard?tab=advertiser');
				return;
			}
		} catch (error) {
			console.error('Error checking advertiser status:', error);
		}
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
			// Submit application using the store
			const applicationData = {
				user_id: user.id,
				...businessForm
			};
			
			advertiserAccount = await advertiserStore.submitApplication(applicationData);
			currentStep = 2;
		} catch (error) {
			console.error('Registration failed:', error);
		} finally {
			submitting = false;
		}
	}

	async function selectPackage(pkg: AdvertiserPackage) {
		selectedPackage = pkg;
		submitting = true;
		
		try {
			// Mock API call to submit complete application
			await new Promise(resolve => setTimeout(resolve, 1500));
			
			// Mark application as submitted
			applicationSubmitted = true;
			currentStep = 3;
		} catch (error) {
			console.error('Package selection failed:', error);
		} finally {
			submitting = false;
		}
	}

	function goToStep(step: number) {
		if (step <= currentStep || (step === 2 && advertiserAccount)) {
			currentStep = step;
		}
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
					<div class="step-label">Application Review</div>
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

						<div class="form-row">
							<div class="form-group">
								<label for="business_address">Business Address</label>
								<input
									type="text"
									id="business_address"
									bind:value={businessForm.business_address}
									placeholder="123 Main St, City, State 12345"
								/>
							</div>
							
							<div class="form-group">
								<label for="tax_id">Tax ID / EIN</label>
								<input
									type="text"
									id="tax_id"
									bind:value={businessForm.tax_id}
									placeholder="12-3456789"
								/>
							</div>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label for="website">Website</label>
								<input
									type="url"
									id="website"
									bind:value={businessForm.website}
									placeholder="https://www.yourcompany.com"
								/>
							</div>
							
							<div class="form-group">
								<label for="industry">Industry *</label>
								<select
									id="industry"
									bind:value={businessForm.industry}
									class:error={formErrors.industry}
									required
								>
									<option value="">Select Industry</option>
									<option value="books">Books & Publishing</option>
									<option value="education">Education</option>
									<option value="religion">Religious Organizations</option>
									<option value="nonprofit">Non-Profit</option>
									<option value="software">Software & Technology</option>
									<option value="retail">Retail & E-commerce</option>
									<option value="services">Professional Services</option>
									<option value="media">Media & Entertainment</option>
									<option value="other">Other</option>
								</select>
								{#if formErrors.industry}
									<span class="error-message">{formErrors.industry}</span>
								{/if}
							</div>
						</div>

						<div class="form-actions">
							<button type="submit" class="submit-btn" disabled={submitting}>
								{#if submitting}
									<LoadingSpinner size="small" color="white" />
									Submitting...
								{:else}
									Continue to Package Selection
								{/if}
							</button>
						</div>
					</form>
				</div>

			{:else if currentStep === 2}
				<!-- Package Selection Step -->
				<div class="packages-section">
					<h2>Choose Your Advertising Package</h2>
					<p>Select the package that best fits your advertising needs. You can upgrade or downgrade your package at any time after approval.</p>
					
					<div class="packages-grid">
						{#each packages as pkg}
							<div class="package-card glass" class:featured={pkg.is_featured}>
								{#if pkg.is_featured}
									<div class="featured-badge">Most Popular</div>
								{/if}
								
								<div class="package-header">
									<h3>{pkg.name}</h3>
									<div class="package-price">
										<span class="price">{formatPrice(pkg.price)}</span>
										<span class="period">/{pkg.billing_cycle}</span>
									</div>
									<p class="package-description">{pkg.description}</p>
								</div>
								
								<div class="package-features">
									<div class="feature">
										<span class="feature-label">Max Campaigns:</span>
										<span class="feature-value">{pkg.limits.max_campaigns}</span>
									</div>
									<div class="feature">
										<span class="feature-label">Ads per Campaign:</span>
										<span class="feature-value">{pkg.limits.max_ads_per_campaign}</span>
									</div>
									<div class="feature">
										<span class="feature-label">Monthly Impressions:</span>
										<span class="feature-value">{formatNumber(pkg.limits.max_monthly_impressions)}</span>
									</div>
									<div class="feature">
										<span class="feature-label">Storage:</span>
										<span class="feature-value">{pkg.limits.max_storage_gb}GB</span>
									</div>
									<div class="feature">
										<span class="feature-label">Support Level:</span>
										<span class="feature-value">{pkg.limits.support_level}</span>
									</div>
									<div class="feature">
										<span class="feature-label">Analytics Retention:</span>
										<span class="feature-value">{pkg.limits.analytics_retention_days} days</span>
									</div>
								</div>
								
								<div class="package-actions">
									<button 
										type="button" 
										class="select-package-btn"
										class:selected={selectedPackage?.id === pkg.id}
										disabled={submitting}
										on:click={() => selectPackage(pkg)}
									>
										{#if submitting && selectedPackage?.id === pkg.id}
											<LoadingSpinner size="small" color="white" />
											Submitting Application...
										{:else if selectedPackage?.id === pkg.id}
											Selected
										{:else}
											Select {pkg.name}
										{/if}
									</button>
								</div>
							</div>
						{/each}
					</div>
				</div>

			{:else if currentStep === 3}
				<!-- Application Submitted -->
				<div class="success-section glass">
					<div class="success-icon">
						<svg width="64" height="64" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
							<path d="M9 12L11 14L15 10M21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3C16.9706 3 21 7.02944 21 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</div>
					
					<h2>Application Submitted Successfully!</h2>
					<p>Thank you for your interest in advertising with Book of Mormon Evidence. Your application has been submitted for review.</p>
					
					<div class="application-summary">
						<h3>Application Summary</h3>
						<div class="summary-item">
							<span class="label">Company:</span>
							<span class="value">{advertiserAccount?.company_name}</span>
						</div>
						<div class="summary-item">
							<span class="label">Contact:</span>
							<span class="value">{advertiserAccount?.contact_name}</span>
						</div>
						<div class="summary-item">
							<span class="label">Email:</span>
							<span class="value">{advertiserAccount?.business_email}</span>
						</div>
						<div class="summary-item">
							<span class="label">Selected Package:</span>
							<span class="value">{selectedPackage?.name} - {formatPrice(selectedPackage?.price || 0)}/month</span>
						</div>
						<div class="summary-item">
							<span class="label">Status:</span>
							<span class="value status-pending">Pending Review</span>
						</div>
					</div>
					
					<div class="next-steps">
						<h3>What Happens Next?</h3>
						<div class="steps-list">
							<div class="next-step">
								<div class="step-icon">1</div>
								<div class="step-content">
									<h4>Application Review</h4>
									<p>Our team will review your application within 2-3 business days to ensure it meets our community guidelines.</p>
								</div>
							</div>
							<div class="next-step">
								<div class="step-icon">2</div>
								<div class="step-content">
									<h4>Approval Notification</h4>
									<p>You'll receive an email notification once your application is approved, along with access to your advertiser dashboard.</p>
								</div>
							</div>
							<div class="next-step">
								<div class="step-icon">3</div>
								<div class="step-content">
									<h4>Create Your First Campaign</h4>
									<p>Once approved, you can log into your advertiser dashboard to create and manage your advertising campaigns.</p>
								</div>
							</div>
						</div>
					</div>
					
					<div class="contact-info">
						<p><strong>Questions?</strong> Contact our advertising team at <a href="mailto:advertising@bome.org">advertising@bome.org</a></p>
					</div>
					
					<div class="success-actions">
						<button type="button" class="dashboard-btn" on:click={() => goto('/dashboard?tab=advertiser&from=advertise')}>
							Return to Dashboard
						</button>
					</div>
				</div>
			{/if}
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
		display: flex;
		align-items: baseline;
		justify-content: center;
		gap: var(--space-xs);
		margin-bottom: var(--space-md);
	}

	.price {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--primary);
	}

	.period {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 400;
	}

	.package-description {
		color: var(--text-secondary);
		text-align: center;
		margin-bottom: var(--space-lg);
		line-height: 1.5;
		font-size: var(--text-sm);
	}

	.package-features {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		margin-bottom: var(--space-xl);
	}

	.feature {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-sm) 0;
		border-bottom: 1px solid var(--border-color);
	}

	.feature-label {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.feature-value {
		color: var(--text-primary);
		font-weight: 600;
		font-size: var(--text-sm);
	}

	.package-actions {
		display: flex;
		justify-content: center;
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
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
	}

	.select-package-btn:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.select-package-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.select-package-btn.selected {
		background: var(--success);
	}

	/* Success Section Styles */
	.success-section {
		padding: var(--space-3xl);
		border-radius: var(--radius-xl);
		text-align: center;
		margin-bottom: var(--space-2xl);
	}

	.success-icon {
		margin: 0 auto var(--space-xl);
		width: 64px;
		height: 64px;
		color: var(--success);
	}

	.success-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.success-section > p {
		color: var(--text-secondary);
		margin-bottom: var(--space-2xl);
		line-height: 1.6;
	}

	.application-summary {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-xl);
		margin-bottom: var(--space-2xl);
		text-align: left;
	}

	.application-summary h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
		text-align: center;
	}

	.summary-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md) 0;
		border-bottom: 1px solid var(--border-color);
	}

	.summary-item:last-child {
		border-bottom: none;
	}

	.summary-item .label {
		color: var(--text-secondary);
		font-weight: 500;
	}

	.summary-item .value {
		color: var(--text-primary);
		font-weight: 600;
	}

	.status-pending {
		color: var(--warning) !important;
		background: rgba(var(--warning-rgb), 0.1);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
	}

	.next-steps {
		text-align: left;
		margin-bottom: var(--space-2xl);
	}

	.next-steps h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
		text-align: center;
	}

	.steps-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.next-step {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
	}

	.step-icon {
		width: 32px;
		height: 32px;
		border-radius: 50%;
		background: var(--primary);
		color: var(--white);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: var(--text-sm);
		flex-shrink: 0;
	}

	.step-content h4 {
		font-size: var(--text-md);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.step-content p {
		color: var(--text-secondary);
		line-height: 1.5;
		font-size: var(--text-sm);
	}

	.contact-info {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		margin-bottom: var(--space-2xl);
	}

	.contact-info p {
		margin: 0;
		color: var(--text-secondary);
		text-align: center;
	}

	.contact-info a {
		color: var(--primary);
		text-decoration: none;
		font-weight: 600;
	}

	.contact-info a:hover {
		text-decoration: underline;
	}

	.success-actions {
		display: flex;
		justify-content: center;
	}

	.dashboard-btn {
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		border: none;
		padding: var(--space-md) var(--space-xl);
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.dashboard-btn:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.hero-stats {
			grid-template-columns: 1fr;
			gap: var(--space-md);
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.packages-grid {
			grid-template-columns: 1fr;
		}

		.package-card.featured {
			transform: none;
		}

		.steps {
			flex-direction: column;
			gap: var(--space-md);
		}

		.step-line {
			display: none;
		}

		.success-section {
			padding: var(--space-2xl) var(--space-lg);
		}

		.next-step {
			flex-direction: column;
			text-align: center;
		}

		.step-content {
			text-align: center;
		}
	}
</style> 