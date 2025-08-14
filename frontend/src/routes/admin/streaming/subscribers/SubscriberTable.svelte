<script lang="ts">
	import type { Subscriber } from '$lib/services/streaming-subscribers';
	import { StreamingSubscriberService } from '$lib/services/streaming-subscribers';
	import { createEventDispatcher } from 'svelte';

	export let subscribers: Subscriber[] = [];
	export let animationDirection: 'left' | 'right' = 'right';
	export let isAnimating = false;
	export let isTransitioning = false;
	export let roles: any[] = [];
	export let selectedSubscribers: Set<number> = new Set();
	export let selectAllSubscribers = false;

	const dispatch = createEventDispatcher<{
		selectItem: { itemId: number; checked: boolean };
		editSubscriber: { subscriber: Subscriber };
	}>();

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getStatusBadgeInfo(status?: string) {
		return StreamingSubscriberService.getSubscriberStatusInfo(status);
	}

	function getRoleName(roleId: string): string {
		const role = roles.find(r => r.id === roleId);
		return role ? role.name : roleId;
	}

	function getRoleColor(roleId: string): string {
		const roleColors: { [key: string]: string } = {
			'super_admin': '#dc2626',
			'admin': '#ea580c',
			'user': '#6b7280',
			'content_editor': '#059669',
			'content_manager': '#0891b2',
			'youtube_manager': '#7c3aed',
			'video_streaming_manager': '#be185d',
			'articles_manager': '#d97706',
			'advertisement_manager': '#059669',
			'events_manager': '#dc2626',
			'marketing_specialist': '#7c3aed',
			'advertiser': '#0891b2',
			'academic_reviewer': '#d97706',
			'research_coordinator': '#059669',
			'security_administrator': '#dc2626',
			'technical_specialist': '#0891b2',
			'analytics_manager': '#7c3aed',
			'financial_administrator': '#059669',
			'user_account_manager': '#d97706',
			'support_specialist': '#0891b2',
			'content_creator': '#7c3aed'
		};
		return roleColors[roleId] || '#6b7280';
	}

	function getRoleIcon(roleId: string): string {
		const roleIcons: { [key: string]: string } = {
			'super_admin': '👑',
			'admin': '🛡️',
			'user': '👤',
			'content_editor': '✏️',
			'content_manager': '📄',
			'youtube_manager': '📹',
			'video_streaming_manager': '🎬',
			'articles_manager': '📰',
			'advertisement_manager': '📢',
			'events_manager': '📅',
			'marketing_specialist': '📊',
			'advertiser': '💰',
			'academic_reviewer': '🎓',
			'research_coordinator': '🔬',
			'security_administrator': '🔒',
			'technical_specialist': '⚙️',
			'analytics_manager': '📈',
			'financial_administrator': '💳',
			'user_account_manager': '👥',
			'support_specialist': '🎧',
			'content_creator': '🎨'
		};
		return roleIcons[roleId] || '⚙️';
	}

	function getStatusBadgeClass(status: string): string {
		const statusClasses: { [key: string]: string } = {
			'active': 'bg-green-100 text-green-800',
			'inactive': 'bg-gray-100 text-gray-800',
			'pending': 'bg-yellow-100 text-yellow-800',
			'suspended': 'bg-red-100 text-red-800'
		};
		return statusClasses[status] || 'bg-gray-100 text-gray-800';
	}

	function getSubscriptionBadgeClass(subscription: string): string {
		const subscriptionClasses: { [key: string]: string } = {
			'premium': 'bg-purple-100 text-purple-800',
			'standard': 'bg-blue-100 text-blue-800',
			'free': 'bg-gray-100 text-gray-800'
		};
		return subscriptionClasses[subscription] || 'bg-gray-100 text-gray-800';
	}

	// Handle edit button click
	function handleEditClick(subscriber: Subscriber) {
		dispatch('editSubscriber', { subscriber });
	}
</script>

