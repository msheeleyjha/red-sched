# Developer Guide

Complete guide for developers working on the Referee Scheduler project.

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Project Structure](#project-structure)
3. [Development Setup](#development-setup)
4. [Adding a New Feature](#adding-a-new-feature)
5. [Testing Guidelines](#testing-guidelines)
6. [Code Patterns](#code-patterns)
7. [Database Migrations](#database-migrations)
8. [Debugging](#debugging)
9. [Common Tasks](#common-tasks)

---

## Architecture Overview

This project uses **Vertical Slice Architecture** where features are organized as self-contained slices rather than horizontal layers.

### Why Vertical Slices?

**Traditional Horizontal Layers** (old approach):
```
backend/
├── models/      # All models
├── handlers/    # All handlers
├── services/    # All services
└── repositories/  # All repositories
```
❌ Problems:
- Hard to find all code for one feature
- Changes touch many directories
- Merge conflicts on shared files
- Unclear feature boundaries

**Vertical Slices** (current approach):
```
backend/features/
├── users/        # Everything for users
├── matches/      # Everything for matches
└── assignments/  # Everything for assignments
```
✅ Benefits:
- All feature code in one place
- Easy to locate and modify
- Parallel development possible
- Clear feature boundaries
- Better testability

### Core Principles

1. **High Cohesion**: Everything related to a feature lives together
2. **Low Coupling**: Features don't depend on each other's internals
3. **Dependency Injection**: Use interfaces for testability
4. **Single Responsibility**: Each layer has one job
5. **Test Coverage**: 100% coverage for handler and service layers

---

## Project Structure

### Backend Structure

```
backend/
├── main.go                       # Entry point (307 lines)
│   ├── Configuration loading
│   ├── Database connection
│   ├── Feature initialization
│   └── Route registration
│
├── shared/                       # Shared infrastructure (used by all features)
│   ├── config/                  # Configuration management
│   │   ├── config.go           # Load environment variables
│   │   └── config_test.go      # Config tests
│   ├── database/                # Database utilities
│   │   ├── db.go               # Connection management
│   │   └── migrations.go       # Migration runner
│   ├── errors/                  # Standard error handling
│   │   ├── errors.go           # AppError type, error constructors
│   │   └── errors_test.go      # Error tests
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go             # Authentication middleware
│   │   ├── rbac.go             # RBAC permission checking
│   │   ├── cors.go             # CORS configuration
│   │   ├── logging.go          # Request logging
│   │   └── *_test.go           # Middleware tests
│   └── utils/                   # Shared utilities
│       ├── ip.go               # IP address extraction
│       └── ip_test.go          # Utility tests
│
└── features/                     # Feature slices (vertical architecture)
    ├── users/                   # User management & profiles
    │   ├── models.go           # Domain models (User, ProfileUpdateRequest)
    │   ├── repository.go       # Data access (RepositoryInterface)
    │   ├── service.go          # Business logic (Service)
    │   ├── service_interface.go # Service contract
    │   ├── handler.go          # HTTP handlers (Handler)
    │   ├── routes.go           # Route registration
    │   ├── service_test.go     # Service layer tests
    │   └── handler_test.go     # Handler layer tests
    │
    ├── matches/                 # Match management
    ├── assignments/             # Referee assignments
    ├── acknowledgment/          # Assignment acknowledgment
    ├── referees/                # Referee management
    ├── availability/            # Match & day availability
    └── eligibility/             # Eligibility checking
```

### Feature Slice Anatomy

Every feature slice follows this structure:

```
features/myfeature/
├── models.go                # Domain models and DTOs
├── repository.go            # Data access layer
├── service.go               # Business logic layer
├── service_interface.go     # Service contract (for DI)
├── handler.go               # HTTP request/response layer
├── routes.go                # Route registration
├── service_test.go          # Service layer unit tests
└── handler_test.go          # Handler layer unit tests
```

**Layer Responsibilities**:
- **Models**: Pure data structures, no logic
- **Repository**: Database queries, no business logic
- **Service**: Business rules, validation, orchestration
- **Handler**: HTTP parsing, JSON encoding, delegates to service
- **Routes**: Route registration with middleware

---

## Development Setup

### Prerequisites

- Docker and Docker Compose
- Go 1.22+ (for local development)
- Google Cloud Project with OAuth2 credentials

### Initial Setup

1. **Clone and configure**:
   ```bash
   git clone <repository-url>
   cd referee-scheduler
   cp .env.example .env
   ```

2. **Add OAuth credentials** to `.env`:
   ```
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

3. **Start services**:
   ```bash
   docker-compose up --build
   ```

4. **Access**:
   - Frontend: http://localhost:3000
   - Backend: http://localhost:8080
   - Health: http://localhost:8080/health

### Creating an Assignor Account

```bash
# 1. Sign in with Google to create account
# 2. Connect to database
docker exec -it referee-scheduler-db psql -U referee_scheduler

# 3. Promote to assignor
UPDATE users SET role = 'assignor', status = 'active' WHERE email = 'your-email@example.com';

# 4. Verify
SELECT id, email, role, status FROM users;

# 5. Exit
\q
```

### Running Backend Locally (Without Docker)

```bash
cd backend

# Install dependencies
go mod download

# Set environment variables (from .env)
export DATABASE_URL="postgres://referee_scheduler:password@localhost:5432/referee_scheduler?sslmode=disable&timezone=America/New_York"
export SESSION_SECRET="your-secret"
export GOOGLE_CLIENT_ID="your-id"
export GOOGLE_CLIENT_SECRET="your-secret"
export GOOGLE_REDIRECT_URL="http://localhost:8080/api/auth/google/callback"
export FRONTEND_URL="http://localhost:3000"

# Run
go run .
```

---

## Adding a New Feature

Follow these steps to add a new feature slice.

### Step 1: Create Directory Structure

```bash
cd backend/features
mkdir myfeature
cd myfeature
```

### Step 2: Create Models (`models.go`)

```go
package myfeature

// Domain model
type MyEntity struct {
    ID        int64  `json:"id"`
    Name      string `json:"name"`
    CreatedAt string `json:"created_at"`
}

// Request DTO
type CreateRequest struct {
    Name string `json:"name"`
}

// Response DTO (if different from domain model)
type MyEntityResponse struct {
    ID   int64  `json:"id"`
    Name string `json:"name"`
}
```

### Step 3: Create Repository (`repository.go`)

```go
package myfeature

import (
    "context"
    "database/sql"
)

// RepositoryInterface defines data access contract
type RepositoryInterface interface {
    Create(ctx context.Context, name string) (*MyEntity, error)
    FindByID(ctx context.Context, id int64) (*MyEntity, error)
    List(ctx context.Context) ([]MyEntity, error)
}

// Repository implements RepositoryInterface
type Repository struct {
    db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
    return &Repository{db: db}
}

// Create inserts a new entity
func (r *Repository) Create(ctx context.Context, name string) (*MyEntity, error) {
    query := `INSERT INTO my_entities (name, created_at) VALUES ($1, NOW()) RETURNING id, name, created_at`
    
    var entity MyEntity
    err := r.db.QueryRowContext(ctx, query, name).Scan(
        &entity.ID,
        &entity.Name,
        &entity.CreatedAt,
    )
    
    return &entity, err
}

// FindByID retrieves an entity by ID
func (r *Repository) FindByID(ctx context.Context, id int64) (*MyEntity, error) {
    query := `SELECT id, name, created_at FROM my_entities WHERE id = $1`
    
    var entity MyEntity
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &entity.ID,
        &entity.Name,
        &entity.CreatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, nil // Not found
    }
    
    return &entity, err
}
```

### Step 4: Create Service Interface (`service_interface.go`)

```go
package myfeature

import "context"

// ServiceInterface defines business logic contract
type ServiceInterface interface {
    Create(ctx context.Context, req CreateRequest) (*MyEntity, error)
    GetByID(ctx context.Context, id int64) (*MyEntity, error)
    List(ctx context.Context) ([]MyEntity, error)
}
```

### Step 5: Create Service (`service.go`)

```go
package myfeature

import (
    "context"
    
    appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// Service implements business logic
type Service struct {
    repo RepositoryInterface
}

// NewService creates a new service
func NewService(repo RepositoryInterface) *Service {
    return &Service{repo: repo}
}

// Create creates a new entity with validation
func (s *Service) Create(ctx context.Context, req CreateRequest) (*MyEntity, error) {
    // Validation
    if req.Name == "" {
        return nil, appErrors.NewBadRequest("Name is required")
    }
    
    // Business logic
    entity, err := s.repo.Create(ctx, req.Name)
    if err != nil {
        return nil, appErrors.NewInternal("Failed to create entity", err)
    }
    
    return entity, nil
}

// GetByID retrieves an entity with error handling
func (s *Service) GetByID(ctx context.Context, id int64) (*MyEntity, error) {
    entity, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, appErrors.NewInternal("Failed to get entity", err)
    }
    
    if entity == nil {
        return nil, appErrors.NewNotFound("MyEntity")
    }
    
    return entity, nil
}
```

### Step 6: Create Handler (`handler.go`)

```go
package myfeature

import (
    "encoding/json"
    "net/http"
    "strconv"
    
    "github.com/gorilla/mux"
    appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// Handler handles HTTP requests
type Handler struct {
    service ServiceInterface
}

// NewHandler creates a new handler
func NewHandler(service ServiceInterface) *Handler {
    return &Handler{service: service}
}

// Create handles POST /api/myfeature
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        appErrors.WriteError(w, appErrors.NewBadRequest("Invalid request body"))
        return
    }
    
    entity, err := h.service.Create(r.Context(), req)
    if err != nil {
        appErrors.WriteError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(entity)
}

// GetByID handles GET /api/myfeature/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    id, err := strconv.ParseInt(vars["id"], 10, 64)
    if err != nil {
        appErrors.WriteError(w, appErrors.NewBadRequest("Invalid ID"))
        return
    }
    
    entity, err := h.service.GetByID(r.Context(), id)
    if err != nil {
        appErrors.WriteError(w, err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(entity)
}
```

### Step 7: Create Routes (`routes.go`)

```go
package myfeature

import (
    "net/http"
    
    "github.com/gorilla/mux"
)

// RegisterRoutes registers all routes for this feature
func (h *Handler) RegisterRoutes(
    r *mux.Router,
    authMiddleware func(http.HandlerFunc) http.HandlerFunc,
    requirePermission func(string, http.HandlerFunc) http.HandlerFunc,
) {
    // Public routes
    r.HandleFunc("/api/myfeature", authMiddleware(h.Create)).Methods("POST")
    r.HandleFunc("/api/myfeature/{id}", authMiddleware(h.GetByID)).Methods("GET")
    
    // Permission-protected route
    r.HandleFunc("/api/myfeature/{id}", requirePermission("can_manage_entities", h.Delete)).Methods("DELETE")
}
```

### Step 8: Write Tests

See [Testing Guidelines](#testing-guidelines) section.

### Step 9: Register in `main.go`

```go
// In main() function, add feature initialization
myfeatureRepo := myfeature.NewRepository(db)
myfeatureService := myfeature.NewService(myfeatureRepo)
myfeatureHandler := myfeature.NewHandler(myfeatureService)
log.Println("MyFeature initialized")

// Register routes
myfeatureHandler.RegisterRoutes(r, authMiddleware, requirePermission)
```

### Step 10: Add Import

```go
import (
    // ... existing imports
    "github.com/msheeley/referee-scheduler/features/myfeature"
)
```

---

## Testing Guidelines

### Test Coverage Requirements

- **Service Layer**: 100% coverage (all business logic tested)
- **Handler Layer**: 100% coverage (all HTTP scenarios tested)
- **Repository Layer**: Integration tests or 0% (uses database mocks)

### Service Tests (`service_test.go`)

```go
package myfeature

import (
    "context"
    "errors"
    "testing"
)

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
    CreateFunc   func(ctx context.Context, name string) (*MyEntity, error)
    FindByIDFunc func(ctx context.Context, id int64) (*MyEntity, error)
}

func (m *mockRepository) Create(ctx context.Context, name string) (*MyEntity, error) {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, name)
    }
    return nil, errors.New("CreateFunc not implemented")
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (*MyEntity, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(ctx, id)
    }
    return nil, errors.New("FindByIDFunc not implemented")
}

func TestService_Create_Success(t *testing.T) {
    repo := &mockRepository{
        CreateFunc: func(ctx context.Context, name string) (*MyEntity, error) {
            return &MyEntity{ID: 1, Name: name}, nil
        },
    }
    
    service := NewService(repo)
    ctx := context.Background()
    
    entity, err := service.Create(ctx, CreateRequest{Name: "Test"})
    
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    
    if entity.ID != 1 {
        t.Errorf("Expected ID=1, got: %d", entity.ID)
    }
    
    if entity.Name != "Test" {
        t.Errorf("Expected Name='Test', got: %s", entity.Name)
    }
}

func TestService_Create_ValidationError(t *testing.T) {
    repo := &mockRepository{}
    service := NewService(repo)
    ctx := context.Background()
    
    _, err := service.Create(ctx, CreateRequest{Name: ""})
    
    if err == nil {
        t.Fatal("Expected validation error, got nil")
    }
}
```

### Handler Tests (`handler_test.go`)

```go
package myfeature

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/gorilla/mux"
)

// mockService implements ServiceInterface for testing
type mockService struct {
    CreateFunc func(ctx context.Context, req CreateRequest) (*MyEntity, error)
}

func (m *mockService) Create(ctx context.Context, req CreateRequest) (*MyEntity, error) {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, req)
    }
    return nil, errors.New("CreateFunc not implemented")
}

func TestHandler_Create_Success(t *testing.T) {
    service := &mockService{
        CreateFunc: func(ctx context.Context, req CreateRequest) (*MyEntity, error) {
            return &MyEntity{ID: 1, Name: req.Name}, nil
        },
    }
    
    handler := NewHandler(service)
    
    reqBody := CreateRequest{Name: "Test"}
    body, _ := json.Marshal(reqBody)
    req := httptest.NewRequest("POST", "/api/myfeature", bytes.NewBuffer(body))
    
    rr := httptest.NewRecorder()
    handler.Create(rr, req)
    
    if rr.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got: %d", rr.Code)
    }
    
    var response MyEntity
    if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
        t.Fatalf("Failed to decode response: %v", err)
    }
    
    if response.Name != "Test" {
        t.Errorf("Expected Name='Test', got: %s", response.Name)
    }
}

