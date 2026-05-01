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
		viewed_by_referee?: boolean;  // Story 5.6: Whether referee has viewed this match
		updated_at?: string;           // Story 5.6: When assignment was last updated
	}

	interface GroupedMatches {
		[date: string]: Match[];
	}

	let matches: Match[] = [];
	let loading = true;
	let error = '';
	let hasProfile = true;
	let unavailableDays: Set<string> = new Set();
	let togglingDayAvailability = false;

	// Pagination
	let currentPage = 1;
	let perPage = 25;
	let totalMatches = 0;
	let totalPages = 0;

	// Filters
	let dateFrom = '';
	let dateTo = '';

	onMount(() => {
		loadMatches();
		loadUnavailableDays();
	});

	async function loadMatches() {
		loading = true;
		error = '';

		try {
			const params = new URLSearchParams();
			params.set('page', currentPage.toString());
			params.set('per_page', perPage.toString());
			if (dateFrom) params.set('date_from', dateFrom);
			if (dateTo) params.set('date_to', dateTo);

			const res = await fetch(`${API_URL}/api/referee/matches?${params.toString()}`, {
				credentials: 'include'
			});

			if (res.ok) {
				const data = await res.json();
				matches = data.matches || [];
				totalMatches = data.total;
				totalPages = data.total_pages;
				currentPage = data.page;
				// If matches is empty on first load with no filters, check if profile exists
				if (matches.length === 0 && totalMatches === 0 && !dateFrom && !dateTo) {
					const profileRes = await fetch(`${API_URL}/api/profile`, { credentials: 'include' });
					if (profileRes.ok) {
						const profile = await profileRes.json();
						hasProfile = !!(profile.first_name && profile.last_name && profile.date_of_birth);
					}
				}
			} else {
				error = 'Failed to load matches';
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
		loadMatches();
	}

	function clearFilters() {
		dateFrom = '';
		dateTo = '';
		currentPage = 1;
		loadMatches();
	}

	function goToPage(page: number) {
		if (page < 1 || page > totalPages) return;
		currentPage = page;
		loadMatches();
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

	async function loadUnavailableDays() {
		try {
			const res = await fetch(`${API_URL}/api/referee/day-unavailability`, {
				credentials: 'include'
			});

			if (res.ok) {
				const days = await res.json();
				unavailableDays = new Set(days.map((d: any) => d.unavailable_date));
			}
		} catch (e) {
			console.error('Failed to load unavailable days:', e);
		}
	}

	async function toggleDayAvailability(date: string) {
		const isCurrentlyUnavailable = unavailableDays.has(date);
		const newState = !isCurrentlyUnavailable;

		if (newState) {
			// Marking as unavailable - confirm action
			if (!confirm(`Mark ${formatDate(date)} as unavailable? This will clear any individual match availability for that day.`)) {
				return;
			}
		} else {
			// Unmarking - confirm action
			if (!confirm(`Clear unavailability for ${formatDate(date)}? You can then mark individual matches as available.`)) {
				return;
			}
		}

		togglingDayAvailability = true;

		try {
			const res = await fetch(`${API_URL}/api/referee/day-unavailability/${date}`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ unavailable: newState })
			});

			if (res.ok) {
				if (newState) {
					unavailableDays.add(date);
				} else {
					unavailableDays.delete(date);
				}
				unavailableDays = unavailableDays; // Trigger reactivity
				await loadMatches(); // Reload matches to reflect the change
			} else {
				alert('Failed to update day availability');
			}
		} catch (e) {
			console.error(e);
			alert('Network error');
		} finally {
			togglingDayAvailability = false;
		}
	}

	async function setAvailability(match: Match, available: boolean | null) {
		try {
			const res = await fetch(`${API_URL}/api/referee/matches/${match.id}/availability`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'include',
				body: JSON.stringify({ available })
			});

			if (res.ok) {
				// Update local state based on tri-state
				if (available === true) {
					match.is_available = true;
					match.is_unavailable = false;
				} else if (available === false) {
					match.is_available = false;
					match.is_unavailable = true;
				} else {
					// null = no preference
					match.is_available = false;
					match.is_unavailable = false;
				}
				matches = matches; // Trigger reactivity
			} else {
				alert('Failed to update availability');
			}
		} catch (e) {
			console.error(e);
			alert('Network error');
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
		// timeString is in format like "08:30:00" or "09:00"
		const parts = timeString.split(':');
		const hour = parseInt(parts[0]);
		const minute = parts[1];
		const ampm = hour >= 12 ? 'PM' : 'AM';
		const displayHour = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
		return `${displayHour}:${minute} ${ampm}`;
	}

	function extractMeetingTime(description: string | null): string | null {
		if (!description) return null;

		// Look for patterns like "Meet: 8:30 AM" or "Meeting time: 9:00AM"
		const meetingPattern = /(?:meet(?:ing)?[:\s]+)(\d{1,2}:\d{2}\s*(?:AM|PM))/i;
		const match = description.match(meetingPattern);
		return match ? match[1] : null;
	}

	function extractField(description: string | null): string | null {
		if (!description) return null;

		// Look for patterns like "Field: 3" or "Field 5"
		const fieldPattern = /field[:\s]+(\w+)/i;
		const match = description.match(fieldPattern);
		return match ? `Field ${match[1]}` : null;
	}

	// Group matches by date, separate assigned from available
	$: groupedMatches = matches.reduce((acc: GroupedMatches, match: Match) => {
		if (!match.is_assigned) {
			const date = match.match_date;
			if (!acc[date]) {
				acc[date] = [];
			}
			acc[date].push(match);
		}
		return acc;
	}, {});

	// Get all unique dates from grouped matches AND unavailable days
	// This ensures date headers show even when day is marked unavailable
	$: allDates = new Set([
		...Object.keys(groupedMatches),
		...Array.from(unavailableDays)
	]);
	$: sortedDates = Array.from(allDates).sort();
