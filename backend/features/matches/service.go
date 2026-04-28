package matches

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"regexp"
	"strings"
	"time"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles match business logic
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new match service
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// getEasternLocation returns US Eastern timezone
// All match dates and times are stored and displayed in US Eastern Time
func getEasternLocation() (*time.Location, error) {
	return time.LoadLocation("America/New_York")
}

// extractAgeGroup extracts age group from team name
// Pattern: "Under {N}" → U{N}
// Examples: "Under 12 Girls - Falcons" → "U12", "Under 8 Boys" → "U8"
func extractAgeGroup(teamName string) *string {
	re := regexp.MustCompile(`(?i)under\s+(\d+)`)
	matches := re.FindStringSubmatch(teamName)
	if len(matches) > 1 {
		ageGroup := "U" + matches[1]
		return &ageGroup
	}
	return nil
}

// ParseCSV parses an uploaded CSV file and returns preview with errors
func (s *Service) ParseCSV(file multipart.File, filename string) (*ImportPreviewResponse, error) {
	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(filename), ".csv") {
		return nil, errors.NewBadRequest("Only .csv files are accepted")
	}

	// Parse CSV
	reader := csv.NewReader(file)

	// Read header row
	headers, err := reader.Read()
	if err != nil {
		return nil, errors.NewBadRequest("Failed to read CSV headers")
	}

	// Map column names to indices (case-insensitive)
	colMap := make(map[string]int)
	for i, header := range headers {
		colMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Validate required columns
	requiredCols := []string{"event_name", "team_name", "start_date", "start_time", "end_time", "location"}
	for _, col := range requiredCols {
		if _, exists := colMap[col]; !exists {
			return nil, errors.NewBadRequest(fmt.Sprintf("Missing required column: %s", col))
		}
	}

	// Parse all rows
	rows := []CSVRow{}
	rowNumber := 1 // Start at 1 (header is row 0)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip malformed rows but continue parsing
			rowNumber++
			continue
		}

		rowNumber++

		// Extract fields
		row := CSVRow{
			RowNumber: rowNumber,
		}

		// Safe column access
		getCol := func(name string) string {
			if idx, ok := colMap[name]; ok && idx < len(record) {
				return strings.TrimSpace(record[idx])
			}
			return ""
		}

		row.EventName = getCol("event_name")
		row.TeamName = getCol("team_name")
		row.StartDate = getCol("start_date")
		row.EndDate = getCol("end_date")
		row.StartTime = getCol("start_time")
		row.EndTime = getCol("end_time")
		row.Description = getCol("description")
		row.Location = getCol("location")
		row.ReferenceID = getCol("reference_id")

		// Validate required fields
		if row.EventName == "" || row.TeamName == "" || row.StartDate == "" ||
			row.StartTime == "" || row.EndTime == "" || row.Location == "" {
			errMsg := "Missing required field(s)"
			row.Error = &errMsg
			rows = append(rows, row)
			continue
		}

		// Extract age group
		row.AgeGroup = extractAgeGroup(row.TeamName)
		if row.AgeGroup == nil {
			errMsg := "Unrecognised age group - could not extract 'Under N' from team name"
			row.Error = &errMsg
		}

		rows = append(rows, row)
	}

	// Check for duplicates
	duplicates := s.detectDuplicates(rows)

	response := &ImportPreviewResponse{
		Rows:       rows,
		Duplicates: duplicates,
	}

	return response, nil
}

// detectDuplicates finds duplicate matches in the upload
func (s *Service) detectDuplicates(rows []CSVRow) []DuplicateMatchGroup {
	duplicates := []DuplicateMatchGroup{}

	// Signal A: Same reference_id
	refIDMap := make(map[string][]CSVRow)
	for _, row := range rows {
		if row.ReferenceID != "" && row.Error == nil {
			refIDMap[row.ReferenceID] = append(refIDMap[row.ReferenceID], row)
		}
	}

	for _, matches := range refIDMap {
		if len(matches) > 1 {
			duplicates = append(duplicates, DuplicateMatchGroup{
				Signal:  "reference_id",
				Matches: matches,
			})
		}
	}

	// Signal B: Same date + start time + location (different reference_id)
	// TODO: Implement in future story

	return duplicates
}

