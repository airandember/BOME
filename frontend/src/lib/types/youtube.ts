export interface YouTubeVideo {
	id: string;
	title: string;
	description: string;
	published_at: string;
	updated_at: string;
	thumbnail_url: string;
	video_url: string;
	embed_url: string;
	duration?: string;
	view_count?: number;
	created_at: string;
}

export interface YouTubeVideosResponse {
	videos: YouTubeVideo[];
	last_updated: string;
	total_count: number;
}

export interface YouTubeStatus {
	channel_id: string;
	channel_name: string;
	total_videos: number;
	last_updated: string;
	webhook_active: boolean;
}

export interface YouTubeSubscriptionResponse {
	success: boolean;
	message: string;
} 