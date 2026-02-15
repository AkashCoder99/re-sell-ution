# ReSellution — EPIC 01 & EPIC 02 (MVP Foundation)


ReSellution is a full-stack local marketplace platform where users can list, discover, and buy/sell pre-owned items in their city.


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

---

## Sprint 1 Completion Summary

### User Stories Addressed

#### Frontend Team (Krishna Chaitanya Kolipakula, Tanmayee Maram)
**All frontend user stories completed:**
- ✅ **F1: Sign up** - Complete registration form with validation
- ✅ **F2: Login** - Login functionality with error handling
- ✅ **F3: Forgot/Reset password** - Password reset flow UI
- ✅ **F4: Logout** - Logout with session cleanup
- ✅ **F5: Choose City / Location** - City selector with popular cities dropdown and manual entry
- ✅ **F6: Basic Profile** - Profile view and edit with city, bio, and photo URL

**Additional Enhancements:**
- Password visibility toggle on login and register forms
- Mock API for frontend-only demos
- Modern UI with glassmorphism, animations, and gradient backgrounds
- Responsive design with proper accessibility (ARIA labels)

#### Backend Team (Akash Singh, Geetha Madhuri Papineni)
**Backend user stories completed:**
- ✅ **B1: Auth service** - Signup, login, logout endpoints implemented
- ✅ **B3: Core schema + migrations** - PostgreSQL schema with migrations
- ✅ **B4: User profile APIs** - GET and PATCH /users/me endpoints

### Issues Successfully Completed

#### Frontend Issues:
1. ✅ User registration form with validation
2. ✅ User login form with error handling
3. ✅ Forgot password UI flow
4. ✅ City selection component (dropdown + manual)
5. ✅ Profile display with user details
6. ✅ Profile edit form (name, city, bio, photo)
7. ✅ Password visibility toggle feature
8. ✅ Mock API for independent testing
9. ✅ Modern UI with animations and gradients
10. ✅ Responsive design and accessibility

#### Backend Issues:
1. ✅ POST /api/v1/auth/register endpoint
2. ✅ POST /api/v1/auth/login endpoint  
3. ✅ GET /api/v1/auth/me endpoint
4. ✅ Database migrations for users table
5. ✅ Password hashing with bcrypt
6. ✅ JWT token generation and validation
7. ✅ CORS configuration for frontend

### Issues Not Completed & Why

**Frontend:**
- ❌ **Password reset backend integration** - Backend endpoint not yet implemented (B2)
- The forgot password UI is complete and shows success message using mock API
- Will be fully functional once backend implements the reset endpoints

**Backend:**
- ❌ **B2: Account recovery (password reset/OTP)** - Deprioritized for Sprint 1
- ❌ **B5: Constraints + indexes** - Basic schema complete, advanced indexes deferred
- ❌ **B6: Observability** - Basic logging in place, metrics deferred
- ❌ **B14: Soft delete + audit fields** - Not critical for MVP auth flow

### Demo Access

**Frontend Demo:**
- URL: `http://localhost:5173` (or `http://127.0.0.1:5173`)
- Mode: Mock API enabled (no backend required)
- Test credentials: Any email/password (data stored in memory)

**Backend Demo:**
- URL: `http://localhost:8080`
- Database: PostgreSQL with migrations applied
- Test with: Postman or cURL

### Technologies Used

**Frontend:**
- React 18 with TypeScript
- Vite for build tooling
- CSS3 with animations
- Mock API for standalone demos

**Backend:**
- Go 1.26
- PostgreSQL 16
- JWT for authentication
- bcrypt for password hashing

