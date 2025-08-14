import adapter from '@sveltejs/adapter-node';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://svelte.dev/docs/kit/integrations
	// for more information about preprocessors
	preprocess: vitePreprocess(),

	kit: {
		// Using adapter-node for DigitalOcean App Platform compatibility
		// This creates a proper Node.js server instead of static-only output
		adapter: adapter({
			// Optional: Configure the adapter for production
			out: 'build',
			precompress: false,
			envPrefix: ''
		})
	}
};

export default config;
