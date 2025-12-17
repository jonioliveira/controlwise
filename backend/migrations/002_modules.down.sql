-- Drop trigger for auto-enabling modules
DROP TRIGGER IF EXISTS trigger_enable_default_modules ON organizations;
DROP FUNCTION IF EXISTS enable_default_modules();

-- Drop trigger for updated_at
DROP TRIGGER IF EXISTS update_organization_modules_updated_at ON organization_modules;

-- Drop indexes
DROP INDEX IF EXISTS idx_organization_modules_is_enabled;
DROP INDEX IF EXISTS idx_organization_modules_module_name;
DROP INDEX IF EXISTS idx_organization_modules_org_id;

-- Drop tables (order matters due to foreign key)
DROP TABLE IF EXISTS organization_modules;
DROP TABLE IF EXISTS available_modules;
