import { writable } from 'svelte/store';

/**
 * Sidebar collapsed state
 * true = collapsed to icons only (when in subsites)
 * false = expanded with full navigation
 */
export const sidebarCollapsed = writable(false);

