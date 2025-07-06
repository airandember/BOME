// Standardized Role System Types
// This file defines the standardized role system that matches the backend implementation

export type PermissionResource = 
	| 'system' | 'users' | 'content' | 'videos' | 'articles' | 'events' | 'advertisements' 
	| 'analytics' | 'financial' | 'security' | 'technical' | 'academic' | 'roles' | 'permissions';

export type PermissionAction = 
	| 'full_access' | 'create' | 'read' | 'update' | 'delete' | 'manage' | 'publish' | 'moderate'
	| 'approve' | 'export' | 'refund' | 'incident' | 'support' | 'review' | 'coordinate';

export type RoleCategory = 
	| 'system' | 'content' | 'subsystem' | 'marketing' | 'user_management' | 'analytics' 
	| 'financial' | 'security' | 'technical' | 'academic' | 'base';

export type Subsystem = 'hub' | 'articles' | 'youtube' | 'streaming' | 'events';

export interface StandardizedPermission {
	id: string;
	resource: PermissionResource;
	action: PermissionAction;
	description: string;
	category: string;
	subsystem: Subsystem | 'all';
}

export interface StandardizedRole {
	id: string;
	name: string;
	slug: string;
	description: string;
	category: RoleCategory;
	level: number;
	permissions: string[];
	isSystemRole: boolean;
	color: string;
	icon: string;
	subsystemAccess: Subsystem[];
	createdAt: string;
	updatedAt: string;
}

export interface UserWithStandardizedRoles {
	id: number;
	email: string;
	firstName: string;
	lastName: string;
	status: string;
	roles: StandardizedRole[];
	roleNames: string[];
	lastLogin: string;
	createdAt: string;
}

// Standardized Role IDs (matching backend)
export const STANDARDIZED_ROLE_IDS = {
	// System Administration (Level 10-9)
	SUPER_ADMIN: 'super_admin',
	SYSTEM_ADMIN: 'system_admin',
	
	// Content Management (Level 8-6)
	CONTENT_MANAGER: 'content_manager',
	CONTENT_EDITOR: 'content_editor',
	CONTENT_CREATOR: 'content_creator',
	
	// Subsystem-Specific Roles (Level 7)
	ARTICLES_MANAGER: 'articles_manager',
	YOUTUBE_MANAGER: 'youtube_manager',
	STREAMING_MANAGER: 'streaming_manager',
	EVENTS_MANAGER: 'events_manager',
	
	// Marketing & Advertising (Level 7-4)
	ADVERTISEMENT_MANAGER: 'advertisement_manager',
	MARKETING_SPECIALIST: 'marketing_specialist',
	
	// User Management (Level 7-5)
	USER_MANAGER: 'user_manager',
	SUPPORT_SPECIALIST: 'support_specialist',
	
	// Analytics & Financial (Level 7)
	ANALYTICS_MANAGER: 'analytics_manager',
	FINANCIAL_ADMIN: 'financial_admin',
	
	// Technical & Security (Level 6-5)
	SECURITY_ADMIN: 'security_admin',
	TECHNICAL_SPECIALIST: 'technical_specialist',
	
	// Academic & Research (Level 6-5)
	ACADEMIC_REVIEWER: 'academic_reviewer',
	RESEARCH_COORDINATOR: 'research_coordinator',
	
	// Base User Roles (Level 3-1)
	ADVERTISER: 'advertiser',
	USER: 'user',
} as const;

export type StandardizedRoleID = typeof STANDARDIZED_ROLE_IDS[keyof typeof STANDARDIZED_ROLE_IDS];

// Role Level Constants
export const ROLE_LEVELS = {
	SUPER_ADMIN: 10,
	SYSTEM_ADMIN: 9,
	CONTENT_MANAGER: 8,
	CONTENT_EDITOR: 7,
	SUBSYSTEM_MANAGER: 7,
	CONTENT_CREATOR: 6,
	SECURITY_ADMIN: 6,
	ACADEMIC_REVIEWER: 6,
	SUPPORT_SPECIALIST: 5,
	TECHNICAL_SPECIALIST: 5,
	RESEARCH_COORDINATOR: 5,
	MARKETING_SPECIALIST: 4,
	ADVERTISER: 3,
	USER: 1,
} as const;

