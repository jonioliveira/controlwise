-- Available modules table (system-level module definitions)
CREATE TABLE available_modules (
    name VARCHAR(50) PRIMARY KEY,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    dependencies JSONB DEFAULT '[]',  -- List of required modules e.g., ["notifications"]
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Organization modules table (per-organization module enablement)
CREATE TABLE organization_modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    module_name VARCHAR(50) NOT NULL REFERENCES available_modules(name),
    is_enabled BOOLEAN DEFAULT TRUE,
    config JSONB DEFAULT '{}',  -- Module-specific settings
    enabled_at TIMESTAMP,
    enabled_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(organization_id, module_name)
);

CREATE INDEX idx_organization_modules_org_id ON organization_modules(organization_id);
CREATE INDEX idx_organization_modules_module_name ON organization_modules(module_name);
CREATE INDEX idx_organization_modules_is_enabled ON organization_modules(is_enabled);

-- Trigger for updated_at
CREATE TRIGGER update_organization_modules_updated_at
    BEFORE UPDATE ON organization_modules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Seed available modules
INSERT INTO available_modules (name, display_name, description, icon, dependencies) VALUES
('construction', 'Gestao de Obras', 'Worksheets, orcamentos, projetos e pagamentos para gestao de obras de construcao', 'Building2', '[]'),
('appointments', 'Agendamentos', 'Calendario, sessoes e gestao de pacientes para terapeutas', 'Calendar', '[]'),
('notifications', 'Notificacoes', 'Notificacoes WhatsApp/SMS para lembretes e confirmacoes', 'Bell', '[]');

-- Function to enable default modules for new organizations
CREATE OR REPLACE FUNCTION enable_default_modules()
RETURNS TRIGGER AS $$
BEGIN
    -- Enable construction module by default for new organizations
    INSERT INTO organization_modules (organization_id, module_name, is_enabled, enabled_at)
    VALUES (NEW.id, 'construction', TRUE, CURRENT_TIMESTAMP);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-enable default modules when organization is created
CREATE TRIGGER trigger_enable_default_modules
    AFTER INSERT ON organizations
    FOR EACH ROW EXECUTE FUNCTION enable_default_modules();
