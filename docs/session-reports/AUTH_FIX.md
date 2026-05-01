# Authentication Fix - /api/auth/me 401 Error

## Issue
User `msheeley@jackhenry.com` was getting a **401 Unauthorized** error with message "user not found in context" when trying to login and access `/api/auth/me`.

## Root Cause
There were **two different auth middleware implementations** in the codebase with incompatible context keys:

### Old Middleware (main.go)
```go
type contextKey string
const userContextKey contextKey = "user"

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    // Stores user with main.contextKey("user")
    ctx = contextWithUser(ctx, user)
}
```

### New Middleware (shared/middleware/auth.go)
```go
type contextKey string
const userContextKey contextKey = "user"

func GetUserFromContext(ctx context.Context) (*User, bool) {
    // Looks for user with middleware.contextKey("user")
    user, ok := ctx.Value(userContextKey).(*User)
}
```

**Even though both had the same value "user", they were DIFFERENT TYPES** (`main.contextKey` vs `middleware.contextKey`), so the context lookup failed.

### The Mismatch
- ✅ Old middleware **stored** user: `main.contextKey("user")`  
- ❌ New handlers **retrieved** user: `middleware.contextKey("user")`  
- **Result**: User not found → 401 error

## Solution
Updated all routes to use the **new AuthMiddleware** from `shared/middleware`:

### Changes Made
1. **main.go**: Changed all feature routes to use `authMW.RequireAuth` instead of old `authMiddleware`
2. **main.go**: Removed old `authMiddleware` function (no longer needed)
3. **availability.go**: Updated to use `middleware.GetUserFromContext()` instead of old context key

### Before
```go
usersHandler.RegisterRoutes(r, authMiddleware)  // OLD
```

### After
```go
usersHandler.RegisterRoutes(r, authMW.RequireAuth)  // NEW
```

## Testing
After restarting the backend server, you should now be able to:

1. ✅ Login with Google OAuth
2. ✅ Successfully call `/api/auth/me` without 401 error
3. ✅ Access all authenticated endpoints
4. ✅ See your user profile data

## Technical Details
- **AuthMiddleware instance**: Created in `main.go` line 97 as `authMW`
- **Proper context key**: Defined in `shared/middleware/auth.go` line 17
- **User retrieval**: All handlers now use `middleware.GetUserFromContext()`
- **Build status**: ✅ Backend compiles successfully

## Files Changed
- `backend/main.go` - Updated routes to use new middleware
- `backend/availability.go` - Updated context retrieval

## Commit
```
5800d77 - Fix: Use proper AuthMiddleware for /api/auth/me endpoint
```

---

**Status**: ✅ FIXED - Login should now work correctly
