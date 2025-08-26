<script lang="ts">
	// @ts-nocheck
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { showToast } from '$lib/toast';

	// Import child components
	import Overview from './overview/+page.svelte';
	import Products from './products/+page.svelte';
	import Customers from './customers/+page.svelte';
	import Coupons from './coupons/+page.svelte';
	import Invoices from './invoices/+page.svelte';
	import Payments from './payments/+page.svelte';
	import Subscriptions from './subscriptions/+page.svelte';
	import Setup from './setup/+page.svelte';

	// State variables using Svelte 5 runes
	let summary = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state('overview');
	let loadingStatus = $state('Initializing...');
	let keyType = $state<'sk' | 'rk' | null>(null);
	let dataTransferActive = $state(false);
	let lastDataTransfer = $state<Date | null>(null);
	let lastActivity = $state(0);
	let activityTimeout: number | undefined = $state(undefined);
	let maxTimeoutId: number | undefined = $state(undefined);
	let progressInterval: number | undefined = $state(undefined);
	let debugMode = $state(false);
	let debugInfo = $state({
		backendUrl: '',
		apiBaseUrl: '',
		frontendHost: '',
		frontendProtocol: ''
	});

	// Setup form state (for main page setup)
	let secret = $state('');
	let saving = $state(false);
	let setupError = $state('');
	let setupSuccess = $state('');

	// Modal state for clear key confirmation
	let showClearModal = $state(false);
	let clearConfirmText = $state('');

	// Customer portal link state
	let portalLink = $state('');
	let savedPortalLink = $state(''); // The saved/persisted value
	let savingPortal = $state(false);
	let portalError = $state('');
	let portalSuccess = $state('');
	let editingPortal = $state(false); // Whether we're in edit mode

	// Debug logging for data changes using $effect
	$effect(() => {
		console.log('=== MAIN STRIPE DEBUG ===');
		console.log('Summary data changed:', summary);
		console.log('Summary enabled:', summary?.enabled);
		console.log('Should show setup form:', !summary?.enabled);
		console.log('Coupons count:', summary?.coupons_count);
		console.log('Coupons array length:', summary?.coupons?.length);
		console.log('Active tab:', activeTab);
		console.log('========================');
	});

	// Tab configuration - dynamically filtered based on capabilities
	const allTabs = [
		{ id: 'overview', name: 'Overview', icon: '📊', component: Overview, capability: null },
		{ id: 'products', name: 'Products', icon: '📦', component: Products, capability: 'products' },
		{ id: 'customers', name: 'Customers', icon: '👥', component: Customers, capability: 'customers' },
		{ id: 'coupons', name: 'Coupons', icon: '🎟️', component: Coupons, capability: 'coupons' },
		{ id: 'invoices', name: 'Invoices', icon: '📄', component: Invoices, capability: 'invoices' },
		{ id: 'payments', name: 'Payments', icon: '💳', component: Payments, capability: 'payment_intents' },
		{ id: 'subscriptions', name: 'Subscriptions', icon: '🔄', component: Subscriptions, capability: 'subscriptions' },
		{ id: 'setup', name: 'Setup', icon: '⚙️', component: Setup, capability: null }
	];
	
	// Filter tabs based on capabilities for restricted keys
	const tabs = $derived(() => {
		if (!summary?.enabled || keyType === 'sk') {
			// Full access for secret keys or when not enabled
			return allTabs;
		}
		
		if (keyType === 'rk' && summary?.capabilities) {
			// Filter tabs based on available capabilities for restricted keys
			return allTabs.filter(tab => {
				if (!tab.capability) return true; // Always show overview and setup
				
				const capability = summary.capabilities[tab.capability];
				return capability && Object.values(capability).some(Boolean);
			});
		}
		
		return allTabs;
	});

	// Derived value for active tab configuration
	const activeTabConfig = $derived(tabs.find(tab => tab.id === activeTab));
	
	// Helper function to clean up timers
	const cleanupTimers = () => {
		if (activityTimeout !== undefined) {
			clearTimeout(activityTimeout);
			activityTimeout = undefined;
		}
		// maxTimeoutId is now disabled - no cleanup needed
		// if (maxTimeoutId !== undefined) {
		// 	clearTimeout(maxTimeoutId);
		// 	maxTimeoutId = undefined;
		// }
		if (progressInterval !== undefined) {
			clearInterval(progressInterval);
			progressInterval = undefined;
		}
		dataTransferActive = false;
	};

	onMount(async () => {
		try {
			// Load debug info
			const { getApiBaseUrl, apiBaseUrl } = await import('$lib/config');
			debugInfo.backendUrl = getApiBaseUrl();
			debugInfo.apiBaseUrl = apiBaseUrl;
			debugInfo.frontendHost = window.location.hostname;
			debugInfo.frontendProtocol = window.location.protocol;
			
			console.log('🚀 Initializing Stripe dashboard...');
			console.log('🔧 Debug info loaded:', debugInfo);
			
			await fetchSummary();
			await loadPortalLink();
			
			// If not enabled, default to setup tab
			if (summary && !summary.enabled) {
				activeTab = 'setup';
				console.log('🔧 Stripe not configured, switching to setup tab');
			} else if (summary && summary.enabled) {
				console.log('✅ Stripe configured successfully, dashboard ready');
			}
		} catch (err) {
			console.error('❌ Failed to initialize Stripe dashboard:', err);
			error = 'Failed to initialize Stripe dashboard';
			loading = false;
		}
	});

	async function fetchSummary() {
		try {
			loading = true;
			error = '';
			loadingStatus = 'Connecting to Stripe...';
			console.log('🔍 Fetching Stripe summary...');
			console.log('🌐 Environment:', window.location.hostname);
			console.log('🔑 Using API base:', apiRequest.toString().includes('localhost') ? 'localhost' : 'production');
			
			// Smart timeout system that detects actual data transfer
			const controller = new AbortController();
			lastActivity = Date.now();
			
			// Function to reset activity timer
			const resetActivityTimer = () => {
				lastActivity = Date.now();
				dataTransferActive = true;
				lastDataTransfer = new Date();
				clearTimeout(activityTimeout);
				
				console.log('📊 Stripe API response detected - keeping connection alive');
				
				// No inactivity timeout since we're getting 200 responses from Stripe
				// The backend is processing multiple API calls and we should wait for all of them
				
				// Reset data transfer indicator after 3 seconds
				setTimeout(() => {
					dataTransferActive = false;
				}, 3000);
			};
			
			// Monitor progress and update status
			progressInterval = setInterval(() => {
				const elapsed = Date.now() - lastActivity;
				if (elapsed < 5000) {
					loadingStatus = 'Receiving data from Stripe...';
				} else if (elapsed < 15000) {
					loadingStatus = 'Processing large dataset...';
				} else if (elapsed < 25000) {
					loadingStatus = 'Still processing (large account detected)...';
				} else {
					loadingStatus = 'Finalizing data transfer...';
				}
			}, 2000);
			
			// Initial activity timer
			resetActivityTimer();
			
			// No maximum timeout - let's see how long it actually takes for this large account
			// maxTimeoutId = setTimeout(() => {
			// 	controller.abort();
			// 	console.warn('⏰ Stripe request reached maximum timeout - account may have massive amounts of data');
			// }, 300000);
			console.log('⏱️ No timeout limit set - measuring actual processing time for large Stripe account');
			
			// Try chunked loading first for large accounts
			let res;
			try {
				loadingStatus = 'Loading Stripe data... This may take several minutes for large accounts';
				res = await apiRequest('/admin/streaming/stripe/summary?limit=100', {
					signal: controller.signal,
					headers: {
						'Accept': 'application/json',
						'X-Requested-With': 'XMLHttpRequest'
					},
					// Add progress tracking with elapsed time
					onProgress: () => {
						lastActivity = Date.now();
						dataTransferActive = true;
						resetActivityTimer();
						const elapsed = Math.floor((Date.now() - startTime) / 1000);
						loadingStatus = `Loading Stripe data... ${elapsed}s elapsed (large account detected)`;
						console.log('📊 Data transfer activity detected at', elapsed, 'seconds');
					}
				});
			} catch (chunkedError) {
				console.warn('⚠️ Chunked loading failed, trying standard approach:', chunkedError);
				loadingStatus = 'Falling back to standard loading...';
				
				// Fallback to standard loading
				res = await apiRequest('/admin/streaming/stripe/summary', {
					signal: controller.signal,
					headers: {
						'Accept': 'application/json',
						'X-Requested-With': 'XMLHttpRequest'
					},
					onProgress: () => {
						resetActivityTimer();
						console.log('📊 Data transfer activity detected');
					}
				});
			}
			
			// Clean up timers
			cleanupTimers();
			
			if (res.ok) {
				loadingStatus = 'Processing Stripe data...';
				
				// Check response size
				const contentLength = res.headers.get('content-length');
				if (contentLength) {
					const sizeKB = Math.round(parseInt(contentLength) / 1024);
					console.log(`📊 Response size: ${sizeKB}KB`);
					if (sizeKB > 1000) {
						loadingStatus = `Processing large response (${sizeKB}KB)...`;
					}
				}
				
				const data = await res.json();
				console.log('📥 Raw API response:', data);
				console.log('📊 Response keys:', Object.keys(data));
				
				if (!data || typeof data !== 'object') {
					throw new Error('Invalid response format from server');
				}
				
				loadingStatus = 'Loading dashboard components...';
				summary = data.summary;
				console.log('📊 Summary assigned:', summary);
				console.log('🔧 Summary enabled:', summary?.enabled);
				
				// Detect key type from summary if available
				if (summary && summary.key_type) {
					keyType = summary.key_type;
				}
				
				if (summary && summary.enabled) {
					loadingStatus = 'Finalizing dashboard...';
					console.log('✅ Stripe dashboard data loaded successfully');
				} else {
					console.log('ℹ️ Stripe not configured yet');
				}
			} else {
				const errorText = await res.text().catch(() => 'Unknown error');
				console.error('❌ API request failed:', {
					status: res.status,
					statusText: res.statusText,
					url: res.url,
					headers: Object.fromEntries(res.headers.entries()),
					errorText
				});
				
				let statusText = '';
				let troubleshooting = '';
				
				switch (res.status) {
					case 400:
						statusText = 'Bad Request - Invalid Stripe key format';
						troubleshooting = 'Check that your Stripe key is correctly formatted (sk_test_... or sk_live_...)';
						break;
					case 401:
						statusText = 'Unauthorized - Invalid Stripe key';
						troubleshooting = 'Your Stripe key may be invalid, expired, or have insufficient permissions';
						break;
					case 403:
						statusText = 'Forbidden - Access denied';
						troubleshooting = 'Your Stripe key may not have the required permissions for this operation';
						break;
					case 404:
						statusText = 'Stripe service not found';
						troubleshooting = 'The backend Stripe service may not be running or configured';
						break;
					case 413:
						statusText = 'Payload too large';
						troubleshooting = 'Your Stripe account data is too large. Try using a restricted key or contact support';
						break;
					case 429:
						statusText = 'Rate limit exceeded';
						troubleshooting = 'Too many requests to Stripe API. Please wait a moment and try again';
						break;
					case 500:
						statusText = 'Server error';
						troubleshooting = 'Internal server error while processing Stripe data';
						break;
					case 503:
						statusText = 'Service unavailable';
						troubleshooting = 'Backend service is temporarily unavailable';
						break;
					case 504:
						statusText = 'Gateway timeout';
						troubleshooting = 'Request to Stripe API timed out. Your account may be very large';
						break;
					default:
						statusText = `HTTP ${res.status}`;
						troubleshooting = 'Unexpected error occurred';
				}
				
				error = `Failed to load Stripe data: ${statusText}. ${troubleshooting}`;
			}
		} catch (err: any) {
			// Clean up timers on error
			cleanupTimers();
			
			if (err.name === 'AbortError') {
				// No timeout set, so this would be a manual abort or network issue
				error = 'Request was aborted. This could be due to a network issue or manual cancellation.';
				console.error('🚫 Stripe summary request was aborted');
				console.log('💡 Check network connection or try refreshing the page');
			} else if (err.message?.includes('fetch')) {
				error = 'Network error. Please check your connection and try again.';
				console.error('🌐 Network error:', err);
			} else {
				error = err.message || 'Failed to load Stripe data';
				console.error('❌ Fetch summary error:', err);
			}
		} finally {
			loading = false;
		}
	}

	async function loadPortalLink() {
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link');
			if (res.ok) {
				const data = await res.json();
				savedPortalLink = data.portal_url;
				portalLink = savedPortalLink; // Initialize portalLink for editing
			} else if (res.status === 500) {
				// Likely means Stripe isn't configured yet - this is normal
				console.warn('Portal link unavailable (normal if Stripe not configured):', res.status);
				savedPortalLink = '';
				portalLink = '';
				portalError = ''; // Clear any previous errors
			} else {
				console.error('Failed to load portal link:', res.status);
				portalError = 'Failed to load customer portal link';
			}
		} catch (err) {
			// Handle network errors or other issues gracefully
			console.warn('Portal link unavailable (normal if Stripe not configured):', err);
			savedPortalLink = '';
			portalLink = '';
			portalError = ''; // Clear any previous errors
		}
	}

	function switchTab(tabId: string) {
		activeTab = tabId;
	}

	// Test connection to backend and Stripe with detailed request logging
	async function testConnection() {
		console.log('🧪 Testing connection...');
		loadingStatus = 'Testing connection...';
		
		try {
			// Import config to get base URL
			const { getApiBaseUrl } = await import('$lib/config');
			const baseUrl = getApiBaseUrl();
			
			console.log('🌐 Testing backend at:', baseUrl);
			
			// Test 1: Backend connectivity (health endpoint is at root level, not /api/v1)
			const healthUrl = `${baseUrl}/health`;
			console.log('🔍 Health check URL:', healthUrl);
			
			const backendTest = await fetch(healthUrl, {
				method: 'GET',
				headers: {
					'Accept': 'application/json',
				},
			});
			
			console.log('📊 Backend response:', {
				ok: backendTest.ok,
				status: backendTest.status,
				statusText: backendTest.statusText,
				url: backendTest.url
			});
			
			if (backendTest.ok) {
				const healthData = await backendTest.json();
				console.log('✅ Backend health data:', healthData);
				loadingStatus = 'Testing Stripe request logging...';
				
				// Test 2: Stripe Request Diagnostics with detailed logging
				if (summary?.enabled) {
					try {
						console.log('🔍 Testing Stripe request with detailed logging...');
						console.log('📤 About to send Stripe API request');
						console.log('🔑 Using key type:', keyType);
						console.log('🌐 Request URL will be:', `${baseUrl}/api/v1/admin/streaming/stripe/summary`);
						
						// Add timestamp for request tracking
						const requestTimestamp = new Date().toISOString();
						console.log('⏰ Request timestamp:', requestTimestamp);
						
						// Test with detailed request monitoring
						const controller = new AbortController();
						const startTime = Date.now();
						
						console.log('📡 Sending request to backend...');
						const stripeTest = await apiRequest('/admin/streaming/stripe/summary', {
							timeout: 30000, // 30 second timeout for diagnostic
							signal: controller.signal,
							// Add custom headers for tracking
							headers: {
								'X-Request-ID': `test-${Date.now()}`,
								'X-Debug-Mode': 'true'
							}
						});
						
						const endTime = Date.now();
						const duration = endTime - startTime;
						
						console.log('📊 Request completed:', {
							ok: stripeTest.ok,
							status: stripeTest.status,
							duration: `${duration}ms`,
							timestamp: requestTimestamp
						});
						
						if (stripeTest.ok) {
							console.log('✅ Stripe request successful - check Stripe logs for request at:', requestTimestamp);
							loadingStatus = `Stripe request successful (${duration}ms) - Check Stripe logs at ${new Date(requestTimestamp).toLocaleTimeString()}`;
							
							// Try to get response data for additional verification
							try {
								const responseData = await stripeTest.json();
								console.log('📄 Response data preview:', {
									hasData: !!responseData,
									keys: responseData ? Object.keys(responseData) : [],
									enabled: responseData?.enabled
								});
							} catch (parseErr) {
								console.log('⚠️ Could not parse response data:', parseErr);
							}
						} else {
							console.error('❌ Stripe request failed:', {
								status: stripeTest.status,
								statusText: stripeTest.statusText,
								duration: `${duration}ms`
							});
							loadingStatus = `Stripe request failed (${stripeTest.status}) - No request should appear in Stripe logs`;
						}
					} catch (stripeErr) {
						console.error('❌ Stripe request error:', stripeErr);
						console.log('🔍 Detailed error analysis:', {
							name: stripeErr instanceof Error ? stripeErr.name : 'Unknown',
							message: stripeErr instanceof Error ? stripeErr.message : String(stripeErr),
							stack: stripeErr instanceof Error ? stripeErr.stack?.split('\n').slice(0, 3) : 'No stack trace'
						});
						
						if (stripeErr instanceof Error && stripeErr.name === 'AbortError') {
							loadingStatus = 'Request aborted - No request should appear in Stripe logs';
						} else {
							loadingStatus = 'Request error - Check console for details';
						}
					}
				} else {
					loadingStatus = 'Backend connected - Stripe not configured';
				}
			} else {
				console.error('❌ Backend health check failed');
				loadingStatus = `Backend connection failed (${backendTest.status})`;
			}
			
		} catch (err) {
			console.error('❌ Connection test failed:', err);
			loadingStatus = `Connection test failed: ${err instanceof Error ? err.message : String(err)}`;
		}
	}

	// Helper function to fetch specific Stripe sections
	async function fetchStripeSection(section: string, limit: number = 100) {
		try {
			console.log(`🔍 Fetching Stripe section: ${section} with limit ${limit}`);
			const startTime = Date.now();
			
			const res = await apiRequest(`/admin/streaming/stripe/summary?section=${section}&limit=${limit}`, {
				headers: {
					'X-Request-ID': `section-${section}-${Date.now()}`,
					'X-Debug-Mode': 'true'
				}
			});
			
			if (res.ok) {
				const data = await res.json();
				const duration = Date.now() - startTime;
				console.log(`✅ Section ${section} loaded in ${duration}ms:`, data.summary);
				return data.summary;
			} else {
				throw new Error(`Failed to fetch ${section}: ${res.status}`);
			}
		} catch (err) {
			console.error(`❌ Error fetching section ${section}:`, err);
			throw err;
		}
	}

	// Setup form functions for main page
	async function saveSecret() {
		if (!secret.trim()) {
			setupError = 'Please enter a Stripe secret key';
			return;
		}
		
		// Validate key format - support both sk_ and rk_ keys
		const isSecretKey = secret.startsWith('sk_test_') || secret.startsWith('sk_live_');
		const isRestrictedKey = secret.startsWith('rk_test_') || secret.startsWith('rk_live_');
		
		if (!isSecretKey && !isRestrictedKey) {
			setupError = 'Invalid Stripe key format. Key should start with sk_test_, sk_live_, rk_test_, or rk_live_';
			return;
		}
		
		// Set key type for later use
		keyType = isSecretKey ? 'sk' : 'rk';
		
		if (secret.length < 20) {
			setupError = 'Stripe key appears to be too short. Please check your key.';
			return;
		}
		
		saving = true;
		setupError = '';
		setupSuccess = '';
		
		console.log('🔐 Saving Stripe secret key...');
		
		try {
			// Add timeout for key saving
			const controller = new AbortController();
			const timeoutId = setTimeout(() => {
				controller.abort();
				console.warn('⏰ Save key request timed out after 30 seconds');
			}, 30000);
			
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ key: secret }),
				signal: controller.signal
			});
			
			clearTimeout(timeoutId);
			
			if (res.ok) {
				console.log('✅ Stripe key saved successfully');
				setupSuccess = 'Stripe key saved successfully! Loading dashboard...';
				secret = '';
				
				// Wait a moment for the backend to process
				await new Promise(resolve => setTimeout(resolve, 1000));
				
				// Refresh the summary with enhanced error handling
				console.log('🔄 Refreshing Stripe dashboard data...');
				await fetchSummary();
				
				// Verify the key was actually saved and working
				if (!summary || !summary.enabled) {
					setupError = 'Key saved but Stripe dashboard failed to load. Please check your key and try refreshing the page.';
					setupSuccess = '';
				} else {
					setupSuccess = 'Stripe connected successfully! Dashboard is now available.';
					console.log('🎉 Stripe dashboard fully loaded and ready');
				}
			} else {
				let errorMessage = 'Failed to save key';
				try {
					const errorData = await res.json();
					errorMessage = errorData.error || errorMessage;
					
					// Provide specific error messages based on status
					if (res.status === 400) {
						errorMessage = errorData.error || 'Invalid Stripe key format';
					} else if (res.status === 401) {
						errorMessage = 'Invalid Stripe key. Please check your key and try again.';
					} else if (res.status === 403) {
						errorMessage = 'Stripe key does not have sufficient permissions';
					} else if (res.status === 500) {
						errorMessage = 'Server error while saving key. Please try again.';
					} else if (res.status === 503) {
						errorMessage = 'Service temporarily unavailable. Please try again later.';
					}
				} catch (parseErr) {
					console.warn('Could not parse error response:', parseErr);
				}
				
				setupError = errorMessage;
				console.error('❌ Save key failed:', res.status, res.statusText, errorMessage);
			}
		} catch (err: any) {
			if (err.name === 'AbortError') {
				setupError = 'Request timed out. Please check your connection and try again.';
				console.error('⏰ Save key request timed out');
			} else if (err.message?.includes('fetch') || err.message?.includes('network')) {
				setupError = 'Network error. Please check your connection and try again.';
				console.error('🌐 Network error while saving key:', err);
			} else {
				setupError = err.message || 'Failed to save key. Please try again.';
				console.error('❌ Unexpected error while saving key:', err);
			}
		} finally {
			saving = false;
		}
	}

	// Show the clear confirmation modal
	function showClearConfirmation() {
		showClearModal = true;
		clearConfirmText = '';
	}

	// Close the modal and reset
	function closeClearModal() {
		showClearModal = false;
		clearConfirmText = '';
	}

	// Confirm and execute the clear action
	async function confirmClearKey() {
		if (clearConfirmText !== 'sk_1337') {
			return; // Don't proceed if confirmation text doesn't match
		}

		// Close modal first
		closeClearModal();

		// Execute the clear action
		saving = true;
		setupError = '';
		setupSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ key: 'sk_1337' })
			});
			
			if (res.ok) {
				// Show toast notification
				showToast('Stripe key cleared successfully!', 'success');
				
				console.log('🔄 Clear key successful, resetting dashboard state...');
				
				// CRITICAL: Set summary to disabled state to trigger line 334 condition
				summary = { enabled: false };
				
				// Switch back to overview tab
				activeTab = 'overview';
				
				// Clear any success messages
				setupSuccess = '';
				setupError = '';
				
				// Also clear the portal link since Stripe is no longer configured
				savedPortalLink = '';
				portalLink = '';
				editingPortal = false;
				portalError = '';
				portalSuccess = '';
				
				console.log('✅ Summary reset to disabled state:', summary);
				console.log('✅ Active tab switched to overview:', activeTab);
				console.log('✅ Portal link cleared along with Stripe key');
			} else {
				const errorData = await res.json();
				setupError = errorData.error || 'Failed to clear key';
				showToast('Failed to clear Stripe key', 'error');
			}
		} catch (err) {
			setupError = 'Failed to clear key';
			showToast('Failed to clear Stripe key', 'error');
			console.error(err);
		} finally {
			saving = false;
		}
	}

	// Save customer portal link
	async function savePortalLink() {
		if (!portalLink.trim()) return;
		
		savingPortal = true;
		portalError = '';
		portalSuccess = '';
		
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link', {
				method: 'POST',
				body: JSON.stringify({ portal_url: portalLink })
			});

			if (res.ok) {
				const data = await res.json();
				savedPortalLink = portalLink; // Use the local value since backend doesn't return it
				editingPortal = false;
				portalSuccess = data.message || 'Customer portal link saved successfully!';
			} else {
				const errorData = await res.json();
				portalError = errorData.error || 'Failed to save portal link';
			}
		} catch (err) {
			portalError = 'Failed to save portal link';
			console.error(err);
		} finally {
			savingPortal = false;
		}
	}

	// Start editing portal link
	function startEditingPortal() {
		editingPortal = true;
		portalLink = savedPortalLink;
		portalError = '';
		portalSuccess = '';
	}

	// Cancel editing portal link
	function cancelEditingPortal() {
		editingPortal = false;
		portalLink = '';
		portalError = '';
		portalSuccess = '';
	}

	// Clear saved portal link
	async function clearPortalLink() {
		try {
			const res = await apiRequest('/admin/streaming/stripe/portal-link', {
				method: 'DELETE'
			});
			
			if (res.ok) {
				savedPortalLink = '';
				portalLink = '';
				editingPortal = false;
				portalSuccess = 'Portal link cleared successfully!';
				portalError = '';
			} else {
				const errorData = await res.json();
				portalError = errorData.error || 'Failed to clear portal link';
			}
		} catch (err) {
			// Even if the API call fails, we can clear the local state
			console.warn('API call to clear portal link failed, clearing local state:', err);
			savedPortalLink = '';
			portalLink = '';
			editingPortal = false;
			portalSuccess = 'Portal link cleared locally (API unavailable)';
			portalError = '';
		}
	}
