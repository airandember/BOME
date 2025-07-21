import { apiRequest } from '$lib/auth';

// Types for admin streaming data
export interface AdminVideoResponse {
	id: number;
	title: string;
	description: string;
	bunny_video_id: string;
	thumbnail_url: string;
	duration: number;
	file_size: number;
	status: string;
	category: string;
	tags: string[];
	view_count: number;
	like_count: number;
	created_by: number;
	created_at: string;
	updated_at: string;
	scheduled_publish_date?: string;
	
	// Admin-specific fields
	processing_progress: number;
	processing_errors: string[];
	upload_status: string;
	upload_progress: number;
	file_format: string;
	resolution: string;
	bitrate: number;
	framerate: number;
	encoding_profile: string;
	storage_location: string;
	cdn_status: string;
	access_control: string;
	monetization: string;
	analytics: Record<string, any>;
	
	// Bunny.net specific data
	bunny_data?: any;
	play_data?: any;
}

export interface AdminStreamingStats {
	total_videos: number;
	ready_videos: number;
	processing_videos: number;
	error_videos: number;
	draft_videos: number;
	scheduled_videos: number;
	total_storage_bytes: number;
	total_duration_seconds: number;
	total_views: number;
	average_file_size_mb: number;
	processing_errors: number;
	upload_queue_size: number;
	cdn_usage_gb: number;
	bandwidth_usage_gb: number;
	last_sync_time: string;
	sync_status: string;
}

export interface AdminVideoFilters {
	page?: number;
	limit?: number;
	status?: string;
	category?: string;
	search?: string;
	sort?: string;
	order?: 'asc' | 'desc';
	include_processing?: boolean;
}

export interface AdminVideoListResponse {
	success: boolean;
	videos: AdminVideoResponse[];
	pagination: {
		current_page: number;
		per_page: number;
		total: number;
		total_pages: number;
		has_more: boolean;
	};
	filters: AdminVideoFilters;
}

export interface AdminVideoResponse {
	success: boolean;
	video: AdminVideoResponse;
}

export interface AdminStatsResponse {
	success: boolean;
	stats: AdminStreamingStats;
}

export interface AdminCacheMetrics {
	hits: number;
	misses: number;
	evictions: number;
	total_size: number;
	max_size: number;
	average_load_time: number;
}

// Admin Streaming Service
export class AdminStreamingService {
	private baseUrl = '/api/v1/admin/streaming';

	// Get admin video list with enhanced filtering
	async getVideos(filters: AdminVideoFilters = {}): Promise<AdminVideoListResponse> {
		const params = new URLSearchParams();
		
		if (filters.page) params.append('page', filters.page.toString());
		if (filters.limit) params.append('limit', filters.limit.toString());
		if (filters.status) params.append('status', filters.status);
		if (filters.category) params.append('category', filters.category);
		if (filters.search) params.append('search', filters.search);
		if (filters.sort) params.append('sort', filters.sort);
		if (filters.order) params.append('order', filters.order);
		if (filters.include_processing !== undefined) {
			params.append('include_processing', filters.include_processing.toString());
		}

		const url = `${this.baseUrl}/videos?${params.toString()}`;
		const response = await apiRequest(url);
		return await response.json() as AdminVideoListResponse;
	}

	// Get single admin video with enhanced data
	async getVideo(id: string | number): Promise<AdminVideoResponse> {
		const url = `${this.baseUrl}/videos/${id}`;
		const response = await apiRequest(url);
		return await response.json() as AdminVideoResponse;
	}

