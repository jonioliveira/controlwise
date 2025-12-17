-- Revert: Restore patient's own name, email, phone columns

-- Add back the columns
ALTER TABLE patients ADD COLUMN name VARCHAR(200);
ALTER TABLE patients ADD COLUMN email VARCHAR(255);
ALTER TABLE patients ADD COLUMN phone VARCHAR(20);

-- Drop the client_id foreign key and index
DROP INDEX IF EXISTS idx_patients_client_id;
ALTER TABLE patients DROP COLUMN IF EXISTS client_id;

-- Recreate phone index
CREATE INDEX idx_patients_phone ON patients(phone);
