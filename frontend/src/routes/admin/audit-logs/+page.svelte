<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	interface AuditLog {
		id: number;
		user_id: number | null;
		user_name: string | null;
		user_email: string | null;
		action_type: 'create' | 'update' | 'delete';
		entity_type: string;
		entity_id: number;
		old_values: any;
		new_values: any;
		ip_address: string | null;
		created_at: string;
	}

	interface AuditLogsResponse {
		logs: AuditLog[];
		total_count: number;
		page: number;
		page_size: number;
		total_pages: number;
	}

	let loading = true;
	let logs: AuditLog[] = [];
	let totalCount = 0;
	let currentPage = 1;
	let pageSize = 100;
	let totalPages = 1;
	let error = '';

	// Filters
	let filterUserId = '';
	let filterEntityType = '';
	let filterActionType = '';
	let filterStartDate = '';
	let filterEndDate = '';

	// Expanded rows
	let expandedRows = new Set<number>();

	// Available entity types (will be populated from API)
	let entityTypes: string[] = [];

	// Export modal state
	let showExportModal = false;
	let exportFormat: 'csv' | 'json' = 'csv';
	let exporting = false;
	let exportWarning = '';

	// Purge modal state
	let showPurgeModal = false;
	let purging = false;
	let purgeResult: any = null;
	let purgeError = '';

	$: currentUserIsSuperAdmin = data.user?.role === 'assignor' || false;

	onMount(async () => {
		await loadAuditLogs();
	});

	async function loadAuditLogs() {
		loading = true;
		error = '';

		try {
			// Build query parameters
			const params = new URLSearchParams({
				page: currentPage.toString(),
				page_size: pageSize.toString()
			});

			if (filterUserId) params.append('user_id', filterUserId);
			if (filterEntityType) params.append('entity_type', filterEntityType);
			if (filterActionType) params.append('action_type', filterActionType);
			if (filterStartDate) params.append('start_date', filterStartDate);
			if (filterEndDate) params.append('end_date', filterEndDate);

			const response = await fetch(`${API_URL}/api/admin/audit-logs?${params}`, {
				credentials: 'include'
			});

			if (response.ok) {
				const data: AuditLogsResponse = await response.json();
				logs = data.logs || [];
				totalCount = data.total_count;
				currentPage = data.page;
				pageSize = data.page_size;
				totalPages = data.total_pages;

				// Extract unique entity types
				const types = new Set(logs.map(log => log.entity_type));
				entityTypes = Array.from(types).sort();
			} else if (response.status === 403) {
				error = 'Access denied. Only System Admins can view audit logs.';
			} else {
				error = 'Failed to load audit logs';
			}
		} catch (err) {
			console.error('Error loading audit logs:', err);
			error = 'Failed to load audit logs';
		} finally {
			loading = false;
		}
	}

	function handleFilterChange() {
		currentPage = 1;
		loadAuditLogs();
	}

	function clearFilters() {
		filterUserId = '';
		filterEntityType = '';
		filterActionType = '';
		filterStartDate = '';
		filterEndDate = '';
		currentPage = 1;
		loadAuditLogs();
	}

	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
			loadAuditLogs();
		}
	}

	function toggleRowExpansion(logId: number) {
		if (expandedRows.has(logId)) {
			expandedRows.delete(logId);
		} else {
			expandedRows.add(logId);
		}
		expandedRows = expandedRows;
	}

	function formatDate(dateString: string): string {
		const date = new Date(dateString);
		return date.toLocaleString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
			second: '2-digit'
		});
	}

	function formatJSON(obj: any): string {
		if (!obj) return 'N/A';
		return JSON.stringify(obj, null, 2);
	}

	function openExportModal() {
		exportWarning = '';
		showExportModal = true;
	}

	function closeExportModal() {
		showExportModal = false;
		exportFormat = 'csv';
		exportWarning = '';
	}

	async function handleExport() {
		exporting = true;
		exportWarning = '';

		try {
			const params = new URLSearchParams({
				format: exportFormat
			});

			if (filterUserId) params.append('user_id', filterUserId);
			if (filterEntityType) params.append('entity_type', filterEntityType);
			if (filterActionType) params.append('action_type', filterActionType);
			if (filterStartDate) params.append('start_date', filterStartDate);
			if (filterEndDate) params.append('end_date', filterEndDate);

			const response = await fetch(`${API_URL}/api/admin/audit-logs/export?${params}`, {
				credentials: 'include'
			});

			if (response.ok) {
				const warningHeader = response.headers.get('X-Export-Warning');
				if (warningHeader) {
					exportWarning = warningHeader;
				}

				const contentDisposition = response.headers.get('Content-Disposition');
				let filename = `audit_logs.${exportFormat}`;
				if (contentDisposition) {
					const matches = /filename="([^"]+)"/.exec(contentDisposition);
					if (matches && matches[1]) {
						filename = matches[1];
					}
				}

				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = filename;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);

				if (!exportWarning) {
					closeExportModal();
				}
			} else if (response.status === 403) {
				exportWarning = 'Access denied. Only System Admins can export audit logs.';
			} else {
				exportWarning = 'Failed to export audit logs';
			}
		} catch (err) {
			console.error('Error exporting audit logs:', err);
			exportWarning = 'Failed to export audit logs';
		} finally {
			exporting = false;
		}
	}

	function openPurgeModal() {
		purgeResult = null;
		purgeError = '';
		showPurgeModal = true;
	}

	function closePurgeModal() {
		showPurgeModal = false;
		purgeResult = null;
		purgeError = '';
	}

	async function handlePurge() {
		purging = true;
		purgeError = '';
		purgeResult = null;

		try {
			const response = await fetch(`${API_URL}/api/admin/audit-logs/purge`, {
				method: 'POST',
				credentials: 'include'
			});

			if (response.ok) {
				purgeResult = await response.json();
				await loadAuditLogs();
			} else if (response.status === 403) {
				purgeError = 'Access denied. Only System Admins can purge audit logs.';
			} else {
				purgeError = 'Failed to purge audit logs';
			}
		} catch (err) {
			console.error('Error purging audit logs:', err);
			purgeError = 'Failed to purge audit logs';
		} finally {
			purging = false;
		}
	}

	function formatDuration(ms: number): string {
		if (ms < 1000) {
			return `${ms}ms`;
		}
		const seconds = (ms / 1000).toFixed(2);
		return `${seconds}s`;
	}
