import { writable, derived } from 'svelte/store';
import { browser } from '$app/environment';
import { apiClient } from '$lib/api/client';
import type { YouTubeVideo, YouTubeVideosResponse, YouTubeStatus } from '$lib/types/youtube';

interface YouTubeStoreState {
	videos: YouTubeVideo[];
	status: YouTubeStatus | null;
	loading: boolean;
	error: string | null;
	lastUpdated: Date | null;
}

const initialState: YouTubeStoreState = {
	videos: [],
	status: null,
	loading: false,
	error: null,
	lastUpdated: null
};

// Create the main YouTube store
export const youtubeStore = writable<YouTubeStoreState>(initialState);

// Derived stores for easy access
export const youtubeVideos = derived(youtubeStore, $store => $store.videos);
export const youtubeStatus = derived(youtubeStore, $store => $store.status);
export const youtubeLoading = derived(youtubeStore, $store => $store.loading);
export const youtubeError = derived(youtubeStore, $store => $store.error);

// Helper functions
export const youtubeActions = {
	// Fetch latest YouTube videos
	async fetchLatestVideos() {
		if (!browser) return;

		youtubeStore.update(state => ({ ...state, loading: true, error: null }));

		try {
			const response = await apiClient.getLatestYouTubeVideos();
			
			if (response.data) {
				youtubeStore.update(state => ({
					...state,
					videos: response.data!.videos,
					lastUpdated: new Date(response.data!.last_updated),
					loading: false,
					error: null
				}));
			} else {
				throw new Error(response.error || 'Failed to fetch YouTube videos');
			}
		} catch (error) {
			console.error('Error fetching YouTube videos:', error);
			youtubeStore.update(state => ({
				...state,
				loading: false,
				error: error instanceof Error ? error.message : 'Failed to fetch YouTube videos'
			}));
		}
	},

	// Fetch all YouTube videos with optional limit
	async fetchVideos(limit?: number) {
		if (!browser) return;

		youtubeStore.update(state => ({ ...state, loading: true, error: null }));

		try {
			const response = await apiClient.getYouTubeVideos(limit);
			
			if (response.data) {
				youtubeStore.update(state => ({
					...state,
					videos: response.data!.videos,
					lastUpdated: new Date(response.data!.last_updated),
					loading: false,
					error: null
				}));
			} else {
				throw new Error(response.error || 'Failed to fetch YouTube videos');
			}
		} catch (error) {
			console.error('Error fetching YouTube videos:', error);
			youtubeStore.update(state => ({
				...state,
				loading: false,
				error: error instanceof Error ? error.message : 'Failed to fetch YouTube videos'
			}));
		}
	},

	// Fetch YouTube integration status
	async fetchStatus() {
		if (!browser) return;

		try {
			const response = await apiClient.getYouTubeStatus();
			
			if (response.data) {
				youtubeStore.update(state => ({
					...state,
					status: response.data!
				}));
			} else {
				throw new Error(response.error || 'Failed to fetch YouTube status');
			}
		} catch (error) {
			console.error('Error fetching YouTube status:', error);
			youtubeStore.update(state => ({
				...state,
				error: error instanceof Error ? error.message : 'Failed to fetch YouTube status'
			}));
		}
	},

	// Subscribe to YouTube channel updates
	async subscribe() {
		if (!browser) return;

		try {
			const response = await apiClient.subscribeToYouTube();
			
			if (response.data?.success) {
				// Refresh status after successful subscription
				await this.fetchStatus();
				return { success: true, message: response.data.message };
			} else {
				throw new Error(response.error || 'Failed to subscribe to YouTube channel');
			}
		} catch (error) {
			console.error('Error subscribing to YouTube:', error);
			const errorMessage = error instanceof Error ? error.message : 'Failed to subscribe to YouTube channel';
			youtubeStore.update(state => ({ ...state, error: errorMessage }));
			return { success: false, message: errorMessage };
		}
	},

	// Unsubscribe from YouTube channel updates
	async unsubscribe() {
		if (!browser) return;

		try {
			const response = await apiClient.unsubscribeFromYouTube();
			
			if (response.data?.success) {
				// Refresh status after successful unsubscription
				await this.fetchStatus();
				return { success: true, message: response.data.message };
			} else {
				throw new Error(response.error || 'Failed to unsubscribe from YouTube channel');
			}
		} catch (error) {
			console.error('Error unsubscribing from YouTube:', error);
			const errorMessage = error instanceof Error ? error.message : 'Failed to unsubscribe from YouTube channel';
			youtubeStore.update(state => ({ ...state, error: errorMessage }));
			return { success: false, message: errorMessage };
		}
	},

	// Clear error state
	clearError() {
		youtubeStore.update(state => ({ ...state, error: null }));
	},

	// Reset store to initial state
	reset() {
		youtubeStore.set(initialState);
	},

	// Get video by ID
	getVideoById(id: string): YouTubeVideo | null {
		let video: YouTubeVideo | null = null;
		youtubeStore.subscribe(state => {
			video = state.videos.find(v => v.id === id) || null;
		})();
		return video;
	},

	// Get latest N videos
	getLatestVideos(count: number): YouTubeVideo[] {
		let videos: YouTubeVideo[] = [];
		youtubeStore.subscribe(state => {
			videos = state.videos
				.sort((a, b) => new Date(b.published_at).getTime() - new Date(a.published_at).getTime())
				.slice(0, count);
		})();
		return videos;
	}
};

// Auto-fetch latest videos on store initialization (client-side only)
if (browser) {
	// Temporarily use mock data instead of API calls
	youtubeStore.update(state => ({
		...state,
		videos: [
			{
				id: "test1",
				title: "Test YouTube Video 1",
				description: "This is a test video from the test channel",
				published_at: "2024-01-15T10:00:00Z",
				updated_at: "2024-01-15T10:00:00Z",
				thumbnail_url: "https://img.youtube.com/vi/test1/maxresdefault.jpg",
				video_url: "https://www.youtube.com/watch?v=test1",
				embed_url: "https://www.youtube.com/embed/test1",
				duration: "10:30",
				view_count: 1500,
				created_at: "2024-01-15T10:00:00Z"
			},
			// Add more test videos as needed
		],
		loading: false,
		error: null,
		lastUpdated: new Date()
	}));
} 