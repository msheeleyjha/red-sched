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
		currentPage = 1; // Reset to first page when filters change
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
		expandedRows = expandedRows; // Trigger reactivity
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

	function getActionColor(action: string): string {
		switch (action) {
			case 'create':
				return 'bg-green-100 text-green-800';
			case 'update':
				return 'bg-blue-100 text-blue-800';
			case 'delete':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
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
			// Build query parameters with same filters as viewer
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
				// Check for warning header
				const warningHeader = response.headers.get('X-Export-Warning');
				if (warningHeader) {
					exportWarning = warningHeader;
				}

				// Get filename from Content-Disposition header
				const contentDisposition = response.headers.get('Content-Disposition');
				let filename = `audit_logs.${exportFormat}`;
				if (contentDisposition) {
					const matches = /filename="([^"]+)"/.exec(contentDisposition);
					if (matches && matches[1]) {
						filename = matches[1];
					}
				}

				// Download file
				const blob = await response.blob();
				const url = window.URL.createObjectURL(blob);
				const a = document.createElement('a');
				a.href = url;
				a.download = filename;
				document.body.appendChild(a);
				a.click();
				document.body.removeChild(a);
				window.URL.revokeObjectURL(url);

				// Close modal if no warning, otherwise keep open to show warning
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
</script>

