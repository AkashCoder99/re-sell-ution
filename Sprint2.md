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
Status: Partial

Implemented
- Feed component with pagination and city/category filters.

Missing / Partial
- Not wired into the main app navigation.
- Default recency sort is not explicitly enforced in the FE.
- Depends on backend browse endpoint (not listed in server routes).

Workflow (component-level)
1. User chooses city/category filters.
2. Results load with pagination.

### F8 (P0) — Category Browsing UI
Status: Partial

Implemented
- Category dropdown filter within the browse component.

Missing / Partial
- No category grid/list landing page.
- No dedicated category listings page.
- Breadcrumbs not implemented.

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
Status: Partial

Implemented
- Mark sold action from the seller dashboard.
- Optional buyer input (manual ID).
- Dashboard state updates after marking sold.

Missing / Partial
- Buyer selection from chats not implemented.
- Feed sync/hide from public feed not implemented.

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

Auth
- POST `/api/v1/auth/register`
- POST `/api/v1/auth/login`
- POST `/api/v1/auth/logout`
- GET `/api/v1/auth/me`
- POST `/api/v1/auth/password/reset/request`
- POST `/api/v1/auth/password/reset/confirm`
- PATCH `/api/v1/users/me`
- PUT `/api/v1/users/me`
- DELETE `/api/v1/users/me`

Categories
- GET `/api/v1/categories`
- GET `/api/v1/categories/tree`

Listings
- GET `/api/v1/listings/search`
  - Query params: `q`, `page`, `limit`, `city`, `category_id`, `lat`, `lng`, `radius_km`
- POST `/api/v1/listings`
- GET `/api/v1/listings/me`
  - Query params: `status`, `page`, `limit`
- PATCH `/api/v1/listings/{id}`
- PATCH `/api/v1/listings/{id}/status`
- DELETE `/api/v1/listings/{id}`
- POST `/api/v1/listings/{id}/images`

Observability
- GET `/health`
- GET `/metrics`
- GET `/metrics/prometheus`
- GET `/metrics/dashboard`
