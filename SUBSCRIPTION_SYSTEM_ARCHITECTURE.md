# Subscription System Architecture

## Overview

The BOME subscription system is built with a clean, linear routing architecture that handles three main components:

1. **Subscription Plans** - The available plans users can subscribe to
2. **Subscriptions** - Active user subscriptions with payment processing
3. **Subscribers** - Users who have active subscriptions (plan_id != null)

## System Components

### 1. Subscription Plans (`/api/admin/subscription-plans`)

**Backend Service**: `services/subscription_plans.go`
**Backend Routes**: `routes/subscription_plans.go`
**Frontend Service**: `services/streaming-subscriptions.ts`

#### Features:
- ✅ Create, read, update, delete subscription plans
- ✅ Plan activation/deactivation
- ✅ Promotion management
- ✅ Feature management
- ✅ Pricing and billing interval configuration
- ✅ Stripe integration

#### API Endpoints:
```
GET    /api/admin/subscription-plans/          # List all plans with filters
POST   /api/admin/subscription-plans/          # Create new plan
GET    /api/admin/subscription-plans/:id       # Get specific plan
PUT    /api/admin/subscription-plans/:id       # Update plan
DELETE /api/admin/subscription-plans/:id       # Soft delete plan
GET    /api/admin/subscription-plans/count     # Get plan count
GET    /api/subscription-plans/active          # Public: Get active plans
GET    /api/subscription-plans/promoted        # Public: Get promoted plans
GET    /api/subscription-plans/:id             # Public: Get specific plan
```

### 2. Subscriptions (`/api/admin/subscriptions`)

**Backend Service**: `services/subscription.go` (existing)
**Backend Routes**: `routes/subscription.go` (existing)
**Database**: `database/subscription.go` (existing)

#### Features:
- ✅ User subscription management
- ✅ Stripe payment processing
- ✅ Subscription lifecycle management
- ✅ Billing history
- ✅ Refund processing
- ✅ Webhook handling

#### API Endpoints:
```
# Customer Routes
GET    /api/subscriptions/                     # Get user's subscription
POST   /api/subscriptions/                     # Create subscription
DELETE /api/subscriptions/                     # Cancel subscription
PUT    /api/subscriptions/                     # Update subscription
GET    /api/subscriptions/history              # Get subscription history
GET    /api/subscriptions/billing              # Get billing info
POST   /api/subscriptions/refund               # Request refund

# Admin Routes
GET    /api/admin/subscriptions/               # List all subscriptions
GET    /api/admin/subscriptions/:id            # Get specific subscription
PUT    /api/admin/subscriptions/:id            # Update subscription (admin)
DELETE /api/admin/subscriptions/:id            # Cancel subscription (admin)
POST   /api/admin/subscriptions/:id/refund     # Process refund (admin)
GET    /api/admin/subscriptions/analytics      # Get analytics
GET    /api/admin/subscriptions/metrics        # Get metrics

# Public Routes
GET    /api/subscription/plans                 # Get available plans
POST   /api/subscription/checkout              # Create checkout session
```

### 3. Subscribers (`/api/admin/subscribers`)

**Backend Service**: `services/subscribers.go` ⭐ **NEW**
**Backend Routes**: `routes/subscribers.go` ⭐ **NEW**
**Frontend Service**: `services/streaming-subscribers.ts` ⭐ **NEW**

#### Features:
- ✅ List all users with active subscriptions
- ✅ Subscriber filtering and search
- ✅ Subscriber statistics and analytics
- ✅ Plan-based subscriber grouping
- ✅ Status-based subscriber filtering
- ✅ Revenue tracking per subscriber

#### API Endpoints:
```
GET    /api/admin/subscribers/                 # List all subscribers with filters
GET    /api/admin/subscribers/:id              # Get specific subscriber
GET    /api/admin/subscribers/count            # Get subscriber count
GET    /api/admin/subscribers/stats            # Get subscriber statistics
GET    /api/admin/subscribers/plan/:planId     # Get subscribers by plan
GET    /api/admin/subscribers/status/:status   # Get subscribers by status
GET    /api/admin/subscribers/search           # Search subscribers
```

## Database Schema

### Subscription Plans Table
```sql
CREATE TABLE subscription_plans (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    short_desc VARCHAR(500),
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    interval VARCHAR(20) NOT NULL,
    interval_count INTEGER DEFAULT 1,
    stripe_price_id VARCHAR(255),
    features JSONB,
    is_active BOOLEAN DEFAULT true,
    is_promoted BOOLEAN DEFAULT false,
    promotion_end_date TIMESTAMP,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

### Subscriptions Table
```sql
CREATE TABLE subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    plan_id INTEGER REFERENCES subscription_plans(id),
    stripe_subscription_id VARCHAR(255),
    stripe_price_id VARCHAR(255),
    status VARCHAR(50) NOT NULL,
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT false,
    cancellation_reason TEXT,
    refund_amount DECIMAL(10,2),
    refund_reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);
