import { writable } from 'svelte/store';

export interface ToastNotification {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	title?: string;
	message: string;
	duration?: number;
	persistent?: boolean;
	showIcon?: boolean;
	showClose?: boolean;
	createdAt: Date;
}

interface ToastStore {
	notifications: ToastNotification[];
}

function createToastStore() {
	const { subscribe, set, update } = writable<ToastStore>({
		notifications: []
	});

	// Generate unique ID for each toast
	function generateId(): string {
		return `toast-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
	}

	return {
		subscribe,

		// Add a new toast notification
		add(notification: Omit<ToastNotification, 'id' | 'createdAt'>): string {
			const id = generateId();
			const toast: ToastNotification = {
				id,
				createdAt: new Date(),
				duration: 5000, // Default 5 seconds
				persistent: false,
				showIcon: true,
				showClose: true,
				...notification
			};

			update(store => ({
				notifications: [...store.notifications, toast]
			}));

			return id;
		},

		// Remove a toast by ID
		remove(id: string) {
			update(store => ({
				notifications: store.notifications.filter(n => n.id !== id)
			}));
		},

		// Clear all notifications
		clear() {
			set({ notifications: [] });
		},

		// Convenience methods for different types
		success(message: string, options?: Partial<Omit<ToastNotification, 'id' | 'type' | 'message' | 'createdAt'>>): string {
			return this.add({
				type: 'success',
				message,
				...options
			});
		},

		error(message: string, options?: Partial<Omit<ToastNotification, 'id' | 'type' | 'message' | 'createdAt'>>): string {
			return this.add({
				type: 'error',
				message,
				persistent: true, // Errors should be persistent by default
				...options
			});
		},

		warning(message: string, options?: Partial<Omit<ToastNotification, 'id' | 'type' | 'message' | 'createdAt'>>): string {
			return this.add({
				type: 'warning',
				message,
				duration: 7000, // Warnings last a bit longer
				...options
			});
		},

		info(message: string, options?: Partial<Omit<ToastNotification, 'id' | 'type' | 'message' | 'createdAt'>>): string {
			return this.add({
				type: 'info',
				message,
				...options
			});
		},

		// API response handlers
		handleApiResponse<T>(response: { data?: T; error?: string }, successMessage?: string): T | null {
			if (response.error) {
				this.error(response.error, {
					title: 'API Error'
				});
				return null;
			}

			if (successMessage && response.data) {
				this.success(successMessage);
			}

			return response.data || null;
		},

		// Network error handler
		handleNetworkError(error: Error) {
			this.error('Network connection failed. Please check your internet connection.', {
				title: 'Connection Error',
				persistent: true
			});
		},

		// Loading state handler
		loading(message: string = 'Loading...'): string {
			return this.info(message, {
				persistent: true,
				showClose: false
			});
		},

		// Update existing toast (useful for loading states)
		update(id: string, updates: Partial<ToastNotification>) {
			update(store => ({
				notifications: store.notifications.map(n => 
					n.id === id ? { ...n, ...updates } : n
				)
			}));
		}
	};
}

export const toastStore = createToastStore(); 