// MOCK DATA FOR ROLE-BASED ACCESS CONTROL SYSTEM
// Section 4.7 - Super Admin Role & Permissions System

import type { 
	Role, Permission, UserWithRoles, RoleAssignment, DashboardWidget, 
	RoleTemplate, PermissionCategory, RoleUsageAnalytics, UserDashboardConfig,
	RoleAuditLog, EmergencyAccess, ApprovalWorkflow, DelegationPermission
} from '../types/roles';

// PERMISSIONS DEFINITIONS
export const MOCK_PERMISSIONS: Permission[] = [
	// User Management Permissions
	{ id: 'users:create', resource: 'users', action: 'create', description: 'Create new user accounts', category: 'core' },
	{ id: 'users:read', resource: 'users', action: 'read', description: 'View user account information', category: 'core' },
	{ id: 'users:update', resource: 'users', action: 'update', description: 'Edit user account details', category: 'core' },
	{ id: 'users:delete', resource: 'users', action: 'delete', description: 'Delete user accounts', category: 'core' },
	{ id: 'users:manage', resource: 'users', action: 'manage', description: 'Full user account management', category: 'core' },

	// Content Management Permissions
	{ id: 'content:create', resource: 'content', action: 'create', description: 'Create new content', category: 'content' },
	{ id: 'content:read', resource: 'content', action: 'read', description: 'View content', category: 'content' },
	{ id: 'content:update', resource: 'content', action: 'update', description: 'Edit existing content', category: 'content' },
	{ id: 'content:delete', resource: 'content', action: 'delete', description: 'Delete content', category: 'content' },
	{ id: 'content:publish', resource: 'content', action: 'publish', description: 'Publish content to live site', category: 'content' },
	{ id: 'content:moderate', resource: 'content', action: 'moderate', description: 'Moderate user-generated content', category: 'content' },

	// Video Management Permissions
	{ id: 'videos:create', resource: 'videos', action: 'create', description: 'Upload new videos', category: 'content' },
	{ id: 'videos:read', resource: 'videos', action: 'read', description: 'View video content and metadata', category: 'content' },
	{ id: 'videos:update', resource: 'videos', action: 'update', description: 'Edit video metadata and settings', category: 'content' },
	{ id: 'videos:delete', resource: 'videos', action: 'delete', description: 'Delete video content', category: 'content' },
	{ id: 'videos:manage', resource: 'videos', action: 'manage', description: 'Full video management capabilities', category: 'content' },

	// Article Management Permissions
	{ id: 'articles:create', resource: 'articles', action: 'create', description: 'Create new articles', category: 'content' },
	{ id: 'articles:read', resource: 'articles', action: 'read', description: 'View articles', category: 'content' },
	{ id: 'articles:update', resource: 'articles', action: 'update', description: 'Edit existing articles', category: 'content' },
	{ id: 'articles:delete', resource: 'articles', action: 'delete', description: 'Delete articles', category: 'content' },
	{ id: 'articles:publish', resource: 'articles', action: 'publish', description: 'Publish articles', category: 'content' },

	// Events Management Permissions
	{ id: 'events:create', resource: 'events', action: 'create', description: 'Create new events', category: 'events' },
	{ id: 'events:read', resource: 'events', action: 'read', description: 'View event information', category: 'events' },
	{ id: 'events:update', resource: 'events', action: 'update', description: 'Edit event details', category: 'events' },
	{ id: 'events:delete', resource: 'events', action: 'delete', description: 'Delete events', category: 'events' },
	{ id: 'events:manage', resource: 'events', action: 'manage', description: 'Full event management', category: 'events' },

	// Advertisement Management Permissions
	{ id: 'advertisements:create', resource: 'advertisements', action: 'create', description: 'Create advertisement campaigns', category: 'marketing' },
	{ id: 'advertisements:read', resource: 'advertisements', action: 'read', description: 'View advertisement data', category: 'marketing' },
	{ id: 'advertisements:update', resource: 'advertisements', action: 'update', description: 'Edit advertisement campaigns', category: 'marketing' },
	{ id: 'advertisements:delete', resource: 'advertisements', action: 'delete', description: 'Delete advertisement campaigns', category: 'marketing' },
	{ id: 'advertisements:approve', resource: 'advertisements', action: 'approve', description: 'Approve advertisement campaigns', category: 'marketing' },
	{ id: 'advertisements:manage', resource: 'advertisements', action: 'manage', description: 'Full advertisement management', category: 'marketing' },

	// Analytics Permissions
	{ id: 'analytics:read', resource: 'analytics', action: 'read', description: 'View analytics and reports', category: 'core' },
	{ id: 'analytics:export', resource: 'analytics', action: 'export', description: 'Export analytics data', category: 'core' },
	{ id: 'analytics:manage', resource: 'analytics', action: 'manage', description: 'Configure analytics settings', category: 'core' },

	// System Administration Permissions
	{ id: 'system:read', resource: 'system', action: 'read', description: 'View system information', category: 'technical' },
	{ id: 'system:update', resource: 'system', action: 'update', description: 'Modify system settings', category: 'technical' },
	{ id: 'system:manage', resource: 'system', action: 'manage', description: 'Full system administration', category: 'technical' },

	// Security Permissions
	{ id: 'security:read', resource: 'security', action: 'read', description: 'View security logs and reports', category: 'technical' },
	{ id: 'security:manage', resource: 'security', action: 'manage', description: 'Manage security settings', category: 'technical' },

	// Billing & Financial Permissions
	{ id: 'billing:read', resource: 'billing', action: 'read', description: 'View billing information', category: 'core' },
	{ id: 'billing:manage', resource: 'billing', action: 'manage', description: 'Manage billing and payments', category: 'core' },

	// Role & Permission Management
	{ id: 'roles:create', resource: 'roles', action: 'create', description: 'Create new roles', category: 'core' },
	{ id: 'roles:read', resource: 'roles', action: 'read', description: 'View role information', category: 'core' },
	{ id: 'roles:update', resource: 'roles', action: 'update', description: 'Edit existing roles', category: 'core' },
	{ id: 'roles:delete', resource: 'roles', action: 'delete', description: 'Delete roles', category: 'core' },
	{ id: 'permissions:manage', resource: 'permissions', action: 'manage', description: 'Manage user permissions', category: 'core' },

	// Backup & Recovery Permissions
	{ id: 'backups:create', resource: 'backups', action: 'create', description: 'Create system backups', category: 'technical' },
	{ id: 'backups:read', resource: 'backups', action: 'read', description: 'View backup information', category: 'technical' },
	{ id: 'backups:manage', resource: 'backups', action: 'manage', description: 'Manage backup systems', category: 'technical' },

	// Logs & Monitoring Permissions
	{ id: 'logs:read', resource: 'logs', action: 'read', description: 'View system logs', category: 'technical' },
	{ id: 'logs:export', resource: 'logs', action: 'export', description: 'Export log data', category: 'technical' },

	// Settings & Configuration Permissions
	{ id: 'settings:read', resource: 'settings', action: 'read', description: 'View system settings', category: 'technical' },
	{ id: 'settings:update', resource: 'settings', action: 'update', description: 'Modify system settings', category: 'technical' },

	// Integration Management Permissions
	{ id: 'integrations:read', resource: 'integrations', action: 'read', description: 'View integration settings', category: 'technical' },
	{ id: 'integrations:manage', resource: 'integrations', action: 'manage', description: 'Manage third-party integrations', category: 'technical' },

	// Notification Management Permissions
	{ id: 'notifications:read', resource: 'notifications', action: 'read', description: 'View notifications', category: 'core' },
	{ id: 'notifications:manage', resource: 'notifications', action: 'manage', description: 'Manage notification settings', category: 'core' }
];

