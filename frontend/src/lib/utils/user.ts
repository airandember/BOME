// User utility functions for audit and display purposes
export interface UserInfo {
	id: number;
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	email_verified: boolean;
}

/**
 * Gets user information from localStorage for audit purposes
 * Used when backend context is missing but user is authenticated in dashboard
 */
export function getUserFromLocalStorage(): UserInfo | null {
	if (typeof window === 'undefined') return null;
	
	try {
		const stored = localStorage.getItem('bome_user_data');
		console.log('STORED: ', stored);
		if (!stored) return null;
		
		const userData = JSON.parse(stored);
		return {
			id: userData.id,
			email: userData.email,
			first_name: userData.first_name || '',
			last_name: userData.last_name || '',
			role: userData.role || 'user',
			email_verified: userData.email_verified || false
		};
	} catch (error) {
		console.error('Failed to parse user data from localStorage:', error);
		return null;
	}
}

/**
 * Gets user display name for audit purposes
 * Returns full name if available, otherwise email
 */
export function getUserDisplayName(): string {
	const user = getUserFromLocalStorage();
	if (!user) return 'Dashboard User';
	
	if (user.first_name && user.last_name) {
		return `${user.first_name} ${user.last_name}`;
	} else if (user.first_name) {
		return user.first_name;
	} else if (user.last_name) {
		return user.last_name;
	} else {
		return user.email;
	}
}

/**
 * Gets user role for audit purposes
 */
export function getUserRole(): string {
	const user = getUserFromLocalStorage();
	return user?.role || 'dashboard';
}

/**
 * Gets user ID for audit purposes
 */
export function getUserId(): number | null {
	const user = getUserFromLocalStorage();
	return user?.id || null;
} 