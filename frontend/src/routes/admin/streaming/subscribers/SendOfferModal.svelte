<script lang="ts">
	import type { SubscriptionOffer } from '$lib/services/subscription-offers';
	import { goto } from '$app/navigation';

	export let isOpen = false;
	export let selectedEmails: string[] = [];
	export let offers: SubscriptionOffer[] = [];
	export let isSubmitting = false;

	// Callback props
	export let onSendOffer: (offerId: number, emails: string[]) => void = () => {};
	export let onCancel: () => void = () => {};

	// Selected offer for sending
	let selectedOfferId: number | null = null;

	// Close modal on backdrop click
	function handleBackdropClick(event: MouseEvent) {
		if (event.target === event.currentTarget) {
			handleCancel();
		}
	}

	function handleCancel() {
		onCancel();
	}

	function handleSendOffer() {
		if (selectedOfferId) {
			onSendOffer(selectedOfferId, selectedEmails);
		}
	}

	function handleCreateNewOffer() {
		goto('/admin/streaming/subscriptions');
	}

	// Get active and inactive offers
	$: activeOffers = offers.filter(offer => offer.is_active);
	$: inactiveOffers = offers.filter(offer => !offer.is_active);
</script>

{#if isOpen}
	<div class="modal-backdrop" on:click={handleBackdropClick}>
		<div class="modal-content">
			<div class="modal-header">
				<h2 class="modal-title">Send Offer to {selectedEmails.length} Users</h2>
				<button type="button" class="close-button" on:click={handleCancel} aria-label="Close">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>

			<div class="modal-body">
				<!-- Create New Offer Button -->
				<div class="create-offer-section">
					<button 
						type="button" 
						class="btn btn-primary create-offer-btn"
						on:click={handleCreateNewOffer}
					>
						➕ Create New Offer
					</button>
					<p class="help-text">Create a new offer and redirect to the offers management page</p>
				</div>

				<!-- Active Offers -->
				<div class="offers-section">
					<h3 class="section-title">Active Offers</h3>
					{#if activeOffers.length === 0}
						<p class="no-offers">No active offers available</p>
					{:else}
						<div class="offers-grid">
							{#each activeOffers as offer}
								<div class="offer-card" class:selected={selectedOfferId === offer.id}>
									<input 
										type="radio" 
										id="offer_{offer.id}" 
										name="selectedOffer" 
										value={offer.id}
										bind:group={selectedOfferId}
										class="offer-radio"
									/>
									<label for="offer_{offer.id}" class="offer-label">
										<div class="offer-header">
											<h4 class="offer-name">{offer.off_name}</h4>
											<span class="offer-status active">Active</span>
										</div>
										<div class="offer-details">
											<p class="offer-description">{offer.off_description || 'No description'}</p>
											<div class="offer-discount">
												{offer.off_discount_type === 'percentage' ? `${offer.off_discount_value}%` : `$${offer.off_discount_value}`} off
											</div>
											{#if offer.off_code}
												<div class="offer-code">Code: {offer.off_code}</div>
											{/if}
										</div>
									</label>
								</div>
							{/each}
						</div>
					{/if}
				</div>

				<!-- Inactive Offers -->
				<div class="offers-section">
					<h3 class="section-title">Inactive Offers</h3>
					{#if inactiveOffers.length === 0}
						<p class="no-offers">No inactive offers</p>
					{:else}
						<div class="offers-grid">
							{#each inactiveOffers as offer}
								<div class="offer-card inactive" class:selected={selectedOfferId === offer.id}>
									<input 
										type="radio" 
										id="offer_{offer.id}" 
										name="selectedOffer" 
										value={offer.id}
										bind:group={selectedOfferId}
										class="offer-radio"
									/>
									<label for="offer_{offer.id}" class="offer-label">
										<div class="offer-header">
											<h4 class="offer-name">{offer.off_name}</h4>
											<span class="offer-status inactive">Inactive</span>
										</div>
										<div class="offer-details">
											<p class="offer-description">{offer.off_description || 'No description'}</p>
											<div class="offer-discount">
												{offer.off_discount_type === 'percentage' ? `${offer.off_discount_value}%` : `$${offer.off_discount_value}`} off
											</div>
											{#if offer.off_code}
												<div class="offer-code">Code: {offer.off_code}</div>
											{/if}
										</div>
									</label>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<div class="modal-actions">
				<button
					type="button"
					class="btn btn-secondary"
					on:click={handleCancel}
					disabled={isSubmitting}
				>
					Cancel
				</button>
				<button
					type="button"
					class="btn btn-primary"
					on:click={handleSendOffer}
					disabled={!selectedOfferId || isSubmitting}
				>
					{isSubmitting ? 'Sending...' : `Send Offer to ${selectedEmails.length} Users`}
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
		max-width: 800px;
		width: 100%;
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

	.create-offer-section {
		margin-bottom: 2rem;
		padding: 1rem;
		background: #f8fafc;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
	}

	.create-offer-btn {
		width: 100%;
		padding: 1rem;
		font-size: 1rem;
		font-weight: 600;
	}

	.help-text {
		margin: 0.5rem 0 0 0;
		font-size: 0.875rem;
		color: #6b7280;
		text-align: center;
	}

	.offers-section {
		margin-bottom: 2rem;
	}

	.section-title {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0 0 1rem 0;
		border-bottom: 2px solid #e5e7eb;
		padding-bottom: 0.5rem;
	}

	.offers-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
		gap: 1rem;
	}

	.offer-card {
		border: 2px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1rem;
		transition: all 0.2s ease;
		cursor: pointer;
		position: relative;
	}

	.offer-card:hover {
		border-color: #2563eb;
		background: #f8fafc;
	}

	.offer-card.selected {
		border-color: #2563eb;
		background: #eff6ff;
	}

	.offer-card.inactive {
		opacity: 0.7;
	}

	.offer-radio {
		position: absolute;
		opacity: 0;
		pointer-events: none;
	}

	.offer-label {
		cursor: pointer;
		display: block;
		width: 100%;
		height: 100%;
	}

	.offer-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 0.5rem;
	}

	.offer-name {
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
		flex: 1;
	}

	.offer-status {
		font-size: 0.75rem;
		font-weight: 600;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		text-transform: uppercase;
	}

	.offer-status.active {
		background: #dcfce7;
		color: #166534;
	}

	.offer-status.inactive {
		background: #fef3c7;
		color: #92400e;
	}

	.offer-details {
		font-size: 0.875rem;
		color: #6b7280;
	}

	.offer-description {
		margin: 0 0 0.5rem 0;
		line-height: 1.4;
	}

	.offer-discount {
		font-weight: 600;
		color: #059669;
		margin-bottom: 0.25rem;
	}

	.offer-code {
		font-family: monospace;
		background: #f3f4f6;
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		display: inline-block;
	}

	.no-offers {
		text-align: center;
		color: #6b7280;
		font-style: italic;
		padding: 2rem;
	}

	.modal-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.75rem;
		padding: 1.5rem;
		border-top: 1px solid #e5e7eb;
	}

	.btn {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	.btn:hover {
		transform: translateY(-1px);
	}

	.btn:focus {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
		transform: none !important;
	}

	.btn-secondary {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	.btn-secondary:hover {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	.btn-primary {
		background: #2563eb;
		color: white;
	}

	.btn-primary:hover {
		background: #1d4ed8;
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	.btn-primary:disabled {
		background: #9ca3af;
		box-shadow: none;
	}

	@media (max-width: 768px) {
		.modal-content {
			margin: 0;
			border-radius: 0;
			max-height: 100vh;
		}

		.offers-grid {
			grid-template-columns: 1fr;
		}

		.modal-actions {
			flex-direction: column;
		}

		.modal-actions .btn {
			width: 100%;
		}
	}
</style> 