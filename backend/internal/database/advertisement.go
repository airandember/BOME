package database

import (
	"database/sql"
	"time"
)

// AdvertiserAccount represents a business account for advertising
type AdvertiserAccount struct {
	ID                int            `json:"id"`
	UserID            int            `json:"user_id"`
	CompanyName       string         `json:"company_name"`
	BusinessEmail     string         `json:"business_email"`
	ContactName       string         `json:"contact_name"`
	ContactPhone      sql.NullString `json:"contact_phone"`
	BusinessAddress   sql.NullString `json:"business_address"`
	TaxID             sql.NullString `json:"tax_id"`
	Website           sql.NullString `json:"website"`
	Industry          sql.NullString `json:"industry"`
	Status            string         `json:"status"` // pending, approved, rejected, suspended
	VerificationNotes sql.NullString `json:"verification_notes"`
	StripeCustomerID  sql.NullString `json:"stripe_customer_id"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// AdCampaign represents an advertising campaign
type AdCampaign struct {
	ID             int            `json:"id"`
	AdvertiserID   int            `json:"advertiser_id"`
	Name           string         `json:"name"`
	Description    sql.NullString `json:"description"`
	Status         string         `json:"status"` // draft, pending, approved, active, paused, completed, rejected
	StartDate      time.Time      `json:"start_date"`
	EndDate        time.Time      `json:"end_date"`
	Budget         float64        `json:"budget"`
	SpentAmount    float64        `json:"spent_amount"`
	TargetAudience sql.NullString `json:"target_audience"`
	BillingType    string         `json:"billing_type"` // weekly, monthly, custom
	BillingRate    float64        `json:"billing_rate"`
	ApprovalNotes  sql.NullString `json:"approval_notes"`
	ApprovedBy     sql.NullInt32  `json:"approved_by"`
	ApprovedAt     sql.NullTime   `json:"approved_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

// Advertisement represents individual ads within a campaign
type Advertisement struct {
	ID         int            `json:"id"`
	CampaignID int            `json:"campaign_id"`
	Title      string         `json:"title"`
	Content    sql.NullString `json:"content"`
	ImageURL   sql.NullString `json:"image_url"`
	ClickURL   string         `json:"click_url"`
	AdType     string         `json:"ad_type"` // banner, large, small
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	Priority   int            `json:"priority"`
	Status     string         `json:"status"` // active, paused, expired
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// AdPlacement represents where ads can be displayed
type AdPlacement struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Location    string    `json:"location"` // header, sidebar, footer, content, video_overlay
	AdType      string    `json:"ad_type"`  // banner, large, small
	MaxWidth    int       `json:"max_width"`
	MaxHeight   int       `json:"max_height"`
	BaseRate    float64   `json:"base_rate"` // base rate per week
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AdSchedule represents when ads should be displayed
type AdSchedule struct {
	ID          int            `json:"id"`
	AdID        int            `json:"ad_id"`
	PlacementID int            `json:"placement_id"`
	StartDate   time.Time      `json:"start_date"`
	EndDate     time.Time      `json:"end_date"`
	DaysOfWeek  sql.NullString `json:"days_of_week"` // JSON array: ["mon","tue","wed"]
	StartTime   sql.NullString `json:"start_time"`   // HH:MM format
	EndTime     sql.NullString `json:"end_time"`     // HH:MM format
	Weight      int            `json:"weight"`       // for rotation priority
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// AdAnalytics represents ad performance metrics
type AdAnalytics struct {
	ID           int       `json:"id"`
	AdID         int       `json:"ad_id"`
	Date         time.Time `json:"date"`
	Impressions  int64     `json:"impressions"`
	Clicks       int64     `json:"clicks"`
	UniqueViews  int64     `json:"unique_views"`
	ViewDuration int64     `json:"view_duration"` // in seconds
	Revenue      float64   `json:"revenue"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AdClick represents individual ad clicks for detailed tracking
type AdClick struct {
	ID        int            `json:"id"`
	AdID      int            `json:"ad_id"`
	UserID    sql.NullInt32  `json:"user_id"` // nullable for anonymous users
	IPAddress string         `json:"ip_address"`
	UserAgent sql.NullString `json:"user_agent"`
	Referrer  sql.NullString `json:"referrer"`
	ClickedAt time.Time      `json:"clicked_at"`
	CreatedAt time.Time      `json:"created_at"`
}

// AdImpression represents individual ad impressions
type AdImpression struct {
	ID           int            `json:"id"`
	AdID         int            `json:"ad_id"`
	UserID       sql.NullInt32  `json:"user_id"` // nullable for anonymous users
	IPAddress    string         `json:"ip_address"`
	UserAgent    sql.NullString `json:"user_agent"`
	ViewDuration int            `json:"view_duration"` // in seconds
	ViewedAt     time.Time      `json:"viewed_at"`
	CreatedAt    time.Time      `json:"created_at"`
}

// AdBilling represents billing records for advertisements
type AdBilling struct {
	ID              int            `json:"id"`
	CampaignID      int            `json:"campaign_id"`
	AdvertiserID    int            `json:"advertiser_id"`
	BillingPeriod   string         `json:"billing_period"` // weekly, monthly, custom
	PeriodStart     time.Time      `json:"period_start"`
	PeriodEnd       time.Time      `json:"period_end"`
	Amount          float64        `json:"amount"`
	TaxAmount       float64        `json:"tax_amount"`
	TotalAmount     float64        `json:"total_amount"`
	Status          string         `json:"status"` // pending, paid, failed, refunded
	StripeInvoiceID sql.NullString `json:"stripe_invoice_id"`
	PaymentIntentID sql.NullString `json:"payment_intent_id"`
	PaidAt          sql.NullTime   `json:"paid_at"`
	DueDate         time.Time      `json:"due_date"`
	InvoiceURL      sql.NullString `json:"invoice_url"`
	Notes           sql.NullString `json:"notes"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

// AdAuditLog represents audit trail for advertisement actions
type AdAuditLog struct {
	ID          int            `json:"id"`
	EntityType  string         `json:"entity_type"` // campaign, advertisement, billing
	EntityID    int            `json:"entity_id"`
	Action      string         `json:"action"` // created, updated, approved, rejected, etc.
	ActorID     int            `json:"actor_id"`
	ActorType   string         `json:"actor_type"` // user, admin, system
	OldValues   sql.NullString `json:"old_values"` // JSON of old values
	NewValues   sql.NullString `json:"new_values"` // JSON of new values
	IPAddress   sql.NullString `json:"ip_address"`
	UserAgent   sql.NullString `json:"user_agent"`
	Description sql.NullString `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
}

// Advertisement database migration SQL statements
const createAdvertiserAccountsTable = `
CREATE TABLE IF NOT EXISTS advertiser_accounts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    company_name VARCHAR(255) NOT NULL,
    business_email VARCHAR(255) UNIQUE NOT NULL,
    contact_name VARCHAR(255) NOT NULL,
    contact_phone VARCHAR(50),
    business_address TEXT,
    tax_id VARCHAR(100),
    website VARCHAR(500),
    industry VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending',
    verification_notes TEXT,
    stripe_customer_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdCampaignsTable = `
CREATE TABLE IF NOT EXISTS ad_campaigns (
    id SERIAL PRIMARY KEY,
    advertiser_id INTEGER REFERENCES advertiser_accounts(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'draft',
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    budget DECIMAL(10,2) NOT NULL,
    spent_amount DECIMAL(10,2) DEFAULT 0,
    target_audience TEXT,
    billing_type VARCHAR(50) DEFAULT 'weekly',
    billing_rate DECIMAL(10,2) NOT NULL,
    approval_notes TEXT,
    approved_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdvertisementsTable = `
CREATE TABLE IF NOT EXISTS advertisements (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    image_url VARCHAR(500),
    click_url VARCHAR(500) NOT NULL,
    ad_type VARCHAR(50) DEFAULT 'banner',
    width INTEGER DEFAULT 728,
    height INTEGER DEFAULT 90,
    priority INTEGER DEFAULT 1,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdPlacementsTable = `
CREATE TABLE IF NOT EXISTS ad_placements (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(100) NOT NULL,
    ad_type VARCHAR(50) DEFAULT 'banner',
    max_width INTEGER DEFAULT 728,
    max_height INTEGER DEFAULT 90,
    base_rate DECIMAL(10,2) DEFAULT 100.00,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdSchedulesTable = `
CREATE TABLE IF NOT EXISTS ad_schedules (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    placement_id INTEGER REFERENCES ad_placements(id) ON DELETE CASCADE,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    days_of_week VARCHAR(255),
    start_time VARCHAR(10),
    end_time VARCHAR(10),
    weight INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdAnalyticsTable = `
CREATE TABLE IF NOT EXISTS ad_analytics (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    date TIMESTAMP NOT NULL,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    unique_views INTEGER DEFAULT 0,
    view_duration INTEGER DEFAULT 0,
    revenue DECIMAL(10,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdClicksTable = `
CREATE TABLE IF NOT EXISTS ad_clicks (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    referrer VARCHAR(500),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdImpressionsTable = `
CREATE TABLE IF NOT EXISTS ad_impressions (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    ip_address INET,
    user_agent TEXT,
    view_duration INTEGER DEFAULT 0,
    viewed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdBillingTable = `
CREATE TABLE IF NOT EXISTS ad_billing (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    advertiser_id INTEGER REFERENCES advertiser_accounts(id) ON DELETE CASCADE,
    billing_period VARCHAR(50) NOT NULL,
    period_start TIMESTAMP NOT NULL,
    period_end TIMESTAMP NOT NULL,
    amount DECIMAL(10,2) NOT NULL,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    stripe_invoice_id VARCHAR(255),
    payment_intent_id VARCHAR(255),
    paid_at TIMESTAMP,
    due_date TIMESTAMP NOT NULL,
    invoice_url VARCHAR(500),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

const createAdAuditLogTable = `
CREATE TABLE IF NOT EXISTS ad_audit_log (
    id SERIAL PRIMARY KEY,
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    action VARCHAR(100) NOT NULL,
    actor_id INTEGER NOT NULL,
    actor_type VARCHAR(50) DEFAULT 'user',
    old_values TEXT,
    new_values TEXT,
    ip_address INET,
    user_agent TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

// CreateIndexes creates indexes for advertisement tables
const createAdvertisementIndexes = `
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_user_id ON advertiser_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_advertiser_accounts_status ON advertiser_accounts(status);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_advertiser_id ON ad_campaigns(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_ad_campaigns_status ON ad_campaigns(status);
CREATE INDEX IF NOT EXISTS idx_advertisements_campaign_id ON advertisements(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ad_schedules_ad_id ON ad_schedules(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_schedules_placement_id ON ad_schedules(placement_id);
CREATE INDEX IF NOT EXISTS idx_ad_analytics_ad_id ON ad_analytics(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_analytics_date ON ad_analytics(date);
CREATE INDEX IF NOT EXISTS idx_ad_clicks_ad_id ON ad_clicks(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_clicks_clicked_at ON ad_clicks(clicked_at);
CREATE INDEX IF NOT EXISTS idx_ad_impressions_ad_id ON ad_impressions(ad_id);
CREATE INDEX IF NOT EXISTS idx_ad_impressions_viewed_at ON ad_impressions(viewed_at);
CREATE INDEX IF NOT EXISTS idx_ad_billing_campaign_id ON ad_billing(campaign_id);
CREATE INDEX IF NOT EXISTS idx_ad_billing_advertiser_id ON ad_billing(advertiser_id);
CREATE INDEX IF NOT EXISTS idx_ad_audit_log_entity ON ad_audit_log(entity_type, entity_id);
`

// SeedAdPlacements inserts default ad placements
func (db *DB) SeedAdPlacements() error {
	placements := []struct {
		Name        string
		Description string
		Location    string
		AdType      string
		MaxWidth    int
		MaxHeight   int
		BaseRate    float64
	}{
		{"Header Banner", "Banner ad displayed in the site header", "header", "banner", 728, 90, 100.00},
		{"Sidebar Large", "Large ad displayed in the sidebar", "sidebar", "large", 300, 250, 150.00},
		{"Sidebar Small", "Small ad displayed in the sidebar", "sidebar", "small", 300, 125, 75.00},
		{"Footer Banner", "Banner ad displayed in the site footer", "footer", "banner", 728, 90, 80.00},
		{"Content Large", "Large ad displayed within content areas", "content", "large", 300, 250, 200.00},
		{"Video Overlay", "Small overlay ad displayed during video playback", "video_overlay", "small", 200, 100, 250.00},
	}

	for _, placement := range placements {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM ad_placements WHERE name = $1", placement.Name).Scan(&count)
		if err != nil {
			return err
		}

		if count == 0 {
			_, err := db.Exec(`
				INSERT INTO ad_placements (name, description, location, ad_type, max_width, max_height, base_rate)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, placement.Name, placement.Description, placement.Location, placement.AdType,
				placement.MaxWidth, placement.MaxHeight, placement.BaseRate)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
