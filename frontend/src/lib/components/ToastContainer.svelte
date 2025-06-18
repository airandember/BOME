<script lang="ts">
	import { toastStore } from '$lib/toast';
	import Toast from './Toast.svelte';

	function handleDismiss(event: CustomEvent) {
		const id = event.detail;
		toastStore.removeToast(id);
	}
</script>

<div class="toast-container">
	{#each $toastStore.toasts as toast (toast.id)}
		<Toast
			id={toast.id}
			type={toast.type}
			message={toast.message}
			duration={toast.duration}
			on:dismiss={() => toastStore.removeToast(toast.id)}
		/>
	{/each}
</div>

<style>
	.toast-container {
		position: fixed;
		top: var(--space-2xl);
		right: var(--space-2xl);
		z-index: var(--z-toast);
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
		pointer-events: none;
	}

	.toast-container > :global(*) {
		pointer-events: auto;
	}

	/* Responsive */
	@media (max-width: 768px) {
		.toast-container {
			top: var(--space-lg);
			right: var(--space-lg);
			left: var(--space-lg);
		}
	}

	@media (max-width: 480px) {
		.toast-container {
			top: var(--space-md);
			right: var(--space-md);
			left: var(--space-md);
		}
	}
</style> 