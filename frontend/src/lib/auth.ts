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
				// TODO: Validate token with backend
				// For now, just set the token
				set({
					user: null,
					token,
					isAuthenticated: true
				});
			}
		},

		getCurrentUser: () => {
			let currentUser = null;
			subscribe(state => {
				currentUser = state.user;
			})();
			return currentUser;
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