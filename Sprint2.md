GitHub Link - https://github.com/AkashCoder99/re-sell-ution 

# Sprint 2 Status Report

This report summarizes what is currently implemented for the Sprint 2 scope below, including frontend and backend coverage, workflow notes, tests, and API documentation.

## Detail Work Completed in Sprint 2

## Frontend Status

### F5 (P0) — Choose City / Location UI
Status: Completed

Implemented
- City picker UI and persistence via profile update.
- City selection stored in user profile and reused by search.

Workflow
1. User selects a city in the city picker.
2. City is saved to the user profile.
3. Search uses the saved city as the default location context.

### F7 (P0) — Home Feed (Local Listings)
Status: COmpleted

Implemented
- Feed component with pagination and city/category filters.

Workflow (component-level)
1. User chooses city/category filters.
2. Results load with pagination.

### F8 (P0) — Category Browsing UI
Status: Completed

Implemented
- Category dropdown filter within the browse component.

Workflow (component-level)
1. User selects a category from the list.
2. Listings update to show category-matched items.

### F9 (P0) — Search UI
Status: Completed

Implemented
- Search bar + results page.
- Recent searches stored locally.
- Loading, empty, and error UX.
- Retry on error.
- Pagination of results.

Workflow
1. User submits a search term.
2. Results load with loading state.
3. Empty or error states show if needed.
4. Recent searches appear on focus when input is empty.

### F10 (P0) — Filters and Sort UI
Status: Completed

Implemented
- Filter modal with price range, category, condition, city.
- Sort options: newest, price low to high, price high to low.
- Applied filter chips + individual removal.
- Clear-all action.
- Filter persistence in localStorage.

Workflow
1. User opens Filters.
2. Sets price/category/condition/city and applies.
3. Results update and filters persist while browsing.

### F11 (P0) — Listing Details Page
Status: Completed (UI)

Implemented
- Image carousel with click and swipe.
- Title, price, condition, description, status.
- Seller card and location.
- CTAs: chat, favorite, report.
- Unavailable listing handling.
- Back navigation.

Workflow
1. User clicks a listing card.
2. Details modal opens with full listing information.
3. User can interact with carousel and CTAs.

### F20 (P1) — Mark as Sold + Buyer Selection
Status: Completed

Implemented
- Mark sold action from the seller dashboard.
- Optional buyer input (manual ID).
- Dashboard state updates after marking sold.

Workflow
1. Seller clicks Mark Sold.
2. Optional buyer ID entered.
3. Listing status updated and removed from active list.

## Backend Status

### B2 (P0) — Account Recovery (Reset/OTP)
Status: Completed

Implemented
- Reset request endpoint with OTP generation.
- Reset confirm endpoint.
- Throttling + expiry enforcement.

Notes
- SMTP delivery is optional and depends on configuration. Without SMTP, OTP is logged.

### B6 (P1) — Observability (Logs + Metrics)
Status: Completed

Implemented
- Structured logs with correlation IDs.
- Metrics endpoint and Prometheus endpoint.
- Dashboard baseline for metrics.

### B7 (P0) — Listings CRUD APIs
Status: Completed

Implemented
- Create listing, update listing, delete listing (soft).
- Change status (active/sold).
- List my listings with status filter.

### B8 (P0) — Image Upload Service
Status: Completed (optional scanning stub)

Implemented
- Multipart uploads with size/type validation.
- Persist listing_images metadata.
- Optional scanning hook placeholder.

### B10 (P1) — Full-text + Geo Strategy
Status: Completed

Implemented
- Full-text GIN index on title/description.
- City and geo indexes (lat/long).
- Search query uses TS query and geo distance.

### B11 (P0) — Categories Service
Status: Completed

Implemented
- Categories list and tree endpoints.
- Seeded category tree.

### B12 (P0) — Filters/Sort Backend Support
Status: Partial

Implemented
- Filters: city, category, geo radius in search query.
- Pagination in search.
- Default sort by relevance and recency.

Missing / Partial
- Price range filter.
- Condition filter.
- Sort by price (low/high).

## Frontend Tests