```

### Users Table (Enhanced)
```sql
-- Users table already exists with plan_id column added via migration
-- The plan_id column links users to their subscription plans
```

## Frontend Integration

### Subscription Plans Dashboard
- **Location**: `/admin/streaming/subscriptions`
- **Features**: Full CRUD operations for subscription plans
- **Styling**: Glass morphism design with modern UI

### Subscribers Dashboard (To be created)
- **Location**: `/admin/streaming/subscribers`
- **Features**: 
  - Subscriber list with filtering
  - Subscriber statistics
  - Plan-based grouping
  - Search functionality
  - Export capabilities

### Subscription Management Dashboard (To be created)
- **Location**: `/admin/streaming/subscription-management`
- **Features**:
  - Active subscription overview
  - Payment processing
  - Refund management
  - Analytics and reporting

## Security & Authentication

### Admin Routes
- All admin routes require authentication (`middleware.AuthRequired()`)
- Admin routes require admin role (`middleware.AdminRequired()`)
- Session activity tracking for audit trails

### Public Routes
- Subscription plan listing is public
- Checkout sessions are public
- User subscription management requires authentication

## Data Flow

### 1. Subscription Plan Creation
```
Admin → Create Plan → Database → Stripe Price ID → Plan Available
```

### 2. User Subscription
```
User → Select Plan → Stripe Checkout → Webhook → Subscription Created → User becomes Subscriber
```

### 3. Subscriber Management
```
Admin → View Subscribers → Filter/Search → Analytics → Reports
```

## API Response Formats

### Subscription Plans
```json
{
  "plans": [
    {
      "id": 1,
      "name": "Premium Plan",
      "description": "Full access to all content",
      "price": 29.99,
      "currency": "USD",
      "interval": "monthly",
      "features": ["HD Quality", "Ad Free", "Downloads"],
      "is_active": true,
      "is_promoted": false
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0
  }
}
```

### Subscribers
```json
{
  "subscribers": [
    {
      "id": 1,
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "plan_id": 1,
      "plan_name": "Premium Plan",
      "plan_price": 29.99,
      "subscription_status": "active",
      "current_period_end": "2024-02-01T00:00:00Z"
    }
  ],
  "pagination": {
    "limit": 10,
    "offset": 0
  }
}
```

### Subscriber Statistics
```json
{
  "stats": {
    "total_subscribers": 150,
    "active_subscribers": 120,
    "trialing_subscribers": 15,
    "past_due_subscribers": 10,
    "canceled_subscribers": 5,
    "monthly_revenue": 3598.80,
    "annual_revenue": 43185.60,
    "average_revenue_per_user": 29.99,
    "churn_rate": 3.33
  }
}
```

## Error Handling

### Standard Error Response
```json
{
  "error": "Error message",
  "details": "Detailed error information",
  "code": "ERROR_CODE"
}
```

### Common Error Codes
- `VALIDATION_ERROR` - Request validation failed
- `NOT_FOUND` - Resource not found
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `PAYMENT_ERROR` - Stripe payment error
- `DATABASE_ERROR` - Database operation failed

## Monitoring & Analytics

### Subscription Analytics
- Revenue tracking
- Churn rate analysis
- Plan popularity metrics
- Payment success rates
- Geographic distribution

### Subscriber Analytics
- Growth trends
- Engagement metrics
- Plan migration patterns
- Lifetime value calculations

## Future Enhancements

### Planned Features
1. **Advanced Analytics Dashboard**
   - Real-time metrics
   - Predictive analytics
   - Custom reporting

2. **Automated Billing Management**
   - Dunning management
   - Automatic retry logic
   - Payment method updates

3. **Multi-currency Support**
   - Dynamic currency conversion
   - Regional pricing
   - Tax calculation

4. **Subscription Tiers**
   - Family plans
   - Enterprise subscriptions
   - Custom pricing

5. **Integration APIs**
   - Webhook endpoints
   - Third-party integrations
   - API rate limiting

## Deployment Notes

### Environment Variables
```bash
# Stripe Configuration
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...

# Database Configuration
DATABASE_URL=postgresql://...

# Application Configuration
ENVIRONMENT=production
SERVER_PORT=8080
```

### Database Migrations
All required migrations are included in the `backend/migrations/` directory:
- `010_create_subscription_plans_table.sql`
- `011_enhance_subscriptions_table.sql`
- `013_add_subscription_analytics.sql`

## Testing

### API Testing
- All endpoints include proper error handling
- Input validation on all requests
- Authentication and authorization checks
- Rate limiting protection

### Frontend Testing
- Component testing for all subscription components
- Integration testing with backend APIs
- User flow testing for subscription process

## Support & Maintenance

### Monitoring
- Health check endpoints
- Error logging and alerting
- Performance monitoring
- Database query optimization

### Backup & Recovery
- Automated database backups
- Point-in-time recovery
- Data export capabilities
- Disaster recovery procedures

---

This architecture provides a robust, scalable foundation for managing subscriptions, plans, and subscribers with clean separation of concerns and comprehensive API coverage. 