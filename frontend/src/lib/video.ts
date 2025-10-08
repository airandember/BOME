import { apiRequest } from './auth';
import { CacheManager } from './performance/optimization';

// Default placeholder image path
const DEFAULT_THUMBNAIL = '/16X10_Placeholder_IMG.png';

// Cache configuration
const CACHE_CONFIG = {
	VIDEO_LIST_TTL: 5 * 60 * 1000, // 5 minutes
	VIDEO_DETAILS_TTL: 10 * 60 * 1000, // 10 minutes
	CATEGORIES_TTL: 30 * 60 * 1000, // 30 minutes
};

// Initialize cache names
const CACHE_NAMES = {
	VIDEO_LIST: 'video_list',
	VIDEO_DETAILS: 'video_details',
	CATEGORIES: 'categories'
};

export interface Video {
	id: number;
	title: string;
	description: string;
	thumbnailUrl: string;
	playbackUrl: string; // Add this
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
	color: string;
	createdAt: string;
	updatedAt: string;
	videoCount: number;
	tagIds: number[];
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
		currentPage: number;
		itemsPerPage: number;
		totalItems: number;
		hasMore: boolean;
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
	guid: string;  // Changed from 'id' to 'guid' to match backend
	name: string;
	videoCount: number;
	totalSize: number;
	dateCreated: string;  // Changed from 'createdAt' to 'dateCreated' to match backend
	lastUpdated: string;  // Changed from 'updatedAt' to 'lastUpdated' to match backend
	previewVideoIds?: string[];  // Array of Bunny video GUIDs for preview
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

// Initialize cache cleanup
setInterval(() => {
	CacheManager.cleanup();
	console.log('Cache cleanup completed');
}, 10 * 60 * 1000); // Run every 10 minutes

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

		// Create cache key
		const cacheKey = `videos_${page}_${limit}_${category || 'all'}_${search || 'none'}`;
		
		// Check cache first
		const cachedData = CacheManager.get(CACHE_NAMES.VIDEO_LIST, cacheKey);
		if (cachedData) {
			console.log('Returning cached video list');
			return cachedData;
		}

		console.log('Fetching videos from:', `/bunny-videos?${params.toString()}`);
		
