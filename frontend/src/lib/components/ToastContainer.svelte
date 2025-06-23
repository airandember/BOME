<script lang="ts">
	import { toastStore } from '$lib/stores/toast';
	import Toast from './Toast.svelte';

	$: notifications = $toastStore.notifications;

	function handleDismiss(id: string) {
		toastStore.remove(id);
	}
</script>

<!-- Toast Container - Fixed position, top-right -->
<div class="toast-container" class:has-toasts={notifications.length > 0}>
	{#each notifications as notification (notification.id)}
		<Toast
			type={notification.type}
			title={notification.title || ''}
			message={notification.message}
			duration={notification.duration || 5000}
			persistent={notification.persistent || false}
			showIcon={notification.showIcon !== false}
			showClose={notification.showClose !== false}
			on:dismiss={() => handleDismiss(notification.id)}
			on:click={() => {
				// Optional: Handle toast click events
				console.log('Toast clicked:', notification);
			}}
		/>
	{/each}
</div>

<style>
	.toast-container {
		position: fixed;
		top: 20px;
		right: 20px;
		z-index: 9999;
		display: flex;
		flex-direction: column;
		gap: 8px;
		pointer-events: none;
		max-width: 100vw;
		max-height: 100vh;
		overflow: hidden;
	}

	.toast-container.has-toasts {
		pointer-events: auto;
	}

	/* Responsive positioning */
	@media (max-width: 768px) {
		.toast-container {
			top: 10px;
			right: 10px;
			left: 10px;
			max-width: none;
		}
	}

	@media (max-width: 480px) {
		.toast-container {
			top: 5px;
			right: 5px;
			left: 5px;
		}
	}

	/* Animation for container */
	.toast-container:empty {
		display: none;
	}
</style> 