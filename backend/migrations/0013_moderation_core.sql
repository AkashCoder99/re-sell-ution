-- B20 PR1: moderation core schema + queue indexes.
--
-- Contract:
-- - target_type: listing | user | message
-- - status: open | in_review | resolved | rejected
-- - action_type: assign | note | hide_listing | warn_user | ban_user | close_report

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY,
    reporter_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type TEXT NOT NULL CHECK (target_type IN ('listing', 'user', 'message')),
    target_listing_id UUID REFERENCES listings(id) ON DELETE CASCADE,
    target_user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    target_message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
    reason_code TEXT,
    reason_text TEXT,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'in_review', 'resolved', 'rejected')),
    priority SMALLINT NOT NULL DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    assigned_admin_id UUID REFERENCES users(id) ON DELETE SET NULL,
    resolution_note TEXT,
    resolved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (
        ((target_listing_id IS NOT NULL)::int +
         (target_user_id IS NOT NULL)::int +
         (target_message_id IS NOT NULL)::int) = 1
    ),
    CHECK (
        (target_type = 'listing' AND target_listing_id IS NOT NULL AND target_user_id IS NULL AND target_message_id IS NULL) OR
        (target_type = 'user' AND target_user_id IS NOT NULL AND target_listing_id IS NULL AND target_message_id IS NULL) OR
        (target_type = 'message' AND target_message_id IS NOT NULL AND target_listing_id IS NULL AND target_user_id IS NULL)
    )
);

CREATE TABLE IF NOT EXISTS moderation_actions (
    id UUID PRIMARY KEY,
    report_id UUID NOT NULL REFERENCES reports(id) ON DELETE CASCADE,
    actor_user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    action_type TEXT NOT NULL CHECK (action_type IN ('assign', 'note', 'hide_listing', 'warn_user', 'ban_user', 'close_report')),
    action_payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_status_created_at
    ON reports (status, created_at DESC, id ASC);

CREATE INDEX IF NOT EXISTS idx_reports_target_listing_id
    ON reports (target_listing_id)
    WHERE target_listing_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reports_target_user_id
    ON reports (target_user_id)
    WHERE target_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reports_target_message_id
    ON reports (target_message_id)
    WHERE target_message_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reports_reporter_created_at
    ON reports (reporter_user_id, created_at DESC, id ASC);

CREATE INDEX IF NOT EXISTS idx_reports_assigned_admin_status
    ON reports (assigned_admin_id, status, created_at DESC, id ASC)
    WHERE assigned_admin_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_reports_open_queue
    ON reports (created_at ASC, priority ASC, id ASC)
    WHERE status IN ('open', 'in_review');

CREATE INDEX IF NOT EXISTS idx_moderation_actions_report_created_at
    ON moderation_actions (report_id, created_at DESC, id ASC);

CREATE INDEX IF NOT EXISTS idx_moderation_actions_actor_created_at
    ON moderation_actions (actor_user_id, created_at DESC, id ASC);
