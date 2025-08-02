<!-- SubscriptionCheck.svelte -->
<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAdmin, initializeAuth, debugTokenStorage } from '$lib/auth';
	import { subscriptionService, type Subscription } from '$lib/subscription';
	import LoadingSpinner from './LoadingSpinner.svelte';

	export let redirectTo: string = '/login';
	export let requireSubscription: boolean = true;
	export let requiredTier: 'free' | 'premium' = 'premium';
	export let checking: boolean = false;

	// Callback props instead of events
	export let onLoadingChange: (data: { loading: boolean }) => void = () => {};
	export let onAccessGranted: () => void = () => {};

	let loading = true;
	let isAuthenticated = false;
	let hasAccess = false;
	let subscription: Subscription | null = null;
	let user: any = null;
	let authInitialized = false;

	$: checking = loading;

	onMount(async () => {
		console.log('🚀 SubscriptionCheck: Component mounted, initializing auth...');
		
		// Debug token storage state
		const debugInfo = debugTokenStorage();
		console.log('🔐 SubscriptionCheck: Token storage state:', debugInfo);
		
		// Ensure auth is initialized before checking access
		await initializeAuth();
		authInitialized = true;
		console.log('✅ SubscriptionCheck: Auth initialized');

		// Subscribe to auth state changes
		auth.subscribe((state: any) => {
			const wasAuthenticated = isAuthenticated;
			isAuthenticated = state.isAuthenticated;
			user = state.user;
			
			console.log('📡 SubscriptionCheck: Auth state update:', { 
				wasAuthenticated, 
				isAuthenticated: state.isAuthenticated, 
				hasUser: !!state.user,
				userRole: state.user?.role 
			});
			
			// Only trigger access check if auth state actually changed
			if (authInitialized && wasAuthenticated !== isAuthenticated) {
				console.log('🔄 SubscriptionCheck: Auth state changed, triggering access check');
				checkAccess();
			}
		});

		// Initial access check
		await checkAccess();
	});

	async function checkAccess() {
		try {
			loading = true;
			if (typeof onLoadingChange === 'function') {
				onLoadingChange({ loading: true });
			}
			
			console.log('🔍 SubscriptionCheck: Checking access...', { 
				isAuthenticated, 
				requireSubscription,
				requiredTier,
				userRole: user?.role
			});

			// Check authentication first
			if (!isAuthenticated) {
				console.log('❌ SubscriptionCheck: User not authenticated, redirecting to:', redirectTo);
				goto(redirectTo);
				return;
			}

			// Admin users always have access - bypass all subscription checks
			if (isAdmin()) {
				console.log('👑 SubscriptionCheck: Admin user detected, granting access immediately');
				hasAccess = true;
				loading = false;
				if (typeof onLoadingChange === 'function') {
					onLoadingChange({ loading: false });
				}
				if (typeof onAccessGranted === 'function') {
					onAccessGranted();
				}
				return;
			}

			// Only check subscription for non-admin users
			if (requireSubscription) {
				console.log('💳 SubscriptionCheck: Checking subscription status for non-admin user...');
				const response = await subscriptionService.getCurrentSubscription();
				subscription = response.subscription;
				
				console.log('💳 SubscriptionCheck: Subscription response:', { 
					hasSubscription: !!subscription,
					status: subscription?.status,
					tier: subscription?.tier || 'free'
				});

				// If no subscription or not active, redirect to subscription page
				if (!subscription || subscription.status !== 'active') {
					console.log('❌ SubscriptionCheck: No active subscription, redirecting to subscription page');
					goto('/subscription');
					return;
				}

				// Check tier requirements
				if (requiredTier === 'premium' && subscription.tier !== 'premium') {
					console.log('❌ SubscriptionCheck: Premium tier required, redirecting to upgrade page');
					goto('/subscription?upgrade=true');
					return;
				}
			}

			console.log('✅ SubscriptionCheck: Access granted');
			hasAccess = true;
			if (typeof onAccessGranted === 'function') {
				onAccessGranted();
			}
		} catch (err) {
			console.error('❌ SubscriptionCheck: Error checking access:', err);
			// If there's an error checking subscription, redirect to subscription page
			// This ensures users without proper subscription access are redirected
			console.log('❌ SubscriptionCheck: Subscription check failed, redirecting to subscription page');
			goto('/subscription');
		} finally {
			loading = false;
			if (typeof onLoadingChange === 'function') {
				onLoadingChange({ loading: false });
			}
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