<script lang="ts">
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import { goto } from '$app/navigation';

	interface OAuth2Provider {
		name: string;
		display_name: string;
		icon: string;
		enabled: boolean;
	}

	let providers: OAuth2Provider[] = [];
	let loading = true;
	let error = '';

	onMount(async () => {
		await loadProviders();
	});

	async function loadProviders() {
		try {
			loading = true;
			const response = await apiRequest('/oauth2/providers');
			
			if (!response.ok) {
				throw new Error('Failed to load OAuth2 providers');
			}
			
			const data = await response.json();
			providers = data.providers || [];
		} catch (err: any) {
			error = err.message || 'Failed to load OAuth2 providers';
			console.error('Error loading OAuth2 providers:', err);
		} finally {
			loading = false;
		}
	}

	async function handleOAuth2Login(provider: string) {
		try {
			// Get current URL for return after OAuth2
			const returnURL = window.location.origin + '/auth/oauth2/callback';
			
			const response = await apiRequest('/oauth2/login', {
				method: 'POST',
				body: JSON.stringify({
					provider: provider,
					return_url: returnURL
				})
			});
			
			if (!response.ok) {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Failed to initiate OAuth2 login');
			}
			
			const data = await response.json();
			
			// Redirect to OAuth2 provider
			window.location.href = data.auth_url;
		} catch (err: any) {
			showToast(err.message || 'OAuth2 login failed', 'error');
			console.error('OAuth2 login error:', err);
		}
	}

	function getProviderIcon(provider: OAuth2Provider): string {
		switch (provider.name) {
			case 'google':
				return '📨';
			case 'facebook':
				return '📘';
			case 'github':
				return '🐙';
			case 'microsoft':
				return '🪟';
			default:
				return provider.icon || '🔗';
		}
	}

	function getProviderColor(provider: OAuth2Provider): string {
		switch (provider.name) {
			case 'google':
				return 'bg-red-500 hover:bg-red-600';
			case 'facebook':
				return 'bg-blue-600 hover:bg-blue-700';
			case 'github':
				return 'bg-gray-800 hover:bg-gray-900';
			case 'microsoft':
				return 'bg-blue-500 hover:bg-blue-600';
			default:
				return 'bg-gray-600 hover:bg-gray-700';
		}
	}
</script>

