package availability

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/msheeley/referee-scheduler/features/eligibility"
	"github.com/msheeley/referee-scheduler/shared/errors"
	"github.com/msheeley/referee-scheduler/shared/middleware"
)

// Handler handles HTTP requests for availability operations
type Handler struct {
	service ServiceInterface
	db      *sql.DB
}

// NewHandler creates a new availability handler
func NewHandler(service ServiceInterface, db ...*sql.DB) *Handler {
	h := &Handler{service: service}
	if len(db) > 0 {
		h.db = db[0]
	}
	return h
}

// ToggleMatchAvailability toggles a referee's availability for a match
func (h *Handler) ToggleMatchAvailability(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Parse match ID from URL
	vars := mux.Vars(r)
	matchID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid match ID"))
		return
	}

	// Parse request body
	var req ToggleMatchAvailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	// Call service
	result, err := h.service.ToggleMatchAvailability(r.Context(), matchID, user.ID, &req)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetDayUnavailability returns all days marked as unavailable for the current referee
func (h *Handler) GetDayUnavailability(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Call service
	days, err := h.service.GetDayUnavailability(r.Context(), user.ID)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(days)
}

// ToggleDayUnavailability toggles a referee's unavailability for a day
func (h *Handler) ToggleDayUnavailability(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		errors.WriteError(w, errors.NewUnauthorized("User not found in context"))
		return
	}

	// Parse date from URL
	vars := mux.Vars(r)
	date := vars["date"]

	// Parse request body
	var req ToggleDayUnavailabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.NewBadRequest("Invalid request body"))
		return
	}

	// Call service
	result, err := h.service.ToggleDayUnavailability(r.Context(), user.ID, date, &req)
	if err != nil {
		errors.WriteError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetRefereeAssignments returns only matches assigned to the current referee
func (h *Handler) GetRefereeAssignments(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized - user not found in context", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()

	page := 1
	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25
	if v := q.Get("per_page"); v != "" {
		if pp, err := strconv.Atoi(v); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	dateFrom := q.Get("date_from")
	dateTo := q.Get("date_to")

	queryArgs := []interface{}{user.ID}
	dateClause := "m.match_date >= CURRENT_DATE"
	argIdx := 2

	if dateFrom != "" {
		dateClause = fmt.Sprintf("m.match_date >= $%d", argIdx)
		queryArgs = append(queryArgs, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		dateClause += fmt.Sprintf(" AND m.match_date <= $%d", argIdx)
		queryArgs = append(queryArgs, dateTo)
		argIdx++
	}

	rows, err := h.db.Query(fmt.Sprintf(`
		SELECT
			m.id, m.event_name, m.team_name, m.age_group,
			m.match_date, m.start_time, m.end_time,
			m.location, m.description, m.status,
			a.position, a.acknowledged, a.acknowledged_at
		FROM matches m
		JOIN assignments a ON a.match_id = m.id AND a.referee_id = $1
		WHERE %s
		  AND m.archived = FALSE
		ORDER BY m.match_date ASC, m.start_time ASC, m.id ASC
	`, dateClause), queryArgs...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matches := []MatchForReferee{}

	for rows.Next() {
		var m MatchForReferee
		var matchDate time.Time
		var description sql.NullString
		var assignedRole sql.NullString
		var acknowledged bool
		var acknowledgedAt sql.NullTime

		err := rows.Scan(
			&m.ID, &m.EventName, &m.TeamName, &m.AgeGroup,
			&matchDate, &m.StartTime, &m.EndTime,
			&m.Location, &description, &m.Status,
			&assignedRole, &acknowledged, &acknowledgedAt,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}

		m.MatchDate = matchDate.Format("2006-01-02")
		if description.Valid {
			m.Description = &description.String
		}

		m.IsAssigned = true
		if assignedRole.Valid {
			m.AssignedRole = &assignedRole.String
		}
		m.Acknowledged = acknowledged
		if acknowledgedAt.Valid {
			ackTime := acknowledgedAt.Time.Format(time.RFC3339)
			m.AcknowledgedAt = &ackTime
		}

		conflictRows, conflictErr := h.db.Query(`
			SELECT
				m2.id, m2.event_name, m2.team_name,
				m2.start_time, mr2.position
			FROM matches m2
			JOIN assignments mr2 ON mr2.match_id = m2.id
			WHERE mr2.referee_id = $1
			  AND m2.id != $2
			  AND m2.status = 'active'
			  AND m2.archived = FALSE
			  AND m2.match_date = $3
			  AND m2.start_time < $5
			  AND m2.end_time > $4
		`, user.ID, m.ID, matchDate, m.StartTime, m.EndTime)

		if conflictErr != nil {
			fmt.Printf("Warning: Failed to check conflicts for match %d: %v\n", m.ID, conflictErr)
		} else {
			conflicts := []ConflictingMatch{}
			for conflictRows.Next() {
				var c ConflictingMatch
				scanErr := conflictRows.Scan(&c.MatchID, &c.EventName, &c.TeamName, &c.StartTime, &c.RoleType)
				if scanErr != nil {
					fmt.Printf("Warning: Failed to scan conflict: %v\n", scanErr)
					continue
				}
				conflicts = append(conflicts, c)
			}
			conflictRows.Close()

			if len(conflicts) > 0 {
				m.HasConflict = true
				m.ConflictingMatches = conflicts
			}
		}

		matches = append(matches, m)
	}

	total := len(matches)
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedRefereeMatchesResponse{
		Matches:    matches[start:end],
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

// GetEligibleMatchesForReferee returns all upcoming matches that the current referee is eligible for
func (h *Handler) GetEligibleMatchesForReferee(w http.ResponseWriter, r *http.Request) {
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized - user not found in context", http.StatusUnauthorized)
		return
	}

	q := r.URL.Query()

	page := 1
	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 25
	if v := q.Get("per_page"); v != "" {
		if pp, err := strconv.Atoi(v); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	dateFrom := q.Get("date_from")
	dateTo := q.Get("date_to")

	var hasProfile bool
	err := h.db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM users
			WHERE id = $1
			  AND first_name IS NOT NULL
			  AND last_name IS NOT NULL
			  AND date_of_birth IS NOT NULL
		)
	`, user.ID).Scan(&hasProfile)

	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	if !hasProfile {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(PaginatedRefereeMatchesResponse{
			Matches: []MatchForReferee{}, Page: 1, PerPage: perPage, Total: 0, TotalPages: 0,
		})
		return
	}

	var referee RefereeProfile
	err = h.db.QueryRow(`
		SELECT id, date_of_birth, certified, cert_expiry
		FROM users
		WHERE id = $1
	`, user.ID).Scan(&referee.ID, &referee.DOB, &referee.Certified, &referee.CertExpiry)

	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	queryArgs := []interface{}{user.ID}
	dateClause := "m.match_date >= CURRENT_DATE"
	argIdx := 2

	if dateFrom != "" {
		dateClause = fmt.Sprintf("m.match_date >= $%d", argIdx)
		queryArgs = append(queryArgs, dateFrom)
		argIdx++
	}
	if dateTo != "" {
		dateClause += fmt.Sprintf(" AND m.match_date <= $%d", argIdx)
		queryArgs = append(queryArgs, dateTo)
		argIdx++
	}

	rows, err := h.db.Query(fmt.Sprintf(`
		SELECT
			m.id, m.event_name, m.team_name, m.age_group,
			m.match_date, m.start_time, m.end_time,
			m.location, m.description, m.status
		FROM matches m
		WHERE %s
		  AND m.status = 'active'
		  AND m.archived = FALSE
		  AND (
			NOT EXISTS (
				SELECT 1 FROM day_unavailability du
				WHERE du.referee_id = $1 AND du.unavailable_date = m.match_date
			)
			OR
			EXISTS (
				SELECT 1 FROM assignments mr
				WHERE mr.match_id = m.id AND mr.referee_id = $1
			)
		  )
		ORDER BY m.match_date ASC, m.start_time ASC, m.id ASC
	`, dateClause), queryArgs...)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	matches := []MatchForReferee{}

	for rows.Next() {
		var m MatchForReferee
		var matchDate time.Time
		var description sql.NullString

		err := rows.Scan(
			&m.ID, &m.EventName, &m.TeamName, &m.AgeGroup,
			&matchDate, &m.StartTime, &m.EndTime,
			&m.Location, &description, &m.Status,
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("Scan error: %v", err), http.StatusInternalServerError)
			return
		}

		m.MatchDate = matchDate.Format("2006-01-02")

		if description.Valid {
			m.Description = &description.String
		}

		eligibleRoles := []string{}

		dobStr := referee.DOB.Format("2006-01-02")

		var certExpiryStr *string
		if referee.CertExpiry.Valid {
			certStr := referee.CertExpiry.Time.Format("2006-01-02")
			certExpiryStr = &certStr
		}

		isEligible, _ := eligibility.CheckEligibility(
			m.AgeGroup, "center", matchDate,
			&dobStr, referee.Certified, certExpiryStr,
		)
		if isEligible {
			eligibleRoles = append(eligibleRoles, "center")
		}

		isEligible, _ = eligibility.CheckEligibility(
			m.AgeGroup, "assistant_1", matchDate,
			&dobStr, referee.Certified, certExpiryStr,
		)
		if isEligible {
			eligibleRoles = append(eligibleRoles, "assistant")
		}

		if len(eligibleRoles) == 0 {
			continue
		}

		m.EligibleRoles = eligibleRoles

		var availableFlag sql.NullBool
		err = h.db.QueryRow(`
			SELECT available
			FROM availability
			WHERE match_id = $1 AND referee_id = $2
		`, m.ID, user.ID).Scan(&availableFlag)

		if err != nil && err != sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		if availableFlag.Valid {
			m.IsAvailable = availableFlag.Bool
			m.IsUnavailable = !availableFlag.Bool
		} else {
			m.IsAvailable = false
			m.IsUnavailable = false
		}

		var assignedRole sql.NullString
		var acknowledged bool
		var acknowledgedAt sql.NullTime
		err = h.db.QueryRow(`
			SELECT position, acknowledged, acknowledged_at
			FROM assignments
			WHERE match_id = $1 AND referee_id = $2
			LIMIT 1
		`, m.ID, user.ID).Scan(&assignedRole, &acknowledged, &acknowledgedAt)

		if err != nil && err != sql.ErrNoRows {
			http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
			return
		}

		if assignedRole.Valid {
			m.IsAssigned = true
			m.AssignedRole = &assignedRole.String
			m.Acknowledged = acknowledged
			if acknowledgedAt.Valid {
				ackTime := acknowledgedAt.Time.Format(time.RFC3339)
				m.AcknowledgedAt = &ackTime
			}

			conflictRows, err := h.db.Query(`
				SELECT
					m2.id, m2.event_name, m2.team_name,
					m2.start_time, mr2.position
				FROM matches m2
				JOIN assignments mr2 ON mr2.match_id = m2.id
				WHERE mr2.referee_id = $1
				  AND m2.id != $2
				  AND m2.status = 'active'
				  AND m2.archived = FALSE
				  AND m2.match_date = $3
				  AND m2.start_time < $5
				  AND m2.end_time > $4
			`, user.ID, m.ID, matchDate, m.StartTime, m.EndTime)

			if err != nil {
				fmt.Printf("Warning: Failed to check conflicts for match %d: %v\n", m.ID, err)
			} else {
				conflicts := []ConflictingMatch{}

				for conflictRows.Next() {
					var c ConflictingMatch
					err := conflictRows.Scan(&c.MatchID, &c.EventName, &c.TeamName, &c.StartTime, &c.RoleType)
					if err != nil {
						fmt.Printf("Warning: Failed to scan conflict: %v\n", err)
						continue
					}
					conflicts = append(conflicts, c)
				}
				conflictRows.Close()

				if len(conflicts) > 0 {
					m.HasConflict = true
					m.ConflictingMatches = conflicts
				}
			}
		}

		matches = append(matches, m)
	}

	total := len(matches)
	totalPages := total / perPage
	if total%perPage > 0 {
		totalPages++
	}

	start := (page - 1) * perPage
	if start > total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedRefereeMatchesResponse{
		Matches:    matches[start:end],
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}
