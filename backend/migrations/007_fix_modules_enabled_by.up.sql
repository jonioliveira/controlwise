-- Fix organization_modules.enabled_by to allow NULL for system admin enablement
-- The enabled_by column references users(id), but system admins are in system_admins table

-- Drop the existing foreign key constraint
ALTER TABLE organization_modules DROP CONSTRAINT IF EXISTS organization_modules_enabled_by_fkey;

-- Add a new column for system admin enablement
ALTER TABLE organization_modules ADD COLUMN IF NOT EXISTS enabled_by_admin UUID REFERENCES system_admins(id);

-- Re-add the foreign key constraint but make it deferrable (optional, for data integrity)
-- We keep enabled_by for when regular users enable modules (org settings)
-- and enabled_by_admin for when system admins enable modules

-- Note: The application logic should:
-- - Use enabled_by when a user enables a module
-- - Use enabled_by_admin when a system admin enables a module
-- - Keep enabled_by as nullable when system admin enables
