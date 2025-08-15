import type { LayoutLoad } from './$types';
import { browser } from '$app/environment';
import { redirect } from '@sveltejs/kit';
import { SecureTokenStorage } from '$lib/auth';

export const load: LayoutLoad = async ({ url }) => {
	// Only run on client side
	if (!browser) {
		return {};
	}

	// Check if user is authenticated
	const tokens = SecureTokenStorage.getTokens();
	const user = SecureTokenStorage.getUser();

	if (tokens && user) {
		// Allow access to all routes for authenticated users
		return {};
	}

	return {};
};
