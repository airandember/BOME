# 🔍 Ghost Products Investigation Checklist

**Date:** October 31, 2025  
**Purpose:** Verify remaining 5 "ghost" products to determine if they're real or actual ghosts  
**Status:** Ready for investigation

---

## ✅ **Already Verified (Removed from Blocklist):**

1. **`prod_HjYKGcWGP9r4EC`** ✅
   - **Status:** REAL historical product
   - **Name:** "Monthly"
   - **Used in:** 10 subscriptions (2020-2021)
   - **Action:** Placeholder added, removed from blocklist

2. **`prod_FvNAlEGGL452nN`** ✅
   - **Status:** REAL historical product  
   - **Name:** "Essential Month"
   - **Used in:** 2 subscriptions (2020-2021, including Murial Murff)
   - **Action:** Placeholder added, removed from blocklist

---

## 🔍 **Need Investigation (Still in Blocklist):**

### **Product 1: `prod_HEmcX1PE8TO2CO`**

**How to investigate:**
1. Go to Stripe Dashboard → Products
2. Search for: `prod_HEmcX1PE8TO2CO`
3. Fill in findings below:

**Findings:**
- [ ] Product exists in Stripe
- [ ] Product deleted/not found

**If exists, record:**
- Product name: `_________________`
- Date created: `_________________`
- Status: Active / Archived / Deleted
- Number of subscriptions: `_____`
- Date range of subscriptions: `_________________`

**Decision:**
- [ ] REAL - Add placeholder and remove from blocklist
- [ ] GHOST - Keep in blocklist

---

### **Product 2: `prod_FvNAeI348dup9w`**

**How to investigate:**
1. Go to Stripe Dashboard → Products
2. Search for: `prod_FvNAeI348dup9w`
3. Fill in findings below:

**Findings:**
- [ ] Product exists in Stripe
- [ ] Product deleted/not found

**If exists, record:**
- Product name: `_________________`
- Date created: `_________________`
- Status: Active / Archived / Deleted
- Number of subscriptions: `_____`
- Date range of subscriptions: `_________________`

**Decision:**
- [ ] REAL - Add placeholder and remove from blocklist
- [ ] GHOST - Keep in blocklist

---

### **Product 3: `prod_HF5YzcBH5Rwr0d`**

**How to investigate:**
1. Go to Stripe Dashboard → Products
2. Search for: `prod_HF5YzcBH5Rwr0d`
3. Fill in findings below:

**Findings:**
- [ ] Product exists in Stripe
- [ ] Product deleted/not found

**If exists, record:**
- Product name: `_________________`
- Date created: `_________________`
- Status: Active / Archived / Deleted
- Number of subscriptions: `_____`
- Date range of subscriptions: `_________________`

**Decision:**
- [ ] REAL - Add placeholder and remove from blocklist
- [ ] GHOST - Keep in blocklist

---

### **Product 4: `prod_GVV5efccnh13h9`**

**How to investigate:**
1. Go to Stripe Dashboard → Products
2. Search for: `prod_GVV5efccnh13h9`
3. Fill in findings below:

**Findings:**
- [ ] Product exists in Stripe
- [ ] Product deleted/not found

**If exists, record:**
- Product name: `_________________`
- Date created: `_________________`
- Status: Active / Archived / Deleted
- Number of subscriptions: `_____`
- Date range of subscriptions: `_________________`

**Decision:**
- [ ] REAL - Add placeholder and remove from blocklist
- [ ] GHOST - Keep in blocklist

---

### **Product 5: `prod_FvNAJgnw48hwpZ`**

**How to investigate:**
1. Go to Stripe Dashboard → Products
2. Search for: `prod_FvNAJgnw48hwpZ`
3. Fill in findings below:

**Findings:**
- [ ] Product exists in Stripe
- [ ] Product deleted/not found

**If exists, record:**
- Product name: `_________________`
- Date created: `_________________`
- Status: Active / Archived / Deleted
- Number of subscriptions: `_____`
- Date range of subscriptions: `_________________`

**Decision:**
- [ ] REAL - Add placeholder and remove from blocklist
- [ ] GHOST - Keep in blocklist

---

## 📋 **Quick Reference: How to Search in Stripe**

### **Method 1: Search Bar**
1. Log into Stripe Dashboard
2. Use search bar at top
3. Paste product ID (e.g., `prod_HEmcX1PE8TO2CO`)
4. Press Enter

### **Method 2: Products List**
1. Stripe Dashboard → Products
2. Click "All products"
3. Use filter or search box
4. Look for the product ID

### **Method 3: Direct URL**
```
https://dashboard.stripe.com/products/[PRODUCT_ID]
```
Example: `https://dashboard.stripe.com/products/prod_HEmcX1PE8TO2CO`

---

## 🎯 **After Investigation:**

### **For Each REAL Product Found:**

1. **Add SQL placeholder:**
```sql
INSERT INTO stripe_products_v2 
(stripe_id, name, description, active, video_approved, stripe_created_at, stripe_updated_at)
VALUES
('[PRODUCT_ID]', '[PRODUCT_NAME] (Legacy)', 'Historical product from [YEAR] - discontinued', false, false, '[DATE]', NOW());
```

2. **Remove from blocklist** in these files:
   - `backend/internal/services/stripe_sync.go`
   - `backend/services/payment/stripe/stripe_sync.go`
   - `backend/subscription/services/stripe_sync.go`

3. **Re-run Stripe Sync** via admin dashboard

### **For Each TRUE GHOST (Not Found):**

1. **Keep in blocklist** - it's protecting you from bad data
2. **Document why** it's blocked (corrupted, test data, etc.)

---

## 📊 **Investigation Summary Template**

After completing investigation, fill this out:

| Product ID | Exists? | Name | Subs | Decision |
|------------|---------|------|------|----------|
| prod_HEmcX1PE8TO2CO | ☐ Yes ☐ No | | | ☐ Real ☐ Ghost |
| prod_FvNAeI348dup9w | ☐ Yes ☐ No | | | ☐ Real ☐ Ghost |
| prod_HF5YzcBH5Rwr0d | ☐ Yes ☐ No | | | ☐ Real ☐ Ghost |
| prod_GVV5efccnh13h9 | ☐ Yes ☐ No | | | ☐ Real ☐ Ghost |
| prod_FvNAJgnw48hwpZ | ☐ Yes ☐ No | | | ☐ Real ☐ Ghost |

**Summary:**
- Real products found: `_____`
- True ghosts found: `_____`
- Action needed: Add `_____` placeholders, keep `_____` blocked

---

## ✅ **Completed Actions So Far:**

1. ✅ Verified 2 products are real (`prod_HjYKGcWGP9r4EC`, `prod_FvNAlEGGL452nN`)
2. ✅ Added SQL placeholders for verified products
3. ✅ Removed verified products from blocklists (all 3 files)
4. ✅ Created investigation checklist

**Next:** Investigate remaining 5 products in Stripe Dashboard

---

## 🚀 **Ready to Test After Investigation:**

Once you've:
1. Investigated all 5 products
2. Added placeholders for any real ones found
3. Removed real ones from blocklists
4. Restarted backend

Then run:
```bash
# Via admin UI
/admin/streaming/stripe → "Simple Sync"

# Check logs for:
✅ "Syncing subscriptions..." (should sync historical ones now)
✅ No "GHOST BLOCKED" for products with placeholders
```

---

**Status:** Investigation in progress  
**Tools:** Stripe Dashboard, this checklist  
**Goal:** Determine if remaining 5 are real historical products or true ghosts

