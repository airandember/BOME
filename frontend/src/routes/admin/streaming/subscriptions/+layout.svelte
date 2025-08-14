<script lang="ts">
	import { page } from '$app/stores';
</script>

<div class="subscriptions-layout">
	<slot />
</div>

<style>
	.subscriptions-layout {
		min-height: 100vh;
		background: #f9fafb;
		padding: 2rem;
	}

	/* SVG max height constraint */
	:global(.svg-max-height) {
		max-height: 100px;
	}

	/* Clean, modern styling matching users page */
	:global(.bg-white) {
		background: white !important;
		border: 1px solid #e5e7eb;
	}

	:global(.bg-gray-50) {
		background: #f9fafb !important;
		border: 1px solid #e5e7eb;
	}

	:global(.border-gray-200) {
		border-color: #e5e7eb !important;
	}

	:global(.border-gray-300) {
		border-color: #d1d5db !important;
	}

	/* Button styling */
	:global(.bg-blue-600) {
		background: #2563eb !important;
		border: none;
		transition: all 0.2s ease;
	}

	:global(.bg-blue-600:hover) {
		background: #1d4ed8 !important;
		transform: translateY(-1px);
	}

	:global(.bg-blue-700) {
		background: #1d4ed8 !important;
	}

	:global(.bg-red-600) {
		background: #dc2626 !important;
		border: none;
		transition: all 0.2s ease;
	}

	:global(.bg-red-600:hover) {
		background: #b91c1c !important;
		transform: translateY(-1px);
	}

	:global(.bg-red-700) {
		background: #b91c1c !important;
	}

	:global(.bg-gray-100) {
		background: #f3f4f6 !important;
		border: 1px solid #e5e7eb;
		transition: all 0.2s ease;
	}

	:global(.bg-gray-100:hover) {
		background: #e5e7eb !important;
		transform: translateY(-1px);
	}

	:global(.bg-red-100) {
		background: #fee2e2 !important;
		border: 1px solid #fecaca;
		transition: all 0.2s ease;
	}

	:global(.bg-red-100:hover) {
		background: #fecaca !important;
		transform: translateY(-1px);
	}

	/* SVG styling */
	:global(svg) {
		color: inherit;
	}

	:global(.text-red-600 svg) {
		color: #dc2626;
	}

	:global(.text-red-700 svg) {
		color: #b91c1c;
	}

	:global(.text-green-500 svg) {
		color: #10b981;
	}

	:global(.text-green-600 svg) {
		color: #059669;
	}

	:global(.text-yellow-500 svg) {
		color: #f59e0b;
	}

	:global(.text-yellow-600 svg) {
		color: #d97706;
	}

	:global(.text-blue-600 svg) {
		color: #2563eb;
	}

	:global(.text-gray-400 svg) {
		color: #9ca3af;
	}

	:global(.text-gray-600 svg) {
		color: #6b7280;
	}

	/* Status badge styling */
	:global(.bg-green-100) {
		background: #dcfce7 !important;
		border: 1px solid #bbf7d0;
		color: #166534;
	}

	:global(.bg-yellow-100) {
		background: #fef3c7 !important;
		border: 1px solid #fde68a;
		color: #92400e;
	}

	:global(.bg-gray-100) {
		background: #f3f4f6 !important;
		border: 1px solid #e5e7eb;
		color: #374151;
	}

	/* Card styling */
	:global(.card) {
		background: white;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
		transition: all 0.2s ease;
	}

	:global(.card:hover) {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	/* Form styling */
	:global(input), :global(select), :global(textarea) {
		background: white !important;
		border-color: #d1d5db !important;
		color: #111827 !important;
		transition: all 0.2s ease;
	}

	:global(input:focus), :global(select:focus), :global(textarea:focus) {
		border-color: #2563eb !important;
		box-shadow: 0 0 0 3px rgba(37, 99, 235, 0.1) !important;
		background: white !important;
	}

	:global(input::placeholder) {
		color: #9ca3af !important;
	}

	/* Text colors */
	:global(.text-gray-700) {
		color: #374151 !important;
	}

	:global(.text-gray-600) {
		color: #6b7280 !important;
	}

	:global(.text-gray-900) {
		color: #111827 !important;
	}

	:global(.text-gray-800) {
		color: #1f2937 !important;
	}

	:global(.text-gray-500) {
		color: #6b7280 !important;
	}

	/* Modal styling */
	:global(.fixed.inset-0) {
		background: rgba(0, 0, 0, 0.5);
	}

	/* Modal container styling */
	:global(.card) {
		background: white !important;
		border-radius: 0.5rem;
		border: 1px solid #e5e7eb;
		box-shadow: 0 10px 25px rgba(0, 0, 0, 0.15);
		transition: all 0.2s ease;
	}

	/* Specific modal styling */
	:global(.fixed.inset-0 .card) {
		background: white !important;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
		border: 1px solid #e5e7eb;
	}

	/* Modal backdrop */
	:global(.fixed.inset-0.bg-black) {
		background: rgba(0, 0, 0, 0.5) !important;
	}

	/* Modal specific styling */
	:global(.modal-backdrop) {
		background: rgba(0, 0, 0, 0.5) !important;
	}

	:global(.modal-content) {
		background: white !important;
		box-shadow: 0 20px 40px rgba(0, 0, 0, 0.25);
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
	}

	/* Loading spinner */
	:global(.animate-spin) {
		border-color: #2563eb;
		border-top-color: transparent;
	}

	/* Line clamp utility */
	:global(.line-clamp-3) {
		display: -webkit-box;
		-webkit-line-clamp: 3;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	/* Checkbox styling */
	:global(input[type="checkbox"]) {
		background: white !important;
		border-color: #d1d5db !important;
		border-radius: 0.25rem;
	}

	:global(input[type="checkbox"]:checked) {
		background: #2563eb !important;
		border-color: #2563eb !important;
	}

	/* Responsive design */
	@media (max-width: 768px) {
		.subscriptions-layout {
			padding: 1rem;
		}

		:global(.card) {
			margin: 0.5rem;
		}
		
		:global(.grid) {
			grid-template-columns: 1fr;
		}
	}

	/* Hover effects for interactive elements */
	:global(button:not(:disabled):hover) {
		transform: translateY(-1px);
	}

	/* Disabled state styling */
	:global(button:disabled) {
		opacity: 0.6;
		cursor: not-allowed;
		transform: none !important;
	}

	/* Focus states for accessibility */
	:global(button:focus), :global(input:focus), :global(select:focus), :global(textarea:focus) {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	/* Smooth transitions */
	:global(*) {
		transition: all 0.2s ease;
	}

	/* Custom scrollbar for modals */
	:global(.overflow-y-auto::-webkit-scrollbar) {
		width: 6px;
	}

	:global(.overflow-y-auto::-webkit-scrollbar-track) {
		background: #f1f5f9;
		border-radius: 3px;
	}

	:global(.overflow-y-auto::-webkit-scrollbar-thumb) {
		background: #cbd5e1;
		border-radius: 3px;
	}

	:global(.overflow-y-auto::-webkit-scrollbar-thumb:hover) {
		background: #94a3b8;
	}

	/* ===== UNIVERSAL ACCORDION STYLE GUIDE ===== */
	
	/* Accordion Container */
	:global(.subscriptions-accordion) {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	/* Individual Accordion Item */
	:global(.subscription-accordion) {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		overflow: hidden;
		transition: all 0.2s ease;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	:global(.subscription-accordion:hover) {
		transform: translateY(-1px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	/* Accordion Header */
	:global(.subscription-header) {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 1.5rem;
		background: none;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		color: #111827;
	}

	:global(.subscription-header:hover) {
		background: #f9fafb;
	}

	:global(.subscription-header:focus) {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	/* Subscription Info Section */
	:global(.subscription-info) {
		display: flex;
		align-items: center;
		gap: 1rem;
		flex: 1;
	}

	:global(.subscription-icon) {
		font-size: 1.5rem;
		flex-shrink: 0;
	}

	:global(.subscription-details) {
		flex: 1;
	}

	:global(.subscription-name) {
		font-size: 1.125rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
		line-height: 1.4;
	}

	:global(.subscription-description) {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0.25rem 0 0 0;
		line-height: 1.4;
	}

	/* Subscription Stats */
	:global(.subscription-stats) {
		display: flex;
		gap: 1rem;
		margin-right: 1rem;
		font-size: 0.875rem;
	}

	:global(.plan-count) {
		color: #111827;
		font-weight: 500;
	}

	:global(.monthly-count) {
		color: #2563eb;
		font-weight: 500;
	}

	:global(.annual-count) {
		color: #7c3aed;
		font-weight: 500;
	}

	/* Accordion Icon */
	:global(.accordion-icon) {
		font-size: 0.875rem;
		color: #6b7280;
		transition: transform 0.2s ease;
	}

	/* Accordion Content */
	:global(.subscription-content) {
		padding: 0 1.5rem 1.5rem 1.5rem;
		background: #f9fafb;
		border-top: 1px solid #e5e7eb;
	}

	/* Plans Grid */
	:global(.plans-grid) {
		display: flex;
		flex-direction: column;
		gap: 2rem;
	}

	/* Plan Section */
	:global(.plan-section) {
		display: flex;
		flex-direction: column;
		gap: 1rem;
	}

	:global(.plan-section-title) {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 1rem;
		font-weight: 600;
		color: #111827;
		margin: 0;
		padding: 0.75rem 0;
		border-bottom: 1px solid #e5e7eb;
	}

	:global(.plan-section-icon) {
		font-size: 1.25rem;
	}

	/* Plans Row */
	:global(.plans-row) {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
		gap: 1rem;
	}

	/* Plan Card */
	:global(.plan-card) {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 0.5rem;
		padding: 1.25rem;
		transition: all 0.2s ease;
		position: relative;
		overflow: hidden;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	:global(.plan-card:hover) {
		transform: translateY(-2px);
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
		border-color: #d1d5db;
	}

	/* Plan Card Variants */
	:global(.plan-card.active-plan) {
		border-left: 4px solid #10b981;
	}

	:global(.plan-card.promoted-plan) {
		border-left: 4px solid #f59e0b;
		background: #fffbeb;
	}

	:global(.plan-card.inactive-plan) {
		border-left: 4px solid #6b7280;
		opacity: 0.7;
	}

	/* Plan Header */
	:global(.plan-header) {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	:global(.plan-icon) {
		font-size: 1.5rem;
		flex-shrink: 0;
	}

	:global(.plan-info) {
		flex: 1;
		min-width: 0;
	}

	:global(.plan-name) {
		font-weight: 600;
		color: #111827;
		margin: 0 0 0.25rem 0;
		font-size: 1rem;
		line-height: 1.4;
	}

	:global(.plan-description) {
		color: #6b7280;
		font-size: 0.875rem;
		margin: 0;
		line-height: 1.4;
		display: -webkit-box;
		-webkit-line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	:global(.plan-price) {
		text-align: right;
		flex-shrink: 0;
	}

	:global(.price-amount) {
		display: block;
		font-weight: 700;
		font-size: 1.125rem;
		color: #111827;
	}

	:global(.price-interval) {
		display: block;
		font-size: 0.75rem;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	/* Plan Stats */
	:global(.plan-stats) {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 0.75rem;
		margin-bottom: 1rem;
		padding: 0.75rem;
		background: #f9fafb;
		border-radius: 0.375rem;
	}

	:global(.stat) {
		display: flex;
		flex-direction: column;
		align-items: center;
		text-align: center;
	}

	:global(.stat-label) {
		font-size: 0.75rem;
		color: #6b7280;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.25rem;
	}

	:global(.stat-value) {
		font-size: 0.875rem;
		font-weight: 600;
		color: #111827;
	}

	/* Status Badges */
	:global(.status-badge) {
		padding: 0.25rem 0.5rem;
		border-radius: 0.25rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.05em;
	}

	:global(.status-badge.active) {
		background: #dcfce7;
		color: #166534;
		border: 1px solid #bbf7d0;
	}

	:global(.status-badge.promoted) {
		background: #fef3c7;
		color: #92400e;
		border: 1px solid #fde68a;
	}

	:global(.status-badge.inactive) {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #e5e7eb;
	}

	/* Plan Footer */
	:global(.plan-footer) {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-top: 1rem;
		border-top: 1px solid #e5e7eb;
	}

	:global(.plan-actions) {
		display: flex;
		gap: 0.5rem;
	}

	/* Button System */
	:global(.btn) {
		padding: 0.5rem 1rem;
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		border: none;
		cursor: pointer;
		transition: all 0.2s ease;
		text-decoration: none;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
	}

	:global(.btn:hover) {
		transform: translateY(-1px);
	}

	:global(.btn:focus) {
		outline: 2px solid #2563eb;
		outline-offset: 2px;
	}

	:global(.btn-sm) {
		padding: 0.375rem 0.75rem;
		font-size: 0.75rem;
	}

	:global(.btn-primary) {
		background: #2563eb;
		color: white;
	}

	:global(.btn-primary:hover) {
		background: #1d4ed8;
		box-shadow: 0 4px 12px rgba(37, 99, 235, 0.3);
	}

	:global(.btn-secondary) {
		background: #f3f4f6;
		color: #374151;
		border: 1px solid #d1d5db;
	}

	:global(.btn-secondary:hover) {
		background: #e5e7eb;
		border-color: #9ca3af;
	}

	:global(.btn-danger) {
		background: #dc2626;
		color: white;
	}

	:global(.btn-danger:hover) {
		background: #b91c1c;
		box-shadow: 0 4px 12px rgba(220, 38, 38, 0.3);
	}

	/* Empty State */
	:global(.empty-state) {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 3rem 1rem;
		text-align: center;
		color: #6b7280;
	}

	:global(.empty-icon) {
		font-size: 3rem;
		margin-bottom: 1rem;
		opacity: 0.5;
	}

	:global(.empty-message) {
		font-size: 1rem;
		margin: 0;
		font-weight: 500;
	}

	/* Responsive Design */
	@media (max-width: 768px) {
		:global(.subscription-header) {
			padding: 1rem;
			flex-direction: column;
			align-items: flex-start;
			gap: 1rem;
		}

		:global(.subscription-stats) {
			margin-right: 0;
			align-self: flex-end;
		}

		:global(.subscription-content) {
			padding: 0 1rem 1rem 1rem;
		}

		:global(.plans-row) {
			grid-template-columns: 1fr;
		}

		:global(.plan-stats) {
			grid-template-columns: repeat(2, 1fr);
		}

		:global(.plan-footer) {
			flex-direction: column;
			gap: 1rem;
			align-items: stretch;
		}

		:global(.plan-actions) {
			justify-content: center;
		}
	}

	/* ===== END UNIVERSAL ACCORDION STYLE GUIDE ===== */

	/* ===== OPTIMISTIC UPDATES STYLING ===== */
	
	/* Plan card updating state */
	:global(.plan-card.updating) {
		position: relative;
		opacity: 0.7;
		pointer-events: none;
	}

	:global(.plan-card.updating::before) {
		content: '';
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(37, 99, 235, 0.1);
		border-radius: 0.5rem;
		z-index: 1;
	}

	/* Optimistic overlay */
	:global(.optimistic-overlay) {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		bottom: 0;
		background: rgba(37, 99, 235, 0.1);
		backdrop-filter: blur(4px);
		border-radius: 0.5rem;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		z-index: 2;
		color: #2563eb;
		font-weight: 500;
	}

	/* Loading spinner */
	:global(.loading-spinner) {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(37, 99, 235, 0.3);
		border-top: 2px solid #2563eb;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 0.5rem;
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	/* Disabled button states */
	:global(.btn:disabled) {
		opacity: 0.5;
		cursor: not-allowed;
		transform: none !important;
	}

	:global(.btn:disabled:hover) {
		transform: none !important;
		box-shadow: none !important;
	}

	/* ===== END OPTIMISTIC UPDATES STYLING ===== */
</style> 
