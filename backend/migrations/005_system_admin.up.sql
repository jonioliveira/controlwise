-- System Administrator Feature
-- Platform-level administrators independent from tenant organizations

-- System administrators table
CREATE TABLE system_admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_system_admins_email ON system_admins(email);
CREATE INDEX idx_system_admins_deleted_at ON system_admins(deleted_at);

-- Audit log for admin actions
CREATE TABLE system_admin_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES system_admins(id),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID,
    details JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sysadmin_audit_logs_admin ON system_admin_audit_logs(admin_id);
CREATE INDEX idx_sysadmin_audit_logs_created ON system_admin_audit_logs(created_at);
CREATE INDEX idx_sysadmin_audit_logs_action ON system_admin_audit_logs(action);
CREATE INDEX idx_sysadmin_audit_logs_entity ON system_admin_audit_logs(entity_type, entity_id);

-- Impersonation session tracking
CREATE TABLE admin_impersonation_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES system_admins(id),
    impersonated_user_id UUID NOT NULL REFERENCES users(id),
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    reason TEXT NOT NULL,
    ip_address VARCHAR(45)
);

CREATE INDEX idx_impersonation_admin ON admin_impersonation_sessions(admin_id);
CREATE INDEX idx_impersonation_user ON admin_impersonation_sessions(impersonated_user_id);
CREATE INDEX idx_impersonation_active ON admin_impersonation_sessions(admin_id) WHERE ended_at IS NULL;

-- System settings (global configuration)
CREATE TABLE system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    description TEXT,
    updated_by UUID REFERENCES system_admins(id),
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add suspend tracking to organizations
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMP;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS suspended_by UUID REFERENCES system_admins(id);
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS suspend_reason TEXT;

-- Add suspend tracking to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS suspended_by UUID REFERENCES system_admins(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS suspend_reason TEXT;
