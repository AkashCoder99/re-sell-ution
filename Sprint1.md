# Sprint 1 Status Report

## User Stories
- **F1 - Sign up UI**: As a new user, I want to sign up with validated inputs so that I can create an account securely.
- **F2 - Login UI**: As a returning user, I want to log in and stay signed in so that I can access my account without repeated authentication.
- **F3 - Forgot/Reset password UI**: As a user who forgot my password, I want to request and confirm a password reset so that I can regain access to my account.
- **F4 - Logout UI**: As a signed-in user, I want to log out and clear my session so that my account remains secure on shared devices.
- **F5 - Choose City / Location UI**: As a user, I want to select and save my city so that content and search results are relevant to my location.
- **F6 - Basic profile UI**: As a user, I want to view and edit my profile details so that my account information is accurate.
- **B1 - Auth service (signup/login/logout)**: As a backend system, I want secure signup, login, and logout endpoints so that user authentication is reliable and safe.
- **B2 - Account recovery (reset/OTP)**: As a backend system, I want OTP-based password recovery with cooldowns and expiry so that account recovery is secure and abuse-resistant.
- **B4 - User profile APIs**: As a client app, I want profile read and update APIs with validation so that profile data stays consistent and valid.
- **B6 - Observability (logs + basic metrics)**: As an engineering team, I want structured logs and request metrics so that I can monitor health and diagnose issues quickly.
- **B3 [DB] - Core schema + migrations**: As a backend system, I want core database schema and migrations so that application data is stored with clear structure and version control.
- **B5 [DB] - Constraints + indexes (MVP)**: As a backend system, I want database constraints and indexes so that data integrity is enforced and queries perform efficiently.
- **B14 [DB] - Soft delete + audit fields**: As an engineering team, I want soft deletes and audit fields so that we preserve history and trace data changes.
- **F12 - Create listing (multi-step form)**: As a seller, I want a multi-step listing form with validation so that I can publish complete and correct listings.
- **F13 - Upload photos UI**: As a seller, I want to upload, reorder, retry, and remove listing photos so that I can present items effectively.
- **F14 - My listings dashboard**: As a seller, I want a dashboard for active, sold, and draft listings so that I can manage my inventory efficiently.
- **F20 - Mark as sold + buyer selection**: As a seller, I want to mark a listing as sold and optionally link a buyer so that listing lifecycle and transaction context are accurate.

## Scope Reviewed
This report covers only the stories listed for:
- PART-01: Auth & User Profile
- PART-02: Database Architecture & Standards
- PART-03: Listings & Media

## PART-01 - Auth & User Profile

### F1 (P0) - Sign up UI
Status: Completed ✅
- Signup form and validation implemented.
- API integration and error handling implemented.
- Post-signup flow implemented (city selection/profile flow).

### F2 (P0) - Login UI
Status: Completed ✅
- Login form implemented.
- Session persistence implemented using localStorage token.

### F3 (P0) - Forgot/Reset password UI
Status: Completed ✅
- Request reset UI implemented.
- Reset UI implemented (OTP + new password).
- Success/failure handling implemented.

### F4 (P0) - Logout UI
Status: Completed ✅
- Logout action implemented in profile actions.
- Session clear + redirect behavior implemented.
- Expired/invalid token handling implemented through session restore failure path.

### F5 (P0) - Choose City / Location UI
Status: Partial
- City picker and persistence implemented.
- Feed/search context update based on city not fully realized because full feed/search flow is not delivered in Sprint 1.

### F6 (P1) - Basic profile UI
Status: Completed ✅
- Profile view implemented.
- Profile edit implemented (name, bio, city, photo URL).
- URL-based photo field is used.

### B1 (P0) - Auth service (signup/login/logout)
Status: Completed ✅
- Signup endpoint implemented with password hashing.
- Login endpoint implemented with token issuance.
- Logout endpoint implemented.

### B2 (P0) - Account recovery (reset/OTP)
Status: Partial
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

## PART-02 - Database Architecture & Standards

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
Status: Completed ✅
- `deleted_at` added for key tables.
- `created_at/updated_at` fields are present.
- Partial query filtering for soft delete exists in implemented user flows.
- `updated_by` style audit attribution implemented.

## PART-03 - Listings & Media

### F12 (P0) - Create listing (multi-step form)
Status: Completed ✅
- Multi-step UI flow implemented (basic -> details -> photos -> review).
- Field validation implemented.
- Submit path exists in frontend API layer.

### F13 (P0) - Upload photos UI
Status: Completed ✅
- Multi-file upload UI implemented with progress, reorder, delete, and retry behavior.
- Current upload is mock-oriented in Sprint 1 flow.

### F14 (P0) - My listings dashboard
Status: Completed ✅
- Active/Sold/Draft tabs implemented.
- Cards, quick actions, empty states, and pagination implemented.

### F20 (P1) - Mark as sold + buyer selection
Status: Partial
- Mark as sold action implemented in dashboard UI.
- Optional buyer field implemented.
- Chat-based buyer selection integration is not implemented.
- Full backend flow for sold-state synchronization is not complete in Sprint 1.

## Sprint 1 Summary
- Completed stories: F1, F2, F3, F4, F6, B1, B3, B4, B5, B14, F12, F13, F14
- Partial stories: F5, B2, B6, F20
- Not Completed stories: None from the listed set were fully skipped, but few are partial due to integration and infrastructure gaps.

## Why Some Stories Are Partial
- Sprint 1 prioritized authentication and database foundation first.
- SMTP configuration is pending, so password reset email delivery is environment-dependent.
- Observability was delivered at basic level without dashboard layer.
