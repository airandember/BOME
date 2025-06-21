<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdvertiserAccount, FormErrors } from '$lib/types/advertising';
	
	let advertiserAccount: AdvertiserAccount | null = null;
	let formData = {
		name: '',
		description: '',
		budget: 100,
		start_date: '',
		end_date: '',
		target_audience: '',
		billing_type: 'weekly',
		billing_rate: 50
	};
	
	let errors: FormErrors = {};
	let loading = false;
	let submitted = false;
	let pageLoading = true;

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			await loadAdvertiserAccount();
		} catch (err) {
			console.error('Error loading advertiser account:', err);
			goto('/advertiser');
		} finally {
			pageLoading = false;
		}
	});

	async function loadAdvertiserAccount() {
		const response = await fetch('/api/v1/advertiser/account', {
			headers: {
				'Authorization': `Bearer ${$auth.token}`
			}
		});

		if (!response.ok) {
			throw new Error('No advertiser account found');
		}

		const data = await response.json();
		advertiserAccount = data.data;

		if (advertiserAccount?.status !== 'approved') {
			goto('/advertiser');
		}

		// Set default start date to today
		const today = new Date();
		formData.start_date = today.toISOString().split('T')[0];
	}

	function validateForm(): boolean {
		errors = {};
		
		if (!formData.name.trim()) {
			errors.name = 'Campaign name is required';
		}
		
		if (formData.budget < 10) {
			errors.budget = 'Minimum budget is $10';
		}
		
		if (formData.billing_rate < 1) {
			errors.billing_rate = 'Minimum billing rate is $1';
		}
		
		if (!formData.start_date) {
			errors.start_date = 'Start date is required';
		}
		
		if (formData.end_date && formData.start_date && new Date(formData.end_date) <= new Date(formData.start_date)) {
			errors.end_date = 'End date must be after start date';
		}
		
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		loading = true;
		
		try {
			const response = await fetch('/api/v1/advertiser/campaigns', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify(formData)
			});

			const data = await response.json();

			if (response.ok) {
				submitted = true;
				// Redirect to campaigns list after a short delay
				setTimeout(() => {
					goto('/advertiser/campaigns');
				}, 2000);
			} else {
				if (data.error) {
					if (typeof data.error === 'object') {
						errors = data.error;
					} else {
						errors.general = data.error;
					}
				} else {
					errors.general = 'Failed to create campaign. Please try again.';
				}
			}
		} catch (error) {
			errors.general = 'Network error. Please check your connection and try again.';
		} finally {
			loading = false;
		}
	}

	function handleCancel() {
		goto('/advertiser/campaigns');
	}

	function updateBillingRate() {
		// Update billing rate based on budget and billing type
		if (formData.billing_type === 'weekly') {
			formData.billing_rate = Math.max(1, formData.budget / 4);
		} else if (formData.billing_type === 'monthly') {
			formData.billing_rate = Math.max(1, formData.budget);
		}
	}
</script>

