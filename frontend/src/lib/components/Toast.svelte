<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { fly, fade } from 'svelte/transition';

	export let type: 'success' | 'error' | 'warning' | 'info' = 'info';
	export let title: string = '';
	export let message: string = '';
	export let duration: number = 5000; // 5 seconds default
	export let persistent: boolean = false; // Don't auto-dismiss
	export let showIcon: boolean = true;
	export let showClose: boolean = true;

	const dispatch = createEventDispatcher();

	let visible = true;
	let timeoutId: number;

	// Auto-dismiss after duration (unless persistent)
	onMount(() => {
		if (!persistent && duration > 0) {
			timeoutId = window.setTimeout(() => {
				dismiss();
			}, duration);
		}

		return () => {
			if (timeoutId) {
				clearTimeout(timeoutId);
			}
		};
	});

	function dismiss() {
		visible = false;
		dispatch('dismiss');
	}

	function handleClick() {
		dispatch('click');
	}

	// Get icon based on type
	function getIcon(type: string): string {
		switch (type) {
			case 'success': return '✓';
			case 'error': return '✕';
			case 'warning': return '⚠';
			case 'info': return 'ℹ';
			default: return 'ℹ';
		}
	}

	// Get colors based on type
	function getTypeClasses(type: string): string {
		switch (type) {
			case 'success': return 'toast-success';
			case 'error': return 'toast-error';
			case 'warning': return 'toast-warning';
			case 'info': return 'toast-info';
			default: return 'toast-info';
		}
	}
</script>

{#if visible}
	<div 
		class="toast {getTypeClasses(type)}"
		transition:fly="{{ y: -50, duration: 300 }}"
		role="alert"
		aria-live="polite"
		on:click={handleClick}
	>
		{#if showIcon}
			<div class="toast-icon">
				{getIcon(type)}
			</div>
		{/if}
		
		<div class="toast-content">
			{#if title}
				<div class="toast-title">{title}</div>
			{/if}
			<div class="toast-message">{message}</div>
		</div>

		{#if showClose}
			<button 
				class="toast-close"
				on:click|stopPropagation={dismiss}
				aria-label="Close notification"
			>
				✕
			</button>
		{/if}

		{#if !persistent && duration > 0}
			<div class="toast-progress">
				<div class="toast-progress-bar" style="animation-duration: {duration}ms;"></div>
			</div>
		{/if}
	</div>
{/if}

<style>
	.toast {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 16px;
		margin-bottom: 8px;
		border-radius: 8px;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
		backdrop-filter: blur(10px);
		border: 1px solid rgba(255, 255, 255, 0.2);
		min-width: 320px;
		max-width: 500px;
		position: relative;
		overflow: hidden;
		cursor: pointer;
		transition: all 0.2s ease;
	}

	.toast:hover {
		transform: translateY(-2px);
		box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
	}

	.toast-icon {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: bold;
		font-size: 14px;
		color: white;
	}

	.toast-content {
		flex: 1;
		min-width: 0;
	}

	.toast-title {
		font-weight: 600;
		font-size: 14px;
		margin-bottom: 4px;
		color: var(--text-primary);
	}

	.toast-message {
		font-size: 13px;
		line-height: 1.4;
		color: var(--text-secondary);
		word-wrap: break-word;
	}

	.toast-close {
		flex-shrink: 0;
		background: none;
		border: none;
		cursor: pointer;
		padding: 4px;
		border-radius: 4px;
		color: var(--text-secondary);
		font-size: 16px;
		line-height: 1;
		transition: all 0.2s ease;
	}

	.toast-close:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.toast-progress {
		position: absolute;
		bottom: 0;
		left: 0;
		right: 0;
		height: 3px;
		background: rgba(255, 255, 255, 0.2);
	}

	.toast-progress-bar {
		height: 100%;
		background: rgba(255, 255, 255, 0.8);
		animation: progress linear forwards;
		transform-origin: left;
	}

	@keyframes progress {
		from { transform: scaleX(1); }
		to { transform: scaleX(0); }
	}

	/* Type-specific styles */
	.toast-success {
		background: linear-gradient(135deg, rgba(34, 197, 94, 0.9), rgba(22, 163, 74, 0.9));
		border-color: rgba(34, 197, 94, 0.3);
	}

	.toast-success .toast-icon {
		background: rgba(255, 255, 255, 0.2);
	}

	.toast-error {
		background: linear-gradient(135deg, rgba(239, 68, 68, 0.9), rgba(220, 38, 38, 0.9));
		border-color: rgba(239, 68, 68, 0.3);
	}

	.toast-error .toast-icon {
		background: rgba(255, 255, 255, 0.2);
	}

	.toast-warning {
		background: linear-gradient(135deg, rgba(245, 158, 11, 0.9), rgba(217, 119, 6, 0.9));
		border-color: rgba(245, 158, 11, 0.3);
	}

	.toast-warning .toast-icon {
		background: rgba(255, 255, 255, 0.2);
	}

	.toast-info {
		background: linear-gradient(135deg, rgba(59, 130, 246, 0.9), rgba(37, 99, 235, 0.9));
		border-color: rgba(59, 130, 246, 0.3);
	}

	.toast-info .toast-icon {
		background: rgba(255, 255, 255, 0.2);
	}

	/* Dark mode adjustments */
	@media (prefers-color-scheme: dark) {
		.toast {
			box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
		}

		.toast:hover {
			box-shadow: 0 6px 16px rgba(0, 0, 0, 0.4);
		}
	}
</style> 