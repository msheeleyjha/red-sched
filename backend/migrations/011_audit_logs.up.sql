-- Create audit_logs table for tracking all data-modifying actions
CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action_type VARCHAR(10) NOT NULL CHECK (action_type IN ('create', 'update', 'delete')),
    entity_type VARCHAR(50) NOT NULL,
    entity_id INTEGER NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_entity_lookup ON audit_logs(entity_type, entity_id);

-- Add comment documenting retention policy (default: 2 years)
COMMENT ON TABLE audit_logs IS 'Audit log entries with 2-year retention policy. Entries older than retention period should be purged by scheduled job.';
COMMENT ON COLUMN audit_logs.action_type IS 'Type of action: create, update, or delete';
COMMENT ON COLUMN audit_logs.entity_type IS 'Type of entity modified (e.g., match, assignment, user, role)';
COMMENT ON COLUMN audit_logs.entity_id IS 'ID of the entity that was modified';
COMMENT ON COLUMN audit_logs.old_values IS 'JSON snapshot of entity state before change (NULL for create)';
COMMENT ON COLUMN audit_logs.new_values IS 'JSON snapshot of entity state after change (NULL for delete)';
COMMENT ON COLUMN audit_logs.ip_address IS 'IP address of user who performed the action';
COMMENT ON COLUMN audit_logs.created_at IS 'Timestamp when action occurred (used for retention policy)';
