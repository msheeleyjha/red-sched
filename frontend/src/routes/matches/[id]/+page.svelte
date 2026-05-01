<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	// Match data
	let match: any = null;
	let matchReport: any = null;
	let myAssignment: any = null;
	let currentUser: any = null;

	// Form state
	let finalScoreHome: number | null = null;
	let finalScoreAway: number | null = null;
	let redCards: number = 0;
	let yellowCards: number = 0;
	let injuries: string = '';
	let otherNotes: string = '';

	// UI state
	let loading = true;
	let submitting = false;
	let error = '';
	let successMessage = '';
	let showForm = false;
	let isEditing = false;

	// Authorization state
	let canSubmitReport = false;
	let isCenterReferee = false;
	let isAssignor = false;
	let authMessage = '';

	const matchId = $page.params.id;

	onMount(async () => {
		await loadData();
		// Mark match as viewed (Story 5.6)
		if (matchId && currentUser) {
			markMatchAsViewed();
		}
	});

	async function loadData() {
		loading = true;
		error = '';

		try {
			// Load current user
			const userRes = await fetch('/api/auth/me', {
				credentials: 'include'
			});
			if (!userRes.ok) throw new Error('Failed to load user data');
			currentUser = await userRes.json();

			// Load match details
			const matchRes = await fetch(`/api/matches/${matchId}`, {
				credentials: 'include'
			});
			if (!matchRes.ok) throw new Error('Failed to load match');
			match = await matchRes.json();

			// Load existing match report if any
			try {
				const reportRes = await fetch(`/api/matches/${matchId}/report`, {
					credentials: 'include'
				});
				if (reportRes.ok) {
					matchReport = await reportRes.json();
					// Pre-populate form with existing data
					finalScoreHome = matchReport.final_score_home;
					finalScoreAway = matchReport.final_score_away;
					redCards = matchReport.red_cards || 0;
					yellowCards = matchReport.yellow_cards || 0;
					injuries = matchReport.injuries || '';
					otherNotes = matchReport.other_notes || '';
				}
			} catch (err) {
				// Report doesn't exist yet, that's okay
			}

			// Check user's assignment for this match
			const assignmentsRes = await fetch(`/api/matches/${matchId}/roles`, {
				credentials: 'include'
			});
			if (assignmentsRes.ok) {
				const assignments = await assignmentsRes.json();
				myAssignment = assignments.find((a: any) => a.assigned_referee_id === currentUser.id);

				if (myAssignment) {
					isCenterReferee = myAssignment.role_type === 'center';
				}
			}

			// Check if user has assignor permissions
			// This is a simplified check - in production you'd check actual permissions
			isAssignor = currentUser.role === 'admin' || currentUser.role === 'assignor';

			// Determine authorization
			canSubmitReport = isCenterReferee || isAssignor;

			if (!canSubmitReport && myAssignment) {
				authMessage = 'Only the center referee can submit match reports';
			} else if (!canSubmitReport && !myAssignment) {
				authMessage = 'You are not assigned to this match';
			}

		} catch (err: any) {
			error = err.message || 'Failed to load match details';
			console.error('Error loading match:', err);
		} finally {
			loading = false;
		}
	}

	function handleShowForm() {
		showForm = true;
		isEditing = !!matchReport;
	}

	function handleCancelForm() {
		showForm = false;
		error = '';
		// Reset form if not editing
		if (!matchReport) {
			resetForm();
		}
	}

	function resetForm() {
		finalScoreHome = null;
		finalScoreAway = null;
		redCards = 0;
		yellowCards = 0;
		injuries = '';
		otherNotes = '';
	}

	async function handleSubmit() {
		error = '';
		successMessage = '';

		// Validate scores (both required)
		if (finalScoreHome === null || finalScoreAway === null) {
			error = 'Both home and away scores are required';
			return;
		}

		if (finalScoreHome < 0 || finalScoreAway < 0) {
			error = 'Scores must be non-negative';
			return;
		}

		if (redCards < 0 || yellowCards < 0) {
			error = 'Card counts must be non-negative';
			return;
		}

		submitting = true;

		try {
			const method = isEditing ? 'PUT' : 'POST';
			const url = `/api/matches/${matchId}/report`;

			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				credentials: 'include',
				body: JSON.stringify({
					final_score_home: finalScoreHome,
					final_score_away: finalScoreAway,
					red_cards: redCards,
					yellow_cards: yellowCards,
					injuries: injuries || null,
					other_notes: otherNotes || null
				})
			});

			if (!response.ok) {
				const errorText = await response.text();
				throw new Error(errorText || 'Failed to submit report');
			}

			matchReport = await response.json();
			successMessage = isEditing
				? 'Match report updated successfully! Match has been archived.'
				: 'Match report submitted successfully! Match has been archived.';
			showForm = false;

			// Reload match to see updated archived status
			setTimeout(() => {
				loadData();
			}, 1500);

		} catch (err: any) {
			error = err.message || 'Failed to submit report';
			console.error('Error submitting report:', err);
		} finally {
			submitting = false;
		}
	}

	function formatDate(dateStr: string): string {
		const date = new Date(dateStr);
		return date.toLocaleDateString('en-US', {
			weekday: 'long',
			year: 'numeric',
			month: 'long',
			day: 'numeric'
		});
	}

	function formatTime(timeStr: string): string {
		if (!timeStr) return '';
		const [hours, minutes] = timeStr.split(':');
		const hour = parseInt(hours);
		const ampm = hour >= 12 ? 'PM' : 'AM';
		const hour12 = hour % 12 || 12;
		return `${hour12}:${minutes} ${ampm}`;
	}

	// Story 5.6: Mark match as viewed when referee visits this page
	async function markMatchAsViewed() {
		try {
			await fetch(`/api/matches/${matchId}/viewed`, {
				method: 'POST',
				credentials: 'include'
			});
			// Don't show error if this fails - it's not critical
		} catch (err) {
			// Silently fail - viewing indicator is not critical functionality
			console.log('Failed to mark match as viewed:', err);
		}
	}
