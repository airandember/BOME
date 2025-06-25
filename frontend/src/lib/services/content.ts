import { apiClient } from '$lib/api/client';
import type { ApiResponse } from '$lib/api/client';

export interface Video {
    id: number;
    title: string;
    description: string;
    duration: string;
    thumbnail: string;
    status: 'draft' | 'pending' | 'published';
    category: string;
    uploaded_by: {
        id: number;
        name: string;
        email: string;
    };
    upload_date: string;
    views: number;
    likes: number;
    comments: number;
    file_size: string;
    resolution: string;
    tags: string[];
}

export interface VideoStats {
    total_videos: number;
    published: number;
    pending: number;
    draft: number;
    total_views: number;
    total_likes: number;
    total_comments: number;
    total_duration: string;
    storage_used: string;
    top_categories: Array<{
        name: string;
        count: number;
        views: number;
    }>;
    recent_uploads: Array<{
        date: string;
        count: number;
    }>;
}

export interface VideoCategory {
    id: number;
    name: string;
    description: string;
    video_count: number;
}

export interface VideoUpdateRequest {
    title?: string;
    description?: string;
    category?: string;
    status?: 'draft' | 'pending' | 'published';
    tags?: string[];
}

export interface BulkVideoOperation {
    operation: 'publish' | 'unpublish' | 'delete';
    video_ids: number[];
}

class ContentService {
    private static instance: ContentService;
    private cache: Map<string, { data: any; timestamp: number }> = new Map();
    private readonly CACHE_DURATION = 5 * 60 * 1000; // 5 minutes

    private constructor() {}

    public static getInstance(): ContentService {
        if (!ContentService.instance) {
            ContentService.instance = new ContentService();
        }
        return ContentService.instance;
    }

    // Video Management
    public async getVideos(params: {
        page?: number;
        limit?: number;
        search?: string;
        category?: string;
        status?: string;
        sort?: string;
        order?: 'asc' | 'desc';
    }): Promise<{ videos: Video[]; pagination: any }> {
        const response = await apiClient.getAdminVideos();
        if (!response.data) {
            throw new Error(response.error || 'Failed to fetch videos');
        }
        return response.data;
    }

    public async getVideo(id: number): Promise<Video> {
        const response = await apiClient.getAdminVideos();
        if (!response.data?.videos) {
            throw new Error(response.error || 'Failed to fetch video');
        }
        const video = response.data.videos.find(v => v.id === id);
        if (!video) {
            throw new Error('Video not found');
        }
        return video;
    }

    public async updateVideo(id: number, data: VideoUpdateRequest): Promise<void> {
        const response = await apiClient.request<void>(`/admin/videos/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            headers: { 'Content-Type': 'application/json' }
        });
        if (response.error) {
            throw new Error(response.error);
        }
    }

    public async deleteVideo(id: number): Promise<void> {
        const response = await apiClient.request<void>(`/admin/videos/${id}`, {
            method: 'DELETE'
        });
        if (response.error) {
            throw new Error(response.error);
        }
    }

    public async bulkVideoOperation(operation: BulkVideoOperation): Promise<void> {
        const response = await apiClient.request<void>('/admin/videos/bulk', {
            method: 'POST',
            body: JSON.stringify(operation),
            headers: { 'Content-Type': 'application/json' }
        });
        if (response.error) {
            throw new Error(response.error);
        }
    }

    public async getVideoStats(): Promise<VideoStats> {
        const cacheKey = 'video_stats';
        const cached = this.cache.get(cacheKey);
        
        if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
            return cached.data;
        }

        const response = await apiClient.getAdminAnalytics();
        if (!response.data?.analytics) {
            throw new Error(response.error || 'Failed to fetch video stats');
        }

        const stats = response.data.analytics;
        this.cache.set(cacheKey, {
            data: stats,
            timestamp: Date.now()
        });

        return stats;
    }

    public async getVideoCategories(): Promise<VideoCategory[]> {
        const cacheKey = 'video_categories';
        const cached = this.cache.get(cacheKey);
        
        if (cached && Date.now() - cached.timestamp < this.CACHE_DURATION) {
            return cached.data;
        }

        const response = await apiClient.getVideoCategories();
        if (!response.data?.categories) {
            throw new Error(response.error || 'Failed to fetch video categories');
        }

        const categories = response.data.categories;
        this.cache.set(cacheKey, {
            data: categories,
            timestamp: Date.now()
        });

        return categories;
    }

    public async uploadVideo(file: File, metadata: {
        title: string;
        description: string;
        category: string;
        tags?: string[];
    }): Promise<Video> {
        const formData = new FormData();
        formData.append('video', file);
        formData.append('title', metadata.title);
        formData.append('description', metadata.description);
        formData.append('category', metadata.category);
        if (metadata.tags) {
            formData.append('tags', JSON.stringify(metadata.tags));
        }

        const response = await apiClient.request<{ video: Video }>('/admin/videos/upload', {
            method: 'POST',
            body: formData,
            headers: { 'Content-Type': 'multipart/form-data' }
        });

        if (!response.data?.video) {
            throw new Error(response.error || 'Failed to upload video');
        }

        return response.data.video;
    }

    public async scheduleVideo(videoId: number, publishDate: Date): Promise<void> {
        const response = await apiClient.request<void>(`/admin/videos/${videoId}/schedule`, {
            method: 'POST',
            body: JSON.stringify({ publish_date: publishDate.toISOString() }),
            headers: { 'Content-Type': 'application/json' }
        });
        if (response.error) {
            throw new Error(response.error);
        }
    }

    public async getScheduledVideos(): Promise<Video[]> {
        const response = await apiClient.request<{ videos: Video[] }>('/admin/videos/scheduled', {
            method: 'GET'
        });
        if (!response.data?.videos) {
            throw new Error(response.error || 'Failed to fetch scheduled videos');
        }
        return response.data.videos;
    }

    // Cache Management
    public clearCache(): void {
        this.cache.clear();
    }

    public clearCacheItem(key: string): void {
        this.cache.delete(key);
    }
}

export const contentService = ContentService.getInstance(); 