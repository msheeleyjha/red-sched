package assignments

import (
	"context"
	"testing"
	"time"
)

// mockRepository is a mock implementation of RepositoryInterface for testing
type mockRepository struct {
	matchExistsFunc                  func(ctx context.Context, matchID int64) (bool, error)
	getMatchTimeWindowFunc           func(ctx context.Context, matchID int64) (*MatchTimeWindow, error)
	getRoleSlotFunc                  func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error)
	updateRoleAssignmentFunc         func(ctx context.Context, roleID int64, refereeID *int64) error
	refereeExistsFunc                func(ctx context.Context, refereeID int64) (bool, error)
	getRefereeExistingRoleOnMatchFunc func(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error)
	findConflictingAssignmentsFunc   func(ctx context.Context, refereeID int64, matchID int64, startTime time.Time, endTime time.Time) ([]ConflictMatch, error)
	logAssignmentFunc                func(ctx context.Context, history *AssignmentHistory) error
}

func (m *mockRepository) MatchExists(ctx context.Context, matchID int64) (bool, error) {
	if m.matchExistsFunc != nil {
		return m.matchExistsFunc(ctx, matchID)
	}
	return false, nil
}

func (m *mockRepository) GetMatchTimeWindow(ctx context.Context, matchID int64) (*MatchTimeWindow, error) {
	if m.getMatchTimeWindowFunc != nil {
		return m.getMatchTimeWindowFunc(ctx, matchID)
	}
	return nil, nil
}

func (m *mockRepository) GetRoleSlot(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
	if m.getRoleSlotFunc != nil {
		return m.getRoleSlotFunc(ctx, matchID, roleType)
	}
	return nil, nil
}

func (m *mockRepository) UpdateRoleAssignment(ctx context.Context, roleID int64, refereeID *int64) error {
	if m.updateRoleAssignmentFunc != nil {
		return m.updateRoleAssignmentFunc(ctx, roleID, refereeID)
	}
	return nil
}

func (m *mockRepository) RefereeExists(ctx context.Context, refereeID int64) (bool, error) {
	if m.refereeExistsFunc != nil {
		return m.refereeExistsFunc(ctx, refereeID)
	}
	return false, nil
}

func (m *mockRepository) GetRefereeExistingRoleOnMatch(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error) {
	if m.getRefereeExistingRoleOnMatchFunc != nil {
		return m.getRefereeExistingRoleOnMatchFunc(ctx, matchID, refereeID, excludeRoleType)
	}
	return nil, nil
}

func (m *mockRepository) FindConflictingAssignments(ctx context.Context, refereeID int64, matchID int64, startTime time.Time, endTime time.Time) ([]ConflictMatch, error) {
	if m.findConflictingAssignmentsFunc != nil {
		return m.findConflictingAssignmentsFunc(ctx, refereeID, matchID, startTime, endTime)
	}
	return []ConflictMatch{}, nil
}

func (m *mockRepository) LogAssignment(ctx context.Context, history *AssignmentHistory) error {
	if m.logAssignmentFunc != nil {
		return m.logAssignmentFunc(ctx, history)
	}
	return nil
}

func (m *mockRepository) GetRefereeMatchHistory(ctx context.Context, refereeID int64) ([]RefereeHistoryMatch, error) {
	return []RefereeHistoryMatch{}, nil
}

func (m *mockRepository) MarkAssignmentAsViewed(ctx context.Context, matchID int64, refereeID int64) error {
	return nil
}

func (m *mockRepository) ResetViewedStatusForMatch(ctx context.Context, matchID int64) error {
	return nil
}

func int64Ptr(i int64) *int64 {
	return &i
}

func stringPtr(s string) *string {
	return &s
}