// STANDARDIZED ROLE DEFINITIONS
export const MOCK_STANDARDIZED_ROLES: Role[] = [
	// System Administration (Level 10-9)
	{
		id: 'super_admin',
		name: 'Super Administrator',
		slug: 'super-administrator',
		description: 'Full system access and role management capabilities',
		category: 'system',
		level: 10,
		permissions: MOCK_PERMISSIONS.map(p => p.id), // All permissions
		isSystemRole: true,
		color: '#dc2626',
		icon: 'crown',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'system-administrator',
		name: 'System Administrator',
		slug: 'system-administrator',
		description: 'Technical system management without role changes',
		category: 'core',
		level: 9,
		permissions: [
			'system:read', 'system:update', 'system:manage',
			'security:read', 'security:manage',
			'backups:create', 'backups:read', 'backups:manage',
			'logs:read', 'logs:export',
			'settings:read', 'settings:update',
			'integrations:read', 'integrations:manage',
			'analytics:read', 'analytics:export'
		],
		isSystemRole: true,
		color: '#7c3aed',
		icon: 'server',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'content-manager',
		name: 'Content Manager',
		slug: 'content-manager',
		description: 'Overall content strategy and oversight',
		category: 'core',
		level: 8,
		permissions: [
			'content:create', 'content:read', 'content:update', 'content:delete', 'content:publish', 'content:moderate',
			'videos:create', 'videos:read', 'videos:update', 'videos:delete', 'videos:manage',
			'articles:create', 'articles:read', 'articles:update', 'articles:delete', 'articles:publish',
			'analytics:read', 'analytics:export'
		],
		isSystemRole: true,
		color: '#059669',
		icon: 'document-text',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'user-account-manager',
		name: 'User Account Manager',
		slug: 'user-account-manager',
		description: 'User management and support operations',
		category: 'core',
		level: 7,
		permissions: [
			'users:create', 'users:read', 'users:update', 'users:delete', 'users:manage',
			'analytics:read',
			'notifications:read', 'notifications:manage'
		],
		isSystemRole: true,
		color: '#2563eb',
		icon: 'users',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'financial-administrator',
		name: 'Financial Administrator',
		slug: 'financial-administrator',
		description: 'Revenue, billing, and financial reporting',
		category: 'core',
		level: 8,
		permissions: [
			'billing:read', 'billing:manage',
			'analytics:read', 'analytics:export',
			'advertisements:read'
		],
		isSystemRole: true,
		color: '#dc2626',
		icon: 'currency-dollar',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'security-administrator',
		name: 'Security Administrator',
		slug: 'security-administrator',
		description: 'Security monitoring and incident response',
		category: 'core',
		level: 9,
		permissions: [
			'security:read', 'security:manage',
			'logs:read', 'logs:export',
			'users:read',
			'system:read'
		],
		isSystemRole: true,
		color: '#dc2626',
		icon: 'shield-check',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'analytics-manager',
		name: 'Analytics Manager',
		slug: 'analytics-manager',
		description: 'Data analysis and reporting across all systems',
		category: 'core',
		level: 7,
		permissions: [
			'analytics:read', 'analytics:export', 'analytics:manage',
			'users:read',
			'content:read',
			'videos:read',
			'articles:read',
			'events:read',
			'advertisements:read'
		],
		isSystemRole: true,
		color: '#059669',
		icon: 'chart-bar',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},

	// Content & Editorial Roles (4.7.2)
	{
		id: 'article-writer',
		name: 'Article Writer',
		slug: 'article-writer',
		description: 'Create and edit blog articles and research content',
		category: 'content',
		level: 4,
		permissions: [
			'articles:create', 'articles:read', 'articles:update',
			'content:create', 'content:read', 'content:update'
		],
		isSystemRole: false,
		color: '#059669',
		icon: 'pencil',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'content-editor',
		name: 'Content Editor',
		slug: 'content-editor',
		description: 'Review, approve, and publish written content',
		category: 'content',
		level: 6,
		permissions: [
			'articles:create', 'articles:read', 'articles:update', 'articles:delete', 'articles:publish',
			'content:create', 'content:read', 'content:update', 'content:delete', 'content:publish', 'content:moderate'
		],
		isSystemRole: false,
		color: '#059669',
		icon: 'check-circle',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'video-content-manager',
		name: 'Video Content Manager',
		slug: 'video-content-manager',
		description: 'Upload, organize, and manage video content',
		category: 'content',
		level: 5,
		permissions: [
			'videos:create', 'videos:read', 'videos:update', 'videos:delete', 'videos:manage',
			'content:create', 'content:read', 'content:update'
		],
		isSystemRole: false,
		color: '#7c3aed',
		icon: 'video-camera',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'content-moderator',
		name: 'Content Moderator',
		slug: 'content-moderator',
		description: 'Review user-generated content and comments',
		category: 'content',
		level: 4,
		permissions: [
			'content:read', 'content:moderate',
			'videos:read',
			'articles:read'
		],
		isSystemRole: false,
		color: '#dc2626',
		icon: 'eye',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},

	// Marketing & Advertisement Roles (4.7.3)
	{
		id: 'advertisement-manager',
		name: 'Advertisement Manager',
		slug: 'advertisement-manager',
		description: 'Full advertisement system oversight',
		category: 'marketing',
		level: 7,
		permissions: [
			'advertisements:create', 'advertisements:read', 'advertisements:update', 'advertisements:delete', 'advertisements:approve', 'advertisements:manage',
			'analytics:read', 'analytics:export',
			'billing:read'
		],
		isSystemRole: false,
		color: '#dc2626',
		icon: 'megaphone',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'marketing-team-member',
		name: 'Marketing Team Member',
		slug: 'marketing-team-member',
		description: 'Campaign creation and advertiser relations',
		category: 'marketing',
		level: 4,
		permissions: [
			'advertisements:create', 'advertisements:read', 'advertisements:update',
			'analytics:read'
		],
		isSystemRole: false,
		color: '#dc2626',
		icon: 'presentation-chart-line',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},

	// Events & Community Roles (4.7.4)
	{
		id: 'events-manager',
		name: 'Events Manager',
		slug: 'events-manager',
		description: 'Full events system management and oversight',
		category: 'events',
		level: 7,
		permissions: [
			'events:create', 'events:read', 'events:update', 'events:delete', 'events:manage',
			'users:read',
			'analytics:read'
		],
		isSystemRole: false,
		color: '#2563eb',
		icon: 'calendar',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'event-coordinator',
		name: 'Event Coordinator',
		slug: 'event-coordinator',
		description: 'Create and manage individual events',
		category: 'events',
		level: 5,
		permissions: [
			'events:create', 'events:read', 'events:update',
			'users:read'
		],
		isSystemRole: false,
		color: '#2563eb',
		icon: 'clipboard-list',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},

	// Technical & Support Roles (4.7.5)
	{
		id: 'video-streaming-specialist',
		name: 'Video Streaming Specialist',
		slug: 'video-streaming-specialist',
		description: 'Manage Bunny.net integration and video technical issues',
		category: 'technical',
		level: 6,
		permissions: [
			'videos:read', 'videos:update', 'videos:manage',
			'integrations:read', 'integrations:manage',
			'analytics:read'
		],
		isSystemRole: false,
		color: '#7c3aed',
		icon: 'play',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'customer-support-lead',
		name: 'Customer Support Lead',
		slug: 'customer-support-lead',
		description: 'Oversee customer support operations',
		category: 'technical',
		level: 6,
		permissions: [
			'users:read', 'users:update',
			'notifications:read', 'notifications:manage',
			'analytics:read'
		],
		isSystemRole: false,
		color: '#059669',
		icon: 'support',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},

	// Academic Roles (4.7.6)
	{
		id: 'academic-reviewer',
		name: 'Academic Reviewer',
		slug: 'academic-reviewer',
		description: 'Review scholarly content for accuracy and quality',
		category: 'academic',
		level: 6,
		permissions: [
			'articles:read', 'articles:update',
			'content:read', 'content:moderate',
			'videos:read'
		],
		isSystemRole: false,
		color: '#059669',
		icon: 'academic-cap',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'subject-matter-expert',
		name: 'Subject Matter Expert',
		slug: 'subject-matter-expert',
		description: 'Provide expertise in specific Book of Mormon research areas',
		category: 'academic',
		level: 7,
		permissions: [
			'articles:create', 'articles:read', 'articles:update',
			'content:create', 'content:read', 'content:update',
			'videos:read'
		],
		isSystemRole: false,
		color: '#059669',
		icon: 'book-open',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	}
];