</script>

<svelte:head>
	<title>My Matches - Referee Scheduler</title>
</svelte:head>

<div class="container">
	<div class="header">
		<div class="header-left">
			<img src="/logo.svg" alt="Logo" class="header-logo" />
			<h1>My Matches</h1>
		</div>
		<a href="/dashboard" class="btn-secondary">← Back to Dashboard</a>
	</div>

	{#if !loading && !error && hasProfile}
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
					<strong>{totalMatches}</strong> match{totalMatches !== 1 ? 'es' : ''} found
					{#if totalPages > 1}
						<span class="page-info">(page {currentPage} of {totalPages})</span>
					{/if}
				</div>
				<div class="filter-actions">
					<div class="per-page-selector">
						<label for="perPage">Per page:</label>
						<select id="perPage" bind:value={perPage} on:change={() => { currentPage = 1; loadMatches(); }}>
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
		<p>Loading matches...</p>
	{:else if error}
		<div class="error">
			<p>{error}</p>
		</div>
	{:else if !hasProfile}
		<div class="info-box">
			<h2>Complete Your Profile</h2>
			<p>You need to complete your profile before you can view available matches.</p>
			<a href="/referee/profile" class="btn-primary">Go to Profile</a>
		</div>
	{:else}
		<!-- Available Matches Section -->
		<section class="available-section">
			<h2>Available Matches</h2>
			<p class="section-description">
				Mark your availability for upcoming matches you're eligible to referee
			</p>

			{#if sortedDates.length === 0}
				<div class="info-box">
					<p>No upcoming matches available at this time.</p>
					<p>Check back later for new matches.</p>
				</div>
			{:else}
				{#each sortedDates as date}
					<div class="date-group">
						<div class="date-header-row">
							<h3 class="date-header">{formatDate(date)}</h3>
							{#if unavailableDays.has(date)}
								<button
									class="btn-day-toggle btn-day-unavailable"
									on:click={() => toggleDayAvailability(date)}
									disabled={togglingDayAvailability}
								>
									Day Marked Unavailable - Click to Clear
								</button>
							{:else}
								<button
									class="btn-day-toggle"
									on:click={() => toggleDayAvailability(date)}
									disabled={togglingDayAvailability}
								>
									Mark Entire Day Unavailable
								</button>
							{/if}
						</div>

						{#if unavailableDays.has(date) && !groupedMatches[date]}
							<!-- Day is marked unavailable and has no matches (filtered out) -->
							<div class="day-unavailable-message">
								<p>You have marked this day as unavailable.</p>
								<p class="small-text">Individual matches for this day are hidden. Click the button above to make yourself available again.</p>
							</div>
						{:else if groupedMatches[date] && groupedMatches[date].length > 0}
							<!-- Show matches for this date -->
							<div class="matches-grid">
								{#each groupedMatches[date] as match}
								<div class="match-card" class:available={match.is_available} class:unavailable={match.is_unavailable}>
									<div class="match-header">
										<div class="match-title">
											<h4>{match.event_name}</h4>
											<span class="age-badge">{match.age_group}</span>
										</div>
										<div class="availability-buttons">
											<button
												class="btn-availability btn-available"
												class:active={match.is_available}
												on:click={() => setAvailability(match, true)}
												title="Mark as available"
											>
												✓
											</button>
											<button
												class="btn-availability btn-unavailable"
												class:active={match.is_unavailable}
												on:click={() => setAvailability(match, false)}
												title="Mark as unavailable"
											>
												✗
											</button>
											<button
												class="btn-availability btn-clear"
												class:active={!match.is_available && !match.is_unavailable}
												on:click={() => setAvailability(match, null)}
												title="Clear preference"
											>
												—
											</button>
										</div>
									</div>

									<div class="match-details">
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
										<div class="detail-row eligible-roles">
											<span class="icon">✓</span>
											<span class="small-text">
												Eligible for:
												{#if match.eligible_roles.includes('center')}
													Center Referee
												{/if}
												{#if match.eligible_roles.includes('center') && match.eligible_roles.includes('assistant')}
													,
												{/if}
												{#if match.eligible_roles.includes('assistant')}
													Assistant Referee
												{/if}
											</span>
										</div>
									</div>
								</div>
							{/each}
						</div>
					{/if}
				</div>
				{/each}
			{/if}
		</section>

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

	h1 {
		margin: 0;
		font-size: 2rem;
	}

	h2 {
		font-size: 1.5rem;
		margin-bottom: 0.5rem;
	}

	.section-description {
		color: #666;
		margin-bottom: 1.5rem;
	}

	.available-section {
		margin-bottom: 3rem;
	}

	.date-group {
		margin-bottom: 2rem;
	}

	.date-header-row {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		padding-bottom: 0.5rem;
		border-bottom: 2px solid var(--border-color);
		gap: 1rem;
		flex-wrap: wrap;
	}

	.date-header {
		font-size: 1.25rem;
		font-weight: 600;
		color: #2c3e50;
		margin: 0;
	}

	.btn-day-toggle {
		padding: 0.5rem 1rem;
		background-color: var(--bg-secondary);
		color: var(--text-primary);
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		font-size: 0.875rem;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		white-space: nowrap;
	}

	.btn-day-toggle:hover:not(:disabled) {
		background-color: var(--border-color);
		border-color: #9ca3af;
	}

	.btn-day-toggle:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.btn-day-unavailable {
		background-color: var(--error-light);
		color: #991b1b;
		border-color: var(--error-color);
		font-weight: 600;
	}

	.btn-day-unavailable:hover:not(:disabled) {
		background-color: #fee2e2;
		border-color: #dc2626;
	}

	.matches-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
		gap: 1rem;
	}

	.match-card {
		background: white;
		border: 2px solid var(--border-color);
		border-radius: 8px;
		padding: 1rem;
		transition: all 0.2s;
	}

	.match-card.available {
		border-color: #10b981;
		background-color: #f0fdf4;
	}

	.match-card.unavailable {
		border-color: var(--error-color);
		background-color: var(--error-light);
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

	.match-title h3,
	.match-title h4 {
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

	.availability-buttons {
		display: flex;
		gap: 0.25rem;
	}

	.btn-availability {
		background: white;
		border: 2px solid #cbd5e1;
		border-radius: 6px;
		padding: 0.5rem 0.75rem;
		font-size: 1.1rem;
		font-weight: 700;
		cursor: pointer;
		transition: all 0.2s;
		min-width: 40px;
		line-height: 1;
	}

	.btn-availability:hover {
		border-color: #94a3b8;
	}

	.btn-available.active {
		background: #10b981;
		color: white;
		border-color: #10b981;
	}

	.btn-available:hover:not(.active) {
		border-color: #10b981;
		color: #10b981;
	}

	.btn-unavailable.active {
		background: var(--error-color);
		color: white;
		border-color: var(--error-color);
	}

	.btn-unavailable:hover:not(.active) {
		border-color: var(--error-color);
		color: var(--error-color);
	}

	.btn-clear {
		color: var(--text-secondary); /* Ensure icon is visible when not active */
	}

	.btn-clear.active {
		background: var(--text-secondary);
		color: white;
		border-color: var(--text-secondary);
	}

	.btn-clear:hover:not(.active) {
		border-color: var(--text-secondary);
		color: var(--text-secondary);
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

	.eligible-roles {
		margin-top: 0.25rem;
		padding-top: 0.5rem;
		border-top: 1px solid var(--border-color);
	}

	.small-text {
		font-size: 0.85rem;
		color: var(--text-secondary);
	}

	.day-unavailable-message {
		background: var(--error-light);
		border: 2px solid var(--error-color);
		border-radius: 8px;
		padding: 1.5rem;
		text-align: center;
		margin-top: 1rem;
	}

	.day-unavailable-message p {
		margin: 0.5rem 0;
		color: #991b1b;
		font-weight: 500;
	}

	.day-unavailable-message .small-text {
		color: #b91c1c;
		font-weight: 400;
	}

	.info-box {
		background: var(--bg-secondary);
		border: 1px solid var(--border-color);
		border-radius: 8px;
		padding: 1.5rem;
		text-align: center;
	}

	.info-box h2 {
		margin-top: 0;
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
	}
</style>
