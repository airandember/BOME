<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	let isLoading = true;
	let events = [];
	let loadingEvents = true;

	onMount(() => {
		console.log('Events page mounted');
		loadEvents();
	});

	async function loadEvents() {
		try {
			loadingEvents = true;
			// TODO: Implement events loading
			await new Promise(resolve => setTimeout(resolve, 1000)); // Simulate API call
			events = [];
		} catch (error) {
			console.error('Error loading events:', error);
			showToast('Failed to load events', 'error');
		} finally {
			loadingEvents = false;
			isLoading = false;
		}
	}
</script>

<svelte:head>
	<title>Events Management - Streaming Admin</title>
	<meta name="description" content="Manage streaming events and live broadcasts" />
</svelte:head>

<div class="events-page">
	<!-- Header -->
	<header class="page-header">
		<div class="header-content">
			<div class="header-left">
				<button class="btn btn-secondary" on:click={() => goto('/admin')}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="19" y1="12" x2="5" y2="12"></line>
						<polyline points="12,19 5,12 12,5"></polyline>
					</svg>
					Go Back to Main Dashboard
				</button>
				<h1>Events Management</h1>
				<p>Manage live streaming events, schedules, and event-based subscriptions</p>
			</div>
			<div class="header-actions">
				<button class="btn btn-primary">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="12" y1="5" x2="12" y2="19"></line>
						<line x1="5" y1="12" x2="19" y2="12"></line>
					</svg>
					Create Event
				</button>
			</div>
		</div>
	</header>

	<!-- Main Content -->
	<div class="main-content">
		{#if isLoading}
			<div class="loading-container">
				<LoadingSpinner />
				<p>Loading events...</p>
			</div>
		{:else}
			<div class="events-section">
				<div class="section-header">
					<h2>Upcoming Events</h2>
					<div class="section-actions">
						<button class="btn btn-secondary" on:click={loadEvents}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 2v6h-6"></path>
								<path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
								<path d="M3 22v-6h6"></path>
								<path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
							</svg>
							Refresh
						</button>
					</div>
				</div>

				{#if loadingEvents}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading events...</p>
					</div>
				{:else if events.length === 0}
					<div class="empty-state">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
							<line x1="16" y1="2" x2="16" y2="6"></line>
							<line x1="8" y1="2" x2="8" y2="6"></line>
							<line x1="3" y1="10" x2="21" y2="10"></line>
						</svg>
						<h3>No Events Scheduled</h3>
						<p>Create your first streaming event to engage with your audience.</p>
						<button class="btn btn-primary">Create First Event</button>
					</div>
				{:else}
					<div class="events-grid">
						{#each events as event}
							<div class="event-card">
								<!-- Event content would go here -->
							</div>
						{/each}
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

<style>
	.events-page {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: 2rem;
	}

	.header-left h1 {
		font-size: 2rem;
		font-weight: bold;
		margin: 1rem 0 0.5rem 0;
		color: #1f2937;
	}

	.header-left p {
		color: #6b7280;
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border-radius: 0.5rem;
		font-weight: 500;
		text-decoration: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-primary {
		background-color: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background-color: #1d4ed8;
	}

	.btn-secondary {
		background-color: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background-color: #e5e7eb;
	}

	.btn svg {
		width: 1.25rem;
		height: 1.25rem;
	}

	.main-content {
		background: white;
		border-radius: 0.75rem;
		box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1);
		padding: 2rem;
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
	}

	.section-header h2 {
		font-size: 1.5rem;
		font-weight: 600;
		color: #1f2937;
		margin: 0;
	}

	.section-actions {
		display: flex;
		gap: 1rem;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		color: #6b7280;
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
		color: #6b7280;
	}

	.empty-state svg {
		width: 4rem;
		height: 4rem;
		margin-bottom: 1rem;
		color: #d1d5db;
	}

	.empty-state h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: #374151;
		margin: 0 0 0.5rem 0;
	}

	.empty-state p {
		margin: 0 0 1.5rem 0;
		max-width: 400px;
	}

	.events-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
	}
</style> 