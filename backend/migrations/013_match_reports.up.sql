-- Migration: Create match_reports table for referee match outcome reporting
-- Epic 5: Match Reporting by Referees
-- Story 5.1: Match Report Database Schema

CREATE TABLE IF NOT EXISTS match_reports (
    id SERIAL PRIMARY KEY,
    match_id INTEGER NOT NULL UNIQUE REFERENCES matches(id) ON DELETE CASCADE,
    submitted_by INTEGER NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    final_score_home INTEGER,
    final_score_away INTEGER,
    red_cards INTEGER DEFAULT 0,
    yellow_cards INTEGER DEFAULT 0,
    injuries TEXT,
    other_notes TEXT,
    submitted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_scores CHECK (final_score_home >= 0 AND final_score_away >= 0),
    CONSTRAINT valid_cards CHECK (red_cards >= 0 AND yellow_cards >= 0)
);

-- Index for finding reports by match
CREATE INDEX idx_match_reports_match_id ON match_reports(match_id);

-- Index for finding reports submitted by a specific user
CREATE INDEX idx_match_reports_submitted_by ON match_reports(submitted_by);

-- Index for finding recent reports
CREATE INDEX idx_match_reports_submitted_at ON match_reports(submitted_at DESC);

-- Comments for documentation
COMMENT ON TABLE match_reports IS 'Match outcome reports submitted by referees';
COMMENT ON COLUMN match_reports.match_id IS 'Foreign key to matches table (one-to-one relationship)';
COMMENT ON COLUMN match_reports.submitted_by IS 'User ID of referee who submitted the report';
COMMENT ON COLUMN match_reports.final_score_home IS 'Final score for home team';
COMMENT ON COLUMN match_reports.final_score_away IS 'Final score for away team';
COMMENT ON COLUMN match_reports.red_cards IS 'Number of red cards issued during match';
COMMENT ON COLUMN match_reports.yellow_cards IS 'Number of yellow cards issued during match';
COMMENT ON COLUMN match_reports.injuries IS 'Description of injuries that occurred';
COMMENT ON COLUMN match_reports.other_notes IS 'Additional notes from the referee';
