<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import FileUpload from '$lib/components/FileUpload.svelte';
	import type { EnhancedAdCampaign, AdvertiserSubscription, AdAsset } from '$lib/types/advertising';
	
	let campaigns: EnhancedAdCampaign[] = [];
	let subscription: AdvertiserSubscription | null = null;
	let loading = true;
	let error = '';
	let showCreateModal = false;
	let showAssetModal = false;
	let selectedCampaign: EnhancedAdCampaign | null = null;
	let creating = false;
	
	// Form data
	let newCampaign = {
		name: '',
		description: '',
		budget: 0,
		start_date: '',
		end_date: '',
		target_audience: '',
		ad_type: 'banner' as const
	};
	
	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}
		
		await Promise.all([
			loadCampaigns(),
			loadSubscription()
		]);
		loading = false;
	});
	
	async function loadCampaigns() {
		try {
			// Mock data - replace with actual API
			campaigns = [
				{
					id: 1,
					advertiser_id: 1,
					name: 'Summer Book Promotion',
					description: 'Promote archaeological findings and Book of Mormon evidence',
					budget: 500.00,
					spent: 234.56,
					start_date: '2024-06-01T00:00:00Z',
					end_date: '2024-08-31T23:59:59Z',
					status: 'active',
					target_audience: 'Religious scholars, history enthusiasts',
					spent_amount: 234.56,
					billing_type: 'monthly',
					billing_rate: 2.50,
					created_at: '2024-05-15T00:00:00Z',
					updated_at: '2024-06-15T00:00:00Z',
					assets: [
						{
							id: 1,
							campaign_id: 1,
							asset_type: 'image',
							file_name: 'banner_728x90.jpg',
							file_path: '/api/assets/banner_728x90.jpg',
							file_size: 245760,
							mime_type: 'image/jpeg',
							width: 728,
							height: 90,
							alt_text: 'Book of Mormon Evidence Banner',
							status: 'approved',
							created_at: '2024-05-15T00:00:00Z',
							updated_at: '2024-05-15T00:00:00Z'
						}
					],
					advertisements: [],
					package_info: {
						package_id: 1,
						package_name: 'Starter',
						remaining_campaigns: 1,
						remaining_ads: 7,
						remaining_storage_gb: 0.7
					},
					asset_requirements: [
						{
							ad_type: 'banner',
							required_assets: [
								{
									asset_type: 'image',
									min_width: 728,
									max_width: 728,
									min_height: 90,
									max_height: 90,
									max_file_size_mb: 5,
									allowed_formats: ['jpg', 'jpeg', 'png', 'webp'],
									is_required: true,
									description: 'Main banner image (728x90px)'
								}
							]
						}
					],
					asset_count: 1,
					has_required_assets: true
				}
			];
		} catch (err) {
			error = 'Failed to load campaigns';
			console.error('Error loading campaigns:', err);
		}
	}
	
	async function loadSubscription() {
		try {
			// Mock data - replace with actual API
			subscription = {
				id: 1,
				advertiser_id: 1,
				package_id: 1,
				stripe_subscription_id: 'sub_1234567890',
				status: 'active',
				current_period_start: '2024-06-01T00:00:00Z',
				current_period_end: '2024-07-01T00:00:00Z',
				cancel_at_period_end: false,
				usage_stats: {
					campaigns_used: 1,
					ads_used: 3,
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
	
	async function handleCreateCampaign() {
		if (!subscription) return;
		
		// Check limits
		if (subscription.usage_stats.campaigns_used >= 3) { // Starter package limit
			showToast('Campaign limit reached. Please upgrade your package.', 'error');
			return;
		}
		
		creating = true;
		
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1500));
			
			showToast('Campaign created successfully!', 'success');
			showCreateModal = false;
			resetForm();
			await loadCampaigns();
			
		} catch (err) {
			showToast('Failed to create campaign', 'error');
			console.error('Error creating campaign:', err);
		} finally {
			creating = false;
		}
	}
	
	function resetForm() {
		newCampaign = {
			name: '',
			description: '',
			budget: 0,
			start_date: '',
			end_date: '',
			target_audience: '',
			ad_type: 'banner'
		};
	}
	
	function handleManageAssets(campaign: EnhancedAdCampaign) {
		selectedCampaign = campaign;
		showAssetModal = true;
	}
	
	function handleAssetUpload(event: CustomEvent<{ assets: AdAsset[] }>) {
		const { assets } = event.detail;
		showToast(`Successfully uploaded ${assets.length} asset(s)`, 'success');
		
		// Update campaign assets
		if (selectedCampaign) {
			selectedCampaign.assets = [...selectedCampaign.assets, ...assets];
			selectedCampaign.asset_count = selectedCampaign.assets.length;
			campaigns = campaigns; // Trigger reactivity
		}
	}
	
	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}
	
	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
	
	function getStatusColor(status: string): string {
		const colors = {
			active: 'var(--success)',
			paused: 'var(--warning)',
			completed: 'var(--info)',
			draft: 'var(--text-secondary)'
		};
		return colors[status as keyof typeof colors] || 'var(--text-secondary)';
	}
	
	function canCreateCampaign(): boolean {
		return subscription ? subscription.usage_stats.campaigns_used < 3 : false;
	}
