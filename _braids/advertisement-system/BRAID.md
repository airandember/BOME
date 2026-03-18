# Braid: advertisement-system

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**Campaign management, ad placement, and revenue optimization**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **backend portion** of the Advertisement System Braid.  
> **Frontend portion**: See `_frontend/braids/advertisement-system/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Server-side advertisement management, campaign tracking, and monetization  
**Technology**: Go, PostgreSQL  
**Complexity**: Medium-High (Campaign Management, Targeting, Analytics)  
**Dependencies**: Auth Braid (advertiser identity), Analytics Braid (tracking)

---

## **File Map** (Production Code)

| Layer | Production Path | Description |
|-------|-----------------|-------------|
| Handlers | `backend/advertisement/handlers/advertisement.go` | Ad API handlers |
| Services | `backend/advertisement/services/advertisement.go` | Ad serving logic |
| Models | `backend/advertisement/models/advertisement.go` | Ad/campaign data models |
| Routes | `backend/internal/routes/advertisement.go` | Ad API routes |
| Database | `backend/internal/database/advertisement.go` | Ad tables |

**Frontend:** `frontend/src/routes/advertise/`, `frontend/src/lib/` (AdDisplay component)

---

## **Key Features**

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

## ðŸ—„ï¸ **Database Schema**

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

## ðŸŒ **API Endpoints**

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

## ðŸ”§ **Backend Services**

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

## ðŸŽ¯ **Ad Selection Algorithm**

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
    
    // 5. Select by weight (price Ã— priority)
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

## ðŸ“Š **Performance Metrics**

### **Key Metrics**:
- **Impressions**: Number of times ad is displayed
- **Clicks**: Number of times ad is clicked
- **CTR** (Click-Through Rate): (Clicks / Impressions) Ã— 100
- **CPC** (Cost Per Click): Total cost / Total clicks
- **CPM** (Cost Per Mille): (Total cost / Impressions) Ã— 1000
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

## ðŸ”’ **Access Control & Security**

### **Advertiser Permissions**:
- âœ… Create/edit/delete own campaigns
- âœ… View own campaign analytics
- âœ… Update own profile
- âŒ Cannot approve own ads
- âŒ Cannot access other advertisers' data

### **Admin Permissions**:
- âœ… Approve/reject advertisements
- âœ… View all campaigns and advertisers
- âœ… Suspend advertiser accounts
- âœ… Manage ad placements
- âœ… Access all analytics

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

## âš¡ **Performance Optimizations**

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

## ðŸ’° **Billing Integration**

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

## ðŸ“ **Known Technical Debt**

### **Current Limitations**:
1. Simple ad selection (no ML/AI)
2. Limited targeting options
3. Manual ad approval process
4. Basic analytics (no attribution)
5. No real-time bidding
6. Limited fraud detection

### **Future Enhancements**:
1. âœ… ML-based ad targeting
2. âœ… Real-time bidding system
3. âœ… Automated content moderation
4. âœ… Advanced attribution modeling
5. âœ… Fraud detection algorithms
6. âœ… A/B testing framework
7. âœ… Programmatic advertising

---

## ðŸ§¬ **Strands (Complete Flows)**

### **1. Campaign Management Strand**:
Complete flow from campaign creation to activation

### **2. Ad Placement Strand**:
Ad serving, impression tracking, and click handling

---

## ðŸš€ **Quick Start**

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
[ðŸ  Master Index](../../BRAIDS_INDEX.md) | [ðŸŽ¨ Frontend Braid](../../_frontend/braids/advertisement-system/BRAID.md)



---

## Frontend Architecture

**Svelte5 UI for advertisers, campaigns, and ad management**

---

## ðŸ”— **Cross-Repository Braid**

> **âš ï¸ IMPORTANT**: This is the **frontend portion** of the Advertisement System Braid.  
> **Backend portion**: See `_braids/advertisement-system/backend/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: User interface for advertisers to manage campaigns and view analytics  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/advertise`, `/advertiser`, `/admin/advertisements`  
**State Management**: Svelte stores for campaign and analytics data

---

## ðŸŽ¯ **Key Features**

### **1. Public Advertise Page**:
- Value proposition
- Pricing information
- Platform statistics
- Get started CTA
- Testimonials from advertisers

### **2. Advertiser Dashboard**:
- Campaign overview cards
- Performance summary
- Quick actions
- Spend overview
- Alert notifications

### **3. Campaign Management**:
- Create/edit campaigns
- Set budgets and schedules
- Campaign status control
- Performance tracking
- Bulk actions

### **4. Ad Creation**:
- Ad creative upload
- Ad copy editor
- Preview component
- Multiple ad formats
- Targeting options

### **5. Analytics Dashboard**:
- Real-time metrics
- Performance charts
- Date range filtering
- Export reports
- Comparison views

---

## ðŸ“„ **Frontend Pages**

### **1. Advertise Landing Page** (`/advertise`)
**File**: `frontend/src/routes/advertise/+page.svelte`

**Features**:
- Hero section with CTA
- Benefits showcase
- Pricing tiers
- Success stories
- FAQ section
- Sign up form

**Example UI**:
```svelte
<div class="advertise-page">
  <section class="hero">
    <h1>Reach Your Target Audience</h1>
    <p>Advertise on BOME and connect with engaged users</p>
    <button class="cta">Start Advertising</button>
  </section>
  
  <section class="benefits">
    <div class="benefit">
      <h3>ðŸŽ¯ Targeted Reach</h3>
      <p>Advanced targeting options</p>
    </div>
    <div class="benefit">
      <h3>ðŸ“Š Real-time Analytics</h3>
      <p>Track performance instantly</p>
    </div>
    <div class="benefit">
      <h3>ðŸ’° Flexible Pricing</h3>
      <p>CPM, CPC, or flat rate</p>
    </div>
  </section>
  
  <section class="pricing">
    <div class="tier">
      <h3>Self-Service</h3>
      <p>$0.50 CPM</p>
      <ul>
        <li>Self-service dashboard</li>
        <li>Basic analytics</li>
        <li>24/7 support</li>
      </ul>
    </div>
  </section>
