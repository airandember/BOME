<script lang="ts">
	const { data } = $props<{ data: any }>();

	const coupons = $derived(data?.coupons || []);
	const couponsCount = $derived(data?.coupons_count || 0);

	// Debug logging
	$effect(() => {
		console.log('=== COUPONS DEBUG ===');
		console.log('Data received:', data);
		console.log('Coupons array:', coupons);
		console.log('Coupons count:', couponsCount);
		console.log('Data type:', typeof data);
		console.log('Data keys:', data ? Object.keys(data) : 'No data');
		console.log('Coupons key exists:', data?.coupons ? 'Yes' : 'No');
		console.log('Coupons key type:', data?.coupons ? typeof data.coupons : 'N/A');
		console.log('Coupons is array:', Array.isArray(data?.coupons));
		console.log('====================');
	});

	function formatCurrency(amount: number, currency: string = 'usd') {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: currency.toUpperCase(),
		}).format(amount / 100); // Convert from cents
	}

	function formatPercentage(percent: number) {
		return `${percent}%`;
	}

	function getDiscountDisplay(coupon: any) {
		if (coupon.PercentOff) {
			return formatPercentage(coupon.PercentOff);
		} else if (coupon.AmountOff) {
			return formatCurrency(coupon.AmountOff, coupon.Currency);
		}
		return 'N/A';
	}

	function getDurationDisplay(duration: string) {
		switch (duration) {
			case 'once':
				return 'Single Use';
			case 'repeating':
				return 'Repeating';
			case 'forever':
				return 'Forever';
			default:
				return duration;
		}
	}

	function getStatusColor(valid: boolean) {
		return valid ? 'var(--success)' : 'var(--error)';
	}

	function getStatusText(valid: boolean) {
		return valid ? 'Active' : 'Inactive';
	}
</script>