</script>

<svelte:head>
	<title>Match Details | Referee Scheduler</title>
</svelte:head>

<div class="container mx-auto px-4 py-8">
	<div class="mb-6">
		<button
			on:click={() => goto('/matches')}
			class="text-blue-600 hover:text-blue-800 flex items-center gap-2"
		>
			← Back to Matches
		</button>
	</div>

	{#if loading}
		<div class="text-center py-12">
			<p class="text-gray-600">Loading match details...</p>
		</div>
	{:else if error && !match}
		<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
			{error}
		</div>
	{:else if match}
		<!-- Match Details Card -->
		<div class="bg-white shadow-md rounded-lg p-6 mb-6">
			<div class="flex justify-between items-start mb-4">
				<h1 class="text-2xl font-bold text-gray-900">
					{match.team_name || match.event_name}
				</h1>
				{#if match.archived}
					<span class="bg-gray-500 text-white px-3 py-1 rounded-full text-sm font-medium">
						Archived
					</span>
				{:else}
					<span class="bg-green-500 text-white px-3 py-1 rounded-full text-sm font-medium">
						Active
					</span>
				{/if}
			</div>

			<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
				<div>
					<p class="text-sm text-gray-600">Date</p>
					<p class="text-lg font-medium">{formatDate(match.match_date)}</p>
				</div>

				<div>
					<p class="text-sm text-gray-600">Time</p>
					<p class="text-lg font-medium">
						{formatTime(match.start_time)}
						{#if match.end_time}
							- {formatTime(match.end_time)}
						{/if}
					</p>
				</div>

				<div>
					<p class="text-sm text-gray-600">Location</p>
					<p class="text-lg font-medium">{match.location || 'TBD'}</p>
				</div>

				<div>
					<p class="text-sm text-gray-600">Age Group</p>
					<p class="text-lg font-medium">{match.age_group || 'N/A'}</p>
				</div>

				{#if match.event_name}
					<div>
						<p class="text-sm text-gray-600">Event</p>
						<p class="text-lg font-medium">{match.event_name}</p>
					</div>
				{/if}
			</div>

			<!-- Assigned Referees -->
			{#if match.roles && match.roles.length > 0}
				<div class="mt-6">
					<h3 class="text-lg font-semibold mb-3">Assigned Referees</h3>
					<div class="space-y-2">
						{#each match.roles as role}
							<div class="flex items-center gap-2">
								<span class="inline-block px-2 py-1 rounded text-sm font-medium
									{role.role_type === 'center' ? 'bg-blue-100 text-blue-800' : 'bg-green-100 text-green-800'}">
									{role.role_type === 'center' ? 'Center Referee' : 'Assistant Referee'}
								</span>
								<span class="text-gray-700">
									{role.assigned_referee_name || 'Unassigned'}
								</span>
							</div>
						{/each}
					</div>
				</div>
			{/if}
		</div>

		<!-- Match Report Section -->
		<div class="bg-white shadow-md rounded-lg p-6">
			<h2 class="text-xl font-bold text-gray-900 mb-4">Match Report</h2>

			<!-- Success Message -->
			{#if successMessage}
				<div class="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded mb-4">
					{successMessage}
				</div>
			{/if}

			<!-- Error Message -->
			{#if error && match}
				<div class="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded mb-4">
					{error}
				</div>
			{/if}

			<!-- Authorization Message -->
			{#if !canSubmitReport && authMessage}
				<div class="bg-yellow-50 border border-yellow-200 text-yellow-800 px-4 py-3 rounded mb-4">
					{authMessage}
				</div>
			{/if}

			<!-- Existing Report Display -->
			{#if matchReport && !showForm}
				<div class="border-t border-gray-200 pt-4">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
						<div>
							<p class="text-sm text-gray-600">Final Score</p>
							<p class="text-2xl font-bold text-gray-900">
								{matchReport.final_score_home ?? '-'} - {matchReport.final_score_away ?? '-'}
							</p>
						</div>

						<div class="grid grid-cols-2 gap-4">
							<div>
								<p class="text-sm text-gray-600">Red Cards</p>
								<p class="text-lg font-medium text-red-600">{matchReport.red_cards || 0}</p>
							</div>
							<div>
								<p class="text-sm text-gray-600">Yellow Cards</p>
								<p class="text-lg font-medium text-yellow-600">{matchReport.yellow_cards || 0}</p>
							</div>
						</div>
					</div>

					{#if matchReport.injuries}
						<div class="mb-4">
							<p class="text-sm text-gray-600">Injuries</p>
							<p class="text-gray-900">{matchReport.injuries}</p>
						</div>
					{/if}

					{#if matchReport.other_notes}
						<div class="mb-4">
							<p class="text-sm text-gray-600">Other Notes</p>
							<p class="text-gray-900 whitespace-pre-wrap">{matchReport.other_notes}</p>
						</div>
					{/if}

					<div class="text-sm text-gray-500 mt-4">
						Submitted {new Date(matchReport.submitted_at).toLocaleString()}
						{#if matchReport.updated_at !== matchReport.submitted_at}
							<br />Last updated {new Date(matchReport.updated_at).toLocaleString()}
						{/if}
					</div>

					{#if canSubmitReport}
						<button
							on:click={handleShowForm}
							class="mt-4 bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 font-medium"
						>
							Edit Report
						</button>
					{/if}
				</div>

			<!-- Submit Report Button (no report exists) -->
			{:else if !showForm && canSubmitReport}
				<p class="text-gray-600 mb-4">No report has been submitted for this match yet.</p>
				<button
					on:click={handleShowForm}
					class="bg-blue-600 text-white px-6 py-2 rounded hover:bg-700 font-medium"
				>
					Submit Report
				</button>

			<!-- Report Submission Form -->
			{:else if showForm}
				<form on:submit|preventDefault={handleSubmit} class="space-y-6">
					<!-- Final Score -->
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="scoreHome" class="block text-sm font-medium text-gray-700 mb-1">
								Home Score <span class="text-red-500">*</span>
							</label>
							<input
								type="number"
								id="scoreHome"
								bind:value={finalScoreHome}
								min="0"
								required
								class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
							/>
						</div>

						<div>
							<label for="scoreAway" class="block text-sm font-medium text-gray-700 mb-1">
								Away Score <span class="text-red-500">*</span>
							</label>
							<input
								type="number"
								id="scoreAway"
								bind:value={finalScoreAway}
								min="0"
								required
								class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
							/>
						</div>
					</div>

					<!-- Cards -->
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="redCards" class="block text-sm font-medium text-gray-700 mb-1">
								Red Cards
							</label>
							<input
								type="number"
								id="redCards"
								bind:value={redCards}
								min="0"
								class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
							/>
						</div>

						<div>
							<label for="yellowCards" class="block text-sm font-medium text-gray-700 mb-1">
								Yellow Cards
							</label>
							<input
								type="number"
								id="yellowCards"
								bind:value={yellowCards}
								min="0"
								class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
							/>
						</div>
					</div>

					<!-- Injuries -->
					<div>
						<label for="injuries" class="block text-sm font-medium text-gray-700 mb-1">
							Injuries
						</label>
						<textarea
							id="injuries"
							bind:value={injuries}
							rows="3"
							placeholder="Describe any injuries that occurred during the match..."
							class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
						></textarea>
					</div>

					<!-- Other Notes -->
					<div>
						<label for="otherNotes" class="block text-sm font-medium text-gray-700 mb-1">
							Other Notes
						</label>
						<textarea
							id="otherNotes"
							bind:value={otherNotes}
							rows="4"
							placeholder="Any other information about the match..."
							class="w-full border border-gray-300 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
						></textarea>
					</div>

					<!-- Form Actions -->
					<div class="flex gap-3">
						<button
							type="submit"
							disabled={submitting}
							class="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700 font-medium disabled:opacity-50 disabled:cursor-not-allowed"
						>
							{submitting ? 'Submitting...' : isEditing ? 'Update Report' : 'Submit Report'}
						</button>
						<button
							type="button"
							on:click={handleCancelForm}
							disabled={submitting}
							class="bg-gray-300 text-gray-700 px-6 py-2 rounded hover:bg-gray-400 font-medium disabled:opacity-50"
						>
							Cancel
						</button>
					</div>

					<p class="text-sm text-gray-500">
						<span class="text-red-500">*</span> Required fields
					</p>

					{#if !isEditing}
						<p class="text-sm text-blue-600">
							Note: Submitting the final score will automatically archive this match.
						</p>
					{/if}
				</form>
			{/if}
		</div>
	{/if}
</div>
