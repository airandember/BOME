<script lang="ts">
	import { apiRequest } from '$lib/auth';
	import { showToast } from '$lib/toast';
	import { onMount } from 'svelte';

	let isExporting = false;
	let exportOptions = {
		format: 'sql',
		includeData: true
	};

	// Check if user is authenticated and is admin
	onMount(() => {
		// This will be handled by the layout if needed
	});

	async function exportDatabase() {
		if (isExporting) return;
		
		isExporting = true;
		showToast('Starting database export...', 'info');

		try {
			const params = new URLSearchParams({
				format: exportOptions.format,
				include_data: exportOptions.includeData.toString()
			});

			const response = await apiRequest(`/admin/database/export?${params}`, {
				method: 'GET'
			});

			if (response.ok) {
				// Create blob and download
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				
				// Get filename from Content-Disposition header
				const contentDisposition = response.headers.get('Content-Disposition');
				const filename = contentDisposition 
					? contentDisposition.split('filename=')[1].replace(/"/g, '')
					: `bome_database_export_${new Date().toISOString().split('T')[0]}.${exportOptions.format}`;
				
				a.download = filename;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);

				showToast('Database exported successfully!', 'success');
			} else {
				const errorData = await response.json();
				throw new Error(errorData.error || 'Export failed');
			}
		} catch (error) {
			console.error('Export error:', error);
			const errorMessage = error instanceof Error ? error.message : 'Unknown error occurred';
			showToast(`Failed to export database: ${errorMessage}`, 'error');
		} finally {
			isExporting = false;
		}
	}
</script>

<svelte:head>
	<title>Database Management - BOME Admin</title>
	<meta name="description" content="Export and manage database for BOME" />
</svelte:head>

