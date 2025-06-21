<script lang="ts">
	import { onMount } from 'svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import AdDisplay from '$lib/components/AdDisplay.svelte';
	import { 
		MOCK_ARTICLES, 
		ARTICLE_CATEGORIES, 
		ARTICLE_TAGS, 
		ARTICLE_AUTHORS,
		searchArticles,
		getArticlesByCategory,
		getArticlesByTag,
		getFeaturedArticles,
		getRecentArticles
	} from '$lib/mockData';

	// State management
	let articles = MOCK_ARTICLES;
	let filteredArticles = articles;
	let searchQuery = '';
	let selectedCategory: number | null = null;
	let selectedTag: string | null = null;
	let currentPage = 1;
	let articlesPerPage = 6;
	let sortBy = 'date'; // 'date', 'popularity', 'title'
	let sortOrder = 'desc'; // 'asc', 'desc'
	let isLoading = false;

	// Computed values
	$: totalPages = Math.ceil(filteredArticles.length / articlesPerPage);
	$: paginatedArticles = filteredArticles.slice(
		(currentPage - 1) * articlesPerPage,
		currentPage * articlesPerPage
	);
	$: featuredArticles = getFeaturedArticles();

	onMount(() => {
		// Initialize with recent articles sorted by date
		applyFilters();
	});

	function applyFilters() {
		isLoading = true;
		let result = [...articles];

		// Apply search filter
		if (searchQuery.trim()) {
			result = searchArticles(searchQuery);
		}

		// Apply category filter
		if (selectedCategory) {
			result = result.filter(article => article.categoryId === selectedCategory);
		}

		// Apply tag filter
		if (selectedTag && selectedTag !== null) {
			result = result.filter(article => article.tags.includes(selectedTag!));
		}

		// Apply sorting
		result.sort((a, b) => {
			let comparison = 0;
			
			switch (sortBy) {
				case 'date':
					comparison = new Date(a.publishedAt).getTime() - new Date(b.publishedAt).getTime();
					break;
				case 'popularity':
					comparison = a.views - b.views;
					break;
				case 'title':
					comparison = a.title.localeCompare(b.title);
					break;
			}

			return sortOrder === 'desc' ? -comparison : comparison;
		});

		filteredArticles = result;
		currentPage = 1; // Reset to first page when filters change
		
		setTimeout(() => {
			isLoading = false;
		}, 300);
	}

	function handleSearch() {
		applyFilters();
	}

	function selectCategory(categoryId: number | null) {
		selectedCategory = categoryId;
		selectedTag = null; // Clear tag filter when selecting category
		applyFilters();
	}

	function selectTag(tag: string | null) {
		selectedTag = tag;
		selectedCategory = null; // Clear category filter when selecting tag
		applyFilters();
	}

	function clearFilters() {
		searchQuery = '';
		selectedCategory = null;
		selectedTag = null;
		sortBy = 'date';
		sortOrder = 'desc';
		applyFilters();
	}

	function changePage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
			// Scroll to top of articles section
			document.getElementById('articles-section')?.scrollIntoView({ behavior: 'smooth' });
		}
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function getAuthorById(authorId: number) {
		return ARTICLE_AUTHORS.find(author => author.id === authorId);
	}

	function getCategoryById(categoryId: number) {
		return ARTICLE_CATEGORIES.find(category => category.id === categoryId);
	}
</script>

<svelte:head>
	<title>Articles & Research - BOME | Book of Mormon Evidence</title>
	<meta name="description" content="Explore scholarly articles, archaeological evidence, and research supporting the Book of Mormon. Expert analysis and latest discoveries." />
	<meta name="keywords" content="Book of Mormon, articles, research, archaeology, evidence, scholarly analysis" />
	<meta property="og:title" content="Articles & Research - BOME" />
	<meta property="og:description" content="Explore scholarly articles, archaeological evidence, and research supporting the Book of Mormon." />
	<meta property="og:url" content="https://bome.org/articles" />
</svelte:head>

<Navigation />

