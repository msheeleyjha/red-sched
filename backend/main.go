package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"github.com/msheeley/referee-scheduler/features/matches"
	"github.com/msheeley/referee-scheduler/features/users"
	"github.com/msheeley/referee-scheduler/shared/config"
	"github.com/msheeley/referee-scheduler/shared/database"
	"github.com/msheeley/referee-scheduler/shared/middleware"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	cfg              *config.Config
	db               *sql.DB
	sessionStore     *sessions.CookieStore
	oauth2Config     *oauth2.Config
	auditLogger      *AuditLogger
	retentionService *AuditRetentionService

	// Middleware instances
	authMW *middleware.AuthMiddleware
	rbacMW *middleware.RBACMiddleware

	// Feature services (temporary, until auth is refactored)
	usersService *users.Service
)

func main() {
	// Load configuration
	cfg = config.Load()

	// Connect to database
	dbConn, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbConn.Close()

	// Set global db variable (will be removed when features are refactored)
	db = dbConn.DB

	// Run migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize audit logger
	auditLogger = NewAuditLogger(db)
	defer auditLogger.Close()
	log.Println("Audit logger initialized")

	// Initialize audit retention service
	retentionService = NewAuditRetentionService(db, cfg.AuditRetentionDays)
	retentionService.Start()
	defer retentionService.Stop()
	log.Println("Audit retention service started")

	// Initialize session store
	sessionStore = sessions.NewCookieStore([]byte(cfg.SessionSecret))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   cfg.IsProduction(),
		SameSite: http.SameSiteLaxMode,
	}

	// Initialize OAuth2 config
	oauth2Config = &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	// Initialize middleware
	authMW = middleware.NewAuthMiddleware(sessionStore, db)
	rbacMW = middleware.NewRBACMiddleware(sessionStore, db)

	// Initialize feature slices
	usersRepo := users.NewRepository(db)
	usersService = users.NewService(usersRepo)
	usersHandler := users.NewHandler(usersService)
	log.Println("Users feature initialized")

	matchesRepo := matches.NewRepository(db)
	matchesService := matches.NewService(matchesRepo)
	matchesHandler := matches.NewHandler(matchesService)
	log.Println("Matches feature initialized")

	// Setup router
	r := mux.NewRouter()

	// Public routes
	r.HandleFunc("/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/auth/google", googleAuthHandler).Methods("GET")
	r.HandleFunc("/api/auth/google/callback", googleCallbackHandler).Methods("GET")
	r.HandleFunc("/api/auth/logout", logoutHandler).Methods("POST")

	// Feature routes
	usersHandler.RegisterRoutes(r, authMiddleware)
	matchesHandler.RegisterRoutes(r, authMiddleware, requirePermission)

	// Referee management routes (assignors only)
	r.HandleFunc("/api/referees", authMiddleware(assignorOnly(listRefereesHandler))).Methods("GET")
	r.HandleFunc("/api/referees/{id}", authMiddleware(assignorOnly(updateRefereeHandler))).Methods("PUT")

	// Match management routes - moved to matches feature slice
	// Old routes commented out (now handled by matchesHandler):
	// r.HandleFunc("/api/matches/import/parse", authMiddleware(assignorOnly(parseCSVHandler))).Methods("POST")
	// r.HandleFunc("/api/matches/import/confirm", authMiddleware(assignorOnly(importMatchesHandler))).Methods("POST")
	// r.HandleFunc("/api/matches", authMiddleware(assignorOnly(listMatchesHandler))).Methods("GET")
	// r.HandleFunc("/api/matches/{id}", authMiddleware(assignorOnly(updateMatchHandler))).Methods("PUT")
	// r.HandleFunc("/api/matches/{match_id}/roles/{role_type}/add", authMiddleware(assignorOnly(addRoleSlotHandler))).Methods("POST")

	// Eligibility check route (still in old code)
	r.HandleFunc("/api/matches/{id}/eligible-referees", authMiddleware(assignorOnly(getEligibleRefereesHandler))).Methods("GET")

	// Referee availability routes
	r.HandleFunc("/api/referee/matches", authMiddleware(getEligibleMatchesForRefereeHandler)).Methods("GET")
	r.HandleFunc("/api/referee/matches/{id}/availability", authMiddleware(toggleAvailabilityHandler)).Methods("POST")
	r.HandleFunc("/api/referee/matches/{match_id}/acknowledge", authMiddleware(acknowledgeAssignmentHandler)).Methods("POST")

	// Day unavailability routes
	r.HandleFunc("/api/referee/day-unavailability", authMiddleware(getDayUnavailabilityHandler)).Methods("GET")
	r.HandleFunc("/api/referee/day-unavailability/{date}", authMiddleware(toggleDayUnavailabilityHandler)).Methods("POST")

	// Assignment routes (assignors only)
	r.HandleFunc("/api/matches/{match_id}/roles/{role_type}/assign", authMiddleware(assignorOnly(assignRefereeHandler))).Methods("POST")
	// r.HandleFunc("/api/matches/{match_id}/roles/{role_type}/add", authMiddleware(assignorOnly(addRoleSlotHandler))).Methods("POST") // Moved to matches feature
	r.HandleFunc("/api/matches/{match_id}/conflicts", authMiddleware(assignorOnly(getConflictingAssignmentsHandler))).Methods("GET")

	// Epic 1: RBAC Admin routes (requires can_assign_roles permission)
	r.HandleFunc("/api/admin/users/{id}/roles", requirePermission("can_assign_roles", assignRoleToUser)).Methods("POST")
	r.HandleFunc("/api/admin/users/{id}/roles/{roleId}", requirePermission("can_assign_roles", revokeRoleFromUser)).Methods("DELETE")
	r.HandleFunc("/api/admin/users/{id}/roles", requirePermission("can_assign_roles", getUserRoles)).Methods("GET")
	r.HandleFunc("/api/admin/roles", requirePermission("can_assign_roles", getAllRoles)).Methods("GET")
	r.HandleFunc("/api/admin/permissions", requirePermission("can_assign_roles", getAllPermissions)).Methods("GET")

	// Epic 2: Audit Logging routes (requires can_view_audit_logs permission)
	r.HandleFunc("/api/admin/audit-logs", requirePermission("can_view_audit_logs", getAuditLogsHandler)).Methods("GET")
	r.HandleFunc("/api/admin/audit-logs/export", requirePermission("can_view_audit_logs", exportAuditLogsHandler)).Methods("GET")
	r.HandleFunc("/api/admin/audit-logs/purge", requirePermission("can_view_audit_logs", purgeAuditLogsHandler)).Methods("POST")

	// Setup CORS
	corsHandler := middleware.NewCORSHandler(cfg.FrontendURL)

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, corsHandler.Handler(r)); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	})
}

