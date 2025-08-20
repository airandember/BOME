<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { masterVideoService } from '$lib/master-video';
	import { showToast } from '$lib/toast';

	// Props
	export let isOpen: boolean = false;

	// Events
	const dispatch = createEventDispatcher<{
		close: void;
	}>();

	// Local state
	let tagAnalytics: any = null;
	let isLoading = true;
	let isTagging = false;

	// Single progress tracking system
	let isProcessing = false;
	let currentCount = 0;
	let totalCount = 0;
	
	// Close modal
	function closeModal() {
		dispatch('close');
	}

	// Load tag analytics
	async function loadTagAnalytics() {
		try {
			isLoading = true;
			const response = await masterVideoService.getTagAnalytics();
			console.log('Tag analytics response:', response);
			if (response.success && response.data) {
				tagAnalytics = response.data;
				console.log('Tag analytics data:', tagAnalytics);
			} else {
				console.error('Tag analytics response failed:', response);
				showToast('❌ Failed to load tag analytics', 'error');
			}
		} catch (error) {
			console.error('Failed to load tag analytics:', error);
			showToast('❌ Failed to load tag analytics', 'error');
		} finally {
			isLoading = false;
		}
	}

	// Tag untagged videos (ADDITIVE - only adds tags to videos without tags)
	async function tagUntaggedVideos() {
		try {
			isTagging = true;
			showToast('🔄 Starting to tag untagged videos...', 'info');
			
			// Get ALL untagged videos (use high limit to get everything)
			const response = await masterVideoService.getUntaggedVideos(10000);
			if (response.success && response.videos) {
				const untaggedVideos = response.videos;
				const totalVideos = untaggedVideos.length;
				showToast(`📝 Found ${totalVideos} untagged videos to process`, 'info');
				
				if (totalVideos === 0) {
					showToast('✅ All videos already have tags!', 'success');
					return;
				}
				
				// Extract video IDs for batch processing
				const videoIDs = untaggedVideos.map(video => video.ID);
				
				// Calculate batch information
				const batchSize = 100;
				const totalBatches = Math.ceil(videoIDs.length / batchSize);
				
				// Initialize progress tracking
				isProcessing = true;
				currentCount = 0;
				totalCount = totalVideos;

				try {
					// Process each batch with progress updates
					for (let batchIndex = 0; batchIndex < totalBatches; batchIndex++) {
						// Get the current batch of video IDs
						const startIdx = batchIndex * batchSize;
						const endIdx = Math.min(startIdx + batchSize, videoIDs.length);
						const batchVideoIDs = videoIDs.slice(startIdx, endIdx);
						
						console.log(`Processing batch ${batchIndex + 1}/${totalBatches} with ${batchVideoIDs.length} videos`);
						
						// Process this batch (ADDITIVE - won't overwrite existing tags)
						const batchResponse = await masterVideoService.batchAutoTagVideos(batchVideoIDs);
						
						if (!batchResponse.success) {
							throw new Error(`Batch ${batchIndex + 1} failed: ${batchResponse.error}`);
						}
						
						// Update progress: add the videos we just processed
						currentCount += batchVideoIDs.length;
						
						// Small delay between batches to prevent overwhelming the server
						if (batchIndex < totalBatches - 1) {
							await new Promise(resolve => setTimeout(resolve, 200));
						}
					}
					
					isProcessing = false;
					showToast('✅ All untagged videos processed successfully!', 'success');
					
				} catch (error) {
					console.error('Batch tagging error:', error);
					isProcessing = false;
					showToast(`❌ Batch tagging failed: ${error}`, 'error');
				}
				
				// Final reload of analytics to show updated stats
				await loadTagAnalytics();
			} else {
				showToast('❌ Failed to get untagged videos', 'error');
			}
		} catch (error) {
			console.error('Failed to tag untagged videos:', error);
			showToast('❌ Failed to tag untagged videos', 'error');
		} finally {
			// Keep progress visible for a moment to show completion
			setTimeout(() => {
				isTagging = false;
				isProcessing = false;
				currentCount = 0;
				totalCount = 0;
			}, 2000);
		}
	}

	// Tag all videos (REPLACES existing tags - clears and repopulates)
	async function tagAllVideos() {
		try {
			isTagging = true;
			showToast('🔄 Starting to tag ALL videos (this will replace existing tags)...', 'warning');
			
			// First get the total count to know how many videos we're dealing with
			const countResponse = await masterVideoService.getMasterVideos({ page: 1, limit: 1 });
			if (!countResponse.success) {
				showToast('❌ Failed to get video count', 'error');
				return;
			}
			
			const totalVideos = countResponse.pagination.total;
			showToast(`📝 Found ${totalVideos} total videos to process (replacing existing tags)`, 'info');
			
			if (totalVideos === 0) {
				showToast('✅ No videos found to tag', 'success');
				return;
			}
			
			// Process videos in batches of 100
			const batchSize = 100;
			const totalBatches = Math.ceil(totalVideos / batchSize);
			
			// Initialize progress tracking
			isProcessing = true;
			currentCount = 0;
			totalCount = totalVideos;

			try {
				// Process each batch with progress updates
				for (let batchIndex = 0; batchIndex < totalBatches; batchIndex++) {
					// Get the current batch of videos
					const page = batchIndex + 1;
					const batchResponse = await masterVideoService.getMasterVideos({ 
						page, 
						limit: batchSize 
					});
					
					if (!batchResponse.success || !batchResponse.videos) {
						throw new Error(`Failed to get batch ${page}: Unknown error`);
					}
					
					const batchVideos = batchResponse.videos;
					const videoIDs = batchVideos.map(video => video.ID);
					
					console.log(`Processing batch ${batchIndex + 1}/${totalBatches} with ${batchVideos.length} videos`);
					
					// FIX: Only send replace=true for the FIRST batch
					const isFirstBatch = batchIndex === 0;
					const tagResponse = await masterVideoService.batchAutoTagVideos(videoIDs, isFirstBatch);
					
					if (!tagResponse.success) {
						throw new Error(`Batch ${batchIndex + 1} failed: ${tagResponse.error}`);
					}
					
					// Update progress: add the videos we just processed
					currentCount += batchVideos.length;
					
					// Small delay between batches to prevent overwhelming the server
					if (batchIndex < totalBatches - 1) {
						await new Promise(resolve => setTimeout(resolve, 200));
					}
				}
				
				isProcessing = false;
				showToast('✅ All videos re-tagged successfully! (existing tags replaced)', 'success');
				
			} catch (error) {
				console.error('Batch tagging error:', error);
				isProcessing = false;
				showToast(`❌ Batch tagging failed: ${error}`, 'error');
			}
			
			// Final reload of analytics to show updated stats
			await loadTagAnalytics();
		} catch (error) {
			console.error('Failed to tag all videos', error);
			showToast('❌ Failed to tag all videos', 'error');
		} finally {
			// Keep progress visible for a moment to show completion
			setTimeout(() => {
				isTagging = false;
				isProcessing = false;
				currentCount = 0;
				totalCount = 0;
			}, 2000);
		}
	}



	// Load analytics when modal opens
	$: if (isOpen && !tagAnalytics) {
		loadTagAnalytics();
	}
