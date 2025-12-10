-- Migration: Add is_archived column to events table
-- Run this on existing PostgreSQL databases to add archive functionality

-- Add is_archived column if it doesn't exist
ALTER TABLE events ADD COLUMN IF NOT EXISTS is_archived BOOLEAN DEFAULT false;

-- Create index for archive filtering
CREATE INDEX IF NOT EXISTS idx_events_is_archived ON events(is_archived);

-- Verify the column was added
SELECT column_name, data_type, column_default 
FROM information_schema.columns 
WHERE table_name = 'events' AND column_name = 'is_archived';
