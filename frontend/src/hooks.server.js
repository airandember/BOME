/** @type {import('@sveltejs/kit').Handle} */
export async function handle({ event, resolve }) {
	// Safety net: If someone hits /api/* directly (without /bome-backend prefix)
	// This should rarely happen now that email URLs are fixed, but provides clear debugging
	if (event.url.pathname.startsWith('/api/')) {
		console.warn(`🚨 [HOOKS] Direct API route accessed: ${event.url.pathname}`);
		console.warn(`💡 [HOOKS] Correct path should be: /bome-backend${event.url.pathname}`);
		
		// Return helpful error message
		return new Response(
			`API Routing Error\n\n` +
			`Requested: ${event.url.pathname}\n` +
			`Correct path: /bome-backend${event.url.pathname}\n\n` +
			`All API calls should go through /bome-backend/api/v1/*`,
			{
				status: 502,
				headers: {
					'Content-Type': 'text/plain',
					'X-SvelteKit-API-Passthrough': 'true',
					'X-Correct-Path': `/bome-backend${event.url.pathname}`
				}
			}
		);
	}

	// Handle all other routes normally
	const response = await resolve(event);
	return response;
}
