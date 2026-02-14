# ReSellution

## Project Description
ReSellution is a full-stack local marketplace platform (similar to OLX) where users can list, discover, and buy/sell pre-owned items in their city.

This first implementation slice includes:
- React + TypeScript frontend login/register UI
- Golang backend auth APIs
- PostgreSQL schema designed for a full marketplace (users, listings, categories, chats, favorites, sessions)

## Roles & Team Members
- Front-End Engineers
  - Krishna Chaitanya Kolipakula
  - Tanmayee Maram
- Back-End Engineers
  - Akash Singh
  - Geetha Madhuri Papineni

## Folder Structure
- `frontend/`: React + TypeScript + Vite app
- `backend/`: Go REST API
- `backend/migrations/0001_init.sql`: PostgreSQL schema

## Database Design (PostgreSQL)
The schema supports current auth plus future OLX-style modules.

Tables:
- `users`: auth + profile basics
- `categories`: listing taxonomy
- `listings`: item posts with condition/status/price/location
- `listing_images`: images for each listing
- `favorites`: saved listings
- `conversations`: buyer-seller conversation per listing
- `messages`: chat messages
- `sessions`: optional token/session tracking for revocation

## Local Setup (macOS)

### 1) Start PostgreSQL and create DB
If you use Homebrew PostgreSQL:

```bash
brew services start postgresql@18
createdb resellution
```

### 2) Backend setup

```bash
cd backend
cp .env.example .env
```

Set env vars in `backend/.env` (especially `TOKEN_SECRET`).

Run migration:

```bash
export DATABASE_URL="postgres://postgres:9999@localhost:5432/resellution?sslmode=disable"
make migrate
```

Run backend:

```bash
export PORT=8080
export DATABASE_URL="postgres://postgres:9999@localhost:5432/resellution?sslmode=disable"
export TOKEN_SECRET="replace-with-a-strong-random-secret"
export TOKEN_EXPIRY_HOURS=24
export CORS_ORIGIN="http://localhost:5173,http://127.0.0.1:5173"
make run
```

Backend URL: `http://localhost:8080`

### 3) Frontend setup

```bash
cd ../frontend
cp .env.example .env
npm install
npm run lint
npm run typecheck
npm run dev
```

Frontend URL: `http://localhost:5173`

## Auth APIs Implemented

### `POST /api/v1/auth/register`
Body:

```json
{
  "full_name": "Akash Singh",
  "email": "akash@example.com",
  "password": "strongpass123"
}
```

### `POST /api/v1/auth/login`
Body:

```json
{
  "email": "akash@example.com",
  "password": "strongpass123"
}
```

### `GET /api/v1/auth/me`
Header:

```text
Authorization: Bearer <token>
```

### `GET /health`
Basic health check endpoint.

## Notes
- ESLint is configured in `frontend/eslint.config.js` with TypeScript + React Hooks rules.
- This sandbox cannot reach external package registries, so `npm install` and `go mod tidy` could not be completed here.
- Run `go mod tidy` inside `backend/` on your machine before first backend run.
