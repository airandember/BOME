package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// StandardizedRole represents a user role in the standardized system
type StandardizedRole struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	Description     string   `json:"description"`
	Category        string   `json:"category"`
	Level           int      `json:"level"`
	Permissions     []string `json:"permissions"`
	IsSystemRole    bool     `json:"isSystemRole"`
	Color           string   `json:"color"`
	Icon            string   `json:"icon"`
	SubsystemAccess []string `json:"subsystemAccess"` // Which subsystems this role can access
	CreatedAt       string   `json:"createdAt"`
	UpdatedAt       string   `json:"updatedAt"`
}

// StandardizedPermission represents a system permission
type StandardizedPermission struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Subsystem   string `json:"subsystem"` // Which subsystem this permission belongs to
}

// STANDARDIZED PERMISSIONS DATA
var STANDARDIZED_PERMISSIONS = []StandardizedPermission{
	// System Administration Permissions (Level 10-9)
	{ID: "system:full_access", Resource: "system", Action: "full_access", Description: "Full system access and control", Category: "system", Subsystem: "all"},
	{ID: "system:manage", Resource: "system", Action: "manage", Description: "System management without role changes", Category: "system", Subsystem: "all"},
	{ID: "system:read", Resource: "system", Action: "read", Description: "View system information", Category: "system", Subsystem: "all"},
	{ID: "system:update", Resource: "system", Action: "update", Description: "Modify system settings", Category: "system", Subsystem: "all"},

	// Role & Permission Management
	{ID: "roles:create", Resource: "roles", Action: "create", Description: "Create new roles", Category: "core", Subsystem: "hub"},
	{ID: "roles:read", Resource: "roles", Action: "read", Description: "View role information", Category: "core", Subsystem: "hub"},
	{ID: "roles:update", Resource: "roles", Action: "update", Description: "Edit existing roles", Category: "core", Subsystem: "hub"},
	{ID: "roles:delete", Resource: "roles", Action: "delete", Description: "Delete roles", Category: "core", Subsystem: "hub"},
	{ID: "permissions:manage", Resource: "permissions", Action: "manage", Description: "Manage user permissions", Category: "core", Subsystem: "hub"},

	// User Management Permissions
	{ID: "users:create", Resource: "users", Action: "create", Description: "Create new user accounts", Category: "core", Subsystem: "hub"},
	{ID: "users:read", Resource: "users", Action: "read", Description: "View user account information", Category: "core", Subsystem: "hub"},
	{ID: "users:update", Resource: "users", Action: "update", Description: "Edit user account details", Category: "core", Subsystem: "hub"},
	{ID: "users:delete", Resource: "users", Action: "delete", Description: "Delete user accounts", Category: "core", Subsystem: "hub"},
	{ID: "users:manage", Resource: "users", Action: "manage", Description: "Full user account management", Category: "core", Subsystem: "hub"},

	// Content Management Permissions (Level 8-6)
	{ID: "content:create", Resource: "content", Action: "create", Description: "Create new content", Category: "content", Subsystem: "all"},
	{ID: "content:read", Resource: "content", Action: "read", Description: "View content", Category: "content", Subsystem: "all"},
	{ID: "content:update", Resource: "content", Action: "update", Description: "Edit existing content", Category: "content", Subsystem: "all"},
	{ID: "content:delete", Resource: "content", Action: "delete", Description: "Delete content", Category: "content", Subsystem: "all"},
	{ID: "content:publish", Resource: "content", Action: "publish", Description: "Publish content to live site", Category: "content", Subsystem: "all"},
	{ID: "content:moderate", Resource: "content", Action: "moderate", Description: "Moderate user-generated content", Category: "content", Subsystem: "all"},

	// Video Management Permissions
	{ID: "videos:create", Resource: "videos", Action: "create", Description: "Upload new videos", Category: "content", Subsystem: "youtube,streaming"},
	{ID: "videos:read", Resource: "videos", Action: "read", Description: "View video content and metadata", Category: "content", Subsystem: "youtube,streaming"},
	{ID: "videos:update", Resource: "videos", Action: "update", Description: "Edit video metadata and settings", Category: "content", Subsystem: "youtube,streaming"},
	{ID: "videos:delete", Resource: "videos", Action: "delete", Description: "Delete video content", Category: "content", Subsystem: "youtube,streaming"},
	{ID: "videos:manage", Resource: "videos", Action: "manage", Description: "Full video management capabilities", Category: "content", Subsystem: "youtube,streaming"},

	// Articles Management Permissions
	{ID: "articles:create", Resource: "articles", Action: "create", Description: "Create new articles", Category: "content", Subsystem: "articles"},
	{ID: "articles:read", Resource: "articles", Action: "read", Description: "View articles", Category: "content", Subsystem: "articles"},
	{ID: "articles:update", Resource: "articles", Action: "update", Description: "Edit existing articles", Category: "content", Subsystem: "articles"},
	{ID: "articles:delete", Resource: "articles", Action: "delete", Description: "Delete articles", Category: "content", Subsystem: "articles"},
	{ID: "articles:publish", Resource: "articles", Action: "publish", Description: "Publish articles", Category: "content", Subsystem: "articles"},
	{ID: "articles:manage", Resource: "articles", Action: "manage", Description: "Full articles management", Category: "content", Subsystem: "articles"},

	// Events Management Permissions
	{ID: "events:create", Resource: "events", Action: "create", Description: "Create new events", Category: "events", Subsystem: "events"},
	{ID: "events:read", Resource: "events", Action: "read", Description: "View event information", Category: "events", Subsystem: "events"},
	{ID: "events:update", Resource: "events", Action: "update", Description: "Edit event details", Category: "events", Subsystem: "events"},
	{ID: "events:delete", Resource: "events", Action: "delete", Description: "Delete events", Category: "events", Subsystem: "events"},
	{ID: "events:manage", Resource: "events", Action: "manage", Description: "Full event management", Category: "events", Subsystem: "events"},

	// Advertisement Management Permissions
	{ID: "advertisements:create", Resource: "advertisements", Action: "create", Description: "Create advertising campaigns", Category: "marketing", Subsystem: "hub"},
	{ID: "advertisements:read", Resource: "advertisements", Action: "read", Description: "View advertising data", Category: "marketing", Subsystem: "hub"},
	{ID: "advertisements:update", Resource: "advertisements", Action: "update", Description: "Edit advertising campaigns", Category: "marketing", Subsystem: "hub"},
	{ID: "advertisements:delete", Resource: "advertisements", Action: "delete", Description: "Delete advertising campaigns", Category: "marketing", Subsystem: "hub"},
	{ID: "advertisements:manage", Resource: "advertisements", Action: "manage", Description: "Full advertisement management", Category: "marketing", Subsystem: "hub"},
	{ID: "advertisements:approve", Resource: "advertisements", Action: "approve", Description: "Approve advertising campaigns", Category: "marketing", Subsystem: "hub"},

	// Analytics Permissions
	{ID: "analytics:read", Resource: "analytics", Action: "read", Description: "View analytics and reports", Category: "core", Subsystem: "all"},
	{ID: "analytics:export", Resource: "analytics", Action: "export", Description: "Export analytics data", Category: "core", Subsystem: "all"},
	{ID: "analytics:manage", Resource: "analytics", Action: "manage", Description: "Configure analytics settings", Category: "core", Subsystem: "all"},

	// Financial Permissions
	{ID: "financial:read", Resource: "financial", Action: "read", Description: "View financial data", Category: "financial", Subsystem: "hub"},
	{ID: "financial:manage", Resource: "financial", Action: "manage", Description: "Manage financial operations", Category: "financial", Subsystem: "hub"},
	{ID: "financial:refund", Resource: "financial", Action: "refund", Description: "Process refunds", Category: "financial", Subsystem: "hub"},

	// Security Permissions
	{ID: "security:read", Resource: "security", Action: "read", Description: "View security logs", Category: "security", Subsystem: "all"},
	{ID: "security:manage", Resource: "security", Action: "manage", Description: "Manage security settings", Category: "security", Subsystem: "all"},
	{ID: "security:incident", Resource: "security", Action: "incident", Description: "Handle security incidents", Category: "security", Subsystem: "all"},

	// Technical Permissions
	{ID: "technical:read", Resource: "technical", Action: "read", Description: "View technical information", Category: "technical", Subsystem: "all"},
	{ID: "technical:manage", Resource: "technical", Action: "manage", Description: "Manage technical systems", Category: "technical", Subsystem: "all"},
	{ID: "technical:support", Resource: "technical", Action: "support", Description: "Provide technical support", Category: "technical", Subsystem: "all"},

	// Academic Permissions
	{ID: "academic:review", Resource: "academic", Action: "review", Description: "Review scholarly content", Category: "academic", Subsystem: "articles"},
	{ID: "academic:coordinate", Resource: "academic", Action: "coordinate", Description: "Coordinate research activities", Category: "academic", Subsystem: "articles"},
	{ID: "academic:manage", Resource: "academic", Action: "manage", Description: "Manage academic content", Category: "academic", Subsystem: "articles"},
}

