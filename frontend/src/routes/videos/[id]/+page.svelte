<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { videoService, type Video, type VideoComment } from '$lib/video';
	import VideoPlayer from '$lib/components/VideoPlayer.svelte';
	import VideoCard from '$lib/components/VideoCard.svelte';
	import { auth } from '$lib/auth';

	let video: Video | null = null;
	let relatedVideos: Video[] = [];
	let comments: VideoComment[] = [];
	let loading = true;
	let error = '';
	let newComment = '';
	let submittingComment = false;
	let isLiked = false;
	let isFavorited = false;
	let user: any = null;
	let isAuthenticated = false;

	$: videoId = $page.params.id;

	onMount(async () => {
		auth.subscribe((state) => {
			user = state.user;
			isAuthenticated = state.isAuthenticated;
		});

		await loadVideo();
	});

	async function loadVideo() {
		try {
			loading = true;
			error = '';

			const videoResponse = await videoService.getVideo(parseInt(videoId));
			video = videoResponse.video;

			// Load related videos
			if (video) {
				const relatedResponse = await videoService.getVideos(1, 6, video.category);
				relatedVideos = (relatedResponse.videos || []).filter((v: Video) => v.id !== video!.id).slice(0, 5);
			}

			// Load comments
			await loadComments();

		} catch (err) {
			error = 'Failed to load video';
			console.error('Error loading video:', err);
		} finally {
			loading = false;
		}
	}

	async function loadComments() {
		try {
			const response = await videoService.getComments(parseInt(videoId));
			comments = response.comments || [];
		} catch (err) {
			console.error('Error loading comments:', err);
		}
	}

	async function handleLike() {
		if (!isAuthenticated) {
			// Redirect to login
			window.location.href = '/login';
			return;
		}

		if (!video) return;

		try {
			if (isLiked) {
				await videoService.unlikeVideo(parseInt(videoId));
				video.likeCount--;
			} else {
				await videoService.likeVideo(parseInt(videoId));
				video.likeCount++;
			}
			isLiked = !isLiked;
		} catch (err) {
			console.error('Error toggling like:', err);
		}
	}

	async function handleFavorite() {
		if (!isAuthenticated) {
			window.location.href = '/login';
			return;
		}

		try {
			if (isFavorited) {
				await videoService.unfavoriteVideo(parseInt(videoId));
			} else {
				await videoService.favoriteVideo(parseInt(videoId));
			}
			isFavorited = !isFavorited;
		} catch (err) {
			console.error('Error toggling favorite:', err);
		}
	}

	async function handleComment() {
		if (!isAuthenticated) {
			window.location.href = '/login';
			return;
		}

		if (!newComment.trim()) return;

		try {
			submittingComment = true;
			await videoService.addComment(parseInt(videoId), newComment);
			newComment = '';
			await loadComments();
		} catch (err) {
			console.error('Error adding comment:', err);
		} finally {
			submittingComment = false;
		}
	}

	function formatDate(dateString: string) {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}
</script>

<svelte:head>
	<title>{video?.title || 'Video'} - Book of Mormon Evidences</title>
	<meta name="description" content={video?.description || 'Watch Book of Mormon evidence videos'} />
</svelte:head>

