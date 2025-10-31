# Customer Linking CLI Tool

This tool helps you link users to their Stripe customers in the v2 schema.

## Usage

### Show Linking Statistics

```bash
cd backend
go run cmd/customer-linking/main.go --stats
```

### Show Unlinked Customers

```bash
go run cmd/customer-linking/main.go --unlinked
```

### Link All Users

This will automatically link all users to their Stripe customers based on email matching:

```bash
go run cmd/customer-linking/main.go --link-all
```

### Link a Specific User

```bash
go run cmd/customer-linking/main.go --user 7113
```

### Pretty Print Output

Add `--pretty` to any command for formatted JSON:

```bash
go run cmd/customer-linking/main.go --stats --pretty
go run cmd/customer-linking/main.go --unlinked --pretty
```

## How It Works

1. **Email Matching**: Finds all Stripe customers with matching email addresses (case-insensitive)
2. **Link Creation**: Creates entries in `user_stripe_customers_v2` table
3. **Primary Selection**: The most recently created customer is marked as `is_primary`
4. **User Update**: Updates `users.stripe_customer_id` with the primary customer

## API Endpoints

You can also use the admin API endpoints:

### Get Statistics
```bash
GET /api/v1/admin/customer-linking/stats
```

### Get Unlinked Customers
```bash
GET /api/v1/admin/customer-linking/unlinked
```

### Link a Specific User
```bash
POST /api/v1/admin/customer-linking/user/:user_id
```

### Link All Users
```bash
POST /api/v1/admin/customer-linking/all
```

### Get User's Customers
```bash
GET /api/v1/admin/customer-linking/user/:user_id/customers
```

### Set Primary Customer
```bash
PUT /api/v1/admin/customer-linking/user/:user_id/primary
{
  "stripe_customer_id": "cus_ABC123"
}
```

## Output Fields

### LinkResult
- `user_id`: User's database ID
- `email`: User's email
- `customers_found`: Total Stripe customers found for this email
- `customers_linked`: Number of customers successfully linked
- `primary_customer`: The Stripe customer ID marked as primary
- `skipped_customers`: Array of customer IDs that failed to link
- `error`: Error message if the operation failed
- `linked_at`: Timestamp of the linking operation

### UnlinkedCustomer
- `stripe_customer_id`: Stripe customer ID
- `email`: Customer's email
- `user_id`: Matching user ID (if found)
- `user_exists`: Whether a user with this email exists
- `has_subscriptions`: Whether customer has active subscriptions
- `created_at`: When the customer was created in Stripe

## Best Practices

1. **Run stats first**: Check how many customers need linking
2. **Review unlinked**: See which customers can't be automatically linked
3. **Test with one user**: Try `--user` flag with a single user first
4. **Bulk link**: Use `--link-all` after testing
5. **Verify**: Check stats again to confirm all customers are linked

## Troubleshooting

### No customers found for user
- Verify email matches between `users` table and Stripe
- Check if Stripe customers exist in `stripe_customers_v2` (run sync first)

### Customers already linked
- The tool is idempotent - safe to run multiple times
- Already linked customers will have their `last_synced_at` updated

### Multiple customers per user
- This is normal and expected
- The most recent customer is automatically set as primary
- Use the API to change primary if needed

