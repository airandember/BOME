## 🚨 **CRITICAL: Missing Video Access Columns**

###Error:
```
⚠️ [Webhook v2] Failed to update video access for subscription: 
pq: column "video_access_source" does not exist
```

### **Problem:**
The `users` table is missing the video access tracking columns that the code expects:
- `manual_video_access` (BOOLEAN)
- `video_access_granted_at` (TIMESTAMPTZ)
- `video_access_source` (VARCHAR)

---

## ✅ **Solution: Apply Migration**

### **Step 1: Apply the Migration**

Run this SQL migration on your database:

```bash
psql -U postgres -d bome_production -f S:\AirEmber\BOME\BOME\backend\migrations\060_add_video_access_tracking.sql
```

**OR** run this SQL directly in your database:

```sql
-- Add manual_video_access column (if it doesn't exist)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS manual_video_access BOOLEAN DEFAULT FALSE;

-- Add video_access_granted_at column (if it doesn't exist)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS video_access_granted_at TIMESTAMPTZ;

-- Add video_access_source column (if it doesn't exist)
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS video_access_source VARCHAR(255);

-- Create index for fast video access lookups
CREATE INDEX IF NOT EXISTS idx_users_manual_video_access 
    ON users(manual_video_access) 
    WHERE manual_video_access = true;

-- Add comments
COMMENT ON COLUMN users.manual_video_access IS 'Manual override for video access (admin granted or subscription-based)';
COMMENT ON COLUMN users.video_access_granted_at IS 'Timestamp when video access was granted';
COMMENT ON COLUMN users.video_access_source IS 'Source of video access grant (e.g., session_verification, webhook, retroactive_linking)';
```

### **Step 2: Restart Backend**

After applying the migration, restart your backend server.

---

## 🎯 **What These Columns Do:**

| Column | Type | Purpose |
|--------|------|---------|
| `manual_video_access` | BOOLEAN | TRUE when user has video access (from subscription or admin grant) |
| `video_access_granted_at` | TIMESTAMPTZ | Timestamp when access was first granted |
| `video_access_source` | VARCHAR(255) | Tracks HOW they got access (e.g., `session_verification:cs_xxx`, `webhook:sub_xxx`, `retroactive_linking`) |

---

## 📊 **How It Works:**

### **When a user subscribes:**
1. Webhook fires: `customer.subscription.created`
2. Code tries to grant video access
3. Updates `users` table:
   ```sql
   UPDATE users 
   SET manual_video_access = true,
       video_access_granted_at = NOW(),
       video_access_source = 'webhook:sub_1SW2lI...'
   WHERE id = 10469
   ```

### **Without these columns:**
❌ SQL error: `column "video_access_source" does not exist`  
❌ User can't access videos despite having active subscription

### **With these columns:**
✅ Video access granted immediately  
✅ Tracking of when and how they got access  
✅ User can watch premium content

---

## 🔍 **Verify It Worked:**

After applying the migration, check the user's access:

```sql
SELECT id, email, manual_video_access, video_access_granted_at, video_access_source
FROM users 
WHERE email = 'bometesting@gmail.com';
```

Should show:
```
id    | email                  | manual_video_access | video_access_granted_at | video_access_source
------|------------------------|---------------------|-------------------------|--------------------
10469 | bometesting@gmail.com  | false               | NULL                    | NULL
```

Then trigger a webhook or session verification, and it should update to:
```
id    | email                  | manual_video_access | video_access_granted_at      | video_access_source
------|------------------------|---------------------|------------------------------|--------------------
10469 | bometesting@gmail.com  | true                | 2025-11-21 22:30:15.123456  | webhook:sub_1SW2lI...
```

---

## 🚨 **This is why the user can't access videos!**

The subscription is **active** in Stripe, but the backend can't update the `users` table because the columns don't exist.

**Apply this migration NOW to fix the issue!** 🚀

