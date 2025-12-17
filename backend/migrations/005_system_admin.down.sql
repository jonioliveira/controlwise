-- Rollback System Administrator Feature

-- Remove suspend tracking from users
ALTER TABLE users DROP COLUMN IF EXISTS suspend_reason;
ALTER TABLE users DROP COLUMN IF EXISTS suspended_by;
ALTER TABLE users DROP COLUMN IF EXISTS suspended_at;

-- Remove suspend tracking from organizations
ALTER TABLE organizations DROP COLUMN IF EXISTS suspend_reason;
ALTER TABLE organizations DROP COLUMN IF EXISTS suspended_by;
ALTER TABLE organizations DROP COLUMN IF EXISTS suspended_at;

-- Drop tables in reverse order (respecting foreign keys)
DROP TABLE IF EXISTS system_settings;
DROP TABLE IF EXISTS admin_impersonation_sessions;
DROP TABLE IF EXISTS system_admin_audit_logs;
DROP TABLE IF EXISTS system_admins;
