import { apiRequest } from '$lib/auth';
import { apiClient } from '$lib/api/client';

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
	Tagged: boolean;
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
export class MasterVideoService {
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
		const response = await apiRequest('/admin/master-videos/sync/from-bunny', {
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
		const response = await apiRequest('/admin/master-videos/sync/to-bunny', {
			method: 'POST'
		});
		return response.json();
	}

	// Check for conflicts
	async checkConflicts(): Promise<{
		success: boolean;
		result: ConflictCheckResult;
	}> {
		const response = await apiRequest('/admin/master-videos/sync/conflicts');
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

	// Smart Tagging Methods
	async autoTagVideo(videoID: number): Promise<{
		success: boolean;
		message?: string;
		result?: {
			tags: string[];
			name: string;
			processed_title: string;
			categorized_tags: Record<string, string>;
		};
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/${videoID}/auto-tag`, {
				method: 'POST'
			});

			const data = await response.json();

			if (data.success) {
				return {
					success: true,
					message: data.message,
					result: data.result
				};
			} else {
				return {
					success: false,
					message: data.message || 'Failed to tag video',
					error: data.error
				};
			}
		} catch (error) {
			console.error('Failed to auto-tag video:', error);
			return {
				success: false,
				message: 'Failed to tag video',
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	async batchAutoTagVideos(videoIDs: number[], replace: boolean = false): Promise<{
		success: boolean;
		message?: string;
		total?: number;
		successful?: number;
		failed?: number;
		results?: Array<{
			video_id: number;
			success: boolean;
			result?: any;
			error?: string;
		}>;
		error?: string;
	}> {
		try {
			// Split into batches of 100 (backend limit)
			const batchSize = 100;
			const batches = [];
			for (let i = 0; i < videoIDs.length; i += batchSize) {
				batches.push(videoIDs.slice(i, i + batchSize));
			}

			let totalSuccessful = 0;
			let totalFailed = 0;
			let allResults: any[] = [];

			// Process each batch
			for (let batchIndex = 0; batchIndex < batches.length; batchIndex++) {
				const batch = batches[batchIndex];
				console.log(`Processing batch ${batchIndex + 1}/${batches.length} with ${batch.length} videos`);

				const response = await apiRequest('/admin/master-videos/batch-auto-tag', {
					method: 'POST',
					body: JSON.stringify({ video_ids: batch, replace: replace })
				});

				const data = await response.json();

				if (data.success) {
					totalSuccessful += data.successful || 0;
					totalFailed += data.failed || 0;
					if (data.results) {
						allResults = allResults.concat(data.results);
					}
				} else {
					// If batch fails, mark all videos in batch as failed
					totalFailed += batch.length;
					allResults = allResults.concat(batch.map(id => ({
						video_id: id,
						success: false,
						error: data.error || 'Batch processing failed'
					})));
				}

				// Add a small delay between batches to be respectful
				if (batchIndex < batches.length - 1) {
					await new Promise(resolve => setTimeout(resolve, 100));
				}
			}

			return {
				success: true,
				message: `Batch tagging completed. ${totalSuccessful} successful, ${totalFailed} failed`,
				total: videoIDs.length,
				successful: totalSuccessful,
				failed: totalFailed,
				results: allResults
			};
		} catch (error) {
			console.error('Failed to batch auto-tag videos:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}



	async getTagAnalytics(): Promise<{
		success: boolean;
		data?: {
			tag_frequency: Array<{ word: string; frequency: number }>;
			categories: Array<{ name: string; description: string; color: string }>;
			total_videos: number;
			tagged_videos: number;
			untagged_videos: number;
			tagging_percentage: number;
		};
		error?: string;
	}> {
		try {
			const response = await apiClient.get('/admin/master-videos/tags/analytics');
			
			if (response.data && (response.data as any).data) {
				return {
					success: true,
					data: (response.data as any).data
				};
			} else {
				return {
					success: false,
					error: response.error || 'Failed to get tag analytics'
				};
			}
		} catch (error) {
			console.error('Failed to get tag analytics:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	async getUntaggedVideos(limit: number = 50): Promise<{
		success: boolean;
		videos?: MasterVideo[];
		count?: number;
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/untagged?limit=${limit}`);
			const data = await response.json();
			
			if (data.success) {
				return {
					success: true,
					videos: data.videos,
					count: data.count
				};
			} else {
				return {
					success: false,
					error: data.message || 'Failed to get untagged videos'
				};
			}
		} catch (error) {
			console.error('Failed to get untagged videos:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	// Tag management methods
	async addTag(tag: string): Promise<Response> {
		try {
			const response = await apiRequest('/admin/master-videos/tags', {
				method: 'POST',
				body: JSON.stringify({ tag })
			});
			return response;
		} catch (error) {
			console.error('Failed to add tag:', error);
			throw error;
		}
	}

	async deleteTag(tagId: number): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/${tagId}`, {
				method: 'DELETE'
			});
			return response;
		} catch (error) {
			console.error('Failed to delete tag:', error);
			throw error;
		}
	}

	async assignTagToCategory(tagId: number, categoryId: number): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/${tagId}/category`, {
				method: 'PUT',
				body: JSON.stringify({ category_id: categoryId })
			});
			return response;
		} catch (error) {
			console.error('Failed to assign tag to category:', error);
			throw error;
		}
	}

	// Category management methods
	async getTagCategories(): Promise<Response> {
		try {
			const response = await apiRequest('/admin/master-videos/tags/categories');
			return response;
		} catch (error) {
			console.error('Failed to get tag categories:', error);
			throw error;
		}
	}

