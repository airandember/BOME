<script lang="ts">
	import { onMount } from 'svelte';
	import { SupportSettingsService } from '$lib/services/support-settings-service';

	let loading = true;
	let saving = false;
	let error = '';
	let success = '';

	// Support settings
	let supportEmail = '';
	let supportPhone = '';
	let supportUrl = '';
	let supportHours = '';
	let supportMessage = '';

	// Track changes
	let hasChanges = false;

	onMount(async () => {
		await loadSettings();
	});

	async function loadSettings() {
		loading = true;
		error = '';

		try {
			const settings = await SupportSettingsService.getAdminSupportSettings();

			supportEmail = settings.email || '';
			supportPhone = settings.phone || '';
			supportUrl = settings.url || '';
			supportHours = settings.hours || '';
			supportMessage = settings.message || 'Please contact our support team for assistance.';
		} catch (err: any) {
			error = err.message || 'Failed to load support settings';
			console.error('Error loading support settings:', err);
		} finally {
			loading = false;
		}
	}

	async function saveSettings() {
		saving = true;
		error = '';
		success = '';

		try {
			// Validate at least one contact method
			if (!supportEmail && !supportPhone && !supportUrl) {
				error = 'Please provide at least one support contact method (email, phone, or URL)';
				saving = false;
				return;
			}

			// Update all support settings
			await SupportSettingsService.updateSupportSettings({
				support_email: supportEmail,
				support_phone: supportPhone,
				support_url: supportUrl,
				support_hours: supportHours,
				support_message: supportMessage
			});

			success = 'Support settings saved successfully!';
			hasChanges = false;

			// Clear success message after 3 seconds
			setTimeout(() => {
				success = '';
			}, 3000);
		} catch (err: any) {
			error = err.message || 'Failed to save support settings';
			console.error('Error saving support settings:', err);
		} finally {
			saving = false;
		}
	}

	function handleInput() {
		hasChanges = true;
	}

	function previewSupportContact() {
		const support = {
			email: supportEmail,
			phone: supportPhone,
			url: supportUrl,
			hours: supportHours,
			message: supportMessage
		};

		alert(
			`Support Contact Preview:\n\n` +
				(support.email ? `Email: ${support.email}\n` : '') +
				(support.phone ? `Phone: ${support.phone}\n` : '') +
				(support.url ? `URL: ${support.url}\n` : '') +
				(support.hours ? `Hours: ${support.hours}\n` : '') +
				`\nMessage: ${support.message}`
		);
	}
</script>

