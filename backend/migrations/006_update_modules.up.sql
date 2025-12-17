-- Update module descriptions to better reflect their purpose
UPDATE available_modules SET
    display_name = 'Construção',
    description = 'Folhas de obra, orçamentos, projetos, tarefas e pagamentos para empresas de construção'
WHERE name = 'construction';

UPDATE available_modules SET
    display_name = 'Saúde/Terapia',
    description = 'Agenda, sessões, pacientes e terapeutas para clínicas e profissionais de saúde'
WHERE name = 'appointments';

UPDATE available_modules SET
    display_name = 'Notificações',
    description = 'Notificações por WhatsApp e lembretes automáticos'
WHERE name = 'notifications';

-- Remove the auto-enable trigger - modules should be set by system admin
DROP TRIGGER IF EXISTS trigger_enable_default_modules ON organizations;
DROP FUNCTION IF EXISTS enable_default_modules();
