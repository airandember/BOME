<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdvertiserAccount, AdCampaign, PlacementPerformance } from '$lib/types/advertising';

	let advertiser: AdvertiserAccount | null = null;
	let campaigns: AdCampaign[] = [];
	let placements: PlacementPerformance[] = [];
	let analytics = {
		totalImpressions: 0,
		totalClicks: 0,
		totalRevenue: 0,
		totalSpent: 0,
		averageCTR: 0,
		activeCampaigns: 0
	};
	let loading = true;
	let error: string | null = null;

	// State management for campaign actions
	let cancellingCampaignId: number | null = null;
	let reviewingCampaignId: number | null = null;
	let reviewingCancelledCampaignId: number | null = null;
	let cancellingAdvertiser = false;

	// Accordion state management for campaigns
	let campaignAccordionState = {
		pending: true,
		approved: false,
		rejected: false,
		cancelled: false
	};

	const advertiserId = parseInt($page.params.id);

	// Mock admin data
	const mockAdmins: { [key: number]: { name: string; email: string } } = {
		1: { name: 'John Smith', email: 'john.smith@bome.com' },
		2: { name: 'Sarah Johnson', email: 'sarah.johnson@bome.com' },
		3: { name: 'Michael Chen', email: 'michael.chen@bome.com' }
	};

	onMount(async () => {
		if (!$auth.isAuthenticated) {
			goto('/login');
			return;
		}

		const user = $auth.user;
		if (!user || user.role !== 'admin') {
			goto('/');
			return;
		}

		try {
			await loadAdvertiserData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	});

	async function loadAdvertiserData() {
		// Mock API call - load advertiser data
		await new Promise(resolve => setTimeout(resolve, 500));

		// Mock advertiser data
		const mockAdvertisers: AdvertiserAccount[] = [
			{
				id: 1,
				user_id: 2,
				company_name: 'TechCorp Solutions',
				business_email: 'contact@techcorp.com',
				contact_name: 'Sarah Johnson',
				contact_phone: '(555) 123-4567',
				business_address: '123 Innovation Drive, Tech City, TC 12345',
				tax_id: '12-3456789',
				website: 'https://techcorp.com',
				industry: 'Technology',
				status: 'approved' as const,
				approved_by: 1,
				approved_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 2,
				user_id: 3,
				company_name: 'Digital Marketing Pro',
				business_email: 'hello@digitalmarketing.com',
				contact_name: 'Michael Chen',
				contact_phone: '(555) 987-6543',
				business_address: '456 Marketing Blvd, Adville, AV 67890',
				tax_id: '98-7654321',
				website: 'https://digitalmarketing.com',
				industry: 'Marketing',
				status: 'approved' as const,
				approved_by: 1,
				approved_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			}
		];

		advertiser = mockAdvertisers.find(adv => adv.id === advertiserId) || null;

		if (!advertiser) {
			throw new Error('Advertiser not found');
		}

		// Mock campaigns for this advertiser
		campaigns = [
			{
				id: 1,
				advertiser_id: advertiserId,
				name: 'Q1 Product Launch Campaign',
				description: 'Promoting our revolutionary new software platform to tech professionals',
				status: 'approved' as const,
				start_date: '2024-03-01',
				end_date: '2024-05-31',
				budget: 15000,
				spent_amount: 8500,
				target_audience: 'Tech professionals, developers, and IT decision makers aged 25-45',
				billing_type: 'monthly',
				billing_rate: 5000,
				approved_by: 1,
				approved_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 4 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 2,
				advertiser_id: advertiserId,
				name: 'Summer Brand Awareness',
				description: 'Building brand recognition and driving website traffic during peak season',
				status: 'active' as const,
				start_date: '2024-06-01',
				end_date: '2024-08-31',
				budget: 8000,
				spent_amount: 2400,
				target_audience: 'Small business owners and marketing professionals',
				billing_type: 'weekly',
				billing_rate: 666.67,
				approved_by: 1,
				approved_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 3,
				advertiser_id: advertiserId,
				name: 'Holiday Special Promotion',
				description: 'Seasonal campaign targeting holiday shoppers with exclusive deals',
				status: 'pending' as const,
				start_date: '2024-11-01',
				end_date: '2024-12-31',
				budget: 12000,
				spent_amount: 0,
				target_audience: 'Holiday shoppers, families, and gift buyers aged 25-55',
				billing_type: 'monthly',
				billing_rate: 6000,
				created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 4,
				advertiser_id: advertiserId,
				name: 'Back to School Campaign',
				description: 'Targeting students and parents for educational technology products',
				status: 'rejected' as const,
				start_date: '2024-08-01',
				end_date: '2024-09-30',
				budget: 5000,
				spent_amount: 0,
				target_audience: 'Students, parents, and educators',
				billing_type: 'monthly',
				billing_rate: 2500,
				approval_notes: 'Campaign content does not align with platform guidelines',
				rejected_by: 2,
				rejected_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			}
		];

		// Mock placement performance data
		placements = [
			{
				placement_id: 1,
				placement_name: 'Homepage Header',
				total_ads: 5,
				active_ads: 3,
				total_impressions: 45000,
				total_clicks: 1350,
				total_revenue: 2700,
				average_ctr: 3.0,
				fill_rate: 85.5
			},
			{
				placement_id: 2,
				placement_name: 'Sidebar Banner',
				total_ads: 3,
				active_ads: 2,
				total_impressions: 28000,
				total_clicks: 840,
				total_revenue: 1680,
				average_ctr: 3.0,
				fill_rate: 92.1
			},
			{
				placement_id: 3,
				placement_name: 'Video Overlay',
				total_ads: 2,
				active_ads: 1,
				total_impressions: 12000,
				total_clicks: 480,
				total_revenue: 960,
				average_ctr: 4.0,
				fill_rate: 78.3
			}
		];

		// Calculate analytics
		analytics = {
			totalImpressions: placements.reduce((sum, p) => sum + p.total_impressions, 0),
			totalClicks: placements.reduce((sum, p) => sum + p.total_clicks, 0),
			totalRevenue: placements.reduce((sum, p) => sum + p.total_revenue, 0),
			totalSpent: campaigns.reduce((sum, c) => sum + c.spent_amount, 0),
			averageCTR: placements.length > 0 ? placements.reduce((sum, p) => sum + p.average_ctr, 0) / placements.length : 0,
			activeCampaigns: campaigns.filter(c => c.status === 'active' || c.status === 'approved').length
		};
	}

	function getAdminName(adminId: number): string {
		return mockAdmins[adminId]?.name || 'Unknown Admin';
	}

	function getStatusBadgeClass(status: string): string {
		switch (status) {
			case 'approved':
			case 'active':
				return 'status-approved';
			case 'pending':
				return 'status-pending';
			case 'rejected':
				return 'status-rejected';
			case 'cancelled':
				return 'status-cancelled';
			case 'paused':
				return 'status-paused';
			case 'completed':
				return 'status-completed';
			default:
				return 'status-default';
		}
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('en-US', {
			style: 'currency',
			currency: 'USD'
		}).format(amount);
	}

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}

	function formatDateTime(dateString: string): string {
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatNumber(num: number): string {
		return new Intl.NumberFormat('en-US').format(num);
	}

	function formatPercentage(num: number): string {
		return `${num.toFixed(1)}%`;
	}

	// Campaign action functions
	async function approveCampaign(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { ...campaign, status: 'approved' as const } : campaign
			);
			
			// Recalculate analytics
			analytics.activeCampaigns = campaigns.filter(c => c.status === 'active' || c.status === 'approved').length;
		} catch (err) {
			error = 'Failed to approve campaign';
		}
	}

	async function rejectCampaign(campaignId: number) {
		const reason = prompt('Please provide a reason for rejection:');
		if (!reason) return;

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { ...campaign, status: 'rejected' as const, approval_notes: reason } : campaign
			);
			
			// Recalculate analytics
			analytics.activeCampaigns = campaigns.filter(c => c.status === 'active' || c.status === 'approved').length;
		} catch (err) {
			error = 'Failed to reject campaign';
		}
	}

	function showCancelConfirmation(campaignId: number) {
		cancellingCampaignId = campaignId;
	}

	function hideCancelConfirmation() {
		cancellingCampaignId = null;
	}

	async function confirmCancelCampaign(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { 
					...campaign, 
					status: 'cancelled' as const,
					cancelled_by: 1, // Mock current admin ID
					cancelled_at: new Date().toISOString()
				} : campaign
			);
			
			cancellingCampaignId = null;
			
			// Recalculate analytics
			analytics.activeCampaigns = campaigns.filter(c => c.status === 'active' || c.status === 'approved').length;
		} catch (err) {
			error = 'Failed to cancel campaign';
		}
	}

	function showReviewConfirmation(campaignId: number) {
		reviewingCampaignId = campaignId;
	}

	function hideReviewConfirmation() {
		reviewingCampaignId = null;
	}

	async function confirmReviewApproval(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { 
					...campaign, 
					status: 'approved' as const,
					approved_by: 1, // Mock current admin ID
					approved_at: new Date().toISOString(),
					// Clear rejection data
					rejected_by: undefined,
					rejected_at: undefined,
					approval_notes: 'Campaign approved after review'
				} : campaign
			);
			
			reviewingCampaignId = null;
			
			// Recalculate analytics
			analytics.activeCampaigns = campaigns.filter(c => c.status === 'active' || c.status === 'approved').length;
		} catch (err) {
			error = 'Failed to approve campaign';
		}
	}

	function showCancelledCampaignReviewConfirmation(campaignId: number) {
		reviewingCancelledCampaignId = campaignId;
	}

	function hideCancelledCampaignReviewConfirmation() {
		reviewingCancelledCampaignId = null;
	}

	async function confirmCancelledCampaignReactivation(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { 
					...campaign, 
					status: 'approved' as const,
					approved_by: 1, // Mock current admin ID
					approved_at: new Date().toISOString(),
					// Clear cancellation data
					cancelled_by: undefined,
					cancelled_at: undefined,
					approval_notes: 'Campaign reactivated after review'
				} : campaign
			);
			
			reviewingCancelledCampaignId = null;
			
			// Recalculate analytics
			analytics.activeCampaigns = campaigns.filter(c => c.status === 'active' || c.status === 'approved').length;
		} catch (err) {
			error = 'Failed to reactivate campaign';
		}
	}

	// Advertiser cancellation functions
	function showAdvertiserCancelConfirmation() {
		cancellingAdvertiser = true;
	}

	function hideAdvertiserCancelConfirmation() {
		cancellingAdvertiser = false;
	}

	async function confirmAdvertiserCancellation() {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			if (advertiser) {
				advertiser = {
					...advertiser,
					status: 'cancelled' as const,
					cancelled_by: 1, // Mock current admin ID
					cancelled_at: new Date().toISOString()
				};
			}
			
			cancellingAdvertiser = false;
		} catch (err) {
			error = 'Failed to cancel advertiser';
		}
	}

	function getCampaignStatusInfo(campaign: AdCampaign): { text: string; date: string; admin: string } | null {
		switch (campaign.status) {
			case 'approved':
				return {
					text: 'Approved',
					date: campaign.approved_at ? formatDateTime(campaign.approved_at) : '',
					admin: campaign.approved_by ? getAdminName(campaign.approved_by) : ''
				};
			case 'rejected':
				return {
					text: 'Rejected',
					date: campaign.rejected_at ? formatDateTime(campaign.rejected_at) : '',
					admin: campaign.rejected_by ? getAdminName(campaign.rejected_by) : ''
				};
			case 'cancelled':
				return {
					text: 'Cancelled',
					date: campaign.cancelled_at ? formatDateTime(campaign.cancelled_at) : '',
					admin: campaign.cancelled_by ? getAdminName(campaign.cancelled_by) : ''
				};
			default:
				return null;
		}
	}

	// Accordion helper functions
	function toggleCampaignAccordion(status: 'pending' | 'approved' | 'rejected' | 'cancelled') {
		campaignAccordionState[status] = !campaignAccordionState[status];
	}

	function getCampaignsByStatus(status: string): AdCampaign[] {
		if (status === 'approved') {
			return campaigns.filter(campaign => campaign.status === 'approved' || campaign.status === 'active');
		}
		return campaigns.filter(campaign => campaign.status === status);
	}

	function getCampaignCounts() {
		return {
			pending: campaigns.filter(camp => camp.status === 'pending').length,
			approved: campaigns.filter(camp => camp.status === 'approved' || camp.status === 'active').length,
			rejected: campaigns.filter(camp => camp.status === 'rejected').length,
			cancelled: campaigns.filter(camp => camp.status === 'cancelled').length
		};
	}

	// Computed campaign counts
	$: campaignCounts = getCampaignCounts();
