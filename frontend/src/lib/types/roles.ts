// ROLE-BASED ACCESS CONTROL (RBAC) TYPES
// Section 4.7 - Super Admin Role & Permissions System

export type PermissionAction = 'create' | 'read' | 'update' | 'delete' | 'approve' | 'publish' | 'moderate' | 'export' | 'manage';

export type PermissionResource = 
	| 'users' | 'content' | 'videos' | 'articles' | 'events' | 'advertisements' 
	| 'analytics' | 'system' | 'security' | 'billing' | 'roles' | 'permissions'
	| 'backups' | 'logs' | 'settings' | 'integrations' | 'notifications';

export interface Permission {
	id: string;
	resource: PermissionResource;
	action: PermissionAction;
	description: string;
	category: 'core' | 'content' | 'marketing' | 'events' | 'technical' | 'academic';
}

export interface Role {
	id: string;
	name: string;
	slug: string;
	description: string;
	category: 'system' | 'core' | 'content' | 'marketing' | 'events' | 'technical' | 'academic' | 'subsystem' | 'user_management' | 'analytics' | 'financial' | 'security' | 'base';
	level: number; // 1-10, higher = more privileges
	permissions: string[]; // Permission IDs
	isSystemRole: boolean; // Cannot be deleted
	color: string; // For UI display
	icon: string; // Icon identifier
	createdAt: string;
	updatedAt: string;
	createdBy?: string;
	updatedBy?: string;
}

export interface UserRole {
	userId: string;
	roleId: string;
	assignedAt: string;
	assignedBy: string;
	expiresAt?: string; // For temporary access
	isActive: boolean;
	notes?: string;
}

export interface RoleAssignment {
	id: string;
	userId: string;
	roleId: string;
	assignedBy: string;
	assignedAt: string;
	expiresAt?: string;
	status: 'active' | 'expired' | 'revoked' | 'pending';
	approvedBy?: string;
	approvedAt?: string;
	revokedBy?: string;
	revokedAt?: string;
	reason?: string;
}

export interface DashboardWidget {
	id: string;
	name: string;
	component: string;
	requiredPermissions: string[];
	category: 'analytics' | 'content' | 'users' | 'system' | 'marketing' | 'events';
	size: 'small' | 'medium' | 'large' | 'full';
	order: number;
	isDefault: boolean;
}

export interface RoleTemplate {
	id: string;
	name: string;
	description: string;
	category: string;
	permissions: string[];
	widgets: string[];
	isBuiltIn: boolean;
}

// Core Admin Roles (4.7.1)
export const CORE_ADMIN_ROLES = [
	'super-administrator',
	'system-administrator', 
	'content-manager',
	'user-account-manager',
	'financial-administrator',
	'security-administrator',
	'analytics-manager'
] as const;

// Content & Editorial Roles (4.7.2)
export const CONTENT_EDITORIAL_ROLES = [
	'article-writer',
	'content-editor',
	'video-content-manager',
	'content-moderator',
	'seo-specialist',
	'research-coordinator',
	'translation-manager'
] as const;

// Marketing & Advertisement Roles (4.7.3)
export const MARKETING_ADVERTISEMENT_ROLES = [
	'advertisement-manager',
	'marketing-team-member',
	'advertisement-reviewer',
	'placement-manager',
	'revenue-analyst',
	'campaign-coordinator',
	'social-media-manager'
] as const;

// Events & Community Roles (4.7.4)
export const EVENTS_COMMUNITY_ROLES = [
	'events-manager',
	'event-coordinator',
	'registration-manager',
	'community-manager',
	'speaker-coordinator',
	'venue-manager',
	'event-marketing-specialist'
] as const;

// Technical & Support Roles (4.7.5)
export const TECHNICAL_SUPPORT_ROLES = [
	'video-streaming-specialist',
	'customer-support-lead',
	'technical-support',
	'quality-assurance',
	'data-analyst',
	'integration-specialist',
	'backup-administrator'
] as const;

