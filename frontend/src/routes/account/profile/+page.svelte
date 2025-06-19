<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { api } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let user: any = null;
	let loading = true;
	let saving = false;
	let error = '';
	let isAuthenticated = false;

	// Form data
	let formData = {
		name: '',
		email: '',
		bio: '',
		location: '',
		website: '',
		phone: ''
	};

	let originalData = { ...formData };
	let hasChanges = false;

	// Subscribe to auth store
	auth.subscribe(state => {
		isAuthenticated = state.isAuthenticated;
		user = state.user;
		if (user) {
			formData = {
				name: user.name || '',
				email: user.email || '',
				bio: user.bio || '',
				location: user.location || '',
				website: user.website || '',
				phone: user.phone || ''
			};
			originalData = { ...formData };
		}
	});

	onMount(async () => {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		await loadProfile();
	});

	const loadProfile = async () => {
		try {
			loading = true;
			const response = await api.get('/api/v1/users/profile');
			
			if (response.user) {
				formData = {
					name: response.user.name || '',
					email: response.user.email || '',
					bio: response.user.bio || '',
					location: response.user.location || '',
					website: response.user.website || '',
					phone: response.user.phone || ''
				};
				originalData = { ...formData };
			}
		} catch (err) {
			error = 'Failed to load profile';
			console.error('Error loading profile:', err);
		} finally {
			loading = false;
		}
	};

	const handleInputChange = () => {
		hasChanges = JSON.stringify(formData) !== JSON.stringify(originalData);
	};

	const handleSave = async () => {
		if (!hasChanges) return;

		try {
			saving = true;
			const response = await api.put('/api/v1/users/profile', formData);
			
			if (response.success) {
				originalData = { ...formData };
				hasChanges = false;
				showToast('Profile updated successfully', 'success');
				
				// Update auth store with new user data
				auth.updateUser(response.user);
			} else {
				showToast('Failed to update profile', 'error');
			}
		} catch (err) {
			showToast('Failed to update profile', 'error');
			console.error('Error updating profile:', err);
		} finally {
			saving = false;
		}
	};

	const handleCancel = () => {
		formData = { ...originalData };
		hasChanges = false;
	};

	const handleBack = () => {
		if (hasChanges) {
			if (confirm('You have unsaved changes. Are you sure you want to leave?')) {
				goto('/account');
			}
		} else {
			goto('/account');
		}
	};

	// Validate email format
	const isValidEmail = (email: string) => {
		const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
		return emailRegex.test(email);
	};

	// Validate website URL
	const isValidWebsite = (website: string) => {
		if (!website) return true; // Optional field
		try {
			new URL(website);
			return true;
		} catch {
			return false;
		}
	};

	$: isFormValid = formData.name.trim() && 
					 formData.email.trim() && 
					 isValidEmail(formData.email) && 
					 isValidWebsite(formData.website);
</script>

<svelte:head>
	<title>Edit Profile - BOME</title>
	<meta name="description" content="Edit your BOME profile information" />
</svelte:head>

