# STRAND: Ad Serving

**Purpose**: Display advertisements to users based on placement and targeting rules.

---

## Implementation Details

### Backend
- **Handlers**: `backend/advertisement/handlers/advertisement.go`
- **Services**: `backend/advertisement/services/advertisement.go`
- **Models**: `backend/advertisement/models/advertisement.go`
- **Routes**: `backend/internal/routes/advertisement.go`
- **Database**: `advertisements`, `ad_campaigns`, `ad_impressions` tables

### Frontend
- **Pages**: `frontend/src/routes/advertise/`
- **Components**: `AdDisplay.svelte`

### Flow
1. Frontend requests ad slot (placement, targeting params)
2. Backend selects eligible ad from campaign
3. Impressions tracked
4. Ad creative returned for display

---

## Status
- [x] Backend Complete
- [x] Frontend Ad Display
- [ ] Full targeting rules documented
- [ ] A/B testing strand

---

## Testing
- Verify ad display on video pages
- Check impression tracking in database
- Validate placement targeting