// Permission IDs (matching backend)
export const STANDARDIZED_PERMISSION_IDS = {
	// System Permissions
	SYSTEM_FULL_ACCESS: 'system:full_access',
	SYSTEM_MANAGE: 'system:manage',
	SYSTEM_READ: 'system:read',
	SYSTEM_UPDATE: 'system:update',
	
	// Role & Permission Management
	ROLES_CREATE: 'roles:create',
	ROLES_READ: 'roles:read',
	ROLES_UPDATE: 'roles:update',
	ROLES_DELETE: 'roles:delete',
	PERMISSIONS_MANAGE: 'permissions:manage',
	
	// User Management
	USERS_CREATE: 'users:create',
	USERS_READ: 'users:read',
	USERS_UPDATE: 'users:update',
	USERS_DELETE: 'users:delete',
	USERS_MANAGE: 'users:manage',
	
	// Content Management
	CONTENT_CREATE: 'content:create',
	CONTENT_READ: 'content:read',
	CONTENT_UPDATE: 'content:update',
	CONTENT_DELETE: 'content:delete',
	CONTENT_PUBLISH: 'content:publish',
	CONTENT_MODERATE: 'content:moderate',
	
	// Video Management
	VIDEOS_CREATE: 'videos:create',
	VIDEOS_READ: 'videos:read',
	VIDEOS_UPDATE: 'videos:update',
	VIDEOS_DELETE: 'videos:delete',
	VIDEOS_MANAGE: 'videos:manage',
	
	// Articles Management
	ARTICLES_CREATE: 'articles:create',
	ARTICLES_READ: 'articles:read',
	ARTICLES_UPDATE: 'articles:update',
	ARTICLES_DELETE: 'articles:delete',
	ARTICLES_PUBLISH: 'articles:publish',
	ARTICLES_MANAGE: 'articles:manage',
	
	// Events Management
	EVENTS_CREATE: 'events:create',
	EVENTS_READ: 'events:read',
	EVENTS_UPDATE: 'events:update',
	EVENTS_DELETE: 'events:delete',
	EVENTS_MANAGE: 'events:manage',
	
	// Advertisement Management
	ADVERTISEMENTS_CREATE: 'advertisements:create',
	ADVERTISEMENTS_READ: 'advertisements:read',
	ADVERTISEMENTS_UPDATE: 'advertisements:update',
	ADVERTISEMENTS_DELETE: 'advertisements:delete',
	ADVERTISEMENTS_MANAGE: 'advertisements:manage',
	ADVERTISEMENTS_APPROVE: 'advertisements:approve',
	
	// Analytics
	ANALYTICS_READ: 'analytics:read',
	ANALYTICS_EXPORT: 'analytics:export',
	ANALYTICS_MANAGE: 'analytics:manage',
	
	// Financial
	FINANCIAL_READ: 'financial:read',
	FINANCIAL_MANAGE: 'financial:manage',
	FINANCIAL_REFUND: 'financial:refund',
	
	// Security
	SECURITY_READ: 'security:read',
	SECURITY_MANAGE: 'security:manage',
	SECURITY_INCIDENT: 'security:incident',
	
	// Technical
	TECHNICAL_READ: 'technical:read',
	TECHNICAL_MANAGE: 'technical:manage',
	TECHNICAL_SUPPORT: 'technical:support',
	
	// Academic
	ACADEMIC_REVIEW: 'academic:review',
	ACADEMIC_COORDINATE: 'academic:coordinate',
	ACADEMIC_MANAGE: 'academic:manage',
} as const;

export type StandardizedPermissionID = typeof STANDARDIZED_PERMISSION_IDS[keyof typeof STANDARDIZED_PERMISSION_IDS];

// Role Categories
export const ROLE_CATEGORIES = {
	SYSTEM: 'system',
	CONTENT: 'content',
	SUBSYSTEM: 'subsystem',
	MARKETING: 'marketing',
	USER_MANAGEMENT: 'user_management',
	ANALYTICS: 'analytics',
	FINANCIAL: 'financial',
	SECURITY: 'security',
	TECHNICAL: 'technical',
	ACADEMIC: 'academic',
	BASE: 'base',
} as const;

// Subsystems
export const SUBSYSTEMS = {
	HUB: 'hub',
	ARTICLES: 'articles',
	YOUTUBE: 'youtube',
	STREAMING: 'streaming',
	EVENTS: 'events',
} as const;

// Helper Functions
export function getRoleLevel(roleId: StandardizedRoleID): number {
	return ROLE_LEVELS[roleId as keyof typeof ROLE_LEVELS] || 1;
}

