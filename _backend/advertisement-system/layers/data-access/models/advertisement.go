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
// Note: Migration constants are now defined in database.go for unified management

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
