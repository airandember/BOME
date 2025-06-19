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
	status: 'pending' | 'approved' | 'rejected';
	verification_notes?: string;
	stripe_customer_id?: string;
	created_at: string;
	updated_at: string;
}

export interface AdCampaign {
	id: number;
	advertiser_id: number;
	name: string;
	description?: string;
	status: 'draft' | 'pending' | 'approved' | 'active' | 'paused' | 'completed' | 'rejected';
	start_date: string;
	end_date?: string;
	budget: number;
	spent_amount: number;
	target_audience?: string;
	billing_type: 'weekly' | 'monthly' | 'custom';
	billing_rate: number;
	approval_notes?: string;
	approved_by?: number;
	approved_at?: string;
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