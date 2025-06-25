<script lang="ts">
	import { onMount } from 'svelte';
	import { designTokenService, type StyleTheme, type DesignToken } from '$lib/services/designTokenService';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	
	// Simple toast implementation for now
	const toast = {
		success: (message: string) => console.log('SUCCESS:', message),
		error: (message: string) => console.error('ERROR:', message)
	};

	let themes: StyleTheme[] = [];
	let currentTheme: StyleTheme | null = null;
	let loading = false;
	let error = '';

	// Modal states
	let showImportModal = false;
	let showCreateModal = false;
	let showPreviewModal = false;
	let previewTheme: StyleTheme | null = null;

	// Form data
	let figmaFileId = '';
	let figmaNodeId = '';
	let themeName = '';
	let themeDescription = '';
	let importData = '';

	onMount(() => {
		loadThemes();
		
		// Listen for theme changes
		const handleThemeChange = (event: Event) => {
			const customEvent = event as CustomEvent;
			toast.success('Theme applied successfully!');
		};
		
		window.addEventListener('themeChanged', handleThemeChange);
		
		return () => {
			window.removeEventListener('themeChanged', handleThemeChange);
		};
	});

	async function loadThemes() {
		loading = true;
		try {
			themes = await designTokenService.getThemes();
			currentTheme = await designTokenService.getCurrentTheme();
		} catch (err) {
			error = 'Failed to load themes';
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function activateTheme(themeId: string) {
		try {
			await designTokenService.activateTheme(themeId);
			await loadThemes();
			toast.success('Theme activated successfully!');
		} catch (err) {
			toast.error('Failed to activate theme');
			console.error(err);
		}
	}

	async function createThemeFromFigma() {
		if (!figmaFileId.trim()) {
			toast.error('Please enter a Figma file ID');
			return;
		}

		loading = true;
		try {
			const theme = await designTokenService.createThemeFromFigma(
				figmaFileId.trim(),
				figmaNodeId.trim() || undefined
			);
			
			if (themeName.trim()) {
				// Update theme name via API if needed
				// For now, the backend handles this during creation
			}

			await loadThemes();
			showCreateModal = false;
			resetCreateForm();
			toast.success('Theme created from Figma successfully!');
		} catch (err) {
			toast.error('Failed to create theme from Figma');
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function updateThemeFromFigma(themeId: string) {
		try {
			loading = true;
			await designTokenService.updateThemeFromFigma(themeId);
			await loadThemes();
			toast.success('Theme updated from Figma successfully!');
		} catch (err) {
			toast.error('Failed to update theme from Figma');
			console.error(err);
		} finally {
			loading = false;
		}
	}

	async function deleteTheme(themeId: string) {
		if (confirm('Are you sure you want to delete this theme?')) {
			try {
				await designTokenService.deleteTheme(themeId);
				await loadThemes();
				toast.success('Theme deleted successfully!');
			} catch (err) {
				toast.error('Failed to delete theme');
				console.error(err);
			}
		}
	}

	async function exportTheme(themeId: string) {
		try {
			const exportData = await designTokenService.exportTheme(themeId);
			const blob = new Blob([exportData], { type: 'application/json' });
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `theme-${themeId}.json`;
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
			toast.success('Theme exported successfully!');
		} catch (err) {
			toast.error('Failed to export theme');
			console.error(err);
		}
	}

	async function importTheme() {
		if (!importData.trim()) {
			toast.error('Please enter theme data');
			return;
		}

		try {
			await designTokenService.importTheme(importData.trim());
			await loadThemes();
			showImportModal = false;
			importData = '';
			toast.success('Theme imported successfully!');
		} catch (err) {
			toast.error('Failed to import theme - invalid data');
			console.error(err);
		}
	}

	function previewThemeTokens(theme: StyleTheme) {
		previewTheme = theme;
		showPreviewModal = true;
	}

	function resetCreateForm() {
		figmaFileId = '';
		figmaNodeId = '';
		themeName = '';
		themeDescription = '';
	}

	function getTokensByCategory(tokens: DesignToken[]) {
		const categories: Record<string, DesignToken[]> = {};
		tokens.forEach(token => {
			if (!categories[token.category]) {
				categories[token.category] = [];
			}
			categories[token.category].push(token);
		});
		return categories;
	}

	function formatTokenValue(token: DesignToken): string {
		if (typeof token.value === 'object') {
			return JSON.stringify(token.value, null, 2);
		}
		return String(token.value);
	}
</script>

<svelte:head>
	<title>Design System Management - BOME Admin</title>
</svelte:head>

<div class="design-system-page">
	<div class="page-header">
		<div class="header-content">
			<div class="header-text">
				<h1>Design System Management</h1>
				<p>Manage Figma design themes and apply dynamic styling across the platform</p>
			</div>
			<div class="header-actions">
				<button class="btn btn-secondary" on:click={() => showImportModal = true}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path>
						<polyline points="7,10 12,15 17,10"></polyline>
						<line x1="12" y1="15" x2="12" y2="3"></line>
					</svg>
					Import Theme
				</button>
				<button class="btn btn-primary" on:click={() => showCreateModal = true}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<circle cx="12" cy="12" r="10"></circle>
						<line x1="12" y1="8" x2="12" y2="16"></line>
						<line x1="8" y1="12" x2="16" y2="12"></line>
					</svg>
					Create from Figma
				</button>
			</div>
		</div>
	</div>

	{#if loading && themes.length === 0}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading design themes...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<p class="error-message">{error}</p>
			<button class="btn btn-primary" on:click={loadThemes}>
				Try Again
			</button>
		</div>
	{:else}
		<!-- Current Theme Status -->
		{#if currentTheme}
			<div class="current-theme-card glass">
				<div class="current-theme-header">
					<div class="current-theme-info">
						<h2>Active Theme</h2>
						<div class="theme-details">
							<h3>{currentTheme.name}</h3>
							<p>{currentTheme.description}</p>
							<div class="theme-meta">
								<span class="meta-item">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="3"></circle>
										<path d="M12 1v6m0 6v6m11-7h-6m-6 0H1"></path>
									</svg>
									{currentTheme.tokens.length} tokens
								</span>
								<span class="meta-item">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect>
										<line x1="16" y1="2" x2="16" y2="6"></line>
										<line x1="8" y1="2" x2="8" y2="6"></line>
										<line x1="3" y1="10" x2="21" y2="10"></line>
									</svg>
									Updated {new Date(currentTheme.updatedAt).toLocaleDateString()}
								</span>
							</div>
						</div>
					</div>
					<div class="current-theme-actions">
						<button class="btn btn-ghost" on:click={() => currentTheme && previewThemeTokens(currentTheme)}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
								<circle cx="12" cy="12" r="3"></circle>
							</svg>
							View Tokens
						</button>
						{#if currentTheme?.figmaFileId}
							<button class="btn btn-outline" on:click={() => currentTheme && updateThemeFromFigma(currentTheme.id.toString())} disabled={loading}>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
									<path d="M21 3v5h-5"></path>
									<path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16"></path>
									<path d="M3 21v-5h5"></path>
								</svg>
								Update from Figma
							</button>
						{/if}
					</div>
				</div>
			</div>
		{:else}
			<div class="no-theme-card glass">
				<div class="no-theme-content">
					<div class="no-theme-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M12 2L2 7l10 5 10-5-10-5z"></path>
							<path d="M2 17l10 5 10-5"></path>
							<path d="M2 12l10 5 10-5"></path>
						</svg>
					</div>
					<h3>No Active Theme</h3>
					<p>Create or activate a theme to customize the platform's design</p>
				</div>
			</div>
		{/if}

		<!-- Themes Grid -->
		<div class="themes-section">
			<div class="section-header">
				<h2>Available Themes</h2>
				<span class="theme-count">{themes.length} theme{themes.length !== 1 ? 's' : ''}</span>
			</div>

			{#if themes.length === 0}
				<div class="empty-state">
					<div class="empty-icon">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
							<circle cx="9" cy="9" r="2"></circle>
							<path d="M21 15l-3.086-3.086a2 2 0 0 0-2.828 0L6 21"></path>
						</svg>
					</div>
					<h3>No Themes Available</h3>
					<p>Create your first theme from Figma or import an existing theme configuration</p>
					<div class="empty-actions">
						<button class="btn btn-primary" on:click={() => showCreateModal = true}>
							Create from Figma
						</button>
						<button class="btn btn-outline" on:click={() => showImportModal = true}>
							Import Theme
						</button>
					</div>
				</div>
			{:else}
				<div class="themes-grid">
					{#each themes as theme (theme.id)}
						<div class="theme-card glass" class:active={theme.isActive}>
							<div class="theme-header">
								<div class="theme-info">
									<h3>{theme.name}</h3>
									<p>{theme.description}</p>
								</div>
								{#if theme.isActive}
									<div class="active-badge">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
											<polyline points="22,4 12,14.01 9,11.01"></polyline>
										</svg>
										Active
									</div>
								{/if}
							</div>

							<div class="theme-stats">
								<div class="stat">
									<span class="stat-value">{theme.tokens.length}</span>
									<span class="stat-label">Tokens</span>
								</div>
								<div class="stat">
									<span class="stat-value">{Object.keys(getTokensByCategory(theme.tokens)).length}</span>
									<span class="stat-label">Categories</span>
								</div>
								{#if theme.figmaFileId}
									<div class="stat figma-linked">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<path d="M5 7h1a2 2 0 0 0 0-4H5a2 2 0 0 0 0 4z"></path>
											<path d="M12 7h1a2 2 0 0 0 0-4h-1v4z"></path>
											<path d="M5 17h1a2 2 0 0 0 0-4H5a2 2 0 0 0 0 4z"></path>
											<path d="M12 17h1a2 2 0 0 0 0-4h-1v4z"></path>
											<circle cx="12" cy="7" r="2"></circle>
										</svg>
										Figma Linked
									</div>
								{/if}
							</div>

							<div class="theme-actions">
								{#if !theme.isActive}
									<button class="btn btn-primary btn-small" on:click={() => activateTheme(theme.id)}>
										Activate
									</button>
								{/if}
								<button class="btn btn-ghost btn-small" on:click={() => previewThemeTokens(theme)}>
									Preview
								</button>
								<div class="theme-menu">
									<button class="btn btn-ghost btn-small menu-trigger">
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<circle cx="12" cy="12" r="1"></circle>
											<circle cx="12" cy="5" r="1"></circle>
											<circle cx="12" cy="19" r="1"></circle>
										</svg>
									</button>
									<div class="menu-dropdown">
										{#if theme.figmaFileId}
											<button on:click={() => updateThemeFromFigma(theme.id)} disabled={loading}>
												<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
													<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8"></path>
													<path d="M21 3v5h-5"></path>
												</svg>
												Update from Figma
											</button>
										{/if}
										<button on:click={() => exportTheme(theme.id)}>
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-15"></path>
												<polyline points="17,10 12,5 7,10"></polyline>
											</svg>
											Export
										</button>
										<button on:click={() => deleteTheme(theme.id)} class="danger">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<polyline points="3,6 5,6 21,6"></polyline>
												<path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path>
											</svg>
											Delete
										</button>
									</div>
								</div>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Create from Figma Modal -->
{#if showCreateModal}
	<div class="modal-overlay" on:click={() => showCreateModal = false}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Create Theme from Figma</h2>
				<button class="modal-close" on:click={() => showCreateModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			<div class="modal-body">
				<div class="form-group">
					<label for="figmaFileId">Figma File ID *</label>
					<input
						id="figmaFileId"
						type="text"
						bind:value={figmaFileId}
						placeholder="Enter Figma file ID"
						class="form-input"
					/>
					<small>You can find this in the Figma file URL</small>
				</div>
				<div class="form-group">
					<label for="figmaNodeId">Figma Node ID (Optional)</label>
					<input
						id="figmaNodeId"
						type="text"
						bind:value={figmaNodeId}
						placeholder="Enter specific node ID"
						class="form-input"
					/>
					<small>Leave empty to import the entire file</small>
				</div>
				<div class="form-group">
					<label for="themeName">Theme Name (Optional)</label>
					<input
						id="themeName"
						type="text"
						bind:value={themeName}
						placeholder="Custom theme name"
						class="form-input"
					/>
				</div>
				<div class="form-group">
					<label for="themeDescription">Description (Optional)</label>
					<textarea
						id="themeDescription"
						bind:value={themeDescription}
						placeholder="Describe this theme"
						class="form-textarea"
						rows="3"
					></textarea>
				</div>
			</div>
			<div class="modal-footer">
				<button class="btn btn-ghost" on:click={() => showCreateModal = false}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={createThemeFromFigma} disabled={loading || !figmaFileId.trim()}>
					{#if loading}
						<LoadingSpinner size="small" color="white" />
					{/if}
					Create Theme
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Import Theme Modal -->
{#if showImportModal}
	<div class="modal-overlay" on:click={() => showImportModal = false}>
		<div class="modal" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Import Theme</h2>
				<button class="modal-close" on:click={() => showImportModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			<div class="modal-body">
				<div class="form-group">
					<label for="importData">Theme Configuration JSON</label>
					<textarea
						id="importData"
						bind:value={importData}
						placeholder="Paste theme JSON configuration here..."
						class="form-textarea"
						rows="10"
					></textarea>
				</div>
			</div>
			<div class="modal-footer">
				<button class="btn btn-ghost" on:click={() => showImportModal = false}>
					Cancel
				</button>
				<button class="btn btn-primary" on:click={importTheme} disabled={!importData.trim()}>
					Import Theme
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Preview Theme Modal -->
{#if showPreviewModal && previewTheme}
	<div class="modal-overlay" on:click={() => showPreviewModal = false}>
		<div class="modal large" on:click|stopPropagation>
			<div class="modal-header">
				<h2>Theme Tokens: {previewTheme.name}</h2>
				<button class="modal-close" on:click={() => showPreviewModal = false}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<line x1="18" y1="6" x2="6" y2="18"></line>
						<line x1="6" y1="6" x2="18" y2="18"></line>
					</svg>
				</button>
			</div>
			<div class="modal-body">
				<div class="token-categories">
					{#each Object.entries(getTokensByCategory(previewTheme.tokens)) as [category, tokens]}
						<div class="token-category">
							<h3>{category}</h3>
							<div class="tokens-list">
								{#each tokens as token}
									<div class="token-item">
										<div class="token-info">
											<span class="token-name">--{token.name}</span>
											<span class="token-type">{token.type}</span>
										</div>
										<div class="token-value">
											{#if token.type === 'color'}
												<div class="color-preview" style="background-color: {token.value}"></div>
											{/if}
											<code>{formatTokenValue(token)}</code>
										</div>
									</div>
								{/each}
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>
{/if}

<style>
	.design-system-page {
		display: flex;
		flex-direction: column;
		gap: var(--space-2xl);
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-text h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.header-text p {
		color: var(--text-secondary);
		margin: var(--space-sm) 0 0 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-4xl);
		gap: var(--space-lg);
	}

	.error-message {
		color: var(--error);
		font-size: var(--text-lg);
	}

	/* Current Theme Card */
	.current-theme-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.current-theme-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-lg);
	}

	.current-theme-info h2 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--success);
		margin: 0 0 var(--space-md) 0;
	}

	.theme-details h3 {
		font-size: var(--text-2xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.theme-details p {
		color: var(--text-secondary);
		margin: 0 0 var(--space-md) 0;
	}

	.theme-meta {
		display: flex;
		gap: var(--space-lg);
		flex-wrap: wrap;
	}

	.meta-item {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		font-size: var(--text-sm);
		color: var(--text-secondary);
	}

	.meta-item svg {
		width: 16px;
		height: 16px;
	}

	.current-theme-actions {
		display: flex;
		gap: var(--space-md);
		flex-shrink: 0;
	}

	/* No Theme Card */
	.no-theme-card {
		padding: var(--space-3xl);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		text-align: center;
	}

	.no-theme-content {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-lg);
	}

	.no-theme-icon {
		width: 80px;
		height: 80px;
		color: var(--text-secondary);
	}

	.no-theme-icon svg {
		width: 100%;
		height: 100%;
	}

	/* Themes Section */
	.themes-section {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
	}

	.section-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.theme-count {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		background: var(--bg-glass);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
	}

	/* Empty State */
	.empty-state {
		text-align: center;
		padding: var(--space-3xl) 0;
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		margin: 0 auto var(--space-lg);
		color: var(--text-secondary);
	}

	.empty-icon svg {
		width: 100%;
		height: 100%;
	}

	.empty-state h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.empty-state p {
		color: var(--text-secondary);
		margin-bottom: var(--space-lg);
	}

	.empty-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: center;
	}

	/* Themes Grid */
	.themes-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: var(--space-lg);
	}

	.theme-card {
		padding: var(--space-xl);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		transition: all var(--transition-normal);
		position: relative;
	}

	.theme-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.theme-card.active {
		border-color: var(--success);
		box-shadow: 0 0 20px rgba(67, 233, 123, 0.2);
	}

	.theme-header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		margin-bottom: var(--space-lg);
	}

	.theme-info h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
	}

	.theme-info p {
		color: var(--text-secondary);
		margin: 0;
		font-size: var(--text-sm);
	}

	.active-badge {
		display: flex;
		align-items: center;
		gap: var(--space-xs);
		background: rgba(67, 233, 123, 0.1);
		color: var(--success);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		flex-shrink: 0;
	}

	.active-badge svg {
		width: 14px;
		height: 14px;
	}

	.theme-stats {
		display: flex;
		gap: var(--space-lg);
		margin-bottom: var(--space-lg);
		flex-wrap: wrap;
	}

	.stat {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.stat-value {
		font-size: var(--text-lg);
		font-weight: 700;
		color: var(--text-primary);
	}

	.stat-label {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.figma-linked {
		flex-direction: row;
		align-items: center;
		gap: var(--space-xs);
		color: var(--primary);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.figma-linked svg {
		width: 16px;
		height: 16px;
	}

	.theme-actions {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		position: relative;
	}

	.theme-menu {
		position: relative;
		margin-left: auto;
	}

	.menu-trigger {
		padding: var(--space-xs);
		width: 32px;
		height: 32px;
	}

	.menu-dropdown {
		position: absolute;
		top: 100%;
		right: 0;
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: var(--shadow-lg);
		min-width: 180px;
		z-index: var(--z-dropdown);
		opacity: 0;
		visibility: hidden;
		transform: translateY(-10px);
		transition: all var(--transition-fast);
	}

	.theme-menu:hover .menu-dropdown {
		opacity: 1;
		visibility: visible;
		transform: translateY(0);
	}

	.menu-dropdown button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		width: 100%;
		padding: var(--space-sm) var(--space-md);
		border: none;
		background: none;
		color: var(--text-primary);
		font-size: var(--text-sm);
		cursor: pointer;
		transition: background var(--transition-fast);
	}

	.menu-dropdown button:hover {
		background: rgba(255, 255, 255, 0.1);
	}

	.menu-dropdown button.danger {
		color: var(--error);
	}

	.menu-dropdown button svg {
		width: 16px;
		height: 16px;
	}

	/* Modal Styles */
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
		z-index: var(--z-modal);
		backdrop-filter: blur(4px);
		padding: var(--space-lg);
	}

	.modal {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
		max-width: 600px;
		width: 100%;
		max-height: 90vh;
		overflow-y: auto;
		box-shadow: var(--shadow-2xl);
	}

	.modal.large {
		max-width: 900px;
	}

	.modal-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-xl);
		border-bottom: 1px solid rgba(255, 255, 255, 0.1);
	}

	.modal-header h2 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.modal-close {
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: var(--space-sm);
		border-radius: var(--radius-md);
		transition: all var(--transition-fast);
	}

	.modal-close:hover {
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
	}

	.modal-close svg {
		width: 20px;
		height: 20px;
	}

	.modal-body {
		padding: var(--space-xl);
	}

	.modal-footer {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-md);
		padding: var(--space-xl);
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Form Styles */
	.form-group {
		margin-bottom: var(--space-lg);
	}

	.form-group label {
		display: block;
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: var(--space-sm);
	}

	.form-input,
	.form-textarea {
		width: 100%;
		padding: var(--space-md);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		background: var(--bg-glass);
		color: var(--text-primary);
		font-size: var(--text-sm);
		transition: all var(--transition-fast);
	}

	.form-input:focus,
	.form-textarea:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
	}

	.form-group small {
		display: block;
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin-top: var(--space-xs);
	}

	/* Token Preview Styles */
	.token-categories {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.token-category h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-md) 0;
		text-transform: capitalize;
	}

	.tokens-list {
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.token-item {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: var(--space-md);
		background: rgba(255, 255, 255, 0.05);
		border-radius: var(--radius-lg);
		gap: var(--space-md);
	}

	.token-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.token-name {
		font-family: var(--font-mono);
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 600;
	}

	.token-type {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.token-value {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
	}

	.color-preview {
		width: 24px;
		height: 24px;
		border-radius: var(--radius-sm);
		border: 1px solid rgba(255, 255, 255, 0.1);
		flex-shrink: 0;
	}

	.token-value code {
		font-family: var(--font-mono);
		font-size: var(--text-xs);
		color: var(--text-secondary);
		background: rgba(255, 255, 255, 0.05);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-sm);
		max-width: 200px;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	/* Button Styles */
	.btn {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-xl);
		font-family: var(--font-sans);
		font-size: var(--text-base);
		font-weight: 600;
		line-height: var(--leading-none);
		text-decoration: none;
		border: none;
		border-radius: var(--radius-lg);
		cursor: pointer;
		transition: all var(--transition-normal);
		position: relative;
		overflow: hidden;
		white-space: nowrap;
		user-select: none;
	}

	.btn:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-primary {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.btn-primary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.btn-secondary {
		background: var(--secondary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.btn-secondary:hover:not(:disabled) {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	.btn-outline {
		background: transparent;
		color: var(--text-primary);
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.btn-outline:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.1);
		transform: translateY(-1px);
	}

	.btn-ghost {
		background: transparent;
		color: var(--text-primary);
	}

	.btn-ghost:hover:not(:disabled) {
		background: rgba(255, 255, 255, 0.1);
	}

	.btn-small {
		padding: var(--space-sm) var(--space-md);
		font-size: var(--text-sm);
	}

	.btn svg {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
	}

	.btn-small svg {
		width: 16px;
		height: 16px;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.header-content {
			flex-direction: column;
			align-items: flex-start;
		}

		.current-theme-header {
			flex-direction: column;
			align-items: flex-start;
		}

		.themes-grid {
			grid-template-columns: 1fr;
		}

		.theme-header {
			flex-direction: column;
			align-items: flex-start;
			gap: var(--space-md);
		}

		.theme-actions {
			width: 100%;
			justify-content: space-between;
		}

		.modal {
			margin: var(--space-lg);
			max-width: none;
		}

		.empty-actions {
			flex-direction: column;
		}
	}
</style> 