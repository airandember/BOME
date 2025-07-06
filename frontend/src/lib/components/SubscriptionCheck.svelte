<!-- SubscriptionCheck.svelte -->
<script lang="ts">
	import { onMount, createEventDispatcher } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAdmin } from '$lib/auth';
	import { subscriptionService, type Subscription } from '$lib/subscription';
	import LoadingSpinner from './LoadingSpinner.svelte';

	export let redirectTo: string = '/login';
	export let requireSubscription: boolean = true;

	const dispatch = createEventDispatcher();
	let loading = true;
	let isAuthenticated = false;
	let hasAccess = false;
	let subscription: Subscription | null = null;
	let user: any = null;

	onMount(async () => {
		auth.subscribe(state => {
			isAuthenticated = state.isAuthenticated;
			user = state.user;
		});

		await checkAccess();
	});

	async function checkAccess() {
		try {
			loading = true;
			dispatch('loadingChange', { loading: true });

			// Check authentication first
			if (!isAuthenticated) {
				goto(redirectTo);
				return;
			}

			// Admin users always have access
			if (isAdmin()) {
				hasAccess = true;
				dispatch('accessGranted');
				return;
			}

			// If subscription is required, check subscription status
			if (requireSubscription) {
				const response = await subscriptionService.getCurrentSubscription();
				subscription = response.subscription;

				// If no subscription or not active, redirect to subscription page
				if (!subscription || subscription.status !== 'active') {
					goto('/subscription');
					return;
				}
			}

			hasAccess = true;
			dispatch('accessGranted');
		} catch (err) {
			console.error('Error checking access:', err);
			goto('/subscription');
		} finally {
			loading = false;
			dispatch('loadingChange', { loading: false });
		}
	}
</script>

{#if loading}
	<div class="loading-container">
		<LoadingSpinner />
		<p>Checking access...</p>
	</div>
{:else if hasAccess}
	<slot></slot>
{/if}

<style>
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem;
		gap: 1rem;
		min-height: 50vh;
	}
</style> 