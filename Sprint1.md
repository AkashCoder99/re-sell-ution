# Sprint 1 Status Report

## Scope Reviewed
This report covers only the stories listed for:
- EPIC-01: Auth & User Profile
- EPIC-02: Database Architecture & Standards
- EPIC-03: Listings & Media

Status legend used in this report:
- Completed
- Partial
- Not Completed

## EPIC-01 - Auth & User Profile

### F1 (P0) - Sign up UI
Status: Completed ✅
- Signup form and validation implemented.
- API integration and error handling implemented.
- Post-signup flow implemented (city selection/profile flow).

### F2 (P0) - Login UI
Status: Completed ✅
- Login form implemented.
- Session persistence implemented using localStorage token.
- Guard behavior implemented in app state flow for authenticated views.

### F3 (P0) - Forgot/Reset password UI
Status: Completed ✅
- Request reset UI implemented.
- Reset UI implemented (OTP + new password).
- Success/failure handling implemented.
- Backend endpoints exist.
- SMTP delivery config is not set yet, so OTP mail sending is not fully operational in deployment environments.

### F4 (P0) - Logout UI
Status: Completed ✅
- Logout action implemented in profile actions.
- Session clear + redirect behavior implemented.
- Expired/invalid token handling implemented through session restore failure path.

### F5 (P0) - Choose City / Location UI
Status: Partial
- City picker and persistence implemented.
- GPS-based suggestion not implemented.
- Feed/search context update based on city not fully realized because full feed/search flow is not delivered in Sprint 1.

### F6 (P1) - Basic profile UI
Status: Completed ✅
- Profile view implemented.
- Profile edit implemented (name, bio, city, photo URL).
- Avatar file upload is not implemented; URL-based photo field is used.

### B1 (P0) - Auth service (signup/login/logout)
Status: Partial
- Signup endpoint implemented with password hashing.
- Login endpoint implemented with token issuance.
- Logout endpoint implemented.
- Refresh token invalidation architecture not implemented (current auth is stateless token flow).
- Global login brute-force/rate-limit middleware not implemented.

### B2 (P0) - Account recovery (reset/OTP)
Status: Completed ✅
- Reset request endpoint implemented.
- Reset confirm endpoint implemented.
- Cooldown, max attempts, and expiry enforcement implemented.
- SMTP config not completed yet; endpoint flow works, but production email delivery depends on SMTP setup.

### B4 (P0) - User profile APIs
Status: Completed ✅
- Get profile endpoint implemented.
- Update profile endpoint implemented for allowed fields.
- Input validation implemented (including city/profile field validation).

### B6 (P1) - Observability (logs + basic metrics)
Status: Partial
- Structured logging implemented.
- Basic request metrics implemented (`/metrics`).
- Correlation IDs not implemented.
- Dashboard baseline not implemented.

## EPIC-02 - Database Architecture & Standards

### B3 [DB] (P0) - Core schema + migrations
Status: Completed ✅
- Core entities implemented: users, listings, categories, listing_images, favorites.
- Chat entities implemented: conversations, messages.
- Notifications scaffold migration present.
- Migration files and category seed scripts present.

### B5 [DB] (P0) - Constraints + indexes (MVP)
Status: Completed ✅
- Unique constraints/indexes for email/phone implemented.
- Core indexes implemented (city/category/status/created patterns).
- Foreign keys and cascade behavior implemented in schema.
- Validation rules implemented using DB checks/enums (for example listing status and price constraints).

### B14 [DB] (P1) - Soft delete + audit fields
Status: Partial
- `deleted_at` added for key tables.
- `created_at/updated_at` fields are present.
- Partial query filtering for soft delete exists in implemented user flows.
- `updated_by` style audit attribution not implemented.

## EPIC-03 - Listings & Media

### F12 (P0) - Create listing (multi-step form)
Status: Partial
- Multi-step UI flow implemented (basic -> details -> photos -> review).
- Field validation implemented.
- Submit path exists in frontend API layer.
- In current backend, listing creation endpoints are not wired, so end-to-end backend integration is not complete.

### F13 (P0) - Upload photos UI
Status: Partial
- Multi-file upload UI implemented with progress, reorder, delete, and retry behavior.
- Current upload is mock-oriented in Sprint 1 flow.
- Backend media upload endpoint integration is not complete.

### F14 (P0) - My listings dashboard
Status: Partial
- Active/Sold/Draft tabs implemented.
- Cards, quick actions, empty states, and pagination implemented.
- End-to-end backend integration for listings APIs is not complete in Sprint 1.

### F20 (P1) - Mark as sold + buyer selection
Status: Partial
- Mark as sold action implemented in dashboard UI.
- Optional buyer field implemented.
- Chat-based buyer selection integration is not implemented.
- Full backend flow for sold-state synchronization is not complete in Sprint 1.

## Sprint 1 Summary
- Completed stories: F1, F2, F3, F4, F6, B2, B3, B4, B5
- Partial stories: F5, B1, B6, B14, F12, F13, F14, F20
- Not Completed stories: None from the listed set were fully skipped, but several are partial due to backend integration and infrastructure gaps.

## Why Some Stories Are Partial
- Sprint 1 prioritized authentication and database foundation first.
- Several EPIC-03 frontend flows were built ahead of full backend listing/media APIs.
- SMTP configuration is pending, so password reset email delivery is environment-dependent.
- Observability was delivered at basic level without correlation IDs/dashboard layer.
