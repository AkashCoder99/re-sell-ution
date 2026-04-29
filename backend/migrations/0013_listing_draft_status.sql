-- Allow seller-owned draft listings that stay hidden from public discovery.

ALTER TABLE listings
    DROP CONSTRAINT IF EXISTS listings_status_check;

ALTER TABLE listings
    ADD CONSTRAINT listings_status_check
    CHECK (status IN ('active', 'reserved', 'sold', 'deleted', 'draft'));
