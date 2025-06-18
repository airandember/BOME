<script lang="ts">
	export let size: 'small' | 'medium' | 'large' = 'medium';
	export let color: 'primary' | 'secondary' | 'white' = 'primary';
</script>

<div class="spinner-container" class:small={size === 'small'} class:medium={size === 'medium'} class:large={size === 'large'}>
	<div class="spinner" class:primary={color === 'primary'} class:secondary={color === 'secondary'} class:white={color === 'white'}>
		<div class="spinner-ring"></div>
		<div class="spinner-ring"></div>
		<div class="spinner-ring"></div>
	</div>
</div>

<style>
	.spinner-container {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.spinner {
		position: relative;
		display: inline-block;
	}

	.spinner-ring {
		position: absolute;
		border-radius: 50%;
		border: 2px solid transparent;
		animation: spin 1.2s linear infinite;
	}

	.spinner-ring:nth-child(1) {
		animation-delay: 0s;
	}

	.spinner-ring:nth-child(2) {
		animation-delay: 0.4s;
	}

	.spinner-ring:nth-child(3) {
		animation-delay: 0.8s;
	}

	/* Size variants */
	.spinner-container.small .spinner {
		width: 24px;
		height: 24px;
	}

	.spinner-container.small .spinner-ring {
		width: 24px;
		height: 24px;
		border-width: 2px;
	}

	.spinner-container.medium .spinner {
		width: 40px;
		height: 40px;
	}

	.spinner-container.medium .spinner-ring {
		width: 40px;
		height: 40px;
		border-width: 3px;
	}

	.spinner-container.large .spinner {
		width: 60px;
		height: 60px;
	}

	.spinner-container.large .spinner-ring {
		width: 60px;
		height: 60px;
		border-width: 4px;
	}

	/* Color variants */
	.spinner.primary .spinner-ring {
		border-top-color: var(--primary);
	}

	.spinner.secondary .spinner-ring {
		border-top-color: var(--secondary);
	}

	.spinner.white .spinner-ring {
		border-top-color: var(--white);
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	/* Pulse effect */
	.spinner::after {
		content: '';
		position: absolute;
		top: 50%;
		left: 50%;
		width: 4px;
		height: 4px;
		background: currentColor;
		border-radius: 50%;
		transform: translate(-50%, -50%);
		animation: pulse 1.2s ease-in-out infinite;
	}

	.spinner.primary::after {
		background: var(--primary);
	}

	.spinner.secondary::after {
		background: var(--secondary);
	}

	.spinner.white::after {
		background: var(--white);
	}

	@keyframes pulse {
		0%, 100% {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
		50% {
			opacity: 0.5;
			transform: translate(-50%, -50%) scale(1.2);
		}
	}

	/* Responsive adjustments */
	@media (max-width: 768px) {
		.spinner-container.large .spinner {
			width: 50px;
			height: 50px;
		}

		.spinner-container.large .spinner-ring {
			width: 50px;
			height: 50px;
			border-width: 3px;
		}
	}

	@media (max-width: 480px) {
		.spinner-container.large .spinner {
			width: 40px;
			height: 40px;
		}

		.spinner-container.large .spinner-ring {
			width: 40px;
			height: 40px;
			border-width: 3px;
		}
	}
</style> 