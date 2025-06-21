package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"bome-backend/internal/database"
)

// AdvertisementService handles advertisement business logic
type AdvertisementService struct {
	db *database.DB
}

// NewAdvertisementService creates a new advertisement service
func NewAdvertisementService(db *database.DB) *AdvertisementService {
	return &AdvertisementService{db: db}
}

// CreateAdvertiserAccount creates a new advertiser account
func (s *AdvertisementService) CreateAdvertiserAccount(userID int, req *CreateAdvertiserRequest) (*database.AdvertiserAccount, error) {
	// Check if user already has an advertiser account
	existing, _ := s.GetAdvertiserAccountByUserID(userID)
	if existing != nil {
		return nil, fmt.Errorf("user already has an advertiser account")
	}

	var id int
	err := s.db.QueryRow(`
		INSERT INTO advertiser_accounts (user_id, company_name, business_email, contact_name, contact_phone, business_address, tax_id, website, industry)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, userID, req.CompanyName, req.BusinessEmail, req.ContactName, req.ContactPhone, req.BusinessAddress, req.TaxID, req.Website, req.Industry).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to create advertiser account: %w", err)
	}

	return s.GetAdvertiserAccountByID(id)
}

// GetAdvertiserAccountByID retrieves an advertiser account by ID
func (s *AdvertisementService) GetAdvertiserAccountByID(id int) (*database.AdvertiserAccount, error) {
	account := &database.AdvertiserAccount{}
	err := s.db.QueryRow(`
		SELECT id, user_id, company_name, business_email, contact_name, contact_phone, business_address, tax_id, website, industry, status, verification_notes, stripe_customer_id, created_at, updated_at
		FROM advertiser_accounts WHERE id = $1
	`, id).Scan(&account.ID, &account.UserID, &account.CompanyName, &account.BusinessEmail, &account.ContactName, &account.ContactPhone, &account.BusinessAddress, &account.TaxID, &account.Website, &account.Industry, &account.Status, &account.VerificationNotes, &account.StripeCustomerID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return account, nil
}

// GetAdvertiserAccountByUserID retrieves an advertiser account by user ID
func (s *AdvertisementService) GetAdvertiserAccountByUserID(userID int) (*database.AdvertiserAccount, error) {
	account := &database.AdvertiserAccount{}
	err := s.db.QueryRow(`
		SELECT id, user_id, company_name, business_email, contact_name, contact_phone, business_address, tax_id, website, industry, status, verification_notes, stripe_customer_id, created_at, updated_at
		FROM advertiser_accounts WHERE user_id = $1
	`, userID).Scan(&account.ID, &account.UserID, &account.CompanyName, &account.BusinessEmail, &account.ContactName, &account.ContactPhone, &account.BusinessAddress, &account.TaxID, &account.Website, &account.Industry, &account.Status, &account.VerificationNotes, &account.StripeCustomerID, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return account, nil
}

// ApproveAdvertiserAccount approves an advertiser account
func (s *AdvertisementService) ApproveAdvertiserAccount(id int, notes string) error {
	// Start a transaction to ensure both updates succeed or fail together
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the advertiser account to find the user_id
	var userID int
	err = tx.QueryRow(`
		SELECT user_id FROM advertiser_accounts WHERE id = $1
	`, id).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get advertiser account: %w", err)
	}

	// Update advertiser account status
	_, err = tx.Exec(`
		UPDATE advertiser_accounts 
		SET status = 'approved', verification_notes = $1, updated_at = NOW()
		WHERE id = $2
	`, notes, id)
	if err != nil {
		return fmt.Errorf("failed to approve advertiser account: %w", err)
	}

	// Update user role to 'advertiser'
	_, err = tx.Exec(`
		UPDATE users 
		SET role = 'advertiser', updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	// Commit the transaction
	return tx.Commit()
}

// RejectAdvertiserAccount rejects an advertiser account
func (s *AdvertisementService) RejectAdvertiserAccount(id int, notes string) error {
	// Start a transaction to ensure both updates succeed or fail together
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the advertiser account to find the user_id
	var userID int
	err = tx.QueryRow(`
		SELECT user_id FROM advertiser_accounts WHERE id = $1
	`, id).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get advertiser account: %w", err)
	}

	// Update advertiser account status
	_, err = tx.Exec(`
		UPDATE advertiser_accounts 
		SET status = 'rejected', verification_notes = $1, updated_at = NOW()
		WHERE id = $2
	`, notes, id)
	if err != nil {
		return fmt.Errorf("failed to reject advertiser account: %w", err)
	}

	// Reset user role back to 'user'
	_, err = tx.Exec(`
		UPDATE users 
		SET role = 'user', updated_at = NOW()
		WHERE id = $1
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	// Commit the transaction
	return tx.Commit()
}

// CreateAdCampaign creates a new advertising campaign
func (s *AdvertisementService) CreateAdCampaign(advertiserID int, req *CreateCampaignRequest) (*database.AdCampaign, error) {
	// Verify advertiser is approved
	advertiser, err := s.GetAdvertiserAccountByID(advertiserID)
	if err != nil {
		return nil, fmt.Errorf("advertiser not found: %w", err)
	}
	if advertiser.Status != "approved" {
		return nil, fmt.Errorf("advertiser account not approved")
	}

	var id int
	err = s.db.QueryRow(`
		INSERT INTO ad_campaigns (advertiser_id, name, description, start_date, end_date, budget, target_audience, billing_type, billing_rate)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, advertiserID, req.Name, req.Description, req.StartDate, req.EndDate, req.Budget, req.TargetAudience, req.BillingType, req.BillingRate).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to create campaign: %w", err)
	}

	return s.GetAdCampaignByID(id)
}

