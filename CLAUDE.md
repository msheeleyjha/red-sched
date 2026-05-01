# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Referee Scheduler** is a web application for managing referee scheduling for club soccer associations. It uses a **Vertical Slice Architecture** where features are self-contained with all layers (models, repository, service, handler, routes) in one directory.

**Tech Stack**:
- Backend: Go 1.22 with Vertical Slice Architecture
- Frontend: SvelteKit with TypeScript
- Database: PostgreSQL 16
- Auth: Google OAuth2
- Deployment: Docker Compose

**Key Constraint**: All dates and times are stored and displayed in **US Eastern Time (America/New_York)** — no timezone conversion is applied since all matches occur locally in Eastern timezone.

---

## Build & Development Commands

### Docker Compose (Full Stack)
```bash
docker-compose up --build        # Start all services (db, backend, frontend)
docker-compose logs -f backend   # Watch backend logs (auto-recompiles)
docker-compose logs -f frontend  # Watch frontend logs (HMR enabled)
docker-compose ps                # Check service status
docker-compose down              # Stop all services
```

### Backend (Go)
```bash
cd backend

# Run all tests with coverage
go test ./... -v -cover

# Test specific feature
go test ./features/users -v
go test ./features/matches -v

# Test handlers and services only (main coverage areas)
go test ./features/... -v

# Build binary
go build -o main .

# Run single test
go test -run TestFunctionName ./features/path -v

# Database access (when containers running)
docker exec -it referee-scheduler-db psql -U referee_scheduler
```

### Frontend (SvelteKit)
```bash
cd frontend

# Check TypeScript and Svelte
npm run check

# Watch mode for development
npm run check:watch

# Build production bundle
npm run build

# Preview production build
npm run preview
```

### Database
```bash
# Connect to running database
docker exec -it referee-scheduler-db psql -U referee_scheduler

# Common queries
\dt              # List tables
\d users         # Describe table structure
SELECT * FROM users;
```

---

## Architecture

### Vertical Slice Pattern

The backend uses **Vertical Slice Architecture** where each feature is self-contained:

```
backend/features/[feature-name]/
├── handler.go      # HTTP request/response handling
├── service.go      # Business logic
├── repository.go   # Data access (SQL queries)
├── models.go       # Domain models
├── routes.go       # Route registration
└── [feature]_test.go  # Tests (aim for 100% handler/service coverage)
```

Current feature slices:
- `auth/` — Google OAuth2, session management, login/logout
- `users/` — User profiles, DOB, certifications
- `matches/` — Match management, CSV import from Stack Team App
- `assignments/` — Referee assignments, conflict detection
- `availability/` — Match & day-level referee availability
- `referees/` — Referee status, grade, role management
- `eligibility/` — Eligibility calculation (certifications, availability)
- `acknowledgment/` — Assignment acknowledgment workflow
- `match_reports/` — Match report generation/export

### Shared Infrastructure
- `shared/config/` — Environment variables, app configuration
- `shared/database/` — Connection pool, migrations runner
- `shared/middleware/` — Auth, RBAC, CORS, logging
- `shared/errors/` — Standard error handling

### Frontend Structure
- `src/routes/dashboard/` — Role-aware landing page
- `src/routes/assignor/` — Assignor UI (matches import, assignments, referee mgmt)
- `src/routes/referee/` — Referee UI (availability, acknowledgments, profile)
- `src/routes/+page.svelte` — Public landing page

---

## Database & Migrations

Migrations run automatically on backend startup. Located in `backend/migrations/`:

```bash
# Create new migration
touch backend/migrations/NNN_description.up.sql
touch backend/migrations/NNN_description.down.sql

# Restart backend for migrations to run
docker-compose restart backend
```

Current schema: Users, Matches, Assignments, Availability, Audit Logs, RBAC tables.

---

## Authentication & Authorization

### Google OAuth2 Flow
1. User clicks login → redirected to `/api/auth/google`
2. Google consent screen → redirected to `/api/auth/google/callback`
3. Backend exchanges code for token, creates/updates user session
4. Frontend redirected to role-appropriate dashboard

### RBAC System (Epic 1)
- **Roles**: `pending_referee`, `referee`, `assignor`
- **Permissions** (granular):
  - `can_assign_referees` — Match/assignment management
  - `can_assign_roles` — User role administration
  - `can_view_audit_logs` — Audit log access
- **Enforcement**: `rbac` middleware checks permission on protected routes

### Session Management
- Cookie-based sessions with Gorilla sessions
- Session secret in `SESSION_SECRET` env var (change in production)

---

## Testing Strategy

### Coverage Goals
- **Handler tests**: All HTTP endpoints (request validation, response codes, happy/sad paths)
- **Service tests**: Business logic, error conditions, conflict detection
- **Repository tests**: Query correctness (optional if complex SQL)
- **Shared tests**: Middleware, config, error handling (31 tests)
- **Feature tests**: 258 tests across all features

### Test Patterns
```go
// handler_test.go
func TestCreateHandler(t *testing.T) {
    // Setup: create test user, match, fixtures
    // Act: call handler with test request
    // Assert: check response status, body, side effects
}

// service_test.go
func TestServiceLogic(t *testing.T) {
    // Test business rules: eligibility, conflicts, availability
}
```

### Running Tests
```bash
go test ./features/... -v              # All feature tests
go test ./features/matches -v          # Single feature
go test -run TestHandlerName -v ./...  # Single test
go test ./... -cover                   # With coverage
```

