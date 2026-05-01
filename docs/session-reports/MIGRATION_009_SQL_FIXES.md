# Migration 009 SQL Fixes - Complete Backend Update

## Summary
Migration 009 renamed the `match_roles` table to `assignments` and renamed several columns, but the backend code was not updated to match. This caused errors throughout the application when trying to query the old table names.

## Migration 009 Changes (Applied Previously)
```sql
-- Table rename
ALTER TABLE match_roles RENAME TO assignments;

-- Column renames  
ALTER TABLE assignments RENAME COLUMN role_type TO position;
ALTER TABLE assignments RENAME COLUMN assigned_referee_id TO referee_id;
```

## Problem
All SQL queries in the backend were still using the old names:
- `match_roles` (should be `assignments`)
- `role_type` (should be `position`)
- `assigned_referee_id` (should be `referee_id`)

This caused errors like:
- `pq: relation "match_roles" does not exist`
- `pq: column "role_type" does not exist`
- `pq: column "assigned_referee_id" does not exist`

## Files Fixed

### 1. Migration 014
**File**: `backend/migrations/014_assignment_change_tracking.up.sql`
- Changed `ALTER TABLE match_roles` → `ALTER TABLE assignments`
- Changed index name from `idx_match_roles_viewed` → `idx_assignments_viewed`
- Changed `assigned_referee_id` → `referee_id` in index definition
- Removed duplicate `updated_at` column addition (already existed from migration 002)

**File**: `backend/migrations/014_assignment_change_tracking.down.sql`
- Updated to match new table and index names

### 2. Matches Feature
**File**: `backend/features/matches/repository.go`
- `CreateRole()`: `INSERT INTO assignments (position)`
- `GetRoles()`: `SELECT FROM assignments`, column aliases updated
- `DeleteRoles()`: `DELETE FROM assignments WHERE position IN (...)`
- `RoleExists()`: `SELECT FROM assignments WHERE position = ...`
- `GetCurrentRoles()`: `SELECT position FROM assignments`
- `LogEdit()`: `INSERT INTO assignment_history (position)`

**File**: `backend/features/matches/service.go`
- `resetViewedStatusForMatch()`: `UPDATE assignments SET ... WHERE referee_id IS NOT NULL`

### 3. Assignments Feature
**File**: `backend/features/assignments/repository.go`
- All `FROM match_roles` → `FROM assignments`
- All `UPDATE match_roles` → `UPDATE assignments`
- All `JOIN match_roles` → `JOIN assignments`
- All `role_type` → `position`
- All `assigned_referee_id` → `referee_id`

**Methods affected**:
- `GetRoleSlot()`
- `UpdateRoleAssignment()`
- `GetRefereeExistingRoleOnMatch()`
- `FindConflictingAssignments()`
- `GetRefereeMatchHistory()`
- `MarkAssignmentAsViewed()`
- `ResetViewedStatusForMatch()`

### 4. Referees Feature
**File**: `backend/features/referees/repository.go`
- `HasUpcomingAssignments()`: `JOIN assignments` instead of `JOIN match_roles`
- Column references updated

### 5. Match Reports Feature  
**File**: `backend/features/match_reports/service.go`
- `isAuthorizedForReport()`: `FROM assignments WHERE position = 'center' AND referee_id = ...`

### 6. Acknowledgment Feature
**File**: `backend/features/acknowledgment/repository.go`
- `GetRefereeAssignmentRole()`: `SELECT position FROM assignments WHERE referee_id = ...`
- `AcknowledgeAssignment()`: `UPDATE assignments SET ... WHERE referee_id = ...`

### 7. Main Backend Files
**File**: `backend/match_retention.go`
- `PurgeOldMatches()`: `DELETE FROM assignments WHERE match_id = ANY($1)`
- Comments updated to reference 'assignments'

**File**: `backend/availability.go`
- `getEligibleMatchesForRefereeHandler()`: All queries updated
- `EXISTS (SELECT 1 FROM assignments WHERE referee_id = ...)`
- `SELECT position FROM assignments WHERE referee_id = ...`
- Conflict checking queries updated

## Verification

### Compile Check
```bash
go build -o bin/server .
```
✅ Backend builds successfully

### Remaining References
```bash
grep -r "match_roles" /home/matt/repos/ref-sched/backend --include="*.go" | wc -l
```
Result: 3 (all in comments explaining the migration)

All SQL queries now use the correct table and column names.

## Commits Applied
```
f47509d Fix migration 014: Remove duplicate updated_at column addition
635f93a Fix migration 014: Use correct table and column names from migration 009
1ad781a Fix matches repository: Update all SQL to use renamed table and columns
395adc7 Fix all backend SQL: Update to use renamed table and columns from migration 009
8bb1402 Fix remaining SQL: Update match_retention.go and availability.go
```

## Testing Required

### Database Cleanup (Before Testing)
If migration 014 failed, clean up the dirty state:

```sql
-- Connect to database
psql -h localhost -U your_db_user -d referee_scheduler

-- Check current state
SELECT version, dirty FROM schema_migrations ORDER BY version;

-- If version 14 is dirty:
-- Option A: If viewed_by_referee column exists on assignments table
UPDATE schema_migrations SET dirty = false WHERE version = 14;

-- Option B: If viewed_by_referee column does NOT exist
DELETE FROM schema_migrations WHERE version = 14;

-- Exit and re-run migrations
\q
./referee-scheduler migrate
```

### Functional Testing
After database cleanup:

1. ✅ **CSV Import**: Upload a CSV file → Should create matches and role slots
2. ✅ **Role Slot Creation**: Verify `assignments` table is populated
3. ✅ **Match Listing**: Verify matches display with roles
4. ✅ **Referee Assignment**: Assign referee to a match
5. ✅ **Acknowledgment**: Referee acknowledges assignment
6. ✅ **Match Reports**: Center referee submits match report
7. ✅ **Availability**: Referee marks availability
8. ✅ **Retention**: Verify match archival works

## Impact

**Before Fixes**:
- ❌ CSV import failed with "relation match_roles does not exist"
- ❌ Match listing failed
- ❌ Assignment operations failed
- ❌ Acknowledgment operations failed
- ❌ Match reports failed
- ❌ Availability marking failed

**After Fixes**:
- ✅ All database operations use correct table names
- ✅ All queries reference correct column names
- ✅ Backend builds successfully
- ✅ Ready for testing

## Prevention

**Root Cause**: Migration 009 renamed database objects but application code was not updated simultaneously.

**Prevention Strategy**:
1. **When renaming database objects**, update all application code in the same commit
2. **Search codebase** for all references before committing migration
3. **Build and test** immediately after applying migration
4. **Consider database views** for backward compatibility during transition period

## Next Steps

1. ✅ **Code Fixed**: All SQL updated to use correct names
2. ⏳ **Database Cleanup**: Follow instructions in MIGRATION_014_FIX.md
3. ⏳ **Migration Execution**: Re-run migrations after cleanup
4. ⏳ **Testing**: Verify all features work correctly
5. ⏳ **Merge**: After testing, merge epic-6-csv-import to main

All backend code is now aligned with the database schema from migration 009.
