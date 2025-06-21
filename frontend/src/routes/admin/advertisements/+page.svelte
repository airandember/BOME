<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/auth';
	import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
	import type { AdvertiserAccount, AdCampaign } from '$lib/types/advertising';
	
	let advertiserAccounts: AdvertiserAccount[] = [];
	let campaigns: AdCampaign[] = [];
	let loading = true;
	let error: string | null = null;
	let activeTab: 'advertisers' | 'campaigns' = 'advertisers';
	let searchQuery = '';
	
	// Accordion state management
	let accordionState = {
		advertisers: {
			pending: true,
			approved: false,
			rejected: false,
			cancelled: false
		},
		campaigns: {
			pending: true,
			approved: false,
			rejected: false,
			cancelled: false
		}
	};
	
	// Action state management
	let cancellingCampaignId: number | null = null;
	let reviewingCampaignId: number | null = null;
	let reviewingAccountId: number | null = null;
	let cancellingAccountId: number | null = null;
	let reviewingCancelledCampaignId: number | null = null;
	let reviewingCancelledAccountId: number | null = null;

	// Mock admin data for displaying admin names
	const mockAdmins: { [key: number]: { name: string; email: string } } = {
		1: { name: 'John Smith', email: 'john.smith@bome.com' },
		2: { name: 'Sarah Johnson', email: 'sarah.johnson@bome.com' },
		3: { name: 'Michael Chen', email: 'michael.chen@bome.com' }
	};

	// Computed filtered data
	$: filteredAdvertiserAccounts = filterAdvertiserAccounts(advertiserAccounts, searchQuery);
	$: filteredCampaigns = filterCampaigns(campaigns, searchQuery);

	// Status counts
	$: advertiserCounts = getAdvertiserCounts(filteredAdvertiserAccounts);
	$: campaignCounts = getCampaignCounts(filteredCampaigns);

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
			await loadData();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			loading = false;
		}
	});

	async function loadData() {
		// Always load advertiser accounts first to have them available for campaign display
		await loadAdvertisers();
		await loadCampaigns();
	}

	async function loadAdvertisers() {
		// Mock data for demonstration - expanded to include more examples
		advertiserAccounts = [
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
				status: 'pending',
				created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
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
				status: 'approved',
				approved_by: 1,
				approved_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 3,
				user_id: 4,
				company_name: 'Global Ventures LLC',
				business_email: 'info@globalventures.com',
				contact_name: 'Emma Rodriguez',
				contact_phone: '(555) 456-7890',
				business_address: '789 Business Park, Enterprise City, EC 13579',
				tax_id: '45-6789012',
				website: 'https://globalventures.com',
				industry: 'Finance',
				status: 'rejected',
				verification_notes: 'Incomplete documentation provided',
				rejected_by: 2,
				rejected_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 4,
				user_id: 5,
				company_name: 'Creative Studios Inc',
				business_email: 'contact@creativestudios.com',
				contact_name: 'David Kim',
				contact_phone: '(555) 789-0123',
				business_address: '321 Creative Ave, Art District, AD 24680',
				tax_id: '67-8901234',
				website: 'https://creativestudios.com',
				industry: 'Design',
				status: 'cancelled',
				cancelled_by: 1,
				cancelled_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 10 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 5,
				user_id: 6,
				company_name: 'NextGen Solutions',
				business_email: 'hello@nextgen.com',
				contact_name: 'Lisa Wang',
				contact_phone: '(555) 234-5678',
				business_address: '654 Future Blvd, Innovation City, IC 97531',
				tax_id: '89-0123456',
				website: 'https://nextgen.com',
				industry: 'Technology',
				status: 'pending',
				created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			}
		];
	}

	async function loadCampaigns() {
		// Mock data for demonstration - expanded to include more examples
		campaigns = [
			{
				id: 1,
				advertiser_id: 1,
				name: 'Q1 Product Launch Campaign',
				description: 'Promoting our revolutionary new software platform to tech professionals',
				status: 'pending',
				start_date: '2024-03-01',
				end_date: '2024-05-31',
				budget: 15000,
				spent_amount: 0,
				target_audience: 'Tech professionals, developers, and IT decision makers aged 25-45',
				billing_type: 'monthly',
				billing_rate: 5000,
				image_url: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
				created_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 2,
				advertiser_id: 2,
				name: 'Summer Brand Awareness',
				description: 'Building brand recognition and driving website traffic during peak season',
				status: 'approved',
				start_date: '2024-06-01',
				end_date: '2024-08-31',
				budget: 8000,
				spent_amount: 2400,
				target_audience: 'Small business owners and marketing professionals',
				billing_type: 'weekly',
				billing_rate: 666.67,
				image_url: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
				approved_by: 1,
				approved_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 4 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 3,
				advertiser_id: 3,
				name: 'Holiday Promotion Campaign',
				description: 'Seasonal campaign targeting holiday shoppers with special offers',
				status: 'rejected',
				start_date: '2024-11-01',
				end_date: '2024-12-31',
				budget: 12000,
				spent_amount: 0,
				target_audience: 'Holiday shoppers, families, and gift buyers aged 30-55',
				billing_type: 'monthly',
				billing_rate: 6000,
				image_url: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
				approval_notes: 'Campaign content does not align with platform guidelines',
				rejected_by: 2,
				rejected_at: new Date(Date.now() - 1 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 4,
				advertiser_id: 4,
				name: 'Creative Portfolio Showcase',
				description: 'Showcasing our design portfolio to attract new clients',
				status: 'cancelled',
				start_date: '2024-04-01',
				end_date: '2024-06-30',
				budget: 5000,
				spent_amount: 1200,
				target_audience: 'Small businesses and startups needing design services',
				billing_type: 'monthly',
				billing_rate: 1666.67,
				image_url: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
				cancelled_by: 1,
				cancelled_at: new Date(Date.now() - 3 * 24 * 60 * 60 * 1000).toISOString(),
				created_at: new Date(Date.now() - 6 * 24 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			},
			{
				id: 5,
				advertiser_id: 5,
				name: 'AI Innovation Summit',
				description: 'Promoting our participation in the upcoming AI innovation conference',
				status: 'pending',
				start_date: '2024-07-01',
				end_date: '2024-07-31',
				budget: 3000,
				spent_amount: 0,
				target_audience: 'Tech enthusiasts, AI researchers, and industry professionals',
				billing_type: 'monthly',
				billing_rate: 3000,
				image_url: '/src/lib/HOMEPAGE_TEST_ASSETS/16X10_Placeholder_IMG.png',
				created_at: new Date(Date.now() - 12 * 60 * 60 * 1000).toISOString(),
				updated_at: new Date().toISOString()
			}
		];
	}

	// Filter functions
	function filterAdvertiserAccounts(accounts: AdvertiserAccount[], query: string): AdvertiserAccount[] {
		if (!query.trim()) return accounts;
		
		const lowerQuery = query.toLowerCase();
		return accounts.filter(account => 
			account.company_name.toLowerCase().includes(lowerQuery) ||
			account.contact_name.toLowerCase().includes(lowerQuery) ||
			account.business_email.toLowerCase().includes(lowerQuery) ||
			(account.industry && account.industry.toLowerCase().includes(lowerQuery))
		);
	}

	function filterCampaigns(campaigns: AdCampaign[], query: string): AdCampaign[] {
		if (!query.trim()) return campaigns;
		
		const lowerQuery = query.toLowerCase();
		return campaigns.filter(campaign => 
			campaign.name.toLowerCase().includes(lowerQuery) ||
			(campaign.description && campaign.description.toLowerCase().includes(lowerQuery)) ||
			(campaign.target_audience && campaign.target_audience.toLowerCase().includes(lowerQuery))
		);
	}

	// Count functions
	function getAdvertiserCounts(accounts: AdvertiserAccount[]) {
		return {
			pending: accounts.filter(acc => acc.status === 'pending').length,
			approved: accounts.filter(acc => acc.status === 'approved').length,
			rejected: accounts.filter(acc => acc.status === 'rejected').length,
			cancelled: accounts.filter(acc => acc.status === 'cancelled').length
		};
	}

	function getCampaignCounts(campaigns: AdCampaign[]) {
		return {
			pending: campaigns.filter(camp => camp.status === 'pending').length,
			approved: campaigns.filter(camp => camp.status === 'approved' || camp.status === 'active').length,
			rejected: campaigns.filter(camp => camp.status === 'rejected').length,
			cancelled: campaigns.filter(camp => camp.status === 'cancelled').length
		};
	}

	// Accordion functions
	function toggleAccordion(section: 'advertisers' | 'campaigns', status: 'pending' | 'approved' | 'rejected' | 'cancelled') {
		accordionState[section][status] = !accordionState[section][status];
	}

	function getAdvertisersByStatus(status: string): AdvertiserAccount[] {
		return filteredAdvertiserAccounts.filter(acc => acc.status === status);
	}

	function getCampaignsByStatus(status: string): AdCampaign[] {
		if (status === 'approved') {
			return filteredCampaigns.filter(camp => camp.status === 'approved' || camp.status === 'active');
		}
		return filteredCampaigns.filter(camp => camp.status === status);
	}

	async function approveAdvertiser(advertiserId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			advertiserAccounts = advertiserAccounts.map(acc => 
				acc.id === advertiserId ? { ...acc, status: 'approved' } : acc
			);
		} catch (err) {
			error = 'Failed to approve advertiser';
		}
	}

	async function rejectAdvertiser(advertiserId: number) {
		const reason = prompt('Please provide a reason for rejection:');
		if (!reason) return;

		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			advertiserAccounts = advertiserAccounts.map(acc => 
				acc.id === advertiserId ? { ...acc, status: 'rejected', verification_notes: reason } : acc
			);
		} catch (err) {
			error = 'Failed to reject advertiser';
		}
	}

	async function approveCampaign(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { ...campaign, status: 'approved' } : campaign
			);
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
				campaign.id === campaignId ? { ...campaign, status: 'rejected', approval_notes: reason } : campaign
			);
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

	function showReviewConfirmation(campaignId: number) {
		reviewingCampaignId = campaignId;
	}

	function hideReviewConfirmation() {
		reviewingCampaignId = null;
	}

	async function confirmCancelCampaign(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { 
					...campaign, 
					status: 'cancelled',
					cancelled_by: 1, // Mock current admin ID
					cancelled_at: new Date().toISOString()
				} : campaign
			);
			
			cancellingCampaignId = null;
		} catch (err) {
			error = 'Failed to cancel campaign';
		}
	}

	async function confirmReviewApproval(campaignId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			campaigns = campaigns.map(campaign => 
				campaign.id === campaignId ? { 
					...campaign, 
					status: 'approved',
					approved_by: 1, // Mock current admin ID
					approved_at: new Date().toISOString(),
					// Clear rejection data
					rejected_by: undefined,
					rejected_at: undefined,
					approval_notes: 'Campaign approved after review'
				} : campaign
			);
			
			reviewingCampaignId = null;
		} catch (err) {
			error = 'Failed to approve campaign';
		}
	}

	function showAccountReviewConfirmation(accountId: number) {
		reviewingAccountId = accountId;
	}

	function hideAccountReviewConfirmation() {
		reviewingAccountId = null;
	}

	async function confirmAccountReviewApproval(accountId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			advertiserAccounts = advertiserAccounts.map(account => 
				account.id === accountId ? { 
					...account, 
					status: 'approved',
					approved_by: 1, // Mock current admin ID
					approved_at: new Date().toISOString(),
					// Clear rejection data
					rejected_by: undefined,
					rejected_at: undefined,
					verification_notes: 'Account approved after review'
				} : account
			);
			
			reviewingAccountId = null;
		} catch (err) {
			error = 'Failed to approve account';
		}
	}

	function showAccountCancelConfirmation(accountId: number) {
		cancellingAccountId = accountId;
	}

	function hideAccountCancelConfirmation() {
		cancellingAccountId = null;
	}

	async function confirmAccountCancellation(accountId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			advertiserAccounts = advertiserAccounts.map(account => 
				account.id === accountId ? { 
					...account, 
					status: 'cancelled',
					cancelled_by: 1, // Mock current admin ID
					cancelled_at: new Date().toISOString()
				} : account
			);
			
			cancellingAccountId = null;
		} catch (err) {
			error = 'Failed to cancel account';
		}
	}

	// Functions for reviewing cancelled campaigns
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
					status: 'approved',
					approved_by: 1, // Mock current admin ID
					approved_at: new Date().toISOString(),
					// Clear cancellation data
					cancelled_by: undefined,
					cancelled_at: undefined,
					approval_notes: 'Campaign reactivated after review'
				} : campaign
			);
			
			reviewingCancelledCampaignId = null;
		} catch (err) {
			error = 'Failed to reactivate campaign';
		}
	}

	// Functions for reviewing cancelled accounts
	function showCancelledAccountReviewConfirmation(accountId: number) {
		reviewingCancelledAccountId = accountId;
	}

	function hideCancelledAccountReviewConfirmation() {
		reviewingCancelledAccountId = null;
	}

	async function confirmCancelledAccountReactivation(accountId: number) {
		try {
			// Mock API call
			await new Promise(resolve => setTimeout(resolve, 500));
			
			advertiserAccounts = advertiserAccounts.map(account => 
				account.id === accountId ? { 
					...account, 
					status: 'approved',
					approved_by: 1, // Mock current admin ID
					approved_at: new Date().toISOString(),
					// Clear cancellation data
					cancelled_by: undefined,
					cancelled_at: undefined,
					verification_notes: 'Account reactivated after review'
				} : account
			);
			
			reviewingCancelledAccountId = null;
		} catch (err) {
			error = 'Failed to reactivate account';
		}
	}

	function switchTab(tab: 'advertisers' | 'campaigns') {
		activeTab = tab;
	}

	function getStatusBadgeClass(status: string): string {
		switch (status) {
			case 'approved':
				return 'status-approved';
			case 'pending':
				return 'status-pending';
			case 'rejected':
				return 'status-rejected';
			case 'active':
				return 'status-active';
			case 'paused':
				return 'status-paused';
			case 'cancelled':
				return 'status-cancelled';
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

	function getTimeAgo(dateString: string): string {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = Math.floor((now.getTime() - date.getTime()) / (1000 * 60 * 60));
		
		if (diffInHours < 24) {
			return `${diffInHours} hours ago`;
		} else {
			const diffInDays = Math.floor(diffInHours / 24);
			return `${diffInDays} days ago`;
		}
	}

	function getAdvertiserInfo(advertiserId: number): AdvertiserAccount | null {
		return advertiserAccounts.find(acc => acc.id === advertiserId) || null;
	}

	function getAdminName(adminId: number): string {
		return mockAdmins[adminId]?.name || 'Unknown Admin';
	}

	function getStatusInfo(campaign: AdCampaign): { text: string; date: string; admin: string } | null {
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

	function getAccountStatusInfo(account: AdvertiserAccount): { text: string; date: string; admin: string } | null {
		switch (account.status) {
			case 'approved':
				return {
					text: 'Approved',
					date: account.approved_at ? formatDateTime(account.approved_at) : '',
					admin: account.approved_by ? getAdminName(account.approved_by) : ''
				};
			case 'rejected':
				return {
					text: 'Rejected',
					date: account.rejected_at ? formatDateTime(account.rejected_at) : '',
					admin: account.rejected_by ? getAdminName(account.rejected_by) : ''
				};
			case 'cancelled':
				return {
					text: 'Cancelled',
					date: account.cancelled_at ? formatDateTime(account.cancelled_at) : '',
					admin: account.cancelled_by ? getAdminName(account.cancelled_by) : ''
				};
			default:
				return null;
		}
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
</script>

<svelte:head>
	<title>Advertisement Management - BOME Admin</title>
</svelte:head>

<div class="advertisement-management">
	<div class="page-header">
		<div class="header-content">
			<div>
				<h1>Advertisement Management</h1>
				<p>Manage advertiser accounts and campaign approvals</p>
			</div>
			<div class="header-actions">
				<button class="btn btn-secondary" on:click={loadData}>
					<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
						<path d="M3 12a9 9 0 0 1 9-9 9.75 9.75 0 0 1 6.74 2.74L21 8" />
						<path d="M21 3v5h-5" />
						<path d="M21 12a9 9 0 0 1-9 9 9.75 9.75 0 0 1-6.74-2.74L3 16" />
						<path d="M3 21v-5h5" />
					</svg>
					Refresh
				</button>
			</div>
		</div>
	</div>

	{#if error}
		<div class="error-message">
			<div class="alert alert-error">
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126zM12 15.75h.007v.008H12v-.008z" />
				</svg>
				<span>{error}</span>
			</div>
		</div>
	{/if}

	<!-- Tab Navigation -->
	<div class="tab-navigation">
		<div class="tab-container">
			<button
				class="tab-button {activeTab === 'advertisers' ? 'active' : ''}"
				on:click={() => switchTab('advertisers')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
					<circle cx="9" cy="7" r="4" />
					<path d="M23 21v-2a4 4 0 0 0-3-3.87" />
					<path d="M16 3.13a4 4 0 0 1 0 7.75" />
				</svg>
				Advertiser Accounts
				<span class="tab-count">{advertiserCounts.pending}</span>
			</button>
			<button
				class="tab-button {activeTab === 'campaigns' ? 'active' : ''}"
				on:click={() => switchTab('campaigns')}
			>
				<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
					<line x1="8" y1="21" x2="16" y2="21" />
					<line x1="12" y1="17" x2="12" y2="21" />
				</svg>
				Campaign Approvals
				<span class="tab-count">{campaignCounts.pending}</span>
			</button>
		</div>
	</div>

	<!-- Search Section -->
	<div class="search-section">
		<div class="search-container">
			<div class="search-input-wrapper">
				<svg class="search-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<circle cx="11" cy="11" r="8" />
					<path d="m21 21-4.35-4.35" />
				</svg>
				<input
					type="text"
					class="search-input"
					placeholder="Search {activeTab === 'advertisers' ? 'advertiser accounts' : 'campaigns'}..."
					bind:value={searchQuery}
				/>
				{#if searchQuery}
					<button class="clear-search" on:click={() => searchQuery = ''}>
						<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<line x1="18" y1="6" x2="6" y2="18" />
							<line x1="6" y1="6" x2="18" y2="18" />
						</svg>
					</button>
				{/if}
			</div>
		</div>
	</div>

	{#if loading}
		<div class="loading-container">
			<LoadingSpinner size="large" color="primary" />
			<p>Loading {activeTab}...</p>
		</div>
	{:else if activeTab === 'advertisers'}
		<!-- Advertiser Accounts Section -->
		<div class="content-section">
			<div class="section-header">
				<h2>Advertiser Accounts</h2>
				<p>Review and approve advertiser account applications</p>
			</div>

			{#if searchQuery.trim()}
				<!-- Search Results - No Accordions -->
				<div class="search-results">
					<div class="search-results-header">
						<h3>Search Results for "{searchQuery}"</h3>
						<p>Found {filteredAdvertiserAccounts.length} advertiser accounts</p>
					</div>
					
					{#if filteredAdvertiserAccounts.length === 0}
						<div class="empty-state">
							<div class="empty-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="11" cy="11" r="8" />
									<path d="m21 21-4.35-4.35" />
								</svg>
							</div>
							<h3>No results found</h3>
							<p>No advertiser accounts match your search criteria.</p>
						</div>
					{:else}
						<div class="accounts-grid">
							{#each filteredAdvertiserAccounts as account}
								{@const accountStatusInfo = getAccountStatusInfo(account)}
								<div class="account-card">
									<div class="card-header">
										<div class="company-info">
											<div class="company-avatar">
												{account.company_name.charAt(0)}
											</div>
											<div class="company-details">
												<h3>{account.company_name}</h3>
												<div class="company-meta">
													<span class="status-badge {getStatusBadgeClass(account.status)}">
														{account.status}
													</span>
													{#if accountStatusInfo}
														<div class="status-info">
															<span class="status-action">{accountStatusInfo.text} by {accountStatusInfo.admin}</span>
															<span class="status-date">{accountStatusInfo.date}</span>
														</div>
													{/if}
												</div>
											</div>
										</div>
										{#if account.status !== 'pending'}
											<div class="view-details-btn">
												<button
													class="btn btn-sm btn-secondary"
													on:click={() => goto(`/admin/advertisers/${account.id}`)}
												>
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<path d="M9 18l6-6-6-6" />
													</svg>
													View Details
												</button>
											</div>
										{/if}
									</div>
									
									<div class="card-content">
										<div class="info-grid">
											<div class="info-item">
												<label>Contact</label>
												<span>{account.contact_name}</span>
											</div>
											<div class="info-item">
												<label>Email</label>
												<span>{account.business_email}</span>
											</div>
											<div class="info-item">
												<label>Phone</label>
												<span>{account.contact_phone || 'Not provided'}</span>
											</div>
											<div class="info-item">
												<label>Industry</label>
												<span>{account.industry || 'Not specified'}</span>
											</div>
											<div class="info-item">
												<label>Website</label>
												<span>{account.website || 'Not provided'}</span>
											</div>
											<div class="info-item">
												<label>Applied</label>
												<span>{getTimeAgo(account.created_at)}</span>
											</div>
										</div>
										
										{#if account.verification_notes}
											<div class="notes-section">
												<label>Notes</label>
												<p>{account.verification_notes}</p>
											</div>
										{/if}
										
										<!-- Action buttons based on status -->
										{#if account.status === 'pending'}
											<div class="card-actions">
												<button
													class="btn btn-sm btn-success"
													on:click={() => approveAdvertiser(account.id)}
												>
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<path d="M9 12l2 2 4-4" />
													</svg>
													Approve
												</button>
												<button
													class="btn btn-sm btn-error"
													on:click={() => rejectAdvertiser(account.id)}
												>
													<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
														<path d="M18 6L6 18M6 6l12 12" />
													</svg>
													Reject
												</button>
											</div>
										{:else if account.status === 'approved'}
											<div class="cancel-section">
												{#if cancellingAccountId === account.id}
													<div class="cancel-confirmation">
														<p class="confirmation-text">Are you sure you want to cancel this advertiser account?</p>
														<div class="confirmation-actions">
															<button
																class="btn btn-sm btn-error"
																on:click={() => confirmAccountCancellation(account.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M9 12l2 2 4-4" />
																</svg>
																Confirm
															</button>
															<button
																class="btn btn-sm btn-secondary"
																on:click={hideAccountCancelConfirmation}
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
														on:click={() => showAccountCancelConfirmation(account.id)}
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
														</svg>
														Cancel Advertiser
													</button>
												{/if}
											</div>
										{:else if account.status === 'rejected'}
											<div class="cancel-section">
												{#if reviewingAccountId === account.id}
													<div class="review-confirmation">
														<p class="confirmation-text">Would you like to approve this advertiser account?</p>
														<div class="confirmation-actions">
															<button
																class="btn btn-sm btn-success"
																on:click={() => confirmAccountReviewApproval(account.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M9 12l2 2 4-4" />
																</svg>
																Approve
															</button>
															<button
																class="btn btn-sm btn-secondary"
																on:click={hideAccountReviewConfirmation}
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
														on:click={() => showAccountReviewConfirmation(account.id)}
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<circle cx="11" cy="11" r="8" />
															<path d="m21 21-4.35-4.35" />
														</svg>
														Review Account
													</button>
												{/if}
											</div>
										{:else if account.status === 'cancelled'}
											<div class="cancel-section">
												{#if reviewingCancelledAccountId === account.id}
													<div class="review-confirmation">
														<p class="confirmation-text">Would you like to reactivate this advertiser account?</p>
														<div class="confirmation-actions">
															<button
																class="btn btn-sm btn-success"
																on:click={() => confirmCancelledAccountReactivation(account.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M9 12l2 2 4-4" />
																</svg>
																Reactivate
															</button>
															<button
																class="btn btn-sm btn-secondary"
																on:click={hideCancelledAccountReviewConfirmation}
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
														on:click={() => showCancelledAccountReviewConfirmation(account.id)}
													>
														<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
															<circle cx="11" cy="11" r="8" />
															<path d="m21 21-4.35-4.35" />
														</svg>
														Review Account
													</button>
												{/if}
											</div>
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{:else}
				<!-- Advertiser Accounts Accordion -->
				<div class="content-section">
					<div class="section-header">
						<h2>Advertiser Accounts</h2>
						<p>Review and approve advertiser account applications</p>
					</div>

					<!-- Pending Advertisers Accordion -->
					<div class="accordion-section">
						<button 
							class="accordion-header {accordionState.advertisers.pending ? 'active' : ''}"
							on:click={() => toggleAccordion('advertisers', 'pending')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10" />
									<path d="M12 6v6l4 2" />
								</svg>
								<span>Pending Approval</span>
								<span class="accordion-count">{advertiserCounts.pending}</span>
							</div>
							<svg class="accordion-chevron {accordionState.advertisers.pending ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.advertisers.pending}
							{@const pendingAdvertisers = getAdvertisersByStatus('pending')}
							<div class="accordion-content">
								{#if pendingAdvertisers.length === 0}
									<div class="empty-state">
										<div class="empty-icon">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<circle cx="12" cy="12" r="10" />
												<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
												<path d="M12 17h.01" />
											</svg>
										</div>
										<h3>No pending advertisers</h3>
										<p>{searchQuery ? 'No pending advertisers match your search.' : 'No advertiser accounts pending approval.'}</p>
									</div>
								{:else}
									<div class="accounts-grid">
										{#each pendingAdvertisers as account}
											<div class="account-card">
												<div class="card-header">
													<div class="company-info">
														<div class="company-avatar">
															{account.company_name.charAt(0)}
														</div>
														<div class="company-details">
															<h3>{account.company_name}</h3>
															<div class="company-meta">
																<span class="status-badge {getStatusBadgeClass(account.status)}">
																	{account.status}
																</span>
															</div>
															<div class="card-actions">
																<button
																	class="btn btn-sm btn-success"
																	on:click={() => approveAdvertiser(account.id)}
																>
																	<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																		<path d="M9 12l2 2 4-4" />
																	</svg>
																	Approve
																</button>
																<button
																	class="btn btn-sm btn-error"
																	on:click={() => rejectAdvertiser(account.id)}
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
												
												<div class="card-content">
													<div class="info-grid">
														<div class="info-item">
															<label>Contact</label>
															<span>{account.contact_name}</span>
														</div>
														<div class="info-item">
															<label>Email</label>
															<span>{account.business_email}</span>
														</div>
														<div class="info-item">
															<label>Phone</label>
															<span>{account.contact_phone || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Industry</label>
															<span>{account.industry || 'Not specified'}</span>
														</div>
														<div class="info-item">
															<label>Website</label>
															<span>{account.website || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Applied</label>
															<span>{getTimeAgo(account.created_at)}</span>
														</div>
													</div>
												</div>
											</div>
										{/each}
									</div>
								{/if}
							</div>
						{/if}
					</div>

					<!-- Approved Advertisers Accordion -->
					<div class="accordion-section">
						<button 
							class="accordion-header {accordionState.advertisers.approved ? 'active' : ''}"
							on:click={() => toggleAccordion('advertisers', 'approved')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 12l2 2 4-4" />
									<circle cx="12" cy="12" r="10" />
								</svg>
								<span>Approved</span>
								<span class="accordion-count">{advertiserCounts.approved}</span>
							</div>
							<svg class="accordion-chevron {accordionState.advertisers.approved ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.advertisers.approved}
							{@const approvedAdvertisers = getAdvertisersByStatus('approved')}
							<div class="accordion-content">
								{#if approvedAdvertisers.length === 0}
									<div class="empty-state">
										<div class="empty-icon">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<circle cx="12" cy="12" r="10" />
												<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
												<path d="M12 17h.01" />
											</svg>
										</div>
										<h3>No approved advertisers</h3>
										<p>{searchQuery ? 'No approved advertisers match your search.' : 'No approved advertiser accounts.'}</p>
									</div>
								{:else}
									<div class="accounts-grid">
										{#each approvedAdvertisers as account}
											{@const accountStatusInfo = getAccountStatusInfo(account)}
											<div class="account-card">
												<div class="card-header">
													<div class="company-info">
														<div class="company-avatar">
															{account.company_name.charAt(0)}
														</div>
														<div class="company-details">
															<h3>{account.company_name}</h3>
															<div class="company-meta">
																<span class="status-badge {getStatusBadgeClass(account.status)}">
																	{account.status}
																</span>
																{#if accountStatusInfo}
																	<div class="status-info">
																		<span class="status-action">{accountStatusInfo.text} by {accountStatusInfo.admin}</span>
																		<span class="status-date">{accountStatusInfo.date}</span>
																	</div>
																{/if}
															</div>
														</div>
													</div>
													<div class="view-details-btn">
														<button
															class="btn btn-sm btn-secondary"
															on:click={() => goto(`/admin/advertisers/${account.id}`)}
														>
															<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																<path d="M9 18l6-6-6-6" />
															</svg>
															View Details
														</button>
													</div>
												</div>
												
												<div class="card-content">
													<div class="info-grid">
														<div class="info-item">
															<label>Contact</label>
															<span>{account.contact_name}</span>
														</div>
														<div class="info-item">
															<label>Email</label>
															<span>{account.business_email}</span>
														</div>
														<div class="info-item">
															<label>Phone</label>
															<span>{account.contact_phone || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Industry</label>
															<span>{account.industry || 'Not specified'}</span>
														</div>
														<div class="info-item">
															<label>Website</label>
															<span>{account.website || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Applied</label>
															<span>{getTimeAgo(account.created_at)}</span>
														</div>
													</div>
													
													<!-- Cancel button for approved accounts -->
													<div class="cancel-section">
														{#if cancellingAccountId === account.id}
															<div class="cancel-confirmation">
																<p class="confirmation-text">Are you sure you want to cancel this advertiser account?</p>
																<div class="confirmation-actions">
																	<button
																		class="btn btn-sm btn-error"
																		on:click={() => confirmAccountCancellation(account.id)}
																	>
																		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																			<path d="M9 12l2 2 4-4" />
																		</svg>
																		Confirm
																	</button>
																	<button
																		class="btn btn-sm btn-secondary"
																		on:click={hideAccountCancelConfirmation}
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
																on:click={() => showAccountCancelConfirmation(account.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
																</svg>
																Cancel Advertiser
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

					<!-- Rejected Advertisers Accordion -->
					<div class="accordion-section">
						<button 
							class="accordion-header {accordionState.advertisers.rejected ? 'active' : ''}"
							on:click={() => toggleAccordion('advertisers', 'rejected')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10" />
									<line x1="15" y1="9" x2="9" y2="15" />
									<line x1="9" y1="9" x2="15" y2="15" />
								</svg>
								<span>Rejected</span>
								<span class="accordion-count">{advertiserCounts.rejected}</span>
							</div>
							<svg class="accordion-chevron {accordionState.advertisers.rejected ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.advertisers.rejected}
							{@const rejectedAdvertisers = getAdvertisersByStatus('rejected')}
							<div class="accordion-content">
								{#if rejectedAdvertisers.length === 0}
									<div class="empty-state">
										<div class="empty-icon">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<circle cx="12" cy="12" r="10" />
												<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
												<path d="M12 17h.01" />
											</svg>
										</div>
										<h3>No rejected advertisers</h3>
										<p>{searchQuery ? 'No rejected advertisers match your search.' : 'No rejected advertiser accounts.'}</p>
									</div>
								{:else}
									<div class="accounts-grid">
										{#each rejectedAdvertisers as account}
											{@const accountStatusInfo = getAccountStatusInfo(account)}
											<div class="account-card">
												<div class="card-header">
													<div class="company-info">
														<div class="company-avatar">
															{account.company_name.charAt(0)}
														</div>
														<div class="company-details">
															<h3>{account.company_name}</h3>
															<div class="company-meta">
																<span class="status-badge {getStatusBadgeClass(account.status)}">
																	{account.status}
																</span>
																{#if accountStatusInfo}
																	<div class="status-info">
																		<span class="status-action">{accountStatusInfo.text} by {accountStatusInfo.admin}</span>
																		<span class="status-date">{accountStatusInfo.date}</span>
																	</div>
																{/if}
															</div>
														</div>
													</div>
												</div>
												
												<div class="card-content">
													<div class="info-grid">
														<div class="info-item">
															<label>Contact</label>
															<span>{account.contact_name}</span>
														</div>
														<div class="info-item">
															<label>Email</label>
															<span>{account.business_email}</span>
														</div>
														<div class="info-item">
															<label>Phone</label>
															<span>{account.contact_phone || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Industry</label>
															<span>{account.industry || 'Not specified'}</span>
														</div>
														<div class="info-item">
															<label>Website</label>
															<span>{account.website || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Applied</label>
															<span>{getTimeAgo(account.created_at)}</span>
														</div>
													</div>
													
													{#if account.verification_notes}
														<div class="notes-section">
															<label>Notes</label>
															<p>{account.verification_notes}</p>
														</div>
													{/if}
													
													<!-- Review button for rejected accounts -->
													<div class="cancel-section">
														{#if reviewingAccountId === account.id}
															<div class="review-confirmation">
																<p class="confirmation-text">Would you like to approve this advertiser account?</p>
																<div class="confirmation-actions">
																	<button
																		class="btn btn-sm btn-success"
																		on:click={() => confirmAccountReviewApproval(account.id)}
																	>
																		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																			<path d="M9 12l2 2 4-4" />
																		</svg>
																		Approve
																	</button>
																	<button
																		class="btn btn-sm btn-secondary"
																		on:click={hideAccountReviewConfirmation}
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
																on:click={() => showAccountReviewConfirmation(account.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<circle cx="11" cy="11" r="8" />
																	<path d="m21 21-4.35-4.35" />
																</svg>
																Review Account
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

					<!-- Cancelled Advertisers Accordion -->
					<div class="accordion-section">
						<button 
							class="accordion-header {accordionState.advertisers.cancelled ? 'active' : ''}"
							on:click={() => toggleAccordion('advertisers', 'cancelled')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10" />
									<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
								</svg>
								<span>Cancelled</span>
								<span class="accordion-count">{advertiserCounts.cancelled}</span>
							</div>
							<svg class="accordion-chevron {accordionState.advertisers.cancelled ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.advertisers.cancelled}
							{@const cancelledAdvertisers = getAdvertisersByStatus('cancelled')}
							<div class="accordion-content">
								{#if cancelledAdvertisers.length === 0}
									<div class="empty-state">
										<div class="empty-icon">
											<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
												<circle cx="12" cy="12" r="10" />
												<path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
												<path d="M12 17h.01" />
											</svg>
										</div>
										<h3>No cancelled advertisers</h3>
										<p>{searchQuery ? 'No cancelled advertisers match your search.' : 'No cancelled advertiser accounts.'}</p>
									</div>
								{:else}
									<div class="accounts-grid">
										{#each cancelledAdvertisers as account}
											{@const accountStatusInfo = getAccountStatusInfo(account)}
											<div class="account-card">
												<div class="card-header">
													<div class="company-info">
														<div class="company-avatar">
															{account.company_name.charAt(0)}
														</div>
														<div class="company-details">
															<h3>{account.company_name}</h3>
															<div class="company-meta">
																<span class="status-badge {getStatusBadgeClass(account.status)}">
																	{account.status}
																</span>
																{#if accountStatusInfo}
																	<div class="status-info">
																		<span class="status-action">{accountStatusInfo.text} by {accountStatusInfo.admin}</span>
																		<span class="status-date">{accountStatusInfo.date}</span>
																	</div>
																{/if}
															</div>
														</div>
													</div>
												</div>
												
												<div class="card-content">
													<div class="info-grid">
														<div class="info-item">
															<label>Contact</label>
															<span>{account.contact_name}</span>
														</div>
														<div class="info-item">
															<label>Email</label>
															<span>{account.business_email}</span>
														</div>
														<div class="info-item">
															<label>Phone</label>
															<span>{account.contact_phone || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Industry</label>
															<span>{account.industry || 'Not specified'}</span>
														</div>
														<div class="info-item">
															<label>Website</label>
															<span>{account.website || 'Not provided'}</span>
														</div>
														<div class="info-item">
															<label>Applied</label>
															<span>{getTimeAgo(account.created_at)}</span>
														</div>
													</div>
													
													<!-- Review button for cancelled accounts -->
													<div class="cancel-section">
														{#if reviewingCancelledAccountId === account.id}
															<div class="review-confirmation">
																<p class="confirmation-text">Would you like to reactivate this advertiser account?</p>
																<div class="confirmation-actions">
																	<button
																		class="btn btn-sm btn-success"
																		on:click={() => confirmCancelledAccountReactivation(account.id)}
																	>
																		<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																			<path d="M9 12l2 2 4-4" />
																		</svg>
																		Reactivate
																	</button>
																	<button
																		class="btn btn-sm btn-secondary"
																		on:click={hideCancelledAccountReviewConfirmation}
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
																on:click={() => showCancelledAccountReviewConfirmation(account.id)}
															>
																<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
																	<circle cx="11" cy="11" r="8" />
																	<path d="m21 21-4.35-4.35" />
																</svg>
																Review Account
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
			{/if}
		</div>
	{:else}
		<!-- Campaign Approvals Section -->
		<div class="content-section">
			<div class="section-header">
				<h2>Campaign Approvals</h2>
				<p>Review and approve advertising campaigns</p>
			</div>

			{#if searchQuery.trim()}
				<!-- Search Results - No Accordions -->
				<div class="search-results">
					<div class="search-results-header">
						<h3>Search Results for "{searchQuery}"</h3>
						<p>Found {filteredCampaigns.length} campaigns</p>
					</div>
					
					{#if filteredCampaigns.length === 0}
						<div class="empty-state">
							<div class="empty-icon">
								<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="11" cy="11" r="8" />
									<path d="m21 21-4.35-4.35" />
								</svg>
							</div>
							<h3>No results found</h3>
							<p>No campaigns match your search criteria.</p>
						</div>
					{:else}
						<div class="campaigns-grid">
							{#each filteredCampaigns as campaign}
								{@const statusInfo = getStatusInfo(campaign)}
								{@const advertiserInfo = getAdvertiserInfo(campaign.advertiser_id)}
								<div class="campaign-card">
									<div class="card-header">
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
													{#if statusInfo}
														<div class="status-info">
															<span class="status-action">{statusInfo.text} by {statusInfo.admin}</span>
															<span class="status-date">{statusInfo.date}</span>
														</div>
													{/if}
												</div>
											</div>
										</div>
									</div>
									
									<div class="card-content">
										{#if campaign.description}
											<div class="campaign-description">
												<p>{campaign.description}</p>
											</div>
										{/if}
										
										<div class="campaign-metrics">
											<div class="metric">
												<label>Budget</label>
												<span class="budget-amount">{formatCurrency(campaign.budget)}</span>
											</div>
											{#if campaign.status !== 'pending'}
												<div class="metric">
													<label>Spent</label>
													<span>{formatCurrency(campaign.spent_amount)}</span>
												</div>
											{/if}
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
										
										<!-- Advertiser info -->
										{#if advertiserInfo}
											<div class="advertiser-info">
												<label>Advertiser</label>
												<div class="advertiser-card">
													<div class="advertiser-avatar">
														{advertiserInfo.company_name.charAt(0)}
													</div>
													<div class="advertiser-details">
														<h4>{advertiserInfo.company_name}</h4>
														<p>{advertiserInfo.contact_name} • {advertiserInfo.business_email}</p>
														<span class="status-badge {getStatusBadgeClass(advertiserInfo.status)}">
															{advertiserInfo.status}
														</span>
													</div>
												</div>
											</div>
										{/if}
										
										<!-- Action buttons based on status -->
										{#if campaign.status === 'pending'}
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
										{:else if campaign.status === 'approved' || campaign.status === 'active'}
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
										{:else if campaign.status === 'rejected'}
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
										{:else if campaign.status === 'cancelled'}
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
										{/if}
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			{:else}
				<!-- Campaign Approvals Accordion -->
				<div class="content-section">
					<div class="section-header">
						<h2>Campaign Approvals</h2>
						<p>Review and approve advertising campaigns</p>
					</div>

					<!-- Pending Campaigns Accordion -->
					<div class="accordion-section">
						<button 
							class="accordion-header {accordionState.campaigns.pending ? 'active' : ''}"
							on:click={() => toggleAccordion('campaigns', 'pending')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10" />
									<path d="M12 6v6l4 2" />
								</svg>
								<span>Pending Approval</span>
								<span class="accordion-count">{campaignCounts.pending}</span>
							</div>
							<svg class="accordion-chevron {accordionState.campaigns.pending ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.campaigns.pending}
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
										<p>{searchQuery ? 'No pending campaigns match your search.' : 'No campaigns pending approval.'}</p>
									</div>
								{:else}
									<div class="campaigns-grid">
										{#each pendingCampaigns as campaign}
											{@const advertiserInfo = getAdvertiserInfo(campaign.advertiser_id)}
											<div class="campaign-card">
												<div class="card-header">
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
												
												<div class="card-content">
													{#if campaign.description}
														<div class="campaign-description">
															<p>{campaign.description}</p>
														</div>
													{/if}
													
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
													
													<!-- Advertiser info -->
													{#if advertiserInfo}
														<div class="advertiser-info">
															<label>Advertiser</label>
															<div class="advertiser-card">
																<div class="advertiser-avatar">
																	{advertiserInfo.company_name.charAt(0)}
																</div>
																<div class="advertiser-details">
																	<h4>{advertiserInfo.company_name}</h4>
																	<p>{advertiserInfo.contact_name} • {advertiserInfo.business_email}</p>
																	<span class="status-badge {getStatusBadgeClass(advertiserInfo.status)}">
																		{advertiserInfo.status}
																	</span>
																</div>
															</div>
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

					<!-- Approved Campaigns Accordion -->
					<div class="accordion-section">
						<button 
							class="accordion-header {accordionState.campaigns.approved ? 'active' : ''}"
							on:click={() => toggleAccordion('campaigns', 'approved')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<path d="M9 12l2 2 4-4" />
									<circle cx="12" cy="12" r="10" />
								</svg>
								<span>Approved</span>
								<span class="accordion-count">{campaignCounts.approved}</span>
							</div>
							<svg class="accordion-chevron {accordionState.campaigns.approved ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.campaigns.approved}
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
										<p>{searchQuery ? 'No approved campaigns match your search.' : 'No approved campaigns.'}</p>
									</div>
								{:else}
									<div class="campaigns-grid">
										{#each approvedCampaigns as campaign}
											{@const statusInfo = getStatusInfo(campaign)}
											{@const advertiserInfo = getAdvertiserInfo(campaign.advertiser_id)}
											<div class="campaign-card">
												<div class="card-header">
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
																{#if statusInfo}
																	<div class="status-info">
																		<span class="status-action">{statusInfo.text} by {statusInfo.admin}</span>
																		<span class="status-date">{statusInfo.date}</span>
																	</div>
																{/if}
															</div>
														</div>
													</div>
												</div>
												
												<div class="card-content">
													{#if campaign.description}
														<div class="campaign-description">
															<p>{campaign.description}</p>
														</div>
													{/if}
													
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
													
													<!-- Advertiser info -->
													{#if advertiserInfo}
														<div class="advertiser-info">
															<label>Advertiser</label>
															<div class="advertiser-card">
																<div class="advertiser-avatar">
																	{advertiserInfo.company_name.charAt(0)}
																</div>
																<div class="advertiser-details">
																	<h4>{advertiserInfo.company_name}</h4>
																	<p>{advertiserInfo.contact_name} • {advertiserInfo.business_email}</p>
																	<span class="status-badge {getStatusBadgeClass(advertiserInfo.status)}">
																		{advertiserInfo.status}
																	</span>
																</div>
															</div>
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
							class="accordion-header {accordionState.campaigns.rejected ? 'active' : ''}"
							on:click={() => toggleAccordion('campaigns', 'rejected')}
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
							<svg class="accordion-chevron {accordionState.campaigns.rejected ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.campaigns.rejected}
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
										<p>{searchQuery ? 'No rejected campaigns match your search.' : 'No rejected campaigns.'}</p>
									</div>
								{:else}
									<div class="campaigns-grid">
										{#each rejectedCampaigns as campaign}
											{@const statusInfo = getStatusInfo(campaign)}
											{@const advertiserInfo = getAdvertiserInfo(campaign.advertiser_id)}
											<div class="campaign-card">
												<div class="card-header">
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
																{#if statusInfo}
																	<div class="status-info">
																		<span class="status-action">{statusInfo.text} by {statusInfo.admin}</span>
																		<span class="status-date">{statusInfo.date}</span>
																	</div>
																{/if}
															</div>
														</div>
													</div>
												</div>
												
												<div class="card-content">
													{#if campaign.description}
														<div class="campaign-description">
															<p>{campaign.description}</p>
														</div>
													{/if}
													
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
													
													<!-- Advertiser info -->
													{#if advertiserInfo}
														<div class="advertiser-info">
															<label>Advertiser</label>
															<div class="advertiser-card">
																<div class="advertiser-avatar">
																	{advertiserInfo.company_name.charAt(0)}
																</div>
																<div class="advertiser-details">
																	<h4>{advertiserInfo.company_name}</h4>
																	<p>{advertiserInfo.contact_name} • {advertiserInfo.business_email}</p>
																	<span class="status-badge {getStatusBadgeClass(advertiserInfo.status)}">
																		{advertiserInfo.status}
																	</span>
																</div>
															</div>
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
							class="accordion-header {accordionState.campaigns.cancelled ? 'active' : ''}"
							on:click={() => toggleAccordion('campaigns', 'cancelled')}
						>
							<div class="accordion-title">
								<svg class="accordion-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
									<circle cx="12" cy="12" r="10" />
									<path d="M3 6h18M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
								</svg>
								<span>Cancelled</span>
								<span class="accordion-count">{campaignCounts.cancelled}</span>
							</div>
							<svg class="accordion-chevron {accordionState.campaigns.cancelled ? 'rotated' : ''}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<path d="M6 9l6 6 6-6" />
							</svg>
						</button>
						{#if accordionState.campaigns.cancelled}
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
										<p>{searchQuery ? 'No cancelled campaigns match your search.' : 'No cancelled campaigns.'}</p>
									</div>
								{:else}
									<div class="campaigns-grid">
										{#each cancelledCampaigns as campaign}
											{@const statusInfo = getStatusInfo(campaign)}
											{@const advertiserInfo = getAdvertiserInfo(campaign.advertiser_id)}
											<div class="campaign-card">
												<div class="card-header">
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
																{#if statusInfo}
																	<div class="status-info">
																		<span class="status-action">{statusInfo.text} by {statusInfo.admin}</span>
																		<span class="status-date">{statusInfo.date}</span>
																	</div>
																{/if}
															</div>
														</div>
													</div>
												</div>
												
												<div class="card-content">
													{#if campaign.description}
														<div class="campaign-description">
															<p>{campaign.description}</p>
														</div>
													{/if}
													
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
													
													<!-- Advertiser info -->
													{#if advertiserInfo}
														<div class="advertiser-info">
															<label>Advertiser</label>
															<div class="advertiser-card">
																<div class="advertiser-avatar">
																	{advertiserInfo.company_name.charAt(0)}
																</div>
																<div class="advertiser-details">
																	<h4>{advertiserInfo.company_name}</h4>
																	<p>{advertiserInfo.contact_name} • {advertiserInfo.business_email}</p>
																	<span class="status-badge {getStatusBadgeClass(advertiserInfo.status)}">
																		{advertiserInfo.status}
																	</span>
																</div>
															</div>
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
			{/if}
		</div>
	{/if}
</div>

<style>
	.advertisement-management {
		display: flex;
		flex-direction: column;
		gap: var(--space-xl);
	}

	.page-header {
		display: flex;
		flex-direction: column;
		gap: var(--space-lg);
	}

	.header-content {
		display: flex;
		align-items: center;
		justify-content: space-between;
		flex-wrap: wrap;
		gap: var(--space-lg);
	}

	.header-content h1 {
		font-size: var(--text-3xl);
		font-weight: 700;
		color: var(--text-primary);
		margin: 0;
	}

	.header-content p {
		color: var(--text-secondary);
		margin: var(--space-sm) 0 0 0;
	}

	.header-actions {
		display: flex;
		gap: var(--space-md);
	}

	.error-message {
		margin-bottom: var(--space-lg);
	}

	.loading-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: var(--space-4xl);
		gap: var(--space-lg);
	}

	/* Tab Navigation */
	.tab-navigation {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.tab-container {
		display: flex;
		gap: var(--space-md);
	}

	.tab-button {
		display: flex;
		align-items: center;
		gap: var(--space-sm);
		padding: var(--space-md) var(--space-lg);
		border: none;
		border-radius: var(--radius-lg);
		background: transparent;
		color: var(--text-secondary);
		font-weight: 500;
		cursor: pointer;
		transition: all var(--transition-normal);
		position: relative;
	}

	.tab-button:hover {
		background: var(--bg-glass-dark);
		color: var(--text-primary);
	}

	.tab-button.active {
		background: var(--primary-gradient);
		color: var(--white);
		box-shadow: var(--shadow-md);
	}

	.tab-button svg {
		width: 20px;
		height: 20px;
	}

	.tab-count {
		background: rgba(255, 255, 255, 0.2);
		color: inherit;
		padding: var(--space-xs) var(--space-sm);
		border-radius: var(--radius-full);
		font-size: var(--text-xs);
		font-weight: 600;
		min-width: 20px;
		text-align: center;
	}

	.tab-button.active .tab-count {
		background: rgba(255, 255, 255, 0.3);
	}

	/* Content Section */
	.content-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-xl);
		border: 1px solid rgba(255, 255, 255, 0.1);
	}

	.section-header {
		margin-bottom: var(--space-xl);
	}

	.section-header h2 {
		font-size: var(--text-2xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.section-header p {
		color: var(--text-secondary);
		margin: 0;
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

	/* Card Grids */
	.accounts-grid,
	.campaigns-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
		gap: var(--space-lg);
	}

	/* Card Styles */
	.account-card,
	.campaign-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.05);
		transition: all var(--transition-normal);
		overflow: hidden;
	}

	.account-card:hover,
	.campaign-card:hover {
		transform: translateY(-2px);
		box-shadow: var(--shadow-lg);
		border-color: rgba(255, 255, 255, 0.1);
	}

	.card-header {
		padding: var(--space-lg);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-md);
	}

	.company-info,
	.campaign-info {
		display: flex;
		align-items: flex-start;
		gap: var(--space-md);
		flex: 1;
	}

	.company-avatar {
		width: 48px;
		height: 48px;
		background: var(--primary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 700;
		color: var(--white);
		font-size: var(--text-lg);
		flex-shrink: 0;
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

	.company-details,
	.campaign-details {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: var(--space-sm);
	}

	.company-details h3,
	.campaign-details h3 {
		font-size: var(--text-lg);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.campaign-meta {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
		align-items: center;
	}

	.company-meta {
		display: flex;
		gap: var(--space-sm);
		flex-wrap: wrap;
		align-items: center;
	}

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

	.status-active {
		background: rgba(59, 130, 246, 0.1);
		color: rgb(59, 130, 246);
	}

	.status-paused {
		background: rgba(107, 114, 128, 0.1);
		color: rgb(107, 114, 128);
	}

	.status-cancelled {
		background: rgba(239, 68, 68, 0.1);
		color: rgb(239, 68, 68);
	}

	.status-default {
		background: rgba(107, 114, 128, 0.1);
		color: rgb(107, 114, 128);
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

	/* Advertiser Info Section */
	.advertiser-info {
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		margin-bottom: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.advertiser-header {
		display: flex;
		align-items: center;
		gap: var(--space-md);
	}

	.advertiser-avatar {
		width: 40px;
		height: 40px;
		background: var(--secondary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: var(--white);
		font-size: var(--text-base);
		flex-shrink: 0;
	}

	.advertiser-details {
		flex: 1;
	}

	.advertiser-details h4 {
		font-size: var(--text-base);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
	}

	.advertiser-contact {
		font-size: var(--text-sm);
		color: var(--text-secondary);
		margin: 0 0 var(--space-xs) 0;
	}

	.advertiser-industry {
		font-size: var(--text-xs);
		color: var(--text-tertiary);
		font-style: italic;
		margin: 0;
	}

	.card-actions {
		display: flex;
		gap: var(--space-sm);
		margin-top: var(--space-sm);
	}

	.view-details-btn {
		display: flex;
		align-items: center;
		margin-left: auto;
	}

	.view-details-btn .btn {
		white-space: nowrap;
	}

	.card-content {
		padding: var(--space-lg);
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
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

	.budget-amount {
		font-weight: 700;
		color: var(--primary);
		font-size: var(--text-base) !important;
	}

	.campaign-description,
	.audience-section,
	.notes-section {
		margin-bottom: var(--space-lg);
	}

	.campaign-description:last-child,
	.audience-section:last-child,
	.notes-section:last-child {
		margin-bottom: 0;
	}

	.campaign-description p,
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

	/* Responsive Design */
	@media (max-width: 1024px) {
		.accounts-grid,
		.campaigns-grid {
			grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		}
	}

	@media (max-width: 768px) {
		.header-content {
			flex-direction: column;
			align-items: stretch;
		}

		.tab-container {
			flex-direction: column;
		}

		.accounts-grid,
		.campaigns-grid {
			grid-template-columns: 1fr;
		}

		.card-header {
			flex-direction: column;
			align-items: stretch;
			gap: var(--space-lg);
		}

		.company-info,
		.campaign-info {
			align-items: center;
		}

		.card-actions {
			justify-content: stretch;
		}

		.card-actions .btn {
			flex: 1;
		}

		.info-grid {
			grid-template-columns: 1fr;
		}
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

	.cancel-confirmation {
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

	.review-confirmation {
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		padding: var(--space-lg);
		border: 1px solid rgba(239, 68, 68, 0.2);
	}

	.review-confirmation p {
		font-size: var(--text-sm);
		color: var(--text-primary);
		margin: 0 0 var(--space-lg) 0;
		text-align: center;
		font-weight: 500;
	}

	.review-confirmation .confirmation-actions {
		display: flex;
		gap: var(--space-md);
		justify-content: center;
	}

	.review-confirmation .confirmation-actions .btn {
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

	/* Search Section */
	.search-section {
		background: var(--bg-glass);
		border-radius: var(--radius-xl);
		padding: var(--space-lg);
		border: 1px solid rgba(255, 255, 255, 0.1);
		margin-bottom: var(--space-xl);
	}

	.search-container {
		max-width: 600px;
		margin: 0 auto;
	}

	.search-input-wrapper {
		position: relative;
		display: flex;
		align-items: center;
	}

	.search-icon {
		position: absolute;
		left: var(--space-md);
		width: 20px;
		height: 20px;
		color: var(--text-secondary);
		z-index: 2;
	}

	.search-input {
		width: 100%;
		padding: var(--space-md) var(--space-md) var(--space-md) 48px;
		background: var(--bg-glass-dark);
		border: 1px solid rgba(255, 255, 255, 0.1);
		border-radius: var(--radius-lg);
		color: var(--text-primary);
		font-size: var(--text-base);
		transition: all var(--transition-normal);
	}

	.search-input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.search-input::placeholder {
		color: var(--text-tertiary);
	}

	.clear-search {
		position: absolute;
		right: var(--space-md);
		width: 20px;
		height: 20px;
		background: none;
		border: none;
		color: var(--text-secondary);
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		transition: color var(--transition-normal);
	}

	.clear-search:hover {
		color: var(--text-primary);
	}

	.clear-search svg {
		width: 16px;
		height: 16px;
	}

	/* Search Results */
	.search-results {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.05);
		overflow: hidden;
	}

	.search-results-header {
		padding: var(--space-xl);
		border-bottom: 1px solid rgba(255, 255, 255, 0.05);
		background: var(--bg-glass);
	}

	.search-results-header h3 {
		font-size: var(--text-xl);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-sm) 0;
	}

	.search-results-header p {
		color: var(--text-secondary);
		margin: 0;
		font-size: var(--text-sm);
	}

	.search-results .accounts-grid,
	.search-results .campaigns-grid {
		padding: var(--space-xl);
	}

	.search-results .empty-state {
		padding: var(--space-4xl);
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
	.accordion-content .accounts-grid > *,
	.accordion-content .campaigns-grid > * {
		animation: slideInUp 0.3s ease-out backwards;
	}

	.accordion-content .accounts-grid > *:nth-child(1),
	.accordion-content .campaigns-grid > *:nth-child(1) {
		animation-delay: 0.1s;
	}

	.accordion-content .accounts-grid > *:nth-child(2),
	.accordion-content .campaigns-grid > *:nth-child(2) {
		animation-delay: 0.15s;
	}

	.accordion-content .accounts-grid > *:nth-child(3),
	.accordion-content .campaigns-grid > *:nth-child(3) {
		animation-delay: 0.2s;
	}

	.accordion-content .accounts-grid > *:nth-child(4),
	.accordion-content .campaigns-grid > *:nth-child(4) {
		animation-delay: 0.25s;
	}

	.accordion-content .accounts-grid > *:nth-child(n+5),
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

	/* Enhanced card animations */
	.account-card,
	.campaign-card {
		background: var(--bg-glass-dark);
		border-radius: var(--radius-xl);
		border: 1px solid rgba(255, 255, 255, 0.05);
		transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
		overflow: hidden;
		position: relative;
	}

	.account-card::before,
	.campaign-card::before {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: linear-gradient(135deg, rgba(255, 255, 255, 0.05) 0%, transparent 50%);
		opacity: 0;
		transition: opacity var(--transition-normal);
		pointer-events: none;
	}

	.account-card:hover,
	.campaign-card:hover {
		transform: translateY(-4px) scale(1.02);
		box-shadow: 
			var(--shadow-lg),
			0 20px 40px rgba(0, 0, 0, 0.1),
			0 0 0 1px rgba(255, 255, 255, 0.1);
		border-color: rgba(255, 255, 255, 0.15);
	}

	.account-card:hover::before,
	.campaign-card:hover::before {
		opacity: 1;
	}

	/* Campaign Metrics */
	.campaign-metrics {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: var(--space-md);
		margin-bottom: var(--space-lg);
		padding: var(--space-md);
		background: var(--bg-glass);
		border-radius: var(--radius-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.metric {
		display: flex;
		flex-direction: column;
		gap: var(--space-xs);
		text-align: center;
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

	/* Advertiser Card in Campaign */
	.advertiser-card {
		display: flex;
		align-items: center;
		gap: var(--space-md);
		padding: var(--space-md);
		background: var(--bg-glass-dark);
		border-radius: var(--radius-lg);
		border: 1px solid rgba(255, 255, 255, 0.05);
	}

	.advertiser-card .advertiser-avatar {
		width: 36px;
		height: 36px;
		background: var(--secondary-gradient);
		border-radius: var(--radius-lg);
		display: flex;
		align-items: center;
		justify-content: center;
		font-weight: 600;
		color: var(--white);
		font-size: var(--text-sm);
		flex-shrink: 0;
	}

	.advertiser-card .advertiser-details {
		flex: 1;
	}

	.advertiser-card .advertiser-details h4 {
		font-size: var(--text-sm);
		font-weight: 600;
		color: var(--text-primary);
		margin: 0 0 var(--space-xs) 0;
	}

	.advertiser-card .advertiser-details p {
		font-size: var(--text-xs);
		color: var(--text-secondary);
		margin: 0 0 var(--space-xs) 0;
	}

	.advertiser-card .status-badge {
		font-size: 10px;
		padding: 2px 6px;
	}

	/* Mobile Responsive Improvements */
	@media (max-width: 640px) {
		.search-input {
			padding-left: 40px;
		}

		.search-icon {
			left: var(--space-sm);
		}

		.clear-search {
			right: var(--space-sm);
		}

		.accordion-title {
			font-size: var(--text-base);
			gap: var(--space-sm);
		}

		.accordion-icon {
			width: 20px;
			height: 20px;
		}

		.campaign-metrics {
			grid-template-columns: repeat(2, 1fr);
		}

		.advertiser-card {
			flex-direction: column;
			align-items: flex-start;
			text-align: left;
		}

		.advertiser-card .advertiser-details {
			width: 100%;
		}
	}
</style> 