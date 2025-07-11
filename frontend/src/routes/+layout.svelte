<script lang="ts">
	import '../app.css';
	import { auth, initializeAuth } from '$lib/auth';
	import { onMount } from 'svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	import { initializeSecurity } from '$lib/utils/security';
	import { authStore } from '$lib/stores/api';

	let mounted = false;

	onMount(async () => {
		// Initialize auth system before mounting components
		await initializeAuth();
		
		// Initialize other services
		initializeSecurity();
		authStore.init();
		
		// Add debug functions to window for debugging
		if (typeof window !== 'undefined') {
			(window as any).debugAuth = () => {
				const { debugTokenStorage, clearAllAuthStorage } = require('$lib/auth');
				return debugTokenStorage();
			};
			(window as any).clearAuth = () => {
				const { clearAllAuthStorage } = require('$lib/auth');
				clearAllAuthStorage();
			};
			console.log('🔧 Debug commands available:');
			console.log('  - window.debugAuth() - Check token storage state');
			console.log('  - window.clearAuth() - Clear all auth storage');
		}
		
		mounted = true;
	});
</script>

<svelte:head>
	<title>BOME - Book of Mormon Evidences</title>
	<meta name="description" content="Discover compelling evidence for the Book of Mormon through our modern streaming platform" />
	<meta name="viewport" content="width=device-width, initial-scale=1" />
	<link rel="icon" href="/favicon.ico" />
	
	<!-- Google Fonts -->
	<link rel="preconnect" href="https://fonts.googleapis.com">
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="">
	<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap" rel="stylesheet">
</svelte:head>

{#if mounted}
	<div class="app">
		<slot />
		<ToastContainer />
	</div>
{:else}
	<div class="loading-screen">
		<div class="loading-content">
			<div class="brand-logo">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
					<path d="M2 17l10 5 10-5"></path>
					<path d="M2 12l10 5 10-5"></path>
				</svg>
			</div>
			<h1>BOME</h1>
			<p>Loading...</p>
		</div>
	</div>
{/if}

<style>
	.app {
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
		transition: all var(--transition-normal);
	}

	.loading-screen {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--primary-gradient);
		color: var(--white);
	}

	.loading-content {
		text-align: center;
		animation: fadeIn 0.6s ease-out;
	}

	.brand-logo {
		width: 80px;
		height: 80px;
		background: rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-lg);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
	}

	.brand-logo svg {
		width: 40px;
		height: 40px;
		color: var(--white);
	}

	.loading-content h1 {
		font-size: var(--text-4xl);
		font-weight: 800;
		margin-bottom: var(--space-md);
		font-family: var(--font-display);
	}

	.loading-content p {
		font-size: var(--text-lg);
		opacity: 0.8;
		margin: 0;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}
</style> 