<div class="profile-page">
	<div class="container">
		<header class="page-header">
			<button class="back-button" on:click={handleBack}>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M19 12H5"></path>
					<path d="M12 19l-7-7 7-7"></path>
				</svg>
				Back to Account
			</button>
			<h1>Edit Profile</h1>
			<p>Update your personal information and preferences</p>
		</header>

		{#if loading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading your profile...</p>
			</div>
		{:else if error}
			<div class="error-container">
				<p class="error-message">{error}</p>
				<button class="btn btn-primary" on:click={loadProfile}>
					Try Again
				</button>
			</div>
		{:else}
			<div class="profile-content">
				<form class="profile-form" on:submit|preventDefault={handleSave}>
					<!-- Basic Information -->
					<div class="form-section">
						<h2>Basic Information</h2>
						
						<div class="form-row">
							<div class="form-group">
								<label for="name">Full Name *</label>
								<input
									id="name"
									type="text"
									bind:value={formData.name}
									on:input={handleInputChange}
									placeholder="Enter your full name"
									required
								/>
							</div>
							
							<div class="form-group">
								<label for="email">Email Address *</label>
								<input
									id="email"
									type="email"
									bind:value={formData.email}
									on:input={handleInputChange}
									placeholder="Enter your email address"
									required
								/>
								{#if formData.email && !isValidEmail(formData.email)}
									<span class="field-error">Please enter a valid email address</span>
								{/if}
							</div>
						</div>
					</div>

					<!-- Additional Information -->
					<div class="form-section">
						<h2>Additional Information</h2>
						
						<div class="form-group">
							<label for="bio">Bio</label>
							<textarea
								id="bio"
								bind:value={formData.bio}
								on:input={handleInputChange}
								placeholder="Tell us a little about yourself..."
								rows="4"
							></textarea>
							<span class="field-hint">Share a brief description about yourself (optional)</span>
						</div>

						<div class="form-row">
							<div class="form-group">
								<label for="location">Location</label>
								<input
									id="location"
									type="text"
									bind:value={formData.location}
									on:input={handleInputChange}
									placeholder="City, State/Country"
								/>
							</div>
							
							<div class="form-group">
								<label for="phone">Phone Number</label>
								<input
									id="phone"
									type="tel"
									bind:value={formData.phone}
									on:input={handleInputChange}
									placeholder="+1 (555) 123-4567"
								/>
							</div>
						</div>

						<div class="form-group">
							<label for="website">Website</label>
							<input
								id="website"
								type="url"
								bind:value={formData.website}
								on:input={handleInputChange}
								placeholder="https://your-website.com"
							/>
							{#if formData.website && !isValidWebsite(formData.website)}
								<span class="field-error">Please enter a valid website URL</span>
							{/if}
						</div>
					</div>

					<!-- Profile Avatar Section -->
					<div class="form-section">
						<h2>Profile Picture</h2>
						<div class="avatar-section">
							<div class="current-avatar">
								<div class="avatar-placeholder">
									{formData.name.charAt(0)?.toUpperCase() || 'U'}
								</div>
							</div>
							<div class="avatar-info">
								<p>Profile pictures are generated from your initials</p>
								<p class="avatar-note">Custom avatar uploads coming soon!</p>
							</div>
						</div>
					</div>

					<!-- Form Actions -->
					<div class="form-actions">
						<button 
							type="button" 
							class="btn btn-outline" 
							on:click={handleCancel}
							disabled={!hasChanges}
						>
							Cancel Changes
						</button>
						<button 
							type="submit" 
							class="btn btn-primary"
							disabled={!hasChanges || !isFormValid || saving}
						>
							{#if saving}
								<LoadingSpinner size="small" />
								Saving...
							{:else}
								Save Changes
							{/if}
						</button>
					</div>
				</form>

				<!-- Account Security -->
				<div class="security-section">
					<h2>Account Security</h2>
					<div class="security-card">
						<div class="security-item">
							<div class="security-info">
								<h3>Password</h3>
								<p>Last changed: Never</p>
							</div>
							<button class="btn btn-outline" disabled>
								Change Password
							</button>
						</div>
						<div class="security-item">
							<div class="security-info">
								<h3>Two-Factor Authentication</h3>
								<p>Not enabled</p>
							</div>
							<button class="btn btn-outline" disabled>
								Enable 2FA
							</button>
						</div>
					</div>
					<p class="security-note">Security features coming soon!</p>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.profile-page {
		min-height: 100vh;
		padding: 2rem 0;
		background: var(--bg-gradient);
	}

	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
		position: relative;
	}

	.back-button {
		position: absolute;
		left: 0;
		top: 0;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.875rem;
		transition: color 0.3s ease;
	}

	.back-button:hover {
		color: var(--primary-color);
	}

	.back-button svg {
		width: 16px;
		height: 16px;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.loading-container,
	.error-container {
		text-align: center;
		padding: 3rem 0;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1rem;
	}

	.profile-content {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.profile-form,
	.security-section {
		background: var(--card-bg);
		border-radius: 20px;
		padding: 2rem;
		box-shadow: var(--neumorphic-shadow);
		border: 1px solid var(--border-color);
	}

	.form-section {
		margin-bottom: 2rem;
	}

	.form-section:last-child {
		margin-bottom: 0;
	}

	.form-section h2 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1.5rem;
		padding-bottom: 0.5rem;
		border-bottom: 1px solid var(--border-color);
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.form-group input,
	.form-group textarea {
		width: 100%;
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 12px;
		background: var(--bg-secondary);
		color: var(--text-primary);
		font-size: 1rem;
		transition: all 0.3s ease;
	}

	.form-group input:focus,
	.form-group textarea:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px rgba(var(--primary-rgb), 0.1);
	}

	.form-group textarea {
		resize: vertical;
		min-height: 100px;
	}

	.field-hint {
		display: block;
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-top: 0.25rem;
	}

	.field-error {
		display: block;
		font-size: 0.875rem;
		color: var(--error-text);
		margin-top: 0.25rem;
	}

	.avatar-section {
		display: flex;
		align-items: center;
		gap: 1.5rem;
	}

	.current-avatar {
		flex-shrink: 0;
	}

	.avatar-placeholder {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: var(--primary-color);
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
		font-weight: 700;
	}

	.avatar-info p {
		margin: 0 0 0.5rem 0;
		color: var(--text-secondary);
	}

	.avatar-note {
		font-size: 0.875rem;
		color: var(--text-tertiary) !important;
		font-style: italic;
	}

	.form-actions {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding-top: 2rem;
		border-top: 1px solid var(--border-color);
	}

	.security-section h2 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1.5rem;
	}

	.security-card {
		background: var(--bg-secondary);
		border-radius: 16px;
		padding: 1.5rem;
		margin-bottom: 1rem;
	}

	.security-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1rem 0;
		border-bottom: 1px solid var(--border-color);
	}

	.security-item:last-child {
		border-bottom: none;
	}

	.security-info h3 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 0.25rem 0;
	}

	.security-info p {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin: 0;
	}

	.security-note {
		font-size: 0.875rem;
		color: var(--text-tertiary);
		font-style: italic;
		text-align: center;
		margin: 0;
	}

	@media (max-width: 768px) {
		.back-button {
			position: static;
			justify-content: center;
			margin-bottom: 1rem;
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.avatar-section {
			flex-direction: column;
			text-align: center;
		}

		.form-actions {
			flex-direction: column;
		}

		.security-item {
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}
	}
</style> 