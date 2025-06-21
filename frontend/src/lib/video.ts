import { api } from './auth';
import { 
	MOCK_VIDEOS, 
	MOCK_CATEGORIES, 
	MOCK_COMMENTS,
	createMockResponse,
	getMockVideos,
	getMockVideo,
	getMockComments,
	getMockAdminVideos,
	getMockVideoStats
} from './mockData';

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

// Video service with comprehensive mock data fallbacks
export const videoService = {
	// Get all videos with pagination and filtering
	getVideos: async (page = 1, limit = 20, category?: string, search?: string) => {
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			
			if (category) params.append('category', category);
			if (search) params.append('search', search);
			
			const response = await api.get(`/api/v1/videos?${params.toString()}`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			return createMockResponse(getMockVideos(page, limit, category, search));
		}
	},

	// Get a single video by ID
	getVideo: async (id: number) => {
		try {
			const response = await api.get(`/api/v1/videos/${id}`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			return createMockResponse(getMockVideo(id));
		}
	},

	// Get video stream URL
	getVideoStream: async (id: number) => {
		try {
			const response = await api.get(`/api/v1/videos/${id}/stream`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			const video = MOCK_VIDEOS.find(v => v.id === id);
			return createMockResponse({ 
				streamUrl: video?.videoUrl || '',
				// Bunny.net typically provides multiple quality streams
				qualityOptions: [
					{
						label: '1080p',
						url: video?.videoUrl?.replace('playlist.m3u8', '1080p/playlist.m3u8') || ''
					},
					{
						label: '720p', 
						url: video?.videoUrl?.replace('playlist.m3u8', '720p/playlist.m3u8') || ''
					},
					{
						label: '480p',
						url: video?.videoUrl?.replace('playlist.m3u8', '480p/playlist.m3u8') || ''
					},
					{
						label: '360p',
						url: video?.videoUrl?.replace('playlist.m3u8', '360p/playlist.m3u8') || ''
					}
				],
				// Bunny.net streaming metadata
				metadata: {
					duration: video?.duration || 0,
					width: 1920,
					height: 1080,
					framerate: 30,
					bitrate: '5000kbps',
					codec: 'H.264'
				}
			});
		}
	},

	// Get video categories
	getCategories: async () => {
		try {
			const response = await api.get('/api/v1/videos/categories');
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			return createMockResponse({ categories: MOCK_CATEGORIES });
		}
	},

	// Search videos
	searchVideos: async (query: string, page = 1, limit = 20) => {
		try {
			const params = new URLSearchParams({
				q: query,
				page: page.toString(),
				limit: limit.toString()
			});
			
			const response = await api.get(`/api/v1/videos/search?${params.toString()}`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			return createMockResponse(getMockVideos(page, limit, undefined, query));
		}
	},

	// Like a video
	likeVideo: async (id: number) => {
		try {
			const response = await api.post(`/api/v1/videos/${id}/like`, {});
			return response;
		} catch (error) {
			console.warn('API call failed, using mock response:', error);
			return createMockResponse({ success: true, message: 'Video liked' }, 200);
		}
	},

	// Unlike a video
	unlikeVideo: async (id: number) => {
		try {
			const response = await api.delete(`/api/v1/videos/${id}/like`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock response:', error);
			return createMockResponse({ success: true, message: 'Video unliked' }, 200);
		}
	},

	// Add to favorites
	favoriteVideo: async (id: number) => {
		try {
			const response = await api.post(`/api/v1/videos/${id}/favorite`, {});
			return response;
		} catch (error) {
			console.warn('API call failed, using mock response:', error);
			return createMockResponse({ success: true, message: 'Video added to favorites' }, 200);
		}
	},

	// Remove from favorites
	unfavoriteVideo: async (id: number) => {
		try {
			const response = await api.delete(`/api/v1/videos/${id}/favorite`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock response:', error);
			return createMockResponse({ success: true, message: 'Video removed from favorites' }, 200);
		}
	},

	// Get video comments
	getComments: async (id: number, page = 1, limit = 20) => {
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			
			const response = await api.get(`/api/v1/videos/${id}/comments?${params.toString()}`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			return createMockResponse(getMockComments(id, page, limit));
		}
	},

	// Add a comment
	addComment: async (id: number, content: string) => {
		try {
			const response = await api.post(`/api/v1/videos/${id}/comment`, { content });
			return response;
		} catch (error) {
			console.warn('API call failed, using mock response:', error);
			const newComment: VideoComment = {
				id: Date.now(),
				videoId: id,
				userId: 1,
				userName: 'Current User',
				content,
				createdAt: new Date().toISOString()
			};
			MOCK_COMMENTS.unshift(newComment);
			return createMockResponse({ success: true, comment: newComment }, 300);
		}
	},

	// Get user favorites
	getFavorites: async (page = 1, limit = 20) => {
		try {
			const params = new URLSearchParams({
				page: page.toString(),
				limit: limit.toString()
			});
			
			const response = await api.get(`/api/v1/users/favorites?${params.toString()}`);
			return response;
		} catch (error) {
			console.warn('API call failed, using mock data:', error);
			// Return first 5 videos as favorites for demo
			const favoriteVideos = MOCK_VIDEOS.slice(0, 5);
			return createMockResponse({
				videos: favoriteVideos,
				pagination: {
					page,
					limit,
					total: favoriteVideos.length,
					totalPages: 1
				}
			});
		}
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
			try {
				const params = new URLSearchParams({
					page: page.toString(),
					limit: limit.toString()
				});
				
				if (filters.search) params.append('search', filters.search);
				if (filters.category) params.append('category', filters.category);
				if (filters.status) params.append('status', filters.status);
				if (filters.sort) params.append('sort', filters.sort);
				if (filters.order) params.append('order', filters.order);
				
				const response = await api.get(`/api/v1/admin/videos?${params.toString()}`);
				return response;
			} catch (error) {
				console.warn('API call failed, using mock data:', error);
				return createMockResponse(getMockAdminVideos(page, limit, filters));
			}
		},

		// Get video statistics
		getStats: async () => {
			try {
				const response = await api.get('/api/v1/admin/videos/stats');
				return response;
			} catch (error) {
				console.warn('API call failed, using mock data:', error);
				return createMockResponse(getMockVideoStats());
			}
		},

		// Get video categories for admin
		getCategories: async () => {
			try {
				const response = await api.get('/api/v1/admin/videos/categories');
				return response;
			} catch (error) {
				console.warn('API call failed, using mock data:', error);
				return createMockResponse({ categories: MOCK_CATEGORIES });
			}
		},

		// Get pending videos
		getPendingVideos: async () => {
			try {
				const response = await api.get('/api/v1/admin/videos/pending');
				return response;
			} catch (error) {
				console.warn('API call failed, using mock data:', error);
				const pendingVideos = MOCK_VIDEOS.slice(0, 3).map(video => ({
					...video,
					status: 'pending',
					uploaded_by: {
						id: Math.floor(Math.random() * 100) + 1,
						name: `Dr. ${['John Smith', 'Sarah Johnson', 'Michael Brown'][Math.floor(Math.random() * 3)]}`,
						email: 'user@byu.edu'
					}
				}));
				return createMockResponse({ videos: pendingVideos });
			}
		},

		// Upload a new video
		uploadVideo: async (formData: FormData) => {
			try {
				const token = localStorage.getItem('token');
				const response = await fetch('/api/v1/admin/videos', {
					method: 'POST',
					headers: {
						'Authorization': token ? `Bearer ${token}` : '',
					},
					body: formData
				});
				return response.json();
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({
					success: true,
					message: 'Video uploaded successfully',
					video: {
						id: Date.now(),
						title: formData.get('title') || 'New Video',
						status: 'pending'
					}
				}, 2000);
			}
		},

		// Update video details
		updateVideo: async (id: number, data: {
			title?: string;
			description?: string;
			category?: string;
			status?: string;
			tags?: string[];
		}) => {
			try {
				const response = await api.put(`/api/v1/admin/videos/${id}`, data);
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ success: true, message: 'Video updated successfully' }, 500);
			}
		},

		// Delete a video
		deleteVideo: async (id: number) => {
			try {
				const response = await api.delete(`/api/v1/admin/videos/${id}`);
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ success: true, message: 'Video deleted successfully' }, 500);
			}
		},

		// Approve a video
		approveVideo: async (id: number) => {
			try {
				const response = await api.post(`/api/v1/admin/videos/${id}/approve`, {});
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ success: true, message: 'Video approved successfully' }, 500);
			}
		},

		// Reject a video
		rejectVideo: async (id: number) => {
			try {
				const response = await api.post(`/api/v1/admin/videos/${id}/reject`, {});
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ success: true, message: 'Video rejected successfully' }, 500);
			}
		},

		// Bulk operations
		bulkOperation: async (operation: 'publish' | 'unpublish' | 'delete', videoIds: number[]) => {
			try {
				const response = await api.post('/api/v1/admin/videos/bulk', {
					operation,
					video_ids: videoIds
				});
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ 
					success: true, 
					message: `Bulk ${operation} completed successfully` 
				}, 1000);
			}
		},

		// Content scheduling
		scheduleVideo: async (id: number, publishDate: string) => {
			try {
				const response = await api.post(`/api/v1/admin/videos/${id}/schedule`, {
					publish_date: publishDate
				});
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ success: true, message: 'Video scheduled successfully' }, 500);
			}
		},

		// Unschedule video
		unscheduleVideo: async (id: number) => {
			try {
				const response = await api.delete(`/api/v1/admin/videos/${id}/schedule`);
				return response;
			} catch (error) {
				console.warn('API call failed, using mock response:', error);
				return createMockResponse({ success: true, message: 'Video unscheduled successfully' }, 500);
			}
		},

		// Get scheduled videos
		getScheduledVideos: async () => {
			try {
				const response = await api.get('/api/v1/admin/videos/scheduled');
				return response;
			} catch (error) {
				console.warn('API call failed, using mock data:', error);
				const scheduledVideos = MOCK_VIDEOS.slice(0, 2).map(video => ({
					...video,
					status: 'scheduled',
					publish_date: new Date(Date.now() + Math.random() * 7 * 24 * 60 * 60 * 1000).toISOString()
				}));
				return createMockResponse({ videos: scheduledVideos });
			}
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

// Export mock data for testing and development
export { MOCK_VIDEOS, MOCK_CATEGORIES, MOCK_COMMENTS }; 