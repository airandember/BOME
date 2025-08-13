<script lang="ts">
	export let title: string;
	export let customers: any[] = [];
	export let showActions = false;
	
	function formatDate(dateString: string): string {
		if (!dateString) return 'N/A';
		return new Date(dateString).toLocaleDateString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric'
		});
	}
</script>

{#if customers.length > 0}
	<div style="margin: 2rem 0;">
		<h2 style="margin-bottom: 1rem; color: #111827;">{title} ({customers.length})</h2>
		<div style="overflow-x: auto; border: 1px solid #e5e7eb; border-radius: 0.5rem;">
			<table style="width: 100%; border-collapse: collapse; font-size: 0.875rem;">
				<thead>
					<tr style="background-color: #f9fafb;">
						<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Customer</th>
						<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Email</th>
						<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Local ID</th>
						<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Role</th>
						<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Plan</th>
						<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Created</th>
						{#if showActions}
							<th style="padding: 0.75rem 1rem; text-align: left; font-weight: 600; color: #6b7280; border-bottom: 1px solid #e5e7eb;">Actions</th>
						{/if}
					</tr>
				</thead>
				<tbody>
					{#each customers as customer}
						<tr style="border-bottom: 1px solid #e5e7eb;">
							<td style="padding: 0.75rem 1rem;">
								<div>
									<div style="font-weight: 600; color: #111827;">{customer.name || 'Unnamed Customer'}</div>
									<div style="font-size: 0.75rem; color: #6b7280; text-transform: uppercase;">
										{customer.source || 'Unknown'}
									</div>
								</div>
							</td>
							<td style="padding: 0.75rem 1rem; color: #111827;">{customer.email}</td>
							<td style="padding: 0.75rem 1rem; color: #111827;">{customer.localId || 'N/A'}</td>
							<td style="padding: 0.75rem 1rem; color: #111827;">{customer.role || 'N/A'}</td>
							<td style="padding: 0.75rem 1rem; color: #111827;">{customer.planName || 'N/A'}</td>
							<td style="padding: 0.75rem 1rem; color: #111827;">{formatDate(customer.createdAt)}</td>
							{#if showActions}
								<td style="padding: 0.75rem 1rem;">
									{#if customer.source === 'stripe'}
										<button style="padding: 0.5rem 1rem; background: #059669; color: white; border: none; border-radius: 0.25rem; font-size: 0.75rem; cursor: pointer;">
											➕ Add User
										</button>
									{:else}
										<span style="color: #059669; font-weight: 600;">✅ Synced</span>
									{/if}
								</td>
							{/if}
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
{/if} 