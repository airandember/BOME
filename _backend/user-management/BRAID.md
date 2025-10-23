# 🧬 User Management Braid - Backend
**User profiles, preferences, RBAC, and account management**

---

## 🔗 **Cross-Repository Braid**

> **⚠️ IMPORTANT**: This is the **backend portion** of the User Management Braid.  
> **Frontend portion**: See `_frontend/braids/user-management/BRAID.md`  
> **Unified context**: Both directories are part of the same braid!

---

## 📋 **Backend Overview**

**Purpose**: Server-side user profile management, RBAC, and account operations  
**Technology**: Go, PostgreSQL, JWT-based authorization  
**Complexity**: Medium-High (RBAC, Profile Management, Admin Operations)  
**Dependencies**: Authentication Braid (user identity)

---

## 🎯 **Key Features**

### **1. User Profile Management**:
- Profile CRUD operations
- Avatar/image upload
- Bio and personal information
- Profile picture management
- Account settings

### **2. Role-Based Access Control (RBAC)**:
- User roles (user, admin, moderator, etc.)
- Permission checking
- Role assignments
- Hierarchical permissions
- Middleware enforcement

### **3. User Preferences**:
- Theme preferences (light/dark)
- Notification settings
- Privacy settings
- Language preferences
- Email preferences

### **4. Admin User Management**:
- User listing and search
- User account suspension/activation
- Role management
- User deletion (GDPR compliance)
- Bulk operations

### **5. Activity Tracking**:
- User action logging
- Audit trail
- Login history
- Activity dashboard
- Security monitoring

---

## 🗄️ **Database Schema**

### **Core Tables**:

**users** (Extended from auth braid):
- `id` - Primary key
- `email` - Email address (unique)
- `first_name` - First name
- `last_name` - Last name
- `profile_picture_url` - Avatar URL
- `bio` - User biography
- `role` - User role (user/admin/etc.)
- `is_active` - Account status
- `email_verified` - Email verification status
- `created_at` - Account creation
- `updated_at` - Last update
- `last_login` - Last login timestamp

**user_preferences**:
- `id` - Primary key
- `user_id` - Foreign key to users
- `theme` - UI theme (light/dark/auto)
- `language` - Preferred language
- `timezone` - User timezone
- `email_notifications` - Email notification settings
- `push_notifications` - Push notification settings
- `privacy_settings` - JSON privacy configuration
- `created_at` - Created timestamp
- `updated_at` - Updated timestamp

**user_roles** (RBAC):
- `id` - Primary key
- `name` - Role name (unique)
- `description` - Role description
- `permissions` - JSON array of permissions
- `is_admin` - Admin flag
- `created_at` - Created timestamp

**user_activity_log**:
- `id` - Primary key
- `user_id` - Foreign key to users
- `action` - Action performed
- `entity_type` - Entity affected
- `entity_id` - Entity ID
- `details` - JSON action details
- `ip_address` - User IP
- `user_agent` - Browser/client
- `created_at` - Action timestamp

---

## 🌐 **API Endpoints**

### **Profile Management**:
```
GET    /api/v1/users/profile        # Get current user profile
PUT    /api/v1/users/profile        # Update profile
POST   /api/v1/users/avatar         # Upload avatar
DELETE /api/v1/users/avatar         # Delete avatar
GET    /api/v1/users/:id/profile    # Get user by ID (public)
```

### **Preferences**:
```
GET    /api/v1/users/preferences    # Get user preferences
PUT    /api/v1/users/preferences    # Update preferences
PATCH  /api/v1/users/preferences/:key  # Update single preference
```

### **Admin User Management**:
```
GET    /api/v1/admin/users          # List users (paginated)
GET    /api/v1/admin/users/:id      # Get user details
PUT    /api/v1/admin/users/:id      # Update user
DELETE /api/v1/admin/users/:id      # Delete user (soft delete)
POST   /api/v1/admin/users/:id/suspend    # Suspend user
POST   /api/v1/admin/users/:id/activate   # Activate user
PUT    /api/v1/admin/users/:id/role       # Change user role
```

### **Roles & Permissions**:
```
GET    /api/v1/roles                # List all roles
GET    /api/v1/roles/:id            # Get role details
POST   /api/v1/roles                # Create role (admin)
PUT    /api/v1/roles/:id            # Update role (admin)
DELETE /api/v1/roles/:id            # Delete role (admin)
GET    /api/v1/users/:id/permissions  # Get user permissions
```

### **Activity Tracking**:
```
GET    /api/v1/users/activity       # Get user activity log
GET    /api/v1/admin/activity       # Get all activity (admin)
```

---

## 🔗 **Integration with Authentication Braid**

### **Shared Concepts**:
- **Users table**: Extended from auth braid
- **JWT tokens**: Include role and permissions
- **Auth middleware**: Enhanced with RBAC checks
- **Session tracking**: Linked to user profiles

### **Dependencies**:
```
Authentication Braid → User Management Braid
├── User identity (user_id)
├── Email verification status
├── JWT token claims (role, permissions)
├── Session management
└── Auth middleware
```

---

## 📁 **File Structure**

### **Backend Routes** (`backend/internal/routes/`):
- `user.go` - User profile endpoints
- `admin.go` - Admin user management (user section)
- `roles.go` - RBAC role management

### **Backend Services** (`backend/internal/services/`):
- User profile business logic
- RBAC permission checking
- Activity logging

### **Backend Database** (`backend/internal/database/`):
- `user.go` - User operations
- Role and permission operations
- Activity log operations

### **Migrations** (`backend/migrations/`):
- `005_add_user_profile_fields.sql`
- User preferences table
- User roles table
- Activity log table

---

## 🔒 **RBAC System**

