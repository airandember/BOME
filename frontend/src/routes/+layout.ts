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
		// Check if user needs to change password
		if (!user.password_changed && url.pathname !== '/change-password') {
			// Redirect to password change page
			throw redirect(302, '/change-password');
		}

		// If user needs to change password and is on change-password page, allow
		if (!user.password_changed && url.pathname === '/change-password') {
			return {};
		}

		// If user has changed password, allow access to all routes
		if (user.password_changed) {
			return {};
		}
	}

	return {};
};