<div class="support-settings-container">
	<div class="header">
		<h1>Support Configuration</h1>
		<p>Configure your support contact information displayed to users</p>
	</div>

	{#if loading}
		<div class="loading">
			<div class="spinner"></div>
			<p>Loading support settings...</p>
		</div>
	{:else}
		<div class="settings-form">
			{#if error}
				<div class="alert alert-error">
					<strong>Error:</strong>
					{error}
				</div>
			{/if}

			{#if success}
				<div class="alert alert-success">
					<strong>Success:</strong>
					{success}
				</div>
			{/if}

			<div class="form-section">
				<h2>Contact Information</h2>
				<p class="section-description">
					Users will see this information when they need support (e.g., multiple subscriptions,
					payment issues)
				</p>

				<div class="form-group">
					<label for="support-email">
						Support Email <span class="required">*</span>
					</label>
					<input
						id="support-email"
						type="email"
						bind:value={supportEmail}
						on:input={handleInput}
						placeholder="support@example.com"
						class="form-control"
					/>
					<small class="form-text">Primary email address for user support inquiries</small>
				</div>

				<div class="form-group">
					<label for="support-phone">Support Phone</label>
					<input
						id="support-phone"
						type="tel"
						bind:value={supportPhone}
						on:input={handleInput}
						placeholder="+1 (555) 123-4567"
						class="form-control"
					/>
					<small class="form-text">Optional phone number for support</small>
				</div>

				<div class="form-group">
					<label for="support-url">Support URL</label>
					<input
						id="support-url"
						type="url"
						bind:value={supportUrl}
						on:input={handleInput}
						placeholder="https://support.example.com"
						class="form-control"
					/>
					<small class="form-text">Optional link to help center, ticket system, or knowledge base</small>
				</div>

				<div class="form-group">
					<label for="support-hours">Support Hours</label>
					<input
						id="support-hours"
						type="text"
						bind:value={supportHours}
						on:input={handleInput}
						placeholder="Monday-Friday 9am-5pm EST"
						class="form-control"
					/>
					<small class="form-text">When your support team is available</small>
				</div>
			</div>

			<div class="form-section">
				<h2>Support Message</h2>
				<p class="section-description">Default message displayed to users when support is needed</p>

				<div class="form-group">
					<label for="support-message">Message</label>
					<textarea
						id="support-message"
						bind:value={supportMessage}
						on:input={handleInput}
						placeholder="Please contact our support team for assistance."
						class="form-control"
						rows="3"
					></textarea>
					<small class="form-text">This message will be shown alongside your contact information</small>
				</div>
			</div>

			<div class="form-actions">
				<button class="btn btn-secondary" on:click={previewSupportContact} disabled={saving}>
					Preview
				</button>
				<button
					class="btn btn-primary"
					on:click={saveSettings}
					disabled={saving || !hasChanges}
				>
					{#if saving}
						<span class="spinner-small"></span>
						Saving...
					{:else}
						Save Settings
					{/if}
				</button>
			</div>

			<div class="info-box">
				<h3>💡 How This Works</h3>
				<p>
					When users encounter issues (e.g., multiple active subscriptions, payment failures), they'll
					see your support contact information with a clear message about how to get help.
				</p>
				<p>
					<strong>Example:</strong> "You have 2 active subscriptions. Please contact support@example.com
					to consolidate them."
				</p>
			</div>
		</div>
	{/if}
</div>

<style>
	.support-settings-container {
		max-width: 800px;
		margin: 0 auto;
		padding: 2rem;
	}

	.header {
		margin-bottom: 2rem;
	}

	.header h1 {
		margin: 0 0 0.5rem 0;
		color: #1a1a1a;
	}

	.header p {
		margin: 0;
		color: #666;
	}

	.loading {
		text-align: center;
		padding: 3rem;
	}

	.spinner {
		border: 4px solid #f3f3f3;
		border-top: 4px solid #3498db;
		border-radius: 50%;
		width: 40px;
		height: 40px;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem auto;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	.settings-form {
		background: white;
		border-radius: 8px;
		padding: 2rem;
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.alert {
		padding: 1rem;
		border-radius: 4px;
		margin-bottom: 1.5rem;
	}

	.alert-error {
		background: #fee;
		border: 1px solid #fcc;
		color: #c00;
	}

	.alert-success {
		background: #efe;
		border: 1px solid #cfc;
		color: #060;
	}

	.form-section {
		margin-bottom: 2rem;
	}

	.form-section h2 {
		margin: 0 0 0.5rem 0;
		font-size: 1.25rem;
		color: #1a1a1a;
	}

	.section-description {
		margin: 0 0 1.5rem 0;
		color: #666;
		font-size: 0.9rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 600;
		color: #333;
	}

	.required {
		color: #c00;
	}

	.form-control {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #ddd;
		border-radius: 4px;
		font-size: 1rem;
		font-family: inherit;
	}

	.form-control:focus {
		outline: none;
		border-color: #3498db;
		box-shadow: 0 0 0 3px rgba(52, 152, 219, 0.1);
	}

	textarea.form-control {
		resize: vertical;
		min-height: 80px;
	}

	.form-text {
		display: block;
		margin-top: 0.25rem;
		color: #666;
		font-size: 0.875rem;
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		margin-top: 2rem;
	}

	.btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 4px;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-primary {
		background: #3498db;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #2980b9;
	}

	.btn-secondary {
		background: #95a5a6;
		color: white;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #7f8c8d;
	}

	.spinner-small {
		border: 2px solid transparent;
		border-top: 2px solid white;
		border-radius: 50%;
		width: 16px;
		height: 16px;
		animation: spin 0.8s linear infinite;
	}

	.info-box {
		margin-top: 2rem;
		padding: 1.5rem;
		background: #f8f9fa;
		border-left: 4px solid #3498db;
		border-radius: 4px;
	}

	.info-box h3 {
		margin: 0 0 1rem 0;
		font-size: 1.1rem;
		color: #1a1a1a;
	}

	.info-box p {
		margin: 0 0 0.5rem 0;
		color: #555;
		line-height: 1.6;
	}

	.info-box p:last-child {
		margin-bottom: 0;
	}
</style>

