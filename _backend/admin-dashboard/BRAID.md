# 🧬 Admin Dashboard Braid - Backend
**Comprehensive administrative interface and system management**

---

## 📋 **Backend Overview**

**Purpose**: Backend support for admin operations, user management, system monitoring, and business operations  
**Technology**: Go, PostgreSQL, RBAC  
**Complexity**: **Very High** (15+ subsystems, Complex UI, Multi-role Access)  

**Critical Files**:
- `backend/internal/routes/admin.go`
- `backend/internal/routes/admin_streaming.go`
- `backend/internal/routes/analytics.go`
- `backend/internal/routes/database_monitoring.go`
- `backend/internal/services/admin_cache.go`

---

## 🎯 **Admin Subsystems** (15+ Areas)

### **1. User Management**:
- View all users
- Edit user details
- Assign/revoke roles
- Suspend/activate accounts
- Reset passwords
- View user activity

### **2. Video Administration**:
- Upload videos
- Edit video metadata
- Approve/reject videos
- Manage categories
- Bulk operations

### **3. Subscription Management**:
- View all subscriptions
- Manage subscription plans
- Process refunds
- Handle cancellations
- Revenue tracking

### **4. Advertisement Management**:
- Approve/reject ads
- Manage advertiser accounts
- View ad performance
- Set pricing

### **5. Analytics & Reporting**:
- Dashboard overview
- User analytics
- Video analytics
- Revenue reports
- Custom reports

### **6. System Monitoring**:
- Health checks
- Performance metrics
- Error logs
- Database monitoring
- API usage

### **7. Stripe Integration**:
- Webhook logs
- Payment tracking
- Subscription sync
- Refund processing

### **8. Content Moderation**:
- Review flagged content
- User reports
- Comment moderation

### **9. Role Management**:
**File**: `backend/internal/routes/roles.go`
- Create/edit roles
- Assign permissions
- RBAC enforcement

### **10. Security & Audit**:
- Audit logs
- Security incidents
- Access control
- Failed login attempts

---

## 🗄️ **Admin-Specific Schema**

### **Admin Roles Table**:
```sql
CREATE TABLE admin_roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB NOT NULL,
    is_system_role BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);

COMMENT ON TABLE admin_roles IS 'Admin role definitions and permissions';
```

### **Admin Activity Log**:
```sql
CREATE TABLE admin_activity_log (
    id SERIAL PRIMARY KEY,
    admin_id INTEGER REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id INTEGER,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_admin_activity_admin_id ON admin_activity_log(admin_id);
CREATE INDEX idx_admin_activity_created_at ON admin_activity_log(created_at);
```

---

## 🌐 **Admin API Endpoints**

### **User Administration**:
```
GET    /api/v1/admin/users              # List all users
GET    /api/v1/admin/users/:id          # Get user details
PUT    /api/v1/admin/users/:id          # Update user
DELETE /api/v1/admin/users/:id          # Delete user
POST   /api/v1/admin/users/:id/suspend  # Suspend user
POST   /api/v1/admin/users/:id/activate # Activate user
```

### **Video Administration**:
```
GET    /api/v1/admin/videos             # List all videos
POST   /api/v1/admin/videos             # Create video
PUT    /api/v1/admin/videos/:id         # Update video
DELETE /api/v1/admin/videos/:id         # Delete video
POST   /api/v1/admin/videos/:id/approve # Approve video
POST   /api/v1/admin/videos/:id/reject  # Reject video
```

### **Subscription Administration**:
```
GET    /api/v1/admin/subscriptions      # List all subscriptions
POST   /api/v1/admin/subscriptions/:id/cancel # Cancel subscription
POST   /api/v1/admin/subscriptions/:id/refund # Issue refund
```

### **Analytics**:
**File**: `backend/internal/routes/analytics.go`
```
GET    /api/v1/admin/analytics          # Dashboard analytics
GET    /api/v1/admin/analytics/users    # User analytics
GET    /api/v1/admin/analytics/videos   # Video analytics
GET    /api/v1/admin/analytics/revenue  # Revenue analytics
```

### **System Monitoring**:
**File**: `backend/internal/routes/database_monitoring.go`
```
GET    /api/v1/admin/system/health      # System health
GET    /api/v1/admin/system/metrics     # Performance metrics
GET    /api/v1/admin/system/errors      # Error logs
GET    /api/v1/admin/database/stats     # Database statistics
```

---

## 🔧 **Backend Services**

