import { writable } from 'svelte/store';

// Types
interface YouTubeVideo {
	id: string;
	title: string;
	description: string;
	published_at: string;
	updated_at: string;
	thumbnail_url: string;
	video_url: string;
	embed_url: string;
	duration: string;
	view_count: number;
	created_at: string;
}

interface YouTubeVideosResponse {
	videos: YouTubeVideo[];
	last_updated: string;
	total_count: number;
	channel?: ChannelInfo;
}

interface ChannelInfo {
	id: string;
	title: string;
	description: string;
	subscriber_count: number;
	video_count: number;
	view_count: number;
	published_at: string;
	country: string;
	custom_url: string;
	thumbnail_url: string;
}

interface YouTubeStatus {
	channel_id: string;
	channel_title: string;
	total_videos: number;
	last_updated: string;
	api_version: string;
	mock_mode: boolean;
	status: string;
	data_source: string;
}

// Store state
interface YouTubeState {
	videos: YouTubeVideo[];
	currentVideo: YouTubeVideo | null;
	channelInfo: ChannelInfo | null;
	status: YouTubeStatus | null;
	categories: string[];
	tags: string[];
	loading: boolean;
	error: string | null;
	searchQuery: string;
	searchResults: YouTubeVideo[];
	thumbnailStore: Map<string, string>;
}

// Initial state
const initialState: YouTubeState = {
	videos: [],
	currentVideo: null,
	channelInfo: null,
	status: null,
	categories: [],
	tags: [],
	loading: false,
	error: null,
	searchQuery: '',
	searchResults: [],
	thumbnailStore: new Map<string, string>()
};

// Create stores
export const youtubeStore = writable<YouTubeState>(initialState);

// API Configuration
const API_BASE_URL = '/api/v1/youtube';

// Helper function for API calls
async function apiCall<T>(endpoint: string): Promise<T> {
	try {
		const response = await fetch(`${API_BASE_URL}${endpoint}`);
		
		if (!response.ok) {
			throw new Error(`HTTP error! status: ${response.status}`);
		}
		
		return await response.json();
	} catch (error) {
		console.error(`API call failed for ${endpoint}:`, error);
		throw error;
	}
}

// YouTube Store Actions
export const youtubeActions = {
	// Set loading state
	setLoading: (loading: boolean) => {
		youtubeStore.update((state: YouTubeState) => ({ ...state, loading }));
	},

	// Set error state
	setError: (error: string | null) => {
		youtubeStore.update((state: YouTubeState) => ({ ...state, error }));
	},

	// Get latest videos
	async getLatestVideos(limit: number = 10): Promise<void> {
		youtubeActions.setLoading(true);
		youtubeActions.setError(null);

		try {
			const response = await apiCall<YouTubeVideosResponse>(`/videos/latest?limit=${limit}`);
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				videos: response.videos,
				loading: false
			}));
		} catch (error) {
			youtubeActions.setError(error instanceof Error ? error.message : 'Failed to load videos');
			youtubeActions.setLoading(false);
		}
	},

	// Get all videos with pagination
	async getAllVideos(limit: number = 20): Promise<void> {
		youtubeActions.setLoading(true);
		youtubeActions.setError(null);

		try {
			const response = await apiCall<YouTubeVideosResponse>(`/videos?limit=${limit}`);
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				videos: response.videos,
				loading: false
			}));
		} catch (error) {
			youtubeActions.setError(error instanceof Error ? error.message : 'Failed to load videos');
			youtubeActions.setLoading(false);
		}
	},

	// Get video by ID
	async getVideoById(id: string): Promise<void> {
		youtubeActions.setLoading(true);
		youtubeActions.setError(null);

		try {
			const video = await apiCall<YouTubeVideo>(`/videos/${id}`);
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				currentVideo: video,
				loading: false
			}));
		} catch (error) {
			youtubeActions.setError(error instanceof Error ? error.message : 'Failed to load video');
			youtubeActions.setLoading(false);
		}
	},

	// Search videos
	async searchVideos(query: string, limit: number = 20): Promise<void> {
		youtubeActions.setLoading(true);
		youtubeActions.setError(null);

		try {
			const response = await apiCall<YouTubeVideosResponse>(`/videos/search?q=${encodeURIComponent(query)}&limit=${limit}`);
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				searchQuery: query,
				searchResults: response.videos,
				loading: false
			}));
		} catch (error) {
			youtubeActions.setError(error instanceof Error ? error.message : 'Failed to search videos');
			youtubeActions.setLoading(false);
		}
	},

	// Get videos by category
	async getVideosByCategory(category: string, limit: number = 20): Promise<void> {
		youtubeActions.setLoading(true);
		youtubeActions.setError(null);

		try {
			const response = await apiCall<YouTubeVideosResponse>(`/videos/category/${encodeURIComponent(category)}?limit=${limit}`);
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				videos: response.videos,
				loading: false
			}));
		} catch (error) {
			youtubeActions.setError(error instanceof Error ? error.message : 'Failed to load category videos');
			youtubeActions.setLoading(false);
		}
	},

	// Get channel info
	async getChannelInfo(): Promise<void> {
		youtubeActions.setLoading(true);
		youtubeActions.setError(null);

		try {
			const channelInfo = await apiCall<ChannelInfo>('/channel');
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				channelInfo,
				loading: false
			}));
		} catch (error) {
			youtubeActions.setError(error instanceof Error ? error.message : 'Failed to load channel info');
			youtubeActions.setLoading(false);
		}
	},

	// Get YouTube status
	async getStatus(): Promise<void> {
		try {
			const status = await apiCall<YouTubeStatus>('/status');
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				status
			}));
		} catch (error) {
			console.error('Failed to load YouTube status:', error);
		}
	},

	// Get all categories
	async getCategories(): Promise<void> {
		try {
			const response = await apiCall<{ categories: string[]; count: number }>('/categories');
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				categories: response.categories
			}));
		} catch (error) {
			console.error('Failed to load categories:', error);
		}
	},

	// Get all tags
	async getTags(): Promise<void> {
		try {
			const response = await apiCall<{ tags: string[]; count: number }>('/tags');
			
			youtubeStore.update((state: YouTubeState) => ({
				...state,
				tags: response.tags
			}));
		} catch (error) {
			console.error('Failed to load tags:', error);
		}
	},

	// Clear current video
	clearCurrentVideo(): void {
		youtubeStore.update((state: YouTubeState) => ({
			...state,
			currentVideo: null
		}));
	},

	// Clear search results
	clearSearch(): void {
		youtubeStore.update((state: YouTubeState) => ({
			...state,
			searchQuery: '',
			searchResults: []
		}));
	},

	// Initialize store with basic data
	async initialize(): Promise<void> {
		await Promise.all([
			youtubeActions.getLatestVideos(10),
			youtubeActions.getChannelInfo(),
			youtubeActions.getStatus(),
			youtubeActions.getCategories(),
			youtubeActions.getTags()
		]);
	}
};