</script>

<svelte:head>
	<title>Audit Logs - Admin - Referee Scheduler</title>
</svelte:head>

<div class="page-header">
	<h2>Audit Logs</h2>
	{#if currentUserIsSuperAdmin}
		<div class="header-actions">
			<button on:click={openExportModal} class="btn btn-primary">Export</button>
			<button on:click={openPurgeModal} class="btn btn-danger">Purge Old Logs</button>
		</div>
	{/if}
</div>

{#if error}
	<div class="alert alert-error" role="alert">{error}</div>
{/if}

{#if !currentUserIsSuperAdmin}
	<div class="alert alert-warning" role="alert">
		<strong>Access Restricted</strong>
		<p>You need System Admin privileges to view audit logs.</p>
	</div>
{:else}
	<!-- Filters -->
	<div class="card filter-card">
		<h3>Filters</h3>
		<div class="filter-grid">
			<div class="filter-field">
				<label for="filter-entity-type">Entity Type</label>
				<select
					id="filter-entity-type"
					bind:value={filterEntityType}
					on:change={handleFilterChange}
				>
					<option value="">All Types</option>
					<option value="user_role">User Role</option>
					<option value="match">Match</option>
					<option value="assignment">Assignment</option>
					<option value="user">User</option>
					{#each entityTypes as type}
						{#if !['user_role', 'match', 'assignment', 'user'].includes(type)}
							<option value={type}>{type}</option>
						{/if}
					{/each}
				</select>
			</div>

			<div class="filter-field">
				<label for="filter-action-type">Action Type</label>
				<select
					id="filter-action-type"
					bind:value={filterActionType}
					on:change={handleFilterChange}
				>
					<option value="">All Actions</option>
					<option value="create">Create</option>
					<option value="update">Update</option>
					<option value="delete">Delete</option>
				</select>
			</div>

			<div class="filter-field">
				<label for="filter-user-id">User ID</label>
				<input
					type="number"
					id="filter-user-id"
					bind:value={filterUserId}
					on:input={handleFilterChange}
					placeholder="Filter by user ID"
				/>
			</div>

			<div class="filter-field">
				<label for="filter-start-date">Start Date</label>
				<input
					type="date"
					id="filter-start-date"
					bind:value={filterStartDate}
					on:change={handleFilterChange}
				/>
			</div>

			<div class="filter-field">
				<label for="filter-end-date">End Date</label>
				<input
					type="date"
					id="filter-end-date"
					bind:value={filterEndDate}
					on:change={handleFilterChange}
				/>
			</div>

			<div class="filter-field filter-actions">
				<button on:click={clearFilters} class="btn btn-secondary">Clear Filters</button>
			</div>
		</div>

		<div class="results-summary">
			Showing {logs.length} of {totalCount} total entries
		</div>
	</div>

	{#if loading}
		<div class="loading-text">Loading audit logs...</div>
	{:else if logs.length === 0}
		<div class="card empty-state">No audit logs found matching your filters.</div>
	{:else}
		<div class="card table-card">
			<table class="data-table">
				<thead>
					<tr>
						<th>Timestamp</th>
						<th>User</th>
						<th>Action</th>
						<th>Entity Type</th>
						<th>Entity ID</th>
						<th>IP Address</th>
						<th>Details</th>
					</tr>
				</thead>
				<tbody>
					{#each logs as log (log.id)}
						<tr>
							<td class="nowrap">{formatDate(log.created_at)}</td>
							<td>
								{#if log.user_name}
									<div class="user-name">{log.user_name}</div>
									<div class="user-email">{log.user_email}</div>
								{:else}
									<span class="text-muted">System</span>
								{/if}
							</td>
							<td>
								<span class="action-badge action-{log.action_type}">
									{log.action_type.toUpperCase()}
								</span>
							</td>
							<td>{log.entity_type}</td>
							<td>{log.entity_id}</td>
							<td class="text-muted">{log.ip_address || 'N/A'}</td>
							<td>
								<button
									on:click={() => toggleRowExpansion(log.id)}
									class="link-btn"
								>
									{expandedRows.has(log.id) ? 'Hide' : 'Show'} JSON
								</button>
							</td>
						</tr>
						{#if expandedRows.has(log.id)}
							<tr class="detail-row">
								<td colspan="7">
									<div class="json-grid">
										<div>
											<h4>Old Values:</h4>
											<pre>{formatJSON(log.old_values)}</pre>
										</div>
										<div>
											<h4>New Values:</h4>
											<pre>{formatJSON(log.new_values)}</pre>
										</div>
									</div>
								</td>
							</tr>
						{/if}
					{/each}
				</tbody>
			</table>
		</div>

		{#if totalPages > 1}
			<div class="card pagination">
				<span class="page-info">Page {currentPage} of {totalPages}</span>
				<div class="page-buttons">
					<button on:click={() => goToPage(1)} disabled={currentPage === 1} class="btn btn-secondary btn-sm">First</button>
					<button on:click={() => goToPage(currentPage - 1)} disabled={currentPage === 1} class="btn btn-secondary btn-sm">Previous</button>
					<button on:click={() => goToPage(currentPage + 1)} disabled={currentPage === totalPages} class="btn btn-secondary btn-sm">Next</button>
					<button on:click={() => goToPage(totalPages)} disabled={currentPage === totalPages} class="btn btn-secondary btn-sm">Last</button>
				</div>
			</div>
		{/if}
	{/if}
{/if}

<!-- Export Modal -->
{#if showExportModal}
	<div class="modal-overlay" on:click={closeExportModal}>
		<div class="modal" on:click|stopPropagation>
			<h3>Export Audit Logs</h3>

			{#if exportWarning}
				<div class="alert alert-warning">{exportWarning}</div>
			{/if}

			<p class="modal-description">Select export format for the current filtered results:</p>

			<div class="export-options">
				<label class="export-option">
					<input type="radio" name="exportFormat" value="csv" bind:group={exportFormat} />
					<div>
						<div class="option-title">CSV (Comma-Separated Values)</div>
						<div class="option-desc">Suitable for Excel and data analysis tools</div>
					</div>
				</label>

				<label class="export-option">
					<input type="radio" name="exportFormat" value="json" bind:group={exportFormat} />
					<div>
						<div class="option-title">JSON (JavaScript Object Notation)</div>
						<div class="option-desc">Suitable for programmatic processing</div>
					</div>
				</label>
			</div>

			<div class="alert alert-info">
				<strong>Note:</strong> Export is limited to 10,000 records. If more records match your filters, only the first 10,000 will be exported.
			</div>

			{#if filterEntityType || filterActionType || filterUserId || filterStartDate || filterEndDate}
				<div class="active-filters">
					<strong>Active Filters:</strong>
					<ul>
						{#if filterEntityType}<li>Entity Type: {filterEntityType}</li>{/if}
						{#if filterActionType}<li>Action Type: {filterActionType}</li>{/if}
						{#if filterUserId}<li>User ID: {filterUserId}</li>{/if}
						{#if filterStartDate}<li>Start Date: {filterStartDate}</li>{/if}
						{#if filterEndDate}<li>End Date: {filterEndDate}</li>{/if}
					</ul>
				</div>
			{/if}

			<div class="modal-actions">
				<button on:click={closeExportModal} class="btn btn-secondary" disabled={exporting}>Cancel</button>
				<button on:click={handleExport} class="btn btn-primary" disabled={exporting}>
					{exporting ? 'Exporting...' : 'Export'}
				</button>
			</div>
		</div>
	</div>
{/if}

<!-- Purge Modal -->
{#if showPurgeModal}
	<div class="modal-overlay" on:click={closePurgeModal}>
		<div class="modal" on:click|stopPropagation>
			<h3>Purge Old Audit Logs</h3>

			{#if purgeError}
				<div class="alert alert-error">{purgeError}</div>
			{/if}

			{#if purgeResult}
				<div class="alert alert-success">
					<strong>Purge Completed Successfully!</strong>
					<ul>
						<li><strong>Deleted:</strong> {purgeResult.deleted_count.toLocaleString()} log entries</li>
						<li><strong>Cutoff Date:</strong> {new Date(purgeResult.cutoff_date).toLocaleDateString()}</li>
						<li><strong>Duration:</strong> {formatDuration(purgeResult.duration_ms)}</li>
					</ul>
				</div>
			{:else if !purging}
				<div class="alert alert-warning">
					<strong>Warning:</strong> This action will permanently delete audit logs older than the configured retention period.
				</div>
				<div class="purge-details">
					<strong>Current Retention Policy:</strong>
					<p>Logs older than <strong>2 years</strong> will be deleted.</p>
				</div>
				<div class="alert alert-info">
					<strong>Note:</strong> This operation may take a few seconds for large datasets. The deletion is batched to minimize database impact.
				</div>
			{/if}

			{#if purging}
				<div class="loading-spinner">
					<div class="spinner"></div>
					<p>Purging old logs...</p>
				</div>
			{/if}

			<div class="modal-actions">
				{#if purgeResult}
					<button on:click={closePurgeModal} class="btn btn-primary">Close</button>
				{:else}
					<button on:click={closePurgeModal} class="btn btn-secondary" disabled={purging}>Cancel</button>
					<button on:click={handlePurge} class="btn btn-danger" disabled={purging}>
						{purging ? 'Purging...' : 'Confirm Purge'}
					</button>
				{/if}
			</div>
		</div>
	</div>
{/if}

<style>
	/* Page Header */
	.page-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	h2 {
		font-size: 1.75rem;
		font-weight: 700;
		color: var(--text-primary);
	}

	.header-actions {
		display: flex;
		gap: 0.5rem;
	}

	/* Filter Card */
	.filter-card {
		margin-bottom: 1rem;
	}

	.filter-card h3 {
		font-size: 1.125rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 0.75rem;
	}

	.filter-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 1rem;
	}

	.filter-field {
		display: flex;
		flex-direction: column;
		gap: 0.375rem;
	}

	.filter-field label {
		font-size: 0.75rem;
		font-weight: 600;
		color: var(--text-secondary);
	}

	.filter-field select,
	.filter-field input {
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		font-size: 0.875rem;
		color: var(--text-primary);
		background-color: white;
		transition: border-color 0.2s;
	}

	.filter-field select:focus,
	.filter-field input:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 2px rgba(37, 99, 235, 0.15);
	}

	.filter-actions {
		display: flex;
		align-items: flex-end;
	}

	.results-summary {
		margin-top: 0.75rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	/* Loading & Empty */
	.loading-text {
		text-align: center;
		padding: 3rem;
		color: var(--text-secondary);
	}

	.empty-state {
		text-align: center;
		padding: 3rem;
		color: var(--text-secondary);
	}

	/* Table */
	.table-card {
		overflow: hidden;
		padding: 0;
	}

	.data-table tr:hover:not(.detail-row) {
		background-color: var(--bg-secondary);
	}

	.nowrap {
		white-space: nowrap;
	}

	.user-name {
		color: var(--text-primary);
		font-weight: 500;
	}

	.user-email {
		color: var(--text-secondary);
		font-size: 0.75rem;
	}

	.text-muted {
		color: var(--text-secondary);
	}

	/* Action Badges */
	.action-badge {
		display: inline-block;
		padding: 0.25rem 0.625rem;
		border-radius: 1rem;
		font-size: 0.7rem;
		font-weight: 600;
	}

	.action-create {
		background-color: #d1fae5;
		color: #065f46;
	}

	.action-update {
		background-color: #dbeafe;
		color: #1e40af;
	}

	.action-delete {
		background-color: #fee2e2;
		color: #991b1b;
	}

	.link-btn {
		background: none;
		border: none;
		color: var(--primary-color);
		font-weight: 500;
		cursor: pointer;
		padding: 0;
		font-size: 0.875rem;
	}

	.link-btn:hover {
		text-decoration: underline;
	}

	/* Detail Row (expanded JSON) */
	.detail-row td {
		background-color: var(--bg-secondary);
		padding: 1rem;
	}

	.json-grid {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 1rem;
	}

	.json-grid h4 {
		font-size: 0.875rem;
		font-weight: 600;
		color: var(--text-secondary);
		margin-bottom: 0.5rem;
	}

	.json-grid pre {
		background-color: var(--bg-primary);
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		padding: 0.75rem;
		font-size: 0.75rem;
		overflow-x: auto;
		white-space: pre-wrap;
		word-break: break-word;
	}

	/* Pagination */
	.pagination {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 1rem;
	}

	.page-info {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.page-buttons {
		display: flex;
		gap: 0.375rem;
	}

	/* Modal */
	.modal {
		max-width: 30rem;
	}

	.modal h3 {
		font-size: 1.25rem;
		font-weight: 600;
		color: var(--text-primary);
		margin-bottom: 1rem;
	}

	.modal-description {
		font-size: 0.875rem;
		color: var(--text-secondary);
		margin-bottom: 1rem;
	}

	/* Export Options */
	.export-options {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		margin-bottom: 1rem;
	}

	.export-option {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		cursor: pointer;
		padding: 0.5rem;
		border-radius: 0.375rem;
		transition: background-color 0.2s;
	}

	.export-option:hover {
		background-color: var(--bg-secondary);
	}

	.export-option input {
		margin-top: 0.25rem;
	}

	.option-title {
		font-weight: 500;
		color: var(--text-primary);
	}

	.option-desc {
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.active-filters {
		background-color: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		padding: 0.75rem 1rem;
		margin-bottom: 1rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.active-filters ul {
		list-style: disc;
		padding-left: 1.25rem;
		margin-top: 0.25rem;
	}

	/* Purge */
	.purge-details {
		padding: 0.75rem 1rem;
		margin-bottom: 1rem;
		font-size: 0.875rem;
		color: var(--text-primary);
	}

	.purge-details p {
		margin-top: 0.25rem;
		color: var(--text-secondary);
	}

	.loading-spinner {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 1rem;
		padding: 2rem;
	}

	.spinner {
		width: 2rem;
		height: 2rem;
		border: 3px solid var(--border-color);
		border-top-color: var(--primary-color);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to { transform: rotate(360deg); }
	}

	@media (max-width: 768px) {
		.filter-grid {
			grid-template-columns: 1fr;
		}

		.json-grid {
			grid-template-columns: 1fr;
		}

		.data-table th,
		.data-table td {
			padding: 0.5rem;
		}

		.pagination {
			flex-direction: column;
			gap: 0.75rem;
		}

		.modal {
			padding: 1rem;
		}
	}
</style>