// PERMISSION CATEGORIES
export const MOCK_PERMISSION_CATEGORIES: PermissionCategory[] = [
	{
		id: 'user-management',
		name: 'User Management',
		description: 'Permissions related to user account management',
		permissions: MOCK_PERMISSIONS.filter(p => p.resource === 'users')
	},
	{
		id: 'content-management',
		name: 'Content Management',
		description: 'Permissions for managing content, articles, and videos',
		permissions: MOCK_PERMISSIONS.filter(p => ['content', 'articles', 'videos'].includes(p.resource))
	},
	{
		id: 'financial-access',
		name: 'Financial Access',
		description: 'Permissions for billing, revenue, and financial data',
		permissions: MOCK_PERMISSIONS.filter(p => ['billing', 'advertisements'].includes(p.resource))
	},
	{
		id: 'system-administration',
		name: 'System Administration',
		description: 'Technical system management permissions',
		permissions: MOCK_PERMISSIONS.filter(p => ['system', 'security', 'backups', 'logs', 'settings', 'integrations'].includes(p.resource))
	},
	{
		id: 'analytics-access',
		name: 'Analytics Access',
		description: 'Permissions for viewing and managing analytics',
		permissions: MOCK_PERMISSIONS.filter(p => p.resource === 'analytics')
	},
	{
		id: 'events-management',
		name: 'Events Management',
		description: 'Permissions for event creation and management',
		permissions: MOCK_PERMISSIONS.filter(p => p.resource === 'events')
	}
];

