GitHub Link - https://github.com/AkashCoder99/re-sell-ution

# Sprint 4 Status Report

This report documents Sprint 4 progress for:
- remaining Sprint 3 work / newly discovered issues
- tests for new functionality
- updated API documentation summary

## Sprint 4 Work Completed

### 1) Remaining Sprint 3 / Newly Discovered Work

### Frontend Status

### F18 (P1) - Favorites / Saved items

**Status:** Completed (UI + integration polish)

**Implemented**

- **Saved page entry + listing cards:** Added `Saved` navigation in profile and a dedicated saved listings page with image/title/price/city cards.
- **Saved count badge:** Profile action now shows live saved count (`Saved (N)`), refreshed from favorites API.
- **View details from Saved:** Added `View details` button that opens `ListingDetails` modal directly from saved cards.
- **Immediate badge refresh after favorite changes:** Favorites badge now updates immediately after favorite/unfavorite from search and after remove from saved page.

**Workflow**

1. User favorites/unfavorites a listing from listing details in search.
2. App refreshes saved count immediately (no need to revisit profile later).
3. User opens Saved page, sees cards, can open full details or remove an item.

---

### F12 (P0) - Create listing (safety polish)

**Status:** Completed (unsaved-changes protection)

**Implemented**

- **Unsaved changes confirm on cancel/navigation:** Added confirmation dialog when leaving create listing flow with unsaved edits.
- **Browser refresh/close warning:** Added `beforeunload` guard so in-progress listing work is not accidentally lost.
- **Preserved cleanup behavior:** New-listing temporary records are still best-effort cleaned on cancel; edit mode does not delete existing listing.

**Workflow**

1. User starts create listing and makes edits.
2. On cancel/leave, user is prompted to confirm discard.
3. On tab refresh/close, browser warns about unsaved changes.


---

### F14 (P0) - My listings dashboard (Edit action wiring)

**Status:** Completed

**Implemented**

- **Edit action wired from dashboard:** `My Listings -> Edit` now opens `CreateListing` with selected listing context.
- **True edit mode in CreateListing:** Form prefilled with listing data; existing images are shown; submit updates the same listing (PATCH) instead of creating a new one.
- **Post-save flow:** Returns to `My Listings` and refreshes list with success messaging.

**Workflow**

1. Seller opens `My Listings` and clicks `Edit` on an active listing.
2. Edit form opens with existing values and media.
3. Seller saves changes and returns to refreshed dashboard state.


---

#### F15 — Start Chat from Listing

**Status:** Completed

**Implemented**

- chat entry from listing details
- chat entry wiring from browse/search paths
- conversation open/send flows integrated in app state

**Workflow**

1. User opens a listing from Browse or Search and taps Chat with Seller on listing details.
2. App checks authentication; if user is not logged in, it shows a message and redirects to login.
3. If logged in, app calls conversation create/open flow with buyer, seller, and listing context.
4. Existing conversation is reused for the same buyer-seller-listing combination; otherwise a new one is created.
5. App sets active conversation in state and navigates user to the chat thread screen.

---

#### F16 — Inbox / Conversation List UX

**Status:** Completed

**Implemented**
- inbox page with conversation list
- unread count badges
- conversation search input
- pagination controls
- open-thread from inbox

**Workflow**

1. User opens Inbox from profile buyer actions.
2. App fetches paginated conversation list and renders participant, listing title, preview, updated time, and unread count.
3. User can type in search input to filter conversation list by listing/user/message text.
4. User can move between pages with Previous/Next controls.
5. On selecting a conversation, app opens that thread and marks messages as read.

---

#### F11 / Listings Integration Follow-up

**Status:** Completed

**Implemented**
- listing details CTA wiring stabilization
- browse-to-chat flow update
- integration cleanup and validation

**Workflow**

1. User opens listing details from Search or Browse cards.
2. Listing details screen consistently renders CTA actions (chat/favorite/report) and listing metadata.
3. Chat CTA from details triggers shared app chat-start handler to keep behavior identical across entry points.
4. Browse-to-chat and Search-to-chat both route to the same thread state/navigation logic.
5. Integration cleanup ensures stable transitions between listing modal, profile, inbox, and chat screens.


---


#### Backend Follow-up: Favorites + Search Enhancements
Status: Completed

Implemented:
- favorites model + handler + route wiring
- favorites HTTP tests
- search filters/sort support extended
- chat model indexing migration for inbox query performance



---

### Backend Status

### B13 (P1) — Listing moderation flags (basic)

**Status:** Completed

**Depends on:** B7 (listings CRUD) — satisfied by existing listing create/update APIs.

**Implemented**

