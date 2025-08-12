package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/coupon"
	"github.com/stripe/stripe-go/v74/promotioncode"
)

// StripeCoupon represents a Stripe coupon
type StripeCoupon struct {
	ID               string            `json:"id"`
	Name             string            `json:"name,omitempty"`
	PercentOff       float64           `json:"percent_off,omitempty"`
	AmountOff        int64             `json:"amount_off,omitempty"`
	Currency         string            `json:"currency,omitempty"`
	Duration         string            `json:"duration"`
	DurationInMonths int64             `json:"duration_in_months,omitempty"`
	MaxRedemptions   int64             `json:"max_redemptions,omitempty"`
	TimesRedeemed    int64             `json:"times_redeemed"`
	Valid            bool              `json:"valid"`
	Metadata         map[string]string `json:"metadata"`
	CreatedAt        time.Time         `json:"created_at"`
}

// StripePromotionCode represents a Stripe promotion code
type StripePromotionCode struct {
	ID             string            `json:"id"`
	Code           string            `json:"code"`
	CouponID       string            `json:"coupon_id"`
	Active         bool              `json:"active"`
	MaxRedemptions int64             `json:"max_redemptions,omitempty"`
	TimesRedeemed  int64             `json:"times_redeemed"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      time.Time         `json:"created_at"`
}

// CouponRestrictions represents advanced coupon restrictions
type CouponRestrictions struct {
	FirstTimeCustomerOnly bool     `json:"first_time_customer_only"`
	CustomerIDs           []string `json:"customer_ids,omitempty"`
	CustomerEmails        []string `json:"customer_emails,omitempty"`
	MinimumAmount         *int64   `json:"minimum_amount,omitempty"`
	MaximumAmount         *int64   `json:"maximum_amount,omitempty"`
	PlanIDs               []string `json:"plan_ids,omitempty"`
	ProductIDs            []string `json:"product_ids,omitempty"`
	Categories            []string `json:"categories,omitempty"`
}

// EnhancedCreateCouponParams represents enhanced coupon creation parameters
type EnhancedCreateCouponParams struct {
	Name             string              `json:"name"`
	PercentOff       *float64            `json:"percent_off,omitempty"`
	AmountOff        *int64              `json:"amount_off,omitempty"`
	Currency         string              `json:"currency,omitempty"`
	Duration         string              `json:"duration"`
	DurationInMonths *int64              `json:"duration_in_months,omitempty"`
	MaxRedemptions   *int64              `json:"max_redemptions,omitempty"`
	RedeemBy         *int64              `json:"redeem_by,omitempty"`
	Metadata         map[string]string   `json:"metadata,omitempty"`
	Restrictions     *CouponRestrictions `json:"restrictions,omitempty"`
}

// CreateCoupon creates a new Stripe coupon
func (s *StripeService) CreateCoupon(name string, percentOff *float64, amountOff *int64, currency string, duration string, durationInMonths *int64, maxRedemptions *int64, metadata map[string]string, redeemBy *int64) (*StripeCoupon, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.CouponParams{
		Duration: stripe.String(duration),
	}

	// Set metadata using individual key-value pairs
	for key, value := range metadata {
		params.AddMetadata(key, value)
	}

	// Set name if provided
	if name != "" {
		params.Name = stripe.String(name)
	}

	// Set discount type - either percent_off or amount_off
	if percentOff != nil {
		params.PercentOff = stripe.Float64(*percentOff)
	} else if amountOff != nil {
		params.AmountOff = stripe.Int64(*amountOff)
		if currency != "" {
			params.Currency = stripe.String(currency)
		}
	} else {
		return nil, fmt.Errorf("either percent_off or amount_off must be specified")
	}

	// Set duration in months for "repeating" duration
	if durationInMonths != nil && duration == "repeating" {
		params.DurationInMonths = stripe.Int64(*durationInMonths)
	}

	// Set max redemptions
	if maxRedemptions != nil {
		params.MaxRedemptions = stripe.Int64(*maxRedemptions)
	}

	// Set expiration timestamp if provided
	if redeemBy != nil {
		params.RedeemBy = stripe.Int64(*redeemBy)
	}

	stripeCoupon, err := coupon.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe coupon: %w", err)
	}

	return s.convertCoupon(stripeCoupon), nil
}

// CreateCouponWithRestrictions creates a new Stripe coupon with advanced restrictions
func (s *StripeService) CreateCouponWithRestrictions(params *EnhancedCreateCouponParams) (*StripeCoupon, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	stripeParams := &stripe.CouponParams{
		Duration: stripe.String(params.Duration),
	}

	// Set metadata
	if params.Metadata != nil {
		for key, value := range params.Metadata {
			stripeParams.AddMetadata(key, value)
		}
	}

	// Set name if provided
	if params.Name != "" {
		stripeParams.Name = stripe.String(params.Name)
	}

	// Set discount type - either percent_off or amount_off
	if params.PercentOff != nil {
		stripeParams.PercentOff = stripe.Float64(*params.PercentOff)
	} else if params.AmountOff != nil {
		stripeParams.AmountOff = stripe.Int64(*params.AmountOff)
		if params.Currency != "" {
			stripeParams.Currency = stripe.String(params.Currency)
		}
	} else {
		return nil, fmt.Errorf("either percent_off or amount_off must be specified")
	}

	// Set duration in months for "repeating" duration
	if params.DurationInMonths != nil && params.Duration == "repeating" {
		stripeParams.DurationInMonths = stripe.Int64(*params.DurationInMonths)
	}

	// Set max redemptions
	if params.MaxRedemptions != nil {
		stripeParams.MaxRedemptions = stripe.Int64(*params.MaxRedemptions)
	}

	// Set expiration timestamp if provided
	if params.RedeemBy != nil {
		stripeParams.RedeemBy = stripe.Int64(*params.RedeemBy)
	}

	// Note: Stripe doesn't support all restrictions at the coupon level
	// Some restrictions need to be implemented at the application level
	// We'll add metadata for restrictions that can't be enforced by Stripe
	if params.Restrictions != nil {
		// Add restriction metadata for application-level enforcement
		if params.Restrictions.FirstTimeCustomerOnly {
			stripeParams.AddMetadata("restriction_first_time_only", "true")
		}
		if len(params.Restrictions.CustomerIDs) > 0 {
			stripeParams.AddMetadata("restriction_customer_ids", strings.Join(params.Restrictions.CustomerIDs, ","))
		}
		if len(params.Restrictions.CustomerEmails) > 0 {
			stripeParams.AddMetadata("restriction_customer_emails", strings.Join(params.Restrictions.CustomerEmails, ","))
		}
		if params.Restrictions.MinimumAmount != nil {
			stripeParams.AddMetadata("restriction_min_amount", strconv.FormatInt(*params.Restrictions.MinimumAmount, 10))
		}
		if params.Restrictions.MaximumAmount != nil {
			stripeParams.AddMetadata("restriction_max_amount", strconv.FormatInt(*params.Restrictions.MaximumAmount, 10))
		}
		if len(params.Restrictions.PlanIDs) > 0 {
			stripeParams.AddMetadata("restriction_plan_ids", strings.Join(params.Restrictions.PlanIDs, ","))
		}
		if len(params.Restrictions.ProductIDs) > 0 {
			stripeParams.AddMetadata("restriction_product_ids", strings.Join(params.Restrictions.ProductIDs, ","))
		}
		if len(params.Restrictions.Categories) > 0 {
			stripeParams.AddMetadata("restriction_categories", strings.Join(params.Restrictions.Categories, ","))
		}
	}

	stripeCoupon, err := coupon.New(stripeParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe coupon with restrictions: %w", err)
	}

	return s.convertCoupon(stripeCoupon), nil
}

// CreatePromotionCode creates a promotion code for a coupon
func (s *StripeService) CreatePromotionCode(couponID, code string, maxRedemptions *int64, metadata map[string]string) (*StripePromotionCode, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.PromotionCodeParams{
		Coupon: stripe.String(couponID),
		Code:   stripe.String(code),
	}

	// Set metadata using individual key-value pairs
	for key, value := range metadata {
		params.AddMetadata(key, value)
	}

	if maxRedemptions != nil {
		params.MaxRedemptions = stripe.Int64(*maxRedemptions)
	}

	stripePromotionCode, err := promotioncode.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create stripe promotion code: %w", err)
	}

	return s.convertPromotionCode(stripePromotionCode), nil
}

// GetCoupon retrieves a Stripe coupon by ID
func (s *StripeService) GetCoupon(couponID string) (*StripeCoupon, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	stripeCoupon, err := coupon.Get(couponID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe coupon: %w", err)
	}

	return s.convertCoupon(stripeCoupon), nil
}

// GetPromotionCode retrieves a Stripe promotion code by ID
func (s *StripeService) GetPromotionCode(promotionCodeID string) (*StripePromotionCode, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	stripePromotionCode, err := promotioncode.Get(promotionCodeID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get stripe promotion code: %w", err)
	}

	return s.convertPromotionCode(stripePromotionCode), nil
}

// UpdateCoupon updates a Stripe coupon (limited fields)
func (s *StripeService) UpdateCoupon(couponID string, name *string, metadata map[string]string) (*StripeCoupon, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.CouponParams{}

	// Set metadata using individual key-value pairs
	for key, value := range metadata {
		params.AddMetadata(key, value)
	}

	if name != nil {
		params.Name = stripe.String(*name)
	}

	stripeCoupon, err := coupon.Update(couponID, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update stripe coupon: %w", err)
	}

	return s.convertCoupon(stripeCoupon), nil
}

// DeleteCoupon deletes a Stripe coupon
func (s *StripeService) DeleteCoupon(couponID string) error {
	if !s.isEnabled {
		return fmt.Errorf("stripe service is disabled")
	}

	_, err := coupon.Del(couponID, nil)
	if err != nil {
		return fmt.Errorf("failed to delete stripe coupon: %w", err)
	}

	return nil
}

// ListCoupons retrieves all coupons
func (s *StripeService) ListCoupons(limit int64) ([]*StripeCoupon, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.CouponListParams{}
	if limit > 0 {
		params.Limit = stripe.Int64(limit)
	}

	iter := coupon.List(params)
	var coupons []*StripeCoupon

	for iter.Next() {
		stripeCoupon := iter.Current().(*stripe.Coupon)
		coupons = append(coupons, s.convertCoupon(stripeCoupon))
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list stripe coupons: %w", err)
	}

	return coupons, nil
}

// ListPromotionCodes retrieves promotion codes for a coupon
func (s *StripeService) ListPromotionCodes(couponID string, limit int64) ([]*StripePromotionCode, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	params := &stripe.PromotionCodeListParams{}
	if couponID != "" {
		params.Coupon = stripe.String(couponID)
	}
	if limit > 0 {
		params.Limit = stripe.Int64(limit)
	}

	iter := promotioncode.List(params)
	var promotionCodes []*StripePromotionCode

	for iter.Next() {
		stripePromotionCode := iter.Current().(*stripe.PromotionCode)
		promotionCodes = append(promotionCodes, s.convertPromotionCode(stripePromotionCode))
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to list stripe promotion codes: %w", err)
	}

	return promotionCodes, nil
}

// GetCouponAnalytics returns analytics data for a coupon
func (s *StripeService) GetCouponAnalytics(couponID string) (map[string]interface{}, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	coupon, err := s.GetCoupon(couponID)
	if err != nil {
		return nil, err
	}

	// Get promotion codes for this coupon
	promotionCodes, err := s.ListPromotionCodes(couponID, 100)
	if err != nil {
		return nil, err
	}

	// Calculate analytics
	totalRedemptions := coupon.TimesRedeemed
	activePromotionCodes := 0
	totalPromotionCodeRedemptions := int64(0)

	for _, pc := range promotionCodes {
		if pc.Active {
			activePromotionCodes++
		}
		totalPromotionCodeRedemptions += pc.TimesRedeemed
	}

	// Calculate conversion rate if max redemptions is set
	var conversionRate *float64
	if coupon.MaxRedemptions > 0 {
		rate := float64(totalRedemptions) / float64(coupon.MaxRedemptions) * 100
		conversionRate = &rate
	}

	analytics := map[string]interface{}{
		"coupon_id":                        coupon.ID,
		"coupon_name":                      coupon.Name,
		"total_redemptions":                totalRedemptions,
		"max_redemptions":                  coupon.MaxRedemptions,
		"conversion_rate_percent":          conversionRate,
		"active_promotion_codes":           activePromotionCodes,
		"total_promotion_code_redemptions": totalPromotionCodeRedemptions,
		"is_valid":                         coupon.Valid,
		"created_at":                       coupon.CreatedAt,
		"duration":                         coupon.Duration,
		"duration_in_months":               coupon.DurationInMonths,
		"discount_type":                    "unknown", // Will be determined below
		"discount_value":                   nil,
		"currency":                         coupon.Currency,
	}

	// Set discount type and value
	if coupon.PercentOff > 0 {
		analytics["discount_type"] = "percentage"
		analytics["discount_value"] = coupon.PercentOff
	} else if coupon.AmountOff > 0 {
		analytics["discount_type"] = "amount"
		analytics["discount_value"] = coupon.AmountOff
	}

	// Add metadata if available
	if len(coupon.Metadata) > 0 {
		analytics["metadata"] = coupon.Metadata
	}

	return analytics, nil
}

// GetCouponPerformanceReport returns a performance report for multiple coupons
func (s *StripeService) GetCouponPerformanceReport(limit int64) ([]map[string]interface{}, error) {
	if !s.isEnabled {
		return nil, fmt.Errorf("stripe service is disabled")
	}

	coupons, err := s.ListCoupons(limit)
	if err != nil {
		return nil, err
	}

	var report []map[string]interface{}
	for _, coupon := range coupons {
		analytics, err := s.GetCouponAnalytics(coupon.ID)
		if err != nil {
			// Log error but continue with other coupons
			continue
		}
		report = append(report, analytics)
	}

	return report, nil
}

// Helper methods to convert Stripe objects to our structs

func (s *StripeService) convertCoupon(stripeCoupon *stripe.Coupon) *StripeCoupon {
	coupon := &StripeCoupon{
		ID:            stripeCoupon.ID,
		Duration:      string(stripeCoupon.Duration),
		TimesRedeemed: stripeCoupon.TimesRedeemed,
		Valid:         stripeCoupon.Valid,
		Metadata:      stripeCoupon.Metadata,
		CreatedAt:     time.Unix(stripeCoupon.Created, 0),
	}

	if stripeCoupon.Name != "" {
		coupon.Name = stripeCoupon.Name
	}

	if stripeCoupon.PercentOff > 0 {
		coupon.PercentOff = stripeCoupon.PercentOff
	}

	if stripeCoupon.AmountOff > 0 {
		coupon.AmountOff = stripeCoupon.AmountOff
		coupon.Currency = string(stripeCoupon.Currency)
	}

	if stripeCoupon.DurationInMonths > 0 {
		coupon.DurationInMonths = stripeCoupon.DurationInMonths
	}

	if stripeCoupon.MaxRedemptions > 0 {
		coupon.MaxRedemptions = stripeCoupon.MaxRedemptions
	}

	return coupon
}

func (s *StripeService) convertPromotionCode(stripePromotionCode *stripe.PromotionCode) *StripePromotionCode {
	promotionCode := &StripePromotionCode{
		ID:            stripePromotionCode.ID,
		Code:          stripePromotionCode.Code,
		CouponID:      stripePromotionCode.Coupon.ID,
		Active:        stripePromotionCode.Active,
		TimesRedeemed: stripePromotionCode.TimesRedeemed,
		Metadata:      stripePromotionCode.Metadata,
		CreatedAt:     time.Unix(stripePromotionCode.Created, 0),
	}

	if stripePromotionCode.MaxRedemptions > 0 {
		promotionCode.MaxRedemptions = stripePromotionCode.MaxRedemptions
	}

	return promotionCode
}

// ValidateCouponData validates coupon data against Stripe's requirements
func ValidateCouponData(discountType string, discountValue float64, duration string, durationInMonths *int64, maxUses *int) error {
	// Validate discount type and value
	switch discountType {
	case "percentage":
		if discountValue <= 0 || discountValue > 100 {
			return fmt.Errorf("percentage discount must be between 0 and 100, got %.2f", discountValue)
		}
	case "amount":
		if discountValue <= 0 {
			return fmt.Errorf("amount discount must be greater than 0, got %.2f", discountValue)
		}
	default:
		return fmt.Errorf("invalid discount type '%s', must be 'percentage' or 'amount'", discountType)
	}

	// Validate duration
	switch duration {
	case "once", "forever":
		// These are valid and don't need duration_in_months
	case "repeating":
		if durationInMonths == nil || *durationInMonths <= 0 {
			return fmt.Errorf("duration_in_months is required for repeating coupons")
		}
		if *durationInMonths > 12 {
			return fmt.Errorf("duration_in_months cannot be greater than 12 for repeating coupons, got %d", *durationInMonths)
		}
	default:
		return fmt.Errorf("invalid duration '%s', must be 'once', 'forever', or 'repeating'", duration)
	}

	// Validate max uses
	if maxUses != nil && *maxUses <= 0 {
		return fmt.Errorf("max_uses must be greater than 0, got %d", maxUses)
	}

	return nil
}

// ValidateCouponRestrictions validates coupon restrictions
func ValidateCouponRestrictions(restrictions *CouponRestrictions) error {
	if restrictions == nil {
		return nil
	}

	// Validate minimum amount
	if restrictions.MinimumAmount != nil && *restrictions.MinimumAmount < 0 {
		return fmt.Errorf("minimum_amount cannot be negative, got %d", *restrictions.MinimumAmount)
	}

	// Validate maximum amount
	if restrictions.MaximumAmount != nil && *restrictions.MaximumAmount < 0 {
		return fmt.Errorf("maximum_amount cannot be negative, got %d", *restrictions.MaximumAmount)
	}

	// Validate amount range
	if restrictions.MinimumAmount != nil && restrictions.MaximumAmount != nil {
		if *restrictions.MinimumAmount >= *restrictions.MaximumAmount {
			return fmt.Errorf("minimum_amount (%d) must be less than maximum_amount (%d)",
				*restrictions.MinimumAmount, *restrictions.MaximumAmount)
		}
	}

	// Validate customer restrictions
	if len(restrictions.CustomerIDs) > 0 && len(restrictions.CustomerEmails) > 0 {
		return fmt.Errorf("cannot specify both customer_ids and customer_emails restrictions")
	}

	// Validate plan restrictions
	if len(restrictions.PlanIDs) > 0 && len(restrictions.ProductIDs) > 0 {
		return fmt.Errorf("cannot specify both plan_ids and product_ids restrictions")
	}

	return nil
}
