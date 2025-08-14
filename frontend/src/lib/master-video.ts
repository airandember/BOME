import { apiRequest } from '$lib/auth';

// Types for master video list
export interface MasterVideo {
	ID: number;
	BunnyVideoID: string;
	Title: string;
	Description: string;
	Category: string;
	Tags: string[];
	Duration: number;
	FileSize: number;
	Resolution: string;
	Framerate: number;
	ThumbnailURL: string;
	VideoURL: string;
	IframeSrc: string;
	PlaybackURL: string;
	Status: string;
	Views: number;
	Likes: number;
	IsPublic: boolean;
	EncodeProgress: number;
	AvailableResolutions: string[];
	CollectionID: string;
	AverageWatchTime: number;
	TotalWatchTime: number;
	LastBunnySync: string;
	LastMasterUpdate: string;
	SyncStatus: 'synced' | 'needs_attention' | 'conflict';
	SyncNotes: string;
	MetadataVersion: number;
	CreatedBy: number;
	CreatedAt: string;
	UpdatedAt: string;
	Vid_Status: boolean;
}

export interface SyncConflict {
	id: number;
	master_video_id: number;
	bunny_video_id: string;
	conflict_type: string;
	field_name: string;
	master_value: string;
	bunny_value: string;
	proposed_action: string;
	admin_notes: string;
	resolved: boolean;
	resolved_by?: number;
	resolved_at?: string;
	created_at: string;
}

export interface SyncResult {
	started_at: string;
	completed_at: string;
	duration: string;
	total_videos: number;
	synced: number;
	updated: number;
	conflicts: number;
	errors: number;
	error_details: string[];
}

export interface ConflictCheckResult {
	started_at: string;
	completed_at: string;
	duration: string;
	total_videos: number;
	conflict_count: number;
	errors: number;
	error_details: string[];
	conflicts: VideoConflict[];
}

export interface VideoConflict {
	video: MasterVideo;
	conflicts: FieldConflict[];
}

export interface FieldConflict {
	field: string;
	master_value: string;
	bunny_value: string;
	conflict_type: string;
}

export interface ConflictResolution {
	action: 'update_master' | 'update_bunny' | 'update_both';
	field: string;
	old_value: string;
	new_value: string;
	notes: string;
}

export interface MasterVideoStats {
	total_videos: number;
	videos_by_status: Record<string, number>;
	videos_by_sync_status: Record<string, number>;
	pending_conflicts: number;
	total_views: number;
	total_duration: number;
	total_file_size: number;
}

// Master Video List Service
class MasterVideoService {
	// Get all master videos with filtering and pagination
	async getMasterVideos(params: {
		page?: number;
		limit?: number;
		category?: string;
		status?: string;
		sync_status?: string;
		vid_status?: string;
		search?: string;
		sort_field?: string;
		sort_direction?: 'asc' | 'desc';
	}): Promise<{
		success: boolean;
		videos: MasterVideo[];
		pagination: {
			current_page: number;
			per_page: number;
			total: number;
			total_pages: number;
			has_more: boolean;
		};
	}> {
		const searchParams = new URLSearchParams();
		if (params.page) searchParams.append('page', params.page.toString());
		if (params.limit) searchParams.append('limit', params.limit.toString());
		if (params.category) searchParams.append('category', params.category);
		if (params.status) searchParams.append('status', params.status);
		if (params.sync_status) searchParams.append('sync_status', params.sync_status);
		if (params.vid_status) searchParams.append('vid_status', params.vid_status);
		if (params.search) searchParams.append('search', params.search);
		if (params.sort_field) searchParams.append('sort_field', params.sort_field);
		if (params.sort_direction) searchParams.append('sort_direction', params.sort_direction);

		const response = await apiRequest(`/admin/master-videos?${searchParams.toString()}`);
		return response.json();
	}

	// Get master video by ID
	async getMasterVideo(id: number): Promise<{
		success: boolean;
		video: MasterVideo;
	}> {
		const response = await apiRequest(`/admin/master-videos/${id}`);
		return response.json();
	}