// DASHBOARD WIDGETS
export const MOCK_DASHBOARD_WIDGETS: DashboardWidget[] = [
	{
		id: 'user-stats',
		name: 'User Statistics',
		component: 'UserStatsWidget',
		requiredPermissions: ['users:read'],
		category: 'users',
		size: 'medium',
		order: 1,
		isDefault: true
	},
	{
		id: 'content-overview',
		name: 'Content Overview',
		component: 'ContentOverviewWidget',
		requiredPermissions: ['content:read'],
		category: 'content',
		size: 'large',
		order: 2,
		isDefault: true
	},
	{
		id: 'revenue-analytics',
		name: 'Revenue Analytics',
		component: 'RevenueAnalyticsWidget',
		requiredPermissions: ['billing:read', 'analytics:read'],
		category: 'analytics',
		size: 'large',
		order: 3,
		isDefault: true
	},
	{
		id: 'system-health',
		name: 'System Health',
		component: 'SystemHealthWidget',
		requiredPermissions: ['system:read'],
		category: 'system',
		size: 'medium',
		order: 4,
		isDefault: true
	},
	{
		id: 'security-alerts',
		name: 'Security Alerts',
		component: 'SecurityAlertsWidget',
		requiredPermissions: ['security:read'],
		category: 'system',
		size: 'medium',
		order: 5,
		isDefault: true
	},
	{
		id: 'advertisement-performance',
		name: 'Advertisement Performance',
		component: 'AdPerformanceWidget',
		requiredPermissions: ['advertisements:read'],
		category: 'marketing',
		size: 'large',
		order: 6,
		isDefault: false
	},
	{
		id: 'event-registrations',
		name: 'Event Registrations',
		component: 'EventRegistrationsWidget',
		requiredPermissions: ['events:read'],
		category: 'events',
		size: 'medium',
		order: 7,
		isDefault: false
	},
	{
		id: 'video-analytics',
		name: 'Video Analytics',
		component: 'VideoAnalyticsWidget',
		requiredPermissions: ['videos:read', 'analytics:read'],
		category: 'analytics',
		size: 'large',
		order: 8,
		isDefault: false
	}
];

