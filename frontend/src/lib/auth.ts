import { writable } from 'svelte/store';
import type { Role } from './types/roles';

export interface User {
	id: number;
	email: string;
	firstName: string;
	lastName: string;
	role: string;
	roles?: Role[];
	emailVerified: boolean;
}

export interface AuthState {
	user: User | null;
	token: string | null;
	isAuthenticated: boolean;
}

// Create auth store
const createAuthStore = () => {
	const { subscribe, set, update } = writable<AuthState>({
		user: null,
		token: null,
		isAuthenticated: false
	});

	return {
		subscribe,
		// Computed property for isAuthenticated
		get isAuthenticated() {
			let authenticated = false;
			subscribe(state => {
				authenticated = state.isAuthenticated;
			})();
			return authenticated;
		},
		login: async (email: string, password: string) => {
			try {
				// Mock admin login for testing - support both admin@bome.com and admin@bome.test
				if ((email === 'admin@bome.com' || email === 'admin@bome.test') && password === 'admin123') {
					// Import role data for super admin assignment
					const { MOCK_ROLES } = await import('./mockData/roles');
					const superAdminRole = MOCK_ROLES.find(r => r.id === 'super-administrator');
					
					const user: User = {
						id: 1,
						email: email, // Use the provided email
						firstName: 'Super',
						lastName: 'Administrator',
						role: 'admin', // Legacy role field
						roles: superAdminRole ? [superAdminRole] : [], // New role management system
						emailVerified: true
					};

					const mockToken = 'mock-admin-token-' + Date.now();
					localStorage.setItem('token', mockToken);
					localStorage.setItem('userData', JSON.stringify(user));

					set({
						user,
						token: mockToken,
						isAuthenticated: true
					});

					return { success: true, user };
				}

				// Mock regular user login for testing
				if ((email === 'user@bome.com' || email === 'user@bome.test') && password === 'user123') {
					const user: User = {
						id: 2,
						email: email, // Use the provided email
						firstName: 'User',
						lastName: 'Account',
						role: 'user', // Regular user role
						roles: [], // No special roles for regular user
						emailVerified: true
					};

					const mockToken = 'mock-user-token-' + Date.now();
					localStorage.setItem('token', mockToken);
					localStorage.setItem('userData', JSON.stringify(user));

					set({
						user,
						token: mockToken,
						isAuthenticated: true
					});

					return { success: true, user };
				}

				// Mock advertiser login for testing - multiple patterns supported
				if ((email === 'advertiser@bome.com' || email === 'advertiser@bome.test') && password === 'advertiser123') {
					const user: User = {
						id: 3,
						email: email, // Use the provided email
						firstName: 'Business',
						lastName: 'Advertiser',
						role: 'advertiser', // Advertiser role
						roles: [], // No special roles for regular advertiser
						emailVerified: true
					};

					const mockToken = 'mock-advertiser-token-' + Date.now();
					localStorage.setItem('token', mockToken);
					localStorage.setItem('userData', JSON.stringify(user));

					set({
						user,
						token: mockToken,
						isAuthenticated: true
					});

					return { success: true, user };
				}

				// Mock business advertiser login - flexible password support
				if (email === 'business@bome.test' && (password === 'business123' || password === 'advertiser123' || password === 'password123')) {
					const user: User = {
						id: 4,
						email: email, // Use the provided email
						firstName: 'Business',
						lastName: 'Owner',
						role: 'advertiser', // Advertiser role
						roles: [], // No special roles for regular advertiser
						emailVerified: true
					};

					const mockToken = 'mock-advertiser-token-' + Date.now();
					localStorage.setItem('token', mockToken);
					localStorage.setItem('userData', JSON.stringify(user));

					set({
						user,
						token: mockToken,
						isAuthenticated: true
					});

					return { success: true, user };
				}

				// Regular API call for other users
				const response = await fetch('/api/v1/auth/login', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({ email, password })
				});

				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Login failed');
				}

				const data = await response.json();
				const user: User = {
					id: data.user.id,
					email: data.user.email,
					firstName: data.user.firstName || '',
					lastName: data.user.lastName || '',
					role: data.user.role,
					roles: data.user.roles,
					emailVerified: data.user.emailVerified || false
				};

				// Store token and user data in localStorage
				localStorage.setItem('token', data.token);
				localStorage.setItem('userData', JSON.stringify(user));

				set({
					user,
					token: data.token,
					isAuthenticated: true
				});

				return { success: true, user };
			} catch (error) {
				return { success: false, error: (error as Error).message };
			}
		},

		register: async (email: string, password: string, firstName: string, lastName: string) => {
			try {
				const response = await fetch('/api/v1/auth/register', {
					method: 'POST',
					headers: {
						'Content-Type': 'application/json'
					},
					body: JSON.stringify({ email, password, firstName, lastName })
				});

				if (!response.ok) {
					const error = await response.json();
					throw new Error(error.error || 'Registration failed');
				}

				return { success: true };
			} catch (error) {
				return { success: false, error: (error as Error).message };
			}
		},

		logout: async () => {
			// Clear localStorage first
			localStorage.removeItem('token');
			localStorage.removeItem('userData');
			
			// Clear the store state
			set({
				user: null,
				token: null,
				isAuthenticated: false
			});
			
			// Small delay to ensure state is updated
			await new Promise(resolve => setTimeout(resolve, 100));
		},

		initialize: async () => {
			const token = localStorage.getItem('token');
			if (token) {
				// Check if it's a mock admin token
				if (token.startsWith('mock-admin-token-')) {
					// Import role data for super admin assignment
					const { MOCK_ROLES } = await import('./mockData/roles');
					const superAdminRole = MOCK_ROLES.find(r => r.id === 'super-administrator');
					
					// Get stored user data to preserve the email
					const storedUserData = localStorage.getItem('userData');
					let email = 'admin@bome.com'; // default
					
					if (storedUserData) {
						try {
							const userData = JSON.parse(storedUserData);
							if (userData.email) {
								email = userData.email;
							}
						} catch (error) {
							console.error('Failed to parse stored user data:', error);
						}
					}

					const mockUser: User = {
						id: 1,
						email: email,
						firstName: 'Super',
						lastName: 'Administrator',
						role: 'admin',
						roles: superAdminRole ? [superAdminRole] : [],
						emailVerified: true
					};

					set({
						user: mockUser,
						token,
						isAuthenticated: true
					});
				} else if (token.startsWith('mock-user-token-')) {
					// Handle mock regular user token
					const storedUserData = localStorage.getItem('userData');
					let email = 'user@bome.com'; // default
					
					if (storedUserData) {
						try {
							const userData = JSON.parse(storedUserData);
							if (userData.email) {
								email = userData.email;
							}
						} catch (error) {
							console.error('Failed to parse stored user data:', error);
						}
					}

					const mockUser: User = {
						id: 2,
						email: email,
						firstName: 'User',
						lastName: 'Account',
						role: 'user',
						roles: [],
						emailVerified: true
					};

					set({
						user: mockUser,
						token,
						isAuthenticated: true
					});
				} else if (token.startsWith('mock-advertiser-token-')) {
					// Handle mock advertiser token
					const storedUserData = localStorage.getItem('userData');
					let email = 'advertiser@bome.com'; // default
					
					if (storedUserData) {
						try {
							const userData = JSON.parse(storedUserData);
							if (userData.email) {
								email = userData.email;
							}
						} catch (error) {
							console.error('Failed to parse stored user data:', error);
						}
					}

					const mockUser: User = {
						id: 3,
						email: email,
						firstName: 'Business',
						lastName: 'Advertiser',
						role: 'advertiser',
						roles: [],
						emailVerified: true
					};

					set({
						user: mockUser,
						token,
						isAuthenticated: true
					});
				} else {
					// For real tokens, we should validate with backend
					// For development, check if we have any stored user data
					const storedUserData = localStorage.getItem('userData');
					
					if (storedUserData) {
						try {
							const userData = JSON.parse(storedUserData);
							// Validate the stored user data structure
							if (userData.id && userData.email && userData.role) {
								set({
									user: userData,
									token,
									isAuthenticated: true
								});
								return;
							}
						} catch (error) {
							console.error('Failed to parse stored user data:', error);
						}
					}
					
					// If no valid stored data, validate token with backend
					// For now, clear the session to force re-login
					localStorage.removeItem('token');
					localStorage.removeItem('userData');
					set({
						user: null,
						token: null,
						isAuthenticated: false
					});
					
					console.log('No valid user data found, session cleared');
				}
			}
		},

		getCurrentUser: () => {
			let currentUser = null;
			subscribe(state => {
				currentUser = state.user;
			})();
			return currentUser;
		},

		updateUser: (updatedUser: Partial<User>) => {
			update(state => ({
				...state,
				user: state.user ? { ...state.user, ...updatedUser } : null
			}));
		}
	};
};

export const auth = createAuthStore();

// API helper with authentication
export const api = {
	get: async (url: string) => {
		const token = localStorage.getItem('token');
		const response = await fetch(url, {
			headers: {
				'Authorization': token ? `Bearer ${token}` : '',
				'Content-Type': 'application/json'
			}
		});
		return response.json();
	},

	post: async (url: string, data: any) => {
		const token = localStorage.getItem('token');
		const response = await fetch(url, {
			method: 'POST',
			headers: {
				'Authorization': token ? `Bearer ${token}` : '',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		});
		return response.json();
	},

	put: async (url: string, data: any) => {
		const token = localStorage.getItem('token');
		const response = await fetch(url, {
			method: 'PUT',
			headers: {
				'Authorization': token ? `Bearer ${token}` : '',
				'Content-Type': 'application/json'
			},
			body: JSON.stringify(data)
		});
		return response.json();
	},

	delete: async (url: string) => {
		const token = localStorage.getItem('token');
		const response = await fetch(url, {
			method: 'DELETE',
			headers: {
				'Authorization': token ? `Bearer ${token}` : '',
				'Content-Type': 'application/json'
			}
		});
		return response.json();
	}
}; 