<script lang="ts">
	import { onMount } from 'svelte';
	import { showToast } from '$lib/toast';
	import type { AdAsset } from '$lib/types/advertising';
	
	let assets: AdAsset[] = [];
	let loading = true;
	let selectedAsset: AdAsset | null = null;
	let showApprovalModal = false;
	let approvalNotes = '';
	let processing = false;
	let filterStatus = 'all';
	let filterType = 'all';
	
	onMount(async () => {
		await loadAssets();
		loading = false;
	});
	
	async function loadAssets() {
		try {
			// Mock data - replace with actual API
			assets = [
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
					description: 'Main promotional banner for summer campaign',
					status: 'pending',
					created_at: '2024-06-15T10:30:00Z',
					updated_at: '2024-06-15T10:30:00Z'
				},
				{
					id: 2,
					campaign_id: 1,
					asset_type: 'image',
					file_name: 'sidebar_300x250.png',
					file_path: '/api/assets/sidebar_300x250.png',
					file_size: 189440,
					mime_type: 'image/png',
					width: 300,
					height: 250,
					alt_text: 'Archaeological Evidence Sidebar',
					description: 'Sidebar advertisement for archaeological findings',
					status: 'approved',
					approved_by: 1,
					approved_at: '2024-06-14T15:45:00Z',
					created_at: '2024-06-14T12:15:00Z',
					updated_at: '2024-06-14T15:45:00Z'
				},
				{
					id: 3,
					campaign_id: 2,
					asset_type: 'video',
					file_name: 'promo_video.mp4',
					file_path: '/api/assets/promo_video.mp4',
					file_size: 15728640,
					mime_type: 'video/mp4',
					width: 1920,
					height: 1080,
					duration: 30,
					alt_text: 'BOME Promotional Video',
					description: 'Short promotional video for video ad placements',
					status: 'rejected',
					approval_notes: 'Video quality is too low and contains copyrighted music. Please resubmit with higher quality and royalty-free audio.',
					rejected_by: 1,
					rejected_at: '2024-06-13T09:20:00Z',
					created_at: '2024-06-13T08:00:00Z',
					updated_at: '2024-06-13T09:20:00Z'
				}
			];
		} catch (err) {
			showToast('Failed to load assets', 'error');
			console.error('Error loading assets:', err);
		}
	}
	
	function handleApprove(asset: AdAsset) {
		selectedAsset = asset;
		approvalNotes = '';
		showApprovalModal = true;
	}
	
	async function processApproval(action: 'approve' | 'reject') {
		if (!selectedAsset) return;
		
		processing = true;
		
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			// Update asset status
			selectedAsset.status = action === 'approve' ? 'approved' : 'rejected';
			if (action === 'approve') {
				selectedAsset.approved_at = new Date().toISOString();
				selectedAsset.approved_by = 1; // Current admin user ID
			} else {
				selectedAsset.rejected_at = new Date().toISOString();
				selectedAsset.rejected_by = 1; // Current admin user ID
				selectedAsset.approval_notes = approvalNotes;
			}
			
			assets = assets; // Trigger reactivity
			
			showToast(`Asset ${action}d successfully`, 'success');
			showApprovalModal = false;
			selectedAsset = null;
			approvalNotes = '';
			
		} catch (err) {
			showToast(`Failed to ${action} asset`, 'error');
			console.error(`Error ${action}ing asset:`, err);
		} finally {
			processing = false;
		}
	}
	
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}
	
	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}
	
	function getStatusColor(status: string): string {
		const colors = {
			pending: 'var(--warning)',
			approved: 'var(--success)',
			rejected: 'var(--error)',
			processing: 'var(--info)'
		};
		return colors[status as keyof typeof colors] || 'var(--text-secondary)';
	}
	
	$: filteredAssets = assets.filter(asset => {
		const statusMatch = filterStatus === 'all' || asset.status === filterStatus;
		const typeMatch = filterType === 'all' || asset.asset_type === filterType;
		return statusMatch && typeMatch;
	});
	
	$: pendingCount = assets.filter(a => a.status === 'pending').length;
</script>

<svelte:head>
	<title>Asset Management - BOME Admin</title>
</svelte:head>