</script>

{#if loading}
	<div class="loading-container">
		<div class="spinner"></div>
		<p class="loading-title">Loading Stripe Dashboard...</p>
		<p class="loading-status">{loadingStatus}</p>
		
		{#if dataTransferActive}
			<div class="data-transfer-indicator">
				<div class="transfer-pulse"></div>
				<span>📊 Data transferring...</span>
			</div>
		{:else if lastDataTransfer}
			<div class="last-transfer-info">
				<small>Last data received: {lastDataTransfer.toLocaleTimeString()}</small>
			</div>
		{/if}
		
		{#if loadingStatus.includes('Processing') || loadingStatus.includes('Loading') || loadingStatus.includes('Receiving')}
			<div class="loading-info">
				<small>⏳ Large Stripe accounts may take several minutes to load</small>
				<small>🔄 Smart timeout: Will only cancel if no data is being transferred</small>
			</div>
		{/if}
	</div>
{:else if error}
	<div class="error-container">
		<h3>⚠️ Error Loading Stripe Dashboard</h3>
		<p class="error-message">{error}</p>
		<div class="error-details">
			<details>
				<summary>🔍 Troubleshooting Tips</summary>
				<div class="troubleshooting-content">
					<ul>
						<li><strong>Network Issues:</strong> Check your internet connection and try again</li>
						<li><strong>Server Issues:</strong> The backend service may be temporarily unavailable</li>
						<li><strong>Invalid Key:</strong> Your Stripe key may be invalid or expired</li>
						<li><strong>Permissions:</strong> Your Stripe key may not have sufficient permissions</li>
						<li><strong>Configuration:</strong> Try clearing your Stripe configuration and re-entering your key</li>
					</ul>
				</div>
			</details>
		</div>
		<div class="error-actions">
			<button class="btn btn-primary" onclick={fetchSummary} disabled={loading}>
				{loading ? '🔄 Retrying...' : '🔄 Retry'}
			</button>
			<button class="btn btn-secondary" onclick={testConnection} disabled={loading}>
				🧪 Test Connection
			</button>
			<button class="btn btn-secondary" onclick={() => debugMode = !debugMode}>
				{debugMode ? '🔍 Hide Debug' : '🔍 Show Debug'}
			</button>
			<button class="btn btn-secondary" onclick={() => showClearModal = true}>
				🗑️ Clear Stripe Configuration
			</button>
		</div>
		
		{#if debugMode}
			<div class="debug-info">
				<h4>🔍 Debug Information</h4>
				<div class="debug-grid">
					<div class="debug-item">
						<strong>Frontend Host:</strong> {debugInfo.frontendHost}
					</div>
					<div class="debug-item">
						<strong>Frontend Protocol:</strong> {debugInfo.frontendProtocol}
					</div>
					<div class="debug-item">
						<strong>Backend URL:</strong> {debugInfo.backendUrl}
					</div>
					<div class="debug-item">
						<strong>API Base URL:</strong> {debugInfo.apiBaseUrl}
					</div>
					<div class="debug-item">
						<strong>Last Activity:</strong> {lastActivity > 0 ? new Date(lastActivity).toLocaleTimeString() : 'None'}
					</div>
					<div class="debug-item">
						<strong>Data Transfer:</strong> {dataTransferActive ? 'Active' : 'Inactive'}
					</div>
					<div class="debug-item">
						<strong>Key Type:</strong> {keyType || 'Unknown'}
					</div>
					<div class="debug-item">
						<strong>Loading Status:</strong> {loadingStatus}
					</div>
					<div class="debug-item">
						<strong>Stripe Key Prefix:</strong> {secret ? secret.substring(0, 8) + '...' : 'None'}
					</div>
					<div class="debug-item">
						<strong>Account Type:</strong> {summary?.customers_total_estimated >= 1000 ? 'Large Account (3000+ customers)' : 'Standard Account'}
					</div>
					<div class="debug-item">
						<strong>Data Mode:</strong> {summary?.customers_total_estimated >= 1000 ? 'Optimized Sampling' : 'Full Data'}
					</div>
				</div>
				<div class="debug-suggestions">
					<h5>💡 Troubleshooting Suggestions:</h5>
					<ul>
						<li><strong>Account-Specific Issue:</strong> This Stripe account appears to have unique limitations</li>
						<li><strong>API Rate Limits:</strong> Account may be hitting Stripe's API rate limits</li>
						<li><strong>Large Data Volume:</strong> Try using a restricted key (rk_) to limit data size</li>
						<li><strong>Account Verification:</strong> Ensure the Stripe account is fully activated</li>
						<li><strong>Key Permissions:</strong> Verify the key has full read/write access</li>
						<li><strong>Test vs Live Mode:</strong> Check if you're using the correct environment key</li>
						<li><strong>Network/Firewall:</strong> Check firewall and proxy settings</li>
					</ul>
				</div>
			</div>
		{/if}
	</div>
{:else}
	<div class="stripe-dashboard">
		<!-- Header -->
		<div class="dashboard-header">
			<!--<div class="header-content">
				<h1>Stripe Dashboard</h1>
				<p>Manage payments, subscriptions, and customer data</p>
			</div>-->
			
			{#if summary?.enabled}
				<div class="header-status">
					<div class="status-indicator connected">
						<div class="status-dot"></div>
						<span>Connected</span>
					</div>
					<div class="environment-badge {summary.environment === 'live' ? 'live' : 'test'}">
						{summary.environment === 'live' ? '🟩 LIVE' : '🟡 TEST'}
					</div>
				</div>
			{:else}
				<div class="header-status">
					<div class="status-indicator disconnected">
						<div class="status-dot"></div>
						<span>Not Connected</span>
					</div>
				</div>
			{/if}
		</div>

		{#if !summary?.enabled}
			<!-- Setup Container for Main Page -->
			<div class="setup-container">
				<div class="setup-header">
					<h1>Setup Stripe</h1>
					<p>Configure your Stripe integration to start processing payments</p>
				</div>

				<div class="setup-section">
					<div class="setup-card">
						<div class="card-header">
							<h2>🔑 Connect Your Stripe Account</h2>
							<p>Enter your Stripe secret key to enable payment processing</p>
						</div>
						
						<form onsubmit={(e) => { e.preventDefault(); saveSecret(); }} class="setup-form">
							<div class="input-group">
								<label for="stripe-key" class="input-label">
									Stripe Secret Key
									<span class="required">*</span>
								</label>
								<input 
									id="stripe-key"
									class="input" 
									type="password" 
									placeholder="sk_test_..., sk_live_..., rk_test_..., or rk_live_..." 
									value={secret}
									oninput={(e) => secret = (e.target as HTMLInputElement).value}
									required
								/>
								<div class="input-help">
									Your secret key will be encrypted and stored securely. It will never be returned or displayed.
								</div>
							</div>

							<button 
								type="submit" 
								class="btn btn-primary btn-lg" 
								disabled={saving || !secret.trim()}
							>
								{#if saving}
									<span class="btn-spinner"></span>
									Connecting...
								{:else}
									🔗 Connect Stripe Account
								{/if}
							</button>
						</form>
						
						{#if setupError}
							<div class="alert alert-error">
								<div class="alert-icon">❌</div>
								<div class="alert-content">
									<strong>Connection Failed</strong>
									<p>{setupError}</p>
								</div>
							</div>
						{/if}
						
						{#if setupSuccess}
							<div class="alert alert-success">
								<div class="alert-icon">✅</div>
								<div class="alert-content">
									<strong>Success!</strong>
									<p>{setupSuccess}</p>
								</div>
							</div>
						{/if}
					</div>

					<!-- Setup Instructions -->
					<div class="instructions-card">
						<h3>🚀 Getting Started</h3>
						<div class="steps">
							<div class="step">
								<div class="step-number">1</div>
								<div class="step-content">
									<h4>Create a Stripe Account</h4>
									<p>Visit <a href="https://stripe.com" target="_blank" rel="noopener">stripe.com</a> to create your account if you haven't already.</p>
								</div>
							</div>
							
							<div class="step">
								<div class="step-number">2</div>
								<div class="step-content">
									<h4>Get Your API Keys</h4>
									<p>In your Stripe Dashboard, go to <strong>Developers → API Keys</strong> to find your secret key.</p>
								</div>
							</div>
							
							<div class="step">
								<div class="step-number">3</div>
								<div class="step-content">
									<h4>Test vs Live Mode</h4>
									<p>Start with test keys (sk_test_...) for development, then switch to live keys (sk_live_...) for production.</p>
								</div>
							</div>
							
							<div class="step">
								<div class="step-number">4</div>
								<div class="step-content">
									<h4>Enter Your Key</h4>
									<p>Paste your secret key above and click "Connect Stripe Account" to get started.</p>
								</div>
							</div>
						</div>
					</div>

					<!-- Security Notice -->
					<div class="security-notice">
						<div class="notice-icon">🔒</div>
						<div class="notice-content">
							<h4>Security & Privacy</h4>
							<ul>
								<li>Your secret key is encrypted using AES-GCM encryption before storage</li>
								<li>Keys are never returned to the frontend or displayed anywhere</li>
								<li>Only you can replace the key by entering a new one</li>
								<li>All communication uses HTTPS encryption</li>
							</ul>
						</div>
					</div>

					<!-- Customer Portal Setup -->
					<div class="instructions-card">
						<h3>🔗 Customer Portal Setup</h3>
						<p>Configure your Stripe customer portal link for subscription management</p>
						{#if savedPortalLink && !editingPortal}
							<!-- Display saved portal link -->
							<div class="saved-portal">
								<div class="saved-portal-display">
									<div class="saved-portal-label">Current Portal URL:</div>
									<div class="saved-portal-url">
										<a href={savedPortalLink} target="_blank" rel="noopener">
											{savedPortalLink}
										</a>
									</div>
								</div>
								<div class="saved-portal-actions">
									<button class="btn btn-outline" onclick={startEditingPortal}>
										✏️ Update Link
									</button>
									<button class="btn btn-secondary" onclick={clearPortalLink}>
										🗑️ Clear Link
									</button>
								</div>
							</div>
						{:else}
							<!-- Edit/Add form -->
							<form onsubmit={(e) => { e.preventDefault(); savePortalLink(); }} class="portal-form">
								<div class="input-group">
									<label for="main-portal-link" class="input-label">
										Customer Portal URL
									</label>
									<input 
										id="main-portal-link"
										class="input" 
										type="url" 
										placeholder="Enter your customer portal URL here" 
										value={portalLink}
										oninput={(e) => portalLink = (e.target as HTMLInputElement).value}
									/>
									<div class="input-help">
										Get this URL from your Stripe Dashboard → Settings → Customer Portal
									</div>
								</div>

								<div class="portal-form-actions">
									<button 
										type="submit" 
										class="btn btn-primary" 
										disabled={savingPortal || !portalLink.trim()}
									>
										{savingPortal ? 'Saving...' : (savedPortalLink ? 'Update Portal Link' : 'Save Portal Link')}
									</button>
									{#if editingPortal}
										<button 
											type="button" 
											class="btn btn-secondary" 
											onclick={cancelEditingPortal}
										>
											Cancel
										</button>
									{/if}
								</div>
							</form>
						{/if}
						
						{#if portalError}
							<div class="alert alert-error">
								<div class="alert-icon">❌</div>
								<div class="alert-content">
									<strong>Error</strong>
									<p>{portalError}</p>
								</div>
							</div>
							
						{/if}
						
						{#if portalSuccess}
							<div class="alert alert-success">
								<div class="alert-icon">✅</div>
								<div class="alert-content">
									<strong>Success!</strong>
									<p>{portalSuccess}</p>
								</div>
							</div>
						{/if}
					</div>
				</div>
			</div>
		{:else}
			<!-- Tab Navigation -->
			<div class="tab-navigation">
				<div class="tab-list">
					{#each tabs() as tab}
						{@const isRestricted = keyType === 'rk' && tab.capability && summary?.capabilities?.[tab.capability] && !Object.values(summary.capabilities[tab.capability]).some(Boolean)}
						<button 
							class="tab-button {activeTab === tab.id ? 'active' : ''}"
							class:disabled={!summary?.enabled && tab.id !== 'overview' && tab.id !== 'setup'}
							class:restricted={isRestricted}
							onclick={() => switchTab(tab.id)}
							disabled={!summary?.enabled && tab.id !== 'overview' && tab.id !== 'setup'}
							title={isRestricted ? 'This functionality is restricted by your Stripe key permissions' : ''}
						>
							<span class="tab-icon">{tab.icon}</span>
							<span class="tab-name">{tab.name}</span>
							{#if !summary?.enabled && tab.id !== 'overview' && tab.id !== 'setup'}
								<span class="tab-lock">🔒</span>
							{:else if isRestricted}
								<span class="tab-restricted">⚠️</span>
							{/if}
						</button>
					{/each}
				</div>
				{#if keyType === 'rk'}
					<div class="key-type-indicator">
						<span class="key-badge restricted">🔑 Restricted Key (rk_)</span>
						<small>Some functionality may be limited</small>
					</div>
				{:else if keyType === 'sk'}
					<div class="key-type-indicator">
						<span class="key-badge full">🔑 Secret Key (sk_)</span>
						<small>Full access enabled</small>
					</div>
				{/if}
			</div>

			<!-- Tab Content -->
			<div class="tab-content">
				{#if keyType === 'rk'}
					{@const currentTab = tabs().find(t => t.id === activeTab)}
					{@const isTabRestricted = currentTab?.capability && summary?.capabilities?.[currentTab.capability] && !Object.values(summary.capabilities[currentTab.capability]).some(Boolean)}
					
					{#if isTabRestricted}
						<div class="restricted-content">
							<div class="restricted-icon">⚠️</div>
							<h2>This Functionality is Restricted by Stripe</h2>
							<p>Your current Stripe restricted key (rk_) does not have permissions to access <strong>{currentTab.name}</strong> functionality.</p>
							<div class="restricted-details">
								<h3>What you can do:</h3>
								<ul>
									<li>Contact your Stripe account administrator to request additional permissions</li>
									<li>Use a secret key (sk_) instead for full access</li>
									<li>Check your Stripe Dashboard settings for key permissions</li>
								</ul>
							</div>
							<div class="restricted-actions">
								<button class="btn btn-primary" onclick={() => switchTab('overview')}>
									← Back to Overview
								</button>
								<button class="btn btn-secondary" onclick={() => switchTab('setup')}>
									⚙️ Update Key
								</button>
							</div>
						</div>
					{:else if activeTab === 'overview'}
						<Overview data={summary} />
					{:else if activeTab === 'products'}
						<Products data={summary} />
					{:else if activeTab === 'customers'}
						<Customers data={summary} />
					{:else if activeTab === 'coupons'}
						<Coupons data={summary} />
					{:else if activeTab === 'invoices'}
						<Invoices data={summary} />
					{:else if activeTab === 'payments'}
						<Payments data={summary} />
					{:else if activeTab === 'subscriptions'}
						<Subscriptions data={summary} />
					{:else if activeTab === 'setup'}
						<Setup data={summary} onClearKey={showClearConfirmation} />
					{/if}
				{:else}
					{#if activeTab === 'overview'}
						<Overview data={summary} />
					{:else if activeTab === 'products'}
						<Products data={summary} />
					{:else if activeTab === 'customers'}
						<Customers data={summary} />
					{:else if activeTab === 'coupons'}
						<Coupons data={summary} />
					{:else if activeTab === 'invoices'}
						<Invoices data={summary} />
					{:else if activeTab === 'payments'}
						<Payments data={summary} />
					{:else if activeTab === 'subscriptions'}
						<Subscriptions data={summary} />
					{:else if activeTab === 'setup'}
						<Setup data={summary} onClearKey={showClearConfirmation} />
					{/if}
				{/if}
			</div>
		{/if}
	</div>
{/if}

<!-- Clear Key Confirmation Modal -->
{#if showClearModal}
	<div class="modal-overlay" onclick={closeClearModal} onkeydown={(e) => e.key === 'Escape' && closeClearModal()} role="dialog" aria-modal="true" tabindex="-1">
		<div class="modal-content" role="document" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h3>⚠️ Clear Stripe Key</h3>
				<button class="modal-close" onclick={closeClearModal}>&times;</button>
			</div>
			
			<div class="modal-body">
				<p><strong>Are you sure you want to clear your Stripe secret key?</strong></p>
				<p>This action will:</p>
				<ul>
					<li>Disable all Stripe payment processing</li>
					<li>Remove your stored secret key</li>
					<li>Return you to the setup screen</li>
				</ul>
				
				<div class="confirmation-input">
					<label for="confirm-text" class="input-label">
						Type <code>sk_1337</code> to confirm:
					</label>
					<input 
						id="confirm-text"
						class="input" 
						type="text" 
						placeholder="sk_1337"
						value={clearConfirmText}
						oninput={(e) => clearConfirmText = (e.target as HTMLInputElement).value}
						onkeydown={(e) => e.key === 'Enter' && clearConfirmText === 'sk_1337' && confirmClearKey()}
					/>
				</div>
			</div>
			
			<div class="modal-footer">
				<button class="btn btn-secondary" onclick={closeClearModal}>
					Cancel
				</button>
				<button 
					class="btn btn-danger" 
					disabled={clearConfirmText !== 'sk_1337' || saving}
					onclick={confirmClearKey}
				>
					{#if saving}
						Clearing...
					{:else}
						🗑️ Clear Key
					{/if}
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.stripe-dashboard {
		min-height: 100vh;
		background: var(--bg-primary);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		min-height: 50vh;
	}

	.loading-title {
		font-size: 1.2rem;
		font-weight: 600;
		margin: var(--space-md) 0 var(--space-sm) 0;
		color: var(--text);
	}

	.loading-status {
		font-size: 1rem;
		color: var(--text-muted);
		margin: 0 0 var(--space-sm) 0;
		font-weight: 500;
	}

	.loading-info {
		margin-top: var(--space-md);
		padding: var(--space-sm) var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.loading-info small {
		color: var(--text-muted);
		font-style: italic;
	}

	.data-transfer-indicator {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		margin: var(--space-md) 0;
		padding: var(--space-sm) var(--space-md);
		background: var(--success-light);
		border-radius: var(--radius-md);
		border: 1px solid var(--success);
		color: var(--success-dark);
		font-weight: 500;
	}

	.transfer-pulse {
		width: 12px;
		height: 12px;
		background: var(--success);
		border-radius: 50%;
		animation: pulse 1s ease-in-out infinite;
	}

	@keyframes pulse {
		0% {
			transform: scale(0.8);
			opacity: 1;
		}
		50% {
			transform: scale(1.2);
			opacity: 0.7;
		}
		100% {
			transform: scale(0.8);
			opacity: 1;
		}
	}

	.last-transfer-info {
		margin: var(--space-sm) 0;
		padding: var(--space-xs) var(--space-sm);
		background: var(--bg-secondary);
		border-radius: var(--radius-sm);
		border: 1px solid var(--border);
	}

	.last-transfer-info small {
		color: var(--text-muted);
		font-size: 0.75rem;
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border);
		border-top: 4px solid var(--primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: var(--space-md);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.error-container h3 {
		color: var(--error);
		margin-bottom: var(--space-md);
	}

	.error-message {
		font-weight: 500;
		margin-bottom: var(--space-lg);
		padding: var(--space-md);
		background: var(--error-light);
		border-radius: var(--radius-md);
		border-left: 4px solid var(--error);
	}

	.error-details {
		margin-bottom: var(--space-lg);
		text-align: left;
	}

	.error-details summary {
		cursor: pointer;
		font-weight: 600;
		color: var(--text);
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--surface);
		border: 1px solid var(--border);
		user-select: none;
	}

	.error-details summary:hover {
		background: var(--surface-hover);
	}

	.troubleshooting-content {
		margin-top: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.troubleshooting-content ul {
		margin: 0;
		padding-left: var(--space-lg);
	}

	.troubleshooting-content li {
		margin-bottom: var(--space-sm);
		line-height: 1.5;
	}

	.error-actions {
		display: flex;
		gap: var(--space-md);
		margin-top: var(--space-md);
		flex-wrap: wrap;
		justify-content: center;
	}

	.debug-info {
		margin-top: var(--space-lg);
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		text-align: left;
		max-width: 800px;
		width: 100%;
	}

	.debug-info h4 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.2rem;
	}

	.debug-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-sm);
		margin-bottom: var(--space-lg);
	}

	.debug-item {
		padding: var(--space-sm);
		background: var(--surface);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
		font-size: 0.9rem;
	}

	.debug-item strong {
		color: var(--text);
		display: block;
		margin-bottom: var(--space-xs);
	}

	.debug-suggestions {
		background: var(--surface);
		padding: var(--space-md);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.debug-suggestions h5 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1rem;
	}

	.debug-suggestions ul {
		margin: 0;
		padding-left: var(--space-lg);
		color: var(--text-muted);
	}

	.debug-suggestions li {
		margin-bottom: var(--space-xs);
		line-height: 1.4;
	}

	.btn {
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 1rem;
		cursor: pointer;
		transition: all 0.2s ease;
		font-weight: 600;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-dark);
		transform: translateY(-1px);
	}

	.btn-spinner {
		display: inline-block;
		width: 16px;
		height: 16px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top: 2px solid white;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-right: var(--space-sm);
	}

	.dashboard-header {
		display: flex;
		justify-content: flex-end;
		align-items: center;
		background: var(--surface);
		border-bottom: 1px solid var(--border);
	}



	.header-status {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.status-indicator {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.9rem;
		font-weight: 600;
	}

	.status-indicator.connected {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.status-indicator.disconnected {
		background: var(--error-light);
		color: var(--error-dark);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: currentColor;
	}

	.environment-badge {
		padding: var(--space-xs) var(--space-md);
		border-radius: var(--radius-md);
		font-size: 0.8rem;
		font-weight: bold;
	}

	.environment-badge.test {
		background: var(--warning);
		color: white;
	}

	.environment-badge.live {
		background: var(--error);
		color: white;
	}

	.tab-navigation {
		background: var(--surface);
		border-bottom: 1px solid var(--border);
		padding: 0 var(--space-lg);
	}

	.tab-list {
		display: flex;
		gap: var(--space-xs);
		overflow-x: auto;
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.tab-list::-webkit-scrollbar {
		display: none;
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-lg);
		border: none;
		background: transparent;
		color: var(--text-muted);
		cursor: pointer;
		border-bottom: 3px solid transparent;
		transition: all 0.2s ease;
		white-space: nowrap;
		font-size: 0.95rem;
		font-weight: 500;
	}

	.tab-button:hover:not(:disabled) {
		color: var(--text);
		background: var(--bg-secondary);
	}

	.tab-button.active {
		color: var(--primary);
		border-bottom-color: var(--primary);
		background: var(--primary-light);
	}

	.tab-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.tab-icon {
		font-size: 1.1rem;
	}

	.tab-name {
		font-weight: inherit;
	}

	.tab-lock {
		font-size: 0.8rem;
		opacity: 0.7;
	}

	.tab-restricted {
		font-size: 0.8rem;
		opacity: 0.8;
		color: var(--warning);
	}

	.tab-button.restricted {
		opacity: 0.7;
		border-left: 3px solid var(--warning);
	}

	.key-type-indicator {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-sm) var(--space-md);
		margin-top: var(--space-sm);
		border-top: 1px solid var(--border);
	}

	.key-badge {
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-md);
		font-size: 0.8rem;
		font-weight: 600;
		margin-bottom: var(--space-xs);
	}

	.key-badge.full {
		background: var(--success-light);
		color: var(--success-dark);
	}

	.key-badge.restricted {
		background: var(--warning-light);
		color: var(--warning-dark);
	}

	.key-type-indicator small {
		color: var(--text-muted);
		font-size: 0.75rem;
	}

	.tab-content {
		flex: 1;
		min-height: calc(100vh - 200px);
		background: var(--bg-primary);
	}

	.restricted-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 50vh;
	}

	.restricted-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.8;
	}

	.restricted-content h2 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.8rem;
		font-weight: 600;
	}

	.restricted-content p {
		margin: 0 0 var(--space-lg) 0;
		color: var(--text-muted);
		font-size: 1.1rem;
		max-width: 600px;
		line-height: 1.5;
	}

	.restricted-details {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		margin-bottom: var(--space-lg);
		text-align: left;
		max-width: 600px;
		width: 100%;
	}

	.restricted-details h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.restricted-details ul {
		margin: 0;
		padding-left: var(--space-lg);
		color: var(--text);
	}

	.restricted-details li {
		margin-bottom: var(--space-sm);
		line-height: 1.5;
	}

	.restricted-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: center;
		flex-wrap: wrap;
	}



	/* New styles for setup container */
	.setup-container {
		padding: var(--space-lg);
		background: var(--surface);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-md);
		margin-top: var(--space-lg);
		border: 1px solid var(--border);
	}

	.setup-header {
		text-align: center;
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-md);
		border-bottom: 1px solid var(--border);
	}

	.setup-header h1 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 2rem;
		font-weight: 700;
	}

	.setup-header p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.setup-section {
		display: flex;
		gap: var(--space-lg);
		flex-wrap: wrap;
		justify-content: center;
	}

	.setup-card {
		flex: 1;
		min-width: 350px;
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		border: 1px solid var(--border);
	}

	.card-header {
		text-align: center;
		margin-bottom: var(--space-lg);
		padding-bottom: var(--space-md);
		border-bottom: 1px solid var(--border);
	}

	.card-header h2 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.8rem;
		font-weight: 700;
	}

	.card-header p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1rem;
	}

	.setup-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.input-group {
		position: relative;
	}

	.input-label {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		font-size: 0.9rem;
		font-weight: 500;
		color: var(--text);
		margin-bottom: var(--space-xs);
	}

	.input-label .required {
		color: var(--error);
		font-size: 0.8rem;
	}

	.input {
		padding: var(--space-md) var(--space-lg);
		border: 1px solid var(--border);
		border-radius: var(--radius-md);
		background: var(--bg-input);
		color: var(--text);
		font-size: 1rem;
		transition: all 0.2s ease;
		width: 100%;
	}

	.input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: var(--shadow-sm);
	}

	.input-help {
		font-size: 0.8rem;
		color: var(--text-muted);
		margin-top: var(--space-xs);
	}

	.btn-lg {
		padding: var(--space-md) var(--space-lg);
		font-size: 1rem;
	}

	.alert {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md);
		border-radius: var(--radius-md);
		margin-top: var(--space-md);
	}

	.alert-success {
		background-color: var(--success-light);
		color: var(--success-dark);
	}

	.alert-error {
		background-color: var(--error-light);
		color: var(--error-dark);
	}

	.alert-icon {
		font-size: 1.2rem;
	}

	.alert-content {
		flex: 1;
	}

	.alert-content strong {
		font-weight: 600;
	}

	.alert-content p {
		margin: 0;
		font-size: 0.9rem;
	}

	.instructions-card {
		flex: 1;
		min-width: 350px;
		padding: var(--space-lg);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
		border: 1px solid var(--border);
	}

	.instructions-card h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
		font-weight: 700;
	}

	.steps {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.step {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
	}

	.step-number {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
		min-width: 20px;
		text-align: center;
	}

	.step-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.step-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1rem;
	}

	.security-notice {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		margin-top: var(--space-lg);
	}

	.notice-icon {
		font-size: 1.5rem;
		color: var(--primary);
	}

	.notice-content h4 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.2rem;
		font-weight: 600;
	}

	.notice-content ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.notice-content li {
		margin-bottom: var(--space-xs);
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.portal-form {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		margin-top: var(--space-lg);
	}

	.portal-form-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
	}

	.saved-portal {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
		border: 1px solid var(--border);
	}

	.saved-portal-display {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.saved-portal-label {
		font-size: 0.9rem;
		color: var(--text-muted);
		font-weight: 500;
	}

	.saved-portal-url {
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.saved-portal-url a {
		color: var(--primary);
		text-decoration: none;
		word-break: break-all;
	}

	.saved-portal-url a:hover {
		text-decoration: underline;
	}

	.saved-portal-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: flex-end;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.6);
		display: flex;
		justify-content: center;
		align-items: center;
		z-index: 1000;
	}

	.modal-content {
		background: var(--bg-primary);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-xl);
		width: 90%;
		max-width: 600px;
		max-height: 90vh;
		display: flex;
		flex-direction: column;
		border: 1px solid var(--border);
		overflow: hidden;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: var(--space-md) var(--space-lg);
		border-bottom: 1px solid var(--border);
		background: var(--surface);
	}

	.modal-header h3 {
		margin: 0;
		color: var(--text);
		font-size: 1.8rem;
		font-weight: 700;
	}

	.modal-close {
		background: none;
		border: none;
		font-size: 2rem;
		color: var(--text-muted);
		cursor: pointer;
		padding: var(--space-xs);
		transition: color 0.2s ease;
	}

	.modal-close:hover {
		color: var(--text);
	}

	.modal-body {
		padding: var(--space-lg);
		overflow-y: auto;
		flex-grow: 1;
	}

	.confirmation-input {
		margin-top: var(--space-md);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-md) var(--space-lg);
		border-top: 1px solid var(--border);
		background: var(--surface);
	}

	.btn-secondary {
		background: var(--bg-secondary);
		color: var(--text);
		border: 1px solid var(--border);
	}

	.btn-secondary:hover {
		background: var(--bg-hover);
	}

	.btn-danger {
		background: var(--error);
		color: white;
	}

	.btn-danger:hover {
		background: var(--error-dark);
	}

	@media (max-width: 768px) {
		.dashboard-header {
			flex-direction: column;
			gap: var(--space-md);
			text-align: center;
		}



		.tab-list {
			justify-content: flex-start;
		}

		.tab-button {
			padding: var(--space-sm) var(--space-md);
			font-size: 0.9rem;
		}

		.setup-section {
			flex-direction: column;
			align-items: center;
		}

		.setup-card,
		.instructions-card,
		.security-notice {
			width: 100%;
			min-width: auto;
		}


	}
</style> 
