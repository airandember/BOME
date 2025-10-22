# 🧬 Advertisement System Braid - Frontend
**Svelte5 UI for advertisers, campaigns, and ad management**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **frontend portion** of the Advertisement System Braid.  
> **Backend portion**: See `_backend/braids/advertisement-system/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Frontend Overview**

**Purpose**: User interface for advertisers to manage campaigns and view analytics  
**Technology**: Svelte 5, TypeScript, TailwindCSS  
**Entry Points**: `/advertise`, `/advertiser`, `/admin/advertisements`  
**State Management**: Svelte stores for campaign and analytics data

---

## 🎯 **Key Features**

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

## 📄 **Frontend Pages**

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
      <h3>🎯 Targeted Reach</h3>
      <p>Advanced targeting options</p>
    </div>
    <div class="benefit">
      <h3>📊 Real-time Analytics</h3>
      <p>Track performance instantly</p>
    </div>
    <div class="benefit">
      <h3>💰 Flexible Pricing</h3>
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

## 🧩 **Frontend Components**

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

## 🗃️ **Frontend Stores**

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

## 🔄 **Data Flow Examples**

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

## 📊 **Analytics Visualization**

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

## 🔒 **Security**

### **Client-Side**:
- ✅ Ad content sanitized
- ✅ Click tracking rate-limited
- ✅ Impression tracking debounced
- ✅ XSS prevention

### **Access Control**:
- ✅ Advertiser can only see own data
- ✅ Admin has full access
- ✅ Public can only see approved ads

---

## 📝 **Known Issues**

### **To Implement**:
1. Real-time campaign updates
2. A/B test interface
3. Advanced targeting UI
4. Ad creative editor
5. Automated optimization suggestions

---

## 🚀 **Quick Links**

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
**Backend Counterpart**: `_backend/braids/advertisement-system/`

---

**Navigate**:  
[🏠 Master Index](../../../BRAIDS_INDEX.md) | [⬅️ Backend Braid](../../_backend/braids/advertisement-system/BRAID.md)

