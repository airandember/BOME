<script lang="ts">
	import { onMount } from 'svelte';
	import { masterVideoService } from '$lib/master-video';
	import { createEventDispatcher } from 'svelte';
	import { toastStore } from '$lib/stores/toast';

	let activeTab = 'tags';
	let tags: any[] = [];
	let categories: any[] = [];
	let loading = false;

	// Tag management
	let newTag = '';
	let selectedTag: any = null;
	let editingTag = false;

	// Category management
	let newCategory = '';
	let newCategoryColor = '#3B82F6';
	let selectedCategory: any = null;
	let editingCategory = false;

	// Article exclusions
	let articleExclusions: any[] = [];
	let newExclusion = '';
	let exclusionsLoading = false;

	// Sorting state
	let tagsSortField: 'status' | 'word' | 'frequency' | null = null;
	let tagsSortDirection: 'asc' | 'desc' = 'asc';
	let exclusionsSortField: 'word' | 'status' | null = null;
	let exclusionsSortDirection: 'asc' | 'desc' = 'asc';

	onMount(() => {
		loadData();
		loadArticleExclusions();
	});

	async function loadData() {
		loading = true;
		try {
			// Load only streaming-specific data
			const [streamingTagsResponse, streamingCategoriesResponse] = await Promise.all([
				masterVideoService.getSubsiteTags('streaming'),
				masterVideoService.getSubsiteCategories('streaming')
			]);

			// Parse responses
			const streamingTagsData = await streamingTagsResponse.json();
			const streamingCategoriesData = await streamingCategoriesResponse.json();

			if (streamingTagsData.success) {
				tags = streamingTagsData.result || [];
			}

			if (streamingCategoriesData.success) {
				categories = streamingCategoriesData.result || [];
			}
		} catch (err) {
			toastStore.error('Failed to load streaming tag data');
			console.error('Error loading data:', err);
		} finally {
			loading = false;
		}
	}

	// Tag management functions
	async function addTag() {
		if (!newTag.trim()) return;
		
		try {
			const response = await masterVideoService.addSubsiteTag('streaming', newTag.trim());
			if (response.ok) {
				const data = await response.json();
				if (data.success) {
					// Add to local state instead of reloading
					const newTagObj = {
						id: data.result?.id || Date.now(),
						word: newTag.trim(),
						frequency: 0,
						category_id: null,
						active_tag: true,
						created_at: new Date().toISOString(),
						updated_at: new Date().toISOString()
					};
					tags = [...tags, newTagObj];
					
					newTag = '';
					toastStore.success('Tag added successfully to streaming subsite');
				} else {
					toastStore.error(data.error || 'Failed to add tag');
				}
			} else {
				toastStore.error('Failed to add tag to streaming subsite');
			}
		} catch (err) {
			toastStore.error('Failed to add tag to streaming subsite');
		}
	}

	async function deleteTag(tag: any) {
		if (confirm(`Are you sure you want to delete "${tag.word}" from streaming?`)) {
			try {
				const response = await masterVideoService.deleteSubsiteTag('streaming', tag.id);
				if (response.ok) {
					// Remove from local state instead of reloading
					tags = tags.filter(t => t.id !== tag.id);
					toastStore.success('Tag removed from streaming subsite');
				} else {
					toastStore.error('Failed to delete tag from streaming subsite');
				}
			} catch (err) {
				toastStore.error('Failed to delete tag from streaming subsite');
			}
		}
	}

	// Category management functions
	async function addCategory() {
		if (!newCategory.trim()) return;
		
		try {
			const response = await masterVideoService.addSubsiteCategory('streaming', {
				name: newCategory.trim(),
				color: newCategoryColor,
				description: `Streaming-specific category: ${newCategory.trim()}`
			});
			if (response.ok) {
				const data = await response.json();
				if (data.success) {
					// Add to local state instead of reloading
					const newCategoryObj = {
						id: data.result?.id || Date.now(),
						name: newCategory.trim(),
						color: newCategoryColor,
						description: `Streaming-specific category: ${newCategory.trim()}`,
						created_at: new Date().toISOString(),
						updated_at: new Date().toISOString()
					};
					categories = [...categories, newCategoryObj];
					
					newCategory = '';
					newCategoryColor = '#3B82F6';
					toastStore.success('Category added successfully to streaming subsite');
				} else {
					toastStore.error(data.error || 'Failed to add category');
				}
			} else {
				toastStore.error('Failed to add category to streaming subsite');
			}
		} catch (err) {
			toastStore.error('Failed to add category to streaming subsite');
		}
	}

	async function deleteCategory(category: any) {
		if (confirm(`Are you sure you want to delete "${category.name}" from streaming?`)) {
			try {
				const response = await masterVideoService.deleteSubsiteCategory('streaming', category.id);
				if (response.ok) {
					// Remove from local state instead of reloading
					categories = categories.filter(c => c.id !== category.id);
					
					// Also remove category_id from tags that had this category
					tags = tags.map(t => 
						t.category_id === category.id 
							? { ...t, category_id: null }
							: t
					);
					
					toastStore.success('Category removed from streaming subsite');
				} else {
					toastStore.error('Failed to delete category from streaming subsite');
				}
			} catch (err) {
				toastStore.error('Failed to delete category from streaming subsite');
			}
		}
	}

	async function assignTagToCategory(tagId: any, categoryId: any) {
		try {
			const response = await masterVideoService.assignSubsiteTagToCategory('streaming', tagId, categoryId);
			if (response.ok) {
				// Update local state instead of reloading
				tags = tags.map(t => 
					t.id === tagId 
						? { ...t, category_id: categoryId === '' ? null : parseInt(categoryId) }
						: t
				);
				
				toastStore.success('Tag assigned to category successfully');
			} else {
				toastStore.error('Failed to assign tag to category');
			}
		} catch (err) {
			toastStore.error('Failed to assign tag to category');
		}
	}



	async function loadArticleExclusions() {
		try {
			exclusionsLoading = true;
			const response = await masterVideoService.getArticleExclusions('streaming');
			if (response.success && response.result) {
				console.log(response);
				articleExclusions = response.result;
			} else {
				toastStore.error(response.error || 'Failed to load article exclusions');
			}
		} catch (err) {
			toastStore.error('Failed to load article exclusions');
		} finally {
			exclusionsLoading = false;
		}
	}

	async function addArticleExclusion() {
		if (!newExclusion.trim()) return;
		
		try {
			const response = await masterVideoService.addArticleExclusion('streaming', newExclusion.trim());
			if (response.success) {
				// Add to local state instead of reloading
				const newExclusionObj = {
					id: Date.now(), // Temporary ID
					Word: newExclusion.trim(),
					Excluded: true,
					subsite_id: 1, // Assuming streaming subsite ID
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};
				articleExclusions = [...articleExclusions, newExclusionObj];
				
				newExclusion = '';
				toastStore.success(response.message || 'Article exclusion added successfully');
			} else {
				toastStore.error(response.error || 'Failed to add article exclusion');
			}
		} catch (err) {
			toastStore.error('Failed to add article exclusion');
		}
	}

	async function toggleArticleExclusion(exclusion: any) {
		try {
			const response = await masterVideoService.toggleArticleExclusion('streaming', exclusion.Word, !exclusion.Excluded);
			if (response.success) {
				// Update local state instead of reloading
				articleExclusions = articleExclusions.map(e => 
					e.Word === exclusion.Word 
						? { ...e, Excluded: !e.Excluded }
						: e
				);
				
				toastStore.success(response.message || 'Article exclusion toggled successfully');
			} else {
				toastStore.error(response.error || 'Failed to toggle article exclusion');
			}
		} catch (err) {
			toastStore.error('Failed to toggle article exclusion');
		}
	}

	async function removeArticleExclusion(exclusion: any) {
		if (confirm(`Are you sure you want to remove "${exclusion.Word}" from exclusions?`)) {
			try {
				const response = await masterVideoService.removeArticleExclusion('streaming', exclusion.Word);
				if (response.success) {
					// Remove from local state instead of reloading
					articleExclusions = articleExclusions.filter(e => e.Word !== exclusion.Word);
					
					toastStore.success(response.message || 'Article exclusion removed successfully');
				} else {
					toastStore.error(response.error || 'Failed to remove article exclusion');
				}
			} catch (err) {
				toastStore.error('Failed to remove article exclusion');
			}
		}
	}

	async function addArticleExclusionFromTag(tag: any) {
		try {
			// First add to exclusions
			const response = await masterVideoService.addArticleExclusion('streaming', tag.word);
			if (response.success) {
				// Then delete the tag from the tags table
				await masterVideoService.deleteSubsiteTag('streaming', tag.id);
				toastStore.success(`"${tag.word}" added to exclusions and removed from tags`);
				
				// Update local state instead of reloading
				tags = tags.filter(t => t.id !== tag.id);
				
				// Add to local exclusions if not already there
				const existingExclusion = articleExclusions.find(e => e.Word === tag.word);
				if (!existingExclusion) {
					articleExclusions = [...articleExclusions, {
						id: Date.now(), // Temporary ID
						Word: tag.word,
						Excluded: true,
						subsite_id: 1, // Assuming streaming subsite ID
						created_at: new Date().toISOString(),
						updated_at: new Date().toISOString()
					}];
				}
			} else {
				toastStore.error(response.error || 'Failed to add article exclusion');
			}
		} catch (err) {
			toastStore.error('Failed to add article exclusion');
			console.error('Error adding article exclusion:', err);
		}
	}

	async function toggleTagActiveStatus(tag: any) {
		try {
			const response = await masterVideoService.toggleTagActiveStatus('streaming', tag.id, !tag.active_tag);
			if (response.success) {
				toastStore.success(`Tag "${tag.word}" ${tag.active_tag ? 'deactivated' : 'activated'} successfully`);
				
				// Update local state instead of reloading
				tags = tags.map(t => 
					t.id === tag.id 
						? { ...t, active_tag: !t.active_tag }
						: t
				);
			} else {
				toastStore.error(response.error || 'Failed to toggle tag status');
			}
		} catch (err) {
			toastStore.error('Failed to toggle tag status');
			console.error('Error toggling tag status:', err);
		}
	}


	// Fix the event target issue in the template
	async function handleCategoryChange(event: Event, tagId: any) {
		const target = event.target as HTMLSelectElement;
		if (target) {
			await assignTagToCategory(tagId, target.value);
		}
	}

	// Sorting functions
