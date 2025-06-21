<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import type { AdvertiserFormData, FormErrors } from '$lib/types/advertising';
	
	let formData: AdvertiserFormData = {
		company_name: '',
		business_email: '',
		contact_name: '',
		contact_phone: '',
		business_address: '',
		tax_id: '',
		website: '',
		industry: ''
	};
	
	let errors: FormErrors = {};
	let loading = false;
	let submitted = false;

	onMount(() => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		// If user is already an advertiser, redirect to dashboard
		if ($auth.user?.role === 'advertiser') {
			goto('/advertiser');
			return;
		}
	});

	function validateForm(): boolean {
		errors = {};
		
		if (!formData.company_name.trim()) {
			errors.company_name = 'Company name is required';
		}
		
		if (!formData.business_email.trim()) {
			errors.business_email = 'Business email is required';
		} else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.business_email)) {
			errors.business_email = 'Please enter a valid email address';
		}
		
		if (!formData.contact_name.trim()) {
			errors.contact_name = 'Contact name is required';
		}
		
		if (formData.website && !/^https?:\/\/.+/.test(formData.website)) {
			errors.website = 'Please enter a valid website URL (including http:// or https://)';
		}
		
		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		loading = true;
		
		try {
			const response = await fetch('/api/v1/advertiser/account', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${$auth.token}`
				},
				body: JSON.stringify(formData)
			});

			const data = await response.json();

			if (data.success) {
				// Update user role locally to 'advertiser' since account was created
				// Note: In production, the role will be updated when admin approves the account
				// For development/testing, we'll update it immediately
				if ($auth.user) {
					auth.updateUser({ role: 'advertiser' });
				}
				
				submitted = true;
				// Redirect to advertiser dashboard after a short delay
				setTimeout(() => {
					goto('/advertiser');
				}, 3000);
			} else {
				errors.general = data.error || 'Failed to create advertiser account';
			}
		} catch (error) {
			errors.general = (error as Error).message;
		} finally {
			loading = false;
		}
	}

	function handleCancel() {
		goto('/advertiser');
	}
</script>

<svelte:head>
	<title>Advertiser Account Setup - BOME</title>
</svelte:head>

<div class="min-h-screen bg-gray-50 py-12">
	<div class="max-w-2xl mx-auto px-4 sm:px-6 lg:px-8">
		{#if submitted}
			<!-- Success state -->
			<div class="text-center">
				<div class="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100">
					<svg class="h-6 w-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
				</div>
				<h2 class="mt-6 text-3xl font-extrabold text-gray-900">Application Submitted!</h2>
				<p class="mt-2 text-sm text-gray-600">
					Your advertiser account application has been submitted for review. You'll receive an email notification once it's been processed.
				</p>
				<div class="mt-6">
					<a href="/advertiser" class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
						Go to Dashboard
					</a>
				</div>
			</div>
		{:else}
			<!-- Form -->
			<div class="bg-white shadow rounded-lg">
				<div class="px-4 py-5 sm:p-6">
					<div class="mb-8">
						<h1 class="text-2xl font-bold text-gray-900">Create Advertiser Account</h1>
						<p class="mt-1 text-sm text-gray-600">
							Fill out the information below to create your business advertising account. All applications are reviewed by our team.
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
						<!-- Company Information -->
						<div>
							<h3 class="text-lg font-medium text-gray-900 mb-4">Company Information</h3>
							<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
								<div class="sm:col-span-2">
									<label for="company_name" class="block text-sm font-medium text-gray-700">
										Company Name *
									</label>
									<div class="mt-1">
										<input
											type="text"
											id="company_name"
											bind:value={formData.company_name}
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
											class:border-red-300={errors.company_name}
											class:ring-red-500={errors.company_name}
											class:focus:ring-red-500={errors.company_name}
											class:focus:border-red-500={errors.company_name}
											required
										/>
										{#if errors.company_name}
											<p class="mt-2 text-sm text-red-600">{errors.company_name}</p>
										{/if}
									</div>
								</div>

								<div>
									<label for="business_email" class="block text-sm font-medium text-gray-700">
										Business Email *
									</label>
									<div class="mt-1">
										<input
											type="email"
											id="business_email"
											bind:value={formData.business_email}
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
											class:border-red-300={errors.business_email}
											class:ring-red-500={errors.business_email}
											class:focus:ring-red-500={errors.business_email}
											class:focus:border-red-500={errors.business_email}
											required
										/>
										{#if errors.business_email}
											<p class="mt-2 text-sm text-red-600">{errors.business_email}</p>
										{/if}
									</div>
								</div>

								<div>
									<label for="industry" class="block text-sm font-medium text-gray-700">
										Industry
									</label>
									<div class="mt-1">
										<select
											id="industry"
											bind:value={formData.industry}
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
										>
											<option value="">Select an industry</option>
											<option value="technology">Technology</option>
											<option value="healthcare">Healthcare</option>
											<option value="finance">Finance</option>
											<option value="retail">Retail</option>
											<option value="education">Education</option>
											<option value="entertainment">Entertainment</option>
											<option value="food-beverage">Food & Beverage</option>
											<option value="automotive">Automotive</option>
											<option value="real-estate">Real Estate</option>
											<option value="travel">Travel & Tourism</option>
											<option value="nonprofit">Non-Profit</option>
											<option value="other">Other</option>
										</select>
									</div>
								</div>

								<div>
									<label for="website" class="block text-sm font-medium text-gray-700">
										Website
									</label>
									<div class="mt-1">
										<input
											type="url"
											id="website"
											bind:value={formData.website}
											placeholder="https://example.com"
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
											class:border-red-300={errors.website}
											class:ring-red-500={errors.website}
											class:focus:ring-red-500={errors.website}
											class:focus:border-red-500={errors.website}
										/>
										{#if errors.website}
											<p class="mt-2 text-sm text-red-600">{errors.website}</p>
										{/if}
									</div>
								</div>

								<div>
									<label for="tax_id" class="block text-sm font-medium text-gray-700">
										Tax ID / EIN
									</label>
									<div class="mt-1">
										<input
											type="text"
											id="tax_id"
											bind:value={formData.tax_id}
											placeholder="XX-XXXXXXX"
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
										/>
									</div>
								</div>
							</div>
						</div>

						<!-- Contact Information -->
						<div>
							<h3 class="text-lg font-medium text-gray-900 mb-4">Contact Information</h3>
							<div class="grid grid-cols-1 gap-6 sm:grid-cols-2">
								<div>
									<label for="contact_name" class="block text-sm font-medium text-gray-700">
										Contact Name *
									</label>
									<div class="mt-1">
										<input
											type="text"
											id="contact_name"
											bind:value={formData.contact_name}
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
											class:border-red-300={errors.contact_name}
											class:ring-red-500={errors.contact_name}
											class:focus:ring-red-500={errors.contact_name}
											class:focus:border-red-500={errors.contact_name}
											required
										/>
										{#if errors.contact_name}
											<p class="mt-2 text-sm text-red-600">{errors.contact_name}</p>
										{/if}
									</div>
								</div>

								<div>
									<label for="contact_phone" class="block text-sm font-medium text-gray-700">
										Contact Phone
									</label>
									<div class="mt-1">
										<input
											type="tel"
											id="contact_phone"
											bind:value={formData.contact_phone}
											placeholder="(555) 123-4567"
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
										/>
									</div>
								</div>

								<div class="sm:col-span-2">
									<label for="business_address" class="block text-sm font-medium text-gray-700">
										Business Address
									</label>
									<div class="mt-1">
										<textarea
											id="business_address"
											rows="3"
											bind:value={formData.business_address}
											class="shadow-sm focus:ring-blue-500 focus:border-blue-500 block w-full sm:text-sm border-gray-300 rounded-md"
											placeholder="Street address, city, state, zip code"
										></textarea>
									</div>
								</div>
							</div>
						</div>

						<!-- Terms and Conditions -->
						<div class="bg-gray-50 rounded-lg p-4">
							<div class="flex">
								<div class="flex-shrink-0">
									<svg class="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
										<path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
									</svg>
								</div>
								<div class="ml-3">
									<h3 class="text-sm font-medium text-gray-800">Review Process</h3>
									<div class="mt-2 text-sm text-gray-700">
										<p>Your application will be reviewed by our team within 1-2 business days. You'll receive an email notification once your account is approved and you can start creating advertising campaigns.</p>
									</div>
								</div>
							</div>
						</div>

						<!-- Form Actions -->
						<div class="flex justify-end space-x-3">
							<button
								type="button"
								on:click={handleCancel}
								class="inline-flex items-center px-4 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
								disabled={loading}
							>
								Cancel
							</button>
							<button
								type="submit"
								class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
								disabled={loading}
							>
								{#if loading}
									<svg class="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
										<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
										<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
									</svg>
									Submitting...
								{:else}
									Submit Application
								{/if}
							</button>
						</div>
					</form>
				</div>
			</div>
		{/if}
	</div>
</div> 