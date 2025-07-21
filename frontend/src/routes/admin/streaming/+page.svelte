<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { MasterVideoService, type MasterVideo } from '$lib/master-video';
	import { stripeFinancialService, type StripeCustomer, type FinancialMetrics, type StripePayment } from '$lib/stripe-financial';

	const masterVideoService = new MasterVideoService();

	// Tab management
	let activeTab = 'overview'; // 'overview', 'videos', 'subscribers', 'financial'
	
	// Overview tab data
	let analytics: any = null;
	let overviewLoading = true;
	
	// Videos tab data
	let videos: MasterVideo[] = [];
	let videosLoading = true;
	let syncing = false;
	let checkingConflicts = false;
	let selectedVideos: Set<number> = new Set();
	let searchQuery = '';
	let currentTab = 'synced'; // 'synced', 'needs-attention', or 'all'
	let currentPage = 1;
	let pageSize = 20;
	let totalVideos = 0;
	let totalPages = 0;
	let sortField: string = 'id';
	let sortDirection: 'asc' | 'desc' = 'asc';
	
	// Subscribers tab data (separated from financial)
	let subscribersLoading = true;
	let subscribers: StripeCustomer[] = [];
	let subscribersPagination = {
		current_page: 1,
		per_page: 20,
		total: 0,
		total_pages: 0
	};
	let subscribersSearch = '';
	let subscribersStatus = '';
	
	// Financial tab data
	let financialLoading = true;
	let financialMetrics: FinancialMetrics | null = null;
	let recentTransactions: StripePayment[] = [];
	let topCustomers: Array<{ customer: StripeCustomer; total_spent: number }> = [];
	let revenueChart: Array<{ date: string; revenue: number; subscriptions: number }> = [];
	let selectedReportPeriod: '7d' | '30d' | '90d' | '1y' | 'all' = '30d';
	let selectedProjectionPeriod: '30d' | '90d' | '6m' | '1y' = '90d';
	
	// Modal state
	let showEditModal = false;
	let editingVideo: MasterVideo | null = null;
	let editForm: Partial<MasterVideo> = {};
	let showCustomerModal = false;
	let selectedCustomer: StripeCustomer | null = null;
	
	// Bulk edit state
	let showBulkEditModal = false;
	let bulkEditForm = {
		title: '',
		category: '',
		description: '',
		tags: ''
	};
	let isApplyingBulkEdit = false;

	let error = '';
	let userRole = '';
	let userPermissions: string[] = [];

	// Sync results
	let lastSyncResult: any = null;
	let lastConflictCheck: any = null;

	onMount(() => {
		console.log('Admin streaming page mounted');
		console.log('Auth debug info:', $auth);
		console.log('Is admin:', $auth.user?.role);
		
		if (!$auth.isAuthenticated || !$auth.user) {
			console.error('User is not authenticated, redirecting...');
			showToast('Access denied. Admin privileges required.');
			goto('/login');
			return;
		}

		// Check if user has admin role
		const adminRoles = [
			'super_admin', 'system_admin', 'content_manager', 
			'articles_manager', 'youtube_manager', 'streaming_manager',
			'events_manager', 'advertisement_manager', 'user_manager',
			'analytics_manager', 'financial_admin', 'admin'
		];
		
		if (!adminRoles.includes($auth.user.role)) {
			console.error('User is not an admin, redirecting...');
			showToast('Access denied. Admin privileges required.');
			goto('/login');
			return;
		}

		userRole = $auth.user.role;
		userPermissions = getUserPermissions($auth.user.role);
		
		testBackendConnection();
	});
	
	function getUserPermissions(role: string): string[] {
		const permissionMap: Record<string, string[]> = {
			'super_admin': ['*'], // All permissions
			'system_admin': ['system:read', 'system:update', 'system:manage', 'analytics:read', 'analytics:export'],
			'content_manager': ['content:read', 'content:create', 'content:update', 'content:delete', 'videos:manage', 'analytics:read'],
			'articles_manager': ['articles:read', 'articles:create', 'articles:update', 'articles:delete', 'analytics:read'],
			'youtube_manager': ['videos:read', 'videos:create', 'videos:update', 'videos:delete', 'analytics:read'],
			'streaming_manager': ['videos:read', 'videos:create', 'videos:update', 'videos:delete', 'analytics:read'],
			'events_manager': ['events:read', 'events:create', 'events:update', 'events:delete', 'analytics:read'],
			'advertisement_manager': ['advertisements:read', 'advertisements:create', 'advertisements:update', 'advertisements:delete', 'analytics:read'],
			'user_manager': ['users:read', 'users:create', 'users:update', 'users:delete', 'analytics:read'],
			'analytics_manager': ['analytics:read', 'analytics:export', 'analytics:manage'],
			'financial_admin': ['financial:read', 'financial:manage', 'analytics:read'],
			'admin': ['*'] // Legacy admin - all permissions
		};
		return permissionMap[role] || [];
	}
	
	async function testBackendConnection() {
		try {
			console.log('Testing backend connectivity...');
			const response = await fetch('http://localhost:8080/health');
			console.log('Health check response:', response.status, response.ok);
			
			if (response.ok) {
				console.log('Backend is reachable, loading data...');
				loadOverviewData();
			} else {
				console.error('Backend health check failed');
				showToast('Backend connection failed');
			}
		} catch (error) {
			console.error('Backend connectivity test failed:', error);
			showToast('Cannot connect to backend');
		}
	}

	async function loadOverviewData() {
		try {
			overviewLoading = true;
			console.log('🔍 Attempting to load dashboard data from API...');
			const data = await masterVideoService.getDashboardAnalytics();
			console.log('✅ Dashboard API response:', data);
			if (data.success) {
				analytics = data.data;
			}
			console.log('📊 Dashboard data loaded:', data);
		} catch (error) {
			console.error('❌ Error loading overview data:', error);
			
			// Fallback to using master video service for basic stats
			try {
				console.log('🔄 Falling back to master video service for basic stats...');
				const statsResponse = await masterVideoService.getStats();
				console.log('📈 Master video stats response:', statsResponse);
				if (statsResponse.success) {
					// Create a basic dashboard data structure from the stats
					analytics = {
						total_users: 0,
						total_videos: statsResponse.stats.total_videos || 0,
						total_views: statsResponse.stats.total_views || 0,
						total_revenue: 0,
						active_subscriptions: 0,
						video_stats: {
							total_videos: statsResponse.stats.total_videos || 0,
							synced_videos: statsResponse.stats.videos_by_sync_status?.synced || 0,
							needs_attention: statsResponse.stats.videos_by_sync_status?.needs_attention || 0,
							total_views: statsResponse.stats.total_views || 0,
							videos_by_status: statsResponse.stats.videos_by_status || {},
							videos_by_sync_status: statsResponse.stats.videos_by_sync_status || {},
							pending_conflicts: statsResponse.stats.pending_conflicts || 0
						},
						view_analytics: {
							total_views: statsResponse.stats.total_views || 0,
							views_today: 0,
							views_week: 0,
							growth_rate: 0
						},
						subscriber_metrics: {
							total_subscribers: 0,
							active_subscriptions: 0,
							monthly_revenue: 0,
							churn_rate: 0
						},
						real_time: {
							active_users: 0,
							recent_activity: []
						}
					};
					console.log('🛠️ Fallback data created:', analytics);
					console.log('📊 statsResponse.stats.total_views:', statsResponse.stats.total_views);
					console.log('📊 view_analytics.total_views:', analytics.view_analytics.total_views);
					showToast('Loaded basic video stats (some features may be limited)');
				} else {
					throw new Error('Failed to load fallback stats');
				}
			} catch (fallbackError) {
				console.error('❌ Error loading fallback data:', fallbackError);
				showToast('Failed to load overview data');
			}
		} finally {
			overviewLoading = false;
		}
	}

	async function loadVideosData() {
		try {
			videosLoading = true;
			await loadVideos();
		} catch (error) {
			console.error('Error loading videos data:', error);
			showToast('Failed to load videos data');
		} finally {
			videosLoading = false;
		}
	}

	async function loadSubscribersData() {
		try {
			subscribersLoading = true;
			
			const response = await stripeFinancialService.getCustomers({
				page: subscribersPagination.current_page,
				limit: subscribersPagination.per_page,
				search: subscribersSearch || undefined,
				status: subscribersStatus || undefined,
				sort_field: 'created',
				sort_direction: 'desc'
			});
			
			if (response.success) {
				subscribers = response.customers;
				subscribersPagination = response.pagination;
			} else {
				throw new Error('Failed to load subscribers');
			}
		} catch (error) {
			console.error('Error loading subscribers data:', error);
			
			// Fallback to mock data when API fails
			console.log('Using fallback subscriber data...');
			subscribers = [
				{ 
					id: 'cus_1', 
					email: 'john@example.com', 
					name: 'John Doe', 
					created: Date.now() - 86400000 * 30,
					payment_methods: [],
					metadata: {}
				},
				{ 
					id: 'cus_2', 
					email: 'jane@example.com', 
					name: 'Jane Smith', 
					created: Date.now() - 86400000 * 15,
					payment_methods: [],
					metadata: {}
				},
				{ 
					id: 'cus_3', 
					email: 'bob@example.com', 
					name: 'Bob Johnson', 
					created: Date.now() - 86400000 * 7,
					payment_methods: [],
					metadata: {}
				}
			];
			
			showToast('Loaded mock subscriber data (real data unavailable)');
		} finally {
			subscribersLoading = false;
		}
	}

	async function loadFinancialData() {
		try {
			financialLoading = true;
			
			const response = await stripeFinancialService.getFinancialDashboard();
			
			if (response.success) {
				financialMetrics = response.data.metrics;
				recentTransactions = response.data.recent_transactions;
				topCustomers = response.data.top_customers;
				revenueChart = response.data.revenue_chart;
			} else {
				throw new Error('Failed to load financial data');
			}
		} catch (error) {
			console.error('Error loading financial data:', error);
			
			// Fallback to mock data when API fails
			console.log('Using fallback financial data...');
			financialMetrics = {
				total_revenue: 12500,
				monthly_recurring_revenue: 8500,
				annual_recurring_revenue: 102000,
				average_revenue_per_user: 125,
				revenue_growth_rate: 15.5,
				total_subscriptions: 68,
				active_subscriptions: 65,
				canceled_subscriptions: 3,
				subscription_growth_rate: 12.3,
				churn_rate: 2.1,
				total_customers: 85,
				new_customers_this_month: 12,
				customer_growth_rate: 18.2,
				customer_lifetime_value: 450,
				payment_success_rate: 98.5,
				average_payment_amount: 125,
				failed_payments: 2,
				refund_rate: 1.2,
				plan_distribution: {
					'Basic Plan': 45,
					'Premium Plan': 35,
					'Pro Plan': 20
				},
				revenue_by_plan: {
					'Basic Plan': 4500,
					'Premium Plan': 5600,
					'Pro Plan': 2400
				}
			};
			
			recentTransactions = [];
			topCustomers = [];
			revenueChart = [];
			
			showToast('Loaded mock financial data (real data unavailable)');
		} finally {
			financialLoading = false;
		}
	}

	async function loadVideos() {
		try {
			// Determine sync status based on current tab
			let syncStatus: string | undefined;
			if (currentTab === 'synced') {
				syncStatus = 'synced';
			} else if (currentTab === 'needs-attention') {
				syncStatus = 'needs_attention';
			}
			
			console.log('Loading videos with params:', {
				page: currentPage,
				limit: pageSize,
				sync_status: syncStatus,
				search: searchQuery || undefined,
				sort_field: sortField,
				sort_direction: sortDirection
			});
			
			const response = await masterVideoService.getMasterVideos({
				page: currentPage,
				limit: pageSize,
				sync_status: syncStatus,
				search: searchQuery || undefined,
				sort_field: sortField,
				sort_direction: sortDirection
			});
			
			console.log('API response:', response);
			
			if (response.success) {
				videos = response.videos || [];
				console.log('Videos:', videos);
				totalVideos = response.pagination.total;
				totalPages = response.pagination.total_pages;
				console.log('Videos loaded:', videos.length, 'Total:', totalVideos);
			} else {
				throw new Error('Failed to load videos');
			}
		} catch (error) {
			console.error('Error loading videos:', error);
			showToast('Failed to load videos');
		}
	}

	function handleTabChange(tab: string) {
		activeTab = tab;
		
		// Load data based on selected tab
		if (tab === 'overview') {
			loadOverviewData();
		} else if (tab === 'videos') {
			loadVideosData();
		} else if (tab === 'subscribers') {
			loadSubscribersData();
		} else if (tab === 'financial') {
			loadFinancialData();
		}
	}

	async function syncVideos() {
		try {
			syncing = true;
			showToast('Syncing videos from Bunny.net...', 'info');
			
			const response = await masterVideoService.syncFromBunny();
			
			if (response.success) {
				lastSyncResult = response.result;
				showToast(`Sync completed: ${response.result.synced} synced, ${response.result.updated} updated, ${response.result.conflicts} conflicts, ${response.result.errors} errors`, 'success');
				
				// Reload data
				loadVideos();
				loadOverviewData();
			} else {
				showToast('Sync failed', 'error');
			}
		} catch (error) {
			showToast('Failed to sync videos from Bunny.net', 'error');
			console.error('Error syncing videos:', error);
		} finally {
			syncing = false;
		}
	}

	async function checkConflicts() {
		try {
			checkingConflicts = true;
			showToast('Checking for conflicts...', 'info');
			
			const response = await masterVideoService.checkConflicts();
			if (response.success) {
				$lastConflictCheck = response.result;
				const message = response.result.conflict_count > 0 
					? `Found ${response.result.conflict_count} conflicts that need attention`
					: 'No conflicts found - all videos are in sync';
				const type = response.result.conflict_count > 0 ? 'warning' : 'success';
				
				showToast(message, type);
				
				// Reload data
				loadVideos();
				loadOverviewData();
			} else {
				showToast('Conflict check failed', 'error');
			}
		} catch (error) {
			showToast('Failed to check conflicts', 'error');
			console.error('Error checking conflicts:', error);
		} finally {
			checkingConflicts = false;
		}
	}

	async function deleteVideo(videoId: number) {
		if (!confirm('Are you sure you want to delete this video? This action cannot be undone.')) {
			return;
		}

		try {
			const response = await masterVideoService.deleteMasterVideo(videoId);
			
			if (response.success) {
				showToast('Video deleted successfully', 'success');
				loadVideos();
				// Reload dashboard data instead of stats
				loadOverviewData();
			} else {
				showToast('Failed to delete video', 'error');
			}
		} catch (error) {
			showToast('Failed to delete video', 'error');
			console.error('Error deleting video:', error);
		}
	}

	async function updateVideo() {
		if (!editingVideo) return;

		try {
			const response = await masterVideoService.updateMasterVideo(editingVideo.ID, editForm);
			
			if (response.success) {
				showToast('Video updated successfully', 'success');
				closeEditModal();
				loadVideos();
				loadOverviewData();
			} else {
				showToast('Failed to update video', 'error');
			}
		} catch (error) {
			showToast('Failed to update video', 'error');
			console.error('Error updating video:', error);
		}
	}

	function openEditModal(video: MasterVideo) {
		editingVideo = video;
		editForm = {
			Title: video.Title,
			Description: video.Description,
			Category: video.Category,
			Tags: video.Tags,
			Duration: video.Duration,
			ThumbnailURL: video.ThumbnailURL,
			VideoURL: video.VideoURL,
			IframeSrc: video.IframeSrc,
			PlaybackURL: video.PlaybackURL,
			Status: video.Status,
			Views: video.Views,
			Likes: video.Likes,
			IsPublic: video.IsPublic,
			CollectionID: video.CollectionID,
			SyncStatus: video.SyncStatus,
			SyncNotes: video.SyncNotes
		};
		showEditModal = true;
	}

	function closeEditModal() {
		showEditModal = false;
		editingVideo = null;
		editForm = {};
	}

	function openBulkEditModal() {
		if (selectedVideos.size === 0) {
			showToast('Please select at least one video for bulk edit', 'warning');
			return;
		}
		showBulkEditModal = true;
		bulkEditForm = {
			title: '',
			category: '',
			description: '',
			tags: ''
		};
	}

	function closeBulkEditModal() {
		showBulkEditModal = false;
		bulkEditForm = {
			title: '',
			category: '',
			description: '',
			tags: ''
		};
	}

	async function applyBulkEdit() {
		if (selectedVideos.size === 0) {
			showToast('No videos selected for bulk edit', 'warning');
			return;
		}

		try {
			isApplyingBulkEdit = true;
			
			// Prepare the bulk edit data
			const bulkEditData: Partial<MasterVideo> = {};
			
			if (bulkEditForm.title.trim()) {
				bulkEditData.Title = bulkEditForm.title.trim();
			}
			if (bulkEditForm.category.trim()) {
				bulkEditData.Category = bulkEditForm.category.trim();
			}
			if (bulkEditForm.description.trim()) {
				bulkEditData.Description = bulkEditForm.description.trim();
			}
			if (bulkEditForm.tags.trim()) {
				bulkEditData.Tags = bulkEditForm.tags.split(',').map(tag => tag.trim()).filter(tag => tag.length > 0);
			}

			// Apply to all selected videos
			const videoIds = Array.from(selectedVideos);
			let successCount = 0;
			let errorCount = 0;

			for (const videoId of videoIds) {
				try {
					const response = await masterVideoService.updateMasterVideo(videoId, bulkEditData);
					if (response.success) {
						successCount++;
					} else {
						errorCount++;
					}
				} catch (error) {
					console.error(`Failed to update video ${videoId}:`, error);
					errorCount++;
				}
			}

			// Show results
			if (successCount > 0) {
				showToast(`Successfully updated ${successCount} video(s)${errorCount > 0 ? `, ${errorCount} failed` : ''}`, 'success');
				selectedVideos.clear();
				loadVideos();
				loadOverviewData();
			} else {
				showToast(`Failed to update any videos (${errorCount} errors)`, 'error');
			}

			closeBulkEditModal();
		} catch (error) {
			console.error('Bulk edit error:', error);
			showToast('Failed to apply bulk edit', 'error');
		} finally {
			isApplyingBulkEdit = false;
		}
	}

	function toggleVideoSelection(videoId: number) {
		if (selectedVideos.has(videoId)) {
			selectedVideos.delete(videoId);
		} else {
			selectedVideos.add(videoId);
		}
		selectedVideos = selectedVideos; // Trigger reactivity
	}

	function selectAllVideos() {
		videos.forEach(video => selectedVideos.add(video.ID));
		selectedVideos = selectedVideos; // Trigger reactivity
	}

	function deselectAllVideos() {
		selectedVideos.clear();
		selectedVideos = selectedVideos; // Trigger reactivity
	}

	function handleModalClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			closeEditModal();
		}
	}

	function formatDuration(seconds: number): string {
		if (!seconds) return '0:00';
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${secs.toString().padStart(2, '0')}`;
	}

	function formatFileSize(bytes: number): string {
		if (!bytes) return '0 B';
		const sizes = ['B', 'KB', 'MB', 'GB'];
		let size = bytes;
		let unitIndex = 0;
		
		while (size >= 1024 && unitIndex < sizes.length - 1) {
			size /= 1024;
			unitIndex++;
		}
		
		return `${size.toFixed(1)} ${sizes[unitIndex]}`;
	}

	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString();
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function getStatusColor(status: string): string {
		switch (status) {
			case 'ready': return 'success';
			case 'processing': case 'transcoding': return 'warning';
			case 'error': case 'upload_failed': return 'error';
			default: return 'info';
		}
	}

	function getSyncStatusColor(syncStatus: string): string {
		switch (syncStatus) {
			case 'synced': return 'success';
			case 'needs_attention': return 'warning';
			case 'conflict': return 'error';
			default: return 'info';
		}
	}

	function getStatusIcon(status: string): string {
		switch (status) {
			case 'ready': return '✓';
			case 'processing': case 'transcoding': return '⏳';
			case 'error': case 'upload_failed': return '✗';
			default: return '○';
		}
	}

	function getSyncStatusIcon(syncStatus: string): string {
		switch (syncStatus) {
			case 'synced': return '✓';
			case 'needs_attention': return '⚠';
			case 'conflict': return '✗';
			default: return '○';
		}
	}

	function handleVideoTabChange(tab: string) {
		currentTab = tab;
		currentPage = 1;
		loadVideos();
	}

	function handlePageChange(page: number) {
		currentPage = page;
		loadVideos();
	}

	function handleSort(field: string) {
		if (sortField === field) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortField = field;
			sortDirection = 'asc';
		}
		currentPage = 1;
		loadVideos();
	}

	function getSortIcon(field: string): string {
		if (sortField !== field) return '↕';
		return sortDirection === 'asc' ? '↑' : '↓';
	}

	function getSortClass(field: string): string {
		if (sortField !== field) return 'sortable';
		return `sortable ${sortDirection}`;
	}

	// Computed properties for video stats
	$: syncedCount = analytics?.video_stats?.synced_videos || 0;
	$: needsAttentionCount = analytics?.video_stats?.needs_attention || 0;
	$: totalViews = analytics?.view_analytics?.total_views || 0;
	$: activeUsers = analytics?.real_time?.active_users || 0;
	$: console.log('🔄 totalViews reactive update:', {
		totalViews,
		viewAnalytics: analytics?.view_analytics,
		viewAnalyticsTotalViews: analytics?.view_analytics?.total_views,
		analytics: analytics
	});
</script>

<svelte:head>
	<title>Streaming Admin Dashboard - BOME</title>
	<meta name="description" content="Manage streaming content, subscribers, and financial data" />
</svelte:head>

<div class="streaming-page">
	<!-- Header -->
	<header class="page-header">
		<div class="header-content">
			<div class="header-left">
				<h1>Streaming Admin Dashboard</h1>
				<p>Manage your streaming platform, videos, subscribers, and financial data</p>
			</div>
			<div class="header-actions">
				{#if activeTab === 'videos'}
					<button class="btn btn-secondary" on:click={checkConflicts} disabled={checkingConflicts}>
						{#if checkingConflicts}
							<LoadingSpinner size="small" />
							Checking...
						{:else}
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
								<line x1="12" y1="9" x2="12" y2="13"></line>
								<line x1="12" y1="17" x2="12.01" y2="17"></line>
							</svg>
							Check Conflicts
						{/if}
					</button>
					<button class="btn btn-secondary" on:click={syncVideos} disabled={syncing}>
						{#if syncing}
							<LoadingSpinner size="small" />
							Syncing...
						{:else}
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M21 2v6h-6"></path>
								<path d="M3 12a9 9 0 0 1 15-6.7L21 8"></path>
								<path d="M3 22v-6h6"></path>
								<path d="M21 12a9 9 0 0 1-15 6.7L3 16"></path>
							</svg>
							Sync from Bunny
						{/if}
					</button>
					<button class="btn btn-primary" on:click={() => goto('/admin/streaming/upload')}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
							<polyline points="7,10 12,15 17,10"></polyline>
							<line x1="12" y1="15" x2="12" y2="3"></line>
						</svg>
						Upload Video
					</button>
				{/if}
			</div>
		</div>
	</header>

	<!-- Main Tabs -->
	<div class="main-tabs">
		<button 
			class="main-tab {activeTab === 'overview' ? 'active' : ''}" 
			on:click={() => handleTabChange('overview')}
		>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
				<circle cx="9" cy="9" r="2"></circle>
				<path d="M21 15l-3.086-3.086a2 2 0 0 0-2.828 0L6 21"></path>
			</svg>
			Overview
		</button>
		<button 
			class="main-tab {activeTab === 'videos' ? 'active' : ''}" 
			on:click={() => handleTabChange('videos')}
		>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<polygon points="23,7 16,12 23,17 23,7"></polygon>
				<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
			</svg>
			Video Management
		</button>
		<button 
			class="main-tab {activeTab === 'subscribers' ? 'active' : ''}" 
			on:click={() => handleTabChange('subscribers')}
		>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
				<circle cx="9" cy="7" r="4"></circle>
				<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
				<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
			</svg>
			Subscribers
		</button>
		<button 
			class="main-tab {activeTab === 'financial' ? 'active' : ''}" 
			on:click={() => handleTabChange('financial')}
		>
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<line x1="12" y1="1" x2="12" y2="23"></line>
				<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
			</svg>
			Financial
		</button>
	</div>

	<!-- Tab Content -->
	<div class="tab-content">
		{#if activeTab === 'overview'}
			<!-- Overview Tab -->
			<div class="overview-tab">
				{#if overviewLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading overview data...</p>
					</div>
				{:else if analytics}
					<!-- Overview Stats Grid -->
					<div class="overview-stats-grid">
						<div class="overview-stat-card">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polygon points="23,7 16,12 23,17 23,7"></polygon>
									<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{analytics.video_stats?.total_videos?.toLocaleString() || '0'}</div>
								<div class="stat-label">Total Videos</div>
								<div class="stat-change">+{analytics.video_stats?.synced_videos || 0} synced</div>
							</div>
						</div>
						
						<div class="overview-stat-card success">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
									<circle cx="9" cy="7" r="4"></circle>
									<path d="M23 21v-2a4 4 0 0 0-3-3.87"></path>
									<path d="M16 3.13a4 4 0 0 1 0 7.75"></path>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{activeUsers}</div>
								<div class="stat-label">Active Users</div>
								<div class="stat-change">+12 this month</div>
							</div>
						</div>
						
						<div class="overview-stat-card warning">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
									<line x1="12" y1="9" x2="12" y2="13"></line>
									<line x1="12" y1="17" x2="12.01" y2="17"></line>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{needsAttentionCount}</div>
								<div class="stat-label">Needs Attention</div>
								<div class="stat-change">{analytics.video_stats?.needs_attention || 0} conflicts</div>
							</div>
						</div>
						
						<div class="overview-stat-card info">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
									<circle cx="12" cy="12" r="3"></circle>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{totalViews.toLocaleString()}</div>
								<div class="stat-label">Total Views</div>
								<div class="stat-change">+{analytics.view_analytics?.views_week?.toLocaleString() || '0'} this week</div>
							</div>
						</div>
					</div>

					<!-- Recent Activity -->
					<div class="recent-activity-section">
						<h3>Recent Activity</h3>
						<div class="activity-list">
							{#each analytics?.real_time?.recent_activity?.slice(0, 3) || [] as activity}
								<div class="activity-item">
									<div class="activity-icon">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<circle cx="12" cy="12" r="10"></circle>
											<line x1="12" y1="16" x2="12" y2="12"></line>
											<line x1="12" y1="8" x2="12.01" y2="8"></line>
										</svg>
									</div>
									<div class="activity-content">
										<div class="activity-message">{activity.message}</div>
										<div class="activity-time">{activity.time}</div>
									</div>
								</div>
							{/each}
						</div>
					</div>
				{:else}
					<div class="error-state">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10"></circle>
							<line x1="15" y1="9" x2="9" y2="15"></line>
							<line x1="9" y1="9" x2="15" y2="15"></line>
						</svg>
						<h3>Failed to load overview data</h3>
						<p>Please try refreshing the page or contact support if the problem persists.</p>
						<button class="btn btn-primary" on:click={loadOverviewData}>Retry</button>
					</div>
				{/if}
			</div>

		{:else if activeTab === 'videos'}
			<!-- Video Management Tab -->
			<div class="videos-tab">
				<!-- Video Stats Cards -->
				{#if analytics}
					<div class="video-stats-grid">
						<div class="video-stat-card">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polygon points="23,7 16,12 23,17 23,7"></polygon>
									<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{analytics.video_stats?.total_videos || 0}</div>
								<div class="stat-label">Total Videos</div>
							</div>
						</div>
						
						<div class="video-stat-card success">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<polyline points="20,6 9,17 4,12"></polyline>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{syncedCount}</div>
								<div class="stat-label">Synced</div>
							</div>
						</div>
						
						<div class="video-stat-card warning">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
									<line x1="12" y1="9" x2="12" y2="13"></line>
									<line x1="12" y1="17" x2="12.01" y2="17"></line>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{needsAttentionCount}</div>
								<div class="stat-label">Needs Attention</div>
							</div>
						</div>
						
						<div class="video-stat-card">
							<div class="stat-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
									<circle cx="12" cy="12" r="3"></circle>
								</svg>
							</div>
							<div class="stat-content">
								<div class="stat-value">{totalViews.toLocaleString()}</div>
								<div class="stat-label">Total Views</div>
							</div>
						</div>
					</div>
				{/if}

				<!-- Video Tabs -->
				<div class="video-tabs">
					<button 
						class="video-tab {currentTab === 'synced' ? 'active' : ''}" 
						on:click={() => handleVideoTabChange('synced')}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<polyline points="20,6 9,17 4,12"></polyline>
						</svg>
						Synced Videos ({syncedCount})
					</button>
					<button 
						class="video-tab {currentTab === 'needs-attention' ? 'active' : ''}" 
						on:click={() => handleVideoTabChange('needs-attention')}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
							<line x1="12" y1="9" x2="12" y2="13"></line>
							<line x1="12" y1="17" x2="12.01" y2="17"></line>
						</svg>
						Needs Attention ({needsAttentionCount})
					</button>
					<button 
						class="video-tab {currentTab === 'all' ? 'active' : ''}" 
						on:click={() => handleVideoTabChange('all')}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
							<circle cx="8.5" cy="8.5" r="1.5"></circle>
							<polyline points="21,15 16,10 5,21"></polyline>
						</svg>
						All Videos ({analytics?.video_stats?.total_videos || 0})
					</button>
				</div>

				<!-- Search -->
				<div class="search-section">
					<input
						type="text"
						placeholder="Search videos..."
						bind:value={searchQuery}
						on:input={() => { 
							currentPage = 1; 
							loadVideos();
							if (searchQuery.trim()) {
								showToast(`Searching for: "${searchQuery}"`, 'info');
							}
						}}
						class="search-input"
					/>
				</div>

				<!-- Bulk Operations -->
				{#if videos.length > 0}
					<div class="bulk-operations">
						<div class="bulk-selection">
							<label class="select-all-checkbox">
								<input 
									type="checkbox" 
									checked={selectedVideos.size === videos.length && videos.length > 0}
									indeterminate={selectedVideos.size > 0 && selectedVideos.size < videos.length}
									on:change={(e) => {
										const target = e.target as HTMLInputElement;
										if (target.checked) {
											selectAllVideos();
										} else {
											deselectAllVideos();
										}
									}}
								/>
								<span>Select All</span>
							</label>
							{#if selectedVideos.size > 0}
								<span class="selected-count">{selectedVideos.size} video(s) selected</span>
								<button class="btn btn-secondary" on:click={deselectAllVideos}>
									Clear Selection
								</button>
							{/if}
						</div>
						
						{#if selectedVideos.size > 0}
							<div class="bulk-actions">
								<button class="btn btn-primary" on:click={openBulkEditModal}>
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
										<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
									</svg>
									Bulk Edit ({selectedVideos.size})
								</button>
							</div>
						{/if}
					</div>
				{/if}

				<!-- Videos Table -->
				<div class="videos-table-container">
					{#if videosLoading}
						<div class="loading-container">
							<LoadingSpinner />
							<p>Loading videos...</p>
						</div>
					{:else if videos.length === 0}
						<div class="empty-state">
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="23,7 16,12 23,17 23,7"></polygon>
								<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
							</svg>
							<h3>No videos found</h3>
							<p>No videos match your current filters. Try adjusting your search criteria or sync from Bunny.net.</p>
							<button class="btn btn-primary" on:click={syncVideos}>
								Sync from Bunny.net
							</button>
						</div>
					{:else}
						<table class="videos-table">
							<thead>
								<tr>
									<th class="checkbox-header">
										<!-- Select all checkbox is handled in bulk operations -->
									</th>
									<th class={getSortClass('id')} on:click={() => handleSort('id')}>
										ID {getSortIcon('id')}
									</th>
									<th>Thumbnail</th>
									<th class={getSortClass('title')} on:click={() => handleSort('title')}>
										Title {getSortIcon('title')}
									</th>
									<th class={getSortClass('category')} on:click={() => handleSort('category')}>
										Category {getSortIcon('category')}
									</th>
									<th class={getSortClass('status')} on:click={() => handleSort('status')}>
										Status {getSortIcon('status')}
									</th>
									<th class={getSortClass('sync_status')} on:click={() => handleSort('sync_status')}>
										Sync Status {getSortIcon('sync_status')}
									</th>
									<th class={getSortClass('views')} on:click={() => handleSort('views')}>
										Views {getSortIcon('views')}
									</th>
									<th class={getSortClass('duration')} on:click={() => handleSort('duration')}>
										Duration {getSortIcon('duration')}
									</th>
									<th class={getSortClass('created_at')} on:click={() => handleSort('created_at')}>
										Created {getSortIcon('created_at')}
									</th>
									<th>Actions</th>
								</tr>
							</thead>
							<tbody>
								{#each videos as video}
									<tr class={selectedVideos.has(video.ID) ? 'selected' : ''}>
										<td class="checkbox-cell">
											<input 
												type="checkbox" 
												checked={selectedVideos.has(video.ID)}
												on:change={() => toggleVideoSelection(video.ID)}
											/>
										</td>
										<td>{video.ID}</td>
										<td class="thumbnail-cell">
											{#if video.ThumbnailURL}
												<img 
													src={video.ThumbnailURL} 
													alt={video.Title} 
													class="video-thumbnail" 
													loading="lazy"
													referrerpolicy="no-referrer"
													on:error={(e) => {
														console.warn('Thumbnail failed to load:', video.ThumbnailURL);
													}}
												/>
												<div class="no-thumbnail" style="display: none;">No Image</div>
											{:else}
												<div class="no-thumbnail">No Image</div>
											{/if}
										</td>
										<td class="title-cell">
											<div class="video-title">{video.Title}</div>
											<div class="video-description">{video.Description}</div>
										</td>
										<td>
											{video.Category}
										</td>
										<td class="td_cell">
											<span class="status-badge {getStatusColor(video.Status)}">
												{getStatusIcon(video.Status)} {video.Status}
											</span>
										</td>
										<td>
											<span class="status-badge {getSyncStatusColor(video.SyncStatus)}">
												{getSyncStatusIcon(video.SyncStatus)} {video.SyncStatus}
											</span>
										</td>
										<td>{video.Views?.toLocaleString() || '0'}</td>
										<td>{formatDuration(video.Duration)}</td>
										<td>{formatDate(video.CreatedAt)}</td>
										<td class="actions-cell">
											<button class="btn btn-sm btn-secondary" on:click={() => openEditModal(video)}>
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
													<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
												</svg>
											</button>
											<button class="btn btn-sm btn-danger" on:click={() => deleteVideo(video.ID)}>
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<polyline points="3,6 5,6 21,6"></polyline>
													<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
												</svg>
											</button>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>

						<!-- Pagination -->
						{#if totalPages > 1}
							<div class="pagination">
								<button 
									class="btn btn-secondary" 
									disabled={currentPage === 1}
									on:click={() => handlePageChange(currentPage - 1)}
								>
									Previous
								</button>
								<span class="page-info">
									Page {currentPage} of {totalPages} ({totalVideos} total videos)
								</span>
								<button 
									class="btn btn-secondary" 
									disabled={currentPage === totalPages}
									on:click={() => handlePageChange(currentPage + 1)}
								>
									Next
								</button>
							</div>
						{/if}
					{/if}
				</div>
			</div>

		{:else if activeTab === 'subscribers'}
			<!-- Subscribers Tab -->
			<div class="subscribers-tab">
				{#if subscribersLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading subscriber data...</p>
					</div>
				{:else}
					<!-- Subscribers List -->
					<div class="subscribers-section">
						<div class="section-header">
							<h3>Subscribers</h3>
							<div class="subscribers-controls">
								<input 
									type="text" 
									placeholder="Search subscribers..." 
									bind:value={subscribersSearch}
									on:input={() => loadSubscribersData()}
								/>
								<select bind:value={subscribersStatus} on:change={() => loadSubscribersData()}>
									<option value="">All Status</option>
									<option value="active">Active</option>
									<option value="canceled">Canceled</option>
									<option value="past_due">Past Due</option>
								</select>
							</div>
						</div>
						
						<div class="subscribers-table-container">
							<table class="subscribers-table">
								<thead>
									<tr>
										<th>Customer ID</th>
										<th>Name</th>
										<th>Email</th>
										<th>Subscription Status</th>
										<th>Plan</th>
										<th>Created</th>
										<th>Actions</th>
									</tr>
								</thead>
								<tbody>
									{#each subscribers as subscriber}
										<tr>
											<td>{subscriber.id}</td>
											<td>{subscriber.name}</td>
											<td>{subscriber.email}</td>
											<td>
												<span class="status-badge {subscriber.subscription?.status === 'active' ? 'success' : 'warning'}">
													{subscriber.subscription?.status || 'No Subscription'}
												</span>
											</td>
											<td>
												<span class="plan-badge">
													{subscriber.subscription?.plan_name || 'N/A'}
												</span>
											</td>
											<td>{new Date(subscriber.created * 1000).toLocaleDateString()}</td>
											<td class="actions-cell">
												<button 
													class="btn btn-sm btn-secondary"
													on:click={() => {
														selectedCustomer = subscriber;
														showCustomerModal = true;
													}}
												>
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
														<path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
													</svg>
												</button>
											</td>
										</tr>
									{/each}
								</tbody>
							</table>
						</div>
						
						<!-- Pagination -->
						{#if subscribersPagination.total_pages > 1}
							<div class="pagination">
								<button 
									class="btn btn-secondary" 
									disabled={subscribersPagination.current_page === 1}
									on:click={() => {
										subscribersPagination.current_page--;
										loadSubscribersData();
									}}
								>
									Previous
								</button>
								<span class="pagination-info">
									Page {subscribersPagination.current_page} of {subscribersPagination.total_pages}
								</span>
								<button 
									class="btn btn-secondary" 
									disabled={subscribersPagination.current_page === subscribersPagination.total_pages}
									on:click={() => {
										subscribersPagination.current_page++;
										loadSubscribersData();
									}}
								>
									Next
								</button>
							</div>
						{/if}
					</div>
				{/if}
			</div>

		{:else if activeTab === 'financial'}
			<!-- Financial Tab -->
			<div class="financial-tab">
				{#if financialLoading}
					<div class="loading-container">
						<LoadingSpinner />
						<p>Loading financial data...</p>
					</div>
				{:else if financialMetrics}
					<!-- Financial Dashboard -->
					<div class="financial-dashboard">
						<!-- Key Metrics -->
						<div class="financial-metrics-grid">
							<div class="metric-card revenue">
								<div class="metric-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<line x1="12" y1="1" x2="12" y2="23"></line>
										<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"></path>
									</svg>
								</div>
								<div class="metric-content">
									<div class="metric-value">{formatCurrency(financialMetrics.total_revenue)}</div>
									<div class="metric-label">Total Revenue</div>
									<div class="metric-change positive">+{financialMetrics.revenue_growth_rate}%</div>
								</div>
							</div>
							
							<div class="metric-card mrr">
								<div class="metric-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2"></path>
										<rect x="8" y="2" width="8" height="4" rx="1" ry="1"></rect>
									</svg>
								</div>
								<div class="metric-content">
									<div class="metric-value">{formatCurrency(financialMetrics.monthly_recurring_revenue)}</div>
									<div class="metric-label">Monthly Recurring Revenue</div>
									<div class="metric-change positive">+{financialMetrics.subscription_growth_rate}%</div>
								</div>
							</div>
							
							<div class="metric-card customers">
								<div class="metric-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path>
										<circle cx="9" cy="7" r="4"></circle>
									</svg>
								</div>
								<div class="metric-content">
									<div class="metric-value">{financialMetrics.total_customers}</div>
									<div class="metric-label">Total Customers</div>
									<div class="metric-change positive">+{financialMetrics.customer_growth_rate}%</div>
								</div>
							</div>
							
							<div class="metric-card churn">
								<div class="metric-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
										<circle cx="12" cy="12" r="3"></circle>
									</svg>
								</div>
								<div class="metric-content">
									<div class="metric-value">{financialMetrics.churn_rate}%</div>
									<div class="metric-label">Churn Rate</div>
									<div class="metric-change negative">-{financialMetrics.churn_rate}%</div>
								</div>
							</div>
						</div>

						<!-- Revenue Chart -->
						<div class="revenue-chart-section">
							<h3>Revenue Trends</h3>
							<div class="chart-container">
								{#if revenueChart.length > 0}
									<div class="revenue-chart">
										<!-- Simple bar chart representation -->
										<div class="chart-bars">
											{#each revenueChart.slice(-7) as dataPoint}
												<div class="chart-bar">
													<div class="bar-fill" style="height: {(dataPoint.revenue / Math.max(...revenueChart.map(d => d.revenue))) * 100}%"></div>
													<div class="bar-label">{new Date(dataPoint.date).toLocaleDateString()}</div>
													<div class="bar-value">{formatCurrency(dataPoint.revenue)}</div>
												</div>
											{/each}
										</div>
									</div>
								{:else}
									<div class="no-data">
										<p>No revenue data available</p>
									</div>
								{/if}
							</div>
						</div>

						<!-- Recent Transactions -->
						<div class="transactions-section">
							<h3>Recent Transactions</h3>
							<div class="transactions-table-container">
								<table class="transactions-table">
									<thead>
										<tr>
											<th>Transaction ID</th>
											<th>Customer</th>
											<th>Amount</th>
											<th>Status</th>
											<th>Date</th>
										</tr>
									</thead>
									<tbody>
										{#each recentTransactions.slice(0, 10) as transaction}
											<tr>
												<td>{transaction.id}</td>
												<td>{transaction.customer_id}</td>
												<td>{formatCurrency(transaction.amount / 100)}</td>
												<td>
													<span class="status-badge {transaction.status === 'succeeded' ? 'success' : 'warning'}">
														{transaction.status}
													</span>
												</td>
												<td>{new Date(transaction.created * 1000).toLocaleDateString()}</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</div>

						<!-- Top Customers -->
						<div class="top-customers-section">
							<h3>Top Customers</h3>
							<div class="customers-grid">
								{#each topCustomers.slice(0, 5) as customerData}
									<div class="customer-card">
										<div class="customer-info">
											<div class="customer-name">{customerData.customer.name}</div>
											<div class="customer-email">{customerData.customer.email}</div>
										</div>
										<div class="customer-revenue">
											<div class="revenue-amount">{formatCurrency(customerData.total_spent)}</div>
											<div class="revenue-label">Total Spent</div>
										</div>
									</div>
								{/each}
							</div>
						</div>
					</div>
				{:else}
					<div class="error-state">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10"></circle>
							<line x1="15" y1="9" x2="9" y2="15"></line>
							<line x1="9" y1="9" x2="15" y2="15"></line>
						</svg>
						<h3>Failed to load financial data</h3>
						<p>Please try refreshing the page or contact support if the problem persists.</p>
						<button class="btn btn-primary" on:click={loadFinancialData}>Retry</button>
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Edit Modal -->
	{#if showEditModal}
		<div class="modal-overlay" on:click={handleModalClick}>
			<div class="modal-content">
				<div class="modal-header">
					<h2>Edit Video: {editingVideo?.Title}</h2>
					<button class="modal-close" on:click={closeEditModal}>×</button>
				</div>

				<!-- Read-only Info Section -->
				<div class="info-section">
					<h3>Video Information</h3>
					<div class="info-grid">
						<div class="info-item">
							<label>ID:</label>
							<span>{editingVideo?.ID}</span>
						</div>
						<div class="info-item">
							<label>Bunny Video ID:</label>
							<span>{editingVideo?.BunnyVideoID}</span>
						</div>
						<div class="info-item">
							<label>Resolution:</label>
							<span>{editingVideo?.Resolution}</span>
						</div>
						<div class="info-item">
							<label>Framerate:</label>
							<span>{editingVideo?.Framerate} fps</span>
						</div>
						<div class="info-item">
							<label>Encode Progress:</label>
							<span>{editingVideo?.EncodeProgress}%</span>
						</div>
						<div class="info-item">
							<label>Available Resolutions:</label>
							<span>{editingVideo?.AvailableResolutions?.join(', ')}</span>
						</div>
						<div class="info-item">
							<label>Average Watch Time:</label>
							<span>{formatDuration(editingVideo?.AverageWatchTime || 0)}</span>
						</div>
						<div class="info-item">
							<label>Total Watch Time:</label>
							<span>{formatDuration(editingVideo?.TotalWatchTime || 0)}</span>
						</div>
					</div>
				</div>

				<form on:submit|preventDefault={updateVideo}>
					<div class="form-grid">
						<!-- Basic Information Section -->
						<div class="form-section">
							<h3 class="section-title">Basic Information</h3>
							<div class="form-row">
								<div class="form-group">
									<label for="edit-title">Title *</label>
									<input type="text" id="edit-title" bind:value={editForm.Title} required />
								</div>
								<div class="form-group">
									<label for="edit-category">Category</label>
									<select id="edit-category" bind:value={editForm.Category}>
										<option value="">-- Select Category --</option>
										<option value="Music">Music</option>
										<option value="Gaming">Gaming</option>
										<option value="Education">Education</option>
										<option value="Entertainment">Entertainment</option>
										<option value="News">News</option>
										<option value="Sports">Sports</option>
										<option value="Technology">Technology</option>
										<option value="Travel">Travel</option>
										<option value="Other">Other</option>
									</select>
								</div>
							</div>
							
							<div class="form-group">
								<label for="edit-description">Description</label>
								<textarea id="edit-description" bind:value={editForm.Description} rows="3" placeholder="Enter video description..."></textarea>
							</div>
							
							<div class="form-group">
								<label for="edit-tags">Tags (comma-separated)</label>
								<input type="text" id="edit-tags" value={editForm.Tags?.join(', ') || ''} placeholder="tag1, tag2, tag3" on:input={(e) => {
									const target = e.target as HTMLInputElement;
									const tags = target.value.split(',').map((tag: string) => tag.trim()).filter((tag: string) => tag.length > 0);
									editForm.Tags = tags;
								}} />
							</div>
						</div>

						<!-- Metrics Section -->
						<div class="form-section">
							<h3 class="section-title">Metrics & Statistics</h3>
							<div class="form-row">
								<div class="form-group">
									<label for="edit-duration">Duration (seconds)</label>
									<input type="number" id="edit-duration" bind:value={editForm.Duration} min="0" />
								</div>
								<div class="form-group">
									<label for="edit-views">Views</label>
									<input type="number" id="edit-views" bind:value={editForm.Views} min="0" />
								</div>
							</div>
							
							<div class="form-row">
								<div class="form-group">
									<label for="edit-likes">Likes</label>
									<input type="number" id="edit-likes" bind:value={editForm.Likes} min="0" />
								</div>
								<div class="form-group">
									<label for="edit-collection-id">Collection ID</label>
									<input type="text" id="edit-collection-id" bind:value={editForm.CollectionID} placeholder="Enter collection ID" />
								</div>
							</div>
						</div>

						<!-- URLs Section -->
						<div class="form-section url-section">
							<h3 class="section-title">Video URLs</h3>
							<div class="url-grid">
								<div class="form-group">
									<label for="edit-thumbnail-url">Thumbnail URL</label>
									<input type="url" id="edit-thumbnail-url" bind:value={editForm.ThumbnailURL} placeholder="https://example.com/thumbnail.jpg" />
								</div>
								
								<div class="form-group">
									<label for="edit-video-url">Video URL</label>
									<input type="url" id="edit-video-url" bind:value={editForm.VideoURL} placeholder="https://example.com/video.mp4" />
								</div>
								
								<div class="form-group">
									<label for="edit-iframe-src">Iframe Source</label>
									<input type="url" id="edit-iframe-src" bind:value={editForm.IframeSrc} placeholder="https://example.com/embed/video" />
								</div>
								
								<div class="form-group">
									<label for="edit-playback-url">Playback URL</label>
									<input type="url" id="edit-playback-url" bind:value={editForm.PlaybackURL} placeholder="https://example.com/playback/video" />
								</div>
							</div>
						</div>

						<!-- Status & Settings Section -->
						<div class="form-section">
							<h3 class="section-title">Status & Settings</h3>
							<div class="form-row">
								<div class="form-group">
									<label for="edit-status">Status</label>
									<select id="edit-status" bind:value={editForm.Status}>
										<option value="ready">Ready</option>
										<option value="processing">Processing</option>
										<option value="transcoding">Transcoding</option>
										<option value="error">Error</option>
										<option value="upload_failed">Upload Failed</option>
										<option value="created">Created</option>
										<option value="uploaded">Uploaded</option>
									</select>
								</div>
								<div class="form-group">
									<label for="edit-sync-status">Sync Status</label>
									<select id="edit-sync-status" bind:value={editForm.SyncStatus}>
										<option value="synced">Synced</option>
										<option value="needs_attention">Needs Attention</option>
										<option value="conflict">Conflict</option>
									</select>
								</div>
							</div>
							
							<div class="form-group">
								<label for="edit-sync-notes">Sync Notes</label>
								<textarea id="edit-sync-notes" bind:value={editForm.SyncNotes} rows="2" placeholder="Add sync notes here..."></textarea>
							</div>
							
							<div class="form-group checkbox-group">
								<label for="edit-is-public">
									<input type="checkbox" id="edit-is-public" bind:checked={editForm.IsPublic} />
									Is Public
								</label>
							</div>
						</div>
					</div>
					
					<div class="modal-actions">
						<button type="button" class="btn btn-secondary" on:click={closeEditModal}>Cancel</button>
						<button type="submit" class="btn btn-primary">Save Changes</button>
					</div>
				</form>
			</div>
		</div>
	{/if}

	<!-- Bulk Edit Modal -->
	{#if showBulkEditModal}
		<div class="modal-overlay" on:click={handleModalClick}>
			<div class="modal-content">
				<div class="modal-header">
					<h2>Bulk Edit Videos ({selectedVideos.size} selected)</h2>
					<button class="modal-close" on:click={closeBulkEditModal}>×</button>
				</div>

				<div class="bulk-edit-info">
					<p>Enter the information you want to apply to all selected videos. Leave fields blank to keep existing values.</p>
				</div>

				<form on:submit|preventDefault={applyBulkEdit}>
					<div class="form-grid">
						<div class="form-section">
							<h3 class="section-title">Video Information</h3>
							<div class="form-row">
								<div class="form-group">
									<label for="bulk-title">Title</label>
									<input 
										type="text" 
										id="bulk-title" 
										bind:value={bulkEditForm.title} 
										placeholder="Leave blank to keep existing titles"
									/>
								</div>
								<div class="form-group">
									<label for="bulk-category">Category</label>
									<select id="bulk-category" bind:value={bulkEditForm.category}>
										<option value="">-- Keep Existing --</option>
										<option value="Music">Music</option>
										<option value="Gaming">Gaming</option>
										<option value="Education">Education</option>
										<option value="Entertainment">Entertainment</option>
										<option value="News">News</option>
										<option value="Sports">Sports</option>
										<option value="Technology">Technology</option>
										<option value="Travel">Travel</option>
										<option value="Other">Other</option>
									</select>
								</div>
							</div>
							
							<div class="form-group">
								<label for="bulk-description">Description</label>
								<textarea 
									id="bulk-description" 
									bind:value={bulkEditForm.description} 
									rows="3" 
									placeholder="Leave blank to keep existing descriptions"
								></textarea>
							</div>
							
							<div class="form-group">
								<label for="bulk-tags">Tags (comma-separated)</label>
								<input 
									type="text" 
									id="bulk-tags" 
									bind:value={bulkEditForm.tags} 
									placeholder="tag1, tag2, tag3 (leave blank to keep existing)"
								/>
							</div>
						</div>
					</div>
					
					<div class="modal-actions">
						<button type="button" class="btn btn-secondary" on:click={closeBulkEditModal}>Cancel</button>
						<button type="submit" class="btn btn-primary" disabled={isApplyingBulkEdit}>
							{#if isApplyingBulkEdit}
								<LoadingSpinner size="small" />
								Applying...
							{:else}
								Apply to {selectedVideos.size} Video(s)
							{/if}
						</button>
					</div>
				</form>
			</div>
		</div>
	{/if}
</div>

<style>
	.streaming-page {
		min-height: 100vh;
		background: var(--bg-primary);
		color: var(--text-primary);
		padding: 2rem;
	}

	.page-header {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.header-content {
		display: flex;
		justify-content: space-between;
		align-items: center;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.header-left h1 {
		font-size: 2rem;
		font-weight: 700;
		margin: 0 0 0.5rem 0;
		color: var(--text-primary);
	}

	.header-left p {
		color: var(--text-secondary);
		margin: 0;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	/* Main Tabs */
	.main-tabs {
		display: flex;
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 0.5rem;
		margin-bottom: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		gap: 0.5rem;
	}

	.main-tab {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 1rem 1.5rem;
		border: none;
		background: transparent;
		color: var(--text-secondary);
		border-radius: 10px;
		cursor: pointer;
		transition: all 0.3s ease;
		font-weight: 500;
		font-size: 1rem;
	}

	.main-tab svg {
		width: 20px;
		height: 20px;
	}

	.main-tab:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.main-tab.active {
		background: var(--primary-gradient);
		color: white;
		box-shadow: 0 4px 15px rgba(99, 102, 241, 0.3);
	}

	/* Tab Content */
	.tab-content {
		min-height: 500px;
	}

	/* Overview Tab */
	.overview-tab {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.overview-stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 1.5rem;
	}

	.overview-stat-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1.5rem;
		transition: all 0.3s ease;
	}

	.overview-stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.overview-stat-card.primary {
		border-left: 4px solid var(--primary-color);
	}

	.overview-stat-card.success {
		border-left: 4px solid #10b981;
	}

	.overview-stat-card.warning {
		border-left: 4px solid #f59e0b;
	}

	.overview-stat-card.info {
		border-left: 4px solid #3b82f6;
	}

	.stat-icon {
		width: 60px;
		height: 60px;
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
		background: var(--primary-gradient);
	}

	.overview-stat-card.success .stat-icon {
		background: linear-gradient(135deg, #10b981, #059669);
	}

	.overview-stat-card.warning .stat-icon {
		background: linear-gradient(135deg, #f59e0b, #d97706);
	}

	.overview-stat-card.info .stat-icon {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
		color: white;
	}

	.stat-content {
		flex: 1;
	}

	.stat-value {
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.stat-label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-weight: 600;
	}

	.stat-change {
		font-size: 0.8rem;
		font-weight: 600;
		color: #10b981;
	}

	/* Quick Actions */
	.quick-actions-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.quick-actions-section h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.quick-actions-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1rem;
	}

	.quick-action-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 1rem;
		padding: 1.5rem;
		background: var(--bg-glass-dark);
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.05);
		cursor: pointer;
		transition: all 0.3s ease;
		color: var(--text-primary);
		text-decoration: none;
	}

	.quick-action-card:hover {
		background: var(--bg-glass);
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.quick-action-card svg {
		width: 32px;
		height: 32px;
		color: var(--primary-color);
	}

	.quick-action-card span {
		font-weight: 600;
		text-align: center;
	}

	/* Recent Activity */
	.recent-activity-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.recent-activity-section h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.activity-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.activity-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 1rem;
		background: var(--bg-glass-dark);
		border-radius: 12px;
		transition: all 0.3s ease;
	}

	.activity-item:hover {
		background: var(--bg-glass);
	}

	.activity-icon {
		width: 40px;
		height: 40px;
		background: var(--primary-gradient);
		border-radius: 10px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.activity-icon svg {
		width: 20px;
		height: 20px;
		color: white;
	}

	.activity-content {
		flex: 1;
	}

	.activity-title {
		font-weight: 600;
		margin-bottom: 0.25rem;
		color: var(--text-primary);
	}

	.activity-description {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.25rem;
	}

	.activity-time {
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	/* Videos Tab */
	.videos-tab {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.video-stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 1.5rem;
	}

	.video-stat-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s ease;
	}

	.video-stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.video-stat-card.success {
		border-left: 4px solid #10b981;
	}

	.video-stat-card.warning {
		border-left: 4px solid #f59e0b;
	}

	/* Video Tabs */
	.video-tabs {
		display: flex;
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 0.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		gap: 0.5rem;
	}

	.video-tab {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1rem;
		border: none;
		background: transparent;
		color: var(--text-secondary);
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.3s ease;
		font-weight: 500;
		font-size: 0.9rem;
	}

	.video-tab svg {
		width: 16px;
		height: 16px;
	}

	.video-tab:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.video-tab.active {
		background: var(--primary-gradient);
		color: white;
		box-shadow: 0 4px 15px rgba(99, 102, 241, 0.3);
	}

	/* Subscribers Tab */
	.subscribers-tab {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.financial-overview {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.financial-overview h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.financial-stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
	}

	.financial-stat-card {
		background: var(--bg-glass-dark);
		border-radius: 15px;
		padding: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s ease;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.financial-stat-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.subscribers-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.subscribers-section h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	/* Common Components */
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 1rem;
	}

	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 1rem;
		text-align: center;
	}

	.error-state svg {
		width: 64px;
		height: 64px;
		color: #ef4444;
	}

	.search-section {
		margin-bottom: 1.5rem;
	}

	.search-input {
		width: 100%;
		max-width: 400px;
		padding: 1rem 1.25rem;
		border: 2px solid var(--border-color);
		border-radius: 10px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 1rem;
		transition: all 0.3s ease;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.15);
	}

	/* Tables */
	.videos-table-container,
	.subscribers-table-container {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.1);
		overflow: hidden;
	}

	.videos-table,
	.subscribers-table {
		width: 100%;
		border-collapse: collapse;
	}

	.videos-table th,
	.videos-table td,
	.subscribers-table th,
	.subscribers-table td {
		padding: 1rem;
		text-align: left;
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.videos-table th,
	.subscribers-table th {
		background: rgba(255, 255, 255, 0.05);
		font-weight: 600;
		color: var(--text-primary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
		font-size: 0.8rem;
	}

	.videos-table th.sortable,
	.subscribers-table th.sortable {
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.videos-table th.sortable:hover,
	.subscribers-table th.sortable:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.videos-table tr:hover,
	.subscribers-table tr:hover {
		background: rgba(255, 255, 255, 0.05);
	}

	.thumbnail-cell {
		width: 80px;
	}

	.td_cell {
		font-size: clamp(0.3rem, 0.6vw, 0.6rem);
		padding:0.5rem !important;
	}

	.video-thumbnail {
		width: 260px;
		height: auto;
		object-fit: cover;
		border-radius: 6px;
	}

	.no-thumbnail {
		width: 60px;
		height: 40px;
		background: var(--bg-secondary);
		border-radius: 6px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.7rem;
		color: var(--text-secondary);
	}

	.title-cell {
		max-width: 300px;
	}

	.video-title {
		font-weight: 600;
		margin-bottom: 0.25rem;
		color: var(--text-primary);
	}

	.video-description {
		font-size: 0.8rem;
		color: var(--text-secondary);
		line-height: 1.4;
	}

	.status-badge {
		display: flex;
		flex-wrap: nowrap;
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: clamp(0.3rem, 0.6vw, 0.6rem);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.success {
		background: rgba(16, 185, 129, 0.2);
		color: #10b981;
	}

	.status-badge.warning {
		background: rgba(245, 158, 11, 0.2);
		color: #f59e0b;
	}

	.status-badge.error {
		background: rgba(239, 68, 68, 0.2);
		color: #ef4444;
	}

	.status-badge.info {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
	}

	.plan-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.plan-badge.premium {
		background: rgba(245, 158, 11, 0.2);
		color: #f59e0b;
	}

	.plan-badge.basic {
		background: rgba(59, 130, 246, 0.2);
		color: #3b82f6;
	}

	.actions-cell {
		display: flex;
		flex-direction: column;
		height: 185.25px;
		gap: 3rem;
		margin-top: 1rem;
		vertical-align: top;
	}

	/* Buttons */
	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		font-size: 0.9rem;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
	}

	.btn-primary {
		background: var(--primary-gradient);
		color: white;
	}

	.btn-primary:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(99, 102, 241, 0.3);
	}

	.btn-secondary {
		background: var(--bg-glass);
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.btn-secondary:hover {
		background: rgba(255, 255, 255, 0.1);
		transform: translateY(-2px);
	}

	.btn-danger {
		background: linear-gradient(135deg, #ef4444, #dc2626);
		color: white;
	}

	.btn-danger:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(239, 68, 68, 0.3);
	}

	.btn-sm {
		padding: 0.5rem;
		font-size: 0.8rem;
	}

	.btn-sm svg {
		width: 16px;
		height: 16px;
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 1rem;
		padding: 2rem;
	}

	.page-info {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	/* Empty State */
	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 300px;
		gap: 1rem;
		text-align: center;
	}

	.empty-state svg {
		width: 64px;
		height: 64px;
		color: var(--text-secondary);
	}

	.empty-state h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0;
		color: var(--text-primary);
	}

	.empty-state p {
		color: var(--text-secondary);
		max-width: 400px;
		line-height: 1.5;
	}

	/* Modal Styles */
	.modal-overlay {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 2rem;
	}

	.modal-content {
		background: var(--bg-primary);
		border: 1px solid var(--border-color);
		border-radius: 12px;
		padding: 2.5rem;
		width: 95%;
		max-width: 1800px;
		max-height: 90vh;
		overflow-y: auto;
		display: flex;
		flex-direction: column;
		gap: 2rem;
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.25);
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.modal-header h2 {
		margin: 0;
		color: white;
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		font-size: 2rem;
		cursor: pointer;
		padding: 0;
		width: 40px;
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		transition: all 0.3s ease;
	}

	.modal-close:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	/* Info Section */
	.info-section {
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 10px;
		padding: 2rem;
		margin-bottom: 2rem;
		transition: all 0.2s ease;
	}

	.info-section:hover {
		border-color: var(--primary-color);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.info-section h3 {
		margin-top: 0;
		margin-bottom: 1.5rem;
		color: var(--text-primary);
		font-size: 1.25rem;
		font-weight: 600;
		padding-bottom: 0.75rem;
		border-bottom: 2px solid var(--border-color);
		position: relative;
	}

	.info-section h3::after {
		content: '';
		position: absolute;
		bottom: -2px;
		left: 0;
		width: 60px;
		height: 2px;
		background: var(--primary-color);
		border-radius: 1px;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
		justify-content: space-between;
		align-items: center;
		text-align: center;
		padding: 0.75rem;
		background: var(--bg-primary);
		border-radius: 6px;
		border: 1px solid var(--border-color);
	}

	.info-item label {
		font-size: 0.875rem;
		color: var(--text-secondary);
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.info-item span {
		font-size: 1rem;
		color: var(--text-primary);
		font-weight: 600;
	}

	/* Form Styles */
	.form-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: 2rem;
		margin-bottom: 2rem;
	}

	.form-section {
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 10px;
		padding: 2rem;
		transition: all 0.2s ease;
	}

	.form-section:hover {
		border-color: var(--primary-color);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.section-title {
		margin-top: 0;
		margin-bottom: 1.5rem;
		color: var(--text-primary);
		font-size: 1.25rem;
		font-weight: 600;
		padding-bottom: 0.75rem;
		border-bottom: 2px solid var(--border-color);
		position: relative;
	}

	.section-title::after {
		content: '';
		position: absolute;
		bottom: -2px;
		left: 0;
		width: 60px;
		height: 2px;
		background: var(--primary-color);
		border-radius: 1px;
	}

	.form-row {
		display: flex;
		gap: 1.5rem;
		margin-bottom: 1.5rem;
	}

	.form-group {
		flex: 1;
		display: flex;
		flex-direction: column;
	}

	.form-group label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.form-group input[type="text"],
	.form-group input[type="number"],
	.form-group input[type="url"],
	.form-group input[type="checkbox"],
	.form-group select,
	.form-group textarea {
		padding: 1rem 1.25rem;
		border: 2px solid var(--border-color);
		border-radius: 8px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 1rem;
		transition: all 0.3s ease;
		width: 100%;
		box-sizing: border-box;
		font-weight: 500;
	}

	.form-group input[type="text"]:focus,
	.form-group input[type="number"]:focus,
	.form-group input[type="url"]:focus,
	.form-group select:focus,
	.form-group textarea:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.15);
		background: var(--bg-primary);
		transform: translateY(-1px);
	}

	.form-group input[type="url"] {
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		font-size: 0.875rem;
		background: var(--bg-primary);
	}

	.form-group textarea {
		resize: vertical;
		min-height: 80px;
		line-height: 1.5;
	}

	.form-group input[type="checkbox"] {
		width: auto;
		margin-top: 0.5rem;
	}

	.checkbox-group {
		flex-direction: row;
		align-items: center;
		gap: 0.5rem;
	}

	.checkbox-group label {
		margin-bottom: 0;
		text-transform: none;
		letter-spacing: normal;
	}

	.url-section .url-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 1.5rem;
		margin-top: 2rem;
		padding-top: 2rem;
		border-top: 2px solid var(--border-color);
	}

	.modal-actions .btn {
		padding: 1rem 2rem;
		font-weight: 600;
		font-size: 1rem;
		border-radius: 8px;
		transition: all 0.3s ease;
		min-width: 120px;
	}

	.modal-actions .btn:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(0, 0, 0, 0.15);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.streaming-page {
			padding: 1rem;
		}

		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.header-actions {
			justify-content: center;
		}

		.main-tabs {
			flex-direction: column;
		}

		.main-tab {
			justify-content: center;
		}

		.overview-stats-grid,
		.video-stats-grid,
		.financial-stats-grid {
			grid-template-columns: 1fr;
		}

		.quick-actions-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.video-tabs {
			flex-direction: column;
		}

		.form-grid {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}

		.url-section .url-grid {
			grid-template-columns: 1fr;
			gap: 1rem;
		}

		.form-section {
			padding: 1.5rem;
		}

		.form-row {
			flex-direction: column;
			gap: 1rem;
		}

		.form-group {
			margin-bottom: 1rem;
		}

		.modal-content {
			width: 98%;
			max-width: none;
			padding: 1.5rem;
			margin: 0.5rem;
		}

		.modal-actions {
			flex-direction: column;
			gap: 1rem;
		}

		.modal-actions .btn {
			width: 100%;
			min-width: auto;
		}

		.videos-table,
		.subscribers-table {
			font-size: 0.8rem;
		}

		.videos-table th,
		.videos-table td,
		.subscribers-table th,
		.subscribers-table td {
			padding: 0.5rem;
		}

		.thumbnail-cell {
			width: 60px;
		}

		.video-thumbnail,
		.no-thumbnail {
			width: 40px;
			height: 30px;
		}

		.title-cell {
			max-width: 150px;
		}
	}

	@media (max-width: 480px) {
		.quick-actions-grid {
			grid-template-columns: 1fr;
		}

		.overview-stat-card,
		.video-stat-card,
		.financial-stat-card {
			flex-direction: column;
			text-align: center;
		}

		.stat-icon {
			width: 50px;
			height: 50px;
		}

		.stat-icon svg {
			width: 24px;
			height: 24px;
		}
	}

	/* Financial Tab Styles */
	.financial-tab {
		padding: 2rem;
	}

	.financial-dashboard {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.financial-metrics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 1.5rem;
	}

	.metric-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		display: flex;
		align-items: center;
		gap: 1rem;
		transition: all 0.3s ease;
	}

	.metric-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.metric-card.revenue {
		border-left: 4px solid #10b981;
	}

	.metric-card.mrr {
		border-left: 4px solid #3b82f6;
	}

	.metric-card.customers {
		border-left: 4px solid #f59e0b;
	}

	.metric-card.churn {
		border-left: 4px solid #ef4444;
	}

	.metric-icon {
		width: 60px;
		height: 60px;
		border-radius: 15px;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.metric-card.revenue .metric-icon {
		background: linear-gradient(135deg, #10b981, #047857);
	}

	.metric-card.mrr .metric-icon {
		background: linear-gradient(135deg, #3b82f6, #1d4ed8);
	}

	.metric-card.customers .metric-icon {
		background: linear-gradient(135deg, #f59e0b, #d97706);
	}

	.metric-card.churn .metric-icon {
		background: linear-gradient(135deg, #ef4444, #dc2626);
	}

	.metric-icon svg {
		width: 28px;
		height: 28px;
		color: white;
	}

	.metric-content {
		flex: 1;
	}

	.metric-value {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.metric-label {
		font-size: 0.9rem;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}

	.metric-change {
		font-size: 0.8rem;
		font-weight: 600;
	}

	.metric-change.positive {
		color: #10b981;
	}

	.metric-change.negative {
		color: #ef4444;
	}

	/* Revenue Chart */
	.revenue-chart-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.revenue-chart-section h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.chart-container {
		height: 300px;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.revenue-chart {
		width: 100%;
		height: 100%;
	}

	.chart-bars {
		display: flex;
		align-items: end;
		justify-content: space-between;
		height: 200px;
		gap: 1rem;
		padding: 1rem 0;
	}

	.chart-bar {
		flex: 1;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.5rem;
	}

	.bar-fill {
		width: 100%;
		background: linear-gradient(135deg, var(--primary-color), var(--primary-dark));
		border-radius: 4px 4px 0 0;
		min-height: 4px;
		transition: all 0.3s ease;
	}

	.bar-label {
		font-size: 0.75rem;
		color: var(--text-secondary);
		text-align: center;
	}

	.bar-value {
		font-size: 0.8rem;
		font-weight: 600;
		color: var(--text-primary);
	}

	.no-data {
		text-align: center;
		color: var(--text-secondary);
	}

	/* Transactions Section */
	.transactions-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.transactions-section h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.transactions-table-container {
		overflow-x: auto;
		border-radius: 10px;
		border: 1px solid var(--border-color);
	}

	.transactions-table {
		width: 100%;
		border-collapse: collapse;
		background: var(--bg-primary);
	}

	.transactions-table th,
	.transactions-table td {
		padding: 1rem;
		text-align: left;
		border-bottom: 1px solid var(--border-color);
	}

	.transactions-table th {
		background: var(--bg-glass-dark);
		font-weight: 600;
		color: var(--text-primary);
		font-size: 0.9rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.transactions-table td {
		color: var(--text-secondary);
		font-size: 0.9rem;
	}

	.transactions-table tr:hover {
		background: var(--bg-glass-dark);
	}

	/* Top Customers Section */
	.top-customers-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.top-customers-section h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0 0 1.5rem 0;
		color: var(--text-primary);
	}

	.customers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1rem;
	}

	.customer-card {
		background: var(--bg-glass-dark);
		border-radius: 10px;
		padding: 1.5rem;
		border: 1px solid var(--border-color);
		display: flex;
		justify-content: space-between;
		align-items: center;
		transition: all 0.3s ease;
	}

	.customer-card:hover {
		background: var(--bg-glass);
		transform: translateY(-2px);
		box-shadow: 0 8px 20px rgba(0, 0, 0, 0.1);
	}

	.customer-info {
		flex: 1;
	}

	.customer-name {
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.25rem;
	}

	.customer-email {
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.customer-revenue {
		text-align: right;
	}

	.revenue-amount {
		font-size: 1.1rem;
		font-weight: 700;
		color: var(--primary-color);
		margin-bottom: 0.25rem;
	}

	.revenue-label {
		font-size: 0.75rem;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	/* Subscribers Tab Updates */
	.subscribers-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.section-header h3 {
		font-size: 1.5rem;
		font-weight: 600;
		margin: 0;
		color: var(--text-primary);
	}

	.subscribers-controls {
		display: flex;
		gap: 1rem;
		align-items: center;
	}

	.subscribers-controls input,
	.subscribers-controls select {
		padding: 0.75rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 0.9rem;
		transition: all 0.3s ease;
	}

	.subscribers-controls input:focus,
	.subscribers-controls select:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 3px rgba(99, 102, 241, 0.1);
	}

	.subscribers-controls input {
		min-width: 200px;
	}

	.subscribers-controls select {
		min-width: 120px;
	}

	/* Status Badge Updates */
	.status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.status-badge.success {
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
		border: 1px solid rgba(16, 185, 129, 0.2);
	}

	.status-badge.warning {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
		border: 1px solid rgba(245, 158, 11, 0.2);
	}

	.plan-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 20px;
		font-size: 0.75rem;
		font-weight: 600;
		background: rgba(99, 102, 241, 0.1);
		color: var(--primary-color);
		border: 1px solid rgba(99, 102, 241, 0.2);
	}

	/* Responsive Design Updates */
	@media (max-width: 768px) {
		.financial-metrics-grid {
			grid-template-columns: 1fr;
		}

		.metric-card {
			flex-direction: column;
			text-align: center;
		}

		.chart-bars {
			flex-direction: column;
			height: auto;
			gap: 0.5rem;
		}

		.chart-bar {
			flex-direction: row;
			justify-content: space-between;
			width: 100%;
		}

		.bar-fill {
			width: 60%;
			height: 20px;
			border-radius: 4px;
		}

		.customers-grid {
			grid-template-columns: 1fr;
		}

		.customer-card {
			flex-direction: column;
			text-align: center;
			gap: 1rem;
		}

		.section-header {
			flex-direction: column;
			align-items: stretch;
		}

		.subscribers-controls {
			flex-direction: column;
		}

		.subscribers-controls input,
		.subscribers-controls select {
			width: 100%;
		}

		.transactions-table {
			font-size: 0.8rem;
		}

		.transactions-table th,
		.transactions-table td {
			padding: 0.5rem;
		}
	}

	/* Bulk Operations Mobile Styles */
	.bulk-operations {
		flex-direction: column;
		align-items: stretch;
	}

	.bulk-selection {
		justify-content: space-between;
	}

	.bulk-actions {
		justify-content: center;
	}

	.videos-table th.checkbox-header,
	.videos-table td.checkbox-cell {
		width: 30px;
	}

	.videos-table td.checkbox-cell input[type="checkbox"] {
		width: 14px;
		height: 14px;
	}
</style>