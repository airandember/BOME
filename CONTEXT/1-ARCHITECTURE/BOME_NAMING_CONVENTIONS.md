# BOME Naming Conventions

## Overview
This document defines standardized naming conventions for the BOME system to ensure consistency, maintainability, and clear communication between frontend and backend components.

---

## 🔧 Backend Naming Conventions

### API Routes (`backend/internal/routes/`)

#### File Naming
```
{domain}_{resource}.go
{domain}_{action}.go
{domain}_{feature}.go
```

**Examples:**
- `subscription_plans.go` - Subscription plan management
- `admin_streaming.go` - Admin streaming features
- `master_video_routes.go` - Master video routing
- `unified_analytics.go` - Unified analytics endpoints
- `standardized_roles.go` - Role management

#### Route Path Structure
```
/api/v1/{admin|public}/{domain}/{resource}/{action}
```

**Patterns:**
- **Admin routes:** `/api/v1/admin/{domain}/{resource}`
- **Public routes:** `/api/v1/{domain}/{resource}`
- **Nested resources:** `/api/v1/admin/{domain}/{resource}/{id}/{sub-resource}`

**Examples:**
```
/api/v1/admin/subscription-plans
/api/v1/admin/streaming/analytics
/api/v1/admin/subscribers
/api/v1/subscription-plans
/api/v1/videos
```

#### Function Naming
```go
// Route handlers
func (h *Handler) {Action}{Resource}(c *gin.Context)
func (h *Handler) {Action}{Resource}{Action}(c *gin.Context)

// Examples:
func (h *Handler) GetSubscriptionPlans(c *gin.Context)
func (h *Handler) CreateSubscriptionPlan(c *gin.Context)
func (h *Handler) UpdateSubscriptionPlan(c *gin.Context)
func (h *Handler) DeleteSubscriptionPlan(c *gin.Context)
func (h *Handler) GetSubscriptionPlanAnalytics(c *gin.Context)
```

### Services (`backend/internal/services/`)

#### File Naming
```
{domain}_{resource}.go
{domain}_{feature}.go
{external_service}.go
```

**Examples:**
- `subscription_plans.go` - Subscription plan business logic
- `subscription_analytics.go` - Subscription analytics
- `stripe.go` - Stripe payment integration
- `bunny.go` - Bunny.net CDN integration
- `youtube.go` - YouTube API integration

#### Struct Naming
```go
type {Resource}Service struct {
    // fields
}

// Examples:
type SubscriptionPlanService struct {}
type SubscriberService struct {}
type AnalyticsService struct {}
```

#### Method Naming
```go
// CRUD operations
func (s *Service) Create{Resource}(ctx context.Context, data *Create{Resource}Request) (*{Resource}, error)
func (s *Service) Get{Resource}(ctx context.Context, id string) (*{Resource}, error)
func (s *Service) List{Resources}(ctx context.Context, filters *{Resource}Filters) (*{Resource}List, error)
func (s *Service) Update{Resource}(ctx context.Context, id string, data *Update{Resource}Request) (*{Resource}, error)
func (s *Service) Delete{Resource}(ctx context.Context, id string) error

// Business logic
func (s *Service) {Action}{Resource}(ctx context.Context, params *{Action}Params) (*{ActionResult}, error)

// Examples:
func (s *SubscriptionPlanService) CreateSubscriptionPlan(ctx context.Context, data *CreateSubscriptionPlanRequest) (*SubscriptionPlan, error)
func (s *SubscriptionPlanService) GetSubscriptionPlanAnalytics(ctx context.Context, params *AnalyticsParams) (*AnalyticsResult, error)
```

#### Interface Naming
```go
type {Resource}ServiceInterface interface {
    // methods
}

// Examples:
type SubscriptionPlanServiceInterface interface {}
type SubscriberServiceInterface interface {}
```

---

## 🎨 Frontend Naming Conventions

### Services (`frontend/src/lib/services/`)

#### File Naming
```
{domain}-{resource}.ts
{domain}-{feature}.ts
{resource}Service.ts
```

**Examples:**
- `streaming-subscriptions.ts` - Streaming subscription management
- `streaming-analytics.ts` - Streaming analytics
- `streaming-customers.ts` - Customer management
- `AnalyticsService.ts` - General analytics service
- `designTokenService.ts` - Design system tokens

#### Class Naming
```typescript
export class {Domain}{Resource}Service {
    // methods
}

// Examples:
export class StreamingSubscriptionService {}
export class StreamingAnalyticsService {}
export class AnalyticsService {}
```

