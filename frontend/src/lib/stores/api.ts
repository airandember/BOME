// Enhanced API-integrated Svelte Stores
import { writable, derived, get } from 'svelte/store';
import { apiClient, type ApiResponse } from '$lib/api/client';
import { toastStore } from '$lib/stores/toast';
import { browser } from '$app/environment';
import { auth } from '$lib/auth';

// Optimistic update utility
function createOptimisticStore<T>(initialState: T) {
	const { subscribe, set, update } = writable<T>(initialState);
	
	return {
		subscribe,
		set,
		update,
		
		// Optimistic update with rollback on failure
		async optimisticUpdate<R>(
			updateFn: (state: T) => T,
			apiCall: () => Promise<ApiResponse<R>>,
			successCallback?: (result: R, state: T) => T,
			errorCallback?: (error: string, state: T) => T
		): Promise<boolean> {
			const originalState = get({ subscribe });
			const optimisticState = updateFn(originalState);
			
			// Apply optimistic update immediately
			set(optimisticState);
			
			try {
				const result = await apiCall();
				
				if (result.error) {
					// Rollback on error
					const finalState = errorCallback ? errorCallback(result.error, originalState) : originalState;
					set(finalState);
					return false;
				}
				
				// Apply success state if provided
				if (successCallback && result.data) {
					const successState = successCallback(result.data, optimisticState);
					set(successState);
				}
				
				return true;
			} catch (error) {
				// Rollback on exception
				const errorMsg = error instanceof Error ? error.message : 'Unknown error';
				const finalState = errorCallback ? errorCallback(errorMsg, originalState) : originalState;
				set(finalState);
				return false;
			}
		}
	};
}

// Authentication store with API integration
interface AuthState {
	isAuthenticated: boolean;
	user: {
		id: number;
		email: string;
		role: string;
		full_name: string;
	} | null;
	token: string | null;
	loading: boolean;
	error: string | null;
}

function createAuthStore() {
	const store = createOptimisticStore<AuthState>({
		isAuthenticated: false,
		user: null,
		token: null,
		loading: false,
		error: null
	});

	// Handle token expiration events
	if (browser) {
		window.addEventListener('auth:token-expired', () => {
			console.log('Token expired, logging out user');
			authStore.logout();
			toastStore.warning('Your session has expired. Please log in again.', {
				title: 'Session Expired',
				persistent: true
			});
		});
	}

	return {
		...store,
		
		// Initialize auth state on app start
		async init() {
			// Main auth system handles initialization
			// This is just a placeholder for compatibility
		},

		// Login with main auth system
		async login(email: string, password: string) {
			return store.optimisticUpdate(
				// Optimistic update: show loading state
				(state) => ({ ...state, loading: true, error: null }),
				// API call - use main auth system
				async () => {
					const result = await auth.login(email, password);
					if (result.success && result.user) {
						return {
							user: {
								id: result.user.id,
								email: result.user.email,
								role: result.user.role,
								full_name: `${result.user.first_name} ${result.user.last_name}`
							},
							token: 'handled-by-main-auth'
						};
					}
					throw new Error(result.error || 'Login failed');
				},
				// Success callback
				(result: any, state) => {
					toastStore.success(`Welcome back, ${result.user.full_name}!`, { 
						title: 'Login Successful' 
					});
					return {
						...state,
						isAuthenticated: true,
						user: result.user,
						token: result.token,
						loading: false,
						error: null
					};
				},
				// Error callback
				(error, state) => {
					toastStore.error(error, { title: 'Login Failed' });
					return {
						...state,
						loading: false,
						error
					};
				}
			);
		},

		// Logout with main auth system
		async logout() {
			await auth.logout();
			store.update(state => ({
				isAuthenticated: false,
				user: null,
				token: null,
				loading: false,
				error: null
			}));
		},

		// Clear error
		clearError() {
			store.update(state => ({ ...state, error: null }));
		}
	};
}

// Video store with API integration and optimistic updates
interface VideoState {
	videos: any[];
	categories: any[];
	currentVideo: any | null;
	loading: boolean;
	error: string | null;
	pagination: {
		current_page: number;
		per_page: number;
		total: number;
		total_pages: number;
	} | null;
}