<div class="oauth2-login">
	{#if loading}
		<div class="loading-container">
			<div class="loading-spinner"></div>
			<p>Loading login options...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary btn-sm" onclick={loadProviders}>
				Try Again
			</button>
		</div>
	{:else if providers.length > 0}
		<div class="oauth2-providers">
			
			
			<div class="providers-grid">
				{#each providers.filter(p => p.enabled) as provider}
					<button 
						class="oauth2-btn {getProviderColor(provider)}"
						onclick={() => handleOAuth2Login(provider.name)}
					>
					<svg viewBox="0 0 32 32" fill="none" xmlns="http://www.w3.org/2000/svg" style="width: 60px; height: 60px;">
						<g id="SVGRepo_bgCarrier" stroke-width="0"></g><g id="SVGRepo_tracerCarrier" stroke-linecap="round" stroke-linejoin="round"></g>
						<g id="SVGRepo_iconCarrier"> <path d="M2 11.9556C2 8.47078 2 6.7284 2.67818 5.39739C3.27473 4.22661 4.22661 3.27473 5.39739 2.67818C6.7284 2 8.47078 2 11.9556 2H20.0444C23.5292 2 25.2716 2 26.6026 2.67818C27.7734 3.27473 28.7253 4.22661 29.3218 5.39739C30 6.7284 30 8.47078 30 11.9556V20.0444C30 23.5292 30 25.2716 29.3218 26.6026C28.7253 27.7734 27.7734 28.7253 26.6026 29.3218C25.2716 30 23.5292 30 20.0444 30H11.9556C8.47078 30 6.7284 30 5.39739 29.3218C4.22661 28.7253 3.27473 27.7734 2.67818 26.6026C2 25.2716 2 23.5292 2 20.0444V11.9556Z" fill="white"></path>
							<path d="M22.0515 8.52295L16.0644 13.1954L9.94043 8.52295V8.52421L9.94783 8.53053V15.0732L15.9954 19.8466L22.0515 15.2575V8.52295Z" fill="#EA4335"></path>
							<path d="M23.6231 7.38639L22.0508 8.52292V15.2575L26.9983 11.459V9.17074C26.9983 9.17074 26.3978 5.90258 23.6231 7.38639Z" fill="#FBBC05"></path>
							<path d="M22.0508 15.2575V23.9924H25.8428C25.8428 23.9924 26.9219 23.8813 26.9995 22.6513V11.459L22.0508 15.2575Z" fill="#34A853"></path>
							<path d="M9.94811 24.0001V15.0732L9.94043 15.0669L9.94811 24.0001Z" fill="#C5221F"></path>
							<path d="M9.94014 8.52404L8.37646 7.39382C5.60179 5.91001 5 9.17692 5 9.17692V11.4651L9.94014 15.0667V8.52404Z" fill="#C5221F"></path>
							<path d="M9.94043 8.52441V15.0671L9.94811 15.0734V8.53073L9.94043 8.52441Z" fill="#C5221F"></path>
							<path d="M5 11.4668V22.6591C5.07646 23.8904 6.15673 24.0003 6.15673 24.0003H9.94877L9.94014 15.0671L5 11.4668Z" fill="#4285F4"></path>
						</g>
					</svg>
						<!--<span class="provider-icon">{getProviderIcon(provider)}</span>-->
						<span class="provider-name">Sign in with {provider.display_name}</span>
						
					</button>
				{/each}
				<div class="divider">
					<span>Or continue with</span>
				</div>
			</div>
			
			{#if providers.filter(p => !p.enabled).length > 0}
				<div class="disabled-providers">
					<p class="disabled-text">Available soon:</p>
					<div class="disabled-grid">
						{#each providers.filter(p => !p.enabled) as provider}
							<div class="disabled-provider">
								<span class="provider-icon">{getProviderIcon(provider)}</span>
								<span class="provider-name">{provider.display_name}</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>
	{:else}
		<div class="no-providers">
			<p>No OAuth2 providers configured</p>
		</div>
	{/if}
</div>

<style>
	.oauth2-login {
		width: 100%;
	}

	.loading-container,
	.error-container,
	.no-providers {
		text-align: center;
		padding: 1rem 0;
	}

	.error-message {
		color: var(--error-text);
		margin-bottom: 1rem;
		font-size: 0.875rem;
	}

	.loading-spinner {
		width: 24px;
		height: 24px;
		border: 2px solid var(--border-color);
		border-top: 2px solid var(--primary-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto 0.5rem auto;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.divider {
		position: relative;
		text-align: center;
		margin: 1.5rem 0;
	}

	.divider::before {
		content: '';
		position: absolute;
		top: 50%;
		left: 0;
		right: 0;
		height: 1px;
		background: var(--border-color);
	}

	.divider span {
		background: var(--card-bg);
		padding: 0 1rem;
		color: var(--text-secondary);
		font-size: 0.875rem;
	}

	.providers-grid {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
	}

	.oauth2-btn {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 1rem;
		border: none;
		border-radius: 0.5rem;
		color: white;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		font-size: 0.875rem;
		width: 100%;
		justify-content: center;
	}

	.oauth2-btn:hover {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
	}

	.provider-icon {
		font-size: 1.25rem;
	}

	.provider-name {
		flex: 1;
		text-align: center;
	}

	.disabled-providers {
		margin-top: 1.5rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border-color);
	}

	.disabled-text {
		font-size: 0.75rem;
		color: var(--text-secondary);
		margin-bottom: 0.75rem;
		text-align: center;
	}

	.disabled-grid {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		justify-content: center;
	}

	.disabled-provider {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 0.75rem;
		background: var(--bg-secondary);
		border-radius: 0.375rem;
		opacity: 0.6;
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.disabled-provider .provider-icon {
		font-size: 1rem;
	}

	.btn {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
	}

	.btn-primary {
		background: var(--primary-color);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-color-dark);
		transform: translateY(-1px);
	}

	.btn-sm {
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
	}

	/* Provider-specific colors */
	.bg-red-500 { background-color: #ef4444; }
	.bg-red-600 { background-color: #dc2626; }
	.hover\:bg-red-600:hover { background-color: #dc2626; }

	.bg-blue-600 { background-color: #2563eb; }
	.bg-blue-700 { background-color: #1d4ed8; }
	.hover\:bg-blue-700:hover { background-color: #1d4ed8; }

	.bg-blue-500 { background-color: #3b82f6; }
	.hover\:bg-blue-600:hover { background-color: #2563eb; }

	.bg-gray-800 { background-color: #1f2937; }
	.bg-gray-900 { background-color: #111827; }
	.hover\:bg-gray-900:hover { background-color: #111827; }

	.bg-gray-600 { background-color: #4b5563; }
	.bg-gray-700 { background-color: #374151; }
	.hover\:bg-gray-700:hover { background-color: #374151; }

	@media (max-width: 640px) {
		.providers-grid {
			gap: 0.5rem;
		}
		
		.oauth2-btn {
			padding: 0.625rem 0.875rem;
			font-size: 0.8125rem;
		}
	}
</style>
