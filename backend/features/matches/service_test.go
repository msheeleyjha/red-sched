package matches

import (
	"bytes"
	"context"
	"mime/multipart"
	"testing"
)

// mockRepository is a mock implementation of RepositoryInterface for testing
type mockRepository struct {
	createFunc         func(ctx context.Context, match *Match) (*Match, error)
	findByIDFunc       func(ctx context.Context, id int64) (*Match, error)
	listFunc           func(ctx context.Context) ([]Match, error)
	updateFunc         func(ctx context.Context, id int64, updates map[string]interface{}) (*Match, error)
	createRoleFunc     func(ctx context.Context, matchID int64, roleType string) error
	getRolesFunc       func(ctx context.Context, matchID int64) ([]MatchRole, error)
	deleteRolesFunc    func(ctx context.Context, matchID int64, roleTypes []string) error
	roleExistsFunc     func(ctx context.Context, matchID int64, roleType string) (bool, error)
	getCurrentRolesFunc func(ctx context.Context, matchID int64) ([]string, error)
	matchExistsFunc    func(ctx context.Context, matchID int64) (bool, error)
	logEditFunc        func(ctx context.Context, matchID int64, actorID int64, changeDescription string) error
}

func (m *mockRepository) Create(ctx context.Context, match *Match) (*Match, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, match)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (*Match, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) List(ctx context.Context) ([]Match, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []Match{}, nil
}

func (m *mockRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*Match, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, updates)
	}
	return nil, nil
}

func (m *mockRepository) CreateRole(ctx context.Context, matchID int64, roleType string) error {
	if m.createRoleFunc != nil {
		return m.createRoleFunc(ctx, matchID, roleType)
	}
	return nil
}

func (m *mockRepository) GetRoles(ctx context.Context, matchID int64) ([]MatchRole, error) {
	if m.getRolesFunc != nil {
		return m.getRolesFunc(ctx, matchID)
	}
	return []MatchRole{}, nil
}

func (m *mockRepository) DeleteRoles(ctx context.Context, matchID int64, roleTypes []string) error {
	if m.deleteRolesFunc != nil {
		return m.deleteRolesFunc(ctx, matchID, roleTypes)
	}
	return nil
}

func (m *mockRepository) RoleExists(ctx context.Context, matchID int64, roleType string) (bool, error) {
	if m.roleExistsFunc != nil {
		return m.roleExistsFunc(ctx, matchID, roleType)
	}
	return false, nil
}

func (m *mockRepository) GetCurrentRoles(ctx context.Context, matchID int64) ([]string, error) {
	if m.getCurrentRolesFunc != nil {
		return m.getCurrentRolesFunc(ctx, matchID)
	}
	return []string{}, nil
}

func (m *mockRepository) MatchExists(ctx context.Context, matchID int64) (bool, error) {
	if m.matchExistsFunc != nil {
		return m.matchExistsFunc(ctx, matchID)
	}
	return false, nil
}

func (m *mockRepository) LogEdit(ctx context.Context, matchID int64, actorID int64, changeDescription string) error {
	if m.logEditFunc != nil {
		return m.logEditFunc(ctx, matchID, actorID, changeDescription)
	}
	return nil
}

func (m *mockRepository) FindByReferenceID(ctx context.Context, referenceID string) (*Match, error) {
	return nil, nil
}

func (m *mockRepository) ListActive(ctx context.Context, params *MatchListParams) ([]Match, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return []Match{}, nil
}

func (m *mockRepository) CountActive(ctx context.Context, params *MatchListParams) (int, error) {
	return 0, nil
}

func (m *mockRepository) ListArchived(ctx context.Context) ([]Match, error) {
	return []Match{}, nil
}

func (m *mockRepository) Archive(ctx context.Context, matchID int64, archivedBy int64) error {
	return nil
}

func (m *mockRepository) Unarchive(ctx context.Context, matchID int64) error {
	return nil
}

func (m *mockRepository) IsReferenceIDExcluded(ctx context.Context, referenceID string) (bool, error) {
	return false, nil
}

func (m *mockRepository) AddExcludedReferenceID(ctx context.Context, referenceID string, reason *string, excludedBy int64) error {
	return nil
}

