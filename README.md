# Referee Scheduler

A web application for managing referee scheduling for club soccer associations.

## Features

- Google OAuth2 authentication
- Role-based access (Assignor, Referee, Pending Referee)
- Match schedule management (CSV import from Stack Team App)
- Referee profile management with certification tracking
- Referee availability marking (per-match and full-day)
- Day-level unavailability tracking
- Assignment workflow with conflict detection
- Assignment acknowledgment by referees
- Overdue acknowledgment tracking (>24 hours)
- Mobile-responsive design

## Important Notes

### Timezone Handling

**All dates and times are stored and displayed in US Eastern Time (America/New_York).**

- Stack Team App CSV exports are in Eastern Time
- All match dates and times are treated as Eastern Time
- No timezone conversion is applied - this is appropriate for a local sports club where all matches occur in the Eastern timezone

## Tech Stack

- **Backend**: Go 1.22 with Vertical Slice Architecture
- **Frontend**: SvelteKit
- **Database**: PostgreSQL 16
- **Auth**: Google OAuth2
- **Deployment**: Docker
- **Testing**: Go testing framework (258 tests, 100% handler/service coverage)

## Architecture

This project uses **Vertical Slice Architecture** where each feature is organized as a self-contained slice with all its layers (models, repository, service, handler, routes, tests).

**Benefits**:
- High cohesion within features
- Low coupling between features
- Easy to locate and modify feature code
- Testable with clear boundaries
- Enables parallel development

**Key Patterns**:
- Dependency injection via interfaces
- Repository pattern for data access
- Service layer for business logic
- Handler layer for HTTP request/response
- Shared infrastructure (config, database, middleware, errors)

📘 **See [ARCHITECTURE.md](ARCHITECTURE.md) for complete architecture documentation**

## Prerequisites

- Docker and Docker Compose
- Google Cloud Project with OAuth2 credentials

## Google OAuth2 Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Google+ API"
   - Click "Enable"
4. Create OAuth2 credentials:
   - Navigate to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "OAuth client ID"
   - Choose "Web application"
   - Add authorized redirect URIs:
     - `http://localhost:8080/api/auth/google/callback` (for local development)
   - Click "Create"
5. Copy the Client ID and Client Secret