// GetAdCampaignByID retrieves a campaign by ID
func (s *AdvertisementService) GetAdCampaignByID(id int) (*database.AdCampaign, error) {
	campaign := &database.AdCampaign{}
	err := s.db.QueryRow(`
		SELECT id, advertiser_id, name, description, status, start_date, end_date, budget, spent_amount, target_audience, billing_type, billing_rate, approval_notes, approved_by, approved_at, created_at, updated_at
		FROM ad_campaigns WHERE id = $1
	`, id).Scan(&campaign.ID, &campaign.AdvertiserID, &campaign.Name, &campaign.Description, &campaign.Status, &campaign.StartDate, &campaign.EndDate, &campaign.Budget, &campaign.SpentAmount, &campaign.TargetAudience, &campaign.BillingType, &campaign.BillingRate, &campaign.ApprovalNotes, &campaign.ApprovedBy, &campaign.ApprovedAt, &campaign.CreatedAt, &campaign.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return campaign, nil
}

// GetCampaignsByAdvertiser retrieves all campaigns for an advertiser
func (s *AdvertisementService) GetCampaignsByAdvertiser(advertiserID int) ([]*database.AdCampaign, error) {
	rows, err := s.db.Query(`
		SELECT id, advertiser_id, name, description, status, start_date, end_date, budget, spent_amount, target_audience, billing_type, billing_rate, approval_notes, approved_by, approved_at, created_at, updated_at
		FROM ad_campaigns WHERE advertiser_id = $1 ORDER BY created_at DESC
	`, advertiserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []*database.AdCampaign
	for rows.Next() {
		campaign := &database.AdCampaign{}
		err := rows.Scan(&campaign.ID, &campaign.AdvertiserID, &campaign.Name, &campaign.Description, &campaign.Status, &campaign.StartDate, &campaign.EndDate, &campaign.Budget, &campaign.SpentAmount, &campaign.TargetAudience, &campaign.BillingType, &campaign.BillingRate, &campaign.ApprovalNotes, &campaign.ApprovedBy, &campaign.ApprovedAt, &campaign.CreatedAt, &campaign.UpdatedAt)
		if err != nil {
			return nil, err
		}
		campaigns = append(campaigns, campaign)
	}
	return campaigns, nil
}

// ApproveCampaign approves an advertising campaign
func (s *AdvertisementService) ApproveCampaign(campaignID, approverID int, notes string) error {
	now := time.Now()
	_, err := s.db.Exec(`
		UPDATE ad_campaigns 
		SET status = 'approved', approval_notes = $1, approved_by = $2, approved_at = $3, updated_at = NOW()
		WHERE id = $4
	`, notes, approverID, now, campaignID)
	return err
}

// RejectCampaign rejects an advertising campaign
func (s *AdvertisementService) RejectCampaign(campaignID, approverID int, notes string) error {
	_, err := s.db.Exec(`
		UPDATE ad_campaigns 
		SET status = 'rejected', approval_notes = $1, approved_by = $2, updated_at = NOW()
		WHERE id = $3
	`, notes, approverID, campaignID)
	return err
}

// CreateAdvertisement creates a new advertisement
func (s *AdvertisementService) CreateAdvertisement(campaignID int, req *CreateAdRequest) (*database.Advertisement, error) {
	// Verify campaign exists and is approved
	campaign, err := s.GetAdCampaignByID(campaignID)
	if err != nil {
		return nil, fmt.Errorf("campaign not found: %w", err)
	}
	if campaign.Status != "approved" {
		return nil, fmt.Errorf("campaign not approved")
	}

	var id int
	err = s.db.QueryRow(`
		INSERT INTO advertisements (campaign_id, title, content, image_url, click_url, ad_type, width, height, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`, campaignID, req.Title, req.Content, req.ImageURL, req.ClickURL, req.AdType, req.Width, req.Height, req.Priority).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to create advertisement: %w", err)
	}

	return s.GetAdvertisementByID(id)
}

// GetAdvertisementByID retrieves an advertisement by ID
func (s *AdvertisementService) GetAdvertisementByID(id int) (*database.Advertisement, error) {
	ad := &database.Advertisement{}
	err := s.db.QueryRow(`
		SELECT id, campaign_id, title, content, image_url, click_url, ad_type, width, height, priority, status, created_at, updated_at
		FROM advertisements WHERE id = $1
	`, id).Scan(&ad.ID, &ad.CampaignID, &ad.Title, &ad.Content, &ad.ImageURL, &ad.ClickURL, &ad.AdType, &ad.Width, &ad.Height, &ad.Priority, &ad.Status, &ad.CreatedAt, &ad.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return ad, nil
}

// GetActiveAdsForPlacement retrieves active ads for a specific placement
func (s *AdvertisementService) GetActiveAdsForPlacement(placementID int) ([]*database.Advertisement, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.campaign_id, a.title, a.content, a.image_url, a.click_url, a.ad_type, a.width, a.height, a.priority, a.status, a.created_at, a.updated_at
		FROM advertisements a
		JOIN ad_schedules s ON a.id = s.ad_id
		JOIN ad_campaigns c ON a.campaign_id = c.id
		WHERE s.placement_id = $1 
		AND a.status = 'active' 
		AND c.status = 'active'
		AND s.is_active = true
		AND NOW() BETWEEN s.start_date AND s.end_date
		ORDER BY a.priority DESC, RANDOM()
	`, placementID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ads []*database.Advertisement
	for rows.Next() {
		ad := &database.Advertisement{}
		err := rows.Scan(&ad.ID, &ad.CampaignID, &ad.Title, &ad.Content, &ad.ImageURL, &ad.ClickURL, &ad.AdType, &ad.Width, &ad.Height, &ad.Priority, &ad.Status, &ad.CreatedAt, &ad.UpdatedAt)
		if err != nil {
			return nil, err
		}
		ads = append(ads, ad)
	}
	return ads, nil
}

// RecordAdImpression records an ad impression
func (s *AdvertisementService) RecordAdImpression(adID int, userID *int, ipAddress, userAgent string, viewDuration int) error {
	_, err := s.db.Exec(`
		INSERT INTO ad_impressions (ad_id, user_id, ip_address, user_agent, view_duration, viewed_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, adID, userID, ipAddress, userAgent, viewDuration)

	if err != nil {
		return err
	}

	// Update daily analytics
	return s.updateDailyAnalytics(adID, "impression", 1)
}

// RecordAdClick records an ad click
func (s *AdvertisementService) RecordAdClick(adID int, userID *int, ipAddress, userAgent, referrer string) error {
	_, err := s.db.Exec(`
		INSERT INTO ad_clicks (ad_id, user_id, ip_address, user_agent, referrer, clicked_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, adID, userID, ipAddress, userAgent, referrer)

	if err != nil {
		return err
	}

	// Update daily analytics
	return s.updateDailyAnalytics(adID, "click", 1)
}

// updateDailyAnalytics updates the daily analytics for an ad
func (s *AdvertisementService) updateDailyAnalytics(adID int, metricType string, value int64) error {
	today := time.Now().Format("2006-01-02")

	var column string
	switch metricType {
	case "impression":
		column = "impressions"
	case "click":
		column = "clicks"
	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}

	_, err := s.db.Exec(fmt.Sprintf(`
		INSERT INTO ad_analytics (ad_id, date, %s)
		VALUES ($1, $2, $3)
		ON CONFLICT (ad_id, date)
		DO UPDATE SET %s = ad_analytics.%s + $3, updated_at = NOW()
	`, column, column, column), adID, today, value)

	return err
}

// GetAdAnalytics retrieves analytics for an advertisement
func (s *AdvertisementService) GetAdAnalytics(adID int, startDate, endDate time.Time) ([]*database.AdAnalytics, error) {
	rows, err := s.db.Query(`
		SELECT id, ad_id, date, impressions, clicks, unique_views, view_duration, revenue, created_at, updated_at
		FROM ad_analytics 
		WHERE ad_id = $1 AND date BETWEEN $2 AND $3
		ORDER BY date DESC
	`, adID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analytics []*database.AdAnalytics
	for rows.Next() {
		analytic := &database.AdAnalytics{}
		err := rows.Scan(&analytic.ID, &analytic.AdID, &analytic.Date, &analytic.Impressions, &analytic.Clicks, &analytic.UniqueViews, &analytic.ViewDuration, &analytic.Revenue, &analytic.CreatedAt, &analytic.UpdatedAt)
		if err != nil {
			return nil, err
		}
		analytics = append(analytics, analytic)
	}
	return analytics, nil
}

// GetCampaignAnalytics retrieves analytics for a campaign
func (s *AdvertisementService) GetCampaignAnalytics(campaignID int, startDate, endDate time.Time) (*CampaignAnalytics, error) {
	var analytics CampaignAnalytics

	err := s.db.QueryRow(`
		SELECT 
			COALESCE(SUM(impressions), 0) as total_impressions,
			COALESCE(SUM(clicks), 0) as total_clicks,
			COALESCE(SUM(unique_views), 0) as total_unique_views,
			COALESCE(SUM(revenue), 0) as total_revenue
		FROM ad_analytics aa
		JOIN advertisements a ON aa.ad_id = a.id
		WHERE a.campaign_id = $1 AND aa.date BETWEEN $2 AND $3
	`, campaignID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).Scan(
		&analytics.TotalImpressions, &analytics.TotalClicks, &analytics.TotalUniqueViews, &analytics.TotalRevenue)

	if err != nil {
		return nil, err
	}

	// Calculate CTR
	if analytics.TotalImpressions > 0 {
		analytics.CTR = float64(analytics.TotalClicks) / float64(analytics.TotalImpressions) * 100
	}

	return &analytics, nil
}

// LogAdAction logs an action for audit purposes
func (s *AdvertisementService) LogAdAction(entityType string, entityID int, action string, actorID int, actorType string, oldValues, newValues interface{}, ipAddress, userAgent, description string) error {
	var oldJSON, newJSON sql.NullString

	if oldValues != nil {
		oldBytes, _ := json.Marshal(oldValues)
		oldJSON = sql.NullString{String: string(oldBytes), Valid: true}
	}

	if newValues != nil {
		newBytes, _ := json.Marshal(newValues)
		newJSON = sql.NullString{String: string(newBytes), Valid: true}
	}

	_, err := s.db.Exec(`
		INSERT INTO ad_audit_log (entity_type, entity_id, action, actor_id, actor_type, old_values, new_values, ip_address, user_agent, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, entityType, entityID, action, actorID, actorType, oldJSON, newJSON, ipAddress, userAgent, description)

	return err
}

// Request/Response structures
type CreateAdvertiserRequest struct {
	CompanyName     string `json:"company_name" validate:"required,max=255"`
	BusinessEmail   string `json:"business_email" validate:"required,email,max=255"`
	ContactName     string `json:"contact_name" validate:"required,max=255"`
	ContactPhone    string `json:"contact_phone,omitempty" validate:"max=50"`
	BusinessAddress string `json:"business_address,omitempty"`
	TaxID           string `json:"tax_id,omitempty" validate:"max=100"`
	Website         string `json:"website,omitempty" validate:"url,max=255"`
	Industry        string `json:"industry,omitempty" validate:"max=100"`
}

type CreateCampaignRequest struct {
	Name           string    `json:"name" validate:"required,max=255"`
	Description    string    `json:"description,omitempty"`
	StartDate      time.Time `json:"start_date" validate:"required"`
	EndDate        time.Time `json:"end_date" validate:"required"`
	Budget         float64   `json:"budget" validate:"required,min=0"`
	TargetAudience string    `json:"target_audience,omitempty"`
	BillingType    string    `json:"billing_type" validate:"required,oneof=weekly monthly custom"`
	BillingRate    float64   `json:"billing_rate" validate:"required,min=0"`
}

type CreateAdRequest struct {
	Title    string `json:"title" validate:"required,max=255"`
	Content  string `json:"content,omitempty"`
	ImageURL string `json:"image_url,omitempty" validate:"url,max=500"`
	ClickURL string `json:"click_url" validate:"required,url,max=500"`
	AdType   string `json:"ad_type" validate:"required,oneof=banner large small"`
	Width    int    `json:"width" validate:"required,min=1"`
	Height   int    `json:"height" validate:"required,min=1"`
	Priority int    `json:"priority" validate:"min=1,max=10"`
}

type CampaignAnalytics struct {
	TotalImpressions int64   `json:"total_impressions"`
	TotalClicks      int64   `json:"total_clicks"`
	TotalUniqueViews int64   `json:"total_unique_views"`
	TotalRevenue     float64 `json:"total_revenue"`
	CTR              float64 `json:"ctr"` // Click-through rate
}
