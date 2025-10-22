import type { Handle } from '@sveltejs/kit';

/**
 * SvelteKit server hooks
 * Handles server-side logic before requests reach routes
 */
export const handle: Handle = async ({ event, resolve }) => {
	// Redirect favicon.ico requests to favicon.png
	if (event.url.pathname === '/favicon.ico') {
		return new Response(null, {
			status: 301,
			headers: {
				Location: '/favicon.png'
			}
		});
	}

	// Continue with normal request handling
	return await resolve(event);
};