// MOCK USERS WITH ROLES
export const MOCK_USERS_WITH_ROLES: UserWithRoles[] = [
	{
		id: '1',
		email: 'super.admin@bome.com',
		firstName: 'Super',
		lastName: 'Administrator',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'super_admin')!],
		permissions: MOCK_PERMISSIONS,
		lastLogin: '2024-12-20T10:30:00Z',
		status: 'active',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-12-20T10:30:00Z'
	},
	{
		id: '1a',
		email: 'admin@bome.test',
		firstName: 'Super',
		lastName: 'Administrator',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'super_admin')!],
		permissions: MOCK_PERMISSIONS,
		lastLogin: '2024-12-20T11:15:00Z',
		status: 'active',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-12-20T11:15:00Z'
	},
	{
		id: '2',
		email: 'content.manager@bome.com',
		firstName: 'Sarah',
		lastName: 'Johnson',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'content-manager')!],
		permissions: MOCK_PERMISSIONS.filter(p => 
			['content:create', 'content:read', 'content:update', 'content:delete', 'content:publish', 'content:moderate',
			 'videos:create', 'videos:read', 'videos:update', 'videos:delete', 'videos:manage',
			 'articles:create', 'articles:read', 'articles:update', 'articles:delete', 'articles:publish',
			 'analytics:read', 'analytics:export'].includes(p.id)
		),
		lastLogin: '2024-12-19T15:45:00Z',
		status: 'active',
		createdAt: '2024-02-15T00:00:00Z',
		updatedAt: '2024-12-19T15:45:00Z'
	},
	{
		id: '3',
		email: 'ad.manager@bome.com',
		firstName: 'Michael',
		lastName: 'Chen',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'advertisement_manager')!],
		permissions: MOCK_PERMISSIONS.filter(p => 
			['advertisements:create', 'advertisements:read', 'advertisements:update', 'advertisements:delete', 'advertisements:approve', 'advertisements:manage',
			 'analytics:read', 'analytics:export', 'billing:read'].includes(p.id)
		),
		lastLogin: '2024-12-20T09:15:00Z',
		status: 'active',
		createdAt: '2024-03-10T00:00:00Z',
		updatedAt: '2024-12-20T09:15:00Z'
	},
	{
		id: '4',
		email: 'events.coordinator@bome.com',
		firstName: 'Emily',
		lastName: 'Rodriguez',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'events_manager')!],
		permissions: MOCK_PERMISSIONS.filter(p => 
			['events:create', 'events:read', 'events:update', 'users:read'].includes(p.id)
		),
		lastLogin: '2024-12-18T14:20:00Z',
		status: 'active',
		createdAt: '2024-04-05T00:00:00Z',
		updatedAt: '2024-12-18T14:20:00Z'
	},
	{
		id: '5',
		email: 'article.writer@bome.com',
		firstName: 'David',
		lastName: 'Thompson',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'content_creator')!],
		permissions: MOCK_PERMISSIONS.filter(p => 
			['articles:create', 'articles:read', 'articles:update', 'content:create', 'content:read', 'content:update'].includes(p.id)
		),
		lastLogin: '2024-12-19T11:30:00Z',
		status: 'active',
		createdAt: '2024-05-20T00:00:00Z',
		updatedAt: '2024-12-19T11:30:00Z'
	},
	{
		id: '6',
		email: 'academic.reviewer@bome.com',
		firstName: 'Dr. Rebecca',
		lastName: 'Williams',
		roles: [MOCK_STANDARDIZED_ROLES.find(r => r.id === 'academic_reviewer')!],
		permissions: MOCK_PERMISSIONS.filter(p => 
			['articles:read', 'articles:update', 'content:read', 'content:moderate', 'videos:read'].includes(p.id)
		),
		lastLogin: '2024-12-17T16:45:00Z',
		status: 'active',
		createdAt: '2024-06-12T00:00:00Z',
		updatedAt: '2024-12-17T16:45:00Z'
	}
];