</script>

<svelte:head>
	<title>{advertiser?.company_name || 'Advertiser'} - BOME Admin</title>
</svelte:head>

<div class="advertiser-detail">
	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading advertiser details...</p>
		</div>
	{:else if error}
		<div class="error-container">
			<div class="alert alert-error">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
				</svg>
				<span>{error}</span>
			</div>
			<button class="btn btn-secondary" on:click={() => goto('/admin/advertisements')}>
				← Back to Advertisements
			</button>
		</div>
	{:else if advertiser}
		<!-- Header -->
		<div class="page-header">
			<div class="header-nav">
				<button class="btn btn-secondary" on:click={() => goto('/admin/advertisements')}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M19 12H5m7-7l-7 7 7 7" />
					</svg>
					Back to Advertisements
				</button>
			</div>
			<div class="header-content">
				<div class="advertiser-header">
					<div class="advertiser-avatar">
						{advertiser.company_name.charAt(0)}
					</div>
					<div class="advertiser-info">
						<h1>{advertiser.company_name}</h1>
						<div class="advertiser-meta">
							<span class="status-badge {getStatusBadgeClass(advertiser.status)}">
								{advertiser.status}
							</span>
							<span class="industry-tag">{advertiser.industry}</span>
						</div>
						<div class="contact-info">
							<span>{advertiser.contact_name}</span>
							<span>•</span>
							<span>{advertiser.business_email}</span>
							{#if advertiser.contact_phone}
								<span>•</span>
								<span>{advertiser.contact_phone}</span>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Analytics Overview -->
		<div class="analytics-section">
			<h2>Performance Overview</h2>
			<div class="analytics-grid">
				<div class="analytics-card">
					<div class="analytics-icon impressions">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
							<circle cx="12" cy="12" r="3" />
						</svg>
					</div>
					<div class="analytics-data">
						<h3>{formatNumber(analytics.totalImpressions)}</h3>
						<p>Total Impressions</p>
					</div>
				</div>
				<div class="analytics-card">
					<div class="analytics-icon clicks">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4" />
							<path d="M21 12c-1 0-3-1-3-3s2-3 3-3 3 1 3 3-2 3-3 3" />
							<path d="M3 12c1 0 3-1 3-3s-2-3-3-3-3 1-3 3 2 3 3 3" />
						</svg>
					</div>
					<div class="analytics-data">
						<h3>{formatNumber(analytics.totalClicks)}</h3>
						<p>Total Clicks</p>
					</div>
				</div>
				<div class="analytics-card">
					<div class="analytics-icon revenue">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="12" y1="1" x2="12" y2="23" />
							<path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
						</svg>
					</div>
					<div class="analytics-data">
						<h3>{formatCurrency(analytics.totalRevenue)}</h3>
						<p>Total Revenue</p>
					</div>
				</div>
				<div class="analytics-card">
					<div class="analytics-icon ctr">
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 3v18h18" />
							<path d="M18.7 8l-5.1 5.2-2.8-2.7L7 14.3" />
						</svg>
					</div>
					<div class="analytics-data">
						<h3>{formatPercentage(analytics.averageCTR)}</h3>
						<p>Average CTR</p>
					</div>
				</div>
			</div>
		</div>

		<!-- Campaigns Section -->
		<div class="campaigns-section">
			<div class="section-header">
				<h2>Campaigns ({campaigns.length})</h2>
				<span class="active-count">{analytics.activeCampaigns} Active</span>
			</div>

			<!-- Pending Campaigns Accordion -->
			<div class="accordion-section">
				<button 
					class="accordion-header {campaignAccordionState.pending ? 'active' : ''}"
					on:click={() => toggleCampaignAccordion('pending')}
				>
					<div class="accordion-title">
						<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10" />
							<path d="M12 6v6l4 2" />
						</svg>
						<span>Pending Approval</span>
						<span class="accordion-count">{campaignCounts.pending}</span>
					</div>
					<svg class="accordion-chevron {campaignAccordionState.pending ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 9l6 6 6-6" />
					</svg>
				</button>
				{#if campaignAccordionState.pending}
					{@const pendingCampaigns = getCampaignsByStatus('pending')}
					<div class="accordion-content">
						{#if pendingCampaigns.length === 0}
							<div class="empty-state">
								<div class="empty-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10" />
										<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
										<path d="M12 17h.01" />
									</svg>
								</div>
								<h3>No pending campaigns</h3>
								<p>No campaigns are currently pending approval.</p>
							</div>
						{:else}
							<div class="campaigns-grid">
								{#each pendingCampaigns as campaign}
									{@const campaignStatusInfo = getCampaignStatusInfo(campaign)}
									<div class="campaign-card">
										<div class="campaign-header">
											<div class="campaign-info">
												<div class="campaign-icon">
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
														<line x1="8" y1="21" x2="16" y2="21" />
														<line x1="12" y1="17" x2="12" y2="21" />
													</svg>
												</div>
												<div class="campaign-details">
													<h3>{campaign.name}</h3>
													<div class="campaign-meta">
														<span class="status-badge {getStatusBadgeClass(campaign.status)}">
															{campaign.status}
														</span>
													</div>
													<div class="card-actions">
														<button
															class="btn btn-sm btn-success"
															on:click={() => approveCampaign(campaign.id)}
														>
															<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M9 12l2 2 4-4" />
															</svg>
															Approve
														</button>
														<button
															class="btn btn-sm btn-error"
															on:click={() => rejectCampaign(campaign.id)}
														>
															<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M18 6L6 18M6 6l12 12" />
															</svg>
															Reject
														</button>
													</div>
												</div>
											</div>
										</div>
										
										<div class="campaign-content">
											<div class="campaign-description">
												<p>{campaign.description}</p>
											</div>
											
											<div class="campaign-metrics">
												<div class="metric">
													<label>Budget</label>
													<span class="budget-amount">{formatCurrency(campaign.budget)}</span>
												</div>
												<div class="metric">
													<label>Duration</label>
													<span>{formatDate(campaign.start_date)} - {campaign.end_date ? formatDate(campaign.end_date) : 'Ongoing'}</span>
												</div>
												<div class="metric">
													<label>Billing</label>
													<span>{formatCurrency(campaign.billing_rate)} / {campaign.billing_type}</span>
												</div>
											</div>
											
											{#if campaign.target_audience}
												<div class="audience-section">
													<label>Target Audience</label>
													<p>{campaign.target_audience}</p>
												</div>
											{/if}
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Approved Campaigns Accordion -->
			<div class="accordion-section">
				<button 
					class="accordion-header {campaignAccordionState.approved ? 'active' : ''}"
					on:click={() => toggleCampaignAccordion('approved')}
				>
					<div class="accordion-title">
						<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M9 12l2 2 4-4" />
							<circle cx="12" cy="12" r="10" />
						</svg>
						<span>Approved</span>
						<span class="accordion-count">{campaignCounts.approved}</span>
					</div>
					<svg class="accordion-chevron {campaignAccordionState.approved ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 9l6 6 6-6" />
					</svg>
				</button>
				{#if campaignAccordionState.approved}
					{@const approvedCampaigns = getCampaignsByStatus('approved')}
					<div class="accordion-content">
						{#if approvedCampaigns.length === 0}
							<div class="empty-state">
								<div class="empty-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10" />
										<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
										<path d="M12 17h.01" />
									</svg>
								</div>
								<h3>No approved campaigns</h3>
								<p>No approved campaigns found.</p>
							</div>
						{:else}
							<div class="campaigns-grid">
								{#each approvedCampaigns as campaign}
									{@const campaignStatusInfo = getCampaignStatusInfo(campaign)}
									<div class="campaign-card">
										<div class="campaign-header">
											<div class="campaign-info">
												<div class="campaign-icon">
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
														<line x1="8" y1="21" x2="16" y2="21" />
														<line x1="12" y1="17" x2="12" y2="21" />
													</svg>
												</div>
												<div class="campaign-details">
													<h3>{campaign.name}</h3>
													<div class="campaign-meta">
														<span class="status-badge {getStatusBadgeClass(campaign.status)}">
															{campaign.status}
														</span>
														{#if campaignStatusInfo}
															<div class="status-info">
																<span class="status-action">{campaignStatusInfo.text} by {campaignStatusInfo.admin}</span>
																<span class="status-date">{campaignStatusInfo.date}</span>
															</div>
														{/if}
													</div>
												</div>
											</div>
										</div>
										
										<div class="campaign-content">
											<div class="campaign-description">
												<p>{campaign.description}</p>
											</div>
											
											<div class="campaign-metrics">
												<div class="metric">
													<label>Budget</label>
													<span class="budget-amount">{formatCurrency(campaign.budget)}</span>
												</div>
												<div class="metric">
													<label>Spent</label>
													<span>{formatCurrency(campaign.spent_amount)}</span>
												</div>
												<div class="metric">
													<label>Duration</label>
													<span>{formatDate(campaign.start_date)} - {campaign.end_date ? formatDate(campaign.end_date) : 'Ongoing'}</span>
												</div>
												<div class="metric">
													<label>Billing</label>
													<span>{formatCurrency(campaign.billing_rate)} / {campaign.billing_type}</span>
												</div>
											</div>
											
											{#if campaign.target_audience}
												<div class="audience-section">
													<label>Target Audience</label>
													<p>{campaign.target_audience}</p>
												</div>
											{/if}
											
											<!-- Cancel button for approved campaigns -->
											<div class="cancel-section">
												{#if cancellingCampaignId === campaign.id}
													<div class="cancel-confirmation">
														<p class="confirmation-text">Are you sure you want to cancel this campaign?</p>
														<div class="confirmation-actions">
															<button
																class="btn btn-sm btn-error"
																on:click={() => confirmCancelCampaign(campaign.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M9 12l2 2 4-4" />
																</svg>
																Confirm
															</button>
															<button
																class="btn btn-sm btn-secondary"
																on:click={hideCancelConfirmation}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M18 6L6 18M6 6l12 12" />
																</svg>
																Cancel
															</button>
														</div>
													</div>
												{:else}
													<button
														class="btn btn-cancel"
														on:click={() => showCancelConfirmation(campaign.id)}
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
														</svg>
														Cancel Campaign
													</button>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Rejected Campaigns Accordion -->
			<div class="accordion-section">
				<button 
					class="accordion-header {campaignAccordionState.rejected ? 'active' : ''}"
					on:click={() => toggleCampaignAccordion('rejected')}
				>
					<div class="accordion-title">
						<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10" />
							<line x1="15" y1="9" x2="9" y2="15" />
							<line x1="9" y1="9" x2="15" y2="15" />
						</svg>
						<span>Rejected</span>
						<span class="accordion-count">{campaignCounts.rejected}</span>
					</div>
					<svg class="accordion-chevron {campaignAccordionState.rejected ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 9l6 6 6-6" />
					</svg>
				</button>
				{#if campaignAccordionState.rejected}
					{@const rejectedCampaigns = getCampaignsByStatus('rejected')}
					<div class="accordion-content">
						{#if rejectedCampaigns.length === 0}
							<div class="empty-state">
								<div class="empty-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10" />
										<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
										<path d="M12 17h.01" />
									</svg>
								</div>
								<h3>No rejected campaigns</h3>
								<p>No rejected campaigns found.</p>
							</div>
						{:else}
							<div class="campaigns-grid">
								{#each rejectedCampaigns as campaign}
									{@const campaignStatusInfo = getCampaignStatusInfo(campaign)}
									<div class="campaign-card">
										<div class="campaign-header">
											<div class="campaign-info">
												<div class="campaign-icon">
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
														<line x1="8" y1="21" x2="16" y2="21" />
														<line x1="12" y1="17" x2="12" y2="21" />
													</svg>
												</div>
												<div class="campaign-details">
													<h3>{campaign.name}</h3>
													<div class="campaign-meta">
														<span class="status-badge {getStatusBadgeClass(campaign.status)}">
															{campaign.status}
														</span>
														{#if campaignStatusInfo}
															<div class="status-info">
																<span class="status-action">{campaignStatusInfo.text} by {campaignStatusInfo.admin}</span>
																<span class="status-date">{campaignStatusInfo.date}</span>
															</div>
														{/if}
													</div>
												</div>
											</div>
										</div>
										
										<div class="campaign-content">
											<div class="campaign-description">
												<p>{campaign.description}</p>
											</div>
											
											<div class="campaign-metrics">
												<div class="metric">
													<label>Budget</label>
													<span class="budget-amount">{formatCurrency(campaign.budget)}</span>
												</div>
												<div class="metric">
													<label>Duration</label>
													<span>{formatDate(campaign.start_date)} - {campaign.end_date ? formatDate(campaign.end_date) : 'Ongoing'}</span>
												</div>
												<div class="metric">
													<label>Billing</label>
													<span>{formatCurrency(campaign.billing_rate)} / {campaign.billing_type}</span>
												</div>
											</div>
											
											{#if campaign.target_audience}
												<div class="audience-section">
													<label>Target Audience</label>
													<p>{campaign.target_audience}</p>
												</div>
											{/if}
											
											{#if campaign.approval_notes}
												<div class="notes-section">
													<label>Notes</label>
													<p>{campaign.approval_notes}</p>
												</div>
											{/if}
											
											<!-- Review button for rejected campaigns -->
											<div class="cancel-section">
												{#if reviewingCampaignId === campaign.id}
													<div class="review-confirmation">
														<p class="confirmation-text">Would you like to approve this campaign?</p>
														<div class="confirmation-actions">
															<button
																class="btn btn-sm btn-success"
																on:click={() => confirmReviewApproval(campaign.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M9 12l2 2 4-4" />
																</svg>
																Approve
															</button>
															<button
																class="btn btn-sm btn-secondary"
																on:click={hideReviewConfirmation}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M18 6L6 18M6 6l12 12" />
																</svg>
																Cancel
															</button>
														</div>
													</div>
												{:else}
													<button
														class="btn btn-review"
														on:click={() => showReviewConfirmation(campaign.id)}
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<circle cx="11" cy="11" r="8" />
															<path d="m21 21-4.35-4.35" />
														</svg>
														Review Campaign
													</button>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Cancelled Campaigns Accordion -->
			<div class="accordion-section">
				<button 
					class="accordion-header {campaignAccordionState.cancelled ? 'active' : ''}"
					on:click={() => toggleCampaignAccordion('cancelled')}
				>
					<div class="accordion-title">
						<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<circle cx="12" cy="12" r="10" />
							<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
						</svg>
						<span>Cancelled</span>
						<span class="accordion-count">{campaignCounts.cancelled}</span>
					</div>
					<svg class="accordion-chevron {campaignAccordionState.cancelled ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M6 9l6 6 6-6" />
					</svg>
				</button>
				{#if campaignAccordionState.cancelled}
					{@const cancelledCampaigns = getCampaignsByStatus('cancelled')}
					<div class="accordion-content">
						{#if cancelledCampaigns.length === 0}
							<div class="empty-state">
								<div class="empty-icon">
									<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
										<circle cx="12" cy="12" r="10" />
										<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
										<path d="M12 17h.01" />
									</svg>
								</div>
								<h3>No cancelled campaigns</h3>
								<p>No cancelled campaigns found.</p>
							</div>
						{:else}
							<div class="campaigns-grid">
								{#each cancelledCampaigns as campaign}
									{@const campaignStatusInfo = getCampaignStatusInfo(campaign)}
									<div class="campaign-card">
										<div class="campaign-header">
											<div class="campaign-info">
												<div class="campaign-icon">
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
														<line x1="8" y1="21" x2="16" y2="21" />
														<line x1="12" y1="17" x2="12" y2="21" />
													</svg>
												</div>
												<div class="campaign-details">
													<h3>{campaign.name}</h3>
													<div class="campaign-meta">
														<span class="status-badge {getStatusBadgeClass(campaign.status)}">
															{campaign.status}
														</span>
														{#if campaignStatusInfo}
															<div class="status-info">
																<span class="status-action">{campaignStatusInfo.text} by {campaignStatusInfo.admin}</span>
																<span class="status-date">{campaignStatusInfo.date}</span>
															</div>
														{/if}
													</div>
												</div>
											</div>
										</div>
										
										<div class="campaign-content">
											<div class="campaign-description">
												<p>{campaign.description}</p>
											</div>
											
											<div class="campaign-metrics">
												<div class="metric">
													<label>Budget</label>
													<span class="budget-amount">{formatCurrency(campaign.budget)}</span>
												</div>
												<div class="metric">
													<label>Spent</label>
													<span>{formatCurrency(campaign.spent_amount)}</span>
												</div>
												<div class="metric">
													<label>Duration</label>
													<span>{formatDate(campaign.start_date)} - {campaign.end_date ? formatDate(campaign.end_date) : 'Ongoing'}</span>
												</div>
												<div class="metric">
													<label>Billing</label>
													<span>{formatCurrency(campaign.billing_rate)} / {campaign.billing_type}</span>
												</div>
											</div>
											
											{#if campaign.target_audience}
												<div class="audience-section">
													<label>Target Audience</label>
													<p>{campaign.target_audience}</p>
												</div>
											{/if}
											
											<!-- Review button for cancelled campaigns -->
											<div class="cancel-section">
												{#if reviewingCancelledCampaignId === campaign.id}
													<div class="review-confirmation">
														<p class="confirmation-text">Would you like to reactivate this campaign?</p>
														<div class="confirmation-actions">
															<button
																class="btn btn-sm btn-success"
																on:click={() => confirmCancelledCampaignReactivation(campaign.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M9 12l2 2 4-4" />
																</svg>
																Reactivate
															</button>
															<button
																class="btn btn-sm btn-secondary"
																on:click={hideCancelledCampaignReviewConfirmation}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M18 6L6 18M6 6l12 12" />
																</svg>
																Cancel
															</button>
														</div>
													</div>
												{:else}
													<button
														class="btn btn-review"
														on:click={() => showCancelledCampaignReviewConfirmation(campaign.id)}
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<circle cx="11" cy="11" r="8" />
															<path d="m21 21-4.35-4.35" />
														</svg>
														Review Campaign
													</button>
												{/if}
											</div>
										</div>
									</div>
								{/each}
							</div>
						{/if}
					</div>
				{/if}
			</div>
		</div>

		<!-- Placements Section -->
		<div class="placements-section">
			<h2>Ad Placements Performance</h2>
			<div class="placements-table">
				<div class="table-header">
					<div class="header-cell">Placement</div>
					<div class="header-cell">Impressions</div>
					<div class="header-cell">Clicks</div>
					<div class="header-cell">CTR</div>
					<div class="header-cell">Revenue</div>
					<div class="header-cell">Fill Rate</div>
				</div>
				{#each placements as placement}
					<div class="table-row">
						<div class="table-cell placement-name">
							<div class="placement-info">
								<span class="name">{placement.placement_name}</span>
								<span class="ads-count">{placement.active_ads}/{placement.total_ads} ads</span>
							</div>
						</div>
						<div class="table-cell">{formatNumber(placement.total_impressions)}</div>
						<div class="table-cell">{formatNumber(placement.total_clicks)}</div>
						<div class="table-cell">{formatPercentage(placement.average_ctr)}</div>
						<div class="table-cell revenue-cell">{formatCurrency(placement.total_revenue)}</div>
						<div class="table-cell">{formatPercentage(placement.fill_rate)}</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- Account Information -->
		<div class="account-info-section">
			<h2>Account Information</h2>
			<div class="info-grid">
				<div class="info-card">
					<h3>Business Details</h3>
					<div class="info-items">
						<div class="info-item">
							<label>Company</label>
							<span>{advertiser.company_name}</span>
						</div>
						<div class="info-item">
							<label>Industry</label>
							<span>{advertiser.industry || 'Not specified'}</span>
						</div>
						<div class="info-item">
							<label>Website</label>
							<span>{advertiser.website || 'Not provided'}</span>
						</div>
						<div class="info-item">
							<label>Tax ID</label>
							<span>{advertiser.tax_id || 'Not provided'}</span>
						</div>
					</div>
				</div>
				<div class="info-card">
					<h3>Contact Information</h3>
					<div class="info-items">
						<div class="info-item">
							<label>Contact Name</label>
							<span>{advertiser.contact_name}</span>
						</div>
						<div class="info-item">
							<label>Email</label>
							<span>{advertiser.business_email}</span>
						</div>
						<div class="info-item">
							<label>Phone</label>
							<span>{advertiser.contact_phone || 'Not provided'}</span>
						</div>
						<div class="info-item">
							<label>Address</label>
							<span>{advertiser.business_address || 'Not provided'}</span>
						</div>
					</div>
				</div>
				<div class="info-card">
					<h3>Account Status</h3>
					<div class="info-items">
						<div class="info-item">
							<label>Status</label>
							<span class="status-badge {getStatusBadgeClass(advertiser.status)}">
								{advertiser.status}
							</span>
						</div>
						{#if advertiser.approved_by}
							<div class="info-item">
								<label>Approved By</label>
								<span>{getAdminName(advertiser.approved_by)}</span>
							</div>
							<div class="info-item">
								<label>Approved Date</label>
								<span>{formatDateTime(advertiser.approved_at || '')}</span>
							</div>
						{/if}
						<div class="info-item">
							<label>Created</label>
							<span>{formatDateTime(advertiser.created_at)}</span>
						</div>
					</div>
				</div>
			</div>
		</div>

		<!-- Cancel Advertiser Section -->
		{#if advertiser.status === 'approved'}
			<div class="cancel-advertiser-section">
				{#if cancellingAdvertiser}
					<div class="cancel-confirmation">
						<p class="confirmation-text">Are you sure you want to cancel this advertiser account? This will affect all their campaigns.</p>
						<div class="confirmation-actions">
							<button
								class="btn btn-error"
								on:click={confirmAdvertiserCancellation}
							>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 12l2 2 4-4" />
								</svg>
								Confirm Cancellation
							</button>
							<button
								class="btn btn-secondary"
								on:click={hideAdvertiserCancelConfirmation}
							>
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M18 6L6 18M6 6l12 12" />
								</svg>
								Cancel
							</button>
						</div>
					</div>
				{:else}
					<button
						class="btn btn-cancel-advertiser"
						on:click={showAdvertiserCancelConfirmation}
					>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
						</svg>
						Cancel Advertiser Account
					</button>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.advertiser-detail {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
		padding: var(--space-xl);
	}

	.loading-container,
	.error-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-4xl);
		gap: var(--space-lg);
	}

	/* Header */
	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.header-nav {
		display: flex;
		align-items: center;
	}

	.header-content {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.advertiser-header {
		display: flex;
		align-items: center;
		gap: var(--space-lg);
	}

	.advertiser-avatar {
		width: 80px;
		height: 80px;
		background: var(--primary-gradient);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		color: var(--white);
		font-size: var(--text-2xl);
	}

	.advertiser-info h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.advertiser-meta {
		display: flex;
		gap: var(--space-md);
		align-items: center;
		margin-bottom: var(--space-sm);
	}

	.industry-tag {
		background: var(--bg-glass-dark);
		color: var(--text-secondary);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 500;
	}

	.contact-info {
		display: flex;
		gap: var(--space-sm);
		color: var(--text-secondary);
		font-size: var(--text-sm);
	}

	/* Analytics Section */
	.analytics-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.analytics-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.analytics-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: var(--space-lg);
	}

	.analytics-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		display: flex;
		align-items: center;
		gap: var(--space-md);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.analytics-icon {
		width: 48px;
		height: 48px;
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
	}

	.analytics-icon.impressions { background: var(--primary-gradient); }
	.analytics-icon.clicks { background: var(--secondary-gradient); }
	.analytics-icon.revenue { background: var(--success); }
	.analytics-icon.ctr { background: var(--info); }

	.analytics-icon svg {
		width: 24px;
		height: 24px;
	}

	.analytics-data h3 {
		font-size: var(--text-xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.analytics-data p {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0;
	}

	/* Campaigns Section */
	.campaigns-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.section-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-lg);
	}

	.section-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.active-count {
		background: var(--success);
		color: var(--white);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
	}

	.campaigns-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: var(--space-lg);
	}

	.campaign-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.campaign-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-md);
		padding-bottom: var(--space-lg);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.campaign-info {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		flex: 1;
		min-width: 0;
	}

	.campaign-icon {
		width: 48px;
		height: 48px;
		background: var(--secondary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--white);
		flex-shrink: 0;
	}

	.campaign-icon svg {
		width: 24px;
		height: 24px;
	}

	.campaign-details {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
		min-width: 0;
	}

	.campaign-header h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
		flex: 1;
		word-wrap: break-word;
	}

	.campaign-meta {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
		align-items: center;
	}

	.status-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
		font-size: var(--text-xs);
	}

	.status-action {
		color: var(--text-secondary);
		font-weight: 500;
	}

	.status-date {
		color: var(--text-tertiary);
		font-weight: 400;
	}

	.card-actions {
		display: flex;
		gap: var(--space-sm);
		margin-top: var(--space-sm);
		justify-content: center;
		flex-wrap: wrap;
	}

	.card-actions .btn {
		flex: 1;
		max-width: 120px;
		min-width: 80px;
	}

	.campaign-content {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.campaign-description {
		color: var(--text-secondary);
		font-size: var(--text-sm);
		line-height: 1.5;
		margin: 0 0 var(--space-md) 0;
	}

	.campaign-description p {
		margin: 0;
	}

	.audience-section,
	.notes-section {
		margin-bottom: var(--space-lg);
	}

	.audience-section:last-child,
	.notes-section:last-child {
		margin-bottom: 0;
	}

	.audience-section p,
	.notes-section p {
		color: var(--text-secondary);
		line-height: 1.5;
		margin: 0;
	}

	.audience-section label,
	.notes-section label {
		display: block;
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: var(--space-xs);
	}

	/* Cancel Section */
	.cancel-section {
		margin-top: var(--space-lg);
		padding-top: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.05);
	}

	.btn-cancel {
		width: 100%;
		background: rgba(239, 68, 68, 0.1);
		color: rgb(239, 68, 68);
		border: 1px solid rgba(239, 68, 68, 0.2);
		transition: all var(--transition-normal);
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		font-size: var(--text-sm);
		font-weight: 500;
	}

	.btn-cancel svg {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	.btn-cancel:hover {
		background: rgba(239, 68, 68, 0.2);
		border-color: rgba(239, 68, 68, 0.3);
		transform: translateY(-1px);
	}

	.cancel-confirmation,
	.review-confirmation {
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		border: 1px solid rgba(239, 68, 68, 0.2);
	}

	.confirmation-text {
		font-size: var(--text-sm);
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
		text-align: center;
		font-weight: 500;
	}

	.confirmation-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: center;
	}

	.confirmation-actions .btn {
		flex: 1;
		max-width: 120px;
	}

	.btn-review {
		width: 100%;
		background: rgba(16, 185, 129, 0.1);
		color: rgb(16, 185, 129);
		border: 1px solid rgba(16, 185, 129, 0.2);
		transition: all var(--transition-normal);
		height: 40px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		font-size: var(--text-sm);
		font-weight: 500;
	}

	.btn-review svg {
		width: 16px;
		height: 16px;
		flex-shrink: 0;
	}

	.btn-review:hover {
		background: rgba(16, 185, 129, 0.2);
		border-color: rgba(16, 185, 129, 0.3);
		transform: translateY(-1px);
	}

	/* Cancel Advertiser Section */
	.cancel-advertiser-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.btn-cancel-advertiser {
		width: 100%;
		background: rgba(239, 68, 68, 0.1);
		color: rgb(239, 68, 68);
		border: 1px solid rgba(239, 68, 68, 0.2);
		transition: all var(--transition-normal);
		height: 48px;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-sm);
		font-size: var(--text-base);
		font-weight: 600;
		border-radius: var(--radius-lg);
	}

	.btn-cancel-advertiser svg {
		width: 20px;
		height: 20px;
		flex-shrink: 0;
	}

	.btn-cancel-advertiser:hover {
		background: rgba(239, 68, 68, 0.2);
		border-color: rgba(239, 68, 68, 0.3);
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
	}

	/* Button Variants */
	.btn-success {
		background: var(--success);
		color: var(--white);
	}

	.btn-success:hover {
		background: #00b894;
	}

	.btn-error {
		background: var(--error);
		color: var(--white);
	}

	.btn-error:hover {
		background: #e55353;
	}

	.campaign-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(100px, 1fr));
		gap: var(--space-md);
	}

	.metric {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.metric label {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.metric span {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
	}

	.budget-amount {
		color: var(--primary) !important;
		font-weight: 700 !important;
	}

	/* Placements Section */
	.placements-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.placements-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.placements-table {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		overflow: hidden;
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.table-header {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr 1fr;
		background: rgba(255, 255, 255, 0.05);
		padding: var(--space-md) var(--space-lg);
		gap: var(--space-md);
	}

	.header-cell {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.table-row {
		display: grid;
		grid-template-columns: 2fr 1fr 1fr 1fr 1fr 1fr;
		padding: var(--space-md) var(--space-lg);
		gap: var(--space-md);
		border-top: 1px solid rgba(255, 255, 255, 0.05);
	}

	.table-cell {
		font-size: var(--text-sm);
		color: var(--text-primary);
		display: flex;
		align-items: center;
	}

	.placement-info {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.placement-info .name {
		font-weight: 500;
	}

	.placement-info .ads-count {
		font-size: var(--text-xs);
		color: var(--text-secondary);
	}

	.revenue-cell {
		color: var(--success);
		font-weight: 600;
	}

	/* Account Info Section */
	.account-info-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.account-info-section h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: var(--space-lg);
	}

	.info-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.info-card h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-md) 0;
	}

	.info-items {
		display: flex;
		flex-direction: column;
		gap: var(--space-md);
	}

	.info-item {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
	}

	.info-item label {
		font-size: var(--text-xs);
		font-weight: 600;
		color: var(--text-secondary);
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.info-item span {
		font-size: var(--text-sm);
		color: var(--text-primary);
		font-weight: 500;
	}

	/* Status Badges */
	.status-badge {
		display: inline-block;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	.status-approved {
		background: rgba(16, 185, 129, 0.1);
		color: rgb(16, 185, 129);
	}

	.status-pending {
		background: rgba(245, 158, 11, 0.1);
		color: rgb(245, 158, 11);
	}

	.status-rejected {
		background: rgba(239, 68, 68, 0.1);
		color: rgb(239, 68, 68);
	}

	.status-cancelled {
		background: rgba(239, 68, 68, 0.1);
		color: rgb(239, 68, 68);
	}

	.status-paused {
		background: rgba(107, 114, 128, 0.1);
		color: rgb(107, 114, 128);
	}

	.status-completed {
		background: rgba(59, 130, 246, 0.1);
		color: rgb(59, 130, 246);
	}

	.status-default {
		background: rgba(107, 114, 128, 0.1);
		color: rgb(107, 114, 128);
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		.advertiser-detail {
			padding: var(--space-lg);
		}

		.advertiser-header {
			flex-direction: column;
			align-items: center;
			text-align: center;
		}

		.analytics-grid {
			grid-template-columns: repeat(2, 1fr);
		}

		.campaigns-grid {
			grid-template-columns: 1fr;
		}

		.info-grid {
			grid-template-columns: 1fr;
		}

		.placements-table {
			overflow-x: auto;
		}

		.header-cell,
		.table-cell {
			min-width: 100px;
		}
	}

	/* Accordion Styles */
	.accordion-section {
		margin-bottom: var(--space-lg);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.05);
		overflow: hidden;
		transition: all var(--transition-normal);
	}

	.accordion-section:hover {
		border-color: rgba(255, 255, 255, 0.1);
		transform: translateY(-1px);
		box-shadow: var(--shadow-lg);
	}

	.accordion-header {
		width: 100%;
		padding: var(--space-lg);
		background: none;
		border: none;
		display: flex;
		align-items: center;
		justify-content: space-between;
		cursor: pointer;
		transition: all var(--transition-normal);
		color: var(--text-primary);
		position: relative;
		overflow: hidden;
	}

	.accordion-header::before {
		content: '';
		position: absolute;
		top: 0;
		left: -100%;
		width: 100%;
		height: 100%;
		background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.03), transparent);
		transition: left 0.6s ease;
	}

	.accordion-header:hover::before {
		left: 100%;
	}

	.accordion-header:hover {
		background: rgba(255, 255, 255, 0.02);
		transform: scale(1.01);
	}

	.accordion-header.active {
		background: rgba(255, 255, 255, 0.05);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
	}

	.accordion-header.active:hover {
		background: rgba(255, 255, 255, 0.07);
	}

	.accordion-title {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		font-weight: 600;
		font-size: var(--text-lg);
		transition: all var(--transition-normal);
	}

	.accordion-header:hover .accordion-title {
		transform: translateX(2px);
	}

	.accordion-icon {
		width: 24px;
		height: 24px;
		color: var(--primary);
		flex-shrink: 0;
		transition: all var(--transition-normal);
	}

	.accordion-header:hover .accordion-icon {
		transform: scale(1.1);
		filter: brightness(1.2);
	}

	.accordion-count {
		background: var(--primary-gradient);
		color: var(--white);
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		min-width: 24px;
		text-align: center;
		transition: all var(--transition-normal);
		box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
	}

	.accordion-header:hover .accordion-count {
		transform: scale(1.05);
		box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
	}

	.accordion-chevron {
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		flex-shrink: 0;
	}

	.accordion-chevron.rotated {
		transform: rotate(180deg);
		color: var(--primary);
	}

	.accordion-header:hover .accordion-chevron {
		color: var(--text-primary);
		transform: scale(1.1);
	}

	.accordion-header:hover .accordion-chevron.rotated {
		transform: rotate(180deg) scale(1.1);
	}

	.accordion-content {
		max-height: 0;
		overflow: hidden;
		transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
		opacity: 0;
		transform: translateY(-10px);
		background: var(--bg-glass);
	}

	.accordion-section:has(.accordion-header.active) .accordion-content {
		max-height: 2000px;
		opacity: 1;
		transform: translateY(0);
		padding: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.05);
	}

	/* Fallback for browsers that don't support :has() */
	.accordion-content.active {
		max-height: 2000px;
		opacity: 1;
		transform: translateY(0);
		padding: var(--space-lg);
		border-top: 1px solid rgba(255, 255, 255, 0.05);
	}

	/* Stagger animation for accordion items */
	.accordion-content .campaigns-grid > * {
		animation: slideInUp 0.3s ease-out backwards;
	}

	.accordion-content .campaigns-grid > *:nth-child(1) {
		animation-delay: 0.1s;
	}

	.accordion-content .campaigns-grid > *:nth-child(2) {
		animation-delay: 0.15s;
	}

	.accordion-content .campaigns-grid > *:nth-child(3) {
		animation-delay: 0.2s;
	}

	.accordion-content .campaigns-grid > *:nth-child(4) {
		animation-delay: 0.25s;
	}

	.accordion-content .campaigns-grid > *:nth-child(n+5) {
		animation-delay: 0.3s;
	}

	@keyframes slideInUp {
		from {
			opacity: 0;
			transform: translateY(20px);
		}
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* Empty State */
	.empty-state {
		text-align: center;
		padding: var(--space-4xl);
	}

	.empty-icon {
		width: 80px;
		height: 80px;
		background: var(--bg-glass-dark);
		border-radius: var(--radius-xl);
		display: flex;
		align-items: center;
		justify-content: center;
		margin: 0 auto var(--space-lg);
		color: var(--text-secondary);
	}

	.empty-icon svg {
		width: 40px;
		height: 40px;
	}

	.empty-state h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.empty-state p {
		color: var(--text-secondary);
		margin: 0;
	}
</style> 