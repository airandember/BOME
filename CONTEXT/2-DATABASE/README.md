# 🗄️ **DATABASE DOCUMENTATION**

**Purpose:** Complete database schema, conventions, and migration references

---

## 📚 **DOCUMENTS IN THIS FOLDER**

### **⭐ MAIN REFERENCE**
- **`DATABASE_SCHEMA.md`** - Complete database documentation
  - 74 tables fully documented
  - All columns, indexes, foreign keys
  - Usage notes for each table
  - Database statistics queries
  - Maintenance commands
  - Critical indexes
  - Naming conventions

---

## 📊 **WHAT'S DOCUMENTED**

### **Major Table Categories:**

1. **Authentication & Users** (5 tables)
   - users, sessions, oauth2_*

2. **Subscriptions** (8 tables)
   - subscription_plans, subscription_offers, subscribers, etc.

3. **Stripe Integration** (7 tables)
   - stripe_customers, stripe_subscriptions, stripe_products, etc.

4. **Video Streaming** (10 tables)
   - master_video_list, video_tags, video_categories, etc.

5. **YouTube Integration** (3 tables)
   - youtube_videos, youtube_config, youtube_sync_log

6. **Advertisement System** (12 tables)
   - advertisement_*, campaign_*, billing_*

7. **Email System** (5 tables)
   - email_templates, email_queue, email_logs, etc.

8. **Analytics** (8 tables)
   - user_activity, video_views, metrics_*

9. **Creator Payouts** (5 tables) ✨ NEW
   - presenters, video_presenters, payout_formulas, presenter_payouts, payout_transactions

10. **System Configuration** (11 tables)
    - public_settings, secure_settings, feature_flags, etc.

---

## 🎯 **WHEN TO USE**

### **Read `DATABASE_SCHEMA.md` when:**
- Creating new tables
- Adding columns to existing tables
- Writing SQL queries
- Understanding data relationships
- Planning migrations
- Optimizing queries
- Debugging database issues

### **Database Conventions:**
```sql
-- Tables: plural, snake_case
users, subscription_plans, master_video_list

-- Columns: snake_case
created_at, is_active, stripe_customer_id

-- Primary Keys: id (BIGSERIAL)
id

-- Foreign Keys: {table}_id
user_id, video_id, plan_id

-- Booleans: is_, has_, can_ prefix
is_active, has_subbed, can_access

-- Timestamps: _at suffix
created_at, updated_at, deleted_at

-- Indexes: idx_{table}_{column}
idx_users_email, idx_videos_bunny_id

-- Unique Constraints: {table}_{column}_key
users_email_key
```

---

## 🔗 **MIGRATION REFERENCES**

Actual migration SQL files are in `backend/migrations/`, but documented here:

### **Creator Payouts:**
- `creator_payout_tables.sql` - 5 new tables
- `creator_payout_functions.sql` - 5 PL/pgSQL functions
- `creator_payout_seed_data.sql` - 4 default formulas

### **YouTube:**
- `youtube_tables.sql` - Config and sync log tables

### **Stripe Ghosts:**
- `stripe_ghosts_table.sql` - Ghost detection table
- `stripe_ghost_functions.sql` - 4 detection functions

### **Indexes & Optimization:**
- Various index creation scripts referenced in docs

---

## 📈 **DATABASE STATS**

- **Total Tables:** 74
- **Total Indexes:** 150+
- **Foreign Keys:** 80+
- **PL/pgSQL Functions:** 10+
- **JSONB Columns:** 15+
- **Array Columns:** 20+

---

## 🔍 **QUICK REFERENCE**

### **Find a table:**
```sql
\dt table_name*
```

### **See table structure:**
```sql
\d table_name
```

### **List all indexes:**
```sql
\di table_name*
```

### **Check foreign keys:**
```sql
SELECT * FROM information_schema.table_constraints 
WHERE constraint_type = 'FOREIGN KEY';
```

### **See functions:**
```sql
\df function_name*
```

---

## 🔗 **RELATED DOCUMENTATION**

- **Architecture:** `../1-ARCHITECTURE/BOME_CONTEXT_STANDARD.md`
- **Migrations:** `../6-MIGRATIONS/` (migration guides)

---

**Location:** `CONTEXT/2-DATABASE/`  
**Files:** 1 comprehensive reference (DATABASE_SCHEMA.md)  
**Status:** Complete ✅  
**Coverage:** All 74 tables documented