// ROLE TEMPLATES
export const MOCK_ROLE_TEMPLATES: RoleTemplate[] = [
	{
		id: 'content-creator-template',
		name: 'Content Creator',
		description: 'Template for users who create articles and video content',
		category: 'content',
		permissions: ['articles:create', 'articles:read', 'articles:update', 'videos:create', 'videos:read', 'videos:update', 'content:create', 'content:read', 'content:update'],
		widgets: ['content-overview', 'video-analytics'],
		isBuiltIn: true
	},
	{
		id: 'marketing-specialist-template',
		name: 'Marketing Specialist',
		description: 'Template for marketing team members',
		category: 'marketing',
		permissions: ['advertisements:create', 'advertisements:read', 'advertisements:update', 'analytics:read'],
		widgets: ['advertisement-performance', 'revenue-analytics'],
		isBuiltIn: true
	},
	{
		id: 'event-organizer-template',
		name: 'Event Organizer',
		description: 'Template for event management staff',
		category: 'events',
		permissions: ['events:create', 'events:read', 'events:update', 'users:read'],
		widgets: ['event-registrations', 'user-stats'],
		isBuiltIn: true
	}
];

// ROLE USAGE ANALYTICS
export const MOCK_ROLE_ANALYTICS: RoleUsageAnalytics[] = [
	{
		roleId: 'super_admin',
		activeUsers: 1,
		totalAssignments: 1,
		averageSessionDuration: 180,
		mostUsedPermissions: ['users:read', 'analytics:read', 'system:read'],
		leastUsedPermissions: ['backups:create', 'logs:export'],
		securityIncidents: 0,
		lastActivity: '2024-12-20T10:30:00Z'
	},
	{
		roleId: 'content_manager',
		activeUsers: 1,
		totalAssignments: 2,
		averageSessionDuration: 120,
		mostUsedPermissions: ['content:read', 'articles:read', 'videos:read'],
		leastUsedPermissions: ['content:delete', 'videos:delete'],
		securityIncidents: 0,
		lastActivity: '2024-12-19T15:45:00Z'
	},
	{
		roleId: 'advertisement_manager',
		activeUsers: 1,
		totalAssignments: 1,
		averageSessionDuration: 90,
		mostUsedPermissions: ['advertisements:read', 'advertisements:approve', 'analytics:read'],
		leastUsedPermissions: ['advertisements:delete'],
		securityIncidents: 0,
		lastActivity: '2024-12-20T09:15:00Z'
	}
];

