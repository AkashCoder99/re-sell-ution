GitHub Link - https://github.com/AkashCoder99/re-sell-ution

# Sprint 4 Status Report

This report documents Sprint 4 progress for:
- remaining Sprint 3 work / newly discovered issues
- tests for new functionality
- updated API documentation summary

## Sprint 4 Work Completed

### 1) Remaining Sprint 3 / Newly Discovered Work

#### F15 — Start Chat from Listing
Status: Completed

Implemented:
- chat entry from listing details
- chat entry wiring from browse/search paths
- conversation open/send flows integrated in app state

Files:
- frontend/src/App.tsx
- frontend/src/components/ListingDetails.tsx
- frontend/src/components/BrowseListings.tsx
- frontend/src/components/SearchListings.tsx
- frontend/src/components/ChatThread.tsx
- frontend/src/types/chat.ts
- frontend/src/api/chat.ts

#### F16 — Inbox / Conversation List UX
Status: Completed

Implemented:
- inbox page with conversation list
- unread count badges
- conversation search input
- pagination controls
- open-thread from inbox

Files:
- frontend/src/components/ConversationInbox.tsx
- frontend/src/api/chat.ts
- frontend/src/styles.css
- frontend/src/App.tsx

#### F11 / Listings Integration Follow-up
Status: Completed

Implemented:
- listing details CTA wiring stabilization
- browse-to-chat flow update
- integration cleanup and validation

Files:
- frontend/src/components/ListingDetails.tsx
- frontend/src/components/BrowseListings.tsx
- frontend/src/App.tsx

#### Backend Follow-up: Favorites + Search Enhancements
Status: Completed

Implemented:
- favorites model + handler + route wiring
- favorites HTTP tests
- search filters/sort support extended
- chat model indexing migration for inbox query performance

Files:
- backend/internal/models/favorite.go
- backend/internal/handlers/favorites.go
- backend/internal/handlers/favorites_http_test.go
- backend/internal/handlers/search.go
- backend/internal/handlers/search_test.go
- backend/internal/models/listing_search.go
- backend/internal/models/listing_search_test.go
- backend/migrations/0011_chat_model_indexes.sql
- backend/cmd/server/main.go
- backend/cmd/dbtool/main.go

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

## Sprint 4 Notes

- Sprint 4 closes key Sprint 3 carryover around chat entry/inbox and browse integration.
- Test coverage was expanded for both unit and e2e paths.
- Front-page README was updated with practical setup, usage, and test commands.
- Team members should continue committing work incrementally for contribution tracking.