<div class="database-export">
	<div class="header">
		<h1>🗄️ Database Management</h1>
		<p>Export and manage your database safely</p>
	</div>

	<div class="export-section">
		<h2>Export Database</h2>
		<p class="section-description">
			Create a complete backup of your database. This includes all users, videos, subscriptions, and system data.
		</p>
		
		<div class="export-options">
			<div class="option-group">
				<div class="group-label">Export Format:</div>
				<div class="radio-group">
					<label class="radio-option">
						<input 
							type="radio" 
							bind:group={exportOptions.format} 
							value="sql"
						/>
						<span class="radio-label">
							<strong>SQL Dump</strong>
							<small>PostgreSQL dump file (.sql) - Best for database restoration</small>
						</span>
					</label>
					<label class="radio-option">
						<input 
							type="radio" 
							bind:group={exportOptions.format} 
							value="json"
						/>
						<span class="radio-label">
							<strong>JSON Export</strong>
							<small>Structured JSON file (.json) - Best for data analysis</small>
						</span>
					</label>
				</div>
			</div>

			<div class="option-group">
				<label class="checkbox-option">
					<input 
						type="checkbox" 
						bind:checked={exportOptions.includeData}
					/>
					<span class="checkbox-label">
						<strong>Include Data</strong>
						<small>Export both database structure and all data. Uncheck for schema-only export.</small>
					</span>
				</label>
			</div>
		</div>

		<button 
			class="export-btn"
			on:click={exportDatabase}
			disabled={isExporting}
		>
			{#if isExporting}
				<span class="spinner"></span>
				Exporting Database...
			{:else}
				📥 Export Database
			{/if}
		</button>

		<div class="export-info">
			<h3>📋 Export Information</h3>
			<div class="info-grid">
				<div class="info-card">
					<h4>🔧 SQL Format</h4>
					<ul>
						<li>Creates a PostgreSQL dump file</li>
						<li>Can be restored with <code>psql</code> command</li>
						<li>Includes table structure and data</li>
						<li>Best for complete database backups</li>
					</ul>
				</div>
				<div class="info-card">
					<h4>📊 JSON Format</h4>
					<ul>
						<li>Creates a structured JSON file</li>
						<li>Human-readable data format</li>
						<li>Easy to parse and analyze</li>
						<li>Best for data analysis and migration</li>
					</ul>
				</div>
			</div>
			
			<div class="security-notice">
				<h4>🔒 Security Notice</h4>
				<p>
					<strong>Important:</strong> Database exports contain sensitive information including user data, 
					passwords (hashed), and system configuration. Store exported files securely and never share 
					them publicly. Only super administrators can perform database exports.
				</p>
			</div>
		</div>
	</div>
</div>

<style>
	.database-export {
		max-width: 1000px;
		margin: 0 auto;
		padding: 2rem;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
	}

	.header {
		text-align: center;
		margin-bottom: 3rem;
	}

	.header h1 {
		font-size: 2.5rem;
		margin-bottom: 0.5rem;
		color: #1f2937;
		font-weight: 700;
	}

	.header p {
		color: #6b7280;
		font-size: 1.1rem;
	}

	.export-section {
		background: white;
		border-radius: 16px;
		padding: 2.5rem;
		box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
		border: 1px solid #e5e7eb;
	}

	.export-section h2 {
		margin-bottom: 0.5rem;
		color: #111827;
		font-size: 1.5rem;
		font-weight: 600;
	}

	.section-description {
		color: #6b7280;
		margin-bottom: 2rem;
		font-size: 1rem;
		line-height: 1.6;
	}

	.export-options {
		display: grid;
		gap: 2rem;
		margin-bottom: 2.5rem;
	}

	.option-group .group-label {
		font-weight: 600;
		color: #374151;
		margin-bottom: 1rem;
		display: block;
		font-size: 1rem;
	}

	.radio-group {
		display: grid;
		gap: 1rem;
	}

	.radio-option, .checkbox-option {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 1rem;
		border: 2px solid #e5e7eb;
		border-radius: 12px;
		cursor: pointer;
		transition: all 0.2s ease;
		background: #fafafa;
	}

	.radio-option:hover, .checkbox-option:hover {
		border-color: #3b82f6;
		background: #f0f9ff;
	}

	.radio-option input[type="radio"]:checked + .radio-label,
	.checkbox-option input[type="checkbox"]:checked + .checkbox-label {
		color: #1f2937;
	}

	.radio-option:has(input:checked), .checkbox-option:has(input:checked) {
		border-color: #3b82f6;
		background: #eff6ff;
	}

	.radio-option input[type="radio"], .checkbox-option input[type="checkbox"] {
		margin: 0;
		width: 18px;
		height: 18px;
		flex-shrink: 0;
		margin-top: 2px;
	}

	.radio-label, .checkbox-label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		flex: 1;
	}

	.radio-label strong, .checkbox-label strong {
		color: #111827;
		font-weight: 600;
	}

	.radio-label small, .checkbox-label small {
		color: #6b7280;
		font-size: 0.875rem;
		line-height: 1.4;
	}

	.export-btn {
		background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
		color: white;
		border: none;
		padding: 1rem 2rem;
		border-radius: 12px;
		font-weight: 600;
		cursor: pointer;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.75rem;
		font-size: 1rem;
		transition: all 0.2s ease;
		box-shadow: 0 4px 6px -1px rgba(59, 130, 246, 0.3);
		min-width: 200px;
		margin: 0 auto 2.5rem;
	}

	.export-btn:hover:not(:disabled) {
		background: linear-gradient(135deg, #2563eb 0%, #1e40af 100%);
		box-shadow: 0 6px 10px -1px rgba(59, 130, 246, 0.4);
		transform: translateY(-1px);
	}

	.export-btn:disabled {
		background: #9ca3af;
		cursor: not-allowed;
		box-shadow: none;
		transform: none;
	}

	.spinner {
		width: 20px;
		height: 20px;
		border: 2px solid rgba(255, 255, 255, 0.3);
		border-top: 2px solid currentColor;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.export-info {
		border-top: 1px solid #e5e7eb;
		padding-top: 2rem;
	}

	.export-info h3 {
		margin-bottom: 1.5rem;
		color: #111827;
		font-size: 1.25rem;
		font-weight: 600;
	}

	.info-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
		gap: 1.5rem;
		margin-bottom: 2rem;
	}

	.info-card {
		background: #f8fafc;
		padding: 1.5rem;
		border-radius: 12px;
		border: 1px solid #e2e8f0;
	}

	.info-card h4 {
		margin-bottom: 1rem;
		color: #1e293b;
		font-size: 1rem;
		font-weight: 600;
	}

	.info-card ul {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.info-card li {
		margin-bottom: 0.5rem;
		color: #64748b;
		font-size: 0.875rem;
		line-height: 1.5;
		position: relative;
		padding-left: 1rem;
	}

	.info-card li:before {
		content: '•';
		color: #3b82f6;
		position: absolute;
		left: 0;
		font-weight: bold;
	}

	.info-card code {
		background: #e2e8f0;
		padding: 0.125rem 0.375rem;
		border-radius: 4px;
		font-family: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', Consolas, 'Courier New', monospace;
		font-size: 0.8125rem;
		color: #1e293b;
	}

	.security-notice {
		background: linear-gradient(135deg, #fef3c7 0%, #fde68a 100%);
		border: 1px solid #f59e0b;
		border-radius: 12px;
		padding: 1.5rem;
	}

	.security-notice h4 {
		margin-bottom: 0.75rem;
		color: #92400e;
		font-size: 1rem;
		font-weight: 600;
	}

	.security-notice p {
		color: #78350f;
		font-size: 0.875rem;
		line-height: 1.6;
		margin: 0;
	}

	.security-notice strong {
		color: #92400e;
	}

	@media (max-width: 768px) {
		.database-export {
			padding: 1rem;
		}

		.header h1 {
			font-size: 2rem;
		}

		.export-section {
			padding: 1.5rem;
		}

		.info-grid {
			grid-template-columns: 1fr;
		}
	}
</style>
