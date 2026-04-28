# Story 4.3: Filter Archived Matches from Active Views - Complete

## Story Overview
**Epic**: 4 - Match Archival & History  
**Story**: 4.3 - Filter Archived Matches from Active Views  
**Status**: ✅ Complete  
**Date**: 2026-04-28

## Objective
Exclude archived matches from active views (dashboard, schedule, referee matches) so users only see upcoming/active matches by default.

## Acceptance Criteria - All Met ✅
- [x] Dashboard API filters `archived = false` by default
- [x] Schedule page API filters `archived = false` by default
- [x] Referee "My Matches" view filters `archived = false` by default
- [x] No changes to assignment logic (archived matches still linked to referees for history)
- [x] Conflict checking excludes archived matches
- [x] Match existence checks exclude archived matches

## Implementation Summary

### 1. Assignor Schedule View (Dashboard)
**File**: `backend/features/matches/service.go`

Updated `ListMatches()` service method to use `ListActive()` instead of `List()`:
- Now calls `s.repo.ListActive(ctx)` which automatically filters `archived = FALSE`
- This is the primary endpoint used by assignors for viewing/scheduling matches
- Added comment documenting this is the default view for active matches only

### 2. Referee "My Matches" View
**File**: `backend/availability.go`

Updated the `/api/referee/matches` endpoint query:
- Added `AND m.archived = FALSE` to main matches query (line 102)
- Ensures referees only see active, non-archived matches in their dashboard
- Updated comment to clarify "non-cancelled, non-archived matches"

### 3. Conflict Detection - Referee View
**File**: `backend/availability.go`

Updated conflict checking query for referee assignments:
- Added `AND m2.archived = FALSE` to conflict detection query (line 251)
- Ensures only active matches are considered when checking for scheduling conflicts
- Prevents false conflict warnings from archived/completed matches

### 4. Conflict Detection - Assignor View
**File**: `backend/features/assignments/repository.go`

Updated `FindConflictingAssignments()` method:
- Added `AND m.archived = FALSE` to conflict query
- Ensures assignors don't see false conflicts from archived matches
- Updated method comment to clarify "active (non-archived) assignments"

### 5. Match Existence Validation
**Files**: 
- `backend/features/assignments/repository.go`
- `backend/features/matches/repository.go`

Updated `MatchExists()` methods:
- Added `AND archived = FALSE` to existence check
- Prevents assigning referees to archived matches
- Prevents updating or adding role slots to archived matches
- Updated comments to clarify "active and not archived"

## Files Changed
- `backend/features/matches/service.go` (modified)
- `backend/features/matches/repository.go` (modified)
- `backend/features/assignments/repository.go` (modified)
- `backend/availability.go` (modified)

## Testing
- ✅ Backend compiles successfully
- ✅ All match list endpoints now filter archived = FALSE by default
- ✅ Conflict detection excludes archived matches
- ✅ Match validation prevents operations on archived matches
- ✅ Assignment data still preserved for archived matches (history intact)

## API Behavior Changes

### Before Story 4.3
- `GET /api/matches` - Returned ALL matches (active + archived)
- `GET /api/referee/matches` - Returned ALL matches referee was eligible for
- Conflict checking considered ALL matches including archived
- Could assign referees to archived matches

### After Story 4.3
- `GET /api/matches` - Returns only ACTIVE matches (archived = FALSE)
- `GET /api/referee/matches` - Returns only ACTIVE matches
- Conflict checking only considers ACTIVE matches
- Cannot assign referees to archived matches (validation fails)

### Archived Match Access
- `GET /api/matches/archived` - New endpoint for viewing match history (from Story 4.1)
- `GET /api/matches/active` - Explicit endpoint for active matches only (from Story 4.1)

## Database Query Performance
All queries updated benefit from the `idx_matches_archived` and `idx_matches_active_date` indexes created in Story 4.1:
- Simple index on `archived` column
- Partial index on `(archived, match_date) WHERE archived = FALSE`
- These indexes make filtering active matches very efficient

## Assignment History Preservation
✅ **Important**: Assignment records (`match_roles` table) are NOT modified when matches are archived.
- Assignments remain linked to matches via `match_id` foreign key
- This preserves complete history for archived matches
- Referees can still see their past assignments in history views (Story 4.5)
- Audit logs for assignments remain intact

## Next Steps
- **Story 4.4**: Archived Match History View (All Users) - Frontend page for viewing archived matches
- **Story 4.5**: Referee Match History View - Personal match history for referees
- **Story 4.6**: Archived Match Retention Policy - Scheduled purge of old data

## Notes
- All active views now consistently filter out archived matches
- No breaking changes to API contracts (same endpoints, just better filtering)
- Performance improved by using partial indexes for active match queries
- Backward compatible: `List()` method still exists for admin purposes if needed
- Assignment logic untouched: matches can still be assigned while active, data preserved when archived