// STANDARDIZED ROLES DATA
var STANDARDIZED_ROLES = []StandardizedRole{
	// System Administration (Level 10-9)
	{
		ID:              "super_admin",
		Name:            "Super Administrator",
		Slug:            "super-administrator",
		Description:     "Full system access and role management capabilities",
		Category:        "system",
		Level:           10,
		Permissions:     getAllStandardizedPermissionIDs(),
		IsSystemRole:    true,
		Color:           "#dc2626",
		Icon:            "crown",
		SubsystemAccess: []string{"hub", "articles", "youtube", "streaming", "events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "system_admin",
		Name:            "System Administrator",
		Slug:            "system-administrator",
		Description:     "Technical system management without role changes",
		Category:        "system",
		Level:           9,
		Permissions:     []string{"system:read", "system:update", "system:manage", "security:read", "security:manage", "technical:read", "technical:manage", "analytics:read", "analytics:export"},
		IsSystemRole:    true,
		Color:           "#7c3aed",
		Icon:            "server",
		SubsystemAccess: []string{"hub", "articles", "youtube", "streaming", "events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Content Management (Level 8-6)
	{
		ID:              "content_manager",
		Name:            "Content Manager",
		Slug:            "content-manager",
		Description:     "Overall content strategy and oversight",
		Category:        "content",
		Level:           8,
		Permissions:     []string{"content:create", "content:read", "content:update", "content:delete", "content:publish", "content:moderate", "videos:create", "videos:read", "videos:update", "videos:delete", "videos:manage", "articles:create", "articles:read", "articles:update", "articles:delete", "articles:publish", "analytics:read", "analytics:export"},
		IsSystemRole:    true,
		Color:           "#059669",
		Icon:            "document-text",
		SubsystemAccess: []string{"articles", "youtube", "streaming"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "content_editor",
		Name:            "Content Editor",
		Slug:            "content-editor",
		Description:     "Review, approve, edit, and publish content",
		Category:        "content",
		Level:           7,
		Permissions:     []string{"content:read", "content:update", "content:publish", "videos:read", "videos:update", "articles:read", "articles:update", "articles:publish", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#059669",
		Icon:            "pencil",
		SubsystemAccess: []string{"articles", "youtube", "streaming"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "content_creator",
		Name:            "Content Creator",
		Slug:            "content-creator",
		Description:     "Create and edit content with limited publishing",
		Category:        "content",
		Level:           6,
		Permissions:     []string{"content:create", "content:read", "content:update", "videos:create", "videos:read", "videos:update", "articles:create", "articles:read", "articles:update"},
		IsSystemRole:    false,
		Color:           "#059669",
		Icon:            "plus-circle",
		SubsystemAccess: []string{"articles", "youtube", "streaming"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Subsystem-Specific Roles (Level 7)
	{
		ID:              "articles_manager",
		Name:            "Articles Manager",
		Slug:            "articles-manager",
		Description:     "Full articles subsystem management",
		Category:        "subsystem",
		Level:           7,
		Permissions:     []string{"articles:create", "articles:read", "articles:update", "articles:delete", "articles:publish", "articles:manage", "content:read", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#1e40af",
		Icon:            "document",
		SubsystemAccess: []string{"articles"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "youtube_manager",
		Name:            "YouTube Manager",
		Slug:            "youtube-manager",
		Description:     "YouTube system management",
		Category:        "subsystem",
		Level:           7,
		Permissions:     []string{"videos:create", "videos:read", "videos:update", "videos:delete", "videos:manage", "content:read", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#dc2626",
		Icon:            "play",
		SubsystemAccess: []string{"youtube"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "streaming_manager",
		Name:            "Video Streaming Manager",
		Slug:            "streaming-manager",
		Description:     "Bunny.net streaming platform management",
		Category:        "subsystem",
		Level:           7,
		Permissions:     []string{"videos:create", "videos:read", "videos:update", "videos:delete", "videos:manage", "content:read", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#7c3aed",
		Icon:            "video",
		SubsystemAccess: []string{"streaming"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "events_manager",
		Name:            "Events Manager",
		Slug:            "events-manager",
		Description:     "Events system management",
		Category:        "subsystem",
		Level:           7,
		Permissions:     []string{"events:create", "events:read", "events:update", "events:delete", "events:manage", "users:read", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#2563eb",
		Icon:            "calendar",
		SubsystemAccess: []string{"events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Marketing & Advertising (Level 7-4)
	{
		ID:              "advertisement_manager",
		Name:            "Advertisement Manager",
		Slug:            "advertisement-manager",
		Description:     "Full advertisement system oversight",
		Category:        "marketing",
		Level:           7,
		Permissions:     []string{"advertisements:create", "advertisements:read", "advertisements:update", "advertisements:delete", "advertisements:manage", "advertisements:approve", "analytics:read", "financial:read"},
		IsSystemRole:    false,
		Color:           "#f59e0b",
		Icon:            "presentation-chart-line",
		SubsystemAccess: []string{"hub"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "marketing_specialist",
		Name:            "Marketing Specialist",
		Slug:            "marketing-specialist",
		Description:     "Campaign creation and advertiser relations",
		Category:        "marketing",
		Level:           4,
		Permissions:     []string{"advertisements:create", "advertisements:read", "advertisements:update", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#f59e0b",
		Icon:            "megaphone",
		SubsystemAccess: []string{"hub"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// User Management (Level 7-5)
	{
		ID:              "user_manager",
		Name:            "User Account Manager",
		Slug:            "user-manager",
		Description:     "User management and support operations",
		Category:        "user_management",
		Level:           7,
		Permissions:     []string{"users:create", "users:read", "users:update", "users:delete", "users:manage", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#2563eb",
		Icon:            "users",
		SubsystemAccess: []string{"hub"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "support_specialist",
		Name:            "Support Specialist",
		Slug:            "support-specialist",
		Description:     "User support and basic account management",
		Category:        "user_management",
		Level:           5,
		Permissions:     []string{"users:read", "users:update", "technical:support"},
		IsSystemRole:    false,
		Color:           "#2563eb",
		Icon:            "life-buoy",
		SubsystemAccess: []string{"hub"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Analytics & Financial (Level 7)
	{
		ID:              "analytics_manager",
		Name:            "Analytics Manager",
		Slug:            "analytics-manager",
		Description:     "Data analysis and reporting across all systems",
		Category:        "analytics",
		Level:           7,
		Permissions:     []string{"analytics:read", "analytics:export", "analytics:manage"},
		IsSystemRole:    false,
		Color:           "#059669",
		Icon:            "chart-bar",
		SubsystemAccess: []string{"hub", "articles", "youtube", "streaming", "events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "financial_admin",
		Name:            "Financial Administrator",
		Slug:            "financial-administrator",
		Description:     "Revenue, billing, and financial reporting",
		Category:        "financial",
		Level:           7,
		Permissions:     []string{"financial:read", "financial:manage", "financial:refund", "analytics:read"},
		IsSystemRole:    false,
		Color:           "#059669",
		Icon:            "credit-card",
		SubsystemAccess: []string{"hub"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Technical & Security (Level 6-5)
	{
		ID:              "security_admin",
		Name:            "Security Administrator",
		Slug:            "security-administrator",
		Description:     "Security monitoring and incident response",
		Category:        "security",
		Level:           6,
		Permissions:     []string{"security:read", "security:manage", "security:incident"},
		IsSystemRole:    false,
		Color:           "#dc2626",
		Icon:            "shield",
		SubsystemAccess: []string{"hub", "articles", "youtube", "streaming", "events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "technical_specialist",
		Name:            "Technical Specialist",
		Slug:            "technical-specialist",
		Description:     "Technical support and maintenance",
		Category:        "technical",
		Level:           5,
		Permissions:     []string{"technical:read", "technical:support"},
		IsSystemRole:    false,
		Color:           "#7c3aed",
		Icon:            "wrench",
		SubsystemAccess: []string{"hub", "articles", "youtube", "streaming", "events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Academic & Research (Level 6-5)
	{
		ID:              "academic_reviewer",
		Name:            "Academic Reviewer",
		Slug:            "academic-reviewer",
		Description:     "Review scholarly content for accuracy and quality",
		Category:        "academic",
		Level:           6,
		Permissions:     []string{"academic:review", "articles:read", "articles:update"},
		IsSystemRole:    false,
		Color:           "#7c2d12",
		Icon:            "academic-cap",
		SubsystemAccess: []string{"articles"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "research_coordinator",
		Name:            "Research Coordinator",
		Slug:            "research-coordinator",
		Description:     "Coordinate academic research and citations",
		Category:        "academic",
		Level:           5,
		Permissions:     []string{"academic:coordinate", "articles:read", "articles:update"},
		IsSystemRole:    false,
		Color:           "#7c2d12",
		Icon:            "book-open",
		SubsystemAccess: []string{"articles"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},

	// Base User Roles (Level 3-1)
	{
		ID:              "advertiser",
		Name:            "Advertiser",
		Slug:            "advertiser",
		Description:     "Create and manage advertising campaigns",
		Category:        "base",
		Level:           3,
		Permissions:     []string{"advertisements:create", "advertisements:read", "advertisements:update"},
		IsSystemRole:    false,
		Color:           "#f59e0b",
		Icon:            "megaphone",
		SubsystemAccess: []string{"hub"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
	{
		ID:              "user",
		Name:            "User",
		Slug:            "user",
		Description:     "Basic platform access",
		Category:        "base",
		Level:           1,
		Permissions:     []string{"content:read"},
		IsSystemRole:    true,
		Color:           "#6b7280",
		Icon:            "user",
		SubsystemAccess: []string{"hub", "articles", "youtube", "streaming", "events"},
		CreatedAt:       "2024-01-01T00:00:00Z",
		UpdatedAt:       "2024-01-01T00:00:00Z",
	},
}

// Helper function to get all standardized permission IDs
func getAllStandardizedPermissionIDs() []string {
	var ids []string
	for _, perm := range STANDARDIZED_PERMISSIONS {
		ids = append(ids, perm.ID)
	}
	return ids
}

// GetStandardizedRolesHandler returns all standardized roles
func GetStandardizedRolesHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"roles": STANDARDIZED_ROLES,
		"total": len(STANDARDIZED_ROLES),
	})
}

// GetStandardizedPermissionsHandler returns all standardized permissions
func GetStandardizedPermissionsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"permissions": STANDARDIZED_PERMISSIONS,
		"total":       len(STANDARDIZED_PERMISSIONS),
	})
}

// GetRoleByID returns a specific role by ID
func GetRoleByID(roleID string) *StandardizedRole {
	for _, role := range STANDARDIZED_ROLES {
		if role.ID == roleID {
			return &role
		}
	}
	return nil
}

// GetPermissionByID returns a specific permission by ID
func GetPermissionByID(permissionID string) *StandardizedPermission {
	for _, perm := range STANDARDIZED_PERMISSIONS {
		if perm.ID == permissionID {
			return &perm
		}
	}
	return nil
}

// HasPermission checks if a role has a specific permission
func HasPermission(roleID string, permissionID string) bool {
	role := GetRoleByID(roleID)
	if role == nil {
		return false
	}

	for _, perm := range role.Permissions {
		if perm == permissionID {
			return true
		}
	}
	return false
}

// GetRolesBySubsystem returns all roles that have access to a specific subsystem
func GetRolesBySubsystem(subsystem string) []StandardizedRole {
	var roles []StandardizedRole
	for _, role := range STANDARDIZED_ROLES {
		for _, access := range role.SubsystemAccess {
			if access == subsystem {
				roles = append(roles, role)
				break
			}
		}
	}
	return roles
}

// GetPermissionsBySubsystem returns all permissions for a specific subsystem
func GetPermissionsBySubsystem(subsystem string) []StandardizedPermission {
	var permissions []StandardizedPermission
	for _, perm := range STANDARDIZED_PERMISSIONS {
		if perm.Subsystem == subsystem || perm.Subsystem == "all" {
			permissions = append(permissions, perm)
		}
	}
	return permissions
}

// SetupStandardizedRolesRoutes sets up all standardized roles routes
func SetupStandardizedRolesRoutes(router *gin.RouterGroup) {
	api := router.Group("/standardized")
	{
		// Standardized roles endpoints
		api.GET("/roles", GetStandardizedRolesHandler)
		api.GET("/permissions", GetStandardizedPermissionsHandler)
		api.GET("/roles/:id", func(c *gin.Context) {
			roleID := c.Param("id")
			role := GetRoleByID(roleID)
			if role == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"role": role})
		})
		api.GET("/permissions/:id", func(c *gin.Context) {
			permissionID := c.Param("id")
			permission := GetPermissionByID(permissionID)
			if permission == nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "Permission not found"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"permission": permission})
		})
		api.GET("/subsystem/:subsystem/roles", func(c *gin.Context) {
			subsystem := c.Param("subsystem")
			roles := GetRolesBySubsystem(subsystem)
			c.JSON(http.StatusOK, gin.H{
				"roles":     roles,
				"subsystem": subsystem,
				"total":     len(roles),
			})
		})
		api.GET("/subsystem/:subsystem/permissions", func(c *gin.Context) {
			subsystem := c.Param("subsystem")
			permissions := GetPermissionsBySubsystem(subsystem)
			c.JSON(http.StatusOK, gin.H{
				"permissions": permissions,
				"subsystem":   subsystem,
				"total":       len(permissions),
			})
		})
		api.POST("/check-permission", func(c *gin.Context) {
			var req struct {
				RoleID       string `json:"roleId"`
				PermissionID string `json:"permissionId"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			hasPermission := HasPermission(req.RoleID, req.PermissionID)
			c.JSON(http.StatusOK, gin.H{
				"roleId":        req.RoleID,
				"permissionId":  req.PermissionID,
				"hasPermission": hasPermission,
			})
		})
	}
}
