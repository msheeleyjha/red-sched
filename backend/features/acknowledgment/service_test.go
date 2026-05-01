package acknowledgment

import (
	"context"
	"errors"
	"testing"
	"time"

	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
	GetRefereeAssignmentRoleFunc func(ctx context.Context, matchID int64, refereeID int64) (*string, error)
	AcknowledgeAssignmentFunc    func(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error
}

func (m *mockRepository) GetRefereeAssignmentRole(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
	if m.GetRefereeAssignmentRoleFunc != nil {
		return m.GetRefereeAssignmentRoleFunc(ctx, matchID, refereeID)
	}
	return nil, errors.New("GetRefereeAssignmentRole not implemented")
}

func (m *mockRepository) AcknowledgeAssignment(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error {
	if m.AcknowledgeAssignmentFunc != nil {
		return m.AcknowledgeAssignmentFunc(ctx, matchID, refereeID, acknowledgedAt)
	}
	return errors.New("AcknowledgeAssignment not implemented")
}

func TestAcknowledgeAssignment_Success(t *testing.T) {
	roleType := "center"
	repo := &mockRepository{
		GetRefereeAssignmentRoleFunc: func(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
			if matchID == 1 && refereeID == 100 {
				return &roleType, nil
			}
			return nil, nil
		},
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error {
			return nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.AcknowledgeAssignment(ctx, 1, 100)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.AcknowledgedAt.IsZero() {
		t.Error("Expected acknowledged_at to be set")
	}
}

func TestAcknowledgeAssignment_NotAssigned(t *testing.T) {
	repo := &mockRepository{
		GetRefereeAssignmentRoleFunc: func(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
			return nil, nil // No assignment found
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.AcknowledgeAssignment(ctx, 1, 100)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	appErr, ok := err.(*appErrors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got: %T", err)
	}

	if appErr.StatusCode != 404 {
		t.Errorf("Expected 404 NotFound, got: %d", appErr.StatusCode)
	}

	if appErr.Message != "Assignment not found" {
		t.Errorf("Expected specific message, got: %s", appErr.Message)
	}
}

func TestAcknowledgeAssignment_GetRoleError(t *testing.T) {
	repo := &mockRepository{
		GetRefereeAssignmentRoleFunc: func(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.AcknowledgeAssignment(ctx, 1, 100)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	appErr, ok := err.(*appErrors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got: %T", err)
	}

	if appErr.StatusCode != 500 {
		t.Errorf("Expected 500 Internal, got: %d", appErr.StatusCode)
	}

	if appErr.Message != "Failed to verify assignment" {
		t.Errorf("Expected specific message, got: %s", appErr.Message)
	}
}

func TestAcknowledgeAssignment_AcknowledgeError(t *testing.T) {
	roleType := "center"
	repo := &mockRepository{
		GetRefereeAssignmentRoleFunc: func(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
			return &roleType, nil
		},
		AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error {
			return errors.New("update failed")
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.AcknowledgeAssignment(ctx, 1, 100)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if result != nil {
		t.Errorf("Expected nil result, got: %v", result)
	}

	appErr, ok := err.(*appErrors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got: %T", err)
	}

	if appErr.StatusCode != 500 {
		t.Errorf("Expected 500 Internal, got: %d", appErr.StatusCode)
	}

	if appErr.Message != "Failed to acknowledge assignment" {
		t.Errorf("Expected specific message, got: %s", appErr.Message)
	}
}

func TestAcknowledgeAssignment_DifferentRoles(t *testing.T) {
	testCases := []struct {
		name     string
		roleType string
	}{
		{"Center role", "center"},
		{"Assistant 1 role", "assistant_1"},
		{"Assistant 2 role", "assistant_2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			roleType := tc.roleType
			repo := &mockRepository{
				GetRefereeAssignmentRoleFunc: func(ctx context.Context, matchID int64, refereeID int64) (*string, error) {
					return &roleType, nil
				},
				AcknowledgeAssignmentFunc: func(ctx context.Context, matchID int64, refereeID int64, acknowledgedAt time.Time) error {
					return nil
				},
			}

			service := NewService(repo)
			ctx := context.Background()

			result, err := service.AcknowledgeAssignment(ctx, 1, 100)

			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}

			if !result.Success {
				t.Error("Expected success to be true")
			}
		})
	}
}