function createVideoStore() {
	const store = createOptimisticStore<VideoState>({
		videos: [],
		categories: [],
		currentVideo: null,
		loading: false,
		error: null,
		pagination: null
	});

	return {
		...store,

		async loadVideos(params?: {
			page?: number;
			limit?: number;
			category?: string;
			search?: string;
		}) {
			return store.optimisticUpdate(
				// Optimistic update: show loading
				(state) => ({ ...state, loading: true, error: null }),
				// API call
				() => apiClient.getVideos(params),
				// Success callback
				(result, state) => ({
					...state,
					videos: result.data,
					pagination: result.pagination,
					loading: false,
					error: null
				}),
				// Error callback
				(error, state) => {
					toastStore.error(error, { title: 'Video Loading Failed' });
					return {
						...state,
						loading: false,
						error
					};
				}
			);
		},

		async loadCategories() {
			return store.optimisticUpdate(
				// No optimistic state change needed for categories
				(state) => state,
				// API call
				() => apiClient.getVideoCategories(),
				// Success callback
				(result, state) => ({
					...state,
					categories: result.categories
				}),
				// Error callback
				(error, state) => {
					toastStore.error(error, { title: 'Categories Loading Failed' });
					return state;
				}
			);
		},

		async loadVideo(id: number) {
			return store.optimisticUpdate(
				// Optimistic update: show loading
				(state) => ({ ...state, loading: true, error: null }),
				// API call
				() => apiClient.getVideo(id),
				// Success callback
				(result, state) => ({
					...state,
					currentVideo: result.video,
					loading: false,
					error: null
				}),
				// Error callback
				(error, state) => {
					toastStore.error(error, { title: 'Video Loading Failed' });
					return {
						...state,
						loading: false,
						error
					};
				}
			);
		},

		// Optimistic video interaction (like/favorite)
		async toggleVideoLike(videoId: number) {
			return store.optimisticUpdate(
				// Optimistic update: toggle like immediately
				(state) => ({
					...state,
					videos: state.videos.map(video => 
						video.id === videoId 
							? { ...video, isLiked: !video.isLiked, likes: video.likes + (video.isLiked ? -1 : 1) }
							: video
					),
					currentVideo: state.currentVideo?.id === videoId 
						? { 
							...state.currentVideo, 
							isLiked: !state.currentVideo.isLiked,
							likes: state.currentVideo.likes + (state.currentVideo.isLiked ? -1 : 1)
						}
						: state.currentVideo
				}),
				// API call (mock for now)
				() => Promise.resolve({ data: { success: true } }),
				// Success callback
				(result, state) => {
					toastStore.success('Video preference updated!');
					return state;
				},
				// Error callback - rollback the optimistic update
				(error, originalState) => {
					toastStore.error('Failed to update video preference');
					return originalState;
				}
			);
		},

		clearError() {
			store.update(state => ({ ...state, error: null }));
		}
	};
}

// Admin store with API integration and optimistic updates
interface AdminState {
	analytics: any | null;
	users: any[];
	videos: any[];
	loading: boolean;
	error: string | null;
}

