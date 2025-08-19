<script lang="ts">
	import { onMount } from 'svelte';
	import { masterVideoService } from '$lib/master-video';
	import { createEventDispatcher } from 'svelte';

	let activeTab = 'tags';
	let tags = [];
	let categories = [];
	let loading = false;
	let error = '';

	// Tag management
	let newTag = '';
	let selectedTag = null;
	let editingTag = false;

	// Category management
	let newCategory = '';
	let newCategoryColor = '#3B82F6';
	let selectedCategory = null;
	let editingCategory = false;

	onMount(() => {
		loadData();
	});

	async function loadData() {
		loading = true;
		try {
			// Load tags and categories
			const [tagsResponse, categoriesResponse] = await Promise.all([
				masterVideoService.getTagAnalytics(),
				masterVideoService.getTagCategories()
			]);

			if (tagsResponse.success) {
				tags = tagsResponse.result?.tag_frequency || [];
			}

			if (categoriesResponse.success) {
				categories = categoriesResponse.result || [];
			}
		} catch (err) {
			error = 'Failed to load data';
			console.error('Error loading data:', err);
		} finally {
			loading = false;
		}
	}

	// Tag management functions
	async function addTag() {
		if (!newTag.trim()) return;
		
		try {
			// Add tag logic here
			await masterVideoService.addTag(newTag.trim());
			newTag = '';
			await loadData();
		} catch (err) {
			error = 'Failed to add tag';
		}
	}

	async function deleteTag(tag) {
		if (confirm(`Are you sure you want to delete "${tag.word}"?`)) {
			try {
				await masterVideoService.deleteTag(tag.id);
				await loadData();
			} catch (err) {
				error = 'Failed to delete tag';
			}
		}
	}

	// Category management functions
	async function addCategory() {
		if (!newCategory.trim()) return;
		
		try {
			await masterVideoService.addTagCategory({
				name: newCategory.trim(),
				color: newCategoryColor
			});
			newCategory = '';
			newCategoryColor = '#3B82F6';
			await loadData();
		} catch (err) {
			error = 'Failed to add category';
		}
	}

	async function deleteCategory(category) {
		if (confirm(`Are you sure you want to delete "${category.name}"?`)) {
			try {
				await masterVideoService.deleteTagCategory(category.id);
				await loadData();
			} catch (err) {
				error = 'Failed to delete category';
			}
		}
	}

	async function assignTagToCategory(tagId, categoryId) {
		try {
			await masterVideoService.assignTagToCategory(tagId, categoryId);
			await loadData();
		} catch (err) {
			error = 'Failed to assign tag to category';
		}
	}
</script>

<svelte:head>
	<title>Tags & Categories - Admin Dashboard</title>
</svelte:head>

