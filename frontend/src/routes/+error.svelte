<script lang="ts">
	import '../app.css'; // Import CSS variables directly for error pages
	import { page } from '$app/stores';
	
	// Type the error properly
	interface AppError extends Error {
		status?: number;
		stack?: string;
	}
</script>

<svelte:head>
	<title>Error - BOME Frontend</title>
</svelte:head>

<div class="error-page">
	<div class="error-container">
		<h1 class="error-title">Something went wrong</h1>
		<p class="error-message">
			{#if $page.error?.message}
				{$page.error.message}
			{:else}
				An unexpected error occurred. Please try refreshing the page.
			{/if}
		</p>
		
		<div class="error-details">
			{#if ($page.error as AppError)?.status}
				<p><strong>Status:</strong> {($page.error as AppError).status}</p>
			{/if}
			{#if ($page.error as AppError)?.stack}
				<details>
					<summary>Technical Details</summary>
					<pre class="error-stack">{($page.error as AppError).stack}</pre>
				</details>
			{/if}
		</div>
		
		<div class="actions">
			<button on:click={() => window.location.reload()}>Refresh Page</button>
			<a href="/" class="btn-home">Back to Home</a>
		</div>
	</div>
</div>

<style>


	.error-title {
		font-family: var(--sales-heading-font);
		color: var(--sales-gold);
		margin-bottom: 1rem;
		text-shadow: 1px 1px 4px rgba(0, 0, 0, 0.75);
	}

	.error-page {
		min-height: 100vh;
		display: flex;
		align-items: center;
		justify-content: center;
		background: linear-gradient(135deg, var(--primary-bom-light) 0%, var(--primary-bom-dark) 50%);
		padding: 1rem;
	}
	
	.error-container {
		max-width: 600px;
		text-align: center;
		background:  linear-gradient(135deg, var(--primary-bom-dark) 0%, var(--primary-bom-light) 70%);
		padding: 2rem;
		border-radius: 8px;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}
	
	h1 {
		font-family: var(--font-sans);
		color: white;
		margin-bottom: 1rem;
	}
	
	.error-message {
		color: var(--sales-gold);
		margin-bottom: 1.5rem;
		font-size: 1.1rem;
	}
	
	.error-details {
		margin-bottom: 2rem;
		text-align: left;
	}
	
	.error-stack {
		background: #f8f9fa;
		padding: 1rem;
		border-radius: 4px;
		font-size: 0.9rem;
		overflow-x: auto;
		margin-top: 0.5rem;
	}
	
	.actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
	}
	
	button, .btn-home {
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 4px;
		background: var(--primary-dark);
		color: white;
		text-decoration: none;
		cursor: pointer;
		font-size: 1rem;
		transition: all 0.3s ease;
	}
	
	button:hover, .btn-home:hover {
		background: var(--primary-bom-dark);
		transform: scale(1.15);
	}
</style>
