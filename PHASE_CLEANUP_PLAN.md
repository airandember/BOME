# 🧹 Phase: Data Cleanup & User Consolidation

**Date:** October 31, 2025  
**Goal:** Clean up 29 users with duplicate customers/subscriptions  
**Status:** Ready to execute

---

## 📊 Current State

From our audit (`subscription-audit-report.json`):

| Issue Type | Count | Action Required |
|------------|-------|-----------------|
| 🔴 Critical (both issues) | 2 | Manual review + consolidation |
| 🟡 Multiple Active Subs | 17 | Contact users, cancel duplicates |
| 🟡 Multiple Customers | 10 | Archive unused customer IDs |
| **TOTAL** | **29** | **All need cleanup** |

---

## 🎯 Cleanup Strategy

### **Phase 1: Communication**
Contact affected users and explain the situation.

### **Phase 2: User Consolidation**
Work through each user systematically.

### **Phase 3: Verification**
Run audit tool again to confirm all issues resolved.

---

## 📧 Phase 1: User Communication

### **Email Template for Multiple Subscriptions:**

```
Subject: Important: Multiple Active Subscriptions Detected

Hi [Name],

We've detected that your account has [X] active subscriptions to Book of Mormon Evidence. This means you may be charged multiple times.

Your subscriptions:
- Subscription 1: [Plan Name] - $[Price]/[interval]
- Subscription 2: [Plan Name] - $[Price]/[interval]

We'd like to consolidate these to a single subscription of your choice. 

Which subscription would you prefer to keep?

Please reply to this email or contact us at [support_email].

Thank you,
BOME Team
```

### **For Critical Users (2):**
Priority email + phone call (if available)

---

## 🔧 Phase 2: Manual Consolidation Steps

### **For Each User:**

#### **Step 1: Identify Primary Subscription**
```bash
# Run audit tool for specific user
cd backend/cmd/subscription-audit
./subscription-audit.exe

# Or query directly:
SELECT 
    ss.stripe_id,
    sp.name,
    ss.status,
    ss.current_period_start,
    ss.current_period_end
FROM stripe_subscriptions_v2 ss
JOIN stripe_customers_v2 sc ON ss.customer_id = sc.id
JOIN user_stripe_customers_v2 usc ON sc.id = usc.stripe_customer_id
LEFT JOIN stripe_prices_v2 spr ON ss.price_id = spr.id
LEFT JOIN stripe_products_v2 sp ON spr.product_id = sp.id
WHERE usc.user_id = [USER_ID]
AND ss.status IN ('active', 'trialing')
ORDER BY ss.stripe_created_at DESC;
```

#### **Step 2: Cancel Duplicate Subscriptions (Stripe Dashboard)**

