-- Workflow Engine Migration
-- Adds configurable workflow support for organizations

-- Message templates (must be created first due to FK reference)
CREATE TABLE message_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    channel VARCHAR(20) NOT NULL, -- 'whatsapp', 'email'
    subject VARCHAR(255), -- for email
    body TEXT NOT NULL,
    variables JSONB DEFAULT '[]'::jsonb, -- available variables: [{name: 'patient_name', description: '...'}]
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, name, channel)
);

-- Core workflow definition
CREATE TABLE workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    module VARCHAR(50) NOT NULL, -- 'appointments', 'construction'
    entity_type VARCHAR(50) NOT NULL, -- 'session', 'budget', 'project'
    is_active BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(organization_id, name)
);

-- Workflow states
CREATE TABLE workflow_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    name VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    state_type VARCHAR(20) NOT NULL, -- 'initial', 'intermediate', 'final'
    color VARCHAR(7), -- hex color for UI
    position INT NOT NULL, -- order in linear flow
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workflow_id, name),
    UNIQUE(workflow_id, position)
);

-- State transitions
CREATE TABLE workflow_transitions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    from_state_id UUID NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    to_state_id UUID NOT NULL REFERENCES workflow_states(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    requires_confirmation BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(workflow_id, from_state_id, to_state_id)
);

-- Triggers (when to execute actions)
CREATE TABLE workflow_triggers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    state_id UUID REFERENCES workflow_states(id) ON DELETE CASCADE,
    transition_id UUID REFERENCES workflow_transitions(id) ON DELETE CASCADE,
    trigger_type VARCHAR(30) NOT NULL, -- 'on_enter', 'on_exit', 'time_before', 'time_after', 'recurring'
    time_offset_minutes INT, -- for time-based triggers (negative = before, positive = after)
    time_field VARCHAR(50), -- field to base time on: 'scheduled_at', 'created_at'
    recurring_cron VARCHAR(100), -- for recurring triggers
    conditions JSONB, -- additional conditions
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CHECK (state_id IS NOT NULL OR transition_id IS NOT NULL)
);

-- Actions to execute
CREATE TABLE workflow_actions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_id UUID NOT NULL REFERENCES workflow_triggers(id) ON DELETE CASCADE,
    action_type VARCHAR(30) NOT NULL, -- 'send_whatsapp', 'send_email', 'update_field', 'create_task'
    action_order INT NOT NULL DEFAULT 0,
    template_id UUID REFERENCES message_templates(id) ON DELETE SET NULL,
    action_config JSONB DEFAULT '{}'::jsonb, -- action-specific configuration
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Scheduled jobs queue
CREATE TABLE scheduled_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    trigger_id UUID NOT NULL REFERENCES workflow_triggers(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    scheduled_for TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed', 'cancelled'
    attempts INT DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

-- Session payments tracking
CREATE TABLE session_payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    amount_cents INT NOT NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'unpaid', -- 'unpaid', 'partial', 'paid'
    payment_method VARCHAR(30), -- 'cash', 'transfer', 'insurance', 'card'
    insurance_provider VARCHAR(100),
    insurance_amount_cents INT,
    due_date DATE,
    paid_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(session_id)
);

-- Workflow execution log
CREATE TABLE workflow_execution_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    workflow_id UUID NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    trigger_id UUID REFERENCES workflow_triggers(id) ON DELETE SET NULL,
    action_id UUID REFERENCES workflow_actions(id) ON DELETE SET NULL,
    event_type VARCHAR(30) NOT NULL, -- 'state_change', 'trigger_fired', 'action_executed', 'action_failed'
    from_state VARCHAR(50),
    to_state VARCHAR(50),
    details JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_message_templates_org ON message_templates(organization_id);
CREATE INDEX idx_workflows_org ON workflows(organization_id);
CREATE INDEX idx_workflows_module ON workflows(organization_id, module);
CREATE INDEX idx_workflow_states_workflow ON workflow_states(workflow_id);
CREATE INDEX idx_workflow_transitions_workflow ON workflow_transitions(workflow_id);
CREATE INDEX idx_workflow_triggers_workflow ON workflow_triggers(workflow_id);
CREATE INDEX idx_workflow_triggers_state ON workflow_triggers(state_id);
CREATE INDEX idx_workflow_actions_trigger ON workflow_actions(trigger_id);
CREATE INDEX idx_scheduled_jobs_pending ON scheduled_jobs(scheduled_for) WHERE status = 'pending';
CREATE INDEX idx_scheduled_jobs_entity ON scheduled_jobs(entity_type, entity_id);
CREATE INDEX idx_scheduled_jobs_org ON scheduled_jobs(organization_id);
CREATE INDEX idx_session_payments_session ON session_payments(session_id);
CREATE INDEX idx_session_payments_status ON session_payments(payment_status);
CREATE INDEX idx_workflow_log_entity ON workflow_execution_log(entity_type, entity_id);
CREATE INDEX idx_workflow_log_workflow ON workflow_execution_log(workflow_id);
CREATE INDEX idx_workflow_log_org ON workflow_execution_log(organization_id);
