CREATE UNIQUE INDEX IF NOT EXISTS idx_users_phone_unique ON users (phone) WHERE phone IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_users_city ON users (city);
CREATE INDEX IF NOT EXISTS idx_listings_city ON listings (city);
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (parent_id);
