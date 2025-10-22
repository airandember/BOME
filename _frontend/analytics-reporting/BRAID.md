# 🧬 Analytics & Reporting Braid - Frontend  
**Data visualization and business intelligence dashboards**

---

## 📋 **Frontend Overview**

**Purpose**: Visualize analytics data for business insights  
**Technology**: Svelte 5, Chart.js/D3.js, TypeScript  
**Entry Points**: `/admin/analytics`, `/admin/streaming/analytics`  

---

## 🎯 **Key Features**

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

## 📄 **Frontend Pages**

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

## 🧩 **Key Components**

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
    {change > 0 ? '↑' : '↓'} {Math.abs(change)}%
  </div>
</div>
```

---

**Last Updated**: October 14, 2025  
**Backend**: `_backend/braids/analytics-reporting/`