		try {
			const response = await apiRequestWithRetry(`/bunny-videos?${params.toString()}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			
			// Cache the response
			CacheManager.set(CACHE_NAMES.VIDEO_LIST, cacheKey, data, CACHE_CONFIG.VIDEO_LIST_TTL);
			
			return data;
		} catch (error) {
			console.error('Error fetching videos:', error);
			throw error;
		}
	},

	// Get a single video by ID or Bunny GUID
	getVideo: async (id: string): Promise<Video> => {
		// Check cache first
		const cachedVideo = CacheManager.get(CACHE_NAMES.VIDEO_DETAILS, id);
		if (cachedVideo) {
			console.log('Returning cached video details for:', id);
			return cachedVideo;
		}

		try {
			// If the ID contains hyphens, it's a Bunny GUID
			const endpoint = id.includes('-') ? `/bunny-videos/${id}` : `/videos/${id}`;
			const response = await apiRequestWithRetry(endpoint);
			
			if (!response.ok) {
				throw await parseApiError(response);
			}
			
			const data = await response.json();
			
			// Ensure we have proper playback URLs
			// For video playback, prioritize HLS stream URL over iframe URL
			const hlsStreamUrl = data.playData?.directPlayUrl || data.directPlayUrl;
			const iframeUrl = data.playData?.iframeSrc || data.iframeSrc || 
							  `https://iframe.mediadelivery.net/play/347378/${data.bunnyVideoId}`;
			
			// Use HLS stream URL for playback if available, otherwise fall back to iframe
			const playbackUrl = hlsStreamUrl || iframeUrl;
			
			const video: Video = {
				id: data.id || data.ID,  // Handle both formats
				title: data.title || data.Title,
				description: data.description || data.Description,
				thumbnailUrl: data.thumbnailUrl || data.ThumbnailURL || getThumbnailUrl(data),
				playbackUrl: playbackUrl, // HLS stream URL for video playback
				videoUrl: playbackUrl,
				duration: data.duration || data.Duration,
				viewCount: data.viewCount || data.ViewCount || 0,
				likeCount: data.likeCount || data.LikeCount || 0,
				category: data.category || data.Category || '',
				tags: data.tags || data.Tags || [],
				status: data.status || data.Status,
				createdAt: data.createdAt || data.CreatedAt,
				updatedAt: data.updatedAt || data.UpdatedAt,
				bunnyVideoId: data.bunnyVideoId || data.BunnyVideoID || id,
				iframeSrc: iframeUrl, // Direct play iframe URL
				directPlayUrl: hlsStreamUrl, // HLS stream URL
				playData: data.playData || data.PlayData
			};

			// Cache the video details
			CacheManager.set(CACHE_NAMES.VIDEO_DETAILS, id, video, CACHE_CONFIG.VIDEO_DETAILS_TTL);

			return video;
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
	},

	// Get videos by collection ID
	getVideosByCollection: async (collectionId: string, page = 1, itemsPerPage = 20): Promise<VideosResponse> => {
		try {
			// Use the dedicated collection videos endpoint
			const response = await apiRequestWithRetry(`/bunny-collections/${collectionId}/videos?page=${page}&per_page=${itemsPerPage}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			console.log('Collection videos response:', data);
			
			// The backend returns the response in the correct format already
			return {
				videos: data.videos || [],
				pagination: {
					currentPage: data.pagination?.current_page || page,
					itemsPerPage: data.pagination?.per_page || itemsPerPage,
					totalItems: data.pagination?.total || 0,
					hasMore: data.pagination?.has_more || false
				}
			};
		} catch (error) {
			console.error('Error fetching videos by collection:', error);
			if (error instanceof Error) {
				error.message = `Failed to fetch collection videos: ${error.message}`;
			}
			throw error;
		}
	},

	// Get tag categories from the streaming admin system
	getTagCategories: async (): Promise<{ categories: VideoCategory[] }> => {
		try {
			const response = await apiRequestWithRetry('/tag-categories');
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			
			console.log('🔍 Raw API response for tag categories:', data);
			
			// Transform the tag categories to match VideoCategory interface
			const categories: VideoCategory[] = (data.result || []).map((category: any) => {
				console.log('🔄 Transforming category:', category);
				const transformed = {
					id: category.id,
					name: category.name,
					description: category.description || '',
					color: category.color || '#3B82F6',
					createdAt: category.created_at || '',
					updatedAt: category.updated_at || '',
					videoCount: category.tag_ids?.length || 0, // Approximate count based on tags
					tagIds: category.tag_ids || []
				};
				console.log('✅ Transformed category:', transformed);
				return transformed;
			});
			
			console.log('📦 Final categories array:', categories);
			return { categories };
		} catch (error) {
			console.error('Error fetching tag categories:', error);
			throw error;
		}
	},

	// Get videos by category tags - now uses direct backend endpoint with tag IDs
	getVideosByTagCategory: async (categoryId: number, page = 1, limit = 20): Promise<VideosResponse> => {
		try {
			// Use the new direct endpoint that queries videos by tag IDs
			const response = await apiRequestWithRetry(`/tag-categories/${categoryId}/videos?page=${page}&limit=${limit}`);
			
			if (!response.ok) {
				const data = await response.json().catch(() => ({}));
				throw parseApiError(response, data);
			}
			
			const data = await response.json();
			
			console.log(`🔍 Raw API response for category ${categoryId} videos:`, data);
			
			if (!data.success) {
				console.error(`❌ API error for category ${categoryId}:`, data.error);
				throw new Error(data.error || 'Failed to fetch videos by category');
			}
			
			// Transform the backend pagination to match our frontend interface
			const backendPagination = data.pagination || {};
			
			// Transform backend MasterVideo format to frontend Video format
			const transformedVideos = (data.result || []).map((backendVideo: any) => ({
				id: backendVideo.ID || backendVideo.id,
				title: backendVideo.Title || backendVideo.title,
				description: backendVideo.Description || backendVideo.description,
				thumbnailUrl: backendVideo.ThumbnailURL || backendVideo.thumbnail_url || backendVideo.thumbnailUrl,
				playbackUrl: backendVideo.PlaybackURL || backendVideo.playback_url || backendVideo.playbackUrl,
				videoUrl: backendVideo.VideoURL || backendVideo.video_url || backendVideo.videoUrl,
				duration: backendVideo.Duration || backendVideo.duration || 0,
				viewCount: backendVideo.Views || backendVideo.views || 0,
				likeCount: backendVideo.Likes || backendVideo.likes || 0,
				category: backendVideo.Category || backendVideo.category || '',
				tags: backendVideo.Tags || backendVideo.tags || [],
				status: backendVideo.Status || backendVideo.status || '',
				createdAt: backendVideo.CreatedAt || backendVideo.created_at || '',
				updatedAt: backendVideo.UpdatedAt || backendVideo.updated_at || '',
				bunnyVideoId: backendVideo.bunnyVideoId || backendVideo.BunnyVideoID || backendVideo.bunny_video_id,
				encodeProgress: backendVideo.EncodeProgress || backendVideo.encode_progress,
				iframeSrc: backendVideo.IframeSrc || backendVideo.iframe_src
			}));

			const result = {
				videos: transformedVideos,
				pagination: {
					currentPage: backendPagination.page || page,
					itemsPerPage: backendPagination.limit || limit,
					totalItems: backendPagination.total || 0,
					hasMore: (backendPagination.page || page) < (backendPagination.total_pages || 0)
				}
			};
			
			console.log(`📦 Transformed response for category ${categoryId}:`, {
				videoCount: result.videos.length,
				pagination: result.pagination,
				firstVideo: result.videos[0] ? {
					id: result.videos[0].id,
					title: result.videos[0].title,
					thumbnailUrl: result.videos[0].thumbnailUrl
				} : null
			});
			
			return result;
		} catch (error) {
			console.error('Error fetching videos by tag category:', error);
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
	formatViewCount: (count: number | undefined | null): string => {
		// Handle undefined, null, or NaN values
		if (count == null || isNaN(count)) {
			return '0';
		}
		
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