---

## Common Development Tasks

### Adding a New Feature
1. Create `backend/features/myfeature/` directory
2. Implement `models.go` → `repository.go` → `service.go` → `handler.go` → `routes.go`
3. Register routes in `backend/main.go` (search for `router.HandleFunc`)
4. Write handler & service tests (100% coverage)
5. Create database migrations if needed
6. Add frontend routes under `src/routes/`

### Modifying Database Schema
1. Create migration files: `NNN_change.up.sql` and `.down.sql`
2. Write SQL in migration files (check other migrations for patterns)
3. Restart backend: `docker-compose restart backend`
4. Verify: `docker exec -it referee-scheduler-db psql -U referee_scheduler`

### Debugging
- Backend logs: `docker-compose logs -f backend`
- Frontend logs: Browser console + `docker-compose logs -f frontend`
- Database: `docker exec -it referee-scheduler-db psql -U referee_scheduler`
- Breakpoints: Use VS Code debugger (Delve for Go)

### CSV Import Flow (Matches)
- Stack Team App exports CSV with match data
- Handler: `POST /api/matches/import/parse` → previews parsed matches
- Handler: `POST /api/matches/import/confirm` → imports matches into database
- Location filtering available in UI (Epic 6)

---

## Code Patterns & Conventions

### Error Handling
```go
// Return wrapped errors with context
if err != nil {
    return fmt.Errorf("getUser: %w", err)
}

// Use shared error types in shared/errors/
if !user.IsActive() {
    return errors.ErrUnauthorized
}
```

### Repository Pattern
```go
// Repository interface (in models.go)
type UserRepository interface {
    FindByID(ctx context.Context, id int) (*User, error)
    Create(ctx context.Context, user *User) error
}

// Implementation (in repository.go)
func (r *repository) FindByID(ctx context.Context, id int) (*User, error) {
    // Query database
}
```

### Handler Pattern
```go
func (h *Handler) getHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Extract user/auth from context (set by middleware)
    // 2. Parse request body/query params
    // 3. Call service
    // 4. Handle errors
    // 5. Write JSON response
}
```

### Middleware Pattern
- Auth: Extracts user from session cookie, stores in context
- RBAC: Checks user permissions, blocks if missing permission
- CORS: Configured for `http://localhost:3000` (dev) and frontend domain (prod)

---

## Important Files & Locations

### Backend
- `backend/main.go` — Application entry point, route registration
- `backend/go.mod` — Dependency versions (Go 1.22)
- `backend/migrations/` — Database schema (auto-run on startup)
- `backend/shared/config/config.go` — Env var loading

### Frontend
- `frontend/package.json` — Dependencies (SvelteKit, TypeScript, Vite)
- `frontend/src/routes/` — Page components (role-based)
- `frontend/vite.config.js` — Vite build config, SvelteKit adapter
- `frontend/tsconfig.json` — TypeScript settings

### Configuration
- `.env` — Local dev secrets (Google OAuth, DB password)
- `.env.example` — Template for `.env`
- `.env.production.example` — Production config template
- `docker-compose.yml` — Service orchestration, port bindings

---

## API Endpoint Reference

All endpoints are documented in [API_REFERENCE.md](docs/architecture/API_REFERENCE.md).

Key groups:
- **Auth**: `/api/auth/*` (public)
- **Users**: `/api/profile`, `/api/admin/*` (authenticated)
- **Matches**: `/api/matches/*` (requires `can_assign_referees`)
- **Referees**: `/api/referees/*` (requires `can_assign_referees`)
- **Assignments**: `/api/matches/{id}/assign` (requires `can_assign_referees`)
- **Availability**: `/api/referee/matches/*` (for referees to mark availability)
- **Acknowledgment**: `/api/referee/matches/{id}/acknowledge` (for referees)

---

## Key Design Decisions

1. **Vertical Slice Architecture**: High cohesion, low coupling, parallel development
2. **Dependency Injection**: Via interfaces, testable without mocks
3. **Repository Pattern**: Data access isolated, easy to test
4. **Cookie Sessions**: Simpler than JWT for this use case, suitable for browser clients
5. **Eastern Time Only**: No timezone conversion — local sports club context
6. **CSV Import**: From Stack Team App, requires location filtering before confirmation

---

## Common Gotchas

1. **Timezone**: All dates are Eastern Time. Stack Team App exports are already in ET, don't convert.
2. **RBAC Permissions**: New endpoints need explicit permission checks in middleware or handler.
3. **Test Coverage**: Handlers and services should have 100% coverage. Use interface mocks for repositories if needed.
4. **Docker Volumes**: Backend code volume allows auto-recompile. Frontend has hot-reload. DB persists in `postgres_data` volume.
5. **Google OAuth**: Redirect URI must match exactly. Local dev uses `http://localhost:8080/api/auth/google/callback`.

---

## References

- [README.md](README.md) — Feature list, setup, troubleshooting
- [ARCHITECTURE.md](docs/architecture/ARCHITECTURE.md) — Detailed architecture documentation
- [API_REFERENCE.md](docs/architecture/API_REFERENCE.md) — All API endpoints
- [DEVELOPER_GUIDE.md](docs/guides/DEVELOPER_GUIDE.md) — Full developer onboarding
- [DOCS_INDEX.md](docs/DOCS_INDEX.md) — Documentation index
- [PROJECT_STATUS.md](docs/PROJECT_STATUS.md) — Current status & completion
