// Unified User Service for Frontend
import { apiRequest } from '$lib/auth';

export interface User {
	id: number;
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	role_id?: string;
	email_verified: boolean;
	is_active: boolean;
	has_subbed?: boolean;
	stripe_customer_id?: string;
	created_at: string;
	updated_at: string;
	last_login?: string;
}

export interface UserStats {
	total_users: number;
	active_users: number;
	verified_users: number;
	admin_users: number;
	recent_signups: number;
}

export interface UserFilters {
	role?: string;
	status?: string;
	search?: string;
	email_verified?: boolean;
	page?: number;
	limit?: number;
}

export interface UsersResponse {
	users: User[];
	pagination: {
		page: number;
		limit: number;
		total: number;
		totalPages: number;
	};
}

export interface CreateUserRequest {
	email: string;
	first_name: string;
	last_name: string;
	role: string;
	role_id?: string;
	email_verified?: boolean;
	is_active?: boolean;
	has_subbed?: boolean;
	stripe_customer_id?: string;
}

export interface UpdateUserRequest {
	role: string;
}

export interface BulkCreateUsersRequest {
	users: CreateUserRequest[];
}

class UserService {
	private baseUrl = '/admin/users';

	/**
	 * Get current user profile (self-service)
	 */
	async getCurrentUser(): Promise<User> {
		try {
			const response = await apiRequest('/users/me');
			if (!response.ok) {
				throw new Error(`Failed to fetch current user: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.user;
		} catch (error) {
			console.error('Error fetching current user:', error);
			throw error;
		}
	}

	/**
	 * Update current user profile (self-service)
	 */
	async updateCurrentUser(updates: Partial<User>): Promise<User> {
		try {
			const response = await apiRequest('/users/me', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(updates)
			});
			
			if (!response.ok) {
				throw new Error(`Failed to update user: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.user;
		} catch (error) {
			console.error('Error updating current user:', error);
			throw error;
		}
	}

	/**
	 * Get all users (admin only)
	 */
	async getAllUsers(filters?: UserFilters): Promise<UsersResponse> {
		try {
			const params = new URLSearchParams();
			
			if (filters?.role) params.append('role', filters.role);
			if (filters?.status) params.append('status', filters.status);
			if (filters?.search) params.append('search', filters.search);
			if (filters?.email_verified !== undefined) params.append('email_verified', filters.email_verified.toString());
			if (filters?.page) params.append('page', filters.page.toString());
			if (filters?.limit) params.append('limit', filters.limit.toString());

			const response = await apiRequest(`${this.baseUrl}?${params}`);
			if (!response.ok) {
				throw new Error(`Failed to fetch users: ${response.statusText}`);
			}
			
			const data = await response.json();
			return {
				users: data.users,
				pagination: data.pagination
			};
		} catch (error) {
			console.error('Error fetching users:', error);
			throw error;
		}
	}

	/**
	 * Get user by ID (admin only)
	 */
	async getUserById(id: number): Promise<User> {
		try {
			const response = await apiRequest(`${this.baseUrl}/${id}`);
			if (!response.ok) {
				throw new Error(`Failed to fetch user: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.user;
		} catch (error) {
			console.error('Error fetching user by ID:', error);
			throw error;
		}
	}

	/**
	 * Create new user (admin only)
	 */
	async createUser(userData: CreateUserRequest): Promise<User> {
		try {
			const response = await apiRequest(this.baseUrl, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(userData)
			});
			
			if (!response.ok) {
				throw new Error(`Failed to create user: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.user;
		} catch (error) {
			console.error('Error creating user:', error);
			throw error;
		}
	}

	/**
	 * Update user (admin only)
	 */
	async updateUser(id: number, updates: UpdateUserRequest): Promise<void> {
		try {
			const response = await apiRequest(`${this.baseUrl}/${id}`, {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify(updates)
			});
			
			if (!response.ok) {
				throw new Error(`Failed to update user: ${response.statusText}`);
			}
		} catch (error) {
			console.error('Error updating user:', error);
			throw error;
		}
	}

	/**
	 * Delete user (admin only)
	 */
	async deleteUser(id: number): Promise<void> {
		try {
			const response = await apiRequest(`${this.baseUrl}/${id}`, {
				method: 'DELETE'
			});
			
			if (!response.ok) {
				throw new Error(`Failed to delete user: ${response.statusText}`);
			}
		} catch (error) {
			console.error('Error deleting user:', error);
			throw error;
		}
	}

	/**
	 * Bulk create users (admin only)
	 */
	async bulkCreateUsers(users: CreateUserRequest[]): Promise<void> {
		try {
			const response = await apiRequest(`${this.baseUrl}/bulk`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ users })
			});
			
			if (!response.ok) {
				throw new Error(`Failed to bulk create users: ${response.statusText}`);
			}
		} catch (error) {
			console.error('Error bulk creating users:', error);
			throw error;
		}
	}

	/**
	 * Get user statistics (admin only)
	 */
	async getUserStats(): Promise<UserStats> {
		try {
			const response = await apiRequest(`${this.baseUrl}/stats`);
			if (!response.ok) {
				throw new Error(`Failed to fetch user stats: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('Error fetching user stats:', error);
			throw error;
		}
	}

	/**
	 * Get available roles (admin only)
	 */
	async getAvailableRoles(): Promise<string[]> {
		try {
			const response = await apiRequest(`${this.baseUrl}/roles`);
			if (!response.ok) {
				throw new Error(`Failed to fetch roles: ${response.statusText}`);
			}
			
			const data = await response.json();
			return data.roles || [];
		} catch (error) {
			console.error('Error fetching roles:', error);
			throw error;
		}
	}
}

// Singleton instance for global use
export const userService = new UserService();