### **Admin Cache Service** (`backend/internal/services/admin_cache.go`):
```go
// Performance optimization for admin dashboard
func GetDashboardMetricsOptimized() (*Metrics, error) {
    cached := cache.Get("admin_dashboard_metrics")
    if cached != nil {
        return cached, nil
    }
    
    metrics := CalculateDashboardMetrics()
    cache.Set("admin_dashboard_metrics", metrics, 5*time.Minute)
    
    return metrics, nil
}
```

---

## 🔒 **RBAC (Role-Based Access Control)**

### **Permission System**:
```go
type Permission string

const (
    ViewUsers     Permission = "view_users"
    EditUsers     Permission = "edit_users"
    DeleteUsers   Permission = "delete_users"
    ViewVideos    Permission = "view_videos"
    EditVideos    Permission = "edit_videos"
    ViewAnalytics Permission = "view_analytics"
    // ... many more
)

func CheckPermission(userID int, permission Permission) (bool, error) {
    user := GetUser(userID)
    role := GetUserRole(user.RoleID)
    
    return role.HasPermission(permission), nil
}
```

### **Admin Middleware**:
```go
func AdminOnly(c *fiber.Ctx) error {
    user := GetCurrentUser(c)
    
    if !user.IsAdmin() {
        return c.Status(403).JSON(fiber.Map{
            "error": "Admin access required"
        })
    }
    
    return c.Next()
}

func RequirePermission(permission Permission) fiber.Handler {
    return func(c *fiber.Ctx) error {
        user := GetCurrentUser(c)
        
        hasPermission, _ := CheckPermission(user.ID, permission)
        if !hasPermission {
            return c.Status(403).JSON(fiber.Map{
                "error": "Insufficient permissions"
            })
        }
        
        return c.Next()
    }
}
```

---

## 📊 **Dashboard Metrics**

### **Key Metrics**:
```go
type DashboardMetrics struct {
    // User Metrics
    TotalUsers       int
    ActiveUsers      int // Last 30 days
    NewUsersToday    int
    
    // Content Metrics
    TotalVideos      int
    PendingVideos    int
    TotalWatchTime   int64 // seconds
    
    // Revenue Metrics
    TotalRevenue     float64
    MonthlyRevenue   float64
    ActiveSubscriptions int
    
    // System Metrics
    SystemUptime     time.Duration
    APIResponseTime  int // milliseconds
    ErrorRate        float64 // percentage
}
```

---

## 🚀 **Admin Operations**

### **Bulk Operations**:
```go
func BulkUpdateVideos(videoIDs []int, updates VideoUpdate) error {
    tx := db.Begin()
    
    for _, id := range videoIDs {
        err := UpdateVideo(id, updates)
        if err != nil {
            tx.Rollback()
            return err
        }
    }
    
    tx.Commit()
    return nil
}
```

### **Data Export**:
```go
func ExportUserData(format string) ([]byte, error) {
    users := GetAllUsers()
    
    switch format {
    case "csv":
        return ExportToCSV(users)
    case "json":
        return json.Marshal(users)
    case "excel":
        return ExportToExcel(users)
    }
}
```

---

## 📝 **Audit Logging**

### **Log Admin Actions**:
```go
func LogAdminAction(adminID int, action string, details map[string]interface{}) {
    log := AdminActivityLog{
        AdminID:    adminID,
        Action:     action,
        Details:    details,
        IPAddress:  GetClientIP(),
        UserAgent:  GetUserAgent(),
        CreatedAt:  time.Now(),
    }
    
    db.Create(&log)
}

// Usage
LogAdminAction(adminID, "DELETE_USER", map[string]interface{}{
    "user_id": deletedUserID,
    "reason": "Terms violation"
})
```

---

## 🎯 **Admin Subsystem Files**

### **Key Routes**:
- `backend/internal/routes/admin.go` - Core admin operations
- `backend/internal/routes/admin_streaming.go` - Video/streaming admin
- `backend/internal/routes/analytics.go` - Analytics endpoints
- `backend/internal/routes/database_monitoring.go` - System monitoring
- `backend/internal/routes/roles.go` - Role management

---

## 📝 **Known Issues**

### **To Implement**:
1. Advanced search filters
2. Automated reports scheduling
3. Two-factor authentication for admins
4. IP whitelisting for admin access
5. Activity dashboard real-time updates
6. Advanced audit trail analysis

---

**Last Updated**: October 14, 2025  
**Status**: Largest braid system  
**Frontend**: `_frontend/braids/admin-dashboard/`

