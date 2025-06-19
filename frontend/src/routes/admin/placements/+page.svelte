<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdPlacement, PlacementPerformance, FormErrors } from '$lib/types/advertising';
	
	let placements: AdPlacement[] = [];
	let performance: PlacementPerformance[] = [];
	let loading = true;
	let error: string | null = null;
	let showCreateModal = false;
	let editingPlacement: AdPlacement | null = null;
	
	let formData = {
		name: '',
		description: '',
		location: 'header',
		ad_type: 'banner',
		max_width: 728,
		max_height: 90,
		base_rate: 100.00,
		is_active: true
	};
	
	let errors: FormErrors = {};
	let submitting = false;

	const locationOptions = [
		{ value: 'header', label: 'Header', description: 'Top of the page header' },
		{ value: 'sidebar', label: 'Sidebar', description: 'Right sidebar area' },
		{ value: 'footer', label: 'Footer', description: 'Bottom of the page footer' },
		{ value: 'content', label: 'Content', description: 'Within content areas' },
		{ value: 'video_overlay', label: 'Video Overlay', description: 'Overlay on video player' },
		{ value: 'between_videos', label: 'Between Videos', description: 'Between video listings' }
	];

	const adTypeOptions = [
		{ value: 'banner', label: 'Banner (728x90)', width: 728, height: 90 },
		{ value: 'large', label: 'Large Rectangle (300x250)', width: 300, height: 250 },
		{ value: 'small', label: 'Small Rectangle (160x120)', width: 160, height: 120 }
	];

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		const user = $auth.user;
		if (!user || user.role !== 'admin') {
			goto('/');
			return;
		}

		try {
			await Promise.all([loadPlacements(), loadPerformance()]);
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	});

	async function loadPlacements() {
		const response = await fetch('/api/v1/admin/ads/placements', {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (response.ok) {
			const data = await response.json();
			placements = data.data || [];
		} else {
			// Mock data for demonstration
			placements = [
				{
					id: 1,
					name: 'Header Banner',
					description: 'Banner ad displayed in the site header',
					location: 'header',
					ad_type: 'banner',
					max_width: 728,
					max_height: 90,
					base_rate: 100.00,
					is_active: true,
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				},
				{
					id: 2,
					name: 'Sidebar Large',
					description: 'Large ad displayed in the sidebar',
					location: 'sidebar',
					ad_type: 'large',
					max_width: 300,
					max_height: 250,
					base_rate: 150.00,
					is_active: true,
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				},
				{
					id: 3,
					name: 'Video Overlay',
					description: 'Small overlay ad during video playback',
					location: 'video_overlay',
					ad_type: 'small',
					max_width: 200,
					max_height: 100,
					base_rate: 250.00,
					is_active: false,
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				}
			];
		}
	}

	async function loadPerformance() {
		const response = await fetch('/api/v1/admin/ads/placements/performance', {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (response.ok) {
			const data = await response.json();
			performance = data.data || [];
		} else {
			// Mock performance data
			performance = [
				{
					placement_id: 1,
					placement_name: 'Header Banner',
					total_ads: 5,
					active_ads: 3,
					total_impressions: 15420,
					total_clicks: 342,
					total_revenue: 856.30,
					average_ctr: 2.22,
					fill_rate: 85.5
				},
				{
					placement_id: 2,
					placement_name: 'Sidebar Large',
					total_ads: 3,
					active_ads: 2,
					total_impressions: 8930,
					total_clicks: 178,
					total_revenue: 445.60,
					average_ctr: 1.99,
					fill_rate: 92.3
				},
				{
					placement_id: 3,
					placement_name: 'Video Overlay',
					total_ads: 0,
					active_ads: 0,
					total_impressions: 0,
					total_clicks: 0,
					total_revenue: 0,
					average_ctr: 0,
					fill_rate: 0
				}
			];
		}
	}

	function handleCreate() {
		resetForm();
		editingPlacement = null;
		showCreateModal = true;
	}

	function handleEdit(placement: AdPlacement) {
		formData = {
			name: placement.name,
			description: placement.description,
			location: placement.location,
			ad_type: placement.ad_type,
			max_width: placement.max_width,
			max_height: placement.max_height,
			base_rate: placement.base_rate,
			is_active: placement.is_active
		};
		editingPlacement = placement;
		showCreateModal = true;
	}

	function resetForm() {
		formData = {
			name: '',
			description: '',
			location: 'header',
			ad_type: 'banner',
			max_width: 728,
			max_height: 90,
			base_rate: 100.00,
			is_active: true
		};
		errors = {};
	}

	function handleAdTypeChange() {
		const adType = adTypeOptions.find(option => option.value === formData.ad_type);
		if (adType) {
			formData.max_width = adType.width;
			formData.max_height = adType.height;
		}
	}

	function validateForm(): boolean {
		errors = {};

		if (!formData.name.trim()) {
			errors.name = 'Placement name is required';
		}

		if (!formData.description.trim()) {
			errors.description = 'Description is required';
		}

		if (formData.max_width < 1 || formData.max_width > 2000) {
			errors.max_width = 'Width must be between 1 and 2000 pixels';
		}

		if (formData.max_height < 1 || formData.max_height > 2000) {
			errors.max_height = 'Height must be between 1 and 2000 pixels';
		}

		if (formData.base_rate < 0) {
			errors.base_rate = 'Base rate must be a positive number';
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		submitting = true;
		errors = {};

		try {
			const url = editingPlacement 
				? `/api/v1/admin/ads/placements/${editingPlacement.id}`
				: '/api/v1/admin/ads/placements';
			
			const method = editingPlacement ? 'PUT' : 'POST';

			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify(formData)
			});

			if (response.ok) {
				await loadPlacements();
				showCreateModal = false;
				resetForm();
			} else {
				const data = await response.json();
				errors.general = data.error || 'Failed to save placement';
			}
		} catch (err) {
			errors.general = 'Network error occurred';
		} finally {
			submitting = false;
		}
	}

	async function togglePlacementStatus(placement: AdPlacement) {
		try {
			const response = await fetch(`/api/v1/admin/ads/placements/${placement.id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify({
					...placement,
					is_active: !placement.is_active
				})
			});

			if (response.ok) {
				await loadPlacements();
			} else {
				error = 'Failed to update placement status';
			}
		} catch (err) {
			error = 'Network error occurred';
		}
	}

	function getPerformanceForPlacement(placementId: number): PlacementPerformance | null {
		return performance.find(p => p.placement_id === placementId) || null;
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
</script>

<svelte:head>
	<title>Placement Management - BOME Admin</title>
</svelte:head>

<div class="placement-management">
	<div class="page-header">
		<div class="header-content">
			<div>
				<h1>Placement Management</h1>
				<p>Manage advertising placements and monitor performance</p>
			</div>
			<button 
				class="btn btn-primary"
				on:click={handleCreate}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 6v6m0 0v6m0-6h6m-6 0H6" />
				</svg>
				Create Placement
			</button>
		</div>
	</div>

	{#if error}
		<div class="error-message">
			<div class="alert alert-error">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
				</svg>
				<span>{error}</span>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading placements...</p>
		</div>
	{:else}
		<!-- Placements Table -->
		<div class="placements-section">
			<div class="section-header">
				<h2>Ad Placements</h2>
				<p>Manage where advertisements can be displayed on your site</p>
			</div>

			<div class="placements-table">
				<div class="table-header">
					<div class="header-cell">Placement</div>
					<div class="header-cell">Location & Type</div>
					<div class="header-cell">Dimensions</div>
					<div class="header-cell">Base Rate</div>
					<div class="header-cell">Performance</div>
					<div class="header-cell">Status</div>
					<div class="header-cell">Actions</div>
				</div>

				{#each placements as placement}
					{@const perf = getPerformanceForPlacement(placement.id)}
					<div class="table-row">
						<div class="cell placement-info">
							<h3>{placement.name}</h3>
							<p>{placement.description}</p>
						</div>
						
						<div class="cell location-info">
							<div class="location-badge {placement.location}">
								{locationOptions.find(l => l.value === placement.location)?.label}
							</div>
							<span class="ad-type">{placement.ad_type}</span>
						</div>
						
						<div class="cell dimensions">
							<span>{placement.max_width} × {placement.max_height}px</span>
						</div>
						
						<div class="cell base-rate">
							<span>{formatCurrency(placement.base_rate)}/week</span>
						</div>
						
						<div class="cell performance">
							{#if perf}
								<div class="perf-metrics">
									<div class="metric">
										<span class="value">{formatNumber(perf.total_impressions)}</span>
										<span class="label">Impressions</span>
									</div>
									<div class="metric">
										<span class="value">{formatPercentage(perf.average_ctr)}</span>
										<span class="label">CTR</span>
									</div>
									<div class="metric">
										<span class="value">{formatCurrency(perf.total_revenue)}</span>
										<span class="label">Revenue</span>
									</div>
								</div>
							{:else}
								<span class="no-data">No data</span>
							{/if}
						</div>
						
						<div class="cell status">
							<button 
								class="status-toggle {placement.is_active ? 'active' : 'inactive'}"
								on:click={() => togglePlacementStatus(placement)}
							>
								{placement.is_active ? 'Active' : 'Inactive'}
							</button>
						</div>
						
						<div class="cell actions">
							<button 
								class="btn btn-sm btn-secondary"
								on:click={() => handleEdit(placement)}
							>
								Edit
							</button>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<!-- Create/Edit Modal -->
{#if showCreateModal}
	<div class="modal-overlay" on:click={() => showCreateModal = false}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h2>{editingPlacement ? 'Edit' : 'Create'} Placement</h2>
				<button class="close-btn" on:click={() => showCreateModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M18 6L6 18M6 6l12 12" />
					</svg>
				</button>
			</div>

			<div class="modal-content">
				{#if errors.general}
					<div class="alert alert-error">
						<span>{errors.general}</span>
					</div>
				{/if}

				<form on:submit|preventDefault={handleSubmit} class="placement-form">
					<div class="form-group">
						<label for="name">Placement Name *</label>
						<input
							type="text"
							id="name"
							bind:value={formData.name}
							class:error={errors.name}
							placeholder="Enter placement name"
							required
						/>
						{#if errors.name}
							<span class="error-text">{errors.name}</span>
						{/if}
					</div>

					<div class="form-group">
						<label for="description">Description *</label>
						<textarea
							id="description"
							bind:value={formData.description}
							class:error={errors.description}
							placeholder="Describe where this placement appears"
							rows="3"
							required
						></textarea>
						{#if errors.description}
							<span class="error-text">{errors.description}</span>
						{/if}
					</div>

					<div class="form-row">
						<div class="form-group">
							<label for="location">Location *</label>
							<select
								id="location"
								bind:value={formData.location}
								required
							>
								{#each locationOptions as option}
									<option value={option.value}>{option.label}</option>
								{/each}
							</select>
						</div>

						<div class="form-group">
							<label for="ad_type">Ad Type *</label>
							<select
								id="ad_type"
								bind:value={formData.ad_type}
								on:change={handleAdTypeChange}
								required
							>
								{#each adTypeOptions as option}
									<option value={option.value}>{option.label}</option>
								{/each}
							</select>
						</div>
					</div>

					<div class="form-row">
						<div class="form-group">
							<label for="max_width">Max Width (px) *</label>
							<input
								type="number"
								id="max_width"
								bind:value={formData.max_width}
								class:error={errors.max_width}
								min="1"
								max="2000"
								required
							/>
							{#if errors.max_width}
								<span class="error-text">{errors.max_width}</span>
							{/if}
						</div>

						<div class="form-group">
							<label for="max_height">Max Height (px) *</label>
							<input
								type="number"
								id="max_height"
								bind:value={formData.max_height}
								class:error={errors.max_height}
								min="1"
								max="2000"
								required
							/>
							{#if errors.max_height}
								<span class="error-text">{errors.max_height}</span>
							{/if}
						</div>
					</div>

					<div class="form-group">
						<label for="base_rate">Base Rate (USD/week) *</label>
						<input
							type="number"
							id="base_rate"
							bind:value={formData.base_rate}
							class:error={errors.base_rate}
							min="0"
							step="0.01"
							required
						/>
						{#if errors.base_rate}
							<span class="error-text">{errors.base_rate}</span>
						{/if}
					</div>

					<div class="form-group">
						<label class="checkbox-label">
							<input
								type="checkbox"
								bind:checked={formData.is_active}
							/>
							<span class="checkmark"></span>
							Active placement
						</label>
					</div>

					<div class="form-actions">
						<button 
							type="button" 
							class="btn btn-secondary"
							on:click={() => showCreateModal = false}
						>
							Cancel
						</button>
						<button 
							type="submit" 
							class="btn btn-primary"
							disabled={submitting}
						>
							{#if submitting}
								<span class="loading-spinner small"></span>
								{editingPlacement ? 'Updating...' : 'Creating...'}
							{:else}
								{editingPlacement ? 'Update' : 'Create'} Placement
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	</div>
{/if}

<style>
	.placement-management {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-content h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.header-content p {
		color: var(--text-secondary);
		margin: var(--space-sm) 0 0 0;
	}

	.error-message {
		margin-bottom: var(--space-lg);
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-4xl);
		gap: var(--space-lg);
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
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

	.placements-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.section-header {
		margin-bottom: var(--space-xl);
	}

	.section-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.section-header p {
		color: var(--text-secondary);
		margin: 0;
	}

	.placements-table {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1.5fr 1fr 1fr 1.5fr 1fr 1fr;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		font-weight: 600;
		font-size: var(--text-sm);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 1.5fr 1fr 1fr 1.5fr 1fr 1fr;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
		transition: all var(--transition-normal);
		align-items: center;
	}

	.table-row:hover {
		background: var(--bg-glass-dark);
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.cell {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.placement-info h3 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.placement-info p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.4;
	}

	.location-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.location-badge.header {
		background: rgba(59, 130, 246, 0.1);
		color: rgb(59, 130, 246);
	}

	.location-badge.sidebar {
		background: rgba(16, 185, 129, 0.1);
		color: rgb(16, 185, 129);
	}

	.location-badge.footer {
		background: rgba(245, 158, 11, 0.1);
		color: rgb(245, 158, 11);
	}

	.location-badge.content {
		background: rgba(139, 92, 246, 0.1);
		color: rgb(139, 92, 246);
	}

	.location-badge.video_overlay {
		background: rgba(239, 68, 68, 0.1);
		color: rgb(239, 68, 68);
	}

	.location-badge.between_videos {
		background: rgba(236, 72, 153, 0.1);
		color: rgb(236, 72, 153);
	}

	.ad-type {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-transform: capitalize;
	}

	.dimensions {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		color: var(--text-primary);
	}

	.base-rate {
		font-weight: 600;
		color: var(--text-primary);
	}

	.perf-metrics {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.metric {
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.metric .value {
		font-weight: 600;
		font-size: var(--text-sm);
		color: var(--text-primary);
	}

	.metric .label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.no-data {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-style: italic;
	}

	.status-toggle {
		padding: var(--space-xs) var(--space-sm);
		border: none;
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.status-toggle.active {
		background: rgba(16, 185, 129, 0.1);
		color: rgb(16, 185, 129);
	}

	.status-toggle.inactive {
		background: rgba(107, 114, 128, 0.1);
		color: rgb(107, 114, 128);
	}

	.status-toggle:hover {
		transform: scale(1.05);
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: var(--z-modal);
		backdrop-filter: blur(4px);
	}

	.modal {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		max-width: 600px;
		width: 90vw;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: var(--shadow-2xl);
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-xl);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.modal-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.close-btn {
		width: 32px;
		height: 32px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-normal);
		color: var(--text-secondary);
	}

	.close-btn:hover {
		background: var(--bg-glass-dark);
		color: var(--text-primary);
		transform: scale(1.05);
	}

	.close-btn svg {
		width: 18px;
		height: 18px;
	}

	.modal-content {
		padding: var(--space-xl);
	}

	.placement-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.form-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-lg);
	}

	.form-group label {
		font-weight: 600;
		color: var(--text-primary);
		font-size: var(--text-sm);
	}

	.form-group input,
	.form-group textarea,
	.form-group select {
		padding: var(--space-md);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-base);
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
	.form-group textarea.error {
		border-color: var(--error);
		box-shadow: 0 0 0 3px rgba(var(--error-rgb), 0.1);
	}

	.error-text {
		color: var(--error);
		font-size: var(--text-sm);
		font-weight: 500;
	}

	.checkbox-label {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		cursor: pointer;
		font-weight: 500;
	}

	.checkbox-label input[type="checkbox"] {
		display: none;
	}

	.checkmark {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(255, 255, 255, 0.2);
		border-radius: var(--radius-sm);
		background: var(--bg-glass);
		transition: all var(--transition-normal);
		position: relative;
	}

	.checkbox-label input[type="checkbox"]:checked + .checkmark {
		background: var(--primary);
		border-color: var(--primary);
	}

	.checkbox-label input[type="checkbox"]:checked + .checkmark::after {
		content: '';
		position: absolute;
		left: 6px;
		top: 2px;
		width: 6px;
		height: 10px;
		border: solid white;
		border-width: 0 2px 2px 0;
		transform: rotate(45deg);
	}

	.form-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
		padding-top: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.table-header,
		.table-row {
			grid-template-columns: 2fr 1fr 1fr 1fr 1fr 1fr;
		}

		.performance {
			display: none;
		}
	}

	@media (max-width: 768px) {
		.table-header {
			display: none;
		}

		.table-row {
			grid-template-columns: 1fr;
			gap: var(--space-lg);
		}

		.form-row {
			grid-template-columns: 1fr;
		}

		.modal {
			width: 95vw;
		}

		.form-actions {
			flex-direction: column;
		}
	}
</style> 