<div class="tags-categories-page">
	<div class="page-header">
		<h1>🏷️ Tags & Categories Management</h1>
		<p>Manage video tags and organize them into categories</p>
	</div>

	{#if error}
		<div class="error-banner">
			{error}
			<button on:click={() => error = ''}>×</button>
		</div>
	{/if}

	<!-- Tab Navigation -->
	<div class="tab-navigation">
		<button 
			class="tab-button" 
			class:active={activeTab === 'tags'}
			on:click={() => activeTab = 'tags'}
		>
			🏷️ Tags ({tags.length})
		</button>
		<button 
			class="tab-button" 
			class:active={activeTab === 'categories'}
			on:click={() => activeTab = 'categories'}
		>
			📁 Categories ({categories.length})
		</button>
	</div>

	<!-- Tags Tab -->
	{#if activeTab === 'tags'}
		<div class="tab-content">
			<div class="section-header">
				<h2>Video Tags</h2>
				<div class="add-tag-form">
					<input 
						type="text" 
						bind:value={newTag} 
						placeholder="Enter new tag..."
						on:keydown={(e) => e.key === 'Enter' && addTag()}
					/>
					<button on:click={addTag} class="add-button">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="5" x2="12" y2="19"></line>
							<line x1="5" y1="12" x2="19" y2="12"></line>
						</svg>
						Add Tag
					</button>
				</div>
			</div>

			{#if loading}
				<div class="loading">Loading tags...</div>
			{:else if tags.length === 0}
				<div class="empty-state">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path>
						<line x1="7" y1="7" x2="7.01" y2="7"></line>
					</svg>
					<p>No tags found</p>
					<p>Tags will appear here after running the smart tagging system</p>
				</div>
			{:else}
				<div class="tags-grid">
					{#each tags as tag}
						<div class="tag-card">
							<div class="tag-info">
								<span class="tag-word">{tag.word}</span>
								<span class="tag-frequency">Used {tag.frequency} times</span>
							</div>
							<div class="tag-actions">
								{#if categories.length > 0}
									<select 
										value={tag.category_id || ''} 
										on:change={(e) => assignTagToCategory(tag.id, e.target.value)}
										class="category-select"
									>
										<option value="">No Category</option>
										{#each categories as category}
											<option value={category.id}>{category.name}</option>
										{/each}
									</select>
								{/if}
								<button on:click={() => deleteTag(tag)} class="delete-button">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
									</svg>
								</button>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	<!-- Categories Tab -->
	{#if activeTab === 'categories'}
		<div class="tab-content">
			<div class="section-header">
				<h2>Tag Categories</h2>
				<div class="add-category-form">
					<input 
						type="text" 
						bind:value={newCategory} 
						placeholder="Category name..."
						on:keydown={(e) => e.key === 'Enter' && addCategory()}
					/>
					<input 
						type="color" 
						bind:value={newCategoryColor}
						class="color-picker"
					/>
					<button on:click={addCategory} class="add-button">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="5" x2="12" y2="19"></line>
							<line x1="5" y1="12" x2="19" y2="12"></line>
						</svg>
						Add Category
					</button>
				</div>
			</div>

			{#if loading}
				<div class="loading">Loading categories...</div>
			{:else if categories.length === 0}
				<div class="empty-state">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 3h18v18H3zM8 12h8M12 8v8"></path>
					</svg>
					<p>No categories found</p>
					<p>Create categories to organize your tags</p>
				</div>
			{:else}
				<div class="categories-grid">
					{#each categories as category}
						<div class="category-card" style="--category-color: {category.color}">
							<div class="category-header">
								<div class="category-color" style="background-color: {category.color}"></div>
								<span class="category-name">{category.name}</span>
								<span class="category-count">
									{tags.filter(t => t.category_id === category.id).length} tags
								</span>
							</div>
							<div class="category-actions">
								<button on:click={() => deleteCategory(category)} class="delete-button">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
									</svg>
								</button>
							</div>
						</div>
					{/each}
				</div>
			{/if}
		</div>
	{/if}
</div>

<style>
	.tags-categories-page {
		padding: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
		text-align: center;
	}

	.page-header h1 {
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
		background: var(--primary-gradient);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
	}

	.page-header p {
		color: var(--text-secondary);
		font-size: 1.1rem;
	}

	.error-banner {
		background: var(--error-bg);
		color: var(--error-text);
		padding: 1rem;
		border-radius: 8px;
		margin-bottom: 2rem;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.error-banner button {
		background: none;
		border: none;
		color: inherit;
		font-size: 1.5rem;
		cursor: pointer;
	}

	.tab-navigation {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		border-bottom: 2px solid var(--border-color);
		padding-bottom: 1rem;
	}

	.tab-button {
		background: none;
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
		transition: all 0.3s ease;
		color: var(--text-secondary);
	}

	.tab-button:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.tab-button.active {
		background: var(--primary-color);
		color: white;
	}

	.tab-content {
		background: var(--bg-glass);
		border-radius: 12px;
		padding: 2rem;
		backdrop-filter: blur(20px);
		border: 1px solid var(--border-color);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.section-header h2 {
		margin: 0;
		font-size: 1.5rem;
		color: var(--text-primary);
	}

	.add-tag-form,
	.add-category-form {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.add-tag-form input,
	.add-category-form input[type="text"] {
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 6px;
		background: var(--bg-primary);
		color: var(--text-primary);
		min-width: 200px;
	}

	.color-picker {
		width: 50px;
		height: 40px;
		border: none;
		border-radius: 6px;
		cursor: pointer;
	}

	.add-button {
		background: var(--primary-color);
		color: white;
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.3s ease;
	}

	.add-button:hover {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}

	.loading {
		text-align: center;
		padding: 2rem;
		color: var(--text-secondary);
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-secondary);
	}

	.empty-state svg {
		width: 64px;
		height: 64px;
		margin-bottom: 1rem;
		opacity: 0.5;
	}

	.tags-grid,
	.categories-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
	}

	.tag-card,
	.category-card {
		background: var(--bg-primary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.3s ease;
	}

	.tag-card:hover,
	.category-card:hover {
		transform: translateY(-2px);
		box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
	}

	.tag-info {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
	}

	.tag-word {
		font-weight: 600;
		color: var(--text-primary);
	}

	.tag-frequency {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.tag-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.category-select {
		padding: 0.25rem 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 4px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 0.9rem;
	}

	.delete-button {
		background: var(--error-bg);
		color: var(--error-text);
		border: none;
		padding: 0.25rem;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.delete-button:hover {
		background: var(--error-text);
		color: white;
	}

	.category-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.category-color {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		border: 2px solid var(--border-color);
	}

	.category-name {
		font-weight: 600;
		color: var(--text-primary);
		flex: 1;
	}

	.category-count {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.category-actions {
		display: flex;
		justify-content: flex-end;
	}

	@media (max-width: 768px) {
		.tags-categories-page {
			padding: 1rem;
		}

		.section-header {
			flex-direction: column;
			align-items: stretch;
		}

		.add-tag-form,
		.add-category-form {
			flex-direction: column;
			align-items: stretch;
		}

		.tags-grid,
		.categories-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
