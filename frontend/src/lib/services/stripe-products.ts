import { apiRequest } from '$lib/auth';

export interface StripeProduct {
	id: number;
	stripe_id: string;
	name: string;
	description: string;
	active: boolean;
	available: boolean;
	video_approved: boolean;
	livemode: boolean;
	created_at: string;
	updated_at: string;
	legacy: boolean;
	price?: {
		id: number;
		stripe_id: string;
		unit_amount: number | null;
		currency: string | null;
		recurring_interval: string | null;
		active: boolean;
	};
}

export interface StripeProductsResponse {
	products: {
		video_approved: StripeProduct[];
		active: StripeProduct[];
		inactive: StripeProduct[];
	};
	total_count: number;
	counts: {
		video_approved: number;
		active: number;
		inactive: number;
	};
}

export class StripeProductsService {
	/**
	 * Get all Stripe products grouped by status for accordion display
	 */
	static async getProductsForAccordion(): Promise<StripeProductsResponse> {
		console.log('🎯 Fetching Stripe products for accordion...');
		
		const response = await apiRequest('/admin/streaming/stripe/products/accordion');
		
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.error || 'Failed to fetch Stripe products');
		}
		
		const data = await response.json();
		console.log('✅ Retrieved Stripe products:', data);
		
		return data;
	}

	/**
	 * Update video approval status for a product
	 */
	static async updateVideoApproval(productId: number, videoApproved: boolean): Promise<StripeProduct> {
		console.log(`🔧 Updating video approval for product ${productId} to ${videoApproved}`);
		
		const response = await apiRequest(`/admin/streaming/stripe/products/video-approval/${productId}`, {
			method: 'PUT',
			body: JSON.stringify({
				video_approved: videoApproved
			})
		});
		
		if (!response.ok) {
			const errorData = await response.json().catch(() => ({}));
			throw new Error(errorData.error || 'Failed to update video approval');
		}
		
		const data = await response.json();
		console.log('✅ Updated product video approval:', data);
		
		return data;
	}

	/**
	 * Format price for display
	 */
	static formatPrice(product: StripeProduct): string {
		if (!product.price || !product.price.unit_amount) {
			return 'No price set';
		}

		const amount = product.price.unit_amount / 100; // Convert cents to dollars
		const currency = product.price.currency?.toUpperCase() || 'USD';
		const interval = product.price.recurring_interval;

		let formattedPrice = `$${amount.toFixed(2)} ${currency}`;
		
		if (interval) {
			formattedPrice += ` / ${interval}`;
		}

		return formattedPrice;
	}

	/**
	 * Get status badge info for a product
	 */
	static getStatusBadge(product: StripeProduct): { text: string; class: string; icon: string } {
		if (product.video_approved) {
			return {
				text: 'Video Approved',
				class: 'badge-success',
				icon: '✅'
			};
		}
		
		if (product.active) {
			return {
				text: 'Active',
				class: 'badge-warning',
				icon: '🟡'
			};
		}
		
		return {
			text: 'Inactive',
			class: 'badge-secondary',
			icon: '⚫'
		};
	}

	/**
	 * Truncate description for display
	 */
	static truncateDescription(description: string, maxLength: number = 150): string {
		if (!description || description.length <= maxLength) {
			return description || 'No description available';
		}
		
		return description.substring(0, maxLength) + '...';
	}
}
