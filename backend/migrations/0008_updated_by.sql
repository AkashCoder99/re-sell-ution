-- Audit: who last updated each record

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id);

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS updated_by UUID REFERENCES users(id);
