# Remaining Backend Root Files - Analysis & Recommendations

After Epic 8 completion, there are still 8 files in the backend root directory. This document analyzes each file and provides recommendations for future cleanup.

**Date**: 2026-04-27  
**Status**: Post Epic 8 Analysis

---

## Summary

| File | Lines | Functions | Status | Recommendation | Effort |
|------|-------|-----------|--------|----------------|--------|
| **main.go** | 307 | 4 handlers + 1 middleware | ✅ Keep | Keep as entry point | N/A |
| **availability.go** | 280 | 1 handler | ⚠️ Partial | Migrate to features/availability/ | 1-2 hours |
| **user.go** | 127 | 4 helpers | ⚠️ Temporary | Delete after auth refactor | Part of auth refactor |
| **audit.go** | 160 | 8 functions | 🔮 Future | Migrate to features/audit/ | 3-4 hours |
| **audit_api.go** | 461 | 7 handlers | 🔮 Future | Migrate to features/audit/ | 3-4 hours |
| **audit_retention.go** | 206 | 6 functions | 🔮 Future | Migrate to features/audit/ | 3-4 hours |
| **rbac.go** | 195 | 6 functions | 🔮 Future | Migrate to shared/middleware/ or features/roles/ | 2-3 hours |
| **roles_api.go** | 333 | 5 handlers | 🔮 Future | Migrate to features/roles/ | 2-3 hours |
| **Total** | **2,069** | **~41** | | | **14-21 hours** |

---

## File-by-File Analysis

### 1. main.go (307 lines) - ✅ KEEP

**Purpose**: Application entry point

**Contents**:
- Configuration loading
- Database connection
- Feature initialization
- Route registration
- OAuth handlers (googleAuthHandler, googleCallbackHandler, logoutHandler)
- healthHandler
- authMiddleware (legacy, to be replaced)

**Recommendation**: **Keep as-is**

**Reason**: This is the entry point. The OAuth handlers could eventually move to features/auth/, but that's a separate refactoring effort.

**Future Work**:
- Create features/auth/ and move OAuth handlers
- Replace legacy authMiddleware with authMW.RequireAuth

---

### 2. availability.go (280 lines) - ⚠️ PARTIALLY MIGRATED

**Purpose**: Referee eligibility for matches

**Contents**:
- `getEligibleMatchesForRefereeHandler` - Still used (registered in main.go line 155)
- ~~`toggleAvailabilityHandler`~~ - **JUST DELETED** (was unused, migrated to features/availability/)

**Current Status**: 
- ✅ Cleaned up unused `toggleAvailabilityHandler` (removed 71 lines)
- ⚠️ One handler remains: `getEligibleMatchesForRefereeHandler`

