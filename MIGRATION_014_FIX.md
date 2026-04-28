# Migration 014 Fix - Database Cleanup Instructions

## Problem
Migration 014 failed because it tried to add an `updated_at` column to the `match_roles` table, but that column already existed from the original table creation in migration 002.

This left the database in a "dirty" state at version 14.

## Root Cause
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

## Fix Applied
The migration script has been corrected to only add the `viewed_by_referee` column:

**Migration 014 (after fix):**
```sql
-- Note: updated_at already exists from migration 002_matches_schema.up.sql
-- We only need to add the viewed_by_referee column

ALTER TABLE match_roles
ADD COLUMN viewed_by_referee BOOLEAN NOT NULL DEFAULT FALSE;
```

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
WHERE table_name = 'match_roles' 
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

Verify the match_roles table has the correct columns:
```sql
\d match_roles
```

Expected columns:
- `id`
- `match_id`
- `role_type`
- `assigned_referee_id`
- `created_at`
- `updated_at` (from migration 002)
- `viewed_by_referee` (from migration 014 - fixed)

And the index:
```sql
\di idx_match_roles_viewed
```

Should show:
```
 Schema |          Name           | Type  |  Owner   |    Table    
--------+-------------------------+-------+----------+-------------
 public | idx_match_roles_viewed  | index | your_user | match_roles
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
1. Migration 014 was created without checking the original table schema
2. The `updated_at` column already existed from the initial table creation

**Lesson learned:** Always check the original table definition before adding columns in migrations.

## Files Changed
- `backend/migrations/014_assignment_change_tracking.up.sql` - Removed duplicate updated_at column addition
- `backend/migrations/014_assignment_change_tracking.down.sql` - Updated to only drop viewed_by_referee column

## Commit
```
f47509d Fix migration 014: Remove duplicate updated_at column addition
```
