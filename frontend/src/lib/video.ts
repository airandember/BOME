import { apiRequest } from './auth';

export interface Video {
	id: number;
	title: string;
	description: string;
	thumbnailUrl: string;
	videoUrl: string;
	duration: number;
	viewCount: number;
	likeCount: number;
	category: string;
	tags: string[];
	createdAt: string;
	updatedAt: string;
}

export interface VideoCategory {
	id: number;
	name: string;
	description: string;
	videoCount: number;
}

export interface VideoComment {
	id: number;
	videoId: number;
	userId: number;
	userName: string;
	content: string;
	createdAt: string;
}

// Video service - only uses real backend data, no mock fallbacks
export const videoService = {
	// Get all videos with pagination and filtering
	getVideos: async (page = 1, limit = 20, category?: string, search?: string) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString(),
		});
		
		if (category) params.append('category', category);
		if (search) params.append('search', search);

		const response = await apiRequest(`/videos?${params.toString()}`);
		return response;
	},

	// Get a single video by ID
	getVideo: async (id: number) => {
		const response = await apiRequest(`/videos/${id}`);
		return response;
	},

	// Get video comments
	getComments: async (videoId: number, page = 1, limit = 20) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString(),
		});

		const response = await apiRequest(`/videos/${videoId}/comments?${params.toString()}`);
		return response;
	},

	// Get video categories
	getCategories: async () => {
		const response = await apiRequest('/videos/categories');
		return response;
	},

	// Search videos
	searchVideos: async (query: string, page = 1, limit = 20) => {
		const params = new URLSearchParams({
			q: query,
			page: page.toString(),
			limit: limit.toString(),
		});

		const response = await apiRequest(`/videos/search?${params.toString()}`);
		return response;
	},

	// Get videos by category
	getVideosByCategory: async (category: string, page = 1, limit = 20) => {
		const params = new URLSearchParams({
			category,
			page: page.toString(),
			limit: limit.toString(),
		});

		const response = await apiRequest(`/videos?${params.toString()}`);
		return response;
	},

	// Get streaming URL for a video
	getStreamUrl: async (videoId: number) => {
		const response = await apiRequest(`/videos/${videoId}/stream`);
		return response;
	}
};

// Video player utilities
export const videoUtils = {
	// Format duration from seconds to MM:SS or HH:MM:SS
	formatDuration: (seconds: number): string => {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
	},

	// Format view count
	formatViewCount: (count: number): string => {
		if (count >= 1000000) {
			return `${(count / 1000000).toFixed(1)}M`;
		} else if (count >= 1000) {
			return `${(count / 1000).toFixed(1)}K`;
		}
		return count.toString();
	},

	// Get video quality options (placeholder for now)
	getQualityOptions: (videoUrl: string) => {
		// In a real implementation, this would return different quality URLs
		return [
			{ label: 'Auto', value: 'auto' },
			{ label: '1080p', value: '1080p' },
			{ label: '720p', value: '720p' },
			{ label: '480p', value: '480p' },
			{ label: '360p', value: '360p' }
		];
	}
};

// No mock data exports - only real backend data 