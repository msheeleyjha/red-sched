package availability

import (
	"context"
	"errors"
	"testing"
	"time"

	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
	ToggleMatchAvailabilityFunc      func(ctx context.Context, matchID int64, refereeID int64, available *bool) error
	MatchExistsAndActiveFunc         func(ctx context.Context, matchID int64) (bool, error)
	GetDayUnavailabilityFunc         func(ctx context.Context, refereeID int64) ([]DayUnavailabilityData, error)
	ToggleDayUnavailabilityFunc      func(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error
	ClearMatchAvailabilityForDayFunc func(ctx context.Context, refereeID int64, date string) error
}

func (m *mockRepository) ToggleMatchAvailability(ctx context.Context, matchID int64, refereeID int64, available *bool) error {
	if m.ToggleMatchAvailabilityFunc != nil {
		return m.ToggleMatchAvailabilityFunc(ctx, matchID, refereeID, available)
	}
	return errors.New("ToggleMatchAvailability not implemented")
}

func (m *mockRepository) MatchExistsAndActive(ctx context.Context, matchID int64) (bool, error) {
	if m.MatchExistsAndActiveFunc != nil {
		return m.MatchExistsAndActiveFunc(ctx, matchID)
	}
	return false, errors.New("MatchExistsAndActive not implemented")
}

func (m *mockRepository) GetDayUnavailability(ctx context.Context, refereeID int64) ([]DayUnavailabilityData, error) {
	if m.GetDayUnavailabilityFunc != nil {
		return m.GetDayUnavailabilityFunc(ctx, refereeID)
	}
	return nil, errors.New("GetDayUnavailability not implemented")
}

func (m *mockRepository) ToggleDayUnavailability(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error {
	if m.ToggleDayUnavailabilityFunc != nil {
		return m.ToggleDayUnavailabilityFunc(ctx, refereeID, date, unavailable, reason)
	}
	return errors.New("ToggleDayUnavailability not implemented")
}

func (m *mockRepository) ClearMatchAvailabilityForDay(ctx context.Context, refereeID int64, date string) error {
	if m.ClearMatchAvailabilityForDayFunc != nil {
		return m.ClearMatchAvailabilityForDayFunc(ctx, refereeID, date)
	}
	return errors.New("ClearMatchAvailabilityForDay not implemented")
}

func TestToggleMatchAvailability_Success(t *testing.T) {
	available := true
	repo := &mockRepository{
		MatchExistsAndActiveFunc: func(ctx context.Context, matchID int64) (bool, error) {
			return true, nil
		},
		ToggleMatchAvailabilityFunc: func(ctx context.Context, matchID int64, refereeID int64, available *bool) error {
			if matchID != 1 || refereeID != 100 {
				t.Errorf("Unexpected parameters: matchID=%d, refereeID=%d", matchID, refereeID)
			}
			if available == nil || *available != true {
				t.Errorf("Expected available=true, got: %v", available)
			}
			return nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &ToggleMatchAvailabilityRequest{
		Available: &available,
	}

	result, err := service.ToggleMatchAvailability(ctx, 1, 100, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Available == nil || *result.Available != true {
		t.Errorf("Expected available=true, got: %v", result.Available)
	}
}

func TestToggleMatchAvailability_MatchNotFound(t *testing.T) {
	available := true
	repo := &mockRepository{
		MatchExistsAndActiveFunc: func(ctx context.Context, matchID int64) (bool, error) {
			return false, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &ToggleMatchAvailabilityRequest{
		Available: &available,
	}

	result, err := service.ToggleMatchAvailability(ctx, 999, 100, req)

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

func TestToggleMatchAvailability_ClearPreference(t *testing.T) {
	repo := &mockRepository{
		MatchExistsAndActiveFunc: func(ctx context.Context, matchID int64) (bool, error) {
			return true, nil
		},
		ToggleMatchAvailabilityFunc: func(ctx context.Context, matchID int64, refereeID int64, available *bool) error {
			if available != nil {
				t.Errorf("Expected available=nil, got: %v", available)
			}
			return nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &ToggleMatchAvailabilityRequest{
		Available: nil, // Clear preference
	}

	result, err := service.ToggleMatchAvailability(ctx, 1, 100, req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Available != nil {
		t.Errorf("Expected nil available, got: %v", result.Available)
	}
}

func TestGetDayUnavailability_Success(t *testing.T) {
	now := time.Now()
	reason := "Vacation"

	mockData := []DayUnavailabilityData{
		{
			ID:              1,
			RefereeID:       100,
			UnavailableDate: now.AddDate(0, 0, 1),
			Reason:          &reason,
			CreatedAt:       now,
		},
		{
			ID:              2,
			RefereeID:       100,
			UnavailableDate: now.AddDate(0, 0, 2),
			Reason:          nil,
			CreatedAt:       now,
		},
	}

	repo := &mockRepository{
		GetDayUnavailabilityFunc: func(ctx context.Context, refereeID int64) ([]DayUnavailabilityData, error) {
			if refereeID != 100 {
				t.Errorf("Expected refereeID=100, got: %d", refereeID)
			}
			return mockData, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.GetDayUnavailability(ctx, 100)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 results, got: %d", len(result))
	}

	if result[0].ID != 1 {
		t.Errorf("Expected ID=1, got: %d", result[0].ID)
	}

	if result[0].Reason == nil || *result[0].Reason != "Vacation" {
		t.Errorf("Expected reason='Vacation', got: %v", result[0].Reason)
	}

	if result[1].Reason != nil {
		t.Errorf("Expected nil reason, got: %v", result[1].Reason)
	}
}

func TestGetDayUnavailability_EmptyResult(t *testing.T) {
	repo := &mockRepository{
		GetDayUnavailabilityFunc: func(ctx context.Context, refereeID int64) ([]DayUnavailabilityData, error) {
			return []DayUnavailabilityData{}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.GetDayUnavailability(ctx, 100)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("Expected 0 results, got: %d", len(result))
	}
}

func TestToggleDayUnavailability_MarkUnavailable(t *testing.T) {
	reason := "Out of town"
	toggleCalled := false
	clearCalled := false

	repo := &mockRepository{
		ToggleDayUnavailabilityFunc: func(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error {
			toggleCalled = true
			if refereeID != 100 {
				t.Errorf("Expected refereeID=100, got: %d", refereeID)
			}
			if date != "2027-12-31" {
				t.Errorf("Expected date=2027-12-31, got: %s", date)
			}
			if !unavailable {
				t.Error("Expected unavailable=true")
			}
			if reason == nil || *reason != "Out of town" {
				t.Errorf("Expected reason='Out of town', got: %v", reason)
			}
			return nil
		},
		ClearMatchAvailabilityForDayFunc: func(ctx context.Context, refereeID int64, date string) error {
			clearCalled = true
			if date != "2027-12-31" {
				t.Errorf("Expected date=2027-12-31, got: %s", date)
			}
			return nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &ToggleDayUnavailabilityRequest{
		Unavailable: true,
		Reason:      &reason,
	}

	result, err := service.ToggleDayUnavailability(ctx, 100, "2027-12-31", req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !toggleCalled {
		t.Error("Expected ToggleDayUnavailability to be called")
	}

	if !clearCalled {
		t.Error("Expected ClearMatchAvailabilityForDay to be called")
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if !result.Success {
		t.Error("Expected success=true")
	}

	if !result.Unavailable {
		t.Error("Expected unavailable=true")
	}

	if result.Date != "2027-12-31" {
		t.Errorf("Expected date=2027-12-31, got: %s", result.Date)
	}
}

func TestToggleDayUnavailability_RemoveUnavailability(t *testing.T) {
	toggleCalled := false
	clearCalled := false

	repo := &mockRepository{
		ToggleDayUnavailabilityFunc: func(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error {
			toggleCalled = true
			if unavailable {
				t.Error("Expected unavailable=false")
			}
			return nil
		},
		ClearMatchAvailabilityForDayFunc: func(ctx context.Context, refereeID int64, date string) error {
			clearCalled = true
			return nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &ToggleDayUnavailabilityRequest{
		Unavailable: false,
		Reason:      nil,
	}

	result, err := service.ToggleDayUnavailability(ctx, 100, "2027-12-31", req)

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if !toggleCalled {
		t.Error("Expected ToggleDayUnavailability to be called")
	}

	if clearCalled {
		t.Error("Expected ClearMatchAvailabilityForDay NOT to be called")
	}

	if result.Unavailable {
		t.Error("Expected unavailable=false")
	}
}

func TestToggleDayUnavailability_InvalidDate(t *testing.T) {
	repo := &mockRepository{}

	service := NewService(repo)
	ctx := context.Background()

	testCases := []struct {
		name string
		date string
	}{
		{"Invalid format", "12/31/2027"},
		{"Invalid date", "2027-13-01"},
		{"Not a date", "invalid"},
		{"Empty string", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &ToggleDayUnavailabilityRequest{
				Unavailable: true,
			}

			result, err := service.ToggleDayUnavailability(ctx, 100, tc.date, req)

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
		})
	}
}

func TestToggleDayUnavailability_RepositoryError(t *testing.T) {
	repo := &mockRepository{
		ToggleDayUnavailabilityFunc: func(ctx context.Context, refereeID int64, date string, unavailable bool, reason *string) error {
			return errors.New("database error")
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	req := &ToggleDayUnavailabilityRequest{
		Unavailable: true,
	}

	result, err := service.ToggleDayUnavailability(ctx, 100, "2027-12-31", req)

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
}
