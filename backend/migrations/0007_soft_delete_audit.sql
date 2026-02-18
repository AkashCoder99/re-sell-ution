-- Soft delete + audit extensions for entities that can be hidden without hard delete.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Keep deleted status and deleted_at consistent going forward.
ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_status_deleted_at_consistency;

ALTER TABLE listings
    ADD CONSTRAINT listings_status_deleted_at_consistency
    CHECK (
        (status = 'deleted' AND deleted_at IS NOT NULL) OR
        (status <> 'deleted' AND deleted_at IS NULL)
    );

CREATE INDEX IF NOT EXISTS idx_users_active_email ON users (email) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
CREATE INDEX IF NOT EXISTS idx_listings_active_status_created ON listings (status, created_at DESC) WHERE deleted_at IS NULL;
