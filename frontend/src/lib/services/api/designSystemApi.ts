/**
 * Design System API Service
 * Handles all HTTP requests for design system functionality
 */

import type { StyleTheme, DesignToken } from '../designTokenService';
import { SecureTokenStorage } from '$lib/auth';

const API_BASE = '/api/v1/admin';

export interface CreateThemeFromFigmaRequest {
	figmaFileId: string;
	figmaNodeId?: string;
	name?: string;
	description?: string;
}

export interface ImportThemeRequest {
	themeData: string;
}

export interface ActivateThemeRequest {
	themeId: number;
}

export interface APIResponse<T = any> {
	success?: boolean;
	message?: string;
	error?: string;
	details?: string;
	data?: T;
	theme?: StyleTheme;
	themes?: StyleTheme[];
	tokens?: DesignToken[];
	token?: DesignToken;
	count?: number;
	preview?: boolean;
}

class DesignSystemApi {
	private async request<T = any>(
		endpoint: string, 
		options: RequestInit = {}
	): Promise<APIResponse<T>> {
		const url = `${API_BASE}${endpoint}`;
		
		const config: RequestInit = {
			headers: {
				'Content-Type': 'application/json',
				...options.headers,
			},
			...options,
		};

		// Add auth token if available
		const token = SecureTokenStorage.getAccessToken();
		if (token) {
			config.headers = {
				...config.headers,
				'Authorization': `Bearer ${token}`,
			};
		}

		try {
			const response = await fetch(url, config);
			const data = await response.json();

			if (!response.ok) {
				throw new Error(data.error || data.message || `HTTP ${response.status}`);
			}

			return data;
		} catch (error) {
			console.error(`API request failed: ${endpoint}`, error);
			throw error;
		}
	}

	private getAuthHeaders(): HeadersInit {
		const token = SecureTokenStorage.getAccessToken();
		return {
			'Content-Type': 'application/json',
			...(token && { 'Authorization': `Bearer ${token}` })
		};
	}

	// Theme Management
	async getThemes(): Promise<StyleTheme[]> {
		const response = await this.request('/design-system/themes');
		return response.themes || [];
	}

	async getActiveTheme(): Promise<StyleTheme | null> {
		const response = await this.request('/design-system/active');
		return response.theme || null;
	}

	async createTheme(theme: Partial<StyleTheme>): Promise<StyleTheme> {
		const response = await this.request('/design-system/themes', {
			method: 'POST',
			body: JSON.stringify(theme),
		});
		return response.theme!;
	}

	async updateTheme(id: number, theme: Partial<StyleTheme>): Promise<StyleTheme> {
		const response = await this.request(`/design-system/themes/${id}`, {
			method: 'PUT',
			body: JSON.stringify(theme),
		});
		return response.theme!;
	}

	async deleteTheme(id: number): Promise<void> {
		await this.request(`/design-system/themes/${id}`, {
			method: 'DELETE',
		});
	}

	async activateTheme(themeId: number): Promise<StyleTheme> {
		const response = await this.request('/design-system/themes/activate', {
			method: 'POST',
			body: JSON.stringify({ themeId }),
		});
		return response.theme!;
	}

	// Figma Integration
	async createThemeFromFigma(request: CreateThemeFromFigmaRequest): Promise<StyleTheme> {
		const response = await this.request('/design-system/figma/import', {
			method: 'POST',
			body: JSON.stringify(request),
		});
		return response.theme!;
	}

	async syncThemeWithFigma(themeId: number): Promise<StyleTheme> {
		const response = await this.request(`/design-system/figma/sync/${themeId}`, {
			method: 'POST',
		});
		return response.theme!;
	}

	async previewFigmaTokens(figmaFileId: string, figmaNodeId?: string): Promise<DesignToken[]> {
		const params = new URLSearchParams({ fileId: figmaFileId });
		if (figmaNodeId) {
			params.append('nodeId', figmaNodeId);
		}

		const response = await this.request(`/design-system/figma/preview?${params}`);
		return response.tokens || [];
	}

	// Theme Import/Export
	async importTheme(themeData: string): Promise<StyleTheme> {
		const response = await this.request('/design-system/themes/import', {
			method: 'POST',
			body: JSON.stringify({ themeData }),
		});
		return response.theme!;
	}

	async exportTheme(themeId: number): Promise<Blob> {
		const url = `${API_BASE}/design-system/themes/${themeId}/export`;
		
		const config: RequestInit = {
			headers: {},
		};

		// Add auth token if available
		const token = SecureTokenStorage.getAccessToken();
		if (token) {
			config.headers = {
				'Authorization': `Bearer ${token}`,
			};
		}

		const response = await fetch(url, config);
		
		if (!response.ok) {
			const error = await response.json();
			throw new Error(error.error || `HTTP ${response.status}`);
		}

		return response.blob();
	}

	// Token Management
	async getThemeTokens(themeId: number): Promise<DesignToken[]> {
		const response = await this.request(`/design-system/themes/${themeId}/tokens`);
		return response.tokens || [];
	}

	async getAllTokens(filters?: {
		themeId?: number;
		type?: string;
		category?: string;
	}): Promise<DesignToken[]> {
		const params = new URLSearchParams();
		if (filters?.themeId) params.append('themeId', filters.themeId.toString());
		if (filters?.type) params.append('type', filters.type);
		if (filters?.category) params.append('category', filters.category);

		const queryString = params.toString();
		const endpoint = `/design-system/tokens${queryString ? `?${queryString}` : ''}`;
		
		const response = await this.request(endpoint);
		return response.tokens || [];
	}

	async createToken(token: Omit<DesignToken, 'id' | 'createdAt' | 'updatedAt'>): Promise<DesignToken> {
		const response = await this.request('/design-system/tokens', {
			method: 'POST',
			body: JSON.stringify(token),
		});
		return response.token!;
	}

	async updateToken(id: number, token: Partial<DesignToken>): Promise<DesignToken> {
		const response = await this.request(`/design-system/tokens/${id}`, {
			method: 'PUT',
			body: JSON.stringify(token),
		});
		return response.token!;
	}

	async deleteToken(id: number): Promise<void> {
		await this.request(`/design-system/tokens/${id}`, {
			method: 'DELETE',
		});
	}

	// Utility Methods
	async downloadThemeExport(themeId: number, filename?: string): Promise<void> {
		try {
			const blob = await this.exportTheme(themeId);
			const url = URL.createObjectURL(blob);
			const link = document.createElement('a');
			link.href = url;
			link.download = filename || `theme-${themeId}-${new Date().toISOString().split('T')[0]}.json`;
			document.body.appendChild(link);
			link.click();
			document.body.removeChild(link);
			URL.revokeObjectURL(url);
		} catch (error) {
			console.error('Failed to download theme export:', error);
			throw error;
		}
	}

	async validateFigmaFileId(figmaFileId: string): Promise<boolean> {
		try {
			await this.previewFigmaTokens(figmaFileId);
			return true;
		} catch (error) {
			return false;
		}
	}
}

export const designSystemApi = new DesignSystemApi(); 