</script>

<svelte:head>
	<title>Campaigns - BOME Advertiser</title>
</svelte:head>

<div class="campaigns-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Advertising Campaigns</h1>
				<p>Manage your advertising campaigns and assets</p>
			</div>
			<div class="header-actions">
				<button 
					class="btn btn-primary" 
					on:click={() => showCreateModal = true}
					disabled={!canCreateCampaign()}
				>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="12" y1="5" x2="12" y2="19"></line>
						<line x1="5" y1="12" x2="19" y2="12"></line>
					</svg>
					Create Campaign
				</button>
			</div>
		</div>
	</div>
	
	{#if subscription}
		<div class="usage-overview glass">
			<h3>Package Usage</h3>
			<div class="usage-grid">
				<div class="usage-item">
					<label>Campaigns</label>
					<span class="usage-value">{subscription.usage_stats.campaigns_used} / 3</span>
					<div class="usage-bar">
						<div class="usage-fill" style="width: {(subscription.usage_stats.campaigns_used / 3) * 100}%"></div>
					</div>
				</div>
				<div class="usage-item">
					<label>Storage</label>
					<span class="usage-value">{subscription.usage_stats.storage_used_gb.toFixed(1)} / 1.0 GB</span>
					<div class="usage-bar">
						<div class="usage-fill" style="width: {(subscription.usage_stats.storage_used_gb / 1.0) * 100}%"></div>
					</div>
				</div>
			</div>
		</div>
	{/if}
	
	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading campaigns...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={() => location.reload()}>
				Try Again
			</button>
		</div>
	{:else if campaigns.length === 0}
		<div class="empty-state">
			<div class="empty-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M3 3h18v18H3zM9 9h6v6H9z"></path>
				</svg>
			</div>
			<h3>No campaigns yet</h3>
			<p>Create your first advertising campaign to get started</p>
			<button class="btn btn-primary" on:click={() => showCreateModal = true}>
				Create Campaign
			</button>
		</div>
	{:else}
		<div class="campaigns-grid">
			{#each campaigns as campaign}
				<div class="campaign-card glass">
					<div class="campaign-header">
						<div class="campaign-info">
							<h3>{campaign.name}</h3>
							<p class="campaign-description">{campaign.description}</p>
						</div>
						<div class="campaign-status">
							<span class="status-badge" style="background: {getStatusColor(campaign.status)}">
								{campaign.status.toUpperCase()}
							</span>
						</div>
					</div>
					
					<div class="campaign-stats">
						<div class="stat-item">
							<label>Budget</label>
							<span class="stat-value">{formatCurrency(campaign.budget)}</span>
						</div>
						<div class="stat-item">
							<label>Spent</label>
							<span class="stat-value">{formatCurrency(campaign.spent)}</span>
						</div>
						<div class="stat-item">
							<label>Assets</label>
							<span class="stat-value">{campaign.asset_count}</span>
						</div>
						<div class="stat-item">
							<label>Required Assets</label>
							<span class="stat-value status-{campaign.has_required_assets ? 'complete' : 'incomplete'}">
								{campaign.has_required_assets ? 'Complete' : 'Incomplete'}
							</span>
						</div>
					</div>
					
					<div class="campaign-dates">
						<div class="date-item">
							<label>Start Date</label>
							<span>{formatDate(campaign.start_date)}</span>
						</div>
						<div class="date-item">
							<label>End Date</label>
							<span>{formatDate(campaign.end_date)}</span>
						</div>
					</div>
					
					<div class="campaign-actions">
						<button class="btn btn-outline" on:click={() => handleManageAssets(campaign)}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
								<polyline points="7,10 12,15 17,10"></polyline>
								<line x1="12" y1="15" x2="12" y2="3"></line>
							</svg>
							Manage Assets
						</button>
						<button class="btn btn-primary">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
								<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
							</svg>
							Edit
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Create Campaign Modal -->
{#if showCreateModal}
	<div class="modal-overlay" on:click={() => showCreateModal = false}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Create New Campaign</h2>
				<button class="modal-close" on:click={() => showCreateModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<form on:submit|preventDefault={handleCreateCampaign}>
				<div class="modal-body">
					<div class="form-group">
						<label for="campaign-name">Campaign Name</label>
						<input 
							id="campaign-name"
							type="text" 
							bind:value={newCampaign.name}
							placeholder="Enter campaign name"
							required
						/>
					</div>
					
					<div class="form-group">
						<label for="campaign-description">Description</label>
						<textarea 
							id="campaign-description"
							bind:value={newCampaign.description}
							placeholder="Describe your campaign"
							rows="3"
						></textarea>
					</div>
					
					<div class="form-row">
						<div class="form-group">
							<label for="campaign-budget">Budget</label>
							<input 
								id="campaign-budget"
								type="number" 
								bind:value={newCampaign.budget}
								placeholder="0.00"
								min="0"
								step="0.01"
								required
							/>
						</div>
						
						<div class="form-group">
							<label for="campaign-ad-type">Ad Type</label>
							<select id="campaign-ad-type" bind:value={newCampaign.ad_type}>
								<option value="banner">Banner</option>
								<option value="large">Large Banner</option>
								<option value="small">Small Banner</option>
							</select>
						</div>
					</div>
					
					<div class="form-row">
						<div class="form-group">
							<label for="campaign-start">Start Date</label>
							<input 
								id="campaign-start"
								type="date" 
								bind:value={newCampaign.start_date}
								required
							/>
						</div>
						
						<div class="form-group">
							<label for="campaign-end">End Date</label>
							<input 
								id="campaign-end"
								type="date" 
								bind:value={newCampaign.end_date}
								required
							/>
						</div>
					</div>
					
					<div class="form-group">
						<label for="campaign-audience">Target Audience</label>
						<input 
							id="campaign-audience"
							type="text" 
							bind:value={newCampaign.target_audience}
							placeholder="Describe your target audience"
						/>
					</div>
				</div>
				
				<div class="modal-footer">
					<button type="button" class="btn btn-outline" on:click={() => showCreateModal = false} disabled={creating}>
						Cancel
					</button>
					<button type="submit" class="btn btn-primary" disabled={creating}>
						{#if creating}
							<div class="loading-spinner small"></div>
							Creating...
						{:else}
							Create Campaign
						{/if}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Asset Management Modal -->
{#if showAssetModal && selectedCampaign}
	<div class="modal-overlay" on:click={() => showAssetModal = false}>
		<div class="modal-content large" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Manage Assets - {selectedCampaign.name}</h2>
				<button class="modal-close" on:click={() => showAssetModal = false} aria-label="Close modal">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-body">
				<div class="asset-requirements">
					<h3>Asset Requirements</h3>
					{#each selectedCampaign.asset_requirements as requirement}
						<div class="requirement-section">
							<h4>{requirement.ad_type.toUpperCase()} Ad Type</h4>
							<div class="requirements-grid">
								{#each requirement.required_assets as assetReq}
									<div class="requirement-item">
										<div class="requirement-header">
											<span class="asset-type">{assetReq.asset_type.toUpperCase()}</span>
											<span class="requirement-status" class:required={assetReq.is_required}>
												{assetReq.is_required ? 'Required' : 'Optional'}
											</span>
										</div>
										<p class="requirement-description">{assetReq.description}</p>
										<div class="requirement-specs">
											{#if assetReq.min_width && assetReq.max_width}
												<span>Size: {assetReq.min_width}×{assetReq.min_height}px</span>
											{/if}
											{#if assetReq.max_file_size_mb}
												<span>Max: {assetReq.max_file_size_mb}MB</span>
											{/if}
											{#if assetReq.allowed_formats}
												<span>Formats: {assetReq.allowed_formats.join(', ')}</span>
											{/if}
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				</div>
				
				<div class="asset-upload">
					<h3>Upload Assets</h3>
					<FileUpload
						campaignId={selectedCampaign.id}
						maxFileSize={5 * 1024 * 1024}
						allowedFormats={['jpg', 'jpeg', 'png', 'webp']}
						on:complete={handleAssetUpload}
						on:error={(e) => showToast(e.detail.error, 'error')}
					/>
				</div>
				
				{#if selectedCampaign.assets.length > 0}
					<div class="existing-assets">
						<h3>Existing Assets</h3>
						<div class="assets-grid">
							{#each selectedCampaign.assets as asset}
								<div class="asset-item">
									<div class="asset-preview">
										{#if asset.asset_type === 'image'}
											<img src={asset.file_path} alt={asset.alt_text || asset.file_name} />
										{/if}
									</div>
									<div class="asset-info">
										<h5>{asset.file_name}</h5>
										<p>{asset.width}×{asset.height}px</p>
										<span class="asset-status status-{asset.status}">{asset.status}</span>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.campaigns-page {
		padding: var(--space-xl);
		background: var(--bg-secondary);
		min-height: 100vh;
	}
	
	.page-header {
		margin-bottom: var(--space-2xl);
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
		margin: 0;
	}
	
	.header-actions {
		display: flex;
		gap: var(--space-md);
	}
	
	.usage-overview {
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
	}
	
	.usage-overview h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.usage-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}
	
	.usage-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}
	
	.usage-item label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 600;
	}
	
	.usage-value {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.usage-bar {
		width: 100%;
		height: 6px;
		background: var(--bg-secondary);
		border-radius: var(--radius-full);
		overflow: hidden;
	}
	
	.usage-fill {
		height: 100%;
		background: var(--primary);
		transition: width var(--transition-fast);
	}
	
	.loading-container,
	.error-container,
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
		text-align: center;
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
	
	.empty-icon {
		width: 64px;
		height: 64px;
		color: var(--text-secondary);
		margin-bottom: var(--space-md);
	}
	
	.empty-icon svg {
		width: 100%;
		height: 100%;
	}
	
	.empty-state h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}
	
	.empty-state p {
		color: var(--text-secondary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.campaigns-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: var(--space-xl);
	}
	
	.campaign-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		transition: all var(--transition-fast);
	}
	
	.campaign-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}
	
	.campaign-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-lg);
	}
	
	.campaign-info h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}
	
	.campaign-description {
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.5;
	}
	
	.status-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--white);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.campaign-stats {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}
	
	.stat-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}
	
	.stat-item label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 600;
	}
	
	.stat-value {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.stat-value.status-complete {
		color: var(--success);
	}
	
	.stat-value.status-incomplete {
		color: var(--error);
	}
	
	.campaign-dates {
		display: flex;
		justify-content: space-between;
		margin-bottom: var(--space-lg);
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-md);
	}
	
	.date-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}
	
	.date-item label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 600;
	}
	
	.date-item span {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
	}
	
	.campaign-actions {
		display: flex;
		gap: var(--space-md);
	}
	
	/* Modal Styles */
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
		max-width: 600px;
		width: 90%;
		max-height: 90vh;
		overflow-y: auto;
	}
	
	.modal-content.large {
		max-width: 900px;
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
	
	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-xl);
		border-top: 1px solid var(--border-color);
	}
	
	/* Form Styles */
	.form-group {
		margin-bottom: var(--space-lg);
	}
	
	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-md);
	}
	
	.form-group label {
		display: block;
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}
	
	.form-group input,
	.form-group textarea,
	.form-group select {
		width: 100%;
		padding: var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: border-color var(--transition-fast);
	}
	
	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		border-color: var(--primary);
	}
	
	/* Asset Management Styles */
	.asset-requirements {
		margin-bottom: var(--space-2xl);
	}
	
	.asset-requirements h3,
	.asset-upload h3,
	.existing-assets h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.requirement-section {
		margin-bottom: var(--space-xl);
	}
	
	.requirement-section h4 {
		font-size: var(--text-md);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-md) 0;
	}
	
	.requirements-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-md);
	}
	
	.requirement-item {
		padding: var(--space-md);
		background: var(--bg-glass);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
	}
	
	.requirement-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-sm);
	}
	
	.asset-type {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.requirement-status {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.requirement-status.required {
		background: var(--error-bg);
		color: var(--error-text);
	}
	
	.requirement-status:not(.required) {
		background: var(--info-bg);
		color: var(--info-text);
	}
	
	.requirement-description {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0 0 var(--space-sm) 0;
	}
	
	.requirement-specs {
		display: flex;
		flex-wrap: wrap;
		gap: var(--space-sm);
	}
	
	.requirement-specs span {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		background: var(--bg-secondary);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
	}
	
	.asset-upload {
		margin-bottom: var(--space-2xl);
	}
	
	.assets-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		gap: var(--space-md);
	}
	
	.asset-item {
		background: var(--bg-glass);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		overflow: hidden;
	}
	
	.asset-preview {
		width: 100%;
		height: 100px;
		overflow: hidden;
	}
	
	.asset-preview img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	
	.asset-info {
		padding: var(--space-sm);
	}
	
	.asset-info h5 {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	
	.asset-info p {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin: 0 0 var(--space-xs) 0;
	}
	
	.asset-status {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.asset-status.status-approved {
		background: var(--success-bg);
		color: var(--success-text);
	}
	
	.asset-status.status-pending {
		background: var(--warning-bg);
		color: var(--warning-text);
	}
	
	.asset-status.status-rejected {
		background: var(--error-bg);
		color: var(--error-text);
	}
	
	@media (max-width: 768px) {
		.campaigns-page {
			padding: var(--space-lg);
		}
		
		.header-content {
			flex-direction: column;
			align-items: stretch;
		}
		
		.campaigns-grid {
			grid-template-columns: 1fr;
		}
		
		.form-row {
			grid-template-columns: 1fr;
		}
		
		.modal-content {
			width: 95%;
		}
		
		.requirements-grid {
			grid-template-columns: 1fr;
		}
		
		.assets-grid {
			grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
		}
	}
</style> 