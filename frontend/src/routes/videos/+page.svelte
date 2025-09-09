<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { replaceState } from '$app/navigation';
	import { videoService, type Video, type VideoCategory, type VideosResponse, type BunnyCollection } from '$lib/video';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import { toastStore } from '$lib/stores/toast';
	import AdDisplay from '$lib/components/AdDisplay.svelte';
	import SubscriptionCheck from '$lib/components/SubscriptionCheck.svelte';
	import { auth, isAdmin } from '$lib/auth';
	import { goto } from '$app/navigation';

	let videos = $state<Video[]>([]);
	let latestVideos = $state<Video[]>([]);
	let searchResults = $state<Video[]>([]);
	let allSearchResults = $state<Video[]>([]); // Store all search results for client-side filtering
	let collections = $state<BunnyCollection[]>([]);
	let categories = $state<VideoCategory[]>([]);
	let loading = $state(false);
	let error = $state('');
	let searchQuery = $state('');
	let selectedCategory = $state('');
	let currentPage = $state(1);
	let hasMore = $state(true);
	let loadingMore = $state(false);
	let authChecking = $state(true);
	let initialDataLoaded = $state(false);
	let activeTab = $state<'latest' | 'collections' | 'categories' | 'allVideos'>('categories');

	// Add reactive logging for state changes
	$effect(() => {
		console.log('🔄 Loading state changed:', loading, 'at', new Date().toISOString());
	});

	$effect(() => {
		console.log('📊 InitialDataLoaded state changed:', initialDataLoaded, 'at', new Date().toISOString());
	});

	// Test effect to see if reactivity is working at all
	$effect(() => {
		console.log('🧪 Test effect running - categories length:', categories.length);
	});
	let scrollThreshold = 800; // pixels from bottom to trigger auto-load (accounts for footer height)
	let searchTimeout: NodeJS.Timeout | null = null;
	let isSearching = $state(false);

	// Set active tab from URL parameter using Svelte 5 $effect
	$effect(() => {
		const tabParam = $page.url.searchParams.get('tab');
		if (tabParam && ['latest', 'collections', 'categories', 'allVideos'].includes(tabParam)) {
			activeTab = tabParam as typeof activeTab;
		}
	});

	// Client-side search filtering function
	function clientSideFilter(videoList: Video[], query: string): Video[] {
		if (!query.trim()) return videoList;
		
		const searchTerms = query.toLowerCase().trim().split(' ').filter(term => term.length > 0);
		
		return videoList.filter(video => {
			const searchableText = [
				video.title,
				video.description,
				video.category,
				...(video.tags || [])
			].join(' ').toLowerCase();
			
			// Check if ANY search term is found in the video's searchable text (more permissive)
			return searchTerms.some(term => searchableText.includes(term));
		});
	}

	// Handle real-time search with client-side filtering using Svelte 5 $effect
	$effect(() => {
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}
		searchTimeout = setTimeout(() => {
			if (searchQuery.length > 0) {
				handleSearch();
			} else if (searchQuery.length === 0) {
				// Clear search when query is empty
				clearSearch();
			}
		}, 300); // 300ms debounce
	});

	// Get the current video list based on tab and search state with client-side filtering
	let currentVideos = $derived(searchQuery.length > 0 ? 
		clientSideFilter(allSearchResults, searchQuery) : 
		(activeTab === 'latest' ? latestVideos : videos));

	// Update searchResults when allSearchResults or searchQuery changes using Svelte 5 $effect
	$effect(() => {
		if (searchQuery.length > 0) {
			searchResults = clientSideFilter(allSearchResults, searchQuery);
		} else {
			searchResults = [];
		}
	});

	onMount(() => {
		console.log('🔄 Videos page mounted');
		console.log('📊 Cache check on mount:', {
			categoriesLength: categories.length,
			videosLength: videos.length,
			initialDataLoaded,
			loading
		});
		
		// Simple: If we have cached data, show it immediately
		if (categories.length > 0 || videos.length > 0) {
			console.log('✅ Using cached data on mount - setting loading=false, initialDataLoaded=true');
			console.log('🔍 BEFORE: loading =', loading, 'initialDataLoaded =', initialDataLoaded);
			console.trace('📍 STACK TRACE: onMount setting loading=false (cached data)');
			loading = false;
			initialDataLoaded = true;
			console.log('🔍 AFTER: loading =', loading, 'initialDataLoaded =', initialDataLoaded);
		} else {
			console.log('❌ No cached data on mount - waiting for auth');
		}
		
		// Add global debug function to window for console access
		if (typeof window !== 'undefined') {
			(window as any).debugVideoSearch = () => {
				console.log('🔍 Video Search Debug Info:', {
					searchQuery,
					isSearching,
					totalVideos: videos.length,
					latestVideos: latestVideos.length,
					allSearchResults: allSearchResults.length,
					filteredSearchResults: searchResults.length,
					currentVideos: currentVideos.length,
					activeTab,
					hasMore,
					currentPage
				});
				
				if (searchQuery) {
					console.log('🔍 Search Query Analysis:', {
						originalQuery: searchQuery,
						searchTerms: searchQuery.toLowerCase().trim().split(' ').filter(term => term.length > 0),
						allResults: allSearchResults.map(v => ({ id: v.id, title: v.title })),
						filteredResults: searchResults.map(v => ({ id: v.id, title: v.title }))
					});
				}
			};
		}
		
		// Add scroll listener for infinite scroll
		const handleScroll = () => {
			// Only auto-load if we're on a tab that supports it, have more content, and not searching
			if ((activeTab === 'latest' || activeTab === 'allVideos') && hasMore && !loadingMore && !isSearching) {
				const scrollTop = window.pageYOffset || document.documentElement.scrollTop;
				const windowHeight = window.innerHeight;
				const documentHeight = document.documentElement.scrollHeight;
				
				// Check if we're within the threshold of the bottom
				if (scrollTop + windowHeight >= documentHeight - scrollThreshold) {
					console.log('🚀 Infinite scroll triggered!', {
						scrollTop,
						windowHeight,
						documentHeight,
						threshold: scrollThreshold,
						hasMore,
						loadingMore,
						activeTab,
						isSearching,
						searchQuery
					});
					loadMore();
				}
			}
		};

		// Add throttled scroll listener
		let scrollTimer: NodeJS.Timeout | null = null;
		const throttledScroll = () => {
			if (scrollTimer) return;
			scrollTimer = setTimeout(() => {
				handleScroll();
				scrollTimer = null;
			}, 100);
		};

		window.addEventListener('scroll', throttledScroll);
		
		// Cleanup function
		return () => {
			window.removeEventListener('scroll', throttledScroll);
			if (scrollTimer) clearTimeout(scrollTimer);
			if (searchTimeout) clearTimeout(searchTimeout);
			// Clean up global debug function
			if (typeof window !== 'undefined') {
				delete (window as any).debugVideoSearch;
			}
		};
	});

	async function loadInitialData() {
		console.log('🔄 loadInitialData called');
		console.log('📊 loadInitialData state check:', {
			initialDataLoaded,
			loading,
			categoriesLength: categories.length,
			videosLength: videos.length
		});
		
		// Simple check: if we already have data, don't reload
		if (initialDataLoaded) {
			console.log('✅ Data already loaded - early return');
			return;
		}

		try {
			console.log('🚀 Starting loadInitialData - setting loading=true');
			console.trace('📍 STACK TRACE: loadInitialData setting loading=true');
			loading = true;
			error = '';

			// Load collections (keep for backward compatibility)
			try {
				const collectionsResponse = await videoService.getCollections();
				collections = collectionsResponse.items || [];
			} catch (err) {
				console.warn('Failed to load collections:', err);
				collections = [];
			}

			// Load tag categories
			try {
				const categoriesResponse = await videoService.getTagCategories();
				categories = categoriesResponse.categories || [];
				console.log('✅ Loaded categories:', categories.length);
			} catch (err) {
				console.error('❌ Failed to load categories:', err);
				categories = [];
			}

			// Load latest videos
			try {
				const response = await videoService.getVideos(1, 6); // Get latest 6 videos
				latestVideos = response.videos || [];
			} catch (err) {
				console.warn('Failed to load latest videos:', err);
				latestVideos = [];
			}

			// Load regular videos
			await loadVideos();
			initialDataLoaded = true;
			console.log('✅ loadInitialData completed successfully');
		} catch (err: any) {
			console.error('❌ loadInitialData failed:', err);
			handleError(err);
		} finally {
			console.log('🏁 loadInitialData finished, setting loading=false');
			console.trace('📍 STACK TRACE: loadInitialData finally block setting loading=false');
			loading = false;
		}
	}

	function handleError(err: any) {
		console.error('Error:', err);
		if (err.error_type === 'authentication_error') {
			error = 'Authentication required. Please log in.';
			toastStore.error('Please log in to view videos.');
		} else if (err.error_type === 'network_error') {
			error = 'Network error. Please check your connection.';
			toastStore.error('Network error. Please check your connection and try again.');
		} else {
			error = err.message || 'An error occurred';
			toastStore.error(error);
		}
	}

	async function loadVideos(reset = false) {
		try {
			if (reset) {
				currentPage = 1;
				if (!isSearching) {
					videos = [];
				}
			}

			console.log('🎬 Loading videos:', {
				page: currentPage,
				category: selectedCategory,
				search: searchQuery,
				isSearching,
				reset
			});

			const response: VideosResponse = await videoService.getVideos(
				currentPage,
				20,
				selectedCategory || undefined,
				isSearching ? searchQuery || undefined : undefined
			);

			console.log('📥 API Response:', {
				videosCount: response.videos?.length || 0,
				//hasMore: response.pagination?.has_more || false,
				//currentPage: response.pagination?.current_page || 1
			});

			const newVideos = response.videos || [];
			
			// Validate that search results actually match the query
			if (isSearching && searchQuery.trim()) {
				const validatedResults = clientSideFilter(newVideos, searchQuery);
				console.log('✅ Search validation:', {
					serverResults: newVideos.length,
					validatedResults: validatedResults.length,
					query: searchQuery
				});
				
				// Use validated results
				const finalResults = validatedResults.length > 0 ? validatedResults : newVideos;
				
				if (reset) {
					allSearchResults = finalResults;
				} else {
					// For search results, prevent duplicates
					const existingIds = new Set(allSearchResults.map(v => v.id));
					const uniqueNewVideos = finalResults.filter(video => !existingIds.has(video.id));
					allSearchResults = [...allSearchResults, ...uniqueNewVideos];
				}
			} else {
				// Regular video loading (not search)
				if (reset) {
					if (activeTab === 'latest') {
						latestVideos = newVideos;
					} else {
						videos = newVideos;
					}
				} else {
					const targetArray = activeTab === 'latest' ? latestVideos : videos;
					const existingIds = new Set(targetArray.map(v => v.id));
					const uniqueNewVideos = newVideos.filter(video => !existingIds.has(video.id));
					
					if (activeTab === 'latest') {
						latestVideos = [...latestVideos, ...uniqueNewVideos];
					} else {
						videos = [...videos, ...uniqueNewVideos];
					}
				}
			}

			// Update pagination info
			//hasMore = response.pagination?.has_more || false;
			console.log('📊 Load complete:', {
				videosLoaded: newVideos.length,
				hasMore,
				totalInArray: isSearching ? allSearchResults.length : 
					(activeTab === 'latest' ? latestVideos.length : videos.length)
			});
			
			// Clear any previous errors on success
			error = '';
		} catch (err: any) {
			console.error('❌ Error loading videos:', err);
			
			// Provide more specific error messages
			if (err.error_type === 'authentication_error') {
				error = 'Authentication required. Please log in.';
				toastStore.error('Please log in to view videos.');
			} else if (err.error_type === 'network_error') {
				error = 'Network error. Please check your connection.';
				toastStore.error('Network error. Please check your connection and try again.');
			} else if (err.status === 429) {
				error = 'Too many requests. Please wait a moment.';
				toastStore.error('Too many requests. Please wait a moment before trying again.');
			} else {
				error = err.message || 'Failed to load videos';
				toastStore.error('Failed to load videos. Please try again.');
			}
		}
	}

	async function handleSearch() {
		if (searchQuery.trim().length === 0) {
			clearSearch();
			return;
		}

		console.log('🔍 Starting search for:', searchQuery);
		isSearching = true;
		currentPage = 1;
		
		// Clear previous search results
		searchResults = [];
		allSearchResults = [];
		
		try {
			// Load all search results at once (higher limit for comprehensive search)
			const response: VideosResponse = await videoService.getVideos(
				1,
				100, // Load more results for search
				undefined,
				searchQuery
			);

			const newVideos = response.videos || [];
			
			// Validate and store search results
			const validatedResults = clientSideFilter(newVideos, searchQuery);
			allSearchResults = validatedResults.length > 0 ? validatedResults : newVideos;
			
			// If we have very few results, also search in the local cache
			if (allSearchResults.length < 10) {
				console.log('🔄 Expanding search with local results');
				const localResults = [
					...videos,
					...latestVideos
				];
				
				// Remove duplicates and add to search results
				const existingIds = new Set(allSearchResults.map(v => v.id));
				const additionalResults = localResults.filter(video => 
					!existingIds.has(video.id) && 
					clientSideFilter([video], searchQuery).length > 0
				);
				
				allSearchResults = [...allSearchResults, ...additionalResults];
			}
			
			// For search, we don't use pagination - show all results
			hasMore = false;
			
			console.log('📊 Search complete:', {
				query: searchQuery,
				totalResults: allSearchResults.length
			});
			
		} catch (err) {
			console.error('❌ Search failed:', err);
			handleError(err);
		} finally {
			// Always stop the search spinner when search completes
			isSearching = false;
		}
	}

	function clearSearch() {
		console.log('🧹 Clearing search');
		isSearching = false;
		searchResults = [];
		allSearchResults = [];
		currentPage = 1;
		
		// Restore pagination for regular browsing
		hasMore = true;
		
		// Don't reload if we already have data for the current tab
		if ((activeTab === 'allVideos' && videos.length === 0) || 
		    (activeTab === 'latest' && latestVideos.length === 0)) {
			loadVideos(true);
		}
	}

	function handleClearSearch() {
		searchQuery = '';
		clearSearch();
	}

	// Debug function to show all search results (including non-matching)
	function showAllSearchResults() {
		if (allSearchResults.length > 0) {
			console.log('🔍 All search results for "' + searchQuery + '":', allSearchResults);
			console.log('✅ Filtered results:', searchResults);
			console.log('📊 Search terms:', searchQuery.toLowerCase().trim().split(' '));
		}
	}

	// Debug function to test client-side filtering
	function debugClientSideFilter(video: Video) {
		const searchTerms = searchQuery.toLowerCase().trim().split(' ').filter(term => term.length > 0);
		const searchableText = [
			video.title,
			video.description,
			video.category,
			...(video.tags || [])
		].join(' ').toLowerCase();
		
		console.log('🔍 Debug filter for:', video.title, {
			searchTerms,
			searchableText,
			matches: searchTerms.some(term => searchableText.includes(term)) // Changed from every to some
		});
		
		return searchTerms.some(term => searchableText.includes(term)); // Changed from every to some
	}

	async function handleCategoryChange() {
		console.log('Filtering by category:', selectedCategory);
		// Don't reset initialDataLoaded - just reload videos for the selected category
		await loadVideos(true);
	}

	async function loadMore() {
		if (loadingMore || !hasMore) return;

		try {
			loadingMore = true;
			currentPage++;
			await loadVideos();
		} finally {
			loadingMore = false;
		}
	}

	function clearFilters() {
		searchQuery = '';
		selectedCategory = '';
		clearSearch();
		// Don't reset initialDataLoaded - just reload videos without filters
		loadVideos(true);
	}

	function handleAuthLoadingChange(data: {loading: boolean}) {
		console.log('🔐 Auth loading change:', data.loading);
		authChecking = data.loading;
	}

	function handleAccessGranted() {
		console.log('✅ Access granted!');
		console.log('📊 Cache check on access granted:', {
			categoriesLength: categories.length,
			videosLength: videos.length,
			initialDataLoaded,
			loading
		});
		
		// Simple: Check cache first, then load if needed
		if (categories.length > 0 || videos.length > 0) {
			console.log('✅ Using cached data - setting loading=false, initialDataLoaded=true');
			console.log('🔍 BEFORE: loading =', loading, 'initialDataLoaded =', initialDataLoaded);
			console.trace('📍 STACK TRACE: handleAccessGranted setting loading=false (cached data)');
			loading = false;
			initialDataLoaded = true;
			console.log('🔍 AFTER: loading =', loading, 'initialDataLoaded =', initialDataLoaded);
		} else {
			console.log('🔄 No cache, loading data');
			loadInitialData();
		}
	}

	function switchTab(tab: typeof activeTab) {
		activeTab = tab;
		
		// Update URL to reflect the active tab using SvelteKit navigation
		const url = new URL($page.url);
		url.searchParams.set('tab', tab);
		replaceState(url.toString(), {});
		
		if (tab === 'latest' || tab === 'allVideos') {
			// Don't clear search when switching between video tabs
			selectedCategory = '';
			
			// If we're not searching, load appropriate data for the tab
			if (!isSearching) {
				// Check if we need to load data for this tab
				if ((tab === 'allVideos' && videos.length === 0) || 
				    (tab === 'latest' && latestVideos.length === 0)) {
					loadVideos(true);
				}
			}
		} else {
			// Clear search when going to collections or categories
			if (searchQuery) {
				searchQuery = '';
				clearSearch();
			}
		}
	}

	// Category section state for lazy loading - using Svelte 5 reactive approach
	let categoryVideoStates = $state(new Map());
	let categoryStateVersion = $state(0); // Force reactivity trigger
	
	function getCategoryVideoState(categoryId: number) {
		if (!categoryVideoStates.has(categoryId)) {
			categoryVideoStates.set(categoryId, {
				loading: false,
				data: null,
				error: null,
				promise: null
			});
		}
		return categoryVideoStates.get(categoryId);
	}
	
	async function loadCategoryVideos(category: VideoCategory) {
		const state = getCategoryVideoState(category.id);
		
		// Return cached result if available
		if (state.data) {
			console.log(`🔄 Returning cached result for category: ${category.name}`);
			return state.data;
		}

		// Return existing promise if already loading
		if (state.promise) {
			console.log(`🔄 Returning existing promise for category: ${category.name}`);
			return state.promise;
		}

		console.log(`🎬 Loading videos for category: ${category.name} (ID: ${category.id})`);
		console.log(`🔍 Full category object:`, category);
		console.log(`🏷️ Category has ${category.tagIds?.length || 0} tag IDs:`, category.tagIds);
		console.log(`🏷️ Raw tag_ids (if exists):`, (category as any).tag_ids);
		
		// Set loading state
		console.log(`🔄 Setting state.loading=true for category: ${category.name}`);
		console.trace('📍 STACK TRACE: loadCategoryVideos setting state.loading=true');
		state.loading = true;
		state.error = null;
		
		// Create and cache the loading promise
		const loadingPromise = videoService.getVideosByTagCategory(category.id, 1, 10)
			.then(response => {
				console.log(`✅ Loaded ${response.videos?.length || 0} videos for category: ${category.name}`);
				console.log(`📊 Response structure:`, { 
					hasVideos: !!response.videos, 
					videoCount: response.videos?.length,
					hasPagination: !!response.pagination 
				});
				
				// Log the actual videos retrieved
				if (response.videos && response.videos.length > 0) {
					console.log(`🎬 Videos for category "${category.name}":`, response.videos.map(video => ({
						id: video.id,
						title: video.title,
						description: video.description?.substring(0, 100) + '...',
						thumbnailUrl: video.thumbnailUrl,
						duration: video.duration,
						tags: video.tags
					})));
				} else {
					console.log(`❌ No videos found for category "${category.name}"`);
				}
				
				// Log pagination details
				if (response.pagination) {
					console.log(`📄 Pagination for "${category.name}":`, response.pagination);
				}
				
				// Update state
				state.loading = false;
				state.data = response;
				state.promise = null;
				
				// Force reactivity by creating new Map reference and incrementing version
				categoryVideoStates = new Map(categoryVideoStates);
				categoryStateVersion++;
				
				console.log(`🔄 Updated state for ${category.name}, forcing reactivity (version: ${categoryStateVersion})`);
				
				return response;
			})
			.catch(error => {
				console.error(`❌ Failed to load videos for category: ${category.name}`, error);
				
				// Update error state
				state.loading = false;
				state.error = error;
				state.promise = null;
				
				// Force reactivity by creating new Map reference and incrementing version
				categoryVideoStates = new Map(categoryVideoStates);
				categoryStateVersion++;
				
				throw error;
			});

		// Cache the loading promise
		state.promise = loadingPromise;
		
		// Force reactivity by creating new Map reference
		categoryVideoStates = new Map(categoryVideoStates);
		
		return loadingPromise;
	}
	
	// Simple category loading when tab becomes active
	$effect(() => {
		if (categories.length > 0 && activeTab === 'categories') {
			categories.forEach(category => {
				const state = getCategoryVideoState(category.id);
				if (!state.data && !state.loading) {
					loadCategoryVideos(category);
				}
			});
		}
	});
