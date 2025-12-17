-- Link patient to client: Patient becomes a healthcare extension of Client
-- Every patient must be linked to a client, and name/email/phone come from the client

-- Add client_id foreign key to patients
ALTER TABLE patients ADD COLUMN client_id UUID REFERENCES clients(id) ON DELETE CASCADE;

-- Create index for efficient lookups
CREATE INDEX idx_patients_client_id ON patients(client_id);

-- Remove redundant columns (name, email, phone now come from client)
ALTER TABLE patients DROP COLUMN IF EXISTS name;
ALTER TABLE patients DROP COLUMN IF EXISTS email;
ALTER TABLE patients DROP COLUMN IF EXISTS phone;

-- Remove old phone uniqueness constraint (now handled by client)
DROP INDEX IF EXISTS idx_patients_phone;

-- Note: After data migration, make client_id NOT NULL:
-- ALTER TABLE patients ALTER COLUMN client_id SET NOT NULL;
