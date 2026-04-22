-- Add explicit available/unavailable tracking to availability table
-- Previously: record exists = available, no record = unavailable/no preference
-- Now: record with available=true = available, available=false = unavailable, no record = no preference

ALTER TABLE availability ADD COLUMN available BOOLEAN NOT NULL DEFAULT true;

-- Create index for filtering by availability status
CREATE INDEX idx_availability_status ON availability(referee_id, available);

-- All existing records will default to available=true (they were implicitly available)