<svelte:head>
	<title>Create New Campaign - BOME Advertiser</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-12">
	<div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8">
		{#if pageLoading}
			<div class="flex justify-center items-center h-64">
				<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
			</div>
		{:else if submitted}
			<!-- Success state -->
			<div class="text-center">
				<div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
					<svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<h2 class="mt-6 text-3xl font-extrabold text-gray-900">Campaign Created!</h2>
				<p class="mt-2 text-sm text-gray-600">
					Your campaign has been submitted for review. You'll receive a notification once it's been approved.
				</p>
				<div class="mt-6">
					<a href="/advertiser/campaigns" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
						View Campaigns
					</a>
				</div>
			</div>
		{:else}
			<!-- Form -->
			<div class="bg-white shadow rounded-lg">
				<div class="px-4 py-5 sm:p-6">
					<div class="mb-8">
						<h1 class="text-2xl font-bold text-gray-900">Create New Campaign</h1>
						<p class="mt-1 text-sm text-gray-600">
							Set up your advertising campaign details. All campaigns require admin approval before going live.
						</p>
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

					<form on:submit|preventDefault={handleSubmit} class="space-y-6">
						<!-- Campaign Name -->
						<div>
							<label for="name" class="block text-sm font-medium text-gray-700">
								Campaign Name *
							</label>
							<div class="mt-1">
								<input
									type="text"
									id="name"
									bind:value={formData.name}
									class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
									class:border-red-300={errors.name}
									class:ring-red-500={errors.name}
									class:focus:ring-red-500={errors.name}
									class:focus:border-red-500={errors.name}
									placeholder="Enter campaign name"
									required
								/>
								{#if errors.name}
									<p class="mt-2 text-sm text-red-600">{errors.name}</p>
								{/if}
							</div>
						</div>

						<!-- Description -->
						<div>
							<label for="description" class="block text-sm font-medium text-gray-700">
								Description
							</label>
							<div class="mt-1">
								<textarea
									id="description"
									rows="3"
									bind:value={formData.description}
									class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
									placeholder="Describe your campaign objectives and target audience"
								></textarea>
								{#if errors.description}
									<p class="mt-2 text-sm text-red-600">{errors.description}</p>
								{/if}
							</div>
						</div>

						<!-- Target Audience -->
						<div>
							<label for="target_audience" class="block text-sm font-medium text-gray-700">
								Target Audience
							</label>
							<div class="mt-1">
								<textarea
									id="target_audience"
									rows="2"
									bind:value={formData.target_audience}
									class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
									placeholder="Describe your target audience demographics and interests"
								></textarea>
								{#if errors.target_audience}
									<p class="mt-2 text-sm text-red-600">{errors.target_audience}</p>
								{/if}
							</div>
						</div>

						<!-- Budget and Billing -->
						<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
							<div>
								<label for="budget" class="block text-sm font-medium text-gray-700">
									Total Budget (USD) *
								</label>
								<div class="mt-1 relative rounded-md shadow-sm">
									<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
										<span class="text-gray-500 sm:text-sm">$</span>
									</div>
									<input
										type="number"
										id="budget"
										min="10"
										step="0.01"
										bind:value={formData.budget}
										on:input={updateBillingRate}
										class="focus:ring-blue-500 focus:border-blue-500 block w-full pl-7 pr-12 sm:text-sm border-gray-300 rounded-md"
										class:border-red-300={errors.budget}
										class:ring-red-500={errors.budget}
										class:focus:ring-red-500={errors.budget}
										class:focus:border-red-500={errors.budget}
										placeholder="100.00"
										required
									/>
									<div class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
										<span class="text-gray-500 sm:text-sm">USD</span>
									</div>
								</div>
								{#if errors.budget}
									<p class="mt-2 text-sm text-red-600">{errors.budget}</p>
								{:else}
									<p class="mt-2 text-sm text-gray-500">Minimum budget is $10.00</p>
								{/if}
							</div>

							<div>
								<label for="billing_type" class="block text-sm font-medium text-gray-700">
									Billing Type *
								</label>
								<div class="mt-1">
									<select
										id="billing_type"
										bind:value={formData.billing_type}
										on:change={updateBillingRate}
										class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
										required
									>
										<option value="weekly">Weekly</option>
										<option value="monthly">Monthly</option>
									</select>
								</div>
								<p class="mt-2 text-sm text-gray-500">How often you'll be charged</p>
							</div>
						</div>

						<!-- Billing Rate -->
						<div>
							<label for="billing_rate" class="block text-sm font-medium text-gray-700">
								{formData.billing_type === 'weekly' ? 'Weekly' : 'Monthly'} Rate (USD) *
							</label>
							<div class="mt-1 relative rounded-md shadow-sm">
								<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
									<span class="text-gray-500 sm:text-sm">$</span>
								</div>
								<input
									type="number"
									id="billing_rate"
									min="1"
									step="0.01"
									bind:value={formData.billing_rate}
									class="focus:ring-blue-500 focus:border-blue-500 block w-full pl-7 pr-12 sm:text-sm border-gray-300 rounded-md"
									class:border-red-300={errors.billing_rate}
									class:ring-red-500={errors.billing_rate}
									class:focus:ring-red-500={errors.billing_rate}
									class:focus:border-red-500={errors.billing_rate}
									placeholder="50.00"
									required
								/>
								<div class="absolute inset-y-0 right-0 pr-3 flex items-center pointer-events-none">
									<span class="text-gray-500 sm:text-sm">USD</span>
								</div>
							</div>
							{#if errors.billing_rate}
								<p class="mt-2 text-sm text-red-600">{errors.billing_rate}</p>
							{:else}
								<p class="mt-2 text-sm text-gray-500">Amount charged per {formData.billing_type === 'weekly' ? 'week' : 'month'}</p>
							{/if}
						</div>

						<!-- Date Range -->
						<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
							<div>
								<label for="start_date" class="block text-sm font-medium text-gray-700">
									Start Date *
								</label>
								<div class="mt-1">
									<input
										type="date"
										id="start_date"
										bind:value={formData.start_date}
										min={new Date().toISOString().split('T')[0]}
										class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
										class:border-red-300={errors.start_date}
										class:ring-red-500={errors.start_date}
										class:focus:ring-red-500={errors.start_date}
										class:focus:border-red-500={errors.start_date}
										required
									/>
									{#if errors.start_date}
										<p class="mt-2 text-sm text-red-600">{errors.start_date}</p>
									{/if}
								</div>
							</div>

							<div>
								<label for="end_date" class="block text-sm font-medium text-gray-700">
									End Date (Optional)
								</label>
								<div class="mt-1">
									<input
										type="date"
										id="end_date"
										bind:value={formData.end_date}
										min={formData.start_date || new Date().toISOString().split('T')[0]}
										class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
										class:border-red-300={errors.end_date}
										class:ring-red-500={errors.end_date}
										class:focus:ring-red-500={errors.end_date}
										class:focus:border-red-500={errors.end_date}
									/>
									{#if errors.end_date}
										<p class="mt-2 text-sm text-red-600">{errors.end_date}</p>
									{:else}
										<p class="mt-2 text-sm text-gray-500">Leave blank for ongoing campaign</p>
									{/if}
								</div>
							</div>
						</div>

						<!-- Form Actions -->
						<div class="flex justify-end space-x-3 pt-6">
							<button
								type="button"
								on:click={handleCancel}
								class="bg-white py-2 px-4 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
							>
								Cancel
							</button>
							<button
								type="submit"
								disabled={loading}
								class="inline-flex justify-center py-2 px-4 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								{#if loading}
									<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
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
	</div>
</div> 