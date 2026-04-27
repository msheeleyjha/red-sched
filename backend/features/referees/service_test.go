package referees

import (
	"context"
	"errors"
	"testing"
	"time"

	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
	ListFunc                     func(ctx context.Context) ([]RefereeListItem, error)
	FindByIDFunc                 func(ctx context.Context, id int64) (*RefereeData, error)
	UpdateFunc                   func(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error)
	HasUpcomingAssignmentsFunc   func(ctx context.Context, refereeID int64) (bool, error)
}

func (m *mockRepository) List(ctx context.Context) ([]RefereeListItem, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx)
	}
	return nil, errors.New("List not implemented")
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (*RefereeData, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, errors.New("FindByID not implemented")
}

func (m *mockRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates)
	}
	return nil, errors.New("Update not implemented")
}

func (m *mockRepository) HasUpcomingAssignments(ctx context.Context, refereeID int64) (bool, error) {
	if m.HasUpcomingAssignmentsFunc != nil {
		return m.HasUpcomingAssignmentsFunc(ctx, refereeID)
	}
	return false, errors.New("HasUpcomingAssignments not implemented")
}

func TestList_Success(t *testing.T) {
	now := time.Now()
	certExpiry := now.AddDate(0, 2, 0) // 2 months from now

	mockReferees := []RefereeListItem{
		{
			ID:         1,
			Email:      "ref1@example.com",
			Name:       "Referee One",
			Role:       "referee",
			Status:     "active",
			CertStatus: "valid",
			CertExpiry: &certExpiry,
			CreatedAt:  now,
		},
		{
			ID:         2,
			Email:      "ref2@example.com",
			Name:       "Referee Two",
			Role:       "pending_referee",
			Status:     "pending",
			CertStatus: "none",
			CreatedAt:  now,
		},
	}

	repo := &mockRepository{
		ListFunc: func(ctx context.Context) ([]RefereeListItem, error) {
			return mockReferees, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	referees, err := service.List(ctx)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(referees) != 2 {
		t.Errorf("Expected 2 referees, got: %d", len(referees))
	}

	if referees[0].Email != "ref1@example.com" {
		t.Errorf("Expected first referee email ref1@example.com, got: %s", referees[0].Email)
	}
}

func TestList_RepositoryError(t *testing.T) {
	repo := &mockRepository{
		ListFunc: func(ctx context.Context) ([]RefereeListItem, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	referees, err := service.List(ctx)

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if referees != nil {
		t.Errorf("Expected nil referees, got: %v", referees)
	}

	appErr, ok := err.(*appErrors.AppError)
	if !ok {
		t.Fatalf("Expected AppError, got: %T", err)
	}

	if appErr.StatusCode != 500 {
		t.Errorf("Expected 500 Internal, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_Success(t *testing.T) {
	grade := "Senior"
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Test Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
		HasUpcomingAssignmentsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
			return &UpdateResult{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Test Referee",
				Role:   "referee",
				Status: "active",
				Grade:  &grade,
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &UpdateRequest{
		Grade: &grade,
	}

	result, err := service.Update(ctx, 1, 100, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Grade == nil || *result.Grade != "Senior" {
		t.Errorf("Expected grade Senior, got: %v", result.Grade)
	}
}

func TestUpdate_RefereeNotFound(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return nil, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	status := "active"
	req := &UpdateRequest{
		Status: &status,
	}

	result, err := service.Update(ctx, 999, 100, req)

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
}

func TestUpdate_NoUpdatesProvided(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Test Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &UpdateRequest{} // No fields set

	result, err := service.Update(ctx, 1, 100, req)

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

	if appErr.StatusCode != 400 {
		t.Errorf("Expected 400 BadRequest, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_CannotModifyOtherAssignor(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     2,
				Email:  "assignor@example.com",
				Name:   "Other Assignor",
				Role:   "assignor",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	status := "inactive"
	req := &UpdateRequest{
		Status: &status,
	}

	// Try to modify assignor ID=2 as user ID=1
	result, err := service.Update(ctx, 2, 1, req)

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

	if appErr.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_CanDemoteOtherAssignor(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     2,
				Email:  "assignor@example.com",
				Name:   "Other Assignor",
				Role:   "assignor",
				Status: "active",
			}, nil
		},
		HasUpcomingAssignmentsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
			return &UpdateResult{
				ID:     2,
				Email:  "assignor@example.com",
				Name:   "Other Assignor",
				Role:   "referee",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	role := "referee"
	req := &UpdateRequest{
		Role: &role,
	}

	// Should be allowed to demote assignor to referee
	result, err := service.Update(ctx, 2, 1, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Role != "referee" {
		t.Errorf("Expected role referee, got: %s", result.Role)
	}
}

func TestUpdate_CannotDeactivateSelf(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "self@example.com",
				Name:   "Self",
				Role:   "assignor",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	status := "inactive"
	req := &UpdateRequest{
		Status: &status,
	}

	// Try to deactivate self (user ID=1, referee ID=1)
	result, err := service.Update(ctx, 1, 1, req)

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

	if appErr.StatusCode != 403 {
		t.Errorf("Expected 403 Forbidden, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_CannotDeactivateWithUpcomingAssignments(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     2,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
		HasUpcomingAssignmentsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
			return true, nil // Has upcoming assignments
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	status := "inactive"
	req := &UpdateRequest{
		Status: &status,
	}

	result, err := service.Update(ctx, 2, 1, req)

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

	if appErr.StatusCode != 400 {
		t.Errorf("Expected 400 BadRequest, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_AutoPromotePendingReferee(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "pending@example.com",
				Name:   "Pending Referee",
				Role:   "pending_referee",
				Status: "pending",
			}, nil
		},
		HasUpcomingAssignmentsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
			// Should have both status=active and role=referee
			if updates["status"] != "active" {
				t.Errorf("Expected status=active in updates, got: %v", updates["status"])
			}
			if updates["role"] != "referee" {
				t.Errorf("Expected role=referee in updates, got: %v", updates["role"])
			}

			return &UpdateResult{
				ID:     1,
				Email:  "pending@example.com",
				Name:   "Pending Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	status := "active"
	req := &UpdateRequest{
		Status: &status,
	}

	result, err := service.Update(ctx, 1, 100, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Role != "referee" {
		t.Errorf("Expected role referee, got: %s", result.Role)
	}

	if result.Status != "active" {
		t.Errorf("Expected status active, got: %s", result.Status)
	}
}

func TestUpdate_AutoActivateWhenPromotingToAssignor(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "pending",
			}, nil
		},
		HasUpcomingAssignmentsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
			// Should have both role=assignor and status=active
			if updates["role"] != "assignor" {
				t.Errorf("Expected role=assignor in updates, got: %v", updates["role"])
			}
			if updates["status"] != "active" {
				t.Errorf("Expected status=active in updates, got: %v", updates["status"])
			}

			return &UpdateResult{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "assignor",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	role := "assignor"
	req := &UpdateRequest{
		Role: &role,
	}

	result, err := service.Update(ctx, 1, 100, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Role != "assignor" {
		t.Errorf("Expected role assignor, got: %s", result.Role)
	}

	if result.Status != "active" {
		t.Errorf("Expected status active, got: %s", result.Status)
	}
}

func TestUpdate_InvalidStatus(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	status := "invalid_status"
	req := &UpdateRequest{
		Status: &status,
	}

	result, err := service.Update(ctx, 1, 100, req)

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

	if appErr.StatusCode != 400 {
		t.Errorf("Expected 400 BadRequest, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_InvalidRole(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	role := "admin"
	req := &UpdateRequest{
		Role: &role,
	}

	result, err := service.Update(ctx, 1, 100, req)

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

	if appErr.StatusCode != 400 {
		t.Errorf("Expected 400 BadRequest, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_InvalidGrade(t *testing.T) {
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	grade := "Expert"
	req := &UpdateRequest{
		Grade: &grade,
	}

	result, err := service.Update(ctx, 1, 100, req)

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

	if appErr.StatusCode != 400 {
		t.Errorf("Expected 400 BadRequest, got: %d", appErr.StatusCode)
	}
}

func TestUpdate_SetGradeToNull(t *testing.T) {
	emptyGrade := ""
	repo := &mockRepository{
		FindByIDFunc: func(ctx context.Context, id int64) (*RefereeData, error) {
			return &RefereeData{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "active",
			}, nil
		},
		HasUpcomingAssignmentsFunc: func(ctx context.Context, refereeID int64) (bool, error) {
			return false, nil
		},
		UpdateFunc: func(ctx context.Context, id int64, updates map[string]interface{}) (*UpdateResult, error) {
			// Should have grade=nil
			if updates["grade"] != nil {
				t.Errorf("Expected grade=nil in updates, got: %v", updates["grade"])
			}

			return &UpdateResult{
				ID:     1,
				Email:  "ref@example.com",
				Name:   "Referee",
				Role:   "referee",
				Status: "active",
				Grade:  nil,
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &UpdateRequest{
		Grade: &emptyGrade,
	}

	result, err := service.Update(ctx, 1, 100, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Grade != nil {
		t.Errorf("Expected nil grade, got: %v", result.Grade)
	}
}

func TestDetermineCertStatus(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name       string
		certified  bool
		certExpiry *time.Time
		expected   string
	}{
		{
			name:      "Not certified",
			certified: false,
			expected:  "none",
		},
		{
			name:       "Certified but no expiry",
			certified:  true,
			certExpiry: nil,
			expected:   "none",
		},
		{
			name:       "Expired",
			certified:  true,
			certExpiry: timePtr(now.AddDate(0, 0, -1)), // Yesterday
			expected:   "expired",
		},
		{
			name:       "Expiring soon (29 days)",
			certified:  true,
			certExpiry: timePtr(now.AddDate(0, 0, 29)),
			expected:   "expiring_soon",
		},
		{
			name:       "Valid (31 days)",
			certified:  true,
			certExpiry: timePtr(now.AddDate(0, 0, 31)),
			expected:   "valid",
		},
		{
			name:       "Valid (60 days)",
			certified:  true,
			certExpiry: timePtr(now.AddDate(0, 0, 60)),
			expected:   "valid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := DetermineCertStatus(tc.certified, tc.certExpiry, now)
			if result != tc.expected {
				t.Errorf("Expected %s, got: %s", tc.expected, result)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
