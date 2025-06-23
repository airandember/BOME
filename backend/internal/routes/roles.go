package routes

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Role represents a user role
type Role struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Slug         string   `json:"slug"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Level        int      `json:"level"`
	Permissions  []string `json:"permissions"`
	IsSystemRole bool     `json:"isSystemRole"`
	Color        string   `json:"color"`
	Icon         string   `json:"icon"`
	CreatedAt    string   `json:"createdAt"`
	UpdatedAt    string   `json:"updatedAt"`
}

// Permission represents a system permission
type Permission struct {
	ID          string `json:"id"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// UserWithRoles represents a user with role assignments
type UserWithRoles struct {
	ID        int      `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Status    string   `json:"status"`
	Roles     []Role   `json:"roles"`
	RoleNames []string `json:"roleNames"`
	LastLogin string   `json:"lastLogin"`
	CreatedAt string   `json:"createdAt"`
}

// MOCK PERMISSIONS DATA
var MOCK_PERMISSIONS_DATA = []Permission{
	// User Management Permissions
	{ID: "users:create", Resource: "users", Action: "create", Description: "Create new user accounts", Category: "core"},
	{ID: "users:read", Resource: "users", Action: "read", Description: "View user account information", Category: "core"},
	{ID: "users:update", Resource: "users", Action: "update", Description: "Edit user account details", Category: "core"},
	{ID: "users:delete", Resource: "users", Action: "delete", Description: "Delete user accounts", Category: "core"},
	{ID: "users:manage", Resource: "users", Action: "manage", Description: "Full user account management", Category: "core"},

	// Content Management Permissions
	{ID: "content:create", Resource: "content", Action: "create", Description: "Create new content", Category: "content"},
	{ID: "content:read", Resource: "content", Action: "read", Description: "View content", Category: "content"},
	{ID: "content:update", Resource: "content", Action: "update", Description: "Edit existing content", Category: "content"},
	{ID: "content:delete", Resource: "content", Action: "delete", Description: "Delete content", Category: "content"},
	{ID: "content:publish", Resource: "content", Action: "publish", Description: "Publish content to live site", Category: "content"},
	{ID: "content:moderate", Resource: "content", Action: "moderate", Description: "Moderate user-generated content", Category: "content"},

	// Video Management Permissions
	{ID: "videos:create", Resource: "videos", Action: "create", Description: "Upload new videos", Category: "content"},
	{ID: "videos:read", Resource: "videos", Action: "read", Description: "View video content and metadata", Category: "content"},
	{ID: "videos:update", Resource: "videos", Action: "update", Description: "Edit video metadata and settings", Category: "content"},
	{ID: "videos:delete", Resource: "videos", Action: "delete", Description: "Delete video content", Category: "content"},
	{ID: "videos:manage", Resource: "videos", Action: "manage", Description: "Full video management capabilities", Category: "content"},

	// Analytics Permissions
	{ID: "analytics:read", Resource: "analytics", Action: "read", Description: "View analytics and reports", Category: "core"},
	{ID: "analytics:export", Resource: "analytics", Action: "export", Description: "Export analytics data", Category: "core"},
	{ID: "analytics:manage", Resource: "analytics", Action: "manage", Description: "Configure analytics settings", Category: "core"},

	// System Administration Permissions
	{ID: "system:read", Resource: "system", Action: "read", Description: "View system information", Category: "technical"},
	{ID: "system:update", Resource: "system", Action: "update", Description: "Modify system settings", Category: "technical"},
	{ID: "system:manage", Resource: "system", Action: "manage", Description: "Full system administration", Category: "technical"},

	// Role & Permission Management
	{ID: "roles:create", Resource: "roles", Action: "create", Description: "Create new roles", Category: "core"},
	{ID: "roles:read", Resource: "roles", Action: "read", Description: "View role information", Category: "core"},
	{ID: "roles:update", Resource: "roles", Action: "update", Description: "Edit existing roles", Category: "core"},
	{ID: "roles:delete", Resource: "roles", Action: "delete", Description: "Delete roles", Category: "core"},
	{ID: "permissions:manage", Resource: "permissions", Action: "manage", Description: "Manage user permissions", Category: "core"},
}

// MOCK ROLES DATA
var MOCK_ROLES_DATA = []Role{
	{
		ID:           "super-administrator",
		Name:         "Super Administrator",
		Slug:         "super-administrator",
		Description:  "Full system access and role management capabilities",
		Category:     "core",
		Level:        10,
		Permissions:  getAllPermissionIDs(),
		IsSystemRole: true,
		Color:        "#dc2626",
		Icon:         "crown",
		CreatedAt:    "2024-01-01T00:00:00Z",
		UpdatedAt:    "2024-01-01T00:00:00Z",
	},
	{
		ID:           "system-administrator",
		Name:         "System Administrator",
		Slug:         "system-administrator",
		Description:  "Technical system management without role changes",
		Category:     "core",
		Level:        9,
		Permissions:  []string{"system:read", "system:update", "system:manage", "analytics:read", "analytics:export"},
		IsSystemRole: true,
		Color:        "#7c3aed",
		Icon:         "server",
		CreatedAt:    "2024-01-01T00:00:00Z",
		UpdatedAt:    "2024-01-01T00:00:00Z",
	},
	{
		ID:           "content-manager",
		Name:         "Content Manager",
		Slug:         "content-manager",
		Description:  "Overall content strategy and oversight",
		Category:     "core",
		Level:        8,
		Permissions:  []string{"content:create", "content:read", "content:update", "content:delete", "content:publish", "content:moderate", "videos:create", "videos:read", "videos:update", "videos:delete", "videos:manage", "analytics:read", "analytics:export"},
		IsSystemRole: true,
		Color:        "#059669",
		Icon:         "document-text",
		CreatedAt:    "2024-01-01T00:00:00Z",
		UpdatedAt:    "2024-01-01T00:00:00Z",
	},
	{
		ID:           "user-account-manager",
		Name:         "User Account Manager",
		Slug:         "user-account-manager",
		Description:  "User management and support operations",
		Category:     "core",
		Level:        7,
		Permissions:  []string{"users:create", "users:read", "users:update", "users:delete", "users:manage", "analytics:read"},
		IsSystemRole: true,
		Color:        "#2563eb",
		Icon:         "users",
		CreatedAt:    "2024-01-01T00:00:00Z",
		UpdatedAt:    "2024-01-01T00:00:00Z",
	},
}

// MOCK USERS WITH ROLES DATA
var MOCK_USERS_WITH_ROLES_DATA = []UserWithRoles{
	{
		ID:        1,
		Email:     "admin@bome.com",
		FirstName: "Super",
		LastName:  "Administrator",
		Status:    "active",
		Roles:     []Role{MOCK_ROLES_DATA[0]},
		RoleNames: []string{"Super Administrator"},
		LastLogin: "2024-04-01T12:00:00Z",
		CreatedAt: "2024-01-01T00:00:00Z",
	},
	{
		ID:        2,
		Email:     "system@bome.com",
		FirstName: "System",
		LastName:  "Administrator",
		Status:    "active",
		Roles:     []Role{MOCK_ROLES_DATA[1]},
		RoleNames: []string{"System Administrator"},
		LastLogin: "2024-03-30T14:22:00Z",
		CreatedAt: "2024-01-05T00:00:00Z",
	},
	{
		ID:        3,
		Email:     "content@bome.com",
		FirstName: "Content",
		LastName:  "Manager",
		Status:    "active",
		Roles:     []Role{MOCK_ROLES_DATA[2]},
		RoleNames: []string{"Content Manager"},
		LastLogin: "2024-03-29T08:15:00Z",
		CreatedAt: "2024-01-10T00:00:00Z",
	},
}

// Helper function to get all permission IDs
func getAllPermissionIDs() []string {
	var ids []string
	for _, perm := range MOCK_PERMISSIONS_DATA {
		ids = append(ids, perm.ID)
	}
	return ids
}

// ROLES ENDPOINTS

// GetRolesHandler returns all roles with optional filtering
func GetRolesHandler(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")

	filteredRoles := MOCK_ROLES_DATA

	// Apply search filter
	if search != "" {
		searchLower := strings.ToLower(search)
		filtered := []Role{}
		for _, role := range filteredRoles {
			if strings.Contains(strings.ToLower(role.Name), searchLower) ||
				strings.Contains(strings.ToLower(role.Description), searchLower) {
				filtered = append(filtered, role)
			}
		}
		filteredRoles = filtered
	}

	// Apply category filter
	if category != "" {
		filtered := []Role{}
		for _, role := range filteredRoles {
			if strings.EqualFold(role.Category, category) {
				filtered = append(filtered, role)
			}
		}
		filteredRoles = filtered
	}

	c.JSON(http.StatusOK, gin.H{"roles": filteredRoles})
}

// GetPermissionsHandler returns all permissions
func GetPermissionsHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"permissions": MOCK_PERMISSIONS_DATA})
}

// GetUsersWithRolesHandler returns users with their role assignments
func GetUsersWithRolesHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	filteredUsers := MOCK_USERS_WITH_ROLES_DATA

	// Apply search filter
	if search != "" {
		searchLower := strings.ToLower(search)
		filtered := []UserWithRoles{}
		for _, user := range filteredUsers {
			if strings.Contains(strings.ToLower(user.Email), searchLower) ||
				strings.Contains(strings.ToLower(user.FirstName), searchLower) ||
				strings.Contains(strings.ToLower(user.LastName), searchLower) {
				filtered = append(filtered, user)
			}
		}
		filteredUsers = filtered
	}

	// Apply pagination
	startIndex := (page - 1) * limit
	endIndex := startIndex + limit
	if startIndex >= len(filteredUsers) {
		filteredUsers = []UserWithRoles{}
	} else if endIndex > len(filteredUsers) {
		filteredUsers = filteredUsers[startIndex:]
	} else {
		filteredUsers = filteredUsers[startIndex:endIndex]
	}

	totalPages := (len(MOCK_USERS_WITH_ROLES_DATA) + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"users": filteredUsers,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      len(MOCK_USERS_WITH_ROLES_DATA),
			"totalPages": totalPages,
		},
	})
}

// GetRoleAnalyticsHandler returns role usage analytics
func GetRoleAnalyticsHandler(c *gin.Context) {
	analytics := gin.H{
		"totalRoles": len(MOCK_ROLES_DATA),
		"systemRoles": func() int {
			count := 0
			for _, role := range MOCK_ROLES_DATA {
				if role.IsSystemRole {
					count++
				}
			}
			return count
		}(),
		"customRoles": func() int {
			count := 0
			for _, role := range MOCK_ROLES_DATA {
				if !role.IsSystemRole {
					count++
				}
			}
			return count
		}(),
		"totalUsers": len(MOCK_USERS_WITH_ROLES_DATA),
		"roleDistribution": []gin.H{
			{"roleName": "Super Administrator", "userCount": 1, "percentage": 33.3},
			{"roleName": "System Administrator", "userCount": 1, "percentage": 33.3},
			{"roleName": "Content Manager", "userCount": 1, "percentage": 33.3},
		},
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

// SetupRolesRoutes sets up all roles management routes
func SetupRolesRoutes(router *gin.RouterGroup) {
	api := router.Group("/api/v1")
	{
		// Roles endpoints
		api.GET("/roles", GetRolesHandler)
		api.GET("/permissions", GetPermissionsHandler)
		api.GET("/users/roles", GetUsersWithRolesHandler)
		api.GET("/roles/analytics", GetRoleAnalyticsHandler)
	}
}
