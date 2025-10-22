/**
 * Creator Payout Service
 * API client for all creator payout endpoints
 */

import { apiClient } from '$lib/api/client';
import type {
	Presenter,
	CreatePresenterInput,
	UpdatePresenterInput,
	PresenterStats,
	VideoPresenter,
	CreateVideoPresenterInput,
	UpdateVideoPresenterInput,
	PayoutFormula,
	CreatePayoutFormulaInput,
	UpdatePayoutFormulaInput,
	PresenterPayout,
	UpdatePayoutStatusInput,
	UpdatePayoutAmountsInput,
	PayoutSummary,
	PayoutTransaction,
	CreatePayoutTransactionInput,
	UpdateTransactionStatusInput
} from '$lib/types/creatorPayout';

class CreatorPayoutService {
	// ========================================
	// PRESENTER MANAGEMENT
	// ========================================

	async getPresenters(activeOnly = false, verifiedOnly = false): Promise<{ presenters: Presenter[]; count: number }> {
		const params = new URLSearchParams();
		if (activeOnly) params.append('active', 'true');
		if (verifiedOnly) params.append('verified', 'true');
		
		const response = await apiClient.get<{ presenters: Presenter[]; count: number }>(`/admin/presenters?${params.toString()}`);
		return response.data || { presenters: [], count: 0 };
	}

	async getPresenterById(id: number): Promise<Presenter> {
		const response = await apiClient.get<Presenter>(`/admin/presenters/${id}`);
		return response.data!;
	}

	async createPresenter(data: CreatePresenterInput): Promise<Presenter> {
		const response = await apiClient.post<Presenter>('/admin/presenters', data);
		return response.data!;
	}

	async updatePresenter(id: number, data: UpdatePresenterInput): Promise<Presenter> {
		const response = await apiClient.put<Presenter>(`/admin/presenters/${id}`, data);
		return response.data!;
	}

	async deletePresenter(id: number): Promise<void> {
		await apiClient.delete(`/admin/presenters/${id}`);
	}

	async verifyPresenter(id: number): Promise<void> {
		await apiClient.post(`/admin/presenters/${id}/verify`, {});
	}

	async getPresenterStats(): Promise<PresenterStats> {
		const response = await apiClient.get<PresenterStats>('/admin/presenters/stats');
		return response.data!;
	}

	async getPresenterVideos(presenterId: number): Promise<{ videos: VideoPresenter[]; count: number }> {
		const response = await apiClient.get<{ videos: VideoPresenter[]; count: number }>(`/admin/presenters/${presenterId}/videos`);
		return response.data || { videos: [], count: 0 };
	}

	async updatePresenterStatistics(presenterId: number): Promise<void> {
		await apiClient.post(`/admin/presenters/${presenterId}/update-stats`, {});
	}

	async updateAllPresenterStatistics(): Promise<void> {
		await apiClient.post('/admin/presenters/update-all-stats', {});
	}

	// ========================================
	// VIDEO-PRESENTER LINKING
	// ========================================

	async getVideoPresenters(videoId: number): Promise<{ presenters: VideoPresenter[]; count: number }> {
		const response = await apiClient.get<{ presenters: VideoPresenter[]; count: number }>(`/admin/videos/${videoId}/presenters`);
		return response.data || { presenters: [], count: 0 };
	}

	async linkPresenterToVideo(data: CreateVideoPresenterInput): Promise<VideoPresenter> {
		const response = await apiClient.post<VideoPresenter>('/admin/video-presenters', data);
		return response.data!;
	}

	async updateVideoPresenter(id: number, data: UpdateVideoPresenterInput): Promise<VideoPresenter> {
		const response = await apiClient.put<VideoPresenter>(`/admin/video-presenters/${id}`, data);
		return response.data!;
	}

	async unlinkPresenterFromVideo(id: number): Promise<void> {
		await apiClient.delete(`/admin/video-presenters/${id}`);
	}

	// ========================================
	// PAYOUT FORMULAS
	// ========================================

	async getFormulas(activeOnly = true): Promise<{ formulas: PayoutFormula[]; count: number }> {
		const params = activeOnly ? '?active=true' : '';
		const response = await apiClient.get<{ formulas: PayoutFormula[]; count: number }>(`/admin/payout-formulas${params}`);
		return response.data || { formulas: [], count: 0 };
	}

	async getFormulaById(id: number): Promise<PayoutFormula> {
		const response = await apiClient.get<PayoutFormula>(`/admin/payout-formulas/${id}`);
		return response.data!;
	}

	async getDefaultFormula(): Promise<PayoutFormula> {
		const response = await apiClient.get<PayoutFormula>('/admin/payout-formulas/default');
		return response.data!;
	}

	async createFormula(data: CreatePayoutFormulaInput): Promise<PayoutFormula> {
		const response = await apiClient.post<PayoutFormula>('/admin/payout-formulas', data);
		return response.data!;
	}

