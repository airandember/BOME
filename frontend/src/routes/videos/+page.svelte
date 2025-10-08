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
	let activeTab = $state<'latest' | 'collections' | 'categories' | 'allVideos'>('latest');

	let scrollThreshold = 800; // pixels from bottom to trigger auto-load (accounts for footer height)
	let isSearching = $state(false);
	let searchTimeout: ReturnType<typeof setTimeout> | null = null; // For debounced search
	
	// Hybrid fuzzy search with Fuse.js - static index only (no rebuilding)
	import Fuse from 'fuse.js';
	
	let fuseIndex: Fuse<Video> | null = null;
	let searchCache = new Map<string, { results: Video[], timestamp: number }>();
	let staticSearchIndex: Video[] = [];
	let searchIndexLoaded = $state(false);
	
	// Fuse.js configuration for optimal search
	const fuseOptions = {
		keys: [
			{ name: 'title', weight: 0.7 },
			{ name: 'description', weight: 0.2 },
			{ name: 'category', weight: 0.05 },
			{ name: 'tags', weight: 0.05 }
		],
		threshold: 0.4, // 0.0 = exact match, 1.0 = match anything
		includeScore: true,
		minMatchCharLength: 2,
		ignoreLocation: true
	};


	// Only set activeTab from URL on initial load, not on programmatic changes
	let initialLoad = $state(true);
	
	$effect(() => {
		if (initialLoad) {
			const tabParam = $page.url.searchParams.get('tab');
			console.log('🔗 Initial URL effect - tabParam:', tabParam);
			if (tabParam && ['latest', 'collections', 'categories', 'allVideos'].includes(tabParam)) {
				console.log('✅ Setting initial activeTab from URL to:', tabParam);
				activeTab = tabParam as typeof activeTab;
			}
			initialLoad = false;
		}
	});

	// Load the static comprehensive search index
	async function loadStaticSearchIndex() {
		try {
			console.log('📥 Loading comprehensive search index...');
			const response = await fetch('/search-index.json');
			
			if (!response.ok) {
				throw new Error(`Failed to load search index: ${response.status}`);
			}
			
			const data = await response.json();
			staticSearchIndex = data.videos || [];
			searchIndexLoaded = true;
			
			console.log('✅ Static search index loaded:', {
				totalVideos: staticSearchIndex.length,
				version: data.version,
				generatedAt: data.generatedAt
			});
			
			// Build the Fuse index ONCE from the static JSON (never rebuild)
			if (staticSearchIndex.length > 0) {
				fuseIndex = new Fuse(staticSearchIndex, fuseOptions);
				console.log('🔍 Static Fuse.js index built ONCE with', staticSearchIndex.length, 'videos');
			}
			
			// Preload thumbnails for the most recent/popular videos for instant display
			const recentVideos = staticSearchIndex
				.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())
				.slice(0, 20); // Preload 20 most recent thumbnails
			
			setTimeout(() => preloadThumbnails(recentVideos), 1000); // Delay to not block initial load
			
		} catch (error) {
			console.warn('⚠️ Failed to load static search index:', error);
			searchIndexLoaded = false;
		}
	}
	
	// Lightning-fast fuzzy search function with conditional thumbnail preloading
	function fuzzySearch(query: string, shouldPreloadThumbnails: boolean = false): Video[] {
		if (!query.trim() || !fuseIndex) return [];
		
		const startTime = performance.now();
		const results = fuseIndex.search(query); // No limit - show ALL matching results
		const searchTime = performance.now() - startTime;
		
		console.log('⚡ Fuzzy search completed in', Math.round(searchTime), 'ms for query:', query);
		
		// Sort by fuzzy match score first, then by date within similar relevance tiers
		const sortedResults = [...results].sort((a, b) => {
			// Primary sort: by relevance score (lower is better in Fuse.js)
			const scoreDiff = (a.score || 0) - (b.score || 0);
			
			// If scores are very similar (within 0.01), sort by date
			if (Math.abs(scoreDiff) < 0.08) {
				const dateA = new Date(a.item.createdAt || 0).getTime();
				const dateB = new Date(b.item.createdAt || 0).getTime();
				return dateB - dateA; // Newest first within similar relevance
			}
			
			return scoreDiff; // Otherwise sort by relevance
		});
		
		// Extract the actual video objects from sorted Fuse results (create new array for immutability)
		const videos = sortedResults.map(result => result.item);
		
		console.log('🎯 Search results sorted:', {
			query,
			totalResults: videos.length,
			bestMatch: videos[0]?.title,
			bestScore: results[0]?.score,
			searchTime: Math.round(searchTime),
			firstFewDates: videos.slice(0, 5).map(v => ({
				title: v.title?.substring(0, 30) + '...',
				date: v.createdAt,
				parsed: new Date(v.createdAt || 0).toISOString()
			})),
			firstFewScores: sortedResults.slice(0, 5).map(r => ({
				title: r.item.title?.substring(0, 30) + '...',
				score: r.score,
				date: r.item.createdAt
			}))
		});
		
		// Only preload thumbnails on final search, not during typing
		if (shouldPreloadThumbnails) {
			preloadThumbnails(videos.slice(0, 12)); // Preload first 12 thumbnails
		}
		
		return videos;
	}
	
	// Preload thumbnails for instant display with better error handling
	function preloadThumbnails(videos: Video[]) {
		videos.forEach((video, index) => {
			// Try multiple thumbnail sources for maximum compatibility
			const thumbnailUrl = video.thumbnailUrl || (video as any).thumbnail || (video as any).bunny?.previewImageUrl;
			
			if (thumbnailUrl) {
				// Add a small delay to prevent overwhelming the CDN
				setTimeout(() => {
					const img = new Image();
					
					// Remove crossOrigin to avoid CORS issues during preloading
					// The actual images will load fine when displayed normally
					
					img.onload = () => {
						// Thumbnail successfully preloaded
						// console.log('✅ Preloaded thumbnail:', thumbnailUrl);
					};
					
					img.onerror = () => {
						// Silently handle preload failures - images will load fine when actually displayed
						// This is normal for CDNs with access controls or CORS restrictions
						// Only log errors occasionally for debugging purposes
						if (Math.random() < 0.01) { // Log only 1% of errors to avoid spam
							console.debug('🔇 Thumbnail preload skipped (CDN access control):', thumbnailUrl.split('/').pop());
						}
					};
					
					img.src = thumbnailUrl;
				}, index * 100); // Increased stagger to 100ms to be gentler on CDN
			}
		});
	}
	
	// Fallback client-side filter (kept for compatibility)
	function clientSideFilter(videoList: Video[], query: string): Video[] {
		if (!query.trim()) return videoList;
		
		const searchTerms = query.toLowerCase().trim().split(' ').filter(term => term.length > 0);
		if (searchTerms.length === 0) return videoList;
		
		return videoList.filter(video => {
			const title = (video.title || '').toLowerCase();
			if (searchTerms.some(term => title.includes(term))) {
				return true;
			}
			
			const description = (video.description || '').toLowerCase();
			const category = (video.category || '').toLowerCase();
			const tags = (video.tags || []).join(' ').toLowerCase();
			
			const otherText = `${description} ${category} ${tags}`;
			return searchTerms.some(term => otherText.includes(term));
		});
	}

	// Handle search input with optimizations
	function handleSearchInput() {
		// Clear existing timeout
		if (searchTimeout) {
			clearTimeout(searchTimeout);
		}
		
		// If empty, clear search immediately
		if (searchQuery.trim() === '') {
			clearSearch();
			return;
		}
		
		// Don't search for queries < 2 characters
		if (searchQuery.trim().length < 2) {
			// Clear results but don't show error toast yet
			searchResults = [];
			allSearchResults = [];
			updateCurrentVideos();
			return;
		}
		
		// Debounce search for 300ms
		searchTimeout = setTimeout(() => {
			handleOptimizedSearch();
		}, 300);
	}
	
	// Handle Enter key press for search
	function handleSearchKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter') {
			event.preventDefault();
			handleOptimizedSearch();
		}
	}

	// Direct state approach - no derived state complexity
	let currentVideos = $state<Video[]>([]);
	
	// Function to update currentVideos based on current state
	function updateCurrentVideos() {
		console.log('🔄 updateCurrentVideos called:', {
			activeTab,
			searchQuery: searchQuery.length,
			videos: videos.length,
			latestVideos: latestVideos.length,
			searchResults: searchResults.length
		});
		
		// If we have a search query, use search results
		if (searchQuery.length > 0) {
			// Use allSearchResults if searchResults is empty but allSearchResults has data
			const resultsToUse = searchResults.length > 0 ? searchResults : allSearchResults;
			console.log('🔍 Setting currentVideos to search results:', {
				searchResults: searchResults.length,
				allSearchResults: allSearchResults.length,
				using: resultsToUse.length
			});
			// Create new array to ensure reactivity
			currentVideos = [...resultsToUse];
		}
		// For allVideos tab, use videos array
		else if (activeTab === 'allVideos') {
			console.log('📺 Setting currentVideos to videos:', videos.length);
			// Create new array to ensure reactivity
			currentVideos = [...videos];
		}
		// For latest tab, use latestVideos array  
		else if (activeTab === 'latest') {
			console.log('📺 Setting currentVideos to latestVideos:', latestVideos.length);
			// Create new array to ensure reactivity
			currentVideos = [...latestVideos];
		}
		// For other tabs (categories, collections), use empty
		else {
			console.log('📂 Setting currentVideos to empty for tab:', activeTab);
			currentVideos = [];
		}
		
		console.log('✅ currentVideos updated to length:', currentVideos.length);
	}
	
	// We'll call updateCurrentVideos() manually where needed instead of using $effect

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
		
		// Load the comprehensive search index immediately for instant search
		loadStaticSearchIndex();
		
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
		//if (typeof window !== 'undefined') {
		//	(window as any).debugVideoSearch = () => {
		//		console.log('🔍 Video Search Debug Info:', {
		//			searchQuery,
		//			isSearching,
		//			totalVideos: videos.length,
		//			latestVideos: latestVideos.length,
		//			allSearchResults: allSearchResults.length,
		//			filteredSearchResults: searchResults.length,
		//			currentVideos: currentVideos.length,
		//			activeTab,
		//			hasMore,
		//			currentPage
		//		});
		//		
		//		if (searchQuery) {
		//			console.log('🔍 Search Query Analysis:', {
		//				originalQuery: searchQuery,
		//				searchTerms: searchQuery.toLowerCase().trim().split(' ').filter(term => term.length > 0),
		//				allResults: allSearchResults.map(v => ({ id: v.id, title: v.title })),
		//				filteredResults: searchResults.map(v => ({ id: v.id, title: v.title }))
		//			});
		//		}
		//	};
		//}
		
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
			
			// Update currentVideos after loading data
			updateCurrentVideos();
			
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
			
			// Update currentVideos after loading more videos
			updateCurrentVideos();
			
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

	// Lightning-fast fuzzy search function
	function handleOptimizedSearch() {
		const query = searchQuery.trim();
		if (query.length === 0) {
			clearSearch();
			return;
		}

		// Check cache first (5 minute cache)
		const cacheKey = query.toLowerCase();
		const cached = searchCache.get(cacheKey);
		const cacheExpiry = 5 * 60 * 1000; // 5 minutes
		
		if (cached && (Date.now() - cached.timestamp) < cacheExpiry) {
			console.log('🚀 Cache hit for search:', query);
			allSearchResults = cached.results;
			updateCurrentVideos();
			
			// Show cached results toast
			toastStore.success(
				`Found ${cached.results.length} video${cached.results.length !== 1 ? 's' : ''} for "${query}" ⚡ Cached`,
				{ duration: 3000 }
			);
			return;
		}

		console.log('🔍 Starting fuzzy search for:', query);
		isSearching = true;
		
		// Show searching toast
		const searchToastId = toastStore.info(`Searching for "${query}"...`, {
			persistent: true,
			showClose: false
		});
		
		// Clear previous search results
		searchResults = [];
		allSearchResults = [];
		
		// Ensure static search index is loaded
		if (!fuseIndex) {
			console.warn('⚠️ Fuse index not ready, search may not work');
			return;
		}
		
		// INSTANT fuzzy search - no API calls needed!
		const startTime = performance.now();
		const results = fuzzySearch(query, true); // Enable thumbnail preloading for final search
		const searchTime = performance.now() - startTime;
		
		// Store results
		allSearchResults = results;
		
		// Cache the results
		searchCache.set(cacheKey, {
			results: allSearchResults,
			timestamp: Date.now()
		});
		
		// For search, we don't use pagination - show all results
		hasMore = false;
		
		console.log('⚡ Fuzzy search complete:', {
			query,
			totalResults: allSearchResults.length,
			searchTime: Math.round(searchTime),
			cached: false
		});
		
		// Stop loading and update UI
		isSearching = false;
		updateCurrentVideos();
		
		// Remove searching toast and show results
		toastStore.remove(searchToastId);
		
		if (allSearchResults.length > 0) {
			toastStore.success(
				`Found ${allSearchResults.length} video${allSearchResults.length !== 1 ? 's' : ''} for "${query}" (${Math.round(searchTime)}ms)`,
				{ duration: 4000 }
			);
		} else {
			toastStore.warning(
				`No videos found for "${query}". Try different keywords or check spelling.`,
				{ duration: 5000 }
			);
		}
	}

	// Keep the old function name for compatibility
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
			// Update currentVideos after search completes
			updateCurrentVideos();
		}
	}

	function clearSearch() {
		console.log('🧹 Clearing search');
		searchResults = [];
		allSearchResults = [];
		isSearching = false;
		currentPage = 1;
		hasMore = true;
		
		// Update currentVideos after clearing search
		updateCurrentVideos();
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
		console.log('🔄 switchTab called with:', tab, 'current activeTab:', activeTab);
		
		// Update the state first
		activeTab = tab;
		console.log('✅ activeTab set to:', activeTab);
		
		// Update URL parameter (this should not trigger the URL effect anymore)
		const url = new URL($page.url);
		url.searchParams.set('tab', tab);
		console.log('🔗 Updating URL to:', url.toString());
		replaceState(url, {});
		
		// Clear search when switching tabs
		if (searchQuery) {
			searchQuery = '';
			clearSearch();
		}
		
		// Load data if needed for video tabs
		if (tab === 'allVideos' && videos.length === 0) {
			loadVideos(true);
		} else if (tab === 'latest' && latestVideos.length === 0) {
			loadVideos(true);
		}
		
		// Update currentVideos after tab switch
		updateCurrentVideos();
		
		console.log('🏁 switchTab complete, final activeTab:', activeTab);
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

						<!-- Debug Info 
						<div style="background: #f0f0f0; padding: 10px; margin: 10px 0; border-radius: 5px; font-family: monospace;">
							<strong>🔍 DEBUG:</strong> activeTab = "{activeTab}" | videos.length = {videos.length} | currentVideos.length = {currentVideos.length} | loading = {loading}
							<br>
							<strong>🔍 SEARCH:</strong> searchQuery = "{searchQuery}" (length: {searchQuery.length}) | searchResults.length = {searchResults.length}
							<br>
							<button onclick={() => loadInitialData()} style="margin-top: 5px; padding: 5px 10px;">🔄 Force Load Data</button>
							<button onclick={() => { searchQuery = ''; }} style="margin-top: 5px; padding: 5px 10px;">🧹 Clear Search</button>
							<button onclick={() => { console.log('🔍 Manual currentVideos check:', { activeTab, videos: videos.length, result: activeTab === 'allVideos' ? videos : [] }); }} style="margin-top: 5px; padding: 5px 10px;">🔍 Test Logic</button>
						</div>-->

						<!-- Navigation Tabs -->
						<div class="hub-tabs">
							<button 
								class="tab-button {activeTab === 'latest' ? 'active' : ''}" 
								onclick={() => switchTab('latest')}
							>
								Latest Videos
							</button>	
							
				
							<button
								class="tab-button {activeTab === 'allVideos' ? 'active' : ''}"
								onclick={() => switchTab('allVideos')}
							>
								All Videos
							</button>
							<button 
								class="tab-button {activeTab === 'categories' ? 'active' : ''}" 
								onclick={() => switchTab('categories')}
							>
								Categories
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
								<!--<h2>Book of Mormon Evidence Videos</h2>-->
								<div class="filters-section">
									<div class="search-bar">
										<div class="search-input-container">
											<input
												type="text"
												placeholder="Search videos"
												bind:value={searchQuery}
												oninput={handleSearchInput}
												onkeydown={handleSearchKeydown}
												class="search-input"
											/>
											{#if searchQuery}
												<button class="btn-clear" onclick={handleClearSearch} title="Clear search">
													✕
												</button>
											{/if}
										</div>
										<button class="btn-primary" onclick={handleOptimizedSearch} disabled={isSearching}>
											{#if isSearching}
												<LoadingSpinner size="small" />
											{:else}
												🔍
											{/if}
											Search
										</button>
									</div>
									<!-- Search status now handled by toast notifications -->
								</div>

								{#if currentVideos.length === 0 && !loading && !loadingMore && !isSearching}
									<div class="no-results">
										{#if searchQuery && searchQuery.trim().length >= 2}
											<p>No videos found for "{searchQuery}".</p>
										{:else if searchQuery && searchQuery.trim().length < 2}
											<!-- Only show Clear Search button for queries under 2 characters -->
										{:else}
											<p>No videos found.</p>
										{/if}
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
								<!--<h2>Latest Uploads</h2>-->
								<div class="filters-section">
									<div class="search-bar">
										<div class="search-input-container">
											<input
												type="text"
												placeholder="Search videos"
												bind:value={searchQuery}
												oninput={handleSearchInput}
												onkeydown={handleSearchKeydown}
												class="search-input"
											/>
											{#if searchQuery}
												<button class="btn-clear" onclick={handleClearSearch} title="Clear search">
													✕
												</button>
											{/if}
										</div>
										<button class="btn-primary" onclick={handleOptimizedSearch} disabled={isSearching}>
											🔍 Search
										</button>
									</div>
									<!-- Search status now handled by toast notifications -->
								</div>

								{#if currentVideos.length === 0 && !loading && !loadingMore && !isSearching}
									<div class="no-results">
										{#if searchQuery && searchQuery.trim().length >= 2}
											<p>No videos found for "{searchQuery}".</p>
										{:else if searchQuery && searchQuery.trim().length < 2}
											<!-- Only show Clear Search button for queries under 2 characters -->
										{:else}
											<p>No videos found.</p>
										{/if}
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
								<!--<h2>Video Collections</h2>-->
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
								<!--<h2>Browse by Category</h2>
								<div class="debug-info">
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

	.search-input-container {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: center;
		position: relative;
		width: 100%;
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
		display: flex;
		justify-content: center;
	}

	/* Search status styles removed - now using toast notifications */

	.btn-link {
		background: none;
		border: none;
		color: inherit;
		text-decoration: underline;
		cursor: pointer;
		font-size: inherit;
		margin-left: 0.5rem;
		opacity: 0.8;
		transition: opacity 0.2s;
	}

	.btn-link:hover {
		opacity: 1;
	}

	.search-input {
		transition: border-color 0.2s ease;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--color-primary);
		box-shadow: 0 0 0 2px rgba(var(--primary-bom-rgb), 0.2);
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
		color: var(--text-primary);
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