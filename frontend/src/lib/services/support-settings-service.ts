import { apiClient } from '$lib/api/client';

export interface SupportSettings {
	email: string | null;
	phone: string | null;
	url: string | null;
	hours: string | null;
	message: string | null;
}

export class SupportSettingsService {
	/**
	 * Get support settings (public, no auth required)
	 */
	static async getSupportSettings(): Promise<SupportSettings> {
		try {
			const response = await apiClient.get<{ success: boolean; support: SupportSettings }>(
				'/system/support'
			);

			if (response.error) {
				console.error('Failed to fetch support settings:', response.error);
				return {
					email: null,
					phone: null,
					url: null,
					hours: null,
					message: 'Please contact our support team for assistance.'
				};
			}

			return (
				response.data?.support || {
					email: null,
					phone: null,
					url: null,
					hours: null,
					message: 'Please contact our support team for assistance.'
				}
			);
		} catch (error) {
			console.error('Error fetching support settings:', error);
			return {
				email: null,
				phone: null,
				url: null,
				hours: null,
				message: 'Please contact our support team for assistance.'
			};
		}
	}

	/**
	 * Update support settings (admin only)
	 */
	static async updateSupportSettings(settings: {
		support_email: string;
		support_phone: string;
		support_url: string;
		support_hours: string;
		support_message: string;
	}): Promise<void> {
		const response = await apiClient.put<{ success: boolean; message: string }>(
			'/admin/support-settings/',
			settings
		);

		if (response.error) {
			throw new Error(response.error);
		}
	}

	/**
	 * Get support settings (admin view)
	 */
	static async getAdminSupportSettings(): Promise<SupportSettings> {
		const response = await apiClient.get<{ success: boolean; support: SupportSettings }>(
			'/admin/support-settings/'
		);

		if (response.error) {
			throw new Error(response.error);
		}

		return (
			response.data?.support || {
				email: null,
				phone: null,
				url: null,
				hours: null,
				message: null
			}
		);
	}
}

