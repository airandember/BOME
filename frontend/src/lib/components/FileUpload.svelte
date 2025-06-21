<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { FileUploadProgress, AdAsset } from '$lib/types/advertising';
	
	export let accept: string = 'image/*,video/*';
	export let maxFileSize: number = 10 * 1024 * 1024; // 10MB default
	export let maxFiles: number = 10;
	export let campaignId: number;
	export let adId: number | undefined = undefined;
	export let assetType: 'image' | 'video' | 'audio' | 'document' | 'banner' | 'logo' = 'image';
	export let className: string = '';
	export let disabled: boolean = false;
	export let showPreview: boolean = true;
	export let allowedFormats: string[] = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'mp4', 'webm', 'pdf'];
	
	const dispatch = createEventDispatcher<{
		upload: { files: File[] };
		progress: { progress: FileUploadProgress[] };
		complete: { assets: AdAsset[] };
		error: { error: string };
	}>();
	
	let fileInput: HTMLInputElement;
	let dragOver = false;
	let uploading = false;
	let uploadProgress: FileUploadProgress[] = [];
	let uploadedAssets: AdAsset[] = [];
	let errors: string[] = [];
	
	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files) {
			handleFiles(Array.from(target.files));
		}
	}
	
	function handleDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		
		if (disabled || !event.dataTransfer) return;
		
		const files = Array.from(event.dataTransfer.files);
		handleFiles(files);
	}
	
	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		if (!disabled) {
			dragOver = true;
		}
	}
	
	function handleDragLeave(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
	}
	
	function handleFiles(files: File[]) {
		errors = [];
		
		// Validate file count
		if (files.length > maxFiles) {
			errors.push(`Maximum ${maxFiles} files allowed`);
			return;
		}
		
		// Validate each file
		const validFiles: File[] = [];
		for (const file of files) {
			const validation = validateFile(file);
			if (validation.valid) {
				validFiles.push(file);
			} else {
				errors.push(`${file.name}: ${validation.error}`);
			}
		}
		
		if (validFiles.length === 0) {
			dispatch('error', { error: 'No valid files selected' });
			return;
		}
		
		if (errors.length > 0) {
			dispatch('error', { error: errors.join(', ') });
		}
		
		uploadFiles(validFiles);
	}
	
	function validateFile(file: File): { valid: boolean; error?: string } {
		// Check file size
		if (file.size > maxFileSize) {
			return {
				valid: false,
				error: `File size exceeds ${formatFileSize(maxFileSize)} limit`
			};
		}
		
		// Check file format
		const extension = file.name.split('.').pop()?.toLowerCase();
		if (!extension || !allowedFormats.includes(extension)) {
			return {
				valid: false,
				error: `File format not allowed. Allowed: ${allowedFormats.join(', ')}`
			};
		}
		
		// Check MIME type
		const allowedMimeTypes = [
			'image/jpeg', 'image/jpg', 'image/png', 'image/gif', 'image/webp',
			'video/mp4', 'video/webm', 'video/quicktime',
			'audio/mpeg', 'audio/wav', 'audio/ogg',
			'application/pdf'
		];
		
		if (!allowedMimeTypes.includes(file.type)) {
			return {
				valid: false,
				error: 'Invalid file type'
			};
		}
		
		return { valid: true };
	}
	
	async function uploadFiles(files: File[]) {
		uploading = true;
		uploadProgress = files.map(file => ({
			file_name: file.name,
			file_size: file.size,
			uploaded_bytes: 0,
			percentage: 0,
			status: 'pending' as const,
		}));
		
		dispatch('upload', { files });
		dispatch('progress', { progress: uploadProgress });
		
		try {
			const uploadPromises = files.map((file, index) => uploadSingleFile(file, index));
			const results = await Promise.allSettled(uploadPromises);
			
			const successfulUploads: AdAsset[] = [];
			const failedUploads: string[] = [];
			
			results.forEach((result, index) => {
				if (result.status === 'fulfilled') {
					successfulUploads.push(result.value);
					uploadProgress[index].status = 'completed';
					uploadProgress[index].percentage = 100;
				} else {
					failedUploads.push(`${files[index].name}: ${result.reason}`);
					uploadProgress[index].status = 'error';
					uploadProgress[index].error_message = result.reason;
				}
			});
			
			uploadedAssets = [...uploadedAssets, ...successfulUploads];
			
			if (failedUploads.length > 0) {
				dispatch('error', { error: failedUploads.join(', ') });
			}
			
			if (successfulUploads.length > 0) {
				dispatch('complete', { assets: successfulUploads });
			}
			
		} catch (error) {
			dispatch('error', { error: error instanceof Error ? error.message : 'Upload failed' });
		} finally {
			uploading = false;
			dispatch('progress', { progress: uploadProgress });
		}
	}
	
	async function uploadSingleFile(file: File, progressIndex: number): Promise<AdAsset> {
		// Update progress to uploading
		uploadProgress[progressIndex].status = 'uploading';
		dispatch('progress', { progress: uploadProgress });
		
		// Create FormData
		const formData = new FormData();
		formData.append('file', file);
		formData.append('campaign_id', campaignId.toString());
		if (adId) formData.append('ad_id', adId.toString());
		formData.append('asset_type', assetType);
		
		// Create XMLHttpRequest for progress tracking
		return new Promise((resolve, reject) => {
			const xhr = new XMLHttpRequest();
			
			xhr.upload.addEventListener('progress', (event) => {
				if (event.lengthComputable) {
					const percentage = Math.round((event.loaded / event.total) * 100);
					uploadProgress[progressIndex].uploaded_bytes = event.loaded;
					uploadProgress[progressIndex].percentage = percentage;
					dispatch('progress', { progress: uploadProgress });
				}
			});
			
			xhr.addEventListener('load', () => {
				if (xhr.status >= 200 && xhr.status < 300) {
					try {
						const response = JSON.parse(xhr.responseText);
						if (response.success) {
							uploadProgress[progressIndex].status = 'processing';
							resolve(response.data);
						} else {
							reject(response.error || 'Upload failed');
						}
					} catch (error) {
						reject('Invalid response format');
					}
				} else {
					reject(`Upload failed with status ${xhr.status}`);
				}
			});
			
			xhr.addEventListener('error', () => {
				reject('Network error during upload');
			});
			
			xhr.addEventListener('abort', () => {
				reject('Upload cancelled');
			});
			
			xhr.open('POST', '/api/v1/advertiser/assets/upload');
			xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('auth_token')}`);
			xhr.send(formData);
		});
	}
	
	function removeFile(index: number) {
		uploadProgress = uploadProgress.filter((_, i) => i !== index);
		uploadedAssets = uploadedAssets.filter((_, i) => i !== index);
	}
	
	function formatFileSize(bytes: number): string {
		if (bytes === 0) return '0 Bytes';
		const k = 1024;
		const sizes = ['Bytes', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
	}
	
	function getFileIcon(fileName: string): string {
		const extension = fileName.split('.').pop()?.toLowerCase();
		const iconMap: { [key: string]: string } = {
			'jpg': '🖼️', 'jpeg': '🖼️', 'png': '🖼️', 'gif': '🖼️', 'webp': '🖼️',
			'mp4': '🎥', 'webm': '🎥', 'mov': '🎥',
			'mp3': '🎵', 'wav': '🎵', 'ogg': '🎵',
			'pdf': '📄', 'doc': '📄', 'docx': '📄'
		};
		return iconMap[extension || ''] || '📎';
	}
	
	function triggerFileSelect() {
		if (!disabled) {
			fileInput.click();
		}
	}
</script>

<div class="file-upload-container {className}">
	<input
		bind:this={fileInput}
		type="file"
		{accept}
		multiple={maxFiles > 1}
		on:change={handleFileSelect}
		class="file-input"
		{disabled}
	/>
	
	<div
		class="upload-zone"
		class:drag-over={dragOver}
		class:disabled
		class:uploading
		on:drop={handleDrop}
		on:dragover={handleDragOver}
		on:dragleave={handleDragLeave}
		on:click={triggerFileSelect}
		role="button"
		tabindex="0"
		on:keydown={(e) => e.key === 'Enter' && triggerFileSelect()}
	>
		<div class="upload-content">
			<div class="upload-icon">
				{#if uploading}
					<div class="loading-spinner"></div>
				{:else}
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="7,10 12,15 17,10"></polyline>
						<line x1="12" y1="15" x2="12" y2="3"></line>
					</svg>
				{/if}
			</div>
			
			<div class="upload-text">
				{#if uploading}
					<h3>Uploading files...</h3>
					<p>Please wait while we process your files</p>
				{:else}
					<h3>Drop files here or click to browse</h3>
					<p>
						Supports: {allowedFormats.join(', ')} 
						• Max {formatFileSize(maxFileSize)} per file
						• Up to {maxFiles} files
					</p>
				{/if}
			</div>
		</div>
	</div>
	
	{#if errors.length > 0}
		<div class="error-messages">
			{#each errors as error}
				<div class="error-message">{error}</div>
			{/each}
		</div>
	{/if}
	
	{#if uploadProgress.length > 0}
		<div class="upload-progress">
			<h4>Upload Progress</h4>
			{#each uploadProgress as progress, index}
				<div class="progress-item">
					<div class="progress-header">
						<div class="file-info">
							<span class="file-icon">{getFileIcon(progress.file_name)}</span>
							<span class="file-name">{progress.file_name}</span>
							<span class="file-size">({formatFileSize(progress.file_size)})</span>
						</div>
						
						<div class="progress-actions">
							<span class="progress-status status-{progress.status}">
								{progress.status}
							</span>
							{#if progress.status === 'completed' || progress.status === 'error'}
								<button 
									class="remove-btn" 
									on:click={() => removeFile(index)}
									aria-label="Remove file"
								>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"></line>
										<line x1="6" y1="6" x2="18" y2="18"></line>
									</svg>
								</button>
							{/if}
						</div>
					</div>
					
					<div class="progress-bar">
						<div 
							class="progress-fill status-{progress.status}" 
							style="width: {progress.percentage}%"
						></div>
					</div>
					
					<div class="progress-details">
						<span class="progress-percentage">{progress.percentage}%</span>
						<span class="progress-bytes">
							{formatFileSize(progress.uploaded_bytes)} / {formatFileSize(progress.file_size)}
						</span>
					</div>
					
					{#if progress.error_message}
						<div class="progress-error">{progress.error_message}</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
	
	{#if showPreview && uploadedAssets.length > 0}
		<div class="uploaded-assets">
			<h4>Uploaded Assets</h4>
			<div class="assets-grid">
				{#each uploadedAssets as asset}
					<div class="asset-preview">
						{#if asset.asset_type === 'image'}
							<img src={asset.file_path} alt={asset.alt_text || asset.file_name} />
						{:else if asset.asset_type === 'video'}
							<video controls>
								<source src={asset.file_path} type={asset.mime_type} />
								Your browser does not support the video tag.
							</video>
						{:else}
							<div class="file-preview">
								<span class="file-icon">{getFileIcon(asset.file_name)}</span>
								<span class="file-name">{asset.file_name}</span>
							</div>
						{/if}
						
						<div class="asset-info">
							<h5>{asset.file_name}</h5>
							<p>{formatFileSize(asset.file_size)}</p>
							{#if asset.width && asset.height}
								<p>{asset.width} × {asset.height}px</p>
							{/if}
							<span class="asset-status status-{asset.status}">{asset.status}</span>
						</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}
</div>

<style>
	.file-upload-container {
		width: 100%;
	}
	
	.file-input {
		display: none;
	}
	
	.upload-zone {
		border: 2px dashed var(--border-color);
		border-radius: var(--radius-lg);
		padding: var(--space-2xl);
		text-align: center;
		cursor: pointer;
		transition: all var(--transition-fast);
		background: var(--bg-glass);
		position: relative;
		min-height: 200px;
		display: flex;
		align-items: center;
		justify-content: center;
	}
	
	.upload-zone:hover:not(.disabled) {
		border-color: var(--primary);
		background: var(--bg-hover);
	}
	
	.upload-zone.drag-over {
		border-color: var(--primary);
		background: var(--primary-bg);
		transform: scale(1.02);
	}
	
	.upload-zone.disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}
	
	.upload-zone.uploading {
		pointer-events: none;
	}
	
	.upload-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-lg);
	}
	
	.upload-icon {
		width: 64px;
		height: 64px;
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		justify-content: center;
	}
	
	.upload-icon svg {
		width: 100%;
		height: 100%;
	}
	
	.loading-spinner {
		width: 32px;
		height: 32px;
		border: 3px solid var(--border-color);
		border-top: 3px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}
	
	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
	
	.upload-text h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}
	
	.upload-text p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
		line-height: 1.5;
	}
	
	.error-messages {
		margin-top: var(--space-md);
	}
	
	.error-message {
		background: var(--error-bg);
		color: var(--error-text);
		padding: var(--space-sm) var(--space-md);
		border-radius: var(--radius-md);
		margin-bottom: var(--space-sm);
		font-size: var(--text-sm);
	}
	
	.upload-progress {
		margin-top: var(--space-xl);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border-color);
	}
	
	.upload-progress h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.progress-item {
		margin-bottom: var(--space-lg);
		padding: var(--space-md);
		background: var(--bg-primary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border-color);
	}
	
	.progress-item:last-child {
		margin-bottom: 0;
	}
	
	.progress-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-sm);
	}
	
	.file-info {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		flex: 1;
	}
	
	.file-icon {
		font-size: var(--text-lg);
	}
	
	.file-name {
		font-weight: 600;
		color: var(--text-primary);
	}
	
	.file-size {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}
	
	.progress-actions {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}
	
	.progress-status {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	.status-pending {
		background: var(--bg-secondary);
		color: var(--text-secondary);
	}
	
	.status-uploading {
		background: var(--info-bg);
		color: var(--info-text);
	}
	
	.status-processing {
		background: var(--warning-bg);
		color: var(--warning-text);
	}
	
	.status-completed {
		background: var(--success-bg);
		color: var(--success-text);
	}
	
	.status-error {
		background: var(--error-bg);
		color: var(--error-text);
	}
	
	.remove-btn {
		width: 24px;
		height: 24px;
		border: none;
		background: var(--bg-glass);
		border-radius: var(--radius-sm);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		color: var(--text-secondary);
		transition: all var(--transition-fast);
	}
	
	.remove-btn:hover {
		background: var(--error-bg);
		color: var(--error-text);
	}
	
	.remove-btn svg {
		width: 16px;
		height: 16px;
	}
	
	.progress-bar {
		width: 100%;
		height: 8px;
		background: var(--bg-secondary);
		border-radius: var(--radius-full);
		overflow: hidden;
		margin-bottom: var(--space-sm);
	}
	
	.progress-fill {
		height: 100%;
		transition: width var(--transition-fast);
		border-radius: var(--radius-full);
	}
	
	.progress-fill.status-uploading {
		background: var(--info);
	}
	
	.progress-fill.status-processing {
		background: var(--warning);
	}
	
	.progress-fill.status-completed {
		background: var(--success);
	}
	
	.progress-fill.status-error {
		background: var(--error);
	}
	
	.progress-details {
		display: flex;
		justify-content: space-between;
		align-items: center;
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}
	
	.progress-error {
		margin-top: var(--space-sm);
		padding: var(--space-sm);
		background: var(--error-bg);
		color: var(--error-text);
		border-radius: var(--radius-sm);
		font-size: var(--text-sm);
	}
	
	.uploaded-assets {
		margin-top: var(--space-xl);
	}
	
	.uploaded-assets h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}
	
	.assets-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: var(--space-lg);
	}
	
	.asset-preview {
		background: var(--bg-glass);
		border: 1px solid var(--border-color);
		border-radius: var(--radius-lg);
		overflow: hidden;
		transition: all var(--transition-fast);
	}
	
	.asset-preview:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}
	
	.asset-preview img,
	.asset-preview video {
		width: 100%;
		height: 120px;
		object-fit: cover;
	}
	
	.file-preview {
		height: 120px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		background: var(--bg-secondary);
	}
	
	.file-preview .file-icon {
		font-size: 2rem;
	}
	
	.file-preview .file-name {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		text-align: center;
		padding: 0 var(--space-sm);
	}
	
	.asset-info {
		padding: var(--space-md);
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
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}
	
	@media (max-width: 768px) {
		.upload-zone {
			padding: var(--space-lg);
			min-height: 150px;
		}
		
		.upload-icon {
			width: 48px;
			height: 48px;
		}
		
		.upload-text h3 {
			font-size: var(--text-lg);
		}
		
		.progress-header {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-sm);
		}
		
		.assets-grid {
			grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		}
	}
</style> 