<div class="coupons-page">
	<div class="page-header">
		<div class="header-content">
			<h1>🎟️ Coupons</h1>
			<p>Manage discount coupons and promotional codes</p>
		</div>
		<div class="header-stats">
			<div class="stat-card">
				<span class="stat-value">{couponsCount}</span>
				<span class="stat-label">Total Coupons</span>
			</div>
			<div class="stat-card">
				<span class="stat-value">{coupons.filter((c: any) => c.Valid).length}</span>
				<span class="stat-label">Active</span>
			</div>
			<div class="stat-card">
				<span class="stat-value">{coupons.filter((c: any) => c.TimesRedeemed > 0).length}</span>
				<span class="stat-label">Used</span>
			</div>
		</div>
		
		<div class="header-actions">
			<button class="btn btn-secondary" onclick={() => window.location.reload()}>
				🔄 Refresh Page
			</button>
			<button class="btn btn-primary">
				➕ Create Coupon
			</button>
		</div>
	</div>

	{#if coupons.length === 0}
		<div class="empty-state">
			<div class="empty-icon">🎟️</div>
			<h3>No Coupons Found</h3>
			<p>You haven't created any coupons yet. Create your first coupon to start offering discounts to customers.</p>
		</div>
	{:else}
		<div class="coupons-grid">
			{#each coupons as coupon}
				<div class="coupon-card">
					<div class="coupon-header">
						<div class="coupon-name">
							<h3>{coupon.Name || 'Unnamed Coupon'}</h3>
							<span class="coupon-id">#{coupon.ID.slice(-8)}</span>
						</div>
						<div class="coupon-status" style="color: {getStatusColor(coupon.Valid)}">
							● {getStatusText(coupon.Valid)}
						</div>
					</div>

					<div class="coupon-details">
						<div class="detail-row">
							<span class="detail-label">Discount:</span>
							<span class="detail-value discount">
								{getDiscountDisplay(coupon)}
							</span>
						</div>

						<div class="detail-row">
							<span class="detail-label">Duration:</span>
							<span class="detail-value">{getDurationDisplay(coupon.Duration)}</span>
						</div>

						{#if coupon.MaxRedemptions}
							<div class="detail-row">
								<span class="detail-label">Max Uses:</span>
								<span class="detail-value">{coupon.MaxRedemptions}</span>
							</div>
						{/if}

						<div class="detail-row">
							<span class="detail-label">Times Used:</span>
							<span class="detail-value">{coupon.TimesRedeemed}</span>
						</div>

						<div class="detail-row">
							<span class="detail-label">Created:</span>
							<span class="detail-value">{new Date(coupon.CreatedAt).toLocaleDateString()}</span>
						</div>
					</div>

					{#if Object.keys(coupon.Metadata || {}).length > 0}
						<div class="coupon-metadata">
							<details>
								<summary class="metadata-summary">📋 View Metadata</summary>
								<div class="metadata-content">
									{#each Object.entries(coupon.Metadata || {}) as [key, value]}
										<div class="metadata-item">
											<span class="metadata-key">{key}:</span>
											<span class="metadata-value">{value}</span>
										</div>
									{/each}
								</div>
							</details>
						</div>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.coupons-page {
		padding: var(--space-lg);
	}

	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-xl);
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-content h1 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 2rem;
		font-weight: 700;
	}

	.header-content p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
	}

	.header-stats {
		display: flex;
		gap: var(--space-md);
		flex-wrap: wrap;
	}

	.stat-card {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: var(--space-md);
		background: var(--surface);
		border-radius: var(--radius-lg);
		border: 1px solid var(--border);
		min-width: 100px;
	}

	.stat-value {
		font-size: 1.5rem;
		font-weight: bold;
		color: var(--primary);
		margin-bottom: var(--space-xs);
	}

	.stat-label {
		font-size: 0.875rem;
		color: var(--text-muted);
		text-align: center;
	}

	.header-actions {
		display: flex;
		justify-content: flex-end;
		margin-top: var(--space-md);
	}

	.btn {
		padding: var(--space-sm) var(--space-lg);
		border: none;
		border-radius: var(--radius-md);
		font-size: 0.875rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.btn-primary {
		background: var(--primary);
		color: white;
	}

	.btn-primary:hover {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.btn-secondary {
		background: var(--secondary);
		color: white;
	}

	.btn-secondary:hover {
		background: var(--secondary-hover);
		transform: translateY(-1px);
	}

	.empty-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-xl);
		text-align: center;
		min-height: 400px;
	}

	.empty-icon {
		font-size: 4rem;
		margin-bottom: var(--space-lg);
		opacity: 0.5;
	}

	.empty-state h3 {
		margin: 0 0 var(--space-md) 0;
		color: var(--text);
		font-size: 1.5rem;
	}

	.empty-state p {
		margin: 0;
		color: var(--text-muted);
		font-size: 1.1rem;
		max-width: 500px;
	}

	.coupons-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: var(--space-lg);
	}

	.coupon-card {
		background: var(--surface);
		border: 1px solid var(--border);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		transition: all 0.2s ease;
	}

	.coupon-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.coupon-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: var(--space-md);
	}

	.coupon-name h3 {
		margin: 0 0 var(--space-xs) 0;
		color: var(--text);
		font-size: 1.25rem;
		font-weight: 600;
	}

	.coupon-id {
		font-size: 0.75rem;
		color: var(--text-muted);
		font-family: var(--font-mono);
	}

	.coupon-status {
		font-size: 0.875rem;
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: var(--space-xs);
	}

	.coupon-details {
		margin-bottom: var(--space-md);
	}

	.detail-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-sm);
		padding: var(--space-xs) 0;
	}

	.detail-label {
		color: var(--text-muted);
		font-size: 0.875rem;
		font-weight: 500;
	}

	.detail-value {
		color: var(--text);
		font-size: 0.875rem;
		font-weight: 600;
	}

	.detail-value.discount {
		color: var(--success);
		font-size: 1rem;
		font-weight: 700;
	}

	.coupon-metadata {
		border-top: 1px solid var(--border);
		padding-top: var(--space-md);
	}

	.metadata-summary {
		cursor: pointer;
		font-weight: 600;
		color: var(--text-muted);
		user-select: none;
		font-size: 0.875rem;
	}

	.metadata-summary:hover {
		color: var(--text);
	}

	.metadata-content {
		margin-top: var(--space-sm);
		padding: var(--space-sm);
		background: var(--bg-secondary);
		border-radius: var(--radius-md);
	}

	.metadata-item {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: var(--space-xs);
		font-size: 0.8125rem;
	}

	.metadata-key {
		color: var(--text-muted);
		font-weight: 500;
	}

	.metadata-value {
		color: var(--text);
		font-family: var(--font-mono);
		word-break: break-all;
	}

	@media (max-width: 768px) {
		.page-header {
			flex-direction: column;
			align-items: center;
			text-align: center;
		}

		.header-stats {
			justify-content: center;
		}

		.coupons-grid {
			grid-template-columns: 1fr;
		}

		.coupon-card {
			padding: var(--space-md);
		}
	}
</style> 