// ImportMatches confirms and imports matches to database
func (s *Service) ImportMatches(ctx context.Context, req *ImportConfirmRequest, currentUserID int64) (*ImportResult, error) {
	imported := 0
	skipped := 0
	errs := []string{}

	// Load US Eastern timezone
	loc, err := getEasternLocation()
	if err != nil {
		return nil, errors.NewInternal("Failed to load timezone", err)
	}

	for _, row := range req.Rows {
		// Skip rows with unresolved errors
		if row.Error != nil {
			skipped++
			continue
		}

		// Parse date in Eastern Time
		var matchDate time.Time
		// Try parsing as YYYY-MM-DD
		parsedDate, err := time.ParseInLocation("2006-01-02", row.StartDate, loc)
		if err != nil {
			// Try parsing as DD/MM/YYYY
			parsedDate, err = time.ParseInLocation("02/01/2006", row.StartDate, loc)
			if err != nil {
				errs = append(errs, fmt.Sprintf("Row %d: Invalid date format: %s", row.RowNumber, row.StartDate))
				skipped++
				continue
			}
		}
		matchDate = parsedDate

		// Create match object
		match := &Match{
			EventName:   row.EventName,
			TeamName:    row.TeamName,
			AgeGroup:    row.AgeGroup,
			MatchDate:   matchDate,
			StartTime:   row.StartTime,
			EndTime:     row.EndTime,
			Location:    row.Location,
			Description: &row.Description,
			ReferenceID: &row.ReferenceID,
			Status:      "active",
			CreatedBy:   currentUserID,
		}

		// Insert match
		createdMatch, err := s.repo.Create(ctx, match)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Row %d: Database error: %s", row.RowNumber, err.Error()))
			skipped++
			continue
		}

		// Create role slots based on age group
		if row.AgeGroup != nil {
			err = s.CreateRoleSlotsForMatch(ctx, createdMatch.ID, *row.AgeGroup)
			if err != nil {
				errs = append(errs, fmt.Sprintf("Row %d: Failed to create role slots: %s", row.RowNumber, err.Error()))
			}
		}

		imported++
	}

	return &ImportResult{
		Imported: imported,
		Skipped:  skipped,
		Errors:   errs,
	}, nil
}

// CreateRoleSlotsForMatch creates appropriate role slots based on age group
// U6/U8: 1 center
// U10: 1 center (assistant slots can be added manually later)
// U12+: 1 center + 2 assistants
func (s *Service) CreateRoleSlotsForMatch(ctx context.Context, matchID int64, ageGroup string) error {
	// Extract numeric age from U6, U8, etc.
	age, err := GetAgeGroupInt(&ageGroup)
	if err != nil {
		return err
	}

	// All matches get a center referee slot
	err = s.repo.CreateRole(ctx, matchID, "center")
	if err != nil {
		return err
	}

	// U12+ matches get 2 assistant referee slots
	if age >= 12 {
		err = s.repo.CreateRole(ctx, matchID, "assistant_1")
		if err != nil {
			return err
		}
		err = s.repo.CreateRole(ctx, matchID, "assistant_2")
		if err != nil {
			return err
		}
	}

	// U10 gets no assistant slots by default (assignor can add manually)
	// U6/U8 get no assistant slots

	return nil
}

// ListMatches returns all matches with role assignments
func (s *Service) ListMatches(ctx context.Context) ([]MatchWithRoles, error) {
	matches, err := s.repo.List(ctx)
	if err != nil {
		return nil, errors.NewInternal("Failed to fetch matches", err)
	}

	result := []MatchWithRoles{}
	for _, match := range matches {
		roles, status, hasOverdueAck, err := s.getMatchRolesAndStatus(ctx, match.ID, match.AgeGroup)
		if err != nil {
			// Log error but continue
			continue
		}

		result = append(result, MatchWithRoles{
			Match:            match,
			Roles:            roles,
			AssignmentStatus: status,
			HasOverdueAck:    hasOverdueAck,
		})
	}

	return result, nil
}

