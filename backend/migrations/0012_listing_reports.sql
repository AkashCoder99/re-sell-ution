-- Listing reports submitted from the buyer-facing listing details UI.

CREATE TABLE IF NOT EXISTS listing_reports (
    id UUID PRIMARY KEY,
    listing_id UUID NOT NULL REFERENCES listings(id) ON DELETE CASCADE,
    reporter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (listing_id, reporter_id)
);

CREATE INDEX IF NOT EXISTS idx_listing_reports_listing_id ON listing_reports (listing_id);
CREATE INDEX IF NOT EXISTS idx_listing_reports_reporter_id ON listing_reports (reporter_id);
CREATE INDEX IF NOT EXISTS idx_listing_reports_created_at ON listing_reports (created_at DESC);
