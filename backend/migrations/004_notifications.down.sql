-- Drop trigger
DROP TRIGGER IF EXISTS trigger_create_session_reminders ON sessions;

-- Drop function
DROP FUNCTION IF EXISTS create_session_reminders();

-- Drop indexes
DROP INDEX IF EXISTS idx_scheduled_reminders_pending;
DROP INDEX IF EXISTS idx_scheduled_reminders_scheduled_for;
DROP INDEX IF EXISTS idx_scheduled_reminders_status;
DROP INDEX IF EXISTS idx_whatsapp_messages_created_at;
DROP INDEX IF EXISTS idx_whatsapp_messages_sid;
DROP INDEX IF EXISTS idx_whatsapp_messages_phone;
DROP INDEX IF EXISTS idx_whatsapp_messages_session_id;
DROP INDEX IF EXISTS idx_whatsapp_messages_org_id;
DROP INDEX IF EXISTS idx_notification_configs_org_id;

-- Drop trigger
DROP TRIGGER IF EXISTS update_notification_configs_updated_at ON notification_configs;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS scheduled_reminders;
DROP TABLE IF EXISTS whatsapp_messages;
DROP TABLE IF EXISTS notification_configs;
