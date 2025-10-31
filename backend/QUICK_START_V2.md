# Quick Start: Populate V2 Tables

Your v2 tables are empty! Here's how to populate them:

## **Step 1: Run Stripe Sync (5-10 minutes)**

This syncs ALL Stripe data to v2 tables:

```bash
cd backend/cmd/stripe-sync
go run main.go
```

**Expected Output**:
```
📦 Syncing products...
✅ Synced 5 products

💰 Syncing prices...
✅ Synced 12 prices

👥 Syncing customers...
✅ Synced 150 customers

📋 Syncing subscriptions...
✅ Synced 145 subscriptions

🎉 Stripe sync complete!
```

---

## **Step 2: Link Users to Customers (1-2 minutes)**

This links users to their Stripe customers by email:

```bash
cd backend/cmd/customer-linking
go run main.go --link-all --pretty
```

**Expected Output**:
```
🔗 Linking all users to Stripe customers...
✅ Linked 148 users to 150 customers
📊 Statistics:
  - Users with customers: 148
  - Users without customers: 2
  - Total stripe customers: 150
  - Unlinked customers: 0
```

---

## **Step 3: Verify**

Check that user 4826 now has linked customers:

```bash
curl http://localhost:8080/api/v1/admin/customer-linking/user/4826 \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"
```

**Expected**:
```json
{
  "user_id": 4826,
  "email": "aaalifesolutions@gmail.com",
  "linked_customers": [
    {
      "stripe_customer_id": "cus_xxxxx",
      "is_primary": true,
      "subscriptions_count": 1
    }
  ]
}
```

---

## **Step 4: Test User Subscription Dashboard**

Now reload `/user/subscriptions` - it should show their subscription!

---

## **Why This Happened**

- ✅ V1 tables have data (old system working)
- ❌ V2 tables are empty (new system not populated yet)
- 🔄 Phase 9 is MIGRATION - we need to sync data from Stripe → V2

**Once v2 is populated, the user subscription dashboard will work!**

---

## **About the Repeated Auth Checks**

The video page is making multiple parallel API calls. This is a frontend optimization opportunity:

**Current**: Each video thumbnail makes its own auth check
**Better**: Batch the auth check or use a shared auth state

But this is a separate optimization - the auth checks are fast (< 10ms each) so not critical.

