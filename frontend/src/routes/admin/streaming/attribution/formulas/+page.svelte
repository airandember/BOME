<script lang="ts">
	import { onMount } from 'svelte';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	
	// Types
	interface AttributionFormula {
		id: number;
		name: string;
		description: string;
		formula_type: string;
		formula_config: Record<string, any>;
		attribution_window_days: number;
		min_watch_percentage: number;
		is_active: boolean;
		is_default: boolean;
		created_at: string;
		updated_at: string;
	}
	
	interface NewFormula {
		name: string;
		description: string;
		formula_type: string;
		formula_config: Record<string, any>;
		attribution_window_days: number;
		min_watch_percentage: number;
	}
	
	// State
	let isLoading = $state(true);
	let formulas = $state<AttributionFormula[]>([]);
	let showCreateModal = $state(false);
	let showEditModal = $state(false);
	let selectedFormula = $state<AttributionFormula | null>(null);
	let error = $state<string | null>(null);
	let successMessage = $state<string | null>(null);
	
	// New formula form
	let newFormula = $state<NewFormula>({
		name: '',
		description: '',
		formula_type: 'last_touch',
		formula_config: {},
		attribution_window_days: 7,
		min_watch_percentage: 25.0
	});
	
	// Formula type definitions
	const formulaTypes = [
		{
			type: 'last_touch',
			name: 'Last Touch',
			description: '100% credit to the last video watched before subscription',
			emoji: '🎯',
			config: {}
		},
		{
			type: 'first_touch',
			name: 'First Touch',
			description: '100% credit to the first video that introduced the user',
			emoji: '🚀',
			config: {}
		},
		{
			type: 'linear',
			name: 'Linear',
			description: 'Equal credit distributed across all videos in the journey',
			emoji: '⚖️',
			config: {}
		},
		{
			type: 'time_decay',
			name: 'Time Decay',
			description: 'More credit to recent videos, exponentially decaying for older ones',
			emoji: '📉',
			config: { decay_rate: 0.5, half_life_days: 3.5 }
		},
		{
			type: 'position_based',
			name: 'Position Based',
			description: '40% first, 40% last, 20% distributed to middle videos',
			emoji: '🎪',
			config: { first_weight: 0.4, last_weight: 0.4, middle_weight: 0.2 }
		},
		{
			type: 'custom',
			name: 'Custom Formula',
			description: 'Define your own attribution logic with custom weights',
			emoji: '🔬',
			config: {}
		}
	];
	
	onMount(async () => {
		await loadFormulas();
	});
	
	async function loadFormulas() {
		isLoading = true;
		error = null;
		
		try {
			const response = await fetch('/api/v1/attribution/formulas');
			if (!response.ok) throw new Error('Failed to fetch formulas');
			
			const data = await response.json();
			formulas = data.formulas || [];
			console.log('✅ Loaded', formulas.length, 'attribution formulas');
		} catch (err) {
			console.error('❌ Failed to load formulas:', err);
			error = 'Failed to load attribution formulas';
		} finally {
			isLoading = false;
		}
	}
	
	function openCreateModal() {
		// Reset form
		newFormula = {
			name: '',
			description: '',
			formula_type: 'last_touch',
			formula_config: {},
			attribution_window_days: 7,
			min_watch_percentage: 25.0
		};
		showCreateModal = true;
	}
	
	function openEditModal(formula: AttributionFormula) {
		selectedFormula = formula;
		showEditModal = true;
	}
	
	async function handleCreateFormula() {
		error = null;
		successMessage = null;
		
		try {
			// Get default config for formula type
			const typeInfo = formulaTypes.find(t => t.type === newFormula.formula_type);
			const formulaConfig = typeInfo ? { ...typeInfo.config } : {};
			
			const response = await fetch('/api/v1/attribution/formulas', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({
					...newFormula,
					formula_config: formulaConfig
				})
			});
			
			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to create formula');
			}
			
			successMessage = 'Formula created successfully!';
			showCreateModal = false;
			await loadFormulas();
		} catch (err: any) {
			error = err.message;
		}
	}
	
	async function handleUpdateFormula(formulaId: number, updates: Partial<AttributionFormula>) {
		error = null;
		successMessage = null;
		
		try {
			const response = await fetch(`/api/v1/attribution/formulas/${formulaId}`, {
				method: 'PATCH',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify(updates)
			});
			
			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to update formula');
			}
			
			successMessage = 'Formula updated successfully!';
			await loadFormulas();
		} catch (err: any) {
			error = err.message;
		}
	}
	
	async function handleDeleteFormula(formulaId: number) {
		if (!confirm('Are you sure you want to delete this formula?')) return;
		
		error = null;
		successMessage = null;
		
		try {
			const response = await fetch(`/api/v1/attribution/formulas/${formulaId}`, {
				method: 'DELETE'
			});
			
			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to delete formula');
			}
			
			successMessage = 'Formula deleted successfully!';
			await loadFormulas();
		} catch (err: any) {
			error = err.message;
		}
	}
	
	async function setAsDefault(formulaId: number) {
		await handleUpdateFormula(formulaId, { is_default: true });
	}
	
	async function toggleActive(formulaId: number, currentActive: boolean) {
		await handleUpdateFormula(formulaId, { is_active: !currentActive });
	}
	
	function getFormulaTypeInfo(type: string) {
		return formulaTypes.find(t => t.type === type) || {
			name: type,
			description: 'Custom attribution formula',
			emoji: '🔧',
			config: {}
		};
	}
	
	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleDateString('en-US', { 
			year: 'numeric', 
			month: 'short', 
			day: 'numeric' 
		});
	}
	
	function closeMessages() {
		error = null;
		successMessage = null;
	}