### **Built-in Roles**:
1. **user** - Standard user (default)
2. **admin** - Full system access
3. **moderator** - Content moderation
4. **support** - Customer support access
5. **analyst** - Analytics and reporting

### **Permission Categories**:
- `users:read` - View user profiles
- `users:write` - Edit users
- `users:delete` - Delete users
- `content:moderate` - Moderate content
- `analytics:view` - View analytics
- `system:admin` - System administration

### **Permission Checking**:
```go
// In middleware
func RequirePermission(permission string) middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := r.Context().Value("user").(*User)
            
            if !user.HasPermission(permission) {
                respondError(w, "Permission denied", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
router.Handle("/admin/users", RequirePermission("users:write"))
```

---

## 🎨 **User Profile Fields**

### **Basic Information**:
- First name, last name
- Email (from auth)
- Profile picture URL
- Bio/description
- Location (optional)
- Website (optional)
- Social media links

### **Account Information**:
- Role and permissions
- Account creation date
- Last login timestamp
- Account status (active/suspended)
- Email verification status

### **Preferences**:
- UI theme
- Language
- Timezone
- Notification settings
- Privacy settings

---

## 🔄 **Data Flow Examples**

### **Profile Update Flow**:
```
1. User edits profile in frontend
2. PUT /api/v1/users/profile
3. Validate input data
4. Check auth (JWT middleware)
5. Update users table
6. Log activity
7. Return updated profile
8. Frontend updates store
```

### **Admin User Suspension**:
```
1. Admin clicks "Suspend User"
2. POST /api/v1/admin/users/:id/suspend
3. Check admin permission
4. Update is_active = false
5. Invalidate user sessions
6. Log admin action
7. Send notification email
8. Return success
```

### **RBAC Permission Check**:
```
1. Request to protected endpoint
2. Auth middleware validates JWT
3. Extract user_id and role
4. RBAC middleware checks permission
5. Query user permissions from role
6. Allow or deny access
7. Log access attempt
8. Proceed or return 403
```

---

## 📊 **Performance Considerations**

### **Caching**:
- User profiles cached (5 min TTL)
- Role permissions cached (15 min TTL)
- Activity logs not cached

### **Database Indexes**:
- `users.email` (unique index)
- `users.role` (for role queries)
- `user_activity_log.user_id` (for activity queries)
- `user_activity_log.created_at` (for time-based queries)

### **Pagination**:
- User listing paginated (50 per page)
- Activity log paginated (100 per page)
- Search results paginated (20 per page)

---

## 🔐 **Security Measures**

### **Data Protection**:
- Profile updates require authentication
- Admin operations require admin role
- Sensitive fields (email) require additional verification
- GDPR compliance (right to be forgotten)

### **Access Control**:
- RBAC enforced at middleware level
- Permission checks before sensitive operations
- Audit log for all admin actions
- Session invalidation on role change

### **Privacy**:
- Optional profile visibility settings
- Email privacy controls
- Activity log retention (90 days)
- User data export capability

---

## 🧬 **Strands (Complete Flows)**

### **1. Profile Management Strand**:
Complete flow from profile edit form to database update

### **2. RBAC System Strand**:
Permission checking from request to database

### **3. User Preferences Strand**:
Settings management from UI to storage

### **4. Admin User Management Strand**:
Admin operations from dashboard to database

---

## 📝 **Known Technical Debt**

### **To Address**:
1. Profile pictures stored externally (Digital Ocean Spaces)
2. Activity log grows large (need archival strategy)
3. Role permissions could use more granularity
4. User search could be optimized with ElasticSearch
5. Profile validation could be more comprehensive

### **Future Enhancements**:
1. Two-factor authentication integration
2. Social login profile sync
3. Advanced privacy controls
4. User blocking/muting features
5. Profile verification badges

---

## 🔗 **Related Braids**

### **Depends On**:
- **Authentication Braid**: User identity, JWT tokens
- **Infrastructure Braid**: Database, file storage

### **Consumed By**:
- **Admin Dashboard Braid**: User management UI
- **Subscription Braid**: User billing info
- **Analytics Braid**: User behavior tracking
- **Content Management Braid**: User-generated content

---

## 📚 **Documentation Files**

### **Layer Documentation**:
- `layers/persistence/schema/` - Database schema docs
- `layers/data-access/` - Go model operations
- `layers/business-logic/` - API handlers and services
- `layers/application/` - API contracts

### **Strand Documentation**:
- `strands/profile-management/` - Profile CRUD flow
- `strands/rbac-system/` - Permission checking flow
- `strands/user-preferences/` - Settings management
- `strands/admin-user-management/` - Admin operations

### **Elastic Bands**:
- `layers/*/ELASTIC-BAND-UP.md` - Upward contracts
- `layers/*/ELASTIC-BAND-DOWN.md` - Downward contracts

---

## 🚀 **Quick Start**

### **Understanding User Management**:
1. Read this BRAID.md (15 min)
2. Check `strands/profile-management/STRAND.md` (15 min)
3. Review RBAC system in `strands/rbac-system/STRAND.md` (15 min)
4. **Total: 45 minutes to understand user management!**

### **Working on Features**:
1. Identify which strand (profile, RBAC, preferences, admin)
2. Read relevant strand document
3. Check elastic bands for layer contracts
4. Follow existing patterns
5. Update documentation

---

**Last Updated**: October 14, 2025  
**Status**: Complete backend documentation  
**Technology**: Go + PostgreSQL  
**Frontend Counterpart**: `_frontend/braids/user-management/`

---

**Navigate**:  
[🏠 Master Index](../../BRAIDS_INDEX.md) | [🎨 Frontend Braid](../../_frontend/braids/user-management/BRAID.md) | [🔗 Auth Braid](../authentication/BRAID.md)

