<!-- SubscriptionCheck.svelte -->
<script lang="ts">
	import { onMount, createEventDispatcher } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth, isAdmin, initializeAuth, debugTokenStorage } from '$lib/auth';
	import { subscriptionService, type Subscription } from '$lib/subscription';
	import LoadingSpinner from './LoadingSpinner.svelte';

	export let redirectTo: string = '/login';
	export let requireSubscription: boolean = true;
	export let requiredTier: 'free' | 'premium' = 'premium';
	export let checking: boolean = false;

	const dispatch = createEventDispatcher();
	let loading = true;
	let isAuthenticated = false;
	let hasAccess = false;
	let subscription: Subscription | null = null;
	let user: any = null;
	let authInitialized = false;

	$: checking = loading;

	// Reactive statement that triggers when auth state changes
	$: if (authInitialized && !loading) {
		console.log('🔍 SubscriptionCheck: Auth state changed, checking access...', { authInitialized, loading, isAuthenticated });
		checkAccess();
	}

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
		auth.subscribe(state => {
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
			dispatch('loadingChange', { loading: true });
			
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

			// Admin users always have access
			if (isAdmin()) {
				console.log('👑 SubscriptionCheck: Admin user detected, granting access');
				hasAccess = true;
				dispatch('accessGranted');
				return;
			}

			// If subscription is required, check subscription status
			if (requireSubscription) {
				console.log('💳 SubscriptionCheck: Checking subscription status...');
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
			dispatch('accessGranted');
		} catch (err) {
			console.error('❌ SubscriptionCheck: Error checking access:', err);
			// If there's an error checking subscription, but user is authenticated,
			// it might be a network issue - don't redirect to login
			if (isAuthenticated) {
				console.warn('⚠️ SubscriptionCheck: Subscription check failed but user is authenticated, allowing access');
				hasAccess = true;
				dispatch('accessGranted');
			} else {
				console.log('❌ SubscriptionCheck: Auth error, redirecting to subscription page');
				goto('/subscription');
			}
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