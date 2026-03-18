# STRAND: Metrics Dashboard

**Purpose**: Aggregate and serve analytics data for business intelligence dashboards.

---

## Implementation Details

### Backend
- **Services**: `backend/internal/services/analytics.go`, `business_intelligence.go`, `subscription_analytics.go`
- **Handlers**: `backend/analytics/handlers/analytics.go`
- **Routes**: `backend/internal/routes/analytics.go`, `unified_analytics.go`
- **Models**: `backend/analytics/models/analytics.go`
- **Database**: `analytics_events`, `daily_metrics` (or equivalent)

### Frontend
- **Pages**: `frontend/src/routes/analytics/`

### Flow
1. Client requests analytics (date range, metrics type)
2. Backend aggregates from events and pre-computed tables
3. Data returned as time-series or summary
4. Dashboard renders charts

---

## Status
- [x] Backend services implemented
- [x] Analytics routes exposed
- [ ] Streaming analytics integration documented
- [ ] Export strand

---

## Testing
- Hit analytics API endpoints with date range
- Verify aggregation correctness
- Check dashboard page loads
