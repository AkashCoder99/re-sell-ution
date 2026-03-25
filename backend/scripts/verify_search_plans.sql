BEGIN;

INSERT INTO users (id, email, password_hash, full_name)
VALUES (
    '00000000-0000-0000-0000-000000999001',
    'search-plan@example.com',
    'verification-only',
    'Search Verification'
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO listings (
    id,
    seller_id,
    title,
    description,
    condition,
    price,
    currency,
    city,
    state,
    latitude,
    longitude,
    status,
    updated_by
)
SELECT
    ('00000000-0000-0000-0000-' || LPAD(g::text, 12, '0'))::uuid,
    '00000000-0000-0000-0000-000000999001',
    CASE
        WHEN g % 17 = 0 THEN 'mountain bike helmet ' || g
        WHEN g % 13 = 0 THEN 'iphone charger ' || g
        WHEN g % 11 = 0 THEN 'study desk ' || g
        ELSE 'marketplace listing ' || g
    END,
    CASE
        WHEN g % 17 = 0 THEN 'Protective bike helmet for daily rides and campus commutes'
        WHEN g % 13 = 0 THEN 'Fast charging adapter and cable bundle'
        WHEN g % 11 = 0 THEN 'Wooden desk with storage shelves'
        ELSE 'General purpose listing used to exercise search indexes'
    END,
    CASE
        WHEN g % 5 = 0 THEN 'like_new'
        WHEN g % 5 = 1 THEN 'good'
        WHEN g % 5 = 2 THEN 'fair'
        WHEN g % 5 = 3 THEN 'new'
        ELSE 'poor'
    END,
    10 + g,
    'INR',
    CASE
        WHEN g % 3 = 0 THEN 'Gainesville'
        WHEN g % 3 = 1 THEN 'Orlando'
        ELSE 'Tampa'
    END,
    'FL',
    CASE
        WHEN g % 3 = 0 THEN 29.6516 + ((g % 25) * 0.001)
        WHEN g % 3 = 1 THEN 28.5383 + ((g % 25) * 0.001)
        ELSE 27.9506 + ((g % 25) * 0.001)
    END,
    CASE
        WHEN g % 3 = 0 THEN -82.3248 + ((g % 25) * 0.001)
        WHEN g % 3 = 1 THEN -81.3792 + ((g % 25) * 0.001)
        ELSE -82.4572 + ((g % 25) * 0.001)
    END,
    'active',
    '00000000-0000-0000-0000-000000999001'
FROM generate_series(1, 3000) AS g;

ANALYZE listings;

\echo 'FTS only plan'
EXPLAIN (ANALYZE, COSTS OFF, BUFFERS)
SELECT l.id
FROM listings l
WHERE l.deleted_at IS NULL
  AND l.status = 'active'
  AND (
      setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
      setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
  ) @@ to_tsquery('simple', 'bike:* & helmet:*')
ORDER BY ts_rank_cd(
    (
        setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
    ),
    to_tsquery('simple', 'bike:* & helmet:*')
) DESC,
l.created_at DESC
LIMIT 20;

\echo 'FTS + city plan'
EXPLAIN (ANALYZE, COSTS OFF, BUFFERS)
SELECT l.id
FROM listings l
WHERE l.deleted_at IS NULL
  AND l.status = 'active'
  AND lower(l.city) = lower('Gainesville')
  AND (
      setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
      setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
  ) @@ to_tsquery('simple', 'bike:* & helmet:*')
ORDER BY ts_rank_cd(
    (
        setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
    ),
    to_tsquery('simple', 'bike:* & helmet:*')
) DESC,
l.created_at DESC
LIMIT 20;

\echo 'FTS + city + radius plan'
EXPLAIN (ANALYZE, COSTS OFF, BUFFERS)
SELECT l.id
FROM listings l
WHERE l.deleted_at IS NULL
  AND l.status = 'active'
  AND lower(l.city) = lower('Gainesville')
  AND l.latitude BETWEEN 29.6516 - (25.0 / 111.0) AND 29.6516 + (25.0 / 111.0)
  AND l.longitude BETWEEN -82.3248 - (25.0 / (111.0 * GREATEST(COS(RADIANS(29.6516)), 0.01)))
                        AND -82.3248 + (25.0 / (111.0 * GREATEST(COS(RADIANS(29.6516)), 0.01)))
  AND (
      6371.0 * ACOS(
          LEAST(
              1.0,
              GREATEST(
                  -1.0,
                  COS(RADIANS(29.6516)) * COS(RADIANS(l.latitude)) * COS(RADIANS(l.longitude) - RADIANS(-82.3248)) +
                  SIN(RADIANS(29.6516)) * SIN(RADIANS(l.latitude))
              )
          )
      )
  ) <= 25.0
  AND (
      setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
      setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
  ) @@ to_tsquery('simple', 'bike:* & helmet:*')
ORDER BY ts_rank_cd(
    (
        setweight(to_tsvector('simple', COALESCE(l.title, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(l.description, '')), 'B')
    ),
    to_tsquery('simple', 'bike:* & helmet:*')
) DESC,
l.created_at DESC
LIMIT 20;

ROLLBACK;
