<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdvertiserAccount } from '$lib/types/advertising';

	let advertiserAccount: AdvertiserAccount | null = null;
	let loading = true;
	let saving = false;
	let error: string | null = null;
	let successMessage: string | null = null;
	let activeTab = 'profile';

	// Form data
	let formData = {
		company_name: 'Mock Business Company',
		business_email: 'business@bome.test',
		contact_name: 'Business Owner',
		contact_phone: '+1 (555) 123-4567',
		business_address: '123 Business St, Business City, BC 12345',
		website: 'https://mockbusiness.com',
		industry: 'Education & Research'
	};

	let billingData = {
		billing_address: '',
		billing_city: '',
		billing_state: '',
		billing_zip: '',
		billing_country: 'United States',
		payment_method: 'credit_card',
		auto_billing: true
	};

	let notificationSettings = {
		email_notifications: true,
		campaign_alerts: true,
		billing_alerts: true,
		performance_reports: false,
		weekly_summary: true
	};

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		if ($auth.user?.role !== 'advertiser' && $auth.user?.role !== 'admin') {
			goto('/advertise');
			return;
		}

		await loadAccountData();
	});

	async function loadAccountData() {
		try {
			loading = true;
			
			// For mock users, create mock account data
			if ($auth.token?.startsWith('mock-advertiser-token-')) {
				await new Promise(resolve => setTimeout(resolve, 800));
				
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
					tax_id: '12-3456789',
					status: 'approved',
					verification_notes: 'Mock account for testing',
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};

				// Populate form data
				formData = {
					company_name: advertiserAccount.company_name,
					business_email: advertiserAccount.business_email,
					contact_name: advertiserAccount.contact_name,
					contact_phone: advertiserAccount.contact_phone || '',
					business_address: advertiserAccount.business_address || '',
					website: advertiserAccount.website || '',
					industry: advertiserAccount.industry || '',
					tax_id: advertiserAccount.tax_id || ''
				};
			} else {
				// Real API call
				const response = await fetch('/api/v1/advertiser/account', {
					headers: {
						'Authorization': `Bearer ${$auth.token}`
					}
				});

				if (response.ok) {
					const data = await response.json();
					advertiserAccount = data.data;
					// Populate form data from API response
					if (advertiserAccount) {
						formData = {
							company_name: advertiserAccount.company_name,
							business_email: advertiserAccount.business_email,
							contact_name: advertiserAccount.contact_name,
							contact_phone: advertiserAccount.contact_phone || '',
							business_address: advertiserAccount.business_address || '',
							website: advertiserAccount.website || '',
							industry: advertiserAccount.industry || '',
							tax_id: advertiserAccount.tax_id || ''
						};
					}
				}
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load account data';
		} finally {
			loading = false;
		}
	}

	async function saveProfile() {
		saving = true;
		error = null;
		successMessage = null;

		try {
			// Simulate API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			successMessage = 'Profile updated successfully!';
		} catch (err) {
			error = 'Failed to save profile';
		} finally {
			saving = false;
		}
	}

	async function saveBilling() {
		try {
			saving = true;
			error = null;
			successMessage = null;

			// For mock users, simulate save
			if ($auth.token?.startsWith('mock-advertiser-token-')) {
				await new Promise(resolve => setTimeout(resolve, 1000));
				successMessage = 'Billing information updated successfully!';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save billing information';
		} finally {
			saving = false;
		}
	}

	async function saveNotifications() {
		try {
			saving = true;
			error = null;
			successMessage = null;

			// For mock users, simulate save
			if ($auth.token?.startsWith('mock-advertiser-token-')) {
				await new Promise(resolve => setTimeout(resolve, 800));
				successMessage = 'Notification settings updated successfully!';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to save notification settings';
		} finally {
			saving = false;
		}
	}

	function clearMessages() {
		error = null;
		successMessage = null;
	}

	function getStatusClass(status: string): string {
		switch (status) {
			case 'approved': return 'status-success';
			case 'pending': return 'status-pending';
			case 'rejected': return 'status-error';
			default: return 'status-secondary';
		}
	}
</script>

<svelte:head>
	<title>Account Settings - Advertiser Dashboard - BOME</title>
</svelte:head>

<Navigation />

<div class="page-container">
	<div class="content-wrapper">
		<!-- Page Header -->
		<div class="page-header">
			<div class="header-content">
				<div class="header-text">
					<h1>Account Settings</h1>
					<p>Manage your advertiser account information and preferences</p>
				</div>
				{#if advertiserAccount}
					<div class="account-status">
						<span class="status-badge {getStatusClass(advertiserAccount.status)}">
							{advertiserAccount.status}
						</span>
					</div>
				{/if}
			</div>
		</div>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
			</div>
		{:else if !advertiserAccount}
			<div class="error-card">
				<h3>Account not found</h3>
				<p>Please complete your advertiser setup first.</p>
				<button class="btn btn-primary" on:click={() => goto('/advertiser/setup')}>
					Setup Account
				</button>
			</div>
		{:else}
			<!-- Tab Navigation -->
			<div class="tab-navigation">
				<button 
					class="tab-button {activeTab === 'profile' ? 'active' : ''}"
					on:click={() => { activeTab = 'profile'; clearMessages(); }}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
						<circle cx="12" cy="7" r="4"/>
					</svg>
					Profile
				</button>
				<button 
					class="tab-button {activeTab === 'billing' ? 'active' : ''}"
					on:click={() => { activeTab = 'billing'; clearMessages(); }}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="1" y="4" width="22" height="16" rx="2" ry="2"/>
						<line x1="1" y1="10" x2="23" y2="10"/>
					</svg>
					Billing
				</button>
				<button 
					class="tab-button {activeTab === 'notifications' ? 'active' : ''}"
					on:click={() => { activeTab = 'notifications'; clearMessages(); }}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/>
						<path d="M13.73 21a2 2 0 0 1-3.46 0"/>
					</svg>
					Notifications
				</button>
			</div>

			<!-- Messages -->
			{#if error}
				<div class="message error">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"/>
						<path d="M15 9l-6 6"/>
						<path d="M9 9l6 6"/>
					</svg>
					{error}
				</div>
			{/if}

			{#if successMessage}
				<div class="message success">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 12l2 2 4-4"/>
						<path d="M21 12c-1.274 4.057-5.064 7-9 7s-7.726-2.943-9-7c1.274-4.057 5.064-7 9-7s7.726 2.943 9 7z"/>
					</svg>
					{successMessage}
				</div>
			{/if}

			<!-- Tab Content -->
			<div class="tab-content">
				{#if activeTab === 'profile'}
					<div class="form-section">
						<div class="section-header">
							<h2>Business Profile</h2>
							<p>Update your business information and contact details</p>
						</div>

						<form on:submit|preventDefault={saveProfile} class="form-grid">
							<div class="form-group">
								<label for="company_name">Company Name *</label>
								<input
									type="text"
									id="company_name"
									bind:value={formData.company_name}
									required
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="business_email">Business Email *</label>
								<input
									type="email"
									id="business_email"
									bind:value={formData.business_email}
									required
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="contact_name">Contact Name *</label>
								<input
									type="text"
									id="contact_name"
									bind:value={formData.contact_name}
									required
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="contact_phone">Contact Phone</label>
								<input
									type="tel"
									id="contact_phone"
									bind:value={formData.contact_phone}
									class="form-input"
								/>
							</div>

							<div class="form-group full-width">
								<label for="business_address">Business Address</label>
								<textarea
									id="business_address"
									bind:value={formData.business_address}
									rows="3"
									class="form-input"
								></textarea>
							</div>

							<div class="form-group">
								<label for="website">Website</label>
								<input
									type="url"
									id="website"
									bind:value={formData.website}
									class="form-input"
									placeholder="https://"
								/>
							</div>

							<div class="form-group">
								<label for="industry">Industry</label>
								<select id="industry" bind:value={formData.industry} class="form-input">
									<option value="">Select Industry</option>
									<option value="Education & Research">Education & Research</option>
									<option value="Religious Organizations">Religious Organizations</option>
									<option value="Publishing & Media">Publishing & Media</option>
									<option value="Technology">Technology</option>
									<option value="Healthcare">Healthcare</option>
									<option value="Finance">Finance</option>
									<option value="Other">Other</option>
								</select>
							</div>

							<div class="form-group">
								<label for="tax_id">Tax ID / EIN</label>
								<input
									type="text"
									id="tax_id"
									bind:value={formData.tax_id}
									class="form-input"
									placeholder="12-3456789"
								/>
							</div>

							<div class="form-actions full-width">
								<button type="submit" class="btn btn-primary" disabled={saving}>
									{#if saving}
										<LoadingSpinner size="small" />
										Saving...
									{:else}
										Save Profile
									{/if}
								</button>
							</div>
						</form>
					</div>
				{:else if activeTab === 'billing'}
					<div class="form-section">
						<div class="section-header">
							<h2>Billing Information</h2>
							<p>Manage your billing address and payment preferences</p>
						</div>

						<form on:submit|preventDefault={saveBilling} class="form-grid">
							<div class="form-group full-width">
								<label for="billing_address">Billing Address</label>
								<input
									type="text"
									id="billing_address"
									bind:value={billingData.billing_address}
									class="form-input"
									placeholder="123 Main Street"
								/>
							</div>

							<div class="form-group">
								<label for="billing_city">City</label>
								<input
									type="text"
									id="billing_city"
									bind:value={billingData.billing_city}
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="billing_state">State/Province</label>
								<input
									type="text"
									id="billing_state"
									bind:value={billingData.billing_state}
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="billing_zip">ZIP/Postal Code</label>
								<input
									type="text"
									id="billing_zip"
									bind:value={billingData.billing_zip}
									class="form-input"
								/>
							</div>

							<div class="form-group">
								<label for="billing_country">Country</label>
								<select id="billing_country" bind:value={billingData.billing_country} class="form-input">
									<option value="United States">United States</option>
									<option value="Canada">Canada</option>
									<option value="United Kingdom">United Kingdom</option>
									<option value="Australia">Australia</option>
								</select>
							</div>

							<div class="form-group">
								<label for="payment_method">Payment Method</label>
								<select id="payment_method" bind:value={billingData.payment_method} class="form-input">
									<option value="credit_card">Credit Card</option>
									<option value="bank_transfer">Bank Transfer</option>
									<option value="paypal">PayPal</option>
								</select>
							</div>

							<div class="form-group full-width">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={billingData.auto_billing}
										class="checkbox-input"
									/>
									<span class="checkbox-custom"></span>
									Enable automatic billing
								</label>
							</div>

							<div class="form-actions full-width">
								<button type="submit" class="btn btn-primary" disabled={saving}>
									{#if saving}
										<LoadingSpinner size="small" />
										Saving...
									{:else}
										Save Billing Info
									{/if}
								</button>
							</div>
						</form>
					</div>
				{:else if activeTab === 'notifications'}
					<div class="form-section">
						<div class="section-header">
							<h2>Notification Preferences</h2>
							<p>Choose how you want to receive updates about your campaigns</p>
						</div>

						<form on:submit|preventDefault={saveNotifications} class="notification-form">
							<div class="notification-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={notificationSettings.email_notifications}
										class="checkbox-input"
									/>
									<span class="checkbox-custom"></span>
									<div class="notification-info">
										<div class="notification-title">Email Notifications</div>
										<div class="notification-desc">Receive general email notifications</div>
									</div>
								</label>
							</div>

							<div class="notification-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={notificationSettings.campaign_alerts}
										class="checkbox-input"
									/>
									<span class="checkbox-custom"></span>
									<div class="notification-info">
										<div class="notification-title">Campaign Alerts</div>
										<div class="notification-desc">Get notified about campaign status changes</div>
									</div>
								</label>
							</div>

							<div class="notification-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={notificationSettings.billing_alerts}
										class="checkbox-input"
									/>
									<span class="checkbox-custom"></span>
									<div class="notification-info">
										<div class="notification-title">Billing Alerts</div>
										<div class="notification-desc">Receive notifications about billing and payments</div>
									</div>
								</label>
							</div>

							<div class="notification-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={notificationSettings.performance_reports}
										class="checkbox-input"
									/>
									<span class="checkbox-custom"></span>
									<div class="notification-info">
										<div class="notification-title">Performance Reports</div>
										<div class="notification-desc">Receive detailed performance reports</div>
									</div>
								</label>
							</div>

							<div class="notification-group">
								<label class="checkbox-label">
									<input
										type="checkbox"
										bind:checked={notificationSettings.weekly_summary}
										class="checkbox-input"
									/>
									<span class="checkbox-custom"></span>
									<div class="notification-info">
										<div class="notification-title">Weekly Summary</div>
										<div class="notification-desc">Get a weekly summary of your campaign performance</div>
									</div>
								</label>
							</div>

							<div class="form-actions">
								<button type="submit" class="btn btn-primary" disabled={saving}>
									{#if saving}
										<LoadingSpinner size="small" />
										Saving...
									{:else}
										Save Preferences
									{/if}
								</button>
							</div>
						</form>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<Footer />

<style>
	.page-container {
		min-height: 100vh;
		background: var(--bg-primary);
		padding-top: 80px;
	}

	.content-wrapper {
		max-width: 1200px;
		margin: 0 auto;
		padding: var(--space-xl) var(--space-lg);
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.page-header {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-lg);
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
	}

	.account-status {
		flex-shrink: 0;
	}

	.status-badge {
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-lg);
		font-size: var(--text-sm);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-success {
		background: rgba(34, 197, 94, 0.2);
		color: #22c55e;
		border: 1px solid rgba(34, 197, 94, 0.3);
	}

	.status-pending {
		background: rgba(245, 158, 11, 0.2);
		color: #f59e0b;
		border: 1px solid rgba(245, 158, 11, 0.3);
	}

	.status-error {
		background: rgba(239, 68, 68, 0.2);
		color: #ef4444;
		border: 1px solid rgba(239, 68, 68, 0.3);
	}

	.tab-navigation {
		display: flex;
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-sm);
		gap: var(--space-sm);
	}

	.tab-button {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-lg);
		background: none;
		border: none;
		border-radius: var(--radius-lg);
		color: var(--text-secondary);
		font-weight: 500;
		transition: all var(--transition-normal);
		cursor: pointer;
	}

	.tab-button svg {
		width: 18px;
		height: 18px;
	}

	.tab-button:hover {
		background: rgba(255, 255, 255, 0.05);
		color: var(--text-primary);
	}

	.tab-button.active {
		background: var(--primary-gradient);
		color: var(--white);
	}

	.message {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md) var(--space-lg);
		border-radius: var(--radius-lg);
		font-weight: 500;
	}

	.message svg {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
	}

	.message.error {
		background: rgba(239, 68, 68, 0.1);
		color: #ef4444;
		border: 1px solid rgba(239, 68, 68, 0.2);
	}

	.message.success {
		background: rgba(34, 197, 94, 0.1);
		color: #22c55e;
		border: 1px solid rgba(34, 197, 94, 0.2);
	}

	.tab-content {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
	}

	.section-header {
		margin-bottom: var(--space-xl);
	}

	.section-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.section-header p {
		color: var(--text-secondary);
	}

	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.form-group.full-width {
		grid-column: 1 / -1;
	}

	.form-group label {
		color: var(--text-primary);
		font-weight: 500;
		font-size: var(--text-sm);
	}

	.form-input {
		padding: var(--space-md);
		background: var(--bg-secondary);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: all var(--transition-normal);
	}

	.form-input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		margin-top: var(--space-lg);
	}

	.notification-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.notification-group {
		padding: var(--space-lg);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.checkbox-label {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		cursor: pointer;
	}

	.checkbox-input {
		display: none;
	}

	.checkbox-custom {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-radius: var(--radius-sm);
		background: transparent;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all var(--transition-normal);
		flex-shrink: 0;
		margin-top: 2px;
	}

	.checkbox-input:checked + .checkbox-custom {
		background: var(--primary);
		border-color: var(--primary);
	}

	.checkbox-input:checked + .checkbox-custom::after {
		content: '✓';
		color: var(--white);
		font-size: 12px;
		font-weight: bold;
	}

	.notification-info {
		flex: 1;
	}

	.notification-title {
		color: var(--text-primary);
		font-weight: 500;
		margin-bottom: var(--space-xs);
	}

	.notification-desc {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.loading-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 400px;
	}

	.error-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(239, 68, 68, 0.2);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		text-align: center;
	}

	@media (max-width: 768px) {
		.content-wrapper {
			padding: var(--space-lg) var(--space-md);
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.tab-navigation {
			flex-direction: column;
		}

		.form-grid {
			grid-template-columns: 1fr;
		}

		.form-actions {
			justify-content: stretch;
		}
	}
</style> 