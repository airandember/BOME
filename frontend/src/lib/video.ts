import { api } from './auth';

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

// Video service
export const videoService = {
	// Get all videos with pagination and filtering
	getVideos: async (page = 1, limit = 20, category?: string, search?: string) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});
		
		if (category) params.append('category', category);
		if (search) params.append('search', search);
		
		return api.get(`/api/v1/videos?${params.toString()}`);
	},

	// Get a single video by ID
	getVideo: async (id: number) => {
		return api.get(`/api/v1/videos/${id}`);
	},

	// Get video stream URL
	getVideoStream: async (id: number) => {
		return api.get(`/api/v1/videos/${id}/stream`);
	},

	// Get video categories
	getCategories: async () => {
		return api.get('/api/v1/videos/categories');
	},

	// Search videos
	searchVideos: async (query: string, page = 1, limit = 20) => {
		const params = new URLSearchParams({
			q: query,
			page: page.toString(),
			limit: limit.toString()
		});
		
		return api.get(`/api/v1/videos/search?${params.toString()}`);
	},

	// Like a video
	likeVideo: async (id: number) => {
		return api.post(`/api/v1/videos/${id}/like`, {});
	},

	// Unlike a video
	unlikeVideo: async (id: number) => {
		return api.delete(`/api/v1/videos/${id}/like`);
	},

	// Add to favorites
	favoriteVideo: async (id: number) => {
		return api.post(`/api/v1/videos/${id}/favorite`, {});
	},

	// Remove from favorites
	unfavoriteVideo: async (id: number) => {
		return api.delete(`/api/v1/videos/${id}/favorite`);
	},

	// Get video comments
	getComments: async (id: number, page = 1, limit = 20) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});
		
		return api.get(`/api/v1/videos/${id}/comments?${params.toString()}`);
	},

	// Add a comment
	addComment: async (id: number, content: string) => {
		return api.post(`/api/v1/videos/${id}/comment`, { content });
	},

	// Get user favorites
	getFavorites: async (page = 1, limit = 20) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString()
		});
		
		return api.get(`/api/v1/users/favorites?${params.toString()}`);
	},

	// Admin video management functions
	admin: {
		// Get all videos for admin with pagination and filtering
		getVideos: async (page = 1, limit = 20, filters: {
			search?: string;
			category?: string;
			status?: string;
			sort?: string;
			order?: string;
		} = {}) => {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			
			if (filters.search) params.append('search', filters.search);
			if (filters.category) params.append('category', filters.category);
			if (filters.status) params.append('status', filters.status);
			if (filters.sort) params.append('sort', filters.sort);
			if (filters.order) params.append('order', filters.order);
			
			return api.get(`/api/v1/admin/videos?${params.toString()}`);
		},

		// Get video statistics
		getStats: async () => {
			return api.get('/api/v1/admin/videos/stats');
		},

		// Get video categories for admin
		getCategories: async () => {
			return api.get('/api/v1/admin/videos/categories');
		},

		// Get pending videos
		getPendingVideos: async () => {
			return api.get('/api/v1/admin/videos/pending');
		},

		// Upload a new video
		uploadVideo: async (formData: FormData) => {
			const token = localStorage.getItem('token');
			const response = await fetch('/api/v1/admin/videos', {
				method: 'POST',
				headers: {
					'Authorization': token ? `Bearer ${token}` : '',
				},
				body: formData
			});
			return response.json();
		},

		// Update video details
		updateVideo: async (id: number, data: {
			title?: string;
			description?: string;
			category?: string;
			status?: string;
			tags?: string[];
		}) => {
			return api.put(`/api/v1/admin/videos/${id}`, data);
		},

		// Delete a video
		deleteVideo: async (id: number) => {
			return api.delete(`/api/v1/admin/videos/${id}`);
		},

		// Approve a video
		approveVideo: async (id: number) => {
			return api.post(`/api/v1/admin/videos/${id}/approve`, {});
		},

		// Reject a video
		rejectVideo: async (id: number) => {
			return api.post(`/api/v1/admin/videos/${id}/reject`, {});
		},

		// Bulk operations
		bulkOperation: async (operation: 'publish' | 'unpublish' | 'delete', videoIds: number[]) => {
			return api.post('/api/v1/admin/videos/bulk', {
				operation,
				video_ids: videoIds
			});
		},

		// Content scheduling
		scheduleVideo: async (id: number, publishDate: string) => {
			return api.post(`/api/v1/admin/videos/${id}/schedule`, {
				publish_date: publishDate
			});
		},

		// Unschedule video
		unscheduleVideo: async (id: number) => {
			return api.delete(`/api/v1/admin/videos/${id}/schedule`);
		},

		// Get scheduled videos
		getScheduledVideos: async () => {
			return api.get('/api/v1/admin/videos/scheduled');
		}
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