import type { StandardizedRole, StandardizedPermission, Department } from '$lib/types/standardized_roles';

// Fallback roles data
export const FALLBACK_ROLES: StandardizedRole[] = [
	{
		id: 'super_admin',
		name: 'Super Administrator',
		slug: 'super-administrator',
		description: 'Full system access and role management capabilities',
		category: 'system',
		level: 10,
		permissions: ['system:full_access'],
		isSystemRole: true,
		color: '#dc2626',
		icon: 'crown',
		subsystemAccess: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		department: { id: 1000, name: 'Core_Admin', icon: '🔧', color: '#dc2626', description: 'Core system administration', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'system_admin',
		name: 'System Administrator',
		slug: 'system-administrator',
		description: 'Technical system management without role changes',
		category: 'system',
		level: 9,
		permissions: ['system:read', 'system:update', 'system:manage'],
		isSystemRole: true,
		color: '#7c3aed',
		icon: 'server',
		subsystemAccess: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		department: { id: 1000, name: 'Core_Admin', icon: '🔧', color: '#dc2626', description: 'Core system administration', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'content_manager',
		name: 'Content Manager',
		slug: 'content-manager',
		description: 'Overall content strategy and oversight',
		category: 'content',
		level: 8,
		permissions: ['content:create', 'content:read', 'content:update', 'content:delete', 'content:publish'],
		isSystemRole: true,
		color: '#059669',
		icon: 'edit',
		subsystemAccess: ['articles', 'youtube', 'streaming'],
		department: { id: 600, name: 'Content_Management', icon: '📚', color: '#7C3AED', description: 'Content creation and management', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'content_editor',
		name: 'Content Editor',
		slug: 'content-editor',
		description: 'Review, approve, edit, and publish content',
		category: 'content',
		level: 7,
		permissions: ['content:read', 'content:update', 'content:publish'],
		isSystemRole: false,
		color: '#059669',
		icon: 'pencil',
		subsystemAccess: ['articles', 'youtube', 'streaming'],
		department: { id: 600, name: 'Content_Management', icon: '📚', color: '#7C3AED', description: 'Content creation and management', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'content_creator',
		name: 'Content Creator',
		slug: 'content-creator',
		description: 'Create and edit content with limited publishing',
		category: 'content',
		level: 6,
		permissions: ['content:create', 'content:read', 'content:update'],
		isSystemRole: false,
		color: '#059669',
		icon: 'plus-circle',
		subsystemAccess: ['articles', 'youtube', 'streaming'],
		department: { id: 600, name: 'Content_Management', icon: '📚', color: '#7C3AED', description: 'Content creation and management', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'articles_manager',
		name: 'Articles Manager',
		slug: 'articles-manager',
		description: 'Full articles subsystem management',
		category: 'subsystem',
		level: 7,
		permissions: ['articles:create', 'articles:read', 'articles:update', 'articles:delete', 'articles:publish', 'articles:manage'],
		isSystemRole: false,
		color: '#1e40af',
		icon: 'document',
		subsystemAccess: ['articles'],
		department: { id: 500, name: 'Content_Strat', icon: '🎨', color: '#2563EB', description: 'Content strategy and planning', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'youtube_manager',
		name: 'YouTube Manager',
		slug: 'youtube-manager',
		description: 'YouTube system management',
		category: 'subsystem',
		level: 7,
		permissions: ['videos:create', 'videos:read', 'videos:update', 'videos:delete', 'videos:manage'],
		isSystemRole: false,
		color: '#dc2626',
		icon: 'play',
		subsystemAccess: ['youtube'],
		department: { id: 500, name: 'Content_Strat', icon: '🎨', color: '#2563EB', description: 'Content strategy and planning', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'streaming_manager',
		name: 'Video Streaming Manager',
		slug: 'streaming-manager',
		description: 'Bunny.net streaming platform management',
		category: 'subsystem',
		level: 7,
		permissions: ['videos:create', 'videos:read', 'videos:update', 'videos:delete', 'videos:manage'],
		isSystemRole: false,
		color: '#7c3aed',
		icon: 'video',
		subsystemAccess: ['streaming'],
		department: { id: 500, name: 'Content_Strat', icon: '🎨', color: '#2563EB', description: 'Content strategy and planning', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'events_manager',
		name: 'Events Manager',
		slug: 'events-manager',
		description: 'Events system management',
		category: 'subsystem',
		level: 7,
		permissions: ['events:create', 'events:read', 'events:update', 'events:delete', 'events:manage'],
		isSystemRole: false,
		color: '#2563eb',
		icon: 'calendar',
		subsystemAccess: ['events'],
		department: { id: 500, name: 'Content_Strat', icon: '🎨', color: '#2563EB', description: 'Content strategy and planning', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'advertisement_manager',
		name: 'Advertisement Manager',
		slug: 'advertisement-manager',
		description: 'Full advertisement system oversight',
		category: 'marketing',
		level: 7,
		permissions: ['advertisements:create', 'advertisements:read', 'advertisements:update', 'advertisements:delete', 'advertisements:manage', 'advertisements:approve'],
		isSystemRole: false,
		color: '#f59e0b',
		icon: 'presentation-chart-line',
		subsystemAccess: ['hub'],
		department: { id: 700, name: 'Marketing', icon: '📢', color: '#f59e0b', description: 'Marketing and advertising', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'marketing_specialist',
		name: 'Marketing Specialist',
		slug: 'marketing-specialist',
		description: 'Campaign creation and advertiser relations',
		category: 'marketing',
		level: 4,
		permissions: ['advertisements:create', 'advertisements:read', 'advertisements:update'],
		isSystemRole: false,
		color: '#f59e0b',
		icon: 'megaphone',
		subsystemAccess: ['hub'],
		department: { id: 700, name: 'Marketing', icon: '📢', color: '#f59e0b', description: 'Marketing and advertising', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'user_manager',
		name: 'User Account Manager',
		slug: 'user-manager',
		description: 'User management and support operations',
		category: 'user_management',
		level: 7,
		permissions: ['users:create', 'users:read', 'users:update', 'users:delete', 'users:manage'],
		isSystemRole: false,
		color: '#2563eb',
		icon: 'users',
		subsystemAccess: ['hub'],
		department: { id: 400, name: 'User_Admin', icon: '👥', color: '#5FB7E0', description: 'User management and administration', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'support_specialist',
		name: 'Support Specialist',
		slug: 'support-specialist',
		description: 'User support and basic account management',
		category: 'user_management',
		level: 5,
		permissions: ['users:read', 'users:update'],
		isSystemRole: false,
		color: '#2563eb',
		icon: 'life-buoy',
		subsystemAccess: ['hub'],
		department: { id: 400, name: 'User_Admin', icon: '👥', color: '#5FB7E0', description: 'User management and administration', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'analytics_manager',
		name: 'Analytics Manager',
		slug: 'analytics-manager',
		description: 'Data analysis and reporting across all systems',
		category: 'analytics',
		level: 7,
		permissions: ['analytics:read', 'analytics:export', 'analytics:manage'],
		isSystemRole: false,
		color: '#059669',
		icon: 'bar-chart',
		subsystemAccess: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		department: { id: 200, name: 'System_Insight', icon: '📈', color: '#D6BD1A', description: 'System analytics and insights', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'financial_admin',
		name: 'Financial Administrator',
		slug: 'financial-administrator',
		description: 'Revenue, billing, and financial reporting',
		category: 'financial',
		level: 7,
		permissions: ['financial:read', 'financial:manage'],
		isSystemRole: false,
		color: '#059669',
		icon: 'credit-card',
		subsystemAccess: ['hub'],
		department: { id: 300, name: 'Finance', icon: '💰', color: '#059669', description: 'Financial operations and billing', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'security_admin',
		name: 'Security Administrator',
		slug: 'security-administrator',
		description: 'Security monitoring and incident response',
		category: 'security',
		level: 6,
		permissions: ['security:read', 'security:manage'],
		isSystemRole: false,
		color: '#dc2626',
		icon: 'shield',
		subsystemAccess: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		department: { id: 100, name: 'Secure_Tech', icon: '🛡️', color: '#3E6313', description: 'Security and technical infrastructure', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'technical_specialist',
		name: 'Technical Specialist',
		slug: 'technical-specialist',
		description: 'Technical support and maintenance',
		category: 'technical',
		level: 5,
		permissions: ['technical:read', 'technical:manage'],
		isSystemRole: false,
		color: '#7c3aed',
		icon: 'wrench',
		subsystemAccess: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		department: { id: 100, name: 'Secure_Tech', icon: '🛡️', color: '#3E6313', description: 'Security and technical infrastructure', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'academic_reviewer',
		name: 'Academic Reviewer',
		slug: 'academic-reviewer',
		description: 'Review scholarly content for accuracy and quality',
		category: 'academic',
		level: 6,
		permissions: ['academic:review'],
		isSystemRole: false,
		color: '#7c2d12',
		icon: 'academic-cap',
		subsystemAccess: ['articles'],
		department: { id: 800, name: 'Academia', icon: '🎓', color: '#7c2d12', description: 'Academic and research operations', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'research_coordinator',
		name: 'Research Coordinator',
		slug: 'research-coordinator',
		description: 'Coordinate academic research and citations',
		category: 'academic',
		level: 5,
		permissions: ['academic:coordinate'],
		isSystemRole: false,
		color: '#7c2d12',
		icon: 'book-open',
		subsystemAccess: ['articles'],
		department: { id: 800, name: 'Academia', icon: '🎓', color: '#7c2d12', description: 'Academic and research operations', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'advertiser',
		name: 'Advertiser',
		slug: 'advertiser',
		description: 'Create and manage advertising campaigns',
		category: 'base',
		level: 3,
		permissions: ['advertisements:create', 'advertisements:read', 'advertisements:update'],
		isSystemRole: false,
		color: '#f59e0b',
		icon: 'megaphone',
		subsystemAccess: ['hub'],
		department: { id: 700, name: 'Marketing', icon: '📢', color: '#f59e0b', description: 'Marketing and advertising', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	},
	{
		id: 'user',
		name: 'User',
		slug: 'user',
		description: 'Basic platform access',
		category: 'base',
		level: 1,
		permissions: ['content:read'],
		isSystemRole: true,
		color: '#6b7280',
		icon: 'user',
		subsystemAccess: ['hub', 'articles', 'youtube', 'streaming', 'events'],
		department: { id: 900, name: 'Base', icon: '🫂', color: '#6b7280', description: 'Base user operations', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	}
];

// Fallback departments data
export const FALLBACK_DEPARTMENTS: Department[] = [
	{ id: 100, name: 'Secure_Tech', icon: '🛡️', color: '#3E6313', description: 'Security and technical infrastructure', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 200, name: 'System_Insight', icon: '📈', color: '#D6BD1A', description: 'System analytics and insights', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 300, name: 'Finance', icon: '💰', color: '#059669', description: 'Financial operations and billing', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 400, name: 'User_Admin', icon: '👥', color: '#5FB7E0', description: 'User management and administration', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 500, name: 'Content_Strat', icon: '🎨', color: '#2563EB', description: 'Content strategy and planning', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 600, name: 'Content_Management', icon: '📚', color: '#7C3AED', description: 'Content creation and management', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 700, name: 'Marketing', icon: '📢', color: '#f59e0b', description: 'Marketing and advertising', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 800, name: 'Academia', icon: '🎓', color: '#7c2d12', description: 'Academic and research operations', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 900, name: 'Base', icon: '🫂', color: '#6b7280', description: 'Base user operations', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' },
	{ id: 1000, name: 'Core_Admin', icon: '🔧', color: '#dc2626', description: 'Core system administration', createdAt: '2024-01-01T00:00:00Z', updatedAt: '2024-01-01T00:00:00Z' }
];

// Fallback users data
export const FALLBACK_USERS = [
	{
		id: 1,
		firstName: 'John',
		lastName: 'Doe',
		email: 'john.doe@example.com',
		role: 'super_admin',
		emailVerified: true,
		status: 'active',
		createdAt: '2024-01-15T10:30:00Z',
		lastLogin: '2024-07-20T09:15:00Z',
		subscription: 'premium',
		department: 'Core_Admin'
	},
	{
		id: 2,
		firstName: 'Jane',
		lastName: 'Smith',
		email: 'jane.smith@example.com',
		role: 'content_manager',
		emailVerified: true,
		status: 'active',
		createdAt: '2024-02-10T14:20:00Z',
		lastLogin: '2024-07-19T16:45:00Z',
		subscription: 'premium',
		department: 'Content_Management'
	},
	{
		id: 3,
		firstName: 'Mike',
		lastName: 'Johnson',
		email: 'mike.johnson@example.com',
		role: 'user_manager',
		emailVerified: true,
		status: 'active',
		createdAt: '2024-03-05T11:15:00Z',
		lastLogin: '2024-07-20T08:30:00Z',
		subscription: 'standard',
		department: 'User_Admin'
	},
	{
		id: 4,
		firstName: 'Sarah',
		lastName: 'Wilson',
		email: 'sarah.wilson@example.com',
		role: 'analytics_manager',
		emailVerified: true,
		status: 'active',
		createdAt: '2024-01-20T09:45:00Z',
		lastLogin: '2024-07-18T13:20:00Z',
		subscription: 'premium',
		department: 'System_Insight'
	},
	{
		id: 5,
		firstName: 'David',
		lastName: 'Brown',
		email: 'david.brown@example.com',
		role: 'user',
		emailVerified: false,
		status: 'pending',
		createdAt: '2024-07-15T16:30:00Z',
		lastLogin: null,
		subscription: 'free',
		department: 'Base'
	}
];

// Helper functions
export function getDepartmentsWithRoles(roles: StandardizedRole[], departments: Department[]) {
	const deptMap = new Map();
	
	// Initialize departments
	departments.forEach(dept => {
		deptMap.set(dept.id, {
			...dept,
			roles: []
		});
	});
	
	// Group roles by department
	roles.forEach(role => {
		if (role.department) {
			const dept = deptMap.get(role.department.id);
			if (dept) {
				dept.roles.push(role);
			}
		}
	});
	
	return Array.from(deptMap.values()).filter(dept => dept.roles.length > 0);
}

export function getDepartmentUserCount(users: any[], departmentName: string) {
	return users.filter(user => user.department === departmentName).length;
} 