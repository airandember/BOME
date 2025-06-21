<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdvertiserPackage, AdAsset } from '$lib/types/advertising';

	export let selectedPackage: AdvertiserPackage;
	export let advertiserAccount: any;

	const dispatch = createEventDispatcher();

	// Campaign form data
	let campaignForm = {
		name: '',
		description: '',
		start_date: '',
		end_date: '',
		budget: selectedPackage.price,
		target_audience: '',
		billing_type: selectedPackage.billing_cycle === 'monthly' ? 'monthly' : 'weekly'
	};

	// Ad creation form
	let adForm = {
		title: '',
		content: '',
		click_url: '',
		ad_type: 'banner' as 'banner' | 'large' | 'small'
	};

	// Image upload state
	let uploadedImages: AdAsset[] = [];
	let uploadProgress: { [key: string]: number } = {};
	let dragOver = false;
	let formErrors: Record<string, string> = {};
	let submitting = false;

	// Ad specifications based on type
	const adSpecs = {
		banner: { width: 728, height: 90, maxSize: 2, description: 'Header/Footer Banner' },
		large: { width: 300, height: 250, maxSize: 2, description: 'Large Rectangle Sidebar' },
		small: { width: 300, height: 125, maxSize: 1, description: 'Small Rectangle Sidebar' }
	};

	// Set default dates
	const today = new Date();
	const nextWeek = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000);
	campaignForm.start_date = today.toISOString().split('T')[0];
	campaignForm.end_date = nextWeek.toISOString().split('T')[0];

	function validateCampaignForm() {
		formErrors = {};

		if (!campaignForm.name.trim()) {
			formErrors.name = 'Campaign name is required';
		}

		if (!campaignForm.description.trim()) {
			formErrors.description = 'Campaign description is required';
		}

		if (!campaignForm.start_date) {
			formErrors.start_date = 'Start date is required';
		}

		if (!campaignForm.end_date) {
			formErrors.end_date = 'End date is required';
		}

		if (campaignForm.start_date && campaignForm.end_date) {
			if (new Date(campaignForm.start_date) >= new Date(campaignForm.end_date)) {
				formErrors.end_date = 'End date must be after start date';
			}
		}

		if (campaignForm.budget < 50) {
			formErrors.budget = 'Minimum budget is $50';
		}

		if (!adForm.title.trim()) {
			formErrors.ad_title = 'Ad title is required';
		}

		if (!adForm.click_url.trim()) {
			formErrors.click_url = 'Click URL is required';
		} else if (!isValidUrl(adForm.click_url)) {
			formErrors.click_url = 'Please enter a valid URL';
		}

		if (uploadedImages.length === 0) {
			formErrors.images = 'At least one ad image is required';
		}

		return Object.keys(formErrors).length === 0;
	}

	function isValidUrl(string: string) {
		try {
			new URL(string);
			return true;
		} catch (_) {
			return false;
		}
	}

	function handleDragOver(event: DragEvent) {
		event.preventDefault();
		dragOver = true;
	}

	function handleDragLeave(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
	}

	function handleDrop(event: DragEvent) {
		event.preventDefault();
		dragOver = false;
		
		const files = Array.from(event.dataTransfer?.files || []);
		handleFiles(files);
	}

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		const files = Array.from(target.files || []);
		handleFiles(files);
	}

	async function handleFiles(files: File[]) {
		const specs = adSpecs[adForm.ad_type];
		
		for (const file of files) {
			// Validate file type
			if (!file.type.startsWith('image/')) {
				alert(`${file.name} is not an image file.`);
				continue;
			}

			// Validate file size
			if (file.size > specs.maxSize * 1024 * 1024) {
				alert(`${file.name} is too large. Maximum size is ${specs.maxSize}MB.`);
				continue;
			}

			// Validate image dimensions
			const isValidDimension = await validateImageDimensions(file, specs);
			if (!isValidDimension) {
				alert(`${file.name} does not meet the required dimensions (${specs.width}x${specs.height}px).`);
				continue;
			}

			// Upload file
			await uploadFile(file);
		}
	}

	function validateImageDimensions(file: File, specs: any): Promise<boolean> {
		return new Promise((resolve) => {
			const img = new Image();
			img.onload = () => {
				resolve(img.width === specs.width && img.height === specs.height);
			};
			img.onerror = () => resolve(false);
			img.src = URL.createObjectURL(file);
		});
	}

	async function uploadFile(file: File) {
		const fileId = Math.random().toString(36).substr(2, 9);
		uploadProgress[fileId] = 0;

		// Simulate upload progress
		const interval = setInterval(() => {
			uploadProgress[fileId] += Math.random() * 30;
			if (uploadProgress[fileId] >= 100) {
				uploadProgress[fileId] = 100;
				clearInterval(interval);
				
				// Create mock uploaded asset
				const asset: AdAsset = {
					id: Math.floor(Math.random() * 1000),
					campaign_id: 0, // Will be set when campaign is created
					asset_type: 'image',
					file_name: file.name,
					file_path: URL.createObjectURL(file),
					file_size: file.size,
					mime_type: file.type,
					width: adSpecs[adForm.ad_type].width,
					height: adSpecs[adForm.ad_type].height,
					alt_text: adForm.title,
					description: adForm.content,
					status: 'pending',
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};

				uploadedImages = [...uploadedImages, asset];
				delete uploadProgress[fileId];
			}
		}, 100);
	}

	function removeImage(index: number) {
		uploadedImages = uploadedImages.filter((_, i) => i !== index);
	}

	async function submitCampaign() {
		if (!validateCampaignForm()) return;

		submitting = true;
		try {
			// Mock API call to create campaign
			await new Promise(resolve => setTimeout(resolve, 2000));
			
			const campaignData = {
				...campaignForm,
				advertiser_id: advertiserAccount.id,
				status: 'pending',
				ad: {
					...adForm,
					assets: uploadedImages
				}
			};

			console.log('Campaign created:', campaignData);
			dispatch('campaignCreated', campaignData);
		} catch (error) {
			console.error('Failed to create campaign:', error);
		} finally {
			submitting = false;
		}
	}

	function goBack() {
		dispatch('goBack');
	}

	$: currentSpecs = adSpecs[adForm.ad_type];
