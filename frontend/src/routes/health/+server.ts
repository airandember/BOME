import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

export const GET: RequestHandler = async () => {
	return json({
		status: 'healthy',
		timestamp: new Date().toISOString(),
		service: 'bome-frontend',
		version: '1.0.0'
	});
};