func TestHandler_Create_InvalidJSON(t *testing.T) {
    service := &mockService{}
    handler := NewHandler(service)
    
    req := httptest.NewRequest("POST", "/api/myfeature", bytes.NewBufferString("invalid json"))
    rr := httptest.NewRecorder()
    
    handler.Create(rr, req)
    
    if rr.Code != http.StatusBadRequest {
        t.Errorf("Expected status 400, got: %d", rr.Code)
    }
}
```

### Running Tests

```bash
# Test specific feature
go test ./features/myfeature -v

# Test all features
go test ./features/... -v

# Test with coverage
go test ./features/myfeature -cover

# Test with coverage report
go test ./features/myfeature -coverprofile=coverage.out
go tool cover -html=coverage.out
```

---

## Code Patterns

### Error Handling Pattern

```go
import appErrors "github.com/msheeley/referee-scheduler/shared/errors"

// Bad request (400)
if input == "" {
    return appErrors.NewBadRequest("Input is required")
}

// Not found (404)
if entity == nil {
    return appErrors.NewNotFound("Entity")
}

// Internal error (500)
if err != nil {
    return appErrors.NewInternal("Failed to save entity", err)
}

// Forbidden (403)
if !hasPermission {
    return appErrors.NewForbidden()
}

// Conflict (409)
if alreadyExists {
    return appErrors.NewConflict("Entity already exists")
}
```

### Repository Query Pattern

```go
// Single row query
func (r *Repository) FindByID(ctx context.Context, id int64) (*Entity, error) {
    query := `SELECT id, name FROM entities WHERE id = $1`
    
    var entity Entity
    err := r.db.QueryRowContext(ctx, query, id).Scan(&entity.ID, &entity.Name)
    
    if err == sql.ErrNoRows {
        return nil, nil // Return nil, not error for "not found"
    }
    
    return &entity, err
}

