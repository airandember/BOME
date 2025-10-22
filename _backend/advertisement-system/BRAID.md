# 🧬 Advertisement System Braid - Backend
**Campaign management, ad placement, and revenue optimization**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **backend portion** of the Advertisement System Braid.  
> **Frontend portion**: See `_frontend/braids/advertisement-system/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Backend Overview**

**Purpose**: Server-side advertisement management, campaign tracking, and monetization  
**Technology**: Go, PostgreSQL  
**Complexity**: Medium-High (Campaign Management, Targeting, Analytics)  
**Dependencies**: Auth Braid (advertiser identity), Analytics Braid (tracking)

---

## 🎯 **Key Features**

### **1. Advertiser Management**:
- Advertiser account creation
- Profile management
- Account verification
- Billing setup
- Access control

### **2. Campaign Management**:
- Create, edit, delete campaigns
- Campaign scheduling (start/end dates)
- Budget management
- Status control (active, paused, completed)
- Campaign goals and KPIs

### **3. Advertisement Creation**:
- Ad content management
- Multiple ad formats (banner, video, native)
- Creative asset upload
- Ad copy and messaging
- Call-to-action buttons

### **4. Ad Placement**:
- Placement positions (homepage, sidebar, video pre-roll)
- Targeting rules (demographics, interests, behavior)
- Frequency capping
- A/B testing support
- Geographic targeting

### **5. Performance Tracking**:
- Impressions
- Clicks (CTR)
- Conversions
- Cost per click (CPC)
- Return on ad spend (ROAS)
- Real-time analytics

---

## 🗄️ **Database Schema**

### **Advertisers Table**:
```sql
CREATE TABLE advertisers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    company_name VARCHAR(255) NOT NULL,
    website VARCHAR(500),
    industry VARCHAR(100),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    billing_address TEXT,
    tax_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'active', 'suspended'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_advertisers_user_id ON advertisers(user_id);
