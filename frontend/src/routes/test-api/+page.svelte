<script lang="ts">
	import { onMount } from 'svelte';
	import { videoService } from '$lib/video';

	let apiResponse = '';
	let loading = false;
	let error = '';

	onMount(async () => {
		await testAPI();
	});

	async function testAPI() {
		loading = true;
		error = '';
		
		try {
			console.log('Testing videos API...');
			const response = await videoService.getVideos(1, 10);
			console.log('API Response:', response);
			apiResponse = JSON.stringify(response, null, 2);
		} catch (err) {
			console.error('API Error:', err);
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head>
	<title>API Test</title>
</svelte:head>

<div class="container">
	<h1>Videos API Test</h1>
	
	<button on:click={testAPI} disabled={loading}>
		{loading ? 'Testing...' : 'Test API'}
	</button>
	
	{#if error}
		<div class="error">
			<h3>Error:</h3>
			<pre>{error}</pre>
		</div>
	{/if}
	
	{#if apiResponse}
		<div class="response">
			<h3>API Response:</h3>
			<pre>{apiResponse}</pre>
		</div>
	{/if}
</div>

<style>
	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
	}
	
	.error {
		background: #ffebee;
		color: #c62828;
		padding: 1rem;
		border-radius: 8px;
		margin: 1rem 0;
	}
	
	.response {
		background: #e8f5e8;
		color: #2e7d32;
		padding: 1rem;
		border-radius: 8px;
		margin: 1rem 0;
	}
	
	pre {
		white-space: pre-wrap;
		word-wrap: break-word;
		font-family: monospace;
		font-size: 0.9rem;
	}
	
	button {
		background: #1976d2;
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
	}
	
	button:disabled {
		background: #ccc;
		cursor: not-allowed;
	}
</style> 