function sortTags(field: 'status' | 'word' | 'frequency') {
	if (tagsSortField === field) {
		tagsSortDirection = tagsSortDirection === 'asc' ? 'desc' : 'asc';
	} else {
		tagsSortField = field;
		tagsSortDirection = 'asc';
	}
	
	// Apply sorting
	tags = [...tags].sort((a, b) => {
		let aVal: any, bVal: any;
		
		switch (field) {
			case 'status':
				aVal = a.active_tag ? 1 : 0;
				bVal = b.active_tag ? 1 : 0;
				break;
			case 'word':
				aVal = a.word.toLowerCase();
				bVal = b.word.toLowerCase();
				break;
			case 'frequency':
				aVal = a.frequency || 0;
				bVal = b.frequency || 0;
				break;
			default:
				return 0;
		}
		
		if (tagsSortDirection === 'asc') {
			return aVal > bVal ? 1 : -1;
		} else {
			return aVal < bVal ? 1 : -1;
		}
	});
}

function sortExclusions(field: 'word' | 'status') {
	if (exclusionsSortField === field) {
		exclusionsSortDirection = exclusionsSortDirection === 'asc' ? 'desc' : 'asc';
	} else {
		exclusionsSortField = field;
		exclusionsSortDirection = 'asc';
	}
	
	// Apply sorting
	articleExclusions = [...articleExclusions].sort((a, b) => {
		let aVal: any, bVal: any;
		
		switch (field) {
			case 'word':
				// Alphanumeric sorting: numbers first, then letters
				aVal = a.Word.toLowerCase();
				bVal = b.Word.toLowerCase();
				
				// Check if both are numbers
				const aNum = parseFloat(aVal);
				const bNum = parseFloat(bVal);
				if (!isNaN(aNum) && !isNaN(bNum)) {
					return tagsSortDirection === 'asc' ? aNum - bNum : bNum - aNum;
				}
				// Check if only a is a number
				if (!isNaN(aNum)) return -1;
				// Check if only b is a number
				if (!isNaN(bNum)) return 1;
				// Both are strings, sort alphabetically
				break;
			case 'status':
				aVal = a.Excluded ? 1 : 0;
				bVal = b.Excluded ? 1 : 0;
				break;
			default:
				return 0;
		}
		
		if (exclusionsSortDirection === 'asc') {
			return aVal > bVal ? 1 : -1;
		} else {
			return aVal < bVal ? 1 : -1;
		}
	});
}


	
</script>

