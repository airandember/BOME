<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { quintOut } from 'svelte/easing';

	export let type: 'success' | 'error' | 'warning' | 'info' = 'info';
	export let message: string;
	export let duration: number = 5000;
	export let onDismiss: () => void = () => {};

	const dispatch = createEventDispatcher();

	let timeoutId: number;

	// Auto-dismiss after duration
	$: if (duration > 0) {
		clearTimeout(timeoutId);
		timeoutId = setTimeout(() => {
			dispatch('dismiss');
		}, duration);
	}

	function handleDismiss() {
		onDismiss();
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' || event.key === ' ') {
			event.preventDefault();
			handleDismiss();
		}
	}

	function getIcon() {
		switch (type) {
			case 'success':
				return `
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
						<polyline points="22,4 12,14.01 9,11.01"></polyline>
					</svg>
				`;
			case 'error':
				return `
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"></circle>
						<line x1="15" y1="9" x2="9" y2="15"></line>
						<line x1="9" y1="9" x2="15" y2="15"></line>
					</svg>
				`;
			case 'warning':
				return `
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"></path>
						<line x1="12" y1="9" x2="12" y2="13"></line>
						<line x1="12" y1="17" x2="12.01" y2="17"></line>
					</svg>
				`;
			case 'info':
			default:
				return `
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"></circle>
						<line x1="12" y1="16" x2="12" y2="12"></line>
						<line x1="12" y1="8" x2="12.01" y2="8"></line>
					</svg>
				`;
		}
	}

	function getTypeStyles() {
		switch (type) {
			case 'success':
				return {
					bg: 'var(--success-bg)',
					border: 'var(--success-color)',
					icon: 'var(--success-color)'
				};
			case 'error':
				return {
					bg: 'var(--error-bg)',
					border: 'var(--error-color)',
					icon: 'var(--error-color)'
				};
			case 'warning':
				return {
					bg: 'var(--warning-bg)',
					border: 'var(--warning-color)',
					icon: 'var(--warning-color)'
				};
			case 'info':
			default:
				return {
					bg: 'var(--primary-bg)',
					border: 'var(--primary-color)',
					icon: 'var(--primary-color)'
				};
		}
	}

	const styles = getTypeStyles();
</script>

<div 
	class="toast glass"
	class:success={type === 'success'}
	class:error={type === 'error'}
	class:warning={type === 'warning'}
	class:info={type === 'info'}
	style="--toast-bg: {styles.bg}; --toast-border: {styles.border}; --toast-icon: {styles.icon}"
	on:click={handleDismiss}
	on:keydown={handleKeydown}
	role="alert"
	aria-live="polite"
	tabindex="0"
	transition:fly={{ y: 50, duration: 300, easing: quintOut }}
>
	<div class="toast-content">
		<div class="toast-icon" class:success={type === 'success'} class:error={type === 'error'} class:warning={type === 'warning'} class:info={type === 'info'}>
			{@html getIcon()}
		</div>
		<div class="toast-message">
			{message}
		</div>
		<button class="toast-close" on:click|stopPropagation={handleDismiss} aria-label="Dismiss notification">
			<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
				<line x1="18" y1="6" x2="6" y2="18"></line>
				<line x1="6" y1="6" x2="18" y2="18"></line>
			</svg>
		</button>
	</div>
	
	{#if duration > 0}
		<div class="toast-progress" style="animation-duration: {duration}ms"></div>
	{/if}
</div>

<style>
	.toast {
		position: relative;
		min-width: 300px;
		max-width: 400px;
		padding: var(--space-lg);
		border-radius: var(--radius-xl);
		border: 1px solid var(--toast-border, rgba(255, 255, 255, 0.1));
		background: var(--toast-bg, var(--bg-glass));
		backdrop-filter: blur(20px);
		-webkit-backdrop-filter: blur(20px);
		box-shadow: var(--shadow-lg);
		cursor: pointer;
		overflow: hidden;
		transition: all var(--transition-normal);
	}

	.toast:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-xl);
	}

	.toast-content {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
	}

	.toast-icon {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: var(--radius-full);
		background: rgba(255, 255, 255, 0.1);
	}

	.toast-icon svg {
		width: 16px;
		height: 16px;
		color: var(--toast-icon, var(--text-primary));
	}

	.toast-message {
		flex: 1;
		color: var(--text-primary);
		font-size: var(--text-sm);
		line-height: var(--leading-relaxed);
		margin: 0;
	}

	.toast-close {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		border: none;
		background: none;
		border-radius: var(--radius-full);
		display: flex;
		align-items: center;
		justify-content: center;
		cursor: pointer;
		transition: all var(--transition-fast);
		color: var(--text-secondary);
	}

	.toast-close:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
		transform: scale(1.1);
	}

	.toast-close svg {
		width: 16px;
		height: 16px;
	}

	.toast-progress {
		position: absolute;
		bottom: 0;
		left: 0;
		height: 3px;
		background: var(--toast-border, var(--primary));
		animation: progress linear forwards;
	}

	@keyframes progress {
		from {
			width: 100%;
		}
		to {
			width: 0%;
		}
	}

	/* Type-specific styles */
	.toast.success {
		border-color: var(--success);
	}

	.toast.success .toast-icon {
		background: rgba(0, 212, 170, 0.1);
	}

	.toast.success .toast-icon svg {
		color: var(--success);
	}

	.toast.error {
		border-color: var(--error);
	}

	.toast.error .toast-icon {
		background: rgba(255, 107, 107, 0.1);
	}

	.toast.error .toast-icon svg {
		color: var(--error);
	}

	.toast.warning {
		border-color: var(--warning);
	}

	.toast.warning .toast-icon {
		background: rgba(255, 167, 38, 0.1);
	}

	.toast.warning .toast-icon svg {
		color: var(--warning);
	}

	.toast.info {
		border-color: var(--primary);
	}

	.toast.info .toast-icon {
		background: rgba(102, 126, 234, 0.1);
	}

	.toast.info .toast-icon svg {
		color: var(--primary);
	}

	/* Responsive */
	@media (max-width: 768px) {
		.toast {
			min-width: 280px;
			max-width: 350px;
			padding: var(--space-md);
		}

		.toast-message {
			font-size: var(--text-xs);
		}
	}

	@media (max-width: 480px) {
		.toast {
			min-width: 260px;
			max-width: 320px;
		}

		.toast-content {
			gap: var(--space-sm);
		}

		.toast-icon {
			width: 20px;
			height: 20px;
		}

		.toast-icon svg {
			width: 14px;
			height: 14px;
		}

		.toast-close {
			width: 20px;
			height: 20px;
		}

		.toast-close svg {
			width: 14px;
			height: 14px;
		}
	}
</style> 