// getMatchRolesAndStatus fetches roles and determines assignment status
func (s *Service) getMatchRolesAndStatus(ctx context.Context, matchID int64, ageGroup *string) ([]MatchRole, string, bool, error) {
	roles, err := s.repo.GetRoles(ctx, matchID)
	if err != nil {
		return nil, "unassigned", false, err
	}

	// Determine if this is a U10 match (ARs are optional)
	isU10OrYounger := false
	if ageGroup != nil {
		age, err := GetAgeGroupInt(ageGroup)
		if err == nil && age <= 10 {
			isU10OrYounger = true
		}
	}

	totalSlots := 0
	assignedSlots := 0
	hasOverdueAck := false

	for _, role := range roles {
		// Check for overdue acknowledgment
		if role.AckOverdue {
			hasOverdueAck = true
		}

		// For U10 and younger, only count center referee toward assignment status
		// ARs are optional and don't affect whether match is "full" or "partial"
		if isU10OrYounger && (role.RoleType == "assistant_1" || role.RoleType == "assistant_2") {
			continue
		}

		totalSlots++
		if role.AssignedRefereeID != nil {
			assignedSlots++
		}
	}

	// Determine assignment status
	status := "unassigned"
	if assignedSlots == totalSlots && totalSlots > 0 {
		status = "full"
	} else if assignedSlots > 0 {
		status = "partial"
	}

	return roles, status, hasOverdueAck, nil
}

// GetMatchWithRoles retrieves a match with its roles
func (s *Service) GetMatchWithRoles(ctx context.Context, matchID int64) (*MatchWithRoles, error) {
	match, err := s.repo.FindByID(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to fetch match", err)
	}
	if match == nil {
		return nil, errors.NewNotFound("Match")
	}

	roles, status, hasOverdueAck, err := s.getMatchRolesAndStatus(ctx, match.ID, match.AgeGroup)
	if err != nil {
		return nil, errors.NewInternal("Failed to fetch match roles", err)
	}

	return &MatchWithRoles{
		Match:            *match,
		Roles:            roles,
		AssignmentStatus: status,
		HasOverdueAck:    hasOverdueAck,
	}, nil
}

