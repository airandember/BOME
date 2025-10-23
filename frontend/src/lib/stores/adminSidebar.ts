import { writable, derived } from 'svelte/store';
import { page } from '$app/stores';

// Main admin sidebar state
export const sidebarCollapsed = writable<boolean>(false);

// Auto-collapse when in subsites
export const autoCollapseForSubsite = derived(
	page,
	($page) => {
		const path = $page.url.pathname;
		// Collapse sidebar if we're in a subsite
		return path.startsWith('/admin/streaming') || 
		       path.startsWith('/admin/articles') || 
		       path.startsWith('/admin/expo');
	}
);

// Combined state: collapsed if manually toggled OR in a subsite
export const shouldCollapse = derived(
	[sidebarCollapsed, autoCollapseForSubsite],
	([$collapsed, $autoCollapse]) => $collapsed || $autoCollapse
);

// Toggle sidebar manually
export function toggleSidebar() {
	sidebarCollapsed.update(collapsed => !collapsed);
}

