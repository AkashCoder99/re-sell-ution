-- Optional buyer reference when a listing is marked sold

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS sold_to_user_id UUID REFERENCES users(id) ON DELETE SET NULL;
