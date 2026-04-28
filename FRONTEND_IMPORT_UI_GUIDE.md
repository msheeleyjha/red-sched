# CSV Import UI Implementation Guide

## Overview
Complete step-by-step guide for implementing the CSV import UI with filter configuration, including home location selection.

## User Flow

```
┌─────────────────────────────────────────────────────────────┐
│ STEP 1: Upload CSV File                                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Import Match Schedule                                       │
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │ [Choose File]  matches.csv                     │         │
│  └────────────────────────────────────────────────┘         │
│                                                              │
│                              [Upload & Preview]              │
│                                                              │
└─────────────────────────────────────────────────────────────┘

                            ↓

┌─────────────────────────────────────────────────────────────┐
│ STEP 2: Preview & Configure Filters                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Import Preview: 50 matches found                            │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐ │
│  │ Filter Options                                         │ │
│  │                                                        │ │
│  │ ☑ Filter practice matches                             │ │
│  │   Matches with "Practice" in team name will be        │ │
│  │   excluded from import                                │ │
│  │                                                        │ │
│  │ ☑ Filter away matches                                 │ │
│  │   Matches at away locations will be excluded          │ │
│  │                                                        │ │
│  │   Select your HOME locations (8 found in file):       │ │
│  │   ┌────────────────────────────────────────────────┐ │ │
│  │   │ ☑ Central Park Field A         (12 matches)   │ │ │
│  │   │ ☐ Lincoln Memorial Field       (5 matches)    │ │ │
│  │   │ ☐ Riverside Tournament Complex (8 matches)    │ │ │
│  │   │ ☑ Smith Complex Field 1        (15 matches)   │ │ │
│  │   │ ☑ Smith Complex Field 2        (10 matches)   │ │ │
│  │   └────────────────────────────────────────────────┘ │ │
│  │                                                        │ │
│  │   [Select All]  [Clear All]                           │ │
│  └────────────────────────────────────────────────────────┘ │
│                                                              │
│  Import Summary:                                             │
│  • Will import: 37 matches                                   │
│  • Will filter: 8 practice matches                           │
│  • Will filter: 5 away matches                               │
│                                                              │
│                    [Confirm Import]  [Cancel]                │
│                                                              │
└─────────────────────────────────────────────────────────────┘

                            ↓

┌─────────────────────────────────────────────────────────────┐
│ STEP 3: Import Complete                                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ✅ Import Successful!                                       │
│                                                              │
│  Results:                                                    │
│  • 30 matches created                                        │
│  • 7 matches updated                                         │
│  • 8 practice matches filtered                               │
│  • 5 away matches filtered                                   │
│                                                              │
│                              [View Matches]                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## API Integration

### Step 1: Upload and Parse CSV

**Request**:
```typescript
const formData = new FormData();
formData.append('file', csvFile);

const response = await fetch('/api/matches/import/preview', {
  method: 'POST',
  body: formData,
});

const preview: ImportPreviewResponse = await response.json();
```

**Response**:
```typescript
interface ImportPreviewResponse {
  rows: CSVRow[];
  duplicates: DuplicateMatchGroup[];
  unique_locations: string[];  // List of all locations in CSV
}

// Example:
{
  "rows": [
    {
      "row_number": 1,
      "team_name": "U12 Girls - Falcons",
      "location": "Smith Complex Field 1",
      "reference_id": "MATCH-001",
      ...
    },
    // ... 49 more rows
  ],
  "duplicates": [],
  "unique_locations": [
    "Central Park Field A",
    "Lincoln Memorial Field",
    "Riverside Tournament Complex",
    "Smith Complex Field 1",
    "Smith Complex Field 2"
  ]
}
```

### Step 2: Configure Filters and Confirm

**Request**:
```typescript
const confirmRequest = {
  rows: preview.rows,
  filters: {
    filter_practices: filterPractices,
    filter_away: filterAway,
    home_locations: selectedHomeLocations,  // User's selections
  },
};

const response = await fetch('/api/matches/import/confirm', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(confirmRequest),
});