<div class="video-page">
	<div class="container">
		{#if loading}
			<div class="loading">
				<div class="loading-spinner"></div>
				<p>Loading video...</p>
			</div>
		{:else if error}
			<div class="error-message">
				{error}
			</div>
		{:else if video}
			<div class="video-content">
				<div class="video-main">
					<div class="video-player-container">
						<VideoPlayer 
							videoUrl={video.videoUrl}
							poster={video.thumbnailUrl}
							width="100%"
							height="600px"
						/>
					</div>

					<div class="video-info">
						<h1 class="video-title">{video.title}</h1>
						
						<div class="video-stats">
							<span class="stat">
								👁️ {video.viewCount.toLocaleString()} views
							</span>
							<span class="stat">
								❤️ {video.likeCount.toLocaleString()} likes
							</span>
							<span class="stat">
								📅 {formatDate(video.createdAt)}
							</span>
						</div>

						<div class="video-actions">
							<button 
								class="action-btn {isLiked ? 'liked' : ''}"
								on:click={handleLike}
							>
								{isLiked ? '❤️' : '🤍'} {isLiked ? 'Liked' : 'Like'}
							</button>
							<button 
								class="action-btn {isFavorited ? 'favorited' : ''}"
								on:click={handleFavorite}
							>
								{isFavorited ? '⭐' : '☆'} {isFavorited ? 'Favorited' : 'Favorite'}
							</button>
							<button class="action-btn">
								📤 Share
							</button>
						</div>

						<div class="video-description">
							<h3>Description</h3>
							<p>{video.description}</p>
						</div>

						{#if video.tags && video.tags.length > 0}
							<div class="video-tags">
								<h3>Tags</h3>
								<div class="tags">
									{#each video.tags as tag}
										<span class="tag">{tag}</span>
									{/each}
								</div>
							</div>
						{/if}
					</div>
				</div>

				<div class="video-sidebar">
					<div class="comments-section">
						<h3>Comments ({comments.length})</h3>
						
						{#if isAuthenticated}
							<div class="comment-form">
								<textarea
									bind:value={newComment}
									placeholder="Add a comment..."
									rows="3"
								></textarea>
								<button 
									class="btn-primary"
									on:click={handleComment}
									disabled={submittingComment || !newComment.trim()}
								>
									{submittingComment ? 'Posting...' : 'Post Comment'}
								</button>
							</div>
						{:else}
							<div class="login-prompt">
								<p>Please <a href="/login">login</a> to add a comment.</p>
							</div>
						{/if}

						<div class="comments-list">
							{#each comments as comment (comment.id)}
								<div class="comment">
									<div class="comment-header">
										<span class="comment-author">{comment.userName}</span>
										<span class="comment-date">{formatDate(comment.createdAt)}</span>
									</div>
									<div class="comment-content">
										{comment.content}
									</div>
								</div>
							{/each}
						</div>
					</div>

					<div class="related-videos">
						<h3>Related Videos</h3>
						<div class="related-grid">
							{#each relatedVideos as relatedVideo (relatedVideo.id)}
								<VideoCard 
									video={relatedVideo} 
									showCategory={false}
									showStats={false}
								/>
							{/each}
						</div>
					</div>
				</div>
			</div>
		{/if}
	</div>
</div>

<style>
	.video-page {
		min-height: 100vh;
		background: var(--bg-color);
		padding: 2rem 0;
	}

	.container {
		max-width: 1400px;
		margin: 0 auto;
		padding: 0 2rem;
	}

	.loading {
		text-align: center;
		padding: 4rem 0;
	}

	.loading-spinner {
		width: 40px;
		height: 40px;
		border: 4px solid var(--border-color);
		border-top: 4px solid var(--accent-color);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin: 0 auto 1rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.loading p {
		color: var(--text-secondary);
		font-size: 1.1rem;
	}

	.error-message {
		background: var(--error-bg);
		color: var(--error-text);
		padding: 1rem;
		border-radius: 12px;
		text-align: center;
		margin-bottom: 2rem;
	}

	.video-content {
		display: grid;
		grid-template-columns: 2fr 1fr;
		gap: 2rem;
	}

	.video-main {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.video-player-container {
		background: var(--card-bg);
		border-radius: 16px;
		overflow: hidden;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.video-info {
		background: var(--card-bg);
		padding: 2rem;
		border-radius: 16px;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.video-title {
		font-size: 1.8rem;
		font-weight: 700;
		color: var(--text-primary);
		margin-bottom: 1rem;
		line-height: 1.3;
	}

	.video-stats {
		display: flex;
		gap: 1.5rem;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
	}

	.stat {
		font-size: 0.9rem;
		color: var(--text-secondary);
		display: flex;
		align-items: center;
		gap: 0.25rem;
	}

	.video-actions {
		display: flex;
		gap: 1rem;
		margin-bottom: 2rem;
		flex-wrap: wrap;
	}

	.action-btn {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		background: var(--card-bg);
		color: var(--text-primary);
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
		display: flex;
		align-items: center;
		gap: 0.5rem;
	}

	.action-btn:hover {
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	.action-btn.liked {
		background: var(--accent-color);
		color: white;
	}

	.action-btn.favorited {
		background: var(--accent-color);
		color: white;
	}

	.video-description {
		margin-bottom: 2rem;
	}

	.video-description h3 {
		font-size: 1.2rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.video-description p {
		color: var(--text-secondary);
		line-height: 1.6;
	}

	.video-tags h3 {
		font-size: 1.2rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.5rem;
	}

	.tags {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.tag {
		background: var(--accent-color);
		color: white;
		padding: 0.25rem 0.75rem;
		border-radius: 12px;
		font-size: 0.8rem;
		font-weight: 500;
	}

	.video-sidebar {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	.comments-section,
	.related-videos {
		background: var(--card-bg);
		padding: 2rem;
		border-radius: 16px;
		box-shadow: 
			8px 8px 16px var(--shadow-dark),
			-8px -8px 16px var(--shadow-light);
	}

	.comments-section h3,
	.related-videos h3 {
		font-size: 1.3rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1.5rem;
	}

	.comment-form {
		margin-bottom: 2rem;
	}

	.comment-form textarea {
		width: 100%;
		padding: 1rem;
		border: none;
		border-radius: 12px;
		background: var(--input-bg);
		color: var(--text-primary);
		font-size: 1rem;
		resize: vertical;
		margin-bottom: 1rem;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light);
	}

	.comment-form textarea:focus {
		outline: none;
		box-shadow: 
			inset 2px 2px 4px var(--shadow-dark),
			inset -2px -2px 4px var(--shadow-light),
			0 0 0 2px var(--accent-color);
	}

	.login-prompt {
		text-align: center;
		padding: 1rem;
		background: var(--input-bg);
		border-radius: 12px;
		margin-bottom: 2rem;
	}

	.login-prompt a {
		color: var(--accent-color);
		text-decoration: none;
		font-weight: 600;
	}

	.comments-list {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.comment {
		padding: 1rem;
		background: var(--input-bg);
		border-radius: 12px;
	}

	.comment-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 0.5rem;
	}

	.comment-author {
		font-weight: 600;
		color: var(--text-primary);
	}

	.comment-date {
		font-size: 0.8rem;
		color: var(--text-secondary);
	}

	.comment-content {
		color: var(--text-secondary);
		line-height: 1.5;
	}

	.related-grid {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	.btn-primary {
		padding: 0.75rem 1.5rem;
		border: none;
		border-radius: 12px;
		background: var(--accent-color);
		color: white;
		font-size: 1rem;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s ease;
		box-shadow: 
			4px 4px 8px var(--shadow-dark),
			-2px -2px 4px var(--shadow-light);
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--accent-hover);
		transform: translateY(-2px);
		box-shadow: 
			6px 6px 12px var(--shadow-dark),
			-3px -3px 6px var(--shadow-light);
	}

	.btn-primary:disabled {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none;
	}

	@media (max-width: 1024px) {
		.video-content {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 768px) {
		.container {
			padding: 0 1rem;
		}

		.video-title {
			font-size: 1.5rem;
		}

		.video-stats {
			flex-direction: column;
			gap: 0.5rem;
		}

		.video-actions {
			flex-direction: column;
		}

		.video-info,
		.comments-section,
		.related-videos {
			padding: 1.5rem;
		}
	}
</style> 