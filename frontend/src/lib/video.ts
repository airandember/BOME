import { apiRequest } from './auth';

// Default placeholder image path
const DEFAULT_THUMBNAIL = '/16X10_Placeholder_IMG.png';

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
	status: string;
	createdAt: string;
	updatedAt: string;
	bunnyVideoId?: string; // Bunny.net GUID
	encodeProgress?: number; // Bunny.net encoding progress
	iframeSrc?: string;
	directPlayUrl?: string;
	resolutions?: string[];
	playData?: VideoPlayData;
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

export interface VideosResponse {
	videos: Video[];
	pagination: {
		current_page: number;
		per_page: number;
		total: number;
		has_more: boolean;
	};
}

export interface ApiError {
	message: string;
	error_type?: string;
	details?: string;
	status?: number;
}

// Collection interfaces
export interface BunnyCollection {
	id: string;
	name: string;
	videoCount: number;
	totalSize: number;
	createdAt: string;
	updatedAt: string;
}

export interface CollectionsResponse {
	totalItems: number;
	currentPage: number;
	itemsPerPage: number;
	items: BunnyCollection[];
}

export interface VideoPlayData {
	videoLibraryId: string;
	guid: string;
	title: string;
	status: number;
	framerate: number;
	width: number;
	height: number;
	duration: number;
	thumbnailCount: number;
	resolutions: string[];
	thumbnailFileName: string;
	hasMP4Fallback: boolean;
	playbackUrl: string;
	iframeSrc: string;
	directPlayUrl: string;
	thumbnailUrl: string;
}

// Enhanced error handling with retry logic
async function apiRequestWithRetry(endpoint: string, options: RequestInit = {}, maxRetries = 3): Promise<Response> {
	let lastError: Error | null = null;
	
	for (let attempt = 1; attempt <= maxRetries; attempt++) {
		try {
			const response = await apiRequest(endpoint, options);
			
			// If successful, return immediately
			if (response.ok) {
				return response;
			}
			
			// For certain status codes, don't retry
			if ([400, 401, 403, 404].includes(response.status)) {
				return response;
			}
			
			// For other errors, throw to trigger retry
			throw new Error(`HTTP ${response.status}: ${response.statusText}`);
			
		} catch (error) {
			lastError = error as Error;
			
			// Don't retry on the last attempt
			if (attempt === maxRetries) {
				break;
			}
			
			// Wait before retrying (exponential backoff)
			const delay = Math.min(1000 * Math.pow(2, attempt - 1), 5000);
			await new Promise(resolve => setTimeout(resolve, delay));
		}
	}
	
	// If we get here, all retries failed
	throw lastError || new Error('Request failed after all retries');
}

// Parse API error response
function parseApiError(response: Response, data?: any): ApiError {
	const error: ApiError = {
		message: 'An unexpected error occurred',
		status: response.status,
	};
	
	if (data) {
		if (data.error) {
			error.message = data.error;
		}
		if (data.error_type) {
			error.error_type = data.error_type;
		}
		if (data.details) {
			error.details = data.details;
		}
	}
	
	// Set specific messages based on status codes
	switch (response.status) {
		case 400:
			error.message = error.message || 'Bad request';
			break;
		case 401:
			error.message = 'Authentication required';
			break;
		case 403:
			error.message = 'Access denied';
			break;
		case 404:
			error.message = 'Resource not found';
			break;
		case 429:
			error.message = 'Too many requests. Please try again later.';
			break;
		case 500:
			error.message = 'Server error. Please try again later.';
			break;
		case 503:
			error.message = 'Service temporarily unavailable';
			break;
	}
	
	return error;
}

// Helper function to ensure valid thumbnail URL
function getThumbnailUrl(video: Partial<Video>): string {
	// If we have a valid thumbnailUrl and it's not 'error', use it, otherwise use placeholder
	return (video.thumbnailUrl && video.thumbnailUrl !== 'error') ? video.thumbnailUrl : DEFAULT_THUMBNAIL;
}

