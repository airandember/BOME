<script lang="ts">
	import type { NonSubscriber } from '$lib/services/streaming-subscribers';
	import { StreamingSubscriberService } from '$lib/services/streaming-subscribers';
	import { showToast } from '$lib/toast';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import { currentUser } from '$lib/auth';
	import { get } from 'svelte/store';

	export let isOpen = false;
	export let nonSubscriber: NonSubscriber | null = null;
	export let subscriptionPlans: any[] = [];
	export let isSubmitting = false;

	// Callback props
	export let onSave: (nonSubscriber: NonSubscriber) => void = () => {};
	export let onCancel: () => void = () => {};

	// State for subscriber history
	let subscriberHistory: any[] = [];
	let historyLoading = false;

	// State for notes
	let showAddNote = false;
	let newNote = '';
	let noteCategory = 'general';
	let isAddingNote = false;

	// Note categories
	const noteCategories = [
		{ value: 'general', label: 'General' },
		{ value: 'support', label: 'Support' },
		{ value: 'billing', label: 'Billing' },
		{ value: 'account_management', label: 'Account Management' },
		{ value: 'technical', label: 'Technical' },
		{ value: 'other', label: 'Other' }
	];

	// Computed properties to separate history and notes
	$: generalHistory = subscriberHistory
		.filter(item => item.action !== 'note_added')
		.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());
	$: notesHistory = subscriberHistory
		.filter(item => item.action === 'note_added')
		.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

	// Load subscriber history when modal opens
	$: if (isOpen && nonSubscriber) {
		loadSubscriberHistory();
	}

	async function loadSubscriberHistory() {
		if (!nonSubscriber) return;
		
		try {
			historyLoading = true;
			subscriberHistory = await StreamingSubscriberService.getSubscriberHistory(nonSubscriber.id);
		} catch (error) {
			console.error('Error loading subscriber history:', error);
			subscriberHistory = [];
		} finally {
			historyLoading = false;
		}
	}

	// Handle suspend subscriber
	async function handleSuspend() {
		if (!nonSubscriber) return;

		try {
			isSubmitting = true;
			const updatedNonSubscriber = await StreamingSubscriberService.suspendSubscriber(nonSubscriber.id);
			onSave(updatedNonSubscriber as NonSubscriber);
			showToast('Non-subscriber suspended successfully', 'success');
		} catch (error) {
			console.error('Error suspending non-subscriber:', error);
			showToast('Failed to suspend non-subscriber', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Handle activate subscriber
	async function handleActivate() {
		if (!nonSubscriber) return;

		try {
			isSubmitting = true;
			const updatedNonSubscriber = await StreamingSubscriberService.activateSubscriber(nonSubscriber.id);
			onSave(updatedNonSubscriber as NonSubscriber);
			showToast('Non-subscriber activated successfully', 'success');
		} catch (error) {
			console.error('Error activating non-subscriber:', error);
			showToast('Failed to activate non-subscriber', 'error');
		} finally {
			isSubmitting = false;
		}
	}

	// Handle adding a note
	async function handleAddNote() {
		if (!nonSubscriber || !newNote.trim()) return;

		try {
			isAddingNote = true;
			
			// Get current user info from auth store
			const currentUserData = get(currentUser);
			
			if (!currentUserData) {
				throw new Error('User not authenticated');
			}
			
			await StreamingSubscriberService.addAdminNote(nonSubscriber.id, {
				admin_id: currentUserData.id,
				admin_name: `${currentUserData.first_name} ${currentUserData.last_name}`,
				note: newNote.trim(),
				category: noteCategory
			});

			showToast('Note added successfully', 'success');
			
			// Clear form and reload history
			newNote = '';
			noteCategory = 'general';
			showAddNote = false;
			
			// Reload history to show the new note
			await loadSubscriberHistory();
		} catch (error) {
			console.error('Error adding note:', error);
			showToast('Failed to add note', 'error');
		} finally {
			isAddingNote = false;
		}
	}

	// Handle cancel
	function handleCancel() {
		onCancel();
	}

	// Close modal on backdrop click
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			handleCancel();
		}
	}

	// Get plan name for display
	function getPlanName(planId: number | null | undefined): string {
		if (!planId) return 'No Plan';
		const plan = subscriptionPlans.find(p => p.id === planId.toString());
		return plan ? plan.name : `Plan ${planId}`;
	}

	// Format date
	function formatDate(dateString: string | null): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	// Get status badge class
	function getStatusBadgeClass(status: string): string {
		const statusClasses: { [key: string]: string } = {
			'active': 'bg-green-100 text-green-800',
			'inactive': 'bg-gray-100 text-gray-800',
			'suspended': 'bg-red-100 text-red-800'
		};
		return statusClasses[status] || 'bg-gray-100 text-gray-800';
	}

	// Get history action class
	function getHistoryActionClass(action: string): string {
		const actionClasses: { [key: string]: string } = {
			'account_created': 'text-blue-600',
			'last_login': 'text-gray-600',
			'subscription_started': 'text-green-600',
			'subscription_ended': 'text-red-600',
			'offer_collected': 'text-purple-600',
			'suspend': 'text-red-600',
			'activate': 'text-green-600',
			'note_added': 'text-blue-600'
		};
		return actionClasses[action] || 'text-gray-600';
	}

	// Get history action icon
	function getHistoryActionIcon(action: string): string {
		const actionIcons: { [key: string]: string } = {
			'account_created': '👤',
			'last_login': '🔐',
			'subscription_started': '✅',
			'subscription_ended': '❌',
			'offer_collected': '🎁',
			'suspend': '⚠️',
			'activate': '✅',
			'note_added': '📝'
		};
		return actionIcons[action] || 'ℹ️';
	}

	// Get history action text
	function getHistoryActionText(action: string): string {
		const actionTexts: { [key: string]: string } = {
			'account_created': 'Account Created',
			'last_login': 'Last Login',
			'subscription_started': 'Subscription Started',
			'subscription_ended': 'Subscription Ended',
			'offer_collected': 'Offer Collected',
			'suspend': 'Suspended',
			'activate': 'Activated',
			'note_added': 'Note Added'
		};
		return actionTexts[action] || action.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
	}
</script>

{#if isOpen && nonSubscriber}
	<div class="modal-backdrop" on:click={handleBackdropClick}>
		<div class="modal-content">
			<div class="modal-header">
				<h2 class="modal-title">Non-Subscriber Information</h2>
				<button type="button" class="close-button" on:click={handleCancel} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
				{#if (nonSubscriber as any).subscription_status === 'suspended'}
					<button
						type="button"
						class="btn btn-success"
						on:click={handleActivate}
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Activating...' : 'Activate User'}
					</button>
				{:else}
					<button
						type="button"
						class="btn btn-warning"
						on:click={handleSuspend}
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Suspending...' : 'Suspend User'}
					</button>
				{/if}
			</div>

			<div class="modal-body">
				<!-- Basic Information Section -->
				<div class="info-section">
					<h3 class="section-title">Basic Information</h3>
					<div class="info-grid">
						<div class="info-item">
							<label class="info-label">Name</label>
							<span class="info-value">{nonSubscriber.first_name} {nonSubscriber.last_name}</span>
						</div>
						<div class="info-item">
							<label class="info-label">Email</label>
							<span class="info-value">{nonSubscriber.email}</span>
						</div>
						<div class="info-item">
							<label class="info-label">Email Verified</label>
							<span class="info-value">
								{#if nonSubscriber.email_verified}
									<span class="verified-badge">✓ Verified</span>
								{:else}
									<span class="unverified-badge">🚫 Not Verified</span>
								{/if}
							</span>
						</div>
						<div class="info-item">
							<label class="info-label">Created</label>
							<span class="info-value">{formatDate(nonSubscriber.created_at)}</span>
						</div>
					</div>
				</div>

				<!-- Subscription Information Section -->
				<div class="info-section">
					<h3 class="section-title">Subscription Information</h3>
					<div class="info-grid">
						<div class="info-item">
							<label class="info-label">Current Plan</label>
							<span class="info-value">{getPlanName(nonSubscriber.sub_id)}</span>
						</div>
						<div class="info-item">
							<label class="info-label">Subscription Status</label>
							<span class="info-value">
								<span class="status-badge bg-gray-100 text-gray-800">
									Non-Subscriber
								</span>
							</span>
						</div>
					</div>
				</div>
				<div class="history-container">
					<!-- Subscriber History Section -->
					<div class="info-section">
						<h3 class="section-title">Subscription History</h3>
						{#if historyLoading}
							<div class="loading-container">
								<LoadingSpinner />
								<p>Loading history...</p>
							</div>
						{:else if generalHistory.length === 0}
							<div class="empty-history">
								<p>No history available</p>
							</div>
						{:else}
							<div class="history-list">
								{#each generalHistory as historyItem}
									<div class="history-item">
										<div class="history-header">
											<span class="history-action {getHistoryActionClass(historyItem.action)}">
												{getHistoryActionIcon(historyItem.action)} {getHistoryActionText(historyItem.action)}
											</span>
											<span class="history-date">{formatDate(historyItem.timestamp)}</span>
										</div>
										<div class="history-description">{historyItem.description}</div>
										{#if historyItem.metadata && Object.keys(historyItem.metadata).length > 0}
											<div class="history-metadata">
												{#if historyItem.metadata.plan_name}
													<div class="metadata-item">
														<span class="metadata-label">Plan:</span>
														<span class="metadata-value">{historyItem.metadata.plan_name}</span>
														{#if historyItem.metadata.plan_price}
															<span class="metadata-price">${historyItem.metadata.plan_price} {historyItem.metadata.plan_currency || 'USD'}</span>
														{/if}
													</div>
												{/if}
												{#if historyItem.metadata.offer_name}
													<div class="metadata-item">
														<span class="metadata-label">Offer:</span>
														<span class="metadata-value">{historyItem.metadata.offer_name}</span>
														{#if historyItem.metadata.offer_description}
															<span class="metadata-description">{historyItem.metadata.offer_description}</span>
														{/if}
													</div>
												{/if}
												{#if historyItem.metadata.category}
													<div class="metadata-item">
														<span class="metadata-label">Category:</span>
														<span class="metadata-value">{historyItem.metadata.category}</span>
													</div>
												{/if}
												{#if historyItem.metadata.admin_name}
													<div class="metadata-item">
														<span class="metadata-label">Added by:</span>
														<span class="metadata-value">{historyItem.metadata.admin_name}</span>
													</div>
												{/if}
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}
					</div>

					<!-- Notes Section -->
					<div class="info-section">
						<div class="section-header">
							<h3 class="section-title">Notes</h3>
							<button
								type="button"
								class="btn btn-primary btn-sm"
								on:click={() => showAddNote = !showAddNote}
								disabled={isAddingNote}
							>
								{showAddNote ? 'Cancel' : 'Add Note'}
							</button>
						</div>

						{#if notesHistory.length === 0}
							<div class="empty-notes">
								<p>No notes available</p>
							</div>
						{:else}
							<div class="notes-list">
								{#each notesHistory as noteItem}
									<div class="note-item">
										<div class="note-header">
											<span class="note-category">
												{noteItem.metadata?.category || 'General'}
											</span>
											<span class="note-date">{formatDate(noteItem.timestamp)}</span>
										</div>
										<div class="note-content">{noteItem.description}</div>
										{#if noteItem.metadata?.admin_name}
											<div class="note-author">
												<span class="note-author-label">Added by:</span>
												<span class="note-author-name">{noteItem.metadata.admin_name}</span>
											</div>
										{/if}
									</div>
								{/each}
							</div>
						{/if}

						{#if showAddNote}
							<div class="add-note-form">
								<div class="form-group">
									<label for="note-category" class="form-label">Category</label>
									<select
										id="note-category"
										bind:value={noteCategory}
										class="form-select"
										disabled={isAddingNote}
									>
										{#each noteCategories as category}
											<option value={category.value}>{category.label}</option>
										{/each}
									</select>
								</div>
								<div class="form-group">
									<label for="new-note" class="form-label">Note</label>
									<textarea
										id="new-note"
										bind:value={newNote}
										class="form-textarea"
										placeholder="Enter your note here..."
										rows="3"
										disabled={isAddingNote}
									></textarea>
								</div>
								<div class="form-actions">
									<button
										type="button"
										class="btn btn-primary"
										on:click={handleAddNote}
										disabled={isAddingNote || !newNote.trim()}
									>
										{isAddingNote ? 'Adding...' : 'Add Note'}
									</button>
									<button
										type="button"
										class="btn btn-secondary"
										on:click={() => {
											showAddNote = false;
											newNote = '';
											noteCategory = 'general';
										}}
										disabled={isAddingNote}
									>
										Cancel
									</button>
								</div>
							</div>
						{/if}
					</div>
				</div>
			</div>

			<div class="modal-actions">
				<button
					type="button"
					class="btn btn-secondary"
					on:click={handleCancel}
					disabled={isSubmitting}
				>
					Close
				</button>
				
			</div>
		</div>
	</div>
{/if}

<style>
	.modal-backdrop {
		position: fixed;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(0, 0, 0, 0.5);
		display: flex;
		align-items: center;
		justify-content: center;
		z-index: 1000;
		padding: 1rem;
	}

	.modal-content {
		background: white;
		border-radius: 0.5rem;
		box-shadow: 0 20px 25px -5px rgba(0, 0, 0, 0.1), 0 10px 10px -5px rgba(0, 0, 0, 0.04);
		max-width: 1800px;
		width: 100%;
		min-height: 80vh;
		max-height: 90vh;
		overflow: hidden;
		display: flex;
		flex-direction: column;
	}

	.modal-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.modal-title {
		font-size: 1.25rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
	}

	.close-button {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 0.375rem;
		transition: all 0.2s ease;
	}

	.close-button:hover {
		background: #f3f4f6;
		color: #374151;
	}

	.modal-body {
		padding: 1.5rem;
		overflow-y: auto;
		flex: 1;
	}

	.info-section {
		margin-bottom: 2rem;
		padding-bottom: 1.5rem;
		border-bottom: 1px solid #e5e7eb;
	}

	.history-container {
		display: flex;
		flex-direction: row;
		gap: 2rem;
		width: 100%;
		height: 100%;
	}

	.history-container .info-section {
		width: 50vw;
	}

	.section-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1rem 0;
		border-bottom: 2px solid #e5e7eb;
		padding-bottom: 0.5rem;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
		gap: 1.5rem;
	}

	.info-item {
		display: flex;
		flex-direction: column;
	}

	.info-label {
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.info-value {
		font-size: 1rem;
		font-weight: 400;
		color: #4b5563;
		word-break: break-word;
	}

	.verified-badge {
		background-color: #d1fae5;
		color: #065f46;
		padding: 0.25rem 0.75rem;
		border-radius: 0.375rem;
		font-weight: 600;
	}

	.unverified-badge {
		background-color: #fef3c7;
		color: #d97706;
		padding: 0.25rem 0.75rem;
		border-radius: 0.375rem;
		font-weight: 600;
	}

	.status-badge {
		padding: 0.25rem 0.75rem;
		border-radius: 0.375rem;
		font-weight: 600;
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 2rem;
	}

	.empty-history {
		text-align: center;
		padding: 2rem;
		color: #6b7280;
	}

	.history-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.history-item {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		display: flex;
		flex-direction: column;
	}

	.history-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
		font-size: 0.9rem;
		color: #4b5563;
	}

	.history-action {
		font-weight: 600;
		color: #111827;
	}

	.history-date {
		font-weight: 400;
		color: #6b7280;
	}

	.history-description {
		font-size: 0.9rem;
		color: #4b5563;
		word-break: break-word;
	}

	.history-metadata {
		margin-top: 0.5rem;
		padding-top: 0.5rem;
		border-top: 1px dashed #e5e7eb;
		font-size: 0.85rem;
		color: #4b5563;
	}

	.metadata-item {
		margin-bottom: 0.25rem;
	}

	.metadata-label {
		font-weight: 500;
		color: #374151;
	}

	.metadata-value {
		font-weight: 400;
		color: #4b5563;
	}

	.metadata-price {
		font-weight: 600;
		color: #065f46;
	}

	.metadata-description {
		font-style: italic;
		color: #6b7280;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
		background: #f9fafb;
	}

	.btn {
		display: inline-flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		text-decoration: none;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover:not(:disabled) {
		background: #1d4ed8;
	}

	.btn-primary:disabled {
		background: #9ca3af;
		cursor: not-allowed;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover:not(:disabled) {
		background: #e5e7eb;
	}

	.btn-secondary:disabled {
		background: #f3f4f6;
		color: #9ca3af;
		cursor: not-allowed;
	}

	.btn-success {
		background: #059669;
		color: white;
	}

	.btn-success:hover:not(:disabled) {
		background: #047857;
	}

	.btn-success:disabled {
		background: #9ca3af;
		cursor: not-allowed;
	}

	.btn-warning {
		background: #d97706;
		color: white;
	}

	.btn-warning:hover:not(:disabled) {
		background: #b45309;
	}

	.btn-warning:disabled {
		background: #9ca3af;
		cursor: not-allowed;
	}

	@media (max-width: 640px) {
		.modal-content {
			margin: 1rem;
			max-height: calc(100vh - 2rem);
		}

		.modal-header,
		.modal-body,
		.modal-actions {
			padding: 1rem;
		}

		.modal-actions {
			flex-direction: column;
		}

		.btn {
			width: 100%;
			justify-content: center;
		}
	}

	/* Note functionality styles */
	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.btn-sm {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	.add-note-form {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1.5rem;
		margin-top: 1rem;
	}

	.form-group {
		margin-bottom: 1rem;
	}

	.form-label {
		display: block;
		font-size: 0.875rem;
		font-weight: 500;
		color: #374151;
		margin-bottom: 0.5rem;
	}

	.form-select {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		background-color: white;
		font-size: 0.875rem;
		color: #374151;
	}

	.form-select:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.form-textarea {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #d1d5db;
		border-radius: 0.375rem;
		background-color: white;
		font-size: 0.875rem;
		color: #374151;
		resize: vertical;
		min-height: 80px;
	}

	.form-textarea:focus {
		outline: none;
		border-color: #2563eb;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1);
	}

	.form-actions {
		display: flex;
		gap: 0.75rem;
		margin-top: 1rem;
	}

	.form-actions .btn {
		padding: 0.5rem 1rem;
		font-size: 0.875rem;
	}

	/* New styles for notes section */
	.empty-notes {
		text-align: center;
		padding: 2rem;
		color: #6b7280;
	}

	.notes-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.note-item {
		background: #f9fafb;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		display: flex;
		flex-direction: column;
	}

	.note-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
		font-size: 0.9rem;
		color: #4b5563;
	}

	.note-category {
		background-color: #e0e7ff;
		color: #1d4ed8;
		padding: 0.25rem 0.75rem;
		border-radius: 0.375rem;
		font-weight: 600;
		font-size: 0.85rem;
	}

	.note-date {
		font-weight: 400;
		color: #6b7280;
	}

	.note-content {
		font-size: 0.9rem;
		color: #4b5563;
		word-break: break-word;
		margin-bottom: 0.5rem;
	}

	.note-author {
		font-size: 0.85rem;
		color: #6b7280;
	}

	.note-author-label {
		font-weight: 500;
		color: #374151;
	}

	.note-author-name {
		font-weight: 400;
		color: #4b5563;
	}
</style> 
