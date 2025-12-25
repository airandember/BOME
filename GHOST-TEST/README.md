# Stripe Ghost Product Test

This test script checks if the known "ghost" product IDs are truly deleted from Stripe or just archived.

## Known Ghost Product IDs Being Tested

- `prod_HEmcX1PE8TO2CO`
- `prod_FvNAeI348dup9w`
- `prod_HF5YzcBH5Rwr0d`
- `prod_GVV5efccnh13h9`
- `prod_FvNAJgnw48hwpZ`

## How to Run

### Option 1: Double-click (Windows)
1. Double-click `run_test.bat`
2. Enter your Stripe secret key when prompted

### Option 2: PowerShell
```powershell
cd "GHOST TEST"
.\run_test.ps1
```

### Option 3: Direct Go
```bash
cd "GHOST TEST"
go run ghost_test.go
```

### Option 4: Set Environment Variable First
```powershell
$env:STRIPE_SECRET_KEY = "sk_live_your_key_here"
go run ghost_test.go
```

## Expected Results

The script will show one of these statuses for each product:

| Status | Meaning |
|--------|---------|
| ✅ ACTIVE | Product exists and is active in Stripe |
| 📦 ARCHIVED | Product exists but is archived (active=false) |
| 👻 DELETED | Product returns `resource_missing` - TRUE GHOST |
| ❌ ERROR | API error (check key/network) |

## What to Send to Stripe Support

After running, a results file will be created: `ghost_test_results_YYYY-MM-DD_HHMMSS.txt`

Send this file to Stripe support as evidence if you see `DELETED` status for any products.

## Key Question for Stripe

If Stripe says these products are "archived," but the API returns `resource_missing` (404), 
then there's a discrepancy that needs to be explained.

