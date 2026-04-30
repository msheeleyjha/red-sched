-- Story 6.5: Rollback excluded_reference_ids table

DROP INDEX IF EXISTS idx_excluded_reference_ids_excluded_by;
DROP INDEX IF EXISTS idx_excluded_reference_ids_reference_id;
DROP TABLE IF EXISTS excluded_reference_ids;
