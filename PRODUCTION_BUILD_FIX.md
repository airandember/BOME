# Production Build Fix - Oct 31, 2025

## Round 1: Type Mismatch Errors

### Issue
Production build was failing with type mismatch errors:

```
cmd/compare-subscribers/main.go:231:88: cannot use v2Sub.Email (variable of type string) as *string value in argument to ptrToString
```

## Round 2: Additional Build Errors

### Issue  
After fixing Round 1, production build failed again with:

```
cmd/compare-subscribers/main.go:27:8: assignment mismatch: 1 variable but database.New returns 2 values
cmd/compare-subscribers/main.go:133:32: cannot use &v2Subscribers[i] as *services.UnifiedSubscriber value
cmd/compare-subscribers/main.go:193:43: cannot use v2Sub as *services.UnifiedSubscriber value
cmd/compare-subscribers/main.go:272:9: invalid operation: operator - not defined on v1Sub.DaysUntilExpiry (variable of type *int)
```

## Root Cause
The `UnifiedSubscriber` struct has these field types:
- `Email`: `string` (not `*string`)
- `FirstName`: `string` (not `*string`)
- `LastName`: `string` (not `*string`)
- `PlanStatus`: `string` (not `*string`)
- `PlanType`: `string` (not `*string`)
- `PlanCurrency`: `string` (not `*string`)
- `PlanInterval`: `string` (not `*string`)

But `PlanName` IS `*string` (nullable).

The `compare-subscribers` tool was incorrectly calling `ptrToString()` on non-pointer string fields.

## Fixes Applied

### Round 1 Fixes
Updated `backend/cmd/compare-subscribers/main.go`:

1. **Lines 210, 217, 224**: Changed `ptrToString(v1Sub.Email)` to `v1Sub.Email`
2. **Lines 230-232**: Changed email comparison to use `v1Sub.Email != v2Sub.Email` directly
3. **Lines 234-239**: Changed full name comparison to concatenate `FirstName + " " + LastName` 
4. **Lines 252-254**: Changed plan status comparison to use `v1Sub.PlanStatus != v2Sub.PlanStatus` directly

### Round 2 Fixes
Updated `backend/cmd/compare-subscribers/main.go`:

1. **Line 28**: Fixed `database.New()` to capture both return values: `db, err := database.New(cfg)`
2. **Line 133**: Added `convertV2ToV1()` call to convert `UnifiedSubscriberV2` to `UnifiedSubscriber`
3. **Line 194**: Added conversion function call in `compareUser()`
4. **Lines 275-285**: Fixed `DaysUntilExpiry` comparison to dereference `*int` pointers before subtraction
5. **Lines 358-413**: Created `convertV2ToV1()` function to handle type differences:
   - Convert `*int` price (cents) to `float64` (dollars)
   - Handle `time.Time` vs `string` date field differences
   - Map `ManualAccessGranted` (v2) to `ManualVideoAccess` (v1)
6. **Line 10**: Added `time` import
7. **Line 57**: Removed duplicate `err` declaration

## Verification
✅ No linter errors
✅ Type mismatches resolved
✅ `PlanName` comparisons (lines 257-259) remain correct as they are `*string`

## Deployment
The fix is ready for production deployment. The build should now succeed.

## Related Files
- `backend/cmd/compare-subscribers/main.go` - Fixed type mismatches
- `backend/internal/services/subscriber_elastic_service.go` - Source of truth for field types

