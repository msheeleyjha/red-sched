<script lang="ts">
	import { onMount } from 'svelte';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	interface HistoryMatch {
		match_id: number;
		event_name: string;
		team_name: string;
		age_group: string | null;
		match_date: string;
		start_time: string;
		end_time: string;
		location: string;
		status: string;
		archived: boolean;
		archived_at: string | null;
		role_type: string;
		acknowledged: boolean;
		acknowledged_at: string | null;
	}

	let loading = true;
	let matches: HistoryMatch[] = [];
	let filteredMatches: HistoryMatch[] = [];
	let error = '';

	// Pagination
	let currentPage = 1;
	let pageSize = 20;
	let totalPages = 1;
	let paginatedMatches: HistoryMatch[] = [];

	// Filters
	let statusFilter = 'all'; // all, active, archived
	let roleFilter = 'all'; // all, center, assistant
	let searchTerm = '';
	let startDate = '';
	let endDate = '';

	// Stats
	let totalMatches = 0;
	let archivedMatches = 0;
	let activeMatches = 0;

	onMount(async () => {
		await loadHistory();
	});

	async function loadHistory() {
		loading = true;
		error = '';

		try {
			const response = await fetch(`${API_URL}/api/referee/my-history`, {
				credentials: 'include'
			});

			if (response.ok) {
				matches = await response.json();
				calculateStats();
				applyFilters();
			} else if (response.status === 401) {
				error = 'Please log in to view your match history';
			} else {
				error = 'Failed to load match history';
			}
		} catch (err) {
			console.error('Error loading match history:', err);
			error = 'Failed to load match history';
		} finally {
			loading = false;
		}
	}

	function calculateStats() {
		totalMatches = matches.length;
		archivedMatches = matches.filter(m => m.archived).length;
		activeMatches = matches.filter(m => !m.archived).length;
	}

	function applyFilters() {
		let filtered = matches;

		// Filter by status (active/archived)
		if (statusFilter === 'active') {
			filtered = filtered.filter(m => !m.archived);
		} else if (statusFilter === 'archived') {
			filtered = filtered.filter(m => m.archived);
		}

		// Filter by role
		if (roleFilter === 'center') {
			filtered = filtered.filter(m => m.role_type === 'center');
		} else if (roleFilter === 'assistant') {
			filtered = filtered.filter(m => m.role_type === 'assistant_1' || m.role_type === 'assistant_2');
		}

		// Search filter
		if (searchTerm) {
			const term = searchTerm.toLowerCase();
			filtered = filtered.filter(m =>
				m.team_name.toLowerCase().includes(term) ||
				m.event_name.toLowerCase().includes(term) ||
				m.location.toLowerCase().includes(term)
			);
		}

		// Date range filter
		if (startDate) {
			filtered = filtered.filter(m => {
				const matchDate = m.match_date.split('T')[0];
				return matchDate >= startDate;
			});
		}

		if (endDate) {
			filtered = filtered.filter(m => {
				const matchDate = m.match_date.split('T')[0];
				return matchDate <= endDate;
			});
		}

		filteredMatches = filtered;
		currentPage = 1;
		updatePagination();
	}

	function updatePagination() {
		totalPages = Math.ceil(filteredMatches.length / pageSize);
		const startIndex = (currentPage - 1) * pageSize;
		const endIndex = startIndex + pageSize;
		paginatedMatches = filteredMatches.slice(startIndex, endIndex);
	}

	function goToPage(page: number) {
		if (page >= 1 && page <= totalPages) {
			currentPage = page;
			updatePagination();
		}
	}

	function clearFilters() {
		statusFilter = 'all';
		roleFilter = 'all';
		searchTerm = '';
		startDate = '';
		endDate = '';
		applyFilters();
	}

	function formatDate(dateString: string): string {
		const datePart = dateString.split('T')[0];
		const [year, month, day] = datePart.split('-').map(Number);
		const date = new Date(year, month - 1, day);
		return date.toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatTime(timeString: string): string {
		const parts = timeString.split(':');
		const hours = parseInt(parts[0]);
		const minutes = parts[1];
		const ampm = hours >= 12 ? 'PM' : 'AM';
		const displayHours = hours % 12 || 12;
		return `${displayHours}:${minutes} ${ampm}`;
	}

	function getRoleName(roleType: string): string {
		const roleMap: Record<string, string> = {
			center: 'Center',
			assistant_1: 'AR1',
			assistant_2: 'AR2'
		};
		return roleMap[roleType] || roleType;
	}

	function getRoleBadgeClass(roleType: string): string {
		if (roleType === 'center') {
			return 'bg-blue-100 text-blue-800';
		}
		return 'bg-green-100 text-green-800';
	}

	function getStatusBadgeClass(archived: boolean): string {
		return archived ? 'bg-gray-100 text-gray-800' : 'bg-green-100 text-green-800';
	}

	function getStatusText(archived: boolean): string {
		return archived ? 'Completed' : 'Upcoming';
	}

	// Reactive statements
	$: {
		statusFilter;
		roleFilter;
		searchTerm;
		startDate;
		endDate;
		applyFilters();
	}

	$: {
		currentPage;
		updatePagination();
	}
</script>

<svelte:head>
	<title>My Match History - Referee Scheduler</title>
</svelte:head>

<div class="container mx-auto px-4 py-6">
	<!-- Header -->
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-gray-900">My Match History</h1>
		<p class="text-gray-600 mt-2">View all matches you've worked as a referee</p>
	</div>

	<!-- Stats Cards -->
	{#if !loading && !error}
		<div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
			<div class="bg-white rounded-lg shadow p-6">
				<div class="text-sm font-medium text-gray-500">Total Matches</div>
				<div class="text-3xl font-bold text-gray-900 mt-2">{totalMatches}</div>
			</div>
			<div class="bg-white rounded-lg shadow p-6">
				<div class="text-sm font-medium text-gray-500">Upcoming</div>
				<div class="text-3xl font-bold text-green-600 mt-2">{activeMatches}</div>
			</div>
			<div class="bg-white rounded-lg shadow p-6">
				<div class="text-sm font-medium text-gray-500">Completed</div>
				<div class="text-3xl font-bold text-gray-600 mt-2">{archivedMatches}</div>
			</div>
		</div>
	{/if}

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
			<!-- Search -->
			<div>
				<label for="search" class="block text-sm font-medium text-gray-700 mb-1">
					Search
				</label>
				<input
					id="search"
					type="text"
					bind:value={searchTerm}
					placeholder="Team, event, location..."
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>

			<!-- Status Filter -->
			<div>
				<label for="status" class="block text-sm font-medium text-gray-700 mb-1">
					Status
				</label>
				<select
					id="status"
					bind:value={statusFilter}
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="all">All Matches</option>
					<option value="active">Upcoming</option>
					<option value="archived">Completed</option>
				</select>
			</div>

			<!-- Role Filter -->
			<div>
				<label for="role" class="block text-sm font-medium text-gray-700 mb-1">
					Role
				</label>
				<select
					id="role"
					bind:value={roleFilter}
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="all">All Roles</option>
					<option value="center">Center Referee</option>
					<option value="assistant">Assistant Referee</option>
				</select>
			</div>

			<!-- Start Date -->
			<div>
				<label for="startDate" class="block text-sm font-medium text-gray-700 mb-1">
					From Date
				</label>
				<input
					id="startDate"
					type="date"
					bind:value={startDate}
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>

			<!-- End Date -->
			<div>
				<label for="endDate" class="block text-sm font-medium text-gray-700 mb-1">
					To Date
				</label>
				<input
					id="endDate"
					type="date"
					bind:value={endDate}
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
			</div>
		</div>

		<!-- Clear Filters Button -->
		<div class="mt-4 flex justify-end">
			<button
				on:click={clearFilters}
				class="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
			>
				Clear Filters
			</button>
		</div>
	</div>

	<!-- Results Summary -->
	{#if !loading && !error}
		<div class="mb-4 text-sm text-gray-600">
			Showing {paginatedMatches.length} of {filteredMatches.length} matches
			{#if filteredMatches.length !== matches.length}
				(filtered from {matches.length} total)
			{/if}
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
			<p class="mt-4 text-gray-600">Loading your match history...</p>
		</div>
	{/if}

	<!-- Error State -->
	{#if error}
		<div class="bg-red-50 border border-red-200 rounded-lg p-4 text-red-800">
			{error}
		</div>
	{/if}

	<!-- Matches Table -->
	{#if !loading && !error && paginatedMatches.length > 0}
		<div class="bg-white rounded-lg shadow overflow-hidden">
			<div class="overflow-x-auto">
				<table class="min-w-full divide-y divide-gray-200">
					<thead class="bg-gray-50">
						<tr>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Date & Time
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Match
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Location
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Role
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Status
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Acknowledged
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each paginatedMatches as match}
							<tr class="hover:bg-gray-50 transition-colors">
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									<div class="font-medium">{formatDate(match.match_date)}</div>
									<div class="text-gray-500">{formatTime(match.start_time)}</div>
								</td>
								<td class="px-6 py-4 text-sm text-gray-900">
									<div class="font-medium">{match.event_name}</div>
									<div class="text-gray-500">{match.team_name}</div>
									{#if match.age_group}
										<div class="text-gray-400 text-xs mt-1">{match.age_group}</div>
									{/if}
								</td>
								<td class="px-6 py-4 text-sm text-gray-500">
									{match.location}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getRoleBadgeClass(match.role_type)}">
										{getRoleName(match.role_type)}
									</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium {getStatusBadgeClass(match.archived)}">
										{getStatusText(match.archived)}
									</span>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{#if match.acknowledged}
										<span class="text-green-600">✓ Yes</span>
									{:else if !match.archived}
										<span class="text-amber-600">Pending</span>
									{:else}
										<span class="text-gray-400">N/A</span>
									{/if}
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="mt-6 flex items-center justify-between">
				<div class="text-sm text-gray-700">
					Page {currentPage} of {totalPages}
				</div>
				<div class="flex gap-2">
					<button
						on:click={() => goToPage(1)}
						disabled={currentPage === 1}
						class="px-3 py-2 text-sm font-medium rounded-md border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						First
					</button>
					<button
						on:click={() => goToPage(currentPage - 1)}
						disabled={currentPage === 1}
						class="px-3 py-2 text-sm font-medium rounded-md border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Previous
					</button>

					<!-- Page numbers -->
					{#each Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
						const offset = Math.max(0, Math.min(currentPage - 3, totalPages - 5));
						return offset + i + 1;
					}) as pageNum}
						<button
							on:click={() => goToPage(pageNum)}
							class="px-3 py-2 text-sm font-medium rounded-md border border-gray-300 {currentPage === pageNum ? 'bg-blue-600 text-white border-blue-600' : 'bg-white hover:bg-gray-50'}"
						>
							{pageNum}
						</button>
					{/each}

					<button
						on:click={() => goToPage(currentPage + 1)}
						disabled={currentPage === totalPages}
						class="px-3 py-2 text-sm font-medium rounded-md border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Next
					</button>
					<button
						on:click={() => goToPage(totalPages)}
						disabled={currentPage === totalPages}
						class="px-3 py-2 text-sm font-medium rounded-md border border-gray-300 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
					>
						Last
					</button>
				</div>
			</div>
		{/if}
	{/if}

	<!-- Empty State -->
	{#if !loading && !error && paginatedMatches.length === 0}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<svg class="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
			</svg>
			<h3 class="mt-2 text-sm font-medium text-gray-900">No matches found</h3>
			<p class="mt-1 text-sm text-gray-500">
				{#if searchTerm || startDate || endDate || statusFilter !== 'all' || roleFilter !== 'all'}
					Try adjusting your filters to see more results.
				{:else}
					You haven't been assigned to any matches yet.
				{/if}
			</p>
		</div>
	{/if}
</div>
