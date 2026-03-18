# Braid: analytics-reporting

**Architecture:** Full-Stack Braid (Frontend to Backend)
**Last Updated:** 2025-10-17

---

## Backend Architecture

**Data analytics, metrics, and business intelligence**

---

## ðŸ“‹ **Backend Overview**

**Purpose**: Collect, process, and serve analytics data for business intelligence  
**Technology**: Go, PostgreSQL, Time-series data  
**Complexity**: High (Data Processing, Aggregation, Real-time Metrics)  

---

## **File Map** (Production Code)

| Layer | Production Path | Description |
|-------|-----------------|-------------|
| Services | `backend/internal/services/analytics.go` | Core analytics logic |
| Services | `backend/internal/services/business_intelligence.go` | BI metrics |
| Services | `backend/analytics/services/analytics.go` | Analytics service |
| Handlers | `backend/analytics/handlers/analytics.go` | Analytics API handlers |
| Routes | `backend/internal/routes/analytics.go`, `unified_analytics.go` | Analytics endpoints |
| Database | `backend/internal/database/` | Analytics tables, metrics |

**Frontend:** `frontend/src/routes/analytics/`

---

## **Key Analytics Domains**

### **1. User Analytics**:
- Active users (DAU, WAU, MAU)
- User growth trends
- Registration funnel
- User retention cohorts
- Geographic distribution
- Device/browser breakdown

### **2. Video Analytics**:
- Video views and watch time
- Completion rates
- Engagement metrics
- Popular content
- Content performance trends

### **3. Subscription Analytics**:
- MRR (Monthly Recurring Revenue)
- Churn rate
- Lifetime value (LTV)
- Conversion rates
- Plan distribution

### **4. Revenue Analytics**:
- Total revenue trends
- Revenue by source (subscriptions, ads)
- ARPU (Average Revenue Per User)
- Revenue forecasting

### **5. System Performance**:
- API response times
- Error rates
- Uptime metrics
- Database performance

---

## ðŸ—„ï¸ **Database Schema**

### **Analytics Events Table**:
```sql
CREATE TABLE analytics_events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(255),
    properties JSONB,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX idx_analytics_events_user_id ON analytics_events(user_id);
CREATE INDEX idx_analytics_events_created_at ON analytics_events(created_at);

-- Partition by month
CREATE TABLE analytics_events_2025_10 PARTITION OF analytics_events
FOR VALUES FROM ('2025-10-01') TO ('2025-11-01');
```

### **Daily Metrics Table** (Aggregated):
```sql
CREATE TABLE daily_metrics (
    id SERIAL PRIMARY KEY,
    date DATE NOT NULL UNIQUE,
    active_users INTEGER DEFAULT 0,
    new_users INTEGER DEFAULT 0,
    total_videos_watched INTEGER DEFAULT 0,
    total_watch_time_seconds BIGINT DEFAULT 0,
    revenue DECIMAL(10, 2) DEFAULT 0,
    new_subscriptions INTEGER DEFAULT 0,
    canceled_subscriptions INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_daily_metrics_date ON daily_metrics(date);
```

---

## ðŸŒ **API Endpoints**

### **Dashboard Analytics**:
```
GET /api/v1/analytics/dashboard          # Overview metrics
GET /api/v1/analytics/users              # User analytics
GET /api/v1/analytics/videos             # Video analytics
GET /api/v1/analytics/subscriptions      # Subscription analytics
GET /api/v1/analytics/revenue            # Revenue analytics
```

### **Unified Analytics**:
**File**: `backend/internal/routes/unified_analytics.go`
```
GET /api/v1/unified-analytics            # Combined metrics
```

### **Stripe Analytics**:
**File**: `backend/internal/routes/stripe_analytics_routes.go`
```
GET /api/v1/stripe-analytics             # Stripe-specific metrics
```

---

## ðŸ”§ **Backend Services**

