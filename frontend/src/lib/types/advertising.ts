export interface AdvertiserAccount {
	id: number;
	user_id: number;
	company_name: string;
	business_email: string;
	contact_name: string;
	contact_phone?: string;
	business_address?: string;
	tax_id?: string;
	website?: string;
	industry?: string;
	status: 'pending' | 'approved' | 'rejected' | 'cancelled';
	verification_notes?: string;
	approved_by?: number;
	approved_at?: string;
	rejected_by?: number;
	rejected_at?: string;
	cancelled_by?: number;
	cancelled_at?: string;
	stripe_customer_id?: string;
	created_at: string;
	updated_at: string;
}

export interface AdCampaign {
	id: number;
	advertiser_id: number;
	name: string;
	description?: string;
	status: 'draft' | 'pending' | 'approved' | 'active' | 'paused' | 'completed' | 'rejected' | 'cancelled';
	start_date: string;
	end_date: string;
	budget: number;
	spent_amount: number;
	target_audience?: string;
	billing_type: 'weekly' | 'monthly' | 'custom';
	billing_rate: number;
	approval_notes?: string;
	approved_by?: number;
	approved_at?: string;
	rejected_by?: number;
	rejected_at?: string;
	cancelled_by?: number;
	cancelled_at?: string;
	created_at: string;
	updated_at: string;
}

export interface Advertisement {
	id: number;
	campaign_id: number;
	title: string;
	content?: string;
	image_url?: string;
	click_url: string;
	ad_type: 'banner' | 'large' | 'small';
	width: number;
	height: number;
	priority: number;
	status: 'active' | 'paused' | 'expired';
	created_at: string;
	updated_at: string;
}

