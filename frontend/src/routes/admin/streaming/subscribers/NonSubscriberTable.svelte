<script lang="ts">
	import type { NonSubscriber } from '$lib/services/streaming-subscribers';
	import { StreamingSubscriberService } from '$lib/services/streaming-subscribers';
	import NonSubscriberEditModal from './NonSubscriberEditModal.svelte';

	export let nonSubscribers: NonSubscriber[] = [];
	export let animationDirection: 'left' | 'right' = 'right';
	export let isAnimating = false;
	export let isTransitioning = false;
	export let roles: any[] = [];
	export let selectedNonSubscribers: Set<number> = new Set();
	export let selectAllNonSubscribers = false;
	export let subscriptionPlans: any[] = [];

	// Remove createEventDispatcher - use callback props instead
	export let onSelectItem: (itemId: number, checked: boolean) => void = () => {};
	export let onNonSubscriberUpdate: (nonSubscriber: NonSubscriber) => void = () => {};

	// Edit modal state
	let showEditModal = false;
	let selectedNonSubscriber: NonSubscriber | null = null;

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
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

	// Handle edit button click
	function handleEditClick(nonSubscriber: NonSubscriber) {
		selectedNonSubscriber = nonSubscriber;
		showEditModal = true;
	}

	// Handle non-subscriber update
	function handleNonSubscriberUpdate(updatedNonSubscriber: NonSubscriber) {
		onNonSubscriberUpdate(updatedNonSubscriber);
		showEditModal = false;
		selectedNonSubscriber = null;
	}

	// Handle cancel edit
	function handleCancelEdit() {
		showEditModal = false;
		selectedNonSubscriber = null;
	}
</script>

<div class="content animate-slide-{animationDirection}" class:animating={isAnimating}>
	{#if isTransitioning}
		<div class="transition-overlay">
			<div class="transition-spinner"></div>
		</div>
	{/if}
	{#if nonSubscribers.length === 0}
		<div class="empty-state">
			<p>No non-subscribers found</p>
		</div>
	{:else}
		<div class="table-container">
			<table class="data-table">
				<thead>
					<tr>
						<th class="checkbox-header">
							<input 
								type="checkbox" 
								checked={selectAllNonSubscribers}
								on:change={(e) => {
									const target = e.target as HTMLInputElement;
									onSelectItem(-1, target?.checked || false);
								}}
							/>
						</th>
						<th>User</th>
						<th>Email</th>
						<th>Verified</th>
						<th>Role</th>
						<th>Has Subbed</th>
						<th>Last Login</th>
						<th>Created</th>
						<th>Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each nonSubscribers as nonSubscriber}
						<tr>
							<td class="checkbox-cell">
								<input 
									type="checkbox" 
									checked={selectedNonSubscribers.has(nonSubscriber.id)}
									on:change={(e) => {
										const target = e.target as HTMLInputElement;
										onSelectItem(nonSubscriber.id, target?.checked || false);
									}}
								/>
							</td>
							<td>
								<div class="user-info">
									<div class="user-avatar">
										{(nonSubscriber.first_name || '').charAt(0)}{(nonSubscriber.last_name || '').charAt(0)}
									</div>
									<div class="user-details">
										<div class="user-name">
											{StreamingSubscriberService.formatNonSubscriberName(nonSubscriber)}
										</div>
										<div class="user-id">ID: {nonSubscriber.id}</div>
									</div>
								</div>
							</td>
							<td>
								<div class="email-cell">
									<span class="email">{nonSubscriber.email}</span>
								</div>
							</td>
							<td>
								<div class="status-cell">
									{#if nonSubscriber.email_verified}
										<span class="verified-badge">✓</span>
									{:else} 
										<span class="unverified-badge">🚫</span>
									{/if}
								</div>
							</td>
							<td>
								<div class="role-badge" style="background: {getRoleColor(nonSubscriber.role)}20; color: {getRoleColor(nonSubscriber.role)}">
									{getRoleIcon(nonSubscriber.role)} {getRoleName(nonSubscriber.role)}
								</div>
							</td>
							<td>
								<span class="has-subbed">
									{#if nonSubscriber.has_subscription_history}
										<span class="has-subbed-badge has-subbed-true">Yes</span>
									{:else}
										<span class="has-subbed-badge has-subbed-false">No</span>
									{/if}
								</span>
							</td>
							<td>
								<span class="last-login">
									{#if nonSubscriber.last_login}
										{formatDate(nonSubscriber.last_login)}
									{:else}
										Never
									{/if}
								</span>
							</td>
							<td>
								<span class="created-date">
									{formatDate(nonSubscriber.created_at)}
								</span>
							</td>
							<td>
								<div class="action-buttons">
									<button class="btn btn-sm btn-secondary" on:click={() => handleEditClick(nonSubscriber)}>
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

{#if showEditModal && selectedNonSubscriber}
	<NonSubscriberEditModal 
		bind:isOpen={showEditModal}
		nonSubscriber={selectedNonSubscriber} 
		{subscriptionPlans} 
		onSave={handleNonSubscriberUpdate} 
		onCancel={handleCancelEdit} 
	/>
{/if}

<style>
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
		appearance: none;
		-webkit-appearance: none;
		position: relative;
		transition: all 0.2s ease;
	}

	.checkbox-header input[type="checkbox"]:checked,
	.checkbox-cell input[type="checkbox"]:checked {
		background: #2563eb;
		border-color: #2563eb;
		position: relative;
	}

	.checkbox-header input[type="checkbox"]:checked::after,
	.checkbox-cell input[type="checkbox"]:checked::after {
		content: '✓';
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		color: white;
		font-size: 0.75rem;
		font-weight: bold;
	}

	.checkbox-header input[type="checkbox"]:hover,
	.checkbox-cell input[type="checkbox"]:hover {
		border-color: #2563eb;
		background: #f8fafc;
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

	.last-login {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.created-date {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.has-subbed {
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.has-subbed-badge {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 500;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.has-subbed-true {
		background: #dcfce7;
		color: #166534;
	}

	.has-subbed-false {
		background: #fef2f2;
		color: #991b1b;
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