</script>

<div class="campaign-creator">
	<div class="creator-header">
		<button class="back-btn" on:click={goBack}>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polyline points="15,18 9,12 15,6"></polyline>
			</svg>
			Back to Packages
		</button>
		<h2>Create Your First Campaign</h2>
		<p>Set up your advertising campaign with the <strong>{selectedPackage.name}</strong> package.</p>
	</div>

	<form on:submit|preventDefault={submitCampaign} class="campaign-form">
		<!-- Campaign Details -->
		<div class="form-section glass">
			<h3>Campaign Details</h3>
			
			<div class="form-row">
				<div class="form-group">
					<label for="campaign_name">Campaign Name *</label>
					<input
						type="text"
						id="campaign_name"
						bind:value={campaignForm.name}
						class:error={formErrors.name}
						placeholder="e.g., Spring Book Launch Campaign"
						required
					/>
					{#if formErrors.name}
						<span class="error-message">{formErrors.name}</span>
					{/if}
				</div>
			</div>

			<div class="form-group">
				<label for="campaign_description">Campaign Description *</label>
				<textarea
					id="campaign_description"
					bind:value={campaignForm.description}
					class:error={formErrors.description}
					placeholder="Describe your campaign goals and target audience"
					rows="4"
					required
				></textarea>
				{#if formErrors.description}
					<span class="error-message">{formErrors.description}</span>
				{/if}
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="start_date">Start Date *</label>
					<input
						type="date"
						id="start_date"
						bind:value={campaignForm.start_date}
						class:error={formErrors.start_date}
						required
					/>
					{#if formErrors.start_date}
						<span class="error-message">{formErrors.start_date}</span>
					{/if}
				</div>

				<div class="form-group">
					<label for="end_date">End Date *</label>
					<input
						type="date"
						id="end_date"
						bind:value={campaignForm.end_date}
						class:error={formErrors.end_date}
						required
					/>
					{#if formErrors.end_date}
						<span class="error-message">{formErrors.end_date}</span>
					{/if}
				</div>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="budget">Monthly Budget *</label>
					<div class="input-with-prefix">
						<span class="prefix">$</span>
						<input
							type="number"
							id="budget"
							bind:value={campaignForm.budget}
							class:error={formErrors.budget}
							min="50"
							step="10"
							required
						/>
					</div>
					{#if formErrors.budget}
						<span class="error-message">{formErrors.budget}</span>
					{/if}
				</div>

				<div class="form-group">
					<label for="billing_type">Billing Cycle</label>
					<select id="billing_type" bind:value={campaignForm.billing_type}>
						<option value="weekly">Weekly</option>
						<option value="monthly">Monthly</option>
					</select>
				</div>
			</div>

			<div class="form-group">
				<label for="target_audience">Target Audience</label>
				<textarea
					id="target_audience"
					bind:value={campaignForm.target_audience}
					placeholder="Describe your target audience (optional)"
					rows="3"
				></textarea>
			</div>
		</div>

		<!-- Ad Creation -->
		<div class="form-section glass">
			<h3>Create Your Advertisement</h3>
			
			<div class="form-group">
				<label for="ad_type">Ad Type *</label>
				<div class="ad-type-selector">
					{#each Object.entries(adSpecs) as [type, specs]}
						<label class="ad-type-option" class:selected={adForm.ad_type === type}>
							<input
								type="radio"
								bind:group={adForm.ad_type}
								value={type}
								on:change={() => uploadedImages = []}
							/>
							<div class="ad-type-info">
								<div class="ad-type-name">{specs.description}</div>
								<div class="ad-type-specs">{specs.width}x{specs.height}px</div>
								<div class="ad-type-size">Max: {specs.maxSize}MB</div>
							</div>
						</label>
					{/each}
				</div>
			</div>

			<div class="form-row">
				<div class="form-group">
					<label for="ad_title">Ad Title *</label>
					<input
						type="text"
						id="ad_title"
						bind:value={adForm.title}
						class:error={formErrors.ad_title}
						placeholder="Your ad title"
						required
					/>
					{#if formErrors.ad_title}
						<span class="error-message">{formErrors.ad_title}</span>
					{/if}
				</div>

				<div class="form-group">
					<label for="click_url">Click URL *</label>
					<input
						type="url"
						id="click_url"
						bind:value={adForm.click_url}
						class:error={formErrors.click_url}
						placeholder="https://yourwebsite.com"
						required
					/>
					{#if formErrors.click_url}
						<span class="error-message">{formErrors.click_url}</span>
					{/if}
				</div>
			</div>

			<div class="form-group">
				<label for="ad_content">Ad Description</label>
				<textarea
					id="ad_content"
					bind:value={adForm.content}
					placeholder="Brief description of your ad (optional)"
					rows="3"
				></textarea>
			</div>
		</div>

		<!-- Image Upload -->
		<div class="form-section glass">
			<h3>Upload Ad Images</h3>
			<p class="upload-info">
				Upload images for your <strong>{currentSpecs.description}</strong> 
				({currentSpecs.width}x{currentSpecs.height}px, max {currentSpecs.maxSize}MB)
			</p>

			<div 
				class="upload-area"
				class:drag-over={dragOver}
				on:dragover={handleDragOver}
				on:dragleave={handleDragLeave}
				on:drop={handleDrop}
			>
				<input
					type="file"
					id="image-upload"
					accept="image/*"
					multiple
					on:change={handleFileSelect}
					style="display: none;"
				/>
				
				<div class="upload-content">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="17,8 12,3 7,8"></polyline>
						<line x1="12" y1="3" x2="12" y2="15"></line>
					</svg>
					<h4>Drop images here or click to browse</h4>
					<p>Supports JPG, PNG, GIF, WebP formats</p>
					<p class="specs">Required: {currentSpecs.width}x{currentSpecs.height}px</p>
				</div>

				<button type="button" class="browse-btn" on:click={() => document.getElementById('image-upload')?.click()}>
					Browse Files
				</button>
			</div>

			{#if formErrors.images}
				<span class="error-message">{formErrors.images}</span>
			{/if}

			<!-- Upload Progress -->
			{#each Object.entries(uploadProgress) as [fileId, progress]}
				<div class="upload-progress">
					<div class="progress-bar">
						<div class="progress-fill" style="width: {progress}%"></div>
					</div>
					<span class="progress-text">{Math.round(progress)}%</span>
				</div>
			{/each}

			<!-- Uploaded Images -->
			{#if uploadedImages.length > 0}
				<div class="uploaded-images">
					<h4>Uploaded Images ({uploadedImages.length})</h4>
					<div class="image-grid">
						{#each uploadedImages as image, index}
							<div class="image-item">
								<img src={image.file_path} alt={image.alt_text} />
								<div class="image-info">
									<div class="image-name">{image.file_name}</div>
									<div class="image-size">{(image.file_size / 1024).toFixed(1)}KB</div>
								</div>
								<button type="button" class="remove-btn" on:click={() => removeImage(index)}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="18" y1="6" x2="6" y2="18"></line>
										<line x1="6" y1="6" x2="18" y2="18"></line>
									</svg>
								</button>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<!-- Package Summary -->
		<div class="package-summary glass">
			<h3>Package Summary</h3>
			<div class="summary-grid">
				<div class="summary-item">
					<span class="label">Package:</span>
					<span class="value">{selectedPackage.name}</span>
				</div>
				<div class="summary-item">
					<span class="label">Monthly Cost:</span>
					<span class="value">${selectedPackage.price}</span>
				</div>
				<div class="summary-item">
					<span class="label">Max Campaigns:</span>
					<span class="value">{selectedPackage.limits.max_campaigns}</span>
				</div>
				<div class="summary-item">
					<span class="label">Monthly Impressions:</span>
					<span class="value">{selectedPackage.limits.max_monthly_impressions.toLocaleString()}</span>
				</div>
			</div>
		</div>

		<!-- Submit Button -->
		<div class="form-actions">
			<button type="submit" class="submit-btn" disabled={submitting}>
				{#if submitting}
					<LoadingSpinner size="small" color="white" />
					Creating Campaign...
				{:else}
					Create Campaign & Submit for Approval
				{/if}
			</button>
		</div>
	</form>
</div>

<style>
	.campaign-creator {
		max-width: 800px;
		margin: 0 auto;
	}

	.creator-header {
		text-align: center;
		margin-bottom: var(--space-2xl);
	}

	.back-btn {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		margin-bottom: var(--space-lg);
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		transition: all var(--transition-normal);
	}

	.back-btn:hover {
		background: var(--bg-glass);
		color: var(--text-primary);
	}

	.back-btn svg {
		width: 16px;
		height: 16px;
	}

	.creator-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-md);
	}

	.creator-header p {
		color: var(--text-secondary);
		line-height: 1.6;
	}

	.campaign-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.form-section {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.form-section h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
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
	.form-group textarea.error,
	.form-group select.error {
		border-color: var(--error);
	}

	.error-message {
		color: var(--error);
		font-size: var(--text-xs);
	}

	.input-with-prefix {
		position: relative;
	}

	.prefix {
		position: absolute;
		left: var(--space-md);
		top: 50%;
		transform: translateY(-50%);
		color: var(--text-secondary);
		font-weight: 500;
	}

	.input-with-prefix input {
		padding-left: calc(var(--space-md) + 20px);
	}

	.ad-type-selector {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-md);
	}

	.ad-type-option {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg);
		border: 2px solid var(--border-color);
		border-radius: var(--radius-lg);
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.ad-type-option:hover {
		border-color: var(--primary);
	}

	.ad-type-option.selected {
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.1);
	}

	.ad-type-option input {
		margin: 0;
	}

	.ad-type-info {
		flex: 1;
	}

	.ad-type-name {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.ad-type-specs {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.ad-type-size {
		color: var(--text-secondary);
		font-size: var(--text-xs);
	}

	.upload-info {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
		line-height: 1.5;
	}

	.upload-area {
		border: 2px dashed var(--border-color);
		border-radius: var(--radius-lg);
		padding: var(--space-2xl);
		text-align: center;
		transition: all var(--transition-normal);
		position: relative;
	}

	.upload-area.drag-over {
		border-color: var(--primary);
		background: rgba(var(--primary-rgb), 0.05);
	}

	.upload-content svg {
		width: 48px;
		height: 48px;
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
	}

	.upload-content h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.upload-content p {
		color: var(--text-secondary);
		margin-bottom: var(--space-xs);
	}

	.specs {
		font-weight: 600;
		color: var(--primary);
	}

	.browse-btn {
		background: var(--primary);
		color: var(--white);
		border: none;
		padding: var(--space-md) var(--space-xl);
		border-radius: var(--radius-md);
		font-weight: 600;
		cursor: pointer;
		transition: all var(--transition-normal);
		margin-top: var(--space-lg);
	}

	.browse-btn:hover {
		background: var(--primary-dark);
		transform: translateY(-2px);
	}

	.upload-progress {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-md);
		margin-top: var(--space-md);
	}

	.progress-bar {
		flex: 1;
		height: 8px;
		background: var(--bg-glass-dark);
		border-radius: var(--radius-full);
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: var(--primary);
		transition: width var(--transition-normal);
	}

	.progress-text {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
	}

	.uploaded-images {
		margin-top: var(--space-xl);
	}

	.uploaded-images h4 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.image-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.image-item {
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		padding: var(--space-md);
		position: relative;
		border: 1px solid var(--border-color);
	}

	.image-item img {
		width: 100%;
		height: 120px;
		object-fit: cover;
		border-radius: var(--radius-md);
		margin-bottom: var(--space-md);
	}

	.image-info {
		text-align: center;
	}

	.image-name {
		font-size: var(--text-sm);
		font-weight: 500;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
		word-break: break-word;
	}

	.image-size {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.remove-btn {
		position: absolute;
		top: var(--space-sm);
		right: var(--space-sm);
		width: 24px;
		height: 24px;
		background: var(--error);
		color: var(--white);
		border: none;
		border-radius: 50%;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: all var(--transition-normal);
	}

	.remove-btn:hover {
		background: var(--error-dark);
		transform: scale(1.1);
	}

	.remove-btn svg {
		width: 12px;
		height: 12px;
	}

	.package-summary {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.package-summary h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-lg);
	}

	.summary-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.summary-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-md);
		border: 1px solid var(--border-color);
	}

	.label {
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	.value {
		color: var(--text-primary);
		font-weight: 600;
		font-size: var(--text-sm);
	}

	.form-actions {
		display: flex;
		justify-content: center;
		margin-top: var(--space-xl);
	}

	.submit-btn {
		background: linear-gradient(135deg, var(--primary) 0%, var(--secondary) 100%);
		color: var(--white);
		border: none;
		padding: var(--space-lg) var(--space-2xl);
		border-radius: var(--radius-md);
		font-weight: 600;
		font-size: var(--text-lg);
		cursor: pointer;
		transition: all var(--transition-normal);
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.submit-btn:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.submit-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	@media (max-width: 768px) {
		.form-row {
			grid-template-columns: 1fr;
		}

		.ad-type-selector {
			grid-template-columns: 1fr;
		}

		.summary-grid {
			grid-template-columns: 1fr;
		}

		.image-grid {
			grid-template-columns: 1fr;
		}
	}
</style> 