-- Revert the fix for organization_modules.enabled_by

-- Drop the system admin column
ALTER TABLE organization_modules DROP COLUMN IF EXISTS enabled_by_admin;

-- Re-add the original foreign key constraint
ALTER TABLE organization_modules
ADD CONSTRAINT organization_modules_enabled_by_fkey
FOREIGN KEY (enabled_by) REFERENCES users(id);