</script>

<svelte:head>
	<title>Video Hub - Book of Mormon Evidences</title>
	<meta name="description" content="Explore our comprehensive collection of Book of Mormon evidence videos, organized by collections, topics, and latest uploads." />
</svelte:head>

<Navigation />

<div class="page-wrapper">
	<main class="main-content-wrapper">
		<SubscriptionCheck 
			redirectTo="/login" 
			requireSubscription={true}
			onLoadingChange={handleAuthLoadingChange}
			onAccessGranted={handleAccessGranted}
		>
			{#if loading}
				<div class="loading-container">
					<LoadingSpinner size="large" />
					<p>Loading video hub...</p>
				</div>
			{:else}
				<div class="video-hub">
					<div class="container">
						<header class="hub-header">
							<!--<h1>Video Hub</h1>-->
							<p>Discover our extensive collection of Book of Mormon evidence videos</p>
						</header>

						<!-- Navigation Tabs -->
						<div class="hub-tabs">
								
							
							<button 
								class="tab-button {activeTab === 'categories' ? 'active' : ''}" 
								onclick={() => switchTab('categories')}
							>
								Categories
							</button>
						
							<button
								class="tab-button {activeTab === 'allVideos' ? 'active' : ''}"
								onclick={() => switchTab('allVideos')}
							>
								All Videos
							</button>
							<button 
								class="tab-button {activeTab === 'latest' ? 'active' : ''}" 
								onclick={() => switchTab('latest')}
							>
								Latest Videos
							</button>
							<button 
								class="tab-button {activeTab === 'collections' ? 'active' : ''}" 
								onclick={() => switchTab('collections')}
							>Collections
							</button>
						</div>

						<!-- All Videos Section -->
						{#if activeTab === 'allVideos'}
						
							<section class="all-videos">
								<h2>Book of Mormon Evidence Videos</h2>
								<div class="filters-section">
									<div class="search-bar">
										<input
											type="text"
											placeholder="Search videos..."
											bind:value={searchQuery}
										/>
										{#if searchQuery}
											<button class="btn-clear" onclick={handleClearSearch} title="Clear search">
												✕
											</button>
										{/if}
										<button class="btn-primary" onclick={handleSearch}>
											🔍 Search
										</button>
									</div>
									{#if searchQuery}
										<div class="search-info">
											{#if isSearching}
												<div class="search-loading">
													<LoadingSpinner size="small" />
													<p>Searching for "{searchQuery}"...</p>
												</div>
											{:else if searchResults.length > 0}
												<p>{searchResults.length} video{searchResults.length !== 1 ? 's' : ''} found</p>
											{:else if searchQuery && !isSearching}
												<p>No videos found for "{searchQuery}"</p>
											{/if}
										</div>
									{/if}
								</div>

								{#if currentVideos.length === 0 && !loading && !loadingMore && !isSearching}
									<div class="no-results">
										<p>No videos found{searchQuery ? ` for "${searchQuery}"` : ''}.</p>
										{#if searchQuery}
											<button class="btn-secondary" onclick={handleClearSearch}>
												Clear Search
											</button>
										{/if}
									</div>
								{:else}
									<div class="video-grid">
										{#each currentVideos as video (video.id)}
											<VideoCard {video} />
										{/each}
									</div>

									{#if hasMore && !isSearching}
										<div class="load-more">
											<button class="btn-secondary" onclick={loadMore} disabled={loadingMore}>
												{loadingMore ? 'Loading...' : 'Load More (or keep scrolling)'}
											</button>
										</div>
									{/if}
								{/if}
							</section>
						{/if}


						<!-- Latest Videos Section -->
						{#if activeTab === 'latest'}
							<section class="latest-videos">
								<h2>Latest Uploads</h2>
								<div class="filters-section">
									<div class="search-bar">
										<input
											type="text"
											placeholder="Search videos..."
											bind:value={searchQuery}
										/>
										{#if searchQuery}
											<button class="btn-clear" onclick={handleClearSearch} title="Clear search">
												✕
											</button>
										{/if}
										<button class="btn-primary" onclick={handleSearch}>
											🔍 Search
										</button>
									</div>
									{#if searchQuery}
										<div class="search-info">
											{#if isSearching}
												<div class="search-loading">
													<LoadingSpinner size="small" />
													<p>Searching for "{searchQuery}"...</p>
												</div>
											{:else if searchResults.length > 0}
												<p>{searchResults.length} video{searchResults.length !== 1 ? 's' : ''} found</p>
											{:else if searchQuery && !isSearching}
												<p>No videos found for "{searchQuery}"</p>
											{/if}
										</div>
									{/if}
								</div>

								{#if currentVideos.length === 0 && !loading && !loadingMore && !isSearching}
									<div class="no-results">
										<p>No videos found{searchQuery ? ` for "${searchQuery}"` : ''}.</p>
										{#if searchQuery}
											<button class="btn-secondary" onclick={handleClearSearch}>
												Clear Search
											</button>
										{/if}
									</div>
								{:else}
									<div class="video-grid">
										{#each currentVideos as video (video.id)}
											<VideoCard {video} />
										{/each}
									</div>

									{#if hasMore && !isSearching}
										<div class="load-more">
											<button class="btn-secondary" onclick={loadMore} disabled={loadingMore}>
												{loadingMore ? 'Loading...' : 'Load More (or keep scrolling)'}
											</button>
										</div>
									{/if}
								{/if}
							</section>
						{/if}

						<!-- Collections Section -->
						{#if activeTab === 'collections'}
							<section class="collections">
								<h2>Video Collections</h2>
								<div class="collections-grid">
									{#each collections as collection (collection.guid)}
										<a 
											href="/videos/collections/{collection.guid}" 
											class="collection-card"
											onclick={(e) => { e.preventDefault(); goto(`/videos/collections/${collection.guid}`); }}
										>
											<div class="collection-image">
												<img 
													src="/HOMEPAGE_TEST_ASSETS/16X10_WORDLESS_Collection_Placeholder_IMG.webp" 
													alt={collection.name}
													loading="lazy"
												/>
												<div class="collection-overlay">
													<span class="video-count">{collection.videoCount} videos</span>
												</div>
											</div>
											<div class="collection-content">
												<h3>{collection.name}</h3>
												<p class="collection-description">
													Explore this collection of Book of Mormon evidence videos
												</p>
											</div>
										</a>
									{/each}
								</div>
							</section>
						{/if}

						<!-- Categories Section -->
						{#if activeTab === 'categories'}
							<section class="categories">
								<h2>Browse by Category</h2>
								<!--<div class="debug-info">
									<p><strong>Debug Info:</strong></p>
									<p>Categories loaded: {categories.length}</p>
									<p>Active tab: {activeTab}</p>
									<p>Categories array: {JSON.stringify(categories.map(c => ({id: c.id, name: c.name, tagIds: c.tagIds})), null, 2)}</p>
								</div>-->
								<div class="categories-container">
									{#each categories as category (category.id)}
										{@const categoryState = categoryVideoStates.get(category.id) || { loading: false, data: null, error: null, promise: null }}
										<!-- Debug: Log what template sees -->
										{console.log(`🎭 Template sees for ${category.name}:`, {
											hasData: !!categoryState.data,
											isLoading: categoryState.loading,
											hasError: !!categoryState.error,
											condition1: !categoryState.data && !categoryState.loading && !categoryState.error,
											condition2: categoryState.loading,
											condition3: categoryState.error,
											condition4: categoryState.data
										})}
										<div class="category-section debug-category" data-category-id={category.id}>
											<div class="category-header">
												<h3>{category.name}</h3>
												<!--<p class="category-description">{category.description}</p>
												<div class="category-stats">
													<span class="tag-count">{(category.tagIds?.length || (category as any).tag_ids?.length || 0)} tags</span>
													<span class="video-count">
														{#if categoryState.data}
															{categoryState.data.videos?.length || 0} videos loaded
														{:else if categoryState.loading}
															Loading videos...
														{:else}
															Click to load videos
														{/if}
													</span>
												</div>-->
											</div>
											
											{#if !categoryState.data && !categoryState.loading && !categoryState.error}
												<!-- Initial state - trigger loading -->
												<div class="category-loading">
													<LoadingSpinner size="small" />
													<p>Loading videos for {category.name}...</p>
												</div>
											{:else if categoryState.loading}
												<div class="category-loading">
													<LoadingSpinner size="small" />
													<p>Loading videos for {category.name}...</p>
												</div>
											{:else if categoryState.error}
												<div class="category-error">
													<p>Failed to load videos for {category.name}.</p>
													<p class="error-details">{categoryState.error.message || 'Unknown error'}</p>
													<button onclick={() => {
														// Reset state and retry
														categoryState.loading = false;
														categoryState.data = null;
														categoryState.error = null;
														categoryState.promise = null;
														categoryVideoStates = new Map(categoryVideoStates);
														loadCategoryVideos(category);
													}}>
														Retry
													</button>
												</div>
											{:else if categoryState.data}
												{#if categoryState.data.videos && categoryState.data.videos.length > 0}
													<div class="carousel-wrapper">
														<!-- Manual scroll buttons for fallback -->
														<button 
															class="scroll-btn scroll-btn-prev"
															onclick={(e) => {
																const target = e.target as HTMLElement;
																const carousel = target?.closest('.carousel-wrapper')?.querySelector('.modern-carousel') as HTMLElement;
																carousel?.scrollBy({ left: -748, behavior: 'smooth' }); // 2 videos: (700px + 24px gap) * 2
															}}
															aria-label="Previous videos"
														>
															‹
														</button>
														
														<div class="modern-carousel">
															{#each categoryState.data.videos as video (video.id)}
																<VideoCard {video} />
															{/each}
															
															<!-- See More Card -->
															<div class="see-more-card" 
																onclick={() => goto(`/videos/categories/${category.id}`)} 
																onkeydown={(e) => e.key === 'Enter' && goto(`/videos/categories/${category.id}`)} 
																role="button" 
																tabindex="0">
																<div class="see-more-content">
																	<div class="see-more-icon">→</div>
																	<h4>See More</h4>
																	<p>View all videos in {category.name}</p>
																	<!--<span class="total-count">
																		{categoryState.data.pagination?.total || categoryState.data.videos.length} total videos
																	</span>-->
																</div>
															</div>
														</div>
														
														<button 
															class="scroll-btn scroll-btn-next"
															onclick={(e) => {
																const target = e.target as HTMLElement;
																const carousel = target?.closest('.carousel-wrapper')?.querySelector('.modern-carousel') as HTMLElement;
																carousel?.scrollBy({ left: 688, behavior: 'smooth' }); // 2 videos: (320px + 24px gap) * 2
															}}
															aria-label="Next videos"
														>
															›
														</button>
													</div>
												{:else}
													<div class="no-videos-message">
														<p>No videos found for this category yet.</p>
														<p class="debug-info">
															Category has {(category.tagIds?.length || (category as any).tag_ids?.length || 0)} tags: 
															{(category.tagIds?.slice(0, 5) || (category as any).tag_ids?.slice(0, 5) || []).join(', ') || 'none'}
															{((category.tagIds?.length || (category as any).tag_ids?.length || 0) > 5) ? '...' : ''}
														</p>
													</div>
												{/if}
											{/if}
										</div>
									{/each}
								</div>
							</section>
						{/if}

						{#if error}
							<div class="error-message">
								<p>{error}</p>
							</div>
						{/if}

						<!-- Ad Placement 
						<div class="ad-placement">
							<AdDisplay placement="videos-footer" />
						</div>-->
					</div>
				</div>
			{/if}
		</SubscriptionCheck>
	</main>
</div>

<Footer />

<style lang="postcss">
	.main-content-wrapper {
		margin-top: 50px;
	}

	.video-hub {
		padding: 2rem 0;
		background: var(--bg-glass);
	}

	.hub-header {
		text-align: center;
		margin-bottom: 2rem;

		h1 {
			font-size: 2.5rem;
			color: var(--color-primary);
			margin-bottom: 1rem;
		}

		p {
			font-size: 1.1rem;
			color: var(--color-text-secondary);
			margin-bottom: 1rem;
		}
	}



	.hub-tabs {
		display: flex;
		justify-content: center;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
		width: 100%;
	}

	.tab-button {
		padding: 0.75rem 1.5rem;
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		color: var(--color-text);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.2s;
		font-weight: 500;

		&:hover {
			background: var(--color-surface-hover);
			border-color: var(--color-primary);
		}

		&.active {
			background: linear-gradient(135deg, var(--primary-bom-light) 0%, var(--primary-bom-dark) 100%);
			color: var(--primary-gold-light);
			border-color: var(--color-primary);
		}
	}

	.all-videos, .latest-videos, .collections, .categories {
		margin-bottom: 2rem;
	}

	.all-videos h2, .latest-videos h2, .collections h2, .categories h2 {
		font-size: 1.8rem;
		color: var(--color-text);
		margin-bottom: 1.5rem;
		text-align: center;
	}

	.filters-section {
		display: flex;
		justify-content: center;
		margin-bottom: 2rem;
	}

	.search-bar {
		display: flex;
		gap: 0.5rem;
		max-width: 400px;
		width: 100%;
		position: relative;
	}

	.search-bar input {
		flex: 1;
		padding: 0.75rem;
		padding-right: 2.5rem; /* Make space for clear button when present */
		border: 1px solid var(--color-border);
		border-radius: 8px;
		font-size: 1rem;
		background: var(--color-surface);
		color: var(--color-text);
	}

	.search-bar input:focus {
		outline: none;
		border-color: var(--color-primary);
	}

	.btn-clear {
		position: absolute;
		right: 70px; /* Position it inside the input, accounting for search button */
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: var(--color-text-secondary);
		cursor: pointer;
		font-size: 1rem;
		padding: 0.25rem;
		transition: color 0.2s;
		z-index: 1;

		&:hover {
			color: var(--color-text);
		}
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
		width: 100%;
	}

	.collections-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
		width: 100%;
	}

	.categories-container {
		display: flex;
		flex-direction: column;
		gap: 3rem;
		margin-bottom: 2rem;
	}

	.category-section {
		
		background: var(--color-surface);
		border: 1px solid var(--color-border);
		border-radius: 16px;
		padding: 2rem;
		transition: all 0.3s ease;
	}

	.category-section:hover {
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
		border-color: var(--color-primary);
	}

	.category-header {
		margin-bottom: 1.5rem;
		text-align: center;
	}

	.category-header h3 {
		font-size: 1.8rem;
		color: var(--primary-gold-dark) !important;
		margin-bottom: 0.5rem;
		font-weight: 600;
	}

	.category-description {
		color: var(--color-text-secondary);
		font-size: 1rem;
		margin: 0;
		line-height: 1.5;
	}

	/* Carousel wrapper for positioning scroll buttons */
	.carousel-wrapper {
		position: relative;
		margin: 1rem 0;
	}

	/* Manual scroll buttons (fallback and always visible) */
	.scroll-btn {
		position: absolute;
		top: 50%;
		transform: translateY(-50%);
		width: 44px;
		height: 44px;
		border-radius: 50%;
		border: 2px solid var(--color-border);
		background: var(--primary-gold-dark);
		color: var(--white);
		font-size: 1.5rem;
		font-weight: bold;
		cursor: pointer;
		z-index: 10;
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		transition: all 0.3s ease;
		opacity: 0.8;
	}

	.scroll-btn:hover {
		background: var(--primary-gold-light);
		border-color: var(--color-primary);
		color: var(--text-primary);
		opacity: 1;
		transform: translateY(-50%) scale(1.1);
	}

	.scroll-btn-prev {
		left: -22px;
	}

	.scroll-btn-next {
		right: -22px;
	}

	/* Modern CSS Carousel Implementation */
	.modern-carousel {
		display: grid;
		grid-auto-flow: column;
		grid-auto-columns: minmax(700px, 1fr);
		gap: 1.5rem;
		overflow-x: auto;
		scroll-snap-type: x mandatory;
		scroll-behavior: smooth;
		padding: 1rem 0;
		
		/* Hide scrollbar for cleaner look */
		scrollbar-width: none;
		-ms-overflow-style: none;
	}

	.modern-carousel::-webkit-scrollbar {
		display: none;
	}

	/* Scroll snap for each item */
	.modern-carousel > :global(*) {
		scroll-snap-align: start;
		flex-shrink: 0;
	}


	.see-more-card {
		min-width: 320px;
		background: linear-gradient(135deg, var(--color-primary-light), var(--color-primary));
		border-radius: 12px;
		padding: 2rem;
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all 0.3s ease;
		min-height: 200px;
		border: 2px dashed var(--color-primary);
		color: var(--color-primary);
		scroll-snap-align: start;
	}

	.see-more-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
		background: var(--color-primary);
		color: white;
	}

	.see-more-content {
		text-align: center;
	}

	.see-more-icon {
		font-size: 2rem;
		margin-bottom: 1rem;
		font-weight: bold;
		color: var(--primary-gold-dark);
	}

	.see-more-content h4 {
		font-size: 1.2rem;
		margin: 0 0 0.5rem 0;
		font-weight: 600;
	}

	.see-more-content p {
		margin: 0 0 0.5rem 0;
		opacity: 0.8;
	}

	.total-count {
		font-size: 0.8rem;
		opacity: 0.7;
		font-weight: 500;
	}

	.category-loading {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 2rem;
		color: var(--color-text-secondary);
	}

	.category-loading p {
		margin: 0;
	}

	.no-videos-message {
		text-align: center;
		padding: 2rem;
		color: var(--color-text-secondary);
		font-style: italic;
	}

	.no-videos-message p {
		margin: 0;
	}

	.category-error {
		text-align: center;
		padding: 2rem;
		color: var(--color-error, #ef4444);
		background: rgba(239, 68, 68, 0.1);
		border-radius: 8px;
	}

	.category-error p {
		margin: 0;
	}



	.load-more {
		display: flex;
		justify-content: center;
		margin-top: 2rem;
	}

	.search-info {
		margin-top: 1rem;
		text-align: center;
		color: var(--color-text-secondary);
		font-size: 0.9rem;
	}

	.search-loading {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		margin-bottom: 0.5rem;
	}

	.search-loading p {
		margin: 0;
	}

	.no-results {
		text-align: center;
		padding: 2rem 0;
		color: var(--color-text-secondary);
		font-size: 1rem;

		p {
			margin-bottom: 1rem;
		}
	}

	.error-message {
		background: #fee;
		border: 1px solid #fcc;
		color: #c33;
		padding: 1rem;
		border-radius: 8px;
		margin: 1rem 0;
		text-align: center;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}

	.loading-container p {
		margin-top: 1rem;
		color: var(--color-text-secondary);
	}

	.ad-placement {
		margin-top: 3rem;
		text-align: center;
	}

	.container {
		width: 95vw;
		max-width: none;
		margin: 0 auto;
		padding: 0 1rem;
	}

	.btn-primary {
		background: var(--color-primary);
		color: white;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		transition: background 0.2s;

		&:hover {
			background: var(--color-primary-hover);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
	}

	.btn-secondary {
		background: var(--color-surface);
		color: var(--color-text);
		border: 1px solid var(--color-border);
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;

		&:hover {
			background: var(--color-surface-hover);
			border-color: var(--color-primary);
		}

		&:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
	}

	.collections-grid {
		display: flex;
		flex-wrap: wrap;
		flex-direction: row;
		gap: 2rem;
		margin-bottom: 2rem;
		width: 100%;
		justify-content: center;
	}

	.collection-card {
		max-height: 400px;
		max-width: 500px;
		background: var(--color-surface);
		border-radius: 16px;
		overflow: hidden;
		transition: all 0.3s ease;
		text-decoration: none;
		color: inherit;
		display: flex;
		flex-direction: column;
		height: 100%;
		border: 1px solid var(--color-border);
		
		&:hover {
			transform: translateY(-8px);
			box-shadow: 0 12px 30px rgba(0, 0, 0, 0.15);
			border-color: var(--color-primary);
			
			.collection-image img {
				transform: scale(1.05);
			}
			
			.collection-overlay {
				background: rgba(0, 0, 0, 0.4);
			}
		}
	}

	.collection-image {
		position: relative;
		aspect-ratio: 16/10;
		overflow: hidden;
		
		img {
			width: 100%;
			height: 100%;
			object-fit: cover;
			transition: transform 0.3s ease;
		}
	}

	.collection-overlay {
		position: absolute;
		inset: 0;
		background: rgba(0, 0, 0, 0.3);
		display: flex;
		align-items: flex-end;
		padding: 1.5rem;
		transition: background 0.3s ease;
	}

	.video-count {
		color: white;
		font-size: 0.9rem;
		font-weight: 500;
		padding: 0.5rem 1rem;
		background: rgba(0, 0, 0, 0.5);
		border-radius: 20px;
		backdrop-filter: blur(4px);
	}

	.collection-content {
		padding: 1.5rem;
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		
		h3 {
			font-size: 1.4rem;
			color: var(--color-text);
			font-weight: 600;
			margin: 0;
		}
		
		.collection-description {
			color: var(--color-text-secondary);
			font-size: 0.95rem;
			line-height: 1.5;
			margin: 0;
		}
	}

	@media (max-width: 768px) {
		.hub-header h1 {
			font-size: 2rem;
		}

		.hub-tabs {
			gap: 0.5rem;
		}

		.tab-button {
			padding: 0.5rem 1rem;
			font-size: 0.9rem;
		}

		.video-grid {
			grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
			gap: 1rem;
		}

		.collections-grid {
			grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
			gap: 1.5rem;
		}

		.collection-content {
			padding: 1.25rem;
			
			h3 {
				font-size: 1.2rem;
			}
		}

		.container {
			width: 98vw;
			padding: 0 0.5rem;
		}
	}

	@media (max-width: 480px) {
		.hub-header h1 {
			font-size: 1.8rem;
		}

		.hub-tabs {
			flex-direction: column;
			align-items: center;
		}

		.tab-button {
			width: 100%;
			max-width: 200px;
		}
	}

	/* Category Section Styles */
	.categories {
		margin-bottom: 3rem;
	}

	.categories h2 {
		font-size: 2rem;
		color: var(--color-text);
		margin-bottom: 2rem;
		text-align: center;
	}

	.categories-container {
		display: flex;
		flex-direction: column;
		gap: 3rem;
	}

	.category-section {
		background: var(--color-surface);
		border-radius: 12px;
		padding: 2rem;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		border: 1px solid var(--color-border);
	}

	.category-header {
		margin-bottom: 2rem;
		text-align: left;

	}

	.category-header h3 {
		font-size: 1.8rem;
		color: var(--color-primary);
		margin: 0 0 0.5rem 0;
		font-weight: 600;
	}

	.category-description {
		color: var(--color-text-secondary);
		font-size: 1rem;
		line-height: 1.5;
		margin: 0 0 1rem 0;
	}

	.category-stats {
		display: flex;
		justify-content: left;
		gap: 1rem;
		font-size: 0.9rem;
	}

	.tag-count, .video-count {
		background: var(--color-primary-light);
		color: var(--color-primary);
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-weight: 500;
	}

	.category-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		padding: 2rem;
		color: var(--color-text-secondary);
	}


	/* Responsive behavior for modern carousel */
	@media (max-width: 768px) {
		.modern-carousel {
			grid-auto-columns: minmax(280px, 1fr);
			gap: 1rem;
		}
		
		.see-more-card {
			min-width: 280px;
			padding: 1.5rem;
		}

		.scroll-btn {
			width: 40px;
			height: 40px;
			font-size: 1.3rem;
		}

		.scroll-btn-prev {
			left: -20px;
		}

		.scroll-btn-next {
			right: -20px;
		}
	}

	@media (max-width: 480px) {
		.modern-carousel {
			grid-auto-columns: minmax(250px, 1fr);
			gap: 0.75rem;
			padding: 0.5rem 0;
		}
		
		.see-more-card {
			min-width: 250px;
			padding: 1rem;
		}

		.scroll-btn {
			width: 36px;
			height: 36px;
			font-size: 1.2rem;
		}

		.scroll-btn-prev {
			left: -18px;
		}

		.scroll-btn-next {
			right: -18px;
		}
	}

	.see-more-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.15);
		background: var(--color-primary);
		color: white;
	}

	.see-more-content {
		text-align: center;
	}

	.see-more-icon {
		font-size: 2rem;
		margin-bottom: 1rem;
		font-weight: bold;
	}

	.see-more-content h4 {
		font-size: 1.2rem;
		margin: 0 0 0.5rem 0;
		font-weight: 600;
	}

	.see-more-content p {
		margin: 0 0 0.5rem 0;
		opacity: 0.8;
	}

	.total-count {
		font-size: 0.8rem;
		opacity: 0.7;
		font-weight: 500;
	}

	.no-videos-message {
		text-align: center;
		padding: 3rem 2rem;
		color: var(--color-text-secondary);
	}

	.debug-info {
		font-size: 0.8rem;
		margin-top: 1rem;
		opacity: 0.7;
		font-family: monospace;
	}

	.category-error {
		text-align: center;
		padding: 2rem;
		background: var(--color-error-light);
		border-radius: 8px;
		color: var(--color-error);
	}

	.error-details {
		font-size: 0.9rem;
		margin: 0.5rem 0 1rem 0;
		opacity: 0.8;
	}

	.category-error button {
		background: var(--color-error);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		font-weight: 500;
		transition: background 0.2s ease;
	}

	.category-error button:hover {
		background: var(--color-error-dark);
	}

	/* Debug styles */
	.debug-info {
		background: #f0f0f0;
		border: 2px solid #007acc;
		padding: 1rem;
		margin: 1rem 0;
		border-radius: 8px;
		font-family: monospace;
		font-size: 0.9rem;
		color: #333;
	}

	.debug-category {
		border-radius: 31px;
		background: var(--bg-glass);
		box-shadow:  11px 11px 18px #676767,
             -11px -11px 18px var(--color-surface-hover);
		margin: 0.5rem 0;
	}

	@media (max-width: 768px) {
		.category-section {
			padding: 1.5rem;
		}

		.category-header h3 {
			font-size: 1.5rem;
		}
	}
</style> 


