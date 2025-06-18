<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/auth';
	import { goto } from '$app/navigation';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

	interface User {
		id: number;
		email: string;
		firstName: string;
		lastName: string;
		role: string;
		emailVerified: boolean;
	}

	interface Video {
		id: number;
		title: string;
		description: string;
		thumbnail: string;
		duration: string;
		views: number;
		likes: number;
		createdAt: string;
		category: string;
	}

	interface Activity {
		type: 'watched' | 'liked' | 'commented' | 'favorited';
		video: Video;
		timestamp: string;
	}

	interface UserStats {
		videosWatched: number;
		totalWatchTime: string;
		favoriteVideos: number;
		commentsPosted: number;
		joinedDate: string;
		subscriptionStatus: string;
	}

	let user: User | null = null;
	let isAuthenticated = false;
	let loading = true;
	let userStats: UserStats | null = null;
	let recentActivity: Activity[] = [];
	let favoriteVideos: Video[] = [];
	let recommendedVideos: Video[] = [];
	let continueWatching: Video[] = [];

	// Subscribe to auth store
	auth.subscribe(state => {
		user = state.user;
		isAuthenticated = state.isAuthenticated;
	});

	onMount(async () => {
		if (!isAuthenticated) {
			goto('/login');
			return;
		}

		try {
			loading = true;
			
			// Mock data for development - would be replaced with real API calls
			await new Promise(resolve => setTimeout(resolve, 1000));
			
			userStats = {
				videosWatched: 47,
				totalWatchTime: '12h 34m',
				favoriteVideos: 8,
				commentsPosted: 23,
				joinedDate: '2024-01-15',
				subscriptionStatus: 'Premium'
			};

			recentActivity = [
				{
					type: 'watched',
					video: {
						id: 1,
						title: 'Archaeological Evidence for the Book of Mormon',
						description: 'Exploring recent archaeological discoveries...',
						thumbnail: '/api/placeholder/320/180',
						duration: '24:15',
						views: 15420,
						likes: 892,
						createdAt: '2024-06-15',
						category: 'Archaeology'
					},
					timestamp: '2 hours ago'
				},
				{
					type: 'liked',
					video: {
						id: 2,
						title: 'DNA and the Book of Mormon',
						description: 'Scientific analysis of genetic evidence...',
						thumbnail: '/api/placeholder/320/180',
						duration: '18:32',
						views: 8750,
						likes: 456,
						createdAt: '2024-06-10',
						category: 'Science'
					},
					timestamp: '1 day ago'
				},
				{
					type: 'commented',
					video: {
						id: 3,
						title: 'Ancient American Civilizations',
						description: 'Comparing Book of Mormon civilizations...',
						thumbnail: '/api/placeholder/320/180',
						duration: '31:45',
						views: 12340,
						likes: 678,
						createdAt: '2024-06-08',
						category: 'History'
					},
					timestamp: '3 days ago'
				}
			];

			favoriteVideos = [
				{
					id: 4,
					title: 'Nephite Metallurgy Evidence',
					description: 'Analysis of ancient metalworking...',
					thumbnail: '/api/placeholder/320/180',
					duration: '22:10',
					views: 9876,
					likes: 543,
					createdAt: '2024-06-05',
					category: 'Archaeology'
				},
				{
					id: 5,
					title: 'Book of Mormon Geography',
					description: 'Mapping ancient civilizations...',
					thumbnail: '/api/placeholder/320/180',
					duration: '28:45',
					views: 11234,
					likes: 721,
					createdAt: '2024-06-01',
					category: 'Geography'
				}
			];

			recommendedVideos = [
				{
					id: 6,
					title: 'New Archaeological Discoveries',
					description: 'Latest findings in Mesoamerica...',
					thumbnail: '/api/placeholder/320/180',
					duration: '19:30',
					views: 5432,
					likes: 298,
					createdAt: '2024-06-18',
					category: 'Archaeology'
				},
				{
					id: 7,
					title: 'Linguistic Analysis of Ancient Texts',
					description: 'Comparing Hebrew and Nephite languages...',
					thumbnail: '/api/placeholder/320/180',
					duration: '26:15',
					views: 7890,
					likes: 412,
					createdAt: '2024-06-16',
					category: 'Linguistics'
				}
			];

			continueWatching = [
				{
					id: 8,
					title: 'The Jaredite Civilization',
					description: 'Exploring the earliest Book of Mormon peoples...',
					thumbnail: '/api/placeholder/320/180',
					duration: '35:20',
					views: 6789,
					likes: 345,
					createdAt: '2024-06-12',
					category: 'History'
				}
			];

		} catch (error) {
			showToast('Failed to load dashboard data', 'error');
			console.error('Error loading dashboard:', error);
		} finally {
			loading = false;
		}
	});

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function formatNumber(num: number) {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function getActivityIcon(type: string) {
		switch (type) {
			case 'watched':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="23,7 16,12 23,17 23,7"></polygon>
					<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
				</svg>`;
			case 'liked':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M20.84 4.61a5.5 5.5 0 0 0-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 0 0-7.78 7.78l1.06 1.06L12 21.23l7.78-7.78 1.06-1.06a5.5 5.5 0 0 0 0-7.78z"></path>
				</svg>`;
			case 'commented':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
				</svg>`;
			case 'favorited':
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"></polygon>
				</svg>`;
			default:
				return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="12" cy="12" r="10"></circle>
					<line x1="12" y1="16" x2="12" y2="12"></line>
					<line x1="12" y1="8" x2="12.01" y2="8"></line>
				</svg>`;
		}
	}

	function getActivityText(activity: Activity) {
		switch (activity.type) {
			case 'watched':
				return `Watched "${activity.video.title}"`;
			case 'liked':
				return `Liked "${activity.video.title}"`;
			case 'commented':
				return `Commented on "${activity.video.title}"`;
			case 'favorited':
				return `Added "${activity.video.title}" to favorites`;
			default:
				return `Interacted with "${activity.video.title}"`;
		}
	}
</script>

<svelte:head>
	<title>Dashboard - BOME</title>
</svelte:head>

{#if loading}
	<div class="loading-container">
		<LoadingSpinner size="large" color="primary" />
		<p>Loading your dashboard...</p>
	</div>
{:else}
	<div class="dashboard">
		<!-- Welcome Header -->
		<div class="welcome-section glass">
			<div class="welcome-content">
				<h1>Welcome back, {user?.firstName || 'User'}!</h1>
				<p>Continue your journey exploring Book of Mormon evidences</p>
			</div>
			<div class="user-avatar-large">
				{user?.firstName?.charAt(0) || 'U'}{user?.lastName?.charAt(0) || ''}
			</div>
		</div>

		<!-- Stats Overview -->
		<div class="stats-grid">
			<div class="stat-card glass">
				<div class="stat-icon videos">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="23,7 16,12 23,17 23,7"></polygon>
						<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{userStats?.videosWatched || 0}</div>
					<div class="stat-label">Videos Watched</div>
				</div>
			</div>

			<div class="stat-card glass">
				<div class="stat-icon time">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"></circle>
						<polyline points="12,6 12,12 16,14"></polyline>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{userStats?.totalWatchTime || '0h 0m'}</div>
					<div class="stat-label">Watch Time</div>
				</div>
			</div>

			<div class="stat-card glass">
				<div class="stat-icon favorites">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"></polygon>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{userStats?.favoriteVideos || 0}</div>
					<div class="stat-label">Favorites</div>
				</div>
			</div>

			<div class="stat-card glass">
				<div class="stat-icon subscription">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
						<polyline points="16,21 12,17 8,21"></polyline>
						<polyline points="12,17 12,3"></polyline>
					</svg>
				</div>
				<div class="stat-content">
					<div class="stat-value">{userStats?.subscriptionStatus || 'Free'}</div>
					<div class="stat-label">Subscription</div>
				</div>
			</div>
		</div>

		<!-- Main Content Grid -->
		<div class="content-grid">
			<!-- Continue Watching -->
			{#if continueWatching.length > 0}
				<div class="section-card glass">
					<div class="section-header">
						<h2>Continue Watching</h2>
						<a href="/videos" class="view-all-link">View All</a>
					</div>
					<div class="video-grid">
						{#each continueWatching as video}
							<div class="video-card">
								<div class="video-thumbnail">
									<img src={video.thumbnail} alt={video.title} />
									<div class="video-duration">{video.duration}</div>
									<div class="progress-bar">
										<div class="progress-fill" style="width: 65%"></div>
									</div>
								</div>
								<div class="video-info">
									<h3 class="video-title">{video.title}</h3>
									<p class="video-meta">{formatNumber(video.views)} views • {video.category}</p>
								</div>
							</div>
						{/each}
					</div>
				</div>
			{/if}

			<!-- Recent Activity -->
			<div class="section-card glass">
				<div class="section-header">
					<h2>Recent Activity</h2>
				</div>
				<div class="activity-list">
					{#each recentActivity as activity}
						<div class="activity-item">
							<div class="activity-icon">
								{@html getActivityIcon(activity.type)}
							</div>
							<div class="activity-content">
								<div class="activity-text">{getActivityText(activity)}</div>
								<div class="activity-time">{activity.timestamp}</div>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Recommended Videos -->
			<div class="section-card glass">
				<div class="section-header">
					<h2>Recommended for You</h2>
					<a href="/videos" class="view-all-link">View All</a>
				</div>
				<div class="video-grid">
					{#each recommendedVideos as video}
						<div class="video-card">
							<div class="video-thumbnail">
								<img src={video.thumbnail} alt={video.title} />
								<div class="video-duration">{video.duration}</div>
							</div>
							<div class="video-info">
								<h3 class="video-title">{video.title}</h3>
								<p class="video-meta">{formatNumber(video.views)} views • {video.category}</p>
							</div>
						</div>
					{/each}
				</div>
			</div>

			<!-- Favorite Videos -->
			<div class="section-card glass">
				<div class="section-header">
					<h2>Your Favorites</h2>
					<a href="/favorites" class="view-all-link">View All</a>
				</div>
				<div class="video-grid">
					{#each favoriteVideos as video}
						<div class="video-card">
							<div class="video-thumbnail">
								<img src={video.thumbnail} alt={video.title} />
								<div class="video-duration">{video.duration}</div>
							</div>
							<div class="video-info">
								<h3 class="video-title">{video.title}</h3>
								<p class="video-meta">{formatNumber(video.views)} views • {video.category}</p>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="quick-actions glass">
			<h2>Quick Actions</h2>
			<div class="action-buttons">
				<a href="/videos" class="action-btn">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="23,7 16,12 23,17 23,7"></polygon>
						<rect x="1" y="5" width="15" height="14" rx="2" ry="2"></rect>
					</svg>
					Browse Videos
				</a>
				<a href="/categories" class="action-btn">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<rect x="3" y="3" width="7" height="7"></rect>
						<rect x="14" y="3" width="7" height="7"></rect>
						<rect x="14" y="14" width="7" height="7"></rect>
						<rect x="3" y="14" width="7" height="7"></rect>
					</svg>
					Explore Categories
				</a>
				<a href="/favorites" class="action-btn">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<polygon points="12,2 15.09,8.26 22,9.27 17,14.14 18.18,21.02 12,17.77 5.82,21.02 7,14.14 2,9.27 8.91,8.26 12,2"></polygon>
					</svg>
					My Favorites
				</a>
				<a href="/subscription" class="action-btn">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M20 7H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2z"></path>
						<polyline points="16,21 12,17 8,21"></polyline>
						<polyline points="12,17 12,3"></polyline>
					</svg>
					Upgrade Plan
				</a>
			</div>
		</div>
	</div>
{/if}

<style>
	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		min-height: 60vh;
		gap: var(--space-lg);
	}

	.loading-container p {
		font-size: var(--text-lg);
		color: var(--text-secondary);
	}

	.dashboard {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
		padding: var(--space-xl);
		max-width: 1200px;
		margin: 0 auto;
	}

	.welcome-section {
		padding: var(--space-2xl);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: space-between;
		background: var(--primary-gradient);
		color: var(--white);
	}

	.welcome-content h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		margin-bottom: var(--space-sm);
	}

	.welcome-content p {
		font-size: var(--text-lg);
		opacity: 0.9;
	}

	.user-avatar-large {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		background: rgba(255, 255, 255, 0.2);
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--white);
		border: 3px solid rgba(255, 255, 255, 0.3);
	}

	.stats-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: var(--space-xl);
	}

	.stat-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		gap: var(--space-lg);
		transition: all var(--transition-normal);
	}

	.stat-card:hover {
		transform: translateY(-4px);
		box-shadow: var(--shadow-lg);
	}

	.stat-icon {
		width: 56px;
		height: 56px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		box-shadow: var(--shadow-md);
	}

	.stat-icon.videos {
		background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
	}

	.stat-icon.time {
		background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
	}

	.stat-icon.favorites {
		background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
	}

	.stat-icon.subscription {
		background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
	}

	.stat-icon svg {
		width: 28px;
		height: 28px;
		color: var(--white);
	}

	.stat-value {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.content-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: var(--space-xl);
	}

	.section-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-xl);
	}

	.section-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.view-all-link {
		color: var(--primary);
		text-decoration: none;
		font-size: var(--text-sm);
		font-weight: 600;
		transition: all var(--transition-normal);
	}

	.view-all-link:hover {
		color: var(--primary-dark);
		transform: translateX(4px);
	}

	.video-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.video-card {
		cursor: pointer;
		transition: all var(--transition-normal);
	}

	.video-card:hover {
		transform: translateY(-4px);
	}

	.video-thumbnail {
		position: relative;
		width: 100%;
		aspect-ratio: 16/9;
		border-radius: var(--radius-lg);
		overflow: hidden;
		margin-bottom: var(--space-md);
	}

	.video-thumbnail img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.video-duration {
		position: absolute;
		bottom: var(--space-sm);
		right: var(--space-sm);
		background: rgba(0, 0, 0, 0.8);
		color: var(--white);
		padding: 2px 6px;
		border-radius: var(--radius-sm);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.progress-bar {
		position: absolute;
		bottom: 0;
		left: 0;
		width: 100%;
		height: 4px;
		background: rgba(0, 0, 0, 0.3);
	}

	.progress-fill {
		height: 100%;
		background: var(--primary);
		transition: width var(--transition-normal);
	}

	.video-title {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xs);
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.video-meta {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin: 0;
	}

	.activity-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.activity-item {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.activity-icon {
		width: 40px;
		height: 40px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.activity-icon svg {
		width: 20px;
		height: 20px;
		color: var(--text-primary);
	}

	.activity-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
		margin-bottom: var(--space-xs);
	}

	.activity-time {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.quick-actions {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
	}

	.quick-actions h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-xl);
	}

	.action-buttons {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.action-btn {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-lg);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		text-decoration: none;
		color: var(--text-primary);
		font-weight: 500;
		transition: all var(--transition-normal);
	}

	.action-btn:hover {
		background: var(--bg-glass-dark);
		transform: translateY(-2px);
		box-shadow: var(--shadow-md);
	}

	.action-btn svg {
		width: 20px;
		height: 20px;
		color: var(--primary);
	}

	@media (max-width: 1024px) {
		.content-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.dashboard {
			padding: var(--space-lg);
		}

		.welcome-section {
			flex-direction: column;
			text-align: center;
			gap: var(--space-lg);
		}

		.stats-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.video-grid {
			grid-template-columns: 1fr;
		}

		.action-buttons {
			grid-template-columns: repeat(2, 1fr);
		}
	}

	@media (max-width: 480px) {
		.stats-grid {
			grid-template-columns: 1fr;
		}

		.action-buttons {
			grid-template-columns: 1fr;
		}

		.stat-card {
			flex-direction: column;
			text-align: center;
			gap: var(--space-md);
		}
	}
</style> 