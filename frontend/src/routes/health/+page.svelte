<script>
	import { onMount } from 'svelte';
	
	let status = 'Checking...';
	let timestamp = '';
	let uptime = '';
	
	onMount(async () => {
		try {
			const response = await fetch('/health');
			const data = await response.json();
			status = data.status;
			timestamp = data.timestamp;
			uptime = new Date().toLocaleString();
		} catch (error) {
			status = 'Error';
			console.error('Health check failed:', error);
		}
	});
</script>

<svelte:head>
	<title>Health Check - BOME Frontend</title>
</svelte:head>

<div class="health-check">
	<h1>BOME Frontend Health Check</h1>
	
	<div class="status-grid">
		<div class="status-item">
			<strong>Status:</strong>
			<span class="status {status === 'healthy' ? 'healthy' : 'error'}">{status}</span>
		</div>
		
		<div class="status-item">
			<strong>Service:</strong>
			<span>BOME Frontend</span>
		</div>
		
		<div class="status-item">
			<strong>Version:</strong>
			<span>1.0.0</span>
		</div>
		
		<div class="status-item">
			<strong>Last Check:</strong>
			<span>{uptime}</span>
		</div>
		
		<div class="status-item">
			<strong>Server Time:</strong>
			<span>{timestamp}</span>
		</div>
	</div>
	
	<div class="actions">
		<button on:click={() => window.location.reload()}>Refresh</button>
		<a href="/" class="btn-home">Go to Homepage</a>
	</div>
</div>

<style>
	.health-check {
		max-width: 600px;
		margin: 2rem auto;
		padding: 2rem;
		background: white;
		border-radius: 8px;
		box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
	}
	
	h1 {
		text-align: center;
		color: #333;
		margin-bottom: 2rem;
	}
	
	.status-grid {
		display: grid;
		gap: 1rem;
		margin-bottom: 2rem;
	}
	
	.status-item {
		display: flex;
		justify-content: space-between;
		padding: 0.75rem;
		background: #f8f9fa;
		border-radius: 4px;
	}
	
	.status.healthy {
		color: #28a745;
		font-weight: bold;
	}
	
	.status.error {
		color: #dc3545;
		font-weight: bold;
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
		background: #007bff;
		color: white;
		text-decoration: none;
		cursor: pointer;
		font-size: 1rem;
	}
	
	button:hover, .btn-home:hover {
		background: #0056b3;
	}
</style>