<main class="blog-container">
	<!-- Hero Section with Featured Articles -->
	<section class="hero-section">
		<div class="hero-content">
			<h1 class="hero-title">Articles & Research</h1>
			<p class="hero-description">
				Explore the latest scholarly articles, archaeological discoveries, and research supporting the Book of Mormon. 
				Expert analysis from leading academics and researchers worldwide.
			</p>
		</div>
		
		<!-- Ad Placement: Header Banner -->
		<div class="ad-placement">
			<AdDisplay placement="articles-header" />
		</div>
	</section>

	<!-- Ad Placement: Between Featured and Filters -->
	<div class="ad-placement">
		<AdDisplay placement="articles-mid" />
	</div>

	<!-- Featured Articles Carousel -->
	{#if featuredArticles.length > 0}
		<section class="featured-section">
			<h2 class="section-title">Featured Articles</h2>
			<div class="featured-grid">
				{#each featuredArticles as article}
					<article class="featured-card">
						<div class="featured-image">
							<img src={article.featuredImage} alt={article.title} />
							<div class="featured-overlay">
								<span class="featured-badge">Featured</span>
							</div>
						</div>
						<div class="featured-content">
							<div class="article-meta">
								<span class="category-badge">
									{getCategoryById(article.categoryId)?.name || 'Uncategorized'}
								</span>
								<span class="read-time">{article.readTime} min read</span>
							</div>
							<h3 class="featured-title">
								<a href="/articles/{article.slug}">{article.title}</a>
							</h3>
							<p class="featured-excerpt">{article.excerpt}</p>
							<div class="featured-footer">
								<div class="author-info">
									<span class="author-name">{getAuthorById(article.authorId)?.name}</span>
									<span class="publish-date">{formatDate(article.publishedAt)}</span>
								</div>
								<div class="article-stats">
									<span class="views">{article.views.toLocaleString()} views</span>
									<span class="likes">{article.likes} likes</span>
								</div>
							</div>
						</div>
					</article>
				{/each}
			</div>
		</section>
	{/if}

	<!-- Ad Placement: Mid-page -->
	<div class="ad-placement">
		<AdDisplay placement="blog-mid" />
	</div>

	<div class="main-content">
		<!-- Sidebar -->
		<aside class="sidebar">
			<!-- Search -->
			<div class="search-section">
				<h3 class="sidebar-title">Search Articles</h3>
				<div class="search-container">
					<input 
						type="text" 
						placeholder="Search articles..." 
						bind:value={searchQuery}
						on:input={handleSearch}
						class="search-input"
					/>
					<button class="search-button" on:click={handleSearch}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="11" cy="11" r="8"></circle>
							<path d="m21 21-4.35-4.35"></path>
						</svg>
					</button>
				</div>
			</div>

			<!-- Categories -->
			<div class="categories-section">
				<h3 class="sidebar-title">Categories</h3>
				<div class="categories-list">
					<button 
						class="category-item" 
						class:active={selectedCategory === null}
						on:click={() => selectCategory(null)}
					>
						All Articles ({articles.length})
					</button>
					{#each ARTICLE_CATEGORIES as category}
						<button 
							class="category-item" 
							class:active={selectedCategory === category.id}
							on:click={() => selectCategory(category.id)}
						>
							{category.name} ({category.count})
						</button>
					{/each}
				</div>
			</div>

			<!-- Popular Tags -->
			<div class="tags-section">
				<h3 class="sidebar-title">Popular Tags</h3>
				<div class="tags-cloud">
					{#each ARTICLE_TAGS.slice(0, 15) as tag}
						<button 
							class="tag-item" 
							class:active={selectedTag === tag}
							on:click={() => selectTag(selectedTag === tag ? null : tag)}
						>
							{tag}
						</button>
					{/each}
				</div>
			</div>

			<!-- Ad Placement: Sidebar -->
			<div class="ad-placement sidebar-ad">
				<AdDisplay placement="articles-sidebar" />
			</div>
		</aside>

		<!-- Articles Section -->
		<section class="articles-section" id="articles-section">
			<!-- Filters and Sorting -->
			<div class="filters-bar">
				<div class="active-filters">
					{#if searchQuery}
						<span class="filter-tag">
							Search: "{searchQuery}"
							<button on:click={() => { searchQuery = ''; applyFilters(); }}>×</button>
						</span>
					{/if}
					{#if selectedCategory}
						<span class="filter-tag">
							Category: {getCategoryById(selectedCategory)?.name}
							<button on:click={() => selectCategory(null)}>×</button>
						</span>
					{/if}
					{#if selectedTag}
						<span class="filter-tag">
							Tag: {selectedTag}
							<button on:click={() => selectTag(null)}>×</button>
						</span>
					{/if}
					{#if searchQuery || selectedCategory || selectedTag}
						<button class="clear-filters" on:click={clearFilters}>Clear All</button>
					{/if}
				</div>

				<div class="sort-controls">
					<label class="sort-label">Sort by:</label>
					<select bind:value={sortBy} on:change={applyFilters} class="sort-select">
						<option value="date">Date</option>
						<option value="popularity">Popularity</option>
						<option value="title">Title</option>
					</select>
					<select bind:value={sortOrder} on:change={applyFilters} class="sort-select">
						<option value="desc">Newest First</option>
						<option value="asc">Oldest First</option>
					</select>
				</div>
			</div>

			<!-- Articles Grid -->
			<div class="articles-grid" class:loading={isLoading}>
				{#if isLoading}
					<div class="loading-spinner">
						<div class="spinner"></div>
						<p>Loading articles...</p>
					</div>
				{:else if paginatedArticles.length === 0}
					<div class="no-results">
						<h3>No articles found</h3>
						<p>Try adjusting your search or filter criteria.</p>
						<button class="clear-button" on:click={clearFilters}>Clear Filters</button>
					</div>
				{:else}
					{#each paginatedArticles as article, index}
						<article class="article-card">
							<div class="article-image">
								<img src={article.featuredImage} alt={article.title} />
								<div class="article-overlay">
									<span class="category-badge">
										{getCategoryById(article.categoryId)?.name}
									</span>
								</div>
							</div>
							<div class="article-content">
								<div class="article-meta">
									<span class="author-name">{getAuthorById(article.authorId)?.name}</span>
									<span class="publish-date">{formatDate(article.publishedAt)}</span>
									<span class="read-time">{article.readTime} min read</span>
								</div>
								<h3 class="article-title">
									<a href="/articles/{article.slug}">{article.title}</a>
								</h3>
								<p class="article-excerpt">{article.excerpt}</p>
								<div class="article-tags">
									{#each article.tags.slice(0, 3) as tag}
										<button 
											class="tag-button" 
											on:click={() => selectTag(tag)}
										>
											{tag}
										</button>
									{/each}
								</div>
								<div class="article-footer">
									<div class="article-stats">
										<span class="views">{article.views.toLocaleString()} views</span>
										<span class="likes">{article.likes} likes</span>
									</div>
									<a href="/articles/{article.slug}" class="read-more">Read More</a>
								</div>
							</div>
						</article>

						<!-- Ad Placement: Between Articles -->
						{#if index === 2}
							<div class="ad-placement feed-ad">
								<AdDisplay placement="articles-feed" />
							</div>
						{/if}
					{/each}
				{/if}
			</div>

			<!-- Pagination -->
			{#if totalPages > 1 && !isLoading}
				<div class="pagination">
					<button 
						class="pagination-button" 
						disabled={currentPage === 1}
						on:click={() => changePage(currentPage - 1)}
					>
						Previous
					</button>
					
					{#each Array(totalPages) as _, i}
						<button 
							class="pagination-button page-number" 
							class:active={currentPage === i + 1}
							on:click={() => changePage(i + 1)}
						>
							{i + 1}
						</button>
					{/each}
					
					<button 
						class="pagination-button" 
						disabled={currentPage === totalPages}
						on:click={() => changePage(currentPage + 1)}
					>
						Next
					</button>
				</div>
			{/if}
		</section>
	</div>

	<!-- Ad Placement: Footer -->
	<div class="ad-placement">
		<AdDisplay placement="articles-footer" />
	</div>
</main>

<Footer />

<style>
	.blog-container {
		min-height: 100vh;
		background: var(--bg-primary);
		padding-top: 80px;
	}

	/* Hero Section */
	.hero-section {
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
		padding: 4rem 2rem;
		text-align: center;
		position: relative;
		overflow: hidden;
	}

	.hero-section::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="1" fill="white" opacity="0.1"/><circle cx="80" cy="30" r="0.5" fill="white" opacity="0.1"/><circle cx="40" cy="60" r="1.5" fill="white" opacity="0.1"/></svg>');
		opacity: 0.3;
	}

	.hero-content {
		position: relative;
		z-index: 2;
		max-width: 800px;
		margin: 0 auto;
	}

	.hero-title {
		font-size: 3.5rem;
		font-weight: 900;
		color: #ffd700;
		text-shadow: 4px 4px 8px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 15px rgba(255, 215, 0, 0.4));
		margin-bottom: 1.5rem;
		line-height: 1.2;
	}

	.hero-description {
		font-size: 1.3rem;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		line-height: 1.6;
		margin-bottom: 2rem;
	}

	/* Featured Section */
	.featured-section {
		padding: 4rem 2rem;
		background: var(--bg-secondary);
	}

	.section-title {
		font-size: 2.5rem;
		font-weight: 700;
		color: var(--text-primary);
		text-align: center;
		margin-bottom: 3rem;
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
	}

	.featured-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
		gap: 2rem;
		max-width: 1200px;
		margin: 0 auto;
	}

	.featured-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		overflow: hidden;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
		transition: all 0.3s ease;
	}

	.featured-card:hover {
		transform: translateY(-5px);
		box-shadow: 0 25px 50px rgba(0, 0, 0, 0.4);
	}

	.featured-image {
		position: relative;
		height: 250px;
		overflow: hidden;
	}

	.featured-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.3s ease;
	}

	.featured-card:hover .featured-image img {
		transform: scale(1.05);
	}

	.featured-overlay {
		position: absolute;
		top: 1rem;
		right: 1rem;
	}

	.featured-badge {
		background: linear-gradient(135deg, #ff6b6b 0%, #ee5a24 100%);
		color: white;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.9rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.featured-content {
		padding: 2rem;
	}

	.featured-title {
		font-size: 1.5rem;
		font-weight: 700;
		margin-bottom: 1rem;
		line-height: 1.3;
	}

	.featured-title a {
		color: var(--text-primary);
		text-decoration: none;
		transition: color 0.3s ease;
	}

	.featured-title a:hover {
		color: var(--accent-primary);
	}

	.featured-excerpt {
		color: var(--text-secondary);
		line-height: 1.6;
		margin-bottom: 1.5rem;
	}

	.featured-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 1rem;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	/* Main Content Layout */
	.main-content {
		display: grid;
		grid-template-columns: 300px 1fr;
		gap: 3rem;
		max-width: 1400px;
		margin: 0 auto;
		padding: 3rem 2rem;
	}

	/* Sidebar */
	.sidebar {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.sidebar-title {
		font-size: 1.3rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
		padding-bottom: 0.5rem;
		border-bottom: 2px solid var(--accent-primary);
	}

	/* Search Section */
	.search-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.search-container {
		position: relative;
	}

	.search-input {
		width: 100%;
		padding: 0.8rem 3rem 0.8rem 1rem;
		border: 1px solid rgba(255, 255, 255, 0.3);
		border-radius: 10px;
		background: rgba(255, 255, 255, 0.1);
		color: var(--text-primary);
		font-size: 1rem;
		backdrop-filter: blur(10px);
		transition: all 0.3s ease;
	}

	.search-input:focus {
		outline: none;
		border-color: var(--accent-primary);
		box-shadow: 0 0 0 3px rgba(255, 107, 107, 0.2);
	}

	.search-button {
		position: absolute;
		right: 0.5rem;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 5px;
		transition: all 0.3s ease;
	}

	.search-button:hover {
		color: var(--accent-primary);
		background: rgba(255, 255, 255, 0.1);
	}

	.search-button svg {
		width: 20px;
		height: 20px;
	}

	/* Categories Section */
	.categories-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.categories-list {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.category-item {
		background: none;
		border: none;
		color: var(--text-secondary);
		padding: 0.8rem 1rem;
		text-align: left;
		border-radius: 10px;
		cursor: pointer;
		transition: all 0.3s ease;
		font-size: 0.95rem;
	}

	.category-item:hover,
	.category-item.active {
		background: rgba(255, 107, 107, 0.2);
		color: var(--text-primary);
		transform: translateX(5px);
	}

	/* Tags Section */
	.tags-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	.tags-cloud {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.tag-item {
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		color: var(--text-secondary);
		padding: 0.4rem 0.8rem;
		border-radius: 20px;
		font-size: 0.85rem;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.tag-item:hover,
	.tag-item.active {
		background: var(--accent-primary);
		color: white;
		border-color: var(--accent-primary);
		transform: scale(1.05);
	}

	/* Articles Section */
	.articles-section {
		min-height: 600px;
	}

	/* Filters Bar */
	.filters-bar {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		padding: 1.5rem;
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
		flex-wrap: wrap;
		gap: 1rem;
	}

	.active-filters {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.filter-tag {
		background: var(--accent-primary);
		color: white;
		padding: 0.4rem 0.8rem;
		border-radius: 20px;
		font-size: 0.9rem;
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.filter-tag button {
		background: none;
		border: none;
		color: white;
		cursor: pointer;
		font-size: 1.2rem;
		padding: 0;
		line-height: 1;
		border-radius: 50%;
		width: 20px;
		height: 20px;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: background 0.3s ease;
	}

	.filter-tag button:hover {
		background: rgba(255, 255, 255, 0.3);
	}

	.clear-filters {
		background: transparent;
		border: 1px solid var(--accent-primary);
		color: var(--accent-primary);
		padding: 0.5rem 1rem;
		border-radius: 20px;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.clear-filters:hover {
		background: var(--accent-primary);
		color: white;
	}

	.sort-controls {
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.sort-label {
		color: var(--text-secondary);
		font-size: 0.9rem;
	}

	.sort-select {
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.3);
		color: var(--text-primary);
		padding: 0.5rem;
		border-radius: 8px;
		font-size: 0.9rem;
		cursor: pointer;
		backdrop-filter: blur(10px);
	}

	/* Articles Grid */
	.articles-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 2rem;
		margin-bottom: 3rem;
	}

	.articles-grid.loading {
		opacity: 0.7;
		pointer-events: none;
	}

	.article-card {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		overflow: hidden;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2);
		transition: all 0.3s ease;
		display: flex;
		flex-direction: column;
	}

	.article-card:hover {
		transform: translateY(-5px);
		box-shadow: 0 20px 45px rgba(0, 0, 0, 0.3);
	}

	.article-image {
		position: relative;
		height: 200px;
		overflow: hidden;
	}

	.article-image img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 0.3s ease;
	}

	.article-card:hover .article-image img {
		transform: scale(1.05);
	}

	.article-overlay {
		position: absolute;
		top: 1rem;
		left: 1rem;
	}

	.category-badge {
		background: rgba(0, 0, 0, 0.7);
		color: white;
		padding: 0.3rem 0.8rem;
		border-radius: 15px;
		font-size: 0.8rem;
		font-weight: 600;
		backdrop-filter: blur(10px);
	}

	.article-content {
		padding: 1.5rem;
		flex: 1;
		display: flex;
		flex-direction: column;
	}

	.article-meta {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 1rem;
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.author-name {
		font-weight: 600;
		color: var(--accent-primary);
	}

	.read-time {
		background: rgba(255, 255, 255, 0.1);
		padding: 0.2rem 0.6rem;
		border-radius: 10px;
	}

	.article-title {
		font-size: 1.3rem;
		font-weight: 700;
		margin-bottom: 1rem;
		line-height: 1.4;
	}

	.article-title a {
		color: var(--text-primary);
		text-decoration: none;
		transition: color 0.3s ease;
	}

	.article-title a:hover {
		color: var(--accent-primary);
	}

	.article-excerpt {
		color: var(--text-secondary);
		line-height: 1.6;
		margin-bottom: 1rem;
		flex: 1;
	}

	.article-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.tag-button {
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		color: var(--text-secondary);
		padding: 0.3rem 0.6rem;
		border-radius: 15px;
		font-size: 0.8rem;
		cursor: pointer;
		transition: all 0.3s ease;
	}

	.tag-button:hover {
		background: var(--accent-primary);
		color: white;
		border-color: var(--accent-primary);
	}

	.article-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 1rem;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.article-stats {
		display: flex;
		gap: 1rem;
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.read-more {
		color: var(--accent-primary);
		text-decoration: none;
		font-weight: 600;
		transition: all 0.3s ease;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		border: 1px solid var(--accent-primary);
	}

	.read-more:hover {
		background: var(--accent-primary);
		color: white;
	}

	/* Loading Spinner */
	.loading-spinner {
		grid-column: 1 / -1;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem;
		color: var(--text-secondary);
	}

	.spinner {
		width: 40px;
		height: 40px;
		border: 3px solid rgba(255, 107, 107, 0.3);
		border-top: 3px solid var(--accent-primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	/* No Results */
	.no-results {
		grid-column: 1 / -1;
		text-align: center;
		padding: 4rem;
		color: var(--text-secondary);
	}

	.no-results h3 {
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.clear-button {
		background: var(--accent-primary);
		color: white;
		border: none;
		padding: 0.8rem 2rem;
		border-radius: 25px;
		cursor: pointer;
		font-weight: 600;
		margin-top: 1rem;
		transition: all 0.3s ease;
	}

	.clear-button:hover {
		background: #e55a5a;
		transform: translateY(-2px);
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.5rem;
		margin-top: 3rem;
	}

	.pagination-button {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border: 1px solid rgba(255, 255, 255, 0.2);
		color: var(--text-primary);
		padding: 0.8rem 1.2rem;
		border-radius: 10px;
		cursor: pointer;
		transition: all 0.3s ease;
		font-weight: 500;
	}

	.pagination-button:hover:not(:disabled) {
		background: var(--accent-primary);
		color: white;
		border-color: var(--accent-primary);
	}

	.pagination-button.active {
		background: var(--accent-primary);
		color: white;
		border-color: var(--accent-primary);
	}

	.pagination-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.pagination-button.page-number {
		min-width: 45px;
		text-align: center;
	}

	/* Ad Placements */
	.ad-placement {
		margin: 2rem 0;
	}

	.sidebar-ad {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
		text-align: center;
	}

	.inline-ad {
		grid-column: 1 / -1;
		background: var(--bg-glass-dark);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
		text-align: center;
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.main-content {
			grid-template-columns: 280px 1fr;
			gap: 2rem;
		}
	}

	@media (max-width: 992px) {
		.main-content {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.sidebar {
			order: 2;
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
			gap: 1.5rem;
		}

		.articles-section {
			order: 1;
		}

		.filters-bar {
			flex-direction: column;
			align-items: stretch;
		}

		.active-filters {
			justify-content: center;
		}

		.sort-controls {
			justify-content: center;
		}
	}

	@media (max-width: 768px) {
		.hero-title {
			font-size: 2.5rem;
		}

		.hero-description {
			font-size: 1.1rem;
		}

		.featured-grid {
			grid-template-columns: 1fr;
		}

		.articles-grid {
			grid-template-columns: 1fr;
		}

		.sidebar {
			grid-template-columns: 1fr;
		}

		.main-content {
			padding: 2rem 1rem;
		}

		.hero-section {
			padding: 3rem 1rem;
		}

		.featured-section {
			padding: 3rem 1rem;
		}
	}

	@media (max-width: 480px) {
		.hero-title {
			font-size: 2rem;
		}

		.filters-bar {
			padding: 1rem;
		}

		.article-content {
			padding: 1rem;
		}

		.featured-content {
			padding: 1.5rem;
		}
	}
</style> 