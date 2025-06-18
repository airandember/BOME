<script lang="ts">
	import { onMount } from 'svelte';
	import { theme } from '$lib/theme';

	let isDark = false;

	onMount(() => {
		// Subscribe to theme changes
		theme.subscribe(state => {
			isDark = state.isDark;
		});
	});

	const toggleTheme = () => {
		theme.toggle();
	};
</script>

<button class="theme-toggle glass" on:click={toggleTheme} aria-label="Toggle theme">
	<div class="toggle-container">
		<svg class="sun-icon" class:active={!isDark} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<circle cx="12" cy="12" r="5"></circle>
			<line x1="12" y1="1" x2="12" y2="3"></line>
			<line x1="12" y1="21" x2="12" y2="23"></line>
			<line x1="4.22" y1="4.22" x2="5.64" y2="5.64"></line>
			<line x1="18.36" y1="18.36" x2="19.78" y2="19.78"></line>
			<line x1="1" y1="12" x2="3" y2="12"></line>
			<line x1="21" y1="12" x2="23" y2="12"></line>
			<line x1="4.22" y1="19.78" x2="5.64" y2="18.36"></line>
			<line x1="18.36" y1="5.64" x2="19.78" y2="4.22"></line>
		</svg>
		
		<svg class="moon-icon" class:active={isDark} viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
			<path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"></path>
		</svg>
		
		<div class="toggle-slider" class:dark={isDark}></div>
	</div>
</button>

<style>
	.theme-toggle {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		cursor: pointer;
		transition: all var(--transition-normal);
		position: relative;
		overflow: hidden;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.theme-toggle:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.theme-toggle:active {
		transform: translateY(0);
	}

	.toggle-container {
		position: relative;
		width: 32px;
		height: 32px;
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		display: flex;
		align-items: center;
		justify-content: center;
		overflow: hidden;
	}

	.sun-icon,
	.moon-icon {
		width: 18px;
		height: 18px;
		position: absolute;
		transition: all var(--transition-bounce);
		color: var(--text-primary);
	}

	.sun-icon {
		opacity: 1;
		transform: scale(1) rotate(0deg);
	}

	.sun-icon.active {
		opacity: 1;
		transform: scale(1) rotate(0deg);
		color: var(--warning);
	}

	.sun-icon:not(.active) {
		opacity: 0.5;
		transform: scale(0.8) rotate(-90deg);
	}

	.moon-icon {
		opacity: 0.5;
		transform: scale(0.8) rotate(90deg);
	}

	.moon-icon.active {
		opacity: 1;
		transform: scale(1) rotate(0deg);
		color: var(--accent);
	}

	.moon-icon:not(.active) {
		opacity: 0.5;
		transform: scale(0.8) rotate(90deg);
	}

	.toggle-slider {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 12px;
		height: 12px;
		background: var(--primary-gradient);
		border-radius: 50%;
		transition: all var(--transition-bounce);
		box-shadow: var(--shadow-sm);
	}

	.toggle-slider.dark {
		transform: translateX(16px);
		background: var(--accent-gradient);
	}

	/* Hover effects */
	.theme-toggle:hover .toggle-slider {
		transform: scale(1.1);
	}

	.theme-toggle:hover .toggle-slider.dark {
		transform: translateX(16px) scale(1.1);
	}

	/* Focus styles */
	.theme-toggle:focus {
		outline: none;
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.2);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.theme-toggle {
			width: 44px;
			height: 44px;
		}

		.toggle-container {
			width: 28px;
			height: 28px;
		}

		.sun-icon,
		.moon-icon {
			width: 16px;
			height: 16px;
		}

		.toggle-slider {
			width: 10px;
			height: 10px;
		}

		.toggle-slider.dark {
			transform: translateX(14px);
		}

		.theme-toggle:hover .toggle-slider.dark {
			transform: translateX(14px) scale(1.1);
		}
	}

	@media (max-width: 480px) {
		.theme-toggle {
			width: 40px;
			height: 40px;
		}

		.toggle-container {
			width: 24px;
			height: 24px;
		}

		.sun-icon,
		.moon-icon {
			width: 14px;
			height: 14px;
		}

		.toggle-slider {
			width: 8px;
			height: 8px;
		}

		.toggle-slider.dark {
			transform: translateX(12px);
		}

		.theme-toggle:hover .toggle-slider.dark {
			transform: translateX(12px) scale(1.1);
		}
	}
</style> 