</script>

{#if isOpen}
	<div class="modal-backdrop" on:click={closeModal} transition:fade={{ duration: 200 }}>
		<div class="modal-content modal-large" on:click|stopPropagation transition:fly={{ y: 20, duration: 200 }}>
			<div class="modal-header">
				<h2>📊 Tag Analytics Dashboard</h2>
				<button class="modal-close" on:click={closeModal}>×</button>
			</div>
			
			<div class="modal-body">
				{#if isLoading}
					<div class="loading-state">
						<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-orange-500 mx-auto"></div>
						<p class="text-center mt-4">Loading tag analytics...</p>
					</div>
				{:else if tagAnalytics}
					<div class="analytics-overview">
						<div class="stats-grid">
							<div class="stat-card">
								<div class="stat-icon">📹</div>
								<div class="stat-content">
									<div class="stat-value">{tagAnalytics.total_videos || 0}</div>
									<div class="stat-label">Total Videos</div>
								</div>
							</div>
							<div class="stat-card">
								<div class="stat-icon">🏷️</div>
								<div class="stat-content">
									<div class="stat-value">{tagAnalytics.tagged_videos || 0}</div>
									<div class="stat-label">Tagged Videos</div>
								</div>
							</div>
							<div class="stat-card">
								<div class="stat-icon">⏳</div>
								<div class="stat-content">
									<div class="stat-value">{tagAnalytics.untagged_videos || 0}</div>
									<div class="stat-label">Untagged Videos</div>
								</div>
							</div>
							<div class="stat-card">
								<div class="stat-icon">📈</div>
								<div class="stat-content">
									<div class="stat-value">{tagAnalytics.tagging_percentage ? tagAnalytics.tagging_percentage.toFixed(1) : '0.0'}%</div>
									<div class="stat-label">Completion Rate</div>
								</div>
							</div>
						</div>
					</div>

					<div class="analytics-sections">
						<!-- Top Tags by Frequency -->
						<div class="analytics-section">
							<div class="section-header">
								<h3 class="text-lg font-semibold text-gray-800">🏷️ Top Tags by Frequency</h3>
								{#if isTagging}
									<span class="live-indicator">🔄 Live Updates</span>
								{/if}
							</div>
							{#if tagAnalytics.tag_frequency && tagAnalytics.tag_frequency.length > 0}
								<div class="tags-grid">
									{#each tagAnalytics.tag_frequency.slice(0, 20) as tag}
										<div class="tag-item">
											<span class="tag-word">{tag.word}</span>
											<span class="tag-frequency">{tag.frequency}</span>
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-gray-500 text-center py-4">No tag frequency data available</p>
							{/if}
						</div>

						<div class="section">
							<div class="section-header">
								<h3>📊 Tag Statistics</h3>
								{#if isTagging}
									<span class="live-indicator">🔄 Live Updates</span>
								{/if}
							</div>
							<div class="stats-grid">
								<div class="stat-item">
									<span class="stat-label">Total Unique Tags:</span>
									<span class="stat-value">{tagAnalytics.total_unique_tags || 0}</span>
								</div>
								<div class="stat-item">
									<span class="stat-label">Most Frequent Tag:</span>
									<span class="stat-value">
										{tagAnalytics.tag_frequency && tagAnalytics.tag_frequency.length > 0 
											? tagAnalytics.tag_frequency[0].word 
											: 'None'}
									</span>
								</div>
							</div>
						</div>
					</div>
				{:else}
					<div class="error-state">
						<p class="text-center text-red-500">Failed to load analytics data</p>
					</div>
				{/if}
			</div>

			<!-- Single Progress Bar -->
			{#if isProcessing}
				<div class="progress-container">
					<div class="progress-header">
						<h4>Tagging in Progress</h4>
						<p class="progress-message">Processing videos: {currentCount} of {totalCount}</p>
					</div>
					
					<div class="progress-bar">
						<div class="progress-fill" style="width: {totalCount > 0 ? (currentCount / totalCount) * 100 : 0}%"></div>
					</div>
					
					<div class="progress-stats">
						<span>{Math.round(totalCount > 0 ? (currentCount / totalCount) * 100 : 0)}% Complete</span>
					</div>
				</div>
			{/if}

			<div class="modal-footer">
				<!-- Tagging Instructions -->
				<div class="tagging-instructions">
					<div class="instruction-item">
						<span class="instruction-icon">🟢</span>
						<strong>Tag Untagged:</strong> Only adds tags to videos that don't have any tags yet (additive)
					</div>
					<div class="instruction-item">
						<span class="instruction-icon">🔵</span>
						<strong>Tag All:</strong> Replaces ALL existing tags with new ones for every video (destructive)
					</div>
				</div>
				
				<div class="tagButtons">
					<button class="btn bg-green-500 text-white px-4 py-2 rounded-lg hover:bg-green-600 transition-colors" on:click={tagUntaggedVideos} disabled={isTagging || isProcessing}>
						{isTagging || isProcessing ? 'Tagging...' : 'Tag Untagged'}
					</button>
					<button class="btn bg-blue-500 text-white px-4 py-2 rounded-lg hover:bg-blue-600 transition-colors" on:click={tagAllVideos} disabled={isTagging || isProcessing}>
						{isTagging || isProcessing ? 'Tagging...' : 'Tag All'}
					</button>
				</div>
				<button class="btn bg-orange-500 text-white px-4 py-2 rounded-lg hover:bg-orange-600 transition-colors" on:click={closeModal}>
					Close Dashboard
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 2rem;
	}

	.modal-content {
		background: var(--bg-primary, rgba(0, 0, 0, 0.9));
		border: 1px solid var(--border-color, rgba(255, 255, 255, 0.1));
		border-radius: 12px;
		padding: 2.5rem;
		width: 95%;
		max-width: 1200px;
		max-height: 90vh;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 2rem;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
	}

	.modal-content.modal-large {
		max-width: 1400px;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.modal-header h2 {
		margin: 0;
		color: var(--text-primary, #ffffff);
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		font-size: 2rem;
		cursor: pointer;
		padding: 0;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		transition: all 0.3s ease;
	}

	.modal-close:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.modal-body {
		flex: 1;
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 200px;
	}

	.animate-spin {
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		from { transform: rotate(0deg); }
		to { transform: rotate(360deg); }
	}

	.analytics-overview {
		margin-bottom: 2rem;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
	}

	.stat-card {
		background: rgba(255, 255, 255, 0.1);
		border-radius: 12px;
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.stat-icon {
		width: 50px;
		height: 50px;
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1.5rem;
		background: rgba(59, 130, 246, 0.2);
	}

	.stat-content {
		flex: 1;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.8rem;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 500;
	}

	.analytics-sections {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 2rem;
	}

	.analytics-section h3 {
		margin: 0 0 1rem 0;
		color: var(--text-primary);
		font-size: 1.1rem;
		font-weight: 600;
	}

	.tag-frequency-list {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		padding: 1rem;
		max-height: 300px;
		overflow-y: auto;
	}

	.tag-frequency-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.5rem;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.tag-frequency-item:last-child {
		border-bottom: none;
	}

	.tag-word {
		color: var(--text-primary);
		font-weight: 500;
	}

	.tag-count {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
		padding: 0.25rem 0.5rem;
		border-radius: 12px;
		font-size: 0.8rem;
		font-weight: 600;
	}

	.categories-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.category-card {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		padding: 1rem;
		border-left: 4px solid #6b7280;
	}

	.category-card h4 {
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
		font-size: 1rem;
		font-weight: 600;
	}

	.category-card p {
		margin: 0;
		color: var(--text-secondary);
		font-size: 0.8rem;
		line-height: 1.4;
	}

	.modal-footer {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
	}

	.btn:hover {
		transform: translateY(-1px);
	}

	.error-state {
		text-align: center;
		padding: 2rem;
	}

	.empty-state {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary);
		font-style: italic;
	}

	.progress-section {
		background: rgba(255, 255, 255, 0.05);
		border-radius: 8px;
		padding: 1rem;
		margin-top: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.progress-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.progress-header h4 {
		margin: 0;
		color: var(--text-primary);
		font-size: 1rem;
		font-weight: 600;
	}

	.progress-message {
		margin: 0;
		color: var(--text-secondary);
		font-size: 0.8rem;
		font-style: italic;
	}

	.progress-bar {
		height: 8px;
		background: rgba(255, 255, 255, 0.2);
		border-radius: 4px;
		margin-bottom: 0.5rem;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: var(--primary-color, #3b82f6);
		border-radius: 4px;
		transition: width 0.3s ease-in-out;
	}

	.progress-stats {
		display: flex;
		justify-content: space-between;
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	.batch-info {
		margin-top: 0.5rem;
		text-align: center;
		padding: 0.5rem;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 6px;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.real-time-updates {
		margin-top: 0.75rem;
		padding-top: 0.75rem;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.update-item {
		padding: 0.5rem 0;
		font-size: 0.85rem;
		color: var(--text-primary);
	}

	.update-item.success {
		color: var(--success-color, #22c55e);
	}

	.update-item.error {
		color: var(--error-color, #ef4444);
	}

	.live-indicator {
		background: var(--primary-color, #3b82f6);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
		margin-left: 1rem;
		white-space: nowrap;
		animation: pulse-live 2s ease-in-out infinite;
	}

	@keyframes pulse-live {
		0%, 100% {
			opacity: 1;
			transform: scale(1);
		}
		50% {
			opacity: 0.8;
			transform: scale(1.05);
		}
	}

	.tags-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: 0.75rem;
		padding: 0.75rem 0;
	}

	.tag-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.5rem 0.75rem;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 6px;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.tag-word {
		font-weight: 600;
		color: var(--text-primary);
	}

	.tag-frequency {
		font-size: 0.875rem;
		color: var(--text-secondary);
		background: rgba(59, 130, 246, 0.2);
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-weight: 600;
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
	}

	.stat-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 0.75rem 1rem;
		background: rgba(255, 255, 255, 0.08);
		border-radius: 8px;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.stat-label {
		font-size: 0.8rem;
		color: var(--text-secondary);
		font-weight: 500;
	}

	.stat-value {
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--text-primary);
	}

	@media (max-width: 768px) {
		.modal-content {
			width: 98%;
			padding: 1.5rem;
			margin: 0.5rem;
		}

		.analytics-sections {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	/* Simple progress bar styles */
	.progress-container {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		padding: 1.5rem;
		margin: 1rem 0;
	}

	.progress-header h4 {
		margin: 0 0 0.5rem 0;
		color: var(--text-primary, #ffffff);
		font-size: 1.1rem;
		font-weight: 600;
	}

	.progress-message {
		margin: 0 0 1rem 0;
		color: var(--text-secondary, #a0a0a0);
		font-size: 0.9rem;
	}

	.progress-bar {
		width: 100%;
		height: 8px;
		background: rgba(255, 255, 255, 0.1);
		border-radius: 4px;
		overflow: hidden;
		margin-bottom: 1rem;
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, #3b82f6, #10b981);
		border-radius: 4px;
		transition: width 0.3s ease;
	}

	.progress-stats {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: 0.875rem;
		color: var(--text-secondary, #a0a0a0);
	}

	/* Tagging Instructions */
	.tagging-instructions {
		background: rgba(255, 255, 255, 0.05);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1rem;
	}

	.instruction-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.instruction-item:last-child {
		margin-bottom: 0;
	}

	.instruction-icon {
		font-size: 1rem;
	}

	.tagButtons {
		display: flex;
		gap: 1rem;
		margin-bottom: 1rem;
	}
</style>
