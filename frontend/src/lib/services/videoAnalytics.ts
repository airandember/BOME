/**
 * Video Analytics Service
 * 
 * Handles all video tracking and analytics for the BOME platform.
 * Tracks views, progress, completion for both authenticated and anonymous users.
 */

import { goto } from '$app/navigation';

// Types
export interface VideoTrackingEvent {
	video_id: number;
	watched_duration: number;
	watched_percentage: number;
	session_id?: string;
}

export interface VideoStats {
	video_id: number;
	total_views: number;
	unique_viewers: number;
	total_watch_time_seconds: number;
	avg_watch_time_seconds: number;
	avg_percentage_watched: number;
	completion_rate: number;
	bounce_rate: number;
	engagement_score: number;
	period: string;
	last_viewed_at: string;
}

export interface WatchHistory {
	id: number;
	user_id: number;
	video_id: number;
	last_position: number;
	completed: boolean;
	first_watched_at: string;
	last_watched_at: string;
	watch_percentage: number;
}

export interface ContinueWatchingVideo {
	video_id: number;
	title: string;
	thumbnail_url: string;
	duration: number;
	last_position: number;
	percentage: number;
	last_watched_at: string;
}

/**
 * Video Analytics Service Class
 * Singleton pattern for consistent tracking across the app
 */
export class VideoAnalyticsService {
	private sessionId: string;
	private lastReportedTime: Map<number, number> = new Map();
	private apiBaseUrl: string = '/api/v1';
	private trackingInterval: number = 10; // Report every 10 seconds
	
	constructor() {
		this.sessionId = this.getOrCreateSessionId();
		console.log('📊 [Video Analytics] Service initialized with session:', this.sessionId);
	}
	
	/**
	 * Track video progress
	 * Called on video player's timeupdate event
	 * Reports every 10 seconds to avoid spamming backend
	 */
	async trackProgress(videoId: number, currentTime: number, duration: number): Promise<void> {
		console.log(`🎬 [FRONTEND] trackProgress called: video=${videoId}, time=${currentTime}s, duration=${duration}s`);
		
		// Don't track if no duration yet
		if (!duration || duration <= 0) {
			console.log(`⏭️  [FRONTEND] Skipping - invalid duration: ${duration}`);
			return;
		}
		
		// Only report if enough time has passed since last report
		const currentSecond = Math.floor(currentTime);
		const lastReported = this.lastReportedTime.get(videoId) || -999; // Start with large negative
		const secondsSinceLastReport = currentSecond - lastReported;
		
		if (secondsSinceLastReport < this.trackingInterval) {
			console.log(`⏭️  [FRONTEND] Skipping - only ${secondsSinceLastReport}s since last report (need ${this.trackingInterval}s)`);
			return;
		}
		
		console.log(`✅ [FRONTEND] ${secondsSinceLastReport}s since last report - tracking now!`);
		this.lastReportedTime.set(videoId, currentSecond);
		
		const percentage = Math.min((currentTime / duration) * 100, 100);
		
		console.log(`📤 [FRONTEND] Sending tracking event: video=${videoId}, time=${currentSecond}s, %=${percentage.toFixed(1)}%`);
		
		try {
			await this.sendTrackingEvent({
				video_id: videoId,
				watched_duration: currentSecond,
				watched_percentage: Math.round(percentage * 100) / 100 // Round to 2 decimals
			});
			
			console.log(`✅ [FRONTEND] Successfully tracked: video=${videoId}, time=${currentSecond}s, %=${percentage.toFixed(1)}%`);
		} catch (error) {
			console.error('❌ [FRONTEND] Failed to track progress:', error);
			// Don't throw - tracking failures shouldn't break video playback
		}
	}
	
