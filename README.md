# ReSellution

A modern local marketplace platform where users can buy and sell pre-owned items in their city.

## Project Description
ReSellution is a full-stack marketplace application similar to OLX, built with modern web technologies. The platform enables users to:
- Create accounts and manage profiles
- Select their city for localized listings
- List, discover, and purchase pre-owned items
- Chat with buyers/sellers (planned)
- Save favorite listings (planned)

## Tech Stack
- **Frontend**: React 18 + TypeScript + Vite
- **Backend**: Go 1.26 + PostgreSQL 16
- **Authentication**: JWT tokens + bcrypt password hashing
- **Styling**: Modern CSS with glassmorphism and animations

## Team Members
- Front-End Engineers
  - Krishna Chaitanya Kolipakula
  - Tanmayee Maram
- Back-End Engineers
  - Akash Singh
  - Geetha Madhuri Papineni

## Folder Structure
- `frontend/`: React + TypeScript + Vite app
- `backend/`: Go REST API
- `backend/migrations`: PostgreSQL schema

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

#### Demo Mode (Frontend Only)
To run the frontend without backend:
```bash
# Create .env.local
echo "VITE_USE_MOCK=true" > .env.local
npm run dev
```
This enables a mock API for testing frontend features independently.

## Notes
- ESLint is configured in `frontend/eslint.config.js` with TypeScript + React Hooks rules.
- This sandbox cannot reach external package registries, so `npm install` and `go mod tidy` could not be completed here.
- Run `go mod tidy` inside `backend/` on your machine before first backend run.