// Multiple rows query
func (r *Repository) List(ctx context.Context) ([]Entity, error) {
    query := `SELECT id, name FROM entities ORDER BY name`
    
    rows, err := r.db.QueryContext(ctx, query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var entities []Entity
    for rows.Next() {
        var entity Entity
        if err := rows.Scan(&entity.ID, &entity.Name); err != nil {
            return nil, err
        }
        entities = append(entities, entity)
    }
    
    return entities, rows.Err()
}
```

### Context Usage Pattern

```go
// Get user from context (set by auth middleware)
import "github.com/msheeley/referee-scheduler/shared/middleware"

user, ok := middleware.GetUserFromContext(r.Context())
if !ok {
    appErrors.WriteError(w, appErrors.NewUnauthorized())
    return
}

// Use user
log.Printf("User %d is making a request", user.ID)
```

---

## Database Migrations

### Creating a Migration

1. **Create migration files**:
   ```bash
   cd backend/migrations
   # Example: 012_add_my_feature.up.sql
   touch 012_add_my_feature.up.sql
   touch 012_add_my_feature.down.sql
   ```

2. **Write up migration** (`012_add_my_feature.up.sql`):
   ```sql
   CREATE TABLE my_entities (
       id BIGSERIAL PRIMARY KEY,
       name VARCHAR(255) NOT NULL,
       created_at TIMESTAMP NOT NULL DEFAULT NOW(),
       updated_at TIMESTAMP NOT NULL DEFAULT NOW()
   );
   
   CREATE INDEX idx_my_entities_name ON my_entities(name);
   ```

3. **Write down migration** (`012_add_my_feature.down.sql`):
   ```sql
   DROP TABLE IF EXISTS my_entities;
   ```

4. **Restart backend** to apply:
   ```bash
   docker-compose restart backend
   docker-compose logs -f backend
   ```

### Migration Best Practices

- **Always write down migrations** for rollback capability
- **Test migrations** before committing
- **Use transactions** where appropriate
- **Add indexes** for frequently queried columns
- **Use constraints** (NOT NULL, UNIQUE, FOREIGN KEY)
- **Don't modify existing migrations** (create new ones)

---

## Debugging

### Backend Debugging

1. **View logs**:
   ```bash
   docker-compose logs -f backend
   ```

2. **Add debug logging**:
   ```go
   import "log"
   
   log.Printf("Debug: variable value = %v", myVar)
   ```

3. **Check database queries**:
   ```go
   log.Printf("Executing query: %s with params: %v", query, params)
   ```

4. **Use delve (Go debugger)**:
   ```bash
   # Install
   go install github.com/go-delve/delve/cmd/dlv@latest
   
   # Run with debugger
   dlv debug
   
   # Set breakpoint
   (dlv) break mypackage.MyFunction
   (dlv) continue
   ```

### Database Debugging

```bash
# Connect to database
docker exec -it referee-scheduler-db psql -U referee_scheduler

# Check table structure
\d my_entities

# View data
SELECT * FROM my_entities;

# Check for locks
SELECT * FROM pg_locks;

# View running queries
SELECT pid, query, state FROM pg_stat_activity WHERE state = 'active';
```

### Common Issues

**Issue**: "database connection refused"
```bash
# Check database is running
docker-compose ps

# Check logs
docker-compose logs db

# Restart database
docker-compose restart db
```

**Issue**: "migration failed"
```bash
# Check migration files
ls -la backend/migrations/

# View schema_migrations table
docker exec -it referee-scheduler-db psql -U referee_scheduler
SELECT * FROM schema_migrations;

# Manually rollback if needed
# Run the down migration manually
```

---

## Common Tasks

### Add a New API Endpoint

1. Add method to service interface
2. Implement method in service
3. Write service tests
4. Add handler method
5. Write handler tests
6. Register route in routes.go

### Add a New Permission

1. Create migration to add permission to `permissions` table
2. Assign permission to roles via migration
3. Use `requirePermission` middleware in routes

### Change Database Schema

1. Create new migration files
2. Write up/down SQL
3. Restart backend
4. Update repository queries
5. Update tests

### Update Shared Infrastructure

1. Modify shared package (config, errors, middleware, utils)
2. Run tests: `go test ./shared/...`
3. Update affected features
4. Re-run all tests: `go test ./...`

### Debug a Failing Test

```bash
# Run test with verbose output
go test ./features/myfeature -v

# Run specific test
go test ./features/myfeature -run TestName -v

# Run with coverage to see what's not covered
go test ./features/myfeature -cover -v

# Add debug logging in test
t.Logf("Debug: value = %v", myVar)
```

---

## Additional Resources

- **[ARCHITECTURE.md](ARCHITECTURE.md)** - Complete architecture documentation
- **[EPIC_8_PROGRESS.md](EPIC_8_PROGRESS.md)** - Refactoring progress and examples
- **[STORY_8.3_COMPLETE.md](STORY_8.3_COMPLETE.md)** - Example of users feature implementation
- **[STORY_8.4_COMPLETE.md](STORY_8.4_COMPLETE.md)** - Example of matches feature implementation

---

## Questions?

For questions or issues:
1. Check existing documentation (ARCHITECTURE.md, story completion docs)
2. Review similar features for examples
3. Check commit history for context
4. Ask in team chat or create an issue

---

**Last Updated**: 2026-04-27  
**Architecture Version**: Vertical Slice Architecture (Epic 8)  
**Test Count**: 258 passing (100% handler/service coverage)