func googleAuthHandler(w http.ResponseWriter, r *http.Request) {
	// Generate random state for CSRF protection
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// Store state in session
	session, _ := sessionStore.Get(r, "auth-session")
	session.Values["oauth_state"] = state
	session.Save(r, w)

	url := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify state
	session, err := sessionStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	state := r.URL.Query().Get("state")
	storedState, ok := session.Values["oauth_state"].(string)
	if !ok || state != storedState {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	token, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user info from Google
	client := oauth2Config.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Failed to get user info", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var userInfo struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		http.Error(w, "Failed to decode user info", http.StatusInternalServerError)
		return
	}

	// Find or create user
	user, err := usersService.FindOrCreate(r.Context(), userInfo.ID, userInfo.Email, userInfo.Name)
	if err != nil {
		log.Printf("Failed to find or create user: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Store user ID in session
	session.Values["user_id"] = user.ID
	session.Values["user_role"] = user.Role
	delete(session.Values, "oauth_state")
	if err := session.Save(r, w); err != nil {
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	// Redirect to frontend
	http.Redirect(w, r, cfg.FrontendURL+"/auth/callback", http.StatusTemporaryRedirect)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, err := sessionStore.Get(r, "auth-session")
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	// Clear session
	session.Values = make(map[interface{}]interface{})
	session.Options.MaxAge = -1
	session.Save(r, w)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "logged out"})
}

func meHandler(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value(userContextKey).(*User)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	})
}

// Middleware
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionStore.Get(r, "auth-session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"].(int64)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get user from database
		user, err := getUserByID(userID)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user to context
		ctx := r.Context()
		ctx = contextWithUser(ctx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// assignorOnly middleware ensures only assignors can access the route
func assignorOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(userContextKey).(*User)

		if user.Role != "assignor" {
			http.Error(w, "Forbidden: Assignor access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}