### **Analytics Service** (`backend/internal/services/analytics.go`):
```go
func TrackEvent(eventType string, userID *int, properties map[string]interface{}) error
func GetDashboardMetrics(dateRange DateRange) (*DashboardMetrics, error)
func GetUserAnalytics(dateRange DateRange) (*UserAnalytics, error)
func GetVideoAnalytics(dateRange DateRange) (*VideoAnalytics, error)
func AggregateDaily Metrics(date time.Time) error
```

---

## ðŸ“Š **Key Metrics**

### **User Metrics**:
- **DAU** (Daily Active Users)
- **MAU** (Monthly Active Users)
- **Retention Rate**: % of users who return
- **Churn Rate**: % of users who leave

### **Revenue Metrics**:
- **MRR** (Monthly Recurring Revenue)
- **ARPU** (Average Revenue Per User) = Revenue / Total Users
- **LTV** (Lifetime Value) = ARPU Ã— Average Customer Lifetime
- **CAC** (Customer Acquisition Cost)

### **Content Metrics**:
- **Average Watch Time**
- **Completion Rate**: % of video watched
- **Engagement Rate**: Likes + Comments / Views

---

**Last Updated**: October 14, 2025  
**Frontend**: `_frontend/braids/analytics-reporting/`



---

## Frontend Architecture

**Data visualization and business intelligence dashboards**

---

## ðŸ“‹ **Frontend Overview**

**Purpose**: Visualize analytics data for business insights  
**Technology**: Svelte 5, Chart.js/D3.js, TypeScript  
**Entry Points**: `/admin/analytics`, `/admin/streaming/analytics`  

---

## ðŸŽ¯ **Key Features**

### **1. Analytics Dashboard**:
- KPI cards (DAU, MAU, Revenue, Churn)
- Trend charts (line, bar, pie)
- Date range picker
- Export reports (CSV, PDF)
- Real-time metrics

### **2. User Analytics**:
- User growth chart
- Retention cohorts
- Geographic heatmap
- Device/browser breakdown

### **3. Video Analytics**:
- Popular videos table
- Watch time trends
- Completion rates
- Engagement metrics

### **4. Revenue Analytics**:
- MRR chart
- Revenue breakdown
- Subscription trends
- Churn analysis

---

## ðŸ“„ **Frontend Pages**

### **Analytics Dashboard** (`/admin/analytics`)
**File**: `frontend/src/routes/admin/analytics/+page.svelte`

```svelte
<script>
  import MetricCard from '$lib/components/MetricCard.svelte';
  import LineChart from '$lib/components/LineChart.svelte';
  
  let metrics = {
    dau: 1250,
    mau: 12500,
    revenue: 45000,
    churn: 2.3
  };
</script>

<div class="analytics-dashboard">
  <h1>Analytics Dashboard</h1>
  
  <div class="metrics-grid">
    <MetricCard title="DAU" value={metrics.dau} change={+12} />
    <MetricCard title="MAU" value={metrics.mau} change={+8} />
    <MetricCard title="Revenue" value={'$' + metrics.revenue} change={+15} />
    <MetricCard title="Churn" value={metrics.churn + '%'} change={-1.2} />
  </div>
  
  <LineChart title="User Growth" data={userGrowthData} />
  <LineChart title="Revenue Trend" data={revenueData} />
</div>
```

---

## ðŸ§© **Key Components**

### **MetricCard**:
```svelte
<script>
  export let title;
  export let value;
  export let change; // percentage change
</script>

<div class="metric-card">
  <h3>{title}</h3>
  <div class="value">{value}</div>
  <div class="change" class:positive={change > 0}>
    {change > 0 ? 'â†‘' : 'â†“'} {Math.abs(change)}%
  </div>
</div>
```

---

**Last Updated**: October 14, 2025  
**Backend**: `_braids/analytics-reporting/backend/`



---

## Integration Notes

- Frontend: `_braids/analytics-reporting/frontend/`
- Backend: `_braids/analytics-reporting/backend/`

This braid represents a complete vertical slice of functionality.