</div>
```

---

### **2. Advertiser Dashboard** (`/advertiser`)
**File**: `frontend/src/routes/advertiser/+page.svelte`

**Features**:
- Summary cards (impressions, clicks, spend)
- Active campaigns list
- Performance chart
- Recent activity
- Quick create campaign button

**Example UI**:
```svelte
<script>
  import { advertiserStore } from '$lib/stores/advertiser';
  import { onMount } from 'svelte';
  
  onMount(async () => {
    await advertiserStore.loadDashboard();
  });
  
  $: metrics = $advertiserStore.dashboardMetrics;
</script>

<div class="advertiser-dashboard">
  <header>
    <h1>Dashboard</h1>
    <button on:click={() => goto('/advertiser/campaigns/new')}>
      + New Campaign
    </button>
  </header>
  
  <!-- Metrics Cards -->
  <div class="metrics-grid">
    <MetricCard
      title="Impressions"
      value={metrics.totalImpressions}
      change={metrics.impressionChange}
    />
    <MetricCard
      title="Clicks"
      value={metrics.totalClicks}
      change={metrics.clickChange}
    />
    <MetricCard
      title="CTR"
      value={metrics.averageCTR + '%'}
      change={metrics.ctrChange}
    />
    <MetricCard
      title="Spend"
      value={'$' + metrics.totalSpend}
      change={metrics.spendChange}
    />
  </div>
  
  <!-- Performance Chart -->
  <div class="chart-container">
    <h2>Performance Over Time</h2>
    <PerformanceChart data={metrics.chartData} />
  </div>
  
  <!-- Active Campaigns -->
  <div class="campaigns-list">
    <h2>Active Campaigns</h2>
    {#each $advertiserStore.activeCampaigns as campaign}
      <CampaignCard {campaign} />
    {/each}
  </div>
</div>
```

---

### **3. Campaign Management** (`/advertiser/campaigns`)
**File**: `frontend/src/routes/advertiser/campaigns/+page.svelte`

**Features**:
- Campaign list with filters
- Status badges (active, paused, draft)
- Performance metrics per campaign
- Bulk actions
- Create campaign button
- Sort and search

---

### **4. Campaign Editor** (`/advertiser/campaigns/new`)
**File**: `frontend/src/routes/advertiser/campaigns/[id]/+page.svelte`

**Features**:
- Campaign name and description
- Objective selection
- Budget settings
- Date range picker
- Save as draft / Launch buttons

**Example UI**:
```svelte
<form on:submit|preventDefault={handleSave}>
  <div class="form-field">
    <label>Campaign Name</label>
    <input bind:value={campaign.name} required />
  </div>
  
  <div class="form-field">
    <label>Objective</label>
    <select bind:value={campaign.objective}>
      <option value="awareness">Brand Awareness</option>
      <option value="traffic">Website Traffic</option>
      <option value="conversions">Conversions</option>
    </select>
  </div>
  
  <div class="form-row">
    <div class="form-field">
      <label>Total Budget</label>
      <input type="number" bind:value={campaign.budget} />
    </div>
    <div class="form-field">
      <label>Daily Budget</label>
      <input type="number" bind:value={campaign.dailyBudget} />
    </div>
  </div>
  
  <div class="form-row">
    <div class="form-field">
      <label>Start Date</label>
      <input type="date" bind:value={campaign.startDate} />
    </div>
    <div class="form-field">
      <label>End Date</label>
      <input type="date" bind:value={campaign.endDate} />
    </div>
  </div>
  
  <div class="form-actions">
    <button type="button" on:click={saveDraft}>
      Save as Draft
    </button>
    <button type="submit" class="primary">
      Launch Campaign
    </button>
  </div>
</form>
```

---

### **5. Ad Analytics** (`/advertiser/analytics`)
**File**: `frontend/src/routes/advertiser/analytics/+page.svelte`

**Features**:
- Date range selector
- Metrics overview
- Performance charts (line, bar, pie)
- Campaign comparison
- Export to CSV
- Filter by campaign

---

### **6. Admin Ad Review** (`/admin/advertisements`)
**File**: `frontend/src/routes/admin/advertisements/+page.svelte`

**Features**:
- Pending ads queue
- Ad preview
- Approve/reject buttons
- Rejection reason form
- Ad details display
- Advertiser information

---

## ðŸ§© **Frontend Components**

### **AdDisplay Component** (Public Facing)
**File**: `frontend/src/lib/components/AdDisplay.svelte`

**Purpose**: Display advertisements on the platform

**Features**:
- Responsive ad container
- Image/video ad support
- Click tracking
- Impression tracking
- Lazy loading

**Usage**:
```svelte
<script>
  import AdDisplay from '$lib/components/AdDisplay.svelte';
</script>

<AdDisplay placement="homepage_hero" />
<AdDisplay placement="sidebar" />
<AdDisplay placement="video_preroll" />
```

**Implementation**:
```svelte
<script lang="ts">
  import { onMount } from 'svelte';
  import { advertiserStore } from '$lib/stores/advertiser';
  
  export let placement: string;
  
  let ad = null;
  let visible = false;
  
  onMount(async () => {
    // Fetch ad for this placement
    ad = await advertiserStore.fetchAd(placement);
    
    // Track impression when visible
    const observer = new IntersectionObserver((entries) => {
      if (entries[0].isIntersecting && !visible) {
        visible = true;
        trackImpression(ad.id);
      }
    });
    
    observer.observe(adContainer);
  });
  
  function handleClick() {
    trackClick(ad.id);
    window.open(ad.clickUrl, '_blank');
  }
</script>

{#if ad}
  <div class="ad-container" bind:this={adContainer}>
    <div class="ad-label">Advertisement</div>
    
    {#if ad.type === 'image'}
      <img
        src={ad.imageUrl}
        alt={ad.title}
        on:click={handleClick}
      />
    {:else if ad.type === 'video'}
      <video src={ad.videoUrl} controls />
    {/if}
    
    <div class="ad-content">
      <h3>{ad.title}</h3>
      <p>{ad.description}</p>
      <button on:click={handleClick}>
        {ad.ctaText || 'Learn More'}
      </button>
    </div>
  </div>
{/if}
```

---

### **CampaignCard Component**
**Purpose**: Display campaign summary

**Features**:
- Campaign name and status
- Key metrics
- Progress bar
- Quick actions (pause, edit, view)

---

### **PerformanceChart Component**
**Purpose**: Visualize campaign performance

**Features**:
- Line chart for trends
- Multiple metrics support
- Responsive
- Tooltip on hover
- Legend

---

## ðŸ—ƒï¸ **Frontend Stores**

### **Advertiser Store** (`$lib/stores/advertiser.ts`)
**Purpose**: Manage advertiser state

**State**:
```typescript
interface AdvertiserState {
  profile: AdvertiserProfile | null;
  dashboardMetrics: DashboardMetrics;
  activeCampaigns: Campaign[];
  loading: boolean;
  error: string | null;
}
```

**Methods**:
```typescript
export const advertiserStore = {
  async loadDashboard() {
    // GET /api/v1/advertisers/dashboard
  },
  
  async fetchAd(placement: string): Promise<Ad> {
    // GET /api/v1/ads/serve?placement={placement}
  },
  
  async trackImpression(adId: number) {
    // POST /api/v1/ads/:id/impression
  },
  
  async trackClick(adId: number) {
    // POST /api/v1/ads/:id/click
  }
};
```

---

### **Campaign Store** (`$lib/stores/campaigns.ts`)
**Purpose**: Manage campaign state

**State**:
```typescript
interface CampaignState {
  campaigns: Campaign[];
  currentCampaign: Campaign | null;
  loading: boolean;
}
```

**Methods**:
```typescript
export const campaigns = {
  async loadCampaigns() {
    // GET /api/v1/campaigns
  },
  
  async createCampaign(data: CampaignData) {
    // POST /api/v1/campaigns
  },
  
  async updateCampaign(id: number, data: CampaignData) {
    // PUT /api/v1/campaigns/:id
  },
  
  async activateCampaign(id: number) {
    // POST /api/v1/campaigns/:id/activate
  },
  
  async pauseCampaign(id: number) {
    // POST /api/v1/campaigns/:id/pause
  }
};
```

---

## ðŸ”„ **Data Flow Examples**

### **Display Advertisement**:
```
1. Page loads with <AdDisplay placement="sidebar" />
2. Component fetches ad for placement
3. API: GET /ads/serve?placement=sidebar
4. Backend selects appropriate ad
5. Ad displayed to user
6. Intersection Observer detects visibility
7. Track impression (async)
8. User clicks ad
9. Track click (async)
10. Open ad link in new tab
```

### **Create Campaign**:
```
1. Advertiser clicks "New Campaign"
2. Fill in campaign form
3. Click "Launch"
4. Store.createCampaign() called
5. API: POST /api/v1/campaigns
6. Backend creates campaign
7. Success response
8. Redirect to campaign details
9. Show success notification
```

---

## ðŸ“Š **Analytics Visualization**

### **Chart Types**:
1. **Line Chart**: Performance over time
2. **Bar Chart**: Campaign comparison
3. **Pie Chart**: Budget allocation
4. **Gauge**: CTR, CPC metrics

### **Libraries**:
- Chart.js or D3.js
- Responsive and interactive
- Export to PNG/CSV

---

## ðŸ”’ **Security**

### **Client-Side**:
- âœ… Ad content sanitized
- âœ… Click tracking rate-limited
- âœ… Impression tracking debounced
- âœ… XSS prevention

### **Access Control**:
- âœ… Advertiser can only see own data
- âœ… Admin has full access
- âœ… Public can only see approved ads

---

## ðŸ“ **Known Issues**

### **To Implement**:
1. Real-time campaign updates
2. A/B test interface
3. Advanced targeting UI
4. Ad creative editor
5. Automated optimization suggestions

---

## ðŸš€ **Quick Links**

**Actual Files**:
- Advertise Page: `frontend/src/routes/advertise/+page.svelte`
- Advertiser Dashboard: `frontend/src/routes/advertiser/+page.svelte`
- Campaigns: `frontend/src/routes/advertiser/campaigns/+page.svelte`
- AdDisplay Component: `frontend/src/lib/components/AdDisplay.svelte`
- Advertiser Store: `frontend/src/lib/stores/advertiser.ts`

---

**Last Updated**: October 14, 2025  
**Status**: Core structure defined  
**Technology**: Svelte 5 + TypeScript  
**Backend Counterpart**: `_braids/advertisement-system/backend/`

---

**Navigate**:  
[ðŸ  Master Index](../../../BRAIDS_INDEX.md) | [â¬…ï¸ Backend Braid](../../_braids/advertisement-system/backend/BRAID.md)



---

## Integration Notes

- Frontend: `_braids/advertisement-system/frontend/`
- Backend: `_braids/advertisement-system/backend/`

This braid represents a complete vertical slice of functionality.

