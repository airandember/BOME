// Enhanced Subscriber Interface for Unified Dashboard
export interface EnhancedSubscriber {
  // Core User Data
  id: number;
  email: string;
  first_name: string;
  last_name: string;
  full_name: string;
  email_verified: boolean;
  role: string;
  
  // Subscription Status (Key Goals)
  has_active_plan: boolean;           // Goal #1: Active plan within date range
  has_video_access: boolean;          // Goal #2: Access to premium videos
  
  // Plan Information
  plan_name: string;
  plan_legacy_status: 'Legacy' | 'Current' | 'Unknown';
  plan_type: 'premium' | 'basic' | 'none';
  plan_start_date: string | null;
  billing_period_start: string | null;
  billing_period_end: string | null;
  plan_status: 'active' | 'expired' | 'trial' | 'cancelled' | 'none';
  plan_price: number;
  plan_currency: string;
  
  // Business Intelligence
  mrr_contribution: number;
  ltv_estimate: number;
  days_until_expiry: number | null;
  subscription_duration_days: number | null;
  
  // Admin Tracking
  last_login: string | null;
  created_at: string;
  updated_at: string;
  manual_access_granted: boolean;
  stripe_customer_id: string | null;
  
  // Computed Fields
  is_expiring_soon: boolean; // Within 7 days
  is_high_value: boolean;    // Above average MRR
  account_age_days: number;
  billing_cycle_length: number | null;
  
  // Additional fields for management
  account_status: string;
  stripe_subscription_id: string | null;
}

export interface SubscriberFilters {
  search?: string;
  plan_type?: 'premium' | 'basic' | 'none' | '';
  has_active_plan?: boolean | null;
  has_video_access?: boolean | null;
  // video_access_source removed - plans are now the only source
  is_expiring_soon?: boolean | null;
  email_verified?: boolean | null;
  role?: string;
  created_date_from?: string;
  created_date_to?: string;
  last_login_from?: string;
  last_login_to?: string;
  min_mrr?: number;
  max_mrr?: number;
}

export interface SubscriberKPIs {
  total_subscribers: number;
  active_subscribers: number;
  video_access_users: number;
  total_mrr: number;
  avg_days_to_expiry: number;
  churn_risk_count: number;
  premium_users: number;
  basic_users: number;
  manual_access_users: number;
}

export interface SubscriberResponse {
  subscribers: EnhancedSubscriber[];
  total_count: number;
  kpis: SubscriberKPIs;
  pagination: {
    page: number;
    limit: number;
    total_pages: number;
    has_next: boolean;
    has_prev: boolean;
  };
}

export interface BulkAction {
  id: string;
  label: string;
  icon: string;
  variant?: 'primary' | 'secondary' | 'danger';
  requiresConfirmation?: boolean;
}

export interface BulkActionRequest {
  action: 'grant_video_access' | 'revoke_video_access' | 'extend_trial' | 'send_email' | 'export_data';
  subscriber_ids: number[];
  parameters?: {
    days?: number;
    email_template?: string;
    export_format?: 'csv' | 'xlsx';
  };
}

export interface QuickFilter {
  label: string;
  icon: string;
  filter: Partial<SubscriberFilters>;
  description: string;
}
