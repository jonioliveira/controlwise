-- Reverse workflow engine migration

-- Drop indexes first
DROP INDEX IF EXISTS idx_workflow_log_org;
DROP INDEX IF EXISTS idx_workflow_log_workflow;
DROP INDEX IF EXISTS idx_workflow_log_entity;
DROP INDEX IF EXISTS idx_session_payments_status;
DROP INDEX IF EXISTS idx_session_payments_session;
DROP INDEX IF EXISTS idx_scheduled_jobs_org;
DROP INDEX IF EXISTS idx_scheduled_jobs_entity;
DROP INDEX IF EXISTS idx_scheduled_jobs_pending;
DROP INDEX IF EXISTS idx_workflow_actions_trigger;
DROP INDEX IF EXISTS idx_workflow_triggers_state;
DROP INDEX IF EXISTS idx_workflow_triggers_workflow;
DROP INDEX IF EXISTS idx_workflow_transitions_workflow;
DROP INDEX IF EXISTS idx_workflow_states_workflow;
DROP INDEX IF EXISTS idx_workflows_module;
DROP INDEX IF EXISTS idx_workflows_org;
DROP INDEX IF EXISTS idx_message_templates_org;

-- Drop tables in reverse order of creation (due to FK dependencies)
DROP TABLE IF EXISTS workflow_execution_log;
DROP TABLE IF EXISTS session_payments;
DROP TABLE IF EXISTS scheduled_jobs;
DROP TABLE IF EXISTS workflow_actions;
DROP TABLE IF EXISTS workflow_triggers;
DROP TABLE IF EXISTS workflow_transitions;
DROP TABLE IF EXISTS workflow_states;
DROP TABLE IF EXISTS workflows;
DROP TABLE IF EXISTS message_templates;
