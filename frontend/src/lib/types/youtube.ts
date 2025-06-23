export interface YouTubeVideo {
	id: string;
	title: string;
	description: string;
	published_at: string;
	updated_at: string;
	thumbnail_url: string;
	video_url: string;
	embed_url: string;
	duration: string;
	view_count: number;
	tags?: string[];
	category?: string;
	created_at: string;
}

export interface YouTubeVideosResponse {
	videos: YouTubeVideo[];
	last_updated: string;
	total_count: number;
	channel?: ChannelInfo;
}

export interface ChannelInfo {
	id: string;
	title: string;
	description: string;
	subscriber_count: number;
	video_count: number;
	view_count: number;
	published_at: string;
	country: string;
	custom_url: string;
	thumbnail_url: string;
}

export interface YouTubeStatus {
	channel_id: string;
	channel_title: string;
	total_videos: number;
	last_updated: string;
	api_version: string;
	mock_mode: boolean;
	status: string;
	data_source: string;
}

export interface YouTubeState {
	videos: YouTubeVideo[];
	currentVideo: YouTubeVideo | null;
	channelInfo: ChannelInfo | null;
	status: YouTubeStatus | null;
	categories: string[];
	tags: string[];
	loading: boolean;
	error: string | null;
	searchQuery: string;
	searchResults: YouTubeVideo[];
}

// Backend data structure (for reference and type transformation)
export interface BackendYouTubeVideo {
	id: string;
	title: string;
	description: string;
	published_at: string;
	updated_at: string;
	thumbnail_url: string;
	video_url: string;
	embed_url: string;
	duration: string;
	view_count: number;
	created_at: string;
} 