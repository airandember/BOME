<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	// Available system settings sections
	const sections = [
		{
			id: 'support',
			title: 'Support Configuration',
			description: 'Configure support contact information displayed to users',
			icon: '📞',
			path: '/admin/system/support',
			color: '#3b82f6'
		}
		// Future sections can be added here:
		// {
		//   id: 'features',
		//   title: 'Feature Flags',
		//   description: 'Enable or disable application features',
		//   icon: '🎛️',
		//   path: '/admin/system/features',
		//   color: '#10b981'
		// }
	];

	function navigateTo(path: string) {
		goto(path);
	}
</script>

<svelte:head>
	<title>System Settings | BOME Admin</title>
</svelte:head>

<div class="page-container">
	<div class="page-header">
		<h1>⚙️ System Settings</h1>
		<p class="subtitle">Configure application-wide settings and preferences</p>
	</div>

	<div class="sections-grid">
		{#each sections as section}
			<button class="section-card" on:click={() => navigateTo(section.path)}>
				<div class="section-icon" style="background: {section.color}20; color: {section.color}">
					{section.icon}
				</div>
				<div class="section-content">
					<h2>{section.title}</h2>
					<p>{section.description}</p>
				</div>
				<div class="section-arrow">→</div>
			</button>
		{/each}
	</div>

	<div class="info-box">
		<h3>💡 About System Settings</h3>
		<p>
			System settings are application-wide configurations that affect how your BOME instance
			operates. These settings are stored in the database and can be updated at any time without
			requiring a deployment.
		</p>
		<p>
			<strong>Current Sections:</strong>
		</p>
		<ul>
			<li>
				<strong>Support Configuration</strong> - Set your support email, phone, and help center URL
				that users see when they need assistance (e.g., when they have multiple active
				subscriptions).
			</li>
		</ul>
		<p>
			<em>More configuration sections will be added here as the system grows.</em>
		</p>
	</div>
</div>

<style>
	.page-container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem;
	}

	.page-header {
		margin-bottom: 3rem;
	}

	.page-header h1 {
		margin: 0 0 0.5rem 0;
		color: #1a1a1a;
		font-size: 2.5rem;
		font-weight: 700;
	}

	.subtitle {
		margin: 0;
		color: #666;
		font-size: 1.125rem;
	}

	.sections-grid {
		display: grid;
		gap: 1.5rem;
		margin-bottom: 3rem;
	}

	.section-card {
		display: grid;
		grid-template-columns: auto 1fr auto;
		align-items: center;
		gap: 1.5rem;
		padding: 1.5rem;
		background: white;
		border: 2px solid #e5e7eb;
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.2s;
		text-align: left;
		width: 100%;
	}

	.section-card:hover {
		border-color: #3b82f6;
		transform: translateY(-2px);
		box-shadow: 0 8px 16px rgba(59, 130, 246, 0.1);
	}

	.section-icon {
		width: 64px;
		height: 64px;
		border-radius: 12px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 2rem;
	}

	.section-content {
		flex: 1;
	}

	.section-content h2 {
		margin: 0 0 0.5rem 0;
		color: #1a1a1a;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.section-content p {
		margin: 0;
		color: #666;
		font-size: 0.95rem;
		line-height: 1.5;
	}

	.section-arrow {
		font-size: 1.5rem;
		color: #9ca3af;
		transition: all 0.2s;
	}

	.section-card:hover .section-arrow {
		color: #3b82f6;
		transform: translateX(4px);
	}

	.info-box {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 2rem;
	}

	.info-box h3 {
		margin: 0 0 1rem 0;
		font-size: 1.25rem;
		color: #1a1a1a;
	}

	.info-box p {
		margin: 0 0 1rem 0;
		color: #555;
		line-height: 1.6;
	}

	.info-box p:last-child {
		margin-bottom: 0;
	}

	.info-box ul {
		margin: 0.5rem 0 1rem 1.5rem;
		color: #555;
	}

	.info-box li {
		margin-bottom: 0.5rem;
		line-height: 1.6;
	}

	.info-box li:last-child {
		margin-bottom: 0;
	}

	@media (max-width: 768px) {
		.page-container {
			padding: 1rem;
		}

		.page-header h1 {
			font-size: 2rem;
		}

		.subtitle {
			font-size: 1rem;
		}

		.section-card {
			grid-template-columns: auto 1fr;
			gap: 1rem;
		}

		.section-arrow {
			display: none;
		}

		.section-icon {
			width: 48px;
			height: 48px;
			font-size: 1.5rem;
		}

		.section-content h2 {
			font-size: 1.1rem;
		}

		.section-content p {
			font-size: 0.9rem;
		}
	}
</style>

