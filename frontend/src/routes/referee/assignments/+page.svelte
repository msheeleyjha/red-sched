<script lang="ts">
	import { onMount } from 'svelte';
	import type { PageData } from './$types';

	export let data: PageData;

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	interface ConflictingMatch {
		match_id: number;
		event_name: string;
		team_name: string;
		start_time: string;
		role_type: string;
	}

	interface Match {
		id: number;
		event_name: string;
		team_name: string;
		age_group: string;
		match_date: string;
		start_time: string;
		end_time: string;
		location: string;
		description: string | null;
		status: string;
		eligible_roles: string[];
		is_available: boolean;
		is_unavailable: boolean;
		is_assigned: boolean;
		assigned_role: string | null;
		acknowledged: boolean;
		acknowledged_at: string | null;
		has_conflict?: boolean;
		conflicting_matches?: ConflictingMatch[];
		viewed_by_referee?: boolean;
		updated_at?: string;
	}

	let matches: Match[] = [];
	let loading = true;
	let error = '';
	let acknowledging = false;

	// Pagination
	let currentPage = 1;
	let perPage = 25;
	let totalMatches = 0;
	let totalPages = 0;

	// Filters
	let dateFrom = '';
	let dateTo = '';

	onMount(() => {
		loadAssignments();
	});

	async function loadAssignments() {
		loading = true;
		error = '';

		try {
			const params = new URLSearchParams();
			params.set('page', currentPage.toString());
			params.set('per_page', perPage.toString());
			if (dateFrom) params.set('date_from', dateFrom);
			if (dateTo) params.set('date_to', dateTo);

			const res = await fetch(`${API_URL}/api/referee/assignments?${params.toString()}`, {
				credentials: 'include'
			});

			if (res.ok) {
				const data = await res.json();
				matches = data.matches || [];
				totalMatches = data.total;
				totalPages = data.total_pages;
				currentPage = data.page;
			} else {
				error = 'Failed to load assignments';
			}
		} catch (e) {
			error = 'Network error';
			console.error(e);
		} finally {
			loading = false;
		}
	}

	function applyFilters() {
		currentPage = 1;
		loadAssignments();
	}

	function clearFilters() {
		dateFrom = '';
		dateTo = '';
		currentPage = 1;
		loadAssignments();
	}

	function goToPage(page: number) {
		if (page < 1 || page > totalPages) return;
		currentPage = page;
		loadAssignments();
	}

	function getNextSaturday(fromDate: Date): Date {
		const d = new Date(fromDate);
		const day = d.getDay();
		const daysUntilSat = (6 - day + 7) % 7 || 7;
		d.setDate(d.getDate() + daysUntilSat);
		return d;
	}

	function formatDateParam(d: Date): string {
		const y = d.getFullYear();
		const m = String(d.getMonth() + 1).padStart(2, '0');
		const day = String(d.getDate()).padStart(2, '0');
		return `${y}-${m}-${day}`;
	}

	function setWeekend(which: 'this' | 'next') {
		const today = new Date();
		const day = today.getDay();
		let saturday: Date;

		if (which === 'this') {
			if (day === 6) {
				saturday = new Date(today);
			} else if (day === 0) {
				saturday = new Date(today);
				saturday.setDate(today.getDate() - 1);
			} else {
				saturday = getNextSaturday(today);
			}
		} else {
			if (day === 6) {
				saturday = new Date(today);
				saturday.setDate(today.getDate() + 7);
			} else if (day === 0) {
				saturday = new Date(today);
				saturday.setDate(today.getDate() + 6);
			} else {
				saturday = getNextSaturday(today);
				saturday.setDate(saturday.getDate() + 7);
			}
		}

		const sunday = new Date(saturday);
		sunday.setDate(saturday.getDate() + 1);

		dateFrom = formatDateParam(saturday);
		dateTo = formatDateParam(sunday);
	}

	async function acknowledgeAssignment(match: Match) {
		acknowledging = true;

		try {
			const res = await fetch(`${API_URL}/api/referee/matches/${match.id}/acknowledge`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include'
			});

			if (res.ok) {
				const data = await res.json();
				match.acknowledged = true;
				match.acknowledged_at = data.acknowledged_at;
				matches = matches;
			} else {
				alert('Failed to acknowledge assignment');
			}
		} catch (e) {
			console.error(e);
			alert('Network error');
		} finally {
			acknowledging = false;
		}
	}

	function formatDate(dateString: string): string {
		const [year, month, day] = dateString.split('-').map(Number);
		const date = new Date(year, month - 1, day);
		return date.toLocaleDateString('en-US', {
			weekday: 'long',
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		});
	}

	function formatShortDate(dateString: string): string {
		const [year, month, day] = dateString.split('-').map(Number);
		const date = new Date(year, month - 1, day);
		return date.toLocaleDateString('en-US', {
			weekday: 'short',
			month: 'short',
			day: 'numeric'
		});
	}

	function formatTime(timeString: string): string {
		const parts = timeString.split(':');
		const hour = parseInt(parts[0]);
		const minute = parts[1];
		const ampm = hour >= 12 ? 'PM' : 'AM';
		const displayHour = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
		return `${displayHour}:${minute} ${ampm}`;
	}

	function extractMeetingTime(description: string | null): string | null {
		if (!description) return null;
		const meetingPattern = /(?:meet(?:ing)?[:\s]+)(\d{1,2}:\d{2}\s*(?:AM|PM))/i;
		const match = description.match(meetingPattern);
		return match ? match[1] : null;
	}

	function extractField(description: string | null): string | null {
		if (!description) return null;
		const fieldPattern = /field[:\s]+(\w+)/i;
		const match = description.match(fieldPattern);
		return match ? `Field ${match[1]}` : null;
	}

	$: unacknowledgedCount = matches.filter(m => !m.acknowledged).length;
