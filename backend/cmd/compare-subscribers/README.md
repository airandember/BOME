# Subscriber Comparison Tool (v1 vs v2)

**Phase 8**: Compare v1 and v2 subscriber data to verify migration readiness.

---

## 🎯 Purpose

This tool compares subscriber data between:
- **v1**: Old fragmented tables (users, stripe_customers, stripe_subscriptions, subscription_plans)
- **v2**: New unified tables (stripe_customers_v2, stripe_subscriptions_v2, stripe_prices_v2, user_stripe_customers_v2)

---

## 🚀 Usage

### Build

```bash
cd backend/cmd/compare-subscribers
go build -o compare-subscribers.exe .
```

### Run

**Compare all subscribers**:
```bash
./compare-subscribers.exe
```

**Compare specific user**:
```bash
./compare-subscribers.exe -user 7374
```

**Show only differences**:
```bash
./compare-subscribers.exe -diff-only
```

**Save to file**:
```bash
./compare-subscribers.exe -output comparison-report.json
```

**Verbose logging**:
```bash
./compare-subscribers.exe -v
```

---

## 📋 Command Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-user` | int | 0 | Compare specific user ID (0 = all users) |
| `-pretty` | bool | true | Pretty print JSON output |
| `-output` | string | "" | Write results to file (default: stdout) |
| `-diff-only` | bool | false | Show only subscribers with differences |
| `-v` | bool | false | Verbose logging |

---

## 📊 Output Format

### JSON Structure

```json
{
  "total_compared": 150,
  "identical_count": 145,
  "different_count": 3,
  "v1_only_count": 1,
  "v2_only_count": 1,
  "subscribers": [
    {
      "user_id": 7374,
      "email": "user@example.com",
      "status": "different",
      "v1_data": { ... },
      "v2_data": { ... },
      "differences": [
        {
          "field": "mrr_contribution",
          "v1_value": 7.97,
          "v2_value": 8.00
        }
      ]
    }
  ],
  "summary": {
    "mrr_contribution": 2,
    "plan_status": 1
  }
}
```

### Status Values

- `identical`: v1 and v2 data match perfectly
- `different`: Data exists in both but has differences
- `v1_only`: User exists in v1 but not v2
- `v2_only`: User exists in v2 but not v1
- `both_missing`: User ID exists but no data in either version (shouldn't happen)

---

## 🔍 Fields Compared

### Critical Fields
- `email`
- `full_name`
- `has_video_access`
- `has_active_plan`
- `plan_status`
- `plan_name`
- `mrr_contribution` (with 0.01 tolerance)
- `arr_contribution` (with 0.01 tolerance)
- `days_until_expiry` (with 1 day tolerance)

### Tolerance Settings

- **Float values** (MRR, ARR): ±$0.01 tolerance (for floating point precision)
- **Days until expiry**: ±1 day tolerance (for timing differences during comparison)

---

## 📈 Success Criteria

### ✅ Ready for Migration
```
✅ Identical: 100% (0 differences)
❌ V1 Only: 0
❌ V2 Only: 0
```

### ⚠️ Review Required
```
✅ Identical: 95%+
⚠️  Different: <5% (minor differences)
```

### ❌ Investigation Needed
```
✅ Identical: <90%
⚠️  Different: >10%
❌ V1/V2 Only: >0
```

---

## 🧪 Example Usage

### 1. Quick Check (All Users)
```bash
./compare-subscribers.exe -diff-only
```
**Output**:
```
📊 COMPARISON SUMMARY
===================================================================
Total Compared:   150
✅ Identical:     148 (98.7%)
⚠️  Different:     2 (1.3%)
❌ V1 Only:       0
❌ V2 Only:       0
===================================================================
📋 Differences by Field:
  - mrr_contribution: 2
===================================================================
⚠️  RESULT: 2 discrepancies found. Review before migration.
===================================================================
```

### 2. Single User Deep Dive
```bash
./compare-subscribers.exe -user 7374 -v
```

### 3. Full Report for Documentation
```bash
./compare-subscribers.exe -output full-comparison-report.json -pretty
```

---

## 🔧 Troubleshooting

### "Failed to initialize database"
- Check `.env` file exists and has correct database credentials
- Ensure PostgreSQL is running

### "Failed to get v1 subscribers"
- Verify v1 tables exist (users, stripe_customers, stripe_subscriptions)
- Check database permissions

### "Failed to get v2 subscribers"
- Verify v2 tables exist (run migration 050_create_stripe_v2_schema.sql)
- Ensure v2 tables have data (run stripe-sync tool first)

### High difference count
- Run `stripe-sync` tool to ensure v2 tables are up to date
- Check for ghost subscriptions in Stripe (see Sub_Ghosts_table.txt)
- Review customer linking status (`customer-linking --stats`)

---

## 🎯 Next Steps After Comparison

1. **If 100% match**: Proceed to Phase 9 (Data Migration)
2. **If <5% differences**: Investigate specific discrepancies, fix, re-run
3. **If >5% differences**: Review data sync process, check for ghost subscriptions

---

## 📝 Related Tools

- `stripe-sync`: Sync Stripe data to v2 tables
- `customer-linking`: Link users to Stripe customers
- `/admin/subscriber-elastic/comparison/:id`: Web UI comparison endpoint

---

**Phase 8.1 Complete!** 🎉

