<script lang="ts">
	import '../app.css';
	import { auth, initializeAuth } from '$lib/auth';
	import { onMount } from 'svelte';
	import ToastContainer from '$lib/components/ToastContainer.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
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
			<h1 class="brand-text">BOOK of MORMON</h1>
			<div class="loading-book">
				<LoadingSpinner size="large" color="white" />
				<p class="loading-text">Loading...</p>

			</div>
			<h1 class="brand-text">EVIDENCE</h1>
		</div>
	</div>
{/if}

<style>

    .brand-text {
		font-size: var(--text-xl);
		color: var(--text-primary);
		font-family: 'Playfair Display', Georgia, serif;
		letter-spacing: 0.2rem;
	}
	
	.app {
		min-height: 100vh;
		background: var(--bg-gray);
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

	.loading-book {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		margin: 0 auto var(--space-xl);
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.08);
		border-radius: var(--radius-xl);
		backdrop-filter: blur(10px);
		-webkit-backdrop-filter: blur(10px);
		width: fit-content;
	}

	/*.loading-content h1 {
		font-size: var(--text-4xl);
		font-weight: 800;
		margin-bottom: var(--space-md);
		font-family: var(--font-display);
	} */

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
