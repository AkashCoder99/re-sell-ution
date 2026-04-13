-- B16: chat data model refinements for inbox queries.

ALTER TABLE conversations
    ADD COLUMN IF NOT EXISTS last_message_at TIMESTAMPTZ;

UPDATE conversations c
SET last_message_at = COALESCE(
    (
        SELECT MAX(m.created_at)
        FROM messages m
        WHERE m.conversation_id = c.id
    ),
    c.updated_at,
    c.created_at
)
WHERE c.last_message_at IS NULL;

ALTER TABLE conversations
    ALTER COLUMN last_message_at SET DEFAULT NOW();

-- Keep compatibility for environments where there may be existing NULL rows.
UPDATE conversations
SET last_message_at = NOW()
WHERE last_message_at IS NULL;

ALTER TABLE conversations
    ALTER COLUMN last_message_at SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_conversations_buyer_last_message_at
    ON conversations (buyer_id, last_message_at DESC);

CREATE INDEX IF NOT EXISTS idx_conversations_seller_last_message_at
    ON conversations (seller_id, last_message_at DESC);
