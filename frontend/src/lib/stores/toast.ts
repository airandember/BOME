import { writable } from 'svelte/store';

export interface Toast {
	id: string;
	type: 'success' | 'error' | 'warning' | 'info';
	message: string;
	duration?: number;
}

function createToastStore() {
	const { subscribe, update } = writable<Toast[]>([]);

	function add(toast: Omit<Toast, 'id'>) {
		const id = Math.random().toString(36).substr(2, 9);
		const newToast: Toast = { ...toast, id };
		
		update(toasts => [...toasts, newToast]);

		// Auto remove after duration
		if (toast.duration !== 0) {
			setTimeout(() => {
				remove(id);
			}, toast.duration || 5000);
		}

		return id;
	}

	function remove(id: string) {
		update(toasts => toasts.filter(t => t.id !== id));
	}

	function success(message: string, duration?: number) {
		return add({ type: 'success', message, duration });
	}

	function error(message: string, duration?: number) {
		return add({ type: 'error', message, duration });
	}

	function warning(message: string, duration?: number) {
		return add({ type: 'warning', message, duration });
	}

	function info(message: string, duration?: number) {
		return add({ type: 'info', message, duration });
	}

	function clear() {
		update(() => []);
	}

	return {
		subscribe,
		add,
		remove,
		success,
		error,
		warning,
		info,
		clear
	};
}

export const toasts = createToastStore(); 