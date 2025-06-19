import { writable } from 'svelte/store';

export interface User {
	id: number;
	email: string;
	firstName: string;
	lastName: string;
	role: string;
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
		login: async (email: string, password: string) => {
			try {
				// Mock admin login for testing
				if (email === 'admin@bome.com' && password === 'admin123') {
					const user: User = {
						id: 1,
						email: 'admin@bome.com',
						firstName: 'Admin',
						lastName: 'User',
						role: 'admin',
						emailVerified: true
					};

					const mockToken = 'mock-admin-token-' + Date.now();
					localStorage.setItem('token', mockToken);

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
					emailVerified: data.user.emailVerified || false
				};

				// Store token in localStorage
				localStorage.setItem('token', data.token);

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

		logout: () => {
			localStorage.removeItem('token');
			set({
				user: null,
				token: null,
				isAuthenticated: false
			});
		},

		initialize: () => {
			const token = localStorage.getItem('token');
			if (token) {
				// Check if it's a mock admin token
				if (token.startsWith('mock-admin-token-')) {
					const mockUser: User = {
						id: 1,
						email: 'admin@bome.com',
						firstName: 'Admin',
						lastName: 'User',
						role: 'admin',
						emailVerified: true
					};

					set({
						user: mockUser,
						token,
						isAuthenticated: true
					});
				} else {
					// TODO: Validate token with backend and get user data
					// For now, create a mock regular user
					const mockUser: User = {
						id: 2,
						email: 'user@example.com',
						firstName: 'Regular',
						lastName: 'User',
						role: 'user',
						emailVerified: true
					};

					set({
						user: mockUser,
						token,
						isAuthenticated: true
					});
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