<div class="content animate-slide-{animationDirection}" class:animating={isAnimating}>
	{#if isTransitioning}
		<div class="transition-overlay">
			<div class="transition-spinner"></div>
		</div>
	{/if}
	{#if subscribers.length === 0}
		<div class="empty-state">
			<p>No subscribers found</p>
		</div>
	{:else}
		<div class="table-container">
			<table class="data-table">
				<thead>
					<tr>
						<th class="checkbox-header">
							<input 
								type="checkbox" 
								checked={selectAllSubscribers}
								on:change={(e) => {
									const target = e.target as HTMLInputElement;
									if (target) {
										dispatch('selectItem', { itemId: -1, checked: target.checked });
									}
								}}
							/>
						</th>
						<th>User</th>
						<th>Email</th>
						<th>Verified</th>
						<th>Role</th>
						<th>Plan</th>
						<th>Plan Status</th>
						<th>Subscription Start</th>
						<th>Subscription End</th>
						<th>Last Login</th>
						<th>Created</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each subscribers as subscriber}
						<tr>
							<td class="checkbox-cell">
								<input 
									type="checkbox" 
									checked={selectedSubscribers.has(subscriber.id)}
									on:change={(e) => {
										const target = e.target as HTMLInputElement;
										if (target) {
											dispatch('selectItem', { itemId: subscriber.id, checked: target.checked });
										}
									}}
								/>
							</td>
							<td>
								<div class="user-info">
									<!--<div class="user-avatar">
										{(subscriber.first_name || '').charAt(0)}{(subscriber.last_name || '').charAt(0)}
									</div>-->
									<div class="user-details">
										<div class="user-name">
											{StreamingSubscriberService.formatSubscriberName(subscriber)}
										</div>
										<div class="user-id">ID: {subscriber.id}</div>
									</div>
								</div>
							</td>
							<td>
								<div class="email-cell">
									<span class="email">{subscriber.email}</span>
								</div>
							</td>
							<td>
								<div class="status-cell">
									{#if subscriber.email_verified}
										<span class="verified-badge">✓</span>
									{:else} 
										<span class="unverified-badge">🚫</span>
									{/if}
								</div>
							</td>
							<td>
								<div class="role-badge" style="background: {getRoleColor(subscriber.role)}20; color: {getRoleColor(subscriber.role)}">
									{getRoleIcon(subscriber.role)} {getRoleName(subscriber.role)}
								</div>
							</td>
							<td>
								{#if subscriber.plan_name}
									<div class="plan-info">
										<span class="plan-name">{subscriber.plan_name}</span>
										<span class="plan-price">
											{StreamingSubscriberService.formatCurrency(subscriber.plan_price, subscriber.plan_currency)}
										</span>
									</div>
								{:else}
									<span class="no-plan">No Plan</span>
								{/if}
							</td>
							<td>
								<div class="status-cell">
									{#if subscriber.subscription_status}
										{@const statusInfo = getStatusBadgeInfo(subscriber.subscription_status)}
										<span class="status-badge {getStatusBadgeClass(subscriber.subscription_status)}">
											{statusInfo.text}
										</span>
									{:else}
										<span class="status-badge bg-gray-100 text-gray-800">
											No Status
										</span>
									{/if}
								</div>
							</td>
							<td>
								<span class="subscription-start">
									{StreamingSubscriberService.formatSubscriptionDates(subscriber).startDate}
								</span>
							</td>
							<td>
								<span class="subscription-end">
									{StreamingSubscriberService.formatSubscriptionDates(subscriber).endDate}
								</span>
							</td>
							<td>
								<span class="last-login">
									{#if subscriber.last_login}
										{formatDate(subscriber.last_login)}
									{:else}
										Never
									{/if}
								</span>
							</td>
							<td>
								<span class="created-date">
									{formatDate(subscriber.created_at)}
								</span>
							</td>
							<td>
								<div class="action-buttons">
									<button class="btn btn-sm btn-secondary" on:click={() => handleEditClick(subscriber)}>
										Edit
									</button>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>

<style>

    .bg-green-100 {
		background-color: #d1fae5;
	} 


	.content {
		background: white;
		border-radius: 0.5rem;
		overflow: hidden;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		transition: all 0.3s ease;
		position: relative;
	}

	.content.animating {
		transition: transform 0.4s ease-in-out, opacity 0.4s ease-in-out;
	}

	.animate-slide-left {
		animation: slideInFromLeft 0.4s ease-in-out;
	}

	.animate-slide-right {
		animation: slideInFromRight 0.4s ease-in-out;
	}

	@keyframes slideInFromLeft {
		0% {
			transform: translateX(-100%);
			opacity: 0;
		}
		100% {
			transform: translateX(0);
			opacity: 1;
		}
	}

	@keyframes slideInFromRight {
		0% {
			transform: translateX(100%);
			opacity: 0;
		}
		100% {
			transform: translateX(0);
			opacity: 1;
		}
	}

	.transition-overlay {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(255, 255, 255, 0.8);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 10;
		border-radius: 0.5rem;
		backdrop-filter: blur(2px);
	}

	.transition-spinner {
		width: 40px;
		height: 40px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.empty-state {
		text-align: center;
		padding: 3rem 2rem;
		color: #6b7280;
	}

	.table-container {
		overflow-x: auto;
	}

	.data-table {
		width: 100%;
		border-collapse: collapse;
	}

	.data-table th,
	.data-table td {
		padding: 1rem;
		text-align: left;
		border-bottom: 1px solid #e5e7eb;
	}

	.data-table th {
		background: #f9fafb;
		font-weight: 600;
		color: #374151;
		font-size: 0.875rem;
	}

	.checkbox-header,
	.checkbox-cell {
		width: 3rem;
		text-align: center;
		padding: 1rem 0.5rem;
	}

	.checkbox-header input[type="checkbox"],
	.checkbox-cell input[type="checkbox"] {
		width: 1.25rem;
		height: 1.25rem;
		border: 2px solid #d1d5db;
		border-radius: 0.25rem;
		background: white;
		cursor: pointer;
	}

	.checkbox-header input[type="checkbox"]:checked,
	.checkbox-cell input[type="checkbox"]:checked {
		background: #2563eb;
		border-color: #2563eb;
	}

	.data-table tr {
		transition: all 0.2s ease;
	}

	.data-table tr:hover {
		background: #f9fafb;
		box-shadow: 2px 2px 8px rgba(0, 0, 0, 0.1);
	}

	.user-info {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.user-avatar {
		width: 2.5rem;
		height: 2.5rem;
		border-radius: 50%;
		background: #3b82f6;
		color: white;
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		font-size: 0.875rem;
	}

	.user-name {
		font-weight: 500;
		color: #111827;
	}

	.user-id {
		font-size: 0.75rem;
		color: #6b7280;
	}

	.email-cell {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.status-cell {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		text-align: center;
		justify-content: center;
	}

	.verified-badge {
		color: #059669;
		font-weight: bold;
	}

	.unverified-badge {
		color: #dc2626;
		font-weight: bold;
	}

	.role-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
		display: inline-flex;
		align-items: center;
		gap: 0.25rem;
	}

	.plan-info {
		display: flex;
		flex-direction: row;
		align-items: center;
		justify-content: space-between;
		gap: 0.25rem;
	}

	.plan-name {
		font-weight: 500;
		color: #111827;
	}

	.plan-price {
		font-size: 0.875rem;
		color: #059669;
		font-weight: 600;
	}

	.no-plan {
		color: #6b7280;
		font-style: italic;
	}

	.status-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
	}

	.last-login {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.created-date {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.action-buttons {
		display: flex;
		gap: 0.5rem;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
	}

	.btn-sm {
		padding: 0.25rem 0.5rem;
		font-size: 0.75rem;
	}

	@media (max-width: 768px) {
		.data-table {
			font-size: 0.875rem;
		}

		.data-table th,
		.data-table td {
			padding: 0.75rem 0.5rem;
		}
	}
</style> 