// UpdateMatch updates a match and reconfigures roles if age group changes
func (s *Service) UpdateMatch(ctx context.Context, matchID int64, req *MatchUpdateRequest, actorID int64) (*MatchWithRoles, error) {
	// Get current match
	currentMatch, err := s.repo.FindByID(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to fetch match", err)
	}
	if currentMatch == nil {
		return nil, errors.NewNotFound("Match")
	}

	// Build update map
	updates := make(map[string]interface{})

	if req.EventName != nil {
		updates["event_name"] = *req.EventName
	}
	if req.TeamName != nil {
		updates["team_name"] = *req.TeamName
	}

	ageGroupChanged := false
	if req.AgeGroup != nil && *req.AgeGroup != "" {
		// Check if age group is changing
		currentAgeGroupStr := ""
		if currentMatch.AgeGroup != nil {
			currentAgeGroupStr = *currentMatch.AgeGroup
		}
		if *req.AgeGroup != currentAgeGroupStr {
			ageGroupChanged = true
		}
		updates["age_group"] = *req.AgeGroup
	}

	if req.MatchDate != nil {
		// Parse the date in Eastern Time
		loc, err := getEasternLocation()
		if err != nil {
			return nil, errors.NewInternal("Failed to load timezone", err)
		}
		parsedDate, err := time.ParseInLocation("2006-01-02", *req.MatchDate, loc)
		if err != nil {
			return nil, errors.NewBadRequest("Invalid date format. Use YYYY-MM-DD")
		}
		updates["match_date"] = parsedDate
	}

	if req.StartTime != nil {
		updates["start_time"] = *req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = *req.EndTime
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Status != nil {
		// Validate status
		validStatuses := map[string]bool{"active": true, "cancelled": true}
		if !validStatuses[*req.Status] {
			return nil, errors.NewBadRequest("Invalid status. Must be: active or cancelled")
		}
		updates["status"] = *req.Status
	}

	if len(updates) == 0 {
		return nil, errors.NewBadRequest("No updates provided")
	}

	// Update match
	_, err = s.repo.Update(ctx, matchID, updates)
	if err != nil {
		return nil, errors.NewInternal("Failed to update match", err)
	}

	// If age group changed, reconfigure role slots
	if ageGroupChanged && req.AgeGroup != nil {
		err = s.reconfigureRoleSlots(ctx, matchID, *req.AgeGroup)
		if err != nil {
			// Log error but don't fail the update
			// In production, use proper logger
		}
	}

	// Log the edit
	changeDesc := s.buildChangeDescription(req)
	err = s.repo.LogEdit(ctx, matchID, actorID, changeDesc)
	if err != nil {
		// Log error but don't fail the update
	}

	// Return updated match with roles
	return s.GetMatchWithRoles(ctx, matchID)
}

// buildChangeDescription builds a human-readable description of changes
func (s *Service) buildChangeDescription(changes *MatchUpdateRequest) string {
	parts := []string{}

	if changes.EventName != nil {
		parts = append(parts, "event_name")
	}
	if changes.TeamName != nil {
		parts = append(parts, "team_name")
	}
	if changes.AgeGroup != nil {
		parts = append(parts, "age_group")
	}
	if changes.MatchDate != nil {
		parts = append(parts, "date")
	}
	if changes.StartTime != nil || changes.EndTime != nil {
		parts = append(parts, "time")
	}
	if changes.Location != nil {
		parts = append(parts, "location")
	}
	if changes.Description != nil {
		parts = append(parts, "description")
	}
	if changes.Status != nil {
		parts = append(parts, "status")
	}

	return "Match edited: " + strings.Join(parts, ", ")
}

// reconfigureRoleSlots adjusts role slots when age group changes
func (s *Service) reconfigureRoleSlots(ctx context.Context, matchID int64, newAgeGroup string) error {
	// Extract numeric age
	age, err := GetAgeGroupInt(&newAgeGroup)
	if err != nil {
		return err
	}

	// Get current role slots
	currentRoles, err := s.repo.GetCurrentRoles(ctx, matchID)
	if err != nil {
		return err
	}

	// Determine required roles for new age group
	requiredRoles := []string{"center"}
	if age >= 12 {
		requiredRoles = append(requiredRoles, "assistant_1", "assistant_2")
	}
	// U10 and below: center only (but keep existing assistants if any)

	// Add missing roles
	for _, required := range requiredRoles {
		found := false
		for _, current := range currentRoles {
			if current == required {
				found = true
				break
			}
		}
		if !found {
			err := s.repo.CreateRole(ctx, matchID, required)
			if err != nil {
				return err
			}
		}
	}

	// For U6/U8: Remove assistant slots if they exist
	if age < 10 {
		err := s.repo.DeleteRoles(ctx, matchID, []string{"assistant_1", "assistant_2"})
		if err != nil {
			return err
		}
	}

	return nil
}

// AddRoleSlot allows assignor to manually add AR slots to matches (e.g., for U10)
func (s *Service) AddRoleSlot(ctx context.Context, matchID int64, roleType string) error {
	// Validate role type
	if roleType != "assistant_1" && roleType != "assistant_2" {
		return errors.NewBadRequest("Can only add assistant referee slots")
	}

	// Verify match exists and is active
	exists, err := s.repo.MatchExists(ctx, matchID)
	if err != nil {
		return errors.NewInternal("Failed to verify match", err)
	}
	if !exists {
		return errors.NewNotFound("Match")
	}

	// Check if role slot already exists
	roleExists, err := s.repo.RoleExists(ctx, matchID, roleType)
	if err != nil {
		return errors.NewInternal("Failed to check role existence", err)
	}
	if roleExists {
		return errors.NewBadRequest("Role slot already exists for this match")
	}

	// Create the role slot
	err = s.repo.CreateRole(ctx, matchID, roleType)
	if err != nil {
		return errors.NewInternal("Failed to create role slot", err)
	}

	return nil
}

// ListActiveMatches retrieves all non-archived matches with their role assignments
func (s *Service) ListActiveMatches(ctx context.Context) ([]MatchWithRoles, error) {
	matches, err := s.repo.ListActive(ctx)
	if err != nil {
		return nil, errors.NewInternal("Failed to list active matches", err)
	}

	return s.enrichMatchesWithRoles(ctx, matches)
}

// ListArchivedMatches retrieves all archived matches with their role assignments
func (s *Service) ListArchivedMatches(ctx context.Context) ([]MatchWithRoles, error) {
	matches, err := s.repo.ListArchived(ctx)
	if err != nil {
		return nil, errors.NewInternal("Failed to list archived matches", err)
	}

	return s.enrichMatchesWithRoles(ctx, matches)
}

// enrichMatchesWithRoles adds role information to a list of matches
func (s *Service) enrichMatchesWithRoles(ctx context.Context, matches []Match) ([]MatchWithRoles, error) {
	result := make([]MatchWithRoles, 0, len(matches))

	for _, match := range matches {
		roles, err := s.repo.GetRoles(ctx, match.ID)
		if err != nil {
			// Log error but continue processing
			continue
		}

		matchWithRoles := MatchWithRoles{
			Match: match,
			Roles: roles,
		}

		// Calculate assignment status
		matchWithRoles.AssignmentStatus = s.calculateAssignmentStatus(roles)

		// Check for overdue acknowledgments
		matchWithRoles.HasOverdueAck = s.hasOverdueAcknowledgment(roles)

		result = append(result, matchWithRoles)
	}

	return result, nil
}

// calculateAssignmentStatus determines if a match is unassigned, partially assigned, or fully assigned
func (s *Service) calculateAssignmentStatus(roles []MatchRole) string {
	if len(roles) == 0 {
		return "unassigned"
	}

	assignedCount := 0
	for _, role := range roles {
		if role.AssignedRefereeID != nil {
			assignedCount++
		}
	}

	if assignedCount == 0 {
		return "unassigned"
	} else if assignedCount < len(roles) {
		return "partial"
	} else {
		return "full"
	}
}

// hasOverdueAcknowledgment checks if any role has an overdue acknowledgment
func (s *Service) hasOverdueAcknowledgment(roles []MatchRole) bool {
	for _, role := range roles {
		if role.AckOverdue {
			return true
		}
	}
	return false
}

// ArchiveMatch archives a match (marks as completed and removes from active views)
func (s *Service) ArchiveMatch(ctx context.Context, matchID int64, userID int64) error {
	// Verify match exists and is active
	match, err := s.repo.FindByID(ctx, matchID)
	if err != nil {
		return errors.NewInternal("Failed to find match", err)
	}
	if match == nil {
		return errors.NewNotFound("Match")
	}

	// Check if already archived
	if match.Archived {
		return errors.NewBadRequest("Match is already archived")
	}

	// Archive the match
	err = s.repo.Archive(ctx, matchID, userID)
	if err != nil {
		return errors.NewInternal("Failed to archive match", err)
	}

	return nil
}

// UnarchiveMatch unarchives a match (for administrative purposes)
func (s *Service) UnarchiveMatch(ctx context.Context, matchID int64) error {
	// Verify match exists
	match, err := s.repo.FindByID(ctx, matchID)
	if err != nil {
		return errors.NewInternal("Failed to find match", err)
	}
	if match == nil {
		return errors.NewNotFound("Match")
	}

	// Check if actually archived
	if !match.Archived {
		return errors.NewBadRequest("Match is not archived")
	}

	// Unarchive the match
	err = s.repo.Unarchive(ctx, matchID)
	if err != nil {
		return errors.NewInternal("Failed to unarchive match", err)
	}

	return nil
}
