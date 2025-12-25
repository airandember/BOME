# Stripe Products V2 Migration Strand

## Purpose
This strand migrates all Stripe product-related endpoints from the legacy `stripe_products` table to the v2 tables (`stripe_products_v2`, `stripe_prices_v2`). The v2 tables provide proper foreign key relationships, better data consistency, and are synced directly from the Stripe API.

## Problem Statement
The legacy `stripe_products` table was inconsistent and not properly linked to price data. The v2 schema provides:
- Proper foreign keys between products and prices
- Audit trails (`first_synced_at`, `last_synced_at`)
- Stripe timestamps (`stripe_created_at`, `stripe_updated_at`)
- Custom fields for video access (`video_approved`, `is_legacy`)

## Implementation Details

### Backend Changes
- **File**: `backend/internal/routes/stripe_analytics_routes.go`
- **Endpoints Updated**:
  - `GET /admin/streaming/stripe/products/accordion` - Now reads from `stripe_products_v2`
  - `GET /admin/streaming/stripe/products/all` - Now reads from `stripe_products_v2`
  - `GET /admin/streaming/stripe/products/available` - Now reads from `stripe_products_v2`
  - `PUT /admin/streaming/stripe/products/video-approval/:id` - Now updates `stripe_products_v2`
  - `PUT /admin/streaming/stripe/products/:stripe_id/availability` - Now updates `stripe_products_v2`
  - `PUT /admin/streaming/stripe/products/bulk-availability` - Now updates `stripe_products_v2`
  - `POST /admin/streaming/stripe/products/update-legacy` - Now updates `stripe_products_v2`
  - `POST /admin/streaming/stripe/products/import-as-plans` - Now reads from V2 and uses correct column names

### Schema Mapping
| Legacy Field (`stripe_products`) | V2 Field (`stripe_products_v2`) |
|----------------------------------|----------------------------------|
| `id` | `id` |
| `stripe_id` | `stripe_id` |
| `name` | `name` |
| `description` | `description` |
| `active` | `active` |
| `available` | `active` (mapped) |
| `video_approved` | `video_approved` |
| `legacy_product` | `is_legacy` |
| `livemode` | N/A (V2 only stores live data) |
| `created_at` | `stripe_created_at` |
| `updated_at` | `last_synced_at` |

### Price Join
The V2 schema includes `stripe_prices_v2` with proper foreign keys:
```sql
LEFT JOIN stripe_prices_v2 pr ON p.id = pr.product_id AND pr.active = true
```

### Frontend Compatibility
All responses now include `"source": "stripe_products_v2"` to indicate data source.
The `available` field is mapped from `active` for frontend compatibility.

## Database Tables

### stripe_products_v2
```sql
CREATE TABLE stripe_products_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(500) NOT NULL,
    description TEXT,
    active BOOLEAN DEFAULT true,
    metadata JSONB,
    stripe_created_at TIMESTAMPTZ NOT NULL,
    stripe_updated_at TIMESTAMPTZ,
    video_approved BOOLEAN DEFAULT false,
    is_legacy BOOLEAN DEFAULT false,
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);
```

### stripe_prices_v2
```sql
CREATE TABLE stripe_prices_v2 (
    id SERIAL PRIMARY KEY,
    stripe_id VARCHAR(255) UNIQUE NOT NULL,
    product_id INTEGER NOT NULL REFERENCES stripe_products_v2(id),
    unit_amount BIGINT NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'usd',
    active BOOLEAN DEFAULT true,
    recurring_interval VARCHAR(20),
    recurring_interval_count INTEGER DEFAULT 1,
    metadata JSONB,
    stripe_created_at TIMESTAMPTZ NOT NULL,
    first_synced_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Data Sync
V2 tables are populated via:
1. **Full Sync**: `StripeSyncV2Service.SyncAll()` - Syncs all products, prices, customers, subscriptions
2. **Single Entity Sync**: Webhook-triggered updates
3. **Manual Sync**: Admin panel sync buttons

## Testing
1. Navigate to `/admin/streaming/subscriptions` in the admin panel
2. Verify the StripeProductsAccordion shows products from V2 tables
3. Test video approval toggle - should update `stripe_products_v2`
4. Test availability toggle - should update `stripe_products_v2.active`
5. Check browser console for `source: "stripe_products_v2"` in API responses

## Status
- [x] Backend routes updated to V2
- [x] Compilation successful
- [ ] Frontend testing required
- [ ] Production deployment pending

## Related Files
- `backend/internal/services/stripe_sync_v2.go` - V2 sync service
- `backend/migrations/050_create_stripe_v2_schema.sql` - V2 schema creation
- `frontend/src/lib/services/stripe-products.ts` - Frontend service (no changes needed)
- `frontend/src/routes/admin/streaming/subscriptions/components/StripeProductsAccordion.svelte` - UI component

## Known Issues
- Legacy `stripe_products` table still exists for backwards compatibility
- Some debug queries may still reference V2 tables explicitly (expected)

## Common Pitfalls
When updating queries from legacy to V2 tables, be aware of column name changes:
- `created_at` → `stripe_created_at` (for Stripe's original creation timestamp)
- `updated_at` → `last_synced_at` (for when we last synced from Stripe)
- `available` → `active` (V2 uses `active` instead of `available`)
- `legacy_product` → `is_legacy`

