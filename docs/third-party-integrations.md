# Third-Party Service Integrations

This document outlines the configuration and integration of third-party services for the BOME streaming platform.

## 1. Bunny.net Video Streaming

### Overview
Bunny.net provides global CDN and video streaming services for our video content.

### Configuration

#### Environment Variables
```bash
# Bunny.net Configuration
BUNNY_STORAGE_ZONE=your-storage-zone
BUNNY_API_KEY=your-api-key
BUNNY_PULL_ZONE=your-pull-zone
BUNNY_STREAM_LIBRARY_ID=your-library-id
BUNNY_STREAM_API_KEY=your-stream-api-key
BUNNY_REGION=de  # Default region (de, ny, la, sg, etc.)
```

#### Setup Steps
1. Create Bunny.net account at https://bunny.net
2. Create a Storage Zone for video uploads
3. Create a Pull Zone for CDN delivery
4. Set up Stream Library for video processing
5. Configure CORS settings for your domain
6. Set up webhook endpoints for video processing events

#### API Integration
- Video upload via Storage API
- Video processing via Stream API
- CDN delivery via Pull Zone URLs
- Webhook handling for processing status

#### Security
- API key authentication
- CORS configuration
- Rate limiting
- Signed URLs for private content

## 2. Stripe Payment Processing

### Overview
Stripe handles all subscription payments, billing, and financial transactions.

### Configuration

#### Environment Variables
```bash
# Stripe Configuration
STRIPE_SECRET_KEY=sk_test_...
STRIPE_PUBLISHABLE_KEY=pk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PRICE_ID_MONTHLY=price_...
STRIPE_PRICE_ID_YEARLY=price_...
STRIPE_CUSTOMER_PORTAL_URL=https://billing.stripe.com/...
```

#### Setup Steps
1. Create Stripe account at https://stripe.com
2. Set up subscription products and pricing
3. Configure webhook endpoints
4. Set up customer portal
5. Configure tax settings
6. Set up payment methods

#### Subscription Plans
- **Basic**: $9.99/month - Access to basic content
- **Premium**: $19.99/month - Full access + exclusive content
- **Annual**: 20% discount on yearly plans

#### Webhook Events
- `customer.subscription.created`
- `customer.subscription.updated`
- `customer.subscription.deleted`
- `invoice.payment_succeeded`
- `invoice.payment_failed`
- `payment_intent.succeeded`
- `payment_intent.payment_failed`

## 3. Digital Ocean Spaces

### Overview
Digital Ocean Spaces provides S3-compatible object storage for backups and static assets.

### Configuration

#### Environment Variables
```bash
# Digital Ocean Spaces Configuration
DO_SPACES_KEY=your-spaces-key
DO_SPACES_SECRET=your-spaces-secret
DO_SPACES_ENDPOINT=nyc3.digitaloceanspaces.com
DO_SPACES_BUCKET=your-bucket-name
DO_SPACES_REGION=nyc3
DO_SPACES_CDN_ENDPOINT=https://your-bucket.nyc3.cdn.digitaloceanspaces.com
```

#### Setup Steps
1. Create Digital Ocean account
2. Create Spaces bucket
3. Generate API keys
4. Configure CORS settings
5. Set up CDN (optional)
6. Configure backup policies

#### Use Cases
- Database backups
- Static asset storage
- User uploads
- System logs
- Configuration files

## 4. Email Service (SendGrid)

### Overview
SendGrid handles transactional emails, notifications, and marketing communications.

### Configuration

#### Environment Variables
```bash
# SendGrid Configuration
SENDGRID_API_KEY=SG.your-api-key
SENDGRID_FROM_EMAIL=noreply@bookofmormonevidence.org
SENDGRID_FROM_NAME=Book of Mormon Evidences
SENDGRID_TEMPLATE_ID_WELCOME=template-id
SENDGRID_TEMPLATE_ID_PASSWORD_RESET=template-id
SENDGRID_TEMPLATE_ID_SUBSCRIPTION=template-id
```

#### Setup Steps
1. Create SendGrid account
2. Verify domain ownership
3. Set up API key
4. Create email templates
5. Configure webhook events
6. Set up bounce handling

#### Email Templates
- Welcome email
- Password reset
- Email verification
- Subscription confirmation
- Payment receipts
- Account updates
- Newsletter

## 5. Integration Architecture

### Service Dependencies
```
Frontend (Svelte) → Backend (Go) → Third-Party Services
                                    ├── Bunny.net (Video)
                                    ├── Stripe (Payments)
                                    ├── Digital Ocean (Storage)
                                    └── SendGrid (Email)
```

### Security Considerations
- All API keys stored in environment variables
- HTTPS-only communication
- Rate limiting on all endpoints
- Input validation and sanitization
- Webhook signature verification
- CORS configuration

### Error Handling
- Retry mechanisms for failed requests
- Circuit breaker patterns
- Graceful degradation
- Comprehensive logging
- Alert notifications

### Monitoring
- API response times
- Error rates
- Usage metrics
- Cost tracking
- Performance alerts

## 6. Development vs Production

### Development Environment
- Use test/staging accounts
- Mock services for local development
- Reduced rate limits
- Debug logging enabled

### Production Environment
- Production API keys
- Full rate limits
- Error monitoring
- Performance optimization
- Backup strategies

## 7. Testing Strategy

### Unit Tests
- Mock third-party API responses
- Test error handling
- Validate webhook processing
- Test retry mechanisms

### Integration Tests
- Test with staging accounts
- Validate webhook flows
- Test payment processing
- Verify email delivery

### Load Testing
- Test API rate limits
- Validate CDN performance
- Test payment processing under load
- Monitor resource usage

## 8. Deployment Checklist

### Pre-Deployment
- [ ] All API keys configured
- [ ] Webhook endpoints registered
- [ ] CORS settings updated
- [ ] Email templates created
- [ ] Payment products configured
- [ ] CDN settings optimized

### Post-Deployment
- [ ] Webhook delivery verified
- [ ] Payment processing tested
- [ ] Email delivery confirmed
- [ ] CDN performance validated
- [ ] Monitoring alerts configured
- [ ] Backup procedures tested 