function createAdminStore() {
	const store = createOptimisticStore<AdminState>({
		analytics: null,
		users: [],
		videos: [],
		loading: false,
		error: null
	});

	return {
		...store,

		async loadAnalytics() {
			return store.optimisticUpdate(
				(state) => ({ ...state, loading: true, error: null }),
				() => apiClient.getAdminAnalytics(),
				(result, state) => ({
					...state,
					analytics: result.analytics,
					loading: false,
					error: null
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Analytics Loading Failed' });
					return { ...state, loading: false, error };
				}
			);
		},

		async loadUsers() {
			return store.optimisticUpdate(
				(state) => state,
				() => apiClient.getAdminUsers(),
				(result, state) => ({
					...state,
					users: result.users
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Users Loading Failed' });
					return state;
				}
			);
		},

		async loadVideos() {
			return store.optimisticUpdate(
				(state) => state,
				() => apiClient.getAdminVideos(),
				(result, state) => ({
					...state,
					videos: result
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Admin Videos Loading Failed' });
					return state;
				}
			);
		},

		// Optimistic user status update
		async updateUserStatus(userId: number, newStatus: string) {
			return store.optimisticUpdate(
				// Optimistic update: change user status immediately
				(state) => ({
					...state,
					users: state.users.map(user => 
						user.id === userId ? { ...user, status: newStatus } : user
					)
				}),
				// API call (mock for now)
				() => Promise.resolve({ data: { success: true } }),
				// Success callback
				(result, state) => {
					toastStore.success(`User status updated to ${newStatus}`);
					return state;
				},
				// Error callback
				(error, originalState) => {
					toastStore.error('Failed to update user status');
					return originalState;
				}
			);
		},

		clearError() {
			store.update(state => ({ ...state, error: null }));
		}
	};
}

// Advertisement store with API integration and optimistic updates
interface AdvertisementState {
	advertisers: any[];
	campaigns: any[];
	currentAdvertiser: any | null;
	currentCampaign: any | null;
	loading: boolean;
	error: string | null;
}

function createAdvertisementStore() {
	const store = createOptimisticStore<AdvertisementState>({
		advertisers: [],
		campaigns: [],
		currentAdvertiser: null,
		currentCampaign: null,
		loading: false,
		error: null
	});

	return {
		...store,

		async loadAdvertisers(status?: string) {
			return store.optimisticUpdate(
				(state) => ({ ...state, loading: true, error: null }),
				() => apiClient.getAdvertisers(status),
				(result, state) => ({
					...state,
					advertisers: result.advertisers,
					loading: false,
					error: null
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Advertisers Loading Failed' });
					return { ...state, loading: false, error };
				}
			);
		},

		async loadAdvertiser(id: number) {
			return store.optimisticUpdate(
				(state) => state,
				() => apiClient.getAdvertiser(id),
				(result, state) => ({
					...state,
					currentAdvertiser: result.advertiser
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Advertiser Loading Failed' });
					return state;
				}
			);
		},

		async loadCampaigns(params?: {
			advertiser_id?: number;
			status?: string;
		}) {
			return store.optimisticUpdate(
				(state) => state,
				() => apiClient.getCampaigns(params),
				(result, state) => ({
					...state,
					campaigns: result.campaigns
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Campaigns Loading Failed' });
					return state;
				}
			);
		},

		async loadCampaign(id: number) {
			return store.optimisticUpdate(
				(state) => state,
				() => apiClient.getCampaign(id),
				(result, state) => ({
					...state,
					currentCampaign: result.campaign
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Campaign Loading Failed' });
					return state;
				}
			);
		},

		// Optimistic campaign status update
		async updateCampaignStatus(campaignId: number, newStatus: string) {
			return store.optimisticUpdate(
				// Optimistic update: change campaign status immediately
				(state) => ({
					...state,
					campaigns: state.campaigns.map(campaign => 
						campaign.id === campaignId ? { ...campaign, status: newStatus } : campaign
					),
					currentCampaign: state.currentCampaign?.id === campaignId 
						? { ...state.currentCampaign, status: newStatus }
						: state.currentCampaign
				}),
				// API call (mock for now)
				() => Promise.resolve({ data: { success: true } }),
				// Success callback
				(result, state) => {
					toastStore.success(`Campaign status updated to ${newStatus}`);
					return state;
				},
				// Error callback
				(error, originalState) => {
					toastStore.error('Failed to update campaign status');
					return originalState;
				}
			);
		},

		clearError() {
			store.update(state => ({ ...state, error: null }));
		}
	};
}

// Dashboard store with API integration
interface DashboardState {
	data: any | null;
	loading: boolean;
	error: string | null;
}

function createDashboardStore() {
	const store = createOptimisticStore<DashboardState>({
		data: null,
		loading: false,
		error: null
	});

	return {
		...store,

		async loadDashboard() {
			return store.optimisticUpdate(
				(state) => ({ ...state, loading: true, error: null }),
				() => apiClient.getDashboard(),
				(result, state) => ({
					...state,
					data: result.data,
					loading: false,
					error: null
				}),
				(error, state) => {
					toastStore.error(error, { title: 'Dashboard Loading Failed' });
					return { ...state, loading: false, error };
				}
			);
		},

		clearError() {
			store.update(state => ({ ...state, error: null }));
		}
	};
}

// Export store instances
export const authStore = createAuthStore();
export const videoStore = createVideoStore();
export const adminStore = createAdminStore();
export const advertisementStore = createAdvertisementStore();
export const dashboardStore = createDashboardStore();

// Derived stores for convenience
export const isAuthenticated = derived(authStore, $auth => $auth.isAuthenticated);
export const currentUser = derived(authStore, $auth => $auth.user);
export const isAdmin = derived(authStore, $auth => {
	if (!$auth.user) return false;
	
	// Admin roles include all roles with level 7+ (subsystem managers and above)
	const adminRoles = [
		'super_admin',           // Level 10: Super Administrator
		'system_admin',          // Level 9: System Administrator
		'content_manager',       // Level 8: Content Manager
		'articles_manager',      // Level 7: Articles Manager
		'youtube_manager',       // Level 7: YouTube Manager
		'streaming_manager',     // Level 7: Video Streaming Manager
		'events_manager',        // Level 7: Events Manager
		'advertisement_manager', // Level 7: Advertisement Manager
		'user_manager',          // Level 7: User Account Manager
		'analytics_manager',     // Level 7: Analytics Manager
		'financial_admin',       // Level 7: Financial Administrator
		'admin'                  // Legacy admin role
	];
	
	return adminRoles.includes($auth.user.role);
}); 