- **Validation rules (price bounds, required fields):** `ListingHandler` uses `validateListingCreate` / `validateListingPatch` for title/description length, allowed condition and create-status values, **price** between **0** and **999999999.99**, 3-letter currency (defaults `INR`), required **city**, optional state/coordinates bounds, optional category UUID validation against taxonomy.
- **Basic prohibited words filter (config-driven):** `LISTING_PROHIBITED_WORDS` (comma-separated, case-insensitive) loaded in config and applied to listing **create** and **patch** when title/description would contain a blocked term.
- **Rate limits for listing creation/edits:** IP-based limiter on `POST /api/v1/listings` and `PATCH /api/v1/listings/{id}` via `LISTING_WRITE_RATE_LIMIT_PER_IP` and `LISTING_WRITE_RATE_LIMIT_WINDOW_MINUTES`.

**Workflow**

1. Seller submits or edits a listing; handler runs schema validation first.
2. If configured prohibited terms appear in title/description, API returns **400** with a clear error.
3. If the client exceeds listing-write rate limits for its IP, API returns **429** with a retry message.

---

### B19 (P1) — Reporting & moderation workflow

**Status:** Completed

**Depends on:** B20 (schema + audit tables), B3 (listing details/report entry from product side).

**Implemented**

- **Create report endpoint:** `POST /api/v1/listings/{id}/report` (authenticated) creates or returns an existing row in the unified **`reports`** table (`ReportStore.CreateListingReport`), with dedupe by reporter + listing; blocks self-reports.
- **Status lifecycle (`open` / `in_review` / `resolved` / `rejected`):** Status stored on `reports`; transitions enforced in the model layer and updated via admin APIs; `resolved_at` / resolution note populated when closing.
- **Admin-only endpoints (resolve, action log):** Admin gate uses **`users.is_admin`** (`UserStore.IsAdminByID`; column added in **`0015_admin_users.sql`**). Routes: **`GET /api/v1/admin/reports`** (filter `status`, pagination), **`GET /api/v1/admin/reports/{id}`**, **`PATCH /api/v1/admin/reports/{id}`** (body `status`, `resolution_note`) for moderation decisions, **`POST /api/v1/admin/reports/{id}/actions`** (body `action_type`, `payload`) for audit entries including notes and other allowed action types.

**Workflow**

1. Buyer reports a listing from the client; API returns **201** with the persisted **`report`** payload.
2. Moderator with `is_admin` lists or opens a report, updates status or resolution note, or posts an audit action.
3. Actions and status changes are persisted for compliance review (`moderation_actions` and updated `reports` row).

---

### B20 [DB] (P1) — Moderation + compliance data

**Status:** Completed

**Depends on:** B3 (domain targets exist: listings, users, messages).

**Implemented**

- **`reports` table linking listing/user/message:** Migration **`0013_moderation_core.sql`** defines `reports` with `target_type` check (`listing` | `user` | `message`), exactly one target FK, status/priority/assignment/resolution columns, and queue indexes.
- **`moderation_actions` / audit table:** Same migration creates **`moderation_actions`** with `action_type` checks and JSON **`action_payload`**, indexed by report and actor/time.
- **Retention policy fields / timestamps:** Migration **`0014_moderation_integrity_retention_backfill.sql`** adds **`retain_until`**, **`purge_after`**, **`is_legal_hold`**, **`legal_hold_reason`**, **`deleted_at`** with temporal/legal-hold constraints; partial unique indexes per reporter+target; legacy **`listing_reports`** backfill into **`reports`** where applicable; convenience view **`v_listing_reports_from_reports`**.

**Workflow**

1. Application migrations (`dbtool migrate`) create moderation schema and extend it for retention.
2. Reporting and admin APIs read/write **`reports`** and **`moderation_actions`** as the system of record for moderation and audit.

---

## Frontend Unit and Cypress Tests

### Frontend Unit Tests (Vitest)
- frontend/src/__tests__/auth.mock.test.ts
- frontend/src/__tests__/BrowseListings.test.tsx
- frontend/src/__tests__/CitySelector.test.tsx
- frontend/src/__tests__/constants.test.ts
- frontend/src/__tests__/CreateListing.test.tsx
- frontend/src/__tests__/ConversationInbox.test.tsx
- frontend/src/__tests__/ListingDetails.test.tsx
- frontend/src/__tests__/logger.test.ts
- frontend/src/__tests__/PhotoUpload.test.tsx
- frontend/src/__tests__/SearchListings.test.tsx
- frontend/src/__tests__/validation.test.ts

### Cypress E2E Tests
- frontend/cypress/e2e/login.cy.ts
- frontend/cypress/e2e/chat-inbox.cy.ts

### Sprint 4 New/Updated Test Coverage Highlights
- Conversation inbox behavior (render, unread badges, pagination callback expectations)
- Listing details + chat CTA coverage
- Search/listing test stabilization for localStorage/runtime variations
- Cypress chat/inbox user journey in preview/mock mode
- Saved/Favorites flow coverage updates (saved list rendering, remove action behavior, view-details modal trigger path)
- Create listing flow test fix for async multi-step transitions (details -> photos -> review -> publish)
- Photo upload UI assertion update to match preview-first behavior (Upload 1 image render)
- Regression check after polish wiring (live saved-count refresh callback from favorite/unfavorite actions)

