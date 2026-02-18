-- Idempotent: uses ON CONFLICT DO NOTHING (skips if slug exists)

INSERT INTO categories (id, name, slug, parent_id) VALUES
    (gen_random_uuid(), 'Electronics', 'electronics', NULL),
    (gen_random_uuid(), 'Furniture', 'furniture', NULL),
    (gen_random_uuid(), 'Clothing & Accessories', 'clothing-accessories', NULL),
    (gen_random_uuid(), 'Books & Media', 'books-media', NULL),
    (gen_random_uuid(), 'Vehicles', 'vehicles', NULL),
    (gen_random_uuid(), 'Home & Garden', 'home-garden', NULL),
    (gen_random_uuid(), 'Sports & Outdoors', 'sports-outdoors', NULL),
    (gen_random_uuid(), 'Toys & Games', 'toys-games', NULL),
    (gen_random_uuid(), 'Health & Beauty', 'health-beauty', NULL),
    (gen_random_uuid(), 'Other', 'other', NULL)
ON CONFLICT (slug) DO NOTHING;
