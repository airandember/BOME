// Mock data for streaming admin dashboard
// This service provides sample data while backend APIs are being developed

export interface MockDashboardData {
	dashboard: {
		metrics: {
			active_subscriptions: number;
			revenue_30_days: { total: number };
			churn_rate: { rate: number };
			new_subscriptions: number;
			total_customers: number;
			avg_revenue_per_user: number;
		};
		recent_events: Array<{
			description: string;
			timestamp: string;
		}>;
	};
}

export interface MockSubscriptionPlan {
	id: string;
	name: string;
	description: string;
	short_desc: string;
	price: number;
	currency: string;
	interval: 'monthly' | 'yearly' | 'weekly' | 'daily';
	interval_count: number;
	features: string[];
	is_active: boolean;
	is_promoted: boolean;
	promotion_end_date: string | null;
	created_at: string;
	updated_at: string;
}

export interface MockSubscriptionPlansResponse {
	subscription_plans: MockSubscriptionPlan[];
	total: number;
	page: number;
	limit: number;
}

export class MockDataService {
	// Mock dashboard data
	static getDashboardData(): MockDashboardData {
		return {
			dashboard: {
				metrics: {
					active_subscriptions: 1247,
					revenue_30_days: { total: 45678.90 },
					churn_rate: { rate: 2.3 },
					new_subscriptions: 89,
					total_customers: 2156,
					avg_revenue_per_user: 36.75
				},
				recent_events: [
					{
						description: "New subscription: Premium Plan - $29.99/month",
						timestamp: "2 minutes ago"
					},
					{
						description: "Payment processed: Basic Plan - $9.99/month",
						timestamp: "5 minutes ago"
					},
					{
						description: "Subscription cancelled: Pro Plan - $19.99/month",
						timestamp: "12 minutes ago"
					},
					{
						description: "New customer registered: john.doe@example.com",
						timestamp: "15 minutes ago"
					},
					{
						description: "Payment failed: Basic Plan - $9.99/month",
						timestamp: "23 minutes ago"
					}
				]
			}
		};
	}

	// Mock subscription plans
	static getSubscriptionPlans(): MockSubscriptionPlansResponse {
		return {
			subscription_plans: [
				{
					id: "1",
					name: "Basic Plan",
					description: "Perfect for getting started with streaming content. Includes access to basic features and limited content library.",
					short_desc: "Essential streaming access",
					price: 9.99,
					currency: "USD",
					interval: "monthly",
					interval_count: 1,
					features: [
						"HD streaming quality",
						"Access to basic content library",
						"Watch on 1 device",
						"Standard customer support"
					],
					is_active: true,
					is_promoted: false,
					promotion_end_date: null,
					created_at: "2024-01-15T10:30:00Z",
					updated_at: "2024-01-15T10:30:00Z"
				},
				{
					id: "2",
					name: "Pro Plan",
					description: "Enhanced streaming experience with premium features, higher quality content, and multi-device support.",
					short_desc: "Premium streaming experience",
					price: 19.99,
					currency: "USD",
					interval: "monthly",
					interval_count: 1,
					features: [
						"4K Ultra HD streaming",
						"Access to premium content library",
						"Watch on 3 devices simultaneously",
						"Priority customer support",
						"Offline downloads",
						"Ad-free experience"
					],
					is_active: true,
					is_promoted: true,
					promotion_end_date: "2024-12-31T23:59:59Z",
					created_at: "2024-01-15T10:30:00Z",
					updated_at: "2024-01-15T10:30:00Z"
				},
				{
					id: "3",
					name: "Premium Plan",
					description: "Ultimate streaming package with exclusive content, highest quality streaming, and unlimited device access.",
					short_desc: "Ultimate streaming package",
					price: 29.99,
					currency: "USD",
					interval: "monthly",
					interval_count: 1,
					features: [
						"8K Ultra HD streaming",
						"Access to exclusive content library",
						"Unlimited device streaming",
						"24/7 premium customer support",
						"Unlimited offline downloads",
						"Ad-free experience",
						"Early access to new content",
						"Exclusive behind-the-scenes content"
					],
					is_active: true,
					is_promoted: false,
					promotion_end_date: null,
					created_at: "2024-01-15T10:30:00Z",
					updated_at: "2024-01-15T10:30:00Z"
				},
				{
					id: "4",
					name: "Annual Basic",
					description: "Save 20% with annual billing for the Basic Plan. All the same features with better value.",
					short_desc: "Annual Basic with savings",
					price: 95.88,
					currency: "USD",
					interval: "yearly",
					interval_count: 1,
					features: [
						"HD streaming quality",
						"Access to basic content library",
						"Watch on 1 device",
						"Standard customer support",
						"20% savings compared to monthly"
					],
					is_active: true,
					is_promoted: false,
					promotion_end_date: null,
					created_at: "2024-01-15T10:30:00Z",
					updated_at: "2024-01-15T10:30:00Z"
				},
				{
					id: "5",
					name: "Student Plan",
					description: "Special pricing for verified students. Same features as Basic Plan at a discounted rate.",
					short_desc: "Student discount plan",
					price: 4.99,
					currency: "USD",
					interval: "monthly",
					interval_count: 1,
					features: [
						"HD streaming quality",
						"Access to basic content library",
						"Watch on 1 device",
						"Standard customer support",
						"Student verification required"
					],
					is_active: true,
					is_promoted: false,
					promotion_end_date: null,
					created_at: "2024-01-15T10:30:00Z",
					updated_at: "2024-01-15T10:30:00Z"
				}
			],
			total: 5,
			page: 1,
			limit: 20
		};
	}

