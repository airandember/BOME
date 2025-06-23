<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(() => {
		// Redirect to unified dashboard with account tab
		goto('/dashboard?tab=account', { replaceState: true });
	});
</script>

<svelte:head>
	<title>Redirecting to Dashboard - BOME</title>
</svelte:head>

<!-- Simple loading state while redirecting -->
<div style="display: flex; align-items: center; justify-content: center; min-height: 50vh; flex-direction: column; gap: 1rem;">
	<div style="width: 40px; height: 40px; border: 3px solid #f3f3f3; border-top: 3px solid #007bff; border-radius: 50%; animation: spin 1s linear infinite;"></div>
	<p>Redirecting to your dashboard...</p>
</div>

<style>
	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}
</style> 