Unit tests
- `re-sell-ution/frontend/src/__tests__/validation.test.ts`
- `re-sell-ution/frontend/src/__tests__/logger.test.ts`
- `re-sell-ution/frontend/src/__tests__/constants.test.ts`
- `re-sell-ution/frontend/src/__tests__/CitySelector.test.tsx`
- `re-sell-ution/frontend/src/__tests__/BrowseListings.test.tsx`
- `re-sell-ution/frontend/src/__tests__/auth.mock.test.ts`
- `re-sell-ution/frontend/src/__tests__/SearchListings.test.tsx`
- `re-sell-ution/frontend/src/__tests__/ListingDetails.test.tsx`
- `re-sell-ution/frontend/src/__tests__/CreateListing.test.tsx`
- `re-sell-ution/frontend/src/__tests__/PhotoUpload.test.tsx` 


Cypress tests
- `re-sell-ution/frontend/cypress/e2e/login.cy.ts`

## Backend Unit Tests

- `re-sell-ution/backend/cmd/server/main_test.go`
- `re-sell-ution/backend/internal/config/config_test.go`
- `re-sell-ution/backend/internal/handlers/auth_http_test.go`
- `re-sell-ution/backend/internal/handlers/auth_validation_test.go`
- `re-sell-ution/backend/internal/handlers/listings_http_test.go`
- `re-sell-ution/backend/internal/handlers/listings_mine_http_test.go`
- `re-sell-ution/backend/internal/handlers/listings_validation_test.go`
- `re-sell-ution/backend/internal/handlers/search_test.go`
- `re-sell-ution/backend/internal/middleware/auth_test.go`
- `re-sell-ution/backend/internal/models/category_test.go`
- `re-sell-ution/backend/internal/observability/middleware_test.go`
- `re-sell-ution/backend/internal/ratelimit/ratelimit_test.go`
- `re-sell-ution/backend/internal/utils/password_test.go`
- `re-sell-ution/backend/internal/utils/token_test.go`

## Backend API Documentation (Current Routes)

## Base URL and versioning

- *Default listen address:* http://localhost:8080 (override with env PORT).
- *GET /* returns JSON service metadata (service, version, pointers to health and api). The API itself lives under */api/v1* (not at / alone).
- *API prefix:* /api/v1.

## Authentication

Protected routes expect a header:

http
Authorization: Bearer <JWT>


The JWT is an HMAC-SHA256 signed token issued by POST /api/v1/auth/register or POST /api/v1/auth/login. The server validates signature and expiry; there is no server-side session store (logout is client-side token disposal).

#Auth
- POST /api/v1/auth/register — Create account; returns JWT and user.
- POST /api/v1/auth/login — Email/password; returns JWT and user.
- POST /api/v1/auth/logout — Acknowledge logout (stateless JWT; Bearer required).
- GET /api/v1/auth/me — Current user profile (Bearer required).
- POST /api/v1/auth/password/reset/request — Request OTP (rate-limited; SMTP optional).
- POST /api/v1/auth/password/reset/confirm — Submit OTP and new password.
- PATCH /api/v1/users/me — Partial profile update (Bearer required).
- PUT /api/v1/users/me — Same as PATCH for profile (Bearer required).
- DELETE /api/v1/users/me — Soft-deactivate account (Bearer required).

#Categories
- GET /api/v1/categories — Flat category list.
- GET /api/v1/categories/tree — Categories nested by parent/child.

#Listings
- GET /api/v1/listings/search — Search active listings (full-text + filters).
  - Query params: q, page, limit, city, category_id, lat, lng, radius_km
- POST /api/v1/listings — Create listing (Bearer required).
- GET /api/v1/listings/me — Seller’s listings with pagination (Bearer required).
  - Query params: status, page, limit
- PATCH /api/v1/listings/{id} — Partial update owned listing (Bearer required).
- PATCH /api/v1/listings/{id}/status — Set status (e.g. active/reserved/sold) (Bearer required).
- DELETE /api/v1/listings/{id} — Soft-delete owned listing (Bearer required).
- POST /api/v1/listings/{id}/images — Add image via JSON image_url or multipart file (Bearer required).

#Observability
- GET /health — Liveness JSON ({"status":"ok"}).
- GET /metrics — JSON application metrics.
- GET /metrics/prometheus — Prometheus text exposition.
- GET /metrics/dashboard — HTML metrics dashboard.