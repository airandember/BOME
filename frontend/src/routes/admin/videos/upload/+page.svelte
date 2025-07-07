<script lang="ts">
	import { onMount } from 'svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { toastStore } from '$lib/stores/toast';
	import { isAdmin } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { videoService } from '$lib/video';

	let loading = false;
	let uploading = false;
	let categories: any[] = [];
	
	// Form data
	let title = '';
	let description = '';
	let category = '';
	let tags: string[] = [];
	let videoFile: File | null = null;
	let selectedTags: string[] = [];

	// Available tags
	const availableTags = [
		'Book of Mormon', 'Archaeology', 'Geography', 'DNA', 'Linguistics',
		'Historical', 'Cultural', 'Religious', 'Evidence', 'Research',
		'Ancient America', 'Mesoamerica', 'North America', 'South America'
	];

	onMount(async () => {
		// Check admin permissions
		if (!isAdmin()) {
			toastStore.error('Access denied. Admin privileges required.');
			goto('/videos');
			return;
		}

		await loadCategories();
	});

	async function loadCategories() {
		try {
			const response = await videoService.getCategories();
			categories = response.categories || [];
		} catch (err) {
			console.error('Error loading categories:', err);
		}
	}

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			videoFile = target.files[0];
		}
	}

	function toggleTag(tag: string) {
		if (selectedTags.includes(tag)) {
			selectedTags = selectedTags.filter(t => t !== tag);
		} else {
			selectedTags = [...selectedTags, tag];
		}
	}

	async function handleSubmit() {
		if (!videoFile) {
			toastStore.error('Please select a video file');
			return;
		}

		if (!title.trim()) {
			toastStore.error('Please enter a title');
			return;
		}

		try {
			uploading = true;

			const formData = new FormData();
			formData.append('video', videoFile);
			formData.append('title', title);
			formData.append('description', description);
			formData.append('category', category);
			formData.append('tags', JSON.stringify(selectedTags));

			const response = await fetch('/api/v1/videos/upload', {
				method: 'POST',
				body: formData,
				credentials: 'include'
			});

			const result = await response.json();

			if (response.ok) {
				toastStore.success('Video uploaded successfully!');
				goto('/videos');
			} else {
				toastStore.error(result.error || 'Upload failed');
			}
		} catch (err) {
			console.error('Upload error:', err);
			toastStore.error('Upload failed. Please try again.');
		} finally {
			uploading = false;
		}
	}
</script>

<svelte:head>
	<title>Upload Video - Admin</title>
	<meta name="description" content="Upload new videos to the platform." />
</svelte:head>

<Navigation />

<div class="page-wrapper">
	<main class="main-content-wrapper">
		{#if loading}
			<div class="loading-container">
				<LoadingSpinner size="large" />
				<p>Loading...</p>
			</div>
		{:else}
			<div class="upload-page">
				<div class="container">
					<header class="page-header">
						<h1>Upload Video</h1>
						<p>Add new videos to the platform</p>
					</header>

					<div class="upload-form-container">
						<form class="upload-form" on:submit|preventDefault={handleSubmit}>
							<div class="form-section">
								<h3>Video File</h3>
								<div class="file-upload">
									<input
										type="file"
										id="video-file"
										accept="video/*"
										on:change={handleFileSelect}
										required
									/>
									<label for="video-file" class="file-label">
										{#if videoFile}
											📹 {videoFile.name}
										{:else}
											�� Choose Video File
										{/if}
									</label>
								</div>
								<p class="file-help">Supported formats: MP4, AVI, MOV, WMV, FLV, WebM, MKV (Max 32MB)</p>
							</div>

							<div class="form-section">
								<h3>Video Information</h3>
								
								<div class="form-group">
									<label for="title">Title *</label>
									<input
										type="text"
										id="title"
										bind:value={title}
										placeholder="Enter video title"
										required
									/>
								</div>

								<div class="form-group">
									<label for="description">Description</label>
									<textarea
										id="description"
										bind:value={description}
										placeholder="Enter video description"
										rows="4"
									></textarea>
								</div>

								<div class="form-group">
									<label for="category">Category</label>
									<select id="category" bind:value={category}>
										<option value="">Select Category</option>
										{#each categories as cat}
											<option value={cat.name}>{cat.name}</option>
										{/each}
									</select>
								</div>
							</div>

							<div class="form-section">
								<h3>Tags</h3>
								<div class="tags-container">
									{#each availableTags as tag}
										<button
											type="button"
											class="tag-button {selectedTags.includes(tag) ? 'selected' : ''}"
											on:click={() => toggleTag(tag)}
										>
											{tag}
										</button>
									{/each}
								</div>
								<p class="tags-help">Click tags to select/deselect them</p>
							</div>

							<div class="form-actions">
								<button type="button" class="btn-secondary" on:click={() => goto('/videos')}>
									Cancel
								</button>
								<button type="submit" class="btn-primary" disabled={uploading}>
									{uploading ? 'Uploading...' : 'Upload Video'}
								</button>
							</div>
						</form>
					</div>
				</div>
			</div>
		{/if}
	</main>
	<Footer />
</div>

<style>
	.page-wrapper {
		display: flex;
		flex-direction: column;
		min-height: 100vh;
	}

	.main-content-wrapper {
		flex: 1 0 auto;
		width: 100%;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 50vh;
		gap: 1rem;
	}

	.upload-page {
		background: var(--bg-color);
		padding: 2rem 0;
		width: 100%;
	}

	.container {
		max-width: 800px;
		margin: 0 auto;
		padding: 0 2rem;
	}

	.page-header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.page-header h1 {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.page-header p {
		font-size: 1.1rem;
		color: var(--text-secondary);
	}

	.upload-form-container {
		background: var(--card-bg);
		padding: 2rem;
		border-radius: 20px;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.upload-form {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.form-section {
		border-bottom: 1px solid var(--border-color);
		padding-bottom: 2rem;
	}

	.form-section:last-child {
		border-bottom: none;
		padding-bottom: 0;
	}

	.form-section h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.form-group {
		margin-bottom: 1.5rem;
	}

	.form-group label {
		display: block;
		font-weight: 500;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.form-group input,
	.form-group textarea,
	.form-group select {
		width: 100%;
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
	}

	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px var(--accent-color);
	}

	.file-upload {
		position: relative;
		margin-bottom: 1rem;
	}

	.file-upload input[type="file"] {
		position: absolute;
		opacity: 0;
		width: 100%;
		height: 100%;
		cursor: pointer;
	}

	.file-label {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 1rem;
		border: 2px dashed var(--border-color);
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.file-label:hover {
		border-color: var(--accent-color);
		background: var(--accent-bg);
	}

	.file-help {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.tags-container {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.tag-button {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 20px;
		background: var(--input-bg);
		color: var(--text-primary);
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.tag-button:hover {
		background: var(--accent-bg);
		border-color: var(--accent-color);
	}

	.tag-button.selected {
		background: var(--accent-color);
		color: white;
		border-color: var(--accent-color);
	}

	.tags-help {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.form-actions {
		display: flex;
		gap: 1rem;
		justify-content: flex-end;
		padding-top: 2rem;
		border-top: 1px solid var(--border-color);
	}

	.btn-primary,
	.btn-secondary {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.btn-primary {
		background: var(--accent-color);
		color: white;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: var(--card-bg);
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover {
		background: var(--input-bg);
	}

	@media (max-width: 768px) {
		.container {
			padding: 0 1rem;
		}

		.upload-form-container {
			padding: 1.5rem;
		}

		.form-actions {
			flex-direction: column;
		}

		.tags-container {
			justify-content: center;
		}
	}
</style> 