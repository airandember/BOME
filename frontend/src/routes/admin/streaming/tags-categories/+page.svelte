<script lang="ts">
	import { onMount } from 'svelte';
	import { masterVideoService } from '$lib/master-video';
	import { createEventDispatcher } from 'svelte';
	import { toastStore } from '$lib/stores/toast';
	import TagsTab from '$lib/components/admin/streaming/TagsTab.svelte';
	import CategoriesTab from '$lib/components/admin/streaming/CategoriesTab.svelte';
	import ExclusionsTab from '$lib/components/admin/streaming/ExclusionsTab.svelte';

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



	onMount(() => {
		loadData();
		loadArticleExclusions();
	});

	async function loadData() {
		loading = true;
		try {
			// Use the new API endpoints
			const [tagsResponse, categoriesResponse] = await Promise.all([
				fetch('/api/v1/tags'),
				fetch('/api/v1/tag-categories')
			]);

			// Parse responses
			const tagsData = await tagsResponse.json();
			const categoriesData = await categoriesResponse.json();

			if (tagsData.success) {
				tags = tagsData.result || [];
				console.log('✅ Loaded tags with new schema:', tags);
				console.log('🔍 Sample tag structure:', tags[0]); // Show first tag structure
				console.log('🔍 Tags with category_ids:', tags.filter(t => t.category_ids && t.category_ids.length > 0));
			} else {
				console.log('❌ Failed to load tags with new schema:', tagsData);
			}

			if (categoriesData.success) {
				categories = categoriesData.result || [];
				console.log('✅ Loaded categories with new schema:', categories);
			}
		} catch (err) {
			toastStore.error('Failed to load tag data');
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
						t.category_ids && t.category_ids.includes(category.id)
							? { ...t, category_ids: t.category_ids.filter((id: number) => id !== category.id) }
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

	// Update the batchUpdateTagCategories function to use the new API
	async function batchUpdateTagCategories(changes: Array<{tagId: number, categoryId: number | null, action: 'add' | 'remove'}>) {
		try {
			// Use the new batch update API endpoint
			const response = await fetch('/api/v1/tag-categories/batch-update', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ changes })
			});

			if (response.ok) {
				const data = await response.json();
				if (data.success) {
					toastStore.success(`Successfully processed ${changes.length} tag changes`);
					
					// Update local state to reflect changes
					updateLocalStateAfterBatchChanges(changes);
					
					// Force reactivity by reassigning the arrays
					tags = [...tags];
					categories = [...categories];
					
					return true;
				} else {
					toastStore.error(data.error || 'Failed to process tag changes');
					return false;
				}
			} else {
				const errorText = await response.text();
				toastStore.error(`Failed to process tag changes (HTTP ${response.status})`);
				console.error('HTTP error:', errorText);
				return false;
			}
		} catch (error) {
			toastStore.error('Failed to process tag changes - network error');
			console.error('Error in batch update:', error);
			return false;
		}
	}

	// Improve assignTagToCategory with better error handling for removals
	async function assignTagToCategory(tagId: any, categoryId: any) {
		try {
			// Send null for removals instead of empty string
			const apiCategoryId = categoryId === null || categoryId === undefined ? null : categoryId;
			
			console.log(`Assigning tag ${tagId} to category: ${apiCategoryId} (${categoryId === null ? 'REMOVAL' : 'ASSIGNMENT'})`);
			
			const response = await masterVideoService.assignSubsiteTagToCategory('streaming', tagId, apiCategoryId);
			
			if (response.ok) {
				const responseData = await response.json();
				
				if (responseData.success) {
					// Update local state immediately for better UX
					if (categoryId === null) {
						// Remove tag from category
						tags = tags.map(t => 
							t.id === tagId 
								? { ...t, category_ids: t.category_ids?.filter((id: number) => id !== tagId) || [] }
								: t
						);
						
						// Also update categories to remove tag from tag_ids array
						categories = categories.map(c => {
							if (c.tag_ids && c.tag_ids.includes(tagId)) {
								return { ...c, tag_ids: c.tag_ids.filter((id: number) => id !== tagId) };
							}
							return c;
						});
						
						toastStore.success('Tag removed from category successfully');
					} else {
						// Add tag to category
						tags = tags.map(t => 
							t.id === tagId 
								? { ...t, category_ids: [...(t.category_ids || []), parseInt(categoryId)] }
								: t
						);
						
						// Also update categories to add tag to tag_ids array
						categories = categories.map(c => {
							if (c.id === categoryId) {
								const currentTagIds = c.tag_ids || [];
								if (!currentTagIds.includes(tagId)) {
									return { ...c, tag_ids: [...currentTagIds, tagId] };
								}
							}
							return c;
						});
						
						toastStore.success('Tag assigned to category successfully');
					}
				} else {
					// Backend returned success: false
					const errorMessage = responseData.error || 'Unknown backend error';
					console.error('Backend error:', responseData);
					
					if (categoryId === null) {
						toastStore.error(`Failed to remove tag from category: ${errorMessage}`);
					} else {
						toastStore.error(`Failed to assign tag to category: ${errorMessage}`);
					}
				}
			} else {
				// HTTP error response
				const errorText = await response.text();
				console.error('HTTP error response:', response.status, errorText);
				
				if (categoryId === null) {
					toastStore.error(`Failed to remove tag from category (HTTP ${response.status})`);
				} else {
					toastStore.error(`Failed to assign tag to category (HTTP ${response.status})`);
				}
			}
		} catch (err) {
			console.error('Network/other error in assignTagToCategory:', err);
			
			if (categoryId === null) {
				toastStore.error('Failed to remove tag from category - network error');
			} else {
				toastStore.error('Failed to assign tag to category - network error');
			}
		}
	}

	// Update local state after batch changes - updated for array relationships
	function updateLocalStateAfterBatchChanges(changes: Array<{tagId: number, categoryId: number | null, action: 'add' | 'remove'}>) {
		changes.forEach(change => {
			if (change.action === 'add' && change.categoryId) {
				// Add category to tag's category_ids array
				tags = tags.map(t => {
					if (t.id === change.tagId) {
						const currentCategoryIds = t.category_ids || [];
						if (!currentCategoryIds.includes(change.categoryId!)) {
							return { ...t, category_ids: [...currentCategoryIds, change.categoryId!] };
						}
					}
					return t;
				});
				
				// Add tag to category's tag_ids array
				categories = categories.map(c => {
					if (c.id === change.categoryId) {
						const currentTagIds = c.tag_ids || [];
						if (!currentTagIds.includes(change.tagId)) {
							return { ...c, tag_ids: [...currentTagIds, change.tagId] };
						}
					}
					return c;
				});
			} else if (change.action === 'remove') {
				if (change.categoryId) {
					// Remove specific category from tag's category_ids array
					tags = tags.map(t => {
						if (t.id === change.tagId && t.category_ids) {
							return { ...t, category_ids: t.category_ids.filter((id: number | null) => id !== change.categoryId) };
						}
						return t;
					});
					
					// Remove tag from category's tag_ids array
					categories = categories.map(c => {
						if (c.id === change.categoryId && c.tag_ids) {
							return { ...c, tag_ids: c.tag_ids.filter((id: number) => id !== change.tagId) };
						}
						return c;
					});
				} else {
					// Remove tag from all categories
					tags = tags.map(t => 
						t.id === change.tagId 
							? { ...t, category_ids: [] }
							: t
					);
					
					// Remove tag from all categories' tag_ids arrays
					categories = categories.map(c => {
						if (c.tag_ids && c.tag_ids.includes(change.tagId)) {
							return { ...c, tag_ids: c.tag_ids.filter((id: number) => id !== change.tagId) };
						}
						return c;
					});
				}
			}
		});
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

	async function updateCategory(category: any) {
		try {
			// For now, update local state directly since we don't have an update API
			// In a real implementation, you would call the backend API here
			categories = categories.map(c => 
				c.id === category.id 
					? { ...c, name: category.name, color: category.color, description: category.description }
					: c
			);
			
			toastStore.success('Category updated successfully');
			
			// TODO: Implement backend API call when available
			// const response = await masterVideoService.updateSubsiteCategory('streaming', category.id, {
			// 	name: category.name,
			// 	color: category.color,
			// 	description: category.description
			// });
		} catch (err) {
			toastStore.error('Failed to update category');
			console.error('Error updating category:', err);
		}
	}


	// Fix the event target issue in the template
	async function handleCategoryChange(event: Event, tagId: any) {
		const target = event.target as HTMLSelectElement;
		if (target) {
			await assignTagToCategory(tagId, target.value);
		}
	}

	// Add new function for querying videos by category
	async function getVideosByCategory(categoryId: number) {
		try {
			// Get the category to find its tags
			const category = categories.find(c => c.id === categoryId);
			if (!category) {
				return [];
			}
			
			// Get tags for this category
			const categoryTags = tags.filter(tag => tag.category_ids && tag.category_ids.includes(categoryId));
			if (categoryTags.length === 0) {
				return [];
			}
			
			// For now, return empty array since we don't have getVideosByTags
			// TODO: Implement when backend supports video queries by tags
			console.log(`Category "${category.name}" has ${categoryTags.length} tags:`, categoryTags.map(t => t.word));
			return [];
			
			// When backend supports it, use:
			// const response = await masterVideoService.getVideosByTags('streaming', categoryTags.map(t => t.id));
			// if (response.ok) {
			// 	const data = await response.json();
			// 	return data.result || [];
			// }
		} catch (error) {
			console.error('Error getting videos by category:', error);
			return [];
		}
	}

	// Add function to get category statistics
	async function getCategoryStats(categoryId: number) {
		try {
			const videos = await getVideosByCategory(categoryId);
			const category = categories.find(c => c.id === categoryId);
			
			return {
				categoryId,
				categoryName: category?.name || 'Unknown',
				videoCount: videos.length,
				tagCount: category?.tag_ids?.length || 0,
				videos: videos.slice(0, 10) // Limit for performance
			};
		} catch (error) {
			console.error('Error getting category stats:', error);
			return null;
		}
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
		<TagsTab 
			{tags}
			{categories}
			{loading}
			bind:newTag
			onAddTag={addTag}
			onDeleteTag={deleteTag}
			onCategoryChange={({ tagId, categoryId }: { tagId: number; categoryId: number }) => assignTagToCategory(tagId, categoryId)}
			onToggleStatus={toggleTagActiveStatus}
			onAddToExclusions={addArticleExclusionFromTag}
		/>
	{/if}

	<!-- Categories Tab -->
	{#if activeTab === 'categories'}
		<CategoriesTab 
			{categories}
			{tags}
			{loading}
			bind:newCategory
			bind:newCategoryColor
			onAddCategory={addCategory}
			onDeleteCategory={deleteCategory}
			onUpdateCategory={updateCategory}
			onBatchTagChanges={batchUpdateTagCategories}
		/>
	{/if}

	<!-- Article Exclusions Tab -->
	{#if activeTab === 'exclusions'}
		<ExclusionsTab 
			{articleExclusions}
			bind:newExclusion
			{exclusionsLoading}
			onAddExclusion={addArticleExclusion}
			onToggleExclusion={toggleArticleExclusion}
			onRemoveExclusion={removeArticleExclusion}
		/>
	{/if}
</div>


<style>



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





	@media (max-width: 768px) {
		.streaming-tags-categories-page {
			padding: 1rem;
		}

		.header-actions {
			flex-direction: column;
			align-items: center;
		}

		.tab-navigation {
			justify-content: center;
		}
	}
</style>