export function getRoleCategory(roleId: StandardizedRoleID): RoleCategory {
	const categoryMap: Record<StandardizedRoleID, RoleCategory> = {
		[STANDARDIZED_ROLE_IDS.SUPER_ADMIN]: 'system',
		[STANDARDIZED_ROLE_IDS.SYSTEM_ADMIN]: 'system',
		[STANDARDIZED_ROLE_IDS.CONTENT_MANAGER]: 'content',
		[STANDARDIZED_ROLE_IDS.CONTENT_EDITOR]: 'content',
		[STANDARDIZED_ROLE_IDS.CONTENT_CREATOR]: 'content',
		[STANDARDIZED_ROLE_IDS.ARTICLES_MANAGER]: 'subsystem',
		[STANDARDIZED_ROLE_IDS.YOUTUBE_MANAGER]: 'subsystem',
		[STANDARDIZED_ROLE_IDS.STREAMING_MANAGER]: 'subsystem',
		[STANDARDIZED_ROLE_IDS.EVENTS_MANAGER]: 'subsystem',
		[STANDARDIZED_ROLE_IDS.ADVERTISEMENT_MANAGER]: 'marketing',
		[STANDARDIZED_ROLE_IDS.MARKETING_SPECIALIST]: 'marketing',
		[STANDARDIZED_ROLE_IDS.USER_MANAGER]: 'user_management',
		[STANDARDIZED_ROLE_IDS.SUPPORT_SPECIALIST]: 'user_management',
		[STANDARDIZED_ROLE_IDS.ANALYTICS_MANAGER]: 'analytics',
		[STANDARDIZED_ROLE_IDS.FINANCIAL_ADMIN]: 'financial',
		[STANDARDIZED_ROLE_IDS.SECURITY_ADMIN]: 'security',
		[STANDARDIZED_ROLE_IDS.TECHNICAL_SPECIALIST]: 'technical',
		[STANDARDIZED_ROLE_IDS.ACADEMIC_REVIEWER]: 'academic',
		[STANDARDIZED_ROLE_IDS.RESEARCH_COORDINATOR]: 'academic',
		[STANDARDIZED_ROLE_IDS.ADVERTISER]: 'base',
		[STANDARDIZED_ROLE_IDS.USER]: 'base',
	};
	
	return categoryMap[roleId] || 'base';
}

export function getSubsystemAccess(roleId: StandardizedRoleID): Subsystem[] {
	const accessMap: Record<StandardizedRoleID, Subsystem[]> = {
		[STANDARDIZED_ROLE_IDS.SUPER_ADMIN]: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		[STANDARDIZED_ROLE_IDS.SYSTEM_ADMIN]: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		[STANDARDIZED_ROLE_IDS.CONTENT_MANAGER]: ['articles', 'youtube', 'streaming'],
		[STANDARDIZED_ROLE_IDS.CONTENT_EDITOR]: ['articles', 'youtube', 'streaming'],
		[STANDARDIZED_ROLE_IDS.CONTENT_CREATOR]: ['articles', 'youtube', 'streaming'],
		[STANDARDIZED_ROLE_IDS.ARTICLES_MANAGER]: ['articles'],
		[STANDARDIZED_ROLE_IDS.YOUTUBE_MANAGER]: ['youtube'],
		[STANDARDIZED_ROLE_IDS.STREAMING_MANAGER]: ['streaming'],
		[STANDARDIZED_ROLE_IDS.EVENTS_MANAGER]: ['events'],
		[STANDARDIZED_ROLE_IDS.ADVERTISEMENT_MANAGER]: ['hub'],
		[STANDARDIZED_ROLE_IDS.MARKETING_SPECIALIST]: ['hub'],
		[STANDARDIZED_ROLE_IDS.USER_MANAGER]: ['hub'],
		[STANDARDIZED_ROLE_IDS.SUPPORT_SPECIALIST]: ['hub'],
		[STANDARDIZED_ROLE_IDS.ANALYTICS_MANAGER]: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		[STANDARDIZED_ROLE_IDS.FINANCIAL_ADMIN]: ['hub'],
		[STANDARDIZED_ROLE_IDS.SECURITY_ADMIN]: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		[STANDARDIZED_ROLE_IDS.TECHNICAL_SPECIALIST]: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		[STANDARDIZED_ROLE_IDS.ACADEMIC_REVIEWER]: ['articles'],
		[STANDARDIZED_ROLE_IDS.RESEARCH_COORDINATOR]: ['articles'],
		[STANDARDIZED_ROLE_IDS.ADVERTISER]: ['hub'],
		[STANDARDIZED_ROLE_IDS.USER]: ['hub', 'articles', 'youtube', 'streaming', 'events'],
	};
	
	return accessMap[roleId] || ['hub'];
}

