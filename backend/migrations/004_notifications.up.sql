-- Notification configuration per organization
CREATE TABLE notification_configs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE UNIQUE,
    -- WhatsApp/Twilio settings
    whatsapp_enabled BOOLEAN DEFAULT FALSE,
    twilio_account_sid VARCHAR(100),
    twilio_auth_token_encrypted VARCHAR(500),  -- Encrypted storage
    twilio_whatsapp_number VARCHAR(20),  -- Format: whatsapp:+351xxxxxxxxx
    -- Reminder settings
    reminder_24h_enabled BOOLEAN DEFAULT TRUE,
    reminder_2h_enabled BOOLEAN DEFAULT TRUE,
    -- Message templates (supports variables like {{patient_name}}, {{date}}, {{time}}, {{therapist}})
    reminder_24h_template TEXT DEFAULT 'Ola {{patient_name}}! Lembrete: tem uma consulta amanha, {{date}} as {{time}} com {{therapist}}. Responda SIM para confirmar ou NAO para cancelar.',
    reminder_2h_template TEXT DEFAULT 'Ola {{patient_name}}! A sua consulta e em 2 horas ({{time}}) com {{therapist}}. Esperamos por si!',
    confirmation_response_template TEXT DEFAULT 'Obrigado {{patient_name}}! A sua consulta foi {{status}}.',
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notification_configs_org_id ON notification_configs(organization_id);

-- Trigger for updated_at
CREATE TRIGGER update_notification_configs_updated_at
    BEFORE UPDATE ON notification_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- WhatsApp messages log (audit trail)
CREATE TABLE whatsapp_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    direction VARCHAR(10) NOT NULL CHECK (direction IN ('outbound', 'inbound')),
    phone_number VARCHAR(20) NOT NULL,
    message_content TEXT,
    message_sid VARCHAR(100),  -- Twilio message SID
    status VARCHAR(20) CHECK (status IN ('queued', 'sending', 'sent', 'delivered', 'read', 'failed', 'undelivered')),
    error_code VARCHAR(20),
    error_message TEXT,
    raw_payload JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_whatsapp_messages_org_id ON whatsapp_messages(organization_id);
CREATE INDEX idx_whatsapp_messages_session_id ON whatsapp_messages(session_id);
CREATE INDEX idx_whatsapp_messages_phone ON whatsapp_messages(phone_number);
CREATE INDEX idx_whatsapp_messages_sid ON whatsapp_messages(message_sid);
CREATE INDEX idx_whatsapp_messages_created_at ON whatsapp_messages(created_at);

-- Scheduled reminders table (for cron job processing)
CREATE TABLE scheduled_reminders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('reminder_24h', 'reminder_2h')),
    scheduled_for TIMESTAMP NOT NULL,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'sent', 'failed', 'skipped')),
    processed_at TIMESTAMP,
    error_message TEXT,
    whatsapp_message_id UUID REFERENCES whatsapp_messages(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, type)
);

CREATE INDEX idx_scheduled_reminders_status ON scheduled_reminders(status);
CREATE INDEX idx_scheduled_reminders_scheduled_for ON scheduled_reminders(scheduled_for);
CREATE INDEX idx_scheduled_reminders_pending ON scheduled_reminders(status, scheduled_for) WHERE status = 'pending';

-- Function to create scheduled reminders when a session is created/updated
CREATE OR REPLACE FUNCTION create_session_reminders()
RETURNS TRIGGER AS $$
BEGIN
    -- Only create reminders for pending or confirmed sessions
    IF NEW.status IN ('pending', 'confirmed') AND NEW.scheduled_at > NOW() THEN
        -- Check if organization has notifications enabled
        IF EXISTS (
            SELECT 1 FROM notification_configs nc
            JOIN organization_modules om ON om.organization_id = nc.organization_id
            WHERE nc.organization_id = NEW.organization_id
                AND nc.whatsapp_enabled = TRUE
                AND om.module_name = 'notifications'
                AND om.is_enabled = TRUE
        ) THEN
            -- Create 24h reminder (only if session is more than 24h away)
            IF NEW.scheduled_at > NOW() + interval '24 hours' THEN
                INSERT INTO scheduled_reminders (session_id, type, scheduled_for)
                VALUES (NEW.id, 'reminder_24h', NEW.scheduled_at - interval '24 hours')
                ON CONFLICT (session_id, type) DO UPDATE
                SET scheduled_for = NEW.scheduled_at - interval '24 hours',
                    status = 'pending',
                    processed_at = NULL,
                    error_message = NULL;
            END IF;

            -- Create 2h reminder (only if session is more than 2h away)
            IF NEW.scheduled_at > NOW() + interval '2 hours' THEN
                INSERT INTO scheduled_reminders (session_id, type, scheduled_for)
                VALUES (NEW.id, 'reminder_2h', NEW.scheduled_at - interval '2 hours')
                ON CONFLICT (session_id, type) DO UPDATE
                SET scheduled_for = NEW.scheduled_at - interval '2 hours',
                    status = 'pending',
                    processed_at = NULL,
                    error_message = NULL;
            END IF;
        END IF;
    ELSE
        -- Cancel pending reminders for cancelled/completed sessions
        UPDATE scheduled_reminders
        SET status = 'skipped'
        WHERE session_id = NEW.id AND status = 'pending';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to create/update reminders on session changes
CREATE TRIGGER trigger_create_session_reminders
    AFTER INSERT OR UPDATE OF scheduled_at, status ON sessions
    FOR EACH ROW EXECUTE FUNCTION create_session_reminders();
