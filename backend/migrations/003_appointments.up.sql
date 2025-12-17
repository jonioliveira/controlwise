-- Patients table (separate from clients, for appointments module)
CREATE TABLE patients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20) NOT NULL,
    date_of_birth DATE,
    notes TEXT,
    emergency_contact VARCHAR(200),
    emergency_phone VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(organization_id, phone)
);

CREATE INDEX idx_patients_org_id ON patients(organization_id);
CREATE INDEX idx_patients_phone ON patients(phone);
CREATE INDEX idx_patients_deleted_at ON patients(deleted_at);

-- Trigger for updated_at
CREATE TRIGGER update_patients_updated_at
    BEFORE UPDATE ON patients
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Therapists table (can be linked to users or standalone)
CREATE TABLE therapists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id),
    name VARCHAR(200) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(20),
    specialty VARCHAR(100),
    working_hours JSONB DEFAULT '{}',  -- {"monday": {"start": "09:00", "end": "18:00"}, ...}
    session_duration_minutes INT DEFAULT 60,
    default_price_cents INT DEFAULT 0,
    timezone VARCHAR(50) DEFAULT 'Europe/Lisbon',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_therapists_org_id ON therapists(organization_id);
CREATE INDEX idx_therapists_user_id ON therapists(user_id);
CREATE INDEX idx_therapists_deleted_at ON therapists(deleted_at);

-- Trigger for updated_at
CREATE TRIGGER update_therapists_updated_at
    BEFORE UPDATE ON therapists
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Sessions table (appointments)
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    therapist_id UUID NOT NULL REFERENCES therapists(id),
    patient_id UUID NOT NULL REFERENCES patients(id),
    scheduled_at TIMESTAMP NOT NULL,
    duration_minutes INT NOT NULL DEFAULT 60,
    price_cents INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed', 'no_show')),
    session_type VARCHAR(50) DEFAULT 'regular',  -- regular, evaluation, follow_up
    notes TEXT,
    cancel_reason TEXT,
    cancelled_at TIMESTAMP,
    cancelled_by UUID REFERENCES users(id),
    completed_at TIMESTAMP,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_sessions_org_id ON sessions(organization_id);
CREATE INDEX idx_sessions_therapist_id ON sessions(therapist_id);
CREATE INDEX idx_sessions_patient_id ON sessions(patient_id);
CREATE INDEX idx_sessions_scheduled_at ON sessions(scheduled_at);
CREATE INDEX idx_sessions_org_scheduled ON sessions(organization_id, scheduled_at);
CREATE INDEX idx_sessions_therapist_scheduled ON sessions(therapist_id, scheduled_at);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_deleted_at ON sessions(deleted_at);

-- Trigger for updated_at
CREATE TRIGGER update_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Session confirmations table (tracks reminder and confirmation messages)
CREATE TABLE session_confirmations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL CHECK (type IN ('reminder_24h', 'reminder_2h', 'manual', 'followup')),
    channel VARCHAR(20) NOT NULL DEFAULT 'whatsapp' CHECK (channel IN ('whatsapp', 'sms', 'email')),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    read_at TIMESTAMP,
    responded_at TIMESTAMP,
    response VARCHAR(20) CHECK (response IN ('confirmed', 'cancelled', 'rescheduled', 'no_response')),
    message_id VARCHAR(100),  -- External provider message ID (Twilio SID)
    error_message TEXT,
    raw_response JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_session_confirmations_session_id ON session_confirmations(session_id);
CREATE INDEX idx_session_confirmations_type ON session_confirmations(type);
CREATE INDEX idx_session_confirmations_sent_at ON session_confirmations(sent_at);

-- Session history table (audit log for session changes)
CREATE TABLE session_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,  -- created, updated, confirmed, cancelled, completed, no_show
    old_values JSONB,
    new_values JSONB,
    changed_by UUID REFERENCES users(id),
    changed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_session_history_session_id ON session_history(session_id);
CREATE INDEX idx_session_history_changed_at ON session_history(changed_at);
