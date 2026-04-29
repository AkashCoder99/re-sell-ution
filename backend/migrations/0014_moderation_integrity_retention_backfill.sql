-- B20 PR2: integrity, retention/compliance fields, and legacy backfill.

-- Retention/compliance metadata for reports.
ALTER TABLE reports
    ADD COLUMN IF NOT EXISTS retain_until TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS purge_after TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS is_legal_hold BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS legal_hold_reason TEXT,
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

-- Basic temporal consistency checks.
ALTER TABLE reports
    DROP CONSTRAINT IF EXISTS reports_retention_temporal_consistency;
ALTER TABLE reports
    ADD CONSTRAINT reports_retention_temporal_consistency
    CHECK (
        purge_after IS NULL OR retain_until IS NULL OR purge_after >= retain_until
    );

ALTER TABLE reports
    DROP CONSTRAINT IF EXISTS reports_legal_hold_reason_consistency;
ALTER TABLE reports
    ADD CONSTRAINT reports_legal_hold_reason_consistency
    CHECK (
        (is_legal_hold = FALSE AND legal_hold_reason IS NULL) OR
        (is_legal_hold = TRUE)
    );

-- Integrity: suppress duplicate reports from same reporter on same target.
-- (Current policy keeps one report per reporter+target pair.)
CREATE UNIQUE INDEX IF NOT EXISTS uq_reports_listing_reporter_target
    ON reports (reporter_user_id, target_listing_id)
    WHERE target_type = 'listing' AND target_listing_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_reports_user_reporter_target
    ON reports (reporter_user_id, target_user_id)
    WHERE target_type = 'user' AND target_user_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_reports_message_reporter_target
    ON reports (reporter_user_id, target_message_id)
    WHERE target_type = 'message' AND target_message_id IS NOT NULL;

-- Useful retention/compliance query indexes.
CREATE INDEX IF NOT EXISTS idx_reports_purge_after
    ON reports (purge_after)
    WHERE purge_after IS NOT NULL AND is_legal_hold = FALSE AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_reports_retain_until
    ON reports (retain_until)
    WHERE retain_until IS NOT NULL AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_reports_legal_hold
    ON reports (is_legal_hold, created_at DESC)
    WHERE is_legal_hold = TRUE;

-- Backfill legacy listing reports into unified reports table.
INSERT INTO reports (
    id,
    reporter_user_id,
    target_type,
    target_listing_id,
    reason_text,
    status,
    priority,
    created_at,
    updated_at
)
SELECT
    lr.id,
    lr.reporter_id,
    'listing',
    lr.listing_id,
    lr.reason,
    'open',
    3,
    lr.created_at,
    lr.created_at
FROM listing_reports lr
ON CONFLICT (id) DO NOTHING;

-- Compatibility convenience view while old listing_reports table still exists.
CREATE OR REPLACE VIEW v_listing_reports_from_reports AS
SELECT
    r.id,
    r.target_listing_id AS listing_id,
    r.reporter_user_id AS reporter_id,
    r.reason_text AS reason,
    r.created_at
FROM reports r
WHERE r.target_type = 'listing'
  AND r.target_listing_id IS NOT NULL;
