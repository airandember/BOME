import type { SubscriptionPlan } from '$lib/services/streaming-subscriptions';

// Store for plans (will be set from the main page)
let plansStore: SubscriptionPlan[] = [];

export function setPlansStore(plans: SubscriptionPlan[]) {
	plansStore = plans;
}

export function getPlanName(planId: number): string {
	const plan = plansStore.find(p => p.id === planId.toString());
	return plan ? plan.name : `Plan ${planId}`;
}

export function getItemName(itemId: string | null | undefined): string {
	if (!itemId) return 'No specific item';
	
	const itemTypes: Record<string, string> = {
		'ebook': 'eBook',
		'dvd': 'DVD', 
		'expo_ticket': 'Expo Ticket',
		'1': 'eBook',
		'2': 'DVD',
		'3': 'Expo Ticket'
	};
	
	return itemTypes[itemId] || itemId;
} 