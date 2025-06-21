<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth } from '$lib/auth';
	import type { AdCampaign, FormErrors } from '$lib/types/advertising';
	
	let campaign: AdCampaign | null = null;
	let formData = {
		title: '',
		content: '',
		image_url: '',
		click_url: '',
		ad_type: 'banner',
		width: 728,
		height: 90,
		priority: 1
	};
	
	let errors: FormErrors = {};
	let loading = false;
	let submitted = false;
	let pageLoading = true;

	$: campaignId = parseInt($page.params.id);

	// Ad type presets
	const adTypePresets = {
		banner: { width: 728, height: 90 },
		large: { width: 300, height: 250 },
		small: { width: 160, height: 120 }
	};

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			await loadCampaign();
		} catch (err) {
			errors.general = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			pageLoading = false;
		}
	});

	async function loadCampaign() {
		const response = await fetch(`/api/v1/advertiser/campaigns/${campaignId}`, {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (response.ok) {
			const data = await response.json();
			campaign = data.data;
			
			// Check if campaign is approved
			if (campaign?.status !== 'approved') {
				throw new Error('Campaign must be approved before creating advertisements');
			}
		} else {
			throw new Error('Campaign not found or access denied');
		}
	}

	function handleAdTypeChange() {
		const preset = adTypePresets[formData.ad_type as keyof typeof adTypePresets];
		if (preset) {
			formData.width = preset.width;
			formData.height = preset.height;
		}
	}

	function validateForm(): boolean {
		errors = {};

		if (!formData.title.trim()) {
			errors.title = 'Title is required';
		} else if (formData.title.length > 255) {
			errors.title = 'Title must be less than 255 characters';
		}

		if (!formData.click_url.trim()) {
			errors.click_url = 'Click URL is required';
		} else if (!isValidUrl(formData.click_url)) {
			errors.click_url = 'Please enter a valid URL';
		}

		if (formData.image_url && !isValidUrl(formData.image_url)) {
			errors.image_url = 'Please enter a valid image URL';
		}

		if (formData.width < 1 || formData.width > 2000) {
			errors.width = 'Width must be between 1 and 2000 pixels';
		}

		if (formData.height < 1 || formData.height > 2000) {
			errors.height = 'Height must be between 1 and 2000 pixels';
		}

		if (formData.priority < 1 || formData.priority > 10) {
			errors.priority = 'Priority must be between 1 and 10';
		}

		return Object.keys(errors).length === 0;
	}

	function isValidUrl(string: string): boolean {
		try {
			new URL(string);
			return true;
		} catch (_) {
			return false;
		}
	}

	async function handleSubmit() {
		if (!validateForm()) return;

		loading = true;
		errors = {};

		try {
			const response = await fetch(`/api/v1/advertiser/campaigns/${campaignId}/ads`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify(formData)
			});

			if (response.ok) {
				submitted = true;
				setTimeout(() => {
					goto(`/advertiser/campaigns/${campaignId}`);
				}, 2000);
			} else {
				const data = await response.json();
				errors.general = data.error || 'Failed to create advertisement';
			}
		} catch (err) {
			errors.general = 'Network error occurred';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>Create Advertisement - {campaign?.name || 'Loading...'}</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-8">
	<div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
		{#if pageLoading}
			<div class="flex justify-center items-center h-64">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		{:else if submitted}
			<div class="text-center">
				<div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
					<svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<h1 class="text-2xl font-bold text-gray-900 mb-2">Advertisement Created!</h1>
				<p class="text-gray-600 mb-6">Your advertisement has been created successfully and is now active.</p>
				<div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto"></div>
				<p class="text-sm text-gray-500 mt-2">Redirecting to campaign...</p>
			</div>
		{:else}
			<!-- Header -->
			<div class="mb-8">
				<nav class="flex mb-4" aria-label="Breadcrumb">
					<ol class="flex items-center space-x-4">
						<li>
							<div>
								<a href="/advertiser" class="text-gray-400 hover:text-gray-500">
									<svg class="flex-shrink-0 h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
										<path d="M10.707 2.293a1 1 0 00-1.414 0l-7 7a1 1 0 001.414 1.414L4 10.414V17a1 1 0 001 1h2a1 1 0 001-1v-2a1 1 0 011-1h2a1 1 0 011 1v2a1 1 0 001 1h2a1 1 0 001-1v-6.586l.293.293a1 1 0 001.414-1.414l-7-7z" />
									</svg>
									<span class="sr-only">Dashboard</span>
								</a>
							</div>
						</li>
						<li>
							<div class="flex items-center">
								<svg class="flex-shrink-0 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
								</svg>
								<a href="/advertiser/campaigns" class="ml-4 text-sm font-medium text-gray-500 hover:text-gray-700">Campaigns</a>
							</div>
						</li>
						<li>
							<div class="flex items-center">
								<svg class="flex-shrink-0 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
								</svg>
								<a href="/advertiser/campaigns/{campaignId}" class="ml-4 text-sm font-medium text-gray-500 hover:text-gray-700">{campaign?.name}</a>
							</div>
						</li>
						<li>
							<div class="flex items-center">
								<svg class="flex-shrink-0 h-5 w-5 text-gray-400" viewBox="0 0 20 20" fill="currentColor">
									<path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
								</svg>
								<span class="ml-4 text-sm font-medium text-gray-500" aria-current="page">New Advertisement</span>
							</div>
						</li>
					</ol>
				</nav>

				<div class="md:flex md:items-center md:justify-between">
					<div class="flex-1 min-w-0">
						<h1 class="text-2xl font-bold leading-7 text-gray-900 sm:text-3xl sm:truncate">
							Create New Advertisement
						</h1>
						<p class="mt-1 text-sm text-gray-500">
							Add a new advertisement to your campaign: {campaign?.name}
						</p>
					</div>
				</div>
			</div>

			{#if errors.general}
				<div class="mb-6 bg-red-50 border border-red-200 rounded-lg p-4">
					<div class="flex">
						<div class="flex-shrink-0">
							<svg class="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
								<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
							</svg>
						</div>
						<div class="ml-3">
							<h3 class="text-sm font-medium text-red-800">Error</h3>
							<div class="mt-2 text-sm text-red-700">
								<p>{errors.general}</p>
							</div>
						</div>
					</div>
				</div>
			{/if}

			<!-- Form -->
			<div class="bg-white shadow rounded-lg">
				<form on:submit|preventDefault={handleSubmit} class="space-y-6 p-6">
					<!-- Advertisement Title -->
					<div>
						<label for="title" class="block text-sm font-medium text-gray-700">Advertisement Title *</label>
						<div class="mt-1">
							<input
								type="text"
								id="title"
								bind:value={formData.title}
								class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								placeholder="Enter a catchy title for your ad"
								required
							/>
						</div>
						{#if errors.title}
							<p class="mt-2 text-sm text-red-600">{errors.title}</p>
						{/if}
					</div>

					<!-- Advertisement Content -->
					<div>
						<label for="content" class="block text-sm font-medium text-gray-700">Advertisement Content</label>
						<div class="mt-1">
							<textarea
								id="content"
								rows="4"
								bind:value={formData.content}
								class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								placeholder="Enter the main text or description for your advertisement (optional)"
							></textarea>
						</div>
						{#if errors.content}
							<p class="mt-2 text-sm text-red-600">{errors.content}</p>
						{/if}
					</div>

					<!-- Ad Type -->
					<div>
						<label for="ad_type" class="block text-sm font-medium text-gray-700">Advertisement Type *</label>
						<div class="mt-1">
							<select
								id="ad_type"
								bind:value={formData.ad_type}
								on:change={handleAdTypeChange}
								class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								required
							>
								<option value="banner">Banner (728x90)</option>
								<option value="large">Large Rectangle (300x250)</option>
								<option value="small">Small Rectangle (160x120)</option>
							</select>
						</div>
						{#if errors.ad_type}
							<p class="mt-2 text-sm text-red-600">{errors.ad_type}</p>
						{/if}
					</div>

					<!-- Dimensions -->
					<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
						<div>
							<label for="width" class="block text-sm font-medium text-gray-700">Width (pixels) *</label>
							<div class="mt-1">
								<input
									type="number"
									id="width"
									bind:value={formData.width}
									min="1"
									max="2000"
									class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
									required
								/>
							</div>
							{#if errors.width}
								<p class="mt-2 text-sm text-red-600">{errors.width}</p>
							{/if}
						</div>

						<div>
							<label for="height" class="block text-sm font-medium text-gray-700">Height (pixels) *</label>
							<div class="mt-1">
								<input
									type="number"
									id="height"
									bind:value={formData.height}
									min="1"
									max="2000"
									class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
									required
								/>
							</div>
							{#if errors.height}
								<p class="mt-2 text-sm text-red-600">{errors.height}</p>
							{/if}
						</div>
					</div>

					<!-- Image URL -->
					<div>
						<label for="image_url" class="block text-sm font-medium text-gray-700">Image URL</label>
						<div class="mt-1">
							<input
								type="url"
								id="image_url"
								bind:value={formData.image_url}
								class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								placeholder="https://example.com/image.jpg"
							/>
						</div>
						<p class="mt-2 text-sm text-gray-500">Optional: URL to an image for your advertisement</p>
						{#if errors.image_url}
							<p class="mt-2 text-sm text-red-600">{errors.image_url}</p>
						{/if}
					</div>

					<!-- Click URL -->
					<div>
						<label for="click_url" class="block text-sm font-medium text-gray-700">Click URL *</label>
						<div class="mt-1">
							<input
								type="url"
								id="click_url"
								bind:value={formData.click_url}
								class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
								placeholder="https://your-website.com/landing-page"
								required
							/>
						</div>
						<p class="mt-2 text-sm text-gray-500">Where users will be taken when they click your advertisement</p>
						{#if errors.click_url}
							<p class="mt-2 text-sm text-red-600">{errors.click_url}</p>
						{/if}
					</div>

					<!-- Priority -->
					<div>
						<label for="priority" class="block text-sm font-medium text-gray-700">Priority</label>
						<div class="mt-1">
							<select
								id="priority"
								bind:value={formData.priority}
								class="block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm"
							>
								<option value={1}>1 - Lowest</option>
								<option value={2}>2</option>
								<option value={3}>3</option>
								<option value={4}>4</option>
								<option value={5}>5 - Normal</option>
								<option value={6}>6</option>
								<option value={7}>7</option>
								<option value={8}>8</option>
								<option value={9}>9</option>
								<option value={10}>10 - Highest</option>
							</select>
						</div>
						<p class="mt-2 text-sm text-gray-500">Higher priority ads are more likely to be shown</p>
						{#if errors.priority}
							<p class="mt-2 text-sm text-red-600">{errors.priority}</p>
						{/if}
					</div>

					<!-- Submit Button -->
					<div class="flex justify-end space-x-3">
						<button
							type="button"
							on:click={() => goto(`/advertiser/campaigns/${campaignId}`)}
							class="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
						>
							Cancel
						</button>
						<button
							type="submit"
							disabled={loading}
							class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{#if loading}
								<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
									<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
									<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
								</svg>
								Creating...
							{:else}
								Create Advertisement
							{/if}
						</button>
					</div>
				</form>
			</div>
		{/if}
	</div>
</div> 