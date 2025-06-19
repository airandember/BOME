// Bunny.net Video Streaming Integration
export interface BunnyConfig {
	libraryId: string;
	apiKey: string;
	cdnUrl: string;
	storageZone: string;
	pullZone: string;
}

export interface VideoUploadOptions {
	title: string;
	description?: string;
	tags?: string[];
	chapters?: Array<{
		title: string;
		start: number;
		end: number;
	}>;
	thumbnailTime?: number;
}

export interface VideoMetadata {
	guid: string;
	title: string;
	description?: string;
	duration: number;
	size: number;
	status: 'uploading' | 'processing' | 'ready' | 'failed';
	thumbnailUrl?: string;
	videoUrl?: string;
	createdAt: string;
	updatedAt: string;
}

export interface StreamingQuality {
	height: number;
	width: number;
	bitrate: number;
	fps: number;
	url: string;
}

class BunnyNetService {
	private config: BunnyConfig;

	constructor() {
		this.config = {
			libraryId: import.meta.env.VITE_BUNNY_LIBRARY_ID || '',
			apiKey: import.meta.env.VITE_BUNNY_API_KEY || '',
			cdnUrl: import.meta.env.VITE_BUNNY_CDN_URL || '',
			storageZone: import.meta.env.VITE_BUNNY_STORAGE_ZONE || '',
			pullZone: import.meta.env.VITE_BUNNY_PULL_ZONE || ''
		};
	}

	private getHeaders(): Record<string, string> {
		return {
			'AccessKey': this.config.apiKey,
			'Content-Type': 'application/json'
		};
	}

