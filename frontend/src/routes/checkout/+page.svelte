<script lang="ts">
import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let portalUrl: string = '';
	let loading = true;
	let error = '';
let iframeBlocked = false;
let iframeEl: HTMLIFrameElement | null = null;
let redirecting = false;

	async function fetchPortalLink() {
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link');
			if (!res.ok) {
				throw new Error(`Failed to load portal link (${res.status})`);
			}
			const data = await res.json();
			// Be flexible about response shape (backend returns { portal_url })
			portalUrl = data?.portal_url || data?.link || data?.portal_link || data?.portalLink || data?.data?.link || '';
			if (!portalUrl) throw new Error('Stripe portal link not configured');
		} catch (e: any) {
			error = e?.message || 'Failed to load portal link';
		} finally {
			loading = false;
		}
	}

	function handleIframeError() {
		// Many hosted billing pages send X-Frame-Options that block embedding
		iframeBlocked = true;
	}

	function openPortalInNewTab() {
		if (portalUrl) window.open(portalUrl, '_blank');
	}

	onMount(async () => {
		await fetchPortalLink();
		// Stripe portal cannot be embedded (frame-ancestors 'none'); redirect instead
		if (portalUrl) {
			redirecting = true;
			try {
				window.location.assign(portalUrl);
			} catch {}
		}
	});
</script>

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" />
		<p>Loading checkout...</p>
	</div>
{:else if error}
	<div class="error-container">
		<p class="error-text">{error}</p>
	</div>
{:else}
	<div class="checkout-container">
		<h1 class="page-title">Checkout</h1>
		{#if portalUrl}
			<div class="fallback">
				<p>{redirecting ? 'Redirecting to billing portal…' : 'Open the billing portal:'}</p>
				<button class="open-btn" on:click={openPortalInNewTab}>Open Billing Portal</button>
				<p class="direct-link"><a href={portalUrl} target="_blank" rel="noopener noreferrer">{portalUrl}</a></p>
			</div>
		{/if}
	</div>
{/if}

<style>
	.checkout-container {
		min-height: 100vh;
		padding: 2rem;
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.page-title {
		margin: 0 0 0.5rem 0;
		font-weight: 700;
		font-size: 1.5rem;
	}

	.portal-frame-wrapper {
		position: relative;
		width: 100%;
		height: 80vh;
		border: 1px solid rgba(255,255,255,0.15);
		border-radius: 12px;
		overflow: hidden;
		background: var(--bg-glass, rgba(255,255,255,0.05));
	}

	.portal-frame {
		width: 100%;
		height: 100%;
		border: none;
	}

	.fallback {
		margin-top: 1rem;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.open-btn {
		align-self: start;
		background: var(--primary);
		color: #fff;
		border: none;
		border-radius: 8px;
		padding: 0.5rem 1rem;
		cursor: pointer;
	}

	.loading-container, .error-container {
		min-height: 60vh;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
	}

	.error-text { color: #ef4444; }

	.direct-link a { color: var(--primary); word-break: break-all; }
</style>