</script>

<svelte:head>
	<title>My Assignments - Referee Scheduler</title>
</svelte:head>

<div class="container">
	<div class="header">
		<div class="header-left">
			<img src="/logo.svg" alt="Logo" class="header-logo" />
			<h1>My Assignments</h1>
		</div>
		<a href="/dashboard" class="btn-secondary">← Back to Dashboard</a>
	</div>

	{#if !loading && !error}
		<div class="filters">
			<div class="filters-row">
				<div class="filter-group">
					<label for="dateFrom">From Date</label>
					<input
						type="date"
						id="dateFrom"
						bind:value={dateFrom}
					/>
				</div>
				<div class="filter-group">
					<label for="dateTo">To Date</label>
					<input
						type="date"
						id="dateTo"
						bind:value={dateTo}
					/>
				</div>
			</div>
			<div class="weekend-shortcuts">
				<span class="shortcut-label">Quick select:</span>
				<button class="btn-shortcut" on:click={() => { setWeekend('this'); applyFilters(); }}>This Weekend</button>
				<button class="btn-shortcut" on:click={() => { setWeekend('next'); applyFilters(); }}>Next Weekend</button>
			</div>
			<div class="filters-footer">
				<div class="stats">
					<strong>{totalMatches}</strong> assignment{totalMatches !== 1 ? 's' : ''}
					{#if unacknowledgedCount > 0}
						<span class="pending-count">({unacknowledgedCount} pending acknowledgment)</span>
					{/if}
					{#if totalPages > 1}
						<span class="page-info">(page {currentPage} of {totalPages})</span>
					{/if}
				</div>
				<div class="filter-actions">
					<div class="per-page-selector">
						<label for="perPage">Per page:</label>
						<select id="perPage" bind:value={perPage} on:change={() => { currentPage = 1; loadAssignments(); }}>
							<option value={25}>25</option>
							<option value={50}>50</option>
							<option value={100}>100</option>
						</select>
					</div>
					<button class="btn-small btn-primary" on:click={applyFilters}>
						Apply Filters
					</button>
					<button class="btn-small btn-secondary" on:click={clearFilters}>
						Clear Filters
					</button>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<p class="loading-text">Loading assignments...</p>
	{:else if error}
		<div class="error">
			<p>{error}</p>
		</div>
	{:else if matches.length === 0}
		<div class="info-box">
			{#if dateFrom || dateTo}
				<p>No assignments found matching your date filters.</p>
				<p>Try adjusting your filters or click Clear Filters.</p>
			{:else}
				<p>You don't have any upcoming assignments.</p>
				<p>Check back later or contact your assignor.</p>
			{/if}
		</div>
	{:else}
		<div class="matches-grid">
			{#each matches as match}
				<div class="match-card" class:has-conflict={match.has_conflict}>
					{#if match.has_conflict && match.conflicting_matches && match.conflicting_matches.length > 0}
						<div class="scheduling-conflict-warning">
							<span class="warning-icon">⚠️</span>
							<div class="warning-text">
								<strong>Scheduling Conflict Detected</strong>
								<p>This assignment overlaps with {match.conflicting_matches.length} other assignment{match.conflicting_matches.length > 1 ? 's' : ''}:</p>
								<ul class="conflict-list">
									{#each match.conflicting_matches as conflict}
										<li>
											<strong>{formatTime(conflict.start_time)}</strong> - {conflict.event_name} ({conflict.team_name})
											{#if conflict.role_type === 'center'}
												as Center Referee
											{:else if conflict.role_type === 'assistant_1'}
												as AR1
											{:else if conflict.role_type === 'assistant_2'}
												as AR2
											{/if}
										</li>
									{/each}
								</ul>
								<p class="conflict-advice">Please contact the assignor immediately to resolve this conflict.</p>
							</div>
						</div>
					{/if}

					<div class="match-header">
						<div class="match-title">
							<h3>{match.event_name}</h3>
							<span class="age-badge">{match.age_group}</span>
							<span class="role-badge">
								{#if match.assigned_role === 'center'}
									Center Referee
								{:else if match.assigned_role === 'assistant_1'}
									Assistant Referee 1
								{:else if match.assigned_role === 'assistant_2'}
									Assistant Referee 2
								{/if}
							</span>
							{#if match.status === 'cancelled'}
								<span class="status-badge cancelled">Cancelled</span>
							{/if}
							{#if match.viewed_by_referee === false}
								<span class="update-badge" title="Match details have been updated since you last viewed it">
									📢 Updated
								</span>
							{/if}
						</div>
					</div>

					<div class="match-details">
						<div class="detail-row">
							<span class="icon">📅</span>
							<span>{formatShortDate(match.match_date)}</span>
						</div>
						<div class="detail-row">
							<span class="icon">🕐</span>
							<span>{formatTime(match.start_time)}</span>
							{#if extractMeetingTime(match.description)}
								<span class="meeting-time">
									(Meet: {extractMeetingTime(match.description)})
								</span>
							{/if}
						</div>
						<div class="detail-row">
							<span class="icon">📍</span>
							<span>{match.location}</span>
							{#if extractField(match.description)}
								<span class="field">• {extractField(match.description)}</span>
							{/if}
						</div>
						<div class="detail-row">
							<span class="icon">⚽</span>
							<span class="team-name">{match.team_name}</span>
						</div>
					</div>

					<div class="acknowledgment-section">
						{#if match.acknowledged}
							<div class="acknowledged-indicator">
								<span class="check-icon">✓</span>
								<span>Confirmed</span>
							</div>
						{:else}
							<button
								class="btn-acknowledge"
								on:click={() => acknowledgeAssignment(match)}
								disabled={acknowledging}
							>
								{acknowledging ? 'Acknowledging...' : 'Acknowledge Assignment'}
							</button>
						{/if}
					</div>
				</div>
			{/each}
		</div>

		{#if totalPages > 1}
			<div class="pagination">
				<button
					class="btn-small btn-secondary"
					on:click={() => goToPage(currentPage - 1)}
					disabled={currentPage <= 1}
				>
					Previous
				</button>

				<div class="page-numbers">
					{#each Array.from({ length: totalPages }, (_, i) => i + 1) as page}
						{#if page === 1 || page === totalPages || (page >= currentPage - 2 && page <= currentPage + 2)}
							<button
								class="page-btn"
								class:active={page === currentPage}
								on:click={() => goToPage(page)}
							>
								{page}
							</button>
						{:else if page === currentPage - 3 || page === currentPage + 3}
							<span class="page-ellipsis">...</span>
						{/if}
					{/each}
				</div>

				<button
					class="btn-small btn-secondary"
					on:click={() => goToPage(currentPage + 1)}
					disabled={currentPage >= totalPages}
				>
					Next
				</button>
			</div>
		{/if}
	{/if}
</div>

<style>
	.container {
		max-width: 1200px;
		margin: 0 auto;
		padding: 2rem 1rem;
	}

	.header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 2rem;
		flex-wrap: wrap;
		gap: 1rem;
	}

	.header-left {
		display: flex;
		align-items: center;
		gap: 1rem;
	}

	.header-logo {
		height: 40px;
		width: auto;
	}

	h1 {
		margin: 0;
		font-size: 2rem;
	}

	.filters {
		background: white;
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1.25rem;
		margin-bottom: 2rem;
	}

	.filters-row {
		display: flex;
		gap: 1.5rem;
		align-items: flex-end;
		flex-wrap: wrap;
	}

	.weekend-shortcuts {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.75rem;
	}

	.shortcut-label {
		font-size: 0.875rem;
		color: var(--text-secondary);
		font-weight: 500;
	}

	.btn-shortcut {
		padding: 0.375rem 0.75rem;
		font-size: 0.8rem;
		border: 1px solid var(--primary-color);
		border-radius: 0.375rem;
		background: white;
		color: var(--primary-color);
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;
	}

	.btn-shortcut:hover {
		background: var(--primary-color);
		color: white;
	}

	.filters-footer {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-top: 1rem;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border-color);
	}

	.filter-actions {
		display: flex;
		gap: 0.5rem;
		align-items: center;
	}

	.per-page-selector {
		display: flex;
		align-items: center;
		gap: 0.375rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.per-page-selector label {
		font-weight: 500;
		white-space: nowrap;
	}

	.per-page-selector select {
		padding: 0.25rem 0.375rem;
		border: 1px solid var(--border-color);
		border-radius: 0.25rem;
		font-size: 0.875rem;
	}

	.filter-group {
		flex: 1;
		min-width: 200px;
	}

	.filter-group label {
		display: block;
		margin-bottom: 0.5rem;
		font-weight: 500;
		color: var(--text-primary);
	}

	.filter-group input[type='date'] {
		width: 100%;
		padding: 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		font-size: 1rem;
		font-family: inherit;
	}

	.stats {
		color: var(--text-secondary);
		padding: 0.75rem 0;
	}

	.pending-count {
		color: var(--warning-color);
		font-weight: 500;
	}

	.page-info {
		color: var(--text-secondary);
	}

	.btn-small {
		padding: 0.375rem 0.75rem;
		font-size: 0.875rem;
		border: none;
		border-radius: 0.25rem;
		cursor: pointer;
		font-weight: 500;
		transition: all 0.2s;
	}

	.btn-small:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.matches-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1rem;
	}

	.match-card {
		background: white;
		border: 2px solid var(--primary-color);
		background-color: var(--primary-light);
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.2s;
	}

	.match-card.has-conflict {
		border: 3px solid var(--error-color);
	}

	.match-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		margin-bottom: 1rem;
		gap: 1rem;
	}

	.match-title {
		flex: 1;
		min-width: 0;
	}

	.match-title h3 {
		margin: 0 0 0.5rem 0;
		font-size: 1.1rem;
		overflow-wrap: break-word;
	}

	.age-badge {
		display: inline-block;
		background: var(--primary-color);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.85rem;
		font-weight: 600;
		margin-right: 0.5rem;
	}

	.role-badge {
		display: inline-block;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.85rem;
		font-weight: 600;
		background: var(--primary-color);
		color: white;
	}

	.status-badge.cancelled {
		display: inline-block;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.85rem;
		font-weight: 600;
		background: var(--error-color);
		color: white;
		margin-left: 0.5rem;
	}

	.update-badge {
		display: inline-block;
		background: var(--warning-color);
		color: white;
		padding: 0.25rem 0.5rem;
		border-radius: 4px;
		font-size: 0.75rem;
		font-weight: 700;
		margin-left: 0.5rem;
		animation: pulse 2s infinite;
	}

	@keyframes pulse {
		0%, 100% { opacity: 1; }
		50% { opacity: 0.7; }
	}

	.scheduling-conflict-warning {
		background: #fee2e2;
		border-left: 4px solid var(--error-color);
		padding: 1rem;
		margin-bottom: 1rem;
		display: flex;
		gap: 0.75rem;
		align-items: flex-start;
		border-radius: 0.375rem;
	}

	.warning-icon {
		font-size: 1.5rem;
		flex-shrink: 0;
	}

	.warning-text {
		flex: 1;
	}

	.warning-text strong {
		display: block;
		color: #991b1b;
		font-size: 0.95rem;
		margin-bottom: 0.25rem;
	}

	.warning-text p {
		color: #7f1d1d;
		font-size: 0.875rem;
		margin: 0;
		line-height: 1.4;
	}

	.conflict-list {
		margin: 0.75rem 0;
		padding-left: 1.25rem;
		color: #7f1d1d;
		font-size: 0.875rem;
		line-height: 1.6;
	}

	.conflict-list li {
		margin-bottom: 0.5rem;
	}

	.conflict-list strong {
		color: #991b1b;
		display: inline;
	}

	.conflict-advice {
		margin-top: 0.75rem !important;
		font-weight: 500;
		color: #991b1b !important;
	}

	.match-details {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.detail-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.95rem;
		flex-wrap: wrap;
	}

	.detail-row .icon {
		font-size: 1rem;
		min-width: 1.5rem;
	}

	.meeting-time {
		color: #059669;
		font-weight: 500;
		font-size: 0.9rem;
	}

	.field {
		color: #666;
		font-size: 0.9rem;
	}

	.team-name {
		color: var(--text-primary);
		font-weight: 500;
	}

	.acknowledgment-section {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--border-color);
	}

	.btn-acknowledge {
		width: 100%;
		padding: 0.75rem 1rem;
		background: var(--primary-color);
		color: white;
		border: none;
		border-radius: 6px;
		font-weight: 600;
		cursor: pointer;
		transition: all 0.2s;
	}

	.btn-acknowledge:hover:not(:disabled) {
		background: var(--primary-hover);
	}

	.btn-acknowledge:disabled {
		opacity: 0.6;
		cursor: not-allowed;
	}

	.acknowledged-indicator {
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 0.5rem;
		padding: 0.75rem;
		background: #d1fae5;
		color: #065f46;
		border-radius: 6px;
		font-weight: 600;
	}

	.check-icon {
		font-size: 1.25rem;
		font-weight: bold;
	}

	.loading-text {
		color: var(--text-secondary);
		text-align: center;
		padding: 2rem;
	}

	.info-box {
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1.5rem;
		text-align: center;
	}

	.error {
		background: var(--error-light);
		border: 1px solid #fecaca;
		border-radius: 8px;
		padding: 1rem;
		color: #991b1b;
	}

	.btn-secondary {
		display: inline-block;
		padding: 0.5rem 1rem;
		border-radius: 6px;
		text-decoration: none;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		background: white;
		color: var(--primary-color);
		border: 2px solid var(--primary-color);
	}

	.btn-secondary:hover {
		background: var(--primary-light);
	}

	.pagination {
		display: flex;
		justify-content: center;
		align-items: center;
		gap: 0.5rem;
		margin-top: 1.5rem;
		padding: 1rem 0;
	}

	.page-numbers {
		display: flex;
		gap: 0.25rem;
		align-items: center;
	}

	.page-btn {
		min-width: 2.25rem;
		height: 2.25rem;
		padding: 0 0.5rem;
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		background: white;
		cursor: pointer;
		font-size: 0.875rem;
		font-weight: 500;
		color: var(--text-primary);
		transition: all 0.2s;
	}

	.page-btn:hover {
		background-color: var(--bg-secondary);
		border-color: var(--primary-color);
	}

	.page-btn.active {
		background-color: var(--primary-color);
		color: white;
		border-color: var(--primary-color);
	}

	.page-ellipsis {
		padding: 0 0.25rem;
		color: var(--text-secondary);
	}

	@media (max-width: 768px) {
		.container {
			padding: 1rem 0.5rem;
		}

		h1 {
			font-size: 1.5rem;
		}

		.matches-grid {
			grid-template-columns: 1fr;
		}

		.header {
			flex-direction: column;
			align-items: flex-start;
		}

		.filters-row {
			flex-direction: column;
		}
	}
</style>