func (m *mockRepository) RemoveExcludedReferenceID(ctx context.Context, referenceID string) error {
	return nil
}

func (m *mockRepository) ListExcludedReferenceIDs(ctx context.Context) ([]ExcludedReferenceID, error) {
	return []ExcludedReferenceID{}, nil
}

func stringPtr(s string) *string {
	return &s
}

func TestService_ParseCSV(t *testing.T) {
	ctx := context.Background()
	service := NewService(&mockRepository{})

	t.Run("successfully parses valid CSV", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location,description,reference_id
Spring League,Under 12 Girls - Falcons,2027-05-15,10:00,11:30,Field A,Championship game,REF-001`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("Expected 1 row, got %d", len(result.Rows))
		}

		row := result.Rows[0]
		if row.EventName != "Spring League" {
			t.Errorf("Expected event_name 'Spring League', got '%s'", row.EventName)
		}
		if row.AgeGroup == nil || *row.AgeGroup != "U12" {
			t.Errorf("Expected age group U12, got %v", row.AgeGroup)
		}
		if row.Error != nil {
			t.Errorf("Expected no error for valid row, got %v", *row.Error)
		}
	})

	t.Run("rejects non-CSV files", func(t *testing.T) {
		_, err := service.ParseCSV(context.Background(), nil, "file.txt")

		if err == nil {
			t.Fatal("Expected error for non-CSV file, got nil")
		}
	})

	t.Run("detects missing required columns", func(t *testing.T) {
		csvContent := `event_name,team_name
Spring League,Under 12 Girls`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		_, err := service.ParseCSV(ctx, file, filename)

		if err == nil {
			t.Fatal("Expected error for missing columns, got nil")
		}
	})

	t.Run("detects rows with missing required fields", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location
Spring League,Under 12 Girls,2027-05-15,10:00,11:30,
Spring League,Under 12 Boys,2027-05-15,12:00,13:30,Field B`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error for parsing, got %v", err)
		}
		if len(result.Rows) != 2 {
			t.Fatalf("Expected 2 rows, got %d", len(result.Rows))
		}

		// First row should have error (missing location)
		if result.Rows[0].Error == nil {
			t.Error("Expected error for row with missing location")
		}

		// Second row should be valid
		if result.Rows[1].Error != nil {
			t.Errorf("Expected no error for valid row, got %v", *result.Rows[1].Error)
		}
	})

	t.Run("detects rows without age group", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location
Spring League,Falcons,2027-05-15,10:00,11:30,Field A`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error for parsing, got %v", err)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("Expected 1 row, got %d", len(result.Rows))
		}

		// Should have error for unrecognized age group
		if result.Rows[0].Error == nil {
			t.Error("Expected error for row without age group")
		}
	})

	t.Run("detects duplicate reference_id", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location,reference_id
Spring League,Under 12 Girls,2027-05-15,10:00,11:30,Field A,REF-001
Spring League,Under 12 Boys,2027-05-15,12:00,13:30,Field B,REF-001`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error for parsing, got %v", err)
		}
		if len(result.Duplicates) != 1 {
			t.Fatalf("Expected 1 duplicate group, got %d", len(result.Duplicates))
		}

		dup := result.Duplicates[0]
		if dup.Signal != "reference_id" {
			t.Errorf("Expected signal 'reference_id', got '%s'", dup.Signal)
		}
		if len(dup.Matches) != 2 {
			t.Errorf("Expected 2 matches in duplicate group, got %d", len(dup.Matches))
		}
	})

	t.Run("detects same-match duplicates with different reference_ids", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location,reference_id
Spring League,Under 12 Girls,2027-05-15,10:00,11:30,Field A,REF-001
Spring League,Under 12 Girls,2027-05-15,10:00,11:30,Field B,REF-002`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error (duplicates should not cause rejection), got %v", err)
		}
		if result == nil {
			t.Fatal("Expected preview response, got nil")
		}

		sameMatchCount := 0
		for _, dup := range result.Duplicates {
			if dup.Signal == "same_match" {
				sameMatchCount++
				if len(dup.Matches) != 2 {
					t.Errorf("Expected 2 matches in same_match group, got %d", len(dup.Matches))
				}
			}
		}
		if sameMatchCount != 1 {
			t.Errorf("Expected 1 same_match duplicate group, got %d", sameMatchCount)
		}
	})

	t.Run("detects same-event duplicates with same gender", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location,reference_id
Game Day,Under 12 Girls - Falcons,2027-05-15,10:00,11:30,Field A,REF-001
Game Day,Under 12 Girls - Hawks,2027-05-15,10:00,11:30,Field A,REF-002`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		sameEventCount := 0
		for _, dup := range result.Duplicates {
			if dup.Signal == "same_event" {
				sameEventCount++
			}
		}
		if sameEventCount != 1 {
			t.Errorf("Expected 1 same_event duplicate group, got %d", sameEventCount)
		}
	})

	t.Run("does not flag same-event with different genders as duplicates", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location,reference_id
Game Day,Under 6 Boys - Falcons,2027-05-15,10:00,11:30,Field A,REF-001
Game Day,Under 6 Girls - Hawks,2027-05-15,10:00,11:30,Field A,REF-002`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		for _, dup := range result.Duplicates {
			if dup.Signal == "same_event" {
				t.Error("Boys and Girls at same event should not be flagged as duplicates")
			}
		}
	})

	t.Run("does not double-flag rows already caught by signal A or B", func(t *testing.T) {
		csvContent := `event_name,team_name,start_date,start_time,end_time,location,reference_id
Spring League,Under 12 Girls,2027-05-15,10:00,11:30,Field A,REF-001
Spring League,Under 12 Girls,2027-05-15,10:00,11:30,Field A,REF-001`

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should be caught by Signal A (same reference_id), not Signal C
		for _, dup := range result.Duplicates {
			if dup.Signal == "same_event" {
				t.Error("Rows already flagged by Signal A should not be double-flagged by Signal C")
			}
		}
	})
}

func TestService_ParseCSV_StripEmojis(t *testing.T) {
	ctx := context.Background()
	service := NewService(&mockRepository{})

	t.Run("strips emojis from event_name", func(t *testing.T) {
		csvContent := "event_name,team_name,start_date,start_time,end_time,location\n\u26BD Spring League \U0001F3C6,Under 12 Girls,2027-05-15,10:00,11:30,Field A"

		file, filename := createMultipartFile(csvContent)
		defer file.Close()

		result, err := service.ParseCSV(ctx, file, filename)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(result.Rows) != 1 {
			t.Fatalf("Expected 1 row, got %d", len(result.Rows))
		}
		if result.Rows[0].EventName != "Spring League" {
			t.Errorf("Expected 'Spring League' after emoji stripping, got '%s'", result.Rows[0].EventName)
		}
	})
}

func TestService_IsPracticeMatch(t *testing.T) {
	service := NewService(&mockRepository{})

	tests := []struct {
		eventName string
		expected  bool
	}{
		{"Spring Practice", true},
		{"PRACTICE SESSION", true},
		{"Team Training", true},
		{"training session", true},
		{"Spring League", false},
		{"Mini Matches", false},
	}

	for _, test := range tests {
		t.Run(test.eventName, func(t *testing.T) {
			result := service.isPracticeMatch(test.eventName)
			if result != test.expected {
				t.Errorf("isPracticeMatch(%q) = %v, want %v", test.eventName, result, test.expected)
			}
		})
	}
}

func TestService_MatchesCustomExcludeTerm(t *testing.T) {
	service := NewService(&mockRepository{})

	t.Run("matches term in event name", func(t *testing.T) {
		if !service.matchesCustomExcludeTerm("Mini Matches", []string{"Mini"}) {
			t.Error("Expected 'Mini Matches' to match term 'Mini'")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		if !service.matchesCustomExcludeTerm("SCRIMMAGE Game", []string{"scrimmage"}) {
			t.Error("Expected case-insensitive match")
		}
	})

	t.Run("no match returns false", func(t *testing.T) {
		if service.matchesCustomExcludeTerm("Spring League", []string{"Mini", "Scrimmage"}) {
			t.Error("Expected no match for 'Spring League'")
		}
	})

	t.Run("ignores empty terms", func(t *testing.T) {
		if service.matchesCustomExcludeTerm("Spring League", []string{"", "  "}) {
			t.Error("Expected empty/whitespace terms to be ignored")
		}
	})
}

func TestService_CreateRoleSlotsForMatch(t *testing.T) {
	ctx := context.Background()

	t.Run("creates center only for U8 match", func(t *testing.T) {
		rolesCreated := []string{}
		mockRepo := &mockRepository{
			createRoleFunc: func(ctx context.Context, matchID int64, roleType string) error {
				rolesCreated = append(rolesCreated, roleType)
				return nil
			},
		}

		service := NewService(mockRepo)
		err := service.CreateRoleSlotsForMatch(ctx, 1, "U8")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(rolesCreated) != 1 {
			t.Fatalf("Expected 1 role, got %d", len(rolesCreated))
		}
		if rolesCreated[0] != "center" {
			t.Errorf("Expected center role, got %s", rolesCreated[0])
		}
	})

	t.Run("creates center only for U10 match", func(t *testing.T) {
		rolesCreated := []string{}
		mockRepo := &mockRepository{
			createRoleFunc: func(ctx context.Context, matchID int64, roleType string) error {
				rolesCreated = append(rolesCreated, roleType)
				return nil
			},
		}

		service := NewService(mockRepo)
		err := service.CreateRoleSlotsForMatch(ctx, 1, "U10")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(rolesCreated) != 1 {
			t.Fatalf("Expected 1 role, got %d", len(rolesCreated))
		}
		if rolesCreated[0] != "center" {
			t.Errorf("Expected center role, got %s", rolesCreated[0])
		}
	})

	t.Run("creates center and 2 assistants for U12 match", func(t *testing.T) {
		rolesCreated := []string{}
		mockRepo := &mockRepository{
			createRoleFunc: func(ctx context.Context, matchID int64, roleType string) error {
				rolesCreated = append(rolesCreated, roleType)
				return nil
			},
		}

		service := NewService(mockRepo)
		err := service.CreateRoleSlotsForMatch(ctx, 1, "U12")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(rolesCreated) != 3 {
			t.Fatalf("Expected 3 roles, got %d", len(rolesCreated))
		}

		expectedRoles := map[string]bool{"center": false, "assistant_1": false, "assistant_2": false}
		for _, role := range rolesCreated {
			expectedRoles[role] = true
		}
		for role, found := range expectedRoles {
			if !found {
				t.Errorf("Expected role %s not created", role)
			}
		}
	})

	t.Run("creates center and 2 assistants for U14 match", func(t *testing.T) {
		rolesCreated := []string{}
		mockRepo := &mockRepository{
			createRoleFunc: func(ctx context.Context, matchID int64, roleType string) error {
				rolesCreated = append(rolesCreated, roleType)
				return nil
			},
		}

		service := NewService(mockRepo)
		err := service.CreateRoleSlotsForMatch(ctx, 1, "U14")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(rolesCreated) != 3 {
			t.Fatalf("Expected 3 roles, got %d", len(rolesCreated))
		}
	})

	t.Run("returns error for invalid age group", func(t *testing.T) {
		service := NewService(&mockRepository{})
		err := service.CreateRoleSlotsForMatch(ctx, 1, "Invalid")

		if err == nil {
			t.Fatal("Expected error for invalid age group, got nil")
		}
	})
}

func TestService_ListMatches(t *testing.T) {
	ctx := context.Background()

	t.Run("returns matches with roles and status", func(t *testing.T) {
		ageGroup := "U12"
		matches := []Match{
			{ID: 1, EventName: "Spring League", TeamName: "Under 12 Girls", AgeGroup: &ageGroup},
		}

		mockRepo := &mockRepository{
			listFunc: func(ctx context.Context) ([]Match, error) {
				return matches, nil
			},
			getRolesFunc: func(ctx context.Context, matchID int64) ([]MatchRole, error) {
				if matchID == 1 {
					refID := int64(10)
					return []MatchRole{
						{ID: 1, MatchID: 1, RoleType: "center", AssignedRefereeID: &refID},
						{ID: 2, MatchID: 1, RoleType: "assistant_1"},
						{ID: 3, MatchID: 1, RoleType: "assistant_2"},
					}, nil
				}
				return []MatchRole{}, nil
			},
		}

		service := NewService(mockRepo)
		result, err := service.ListMatches(ctx, nil)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(result.Matches) != 1 {
			t.Fatalf("Expected 1 match, got %d", len(result.Matches))
		}

		match := result.Matches[0]
		if len(match.Roles) != 3 {
			t.Errorf("Expected 3 roles, got %d", len(match.Roles))
		}
		if match.AssignmentStatus != "partial" {
			t.Errorf("Expected status 'partial', got '%s'", match.AssignmentStatus)
		}
	})

	t.Run("calculates correct status for U10 match", func(t *testing.T) {
		ageGroup := "U10"
		matches := []Match{
			{ID: 1, EventName: "Spring League", TeamName: "Under 10 Boys", AgeGroup: &ageGroup},
		}

		mockRepo := &mockRepository{
			listFunc: func(ctx context.Context) ([]Match, error) {
				return matches, nil
			},
			getRolesFunc: func(ctx context.Context, matchID int64) ([]MatchRole, error) {
				if matchID == 1 {
					refID := int64(10)
					return []MatchRole{
						{ID: 1, MatchID: 1, RoleType: "center", AssignedRefereeID: &refID},
						{ID: 2, MatchID: 1, RoleType: "assistant_1"},
						{ID: 3, MatchID: 1, RoleType: "assistant_2"},
					}, nil
				}
				return []MatchRole{}, nil
			},
		}

		service := NewService(mockRepo)
		result, err := service.ListMatches(ctx, nil)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(result.Matches) != 1 {
			t.Fatalf("Expected 1 match, got %d", len(result.Matches))
		}

		match := result.Matches[0]
		if match.AssignmentStatus != "full" {
			t.Errorf("Expected status 'full' for U10 with center assigned, got '%s'", match.AssignmentStatus)
		}
	})
}

func TestService_UpdateMatch(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates match", func(t *testing.T) {
		ageGroup := "U12"
		currentMatch := &Match{
			ID: 1, EventName: "Spring League", TeamName: "Under 12 Girls", AgeGroup: &ageGroup,
		}

		updatedCalled := false
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*Match, error) {
				if id == 1 {
					return currentMatch, nil
				}
				return nil, nil
			},
			updateFunc: func(ctx context.Context, id int64, updates map[string]interface{}) (*Match, error) {
				updatedCalled = true
				currentMatch.EventName = updates["event_name"].(string)
				return currentMatch, nil
			},
			getRolesFunc: func(ctx context.Context, matchID int64) ([]MatchRole, error) {
				return []MatchRole{}, nil
			},
			logEditFunc: func(ctx context.Context, matchID int64, actorID int64, changeDescription string) error {
				return nil
			},
		}

		service := NewService(mockRepo)
		newEventName := "Summer League"
		req := &MatchUpdateRequest{
			EventName: &newEventName,
		}

		result, err := service.UpdateMatch(ctx, 1, req, 10)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !updatedCalled {
			t.Error("Expected update to be called")
		}
		if result.EventName != "Summer League" {
			t.Errorf("Expected event name 'Summer League', got '%s'", result.EventName)
		}
	})

	t.Run("returns error for invalid match ID", func(t *testing.T) {
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*Match, error) {
				return nil, nil // Match not found
			},
		}

		service := NewService(mockRepo)
		req := &MatchUpdateRequest{EventName: stringPtr("New Name")}

		_, err := service.UpdateMatch(ctx, 999, req, 10)

		if err == nil {
			t.Fatal("Expected error for invalid match ID, got nil")
		}
	})

	t.Run("returns error for invalid date format", func(t *testing.T) {
		ageGroup := "U12"
		currentMatch := &Match{ID: 1, AgeGroup: &ageGroup}

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*Match, error) {
				return currentMatch, nil
			},
		}

		service := NewService(mockRepo)
		req := &MatchUpdateRequest{
			MatchDate: stringPtr("invalid-date"),
		}

		_, err := service.UpdateMatch(ctx, 1, req, 10)

		if err == nil {
			t.Fatal("Expected error for invalid date format, got nil")
		}
	})

	t.Run("returns error for invalid status", func(t *testing.T) {
		ageGroup := "U12"
		currentMatch := &Match{ID: 1, AgeGroup: &ageGroup}

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*Match, error) {
				return currentMatch, nil
			},
		}

		service := NewService(mockRepo)
		req := &MatchUpdateRequest{
			Status: stringPtr("invalid"),
		}

		_, err := service.UpdateMatch(ctx, 1, req, 10)

		if err == nil {
			t.Fatal("Expected error for invalid status, got nil")
		}
	})

	t.Run("returns error for no updates", func(t *testing.T) {
		ageGroup := "U12"
		currentMatch := &Match{ID: 1, AgeGroup: &ageGroup}

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*Match, error) {
				return currentMatch, nil
			},
		}

		service := NewService(mockRepo)
		req := &MatchUpdateRequest{}

		_, err := service.UpdateMatch(ctx, 1, req, 10)

		if err == nil {
			t.Fatal("Expected error for no updates, got nil")
		}
	})
}

func TestService_AddRoleSlot(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully adds assistant role", func(t *testing.T) {
		roleCreated := false
		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			roleExistsFunc: func(ctx context.Context, matchID int64, roleType string) (bool, error) {
				return false, nil
			},
			createRoleFunc: func(ctx context.Context, matchID int64, roleType string) error {
				roleCreated = true
				return nil
			},
		}

		service := NewService(mockRepo)
		err := service.AddRoleSlot(ctx, 1, "assistant_1")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !roleCreated {
			t.Error("Expected role to be created")
		}
	})

	t.Run("rejects center role", func(t *testing.T) {
		service := NewService(&mockRepository{})
		err := service.AddRoleSlot(ctx, 1, "center")

		if err == nil {
			t.Fatal("Expected error for center role, got nil")
		}
	})

	t.Run("returns error for non-existent match", func(t *testing.T) {
		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return false, nil
			},
		}

		service := NewService(mockRepo)
		err := service.AddRoleSlot(ctx, 999, "assistant_1")

		if err == nil {
			t.Fatal("Expected error for non-existent match, got nil")
		}
	})

	t.Run("returns error for existing role", func(t *testing.T) {
		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			roleExistsFunc: func(ctx context.Context, matchID int64, roleType string) (bool, error) {
				return true, nil
			},
		}

		service := NewService(mockRepo)
		err := service.AddRoleSlot(ctx, 1, "assistant_1")

		if err == nil {
			t.Fatal("Expected error for existing role, got nil")
		}
	})
}

func TestService_ImportMatches(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully imports valid matches", func(t *testing.T) {
		createCount := 0
		roleCreateCount := 0

		mockRepo := &mockRepository{
			createFunc: func(ctx context.Context, match *Match) (*Match, error) {
				createCount++
				match.ID = int64(createCount)
				return match, nil
			},
			createRoleFunc: func(ctx context.Context, matchID int64, roleType string) error {
				roleCreateCount++
				return nil
			},
		}

		service := NewService(mockRepo)
		ageGroup := "U12"
		req := &ImportConfirmRequest{
			Rows: []CSVRow{
				{
					EventName:   "Spring League",
					TeamName:    "Under 12 Girls",
					StartDate:   "2027-05-15",
					StartTime:   "10:00",
					EndTime:     "11:30",
					Location:    "Field A",
					Description: "Game 1",
					ReferenceID: "REF-001",
					AgeGroup:    &ageGroup,
					RowNumber:   2,
				},
			},
		}

		result, err := service.ImportMatches(ctx, req, 1)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Imported != 1 {
			t.Errorf("Expected 1 imported, got %d", result.Imported)
		}
		if result.Skipped != 0 {
			t.Errorf("Expected 0 skipped, got %d", result.Skipped)
		}
		if createCount != 1 {
			t.Errorf("Expected 1 match created, got %d", createCount)
		}
		// U12 should create 3 roles (center + 2 ARs)
		if roleCreateCount != 3 {
			t.Errorf("Expected 3 roles created for U12, got %d", roleCreateCount)
		}
	})

	t.Run("skips rows with errors", func(t *testing.T) {
		mockRepo := &mockRepository{}
		service := NewService(mockRepo)

		errorMsg := "Missing required fields"
		req := &ImportConfirmRequest{
			Rows: []CSVRow{
				{
					EventName: "Spring League",
					Error:     &errorMsg,
					RowNumber: 2,
				},
			},
		}

		result, err := service.ImportMatches(ctx, req, 1)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Imported != 0 {
			t.Errorf("Expected 0 imported, got %d", result.Imported)
		}
		if result.Skipped != 1 {
			t.Errorf("Expected 1 skipped, got %d", result.Skipped)
		}
	})

	t.Run("handles invalid date format", func(t *testing.T) {
		mockRepo := &mockRepository{}
		service := NewService(mockRepo)

		ageGroup := "U12"
		req := &ImportConfirmRequest{
			Rows: []CSVRow{
				{
					EventName:   "Spring League",
					TeamName:    "Under 12 Girls",
					StartDate:   "invalid-date",
					StartTime:   "10:00",
					EndTime:     "11:30",
					Location:    "Field A",
					AgeGroup:    &ageGroup,
					RowNumber:   2,
				},
			},
		}

		result, err := service.ImportMatches(ctx, req, 1)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Skipped != 1 {
			t.Errorf("Expected 1 skipped, got %d", result.Skipped)
		}
		if len(result.Errors) == 0 {
			t.Error("Expected error message for invalid date")
		}
	})
}

func TestExtractGender(t *testing.T) {
	tests := []struct {
		teamName string
		expected string
	}{
		{"Under 12 Girls - Falcons", "girls"},
		{"Under 6 Boys", "boys"},
		{"Under 10 BOYS - Hawks", "boys"},
		{"under 8 girls", "girls"},
		{"Senior Team", ""},
		{"U12 Girls", ""},
	}

	for _, test := range tests {
		t.Run(test.teamName, func(t *testing.T) {
			result := extractGender(test.teamName)
			if result != test.expected {
				t.Errorf("extractGender(%q) = %q, want %q", test.teamName, result, test.expected)
			}
		})
	}
}

func TestExtractAgeGroup(t *testing.T) {
	tests := []struct {
		teamName string
		expected *string
	}{
		{"Under 12 Girls - Falcons", stringPtr("U12")},
		{"Under 8 Boys", stringPtr("U8")},
		{"UNDER 10 Mixed", stringPtr("U10")},
		{"under 14 Elite", stringPtr("U14")},
		{"Senior Team", nil},
		{"U12 Girls", nil}, // Doesn't match "Under N" pattern
	}

	for _, test := range tests {
		t.Run(test.teamName, func(t *testing.T) {
			result := extractAgeGroup(test.teamName)

			if test.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", *result)
			} else if test.expected != nil && result == nil {
				t.Errorf("Expected %v, got nil", *test.expected)
			} else if test.expected != nil && result != nil && *test.expected != *result {
				t.Errorf("Expected %v, got %v", *test.expected, *result)
			}
		})
	}
}

func TestGetAgeGroupInt(t *testing.T) {
	tests := []struct {
		name     string
		ageGroup *string
		expected int
		hasError bool
	}{
		{"U12", stringPtr("U12"), 12, false},
		{"U8", stringPtr("U8"), 8, false},
		{"U14", stringPtr("U14"), 14, false},
		{"nil", nil, 0, true},
		{"invalid format", stringPtr("Invalid"), 0, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := GetAgeGroupInt(test.ageGroup)

			if test.hasError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if result != test.expected {
					t.Errorf("Expected %d, got %d", test.expected, result)
				}
			}
		})
	}
}

// Helper function to create a multipart file from string content
func createMultipartFile(content string) (multipart.File, string) {
	buffer := bytes.NewBufferString(content)
	return &mockMultipartFile{reader: buffer}, "test.csv"
}

// mockMultipartFile implements multipart.File interface
type mockMultipartFile struct {
	reader *bytes.Buffer
}

func (m *mockMultipartFile) Read(p []byte) (n int, err error) {
	return m.reader.Read(p)
}

func (m *mockMultipartFile) ReadAt(p []byte, off int64) (n int, err error) {
	return 0, nil
}

func (m *mockMultipartFile) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (m *mockMultipartFile) Close() error {
	return nil
}
