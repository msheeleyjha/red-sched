package middleware

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/sessions"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	userContextKey        contextKey = "user"
	userPermissionsKey    contextKey = "userPermissions"
)

// AuthMiddleware provides authentication middleware
type AuthMiddleware struct {
	sessionStore *sessions.CookieStore
	db           *sql.DB
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(sessionStore *sessions.CookieStore, db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{
		sessionStore: sessionStore,
		db:           db,
	}
}

// RequireAuth middleware ensures user is authenticated
func (am *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := am.sessionStore.Get(r, "auth-session")
		if err != nil {
			log.Printf("Session error: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(int64)
		if !ok || userID == 0 {
			http.Error(w, "Unauthorized - not authenticated", http.StatusUnauthorized)
			return
		}

		// Get user from database and add to context
		user, err := am.getUserByID(userID)
		if err != nil {
			log.Printf("Failed to get user: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetCurrentUserID gets the current user ID from session
func (am *AuthMiddleware) GetCurrentUserID(r *http.Request) (int64, error) {
	session, err := am.sessionStore.Get(r, "auth-session")
	if err != nil {
		return 0, fmt.Errorf("session error: %w", err)
	}

	userID, ok := session.Values["user_id"].(int64)
	if !ok || userID == 0 {
		return 0, fmt.Errorf("user not authenticated")
	}

	return userID, nil
}

// User represents a user (temporary, will be moved to features/users/models.go)
type User struct {
	ID    int64
	Email string
	Name  string
	Role  string // Legacy field, kept for backward compatibility
}

// getUserByID retrieves a user from the database
func (am *AuthMiddleware) getUserByID(userID int64) (*User, error) {
	var user User
	err := am.db.QueryRow(
		"SELECT id, email, name, role FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.Email, &user.Name, &user.Role)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	return &user, nil
}

// GetUserFromContext retrieves the user from request context
func GetUserFromContext(ctx context.Context) (*User, bool) {
	user, ok := ctx.Value(userContextKey).(*User)
	return user, ok
}

// SetUserInContext adds a user to the context (for testing)
func SetUserInContext(ctx context.Context, user *User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
