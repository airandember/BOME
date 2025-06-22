import { writable } from 'svelte/store';
import type { AdvertiserAccount } from '$lib/types/advertising';

export interface AdvertiserState {
	accounts: AdvertiserAccount[];
	loading: boolean;
	error: string | null;
}

const initialState: AdvertiserState = {
	accounts: [],
	loading: false,
	error: null
};

function createAdvertiserStore() {
	const { subscribe, set, update } = writable<AdvertiserState>(initialState);

	// Mock data storage - in production this would be replaced with API calls
	let mockAccounts: AdvertiserAccount[] = [
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
		}
	];

	return {
		subscribe,
		
		// Load all advertiser accounts (for admin)
		async loadAll() {
			update(state => ({ ...state, loading: true, error: null }));
			
			try {
				// Simulate API delay
				await new Promise(resolve => setTimeout(resolve, 500));
				
				update(state => ({
					...state,
					accounts: mockAccounts,
					loading: false
				}));
			} catch (error) {
				update(state => ({
					...state,
					error: error instanceof Error ? error.message : 'Failed to load accounts',
					loading: false
				}));
			}
		},

		// Get account by user ID (for user dashboard)
		async getByUserId(userId: number): Promise<AdvertiserAccount | null> {
			// Simulate API delay
			await new Promise(resolve => setTimeout(resolve, 300));
			
			const account = mockAccounts.find(acc => acc.user_id === userId);
			return account || null;
		},

		// Submit new advertiser application
		async submitApplication(applicationData: Partial<AdvertiserAccount>): Promise<AdvertiserAccount> {
			update(state => ({ ...state, loading: true, error: null }));
			
			try {
				// Simulate API delay
				await new Promise(resolve => setTimeout(resolve, 1000));
				
				const newAccount: AdvertiserAccount = {
					id: mockAccounts.length + 1,
					user_id: applicationData.user_id!,
					company_name: applicationData.company_name!,
					business_email: applicationData.business_email!,
					contact_name: applicationData.contact_name!,
					contact_phone: applicationData.contact_phone || '',
					business_address: applicationData.business_address || '',
					tax_id: applicationData.tax_id || '',
					website: applicationData.website || '',
					industry: applicationData.industry || '',
					status: 'pending',
					created_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};

				mockAccounts.push(newAccount);
				
				update(state => ({
					...state,
					accounts: [...mockAccounts],
					loading: false
				}));

				return newAccount;
			} catch (error) {
				update(state => ({
					...state,
					error: error instanceof Error ? error.message : 'Failed to submit application',
					loading: false
				}));
				throw error;
			}
		},

		// Approve advertiser account (admin action)
		async approveAccount(accountId: number, adminId: number): Promise<void> {
			update(state => ({ ...state, loading: true, error: null }));
			
			try {
				// Simulate API delay
				await new Promise(resolve => setTimeout(resolve, 800));
				
				const accountIndex = mockAccounts.findIndex(acc => acc.id === accountId);
				if (accountIndex === -1) {
					throw new Error('Account not found');
				}

				mockAccounts[accountIndex] = {
					...mockAccounts[accountIndex],
					status: 'approved',
					approved_by: adminId,
					approved_at: new Date().toISOString(),
					updated_at: new Date().toISOString()
				};

				update(state => ({
					...state,
					accounts: [...mockAccounts],
					loading: false
				}));
			} catch (error) {
				update(state => ({
					...state,
					error: error instanceof Error ? error.message : 'Failed to approve account',
					loading: false
				}));
				throw error;
			}
		},

		// Reject advertiser account (admin action)
		async rejectAccount(accountId: number, adminId: number, reason: string): Promise<void> {
			update(state => ({ ...state, loading: true, error: null }));
			
			try {
				// Simulate API delay
				await new Promise(resolve => setTimeout(resolve, 800));
				
				const accountIndex = mockAccounts.findIndex(acc => acc.id === accountId);
				if (accountIndex === -1) {
					throw new Error('Account not found');
				}

				mockAccounts[accountIndex] = {
					...mockAccounts[accountIndex],
					status: 'rejected',
					rejected_by: adminId,
					rejected_at: new Date().toISOString(),
					verification_notes: reason,
					updated_at: new Date().toISOString()
				};

				update(state => ({
					...state,
					accounts: [...mockAccounts],
					loading: false
				}));
			} catch (error) {
				update(state => ({
					...state,
					error: error instanceof Error ? error.message : 'Failed to reject account',
					loading: false
				}));
				throw error;
			}
		},

		// Reset store
		reset() {
			set(initialState);
		}
	};
}

export const advertiserStore = createAdvertiserStore(); 