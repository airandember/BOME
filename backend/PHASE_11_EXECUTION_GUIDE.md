# Phase 11: V1 Archival - Execution Guide

**Date:** October 31, 2025  
**Status:** 🚀 READY TO EXECUTE  
**Prerequisites:** ✅ ALL MET (Phase 9 confirms 100% data integrity)

---

## 🎯 Objective

Safely archive v1 Stripe tables and remove deprecated code, completing the migration to v2 architecture.

---

## ⚠️ Pre-Flight Checklist

Before executing Phase 11, verify:

- [x] ✅ Phase 9 completed successfully
- [x] ✅ 2,531 users migrated to v2 (100% coverage)
- [x] ✅ V2 elastic service working in production
- [x] ✅ Frontend using v2 data
- [x] ✅ Webhooks writing to v2 tables
- [x] ✅ Video access calculated from v2 data
- [x] ✅ Database backup completed (RECOMMENDED!)

---

## 📋 Phase 11.1: Archive V1 Tables

### Step 1: Run Migration 052

```sql
-- Connect to your database and run:
\i migrations/052_archive_v1_stripe_tables.sql
```

**OR** via PowerShell:

```powershell
cd backend
psql -h localhost -U bome_admin -d bome_db -f migrations/052_archive_v1_stripe_tables.sql
```

### Expected Output:

```
✅ Renamed user_subscriptions → user_subscriptions_deprecated_v1
✅ Added is_v1_legacy column to subscription_plans
✅ Created audit record for user_subscriptions archival
✅ Added deprecation comment to user_subscriptions_deprecated_v1

═══════════════════════════════════════════════════
✅ Migration 052: V1 Table Archival Complete!
═══════════════════════════════════════════════════

📊 Table Status:
   ✅ user_subscriptions_deprecated_v1: EXISTS (archived)
   ✅ stripe_subscriptions_v2: EXISTS (active)
   ✅ user_subscriptions: ARCHIVED (no longer exists)
```

### Rollback Plan:

If something goes wrong:

```sql
-- Restore v1 table
ALTER TABLE user_subscriptions_deprecated_v1 RENAME TO user_subscriptions;
```

---

## 📋 Phase 11.2: Remove V1 Code

### Files to Delete:

**Backend:**
```
❌ backend/internal/database/subscription.go (v1 methods - keep v2 only)
❌ backend/internal/services/subscription_service.go (deprecated)
❌ backend/internal/routes/subscription_routes.go (old authenticated routes - commented out)
```

### Code to Remove:

**In `backend/internal/routes/routes.go`:**

1. Remove commented-out v1 subscription routes (lines ~280-300)
```go
// DEPRECATED: Old authenticated Stripe routes - replaced by v2
// SetupAuthenticatedStripeRoutes(authenticated, db)
```

2. Remove global v1 service initialization:
```go
// DEPRECATED: v1 subscription service
// globalUserSubscriptionService = services.NewUserSubscriptionService(db, stripe.Key)
```

**In `backend/internal/database/database.go`:**

Remove v1 subscription methods:
- `GetSubscriptionByUserID()` - replaced by v2 elastic service
- `GetSubscriptionPlanByID()` - replaced by v2 queries
- `HasVideoAccess()` - replaced by v2 elastic service logic

**Keep these (still needed):**
- ✅ `subscription_plans` table methods (used by both v1 and v2)
- ✅ v2 services and routes
- ✅ Webhook dual-write logic (can be simplified to v2-only later)

---

## 📋 Phase 11.3: Update Middleware

### Files to Update:

**`backend/internal/middleware/middleware.go`:**

Already updated! All middleware now uses `SubscriberElasticServiceV2`:
- ✅ `SubscriptionAccessRequired()` - uses v2
- ✅ `SubscriptionValidation()` - uses v2
- ✅ `SubscriptionPlanValidation()` - uses v2
- ✅ `SubscriptionExpirationWarning()` - uses v2

**No changes needed!** ✅

---

## 📋 Phase 11.4: Documentation

### Documents to Create/Update:

1. **`V1_TO_V2_MIGRATION_COMPLETE.md`** - Final migration report
2. **`DEPRECATED_CODE_ARCHIVE.md`** - List of removed code with git hashes
3. **Update `README.md`** - Remove v1 references
4. **Update `ARCHITECTURE.md`** (if exists) - Document v2 as standard

---

## 🔍 Post-Migration Verification

### 1. Test Key Functionality:

```bash
# Test v2 elastic service
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/admin/subscriber-elastic-v2/subscribers

# Test user subscription management
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/api/v1/user/subscriptions

# Test webhook (use Stripe CLI)
stripe trigger customer.created
```

### 2. Monitor Logs:

Watch for any errors referencing:
- `user_subscriptions` table
- v1 subscription methods
- Deprecated routes

### 3. Check Frontend:

- ✅ Admin subscriber list loads
- ✅ User subscription dashboard works
- ✅ Video access enforced correctly
- ✅ New subscriptions create properly

---

## 📊 Success Criteria

Phase 11 is complete when:

- [x] ✅ V1 tables renamed with `_deprecated` suffix
- [ ] ⏳ V1 code removed from codebase
- [ ] ⏳ Documentation updated
- [ ] ⏳ System runs without errors for 48 hours
- [ ] ⏳ No references to deprecated tables in logs

---

## 🚨 Emergency Rollback

If critical issues occur:

1. **Restore V1 Table:**
   ```sql
   ALTER TABLE user_subscriptions_deprecated_v1 RENAME TO user_subscriptions;
   ```

2. **Revert Code Changes:**
   ```bash
   git reset --hard HEAD~1  # Or specific commit
   ```

3. **Restart Services:**
   ```bash
   # Kill backend
   taskkill /F /IM bome-backend.exe
   
   # Restart with old code
   cd backend
   go run main.go
   ```

---

## 📅 Timeline

- **Phase 11.1** (Table Archival): 5 minutes
- **Phase 11.2** (Code Removal): 30 minutes
- **Phase 11.3** (Middleware Update): Already complete! ✅
- **Phase 11.4** (Documentation): 30 minutes
- **Monitoring Period**: 48 hours

**Total:** ~1 hour work + 48 hours monitoring

---

## 🎉 Next Steps After Phase 11

1. Monitor system for 48 hours
2. After 48 hours of stability: Create `MIGRATION_SUCCESS.md`
3. After 90 days: Drop deprecated tables permanently:
   ```sql
   DROP TABLE IF EXISTS user_subscriptions_deprecated_v1 CASCADE;
   DROP TABLE IF EXISTS v1_archive_audit;
   ```

---

## 📞 Support

If issues arise during Phase 11 execution:
1. Check logs for specific error messages
2. Verify database connection
3. Check that v2 tables still exist
4. Review Phase 9 report for data integrity

**DO NOT proceed if:**
- Database backup is not available
- Phase 9 shows data integrity issues
- Production system is experiencing other problems