	/**
	 * Mark video as complete
	 * Called when video ends or user reaches 95%+
	 */
	async markComplete(videoId: number, duration: number): Promise<void> {
		console.log(`✅ [Video Analytics] Marking video ${videoId} as complete`);
		
		try {
			await this.sendTrackingEvent({
				video_id: videoId,
				watched_duration: Math.floor(duration),
				watched_percentage: 100
			});
			
			// Also call the dedicated complete endpoint if authenticated
			if (this.isAuthenticated()) {
				await fetch(`${this.apiBaseUrl}/videos/${videoId}/complete`, {
					method: 'POST',
					headers: this.getHeaders()
				});
			}
			
			console.log(`✅ [Video Analytics] Video ${videoId} marked complete`);
		} catch (error) {
			console.warn('⚠️ [Video Analytics] Failed to mark complete:', error);
		}
	}
	
	/**
	 * Get watch history for a specific video
	 * Returns resume position if available
	 */
	async getWatchHistory(videoId: number): Promise<WatchHistory | null> {
		if (!this.isAuthenticated()) {
			return null; // Watch history only for authenticated users
		}
		
		try {
			const response = await fetch(`${this.apiBaseUrl}/videos/${videoId}/watch-history`, {
				headers: this.getHeaders()
			});
			
			if (!response.ok) {
				if (response.status === 404) {
					return null; // No history found
				}
				throw new Error(`HTTP ${response.status}`);
			}
			
			const data = await response.json();
			return data;
		} catch (error) {
			console.error('❌ [Video Analytics] Failed to get watch history:', error);
			return null;
		}
	}
	