## Local Development Setup

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd referee-scheduler
   ```

2. Create a `.env` file from the example:
   ```bash
   cp .env.example .env
   ```

3. Edit `.env` and add your Google OAuth2 credentials:
   ```
   GOOGLE_CLIENT_ID=your-client-id-here.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret-here
   ```

4. Start the application with Docker Compose:
   ```bash
   docker-compose up --build
   ```

5. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health check: http://localhost:8080/health

## Database Migrations

Migrations run automatically when the backend starts. Migration files are located in `backend/migrations/`.

To create a new migration:
1. Create two files in `backend/migrations/`:
   - `XXX_description.up.sql` (for applying the migration)
   - `XXX_description.down.sql` (for rolling back)
2. Restart the backend container

## Creating an Assignor Account

By default, new users are created with the `pending_referee` role. To create an assignor account:

1. Sign in with Google to create your user account
2. Connect to the database:
   ```bash
   docker exec -it referee-scheduler-db psql -U referee_scheduler
   ```
3. Update your user role:
   ```sql
   UPDATE users SET role = 'assignor', status = 'active' WHERE email = 'your-email@example.com';
   ```
4. Exit the database:
   ```
   \q
   ```
5. Sign out and sign back in to see the assignor dashboard

## Project Structure

```
referee-scheduler/
├── backend/                      # Go backend
│   ├── main.go                  # Application entry point (307 lines)
│   ├── shared/                  # Shared infrastructure
│   │   ├── config/             # Configuration management
│   │   ├── database/           # Database connection & migrations
│   │   ├── errors/             # Standard error handling
│   │   ├── middleware/         # HTTP middleware (auth, RBAC, CORS)
│   │   └── utils/              # Shared utilities
│   ├── features/                # Feature slices (vertical architecture)
│   │   ├── users/              # User management & profiles
│   │   ├── matches/            # Match management & CSV import
│   │   ├── assignments/        # Referee assignments
│   │   ├── acknowledgment/     # Assignment acknowledgment
│   │   ├── referees/           # Referee management
│   │   ├── availability/       # Match & day availability
│   │   └── eligibility/        # Eligibility checking
│   ├── migrations/              # Database migrations
│   ├── Dockerfile               # Backend container config
│   └── go.mod                   # Go dependencies
├── frontend/                     # SvelteKit frontend
│   ├── src/
│   │   ├── routes/              # SvelteKit routes
│   │   ├── app.html             # HTML template
│   │   └── app.css              # Global styles
│   ├── Dockerfile               # Frontend container config
│   └── package.json             # Node dependencies
├── docker-compose.yml            # Docker orchestration
├── .env.example                  # Environment variables template
├── ARCHITECTURE.md               # Architecture documentation
└── README.md                     # This file
```

**Architecture**: This project uses **Vertical Slice Architecture** where each feature is self-contained with its own models, repository, service, handler, and tests. See [ARCHITECTURE.md](ARCHITECTURE.md) for details.

## Development

### Quick Start for Developers

1. **Backend Development**: Go automatically recompiles on changes
   ```bash
   docker-compose logs -f backend
   ```

2. **Frontend Development**: Hot module replacement enabled
   ```bash
   docker-compose logs -f frontend
   ```

3. **Run Tests**:
   ```bash
   cd backend
   go test ./features/...  # Feature tests (258 tests)
   go test ./shared/...    # Shared package tests (31 tests)
   ```

4. **Database Access**:
   ```bash
   docker exec -it referee-scheduler-db psql -U referee_scheduler
   ```

### Adding a New Feature

Follow the vertical slice pattern:

1. Create feature directory: `backend/features/myfeature/`
2. Add models, repository, service, handler, routes
3. Write tests (aim for 100% handler/service coverage)
4. Register routes in `main.go`
5. Update documentation

📘 **See [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md) for complete developer onboarding**

### Common Commands

**Database**:
- `\dt` - List all tables
- `\d users` - Describe users table
- `SELECT * FROM users;` - View all users

**Testing**:
- `go test ./features/users -v` - Test specific feature
- `go test ./... -cover` - Test with coverage report
- `go build` - Verify compilation

## API Endpoints

All endpoints organized by feature slice. See [ARCHITECTURE.md](ARCHITECTURE.md) for implementation details.

### 🔐 Authentication (Public)
- `GET /health` - Health check
- `GET /api/auth/google` - Initiate Google OAuth2 flow
- `GET /api/auth/google/callback` - OAuth2 callback handler
- `POST /api/auth/logout` - Sign out and clear session
- `GET /api/auth/me` - Get current authenticated user

### 👤 Users & Profiles (Authenticated)
- `GET /api/profile` - Get current user's full profile
- `PUT /api/profile` - Update profile (name, DOB, certification)

### ⚽ Matches (Permission: `can_assign_referees`)
- `POST /api/matches/import/parse` - Parse CSV file for preview
- `POST /api/matches/import/confirm` - Confirm and import matches
- `GET /api/matches` - List all matches with filters
- `PUT /api/matches/{id}` - Update match details (date, time, location)
- `POST /api/matches/{match_id}/roles/{role_type}/add` - Add role slot to match

### 👥 Referees (Permission: `can_assign_referees`)
- `GET /api/referees` - List all referees with status filtering
- `PUT /api/referees/{id}` - Update referee (status, grade, role)

### ✅ Eligibility (Permission: `can_assign_referees`)
- `GET /api/matches/{id}/eligible-referees?role={role_type}` - Get eligible referees for match/role

### 📋 Assignments (Permission: `can_assign_referees`)
- `POST /api/matches/{match_id}/roles/{role_type}/assign` - Assign/reassign/remove referee
- `GET /api/matches/{match_id}/conflicts?referee_id={id}&role_type={type}` - Check assignment conflicts

### 📅 Availability (Authenticated Referees)
- `GET /api/referee/matches` - Get eligible matches for current referee
- `POST /api/referee/matches/{id}/availability` - Toggle match availability (available/unavailable/clear)
- `GET /api/referee/day-unavailability` - Get all unavailable dates
- `POST /api/referee/day-unavailability/{date}` - Toggle full-day unavailability

### ✔️ Acknowledgment (Authenticated Referees)
- `POST /api/referee/matches/{match_id}/acknowledge` - Acknowledge assignment

### 🔑 RBAC Administration (Permission: `can_assign_roles`)
- `GET /api/admin/roles` - List all roles
- `GET /api/admin/permissions` - List all permissions
- `GET /api/admin/users/{id}/roles` - Get user's assigned roles
- `POST /api/admin/users/{id}/roles` - Assign role to user
- `DELETE /api/admin/users/{id}/roles/{roleId}` - Revoke role from user

### 📊 Audit Logging (Permission: `can_view_audit_logs`)
- `GET /api/admin/audit-logs` - Query audit logs with filters
- `GET /api/admin/audit-logs/export` - Export audit logs as CSV
- `POST /api/admin/audit-logs/purge` - Purge old audit logs

## User Roles & Permissions

### User Roles
- **pending_referee**: New user awaiting assignor approval (read-only access)
- **referee**: Active referee who can view matches and mark availability
- **assignor**: Admin who can manage referees and assign matches

### RBAC System (Epic 1)
The system uses **Role-Based Access Control (RBAC)** with granular permissions:

**Key Permissions**:
- `can_assign_referees` - Manage matches, assignments, and referee details
- `can_assign_roles` - Manage user roles and permissions
- `can_view_audit_logs` - Access audit log system

**Role Assignment**:
- Assignors can grant roles to users via the admin interface
- Multiple roles can be assigned to a single user
- Permissions are checked on every API request

See [EPIC_1_SUMMARY.md](EPIC_1_SUMMARY.md) for RBAC implementation details.

## Troubleshooting

### "Failed to connect to database"
- Ensure PostgreSQL container is running: `docker-compose ps`
- Check database logs: `docker-compose logs db`

### "OAuth2 error" or "Invalid credentials"
- Verify your Google OAuth2 credentials in `.env`
- Ensure the redirect URI matches exactly in Google Cloud Console

### Frontend can't connect to backend
- Ensure both containers are running: `docker-compose ps`
- Check CORS configuration in `backend/main.go`
- Verify `VITE_API_URL` in docker-compose.yml

## Documentation

See **[DOCS_INDEX.md](DOCS_INDEX.md)** for a complete guide to all documentation.

Key documents:
- **[GETTING_STARTED.md](GETTING_STARTED.md)** - Setup and usage guide
- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Production deployment guide
- **[PROJECT_STATUS.md](PROJECT_STATUS.md)** - Current status (91% complete)
- **[STORIES.md](STORIES.md)** - All epics and user stories

## Project Status

✅ **Epics 1-7 Complete** - All core features implemented  
🚧 **Epic 8 In Progress (91%)** - Backend refactoring to vertical slice architecture

**Recent Milestones**:
- ✅ Epic 1: Role-Based Access Control (RBAC)
- ✅ Epic 2: Audit Logging & Retention
- ✅ Epic 3-6: Core feature set (matches, assignments, availability)
- ✅ Epic 7: Self-hosted deployment infrastructure
- 🚧 Epic 8: Vertical slice architecture migration (7/9 stories complete)

**Current Architecture**: Vertical Slice Architecture with 7 feature slices and shared infrastructure. See [EPIC_8_PROGRESS.md](EPIC_8_PROGRESS.md) for details.

See [PROJECT_STATUS.md](PROJECT_STATUS.md) for detailed status.

## License

Private - Club use only
