# 🧬 Analytics & Reporting Braid - Backend  
**Data analytics, metrics, and business intelligence**

---

## 📋 **Backend Overview**

**Purpose**: Collect, process, and serve analytics data for business intelligence  
**Technology**: Go, PostgreSQL, Time-series data  
**Complexity**: High (Data Processing, Aggregation, Real-time Metrics)  

**Critical Files**:
- `backend/internal/services/analytics.go`
- `backend/internal/services/business_intelligence.go`
- `backend/internal/services/subscription_analytics.go`
- `backend/internal/routes/analytics.go`
- `backend/internal/database/analytics.go`

---

## 🎯 **Key Analytics Domains**

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

## 🗄️ **Database Schema**

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

## 🌐 **API Endpoints**

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

## 🔧 **Backend Services**

### **Analytics Service** (`backend/internal/services/analytics.go`):
```go
func TrackEvent(eventType string, userID *int, properties map[string]interface{}) error
func GetDashboardMetrics(dateRange DateRange) (*DashboardMetrics, error)
func GetUserAnalytics(dateRange DateRange) (*UserAnalytics, error)
func GetVideoAnalytics(dateRange DateRange) (*VideoAnalytics, error)
func AggregateDaily Metrics(date time.Time) error
```

---

## 📊 **Key Metrics**

### **User Metrics**:
- **DAU** (Daily Active Users)
- **MAU** (Monthly Active Users)
- **Retention Rate**: % of users who return
- **Churn Rate**: % of users who leave

### **Revenue Metrics**:
- **MRR** (Monthly Recurring Revenue)
- **ARPU** (Average Revenue Per User) = Revenue / Total Users
- **LTV** (Lifetime Value) = ARPU × Average Customer Lifetime
- **CAC** (Customer Acquisition Cost)

### **Content Metrics**:
- **Average Watch Time**
- **Completion Rate**: % of video watched
- **Engagement Rate**: Likes + Comments / Views

---

**Last Updated**: October 14, 2025  
**Frontend**: `_frontend/braids/analytics-reporting/`

