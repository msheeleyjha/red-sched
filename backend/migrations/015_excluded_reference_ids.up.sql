-- Story 6.5: Create table for permanently excluded reference IDs
-- These are match reference IDs that should never be imported from CSV

CREATE TABLE excluded_reference_ids (
    id SERIAL PRIMARY KEY,
    reference_id VARCHAR(255) NOT NULL UNIQUE,
    reason TEXT,
    excluded_by INT REFERENCES users(id) ON DELETE SET NULL,
    excluded_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for fast lookup during CSV import
CREATE INDEX idx_excluded_reference_ids_reference_id ON excluded_reference_ids(reference_id);

-- Index for filtering by user
CREATE INDEX idx_excluded_reference_ids_excluded_by ON excluded_reference_ids(excluded_by);

COMMENT ON TABLE excluded_reference_ids IS 'Story 6.5: Permanently excluded match reference IDs that should not be imported';
COMMENT ON COLUMN excluded_reference_ids.reference_id IS 'The reference_id to exclude from future imports';
COMMENT ON COLUMN excluded_reference_ids.reason IS 'Optional reason why this reference_id is excluded (e.g., "recurring practice", "tournament cancelled")';
COMMENT ON COLUMN excluded_reference_ids.excluded_by IS 'User who added this exclusion';
COMMENT ON COLUMN excluded_reference_ids.excluded_at IS 'When this reference_id was excluded';
