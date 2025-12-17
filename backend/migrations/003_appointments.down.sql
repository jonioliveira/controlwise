-- Drop indexes
DROP INDEX IF EXISTS idx_session_history_changed_at;
DROP INDEX IF EXISTS idx_session_history_session_id;
DROP INDEX IF EXISTS idx_session_confirmations_sent_at;
DROP INDEX IF EXISTS idx_session_confirmations_type;
DROP INDEX IF EXISTS idx_session_confirmations_session_id;
DROP INDEX IF EXISTS idx_sessions_deleted_at;
DROP INDEX IF EXISTS idx_sessions_status;
DROP INDEX IF EXISTS idx_sessions_therapist_scheduled;
DROP INDEX IF EXISTS idx_sessions_org_scheduled;
DROP INDEX IF EXISTS idx_sessions_scheduled_at;
DROP INDEX IF EXISTS idx_sessions_patient_id;
DROP INDEX IF EXISTS idx_sessions_therapist_id;
DROP INDEX IF EXISTS idx_sessions_org_id;
DROP INDEX IF EXISTS idx_therapists_deleted_at;
DROP INDEX IF EXISTS idx_therapists_user_id;
DROP INDEX IF EXISTS idx_therapists_org_id;
DROP INDEX IF EXISTS idx_patients_deleted_at;
DROP INDEX IF EXISTS idx_patients_phone;
DROP INDEX IF EXISTS idx_patients_org_id;

-- Drop triggers
DROP TRIGGER IF EXISTS update_sessions_updated_at ON sessions;
DROP TRIGGER IF EXISTS update_therapists_updated_at ON therapists;
DROP TRIGGER IF EXISTS update_patients_updated_at ON patients;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS session_history;
DROP TABLE IF EXISTS session_confirmations;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS therapists;
DROP TABLE IF EXISTS patients;
