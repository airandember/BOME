<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import Navigation from '$lib/components/Navigation.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import AdDisplay from '$lib/components/AdDisplay.svelte';
	import { 
		MOCK_ARTICLES, 
		ARTICLE_CATEGORIES, 
		ARTICLE_AUTHORS,
		getArticlesByCategory,
		getArticlesByAuthor
	} from '$lib/mockData';

	let article: any = null;
	let author: any = null;
	let category: any = null;
	let relatedArticles: any[] = [];
	let authorArticles: any[] = [];
	let isLoading = true;
	let error = '';

	$: slug = $page.params.slug;

	onMount(() => {
		loadArticle();
	});

	function loadArticle() {
		isLoading = true;
		error = '';

		// Find article by slug
		const foundArticle = MOCK_ARTICLES.find(a => a.slug === slug);
		
		if (!foundArticle) {
			error = 'Article not found';
			isLoading = false;
			return;
		}

		article = foundArticle;
		author = ARTICLE_AUTHORS.find(a => a.id === article.authorId);
		category = ARTICLE_CATEGORIES.find(c => c.id === article.categoryId);

		// Get related articles (same category, excluding current)
		relatedArticles = getArticlesByCategory(article.categoryId)
			.filter(a => a.id !== article.id)
			.slice(0, 3);

		// Get other articles by same author (excluding current)
		authorArticles = getArticlesByAuthor(article.authorId)
			.filter(a => a.id !== article.id)
			.slice(0, 2);

		isLoading = false;

		// Scroll to top
		window.scrollTo(0, 0);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function shareArticle(platform: string) {
		const url = encodeURIComponent(window.location.href);
		const title = encodeURIComponent(article.title);
		
		let shareUrl = '';
		
		switch (platform) {
			case 'twitter':
				shareUrl = `https://twitter.com/intent/tweet?url=${url}&text=${title}`;
				break;
			case 'facebook':
				shareUrl = `https://www.facebook.com/sharer/sharer.php?u=${url}`;
				break;
			case 'linkedin':
				shareUrl = `https://www.linkedin.com/sharing/share-offsite/?url=${url}`;
				break;
			case 'email':
				shareUrl = `mailto:?subject=${title}&body=Check out this article: ${url}`;
				break;
		}
		
		if (shareUrl) {
			window.open(shareUrl, '_blank', 'width=600,height=400');
		}
	}

	function copyLink() {
		navigator.clipboard.writeText(window.location.href).then(() => {
			// Could add a toast notification here
			alert('Link copied to clipboard!');
		});
	}
</script>

<svelte:head>
	{#if article}
		<title>{article.title} - BOME | Book of Mormon Evidence</title>
		<meta name="description" content={article.excerpt} />
		<meta name="keywords" content={article.tags.join(', ')} />
		<meta property="og:title" content={article.title} />
		<meta property="og:description" content={article.excerpt} />
		<meta property="og:image" content={article.featuredImage} />
		<meta property="og:type" content="article" />
		<meta name="twitter:card" content="summary_large_image" />
		<meta name="twitter:title" content={article.title} />
		<meta name="twitter:description" content={article.excerpt} />
		<meta name="twitter:image" content={article.featuredImage} />
	{:else}
		<title>Article - BOME | Book of Mormon Evidence</title>
	{/if}
</svelte:head>

<Navigation />

<main class="article-container">
	{#if isLoading}
		<div class="loading-container">
			<div class="loading-spinner">
				<div class="spinner"></div>
				<p>Loading article...</p>
			</div>
		</div>
	{:else if error}
		<div class="error-container">
			<div class="error-content">
				<h1>Article Not Found</h1>
				<p>The article you're looking for doesn't exist or has been moved.</p>
				<a href="/articles" class="back-button">← Back to Articles</a>
			</div>
		</div>
	{:else if article}
		<!-- Article Header -->
		<header class="article-header">
			<div class="header-content">
				<nav class="breadcrumb">
					<a href="/">Home</a>
					<span class="separator">›</span>
					<a href="/articles">Articles</a>
					<span class="separator">›</span>
					<a href="/articles?category={category?.id}">{category?.name}</a>
					<span class="separator">›</span>
					<span class="current">{article.title}</span>
				</nav>

				<div class="article-meta-header">
					<span class="category-badge">{category?.name}</span>
					<div class="article-tags-header">
						{#each article.tags.slice(0, 3) as tag}
							<span class="tag-badge">{tag}</span>
						{/each}
					</div>
				</div>

				<h1 class="article-title">{article.title}</h1>
				<p class="article-excerpt">{article.excerpt}</p>

				<div class="article-info">
					<div class="author-info">
						<img src={author?.avatar} alt={author?.name} class="author-avatar" />
						<div class="author-details">
							<h3 class="author-name">{author?.name}</h3>
							<p class="author-title">{author?.title}</p>
							<p class="author-institution">{author?.institution}</p>
						</div>
					</div>
					<div class="article-stats">
						<div class="stat-item">
							<span class="stat-label">Published</span>
							<span class="stat-value">{formatDate(article.publishedAt)}</span>
						</div>
						<div class="stat-item">
							<span class="stat-label">Read Time</span>
							<span class="stat-value">{article.readTime} min</span>
						</div>
						<div class="stat-item">
							<span class="stat-label">Views</span>
							<span class="stat-value">{article.views.toLocaleString()}</span>
						</div>
						<div class="stat-item">
							<span class="stat-label">Likes</span>
							<span class="stat-value">{article.likes}</span>
						</div>
					</div>
				</div>
			</div>

			<!-- Featured Image -->
			<div class="featured-image-container">
				<img src={article.featuredImage} alt={article.title} class="featured-image" />
			</div>
		</header>

		<div class="content-wrapper">
			<!-- Main Article Content -->
			<article class="article-content">
				<!-- Ad Placement: Top of Article -->
				<div class="ad-placement">
					<AdDisplay placement="article-top" />
				</div>

				<div class="article-body">
					{@html article.content.replace(/\n/g, '</p><p>')}
				</div>

				<!-- Article Tags -->
				<div class="article-tags-section">
					<h3>Tags</h3>
					<div class="tags-list">
						{#each article.tags as tag}
							<a href="/articles?tag={encodeURIComponent(tag)}" class="tag-link">{tag}</a>
						{/each}
					</div>
				</div>

				<!-- Share Section -->
				<div class="share-section">
					<h3>Share this article</h3>
					<div class="share-buttons">
						<button class="share-button twitter" on:click={() => shareArticle('twitter')}>
							<svg viewBox="0 0 24 24" fill="currentColor">
								<path d="M23.953 4.57a10 10 0 01-2.825.775 4.958 4.958 0 002.163-2.723c-.951.555-2.005.959-3.127 1.184a4.92 4.92 0 00-8.384 4.482C7.69 8.095 4.067 6.13 1.64 3.162a4.822 4.822 0 00-.666 2.475c0 1.71.87 3.213 2.188 4.096a4.904 4.904 0 01-2.228-.616v.06a4.923 4.923 0 003.946 4.827 4.996 4.996 0 01-2.212.085 4.936 4.936 0 004.604 3.417 9.867 9.867 0 01-6.102 2.105c-.39 0-.779-.023-1.17-.067a13.995 13.995 0 007.557 2.209c9.053 0 13.998-7.496 13.998-13.985 0-.21 0-.42-.015-.63A9.935 9.935 0 0024 4.59z"/>
							</svg>
							Twitter
						</button>
						<button class="share-button facebook" on:click={() => shareArticle('facebook')}>
							<svg viewBox="0 0 24 24" fill="currentColor">
								<path d="M24 12.073c0-6.627-5.373-12-12-12s-12 5.373-12 12c0 5.99 4.388 10.954 10.125 11.854v-8.385H7.078v-3.47h3.047V9.43c0-3.007 1.792-4.669 4.533-4.669 1.312 0 2.686.235 2.686.235v2.953H15.83c-1.491 0-1.956.925-1.956 1.874v2.25h3.328l-.532 3.47h-2.796v8.385C19.612 23.027 24 18.062 24 12.073z"/>
							</svg>
							Facebook
						</button>
						<button class="share-button linkedin" on:click={() => shareArticle('linkedin')}>
							<svg viewBox="0 0 24 24" fill="currentColor">
								<path d="M20.447 20.452h-3.554v-5.569c0-1.328-.027-3.037-1.852-3.037-1.853 0-2.136 1.445-2.136 2.939v5.667H9.351V9h3.414v1.561h.046c.477-.9 1.637-1.85 3.37-1.85 3.601 0 4.267 2.37 4.267 5.455v6.286zM5.337 7.433c-1.144 0-2.063-.926-2.063-2.065 0-1.138.92-2.063 2.063-2.063 1.14 0 2.064.925 2.064 2.063 0 1.139-.925 2.065-2.064 2.065zm1.782 13.019H3.555V9h3.564v11.452zM22.225 0H1.771C.792 0 0 .774 0 1.729v20.542C0 23.227.792 24 1.771 24h20.451C23.2 24 24 23.227 24 22.271V1.729C24 .774 23.2 0 22.222 0h.003z"/>
							</svg>
							LinkedIn
						</button>
						<button class="share-button email" on:click={() => shareArticle('email')}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"></path>
								<polyline points="22,6 12,13 2,6"></polyline>
							</svg>
							Email
						</button>
						<button class="share-button copy" on:click={copyLink}>
							<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path>
								<path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>
							</svg>
							Copy Link
						</button>
					</div>
				</div>

				<!-- Ad Placement: Bottom of Article -->
				<div class="ad-placement">
					<AdDisplay placement="article-bottom" />
				</div>
			</article>

			<!-- Sidebar -->
			<aside class="article-sidebar">
				<!-- Author Bio -->
				<div class="author-bio-card">
					<h3>About the Author</h3>
					<div class="author-bio">
						<img src={author?.avatar} alt={author?.name} class="author-avatar-large" />
						<div class="author-bio-content">
							<h4>{author?.name}</h4>
							<p class="author-credentials">{author?.title}<br>{author?.institution}</p>
							<p class="author-description">{author?.bio}</p>
							<div class="author-stats">
								<span>{author?.articlesCount} Articles</span>
							</div>
						</div>
					</div>
				</div>

				<!-- More from Author -->
				{#if authorArticles.length > 0}
					<div class="related-section">
						<h3>More from {author?.name}</h3>
						<div class="related-articles">
							{#each authorArticles as relatedArticle}
								<article class="related-article">
									<img src={relatedArticle.featuredImage} alt={relatedArticle.title} class="related-image" />
									<div class="related-content">
										<h4><a href="/articles/{relatedArticle.slug}">{relatedArticle.title}</a></h4>
										<p class="related-meta">{formatDate(relatedArticle.publishedAt)} • {relatedArticle.readTime} min read</p>
									</div>
								</article>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Related Articles -->
				{#if relatedArticles.length > 0}
					<div class="related-section">
						<h3>Related Articles</h3>
						<div class="related-articles">
							{#each relatedArticles as relatedArticle}
								<article class="related-article">
									<img src={relatedArticle.featuredImage} alt={relatedArticle.title} class="related-image" />
									<div class="related-content">
										<h4><a href="/articles/{relatedArticle.slug}">{relatedArticle.title}</a></h4>
										<p class="related-meta">{formatDate(relatedArticle.publishedAt)} • {relatedArticle.readTime} min read</p>
									</div>
								</article>
							{/each}
						</div>
					</div>
				{/if}

				<!-- Ad Placement: Sidebar -->
				<div class="ad-placement">
					<AdDisplay placement="article-sidebar" />
				</div>
			</aside>
		</div>
	{/if}
</main>

<Footer />

<style>
	.article-container {
		min-height: 100vh;
		background: var(--bg-primary);
		padding-top: 80px;
	}

	/* Loading and Error States */
	.loading-container,
	.error-container {
		display: flex;
		justify-content: center;
		align-items: center;
		min-height: 60vh;
		padding: 2rem;
	}

	.loading-spinner {
		display: flex;
		flex-direction: column;
		align-items: center;
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

	.error-content {
		text-align: center;
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 3rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.error-content h1 {
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.error-content p {
		color: var(--text-secondary);
		margin-bottom: 2rem;
	}

	.back-button {
		display: inline-block;
		background: var(--accent-primary);
		color: white;
		padding: 0.8rem 2rem;
		border-radius: 25px;
		text-decoration: none;
		font-weight: 600;
		transition: all 0.3s ease;
	}

	.back-button:hover {
		background: #e55a5a;
		transform: translateY(-2px);
	}

	/* Article Header */
	.article-header {
		background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
		padding: 3rem 2rem;
		position: relative;
		overflow: hidden;
	}

	.article-header::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="20" cy="20" r="1" fill="white" opacity="0.1"/><circle cx="80" cy="30" r="0.5" fill="white" opacity="0.1"/></svg>');
		opacity: 0.3;
	}

	.header-content {
		position: relative;
		z-index: 2;
		max-width: 1000px;
		margin: 0 auto;
	}

	.breadcrumb {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 2rem;
		font-size: 0.9rem;
		color: rgba(255, 255, 255, 0.8);
	}

	.breadcrumb a {
		color: var(--accent-primary);
		text-decoration: none;
		transition: color 0.3s ease;
	}

	.breadcrumb a:hover {
		color: #ff8a8a;
	}

	.breadcrumb span:last-child {
		color: rgba(255, 255, 255, 0.6);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		max-width: 200px;
	}

	.article-meta-header {
		display: flex;
		align-items: center;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
	}

	.category-badge {
		background: var(--accent-primary);
		color: white;
		padding: 0.5rem 1rem;
		border-radius: 20px;
		font-size: 0.9rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.article-tags-header {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.tag-badge {
		background: rgba(255, 255, 255, 0.2);
		color: white;
		padding: 0.3rem 0.8rem;
		border-radius: 15px;
		font-size: 0.8rem;
		font-weight: 500;
		border: 1px solid rgba(255, 255, 255, 0.3);
	}

	.article-title {
		font-size: 3rem;
		font-weight: 900;
		color: #ffd700;
		text-shadow: 4px 4px 8px rgba(0, 0, 0, 0.9);
		filter: drop-shadow(0 0 15px rgba(255, 215, 0, 0.4));
		margin-bottom: 1.5rem;
		line-height: 1.2;
	}

	.article-excerpt {
		font-size: 1.3rem;
		color: rgba(255, 255, 255, 0.95);
		text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.8);
		line-height: 1.6;
		margin-bottom: 2rem;
	}

	.article-info {
		display: grid;
		grid-template-columns: auto 1fr;
		gap: 2rem;
		align-items: center;
	}

	.author-info {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.author-avatar {
		width: 60px;
		height: 60px;
		border-radius: 50%;
		border: 3px solid rgba(255, 255, 255, 0.3);
		box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
	}

	.author-details h3 {
		color: white;
		font-size: 1.1rem;
		font-weight: 700;
		margin-bottom: 0.3rem;
	}

	.author-title,
	.author-institution {
		color: rgba(255, 255, 255, 0.8);
		font-size: 0.9rem;
		margin: 0;
	}

	.article-stats {
		display: flex;
		gap: 2rem;
		justify-self: end;
	}

	.stat-item {
		text-align: center;
		background: rgba(255, 255, 255, 0.1);
		backdrop-filter: blur(10px);
		padding: 0.8rem 1rem;
		border-radius: 15px;
		border: 1px solid rgba(255, 255, 255, 0.2);
	}

	.stat-label {
		display: block;
		color: rgba(255, 255, 255, 0.7);
		font-size: 0.8rem;
		margin-bottom: 0.3rem;
		text-transform: uppercase;
		letter-spacing: 0.5px;
	}

	.stat-value {
		display: block;
		color: white;
		font-weight: 700;
		font-size: 1rem;
	}

	.featured-image-container {
		margin-top: 3rem;
		border-radius: 20px;
		overflow: hidden;
		box-shadow: 0 20px 60px rgba(0, 0, 0, 0.4);
	}

	.featured-image {
		width: 100%;
		height: 400px;
		object-fit: cover;
		display: block;
	}

	/* Content Wrapper */
	.content-wrapper {
		display: grid;
		grid-template-columns: 1fr 350px;
		gap: 4rem;
		max-width: 1400px;
		margin: 0 auto;
		padding: 4rem 2rem;
	}

	/* Article Content */
	.article-content {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 3rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
	}

	.article-body {
		color: var(--text-primary);
		line-height: 1.8;
		font-size: 1.1rem;
		margin-bottom: 3rem;
	}

	.article-body :global(p) {
		margin-bottom: 1.5rem;
	}

	.article-body :global(h2) {
		font-size: 1.8rem;
		font-weight: 700;
		color: var(--text-primary);
		margin: 2.5rem 0 1.5rem 0;
		border-bottom: 2px solid var(--accent-primary);
		padding-bottom: 0.5rem;
	}

	.article-body :global(h3) {
		font-size: 1.5rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 2rem 0 1rem 0;
	}

	.article-body :global(ul),
	.article-body :global(ol) {
		margin-bottom: 1.5rem;
		padding-left: 2rem;
	}

	.article-body :global(li) {
		margin-bottom: 0.5rem;
		color: var(--text-secondary);
	}

	.article-body :global(blockquote) {
		background: rgba(255, 107, 107, 0.1);
		border-left: 4px solid var(--accent-primary);
		padding: 1.5rem;
		margin: 2rem 0;
		border-radius: 0 15px 15px 0;
		font-style: italic;
		color: var(--text-secondary);
	}

	.article-tags-section {
		margin-bottom: 3rem;
		padding-top: 2rem;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.article-tags-section h3 {
		color: var(--text-primary);
		font-size: 1.3rem;
		font-weight: 700;
		margin-bottom: 1rem;
	}

	.tags-list {
		display: flex;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.tag-link {
		background: rgba(255, 255, 255, 0.1);
		border: 1px solid rgba(255, 255, 255, 0.2);
		color: var(--text-secondary);
		padding: 0.5rem 1rem;
		border-radius: 20px;
		text-decoration: none;
		font-size: 0.9rem;
		transition: all 0.3s ease;
	}

	.tag-link:hover {
		background: var(--accent-primary);
		color: white;
		border-color: var(--accent-primary);
		transform: translateY(-2px);
	}

	/* Share Section */
	.share-section {
		margin-bottom: 3rem;
		padding-top: 2rem;
		border-top: 1px solid rgba(255, 255, 255, 0.1);
	}

	.share-section h3 {
		color: var(--text-primary);
		font-size: 1.3rem;
		font-weight: 700;
		margin-bottom: 1rem;
	}

	.share-buttons {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	.share-button {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.8rem 1.2rem;
		border: none;
		border-radius: 25px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.3s ease;
		text-decoration: none;
		font-size: 0.9rem;
	}

	.share-button svg {
		width: 18px;
		height: 18px;
	}

	.share-button.twitter {
		background: #1da1f2;
		color: white;
	}

	.share-button.facebook {
		background: #1877f2;
		color: white;
	}

	.share-button.linkedin {
		background: #0a66c2;
		color: white;
	}

	.share-button.email {
		background: #34495e;
		color: white;
	}

	.share-button.copy {
		background: var(--accent-primary);
		color: white;
	}

	.share-button:hover {
		transform: translateY(-2px);
		box-shadow: 0 5px 15px rgba(0, 0, 0, 0.3);
	}

	/* Sidebar */
	.article-sidebar {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.author-bio-card,
	.related-section {
		background: var(--bg-glass);
		backdrop-filter: blur(20px);
		border-radius: 20px;
		padding: 2rem;
		border: 1px solid rgba(255, 255, 255, 0.2);
		box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2);
	}

	.author-bio-card h3,
	.related-section h3 {
		color: var(--text-primary);
		font-size: 1.3rem;
		font-weight: 700;
		margin-bottom: 1.5rem;
		padding-bottom: 0.5rem;
		border-bottom: 2px solid var(--accent-primary);
	}

	.author-bio {
		text-align: center;
	}

	.author-avatar-large {
		width: 80px;
		height: 80px;
		border-radius: 50%;
		border: 3px solid var(--accent-primary);
		margin-bottom: 1rem;
		box-shadow: 0 4px 15px rgba(0, 0, 0, 0.3);
	}

	.author-bio-content h4 {
		color: var(--text-primary);
		font-size: 1.2rem;
		font-weight: 700;
		margin-bottom: 0.5rem;
	}

	.author-credentials {
		color: var(--accent-primary);
		font-size: 0.9rem;
		font-weight: 600;
		margin-bottom: 1rem;
		line-height: 1.4;
	}

	.author-description {
		color: var(--text-secondary);
		line-height: 1.6;
		margin-bottom: 1rem;
		font-size: 0.95rem;
	}

	.author-stats {
		color: var(--text-secondary);
		font-size: 0.9rem;
	}

	/* Related Articles */
	.related-articles {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.related-article {
		display: flex;
		gap: 1rem;
		padding: 1rem;
		background: rgba(255, 255, 255, 0.05);
		border-radius: 15px;
		transition: all 0.3s ease;
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.related-article:hover {
		background: rgba(255, 255, 255, 0.1);
		transform: translateY(-2px);
	}

	.related-image {
		width: 80px;
		height: 60px;
		object-fit: cover;
		border-radius: 10px;
		flex-shrink: 0;
	}

	.related-content {
		flex: 1;
	}

	.related-content h4 {
		font-size: 1rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
		line-height: 1.3;
	}

	.related-content h4 a {
		color: var(--text-primary);
		text-decoration: none;
		transition: color 0.3s ease;
	}

	.related-content h4 a:hover {
		color: var(--accent-primary);
	}

	.related-meta {
		color: var(--text-secondary);
		font-size: 0.8rem;
		margin: 0;
	}

	/* Ad Placements */
	.ad-placement {
		margin: 3rem 0;
		text-align: center;
		background: var(--bg-glass-dark);
		backdrop-filter: blur(20px);
		border-radius: 15px;
		padding: 1.5rem;
		border: 1px solid rgba(255, 255, 255, 0.1);
		box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
	}

	/* Responsive Design */
	@media (max-width: 1200px) {
		.content-wrapper {
			grid-template-columns: 1fr 300px;
			gap: 3rem;
		}
	}

	@media (max-width: 992px) {
		.content-wrapper {
			grid-template-columns: 1fr;
			gap: 2rem;
		}

		.article-sidebar {
			order: 2;
			display: grid;
			grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
			gap: 2rem;
		}

		.article-content {
			order: 1;
		}

		.article-info {
			grid-template-columns: 1fr;
			gap: 1.5rem;
		}

		.article-stats {
			justify-self: start;
		}
	}

	@media (max-width: 768px) {
		.article-header {
			padding: 2rem 1rem;
		}

		.article-title {
			font-size: 2.2rem;
		}

		.article-excerpt {
			font-size: 1.1rem;
		}

		.content-wrapper {
			padding: 2rem 1rem;
		}

		.article-content {
			padding: 2rem;
		}

		.author-bio-card,
		.related-section {
			padding: 1.5rem;
		}

		.article-stats {
			display: grid;
			grid-template-columns: repeat(2, 1fr);
			gap: 1rem;
		}

		.share-buttons {
			justify-content: center;
		}

		.breadcrumb span:last-child {
			max-width: 150px;
		}
	}

	@media (max-width: 480px) {
		.article-title {
			font-size: 1.8rem;
		}

		.article-content {
			padding: 1.5rem;
		}

		.author-bio-card,
		.related-section {
			padding: 1rem;
		}

		.article-info {
			gap: 1rem;
		}

		.author-info {
			flex-direction: column;
			text-align: center;
		}

		.article-stats {
			grid-template-columns: 1fr;
		}

		.share-buttons {
			flex-direction: column;
			align-items: center;
		}

		.share-button {
			width: 100%;
			max-width: 200px;
			justify-content: center;
		}
	}
</style> 