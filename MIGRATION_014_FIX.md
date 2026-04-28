# Migration 014 Fix - Database Cleanup Instructions

## Problem
Migration 014 failed for TWO reasons:

1. **Duplicate column**: Tried to add an `updated_at` column that already existed from migration 002
2. **Wrong table name**: Referenced `match_roles` table, but it was renamed to `assignments` in migration 009

This left the database in a "dirty" state at version 14.

## Root Causes

### Issue 1: Duplicate Column
**Migration 014 (before fix):**
```sql
ALTER TABLE match_roles
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;  -- ❌ Column already exists!
```

**Original table creation (migration 002):**
```sql
CREATE TABLE match_roles (
    id BIGSERIAL PRIMARY KEY,
    ...
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),  -- ✅ Already defined here
    ...
);
```

### Issue 2: Wrong Table/Column Names
**Migration 014 (before fix):**
```sql
ALTER TABLE match_roles  -- ❌ Table renamed to 'assignments' in migration 009
ADD COLUMN viewed_by_referee BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_match_roles_viewed ON match_roles(assigned_referee_id, viewed_by_referee);
--                                                   ^^^^^^^^^^^^^^^^^^^ ❌ Column renamed to 'referee_id' in migration 009
```

**Migration 009 renamed:**
- `match_roles` → `assignments`
- `role_type` → `position`
- `assigned_referee_id` → `referee_id`

## Fix Applied
The migration script has been corrected to:
1. Remove the duplicate `updated_at` column addition
2. Use the correct table name `assignments` (renamed in migration 009)
3. Use the correct column name `referee_id` (renamed in migration 009)

**Migration 014 (after fix):**
```sql
-- Note: Table was renamed from match_roles to assignments in migration 009
-- Note: updated_at already exists from the original table creation in migration 002
-- Note: assigned_referee_id was renamed to referee_id in migration 009
-- We only need to add the viewed_by_referee column

ALTER TABLE assignments
ADD COLUMN viewed_by_referee BOOLEAN NOT NULL DEFAULT FALSE;

-- Create index for efficient querying of unviewed assignments
CREATE INDEX idx_assignments_viewed ON assignments(referee_id, viewed_by_referee) 
WHERE viewed_by_referee = FALSE;
```

**Service layer also fixed:**
The `resetViewedStatusForMatch()` method in `service.go` was also updated to use the correct table and column names.

## Database Cleanup Steps

### Step 1: Connect to the database
```bash
# Using psql
psql -h localhost -U your_db_user -d referee_scheduler

# Or if using Docker
docker exec -it your_postgres_container psql -U your_db_user -d referee_scheduler
```

### Step 2: Check the current migration state
```sql
SELECT version, dirty FROM schema_migrations;
```

You should see:
```
 version | dirty 
---------+-------
      14 | true   <-- Dirty state!
```

### Step 3: Clean up the dirty state

**Option A: If migration 014 partially succeeded (viewed_by_referee column exists)**

Check if the column exists:
```sql
SELECT column_name 
FROM information_schema.columns 
WHERE table_name = 'assignments' 
  AND column_name = 'viewed_by_referee';
```

If it exists, mark migration as complete and clean:
```sql
UPDATE schema_migrations SET dirty = false WHERE version = 14;
```

Then you're done! The migration is complete.

**Option B: If migration 014 completely failed (viewed_by_referee column does NOT exist)**

Remove the failed migration record:
```sql
DELETE FROM schema_migrations WHERE version = 14;
```

Exit psql and re-run migrations:
```bash
# Exit psql
\q

# Re-run migrations (from backend directory)
cd /home/matt/repos/ref-sched/backend
./referee-scheduler migrate
```

The corrected migration 014 will now run successfully.

### Step 4: Verify the fix

Check that migrations completed successfully:
```sql
SELECT version, dirty FROM schema_migrations ORDER BY version;
```

You should see:
```
 version | dirty 
---------+-------
       1 | false
       2 | false
       ...
      14 | false  <-- Clean!
      15 | false  <-- If Epic 6 migrations ran
```

Verify the assignments table has the correct columns:
```sql
\d assignments
```

Expected columns:
- `id`
- `match_id`
- `position` (renamed from role_type in migration 009)
- `referee_id` (renamed from assigned_referee_id in migration 009)
- `created_at`
- `updated_at` (from migration 002)
- `viewed_by_referee` (from migration 014 - fixed)

And the index:
```sql
\di idx_assignments_viewed
```

Should show:
```
 Schema |          Name           | Type  |  Owner   |    Table     
--------+-------------------------+-------+----------+--------------
 public | idx_assignments_viewed  | index | your_user | assignments
```

### Step 5: Continue with remaining migrations

If you haven't run migration 015 yet (Epic 6):
```bash
cd /home/matt/repos/ref-sched/backend
./referee-scheduler migrate
```

This will run migration 015 (excluded_reference_ids table).

## Summary

**Migration fixed:** ✅  
**Database cleanup:** Manual steps required above  
**Next steps:** Re-run migrations after cleanup

## Prevention

This issue occurred because:
1. Migration 014 was created without checking the original table schema (duplicate `updated_at` column)
2. Migration 014 was created without checking migration history (table renamed in migration 009)
3. Service code was not updated when migration was created

**Lessons learned:** 
- Always check the complete migration history before creating new migrations
- Verify table and column names match the current schema (after all previous migrations)
- Update application code that references the same database objects

## Files Changed
- `backend/migrations/014_assignment_change_tracking.up.sql` - Removed duplicate updated_at, fixed table/column names
- `backend/migrations/014_assignment_change_tracking.down.sql` - Updated to use correct table/column names
- `backend/features/matches/service.go` - Fixed resetViewedStatusForMatch() to use correct table/column names

## Commits
```
f47509d Fix migration 014: Remove duplicate updated_at column addition
635f93a Fix migration 014: Use correct table and column names from migration 009
```