<div class="container mx-auto p-6 max-w-7xl">
	<div class="flex justify-between items-center mb-6">
		<h1 class="text-3xl font-bold">Audit Logs</h1>
		{#if currentUserIsSuperAdmin}
			<button
				on:click={openExportModal}
				class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
			>
				Export
			</button>
		{/if}
	</div>

	{#if error}
		<div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4" role="alert">
			{error}
		</div>
	{/if}

	{#if !currentUserIsSuperAdmin}
		<div class="bg-yellow-100 border border-yellow-400 text-yellow-700 px-4 py-3 rounded mb-4" role="alert">
			<p class="font-bold">Access Restricted</p>
			<p>You need System Admin privileges to view audit logs.</p>
		</div>
	{:else}
		<!-- Filters -->
		<div class="bg-white shadow-md rounded-lg p-4 mb-4">
			<h2 class="text-lg font-semibold mb-3">Filters</h2>
			<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
				<!-- Entity Type Filter -->
				<div>
					<label for="filter-entity-type" class="block text-sm font-medium text-gray-700 mb-1">
						Entity Type
					</label>
					<select
						id="filter-entity-type"
						bind:value={filterEntityType}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
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

				<!-- Action Type Filter -->
				<div>
					<label for="filter-action-type" class="block text-sm font-medium text-gray-700 mb-1">
						Action Type
					</label>
					<select
						id="filter-action-type"
						bind:value={filterActionType}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All Actions</option>
						<option value="create">Create</option>
						<option value="update">Update</option>
						<option value="delete">Delete</option>
					</select>
				</div>

				<!-- User ID Filter -->
				<div>
					<label for="filter-user-id" class="block text-sm font-medium text-gray-700 mb-1">
						User ID
					</label>
					<input
						type="number"
						id="filter-user-id"
						bind:value={filterUserId}
						on:input={handleFilterChange}
						placeholder="Filter by user ID"
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<!-- Start Date Filter -->
				<div>
					<label for="filter-start-date" class="block text-sm font-medium text-gray-700 mb-1">
						Start Date
					</label>
					<input
						type="date"
						id="filter-start-date"
						bind:value={filterStartDate}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<!-- End Date Filter -->
				<div>
					<label for="filter-end-date" class="block text-sm font-medium text-gray-700 mb-1">
						End Date
					</label>
					<input
						type="date"
						id="filter-end-date"
						bind:value={filterEndDate}
						on:change={handleFilterChange}
						class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
					/>
				</div>

				<!-- Clear Filters Button -->
				<div class="flex items-end">
					<button
						on:click={clearFilters}
						class="w-full px-4 py-2 bg-gray-200 text-gray-700 rounded-md hover:bg-gray-300 focus:outline-none focus:ring-2 focus:ring-gray-500"
					>
						Clear Filters
					</button>
				</div>
			</div>

			<!-- Results Summary -->
			<div class="mt-3 text-sm text-gray-600">
				Showing {logs.length} of {totalCount} total entries
			</div>
		</div>

		{#if loading}
			<div class="text-center py-8">
				<p>Loading audit logs...</p>
			</div>
		{:else if logs.length === 0}
			<div class="bg-white shadow-md rounded-lg p-8 text-center text-gray-500">
				<p>No audit logs found matching your filters.</p>
			</div>
		{:else}
			<!-- Audit Logs Table -->
			<div class="bg-white shadow-md rounded-lg overflow-hidden">
				<table class="min-w-full">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Timestamp
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								User
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Action
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Entity Type
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Entity ID
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								IP Address
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Details
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each logs as log (log.id)}
							<tr class="hover:bg-gray-50">
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									{formatDate(log.created_at)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									{#if log.user_name}
										<div class="text-gray-900">{log.user_name}</div>
										<div class="text-gray-500 text-xs">{log.user_email}</div>
									{:else}
										<span class="text-gray-400">System</span>
									{/if}
								</td>
								<td class="px-6 py-4 whitespace-nowrap">
									<span class="px-2 py-1 text-xs font-semibold rounded-full {getActionColor(log.action_type)}">
										{log.action_type.toUpperCase()}
									</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									{log.entity_type}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									{log.entity_id}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{log.ip_address || 'N/A'}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<button
										on:click={() => toggleRowExpansion(log.id)}
										class="text-blue-600 hover:text-blue-900 font-medium"
									>
										{expandedRows.has(log.id) ? 'Hide' : 'Show'} JSON
									</button>
								</td>
							</tr>
							{#if expandedRows.has(log.id)}
								<tr>
									<td colspan="7" class="px-6 py-4 bg-gray-50">
										<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
											<!-- Old Values -->
											<div>
												<h4 class="font-semibold text-sm text-gray-700 mb-2">Old Values:</h4>
												<pre class="bg-white p-3 rounded border border-gray-300 text-xs overflow-x-auto">{formatJSON(log.old_values)}</pre>
											</div>
											<!-- New Values -->
											<div>
												<h4 class="font-semibold text-sm text-gray-700 mb-2">New Values:</h4>
												<pre class="bg-white p-3 rounded border border-gray-300 text-xs overflow-x-auto">{formatJSON(log.new_values)}</pre>
											</div>
										</div>
									</td>
								</tr>
							{/if}
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="bg-white shadow-md rounded-lg p-4 mt-4">
					<div class="flex justify-between items-center">
						<div class="text-sm text-gray-600">
							Page {currentPage} of {totalPages}
						</div>
						<div class="flex gap-2">
							<button
								on:click={() => goToPage(1)}
								disabled={currentPage === 1}
								class="px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								First
							</button>
							<button
								on:click={() => goToPage(currentPage - 1)}
								disabled={currentPage === 1}
								class="px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								Previous
							</button>
							<button
								on:click={() => goToPage(currentPage + 1)}
								disabled={currentPage === totalPages}
								class="px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								Next
							</button>
							<button
								on:click={() => goToPage(totalPages)}
								disabled={currentPage === totalPages}
								class="px-3 py-1 bg-gray-200 text-gray-700 rounded hover:bg-gray-300 disabled:opacity-50 disabled:cursor-not-allowed"
							>
								Last
							</button>
						</div>
					</div>
				</div>
			{/if}
		{/if}
	{/if}
</div>

<style>
	/* Additional styles if needed */
</style>

<!-- Export Modal -->
{#if showExportModal}
	<div class="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50" on:click={closeExportModal}>
		<div class="relative top-20 mx-auto p-5 border w-full max-w-md shadow-lg rounded-md bg-white" on:click|stopPropagation>
			<div class="mt-3">
				<h3 class="text-lg font-medium leading-6 text-gray-900 mb-4">
					Export Audit Logs
				</h3>

				{#if exportWarning}
					<div class="bg-yellow-100 border border-yellow-400 text-yellow-700 px-4 py-3 rounded mb-4" role="alert">
						<p class="text-sm">{exportWarning}</p>
					</div>
				{/if}

				<div class="mb-4">
					<p class="text-sm text-gray-600 mb-3">
						Select export format for the current filtered results:
					</p>

					<div class="space-y-2">
						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="exportFormat"
								value="csv"
								bind:group={exportFormat}
								class="mr-3"
							/>
							<div>
								<div class="font-medium">CSV (Comma-Separated Values)</div>
								<div class="text-sm text-gray-600">
									Suitable for Excel and data analysis tools
								</div>
							</div>
						</label>

						<label class="flex items-center cursor-pointer">
							<input
								type="radio"
								name="exportFormat"
								value="json"
								bind:group={exportFormat}
								class="mr-3"
							/>
							<div>
								<div class="font-medium">JSON (JavaScript Object Notation)</div>
								<div class="text-sm text-gray-600">
									Suitable for programmatic processing
								</div>
							</div>
						</label>
					</div>
				</div>

				<div class="bg-blue-50 border border-blue-200 px-4 py-3 rounded mb-4">
					<p class="text-sm text-blue-800">
						<strong>Note:</strong> Export is limited to 10,000 records. If more records
						match your filters, only the first 10,000 will be exported.
					</p>
				</div>

				{#if filterEntityType || filterActionType || filterUserId || filterStartDate || filterEndDate}
					<div class="bg-gray-50 border border-gray-200 px-4 py-3 rounded mb-4">
						<p class="text-sm text-gray-700">
							<strong>Active Filters:</strong>
						</p>
						<ul class="text-sm text-gray-600 mt-1 list-disc list-inside">
							{#if filterEntityType}
								<li>Entity Type: {filterEntityType}</li>
							{/if}
							{#if filterActionType}
								<li>Action Type: {filterActionType}</li>
							{/if}
							{#if filterUserId}
								<li>User ID: {filterUserId}</li>
							{/if}
							{#if filterStartDate}
								<li>Start Date: {filterStartDate}</li>
							{/if}
							{#if filterEndDate}
								<li>End Date: {filterEndDate}</li>
							{/if}
						</ul>
					</div>
				{/if}

				<div class="flex justify-end gap-3 mt-4">
					<button
						on:click={closeExportModal}
						class="px-4 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400 disabled:opacity-50"
						disabled={exporting}
					>
						Cancel
					</button>
					<button
						on:click={handleExport}
						class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:bg-gray-400"
						disabled={exporting}
					>
						{exporting ? 'Exporting...' : 'Export'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