export interface AdPlacement {
	id: number;
	name: string;
	description: string;
	location: 'header' | 'sidebar' | 'footer' | 'content' | 'video_overlay' | 'between_videos';
	ad_type: 'banner' | 'large' | 'small';
	max_width: number;
	max_height: number;
	base_rate: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface AdSchedule {
	id: number;
	ad_id: number;
	placement_id: number;
	start_date: string;
	end_date: string;
	days_of_week?: string[];
	start_time?: string;
	end_time?: string;
	weight: number;
	is_active: boolean;
	created_at: string;
	updated_at: string;
}

export interface AdAnalytics {
	id: number;
	ad_id: number;
	date: string;
	impressions: number;
	clicks: number;
	unique_views: number;
	view_duration: number;
	revenue: number;
	created_at: string;
	updated_at: string;
}

export interface CampaignAnalytics {
	total_impressions: number;
	total_clicks: number;
	total_unique_views: number;
	total_revenue: number;
	ctr: number; // Click-through rate
}

export interface AdvancedAnalytics {
	campaign_id: number;
	date_range: {
		start: string;
		end: string;
	};
	metrics: {
		impressions: number;
		clicks: number;
		ctr: number;
		unique_views: number;
		revenue: number;
		cost_per_click: number;
		cost_per_impression: number;
		return_on_ad_spend: number;
	};
	demographics: {
		age_groups: Array<{ range: string; percentage: number }>;
		geographic: Array<{ location: string; impressions: number; clicks: number }>;
		device_types: Array<{ device: string; percentage: number }>;
	};
	performance_by_placement: Array<{
		placement_id: number;
		placement_name: string;
		impressions: number;
		clicks: number;
		ctr: number;
		revenue: number;
	}>;
	hourly_performance: Array<{
		hour: number;
		impressions: number;
		clicks: number;
		ctr: number;
	}>;
	daily_performance: Array<{
		date: string;
		impressions: number;
		clicks: number;
		revenue: number;
		ctr: number;
	}>;
}

export interface PlacementPerformance {
	placement_id: number;
	placement_name: string;
	total_ads: number;
	active_ads: number;
	total_impressions: number;
	total_clicks: number;
	total_revenue: number;
	average_ctr: number;
	fill_rate: number;
}

export interface AdvertiserFormData {
	company_name: string;
	business_email: string;
	contact_name: string;
	contact_phone: string;
	business_address: string;
	tax_id: string;
	website: string;
	industry: string;
}

export interface FormErrors {
	[key: string]: string;
}

export interface DashboardAnalytics {
	totalImpressions: number;
	totalClicks: number;
	totalSpent: number;
	activeCampaigns: number;
}

export interface AdminAnalytics {
	total_advertisers: number;
	active_campaigns: number;
	total_revenue: number;
	pending_approvals: number;
	top_performing_placements: Array<{
		placement_id: number;
		name: string;
		revenue: number;
		impressions: number;
	}>;
	revenue_by_month: Array<{
		month: string;
		revenue: number;
		advertisers: number;
	}>;
}

export interface ExportOptions {
	format: 'csv' | 'pdf' | 'excel';
	date_range: {
		start: string;
		end: string;
	};
	metrics: string[];
	group_by?: 'day' | 'week' | 'month';
	filters?: {
		campaign_ids?: number[];
		placement_ids?: number[];
		advertiser_ids?: number[];
	};
}

export interface AdServeRequest {
	placement_id: number;
	user_id?: number;
	device_type?: string;
	user_agent?: string;
	referrer?: string;
	location?: {
		country?: string;
		region?: string;
		city?: string;
	};
}

export interface AdServeResponse {
	ad?: Advertisement;
	placement: AdPlacement;
	tracking_data: {
		impression_url: string;
		click_url: string;
		view_tracking: boolean;
	};
}

export interface AdvertiserPackage {
	id: number;
	name: string;
	description: string;
	price: number;
	billing_cycle: 'weekly' | 'monthly' | 'quarterly' | 'yearly';
	features: AdvertiserPackageFeature[];
	limits: AdvertiserPackageLimits;
	is_active: boolean;
	is_featured: boolean;
	sort_order: number;
	created_at: string;
	updated_at: string;
}

export interface AdvertiserPackageFeature {
	id: number;
	package_id: number;
	name: string;
	description: string;
	is_included: boolean;
	limit_value?: number;
	limit_type?: 'campaigns' | 'ads' | 'impressions' | 'clicks' | 'storage';
}

export interface AdvertiserPackageLimits {
	max_campaigns: number;
	max_ads_per_campaign: number;
	max_monthly_impressions: number;
	max_file_size_mb: number;
	max_storage_gb: number;
	allowed_ad_types: ('banner' | 'large' | 'small' | 'video' | 'interactive')[];
	allowed_placements: string[];
	priority_boost: number;
	analytics_retention_days: number;
	support_level: 'basic' | 'priority' | 'premium';
}

export interface AdAsset {
	id: number;
	campaign_id: number;
	ad_id?: number;
	asset_type: 'image' | 'video' | 'audio' | 'document' | 'banner' | 'logo';
	file_name: string;
	file_path: string;
	file_size: number;
	mime_type: string;
	width?: number;
	height?: number;
	duration?: number;
	alt_text?: string;
	description?: string;
	status: 'pending' | 'approved' | 'rejected' | 'processing';
	approval_notes?: string;
	approved_by?: number;
	approved_at?: string;
	rejected_by?: number;
	rejected_at?: string;
	created_at: string;
	updated_at: string;
}

export interface AssetUploadRequest {
	campaign_id: number;
	ad_id?: number;
	asset_type: 'image' | 'video' | 'audio' | 'document' | 'banner' | 'logo';
	alt_text?: string;
	description?: string;
}

export interface AssetUploadResponse {
	asset: AdAsset;
	upload_url: string;
	upload_fields: Record<string, string>;
}

export interface CampaignAssetRequirement {
	ad_type: 'banner' | 'large' | 'small' | 'video' | 'interactive';
	required_assets: {
		asset_type: 'image' | 'video' | 'audio' | 'document' | 'banner' | 'logo';
		min_width?: number;
		max_width?: number;
		min_height?: number;
		max_height?: number;
		max_file_size_mb?: number;
		allowed_formats?: string[];
		is_required: boolean;
		description: string;
	}[];
}

// Enhanced Advertisement interface with assets
export interface EnhancedAdvertisement extends Advertisement {
	assets: AdAsset[];
	primary_asset?: AdAsset;
	asset_count: number;
	has_required_assets: boolean;
}

// Enhanced Campaign interface with assets and package info
export interface EnhancedAdCampaign extends AdCampaign {
	assets: AdAsset[];
	advertisements: EnhancedAdvertisement[];
	spent: number;
	asset_count: number;
	has_required_assets: boolean;
	package_info?: {
		package_id: number;
		package_name: string;
		remaining_campaigns: number;
		remaining_ads: number;
		remaining_storage_gb: number;
	};
	asset_requirements: CampaignAssetRequirement[];
}

export interface AdvertiserSubscription {
	id: number;
	advertiser_id: number;
	package_id: number;
	stripe_subscription_id: string;
	status: 'active' | 'canceled' | 'past_due' | 'unpaid' | 'trialing';
	current_period_start: string;
	current_period_end: string;
	cancel_at_period_end: boolean;
	usage_stats: {
		campaigns_used: number;
		ads_used: number;
		storage_used_gb: number;
		monthly_impressions: number;
		monthly_clicks: number;
	};
	created_at: string;
	updated_at: string;
}

// File upload utilities
export interface FileUploadProgress {
	file_name: string;
	file_size: number;
	uploaded_bytes: number;
	percentage: number;
	status: 'pending' | 'uploading' | 'processing' | 'completed' | 'error';
	error_message?: string;
}

export interface BulkAssetUpload {
	campaign_id: number;
	files: File[];
	progress: FileUploadProgress[];
	total_size: number;
	completed_count: number;
	failed_count: number;
	status: 'pending' | 'uploading' | 'processing' | 'completed' | 'error';
} 