	// Mock customer data
	static getCustomers() {
		return {
			customers: [
				{
					id: "1",
					email: "john.doe@example.com",
					name: "John Doe",
					status: "active",
					subscription: {
						id: "sub_1",
						plan_name: "Pro Plan",
						price: 19.99,
						currency: "USD",
						interval: "monthly",
						status: "active",
						current_period_start: "2024-01-01T00:00:00Z",
						current_period_end: "2024-02-01T00:00:00Z",
						cancel_at_period_end: false
					},
					created_at: "2023-06-15T10:30:00Z",
					updated_at: "2024-01-15T10:30:00Z",
					last_login: "2024-01-15T09:45:00Z",
					total_spent: 239.88,
					subscription_count: 1
				},
				{
					id: "2",
					email: "jane.smith@example.com",
					name: "Jane Smith",
					status: "active",
					subscription: {
						id: "sub_2",
						plan_name: "Premium Plan",
						price: 29.99,
						currency: "USD",
						interval: "monthly",
						status: "active",
						current_period_start: "2024-01-01T00:00:00Z",
						current_period_end: "2024-02-01T00:00:00Z",
						cancel_at_period_end: false
					},
					created_at: "2023-08-20T14:15:00Z",
					updated_at: "2024-01-15T10:30:00Z",
					last_login: "2024-01-15T08:30:00Z",
					total_spent: 179.94,
					subscription_count: 1
				}
			],
			total: 2,
			page: 1,
			limit: 20
		};
	}

	// Mock analytics data
	static getAnalyticsOverview() {
		return {
			revenue_trend: [
				{ date: "2024-01-01", amount: 1500, currency: "USD", subscriptions: 50 },
				{ date: "2024-01-02", amount: 1650, currency: "USD", subscriptions: 55 },
				{ date: "2024-01-03", amount: 1800, currency: "USD", subscriptions: 60 }
			],
			subscription_trend: [
				{ date: "2024-01-01", new_subscriptions: 5, cancelled_subscriptions: 2, active_subscriptions: 50 },
				{ date: "2024-01-02", new_subscriptions: 7, cancelled_subscriptions: 1, active_subscriptions: 55 },
				{ date: "2024-01-03", new_subscriptions: 6, cancelled_subscriptions: 3, active_subscriptions: 60 }
			],
			churn_trend: [
				{ date: "2024-01-01", churn_rate: 2.1, cancelled_count: 2, total_subscribers: 50 },
				{ date: "2024-01-02", churn_rate: 1.8, cancelled_count: 1, total_subscribers: 55 },
				{ date: "2024-01-03", churn_rate: 2.5, cancelled_count: 3, total_subscribers: 60 }
			],
			summary: {
				total_revenue: 45678.90,
				total_subscriptions: 1247,
				avg_churn_rate: 2.3,
				growth_rate: 8.5
			}
		};
	}

	// Simulate API delay
	static async delay(ms: number = 500): Promise<void> {
		return new Promise(resolve => setTimeout(resolve, ms));
	}

	// Simulate API error (randomly)
	static async simulateError(probability: number = 0.1): Promise<void> {
		if (Math.random() < probability) {
			throw new Error("Simulated API error");
		}
	}
} 