// AUDIT LOGS
export const MOCK_ROLE_AUDIT_LOGS: RoleAuditLog[] = [
	{
		id: '1',
		action: 'assign',
		entityType: 'user_role',
		entityId: '2',
		userId: '1',
		targetUserId: '2',
		changes: {
			role: { old: null, new: 'content_manager' }
		},
		reason: 'Initial role assignment',
		timestamp: '2024-02-15T10:00:00Z',
		ipAddress: '192.168.1.100',
		userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
	},
	{
		id: '2',
		action: 'create',
		entityType: 'role',
		entityId: 'marketing-team-member',
		userId: '1',
		changes: {
			name: { old: null, new: 'Marketing Team Member' },
			permissions: { old: [], new: ['advertisements:create', 'advertisements:read', 'advertisements:update'] }
		},
		timestamp: '2024-03-01T14:30:00Z',
		ipAddress: '192.168.1.100',
		userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
	},
	{
		id: '3',
		action: 'permission_change',
		entityType: 'role',
		entityId: 'content_creator',
		userId: '1',
		changes: {
			permissions: { 
				old: ['articles:create', 'articles:read', 'articles:update'], 
				new: ['articles:create', 'articles:read', 'articles:update', 'content:create', 'content:read', 'content:update'] 
			}
		},
		reason: 'Added content management permissions',
		timestamp: '2024-05-25T11:15:00Z',
		ipAddress: '192.168.1.100',
		userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
	}
];