	async updateFormula(id: number, data: UpdatePayoutFormulaInput): Promise<PayoutFormula> {
		const response = await apiClient.put<PayoutFormula>(`/admin/payout-formulas/${id}`, data);
		return response.data!;
	}

	async deleteFormula(id: number): Promise<void> {
		await apiClient.delete(`/admin/payout-formulas/${id}`);
	}

	async setDefaultFormula(id: number): Promise<void> {
		await apiClient.post(`/admin/payout-formulas/${id}/set-default`, {});
	}

	// ========================================
	// PAYOUT MANAGEMENT
	// ========================================

	async generateMonthlyPayouts(month: string): Promise<{ generated_count: number; total_amount: number; details: any }> {
		const response = await apiClient.post<{ generated_count: number; total_amount: number; details: any }>('/admin/payouts/generate', { month });
		return response.data!;
	}

	async calculatePresenterPayout(presenterId: number, month: string, formulaId: number): Promise<any> {
		const response = await apiClient.post('/admin/payouts/calculate', {
			presenter_id: presenterId,
			month,
			formula_id: formulaId
		});
		return response.data;
	}

	async getPayoutById(id: number): Promise<PresenterPayout> {
		const response = await apiClient.get<PresenterPayout>(`/admin/payouts/${id}`);
		return response.data!;
	}

	async getPresenterPayouts(presenterId: number): Promise<{ payouts: PresenterPayout[]; count: number }> {
		const response = await apiClient.get<{ payouts: PresenterPayout[]; count: number }>(`/admin/payouts/presenter/${presenterId}`);
		return response.data || { payouts: [], count: 0 };
	}

	async getPayoutsByMonth(month: string, status?: string): Promise<{ payouts: any[]; count: number; month: string }> {
		const params = status ? `?status=${status}` : '';
		const response = await apiClient.get<{ payouts: any[]; count: number; month: string }>(`/admin/payouts/month/${month}${params}`);
		return response.data || { payouts: [], count: 0, month };
	}

	async getPayoutSummary(month: string): Promise<PayoutSummary> {
		const response = await apiClient.get<PayoutSummary>(`/admin/payouts/month/${month}/summary`);
		return response.data!;
	}

	async updatePayoutStatus(id: number, data: UpdatePayoutStatusInput): Promise<PresenterPayout> {
		const response = await apiClient.put<PresenterPayout>(`/admin/payouts/${id}/status`, data);
		return response.data!;
	}

	async updatePayoutAmounts(id: number, data: UpdatePayoutAmountsInput): Promise<PresenterPayout> {
		const response = await apiClient.put<PresenterPayout>(`/admin/payouts/${id}/amounts`, data);
		return response.data!;
	}

	async approvePayouts(payoutIds: number[]): Promise<{ message: string; approved_count: number }> {
		const response = await apiClient.post<{ message: string; approved_count: number }>('/admin/payouts/approve', { payout_ids: payoutIds });
		return response.data!;
	}

	async deletePayout(id: number): Promise<void> {
		await apiClient.delete(`/admin/payouts/${id}`);
	}

	// ========================================
	// TRANSACTIONS
	// ========================================

	async createTransaction(data: CreatePayoutTransactionInput): Promise<PayoutTransaction> {
		const response = await apiClient.post<PayoutTransaction>('/admin/payout-transactions', data);
		return response.data!;
	}

	async getTransactionsByPayout(payoutId: number): Promise<{ transactions: PayoutTransaction[]; count: number }> {
		const response = await apiClient.get<{ transactions: PayoutTransaction[]; count: number }>(`/admin/payout-transactions/payout/${payoutId}`);
		return response.data || { transactions: [], count: 0 };
	}

	async getTransactionsByPresenter(presenterId: number): Promise<{ transactions: PayoutTransaction[]; count: number }> {
		const response = await apiClient.get<{ transactions: PayoutTransaction[]; count: number }>(`/admin/payout-transactions/presenter/${presenterId}`);
		return response.data || { transactions: [], count: 0 };
	}

	async getRecentTransactions(limit = 50, status?: string): Promise<{ transactions: PayoutTransaction[]; count: number }> {
		const params = new URLSearchParams();
		params.append('limit', limit.toString());
		if (status) params.append('status', status);
		const response = await apiClient.get<{ transactions: PayoutTransaction[]; count: number }>(`/admin/payout-transactions/recent?${params.toString()}`);
		return response.data || { transactions: [], count: 0 };
	}

	async updateTransactionStatus(id: number, data: UpdateTransactionStatusInput): Promise<PayoutTransaction> {
		const response = await apiClient.put<PayoutTransaction>(`/admin/payout-transactions/${id}/status`, data);
		return response.data!;
	}
}

export const creatorPayoutService = new CreatorPayoutService();

