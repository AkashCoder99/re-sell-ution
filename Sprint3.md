GitHub Link - https://github.com/AkashCoder99/re-sell-ution

# Sprint 3 Status Report

This report summarizes the current implementation status for the Sprint 3 scope below, including frontend and backend coverage, workflow notes, tests, and API documentation.

---

## Frontend Status

### F11 (P0) — Listing details page

**Depends on:** B7, B8

**Status:** Completed

**Implemented**

- **Image carousel + details layout:** `ListingDetails` with prev/next controls, swipe-friendly touch handling, title, price, condition, description, status, relative listing age.
- **Seller card + location:** Seller-focused block and city/location presentation (uses listing + seller context from feed/search data).
- **CTA: chat / favorite / report:** Chat (via `onStartChat`), favorite toggle with loading/error, report with confirmation messaging; unavailable (`deleted`) listing handling and back navigation.

**Workflow**

1. User opens a listing from browse or search.
2. Details overlay shows images, metadata, seller/location, and CTAs.
3. Chat, favorite, and report invoke parent handlers (real API when not in mock mode).

---

### F15 (P0) — Start chat from listing

**Depends on:** B15, B7

**Status:** Completed

**Implemented**

- **“Chat” CTA on listing detail:** Wired through `ListingDetails` → `onStartChat(listing)`.
- **Auth gating (login prompt):** `App.handleStartChat` checks `isAuthenticated`; sets user-visible message, switches to login, throws so the listing UI can show an error state.
- **Create/open conversation flow:** Calls `getOrCreateConversation` with listing + participant fields, sets active conversation, refreshes inbox key, navigates to `chat` view.

**Workflow**

1. Logged-in user taps Chat → backend creates or returns conversation → thread opens.
2. Logged-out user taps Chat → message + redirect to login.

---

### F16 (P1) — Chat inbox UI

**Depends on:** B15, B16

**Status:** Completed

**Implemented**

- **Conversation list + unread badges:** `ConversationInbox` lists threads with preview, time, and `unread_count` badge when &gt; 0.
- **Search conversations (optional):** Search box filters via `q` / client `query` passed to `listConversations` (backend supports search on list).
- **Pagination:** Previous/Next with page label (`PAGE_SIZE` 6) and `total_pages` from API.

**Workflow**

1. User opens Inbox from profile navigation.
2. Conversations load with loading/error states; unread counts display when provided by API or mock.

---

### F17 (P1) — Chat thread UI

**Depends on:** B15

**Status:** Completed (minor gaps)

**Implemented**

- **Message composer + send:** `ChatThread` form with draft, submit, `Sending...` disabled state.
- **Message rendering + timestamps:** Bubbles for mine vs theirs; time via `toLocaleTimeString`.
- **Retry on failure:** Not implemented in `ChatThread` (failed send does not offer a dedicated retry button; user can resubmit text manually).

**Loading states:** Send button shows sending state; thread does not show a separate “loading history” skeleton (messages come from conversation payload).

---

## Backend Status

### B12 (P0) — Filters/sort backend support

**Depends on:** B5

**Status:** Completed

**Implemented**

- **Filter by price / category / condition / city:** `GET /api/v1/listings/search` query params: `city`, `category_id`, `condition`, `min_price`, `max_price`, plus existing geo (`lat`, `lng`, `radius_km`) and full-text `q`.
- **Sorting:** `sort` = `relevance` (default), `created_at_desc`, `created_at_asc`, `price_asc`, `price_desc`.
- **Pagination + stable ordering:** `page`, `limit` (capped); `ORDER BY` includes tie-breaker `l.id` for consistent pages.

---

### B15 (P0) — Conversations + messages APIs

**Depends on:** B16

**Status:** Completed

**Implemented**

- **Create / get conversation (listing + participants):** `POST /api/v1/chat/conversations`, `GET /api/v1/chat/conversations/{id}` (auth required; buyer/seller enforced in store).
- **Send message + validate participants:** `POST /api/v1/chat/conversations/{id}/messages` with body text; participant checks in model layer.
- **Fetch messages (paginated):** `GET /api/v1/chat/conversations/{id}/messages` with `page` / `limit`.
- **Mark read:** `PATCH /api/v1/chat/conversations/{id}/read`.
- **List inbox:** `GET /api/v1/chat/conversations` with pagination and optional search `q`.

---

### B16 [DB] (P1) — Chat data model

**Depends on:** B3

**Status:** Completed

**Implemented**

