<script lang="ts">
	import { goto } from '$app/navigation';

	const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

	let step: 'upload' | 'preview' | 'complete' = 'upload';
	let file: File | null = null;
	let uploading = false;
	let importing = false;
	let error = '';

	// Preview data
	let rows: any[] = [];
	let duplicates: any[] = [];
	let validRows: any[] = [];
	let errorRows: any[] = [];
	let uniqueLocations: string[] = [];

	// Duplicate resolution: track which row the user picks per duplicate group
	// Key: group index, Value: set of selected row numbers
	let duplicateSelections: Record<number, Set<number>> = {};

	// Filter options (Story 6.4)
	let filterPractices = false;
	let filterAway = false;
	let homeLocations: string[] = [];
	let customExcludeText = '';

	// Section collapse state (all start collapsed)
	let showDuplicates = false;
	let showFilters = false;
	let showErrors = false;
	let showValidMatches = false;

	// Import results
	let importResult: any = null;

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files[0]) {
			file = target.files[0];

			// Validate file extension
			if (!file.name.toLowerCase().endsWith('.csv')) {
				error = 'Only .csv files are accepted';
				file = null;
				return;
			}

			error = '';
		}
	}

	async function handleUpload() {
		if (!file) {
			error = 'Please select a file';
			return;
		}

		uploading = true;
		error = '';

		try {
			const formData = new FormData();
			formData.append('file', file);

			const response = await fetch(`${API_URL}/api/matches/import/parse`, {
				method: 'POST',
				credentials: 'include',
				body: formData
			});

			if (response.ok) {
				const data = await response.json();
				rows = data.rows || [];
				duplicates = data.duplicates || [];
				uniqueLocations = data.unique_locations || [];

				// Collect row numbers involved in duplicate groups
				const duplicateRowNumbers = new Set<number>();
				duplicates.forEach((group: any) => {
					(group.matches || []).forEach((m: any) => duplicateRowNumbers.add(m.row_number));
				});

				// Separate valid, error, and duplicate rows
				validRows = rows.filter((r) => !r.error && !duplicateRowNumbers.has(r.row_number));
				errorRows = rows.filter((r) => r.error);

				// Initialize duplicate selections: first row in each group selected by default
				duplicateSelections = {};
				duplicates.forEach((group: any, i: number) => {
					const matches = group.matches || [];
					if (matches.length > 0) {
						duplicateSelections[i] = new Set([matches[0].row_number]);
					}
				});

				step = 'preview';
			} else {
				const text = await response.text();
				error = text || 'Failed to parse CSV';
			}
		} catch (err) {
			error = 'Failed to upload file';
		} finally {
			uploading = false;
		}
	}

	async function handleConfirmImport() {
		importing = true;
		error = '';

		try {
			const response = await fetch(`${API_URL}/api/matches/import/confirm`, {
				method: 'POST',
				credentials: 'include',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					rows: rowsToImport,
					resolutions: {},
					filters: {
						filter_practices: filterPractices,
						filter_away: filterAway,
						home_locations: homeLocations,
						custom_exclude_terms: customExcludeTerms
					}
				})
			});

			if (response.ok) {
				importResult = await response.json();
				step = 'complete';
			} else {
				const text = await response.text();
				error = text || 'Failed to import matches';
			}
		} catch (err) {
			error = 'Failed to import matches';
		} finally {
			importing = false;
		}
	}

	function toggleDuplicateSelection(groupIndex: number, rowNumber: number) {
		const current = duplicateSelections[groupIndex] || new Set();
		if (current.has(rowNumber)) {
			current.delete(rowNumber);
		} else {
			current.add(rowNumber);
		}
		duplicateSelections[groupIndex] = new Set(current);
		duplicateSelections = { ...duplicateSelections };
	}

	// Rows selected from duplicate groups to include in import
	$: resolvedDuplicateRows = (() => {
		const resolved: any[] = [];
		duplicates.forEach((group: any, i: number) => {
			const selected = duplicateSelections[i] || new Set();
			(group.matches || []).forEach((m: any) => {
				if (selected.has(m.row_number)) {
					resolved.push(m);
				}
			});
		});
		return resolved;
	})();

	// All rows that will be sent to the confirm endpoint
	$: rowsToImport = [...validRows, ...resolvedDuplicateRows];

	function handleStartOver() {
		step = 'upload';
		file = null;
		rows = [];
		duplicates = [];
		validRows = [];
		errorRows = [];
		uniqueLocations = [];
		duplicateSelections = {};
		showDuplicates = false;
		showFilters = false;
		showErrors = false;
		showValidMatches = false;
		filterPractices = false;
		filterAway = false;
		homeLocations = [];
		customExcludeText = '';
		importResult = null;
		error = '';
	}

	// Helper functions for location selection
	function toggleLocation(location: string) {
		if (homeLocations.includes(location)) {
			homeLocations = homeLocations.filter((l) => l !== location);
		} else {
			homeLocations = [...homeLocations, location];
		}
	}

	function selectAllLocations() {
		homeLocations = [...uniqueLocations];
	}

	function clearAllLocations() {
		homeLocations = [];
	}

	// Parse comma-separated custom exclude terms
	$: customExcludeTerms = customExcludeText
		.split(',')
		.map((t) => t.trim())
		.filter((t) => t.length > 0);

	// Computed values for preview
	$: filteredCount = (() => {
		if (!filterPractices && !filterAway && customExcludeTerms.length === 0) return 0;

		let count = 0;
		rowsToImport.forEach((row) => {
			const eventLower = (row.event_name || '').toLowerCase();
			if (filterPractices && (eventLower.includes('practice') || eventLower.includes('training'))) {
				count++;
				return;
			}
			if (customExcludeTerms.length > 0) {
				const matched = customExcludeTerms.some((term) => eventLower.includes(term.toLowerCase()));
				if (matched) {
					count++;
					return;
				}
			}
			if (filterAway && homeLocations.length > 0) {
				const isHome = homeLocations.some((loc) => row.location?.includes(loc));
				if (!isHome) count++;
			}
		});
		return count;
	})();

	$: willImportCount = rowsToImport.length - filteredCount;

	// Rows that will actually be imported (respects all filters)
	$: displayRows = rowsToImport.filter((row) => {
		const eventLower = (row.event_name || '').toLowerCase();
		if (filterPractices && (eventLower.includes('practice') || eventLower.includes('training'))) {
			return false;
		}
		if (customExcludeTerms.length > 0) {
			if (customExcludeTerms.some((term) => eventLower.includes(term.toLowerCase()))) {
				return false;
			}
		}
		if (filterAway && homeLocations.length > 0) {
			const isHome = homeLocations.some((loc) => row.location?.includes(loc));
			if (!isHome) return false;
		}
		return true;
	});
