# 🚀 Resend Subscription Tier Support

## Overview
The BOME email system is designed to dynamically handle different Resend subscription tiers without code changes.

## Supported Tiers

### 🆓 **Free Tier** (Default)
- **Daily Limit**: 100 emails
- **Monthly Limit**: 3,000 emails
- **Contacts**: Unlimited
- **Domain Verification**: Required for production

### 💎 **Pro Tier** ($20/month)
- **Daily Limit**: 1,000 emails  
- **Monthly Limit**: 50,000 emails
- **All Free features**: Plus advanced analytics

### 🚀 **Business Tier** ($85/month)
- **Daily Limit**: 5,000 emails
- **Monthly Limit**: 100,000 emails
- **All Pro features**: Plus dedicated IP

### 🏢 **Enterprise Tier** (Custom)
- **Daily Limit**: Custom (10,000+)
- **Monthly Limit**: Custom (500,000+)
- **All Business features**: Plus custom integrations

## How to Upgrade

### 1. **Upgrade in Resend Dashboard**
```bash
# Visit: https://resend.com/settings/billing
# Select your desired tier
# Complete payment process
```

### 2. **Update BOME Configuration**
```bash
# Run the setup command with new limits
export RESEND_DAILY_LIMIT=1000      # Pro tier example
export RESEND_MONTHLY_LIMIT=50000   # Pro tier example
./setup-resend
```

### 3. **Or Update via Admin Dashboard**
- Visit `/admin/streaming/email`
- Update "Daily Email Limit" and "Monthly Email Limit"
- Click "Update Settings"

## Configuration Keys

The system uses these database settings:

```sql
-- Daily limits
resend_daily_limit = "100"    -- Free: 100, Pro: 1000, Business: 5000
resend_monthly_limit = "3000" -- Free: 3000, Pro: 50000, Business: 100000
email_enabled = "true"        -- Enable/disable email sending
```

## Automatic Detection

The system will:
- ✅ **Respect your configured limits**
- ✅ **Show usage percentages correctly**
- ✅ **Prevent sending when limits reached**
- ✅ **Display monthly tracking for all tiers**
- ✅ **Scale UI automatically**

## Benefits of Each Tier

### Free → Pro Upgrade Benefits
- **10x daily capacity** (100 → 1,000)
- **16x monthly capacity** (3K → 50K)
- **Advanced analytics**
- **Priority support**

### Pro → Business Upgrade Benefits  
- **5x daily capacity** (1K → 5K)
- **2x monthly capacity** (50K → 100K)
- **Dedicated IP address**
- **Enhanced deliverability**

## Monitoring

The admin dashboard automatically shows:
- **Real-time usage** against your configured limits
- **Monthly progress** tracking
- **Color-coded warnings** when approaching limits
- **Remaining capacity** calculations

## No Code Changes Required! 🎉

The system is designed to handle any Resend tier upgrade without:
- ❌ Code modifications
- ❌ Database migrations  
- ❌ Service restarts
- ❌ Downtime

Just update the limits in settings and you're ready to scale! 🚀
