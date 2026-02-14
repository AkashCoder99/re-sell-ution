# ReSellution — EPIC 01 & EPIC 02 (MVP Foundation)


ReSellution is a full-stack local marketplace platform (OLX-style) where users can list, discover, and buy/sell pre-owned items in their city.


This README covers **only**:
- **EPIC-01: Auth & User Profile**
- **EPIC-02: Database Architecture & Standards**


---


## Scope


### EPIC-01 — Auth & User Profile
**Goal:** Users can sign up/login, manage sessions, set city, and maintain a basic public profile.


**Frontend stories**
- F1: Sign up
- F2: Login
- F3: Forgot/Reset password
- F4: Logout
- F5: Choose City / Location
- F6: Basic Profile


**Backend stories**
- B1: Auth service (signup/login/logout)
- B2: Account recovery (password reset / OTP)
- B4: User profile APIs
- B6: Observability (logs + basic metrics)


---


### EPIC-02 — Database Architecture & Standards
**Goal:** Establish clean schema, migrations, indexing, and auditing standards.


**Backend/DB stories**
- B3 [DB]: Core schema + migrations
- B5 [DB]: Constraints + indexes (MVP)
- B14 [DB]: Soft delete + audit fields


---


## High-Level Architecture (Foundation)


### Components
- **Frontend**
  - Auth screens (Signup/Login/Forgot)
  - City selection + persistence
  - Profile view/edit
- **Backend**
  - Auth endpoints (token-based auth)
  - Password reset/OTP flows
  - User profile endpoints
  - Request logging + metrics
- **Database**
  - Core schema for users and base marketplace entities (seeded early)
  - Indexing/constraints for correctness and performance
  - Soft-delete + audit fields to support recovery and traceability


---


## User Flows


### 1) Signup - City Select - Home
1. User signs up (email/phone + password).
2. On success, user selects city (manual or GPS suggestion).
3. City is saved and used as default browsing context.


### 2) Login - Continue Session
1. User logs in.
2. Session persists securely.
3. User lands on local home feed context (based on saved city).


### 3) Forgot Password - Reset
1. User requests reset via email/phone.
2. Receives time-bound token/OTP.
3. Sets new password and can log in.


### 4) Profile View/Edit
1. User opens profile page.
2. Can edit name/photo/bio/city (as allowed).
3. Saves changes and sees updated profile instantly.


---


## Acceptance Criteria (Epic-Level)


### EPIC-01 — Auth & User Profile
- Signup and login succeed with validated inputs.
- Sessions persist and can be invalidated (logout).
- Password reset flow is time-bound and rate-limited.
- City selection persists and is used by default.
- Profile can be fetched/updated with field-level validation.


### EPIC-02 — DB Architecture & Standards
- Migrations run cleanly from empty DB to latest schema.
- Uniqueness and referential integrity constraints are enforced.
- Core indexes exist to support common access patterns.
- Soft delete + audit fields exist (where applicable) and are respected in queries.


---


## API Surface (Foundation)


> Exact routes may vary by your framework; keep these as contract targets.


### Auth
- `POST /auth/signup`
- `POST /auth/login`
- `POST /auth/logout`
- `POST /auth/refresh` (optional)


### Password Recovery
- `POST /auth/password/reset/request`
- `POST /auth/password/reset/confirm`


### User Profile
- `GET /users/me`
- `PATCH /users/me`


### City / Location
- `PATCH /users/me/city` (or part of `PATCH /users/me`)


---


## Database Architecture (EPIC-02)


### Core Entities (initial)
Even though EPIC-01 focuses on Auth/Profile, the DB foundation should scaffold the core marketplace to avoid rework later.


**Minimum tables**
- `users`
- `cities` (or `locations`)
- `categories` 
- `listings` 
- `listing_images` 
- `favorites` 
- `conversations`, `messages`

### Constraints & Indexes (examples)
- Unique: `users.email`, `users.phone`
- Indexes:
  - `users.city_id`
  - `listings.city_id`, `listings.category_id`, `listings.created_at`
  - `categories.parent_id`
- Audit/soft delete:
  - `created_at`, `updated_at`, `deleted_at`
  - `updated_by` 


---

- Duplicate accounts (email/phone collisions) must return user-friendly errors.
- Token expiry and refresh edge cases (expired access token but valid refresh).
- Password reset abuse (rate limit + cooldown).
- City selection conflicts (manual vs GPS suggestion).
- Profile updates should block unsafe fields (role, auth flags, etc.).
