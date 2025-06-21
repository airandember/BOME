import { redirect } from '@sveltejs/kit';

export async function load() {
	// Redirect from old /blog URL to new /articles URL
	throw redirect(301, '/articles');
} 