// Specialized Academic Roles (4.7.6)
export const ACADEMIC_ROLES = [
	'academic-reviewer',
	'citation-manager',
	'peer-review-coordinator',
	'research-database-manager',
	'academic-partnership-coordinator',
	'scholarly-communication-specialist',
	'subject-matter-expert'
] as const;

export type CoreAdminRole = typeof CORE_ADMIN_ROLES[number];
export type ContentEditorialRole = typeof CONTENT_EDITORIAL_ROLES[number];
export type MarketingAdvertisementRole = typeof MARKETING_ADVERTISEMENT_ROLES[number];
export type EventsCommunityRole = typeof EVENTS_COMMUNITY_ROLES[number];
export type TechnicalSupportRole = typeof TECHNICAL_SUPPORT_ROLES[number];
export type AcademicRole = typeof ACADEMIC_ROLES[number];

export type SystemRole = 
	| CoreAdminRole 
	| ContentEditorialRole 
	| MarketingAdvertisementRole 
	| EventsCommunityRole 
	| TechnicalSupportRole 
	| AcademicRole;

// Permission Categories (4.7.7)
export interface PermissionCategory {
	id: string;
	name: string;
	description: string;
	permissions: Permission[];
}

// Advanced Permission Features (4.7.9)
export interface TimeBasedPermission {
	permissionId: string;
	startTime: string;
	endTime: string;
	timezone: string;
	isRecurring: boolean;
	recurringPattern?: 'daily' | 'weekly' | 'monthly';
}

export interface LocationBasedPermission {
	permissionId: string;
	allowedLocations: string[]; // IP ranges or geographic locations
	restrictedLocations: string[];
}

export interface ApprovalWorkflow {
	id: string;
	name: string;
	requiredPermission: string;
	approvers: string[]; // Role IDs that can approve
	requiredApprovals: number;
	timeoutHours: number;
	escalationRoles: string[];
}

export interface DelegationPermission {
	id: string;
	delegatorUserId: string;
	delegateUserId: string;
	permissions: string[];
	startDate: string;
	endDate: string;
	reason: string;
	isActive: boolean;
}

// Emergency Access (Break-glass)
export interface EmergencyAccess {
	id: string;
	userId: string;
	reason: string;
	requestedPermissions: string[];
	approvedBy?: string;
	approvedAt?: string;
	expiresAt: string;
	status: 'pending' | 'approved' | 'denied' | 'expired' | 'revoked';
	auditLog: EmergencyAccessAudit[];
}

export interface EmergencyAccessAudit {
	id: string;
	action: string;
	timestamp: string;
	userId: string;
	details: Record<string, any>;
}

// Role Analytics
export interface RoleUsageAnalytics {
	roleId: string;
	activeUsers: number;
	totalAssignments: number;
	averageSessionDuration: number;
	mostUsedPermissions: string[];
	leastUsedPermissions: string[];
	securityIncidents: number;
	lastActivity: string;
}

// User with Role Information
export interface UserWithRoles {
	id: string;
	email: string;
	firstName: string;
	lastName: string;
	roles: Role[];
	permissions: Permission[];
	lastLogin?: string;
	status: 'active' | 'inactive' | 'suspended';
	createdAt: string;
	updatedAt: string;
}

// Dashboard Configuration
export interface UserDashboardConfig {
	userId: string;
	widgets: DashboardWidget[];
	layout: 'grid' | 'list' | 'custom';
	theme: 'light' | 'dark' | 'auto';
	notifications: {
		email: boolean;
		push: boolean;
		sms: boolean;
	};
	quickActions: string[];
	customizations: Record<string, any>;
}

// Audit Trail
export interface RoleAuditLog {
	id: string;
	action: 'create' | 'update' | 'delete' | 'assign' | 'revoke' | 'permission_change';
	entityType: 'role' | 'permission' | 'user_role' | 'delegation';
	entityId: string;
	userId: string; // Who performed the action
	targetUserId?: string; // Who was affected (for assignments)
	changes: Record<string, { old: any; new: any }>;
	reason?: string;
	timestamp: string;
	ipAddress: string;
	userAgent: string;
} 