func TestService_AssignReferee(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully assigns referee to empty slot", func(t *testing.T) {
		updatedCalled := false
		loggedAction := ""

		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			getRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
				return &RoleSlot{
					ID:                1,
					MatchID:           matchID,
					RoleType:          roleType,
					AssignedRefereeID: nil, // Empty slot
				}, nil
			},
			refereeExistsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
				return true, nil
			},
			getRefereeExistingRoleOnMatchFunc: func(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error) {
				return nil, nil // No existing role
			},
			updateRoleAssignmentFunc: func(ctx context.Context, roleID int64, refereeID *int64) error {
				updatedCalled = true
				return nil
			},
			logAssignmentFunc: func(ctx context.Context, history *AssignmentHistory) error {
				loggedAction = history.Action
				return nil
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: int64Ptr(10)}

		result, err := service.AssignReferee(ctx, 1, "center", req, 5)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !updatedCalled {
			t.Error("Expected update to be called")
		}
		if result.Action != "assigned" {
			t.Errorf("Expected action 'assigned', got '%s'", result.Action)
		}
		if loggedAction != "assigned" {
			t.Errorf("Expected logged action 'assigned', got '%s'", loggedAction)
		}
	})

	t.Run("successfully reassigns referee in occupied slot", func(t *testing.T) {
		oldRefereeID := int64(5)
		newRefereeID := int64(10)
		loggedAction := ""

		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			getRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
				return &RoleSlot{
					ID:                1,
					MatchID:           matchID,
					RoleType:          roleType,
					AssignedRefereeID: &oldRefereeID, // Already occupied
				}, nil
			},
			refereeExistsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
				return true, nil
			},
			getRefereeExistingRoleOnMatchFunc: func(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error) {
				return nil, nil
			},
			updateRoleAssignmentFunc: func(ctx context.Context, roleID int64, refereeID *int64) error {
				return nil
			},
			logAssignmentFunc: func(ctx context.Context, history *AssignmentHistory) error {
				loggedAction = history.Action
				if history.OldRefereeID == nil || *history.OldRefereeID != oldRefereeID {
					t.Error("Old referee ID not logged correctly")
				}
				if history.NewRefereeID == nil || *history.NewRefereeID != newRefereeID {
					t.Error("New referee ID not logged correctly")
				}
				return nil
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: &newRefereeID}

		result, err := service.AssignReferee(ctx, 1, "center", req, 5)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Action != "reassigned" {
			t.Errorf("Expected action 'reassigned', got '%s'", result.Action)
		}
		if loggedAction != "reassigned" {
			t.Errorf("Expected logged action 'reassigned', got '%s'", loggedAction)
		}
	})

	t.Run("successfully removes referee assignment", func(t *testing.T) {
		oldRefereeID := int64(5)
		loggedAction := ""

		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			getRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
				return &RoleSlot{
					ID:                1,
					MatchID:           matchID,
					RoleType:          roleType,
					AssignedRefereeID: &oldRefereeID,
				}, nil
			},
			updateRoleAssignmentFunc: func(ctx context.Context, roleID int64, refereeID *int64) error {
				if refereeID != nil {
					t.Error("Expected nil referee ID for removal")
				}
				return nil
			},
			logAssignmentFunc: func(ctx context.Context, history *AssignmentHistory) error {
				loggedAction = history.Action
				if history.NewRefereeID != nil {
					t.Error("Expected nil new referee ID for removal")
				}
				return nil
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: nil} // Remove assignment

		result, err := service.AssignReferee(ctx, 1, "center", req, 5)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.Action != "unassigned" {
			t.Errorf("Expected action 'unassigned', got '%s'", result.Action)
		}
		if loggedAction != "unassigned" {
			t.Errorf("Expected logged action 'unassigned', got '%s'", loggedAction)
		}
	})

	t.Run("returns error for invalid role type", func(t *testing.T) {
		service := NewService(&mockRepository{})
		req := &AssignmentRequest{RefereeID: int64Ptr(10)}

		_, err := service.AssignReferee(ctx, 1, "invalid_role", req, 5)

		if err == nil {
			t.Fatal("Expected error for invalid role type, got nil")
		}
	})

	t.Run("returns error for non-existent match", func(t *testing.T) {
		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return false, nil
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: int64Ptr(10)}

		_, err := service.AssignReferee(ctx, 999, "center", req, 5)

		if err == nil {
			t.Fatal("Expected error for non-existent match, got nil")
		}
	})

	t.Run("returns error for non-existent role slot", func(t *testing.T) {
		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			getRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
				return nil, nil // Role slot not found
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: int64Ptr(10)}

		_, err := service.AssignReferee(ctx, 1, "assistant_1", req, 5)

		if err == nil {
			t.Fatal("Expected error for non-existent role slot, got nil")
		}
	})

	t.Run("returns error for non-existent referee", func(t *testing.T) {
		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			getRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
				return &RoleSlot{ID: 1, MatchID: matchID, RoleType: roleType}, nil
			},
			refereeExistsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
				return false, nil // Referee not found
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: int64Ptr(999)}

		_, err := service.AssignReferee(ctx, 1, "center", req, 5)

		if err == nil {
			t.Fatal("Expected error for non-existent referee, got nil")
		}
	})

	t.Run("returns error when referee already has different role on same match", func(t *testing.T) {
		existingRole := "assistant_1"

		mockRepo := &mockRepository{
			matchExistsFunc: func(ctx context.Context, matchID int64) (bool, error) {
				return true, nil
			},
			getRoleSlotFunc: func(ctx context.Context, matchID int64, roleType string) (*RoleSlot, error) {
				return &RoleSlot{ID: 1, MatchID: matchID, RoleType: roleType}, nil
			},
			refereeExistsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
				return true, nil
			},
			getRefereeExistingRoleOnMatchFunc: func(ctx context.Context, matchID int64, refereeID int64, excludeRoleType string) (*string, error) {
				return &existingRole, nil // Already has assistant_1 role
			},
		}

		service := NewService(mockRepo)
		req := &AssignmentRequest{RefereeID: int64Ptr(10)}

		_, err := service.AssignReferee(ctx, 1, "center", req, 5)

		if err == nil {
			t.Fatal("Expected error for duplicate role, got nil")
		}
	})
}