CREATE INDEX idx_advertisers_status ON advertisers(status);
```

---

### **Campaigns Table**:
```sql
CREATE TABLE campaigns (
    id SERIAL PRIMARY KEY,
    advertiser_id INTEGER REFERENCES advertisers(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    objective VARCHAR(100), -- 'awareness', 'traffic', 'conversions'
    budget DECIMAL(10, 2),
    daily_budget DECIMAL(10, 2),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(50) DEFAULT 'draft', -- 'draft', 'active', 'paused', 'completed'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_campaigns_advertiser_id ON campaigns(advertiser_id);
CREATE INDEX idx_campaigns_status ON campaigns(status);
CREATE INDEX idx_campaigns_dates ON campaigns(start_date, end_date);
```

---

### **Advertisements Table**:
**File**: `backend/internal/database/advertisement.go`

```sql
CREATE TABLE advertisements (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    ad_type VARCHAR(50), -- 'banner', 'video', 'native', 'sponsored'
    title VARCHAR(255),
    description TEXT,
    image_url VARCHAR(500),
    video_url VARCHAR(500),
    click_url VARCHAR(500) NOT NULL,
    cta_text VARCHAR(100), -- Call-to-action text
    status VARCHAR(50) DEFAULT 'pending_review', -- 'pending_review', 'approved', 'rejected', 'active', 'paused'
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_advertisements_campaign_id ON advertisements(campaign_id);
CREATE INDEX idx_advertisements_status ON advertisements(status);
CREATE INDEX idx_advertisements_ad_type ON advertisements(ad_type);
```

---

### **Ad Placements Table**:
```sql
CREATE TABLE ad_placements (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    location VARCHAR(100), -- 'homepage_hero', 'sidebar', 'video_preroll', 'article_inline'
    description TEXT,
    dimensions VARCHAR(50), -- '300x250', '728x90', 'responsive'
    price_per_impression DECIMAL(10, 4),
    price_per_click DECIMAL(10, 2),
    max_ads_per_page INTEGER DEFAULT 1,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

### **Ad Impressions Table** (Analytics):
```sql
CREATE TABLE ad_impressions (
    id SERIAL PRIMARY KEY,
    advertisement_id INTEGER REFERENCES advertisements(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    placement_id INTEGER REFERENCES ad_placements(id),
    ip_address VARCHAR(45),
    user_agent TEXT,
    device_type VARCHAR(50), -- 'desktop', 'mobile', 'tablet'
    country VARCHAR(100),
    city VARCHAR(100),
    clicked BOOLEAN DEFAULT false,
    clicked_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_ad_impressions_ad_id ON ad_impressions(advertisement_id);
CREATE INDEX idx_ad_impressions_created_at ON ad_impressions(created_at);
CREATE INDEX idx_ad_impressions_clicked ON ad_impressions(clicked);

-- Partitioning by date for performance
CREATE TABLE ad_impressions_2025_10 PARTITION OF ad_impressions
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
```

---

### **Campaign Analytics Table** (Aggregated):
```sql
CREATE TABLE campaign_analytics (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES campaigns(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    conversions INTEGER DEFAULT 0,
    cost DECIMAL(10, 2) DEFAULT 0,
    revenue DECIMAL(10, 2) DEFAULT 0,
    ctr DECIMAL(5, 2), -- Click-through rate
    cpc DECIMAL(10, 2), -- Cost per click
    cpa DECIMAL(10, 2), -- Cost per acquisition
    roas DECIMAL(10, 2), -- Return on ad spend
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(campaign_id, date)
);

CREATE INDEX idx_campaign_analytics_campaign_id ON campaign_analytics(campaign_id);
CREATE INDEX idx_campaign_analytics_date ON campaign_analytics(date);
```

---

## 🌐 **API Endpoints**

### **Advertiser Management**:
```
POST   /api/v1/advertisers/register      # Register as advertiser
GET    /api/v1/advertisers/profile       # Get advertiser profile
PUT    /api/v1/advertisers/profile       # Update profile
GET    /api/v1/advertisers/:id           # Get advertiser (admin)
```

### **Campaigns**:
```
GET    /api/v1/campaigns                 # List campaigns (advertiser's)
GET    /api/v1/campaigns/:id             # Get campaign details
POST   /api/v1/campaigns                 # Create campaign
PUT    /api/v1/campaigns/:id             # Update campaign
DELETE /api/v1/campaigns/:id             # Delete campaign
POST   /api/v1/campaigns/:id/activate    # Activate campaign
POST   /api/v1/campaigns/:id/pause       # Pause campaign
```

### **Advertisements**:
**File**: `backend/internal/routes/advertisement.go`

```
GET    /api/v1/advertisements            # List ads
GET    /api/v1/advertisements/:id        # Get ad details
POST   /api/v1/advertisements            # Create ad
PUT    /api/v1/advertisements/:id        # Update ad
DELETE /api/v1/advertisements/:id        # Delete ad
POST   /api/v1/advertisements/:id/submit # Submit for review
```

### **Ad Serving** (Public):
```
GET    /api/v1/ads/serve?placement=homepage_hero    # Serve ad for placement
POST   /api/v1/ads/:id/impression                   # Track impression
POST   /api/v1/ads/:id/click                        # Track click
```

### **Analytics**:
```
GET    /api/v1/campaigns/:id/analytics   # Campaign performance
GET    /api/v1/advertisements/:id/analytics # Ad performance
GET    /api/v1/advertisers/dashboard     # Advertiser dashboard data
```

### **Admin**:
```
GET    /api/v1/admin/advertisements/pending  # Pending ads for review
POST   /api/v1/admin/advertisements/:id/approve  # Approve ad
POST   /api/v1/admin/advertisements/:id/reject   # Reject ad
GET    /api/v1/admin/advertisers          # List all advertisers
```

---

## 🔧 **Backend Services**

### **Advertisement Service** (`backend/internal/services/advertisement.go`):

**Key Functions**:
```go
// Ad serving
func ServeAd(placementID string, context AdContext) (*Advertisement, error)

// Impression tracking
func TrackImpression(adID int, userID *int, metadata ImpressionMetadata) error

// Click tracking
func TrackClick(adID int, userID *int, metadata ClickMetadata) error

// Ad selection algorithm
func SelectAdForPlacement(placement *AdPlacement, context AdContext) (*Advertisement, error)
```

### **Campaign Service**:
```go
// Campaign CRUD
func CreateCampaign(campaign *Campaign) error
func UpdateCampaign(campaign *Campaign) error
func GetCampaign(id int) (*Campaign, error)
func DeleteCampaign(id int) error

// Campaign lifecycle
func ActivateCampaign(id int) error
func PauseCampaign(id int) error
func CompleteCampaign(id int) error

// Budget management
func CheckBudgetAvailable(campaignID int, amount float64) (bool, error)
func DeductBudget(campaignID int, amount float64) error
```

### **Analytics Service**:
```go
// Real-time metrics
func GetCampaignMetrics(campaignID int, dateRange DateRange) (*Metrics, error)
func GetAdMetrics(adID int, dateRange DateRange) (*Metrics, error)

// Aggregation
func AggregateImpressions(campaignID int, date time.Time) error
func CalculatePerformanceMetrics(campaignID int) error

// Reporting
func GenerateAdvertiserReport(advertiserID int, period string) (*Report, error)
```

---

## 🎯 **Ad Selection Algorithm**

### **Simple Algorithm** (Current):
```go
func SelectAdForPlacement(placement *AdPlacement, context AdContext) (*Advertisement, error) {
    // 1. Get active campaigns with budget
    campaigns := GetActiveCampaignsWithBudget()
    
    // 2. Filter by placement compatibility
    compatibleAds := FilterByPlacementType(campaigns, placement.AdType)
    
    // 3. Apply targeting rules
    targetedAds := ApplyTargeting(compatibleAds, context)
    
    // 4. Check frequency caps
    eligibleAds := CheckFrequencyCaps(targetedAds, context.UserID)
    
    // 5. Select by weight (price × priority)
    selectedAd := WeightedRandomSelection(eligibleAds)
    
    return selectedAd, nil
}
```

### **Targeting Rules**:
```go
type TargetingRules struct {
    Countries    []string
    AgeRange     *AgeRange
    Interests    []string
    DeviceTypes  []string
    TimeOfDay    *TimeRange
}

func ApplyTargeting(ads []*Advertisement, context AdContext) []*Advertisement {
    filtered := []Advertisement{}
    
    for _, ad := range ads {
        if !ad.Rules.Matches(context) {
            continue
        }
        filtered = append(filtered, ad)
    }
    
    return filtered
}
```

---

## 📊 **Performance Metrics**

### **Key Metrics**:
- **Impressions**: Number of times ad is displayed
- **Clicks**: Number of times ad is clicked
- **CTR** (Click-Through Rate): (Clicks / Impressions) × 100
- **CPC** (Cost Per Click): Total cost / Total clicks
- **CPM** (Cost Per Mille): (Total cost / Impressions) × 1000
- **Conversions**: Goal completions (signups, purchases, etc.)
- **CPA** (Cost Per Acquisition): Total cost / Conversions
- **ROAS** (Return on Ad Spend): Revenue / Cost

### **Calculation Examples**:
```go
type CampaignMetrics struct {
    Impressions  int
    Clicks       int
    Conversions  int
    Cost         float64
    Revenue      float64
    CTR          float64 // (Clicks / Impressions) * 100
    CPC          float64 // Cost / Clicks
    CPA          float64 // Cost / Conversions
    ROAS         float64 // Revenue / Cost
}

func CalculateMetrics(campaignID int, dateRange DateRange) (*CampaignMetrics, error) {
    impressions := CountImpressions(campaignID, dateRange)
    clicks := CountClicks(campaignID, dateRange)
    conversions := CountConversions(campaignID, dateRange)
    cost := CalculateCost(campaignID, dateRange)
    revenue := CalculateRevenue(campaignID, dateRange)
    
    metrics := &CampaignMetrics{
        Impressions: impressions,
        Clicks:      clicks,
        Conversions: conversions,
        Cost:        cost,
        Revenue:     revenue,
        CTR:         (float64(clicks) / float64(impressions)) * 100,
        CPC:         cost / float64(clicks),
        CPA:         cost / float64(conversions),
        ROAS:        revenue / cost,
    }
    
    return metrics, nil
}
```

---

## 🔒 **Access Control & Security**

### **Advertiser Permissions**:
- ✅ Create/edit/delete own campaigns
- ✅ View own campaign analytics
- ✅ Update own profile
- ❌ Cannot approve own ads
- ❌ Cannot access other advertisers' data

### **Admin Permissions**:
- ✅ Approve/reject advertisements
- ✅ View all campaigns and advertisers
- ✅ Suspend advertiser accounts
- ✅ Manage ad placements
- ✅ Access all analytics

### **Ad Content Moderation**:
```go
func ReviewAdvertisement(adID int, approved bool, reason string) error {
    ad := GetAdvertisement(adID)
    
    if approved {
        ad.Status = "approved"
        NotifyAdvertiser(ad.CampaignID, "Ad approved")
    } else {
        ad.Status = "rejected"
        ad.RejectionReason = reason
        NotifyAdvertiser(ad.CampaignID, "Ad rejected: " + reason)
    }
    
    return UpdateAdvertisement(ad)
}
```

---

## ⚡ **Performance Optimizations**

### **Caching**:
- **Active campaigns**: 5-minute cache
- **Ad placements**: 10-minute cache
- **Advertiser profiles**: 15-minute cache

### **Database Optimization**:
- Indexed fields for fast lookups
- Partitioned impression tables by date
- Aggregated analytics tables
- Read replicas for analytics queries

### **Ad Serving**:
- <50ms response time for ad selection
- CDN for ad creative assets
- Async impression/click tracking

---

## 💰 **Billing Integration**

### **Pricing Models**:
1. **CPM** (Cost Per Mille): Pay per 1,000 impressions
2. **CPC** (Cost Per Click): Pay per click
3. **CPA** (Cost Per Acquisition): Pay per conversion
4. **Flat Rate**: Fixed price for time period

### **Budget Management**:
```go
func CheckAndDeductBudget(campaignID int, cost float64) (bool, error) {
    campaign := GetCampaign(campaignID)
    
    // Check daily budget
    todaySpend := GetTodaySpend(campaignID)
    if todaySpend + cost > campaign.DailyBudget {
        return false, errors.New("daily budget exceeded")
    }
    
    // Check total budget
    totalSpend := GetTotalSpend(campaignID)
    if totalSpend + cost > campaign.Budget {
        return false, errors.New("total budget exceeded")
    }
    
    // Deduct budget
    err := DeductBudget(campaignID, cost)
    return true, err
}
```

---

## 📝 **Known Technical Debt**

### **Current Limitations**:
1. Simple ad selection (no ML/AI)
2. Limited targeting options
3. Manual ad approval process
4. Basic analytics (no attribution)
5. No real-time bidding
6. Limited fraud detection

### **Future Enhancements**:
1. ✅ ML-based ad targeting
2. ✅ Real-time bidding system
3. ✅ Automated content moderation
4. ✅ Advanced attribution modeling
5. ✅ Fraud detection algorithms
6. ✅ A/B testing framework
7. ✅ Programmatic advertising

---

## 🧬 **Strands (Complete Flows)**

### **1. Campaign Management Strand**:
Complete flow from campaign creation to activation

### **2. Ad Placement Strand**:
Ad serving, impression tracking, and click handling

---

## 🚀 **Quick Start**

### **Understanding Advertisement System** (15 min):
1. Read this BRAID.md (7 min)
2. Check database schema (5 min)
3. Review ad serving algorithm (3 min)

---

**Last Updated**: October 14, 2025  
**Status**: Core implementation documented  
**Technology**: Go + PostgreSQL  
**Frontend Counterpart**: `_frontend/braids/advertisement-system/`

---

**Navigate**:  
[🏠 Master Index](../../BRAIDS_INDEX.md) | [🎨 Frontend Braid](../../_frontend/braids/advertisement-system/BRAID.md)