<div class="assets-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Asset Management</h1>
				<p>Review and approve advertiser asset uploads</p>
			</div>
			<div class="header-stats">
				<div class="stat-card">
					<span class="stat-number">{pendingCount}</span>
					<span class="stat-label">Pending Approval</span>
				</div>
			</div>
		</div>
	</div>
	
	<div class="filters-section glass">
		<div class="filters">
			<div class="filter-group">
				<label for="status-filter">Status</label>
				<select id="status-filter" bind:value={filterStatus}>
					<option value="all">All Statuses</option>
					<option value="pending">Pending</option>
					<option value="approved">Approved</option>
					<option value="rejected">Rejected</option>
				</select>
			</div>
			
			<div class="filter-group">
				<label for="type-filter">Asset Type</label>
				<select id="type-filter" bind:value={filterType}>
					<option value="all">All Types</option>
					<option value="image">Images</option>
					<option value="video">Videos</option>
					<option value="audio">Audio</option>
					<option value="document">Documents</option>
				</select>
			</div>
		</div>
		
		<div class="filter-results">
			<span class="results-count">
				{filteredAssets.length} asset{filteredAssets.length !== 1 ? 's' : ''}
			</span>
		</div>
	</div>
	
	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading assets...</p>
		</div>
	{:else if filteredAssets.length === 0}
		<div class="empty-state glass">
			<div class="empty-icon">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
					<polyline points="7,10 12,15 17,10"></polyline>
					<line x1="12" y1="15" x2="12" y2="3"></line>
				</svg>
			</div>
			<h3>No assets found</h3>
			<p>No assets match your current filters</p>
		</div>
	{:else}
		<div class="assets-grid">
			{#each filteredAssets as asset}
				<div class="asset-card glass">
					<div class="asset-preview">
						{#if asset.asset_type === 'image'}
							<img src={asset.file_path} alt={asset.alt_text || asset.file_name} />
						{:else if asset.asset_type === 'video'}
							<video controls>
								<source src={asset.file_path} type={asset.mime_type} />
								Your browser does not support the video tag.
							</video>
						{:else}
							<div class="file-placeholder">
								<div class="file-icon">
									{#if asset.asset_type === 'audio'}
										🎵
									{:else if asset.asset_type === 'document'}
										📄
									{:else}
										📎
									{/if}
								</div>
								<span class="file-type">{asset.asset_type.toUpperCase()}</span>
							</div>
						{/if}
						
						<div class="asset-status-overlay">
							<span class="status-badge" style="background: {getStatusColor(asset.status)}">
								{asset.status.toUpperCase()}
							</span>
						</div>
					</div>
					
					<div class="asset-info">
						<h4>{asset.file_name}</h4>
						
						{#if asset.description}
							<p class="asset-description">{asset.description}</p>
						{/if}
						
						<div class="asset-details">
							<div class="detail-item">
								<label>Size</label>
								<span>{formatFileSize(asset.file_size)}</span>
							</div>
							
							{#if asset.width && asset.height}
								<div class="detail-item">
									<label>Dimensions</label>
									<span>{asset.width} × {asset.height}px</span>
								</div>
							{/if}
							
							{#if asset.duration}
								<div class="detail-item">
									<label>Duration</label>
									<span>{asset.duration}s</span>
								</div>
							{/if}
							
							<div class="detail-item">
								<label>Type</label>
								<span>{asset.mime_type}</span>
							</div>
							
							<div class="detail-item">
								<label>Uploaded</label>
								<span>{formatDate(asset.created_at)}</span>
							</div>
						</div>
						
						{#if asset.approval_notes}
							<div class="approval-notes">
								<label>Notes</label>
								<p>{asset.approval_notes}</p>
							</div>
						{/if}
						
						{#if asset.status === 'approved' && asset.approved_at}
							<div class="approval-info success">
								<span>✓ Approved on {formatDate(asset.approved_at)}</span>
							</div>
						{:else if asset.status === 'rejected' && asset.rejected_at}
							<div class="approval-info error">
								<span>✗ Rejected on {formatDate(asset.rejected_at)}</span>
							</div>
						{/if}
					</div>
					
					<div class="asset-actions">
						{#if asset.status === 'pending'}
							<button class="btn btn-success" on:click={() => handleApprove(asset)}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20,6 9,17 4,12"></polyline>
								</svg>
								Review
							</button>
						{:else if asset.status === 'approved'}
							<button class="btn btn-outline" on:click={() => handleApprove(asset)}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<line x1="18" y1="6" x2="6" y2="18"></line>
									<line x1="6" y1="6" x2="18" y2="18"></line>
								</svg>
								Reject
							</button>
						{:else if asset.status === 'rejected'}
							<button class="btn btn-outline" on:click={() => handleApprove(asset)}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20,6 9,17 4,12"></polyline>
								</svg>
								Approve
							</button>
						{/if}
						
						<button class="btn btn-outline">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
								<circle cx="12" cy="12" r="3"></circle>
							</svg>
							View
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Approval Modal -->
{#if showApprovalModal && selectedAsset}
	<div class="modal-overlay" on:click={() => showApprovalModal = false}>
		<div class="modal-content" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Review Asset - {selectedAsset.file_name}</h2>
				<button class="modal-close" on:click={() => showApprovalModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			
			<div class="modal-body">
				<div class="asset-preview-large">
					{#if selectedAsset.asset_type === 'image'}
						<img src={selectedAsset.file_path} alt={selectedAsset.alt_text || selectedAsset.file_name} />
					{:else if selectedAsset.asset_type === 'video'}
						<video controls>
							<source src={selectedAsset.file_path} type={selectedAsset.mime_type} />
						</video>
					{:else}
						<div class="file-placeholder large">
							<div class="file-icon">
								{#if selectedAsset.asset_type === 'audio'}
									🎵
								{:else if selectedAsset.asset_type === 'document'}
									📄
								{:else}
									📎
								{/if}
							</div>
							<span class="file-type">{selectedAsset.asset_type.toUpperCase()}</span>
							<span class="file-name">{selectedAsset.file_name}</span>
						</div>
					{/if}
				</div>
				
				<div class="asset-details-modal">
					<h3>Asset Details</h3>
					<div class="details-grid">
						<div class="detail-item">
							<label>File Name</label>
							<span>{selectedAsset.file_name}</span>
						</div>
						<div class="detail-item">
							<label>File Size</label>
							<span>{formatFileSize(selectedAsset.file_size)}</span>
						</div>
						<div class="detail-item">
							<label>MIME Type</label>
							<span>{selectedAsset.mime_type}</span>
						</div>
						{#if selectedAsset.width && selectedAsset.height}
							<div class="detail-item">
								<label>Dimensions</label>
								<span>{selectedAsset.width} × {selectedAsset.height}px</span>
							</div>
						{/if}
						{#if selectedAsset.duration}
							<div class="detail-item">
								<label>Duration</label>
								<span>{selectedAsset.duration} seconds</span>
							</div>
						{/if}
					</div>
					
					{#if selectedAsset.description}
						<div class="description-section">
							<label>Description</label>
							<p>{selectedAsset.description}</p>
						</div>
					{/if}
					
					{#if selectedAsset.alt_text}
						<div class="alt-text-section">
							<label>Alt Text</label>
							<p>{selectedAsset.alt_text}</p>
						</div>
					{/if}
				</div>
				
				<div class="approval-section">
					<label for="approval-notes">Notes (optional for approval, required for rejection)</label>
					<textarea 
						id="approval-notes"
						bind:value={approvalNotes} 
						placeholder="Add notes about this asset review..."
						rows="4"
					></textarea>
				</div>
			</div>
			
			<div class="modal-footer">
				<button class="btn btn-outline" on:click={() => showApprovalModal = false} disabled={processing}>
					Cancel
				</button>
				
				{#if selectedAsset.status !== 'rejected'}
					<button 
						class="btn btn-error" 
						on:click={() => processApproval('reject')} 
						disabled={processing || !approvalNotes.trim()}
					>
						{#if processing}
							<div class="loading-spinner small"></div>
							Rejecting...
						{:else}
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<line x1="18" y1="6" x2="6" y2="18"></line>
								<line x1="6" y1="6" x2="18" y2="18"></line>
							</svg>
							Reject
						{/if}
					</button>
				{/if}
				
				{#if selectedAsset.status !== 'approved'}
					<button class="btn btn-success" on:click={() => processApproval('approve')} disabled={processing}>
						{#if processing}
							<div class="loading-spinner small"></div>
							Approving...
						{:else}
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="20,6 9,17 4,12"></polyline>
							</svg>
							Approve
						{/if}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	.assets-page {
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
	
	.header-stats {
		display: flex;
		gap: var(--space-md);
	}
	
	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-lg);
		background: var(--bg-glass);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		min-width: 120px;
	}
	
	.stat-number {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--warning);
		margin-bottom: var(--space-xs);
	}
	
	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		text-align: center;
	}
	
	.filters-section {
		padding: var(--space-lg);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border-color);
		margin-bottom: var(--space-xl);
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	
	.filters {
		display: flex;
		gap: var(--space-lg);
	}
	
	.filter-group {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}
	
	.filter-group label {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
	}
	
	.filter-group select {
		padding: var(--space-sm) var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: var(--text-sm);
	}
	
	.results-count {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		font-weight: 500;
	}
	
	.loading-container,
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 400px;
		gap: var(--space-lg);
		text-align: center;
	}
	
	.empty-state {
		padding: var(--space-2xl);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border-color);
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
		margin: 0;
	}
	
	.empty-state p {
		color: var(--text-secondary);
		margin: 0;
	}
	
	.assets-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: var(--space-xl);
	}
	
	.asset-card {
		border-radius: var(--radius-xl);
		border: 1px solid var(--border-color);
		overflow: hidden;
		transition: all var(--transition-fast);
	}
	
	.asset-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}
	
	.asset-preview {
		position: relative;
		width: 100%;
		height: 200px;
		overflow: hidden;
		background: var(--bg-secondary);
	}
	
	.asset-preview img,
	.asset-preview video {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	
	.file-placeholder {
		width: 100%;
		height: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		background: var(--bg-glass);
	}
	
	.file-placeholder.large {
		height: 300px;
		gap: var(--space-md);
	}
	
	.file-icon {
		font-size: 3rem;
	}
	
	.file-type {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.file-name {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-align: center;
		word-break: break-all;
		padding: 0 var(--space-sm);
	}
	
	.asset-status-overlay {
		position: absolute;
		top: var(--space-sm);
		right: var(--space-sm);
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
	
	.asset-info {
		padding: var(--space-lg);
	}
	
	.asset-info h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
		word-break: break-all;
	}
	
	.asset-description {
		color: var(--text-secondary);
		margin: 0 0 var(--space-md) 0;
		line-height: 1.5;
	}
	
	.asset-details {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--space-sm);
		margin-bottom: var(--space-md);
	}
	
	.detail-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}
	
	.detail-item label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.detail-item span {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
	}
	
	.approval-notes {
		margin-bottom: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--info);
	}
	
	.approval-notes label {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		display: block;
		margin-bottom: var(--space-xs);
	}
	
	.approval-notes p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.5;
	}
	
	.approval-info {
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		font-size: var(--text-sm);
		font-weight: 500;
		margin-bottom: var(--space-md);
	}
	
	.approval-info.success {
		background: var(--success-bg);
		color: var(--success-text);
		border-left: 4px solid var(--success);
	}
	
	.approval-info.error {
		background: var(--error-bg);
		color: var(--error-text);
		border-left: 4px solid var(--error);
	}
	
	.asset-actions {
		padding: var(--space-lg);
		border-top: 1px solid var(--border-color);
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
		max-width: 800px;
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
	
	.asset-preview-large {
		margin-bottom: var(--space-xl);
		border-radius: var(--radius-lg);
		overflow: hidden;
		background: var(--bg-secondary);
	}
	
	.asset-preview-large img,
	.asset-preview-large video {
		width: 100%;
		max-height: 400px;
		object-fit: contain;
	}
	
	.asset-details-modal h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.details-grid {
		display: grid;
		grid-template-columns: repeat(2, 1fr);
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
	}
	
	.description-section,
	.alt-text-section {
		margin-bottom: var(--space-lg);
	}
	
	.description-section label,
	.alt-text-section label {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		display: block;
		margin-bottom: var(--space-sm);
	}
	
	.description-section p,
	.alt-text-section p {
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.5;
	}
	
	.approval-section {
		margin-top: var(--space-xl);
		padding-top: var(--space-xl);
		border-top: 1px solid var(--border-color);
	}
	
	.approval-section label {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		display: block;
		margin-bottom: var(--space-sm);
	}
	
	.approval-section textarea {
		width: 100%;
		padding: var(--space-md);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-md);
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: var(--text-sm);
		resize: vertical;
		min-height: 100px;
	}
	
	.approval-section textarea:focus {
		outline: none;
		border-color: var(--primary);
	}
	
	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-xl);
		border-top: 1px solid var(--border-color);
	}
	
	@media (max-width: 768px) {
		.assets-page {
			padding: var(--space-lg);
		}
		
		.header-content {
			flex-direction: column;
			align-items: stretch;
		}
		
		.filters-section {
			flex-direction: column;
			align-items: stretch;
			gap: var(--space-md);
		}
		
		.filters {
			flex-direction: column;
			gap: var(--space-md);
		}
		
		.assets-grid {
			grid-template-columns: 1fr;
		}
		
		.asset-details {
			grid-template-columns: 1fr;
		}
		
		.modal-content {
			width: 95%;
		}
		
		.details-grid {
			grid-template-columns: 1fr;
		}
	}
</style> 