</script>

<svelte:head>
	<title>Import Match Schedule - Referee Scheduler</title>
</svelte:head>

<div class="container">
	<div class="header">
		<div class="header-left">
			<img src="/logo.svg" alt="Logo" class="header-logo" />
			<h1>Import Match Schedule</h1>
		</div>
		<button on:click={() => goto('/dashboard')} class="btn btn-secondary">Back to Dashboard</button>
	</div>

	{#if error}
		<div class="alert alert-error">{error}</div>
	{/if}

	{#if step === 'upload'}
		<div class="card">
			<h2>Upload CSV File</h2>
			<p class="instructions">
				Upload a CSV export from Stack Team App. The file must include columns: event_name,
				team_name, start_date, start_time, end_time, location.
			</p>

			<div class="upload-section">
				<input
					type="file"
					accept=".csv"
					on:change={handleFileSelect}
					id="csvFile"
					class="file-input"
				/>
				<label for="csvFile" class="file-label">
					{file ? file.name : 'Choose CSV file...'}
				</label>
			</div>

			{#if file}
				<div class="actions">
					<button on:click={handleUpload} class="btn btn-primary" disabled={uploading}>
						{uploading ? 'Parsing...' : 'Parse CSV'}
					</button>
				</div>
			{/if}
		</div>
	{:else if step === 'preview'}
		<div class="card">
			<h2>Import Preview</h2>
			<p class="summary">
				<strong>{validRows.length}</strong> unique matches •
				<strong class="error-count">{errorRows.length}</strong> rows with errors
				{#if duplicates.length > 0}
					• <strong class="duplicate-count">{duplicates.length}</strong> duplicate group{duplicates.length > 1 ? 's' : ''} ({resolvedDuplicateRows.length} selected to keep)
				{/if}
			</p>

			{#if duplicates.length > 0}
				<div class="section duplicate-section">
					<button class="section-toggle" on:click={() => (showDuplicates = !showDuplicates)}>
						<span class="toggle-icon">{showDuplicates ? '▾' : '▸'}</span>
						<h3>Duplicates Detected ({duplicates.length} group{duplicates.length > 1 ? 's' : ''})</h3>
					</button>
					{#if showDuplicates}
					<p class="section-info">
						The following rows appear to be duplicates. Select which row(s) to keep from each group.
						Unselected rows will be skipped.
					</p>

					{#each duplicates as group, groupIndex}
						<div class="duplicate-group">
							<div class="duplicate-group-header">
								{#if group.signal === 'reference_id'}
									<strong>Same Reference ID: {group.matches[0]?.reference_id}</strong>
								{:else if group.signal === 'same_event'}
									<strong>Same event, age group, location, date and time</strong>
								{:else}
									<strong>Same match (team, date, time) with different reference IDs</strong>
								{/if}
							</div>
							<div class="table-container">
								<table class="preview-table">
									<thead>
										<tr>
											<th>Keep</th>
											<th>Row</th>
											<th>Event</th>
											<th>Team</th>
											<th>Date</th>
											<th>Time</th>
											<th>Location</th>
											<th>Ref ID</th>
										</tr>
									</thead>
									<tbody>
										{#each group.matches as match}
											<tr class:duplicate-selected={duplicateSelections[groupIndex]?.has(match.row_number)}>
												<td>
													<input
														type="checkbox"
														checked={duplicateSelections[groupIndex]?.has(match.row_number)}
														on:change={() => toggleDuplicateSelection(groupIndex, match.row_number)}
													/>
												</td>
												<td>{match.row_number}</td>
												<td>{match.event_name}</td>
												<td>{match.team_name}</td>
												<td>{match.start_date}</td>
												<td>{match.start_time} - {match.end_time}</td>
												<td>{match.location}</td>
												<td>{match.reference_id || '—'}</td>
											</tr>
										{/each}
									</tbody>
								</table>
							</div>
						</div>
					{/each}
					{/if}
				</div>
			{/if}

			<!-- Filter Options (Story 6.4) -->
			<div class="section filter-section">
				<button class="section-toggle" on:click={() => (showFilters = !showFilters)}>
					<span class="toggle-icon">{showFilters ? '▾' : '▸'}</span>
					<h3>Filter Options{filteredCount > 0 ? ` (${filteredCount} filtered)` : ''}</h3>
				</button>
				{#if showFilters}
				<p class="section-info">
					Optionally filter out practice matches and away matches before importing.
				</p>

				<div class="filter-options">
					<!-- Practice Filter -->
					<label class="filter-checkbox">
						<input type="checkbox" bind:checked={filterPractices} />
						<span class="filter-label">
							<strong>Filter Practices & Training</strong>
							<span class="filter-description"
								>Skip matches with "Practice" or "Training" in the event name</span
							>
						</span>
					</label>

					<!-- Custom Exclude Terms -->
					<div class="custom-exclude-panel">
						<label for="customExclude">
							<strong>Custom Exclude Terms</strong>
							<span class="filter-description"
								>Comma-separated terms to match against event name (e.g. "Mini, Scrimmage")</span
							>
						</label>
						<input
							type="text"
							id="customExclude"
							bind:value={customExcludeText}
							placeholder="e.g. Mini, Scrimmage, Friendly"
							class="text-input"
						/>
						{#if customExcludeTerms.length > 0}
							<div class="exclude-tags">
								{#each customExcludeTerms as term}
									<span class="exclude-tag">{term}</span>
								{/each}
							</div>
						{/if}
					</div>

					<!-- Away Match Filter -->
					<label class="filter-checkbox">
						<input type="checkbox" bind:checked={filterAway} />
						<span class="filter-label">
							<strong>Filter Away Matches</strong>
							<span class="filter-description">Skip matches not at home locations</span>
						</span>
					</label>

					<!-- Home Locations Selection (shown when filterAway is checked) -->
					{#if filterAway && uniqueLocations.length > 0}
						<div class="home-locations-panel">
							<div class="locations-header">
								<h4>Select Home Locations ({homeLocations.length} of {uniqueLocations.length} selected)</h4>
								<div class="locations-actions">
									<button
										type="button"
										on:click={selectAllLocations}
										class="btn-link"
										disabled={homeLocations.length === uniqueLocations.length}
									>
										Select All
									</button>
									<button
										type="button"
										on:click={clearAllLocations}
										class="btn-link"
										disabled={homeLocations.length === 0}
									>
										Clear All
									</button>
								</div>
							</div>

							<div class="locations-grid">
								{#each uniqueLocations as location}
									<label class="location-checkbox">
										<input
											type="checkbox"
											checked={homeLocations.includes(location)}
											on:change={() => toggleLocation(location)}
										/>
										<span class="location-name">{location}</span>
									</label>
								{/each}
							</div>

							{#if homeLocations.length === 0}
								<div class="alert alert-warning">
									<strong>⚠️ No home locations selected</strong>
									<p>All matches will be filtered as away matches. Select at least one home location.</p>
								</div>
							{/if}
						</div>
					{/if}

					<!-- Filter Preview -->
					{#if filterPractices || customExcludeTerms.length > 0 || (filterAway && homeLocations.length > 0)}
						<div class="filter-preview">
							<strong>Filter Preview:</strong>
							<div class="filter-stats">
								<span class="stat-item">
									<span class="stat-value">{willImportCount}</span> matches will be imported
								</span>
								<span class="stat-item filtered">
									<span class="stat-value">{filteredCount}</span> will be filtered
								</span>
							</div>
						</div>
					{/if}
				</div>
				{/if}
			</div>

			{#if errorRows.length > 0}
				<div class="section">
					<button class="section-toggle" on:click={() => (showErrors = !showErrors)}>
						<span class="toggle-icon">{showErrors ? '▾' : '▸'}</span>
						<h3>Rows with Errors ({errorRows.length})</h3>
					</button>
					{#if showErrors}
					<p class="section-info">These rows will be skipped:</p>
					<div class="table-container">
						<table class="preview-table">
							<thead>
								<tr>
									<th>Row</th>
									<th>Team</th>
									<th>Date</th>
									<th>Time</th>
									<th>Age Group</th>
									<th>Error</th>
								</tr>
							</thead>
							<tbody>
								{#each errorRows as row}
									<tr class="error-row">
										<td>{row.row_number}</td>
										<td>{row.team_name}</td>
										<td>{row.start_date}</td>
										<td>{row.start_time}</td>
										<td>
											{#if row.age_group}
												{row.age_group}
											{:else}
												<span class="text-muted">—</span>
											{/if}
										</td>
										<td class="error-cell">{row.error}</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
					{/if}
				</div>
			{/if}

			<div class="section">
				<button class="section-toggle" on:click={() => (showValidMatches = !showValidMatches)}>
					<span class="toggle-icon">{showValidMatches ? '▾' : '▸'}</span>
					<h3>Matches to Import ({displayRows.length})</h3>
				</button>
				{#if showValidMatches}
				<div class="table-container">
					<table class="preview-table">
						<thead>
							<tr>
								<th>Row</th>
								<th>Status</th>
								<th>Event Name</th>
								<th>Team</th>
								<th>Age Group</th>
								<th>Date</th>
								<th>Time</th>
								<th>Location</th>
							</tr>
						</thead>
						<tbody>
							{#each displayRows as row}
								<tr>
									<td>{row.row_number}</td>
									<td>
										{#if row.exists_in_db}
											<span class="badge badge-update">Update</span>
										{:else}
											<span class="badge badge-new">New</span>
										{/if}
									</td>
									<td>{row.event_name}</td>
									<td>{row.team_name}</td>
									<td>{row.age_group || '—'}</td>
									<td>{row.start_date}</td>
									<td>{row.start_time} - {row.end_time}</td>
									<td>{row.location}</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
				{/if}
			</div>

			<div class="actions">
				<button on:click={handleStartOver} class="btn btn-secondary">Cancel</button>
				<button
					on:click={handleConfirmImport}
					class="btn btn-primary"
					disabled={importing || rowsToImport.length === 0 || (filterAway && homeLocations.length === 0)}
				>
					{#if importing}
						Importing...
					{:else if filteredCount > 0}
						Import {willImportCount} Matches (Filter {filteredCount})
					{:else}
						Import {rowsToImport.length} Matches
					{/if}
				</button>
			</div>
		</div>
	{:else if step === 'complete'}
		<div class="card">
			<h2>✅ Import Complete</h2>

			<div class="result-summary">
				{#if importResult.created !== undefined}
					<div class="result-item success">
						<span class="result-label">Created:</span>
						<span class="result-value">{importResult.created}</span>
					</div>
				{/if}
				{#if importResult.updated !== undefined && importResult.updated > 0}
					<div class="result-item updated">
						<span class="result-label">Updated:</span>
						<span class="result-value">{importResult.updated}</span>
					</div>
				{/if}
				{#if importResult.imported !== undefined}
					<div class="result-item success">
						<span class="result-label">Total Imported:</span>
						<span class="result-value">{importResult.imported}</span>
					</div>
				{/if}
				{#if importResult.skipped !== undefined && importResult.skipped > 0}
					<div class="result-item">
						<span class="result-label">Skipped (Errors):</span>
						<span class="result-value">{importResult.skipped}</span>
					</div>
				{/if}
				{#if importResult.filtered !== undefined && importResult.filtered > 0}
					<div class="result-item filtered">
						<span class="result-label">Filtered:</span>
						<span class="result-value">{importResult.filtered}</span>
					</div>
				{/if}
				{#if importResult.excluded !== undefined && importResult.excluded > 0}
					<div class="result-item excluded">
						<span class="result-label">Excluded:</span>
						<span class="result-value">{importResult.excluded}</span>
					</div>
				{/if}
			</div>

			{#if importResult.errors && importResult.errors.length > 0}
				<div class="alert alert-warning">
					<strong>Import Errors:</strong>
					<ul>
						{#each importResult.errors as err}
							<li>{err}</li>
						{/each}
					</ul>
				</div>
			{/if}

			<div class="actions">
				<button on:click={handleStartOver} class="btn btn-secondary">Import Another File</button>
				<button on:click={() => goto('/assignor')} class="btn btn-primary">
					View Schedule
				</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.container {
		max-width: 1400px;
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
		font-size: 2rem;
		font-weight: 700;
		color: var(--text-primary);
	}

	h2 {
		font-size: 1.5rem;
		font-weight: 600;
		margin-bottom: 1rem;
		color: var(--text-primary);
	}

	h3 {
		font-size: 1.25rem;
		font-weight: 600;
		margin-bottom: 0.5rem;
		color: var(--text-primary);
	}

	.instructions {
		color: var(--text-secondary);
		margin-bottom: 1.5rem;
		line-height: 1.6;
	}

	.upload-section {
		margin-bottom: 1.5rem;
	}

	.file-input {
		display: none;
	}

	.file-label {
		display: inline-block;
		padding: 0.75rem 1.5rem;
		background-color: white;
		border: 2px dashed var(--border-color);
		border-radius: 0.375rem;
		cursor: pointer;
		color: var(--text-secondary);
		transition: all 0.2s;
	}

	.file-label:hover {
		border-color: var(--primary-color);
		color: var(--primary-color);
	}

	.summary {
		padding: 1rem;
		background-color: var(--bg-secondary);
		border-radius: 0.375rem;
		margin-bottom: 1.5rem;
	}

	.error-count {
		color: var(--error-color);
	}

	.duplicate-count {
		color: #d97706;
	}

	.section {
		margin-bottom: 2rem;
	}

	.section-toggle {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0;
		width: 100%;
		text-align: left;
	}

	.section-toggle:hover {
		opacity: 0.8;
	}

	.section-toggle h3 {
		margin-bottom: 0;
	}

	.toggle-icon {
		font-size: 1rem;
		color: var(--text-secondary);
		flex-shrink: 0;
	}

	.section-info {
		color: var(--text-secondary);
		margin-bottom: 1rem;
		font-size: 0.875rem;
	}

	.table-container {
		overflow-x: auto;
		margin-bottom: 1rem;
	}

	.preview-table {
		width: 100%;
		border-collapse: collapse;
		font-size: 0.875rem;
	}

	.preview-table th {
		text-align: left;
		padding: 0.75rem;
		border-bottom: 2px solid var(--border-color);
		font-weight: 600;
		color: var(--text-primary);
		background-color: var(--bg-secondary);
	}

	.preview-table td {
		padding: 0.75rem;
		border-bottom: 1px solid var(--border-color);
	}

	.preview-table tr:hover {
		background-color: var(--bg-secondary);
	}

	.error-row {
		background-color: #fee;
	}

	.error-row:hover {
		background-color: #fdd;
	}

	.error-cell {
		color: var(--error-color);
		font-weight: 500;
	}

	.text-muted {
		color: var(--text-secondary);
		font-style: italic;
	}

	.badge {
		display: inline-block;
		padding: 0.125rem 0.5rem;
		border-radius: 1rem;
		font-size: 0.75rem;
		font-weight: 600;
		text-transform: uppercase;
		letter-spacing: 0.025em;
	}

	.badge-new {
		background-color: #dcfce7;
		color: #166534;
	}

	.badge-update {
		background-color: #dbeafe;
		color: #1e40af;
	}

	.actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
		margin-top: 1.5rem;
	}

	.result-summary {
		display: flex;
		gap: 2rem;
		margin: 2rem 0;
	}

	.result-item {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.result-item.success .result-value {
		color: var(--success-color);
	}

	.result-label {
		font-weight: 600;
		color: var(--text-primary);
	}

	.result-value {
		font-size: 2rem;
		font-weight: 700;
	}

	.alert {
		padding: 1rem;
		border-radius: 0.375rem;
		margin-bottom: 1.5rem;
	}

	.alert-error {
		background-color: #fee;
		color: var(--error-color);
		border: 1px solid var(--error-color);
	}

	.alert-warning {
		background-color: #fffbeb;
		color: #92400e;
		border: 1px solid #fbbf24;
	}

	.alert ul {
		margin-top: 0.5rem;
		margin-left: 1.5rem;
	}

	.alert li {
		margin-bottom: 0.25rem;
	}

	.btn-secondary {
		background-color: white;
		color: var(--text-primary);
		border: 1px solid var(--border-color);
	}

	.btn-secondary:hover {
		background-color: var(--bg-secondary);
	}

	/* Duplicate Resolution Styles */
	.duplicate-section {
		background-color: #fffbeb;
		padding: 1.5rem;
		border-radius: 0.5rem;
		border: 1px solid #fbbf24;
	}

	.duplicate-group {
		background-color: white;
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		margin-bottom: 1rem;
		overflow: hidden;
	}

	.duplicate-group-header {
		padding: 0.75rem 1rem;
		background-color: #fef3c7;
		border-bottom: 1px solid #fbbf24;
		color: #92400e;
		font-size: 0.875rem;
	}

	.duplicate-selected {
		background-color: #f0fdf4;
	}

	.duplicate-group .preview-table td input[type='checkbox'] {
		width: 1.125rem;
		height: 1.125rem;
		cursor: pointer;
	}

	/* Filter Options Styles */
	.filter-section {
		background-color: var(--bg-secondary);
		padding: 1.5rem;
		border-radius: 0.5rem;
		border: 1px solid var(--border-color);
	}

	.filter-options {
		display: flex;
		flex-direction: column;
		gap: 1.5rem;
	}

	.filter-checkbox {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		cursor: pointer;
		padding: 1rem;
		background-color: white;
		border-radius: 0.375rem;
		border: 1px solid var(--border-color);
		transition: all 0.2s;
	}

	.filter-checkbox:hover {
		border-color: var(--primary-color);
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}

	.filter-checkbox input[type='checkbox'] {
		margin-top: 0.25rem;
		width: 1.125rem;
		height: 1.125rem;
		cursor: pointer;
	}

	.filter-label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
		flex: 1;
	}

	.filter-description {
		color: var(--text-secondary);
		font-size: 0.875rem;
		font-weight: normal;
	}

	.custom-exclude-panel {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
		padding: 1rem;
		background-color: white;
		border-radius: 0.375rem;
		border: 1px solid var(--border-color);
	}

	.custom-exclude-panel label {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;
	}

	.text-input {
		padding: 0.5rem 0.75rem;
		border: 1px solid var(--border-color);
		border-radius: 0.375rem;
		font-size: 0.875rem;
		width: 100%;
		box-sizing: border-box;
	}

	.text-input:focus {
		outline: none;
		border-color: var(--primary-color);
		box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.15);
	}

	.exclude-tags {
		display: flex;
		gap: 0.5rem;
		flex-wrap: wrap;
	}

	.exclude-tag {
		display: inline-block;
		padding: 0.25rem 0.625rem;
		background-color: #fef3c7;
		border: 1px solid #fbbf24;
		border-radius: 1rem;
		font-size: 0.75rem;
		color: #92400e;
		font-weight: 500;
	}

	.home-locations-panel {
		background-color: white;
		padding: 1.5rem;
		border-radius: 0.375rem;
		border: 1px solid var(--border-color);
		margin-left: 2rem;
	}

	.locations-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 1rem;
		flex-wrap: wrap;
		gap: 0.5rem;
	}

	.locations-header h4 {
		font-size: 1rem;
		font-weight: 600;
		color: var(--text-primary);
		margin: 0;
	}

	.locations-actions {
		display: flex;
		gap: 1rem;
	}

	.btn-link {
		background: none;
		border: none;
		color: var(--primary-color);
		cursor: pointer;
		font-size: 0.875rem;
		padding: 0.25rem 0.5rem;
		text-decoration: underline;
		transition: opacity 0.2s;
	}

	.btn-link:hover:not(:disabled) {
		opacity: 0.8;
	}

	.btn-link:disabled {
		color: var(--text-secondary);
		cursor: not-allowed;
		text-decoration: none;
	}

	.locations-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
		gap: 0.75rem;
		margin-bottom: 1rem;
	}

	.location-checkbox {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem;
		background-color: var(--bg-secondary);
		border-radius: 0.25rem;
		cursor: pointer;
		transition: all 0.2s;
	}

	.location-checkbox:hover {
		background-color: #e5e7eb;
	}

	.location-checkbox input[type='checkbox'] {
		cursor: pointer;
	}

	.location-name {
		font-size: 0.875rem;
		color: var(--text-primary);
	}

	.filter-preview {
		background-color: #eff6ff;
		border: 1px solid #3b82f6;
		border-radius: 0.375rem;
		padding: 1rem;
	}

	.filter-preview strong {
		color: var(--text-primary);
		display: block;
		margin-bottom: 0.5rem;
	}

	.filter-stats {
		display: flex;
		gap: 2rem;
		flex-wrap: wrap;
	}

	.stat-item {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.875rem;
		color: var(--text-secondary);
	}

	.stat-item .stat-value {
		font-size: 1.25rem;
		font-weight: 700;
		color: var(--success-color);
	}

	.stat-item.filtered .stat-value {
		color: #d97706;
	}

	.result-item.updated .result-value {
		color: #3b82f6;
	}

	.result-item.filtered .result-value {
		color: #d97706;
	}

	.result-item.excluded .result-value {
		color: #6b7280;
	}

	@media (max-width: 768px) {
		.header {
			flex-direction: column;
			align-items: flex-start;
		}

		.preview-table {
			font-size: 0.75rem;
		}

		.preview-table th,
		.preview-table td {
			padding: 0.5rem;
		}

		.result-summary {
			flex-direction: column;
			gap: 1rem;
		}

		.actions {
			flex-direction: column;
		}

		.btn {
			width: 100%;
		}

		.home-locations-panel {
			margin-left: 0;
		}

		.locations-grid {
			grid-template-columns: 1fr;
		}

		.locations-header {
			flex-direction: column;
			align-items: flex-start;
		}

		.filter-stats {
			flex-direction: column;
			gap: 0.5rem;
		}
	}
</style>