#### Method Naming
```typescript
// CRUD operations
static async get{Resources}(params?: {Resource}Filters): Promise<{Resource}[]>
static async get{Resource}(id: string): Promise<{Resource}>
static async create{Resource}(data: Create{Resource}Data): Promise<{Resource}>
static async update{Resource}(id: string, data: Update{Resource}Data): Promise<{Resource}>
static async delete{Resource}(id: string): Promise<void>

// Business logic
static async {action}{Resource}(params: {Action}Params): Promise<{ActionResult}>

// Examples:
static async getSubscriptionPlans(params?: SubscriptionPlanFilters): Promise<SubscriptionPlan[]>
static async createSubscriptionPlan(data: CreateSubscriptionPlanData): Promise<SubscriptionPlan>
static async getSubscriptionAnalytics(params: AnalyticsParams): Promise<AnalyticsResult>
```

#### Interface Naming
```typescript
export interface {Resource} {
    // properties
}

export interface Create{Resource}Data {
    // properties
}

export interface Update{Resource}Data {
    // properties
}

export interface {Resource}Filters {
    // properties
}

// Examples:
export interface SubscriptionPlan {}
export interface CreateSubscriptionPlanData {}
export interface SubscriptionPlanFilters {}
```

### Routes (`frontend/src/routes/`)

#### Directory Naming
```
{domain}/{feature}/
{domain}/{resource}/
{action}/
```

**Examples:**
- `admin/streaming/subscriptions/` - Admin streaming subscriptions
- `admin/streaming/analytics/` - Admin streaming analytics
- `videos/` - Video management
- `account/` - Account management
- `login/` - Authentication

#### Page File Naming
```
+page.svelte          // Main page
+layout.svelte        // Layout wrapper
+error.svelte         // Error page
+loading.svelte       // Loading state
```

---

## 🔗 Middleware Naming Conventions

### Backend Middleware (`backend/internal/middleware/`)

#### File Naming
```
{feature}_middleware.go
{action}_middleware.go
```

**Examples:**
- `auth_middleware.go` - Authentication middleware
- `cors_middleware.go` - CORS handling
- `rate_limit_middleware.go` - Rate limiting
- `logging_middleware.go` - Request logging

#### Function Naming
```go
func {Action}Middleware() gin.HandlerFunc
func {Action}Middleware(config *{Action}Config) gin.HandlerFunc

// Examples:
func AuthMiddleware() gin.HandlerFunc
func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc
func CORSMiddleware() gin.HandlerFunc
```

### Frontend Middleware (`frontend/src/lib/middleware/`)

#### File Naming
```
{feature}.ts
{action}.ts
```

**Examples:**
- `auth.ts` - Authentication middleware
- `guards.ts` - Route guards
- `validation.ts` - Form validation

#### Function Naming
```typescript
export function {action}{Feature}(): void
export function {action}{Feature}(config: {Feature}Config): void

// Examples:
export function requireAuth(): void
export function requireAdmin(): void
export function validateForm(config: ValidationConfig): void
```

---

## 📊 Database Naming Conventions

### Table Naming
```
{domain}_{resource}s
{domain}_{feature}
```

**Examples:**
- `subscription_plans` - Subscription plans table
- `subscribers` - Subscribers table
- `streaming_analytics` - Streaming analytics table
- `user_roles` - User roles table

### Column Naming
```
{resource}_{property}
{action}_{timestamp}
```

**Examples:**
- `plan_name` - Plan name column
- `subscription_status` - Subscription status
- `created_at` - Creation timestamp
- `updated_at` - Update timestamp

---

## 🔄 API Response Naming Conventions

### Success Response Structure
```json
{
  "data": {
    // Response data
  },
  "message": "Success message",
  "status": "success"
}
```

### Error Response Structure
```json
{
  "error": "Error message",
  "status": "error",
  "code": "ERROR_CODE"
}
```

### List Response Structure
```json
{
  "data": {
    "{resources}": [
      // Array of resources
    ],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "pages": 10
    }
  }
}
```

---

## 🎯 Best Practices

### 1. Consistency
- Use the same naming pattern across similar components
- Maintain consistency between frontend and backend naming
- Follow established patterns in the codebase

### 2. Clarity
- Use descriptive names that clearly indicate purpose
- Avoid abbreviations unless widely understood
- Use domain-specific terminology consistently

### 3. Scalability
- Design naming conventions that scale with the application
- Consider future features when establishing patterns
- Use modular naming that supports feature expansion

### 4. Documentation
- Document any deviations from these conventions
- Update this document when new patterns emerge
- Use comments to explain complex naming decisions

---

## 📝 Quick Reference

| Component | Pattern | Example |
|-----------|---------|---------|
| Backend Routes | `{domain}_{resource}.go` | `subscription_plans.go` |
| Backend Services | `{domain}_{resource}.go` | `subscription_plans.go` |
| Frontend Services | `{domain}-{resource}.ts` | `streaming-subscriptions.ts` |
| Frontend Routes | `{domain}/{resource}/` | `admin/streaming/subscriptions/` |
| Database Tables | `{domain}_{resource}s` | `subscription_plans` |
| API Endpoints | `/api/v1/{admin\|public}/{domain}/{resource}` | `/api/v1/admin/subscription-plans` | 