func TestService_CheckConflicts(t *testing.T) {
	ctx := context.Background()

	t.Run("returns no conflicts when none exist", func(t *testing.T) {
		mockRepo := &mockRepository{
			getMatchTimeWindowFunc: func(ctx context.Context, matchID int64) (*MatchTimeWindow, error) {
				return &MatchTimeWindow{
					MatchID: matchID,
					Start:   time.Date(2027, 5, 15, 10, 0, 0, 0, time.UTC),
					End:     time.Date(2027, 5, 15, 11, 30, 0, 0, time.UTC),
				}, nil
			},
			findConflictingAssignmentsFunc: func(ctx context.Context, refereeID int64, matchID int64, startTime time.Time, endTime time.Time) ([]ConflictMatch, error) {
				return []ConflictMatch{}, nil
			},
		}

		service := NewService(mockRepo)
		result, err := service.CheckConflicts(ctx, 1, 10)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result.HasConflict {
			t.Error("Expected no conflicts")
		}
		if len(result.Conflicts) != 0 {
			t.Errorf("Expected 0 conflicts, got %d", len(result.Conflicts))
		}
	})

	t.Run("returns conflicts when they exist", func(t *testing.T) {
		mockRepo := &mockRepository{
			getMatchTimeWindowFunc: func(ctx context.Context, matchID int64) (*MatchTimeWindow, error) {
				return &MatchTimeWindow{
					MatchID: matchID,
					Start:   time.Date(2027, 5, 15, 10, 0, 0, 0, time.UTC),
					End:     time.Date(2027, 5, 15, 11, 30, 0, 0, time.UTC),
				}, nil
			},
			findConflictingAssignmentsFunc: func(ctx context.Context, refereeID int64, matchID int64, startTime time.Time, endTime time.Time) ([]ConflictMatch, error) {
				return []ConflictMatch{
					{
						MatchID:   2,
						EventName: "Other League",
						TeamName:  "Under 12 Boys",
						MatchDate: "2027-05-15",
						StartTime: "10:30",
						RoleType:  "center",
					},
				}, nil
			},
		}

		service := NewService(mockRepo)
		result, err := service.CheckConflicts(ctx, 1, 10)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !result.HasConflict {
			t.Error("Expected conflicts to be detected")
		}
		if len(result.Conflicts) != 1 {
			t.Errorf("Expected 1 conflict, got %d", len(result.Conflicts))
		}
		if result.Conflicts[0].MatchID != 2 {
			t.Errorf("Expected conflict match ID 2, got %d", result.Conflicts[0].MatchID)
		}
	})

	t.Run("returns error for non-existent match", func(t *testing.T) {
		mockRepo := &mockRepository{
			getMatchTimeWindowFunc: func(ctx context.Context, matchID int64) (*MatchTimeWindow, error) {
				return nil, nil // Match not found
			},
		}

		service := NewService(mockRepo)
		_, err := service.CheckConflicts(ctx, 999, 10)

		if err == nil {
			t.Fatal("Expected error for non-existent match, got nil")
		}
	})
}

func TestValidRoleTypes(t *testing.T) {
	validRoles := []string{"center", "assistant_1", "assistant_2"}
	invalidRoles := []string{"invalid", "referee", "center_referee", ""}

	for _, role := range validRoles {
		t.Run("valid_"+role, func(t *testing.T) {
			if !ValidRoleTypes[role] {
				t.Errorf("Expected %s to be valid", role)
			}
		})
	}

	for _, role := range invalidRoles {
		t.Run("invalid_"+role, func(t *testing.T) {
			if ValidRoleTypes[role] {
				t.Errorf("Expected %s to be invalid", role)
			}
		})
	}
}

func TestRoleTypeDisplayNames(t *testing.T) {
	expectedNames := map[string]string{
		"center":      "Center Referee",
		"assistant_1": "Assistant Referee 1",
		"assistant_2": "Assistant Referee 2",
	}

	for roleType, expectedName := range expectedNames {
		t.Run(roleType, func(t *testing.T) {
			actualName := RoleTypeDisplayNames[roleType]
			if actualName != expectedName {
				t.Errorf("Expected display name '%s', got '%s'", expectedName, actualName)
			}
		})
	}
}