// Utility functions for components
export const youtubeUtils = {
	// Format duration from PT15M42S to readable format
	formatDuration(duration: string): string {
		const match = duration.match(/PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?/);
		if (!match) return duration;

		const hours = parseInt(match[1] || '0');
		const minutes = parseInt(match[2] || '0');
		const seconds = parseInt(match[3] || '0');

		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${seconds.toString().padStart(2, '0')}`;
	},

	// Format view count
	formatViewCount(count: number): string {
		if (count >= 1000000) {
			return `${(count / 1000000).toFixed(1)}M views`;
		} else if (count >= 1000) {
			return `${(count / 1000).toFixed(1)}K views`;
		}
		return `${count} views`;
	},

	// Format published date
	formatPublishedDate(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diffTime = Math.abs(now.getTime() - date.getTime());
		const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

		if (diffDays === 1) return '1 day ago';
		if (diffDays < 7) return `${diffDays} days ago`;
		if (diffDays < 30) return `${Math.floor(diffDays / 7)} weeks ago`;
		if (diffDays < 365) return `${Math.floor(diffDays / 30)} months ago`;
		return `${Math.floor(diffDays / 365)} years ago`;
	},

	// Local thumbnail URL cache
	thumbnailCache: new Map<string, string>(),

	// Get video thumbnail with fallback
	getThumbnail(video: YouTubeVideo, quality: 'default' | 'medium' | 'high' | 'maxres' = 'medium'): string {
		const PLACEHOLDER_IMG = '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_WEBP.webp'; // Using WEBP format for better performance

		if (!video?.thumbnail_url) {
			return PLACEHOLDER_IMG;
		}

		try {
			// Extract video ID from the thumbnail URL
			const videoId = video.thumbnail_url.split('/').slice(-2, -1)[0];
			
			// Check if it looks like a valid YouTube ID (11 characters, alphanumeric + _ and -)
			if (!videoId || !/^[a-zA-Z0-9_-]{11}$/.test(videoId)) {
				console.debug('Invalid or mock YouTube ID:', videoId);
				return PLACEHOLDER_IMG;
			}

			// Check if we already have a cached URL for this video
			const cachedUrl = youtubeUtils.thumbnailCache.get(videoId);
			if (cachedUrl) {
				return cachedUrl;
			}

			// YouTube thumbnails are available in these formats:
			// - maxresdefault.jpg (1280x720) - not always available
			// - sddefault.jpg (640x480) - not always available
			// - hqdefault.jpg (480x360) - always available
			// - mqdefault.jpg (320x180) - always available
			// - default.jpg (120x90) - always available
			const qualityMap = {
				maxres: ['maxresdefault', 'sddefault', 'hqdefault', 'mqdefault', 'default'],
				high: ['hqdefault', 'mqdefault', 'default'],
				medium: ['mqdefault', 'default'],
				default: ['default']
			};

			// Get the quality chain for the requested quality
			const qualities = qualityMap[quality];
			let currentQualityIndex = 0;

			const checkImage = (url: string): Promise<string> => {
				return new Promise((resolve) => {
					const img = new Image();
					img.onload = () => resolve(url);
					img.onerror = () => {
						currentQualityIndex++;
						if (currentQualityIndex < qualities.length) {
							const nextUrl = `https://img.youtube.com/vi/${videoId}/${qualities[currentQualityIndex]}.jpg`;
							resolve(checkImage(nextUrl));
						} else {
							resolve(PLACEHOLDER_IMG);
						}
					};
					img.src = url;
				});
			};

			// Start with the first quality
			const initialUrl = `https://img.youtube.com/vi/${videoId}/${qualities[0]}.jpg`;
			
			// Start the check process and update the cache when complete
			checkImage(initialUrl).then(finalUrl => {
				youtubeUtils.thumbnailCache.set(videoId, finalUrl);
			});

			// Return either cached URL or initial URL while checking
			return youtubeUtils.thumbnailCache.get(videoId) || initialUrl;
		} catch (error) {
			console.error('Error getting thumbnail:', error);
			return PLACEHOLDER_IMG;
		}
	}
};

// Export types for use in components
export type { YouTubeVideo, YouTubeVideosResponse, ChannelInfo, YouTubeStatus, YouTubeState }; 