	async addTagCategory(category: { name: string; color: string }): Promise<Response> {
		try {
			const response = await apiRequest('/admin/master-videos/tags/categories', {
				method: 'POST',
				body: JSON.stringify(category)
			});
			return response;
		} catch (error) {
			console.error('Failed to add tag category:', error);
			throw error;
		}
	}

	async deleteTagCategory(categoryId: number): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/categories/${categoryId}`, {
				method: 'DELETE'
			});
			return response;
		} catch (error) {
			console.error('Failed to delete tag category:', error);
			throw error;
		}
	}

	// Subsite-specific tag management methods
	async getSubsiteTags(subsite: string): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}`);
			return response;
		} catch (error) {
			console.error(`Failed to get ${subsite} tags:`, error);
			throw error;
		}
	}

	async addSubsiteTag(subsite: string, tag: string): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}`, {
				method: 'POST',
				body: JSON.stringify({ tag })
			});
			return response;
		} catch (error) {
			console.error(`Failed to add ${subsite} tag:`, error);
			throw error;
		}
	}

	async deleteSubsiteTag(subsite: string, tagId: number): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}/${tagId}`, {
				method: 'DELETE'
			});
			return response;
		} catch (error) {
			console.error(`Failed to delete ${subsite} tag:`, error);
			throw error;
		}
	}

	async assignSubsiteTagToCategory(subsite: string, tagId: number, categoryId: number): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}/${tagId}/category`, {
				method: 'PUT',
				body: JSON.stringify({ category_id: categoryId })
			});
			return response;
		} catch (error) {
			console.error(`Failed to assign ${subsite} tag to category:`, error);
			throw error;
		}
	}

	async toggleTagActiveStatus(subsite: string, tagId: number, active: boolean): Promise<{
		success: boolean;
		message?: string;
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}/${tagId}/toggle-active`, {
				method: 'PUT'
			});
			const data = await response.json();
			
			if (data.success) {
				return {
					success: true,
					message: data.message
				};
			} else {
				return {
					success: false,
					error: data.error || 'Failed to toggle tag active status'
				};
			}
		} catch (error) {
			console.error('Failed to toggle tag active status:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	// Subsite-specific category management methods
	async getSubsiteCategories(subsite: string): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}/categories`);
			return response;
		} catch (error) {
			console.error(`Failed to get ${subsite} categories:`, error);
			throw error;
		}
	}

	async addSubsiteCategory(subsite: string, category: { name: string; color: string; description?: string }): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}/categories`, {
				method: 'POST',
				body: JSON.stringify(category)
			});
			return response;
		} catch (error) {
			console.error(`Failed to add ${subsite} category:`, error);
			throw error;
		}
	}

	async deleteSubsiteCategory(subsite: string, categoryId: number): Promise<Response> {
		try {
			const response = await apiRequest(`/admin/master-videos/tags/subsites/${subsite}/categories/${categoryId}`, {
				method: 'DELETE'
			});
			return response;
		} catch (error) {
			console.error(`Failed to delete ${subsite} category:`, error);
			throw error;
		}
	}

	// Article Exclusions Management
	async getArticleExclusions(subsite: string): Promise<{
		success: boolean;
		result?: Array<{
			id: number;
			word: string;
			excluded: boolean;
			subsite_id: number;
			created_at: string;
			updated_at: string;
		}>;
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/article-exclusions/${subsite}`);
			const data = await response.json();
			
			if (data.success) {
				return {
					success: true,
					result: data.result
				};
			} else {
				return {
					success: false,
					error: data.error || 'Failed to get article exclusions'
				};
			}
		} catch (error) {
			console.error('Failed to get article exclusions:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	async addArticleExclusion(subsite: string, word: string): Promise<{
		success: boolean;
		message?: string;
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/article-exclusions/${subsite}`, {
				method: 'POST',
				body: JSON.stringify({ word })
			});
			const data = await response.json();
			
			if (data.success) {
				return {
					success: true,
					message: data.message
				};
			} else {
				return {
					success: false,
					error: data.error || 'Failed to add article exclusion'
				};
			}
		} catch (error) {
			console.error('Failed to add article exclusion:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	async toggleArticleExclusion(subsite: string, word: string, excluded: boolean): Promise<{
		success: boolean;
		message?: string;
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/article-exclusions/${subsite}/toggle`, {
				method: 'PUT',
				body: JSON.stringify({ word, excluded })
			});
			const data = await response.json();
			
			if (data.success) {
				return {
					success: true,
					message: data.message
				};
			} else {
				return {
					success: false,
					error: data.error || 'Failed to toggle article exclusion'
				};
			}
		} catch (error) {
			console.error('Failed to toggle article exclusion:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}

	async removeArticleExclusion(subsite: string, word: string): Promise<{
		success: boolean;
		message?: string;
		error?: string;
	}> {
		try {
			const response = await apiRequest(`/admin/master-videos/article-exclusions/${subsite}/${encodeURIComponent(word)}`, {
				method: 'DELETE'
			});
			const data = await response.json();
			
			if (data.success) {
				return {
					success: true,
					message: data.message
				};
			} else {
				return {
					success: false,
					error: data.error || 'Failed to remove article exclusion'
				};
			}
		} catch (error) {
			console.error('Failed to remove article exclusion:', error);
			return {
				success: false,
				error: error instanceof Error ? error.message : 'Unknown error'
			};
		}
	}
}

// Export the singleton instance
export const masterVideoService = new MasterVideoService(); 
