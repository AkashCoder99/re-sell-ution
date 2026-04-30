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

---

## Backend Unit Tests

- backend/cmd/server/main_test.go
- backend/internal/config/config_test.go
- backend/internal/handlers/auth_http_test.go
- backend/internal/handlers/auth_validation_test.go
- backend/internal/handlers/favorites_http_test.go
- backend/internal/handlers/listings_http_test.go
- backend/internal/handlers/listings_mine_http_test.go
- backend/internal/handlers/listings_validation_test.go
- backend/internal/handlers/search_test.go
- backend/internal/middleware/auth_test.go
- backend/internal/models/category_test.go
- backend/internal/models/listing_search_test.go
- backend/internal/observability/middleware_test.go
- backend/internal/ratelimit/ratelimit_test.go
- backend/internal/utils/password_test.go
- backend/internal/utils/token_test.go

---

## Updated Backend API Documentation Summary

Authoritative source:
- backend/openapi/openapi.yaml

Runtime docs endpoints:
- GET /docs
- GET /openapi.yaml

### System
- GET /
- GET /health
- GET /metrics
- GET /metrics/prometheus
- GET /metrics/dashboard

### Auth and Users
- POST /api/v1/auth/register
- POST /api/v1/auth/login
- POST /api/v1/auth/logout
- POST /api/v1/auth/password/reset/request
- POST /api/v1/auth/password/reset/confirm
- GET /api/v1/auth/me
- PATCH /api/v1/users/me
- PUT /api/v1/users/me
- DELETE /api/v1/users/me

### Categories
- GET /api/v1/categories
- GET /api/v1/categories/tree

### Listings
- GET /api/v1/listings/browse
- GET /api/v1/listings/search
- POST /api/v1/listings
- GET /api/v1/listings/me
- PATCH /api/v1/listings/{id}
- PATCH /api/v1/listings/{id}/status
- DELETE /api/v1/listings/{id}
- POST /api/v1/listings/{id}/images

### Favorites
- GET /api/v1/favorites
- PUT /api/v1/favorites/{listing_id}
- DELETE /api/v1/favorites/{listing_id}

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