</script>

<svelte:head>
	<title>Attribution Formulas - Admin Dashboard</title>
</svelte:head>

<div class="formulas-page">
	<!-- Header -->
	<div class="page-header">
		<div>
			<h1>🧮 Revenue Attribution Formulas</h1>
			<p class="subtitle">Configure how subscription revenue is attributed to videos</p>
		</div>
		<button class="btn-primary" onclick={openCreateModal}>
			<span class="btn-icon">➕</span>
			Create Formula
		</button>
	</div>
	
	<!-- Messages -->
	{#if error}
		<div class="alert alert-error">
			<span class="alert-icon">❌</span>
			<span class="alert-message">{error}</span>
			<button class="alert-close" onclick={closeMessages}>×</button>
		</div>
	{/if}
	
	{#if successMessage}
		<div class="alert alert-success">
			<span class="alert-icon">✅</span>
			<span class="alert-message">{successMessage}</span>
			<button class="alert-close" onclick={closeMessages}>×</button>
		</div>
	{/if}
	
	<!-- Loading State -->
	{#if isLoading}
		<div class="loading-container">
			<LoadingSpinner />
			<p>Loading attribution formulas...</p>
		</div>
	{:else}
		<!-- Formula Cards -->
		<div class="formulas-grid">
	{#each formulas as formula}
		{@const typeInfo = getFormulaTypeInfo(formula.formula_type)}
		<div class="formula-card {formula.is_default ? 'default' : ''} {formula.is_active ? '' : 'inactive'}">
			<!-- Card Header -->
			<div class="card-header">
						<div class="card-title-row">
							<span class="formula-emoji">{typeInfo.emoji}</span>
							<div>
								<h3>{formula.name}</h3>
								<div class="card-badges">
									{#if formula.is_default}
										<span class="badge badge-primary">⭐ Default</span>
									{/if}
									{#if !formula.is_active}
										<span class="badge badge-inactive">⏸️ Inactive</span>
									{/if}
									<span class="badge badge-type">{typeInfo.name}</span>
								</div>
							</div>
						</div>
						<div class="card-actions">
							<button class="btn-icon-action" onclick={() => openEditModal(formula)} title="Edit formula">
								✏️
							</button>
							<button 
								class="btn-icon-action" 
								onclick={() => toggleActive(formula.id, formula.is_active)}
								title={formula.is_active ? 'Deactivate' : 'Activate'}
							>
								{formula.is_active ? '⏸️' : '▶️'}
							</button>
							{#if !formula.is_default}
								<button 
									class="btn-icon-action" 
									onclick={() => handleDeleteFormula(formula.id)}
									title="Delete formula"
								>
									🗑️
								</button>
							{/if}
						</div>
					</div>
					
					<!-- Card Body -->
					<div class="card-body">
						<p class="formula-description">{formula.description || typeInfo.description}</p>
						
						<!-- Configuration -->
						<div class="formula-config">
							<div class="config-item">
								<span class="config-label">Attribution Window:</span>
								<span class="config-value">{formula.attribution_window_days} days</span>
							</div>
							<div class="config-item">
								<span class="config-label">Min Watch %:</span>
								<span class="config-value">{formula.min_watch_percentage}%</span>
							</div>
							
							{#if Object.keys(formula.formula_config).length > 0}
								<div class="config-advanced">
									<span class="config-label">Formula Config:</span>
									<pre class="config-json">{JSON.stringify(formula.formula_config, null, 2)}</pre>
								</div>
							{/if}
						</div>
						
						<!-- Metadata -->
						<div class="formula-meta">
							<span class="meta-item">Created: {formatDate(formula.created_at)}</span>
							{#if formula.updated_at !== formula.created_at}
								<span class="meta-item">Updated: {formatDate(formula.updated_at)}</span>
							{/if}
						</div>
					</div>
					
					<!-- Card Footer -->
					{#if !formula.is_default}
						<div class="card-footer">
							<button class="btn-secondary" onclick={() => setAsDefault(formula.id)}>
								Set as Default
							</button>
						</div>
					{/if}
				</div>
			{/each}
		</div>
		
		{#if formulas.length === 0}
			<div class="empty-state">
				<div class="empty-icon">🧮</div>
				<h3>No Attribution Formulas</h3>
				<p>Create your first attribution formula to start tracking video revenue impact</p>
				<button class="btn-primary" onclick={openCreateModal}>
					Create Formula
				</button>
			</div>
		{/if}
	{/if}
</div>

<!-- Create Formula Modal -->
{#if showCreateModal}
	<div class="modal-overlay" onclick={() => showCreateModal = false}>
		<div class="modal" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h2>Create Attribution Formula</h2>
				<button class="modal-close" onclick={() => showCreateModal = false}>×</button>
			</div>
			
			<form onsubmit={(e) => { e.preventDefault(); handleCreateFormula(); }}>
				<div class="modal-body">
					<!-- Formula Name -->
					<div class="form-group">
						<label for="formula-name">Formula Name</label>
						<input 
							id="formula-name"
							type="text" 
							bind:value={newFormula.name}
							placeholder="e.g., My Custom Attribution"
							required
						/>
					</div>
					
					<!-- Formula Description -->
					<div class="form-group">
						<label for="formula-description">Description</label>
						<textarea 
							id="formula-description"
							bind:value={newFormula.description}
							placeholder="Describe how this formula works..."
							rows="3"
						></textarea>
					</div>
					
					<!-- Formula Type -->
					<div class="form-group">
						<label for="formula-type">Formula Type</label>
						<select id="formula-type" bind:value={newFormula.formula_type}>
							{#each formulaTypes as type}
								<option value={type.type}>
									{type.emoji} {type.name}
								</option>
							{/each}
						</select>
						<p class="form-hint">
							{formulaTypes.find(t => t.type === newFormula.formula_type)?.description}
						</p>
					</div>
					
					<!-- Attribution Window -->
					<div class="form-row">
						<div class="form-group">
							<label for="window-days">Attribution Window (Days)</label>
							<input 
								id="window-days"
								type="number" 
								bind:value={newFormula.attribution_window_days}
								min="1"
								max="90"
								required
							/>
							<p class="form-hint">How far back to look for video views</p>
						</div>
						
						<div class="form-group">
							<label for="min-watch">Min Watch Percentage</label>
							<input 
								id="min-watch"
								type="number" 
								bind:value={newFormula.min_watch_percentage}
								min="0"
								max="100"
								step="0.1"
								required
							/>
							<p class="form-hint">Minimum % watched to count</p>
						</div>
					</div>
					
			<!-- Formula Preview -->
			<div class="formula-preview">
				<h4>📊 Formula Preview</h4>
				{#if true}
					{@const selectedType = formulaTypes.find(t => t.type === newFormula.formula_type)}
					<div class="preview-content">
						<div class="preview-header">
							<span class="preview-emoji">{selectedType?.emoji}</span>
							<strong>{selectedType?.name}</strong>
						</div>
						<p class="preview-description">{selectedType?.description}</p>
						
						{#if Object.keys(selectedType?.config || {}).length > 0}
							<div class="preview-config">
								<strong>Default Configuration:</strong>
								<pre>{JSON.stringify(selectedType?.config, null, 2)}</pre>
							</div>
						{/if}
						
						<div class="preview-settings">
							<div class="preview-setting">
								<span class="setting-icon">📅</span>
								<span>Looks back <strong>{newFormula.attribution_window_days} days</strong></span>
							</div>
							<div class="preview-setting">
								<span class="setting-icon">⏱️</span>
								<span>Requires <strong>{newFormula.min_watch_percentage}% watch</strong></span>
							</div>
						</div>
					</div>
				{/if}
			</div>
				
				<div class="modal-footer">
					<button type="button" class="btn-secondary" onclick={() => showCreateModal = false}>
						Cancel
					</button>
					<button type="submit" class="btn-primary">
						Create Formula
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}

<!-- Edit Formula Modal -->
{#if showEditModal && selectedFormula}
	<div class="modal-overlay" onclick={() => showEditModal = false}>
		<div class="modal" onclick={(e) => e.stopPropagation()}>
			<div class="modal-header">
				<h2>Edit Formula: {selectedFormula.name}</h2>
				<button class="modal-close" onclick={() => showEditModal = false}>×</button>
			</div>
			
			<div class="modal-body">
				<p class="edit-notice">
					ℹ️ Editing formulas will not recalculate existing attributions.
					This will only affect future attribution calculations.
				</p>
				
				<!-- Edit form here (simplified for now) -->
				<div class="form-group">
					<label>Formula Type</label>
					<p class="readonly-value">{getFormulaTypeInfo(selectedFormula.formula_type).name}</p>
				</div>
				
				<div class="form-group">
					<label>Attribution Window</label>
					<p class="readonly-value">{selectedFormula.attribution_window_days} days</p>
				</div>
				
				<div class="form-group">
					<label>Min Watch Percentage</label>
					<p class="readonly-value">{selectedFormula.min_watch_percentage}%</p>
				</div>
				
				<p class="edit-hint">Full editing interface coming soon. For now, create a new formula.</p>
			</div>
			
			<div class="modal-footer">
				<button class="btn-secondary" onclick={() => showEditModal = false}>
					Close
				</button>
			</div>
		</div>
	</div>
{/if}

<style>
	.formulas-page {
		padding: 2rem;
		max-width: 1400px;
		margin: 0 auto;
	}
	
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 2rem;
		gap: 2rem;
	}
	
	.page-header h1 {
		font-size: 2rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.subtitle {
		color: #666;
		margin: 0;
		font-size: 0.95rem;
	}
	
	.btn-primary {
		padding: 0.75rem 1.5rem;
		background: #d4a574;
		color: white;
		border: none;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.2s;
	}
	
	.btn-primary:hover {
		background: #c89860;
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(212, 165, 116, 0.3);
	}
	
	.btn-secondary {
		padding: 0.75rem 1.5rem;
		background: #f5f5f5;
		color: #333;
		border: 1px solid #e0e0e0;
		border-radius: 8px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}
	
	.btn-secondary:hover {
		background: #e0e0e0;
	}
	
	.btn-icon {
		font-size: 1.2rem;
	}
	
	/* Alerts */
	.alert {
		padding: 1rem 1.5rem;
		border-radius: 8px;
		margin-bottom: 1.5rem;
		display: flex;
		align-items: center;
		gap: 1rem;
	}
	
	.alert-error {
		background: #fee;
		border: 1px solid #fcc;
		color: #c00;
	}
	
	.alert-success {
		background: #efe;
		border: 1px solid #cfc;
		color: #060;
	}
	
	.alert-icon {
		font-size: 1.25rem;
	}
	
	.alert-message {
		flex: 1;
	}
	
	.alert-close {
		background: none;
		border: none;
		font-size: 1.5rem;
		cursor: pointer;
		color: inherit;
		opacity: 0.6;
	}
	
	.alert-close:hover {
		opacity: 1;
	}
	
	/* Loading */
	.loading-container {
		text-align: center;
		padding: 4rem 2rem;
	}
	
	/* Formulas Grid */
	.formulas-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: 1.5rem;
	}
	
	.formula-card {
		background: white;
		border-radius: 12px;
		border: 2px solid #e0e0e0;
		overflow: hidden;
		transition: all 0.3s;
	}
	
	.formula-card:hover {
		transform: translateY(-4px);
		box-shadow: 0 8px 24px rgba(0, 0, 0, 0.1);
	}
	
	.formula-card.default {
		border-color: #d4a574;
		background: linear-gradient(to bottom, #fffbf5, white);
	}
	
	.formula-card.inactive {
		opacity: 0.6;
	}
	
	.card-header {
		padding: 1.5rem;
		border-bottom: 1px solid #f0f0f0;
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
	}
	
	.card-title-row {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
		flex: 1;
	}
	
	.formula-emoji {
		font-size: 2rem;
	}
	
	.card-header h3 {
		font-size: 1.25rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.card-badges {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}
	
	.badge {
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 600;
	}
	
	.badge-primary {
		background: linear-gradient(135deg, #ffd700, #ffed4e);
		color: #1a1a1a;
	}
	
	.badge-inactive {
		background: #f5f5f5;
		color: #999;
	}
	
	.badge-type {
		background: #e3f2fd;
		color: #1976d2;
	}
	
	.card-actions {
		display: flex;
		gap: 0.5rem;
	}
	
	.btn-icon-action {
		background: none;
		border: none;
		font-size: 1.25rem;
		cursor: pointer;
		padding: 0.25rem;
		opacity: 0.6;
		transition: opacity 0.2s;
	}
	
	.btn-icon-action:hover {
		opacity: 1;
	}
	
	.card-body {
		padding: 1.5rem;
	}
	
	.formula-description {
		color: #666;
		margin: 0 0 1rem 0;
		line-height: 1.5;
	}
	
	.formula-config {
		background: #f9f9f9;
		border-radius: 8px;
		padding: 1rem;
		margin: 1rem 0;
	}
	
	.config-item {
		display: flex;
		justify-content: space-between;
		padding: 0.5rem 0;
		border-bottom: 1px solid #eee;
	}
	
	.config-item:last-child {
		border-bottom: none;
	}
	
	.config-label {
		font-weight: 600;
		color: #666;
	}
	
	.config-value {
		font-weight: 600;
		color: #1a1a1a;
	}
	
	.config-advanced {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid #eee;
	}
	
	.config-json {
		background: white;
		border: 1px solid #e0e0e0;
		border-radius: 4px;
		padding: 0.75rem;
		font-size: 0.85rem;
		margin: 0.5rem 0 0 0;
		overflow-x: auto;
	}
	
	.formula-meta {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid #f0f0f0;
		font-size: 0.85rem;
		color: #999;
	}
	
	.meta-item {
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}
	
	.card-footer {
		padding: 1rem 1.5rem;
		border-top: 1px solid #f0f0f0;
		background: #fafafa;
	}
	
	/* Empty State */
	.empty-state {
		text-align: center;
		padding: 4rem 2rem;
	}
	
	.empty-icon {
		font-size: 4rem;
		margin-bottom: 1rem;
		opacity: 0.3;
	}
	
	.empty-state h3 {
		font-size: 1.5rem;
		color: #1a1a1a;
		margin: 0 0 0.5rem 0;
	}
	
	.empty-state p {
		color: #666;
		margin: 0 0 2rem 0;
	}
	
	/* Modal */
	.modal-overlay {
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
		padding: 2rem;
	}
	
	.modal {
		background: white;
		border-radius: 12px;
		max-width: 700px;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
	}
	
	.modal-header {
		padding: 1.5rem 2rem;
		border-bottom: 1px solid #f0f0f0;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}
	
	.modal-header h2 {
		font-size: 1.5rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0;
	}
	
	.modal-close {
		background: none;
		border: none;
		font-size: 2rem;
		cursor: pointer;
		color: #999;
		line-height: 1;
	}
	
	.modal-close:hover {
		color: #333;
	}
	
	.modal-body {
		padding: 2rem;
	}
	
	.modal-footer {
		padding: 1.5rem 2rem;
		border-top: 1px solid #f0f0f0;
		display: flex;
		justify-content: flex-end;
		gap: 1rem;
	}
	
	/* Forms */
	.form-group {
		margin-bottom: 1.5rem;
	}
	
	.form-group label {
		display: block;
		font-weight: 600;
		color: #333;
		margin-bottom: 0.5rem;
	}
	
	.form-group input,
	.form-group textarea,
	.form-group select {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid #e0e0e0;
		border-radius: 6px;
		font-size: 0.95rem;
		font-family: inherit;
	}
	
	.form-group input:focus,
	.form-group textarea:focus,
	.form-group select:focus {
		outline: none;
		border-color: #d4a574;
		box-shadow: 0 0 0 3px rgba(212, 165, 116, 0.1);
	}
	
	.form-hint {
		font-size: 0.85rem;
		color: #999;
		margin: 0.5rem 0 0 0;
	}
	
	.form-row {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}
	
	/* Formula Preview */
	.formula-preview {
		background: linear-gradient(135deg, #f5f7fa 0%, #f9fafb 100%);
		border: 2px dashed #d0d5dd;
		border-radius: 12px;
		padding: 1.5rem;
		margin-top: 2rem;
	}
	
	.formula-preview h4 {
		font-size: 1.1rem;
		font-weight: 700;
		color: #1a1a1a;
		margin: 0 0 1rem 0;
	}
	
	.preview-content {
		background: white;
		border-radius: 8px;
		padding: 1.5rem;
	}
	
	.preview-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}
	
	.preview-emoji {
		font-size: 2rem;
	}
	
	.preview-description {
		color: #666;
		margin: 0 0 1rem 0;
		line-height: 1.6;
	}
	
	.preview-config {
		background: #f9f9f9;
		border-radius: 6px;
		padding: 1rem;
		margin: 1rem 0;
	}
	
	.preview-config strong {
		display: block;
		margin-bottom: 0.5rem;
		color: #333;
	}
	
	.preview-config pre {
		margin: 0;
		font-size: 0.85rem;
		color: #666;
	}
	
	.preview-settings {
		display: flex;
		flex-direction: column;
		gap: 0.75rem;
		margin-top: 1rem;
	}
	
	.preview-setting {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem;
		background: #f9f9f9;
		border-radius: 6px;
	}
	
	.setting-icon {
		font-size: 1.25rem;
	}
	
	.preview-setting strong {
		color: #d4a574;
	}
	
	/* Edit Modal */
	.edit-notice {
		background: #e3f2fd;
		border-left: 4px solid #1976d2;
		padding: 1rem;
		border-radius: 4px;
		margin-bottom: 1.5rem;
		color: #0d47a1;
	}
	
	.readonly-value {
		font-weight: 600;
		color: #666;
		margin: 0.5rem 0 0 0;
	}
	
	.edit-hint {
		text-align: center;
		color: #999;
		margin-top: 2rem;
		font-style: italic;
	}
	
	/* Responsive */
	@media (max-width: 768px) {
		.formulas-page {
			padding: 1rem;
		}
		
		.page-header {
			flex-direction: column;
			align-items: stretch;
		}
		
		.formulas-grid {
			grid-template-columns: 1fr;
		}
		
		.form-row {
			grid-template-columns: 1fr;
		}
	}
</style>
