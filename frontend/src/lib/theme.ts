import { writable } from 'svelte/store';

interface ThemeState {
	isDark: boolean;
}

// Create the theme store
function createThemeStore() {
	const { subscribe, set, update } = writable<ThemeState>({ isDark: false });

	// Initialize theme from localStorage or system preference
	function init() {
		if (typeof window === 'undefined') return;

		const savedTheme = localStorage.getItem('theme');
		let isDark = false;

		if (savedTheme) {
			isDark = savedTheme === 'dark';
		} else {
			// Check system preference
			isDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
		}

		set({ isDark });
		applyTheme(isDark);
	}

	// Apply theme to document
	function applyTheme(isDark: boolean) {
		if (typeof document === 'undefined') return;

		const root = document.documentElement;
		
		if (isDark) {
			root.setAttribute('data-theme', 'dark');
		} else {
			root.setAttribute('data-theme', 'light');
		}
	}

	// Toggle theme
	function toggle() {
		update(state => {
			const newState = { isDark: !state.isDark };
			
			// Save to localStorage
			if (typeof window !== 'undefined') {
				localStorage.setItem('theme', newState.isDark ? 'dark' : 'light');
			}
			
			// Apply theme
			applyTheme(newState.isDark);
			
			return newState;
		});
	}

	// Set specific theme
	function setTheme(isDark: boolean) {
		set({ isDark });
		
		// Save to localStorage
		if (typeof window !== 'undefined') {
			localStorage.setItem('theme', isDark ? 'dark' : 'light');
		}
		
		// Apply theme
		applyTheme(isDark);
	}

	return {
		subscribe,
		init,
		toggle,
		setTheme
	};
}

export const theme = createThemeStore();

// Initialize theme on client side
if (typeof window !== 'undefined') {
	theme.init();
} 