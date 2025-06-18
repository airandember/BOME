import { writable } from 'svelte/store';

export interface Toast {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	message: string;
	duration?: number;
}

interface ToastStore {
	toasts: Toast[];
}

function createToastStore() {
	const { subscribe, update } = writable<ToastStore>({ toasts: [] });

	function addToast(toast: Omit<Toast, 'id'>) {
		const id = Math.random().toString(36).substr(2, 9);
		const newToast: Toast = {
			...toast,
			id,
			duration: toast.duration ?? 5000
		};

		update(store => ({
			...store,
			toasts: [...store.toasts, newToast]
		}));

		return id;
	}

	function removeToast(id: string) {
		update(store => ({
			...store,
			toasts: store.toasts.filter(toast => toast.id !== id)
		}));
	}

	function clearToasts() {
		update(store => ({
			...store,
			toasts: []
		}));
	}

	return {
		subscribe,
		addToast,
		removeToast,
		clearToasts
	};
}

export const toastStore = createToastStore();

// Convenience functions
export function showToast(message: string, type: Toast['type'] = 'info', duration?: number) {
	return toastStore.addToast({ message, type, duration });
}

export function showSuccess(message: string, duration?: number) {
	return showToast(message, 'success', duration);
}

export function showError(message: string, duration?: number) {
	return showToast(message, 'error', duration);
}

export function showWarning(message: string, duration?: number) {
	return showToast(message, 'warning', duration);
}

export function showInfo(message: string, duration?: number) {
	return showToast(message, 'info', duration);
} 