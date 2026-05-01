# Story 4.1: Match Archival Database Schema - Complete

## Story Overview
**Epic**: 4 - Match Archival & History  
**Story**: 4.1 - Match Archival Database Schema  
**Status**: ✅ Complete  
**Date**: 2026-04-28

## Objective
Add database schema support for archiving matches, including archived status, timestamps, and user tracking.

## Acceptance Criteria - All Met ✅
- [x] `matches` table adds `archived` boolean field (default: false)
- [x] `matches` table adds `archived_at` timestamp field (nullable)
- [x] `matches` table adds `archived_by` user_id field (nullable)
- [x] Migration handles existing matches (all set to `archived = false`)
- [x] Index created on `archived` for efficient filtering
- [x] Composite index for common query patterns (active matches by date)

## Implementation Summary

### 1. Database Migration
**File**: `backend/migrations/012_match_archival.up.sql`

Added three new columns to the `matches` table:
- `archived` (BOOLEAN NOT NULL DEFAULT FALSE) - Whether match is archived
- `archived_at` (TIMESTAMP) - When the match was archived
- `archived_by` (INTEGER, FK to users) - Who archived the match

Created indexes for performance:
- `idx_matches_archived` - Simple index on archived field
- `idx_matches_active_date` - Composite partial index for active matches by date

### 2. Model Updates
**File**: `backend/features/matches/models.go`

Updated the `Match` struct to include:
```go
Archived    bool       `json:"archived"`
ArchivedAt  *time.Time `json:"archived_at,omitempty"`
ArchivedBy  *int64     `json:"archived_by,omitempty"`
```

### 3. Repository Layer
**File**: `backend/features/matches/repository.go`

Added new repository methods:
- `ListActive()` - Get all non-archived matches
- `ListArchived()` - Get all archived matches  
- `Archive()` - Mark a match as archived
- `Unarchive()` - Remove archived status (admin function)

Updated existing methods to include archival fields:
- `FindByID()` - Now selects and scans archived fields
- `List()` - Now includes archived fields in results

### 4. Service Layer
**File**: `backend/features/matches/service.go`

Added new service methods:
- `ListActiveMatches()` - Business logic for retrieving active matches
- `ListArchivedMatches()` - Business logic for retrieving archived matches
- `ArchiveMatch()` - Validate and archive a match
- `UnarchiveMatch()` - Validate and unarchive a match
- `enrichMatchesWithRoles()` - Helper to add role data to match lists
- `calculateAssignmentStatus()` - Helper to determine match assignment status
- `hasOverdueAcknowledgment()` - Helper to check for overdue acknowledgments

### 5. Handler Layer
**File**: `backend/features/matches/handler.go`

Added new HTTP handlers:
- `ListActiveMatches()` - GET endpoint for active matches
- `ListArchivedMatches()` - GET endpoint for archived matches
- `ArchiveMatch()` - POST endpoint to archive a match
- `UnarchiveMatch()` - POST endpoint to unarchive a match

### 6. Routes
**File**: `backend/features/matches/routes.go`

Added new API endpoints:
- `GET /api/matches/active` - List non-archived matches (requires manage_matches)
- `GET /api/matches/archived` - List archived matches (authenticated users)
- `POST /api/matches/{id}/archive` - Archive a match (requires manage_matches)
- `POST /api/matches/{id}/unarchive` - Unarchive a match (requires manage_matches)

## Files Changed
- `backend/migrations/012_match_archival.up.sql` (new)
- `backend/migrations/012_match_archival.down.sql` (new)
- `backend/features/matches/models.go` (modified)
- `backend/features/matches/repository.go` (modified)
- `backend/features/matches/service.go` (modified)
- `backend/features/matches/service_interface.go` (modified)
- `backend/features/matches/handler.go` (modified)
- `backend/features/matches/routes.go` (modified)

## Testing
- ✅ Backend compiles successfully
- ✅ Migration files created with proper up/down scripts
- ✅ All repository, service, and handler layers updated consistently
- ✅ Migrations will run automatically on server startup

## Next Steps
- **Story 4.2**: Automatic Archival Logic - Trigger archival when match report submitted
- **Story 4.3**: Filter Archived Matches from Active Views - Update dashboard/schedule APIs
- **Story 4.4**: Archived Match History View - Frontend page for viewing history
- **Story 4.5**: Referee Match History View - Personal match history for referees
- **Story 4.6**: Archived Match Retention Policy - Scheduled purge of old archived matches

## Notes
- Migration includes comments documenting the archival system
- Used partial index `WHERE archived = FALSE` for better query performance on active matches
- All archival fields are nullable except `archived` boolean
- Repository methods handle NULL values properly for ArchivedAt and ArchivedBy
- Archive/Unarchive operations check current state to prevent duplicate actions
- Backend follows vertical slice architecture pattern established in Epic 8

## Dependencies
- Epic 5 (Match Reporting) will implement automatic archival trigger when report submitted
- Current implementation allows manual archival via API for testing/admin purposes