	// Update master video
	async updateMasterVideo(id: number, data: Partial<MasterVideo>): Promise<{
		success: boolean;
		message: string;
		video: MasterVideo;
	}> {
		const response = await apiRequest(`/admin/master-videos/${id}`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		});
		return response.json();
	}

	// Toggle video status (vid_status only)
	async toggleVideoStatus(id: number, vidStatus: boolean): Promise<{
		success: boolean;
		message: string;
		video: MasterVideo;
	}> {
		const response = await apiRequest(`/admin/master-videos/${id}/toggle-status`, {
			method: 'PUT',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify({ vid_status: vidStatus })
		});
		return response.json();
	}

	// Delete master video
	// async deleteMasterVideo(id: number): Promise<{
	// 	success: boolean;
	// 	message: string;
	// }> {
	// 	const response = await apiRequest(`/admin/master-videos/${id}`, {
	// 		method: 'DELETE'
	// 	});
	// 	return response.json();
	// }

	// Sync from Bunny.net to master list
	async syncFromBunny(): Promise<{
		success: boolean;
		message: string;
		result: SyncResult;
	}> {
		const response = await apiRequest('/admin/sync/from-bunny', {
			method: 'POST'
		});
		return response.json();
	}

	// Sync from master list to Bunny.net
	async syncToBunny(): Promise<{
		success: boolean;
		message: string;
		result: SyncResult;
	}> {
		const response = await apiRequest('/admin/sync/to-bunny', {
			method: 'POST'
		});
		return response.json();
	}

	// Check for conflicts
	async checkConflicts(): Promise<{
		success: boolean;
		result: ConflictCheckResult;
	}> {
		const response = await apiRequest('/admin/sync/conflicts');
		return response.json();
	}

	// Get all conflicts
	async getConflicts(): Promise<{
		success: boolean;
		conflicts: SyncConflict[];
	}> {
		const response = await apiRequest('/admin/conflicts');
		return response.json();
	}

	// Resolve a specific conflict
	async resolveConflict(conflictId: number, resolution: ConflictResolution): Promise<{
		success: boolean;
		message: string;
	}> {
		const response = await apiRequest(`/admin/conflicts/${conflictId}/resolve`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(resolution)
		});
		return response.json();
	}

	// Get master video statistics
	async getStats(): Promise<{
		success: boolean;
		stats: MasterVideoStats;
	}> {
		const response = await apiRequest('/admin/stats/master-videos');
		return response.json();
	}

	// Upload video
	async uploadVideo(formData: FormData): Promise<{
		success: boolean;
		message: string;
		video?: MasterVideo;
		error?: string;
	}> {
		try {
			const response = await apiRequest('/videos/upload', {
				method: 'POST',
				body: formData
			});
			
			const data = await response.json();
			
			return {
				success: data.success || false,
				message: data.message || 'Upload completed',
				video: data.video,
				error: data.error
			};
		} catch (error) {
			return {
				success: false,
				message: 'Upload failed',
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	// Get comprehensive dashboard analytics (extends the working stats method)
	async getDashboardAnalytics(): Promise<{
		success: boolean;
		data: {
			total_users: number;
			total_videos: number;
			total_views: number;
			total_revenue: number;
			active_subscriptions: number;
			video_stats: {
				total_videos: number;
				synced_videos: number;
				needs_attention: number;
				total_views: number;
				videos_by_status: Record<string, number>;
				videos_by_sync_status: Record<string, number>;
				pending_conflicts: number;
			};
			view_analytics: {
				total_views: number;
				views_today: number;
				views_week: number;
				growth_rate: number;
			};
			subscriber_metrics: {
				total_subscribers: number;
				active_subscriptions: number;
				monthly_revenue: number;
				churn_rate: number;
			};
			real_time: {
				active_users: number;
				recent_activity: Array<{
					type: string;
					message: string;
					time: string;
					user?: string;
				}>;
			};
		};
	}> {
		try {
			// Get the working video stats
			const videoStatsResponse = await this.getStats();
			
			if (!videoStatsResponse.success) {
				throw new Error('Failed to get video stats');
			}

			// Get additional analytics from existing endpoints
			const [viewAnalyticsResponse, subscriberMetricsResponse] = await Promise.allSettled([
				apiRequest('/admin/streaming/dashboard').then(r => r.json()),
				apiRequest('/admin/analytics').then(r => r.json())
			]);

			// Debug logging (can be removed once stable)
			console.log('🔍 Dashboard API responses:');
			console.log('View analytics response:', viewAnalyticsResponse);
			console.log('Subscriber metrics response:', subscriberMetricsResponse);

			// Build comprehensive analytics from working data
			const videoStats = videoStatsResponse.stats;
			
			// Safely extract view analytics with fallbacks
			const viewAnalytics = viewAnalyticsResponse.status === 'fulfilled' 
				? viewAnalyticsResponse.value?.data?.view_analytics || {}
				: {};
			
			// Safely extract subscriber metrics with fallbacks
			const subscriberMetrics = subscriberMetricsResponse.status === 'fulfilled'
				? subscriberMetricsResponse.value?.data?.subscriptions || {}
				: {};

			return {
				success: true,
				data: {
					total_users: subscriberMetricsResponse.status === 'fulfilled' 
						? subscriberMetricsResponse.value?.data?.users?.total || 0 
						: 0,
					total_videos: videoStats.total_videos,
					total_views: videoStats.total_views,
					total_revenue: subscriberMetrics.monthly_revenue || 0,
					active_subscriptions: subscriberMetrics.active_subscriptions || 0,
					video_stats: {
						total_videos: videoStats.total_videos,
						synced_videos: videoStats.videos_by_sync_status?.synced || 0,
						needs_attention: videoStats.videos_by_sync_status?.needs_attention || 0,
						total_views: videoStats.total_views,
						videos_by_status: videoStats.videos_by_status,
						videos_by_sync_status: videoStats.videos_by_sync_status,
						pending_conflicts: videoStats.pending_conflicts
					},
					view_analytics: {
						total_views: viewAnalytics.total_views || videoStats.total_views,
						views_today: viewAnalytics.views_today || 0,
						views_week: viewAnalytics.views_week || 0,
						growth_rate: viewAnalytics.growth_rate || 0
					},
					subscriber_metrics: {
						total_subscribers: subscriberMetrics.total_subscribers || 0,
						active_subscriptions: subscriberMetrics.active_subscriptions || 0,
						monthly_revenue: subscriberMetrics.monthly_revenue || 0,
						churn_rate: subscriberMetrics.churn_rate || 0
					},
					real_time: {
						active_users: viewAnalyticsResponse.status === 'fulfilled' 
							? viewAnalyticsResponse.value?.data?.active_users || 0 
							: 0,
						recent_activity: viewAnalyticsResponse.status === 'fulfilled'
							? viewAnalyticsResponse.value?.data?.recent_activity || []
							: []
					}
				}
			};
		} catch (error) {
			console.error('Failed to get dashboard analytics:', error);
			return {
				success: false,
				data: {
					total_users: 0,
					total_videos: 0,
					total_views: 0,
					total_revenue: 0,
					active_subscriptions: 0,
					video_stats: {
						total_videos: 0,
						synced_videos: 0,
						needs_attention: 0,
						total_views: 0,
						videos_by_status: {},
						videos_by_sync_status: {},
						pending_conflicts: 0
					},
					view_analytics: {
						total_views: 0,
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
				}
			};
		}
	}

	// Get basic stats for simple dashboards
	async getBasicStats(): Promise<{
		success: boolean;
		data: {
			total_users: number;
			total_videos: number;
			total_views: number;
			total_revenue: number;
		};
	}> {
		try {
			const videoStatsResponse = await this.getStats();
			
			if (!videoStatsResponse.success) {
				throw new Error('Failed to get video stats');
			}

			const videoStats = videoStatsResponse.stats;

			return {
				success: true,
				data: {
					total_users: 0, // Will be populated from other endpoints if available
					total_videos: videoStats.total_videos,
					total_views: videoStats.total_views,
					total_revenue: 0 // Will be populated from other endpoints if available
				}
			};
		} catch (error) {
			console.error('Failed to get basic stats:', error);
			return {
				success: false,
				data: {
					total_users: 0,
					total_videos: 0,
					total_views: 0,
					total_revenue: 0
				}
			};
		}
	}
}

// Export both the class and singleton instance
export { MasterVideoService };
export const masterVideoService = new MasterVideoService(); 
