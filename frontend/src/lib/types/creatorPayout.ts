/**
 * Creator Payout System Types
 */

// ========================================
// PRESENTER TYPES
// ========================================

export interface Presenter {
	id: number;
	user_id?: number;
	name: string;
	email?: string;
	bio?: string;
	avatar_url?: string;
	payment_method?: string;
	stripe_connect_id?: string;
	paypal_email?: string;
	tax_id?: string;
	bank_account_last4?: string;
	address_line1?: string;
	address_line2?: string;
	city?: string;
	state?: string;
	postal_code?: string;
	country: string;
	is_active: boolean;
	verified: boolean;
	verified_at?: string;
	verified_by?: number;
	total_videos: number;
	total_views: number;
	total_watch_minutes: number;
	total_earnings: number;
	lifetime_paid: number;
	notes?: string;
	internal_notes?: string;
	created_at: string;
	updated_at: string;
}

export interface CreatePresenterInput {
	user_id?: number;
	name: string;
	email?: string;
	bio?: string;
	avatar_url?: string;
	payment_method?: string;
	stripe_connect_id?: string;
	paypal_email?: string;
	tax_id?: string;
	address_line1?: string;
	address_line2?: string;
	city?: string;
	state?: string;
	postal_code?: string;
	country?: string;
	notes?: string;
	internal_notes?: string;
}

export interface UpdatePresenterInput {
	name?: string;
	email?: string;
	bio?: string;
	avatar_url?: string;
	payment_method?: string;
	stripe_connect_id?: string;
	paypal_email?: string;
	tax_id?: string;
	address_line1?: string;
	address_line2?: string;
	city?: string;
	state?: string;
	postal_code?: string;
	country?: string;
	is_active?: boolean;
	notes?: string;
	internal_notes?: string;
}

export interface PresenterStats {
	total_presenters: number;
	active_presenters: number;
	verified_presenters: number;
	total_videos: number;
	total_views: number;
	total_earnings: number;
	total_paid: number;
	pending_payouts: number;
}

// ========================================
// VIDEO-PRESENTER TYPES
// ========================================

export interface VideoPresenter {
	id: number;
	video_id: number;
	video_title?: string;
	presenter_id: number;
	presenter_name?: string;
	role: string;
	attribution_percentage: number;
	is_primary: boolean;
	display_order: number;
	notes?: string;
	added_at: string;
}

export interface CreateVideoPresenterInput {
	video_id: number;
	presenter_id: number;
	role?: string;
	attribution_percentage?: number;
	is_primary?: boolean;
	display_order?: number;
	notes?: string;
}

export interface UpdateVideoPresenterInput {
	role?: string;
	attribution_percentage?: number;
	is_primary?: boolean;
	display_order?: number;
	notes?: string;
}

// ========================================
// PAYOUT FORMULA TYPES
// ========================================

export interface PayoutFormula {
	id: number;
	name: string;
	description?: string;
	formula_type: 'per_view' | 'per_watch_minute' | 'tier_based' | 'flat_rate' | 'hybrid';
	base_rate: number;
	tier_config?: any;
	subscriber_multiplier: number;
	completion_multiplier: number;
	engagement_multiplier: number;
	completion_threshold: number;
	engagement_threshold: number;
	min_payout: number;
	max_payout?: number;
	is_active: boolean;
	is_default: boolean;
	effective_date?: string;
	expiration_date?: string;
	created_by?: number;
	created_at: string;
	updated_at: string;
}

export interface CreatePayoutFormulaInput {
	name: string;
	description?: string;
	formula_type: string;
	base_rate: number;
	tier_config?: any;
	subscriber_multiplier?: number;
	completion_multiplier?: number;
	engagement_multiplier?: number;
	completion_threshold?: number;
	engagement_threshold?: number;
	min_payout?: number;
	max_payout?: number;
	is_active?: boolean;
	is_default?: boolean;
	effective_date?: string;
	expiration_date?: string;
}

export interface UpdatePayoutFormulaInput {
	name?: string;
	description?: string;
	base_rate?: number;
	tier_config?: any;
	subscriber_multiplier?: number;
	completion_multiplier?: number;
	engagement_multiplier?: number;
	completion_threshold?: number;
	engagement_threshold?: number;
	min_payout?: number;
	max_payout?: number;
	is_active?: boolean;
	is_default?: boolean;
	effective_date?: string;
	expiration_date?: string;
}

// ========================================
// PAYOUT TYPES
// ========================================

export interface PresenterPayout {
	id: number;
	presenter_id: number;
	presenter_name?: string;
	presenter_email?: string;
	formula_id?: number;
	formula_name?: string;
	payout_month: string;
	total_videos: number;
	total_views: number;
	total_watch_minutes: number;
	unique_viewers: number;
	subscriber_views: number;
	avg_completion_rate: number;
	total_engagement: number;
	base_amount: number;
	bonus_amount: number;
	adjustment_amount: number;
	deductions: number;
	final_amount: number;
	currency: string;
	status: 'pending' | 'approved' | 'processing' | 'paid' | 'failed' | 'cancelled' | 'on_hold';
	payment_method?: string;
	payment_reference?: string;
	payment_fee: number;
	paid_at?: string;
	calculation_data?: any;
	notes?: string;
	admin_notes?: string;
	calculated_by?: number;
	calculated_at?: string;
	approved_by?: number;
	approved_at?: string;
	paid_by?: number;
	created_at: string;
	updated_at: string;
}

export interface UpdatePayoutStatusInput {
	status: string;
	payment_method?: string;
	payment_reference?: string;
	payment_fee?: number;
	notes?: string;
	admin_notes?: string;
}

export interface UpdatePayoutAmountsInput {
	adjustment_amount: number;
	deductions: number;
	admin_notes?: string;
}

export interface PayoutSummary {
	payout_month: string;
	total_presenters: number;
	total_videos: number;
	total_views: number;
	total_amount: number;
	pending_amount: number;
	approved_amount: number;
	paid_amount: number;
	status_breakdown: any;
}

// ========================================
// TRANSACTION TYPES
// ========================================

export interface PayoutTransaction {
	id: number;
	payout_id?: number;
	presenter_id: number;
	presenter_name?: string;
	transaction_type: 'payment' | 'adjustment' | 'refund' | 'chargeback' | 'bonus' | 'correction';
	amount: number;
	currency: string;
	payment_method?: string;
	payment_provider?: string;
	provider_transaction_id?: string;
	status: 'pending' | 'processing' | 'completed' | 'failed' | 'reversed';
	description?: string;
	notes?: string;
	error_message?: string;
	fee_amount: number;
	net_amount: number;
	processed_by?: number;
	processed_at?: string;
	created_at: string;
	updated_at?: string;
}

export interface CreatePayoutTransactionInput {
	payout_id?: number;
	presenter_id: number;
	transaction_type: string;
	amount: number;
	currency?: string;
	payment_method?: string;
	payment_provider?: string;
	provider_transaction_id?: string;
	description?: string;
	notes?: string;
	fee_amount?: number;
}

export interface UpdateTransactionStatusInput {
	status: string;
	error_message?: string;
}

// ========================================
// DASHBOARD TYPES
// ========================================

export interface DashboardStats {
	presenter_stats: PresenterStats;
	current_month_summary?: PayoutSummary;
	recent_payouts: PresenterPayout[];
	recent_transactions: PayoutTransaction[];
	default_formula?: PayoutFormula;
}