	// Update admin video
	async updateVideo(id: string | number, updateData: Record<string, any>): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/videos/${id}`;
		const response = await apiRequest(url, {
			method: 'PUT',
			body: JSON.stringify(updateData)
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Delete admin video
	async deleteVideo(id: string | number): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/videos/${id}`;
		const response = await apiRequest(url, {
			method: 'DELETE'
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Get admin streaming stats
	async getStats(): Promise<AdminStatsResponse> {
		const url = `${this.baseUrl}/stats`;
		const response = await apiRequest(url);
		return await response.json() as AdminStatsResponse;
	}

	// Get admin video analytics
	async getVideoAnalytics(id: string | number): Promise<any> {
		const url = `${this.baseUrl}/videos/${id}/analytics`;
		const response = await apiRequest(url);
		return await response.json();
	}

	// Sync with Bunny.net
	async syncWithBunny(): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/sync`;
		const response = await apiRequest(url, {
			method: 'POST'
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Get sync status
	async getSyncStatus(): Promise<any> {
		const url = `${this.baseUrl}/sync/status`;
		const response = await apiRequest(url);
		return await response.json();
	}

	// Retry video processing
	async retryVideoProcessing(id: string | number): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/videos/${id}/retry-processing`;
		const response = await apiRequest(url, {
			method: 'POST'
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Get upload queue
	async getUploadQueue(): Promise<any> {
		const url = `${this.baseUrl}/uploads`;
		const response = await apiRequest(url);
		return await response.json();
	}

	// Cancel upload
	async cancelUpload(id: string | number): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/uploads/${id}/cancel`;
		const response = await apiRequest(url, {
			method: 'POST'
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Get CDN usage
	async getCDNUsage(): Promise<any> {
		const url = `${this.baseUrl}/cdn/usage`;
		const response = await apiRequest(url);
		return await response.json();
	}

	// Get storage usage
	async getStorageUsage(): Promise<any> {
		const url = `${this.baseUrl}/storage/usage`;
		const response = await apiRequest(url);
		return await response.json();
	}

	// Get encoding profiles
	async getEncodingProfiles(): Promise<any> {
		const url = `${this.baseUrl}/encoding/profiles`;
		const response = await apiRequest(url);
		return await response.json();
	}

	// Re-encode video
	async reEncodeVideo(id: string | number, profile?: string): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/videos/${id}/re-encode`;
		const body = profile ? { profile } : {};
		const response = await apiRequest(url, {
			method: 'POST',
			body: JSON.stringify(body)
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Get cache metrics
	async getCacheMetrics(): Promise<{ success: boolean; metrics: AdminCacheMetrics }> {
		const url = `${this.baseUrl}/cache/metrics`;
		const response = await apiRequest(url);
		return await response.json() as { success: boolean; metrics: AdminCacheMetrics };
	}

	// Clear admin cache
	async clearCache(): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/cache/clear`;
		const response = await apiRequest(url, {
			method: 'POST'
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Bulk operations
	async bulkOperation(operation: string, videoIds: (string | number)[]): Promise<{ success: boolean; message: string }> {
		const url = `${this.baseUrl}/videos/bulk`;
		const response = await apiRequest(url, {
			method: 'POST',
			body: JSON.stringify({
				operation,
				video_ids: videoIds
			})
		});
		return await response.json() as { success: boolean; message: string };
	}

	// Utility methods for admin dashboard

	// Get video status counts
	async getVideoStatusCounts(): Promise<Record<string, number>> {
		const stats = await this.getStats();
		return {
			total: stats.stats.total_videos,
			ready: stats.stats.ready_videos,
			processing: stats.stats.processing_videos,
			error: stats.stats.error_videos,
			draft: stats.stats.draft_videos,
			scheduled: stats.stats.scheduled_videos
		};
	}

	// Get storage summary
	async getStorageSummary(): Promise<{
		totalStorage: number;
		averageFileSize: number;
		cdnUsage: number;
		bandwidthUsage: number;
	}> {
		const stats = await this.getStats();
		return {
			totalStorage: stats.stats.total_storage_bytes,
			averageFileSize: stats.stats.average_file_size_mb,
			cdnUsage: stats.stats.cdn_usage_gb,
			bandwidthUsage: stats.stats.bandwidth_usage_gb
		};
	}

	// Format file size for display
	formatFileSize(bytes: number): string {
		const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
		if (bytes === 0) return '0 B';
		const i = Math.floor(Math.log(bytes) / Math.log(1024));
		return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${sizes[i]}`;
	}

	// Format duration for display
	formatDuration(seconds: number): string {
		const hours = Math.floor(seconds / 3600);
		const minutes = Math.floor((seconds % 3600) / 60);
		const secs = seconds % 60;
		
		if (hours > 0) {
			return `${hours}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
		}
		return `${minutes}:${secs.toString().padStart(2, '0')}`;
	}

	// Get video status color for UI
	getStatusColor(status: string): string {
		switch (status.toLowerCase()) {
			case 'ready':
				return 'green';
			case 'processing':
			case 'transcoding':
				return 'yellow';
			case 'error':
			case 'upload_failed':
				return 'red';
			case 'draft':
				return 'gray';
			case 'scheduled':
				return 'blue';
			default:
				return 'gray';
		}
	}

	// Get upload status color for UI
	getUploadStatusColor(status: string): string {
		switch (status.toLowerCase()) {
			case 'ready':
				return 'green';
			case 'uploaded':
			case 'processing':
			case 'transcoding':
				return 'yellow';
			case 'error':
			case 'upload_failed':
				return 'red';
			case 'created':
				return 'blue';
			default:
				return 'gray';
		}
	}
}

// Export singleton instance
export const adminStreamingService = new AdminStreamingService(); 