**Recommendation**: **Migrate to features/availability/**

**Migration Plan**:
1. Move `getEligibleMatchesForRefereeHandler` to features/availability/handler.go
2. Add method to availability handler
3. Register route in features/availability/routes.go
4. Update main.go to remove old route
5. Delete availability.go

**Effort**: 1-2 hours

**Benefit**: Complete availability feature in one place, delete 280-line file

**Priority**: Medium (not urgent, but easy win)

---

### 3. user.go (127 lines) - ⚠️ TEMPORARY KEEP

**Purpose**: User helper functions for authentication

**Contents**:
```go
type User struct { ... }              // User model (also in features/users/)
type contextKey string                // Context key type
const userContextKey contextKey = "user"
func contextWithUser(ctx, user) ctx   // Used by main.go authMiddleware
func findOrCreateUser(...) (*User, error)  // Used by OAuth callback
func getUserByGoogleID(...) (*User, error) // Used by findOrCreateUser
func getUserByID(...) (*User, error)  // Used by main.go authMiddleware
```

**Currently Used By**:
- `main.go` authMiddleware (lines 338, 346)
- `main.go` googleCallbackHandler (line 275)

**Recommendation**: **Keep temporarily, delete after auth refactor**

**Why Keep**: Main.go still uses these helper functions for the legacy auth middleware.

**Future Migration Plan** (Auth Refactor):
1. Create features/auth/ for OAuth handlers
2. Move OAuth logic to features/auth/
3. Update features to use authMW.RequireAuth (from shared/middleware/)
4. Remove legacy authMiddleware from main.go
5. Delete user.go

**Effort**: Part of larger auth refactoring (3-4 hours total)

**Benefit**: Standardized auth, no duplicate implementations

**Priority**: Low (works fine for now, can wait)

---

### 4. audit.go (160 lines) - 🔮 FUTURE MIGRATION

**Purpose**: Audit logger core functionality

**Contents**:
```go
type AuditLogger struct { ... }
func NewAuditLogger(db) *AuditLogger
func (al *AuditLogger) Log(...) error
func (al *AuditLogger) LogWithDetails(...) error
func (al *AuditLogger) Close()
// ... 3 more helper functions
```

**Currently Used By**:
- main.go (initialization)
- Other parts of the codebase for audit logging

**Recommendation**: **Migrate to features/audit/**

**Migration Plan**:
1. Create features/audit/
2. Move AuditLogger to features/audit/logger.go
3. Move audit API handlers to features/audit/handler.go
4. Move retention service to features/audit/retention.go
5. Create routes, tests
6. Update main.go
7. Delete audit.go, audit_api.go, audit_retention.go

**Effort**: 3-4 hours (all audit files together)

**Benefit**: Complete audit feature in one place, well-tested

**Priority**: Medium (nice to have, not urgent)

---

### 5. audit_api.go (461 lines) - 🔮 FUTURE MIGRATION

**Purpose**: Audit log API handlers

**Contents**:
```go
func getAuditLogsHandler(w, r)         // GET /api/admin/audit-logs
func exportAuditLogsHandler(w, r)      // GET /api/admin/audit-logs/export
func purgeAuditLogsHandler(w, r)       // POST /api/admin/audit-logs/purge
// ... plus helper functions for CSV generation, filtering, etc.
```

**Recommendation**: **Migrate to features/audit/** (same as audit.go)

**Effort**: Included in audit.go migration (3-4 hours total)

---

### 6. audit_retention.go (206 lines) - 🔮 FUTURE MIGRATION

**Purpose**: Audit log retention service (automatic purging)

**Contents**:
```go
type AuditRetentionService struct { ... }
func NewAuditRetentionService(...) *AuditRetentionService
func (s *AuditRetentionService) Start()
func (s *AuditRetentionService) Stop()
func (s *AuditRetentionService) runDailyCleanup()
func (s *AuditRetentionService) purgeOldLogs(...) int
```

**Recommendation**: **Migrate to features/audit/** (same as audit.go)

**Effort**: Included in audit.go migration (3-4 hours total)

---

### 7. rbac.go (195 lines) - 🔮 FUTURE MIGRATION

**Purpose**: RBAC helper functions and middleware

**Contents**:
```go
func requirePermission(permission, handler) http.HandlerFunc  // Middleware
func getUserPermissions(userID) ([]string, error)            // Helper
func hasPermission(userID, permission) (bool, error)         // Helper
func assignRoleToUser(w, r)                                  // POST /api/admin/users/{id}/roles
func revokeRoleFromUser(w, r)                                // DELETE /api/admin/users/{id}/roles/{roleId}
func getUserRoles(w, r)                                      // GET /api/admin/users/{id}/roles
```

**Recommendation**: **Split between shared/middleware/ and features/roles/**

**Migration Plan**:
1. Move RBAC middleware functions to shared/middleware/rbac.go (already partially there)
2. Move role assignment handlers to features/roles/
3. Update main.go
4. Delete rbac.go

**Effort**: 2-3 hours

**Benefit**: Clear separation of middleware vs. API handlers

**Priority**: Medium (works fine, but could be cleaner)

---

### 8. roles_api.go (333 lines) - 🔮 FUTURE MIGRATION

**Purpose**: Roles and permissions API handlers

**Contents**:
```go
func getAllRoles(w, r)                  // GET /api/admin/roles
func getAllPermissions(w, r)            // GET /api/admin/permissions
// ... plus helper functions
```

**Recommendation**: **Migrate to features/roles/**

**Migration Plan**:
1. Create features/roles/
2. Move all role/permission handlers
3. Combine with rbac.go role assignment handlers
4. Create routes, tests
5. Delete roles_api.go

**Effort**: 2-3 hours (combined with rbac.go)

**Benefit**: Complete roles feature in one place

**Priority**: Medium (works fine, but could be cleaner)

---

## Immediate Actions Taken (Post Epic 8)

### ✅ Cleaned availability.go
- **Deleted**: `toggleAvailabilityHandler` function (71 lines)
- **Removed**: Unused imports (strconv, gorilla/mux)
- **Result**: 353 lines → 280 lines (20% reduction)
- **Status**: Build passing ✅, tests passing ✅

---

## Recommended Migration Order

### Phase 1: Quick Wins (Optional, Low Priority)

#### 1.1 Complete Availability Migration (1-2 hours)
**Why**: Only one handler left, easy to finish
**Impact**: Delete 280-line file, complete features/availability/

**Steps**:
1. Move `getEligibleMatchesForRefereeHandler` to features/availability/
2. Update route in main.go
3. Delete availability.go

---

### Phase 2: Feature Migrations (Medium Priority, Future Work)

#### 2.1 Migrate Audit Feature (3-4 hours)
**Why**: Large, well-defined feature (827 lines total)
**Impact**: Clean vertical slice for audit logging

**Files to migrate**:
- audit.go (160 lines)
- audit_api.go (461 lines)
- audit_retention.go (206 lines)

**Steps**:
1. Create features/audit/
2. Create models.go (AuditLog, AuditLogger structs)
3. Create repository.go (database queries)
4. Create service.go (logger, retention service)
5. Create handler.go (API handlers)
6. Create routes.go
7. Create tests (service_test.go, handler_test.go)
8. Update main.go
9. Delete old files

#### 2.2 Migrate Roles/RBAC Feature (2-3 hours)
**Why**: Complete RBAC in one place (528 lines total)
**Impact**: Clear separation of middleware vs. feature

**Files to migrate**:
- rbac.go (195 lines) - Split between middleware and features/roles/
- roles_api.go (333 lines)

**Steps**:
1. Create features/roles/
2. Move role assignment handlers from rbac.go
3. Move role/permission query handlers from roles_api.go
4. Create routes, tests
5. Ensure RBAC middleware stays in shared/middleware/
6. Delete old files

---

### Phase 3: Auth Refactor (Lower Priority, Larger Effort)

#### 3.1 Refactor Auth to Feature Slice (3-4 hours)
**Why**: Standardize auth, remove duplicate implementations
**Impact**: Delete user.go, cleaner auth flow

**Steps**:
1. Create features/auth/
2. Move OAuth handlers from main.go
3. Move user helpers from user.go
4. Update all features to use authMW.RequireAuth
5. Remove legacy authMiddleware from main.go
6. Delete user.go

---

## Summary Statistics

### Current State (Post Cleanup)

| Metric | Value |
|--------|-------|
| Files in backend root | 8 |
| Total lines | 2,069 |
| Handlers | ~20 |
| Helper functions | ~21 |

### After All Migrations

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Files in backend root | 8 | 1 | -7 (-87.5%) |
| Lines in root | 2,069 | 307 | -1,762 (-85.2%) |
| Feature slices | 7 | 9 | +2 (audit, roles) |
| Complete features | 7 | 9 | +2 |

---

## Recommendations

### Immediate (Do Now)
1. ✅ **Clean up availability.go** - DONE! Removed toggleAvailabilityHandler

### Short Term (Next Sprint, Optional)
2. **Complete availability migration** - Easy 1-2 hour task, completes the feature

### Medium Term (Future Epics, When Needed)
3. **Migrate audit feature** - When audit needs updates or enhancements
4. **Migrate roles/RBAC feature** - When RBAC needs changes
5. **Refactor auth** - When standardizing auth becomes necessary

### Long Term (Nice to Have)
6. Keep only main.go in backend root (all features in slices)

---

## Conclusion

The remaining 8 files in the backend root can be categorized as:

1. **Keep**: main.go (entry point)
2. **Quick win**: availability.go (1-2 hours to complete migration)
3. **Temporary**: user.go (delete after auth refactor)
4. **Future migrations**: audit files, rbac files (total 14-21 hours)

**Key Insight**: These files are not urgent. They work fine as-is. Migrate them when you're making changes to those features anyway, not as a standalone cleanup effort.

**Total remaining cleanup effort**: 14-21 hours (can be spread across multiple sprints)

---

**Last Updated**: 2026-04-27  
**Status**: Post Epic 8 Analysis  
**Next Recommended Action**: Complete availability.go migration (1-2 hours, optional)