// Video service - only uses real backend data, no mock fallbacks
export const videoService = {
	// Get all collections with pagination
	getCollections: async (page = 1, perPage = 20): Promise<CollectionsResponse> => {
		const params = new URLSearchParams({
			page: page.toString(),
			per_page: perPage.toString()
		});

		try {
			const response = await apiRequestWithRetry(`/bunny-collections?${params.toString()}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			return await response.json();
		} catch (error) {
			console.error('Error fetching collections:', error);
			throw error;
		}
	},

	// Get a single collection by ID
	getCollection: async (id: string): Promise<BunnyCollection> => {
		try {
			const response = await apiRequestWithRetry(`/bunny-collections/${id}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			return await response.json();
		} catch (error) {
			console.error('Error fetching collection:', error);
			throw error;
		}
	},

	// Get all videos with pagination and filtering
	getVideos: async (page = 1, limit = 20, category?: string, search?: string): Promise<VideosResponse> => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString(),
		});
		
		if (category) params.append('category', category);
		if (search) params.append('search', search);

		console.log('Fetching videos from:', `/bunny-videos?${params.toString()}`);
		
		try {
			const response = await apiRequestWithRetry(`/bunny-videos?${params.toString()}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			
			// Process videos to ensure proper thumbnail URLs
			if (data.videos && Array.isArray(data.videos)) {
				data.videos = data.videos.map((video: Partial<Video>) => ({
					...video,
					thumbnailUrl: getThumbnailUrl(video)
				}));
			}
			
			console.log('Videos API response:', data);
			return data;
		} catch (error) {
			console.error('Error fetching videos:', error);
			throw error;
		}
	},

	// Get a single video by ID or Bunny GUID
	getVideo: async (id: string): Promise<Video> => {
		try {
			const response = await apiRequestWithRetry(`/videos/${id}`);
			
			if (!response.ok) {
				throw await parseApiError(response);
			}
			
			const data = await response.json();
			
			// Ensure we have proper playback URLs
			const playbackUrl = data.playData?.directPlayUrl || data.directPlayUrl || 
							  `https://iframe.mediadelivery.net/embed/${data.bunnyVideoId}`;
			
			return {
				id: data.ID,
				title: data.Title,
				description: data.Description,
				thumbnailUrl: data.ThumbnailURL,
				videoUrl: playbackUrl,
				duration: data.Duration,
				viewCount: data.ViewCount,
				likeCount: data.LikeCount,
				category: data.Category,
				tags: data.Tags,
				status: data.Status,
				createdAt: data.CreatedAt,
				updatedAt: data.UpdatedAt,
				bunnyVideoId: data.BunnyVideoID,
				playData: data.playData
			};
		} catch (error) {
			console.error('Error fetching video:', error);
			throw error;
		}
	},

	// Get video comments
	getComments: async (videoId: number, page = 1, limit = 20) => {
		const params = new URLSearchParams({
			page: page.toString(),
			limit: limit.toString(),
		});

		try {
			const response = await apiRequestWithRetry(`/videos/${videoId}/comments?${params.toString()}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error fetching comments:', error);
			throw error;
		}
	},

	// Get video categories
	getCategories: async () => {
		try {
			const response = await apiRequestWithRetry('/videos/categories');
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error fetching categories:', error);
			throw error;
		}
	},

	// Search videos
	searchVideos: async (query: string, page = 1, limit = 20): Promise<VideosResponse> => {
		const params = new URLSearchParams({
			q: query,
			page: page.toString(),
			limit: limit.toString(),
		});

		try {
			const response = await apiRequestWithRetry(`/videos/search?${params.toString()}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error searching videos:', error);
			throw error;
		}
	},

	// Get videos by category
	getVideosByCategory: async (category: string, page = 1, limit = 20): Promise<VideosResponse> => {
		const params = new URLSearchParams({
			category,
			page: page.toString(),
			limit: limit.toString(),
		});

		try {
			const response = await apiRequestWithRetry(`/videos?${params.toString()}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error fetching videos by category:', error);
			throw error;
		}
	},

	// Get streaming URL for a video
	getStreamUrl: async (videoId: number) => {
		try {
			const response = await apiRequestWithRetry(`/videos/${videoId}/stream`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error getting stream URL:', error);
			throw error;
		}
	},

	// Sync Bunny.net videos (admin only)
	syncBunnyVideos: async (): Promise<any> => {
		try {
			const response = await apiRequestWithRetry('/sync-bunny-videos', {
				method: 'POST',
			});
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error syncing Bunny videos:', error);
			throw error;
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

// No mock data exports - only real backend data 