- **DB: `conversations`:** In `migrations/0001_init.sql` — `listing_id`, `buyer_id`, `seller_id`, `created_at`, `updated_at`, unique `(listing_id, buyer_id)`.
- **DB: `messages`:** `conversation_id`, `sender_id`, `body`, `is_read`, `created_at`.
- **DB: indexes for inbox / ordering:** `idx_messages_conversation_created`; `migrations/0011_chat_model_indexes.sql` adds `last_message_at` on `conversations`, backfill, and indexes on `(buyer_id, last_message_at DESC)` and `(seller_id, last_message_at DESC)`.

---

### B17 (P2) — Notifications service

**Depends on:** B15, B3

**Status:** Completed (push optional / not wired)

**Implemented**

- **Notify on new message:** On new message, model layer inserts a row into `notifications` for the recipient (see `internal/models/conversation.go`) and may prepare email delivery metadata.
- **Store notifications (read/unread):** `notifications` table (see `migrations/0006_notifications.sql`); `GET /api/v1/notifications`, `PATCH /api/v1/notifications/{id}/read`, `PATCH /api/v1/notifications/read-all`.
- **Push/email hooks (optional):** SMTP email for new message when configured; no mobile push integration in repo.

---

## Frontend tests (parity with backend test coverage style)

Unit / component tests (Vitest + Testing Library), similar in role to backend `*_test.go` files:

| Area | Test file |
|------|-----------|
| Validation helpers | `frontend/src/__tests__/validation.test.ts` |
| Logger | `frontend/src/__tests__/logger.test.ts` |
| Constants / API base | `frontend/src/__tests__/constants.test.ts` |
| Auth (mock) | `frontend/src/__tests__/auth.mock.test.ts` |
| City selector | `frontend/src/__tests__/CitySelector.test.tsx` |
| Browse listings | `frontend/src/__tests__/BrowseListings.test.tsx` |
| Search listings | `frontend/src/__tests__/SearchListings.test.tsx` |
| Listing details (F11) | `frontend/src/__tests__/ListingDetails.test.tsx` |
| Create listing | `frontend/src/__tests__/CreateListing.test.tsx` |
| Photo upload | `frontend/src/__tests__/PhotoUpload.test.tsx` |
| Conversation inbox (F16) | `frontend/src/__tests__/ConversationInbox.test.tsx` |


**E2E (Cypress)**

- `frontend/cypress/e2e/login.cy.ts` (existing).

---

## Backend unit tests (reference list)

- `backend/cmd/server/main_test.go`
- `backend/internal/config/config_test.go`
- `backend/internal/handlers/auth_http_test.go`
- `backend/internal/handlers/auth_validation_test.go`
- `backend/internal/handlers/listings_http_test.go`
- `backend/internal/handlers/listings_mine_http_test.go`
- `backend/internal/handlers/listings_validation_test.go`
- `backend/internal/handlers/search_test.go`
- `backend/internal/handlers/conversations_http_test.go`
- `backend/internal/handlers/favorites_http_test.go`
- `backend/internal/handlers/listing_reports_http_test.go`
- `backend/internal/handlers/notifications_http_test.go`
- `backend/internal/middleware/auth_test.go`
- `backend/internal/models/category_test.go`
- `backend/internal/models/listing_search_test.go`
- `backend/internal/observability/middleware_test.go`
- `backend/internal/ratelimit/ratelimit_test.go`
- `backend/internal/utils/password_test.go`
- `backend/internal/utils/token_test.go`

---

## Backend API Documentation (Current Routes)

**Base URL:** `http://localhost:8080` (override with `PORT`). **Prefix:** `/api/v1`.

**Chat (Bearer required)**

- `GET /api/v1/chat/conversations` — Inbox (`page`, `limit`, `q`).
- `POST /api/v1/chat/conversations` — Create or get by `listing_id`.
- `GET /api/v1/chat/conversations/{id}` — Conversation detail.
- `GET /api/v1/chat/conversations/{id}/messages` — Paginated messages.
- `POST /api/v1/chat/conversations/{id}/messages` — Send message.
- `PATCH /api/v1/chat/conversations/{id}/read` — Mark read.

**Search (public)**

- `GET /api/v1/listings/search` — See B12 for `q`, filters, `sort`, `page`, `limit`.

**Notifications (Bearer required)**

- `GET /api/v1/notifications` — List (`page`, `limit`, optional `unread`).
- `PATCH /api/v1/notifications/{id}/read` — Mark one read.
- `PATCH /api/v1/notifications/read-all` — Mark all read.