// HELPER FUNCTIONS
export function getUserPermissions(userId: string): Permission[] {
	const user = MOCK_USERS_WITH_ROLES.find(u => u.id === userId);
	return user?.permissions || [];
}

export function hasPermission(userId: string, permissionId: string): boolean {
	const permissions = getUserPermissions(userId);
	return permissions.some(p => p.id === permissionId);
}

export function getUserRoles(userId: string): Role[] {
	const user = MOCK_USERS_WITH_ROLES.find(u => u.id === userId);
	return user?.roles || [];
}

export function getRolesByCategory(category: string): Role[] {
	return MOCK_STANDARDIZED_ROLES.filter(role => role.category === category);
}

export function getPermissionsByCategory(category: string): Permission[] {
	return MOCK_PERMISSIONS.filter(permission => permission.category === category);
}

export function canAccessWidget(userId: string, widgetId: string): boolean {
	const widget = MOCK_DASHBOARD_WIDGETS.find(w => w.id === widgetId);
	if (!widget) return false;
	
	const userPermissions = getUserPermissions(userId);
	return widget.requiredPermissions.every(reqPermission => 
		userPermissions.some(p => p.id === reqPermission)
	);
}

export function getUserDashboardWidgets(userId: string): DashboardWidget[] {
	return MOCK_DASHBOARD_WIDGETS.filter(widget => canAccessWidget(userId, widget.id));
}

export function searchRoles(query: string): Role[] {
	const lowerQuery = query.toLowerCase();
	return MOCK_STANDARDIZED_ROLES.filter(role => 
		role.name.toLowerCase().includes(lowerQuery) ||
		role.description.toLowerCase().includes(lowerQuery) ||
		role.category.toLowerCase().includes(lowerQuery)
	);
}

export function searchUsers(query: string): UserWithRoles[] {
	const lowerQuery = query.toLowerCase();
	return MOCK_USERS_WITH_ROLES.filter(user => 
		user.email.toLowerCase().includes(lowerQuery) ||
		user.firstName.toLowerCase().includes(lowerQuery) ||
		user.lastName.toLowerCase().includes(lowerQuery) ||
		user.roles.some(role => role.name.toLowerCase().includes(lowerQuery))
	);
}

export function getRoleHierarchy(): Role[] {
	return MOCK_STANDARDIZED_ROLES.sort((a, b) => b.level - a.level);
}

export function getSystemRoles(): Role[] {
	return MOCK_STANDARDIZED_ROLES.filter(role => role.isSystemRole);
}

export function getCustomRoles(): Role[] {
	return MOCK_STANDARDIZED_ROLES.filter(role => !role.isSystemRole);
}

// Mock API Response Helpers
export function createMockRoleResponse<T>(data: T, delay: number = 500): Promise<T> {
	return new Promise((resolve) => {
		setTimeout(() => resolve(data), delay);
	});
}

export function createMockRoleErrorResponse(message: string, delay: number = 500): Promise<never> {
	return new Promise((_, reject) => {
		setTimeout(() => reject(new Error(message)), delay);
	});
}

export function getUserByEmail(email: string): UserWithRoles | undefined {
	return MOCK_USERS_WITH_ROLES.find(user => user.email === email);
}

export function hasUserSuperAdminRole(email: string): boolean {
	const user = getUserByEmail(email);
	if (!user) return false;
	
	return user.roles.some(role => role.id === 'super_admin');
}

export function getUserRoleNames(email: string): string[] {
	const user = getUserByEmail(email);
	if (!user) return [];
	
	return user.roles.map(role => role.name);
} 