	/**
	 * Get "Continue Watching" list
	 * Returns videos user started but didn't finish
	 */
	async getContinueWatching(limit: number = 20): Promise<ContinueWatchingVideo[]> {
		if (!this.isAuthenticated()) {
			return [];
		}
		
		try {
			const response = await fetch(`${this.apiBaseUrl}/videos/continue-watching?limit=${limit}`, {
				headers: this.getHeaders()
			});
			
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}
			
			const data = await response.json();
			return data.videos || [];
		} catch (error) {
			console.error('❌ [Video Analytics] Failed to get continue watching:', error);
			return [];
		}
	}
	
	/**
	 * Get video statistics
	 * Requires authentication
	 */
	async getVideoStats(videoId: number, period: string = '7d'): Promise<VideoStats | null> {
		if (!this.isAuthenticated()) {
			return null;
		}
		
		try {
			const response = await fetch(`${this.apiBaseUrl}/analytics/video/${videoId}/stats?period=${period}`, {
				headers: this.getHeaders()
			});
			
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}
			
			return await response.json();
		} catch (error) {
			console.error('❌ [Video Analytics] Failed to get stats:', error);
			return null;
		}
	}
	
	/**
	 * Get trending videos
	 * Public endpoint, no auth required
	 */
	async getTrendingVideos(limit: number = 10): Promise<any[]> {
		try {
			const response = await fetch(`${this.apiBaseUrl}/analytics/trending?limit=${limit}`);
			
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}
			
			const data = await response.json();
			return data.trending || [];
		} catch (error) {
			console.error('❌ [Video Analytics] Failed to get trending:', error);
			return [];
		}
	}
	
	/**
	 * Get top (most watched) videos
	 * Public endpoint, no auth required
	 */
	async getTopVideos(limit: number = 25, days: number = 30): Promise<any[]> {
		try {
			const response = await fetch(`${this.apiBaseUrl}/analytics/top?limit=${limit}&days=${days}`);
			
			if (!response.ok) {
				throw new Error(`HTTP ${response.status}`);
			}
			
			const data = await response.json();
			return data.videos || [];
		} catch (error) {
			console.error('❌ [Video Analytics] Failed to get top videos:', error);
			return [];
		}
	}
	
	/**
	 * Clear watch history for a video
	 * Removes resume point
	 */
	async clearHistory(videoId: number): Promise<boolean> {
		if (!this.isAuthenticated()) {
			return false;
		}
		
		try {
			const response = await fetch(`${this.apiBaseUrl}/videos/${videoId}/watch-history`, {
				method: 'DELETE',
				headers: this.getHeaders()
			});
			
			return response.ok;
		} catch (error) {
			console.error('❌ [Video Analytics] Failed to clear history:', error);
			return false;
		}
	}
	
	// Private helper methods
	
	/**
	 * Send tracking event to backend
	 */
	private async sendTrackingEvent(event: VideoTrackingEvent): Promise<void> {
		console.log(`🌐 [FRONTEND→BACKEND] Preparing request to ${this.apiBaseUrl}/analytics/video/track`, event);
		
		const headers = this.getHeaders();
		
		// Add session ID for anonymous users
		if (!this.isAuthenticated()) {
			headers['X-Session-ID'] = this.sessionId;
			console.log(`🔑 [FRONTEND] Using session ID: ${this.sessionId}`);
		} else {
			console.log(`🔑 [FRONTEND] Using JWT authentication`);
		}
		
		const requestStart = performance.now();
		const response = await fetch(`${this.apiBaseUrl}/analytics/video/track`, {
			method: 'POST',
			headers: headers,
			body: JSON.stringify(event)
		});
		const requestDuration = performance.now() - requestStart;
		
		console.log(`📥 [FRONTEND←BACKEND] Response received in ${requestDuration.toFixed(2)}ms: ${response.status}`);
		
		if (!response.ok) {
			throw new Error(`HTTP ${response.status}: ${await response.text()}`);
		}
		
		// Check for throttled response (circuit breaker open)
		try {
			const data = await response.json();
			console.log(`📋 [FRONTEND] Response data:`, data);
			
			if (data.status === 'throttled') {
				console.warn('⚠️ [FRONTEND] Analytics temporarily throttled:', data.message);
				// Still treat as success - don't break video playback
				// Circuit breaker will recover automatically
			} else if (data.status === 'tracked') {
				console.log(`✅ [FRONTEND] Backend confirmed tracking for video ${data.video_id}`);
			}
		} catch (e) {
			// If response is not JSON or parsing fails, that's fine
			// Backend might return empty body on success
			console.log(`ℹ️  [FRONTEND] Non-JSON response (likely empty body - OK)`);
		}
	}
	
	/**
	 * Get or create session ID for anonymous tracking
	 */
	private getOrCreateSessionId(): string {
		// Check if we're in a browser environment
		if (typeof window === 'undefined' || typeof sessionStorage === 'undefined') {
			// SSR environment - return a temporary session ID
			return `sess_ssr_${Date.now()}`;
		}
		
		const storageKey = 'video_session_id';
		
		// Try to get existing session ID
		let sessionId = sessionStorage.getItem(storageKey);
		
		if (!sessionId) {
			// Generate new session ID
			sessionId = `sess_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
			sessionStorage.setItem(storageKey, sessionId);
			console.log('📊 [Video Analytics] Created new session:', sessionId);
		}
		
		return sessionId;
	}
	
	/**
	 * Check if user is authenticated
	 */
	private isAuthenticated(): boolean {
		return !!this.getAuthToken();
	}
	
	/**
	 * Get authentication token
	 */
	private getAuthToken(): string | null {
		// Check if we're in a browser environment
		if (typeof window === 'undefined' || typeof localStorage === 'undefined') {
			return null; // SSR - no auth token available
		}
		
		// Get tokens from BOME auth storage (matches auth.ts structure)
		try {
			const authDataStr = localStorage.getItem('bome_auth_data');
			if (authDataStr) {
				const authData = JSON.parse(authDataStr);
				return authData.access_token || null;
			}
		} catch (error) {
			console.error('Failed to parse auth data:', error);
		}
		
		return null;
	}
	
	/**
	 * Get HTTP headers for API requests
	 */
	private getHeaders(): Record<string, string> {
		const headers: Record<string, string> = {
			'Content-Type': 'application/json'
		};
		
		const token = this.getAuthToken();
		if (token) {
			headers['Authorization'] = `Bearer ${token}`;
		}
		
		return headers;
	}
	
	/**
	 * Reset tracking state for a video
	 * Useful when video is restarted
	 */
	resetTracking(videoId: number): void {
		this.lastReportedTime.delete(videoId);
		console.log(`🔄 [Video Analytics] Reset tracking for video ${videoId}`);
	}
}

// Export singleton instance
export const videoAnalytics = new VideoAnalyticsService();

// Export for testing/mocking
export default videoAnalytics;

