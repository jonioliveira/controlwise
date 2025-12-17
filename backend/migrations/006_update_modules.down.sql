-- Restore original module descriptions
UPDATE available_modules SET
    display_name = 'Gestao de Obras',
    description = 'Worksheets, orcamentos, projetos e pagamentos para gestao de obras de construcao'
WHERE name = 'construction';

UPDATE available_modules SET
    display_name = 'Agendamentos',
    description = 'Calendario, sessoes e gestao de pacientes para terapeutas'
WHERE name = 'appointments';

UPDATE available_modules SET
    display_name = 'Notificacoes',
    description = 'Notificacoes WhatsApp/SMS para lembretes e confirmacoes'
WHERE name = 'notifications';

-- Restore the auto-enable trigger
CREATE OR REPLACE FUNCTION enable_default_modules()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO organization_modules (organization_id, module_name, is_enabled, enabled_at)
    VALUES (NEW.id, 'construction', TRUE, CURRENT_TIMESTAMP);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_enable_default_modules
    AFTER INSERT ON organizations
    FOR EACH ROW EXECUTE FUNCTION enable_default_modules();