export function hasPermission(roleId: StandardizedRoleID, permissionId: StandardizedPermissionID): boolean {
	// This would typically be checked against the backend
	// For now, we'll implement basic permission checking based on role levels
	const roleLevel = getRoleLevel(roleId);
	
	// Super admin has all permissions
	if (roleLevel === 10) return true;
	
	// System admin has most permissions except role management
	if (roleLevel === 9) {
		return !permissionId.startsWith('roles:') && !permissionId.startsWith('permissions:');
	}
	
	// Content manager has content-related permissions
	if (roleLevel === 8) {
		return permissionId.startsWith('content:') || 
			   permissionId.startsWith('videos:') || 
			   permissionId.startsWith('articles:') ||
			   permissionId.startsWith('analytics:');
	}
	
	// Subsystem managers have subsystem-specific permissions
	if (roleLevel === 7) {
		if (roleId === STANDARDIZED_ROLE_IDS.ARTICLES_MANAGER) {
			return permissionId.startsWith('articles:') || permissionId.startsWith('content:') || permissionId.startsWith('analytics:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.YOUTUBE_MANAGER || roleId === STANDARDIZED_ROLE_IDS.STREAMING_MANAGER) {
			return permissionId.startsWith('videos:') || permissionId.startsWith('content:') || permissionId.startsWith('analytics:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.EVENTS_MANAGER) {
			return permissionId.startsWith('events:') || permissionId.startsWith('users:') || permissionId.startsWith('analytics:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.ADVERTISEMENT_MANAGER) {
			return permissionId.startsWith('advertisements:') || permissionId.startsWith('analytics:') || permissionId.startsWith('financial:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.USER_MANAGER) {
			return permissionId.startsWith('users:') || permissionId.startsWith('analytics:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.ANALYTICS_MANAGER) {
			return permissionId.startsWith('analytics:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.FINANCIAL_ADMIN) {
			return permissionId.startsWith('financial:') || permissionId.startsWith('analytics:');
		}
	}
	
	// Lower level roles have limited permissions
	if (roleLevel <= 6) {
		if (roleId === STANDARDIZED_ROLE_IDS.CONTENT_EDITOR) {
			return permissionId.startsWith('content:') || 
				   permissionId.startsWith('videos:') || 
				   permissionId.startsWith('articles:') ||
				   permissionId === STANDARDIZED_PERMISSION_IDS.ANALYTICS_READ;
		}
		if (roleId === STANDARDIZED_ROLE_IDS.CONTENT_CREATOR) {
			return permissionId.startsWith('content:') || 
				   permissionId.startsWith('videos:') || 
				   permissionId.startsWith('articles:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.SECURITY_ADMIN) {
			return permissionId.startsWith('security:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.ACADEMIC_REVIEWER) {
			return permissionId.startsWith('academic:') || 
				   permissionId.startsWith('articles:');
		}
	}
	
	if (roleLevel <= 5) {
		if (roleId === STANDARDIZED_ROLE_IDS.SUPPORT_SPECIALIST) {
			return permissionId.startsWith('users:') || permissionId === STANDARDIZED_PERMISSION_IDS.TECHNICAL_SUPPORT;
		}
		if (roleId === STANDARDIZED_ROLE_IDS.TECHNICAL_SPECIALIST) {
			return permissionId.startsWith('technical:');
		}
		if (roleId === STANDARDIZED_ROLE_IDS.RESEARCH_COORDINATOR) {
			return permissionId.startsWith('academic:') || permissionId.startsWith('articles:');
		}
	}
	
	if (roleLevel <= 4) {
		if (roleId === STANDARDIZED_ROLE_IDS.MARKETING_SPECIALIST) {
			return permissionId.startsWith('advertisements:') || permissionId === STANDARDIZED_PERMISSION_IDS.ANALYTICS_READ;
		}
	}
	
	if (roleLevel <= 3) {
		if (roleId === STANDARDIZED_ROLE_IDS.ADVERTISER) {
			return permissionId.startsWith('advertisements:');
		}
	}
	
	// Base user has minimal permissions
	if (roleLevel <= 1) {
		if (roleId === STANDARDIZED_ROLE_IDS.USER) {
			return permissionId === STANDARDIZED_PERMISSION_IDS.CONTENT_READ;
		}
	}
	
	return false;
}

export function canAccessSubsystem(roleId: StandardizedRoleID, subsystem: Subsystem): boolean {
	const access = getSubsystemAccess(roleId);
	return access.includes(subsystem);
}

export function getRolesBySubsystem(subsystem: Subsystem): StandardizedRoleID[] {
	const allRoles = Object.values(STANDARDIZED_ROLE_IDS);
	return allRoles.filter(roleId => canAccessSubsystem(roleId, subsystem));
}

export function getRolesByCategory(category: RoleCategory): StandardizedRoleID[] {
	const allRoles = Object.values(STANDARDIZED_ROLE_IDS);
	return allRoles.filter(roleId => getRoleCategory(roleId) === category);
}

export function getRolesByLevel(minLevel: number, maxLevel?: number): StandardizedRoleID[] {
	const allRoles = Object.values(STANDARDIZED_ROLE_IDS);
	return allRoles.filter(roleId => {
		const level = getRoleLevel(roleId);
		return level >= minLevel && (!maxLevel || level <= maxLevel);
	});
} 