-- Add custom_name column to users table
-- This field is managed by admin directly in the database
-- When not empty, it overrides the LINE display name in LINEUP mode

ALTER TABLE users ADD COLUMN IF NOT EXISTS custom_name VARCHAR(100) DEFAULT '';

COMMENT ON COLUMN users.custom_name IS 'Admin-managed display name override. When not empty, this name is used instead of line_display_name in LINEUP displays.';
