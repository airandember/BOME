import { apiRequest } from '$lib/auth';

export interface SyncJob {
	id: string;
	type: string;
	entity_type: string;
	status: 'running' | 'completed' | 'failed';
	progress_percent: number;
	processed_items: number;
	total_items: number;
	started_at: string;
	completed_at?: string;
	error_message?: string;
}

export interface SyncStatus {
	sync_running: boolean;
	recent_jobs: SyncJob[];
	current_job?: {
		id: string;
		type: string;
		entity_type: string;
		progress_percent: number;
		processed_items: number;
		total_items: number;
		started_at: string;
	};
}

export interface SyncResponse {
	message: string;
	note?: string;
	since?: string;
}

export class StripeSyncService {
	private static readonly BASE_URL = '/api/v1/admin/streaming/stripe/sync';

	/**
	 * Trigger an incremental sync (last 24 hours by default)
	 */
	static async triggerIncrementalSync(since?: string): Promise<SyncResponse> {
		const url = since ? `${this.BASE_URL}/incremental?since=${since}` : `${this.BASE_URL}/incremental`;
		const response = await apiRequest(url, {
			method: 'POST'
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: 'Failed to trigger sync' }));
			throw new Error(error.error || 'Failed to trigger incremental sync');
		}

		return response.json();
	}

	/**
	 * Trigger a full initial sync (1.5 years of data)
	 */
	static async triggerInitialSync(): Promise<SyncResponse> {
		const response = await apiRequest(`${this.BASE_URL}/initial`, {
			method: 'POST'
		});

		if (!response.ok) {
			const error = await response.json().catch(() => ({ error: 'Failed to trigger sync' }));
			throw new Error(error.error || 'Failed to trigger initial sync');
		}

		return response.json();
	}

	/**
	 * Get current sync status and progress
	 */
	static async getSyncStatus(): Promise<SyncStatus> {
		const response = await apiRequest(`${this.BASE_URL}/status`);

		if (!response.ok) {
			throw new Error('Failed to get sync status');
		}

		return response.json();
	}

	/**
	 * Get sync job history
	 */
	static async getSyncJobs(limit: number = 10): Promise<{ jobs: SyncJob[]; count: number }> {
		const response = await apiRequest(`${this.BASE_URL}/jobs?limit=${limit}`);

		if (!response.ok) {
			throw new Error('Failed to get sync jobs');
		}

		return response.json();
	}

	/**
	 * Poll sync status until completion
	 */
	static async pollSyncStatus(
		onProgress: (status: SyncStatus) => void,
		intervalMs: number = 2000,
		maxAttempts: number = 300 // 10 minutes max
	): Promise<SyncStatus> {
		let attempts = 0;

		return new Promise((resolve, reject) => {
			const poll = async () => {
				try {
					attempts++;
					const status = await this.getSyncStatus();
					onProgress(status);

					// If no sync is running, we're done
					if (!status.sync_running) {
						resolve(status);
						return;
					}

					// Check for timeout
					if (attempts >= maxAttempts) {
						reject(new Error('Sync polling timeout - sync may still be running'));
						return;
					}

					// Continue polling
					setTimeout(poll, intervalMs);
				} catch (error) {
					reject(error);
				}
			};

			poll();
		});
	}
}