	async uploadVideo(file: File, options: VideoUploadOptions, onProgress?: (progress: number) => void): Promise<VideoMetadata> {
		try {
			// Step 1: Create video entry
			const createResponse = await fetch(`https://video.bunnycdn.com/library/${this.config.libraryId}/videos`, {
				method: 'POST',
				headers: this.getHeaders(),
				body: JSON.stringify({
					title: options.title,
					description: options.description,
					tags: options.tags?.join(',')
				})
			});

			if (!createResponse.ok) {
				throw new Error('Failed to create video entry');
			}

			const videoData = await createResponse.json();
			const videoGuid = videoData.guid;

			// Step 2: Upload video file
			const uploadUrl = `https://video.bunnycdn.com/library/${this.config.libraryId}/videos/${videoGuid}`;
			
			return new Promise((resolve, reject) => {
				const xhr = new XMLHttpRequest();

				xhr.upload.addEventListener('progress', (e) => {
					if (e.lengthComputable && onProgress) {
						const progress = (e.loaded / e.total) * 100;
						onProgress(progress);
					}
				});

				xhr.addEventListener('load', async () => {
					if (xhr.status >= 200 && xhr.status < 300) {
						try {
							// Step 3: Set thumbnail time if specified
							if (options.thumbnailTime) {
								await this.setThumbnailTime(videoGuid, options.thumbnailTime);
							}

							// Step 4: Add chapters if specified
							if (options.chapters && options.chapters.length > 0) {
								await this.addChapters(videoGuid, options.chapters);
							}

							const metadata = await this.getVideoMetadata(videoGuid);
							resolve(metadata);
						} catch (error) {
							reject(error);
						}
					} else {
						reject(new Error('Video upload failed'));
					}
				});

				xhr.addEventListener('error', () => {
					reject(new Error('Video upload failed'));
				});

				xhr.open('PUT', uploadUrl);
				xhr.setRequestHeader('AccessKey', this.config.apiKey);
				xhr.send(file);
			});

		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Video upload failed: ${errorMessage}`);
		}
	}

	async getVideoMetadata(videoGuid: string): Promise<VideoMetadata> {
		try {
			const response = await fetch(`https://video.bunnycdn.com/library/${this.config.libraryId}/videos/${videoGuid}`, {
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to fetch video metadata');
			}

			const data = await response.json();
			
			return {
				guid: data.guid,
				title: data.title,
				description: data.description,
				duration: data.length,
				size: data.storageSize,
				status: this.mapStatus(data.status),
				thumbnailUrl: data.thumbnailFileName ? `${this.config.cdnUrl}/${videoGuid}/${data.thumbnailFileName}` : undefined,
				videoUrl: `${this.config.cdnUrl}/${videoGuid}/playlist.m3u8`,
				createdAt: data.dateUploaded,
				updatedAt: data.dateUploaded
			};
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get video metadata: ${errorMessage}`);
		}
	}

	async getStreamingQualities(videoGuid: string): Promise<StreamingQuality[]> {
		try {
			const metadata = await this.getVideoMetadata(videoGuid);
			
			// Mock quality options - in real implementation, these would come from Bunny.net
			const qualities: StreamingQuality[] = [
				{
					height: 1080,
					width: 1920,
					bitrate: 5000,
					fps: 30,
					url: `${this.config.cdnUrl}/${videoGuid}/1080p.mp4`
				},
				{
					height: 720,
					width: 1280,
					bitrate: 2500,
					fps: 30,
					url: `${this.config.cdnUrl}/${videoGuid}/720p.mp4`
				},
				{
					height: 480,
					width: 854,
					bitrate: 1000,
					fps: 30,
					url: `${this.config.cdnUrl}/${videoGuid}/480p.mp4`
				},
				{
					height: 360,
					width: 640,
					bitrate: 500,
					fps: 30,
					url: `${this.config.cdnUrl}/${videoGuid}/360p.mp4`
				}
			];

			return qualities;
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get streaming qualities: ${errorMessage}`);
		}
	}

	async deleteVideo(videoGuid: string): Promise<void> {
		try {
			const response = await fetch(`https://video.bunnycdn.com/library/${this.config.libraryId}/videos/${videoGuid}`, {
				method: 'DELETE',
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to delete video');
			}
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to delete video: ${errorMessage}`);
		}
	}

	async setThumbnailTime(videoGuid: string, timeInSeconds: number): Promise<void> {
		try {
			const response = await fetch(`https://video.bunnycdn.com/library/${this.config.libraryId}/videos/${videoGuid}/thumbnail`, {
				method: 'POST',
				headers: this.getHeaders(),
				body: JSON.stringify({
					thumbnailTime: timeInSeconds
				})
			});

			if (!response.ok) {
				throw new Error('Failed to set thumbnail time');
			}
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to set thumbnail time: ${errorMessage}`);
		}
	}

	async addChapters(videoGuid: string, chapters: Array<{ title: string; start: number; end: number }>): Promise<void> {
		try {
			const response = await fetch(`https://video.bunnycdn.com/library/${this.config.libraryId}/videos/${videoGuid}/chapters`, {
				method: 'POST',
				headers: this.getHeaders(),
				body: JSON.stringify({
					chapters: chapters
				})
			});

			if (!response.ok) {
				throw new Error('Failed to add chapters');
			}
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to add chapters: ${errorMessage}`);
		}
	}

	async getVideoStats(videoGuid: string, dateFrom?: string, dateTo?: string): Promise<any> {
		try {
			const params = new URLSearchParams();
			if (dateFrom) params.append('dateFrom', dateFrom);
			if (dateTo) params.append('dateTo', dateTo);

			const response = await fetch(`https://video.bunnycdn.com/library/${this.config.libraryId}/videos/${videoGuid}/statistics?${params}`, {
				headers: this.getHeaders()
			});

			if (!response.ok) {
				throw new Error('Failed to get video statistics');
			}

			return await response.json();
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : 'Unknown error';
			throw new Error(`Failed to get video statistics: ${errorMessage}`);
		}
	}

	private mapStatus(bunnyStatus: number): VideoMetadata['status'] {
		switch (bunnyStatus) {
			case 0:
			case 1:
				return 'uploading';
			case 2:
			case 3:
				return 'processing';
			case 4:
				return 'ready';
			case 5:
			case 6:
				return 'failed';
			default:
				return 'processing';
		}
	}

	// Generate signed URLs for secure video access
	generateSignedUrl(videoGuid: string, expirationTime: number = 3600): string {
		const expires = Math.floor(Date.now() / 1000) + expirationTime;
		const token = this.generateSecurityToken(videoGuid, expires);
		
		return `${this.config.cdnUrl}/${videoGuid}/playlist.m3u8?token=${token}&expires=${expires}`;
	}

	private generateSecurityToken(videoGuid: string, expires: number): string {
		// Mock token generation - in real implementation, use proper signing
		const data = `${videoGuid}-${expires}`;
		return btoa(data);
	}
}

export const bunnyService = new BunnyNetService(); 