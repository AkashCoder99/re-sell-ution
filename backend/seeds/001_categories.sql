-- OLX-style taxonomy with parent-child relationships.

INSERT INTO categories (id, name, slug, parent_id) VALUES
    (gen_random_uuid(), 'Mobiles & Tablets', 'mobiles-tablets', NULL),
    (gen_random_uuid(), 'Electronics & Appliances', 'electronics-appliances', NULL),
    (gen_random_uuid(), 'Furniture', 'furniture', NULL),
    (gen_random_uuid(), 'Home & Garden', 'home-garden', NULL),
    (gen_random_uuid(), 'Fashion', 'fashion', NULL),
    (gen_random_uuid(), 'Vehicles', 'vehicles', NULL),
    (gen_random_uuid(), 'Books, Sports & Hobbies', 'books-sports-hobbies', NULL),
    (gen_random_uuid(), 'Kids', 'kids', NULL),
    (gen_random_uuid(), 'Pets', 'pets', NULL),
    (gen_random_uuid(), 'Services', 'services', NULL)
ON CONFLICT (slug) DO UPDATE
SET
    name = EXCLUDED.name,
    parent_id = EXCLUDED.parent_id;

INSERT INTO categories (id, name, slug, parent_id)
SELECT
    gen_random_uuid(),
    child.name,
    child.slug,
    parent.id
FROM (
    VALUES
        ('Smartphones', 'smartphones', 'mobiles-tablets'),
        ('Tablets', 'tablets', 'mobiles-tablets'),
        ('Mobile Accessories', 'mobile-accessories', 'mobiles-tablets'),
        ('Laptops & Computers', 'laptops-computers', 'electronics-appliances'),
        ('TV, Video & Audio', 'tv-video-audio', 'electronics-appliances'),
        ('Cameras & Accessories', 'cameras-accessories', 'electronics-appliances'),
        ('Home Appliances', 'home-appliances', 'electronics-appliances'),
        ('Sofas & Chairs', 'sofas-chairs', 'furniture'),
        ('Beds & Wardrobes', 'beds-wardrobes', 'furniture'),
        ('Tables & Dining', 'tables-dining', 'furniture'),
        ('Kitchen & Dining', 'kitchen-dining', 'home-garden'),
        ('Decor & Lighting', 'decor-lighting', 'home-garden'),
        ('Garden Tools', 'garden-tools', 'home-garden'),
        ('Men', 'fashion-men', 'fashion'),
        ('Women', 'fashion-women', 'fashion'),
        ('Footwear & Accessories', 'fashion-accessories', 'fashion'),
        ('Cars', 'cars', 'vehicles'),
        ('Motorcycles', 'motorcycles', 'vehicles'),
        ('Bicycles', 'bicycles', 'vehicles'),
        ('Books', 'books', 'books-sports-hobbies'),
        ('Sports Equipment', 'sports-equipment', 'books-sports-hobbies'),
        ('Music & Hobbies', 'music-hobbies', 'books-sports-hobbies'),
        ('Baby Gear', 'baby-gear', 'kids'),
        ('Toys & Games', 'toys-games', 'kids'),
        ('Kids Fashion', 'kids-fashion', 'kids'),
        ('Pet Supplies', 'pet-supplies', 'pets'),
        ('Pet Adoption', 'pet-adoption', 'pets'),
        ('Home Services', 'home-services', 'services'),
        ('Repairs', 'repairs', 'services'),
        ('Tuition', 'tuition', 'services')
) AS child(name, slug, parent_slug)
JOIN categories parent ON parent.slug = child.parent_slug
ON CONFLICT (slug) DO UPDATE
SET
    name = EXCLUDED.name,
    parent_id = EXCLUDED.parent_id;