<svelte:head>
	<title>Streaming Tags & Categories - Admin Dashboard</title>
</svelte:head>

<div class="streaming-tags-categories-page">
	<!--<div class="page-header">
		 Global Integration not needed for now<div class="header-content">
			<h1>🎥 Streaming Tags & Categories</h1>
			<p>Manage tags and categories specifically for the streaming subsite</p>
			 -->
			<!--<div class="header-actions">
				<button 
					class="global-toggle" 
					class:active={showGlobalTags}
					on:click={() => showGlobalTags = !showGlobalTags}
				>
					{showGlobalTags ? 'Hide' : 'Show'} Global Integration
				</button>
				<a href="/admin/tags-categories" class="global-link">
					🌐 View Global Tags
				</a>
			</div>
		</div>
	</div>-->



	<!-- Tab Navigation -->
	<div class="tab-navigation">
		<button 
			class="tab-button" 
			class:active={activeTab === 'tags'}
			on:click={() => activeTab = 'tags'}
		>
			🏷️ Streaming Tags ({tags.length})
		</button>
		<button 
			class="tab-button" 
			class:active={activeTab === 'categories'}
			on:click={() => activeTab = 'categories'}
		>
			📁 Streaming Categories ({categories.length})
		</button>
		<button 
			class="tab-button" 
			class:active={activeTab === 'exclusions'}
			on:click={() => activeTab = 'exclusions'}
		>
			🚫 Exclusions ({articleExclusions.length})
		</button>
	</div>

	<!-- Tags Tab -->
	{#if activeTab === 'tags'}
		<div class="tab-content">
			<div class="section-header">
				<h2>Streaming Video Tags</h2>
				<div class="add-tag-form">
					<input 
						type="text" 
						bind:value={newTag} 
						placeholder="Add new streaming tag..."
						on:keydown={(e) => e.key === 'Enter' && addTag()}
					/>
					<button on:click={addTag} class="add-button">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="5" x2="12" y2="19"></line>
							<line x1="5" y1="12" x2="19" y2="12"></line>
						</svg>
						Add to Streaming
					</button>
				</div>
			</div>

			<div class="tags-info">
				<p><strong>Tag Actions:</strong></p>
				<ul>
					<li><span class="info-icon">🟢/🔴</span> <strong>Checkbox:</strong> Toggle tag active status (controls if tag is used for new videos)</li>
					<li><span class="info-icon">🚫</span> <strong>Exclude:</strong> Add word to exclusions and remove tag completely</li>
					<li><span class="info-icon">🗑️</span> <strong>Delete:</strong> Remove tag from streaming subsite</li>
				</ul>
			</div>

			{#if loading}
				<div class="loading">Loading streaming tags...</div>
			{:else if tags.length === 0}
				<div class="empty-state">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"></path>
						<line x1="7" y1="7" x2="7.01" y2="7"></line>
					</svg>
					<p>No streaming-specific tags found</p>
					<p>Add tags to get started</p>
				</div>
			{:else}
			<div class="exclusions-table-container">
				<table class="exclusions-table">
					<thead>
						<tr>
							<th class="exclusion-head sortable" on:click={() => sortTags('status')}>
								Status
								<span class="sort-indicator">
									{#if tagsSortField === 'status'}
										{tagsSortDirection === 'asc' ? '↑' : '↓'}
									{:else}
										↕
									{/if}
								</span>
							</th>
							<th class="exclusion-head sortable" on:click={() => sortTags('word')}>
								Tag Word
								<span class="sort-indicator">
									{#if tagsSortField === 'word'}
										{tagsSortDirection === 'asc' ? '↑' : '↓'}
									{:else}
										↕
									{/if}
								</span>
							</th>
							<th class="exclusion-head sortable" on:click={() => sortTags('frequency')}>
								Frequency
								<span class="sort-indicator">
									{#if tagsSortField === 'frequency'}
										{tagsSortDirection === 'asc' ? '↑' : '↓'}
									{:else}
										↕
									{/if}
								</span>
							</th>
							<th class="exclusion-head">Category</th>
							<th class="exclusion-head">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each tags as tag}
							<tr class="exclusion-row">
								<td class="exclusion-status">
									<label class="status-toggle">
										<input 
											type="checkbox" 
											checked={tag.active_tag} 
											on:change={() => toggleTagActiveStatus(tag)}
											class="status-checkbox"
											aria-label="Toggle tag active status"
										/>
										<span class="status-indicator" class:active={tag.active_tag}>
											{tag.active_tag ? '🟢 Active' : '🔴 Inactive'}
										</span>
									</label>
								</td>
								<td class="exclusion-word">
									<span class="word-text">{tag.word}</span>
								</td>
								<td class="exclusion-status">
									<span class="tag-frequency-display">
										{tag.frequency || 0} uses
									</span>
								</td>
								<td class="exclusion-status">
									{#if categories.length > 0}
										<select 
											value={tag.category_id || ''} 
											on:change={(e) => handleCategoryChange(e, tag.id)}
											class="category-select"
										>
											<option value="">No Category</option>
											{#each categories as category}
												<option value={category.id}>{category.name}</option>
											{/each}
										</select>
									{/if}
								</td>
								<td class="exclusion-actions">
									<button 
										on:click={() => addArticleExclusionFromTag(tag)} 
										class="remove-button"
										title="Add to exclusions and remove tag"
									>
									🚫
									</button>
									<button 
										on:click={() => deleteTag(tag)} 
										class="remove-button"
										title="Delete tag"
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="3,6 5,6 21,6"></polyline>
											<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
										</svg>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
				<!--<div class="tags-grid">
					{#each tags as tag}
						<div class="tag-card">
							<div class="tag-info">
																<div class="tag-header">
									<div class="checkbox-container">
										<input 
											type="checkbox" 
											checked={tag.active_tag} 
											on:change={() => toggleTagActiveStatus(tag)}
											class="active-checkbox"
											aria-label="Toggle tag active status"
										/>
										<span class="checkbox-label" title="Toggle tag active status - controls whether this tag is used for new videos">
											{tag.active_tag ? '🟢' : '🔴'}
										</span>
									</div>
									<span class="tag-word">{tag.word}</span>
								</div>
								<span class="tag-frequency">Used {tag.frequency || 0} times</span>
								<span class="tag-source">Streaming</span>
							</div>
							<div class="tag-actions">
								{#if categories.length > 0}
									<select 
										value={tag.category_id || ''} 
										on:change={(e) => handleCategoryChange(e, tag.id)}
										class="category-select"
									>
										<option value="">No Category</option>
										{#each categories as category}
											<option value={category.id}>{category.name}</option>
										{/each}
									</select>
								{/if}
								<button 
									on:click={() => addArticleExclusionFromTag(tag)} 
									class="exclusion-button"
									title="Add to exclusions and remove tag"
								>
									<span class="button-icon">🚫</span>
									<span class="button-text">Exclude</span>
								</button>
								<button on:click={() => deleteTag(tag)} class="delete-button">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
									</svg>
								</button>
							</div>
						</div>
					{/each}
				</div>-->
			{/if}
		</div>
	{/if}

	<!-- Categories Tab -->
	{#if activeTab === 'categories'}
		<div class="tab-content">
			<div class="section-header">
				<h2>Streaming Tag Categories</h2>
				<div class="add-category-form">
					<input 
						type="text" 
						bind:value={newCategory} 
						placeholder="Streaming category name..."
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
						Add to Streaming
					</button>
				</div>
			</div>

			{#if loading}
				<div class="loading">Loading streaming categories...</div>
			{:else if categories.length === 0}
				<div class="empty-state">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 3h18v18H3zM8 12h8M12 8v8"></path>
					</svg>
					<p>No streaming-specific categories found</p>
					<p>Create categories or import from the global system</p>
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
							<div class="category-description">
								{category.description || 'Streaming-specific category'}
							</div>
							<div class="category-actions">
								<button on:click={() => deleteCategory(category)} class="delete-button">
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

	<!-- Article Exclusions Tab -->
	{#if activeTab === 'exclusions'}
		<div class="tab-content">
			<div class="section-header">
				<h2>Article Exclusion List</h2>
				<div class="add-exclusion-form">
					<input 
						type="text" 
						bind:value={newExclusion} 
						placeholder="Enter article URL or pattern to exclude..."
						on:keydown={(e) => e.key === 'Enter' && addArticleExclusion()}
					/>
					<button on:click={addArticleExclusion} class="add-button">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="5" x2="12" y2="19"></line>
							<line x1="5" y1="12" x2="19" y2="12"></line>
						</svg>
						Add Exclusion
					</button>
				</div>
			</div>

			{#if exclusionsLoading}
				<div class="loading">Loading article exclusions...</div>
			{:else if articleExclusions.length === 0}
				<div class="empty-state">
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M9 12l2 2 4-4"></path>
						<path d="M21 12c-1 0-2-1-2-2s1-2 2-2 2 1 2 2-1 2-2 2z"></path>
						<path d="M3 12c1 0 2-1 2-2s-1-2-2-2-2 1-2 2 1 2 2 2z"></path>
					</svg>
					<p>No article exclusions found</p>
					<p>Add exclusions to prevent certain articles from being processed</p>
				</div>
			{:else}
			<div class="exclusions-table-container">
				<table class="exclusions-table">
					<thead>
						<tr>
							<th class="exclusion-head sortable" on:click={() => sortExclusions('word')}>
								Word/Pattern
								<span class="sort-indicator">
									{#if exclusionsSortField === 'word'}
										{exclusionsSortDirection === 'asc' ? '↑' : '↓'}
									{:else}
										↕
									{/if}
								</span>
							</th>
							<th class="exclusion-head sortable" on:click={() => sortExclusions('status')}>
								Status
								<span class="sort-indicator">
									{#if exclusionsSortField === 'status'}
										{exclusionsSortDirection === 'asc' ? '↑' : '↓'}
									{:else}
										↕
									{/if}
								</span>
							</th>
							<th class="exclusion-head">Actions</th>
						</tr>
					</thead>
					<tbody>
						{#each articleExclusions as exclusion}
							<tr class="exclusion-row">
								<td class="exclusion-word">
									<span class="word-text">{exclusion.Word}</span>
								</td>
								<td class="exclusion-status">
									<label class="status-toggle">
									<input 
											type="checkbox" 
											checked={exclusion.Excluded} 
											on:change={() => toggleArticleExclusion(exclusion)}
											class="status-checkbox"
											aria-label="Toggle exclusion status"
										/>
										
										<span class="status-indicator" class:active={exclusion.Excluded}>
											{exclusion.Excluded ? '🚫 Excluded' : '✅ Included'}
										</span>
									</label>
								</td>
								<td class="exclusion-actions">
									
									<button 
										on:click={() => removeArticleExclusion(exclusion)} 
										class="remove-button"
										title="Remove exclusion"
									>
										<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
											<polyline points="3,6 5,6 21,6"></polyline>
											<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
										</svg>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
				
				<!--<div class="exclusions-grid">
					{#each articleExclusions as exclusion}
						<div class="exclusion-card">
							<div class="exclusion-info">
								<div class="exclusion-header">
									<span class="exclusion-word">{exclusion.Word}</span>
									<input 
										type="checkbox" 
										checked={exclusion.Excluded} 
										on:change={() => toggleArticleExclusion(exclusion)}
										class="exclusion-checkbox"
										aria-label="Toggle exclusion status"
									/>
									
								</div>
								<span class="exclusion-type">
									{exclusion.Excluded ? 'Excluded' : 'Included'}
								</span>
							</div>
							<div class="exclusion-actions">
								<button on:click={() => removeArticleExclusion(exclusion)} class="delete-button">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<polyline points="3,6 5,6 21,6"></polyline>
										<path d="M19,6v14a2,2,0,0,1-2,2H7a2,2,0,0,1-2-2V6m3,0V4a2,2,0,0,1,2-2h4a2,2,0,0,1,2,2V6"></path>
									</svg>
								</button>
							</div>
						</div>
					{/each}
				</div>-->
			{/if}
		</div>
	{/if}
</div>


<style>
	.sort-indicator {
		cursor: pointer;
	}


	.streaming-tags-categories-page {
		padding: 2rem;
		max-width:100%;
		margin: 0 auto;
	}

	.page-header {
		margin-bottom: 2rem;
		text-align: center;
	}

	.header-content h1 {
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
		background: var(--primary-gradient);
		-webkit-background-clip: text;
		-webkit-text-fill-color: transparent;
	}

	.header-content p {
		color: var(--text-secondary);
		font-size: 1.1rem;
		margin-bottom: 1.5rem;
	}

	.header-actions {
		display: flex;
		gap: 1rem;
		justify-content: center;
		flex-wrap: wrap;
	}

	.global-toggle {
		background: var(--secondary-color);
		color:  var(--text-secondary);
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.global-toggle:hover,
	.global-toggle.active {
		background: var(--secondary-hover);
		transform: translateY(-1px);
	}

	.global-link {
		background: var(--primary-color);
		color:  var(--text-secondary);
		text-decoration: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		transition: all 0.3s ease;
	}

	.global-link:hover {
		background: var(--primary-hover);
		transform: translateY(-1px);
	}



	.tab-navigation {
		display: flex;
		gap: 1rem;
		border-bottom: 2px solid var(--border-color);
		flex-wrap: wrap;
		padding-left: 1rem;
	}

	.tab-button {
		background: none;
		border: none;
		padding: 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		font-size: 1rem;
		transition: all 0.3s ease;
		color: var(--text-secondary);
		white-space: nowrap;
	}

	.tab-button:hover {
		background: var(--bg-hover);
		color: var(--text-primary);
	}

	.tab-button.active {
		background: var(--primary-color);
		color: var(--text-secondary);
		border-radius: 9px 9px 0 0;
		background: linear-gradient(145deg, var(--bg-tertiary), var(--bg-secondary));
		box-shadow:  10px 10px 15px var(--bg-quaternary),
					-10px -10px 15px var(--bg-primary);
	}

	.tab-content {
		background: var(--bg-glass);
		border-radius: 12px;
		padding: 2rem;
		backdrop-filter: blur(80px);
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
		flex-wrap: wrap;
	}

	.add-tag-form input,
	.add-category-form input[type="text"],
	.add-exclusion-form input[type="text"] {
		font-size: 1.1rem;
		padding: 0.5rem 1rem;
		border: 1px solid var(--border-color);
		border-radius: 11px;
		min-height: 65px;
		color: var(--text-primary);
		min-width: 200px;
		border-radius: 11px;
		background: var(--bg-primary);
		box-shadow: inset 5px 5px 10px var(--bg-senary),
            inset -5px -5px 10px var(--bg-secondary);
	}

	.color-picker {
		width: 50px;
		height: 40px;
		border: none;
		border-radius: 6px;
		cursor: pointer;
	}

	.add-button {
		background: var(--bg-glass-dark);
		color: var(--text-secondary);
		border: none;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		cursor: pointer;
		display: flex;
		align-items: center;
		gap: 0.5rem;
		transition: all 0.3s ease;
		white-space: nowrap;
		min-height: 65px;
		box-shadow: 0 0 0  var(--bg-senary),
            0 0 0 var(--bg-secondary);
	}

	.add-button:hover {
		background: var(--bg-glass);
		transform: translateY(-1px);
		box-shadow: 5px 5px 10px  var(--bg-senary),
            -5px -5px 10px var(--bg-secondary);
	}

	.add-button:active {
		transform: translateY(1px);
		border-radius: 11px;
		background: var(--bg-glass-dark);
		box-shadow: inset 5px 5px 10px var(--bg-senary),
            inset -5px -5px 10px var(--bg-secondary);
	}

	.tags-info {
		background: var(--bg-hover);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1rem;
		margin-bottom: 1.5rem;
		font-size: 0.9rem;
	}

	.tags-info p {
		margin: 0 0 0.75rem 0;
		color: var(--text-primary);
	}

	.tags-info ul {
		margin: 0;
		padding-left: 1.5rem;
		color: var(--text-secondary);
	}

	.tags-info li {
		margin-bottom: 0.5rem;
	}

	.info-icon {
		margin-right: 0.5rem;
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

	.import-button {
		background: var(--secondary-color);
		color:  var(--text-secondary);
		border: none;
		padding: 0.75rem 1.5rem;
		border-radius: 8px;
		cursor: pointer;
		transition: all 0.3s ease;
		margin-top: 1rem;
	}

	.import-button:hover {
		background: var(--secondary-hover);
		transform: translateY(-1px);
	}

	.tags-grid,
	.categories-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
	}

	.exclusions-grid {
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
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.tag-header {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.checkbox-container {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.checkbox-label {
		font-size: 0.8rem;
		cursor: help;
	}

	.active-checkbox {
		width: 18px;
		height: 18px;
		cursor: pointer;
		accent-color: var(--primary-color);
	}

	.tag-word {
		font-weight: 600;
		color: var(--text-primary);
		font-size: 1.1rem;
	}

	.tag-frequency {
		font-size: 0.9rem;
		color: var(--text-secondary);
	}

	.tag-source {
		background: var(--primary-color);
		color:  var(--text-secondary);
		padding: 0.25rem 0.5rem;
		border-radius: 12px;
		font-size: 0.75rem;
		font-weight: 500;
		align-self: flex-start;
	}

	.tag-active-toggle {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.5rem;
		cursor: pointer;
	}

	.toggle-switch {
		position: relative;
		width: 40px;
		height: 20px;
		background-color: var(--border-color);
		border-radius: 10px;
		cursor: pointer;
		transition: background-color 0.3s ease;
	}

	.toggle-switch::after {
		content: "";
		position: absolute;
		width: 16px;
		height: 16px;
		background-color: var(--text-secondary);
		border-radius: 50%;
		top: 2px;
		left: 2px;
		transition: transform 0.3s ease;
	}

	.toggle-switch.checked {
		background-color: var(--primary-color);
	}

	.toggle-switch.checked::after {
		transform: translateX(20px);
	}

	.tag-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
		flex-wrap: wrap;
	}

	.category-select {
		padding: 0.25rem 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 4px;
		background: var(--bg-primary);
		color: var(--text-primary);
		font-size: 0.9rem;
		min-width: 120px;
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
		color:  var(--text-secondary);
	}

	.exclusion-button {
		background: var(--secondary-color);
		color: var(--text-secondary);
		border: none;
		padding: 0.5rem 0.75rem;
		border-radius: 4px;
		cursor: pointer;
		transition: all 0.3s ease;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.exclusion-button:hover {
		background: var(--primary-color);
		color: var(--text-primary);
	}

	.button-icon {
		font-size: 1rem;
	}

	.button-text {
		font-size: 0.8rem;
		font-weight: 500;
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

	.category-description {
		color: var(--text-secondary);
		font-size: 0.9rem;
		margin-bottom: 1rem;
		font-style: italic;
	}

	.category-actions {
		display: flex;
		justify-content: flex-end;
	}

	.exclusion-card {
	background: var(--bg-primary);
	border: 1px solid var(--border-color);
	border-radius: 8px;
	padding: 1rem;
	transition: all 0.3s ease;
}

.exclusion-card:hover {
	transform: translateY(-2px);
	box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
}

.exclusion-info {
	display: flex;
	flex-direction: column;
	gap: 0.5rem;
	margin-bottom: 1rem;
}

.exclusion-header {
	display: flex;
	align-items: center;
	gap: 0.75rem;
}

.exclusion-checkbox {
	width: 18px;
	height: 18px;
	cursor: pointer;
	accent-color: var(--primary-color);
}

.exclusion-word {
	font-weight: 600;
	color: var(--text-primary);
	font-size: 1rem;
}

.exclusion-type {
	background: var(--secondary-color);
	color: var(--text-secondary);
	padding: 0.25rem 0.5rem;
	border-radius: 12px;
	font-size: 0.75rem;
	font-weight: 500;
	align-self: flex-start;
}

.exclusion-actions {
	display: flex;
	justify-content: flex-end;
}

.add-exclusion-form {
	display: flex;
	gap: 0.5rem;
	align-items: center;
	flex-wrap: wrap;
}

.add-exclusion-form input {
	padding: 0.5rem 1rem;
	border: 1px solid var(--border-color);
	border-radius: 6px;
	background: var(--bg-primary);
	color: var(--text-primary);
	min-width: 300px;
}

	/* Enhanced Table Styles */
.exclusions-table-container {
	overflow-x: auto;
	border-radius: 8px;
	border: 1px solid var(--border-color);
}

.exclusions-table {
	width: 100%;
	border-collapse: collapse;
	background: var(--bg-tertiary);
}

.exclusions-table th {
	background: var(--bg-hover);
	color: var(--text-primary);
	font-weight: 600;
	padding: 1rem;
	text-align: left;
	border-bottom: 2px solid var(--border-color);
	font-size: 0.9rem;
	text-transform: uppercase;
	letter-spacing: 0.5px;
}

.exclusions-table td {
	padding: 1rem;
	border-bottom: 1px solid var(--border-color);
	vertical-align: middle;
}

thead {
	background-color: var(--bg-quinary);
}

.exclusion-head {
	color: var(--text-inverse) !important;
	font-weight: 600;
}

.exclusion-row {
	border-bottom: 5px solid var(--bg-primary);
	color: var(--text-tertiary);
}

.exclusion-row:hover {
	background: var(--bg-hover);
	transition: background-color 0.2s ease;
}

.exclusion-word .word-text {
	font-weight: 600;
	color: var(--text-primary);
	font-family: 'Courier New', monospace;
	background: rgba(59, 130, 246, 0.1);
	padding: 0.25rem 0.5rem;
	border-radius: 4px;
	border: 1px solid rgba(59, 130, 246, 0.2);
}

.status-toggle {
	display: flex;
	align-items: center;
	gap: 0.5rem;
	cursor: pointer;
}

.status-checkbox {
	width: 18px;
	height: 18px;
	cursor: pointer;
	accent-color: var(--primary-color);
}

.status-indicator {
	padding: 0.25rem 0.75rem;
	border-radius: 12px;
	font-size: 1.25rem;
	font-weight: 500;
	transition: all 0.2s ease;
}

.status-indicator.active {
	background: var(--error-bg);
	color: var(--error-text);
}

.status-indicator:not(.active) {
	background: var(--success-bg);
	color: var(--success-text);
}

.remove-button {
	background: var(--error-bg);
	color: var(--text-secondary);
	border: none;
	padding: 0.5rem;
	border-radius: 6px;
	cursor: pointer;
	transition: all 0.2s ease;
	display: flex;
	align-items: center;
	justify-content: center;
}

.remove-button:hover {
	background: var(--error-text);
	color: red;
	transform: scale(1.05);
}

.remove-button svg {
	width: 16px;
	height: 16px;
}
/* Tag-specific table styles */
.tag-frequency-display {
	font-size: 0.9rem;
	color: var(--text-secondary);
	background: rgba(59, 130, 246, 0.1);
	padding: 0.25rem 0.5rem;
	border-radius: 4px;
	border: 1px solid rgba(59, 130, 246, 0.2);
	font-weight: 500;
}

.category-select {
	padding: 0.25rem 0.5rem;
	border: 1px solid var(--border-color);
	border-radius: 4px;
	background: var(--bg-primary);
	color: var(--text-primary);
	font-size: 0.9rem;
	min-width: 120px;
}

	@media (max-width: 768px) {
		.streaming-tags-categories-page {
			padding: 1rem;
		}

		.header-actions {
			flex-direction: column;
			align-items: center;
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
		.categories-grid,
		.global-items-grid {
			grid-template-columns: 1fr;
		}

		.tab-navigation {
			justify-content: center;
		}

		.exclusions-table th,
	.exclusions-table td {
		padding: 0.75rem 0.5rem;
		font-size: 0.85rem;
	}
	
	.exclusion-word .word-text {
		font-size: 0.8rem;
		padding: 0.2rem 0.4rem;
	}
	}
</style>
