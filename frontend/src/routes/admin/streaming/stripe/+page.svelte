<script lang="ts">
	// @ts-nocheck
	import { onMount } from 'svelte';
	import { apiRequest } from '$lib/auth';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { showToast } from '$lib/toast';

	// Import child components
	import AnalyticsOverview from './overview/AnalyticsOverview.svelte';
	import MetadataHealth from './metadata/MetadataHealth.svelte';
	import Setup from './setup/+page.svelte';
	import SimpleSync from '../simple-sync/+page.svelte';
	import EmailUsagePanel from './components/EmailUsagePanel.svelte';

	// State variables using Svelte 5 runes
	let summary = $state<any>(null);
	let loading = $state(true);
	let error = $state('');
	let activeTab = $state('analytics');
	let loadingStatus = $state('Initializing...');
	let keyType = $state<'sk' | 'rk' | null>(null);
	let dataTransferActive = $state(false);
	let lastDataTransfer = $state<Date | null>(null);
	let lastActivity = $state(0);
	let activityTimeout = $state<number | undefined>(undefined);
	let maxTimeoutId = $state<number | undefined>(undefined);
	let progressInterval = $state<number | undefined>(undefined);
	let debugMode = $state(false);
	let debugInfo = $state({
		backendUrl: '',
		apiBaseUrl: '',
		frontendHost: '',
		frontendProtocol: ''
	});

	// NEW: Two-state pattern for Stripe data
	let stripeDataIncoming = $state({
		customers: [],
		subscriptions: [],
		products: [],
		prices: [],
		invoices: [],
		paymentIntents: [],
		coupons: [],
		lastUpdated: null,
		isLoading: true,
		progress: {
			total: 7,
			completed: 0,
			current: 'Initializing...'
		}
	});

	let stripeData = $state({
		customers: [],
		subscriptions: [],
		products: [],
		prices: [],
		invoices: [],
		paymentIntents: [],
		coupons: [],
		lastUpdated: null,
		isLoading: false,
		progress: {
			total: 7,
			completed: 0,
			current: 'Initializing...'
		}
	});

	// Add a state for tracking the current loading phase
	let currentLoadingPhase = $state('initializing');

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
		console.log('StripeDataIncoming progress:', stripeDataIncoming.progress);
		console.log('StripeData ready:', !stripeData.isLoading);
		console.log('========================');
	});

	// Tab configuration - dynamically filtered based on capabilities
	const allTabs = [
		{ id: 'analytics', name: 'Analytics', icon: '📈', component: AnalyticsOverview, capability: null },
		{ id: 'metadata', name: 'Metadata Health', icon: '🏥', component: MetadataHealth, capability: null },
		{ id: 'setup', name: 'Setup', icon: '⚙️', component: Setup, capability: null },
		{ id: 'simple-sync', name: 'Simple Sync', icon: '🔄', component: SimpleSync, capability: null }
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

	// Add this test function to debug the apiRequest
	async function testApiRequest() {
		try {
			console.log(' Testing apiRequest function...');
			const res = await apiRequest('/health'); // Remove /api/v1 prefix
			console.log('✅ apiRequest test successful:', res.status);
		} catch (err) {
			console.error('❌ apiRequest test failed:', err);
		}
	}

	// Call this in onMount to test
	onMount(async () => {
		//await testApiRequest();
		try {
			// Load debug info
			const { getApiBaseUrl, apiBaseUrl } = await import('$lib/config');
			debugInfo.backendUrl = getApiBaseUrl();
			debugInfo.apiBaseUrl = apiBaseUrl;
			debugInfo.frontendHost = window.location.hostname;
			debugInfo.frontendProtocol = window.location.protocol;
			
			console.log('🚀 Initializing Stripe dashboard...');
			console.log('🔧 Debug info loaded:', debugInfo);
			
			// 🎯 EXPLICIT STATUS CHECK: Use dedicated status endpoint first (no assumptions)
			console.log('🔍 [ONMOUNT] Checking Stripe configuration status via /stripe/status...');
			console.log('🔍 [ONMOUNT] This is the onMount function - NOT triggered by manual sync');
			
			// First, check explicit status (always returns 200 OK with clear info)
			const statusRes = await apiRequest('/admin/streaming/stripe/status');
			
			if (statusRes.ok) {
				const statusData = await statusRes.json();
				console.log('✅ [STATUS] Explicit status received:', statusData);
				
				if (statusData.configured) {
					// Stripe is configured - now get the dashboard data
					console.log('🚀 [DASH] Stripe configured - fetching dashboard data...');
					const dashRes = await apiRequest('/admin/streaming/stripe/dash');
					
					if (dashRes.ok) {
						const dashData = await dashRes.json();
						if (dashData.enabled) {
							// 🚀 DASH MODE: Use the fast data we already have!
							console.log('🎯 Stripe configured - using lightning-fast dash data!');
							console.log('⚡ No need for slow comprehensive loading - dash has everything!');
							
							// Use the dash data directly instead of calling loadAllStripeData
							summary = {
								enabled: true,
								...dashData
							};
							
							// Process the dash data into our frontend structure
							await processStripeData(summary);
							
							// Set loading to false since we're done
							loading = false;
							loadingStatus = 'Dashboard ready via dash!';
						} else {
							// ⚙️ Stripe not configured - skip to setup (no loading needed)
							console.log('⚙️ Stripe not configured - showing setup screen');
							loading = false;
						}
					} else {
						// Dashboard call failed - but we know Stripe is configured
						console.log('⚠️ Stripe configured but dashboard failed - trying fallback...');
						await loadAllStripeData();
					}
				} else {
					// Stripe explicitly not configured
					console.log('⚙️ [STATUS] Stripe not configured - showing setup screen');
					summary = { enabled: false };
					loading = false;
				}
			} else {
				// Status endpoint failed - fallback to old method
				console.log('⚠️ [STATUS] Status endpoint failed - falling back to /stripe/dash...');
				const dashRes = await apiRequest('/admin/streaming/stripe/dash');
				
				if (dashRes.ok) {
				const dashData = await dashRes.json();
				if (dashData.enabled) {
					// 🚀 DASH MODE: Use the fast data we already have!
					console.log('🎯 Stripe configured - using lightning-fast dash data!');
					console.log('⚡ No need for slow comprehensive loading - dash has everything!');
					
					// Use the dash data directly instead of calling loadAllStripeData
					summary = {
						enabled: true,
						...dashData
					};
					
					// Process the dash data into our frontend structure
					await processStripeData(summary);
					
					// Set loading to false since we're done
					loading = false;
					loadingStatus = 'Dashboard ready via dash!';
					currentLoadingPhase = 'complete';
					
					console.log('🎉 Dashboard ready using fast dash data!');
				} else {
					// ⚙️ Stripe not configured - skip to setup (no loading needed)
					console.log('⚙️ Stripe not configured - showing setup screen');
					loading = false;
				}
			} else {
				// Check if it's a 503 (Service Unavailable) which means Stripe is not configured
				if (dashRes.status === 503) {
					console.log('⚙️ Stripe service unavailable (not configured) - showing setup screen');
					summary = { enabled: false };
					loading = false;
				} else if (dashRes.status === 504) {
					// 504 Gateway Timeout - could be Stripe disabled or slow API
					console.log('⏰ Gateway timeout (504) - assuming Stripe not configured, showing setup screen');
					console.log('💡 If Stripe IS configured, this indicates a performance issue that needs investigation');
					summary = { enabled: false };
					loading = false;
				} else {
					// ❌ Other API error - show error state
					console.log('❌ Failed to check Stripe status - showing error state');
					error = 'Failed to connect to Stripe service';
					loading = false;
				}
			}
			}
			
		await loadPortalLink();
		
			// Always default to analytics tab for fast access
			activeTab = 'analytics';
			console.log('✅ Analytics tab set as default for optimal user experience');
			
		} catch (err) {
			console.error('❌ Failed to initialize Stripe dashboard:', err);
			error = 'Failed to initialize Stripe dashboard';
			loading = false;
		}
	});

	// 🚀 V2 ANALYTICS: Lightning-fast analytics refresh using the v2 endpoint
	async function loadAnalyticsData() {
		console.log("🔄 Loading v2 analytics data...")
		try {
			const response = await apiRequest('/admin/streaming/stripe/v2/analytics')
			const data = await response.json()
			console.log("📊 V2 Analytics data received:", data)
			
			if (data && data.enabled) {
				console.log("✅ Stripe is enabled, processing v2 analytics...")
				
				// Log each section of v2 data
				if (data.balance) {
					console.log("💰 Balance data:", data.balance)
				}
				if (data.customer_analytics) {
					console.log("👥 Customer analytics:", data.customer_analytics)
				}
				if (data.subscription_health) {
					console.log("📋 Subscription health:", data.subscription_health)
				}
				if (data.mrr_analytics) {
					console.log("📈 MRR analytics:", data.mrr_analytics)
				}
				if (data.revenue_analytics) {
					console.log("💳 Revenue analytics:", data.revenue_analytics)
				}
				if (data.product_performance) {
					console.log("📦 Product performance:", data.product_performance)
				}
				if (data.payment_analytics) {
					console.log("💰 Payment analytics:", data.payment_analytics)
				}
				
				// Process the v2 data
				processStripeDataV2(data)
			} else {
				console.log("❌ Stripe not enabled or no data received")
			}
		} catch (error) {
			console.error("❌ Failed to load analytics data:", error)
		}
	}

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
			
			console.error('🔍 Detailed error analysis:', {
				name: err.name,
				message: err.message,
				stack: err.stack,
				cause: err.cause,
				type: typeof err
			});
			
			if (err.name === 'AbortError') {
				error = 'Request was aborted. This could be due to a network issue or manual cancellation.';
				console.error('🚫 Stripe summary request was aborted');
			} else if (err.message?.includes('fetch')) {
				error = 'Network error. Please check your connection and try again.';
				console.error('🌐 Network error details:', err);
				
				// Add more specific network error details
				if (err.message.includes('Failed to fetch')) {
					console.error('🔍 Failed to fetch - possible causes:');
					console.error('  - Network connectivity issue');
					console.error('  - CORS problem');
					console.error('  - Invalid URL');
					console.error('  - Request being blocked by browser/firewall');
				}
			} else {
				error = err.message || 'Failed to load Stripe data';
				console.error('❌ Unexpected error:', err);
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
						console.log('🌐 Request URL will be:', `${baseUrl}/api/v1/admin/streaming/stripe/dash`);
						
						// Add timestamp for request tracking
						const requestTimestamp = new Date().toISOString();
						console.log('⏰ Request timestamp:', requestTimestamp);
						
						// Test with detailed request monitoring
						const controller = new AbortController();
						const startTime = Date.now();
						
						console.log('📡 Sending request to backend...');
						const stripeTest = await apiRequest('/admin/streaming/stripe/dash', {
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
	// NOTE: This is no longer used with the single-call approach
	// Keeping it for reference but it's not called anymore
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
			}
			
			throw new Error(`Failed to fetch section ${section}: ${res.status}`);
		} catch (err) {
			console.error(`❌ Error fetching section ${section}:`, err);
			throw err;
		}
	}

	// NEW: Three-Phase Data Loading Implementation
	// NOTE: Phase 1 is now handled by fetchStripeSummary + processStripeData
	// Keeping Phase 2 and 3 for the atomic transfer pattern
	
	async function loadStripeDataPhase2() {
		console.log('🎯 Phase 2: Transferring data to presentation state...');
		
		// Atomic transfer of all data
		stripeData.products = [...stripeDataIncoming.products];
		stripeData.prices = [...stripeDataIncoming.prices];
		stripeData.customers = [...stripeDataIncoming.customers];
		stripeData.subscriptions = [...stripeDataIncoming.subscriptions];
		stripeData.paymentIntents = [...stripeDataIncoming.paymentIntents];
		stripeData.invoices = [...stripeDataIncoming.invoices];
		stripeData.coupons = [...stripeDataIncoming.coupons];
		
		// Update progress and status
		stripeData.isLoading = false;
		stripeData.progress.completed = 7; // Add this line
		stripeData.progress.current = 'Data transfer complete'; // Add this line
		stripeData.lastUpdated = new Date();
		
		console.log('✅ Phase 2 complete: Data transferred to presentation state');
	}

	async function loadStripeDataPhase3() {
		console.log('🎯 Phase 3: Finalizing dashboard...');
		
		// Update summary with collected data
		if (summary) {
			summary.customers = stripeData.customers;
			summary.subscriptions = stripeData.subscriptions;
			summary.products = stripeData.products;
			summary.prices = stripeData.prices;
			summary.invoices = stripeData.invoices;
			summary.payment_intents = stripeData.paymentIntents;
			summary.coupons = stripeData.coupons;
		}
		
		console.log('✅ Phase 3 complete: Dashboard ready');
		return true;
	}

	// Individual data loading functions with progress tracking
	// NOTE: These are no longer used with the single-call approach
	// Keeping them for reference but they're not called anymore
	
	async function loadProducts() {
		try {
			updateProgress('Loading products...', 0);
			const data = await fetchStripeSection('products', 100);
			stripeDataIncoming.products = data.products || [];
			updateProgress('Products loaded', 1);
			console.log(`📦 Products loaded: ${stripeDataIncoming.products.length} items`);
		} catch (err) {
			console.error('❌ Failed to load products:', err);
			stripeDataIncoming.products = [];
			updateProgress('Products failed', 1);
		}
	}

	async function loadPrices() {
		try {
			updateProgress('Loading prices...', 1);
			const data = await fetchStripeSection('prices', 100);
			stripeDataIncoming.prices = data.prices || [];
			updateProgress('Prices loaded', 2);
			console.log(`💰 Prices loaded: ${stripeDataIncoming.prices.length} items`);
		} catch (err) {
			console.error('❌ Failed to load prices:', err);
			stripeDataIncoming.prices = [];
			updateProgress('Prices failed', 2);
		}
	}

	async function loadCustomers() {
		try {
			updateProgress('Loading customers...', 2);
			const data = await fetchStripeSection('customers', 100);
			stripeDataIncoming.customers = data.customers || [];
			updateProgress('Customers loaded', 3);
			console.log(`👥 Customers loaded: ${stripeDataIncoming.customers.length} items`);
		} catch (err) {
			console.error('❌ Failed to load customers:', err);
			stripeDataIncoming.customers = [];
			updateProgress('Customers failed', 3);
		}
	}

	async function loadSubscriptions() {
		try {
			updateProgress('Loading subscriptions...', 3);
			const data = await fetchStripeSection('subscriptions', 100);
			stripeDataIncoming.subscriptions = data.subscriptions || [];
			updateProgress('Subscriptions loaded', 4);
			console.log(`📋 Subscriptions loaded: ${stripeDataIncoming.subscriptions.length} items`);
		} catch (err) {
			console.error('❌ Failed to load subscriptions:', err);
			stripeDataIncoming.subscriptions = [];
			updateProgress('Subscriptions failed', 4);
		}
	}

	async function loadPaymentIntents() {
		try {
			updateProgress('Loading payment intents...', 4);
			const data = await fetchStripeSection('payment_intents', 100);
			stripeDataIncoming.paymentIntents = data.payment_intents || [];
			updateProgress('Payment intents loaded', 5);
			console.log(`💳 Payment intents loaded: ${stripeDataIncoming.paymentIntents.length} items`);
		} catch (err) {
			console.error('❌ Failed to load payment intents:', err);
			stripeDataIncoming.paymentIntents = [];
			updateProgress('Payment intents failed', 5);
		}
	}

	async function loadInvoices() {
		try {
			updateProgress('Loading invoices...', 5);
			const data = await fetchStripeSection('invoices', 100);
			stripeDataIncoming.invoices = data.invoices || [];
			updateProgress('Invoices loaded', 6);
			console.log(`🧾 Invoices loaded: ${stripeDataIncoming.invoices.length} items`);
		} catch (err) {
			console.error('❌ Failed to load invoices:', err);
			stripeDataIncoming.invoices = [];
			updateProgress('Invoices failed', 6);
		}
	}

	async function loadCoupons() {
		try {
			updateProgress('Loading coupons...', 6);
			const data = await fetchStripeSection('coupons', 100);
			stripeDataIncoming.coupons = data.coupons || [];
			updateProgress('Coupons loaded', 7);
			console.log(`🎟️ Coupons loaded: ${stripeDataIncoming.coupons.length} items`);
		} catch (err) {
			console.error('❌ Failed to load coupons:', err);
			stripeDataIncoming.coupons = [];
			updateProgress('Coupons failed', 7);
		}
	}

	function updateProgress(current: string, completed: number) {
		stripeDataIncoming.progress.current = current;
		stripeDataIncoming.progress.completed = completed;
		console.log(`📊 Progress: ${completed}/${stripeDataIncoming.progress.total} - ${current}`);
	}

	// Main data loading orchestrator
	async function loadAllStripeData() {
		try {
			console.log('🚀 Starting three-phase Stripe data loading...');
			
			// Set loading state
			loading = true;
			loadingStatus = 'Starting three-phase data loading...';
			currentLoadingPhase = 'fetching';
			
			// Phase 1: Single backend call that gets everything
			console.log('📡 Phase 1: Fetching all Stripe data in single call...');
			loadingStatus = 'Fetching all Stripe data...';
			const fetchedSummary = await fetchStripeSummary();
			
			// CRITICAL FIX: Update the global summary variable
			summary = fetchedSummary;
			console.log('🔧 Global summary updated:', summary);
			console.log('🔧 Summary enabled:', summary?.enabled);
			
			// Phase 2: Process the data into our structure
			console.log('🔄 Phase 2: Processing data into frontend structure...');
			currentLoadingPhase = 'processing';
			loadingStatus = 'Processing data...';
			await processStripeData(fetchedSummary);
			
			// Phase 3: Transfer to presentation state
			console.log('🎯 Phase 3: Transferring to presentation state...');
			currentLoadingPhase = 'transferring';
			loadingStatus = 'Transferring data...';
			await loadStripeDataPhase2();
			
			// Phase 4: Finalize dashboard
			currentLoadingPhase = 'finalizing';
			loadingStatus = 'Finalizing dashboard...';
			await loadStripeDataPhase3();
			
			// Complete loading
			loading = false;
			loadingStatus = 'Dashboard ready!';
			currentLoadingPhase = 'complete';
			
			console.log('🎉 All phases complete! Dashboard ready.');
			return true;
		} catch (err) {
			console.error('❌ Failed to load Stripe data:', err);
			error = 'Failed to load Stripe data: ' + (err instanceof Error ? err.message : String(err));
			loading = false;
			currentLoadingPhase = 'error';
			return false;
		}
	}

	async function fetchStripeSummary() {
		console.log('📡 Fetching complete Stripe summary...');
		try {
			// Fix: Remove the duplicate /api/v1 prefix since apiRequest adds it
			const endpoint = '/admin/streaming/stripe/summary?limit=100';
			console.log('🔗 Endpoint:', endpoint);
			console.log('🌐 Full URL will be:', `${debugInfo.apiBaseUrl}${endpoint}`);
			
			const res = await apiRequest(endpoint);
			console.log('📡 Response received:', res.status, res.statusText);
			
			if (res.ok) {
				const data = await res.json();
				console.log('✅ Stripe summary fetched successfully:', data.summary);
				return data.summary;
			} else {
				console.error('❌ Response not OK:', res.status, res.statusText);
				const errorText = await res.text();
				console.error('❌ Error response body:', errorText);
				throw new Error(`Failed to fetch Stripe summary: ${res.status} - ${res.statusText}`);
			}
		} catch (err) {
			console.error('❌ Exception in fetchStripeSummary:', err);
			console.error('🔍 Error details:', {
				name: err.name,
				message: err.message,
				cause: err.cause
			});
			throw err;
		}
	}

	async function processStripeData(summary: any) {
		console.log('🔄 Processing Stripe data into frontend structure...');
		stripeDataIncoming.isLoading = true;
		stripeDataIncoming.progress.current = 'Processing data...';

		// CRITICAL FIX: Detect environment from the key type
		if (summary && !summary.environment) {
			// Determine environment from the key type
			if (keyType === 'sk' || keyType === 'rk') {
				// Check if the key contains 'live' to determine environment
				summary.environment = secret.includes('live') ? 'live' : 'test';
				console.log('🔧 Environment detected:', summary.environment, 'from key type:', keyType);
			}
		}

		stripeDataIncoming.products = summary.products || [];
		stripeDataIncoming.prices = summary.prices || [];
		stripeDataIncoming.customers = summary.customers || [];
		stripeDataIncoming.subscriptions = summary.subscriptions || [];
		stripeDataIncoming.paymentIntents = summary.payment_intents || [];
		stripeDataIncoming.invoices = summary.invoices || [];
		stripeDataIncoming.coupons = summary.coupons || [];

		// Update both progress objects
		stripeDataIncoming.progress.completed = 7;
		stripeData.progress.completed = 7;
		stripeDataIncoming.progress.current = 'All data processed';
		stripeData.progress.current = 'All data processed';

		console.log('✅ Data processing complete:', {
			products: stripeDataIncoming.products.length,
			prices: stripeDataIncoming.prices.length,
			customers: stripeDataIncoming.customers.length,
			subscriptions: stripeDataIncoming.subscriptions.length,
			paymentIntents: stripeDataIncoming.paymentIntents.length,
			invoices: stripeDataIncoming.invoices.length,
			coupons: stripeDataIncoming.coupons.length
		});
	}

	// Process v2 analytics data into frontend structure
	async function processStripeDataV2(data: any) {
		console.log('🔄 Processing v2 analytics data into frontend structure...');
		
		// Update summary with v2 analytics data
		summary = {
			enabled: true,
			environment: data.environment || 'test',
			...data
		};

		// Convert v2 analytics to legacy format for existing components
		stripeDataIncoming.isLoading = false;
		stripeDataIncoming.progress.current = 'V2 analytics processed';
		stripeDataIncoming.progress.completed = 7;

		// Map v2 data to existing structure
		if (data.customer_analytics) {
			stripeDataIncoming.customers = Array(data.customer_analytics.total_customers || 0).fill({
				id: 'sample',
				email: 'sample@example.com',
				name: 'Sample Customer'
			});
		}

		if (data.subscription_health) {
			stripeDataIncoming.subscriptions = Array(data.subscription_health.active_subscriptions || 0).fill({
				id: 'sample',
				status: 'active'
			});
		}

		if (data.product_performance) {
			stripeDataIncoming.products = Array(data.product_performance.active_products || 0).fill({
				id: 'sample',
				name: 'Sample Product',
				active: true
			});
			
			stripeDataIncoming.prices = Array(data.product_performance.active_prices || 0).fill({
				id: 'sample',
				unit_amount: 1000,
				currency: 'usd'
			});
		}

		// Set other arrays to empty for now
		stripeDataIncoming.paymentIntents = [];
		stripeDataIncoming.invoices = [];
		stripeDataIncoming.coupons = [];

		console.log('✅ V2 analytics processing complete:', {
			balance: data.balance?.available_usd,
			customers: data.customer_analytics?.total_customers,
			subscriptions: data.subscription_health?.active_subscriptions,
			products: data.product_performance?.active_products,
			mrr: data.mrr_analytics?.estimated_mrr,
			fetchTime: data.total_fetch_time
		});
	}

	// Function to handle loading state transitions
	function setLoadingState(isLoading: boolean, status: string) {
		loading = isLoading;
		loadingStatus = status;
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
				
				// 🚀 DASH MODE: Use lightning-fast dashboard endpoint after key save
				console.log('🚀 [KEY-SAVE] Key saved - using fast dash endpoint for immediate data...');
				console.log('🚀 [KEY-SAVE] This is the key save function - triggered by user saving key');
				const dashRes = await apiRequest('/admin/streaming/stripe/dash');
				
				if (dashRes.ok) {
					const dashData = await dashRes.json();
					if (dashData.enabled) {
						// Use the dash data directly
						summary = {
							enabled: true,
							...dashData
						};
						
						// Process the dash data into our frontend structure
						await processStripeData(summary);
						
						console.log('🎉 Dashboard loaded via fast dash after key save!');
					}
				}
				
				// Show success message
				setupSuccess = 'Stripe connected successfully! Dashboard loaded.';
				
				// Clear success message after 3 seconds
				setTimeout(() => {
					setupSuccess = '';
				}, 3000);
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
			console.log('🔄 Sending clear key request to backend...');
			const res = await apiRequest('/admin/streaming/stripe/secret', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ key: 'sk_1337' })
			});
			
			console.log('📡 Clear key response status:', res.status);
			console.log('📡 Clear key response ok:', res.ok);
			
			if (res.ok) {
				const responseData = await res.json();
				console.log('✅ Clear key response data:', responseData);
				// Show toast notification
				showToast('Stripe key cleared successfully!', 'success');
				
				console.log('🔄 Clear key successful, resetting dashboard state...');
				
				// CRITICAL: Set summary to disabled state to trigger line 334 condition
				summary = { enabled: false };
				
				// Switch back to analytics tab
				activeTab = 'analytics';
				
				// Clear any success messages
				setupSuccess = '';
				setupError = '';
				
				// Also clear the portal link since Stripe is no longer configured
				savedPortalLink = '';
				
				// Reset all Stripe data state
				stripeData = {
					customers: [],
					subscriptions: [],
					products: [],
					prices: [],
					invoices: [],
					paymentIntents: [],
					coupons: [],
					lastUpdated: null,
					isLoading: false,
					progress: {
						total: 7,
						completed: 0,
						current: 'Initializing...'
					}
				};
				
				// Reset loading states
				loading = false;
				error = '';
				keyType = null;
				
				// No forced reload - let the reactive state handle the UI updates
				console.log('✅ Key cleared - state updated, no reload needed');
				
				console.log('✅ Summary reset to disabled state:', summary);
				console.log('✅ Active tab switched to analytics:', activeTab);
				console.log('✅ Portal link cleared along with Stripe key');
			} else {
				const errorData = await res.json();
				console.error('❌ Clear key failed with status:', res.status);
				console.error('❌ Clear key error data:', errorData);
				setupError = errorData.error || 'Failed to clear key';
				showToast('Failed to clear Stripe key', 'error');
			}
		} catch (err) {
			console.error('❌ Clear key request failed with exception:', err);
			setupError = 'Failed to clear key: ' + (err.message || 'Unknown error');
			showToast('Failed to clear Stripe key', 'error');
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

	// Add this function to test our new analytics endpoints
	async function testAnalyticsEndpoints() {
		console.log(" Testing new Stripe Analytics endpoints...")
		
		try {
			// Test the fast dashboard endpoint
			console.log("📊 Testing /stripe/dash endpoint...")
			const dashResponse = await apiRequest('/admin/streaming/stripe/dash')
			console.log("✅ /stripe/dash Response:", dashResponse)
			
			// Test the comprehensive analytics endpoint
			console.log("📈 Testing /stripe/analytics endpoint...")
			const analyticsResponse = await apiRequest('/admin/streaming/stripe/analytics')
			console.log("✅ /stripe/analytics Response:", analyticsResponse)
			
			// Test individual analytics methods
			console.log(" Testing individual analytics methods...")
			
			// Test subscription metrics
			const subscriptionMetrics = await apiRequest('/admin/streaming/stripe/subscription-metrics')
			console.log("📋 Subscription Metrics:", subscriptionMetrics)
			
			// Test customer analytics
			const customerAnalytics = await apiRequest('/admin/streaming/stripe/customer-analytics')
			console.log(" Customer Analytics:", customerAnalytics)
			
			// Test revenue analytics
			const revenueAnalytics = await apiRequest('/admin/streaming/stripe/revenue-analytics')
			console.log("💰 Revenue Analytics:", revenueAnalytics)
			
			console.log("🎉 All analytics endpoints tested successfully!")
			
		} catch (error) {
			console.error("❌ Analytics endpoint test failed:", error)
		}
	}

	// Add this to your onMount or create a test button
	$effect(() => {
		if (stripeData.enabled) {
			console.log(" Stripe is enabled, testing analytics endpoints...")
			// Wait a bit for the initial load to complete
			setTimeout(() => {
				testAnalyticsEndpoints()
			}, 2000)
		}
	})
</script>

{#if loading}
	<div class="loading-container">
		<div class="spinner"></div>
		<p class="loading-title">Loading Stripe Dashboard...</p>
		<p class="loading-status">{loadingStatus}</p>
		
		<!-- NEW: Three-Phase Progress Indicator -->
		{#if stripeDataIncoming.isLoading}
			<div class="phase-progress">
				<div class="progress-header">
					<h3>🚀 Loading Stripe Data</h3>
					<p class="progress-status">
						{#if currentLoadingPhase === 'fetching'}
							📡 Fetching all data from Stripe...
						{:else if currentLoadingPhase === 'processing'}
							🔄 Processing data into frontend structure...
						{:else if currentLoadingPhase === 'transferring'}
							🎯 Transferring to presentation state...
						{:else if currentLoadingPhase === 'finalizing'}
							✨ Finalizing dashboard...
						{:else}
							{stripeDataIncoming.progress.current}
						{/if}
					</p>
				</div>
				
				<div class="progress-bar">
					<div class="progress-fill" style="width: {(stripeDataIncoming.progress.completed / stripeDataIncoming.progress.total) * 100}%"></div>
				</div>
				
				<div class="progress-details">
					<span class="progress-text">
						{#if currentLoadingPhase === 'fetching'}
							Fetching data from Stripe API...
						{:else if currentLoadingPhase === 'processing'}
							Processing {stripeDataIncoming.products.length + stripeDataIncoming.prices.length + stripeDataIncoming.customers.length + stripeDataIncoming.subscriptions.length + stripeDataIncoming.paymentIntents.length + stripeDataIncoming.invoices.length + stripeDataIncoming.coupons.length} items...
						{:else if currentLoadingPhase === 'transferring'}
							Transferring data to presentation...
						{:else if currentLoadingPhase === 'finalizing'}
							Preparing dashboard...
						{:else}
							{stripeDataIncoming.progress.completed} of {stripeDataIncoming.progress.total} sections complete
						{/if}
					</span>
					<span class="progress-percentage">
						{#if currentLoadingPhase === 'fetching'}
							⏳
						{:else if currentLoadingPhase === 'processing'}
							🔄
						{:else if currentLoadingPhase === 'transferring'}
							🎯
						{:else if currentLoadingPhase === 'finalizing'}
							✨
						{:else}
							{Math.round((stripeDataIncoming.progress.completed / stripeDataIncoming.progress.total) * 100)}%
						{/if}
					</span>
				</div>
				
				<div class="phase-info">
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 1 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 1}
								✅
							{:else}
								📦
							{/if}
						</span>
						<span class="phase-name">Products</span>
					</div>
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 2 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 2}
								✅
							{:else}
								💰
							{/if}
						</span>
						<span class="phase-name">Prices</span>
					</div>
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 3 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 3}
								✅
							{:else}
								👥
							{/if}
						</span>
						<span class="phase-name">Customers</span>
					</div>
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 4 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 4}
								✅
							{:else}
								📋
							{/if}
						</span>
						<span class="phase-name">Subscriptions</span>
					</div>
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 5 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 5}
								✅
							{:else}
								💳
							{/if}
						</span>
						<span class="phase-name">Payments</span>
					</div>
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 6 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 6}
								✅
							{:else}
								🧾
							{/if}
						</span>
						<span class="phase-name">Invoices</span>
					</div>
					<div class="phase-item {currentLoadingPhase === 'fetching' ? 'active' : (stripeDataIncoming.progress.completed >= 7 ? 'complete' : '')}">
						<span class="phase-icon">
							{#if currentLoadingPhase === 'fetching'}
								📡
							{:else if stripeDataIncoming.progress.completed >= 7}
								✅
							{:else}
								🎟️
							{/if}
						</span>
						<span class="phase-name">Coupons</span>
					</div>
				</div>
			</div>
		{/if}
		
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
			<button class="btn btn-primary" onclick={loadAnalyticsData} disabled={loading}>
				{loading ? '⚡ Retrying...' : '⚡ Fast Retry'}
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
					
					
					<!-- 🔄 SMART REFRESH: Different refresh strategies for different needs -->
					{#if stripeData.lastUpdated}
						<div class="data-freshness">
							<span class="freshness-icon">🕒</span>
							<span class="freshness-text">Last updated: {stripeData.lastUpdated.toLocaleTimeString()}</span>
							
							<!-- 🚀 Quick Analytics Refresh (for Analytics tab) -->
							{#if activeTab === 'analytics'}
								<button 
									class="refresh-btn analytics-refresh" 
									onclick={loadAnalyticsData}
									disabled={loading}
									title="Quick analytics refresh"
								>
									{loading ? '⚡' : '⚡'}
								</button>
							{:else}
								<!-- 🚀 Fast Data Refresh (for other tabs) -->
								<button 
									class="refresh-btn full-refresh" 
									onclick={loadAnalyticsData}
									disabled={loading}
									title="Fast data refresh via dash endpoint"
								>
									{loading ? '⚡' : '⚡'}
								</button>
							{/if}
					</div>
					{/if}
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
							class:disabled={!summary?.enabled && tab.id !== 'analytics' && tab.id !== 'setup' && tab.id !== 'simple-sync'}
							class:restricted={isRestricted}
							onclick={() => switchTab(tab.id)}
							disabled={!summary?.enabled && tab.id !== 'analytics' && tab.id !== 'setup' && tab.id !== 'simple-sync'}
							title={isRestricted ? 'This functionality is restricted by your Stripe key permissions' : ''}
						>
							<span class="tab-icon">{tab.icon}</span>
							<span class="tab-name">{tab.name}</span>
							{#if !summary?.enabled && tab.id !== 'analytics' && tab.id !== 'setup' && tab.id !== 'simple-sync'}
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
				{#if activeTab === 'analytics'}
					<AnalyticsOverview {summary} {stripeData} />
				{:else if activeTab === 'metadata'}
					<MetadataHealth />
				{:else if activeTab === 'setup'}
					<Setup {summary} onClearKey={() => showClearModal = true} />
				{:else if activeTab === 'simple-sync'}
					<SimpleSync />
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
		transition: all 0.3s ease;
	}

	.status-indicator.connected {
		background: var(--success);
		color: white;
		box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
		animation: pulse-green 2s ease-in-out infinite;
	}

	.status-indicator.disconnected {
		background: var(--error);
		color: white;
		box-shadow: 0 2px 8px rgba(239, 68, 68, 0.3);
	}

	.status-dot {
		width: 10px;
		height: 10px;
		border-radius: 50%;
		background: currentColor;
		animation: pulse-dot 2s ease-in-out infinite;
	}

	.status-indicator.connected .status-dot {
		background: white;
		box-shadow: 0 0 10px rgba(255, 255, 255, 0.8);
	}

	.status-indicator.disconnected .status-dot {
		background: white;
		opacity: 0.9;
	}


	/* Enhanced animations for connected status */
	@keyframes pulse-green {
		0%, 100% {
			box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
			transform: scale(1);
		}
		50% {
			box-shadow: 0 4px 16px rgba(16, 185, 129, 0.5);
			transform: scale(1.02);
		}
	}

	@keyframes pulse-dot {
		0%, 100% {
			opacity: 1;
			transform: scale(1);
		}
		50% {
			opacity: 0.8;
			transform: scale(1.1);
		}
	}

	/* Hover effects for better interactivity */
	.status-indicator.connected:hover {
		background: var(--success-dark);
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);
	}

	.status-indicator.disconnected:hover {
		background: var(--error-dark);
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(239, 68, 68, 0.4);
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
		font-size: 1.1rem;
		font-weight: 600;
	}

	.step-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 0.9rem;
		line-height: 1.4;
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

	/* NEW: Three-Phase Progress Indicator Styles */
	.phase-progress {
		background: var(--bg-secondary);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		margin: var(--space-lg) 0;
		max-width: 900px;
		margin-left: auto;
		margin-right: auto;
	}

	.progress-header {
		text-align: center;
		margin-bottom: var(--space-lg);
	}

	.progress-header h3 {
		margin: 0 0 var(--space-sm) 0;
		color: var(--text);
		font-size: 1.5rem;
		font-weight: 600;
	}

	.progress-status {
		margin: 0;
		color: var(--text-muted);
		font-size: 1rem;
	}

	.progress-bar {
		width: 100%;
		height: 12px;
		background: var(--bg-input);
		border-radius: 6px;
		overflow: hidden;
		margin-bottom: var(--space-md);
	}

	.progress-fill {
		height: 100%;
		background: linear-gradient(90deg, var(--primary), var(--primary-light));
		transition: width 0.3s ease;
		border-radius: 6px;
	}

	.progress-details {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-lg);
	}

	.progress-text {
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.progress-percentage {
		color: var(--primary);
		font-weight: 600;
		font-size: 1.1rem;
	}

	.phase-info {
		display:flex;
		flex-wrap: wrap;
		gap: var(--space-sm);
	}

	.phase-item {
		display: flex;
			flex-direction: column;
			align-items: center;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		background: var(--bg-input);
		border: 1px solid var(--border);
		transition: all 0.2s ease;
	}

	.phase-item.complete {
		background: var(--success-light);
		border-color: var(--success);
		color: var(--success-dark);
	}

	.phase-item.active {
		background: var(--primary-light);
		border-color: var(--primary);
		color: var(--primary-dark);
		animation: pulse 2s ease-in-out infinite;
	}

	@keyframes pulse {
		0% {
			transform: scale(1);
			opacity: 1;
		}
		50% {
			transform: scale(1.05);
			opacity: 0.8;
		}
		100% {
			transform: scale(1);
			opacity: 1;
		}
	}

	.phase-icon {
		font-size: 1.5rem;
		margin-bottom: var(--space-xs);
	}

	.phase-name {
		font-size: 0.8rem;
		font-weight: 500;
		text-align: center;
	}

	.data-freshness {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		margin-top: var(--space-sm);
		color: var(--text-muted);
		font-size: 0.9rem;
	}

	.freshness-icon {
		font-size: 1.2rem;
		color: var(--primary);
	}

	.freshness-text {
		font-weight: 600;
	}

	.refresh-btn {
		background: none;
		border: none;
		font-size: 1.2rem;
		color: var(--text-muted);
		cursor: pointer;
		margin-left: var(--space-sm);
		transition: color 0.2s ease;
	}

	.refresh-btn:hover {
		color: var(--text);
	}

	/* 🎨 BEAST MODE REFRESH BUTTON STYLES */
	.refresh-btn.analytics-refresh {
		background: linear-gradient(45deg, #10b981, #059669);
		color: white;
		border-radius: 50%;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1rem;
		transition: all 0.3s ease;
		box-shadow: 0 2px 8px rgba(16, 185, 129, 0.3);
	}

	.refresh-btn.analytics-refresh:hover:not(:disabled) {
		transform: scale(1.1) rotate(180deg);
		box-shadow: 0 4px 16px rgba(16, 185, 129, 0.5);
	}

	.refresh-btn.full-refresh {
		background: linear-gradient(45deg, #2563eb, #1d4ed8);
		color: white;
		border-radius: 50%;
		width: 32px;
		height: 32px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 1rem;
		transition: all 0.3s ease;
		box-shadow: 0 2px 8px rgba(37, 99, 235, 0.3);
	}

	.refresh-btn.full-refresh:hover:not(:disabled) {
		transform: scale(1.1) rotate(360deg);
		box-shadow: 0 4px 16px rgba(37, 99, 235, 0.5);
	}

	.refresh-btn:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none !important;
	}

	.test-section {
		margin: 1rem 0;
		padding: 1rem;
		background: #f8f9fa;
		border-radius: 8px;
		border: 1px solid #dee2e6;
	}

	.test-button {
		background: #007bff;
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 4px;
		cursor: pointer;
		font-size: 0.9rem;
	}

	.test-button:hover {
		background: #0056b3;
	}
</style> 

<!-- Add this button somewhere in your template for testing -->
{#if stripeData.enabled}
	<div class="test-section">
		<button 
			onclick={testAnalyticsEndpoints}
			class="test-button"
		>
			 Test Analytics Endpoints
		</button>
	</div>
{/if}