1. Go to Stripe Dashboard → Customers
2. Search for customer email
3. Click on customer
4. View subscriptions
5. For each duplicate subscription:
   - Click "Cancel subscription"
   - Choose "Cancel immediately" (don't wait for period end)
   - Add cancellation note: "Duplicate subscription - user consolidated to [primary_sub_id]"

#### **Step 3: Archive Duplicate Customer IDs (If Applicable)**

1. In Stripe Dashboard → Customers
2. Find customer with no active subscriptions
3. Click "⋮" menu → "Archive customer"
4. Add note: "Duplicate customer - user consolidated to [primary_cus_id]"

#### **Step 4: Update Database (After Stripe Changes)**

```bash
# Run Simple Sync to update v2 tables
curl http://localhost:8080/api/v1/admin/stripe-sync-v2/sync \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -X POST

# Or use admin UI: /admin/streaming/stripe → "Simple Sync"
```

#### **Step 5: Verify User**

```bash
# Check user's subscription status
curl http://localhost:8080/api/v1/admin/subscriber-elastic-v2/subscriber/[USER_ID] \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Should show:
# - has_active_plan: true
# - Single subscription
# - Single customer
```

---

## 📋 Cleanup Checklist

### **🔴 Critical Priority (2 users)**

- [ ] **jameskersey2@gmail.com** (User 4891)
  - [ ] Email sent
  - [ ] User responded with preference
  - [ ] Canceled 2 duplicate subscriptions
  - [ ] Archived 2 duplicate customers
  - [ ] Simple Sync completed
  - [ ] Verified in admin dashboard

- [ ] **pdm1441@gmail.com** (User 4881)
  - [ ] Email sent
  - [ ] User responded with preference
  - [ ] Canceled 1 duplicate subscription
  - [ ] Archived 1 duplicate customer
  - [ ] Simple Sync completed
  - [ ] Verified in admin dashboard

### **🟡 High Priority (17 users with multiple active subs)**

Track progress in CSV: `subscription-audit-report.csv`

- [ ] kjoelwa@me.com (3 active subs)
- [ ] benheaton1@gmail.com (2 active subs)
- [ ] clay.inspire@gmail.com (2 active subs)
- [ ] dbarger11@cox.net (2 active subs)
- [ ] emersonrowley@hotmail.com (2 active subs)
- [ ] floydwgowans@gmail.com (2 active subs)
- [ ] hushpuppi2001@yahoo.com (2 active subs)
- [ ] jam777jam777@netscape.net (2 active subs)
- [ ] james.hewitt.329@gmail.com (2 active subs)
- [ ] lorisessentialoils@gmail.com (2 active subs)
- [ ] lry@ebbe-america.com (2 active subs)
- [ ] lwinkelkotter@gmail.com (2 active subs)
- [ ] lyman.stevens@comcast.net (2 active subs)
- [ ] maryidabush@gmail.com (2 active subs)
- [ ] mike.armatage@gmail.com (2 active subs)
- [ ] pthooah@gmail.com (2 active subs)
- [ ] steveevans@outlook.com (2 active subs)

### **🟡 Medium Priority (10 users with multiple customers only)**

- [ ] ericgessel@gmail.com (2 customers, 1 active sub)
- [ ] gay.martin@gmail.com (2 customers, 1 active sub)
- [ ] joyfullavatar@gmail.com (2 customers, 1 active sub)
- [ ] robberch@gmail.com (2 customers, 1 active sub)
- [ ] dbates62@hotmail.com (2 customers, 0 active - unpaid)
- [ ] garrettreichert@hotmail.com (2 customers, 0 active)
- [ ] jillypill1@yahoo.com (2 customers, 0 active - past_due)
- [ ] lbar3351@gmail.com (2 customers, 0 active)
- [ ] shauna_math@outlook.com (2 customers, 0 active)
- [ ] sherryjohns@hotmail.com (2 customers, 0 active)

---

## 🤖 Semi-Automated Approach (Optional)

### **Create Admin Tool for Bulk Operations:**

Could create an admin UI page `/admin/subscription-cleanup` that:

1. Shows list of users with issues
2. For each user, displays:
   - All customers
   - All subscriptions
   - Radio buttons to select "primary" ones
3. "Consolidate" button that:
   - Calls Stripe API to cancel non-primary subscriptions
   - Archives non-primary customers
   - Updates database
   - Logs all actions

**Pros:**
- Faster cleanup
- Consistent process
- Audit trail

**Cons:**
- Requires development time
- Need to be careful with Stripe API rate limits
- Should still email users first

---

## 🎯 Success Criteria

After cleanup is complete:

1. **Run audit tool again:**
   ```bash
   cd backend/cmd/subscription-audit
   ./subscription-audit.exe
   ```

2. **Expected result:**
   ```
   📊 SUMMARY:
   Total Users with Issues: 0
   ```

3. **Verify in admin dashboard:**
   - All users show single customer
   - All users show single active subscription (or none)
   - No duplicate entries in admin table

4. **Financial verification:**
   - No users being charged multiple times
   - All active subscriptions match user expectations

---

## 💰 Financial Impact

### **Before Cleanup:**
- 19 users being overcharged (multiple active subs)
- Estimated ~$210/month extra charges
- Estimated ~$2,520/year extra charges

### **After Cleanup:**
- 0 users overcharged
- All charges accurate
- Better customer trust

---

## 🚀 Post-Cleanup Actions

1. **Email all affected users:**
   - Confirm consolidation complete
   - Show new single subscription details
   - Apologize for any confusion
   - Offer refund for duplicate charges (if applicable)

2. **Monitor for new issues:**
   - Run audit tool weekly
   - Should see 0 issues (system prevents new duplicates)

3. **Documentation:**
   - Update this checklist as you go
   - Keep email templates for future use
   - Document any edge cases encountered

---

## 📞 Support Information

**For users with questions:**
- Email: support@bookofmormonevidence.org
- Include their email address
- Include their preferred subscription details

**For admin assistance:**
- Use admin dashboard: `/admin/streaming/stripe`
- Use audit tool: `backend/cmd/subscription-audit/`
- Use Simple Sync to refresh data after Stripe changes

---

## 🎯 Recommended Timeline

| Task | Timeline | Effort |
|------|----------|--------|
| Email critical users (2) | Day 1 | 30 min |
| Wait for responses | Day 1-2 | - |
| Consolidate critical users | Day 2 | 1 hour |
| Email high priority users (17) | Day 2-3 | 2 hours |
| Wait for responses | Day 3-5 | - |
| Consolidate high priority | Day 5-7 | 3 hours |
| Archive duplicate customers (10) | Day 7 | 1 hour |
| Final verification | Day 8 | 1 hour |
| Follow-up emails | Day 8 | 1 hour |

**Total Active Time:** ~9 hours  
**Total Calendar Time:** ~8 days (with user response wait times)

---

**Ready to start?** Begin with the 2 critical users, then work through high priority, then medium priority. Use the CSV file to track progress!

**Tool:** `backend/cmd/subscription-audit/subscription-audit.exe`  
**CSV:** `backend/cmd/subscription-audit/subscription-audit-report.csv`  
**JSON:** `backend/cmd/subscription-audit/subscription-audit-report.json`

