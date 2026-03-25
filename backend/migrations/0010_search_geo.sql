-- Search and location support for public listing discovery.

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION;

ALTER TABLE listings
    ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION;

ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_latitude_valid;

ALTER TABLE listings
    ADD CONSTRAINT listings_latitude_valid
    CHECK (latitude IS NULL OR latitude BETWEEN -90 AND 90);

ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_longitude_valid;

ALTER TABLE listings
    ADD CONSTRAINT listings_longitude_valid
    CHECK (longitude IS NULL OR longitude BETWEEN -180 AND 180);

ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_coordinates_pair;

ALTER TABLE listings
    ADD CONSTRAINT listings_coordinates_pair
    CHECK (
        (latitude IS NULL AND longitude IS NULL) OR
        (latitude IS NOT NULL AND longitude IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS idx_listings_active_city_lower
    ON listings (lower(city))
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_listings_active_coordinates
    ON listings (latitude, longitude)
    WHERE deleted_at IS NULL AND latitude IS NOT NULL AND longitude IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_listings_search_document_active
    ON listings
    USING GIN (
        (
            setweight(to_tsvector('simple', COALESCE(title, '')), 'A') ||
            setweight(to_tsvector('simple', COALESCE(description, '')), 'B')
        )
    )
    WHERE deleted_at IS NULL;
