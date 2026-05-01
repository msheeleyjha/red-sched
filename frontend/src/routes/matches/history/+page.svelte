<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	interface ArchivedMatch {
		id: number;
		event_name: string;
		team_name: string;
		age_group: string | null;
		match_date: string;
		start_time: string;
		end_time: string;
		location: string;
		archived_at: string | null;
		archived_by: number | null;
		roles: any[];
	}

	let loading = true;
	let matches: ArchivedMatch[] = [];
	let filteredMatches: ArchivedMatch[] = [];
	let error = '';

	// Pagination
	let currentPage = 1;
	let pageSize = 50;
	let totalPages = 1;
	let paginatedMatches: ArchivedMatch[] = [];

	// Filters
	let searchTerm = '';
	let startDate = '';
	let endDate = '';
	let ageGroupFilter = 'all';

	// Available age groups (extracted from data)
	let ageGroups: string[] = [];

	onMount(async () => {
		await loadArchivedMatches();
	});

	async function loadArchivedMatches() {
		loading = true;
		error = '';

		try {
			const response = await fetch(`${API_URL}/api/matches/archived`, {
				credentials: 'include'
			});

			if (response.ok) {
				matches = await response.json();
				extractAgeGroups();
				applyFilters();
			} else if (response.status === 401) {
				error = 'Please log in to view match history';
			} else {
				error = 'Failed to load archived matches';
			}
		} catch (err) {
			console.error('Error loading archived matches:', err);
			error = 'Failed to load archived matches';
		} finally {
			loading = false;
		}
	}

	function extractAgeGroups() {
		const groups = new Set<string>();
		matches.forEach(match => {
			if (match.age_group) {
				groups.add(match.age_group);
			}
		});
		ageGroups = Array.from(groups).sort();
	}

	function applyFilters() {
		let filtered = matches;

		// Search by team name
		if (searchTerm) {
			const term = searchTerm.toLowerCase();
			filtered = filtered.filter(m =>
				m.team_name.toLowerCase().includes(term) ||
				m.event_name.toLowerCase().includes(term) ||
				m.location.toLowerCase().includes(term)
			);
		}

		// Filter by date range
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

		// Filter by age group
		if (ageGroupFilter !== 'all') {
			filtered = filtered.filter(m => m.age_group === ageGroupFilter);
		}

		filteredMatches = filtered;
		currentPage = 1; // Reset to first page
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
		searchTerm = '';
		startDate = '';
		endDate = '';
		ageGroupFilter = 'all';
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

	function formatDateTime(dateString: string | null): string {
		if (!dateString) return 'N/A';
		const date = new Date(dateString);
		return date.toLocaleString('en-US', {
			month: 'short',
			day: 'numeric',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function getRefereeNames(roles: any[]): string {
		if (!roles || roles.length === 0) return 'No assignments';

		const assignedRoles = roles.filter(r => r.assigned_referee_name);
		if (assignedRoles.length === 0) return 'No assignments';

		return assignedRoles.map(r => r.assigned_referee_name).join(', ');
	}

	function viewMatchDetails(matchId: number) {
		// TODO: Navigate to match detail view when implemented
		console.log('View match details:', matchId);
	}

	// Reactive statements
	$: {
		searchTerm;
		startDate;
		endDate;
		ageGroupFilter;
		applyFilters();
	}

	$: {
		currentPage;
		updatePagination();
	}
</script>

<div class="container mx-auto px-4 py-6">
	<div class="mb-6">
		<h1 class="text-3xl font-bold text-gray-900">Match History</h1>
		<p class="text-gray-600 mt-2">View archived and completed matches</p>
	</div>

	<!-- Filters -->
	<div class="bg-white rounded-lg shadow p-6 mb-6">
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
			<!-- Search -->
			<div>
				<label for="search" class="block text-sm font-medium text-gray-700 mb-1">
					Search
				</label>
				<input
					id="search"
					type="text"
					bind:value={searchTerm}
					placeholder="Team name, event, location..."
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				/>
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

			<!-- Age Group -->
			<div>
				<label for="ageGroup" class="block text-sm font-medium text-gray-700 mb-1">
					Age Group
				</label>
				<select
					id="ageGroup"
					bind:value={ageGroupFilter}
					class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
				>
					<option value="all">All Age Groups</option>
					{#each ageGroups as ageGroup}
						<option value={ageGroup}>{ageGroup}</option>
					{/each}
				</select>
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
			Showing {paginatedMatches.length} of {filteredMatches.length} archived matches
			{#if filteredMatches.length !== matches.length}
				(filtered from {matches.length} total)
			{/if}
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="bg-white rounded-lg shadow p-12 text-center">
			<div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"></div>
			<p class="mt-4 text-gray-600">Loading archived matches...</p>
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
								Age Group
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Location
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Referees
							</th>
							<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
								Archived
							</th>
						</tr>
					</thead>
					<tbody class="bg-white divide-y divide-gray-200">
						{#each paginatedMatches as match}
							<tr
								class="hover:bg-gray-50 cursor-pointer transition-colors"
								on:click={() => viewMatchDetails(match.id)}
							>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									<div class="font-medium">{formatDate(match.match_date)}</div>
									<div class="text-gray-500">{formatTime(match.start_time)}</div>
								</td>
								<td class="px-6 py-4 text-sm text-gray-900">
									<div class="font-medium">{match.event_name}</div>
									<div class="text-gray-500">{match.team_name}</div>
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
									{match.age_group || 'N/A'}
								</td>
								<td class="px-6 py-4 text-sm text-gray-500">
									{match.location}
								</td>
								<td class="px-6 py-4 text-sm text-gray-500">
									{getRefereeNames(match.roles)}
								</td>
								<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
									{formatDateTime(match.archived_at)}
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
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
			</svg>
			<h3 class="mt-2 text-sm font-medium text-gray-900">No archived matches found</h3>
			<p class="mt-1 text-sm text-gray-500">
				{#if searchTerm || startDate || endDate || ageGroupFilter !== 'all'}
					Try adjusting your filters to see more results.
				{:else}
					There are no archived matches yet.
				{/if}
			</p>
		</div>
	{/if}
</div>

<style>
	/* Add any custom styles here */
</style>