---

## Backend Unit Tests

- backend/cmd/server/main_test.go
- backend/internal/config/config_test.go
- backend/internal/handlers/auth_http_test.go
- backend/internal/handlers/auth_validation_test.go
- backend/internal/handlers/categories_http_test.go
- backend/internal/handlers/conversations_http_test.go
- backend/internal/handlers/favorites_http_test.go
- backend/internal/handlers/listings_http_test.go
- backend/internal/handlers/listings_mine_http_test.go
- backend/internal/handlers/listings_validation_test.go
- backend/internal/handlers/listing_reports_http_test.go
- backend/internal/handlers/notifications_http_test.go
- backend/internal/handlers/search_test.go
- backend/internal/middleware/auth_test.go
- backend/internal/models/category_test.go
- backend/internal/models/conversation_helpers_test.go
- backend/internal/models/listing_search_test.go
- backend/internal/models/moderation_test.go
- backend/internal/models/notification_test.go
- backend/internal/observability/middleware_test.go
- backend/internal/ratelimit/ratelimit_test.go
- backend/internal/utils/password_test.go
- backend/internal/utils/token_test.go

---

## Updated Backend API Documentation Summary

Runtime docs endpoints:
- GET /docs — Interactive OpenAPI explorer in the browser
- GET /openapi.yaml — Download OpenAPI spec as YAML file

### System
- GET / — JSON service banner and link shortcuts
- GET /health — Liveness probe for uptime checks
- GET /metrics — JSON snapshot of app counters and gauges
- GET /metrics/prometheus — Prometheus text-format scrape endpoint
- GET /metrics/dashboard — HTML page with live metric charts
- GET /uploads/… — Serves stored listing images by file path

### Auth and Users
- POST /api/v1/auth/register — Create account; returns JWT and user
- POST /api/v1/auth/login — Authenticate email/password; returns JWT
- POST /api/v1/auth/logout — Acknowledge logout; JWT discarded client-side
- POST /api/v1/auth/password/reset/request — Send or log OTP for reset
- POST /api/v1/auth/password/reset/confirm — Verify OTP; set new password
- GET /api/v1/auth/me — Return current user from Bearer token
- PATCH /api/v1/users/me — Partial update of profile fields
- PUT /api/v1/users/me — Full profile replace via same handler
- DELETE /api/v1/users/me — Soft-deactivate the logged-in account

### Categories
- GET /api/v1/categories — Flat list of marketplace categories
- GET /api/v1/categories/tree — Categories nested by parent and child

### Listings
- GET /api/v1/listings/browse — Paginated active listings by city or filters
- GET /api/v1/listings/search — Full-text search with filters and sort
- POST /api/v1/listings — Create listing; optional draft or active
- GET /api/v1/listings/me — Seller’s listings with status pagination
- PATCH /api/v1/listings/{id} — Edit fields on owned listing
- PATCH /api/v1/listings/{id}/status — Set active, sold, draft, etc.
- DELETE /api/v1/listings/{id} — Soft-delete listing owned by seller
- POST /api/v1/listings/{id}/images — Multipart or JSON image attach
- POST /api/v1/listings/{id}/report — File moderation report for listing

### Moderation (admin; requires `users.is_admin`)

- GET /api/v1/admin/reports — Paginated queue; optional status filter
- GET /api/v1/admin/reports/{id} — Load one report including targets
- PATCH /api/v1/admin/reports/{id} — Change status and resolution note
- POST /api/v1/admin/reports/{id}/actions — Record audit action or note

### Favorites
- GET /api/v1/favorites — List saved listings with pagination
- GET /api/v1/favorites/{listing_id} — Whether current user favorited listing
- PUT /api/v1/favorites/{listing_id} — Save listing to favorites
- DELETE /api/v1/favorites/{listing_id} — Remove listing from favorites

### Chat
- GET /api/v1/chat/conversations — Paginated inbox of buyer-seller threads
- POST /api/v1/chat/conversations — Open or reuse conversation for listing
- GET /api/v1/chat/conversations/{id} — Conversation detail and listing link
- GET /api/v1/chat/conversations/{id}/messages — Paginated message history
- POST /api/v1/chat/conversations/{id}/messages — Send message in thread
- PATCH /api/v1/chat/conversations/{id}/read — Mark messages read up to time

### Notifications
- GET /api/v1/notifications — List notifications for current user
- PATCH /api/v1/notifications/{id}/read — Mark single notification read
- PATCH /api/v1/notifications/read-all — Mark every notification read

---

## Commands for Sprint 4 Demo / Validation

Frontend app (mock mode):

```bash
cd frontend
VITE_USE_MOCK=true npm run dev -- --host 0.0.0.0 --port 5173 --strictPort
```

Frontend unit tests:

```bash
cd frontend
npm run test
```

Cypress run:

```bash
cd frontend
npm run cypress:run
```

Backend unit tests:

```bash
cd backend
go test ./...
```

---