const result: ImportResult = await response.json();
```

**Response**:
```typescript
interface ImportResult {
  created: number;
  updated: number;
  skipped: number;
  filtered: number;
  excluded: number;
  errors: string[];
}

// Example:
{
  "created": 30,
  "updated": 7,
  "skipped": 0,
  "filtered": 13,  // 8 practices + 5 away
  "excluded": 0,
  "errors": []
}
```

## React Implementation

### Complete Component

```typescript
import React, { useState, useMemo } from 'react';

interface CSVRow {
  row_number: number;
  team_name: string;
  location: string;
  start_date: string;
  start_time: string;
  reference_id: string;
  error?: string;
  filter_reason?: string;
}

interface ImportPreviewResponse {
  rows: CSVRow[];
  duplicates: any[];
  unique_locations: string[];
}

interface ImportResult {
  created: number;
  updated: number;
  skipped: number;
  filtered: number;
  excluded: number;
  errors: string[];
}

export function CSVImportPage() {
  // Step 1: Upload state
  const [file, setFile] = useState<File | null>(null);
  const [isUploading, setIsUploading] = useState(false);

  // Step 2: Preview state
  const [preview, setPreview] = useState<ImportPreviewResponse | null>(null);
  const [filterPractices, setFilterPractices] = useState(true);
  const [filterAway, setFilterAway] = useState(false);
  const [homeLocations, setHomeLocations] = useState<string[]>([]);

  // Step 3: Result state
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const [isImporting, setIsImporting] = useState(false);

  // Calculate location statistics
  const locationStats = useMemo(() => {
    if (!preview) return {};
    
    const stats: Record<string, number> = {};
    preview.rows.forEach(row => {
      if (row.location && !row.error) {
        stats[row.location] = (stats[row.location] || 0) + 1;
      }
    });
    return stats;
  }, [preview]);

  // Calculate import preview counts
  const importPreview = useMemo(() => {
    if (!preview) return { willImport: 0, practices: 0, away: 0 };

    let practices = 0;
    let away = 0;

    preview.rows.forEach(row => {
      if (row.error) return;

      // Check practice filter
      if (filterPractices && row.team_name.toLowerCase().includes('practice')) {
        practices++;
        return;
      }

      // Check away filter
      if (filterAway && homeLocations.length > 0) {
        const isHome = homeLocations.some(loc => 
          row.location.toLowerCase().includes(loc.toLowerCase())
        );
        if (!isHome) {
          away++;
          return;
        }
      }
    });

    const willImport = preview.rows.filter(r => !r.error).length - practices - away;
    return { willImport, practices, away };
  }, [preview, filterPractices, filterAway, homeLocations]);

  // Step 1: Upload and parse CSV
  const handleUpload = async () => {
    if (!file) return;

    setIsUploading(true);
    try {
      const formData = new FormData();
      formData.append('file', file);

      const response = await fetch('/api/matches/import/preview', {
        method: 'POST',
        body: formData,
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Upload failed');
      }

      const data: ImportPreviewResponse = await response.json();
      setPreview(data);
      
      // Auto-select all locations as home by default
      setHomeLocations(data.unique_locations);
    } catch (error) {
      console.error('Upload failed:', error);
      alert(`Upload failed: ${error.message}`);
    } finally {
      setIsUploading(false);
    }
  };

  // Step 2: Confirm import with filters
  const handleConfirmImport = async () => {
    if (!preview) return;

    setIsImporting(true);
    try {
      const response = await fetch('/api/matches/import/confirm', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          rows: preview.rows,
          filters: {
            filter_practices: filterPractices,
            filter_away: filterAway,
            home_locations: homeLocations,
          },
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Import failed');
      }

      const result: ImportResult = await response.json();
      setImportResult(result);
      setPreview(null);  // Clear preview
    } catch (error) {
      console.error('Import failed:', error);
      alert(`Import failed: ${error.message}`);
    } finally {
      setIsImporting(false);
    }
  };

  // Helper: Toggle location selection
  const toggleLocation = (location: string) => {
    setHomeLocations(prev =>
      prev.includes(location)
        ? prev.filter(l => l !== location)
        : [...prev, location]
    );
  };

  // Helper: Select all locations
  const selectAllLocations = () => {
    if (preview) {
      setHomeLocations(preview.unique_locations);
    }
  };

  // Helper: Clear all locations
  const clearAllLocations = () => {
    setHomeLocations([]);
  };

  // Render: Upload step
  if (!preview && !importResult) {
    return (
      <div className="import-page">
        <h1>Import Match Schedule</h1>
        
        <div className="upload-section">
          <input
            type="file"
            accept=".csv"
            onChange={e => setFile(e.target.files?.[0] || null)}
          />
          
          <button
            onClick={handleUpload}
            disabled={!file || isUploading}
          >
            {isUploading ? 'Uploading...' : 'Upload & Preview'}
          </button>
        </div>

        <div className="help-text">
          <p>Upload a CSV file exported from Stack Team App.</p>
          <p>Required columns: event_name, team_name, start_date, start_time, end_time, location</p>
        </div>
      </div>
    );
  }

  // Render: Preview and filter configuration step
  if (preview) {
    return (
      <div className="import-page">
        <h1>Import Preview: {preview.rows.length} matches found</h1>

        <div className="filter-section">
          <h2>Filter Options</h2>

          {/* Practice filter */}
          <div className="filter-option">
            <label>
              <input
                type="checkbox"
                checked={filterPractices}
                onChange={e => setFilterPractices(e.target.checked)}
              />
              <strong>Filter practice matches</strong>
            </label>
            <p className="filter-help">
              Matches with "Practice" in team name will be excluded from import
            </p>
          </div>

          {/* Away filter */}
          <div className="filter-option">
            <label>
              <input
                type="checkbox"
                checked={filterAway}
                onChange={e => setFilterAway(e.target.checked)}
              />
              <strong>Filter away matches</strong>
            </label>
            <p className="filter-help">
              Matches at away locations will be excluded from import
            </p>

            {filterAway && (
              <div className="home-locations">
                <h3>
                  Select your HOME locations ({preview.unique_locations.length} found in file):
                </h3>

                <div className="location-actions">
                  <button onClick={selectAllLocations}>Select All</button>
                  <button onClick={clearAllLocations}>Clear All</button>
                  <span className="selected-count">
                    {homeLocations.length} of {preview.unique_locations.length} selected
                  </span>
                </div>

                <div className="location-list">
                  {preview.unique_locations.map(location => (
                    <label key={location} className="location-item">
                      <input
                        type="checkbox"
                        checked={homeLocations.includes(location)}
                        onChange={() => toggleLocation(location)}
                      />
                      <span className="location-name">{location}</span>
                      <span className="location-count">
                        ({locationStats[location] || 0} matches)
                      </span>
                    </label>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Import preview summary */}
        <div className="import-summary">
          <h3>Import Summary:</h3>
          <ul>
            <li>
              <strong>Will import: {importPreview.willImport} matches</strong>
            </li>
            {importPreview.practices > 0 && (
              <li>Will filter: {importPreview.practices} practice matches</li>
            )}
            {importPreview.away > 0 && (
              <li>Will filter: {importPreview.away} away matches</li>
            )}
          </ul>
        </div>

        {/* Action buttons */}
        <div className="import-actions">
          <button
            className="btn-primary"
            onClick={handleConfirmImport}
            disabled={isImporting}
          >
            {isImporting ? 'Importing...' : 'Confirm Import'}
          </button>
          <button
            className="btn-secondary"
            onClick={() => setPreview(null)}
            disabled={isImporting}
          >
            Cancel
          </button>
        </div>
      </div>
    );
  }

  // Render: Import complete step
  if (importResult) {
    return (
      <div className="import-page">
        <h1>✅ Import Successful!</h1>

        <div className="import-results">
          <h2>Results:</h2>
          <ul>
            <li>
              <strong>{importResult.created} matches created</strong>
            </li>
            <li>
              <strong>{importResult.updated} matches updated</strong>
            </li>
            {importResult.filtered > 0 && (
              <li>{importResult.filtered} matches filtered</li>
            )}
            {importResult.excluded > 0 && (
              <li>{importResult.excluded} matches excluded</li>
            )}
            {importResult.skipped > 0 && (
              <li>{importResult.skipped} rows skipped (errors)</li>
            )}
          </ul>

          {importResult.errors.length > 0 && (
            <div className="import-errors">
              <h3>Warnings:</h3>
              <ul>
                {importResult.errors.map((error, i) => (
                  <li key={i}>{error}</li>
                ))}
              </ul>
            </div>
          )}
        </div>

        <div className="import-actions">
          <button
            className="btn-primary"
            onClick={() => window.location.href = '/matches'}
          >
            View Matches
          </button>
          <button
            className="btn-secondary"
            onClick={() => {
              setImportResult(null);
              setFile(null);
            }}
          >
            Import Another File
          </button>
        </div>
      </div>
    );
  }

  return null;
}
```

### CSS Styling

```css
/* Import page layout */
.import-page {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
}

.import-page h1 {
  margin-bottom: 1.5rem;
}

/* Upload section */
.upload-section {
  display: flex;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1rem;
}

.help-text {
  color: #666;
  font-size: 0.9rem;
  margin-top: 1rem;
}

/* Filter section */
.filter-section {
  background: #f5f5f5;
  padding: 1.5rem;
  border-radius: 8px;
  margin-bottom: 1.5rem;
}

.filter-section h2 {
  margin-top: 0;
  margin-bottom: 1rem;
}

.filter-option {
  margin-bottom: 1.5rem;
}

.filter-option label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
}

.filter-help {
  margin: 0.5rem 0 0 1.75rem;
  color: #666;
  font-size: 0.9rem;
}

/* Home locations section */
.home-locations {
  margin-top: 1rem;
  padding: 1rem;
  background: white;
  border-radius: 4px;
  border: 1px solid #ddd;
}

.home-locations h3 {
  margin: 0 0 1rem 0;
  font-size: 1rem;
}

.location-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  margin-bottom: 1rem;
}

.selected-count {
  margin-left: auto;
  color: #666;
  font-size: 0.9rem;
}

.location-list {
  max-height: 300px;
  overflow-y: auto;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 0.5rem;
}

.location-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  border-radius: 4px;
}

.location-item:hover {
  background: #f9f9f9;
}

.location-name {
  flex: 1;
}

.location-count {
  color: #666;
  font-size: 0.9rem;
}

/* Import summary */
.import-summary {
  background: #e3f2fd;
  padding: 1rem;
  border-radius: 4px;
  margin-bottom: 1.5rem;
}

.import-summary h3 {
  margin-top: 0;
}

.import-summary ul {
  margin-bottom: 0;
}

/* Import actions */
.import-actions {
  display: flex;
  gap: 1rem;
}

.btn-primary {
  background: #1976d2;
  color: white;
  padding: 0.75rem 1.5rem;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-primary:hover:not(:disabled) {
  background: #1565c0;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background: white;
  color: #333;
  padding: 0.75rem 1.5rem;
  border: 1px solid #ddd;
  border-radius: 4px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-secondary:hover:not(:disabled) {
  background: #f5f5f5;
}

/* Import results */
.import-results h2 {
  margin-bottom: 1rem;
}

.import-results ul {
  list-style: none;
  padding: 0;
}

.import-results ul li {
  padding: 0.5rem 0;
  border-bottom: 1px solid #eee;
}

.import-errors {
  margin-top: 1.5rem;
  padding: 1rem;
  background: #fff3cd;
  border-radius: 4px;
}

.import-errors h3 {
  margin-top: 0;
  color: #856404;
}
```

## Vue.js Implementation (Alternative)

```vue
<template>
  <div class="import-page">
    <!-- Step 1: Upload -->
    <div v-if="!preview && !importResult">
      <h1>Import Match Schedule</h1>
      
      <div class="upload-section">
        <input
          type="file"
          accept=".csv"
          @change="handleFileSelect"
        />
        
        <button
          @click="handleUpload"
          :disabled="!file || isUploading"
        >
          {{ isUploading ? 'Uploading...' : 'Upload & Preview' }}
        </button>
      </div>
    </div>

    <!-- Step 2: Preview and Configure -->
    <div v-if="preview">
      <h1>Import Preview: {{ preview.rows.length }} matches found</h1>

      <div class="filter-section">
        <h2>Filter Options</h2>

        <!-- Practice filter -->
        <div class="filter-option">
          <label>
            <input
              type="checkbox"
              v-model="filterPractices"
            />
            <strong>Filter practice matches</strong>
          </label>
          <p class="filter-help">
            Matches with "Practice" in team name will be excluded
          </p>
        </div>

        <!-- Away filter -->
        <div class="filter-option">
          <label>
            <input
              type="checkbox"
              v-model="filterAway"
            />
            <strong>Filter away matches</strong>
          </label>

          <div v-if="filterAway" class="home-locations">
            <h3>Select your HOME locations ({{ preview.unique_locations.length }} found):</h3>

            <div class="location-actions">
              <button @click="selectAllLocations">Select All</button>
              <button @click="clearAllLocations">Clear All</button>
              <span class="selected-count">
                {{ homeLocations.length }} of {{ preview.unique_locations.length }} selected
              </span>
            </div>

            <div class="location-list">
              <label
                v-for="location in preview.unique_locations"
                :key="location"
                class="location-item"
              >
                <input
                  type="checkbox"
                  :value="location"
                  v-model="homeLocations"
                />
                <span class="location-name">{{ location }}</span>
                <span class="location-count">
                  ({{ locationStats[location] || 0 }} matches)
                </span>
              </label>
            </div>
          </div>
        </div>
      </div>

      <!-- Import summary -->
      <div class="import-summary">
        <h3>Import Summary:</h3>
        <ul>
          <li><strong>Will import: {{ importPreviewCounts.willImport }} matches</strong></li>
          <li v-if="importPreviewCounts.practices > 0">
            Will filter: {{ importPreviewCounts.practices }} practice matches
          </li>
          <li v-if="importPreviewCounts.away > 0">
            Will filter: {{ importPreviewCounts.away }} away matches
          </li>
        </ul>
      </div>

      <div class="import-actions">
        <button
          class="btn-primary"
          @click="handleConfirmImport"
          :disabled="isImporting"
        >
          {{ isImporting ? 'Importing...' : 'Confirm Import' }}
        </button>
        <button
          class="btn-secondary"
          @click="cancelPreview"
          :disabled="isImporting"
        >
          Cancel
        </button>
      </div>
    </div>

    <!-- Step 3: Results -->
    <div v-if="importResult">
      <h1>✅ Import Successful!</h1>
      
      <div class="import-results">
        <h2>Results:</h2>
        <ul>
          <li><strong>{{ importResult.created }} matches created</strong></li>
          <li><strong>{{ importResult.updated }} matches updated</strong></li>
          <li v-if="importResult.filtered > 0">
            {{ importResult.filtered }} matches filtered
          </li>
        </ul>
      </div>

      <div class="import-actions">
        <button class="btn-primary" @click="goToMatches">
          View Matches
        </button>
        <button class="btn-secondary" @click="reset">
          Import Another File
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue';

interface ImportPreviewResponse {
  rows: any[];
  duplicates: any[];
  unique_locations: string[];
}

const file = ref<File | null>(null);
const preview = ref<ImportPreviewResponse | null>(null);
const filterPractices = ref(true);
const filterAway = ref(false);
const homeLocations = ref<string[]>([]);
const isUploading = ref(false);
const isImporting = ref(false);
const importResult = ref(null);

// Calculate location statistics
const locationStats = computed(() => {
  if (!preview.value) return {};
  
  const stats: Record<string, number> = {};
  preview.value.rows.forEach(row => {
    if (row.location && !row.error) {
      stats[row.location] = (stats[row.location] || 0) + 1;
    }
  });
  return stats;
});

// Calculate import preview
const importPreviewCounts = computed(() => {
  if (!preview.value) return { willImport: 0, practices: 0, away: 0 };

  let practices = 0;
  let away = 0;

  preview.value.rows.forEach(row => {
    if (row.error) return;

    if (filterPractices.value && row.team_name.toLowerCase().includes('practice')) {
      practices++;
      return;
    }

    if (filterAway.value && homeLocations.value.length > 0) {
      const isHome = homeLocations.value.some(loc =>
        row.location.toLowerCase().includes(loc.toLowerCase())
      );
      if (!isHome) {
        away++;
        return;
      }
    }
  });

  const willImport = preview.value.rows.filter(r => !r.error).length - practices - away;
  return { willImport, practices, away };
});

const handleFileSelect = (event: Event) => {
  const target = event.target as HTMLInputElement;
  file.value = target.files?.[0] || null;
};

const handleUpload = async () => {
  if (!file.value) return;

  isUploading.value = true;
  try {
    const formData = new FormData();
    formData.append('file', file.value);

    const response = await fetch('/api/matches/import/preview', {
      method: 'POST',
      body: formData,
    });

    const data = await response.json();
    preview.value = data;
    
    // Auto-select all locations as home
    homeLocations.value = [...data.unique_locations];
  } catch (error) {
    alert('Upload failed');
  } finally {
    isUploading.value = false;
  }
};

const handleConfirmImport = async () => {
  if (!preview.value) return;

  isImporting.value = true;
  try {
    const response = await fetch('/api/matches/import/confirm', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        rows: preview.value.rows,
        filters: {
          filter_practices: filterPractices.value,
          filter_away: filterAway.value,
          home_locations: homeLocations.value,
        },
      }),
    });

    importResult.value = await response.json();
    preview.value = null;
  } finally {
    isImporting.value = false;
  }
};

const selectAllLocations = () => {
  if (preview.value) {
    homeLocations.value = [...preview.value.unique_locations];
  }
};

const clearAllLocations = () => {
  homeLocations.value = [];
};

const cancelPreview = () => {
  preview.value = null;
};

const goToMatches = () => {
  window.location.href = '/matches';
};

const reset = () => {
  file.value = null;
  preview.value = null;
  importResult.value = null;
  filterPractices.value = true;
  filterAway.value = false;
  homeLocations.value = [];
};
</script>
```

## Key Features

### 1. Auto-Select All Locations
When the CSV is uploaded, **all locations are auto-selected as "home"** by default:
```typescript
// Auto-select all locations as home by default
setHomeLocations(data.unique_locations);
```

This makes sense because:
- Most imports are likely all home matches
- User can easily uncheck away venues
- Better UX than starting with nothing selected

### 2. Location Match Count
Show how many matches are at each location:
```typescript
<span className="location-count">
  ({locationStats[location] || 0} matches)
</span>
```

Helps user understand the impact of their selection.

### 3. Select All / Clear All
Quick buttons for bulk selection:
```typescript
<button onClick={selectAllLocations}>Select All</button>
<button onClick={clearAllLocations}>Clear All</button>
```

### 4. Live Import Preview
Updates in real-time as filters change:
```typescript
Import Summary:
• Will import: 37 matches
• Will filter: 8 practice matches
• Will filter: 5 away matches
```

Shows exactly what will happen before confirming.

### 5. Selected Count
Shows progress:
```
3 of 8 locations selected
```

## Testing Checklist

### Backend (Already Complete ✅)
- [x] `unique_locations` returned in preview response
- [x] Locations alphabetically sorted
- [x] Empty locations excluded
- [x] Duplicates removed
- [x] Only valid rows included

### Frontend (To Implement)
- [ ] Upload CSV file
- [ ] Display unique locations list
- [ ] Check/uncheck individual locations
- [ ] "Select All" button works
- [ ] "Clear All" button works
- [ ] Match counts displayed per location
- [ ] Import preview updates live
- [ ] Confirm import sends correct filters
- [ ] Display import results

## Summary

The backend is **100% ready** to support this UI flow. The `unique_locations` array provides everything needed to build a great user experience.

**What the backend gives you:**
- ✅ List of all locations in the CSV
- ✅ Alphabetically sorted
- ✅ Deduplicated
- ✅ Only from valid rows

**What the frontend needs to do:**
1. Display locations as checkboxes
2. Let user select which are "home"
3. Send selected locations in `filters.home_locations`
4. Show import preview

The complete React and Vue examples above show exactly how to implement this!
