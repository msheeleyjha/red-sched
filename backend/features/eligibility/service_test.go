package eligibility

import (
	"context"
	"errors"
	"testing"
	"time"

	appErrors "github.com/msheeley/referee-scheduler/shared/errors"
)

// mockRepository implements RepositoryInterface for testing
type mockRepository struct {
	GetMatchDataFunc      func(ctx context.Context, matchID int64) (*MatchData, error)
	GetActiveRefereesFunc func(ctx context.Context, matchID int64) ([]RefereeData, error)
}

func (m *mockRepository) GetMatchData(ctx context.Context, matchID int64) (*MatchData, error) {
	if m.GetMatchDataFunc != nil {
		return m.GetMatchDataFunc(ctx, matchID)
	}
	return nil, errors.New("GetMatchData not implemented")
}

func (m *mockRepository) GetActiveReferees(ctx context.Context, matchID int64) ([]RefereeData, error) {
	if m.GetActiveRefereesFunc != nil {
		return m.GetActiveRefereesFunc(ctx, matchID)
	}
	return nil, errors.New("GetActiveReferees not implemented")
}

func TestGetEligibleReferees_Success(t *testing.T) {
	dob := "1990-01-01"
	certExpiry := "2028-12-31"
	grade := "Senior"

	repo := &mockRepository{
		GetMatchDataFunc: func(ctx context.Context, matchID int64) (*MatchData, error) {
			return &MatchData{
				ID:        1,
				AgeGroup:  "U12",
				MatchDate: "2027-06-15",
			}, nil
		},
		GetActiveRefereesFunc: func(ctx context.Context, matchID int64) ([]RefereeData, error) {
			return []RefereeData{
				{
					ID:          100,
					FirstName:   "John",
					LastName:    "Doe",
					Email:       "john@example.com",
					Grade:       &grade,
					DateOfBirth: &dob,
					Certified:   true,
					CertExpiry:  &certExpiry,
					IsAvailable: true,
				},
			}, nil
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.GetEligibleReferees(ctx, 1, "center")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("Expected 1 referee, got: %d", len(result))
	}

	if result[0].ID != 100 {
		t.Errorf("Expected ID=100, got: %d", result[0].ID)
	}

	if !result[0].IsEligible {
		t.Errorf("Expected referee to be eligible, got ineligible: %v", result[0].IneligibleReason)
	}

	if result[0].AgeAtMatch == nil {
		t.Error("Expected AgeAtMatch to be set")
	}
}

func TestGetEligibleReferees_InvalidRoleType(t *testing.T) {
	repo := &mockRepository{}
	service := NewService(repo)
	ctx := context.Background()

	result, err := service.GetEligibleReferees(ctx, 1, "invalid_role")

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

func TestGetEligibleReferees_MatchNotFound(t *testing.T) {
	repo := &mockRepository{
		GetMatchDataFunc: func(ctx context.Context, matchID int64) (*MatchData, error) {
			return nil, nil // Match not found
		},
	}

	service := NewService(repo)
	ctx := context.Background()

	result, err := service.GetEligibleReferees(ctx, 999, "center")

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

func TestCalculateAgeAtDate(t *testing.T) {
	testCases := []struct {
		name         string
		birthDate    string
		targetDate   string
		expectedAge  int
	}{
		{
			name:        "Simple case",
			birthDate:   "1990-01-01",
			targetDate:  "2020-01-01",
			expectedAge: 30,
		},
		{
			name:        "Birthday not yet occurred",
			birthDate:   "1990-06-15",
			targetDate:  "2020-06-14",
			expectedAge: 29,
		},
		{
			name:        "Birthday today",
			birthDate:   "1990-06-15",
			targetDate:  "2020-06-15",
			expectedAge: 30,
		},
		{
			name:        "Birthday passed",
			birthDate:   "1990-06-15",
			targetDate:  "2020-06-16",
			expectedAge: 30,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			birth, _ := time.Parse("2006-01-02", tc.birthDate)
			target, _ := time.Parse("2006-01-02", tc.targetDate)

			age := CalculateAgeAtDate(birth, target)

			if age != tc.expectedAge {
				t.Errorf("Expected age %d, got: %d", tc.expectedAge, age)
			}
		})
	}
}

func TestCheckEligibility_U10AndYounger(t *testing.T) {
	testCases := []struct {
		name             string
		ageGroup         string
		roleType         string
		matchDate        string
		dob              string
		certified        bool
		certExpiry       *string
		expectedEligible bool
		expectedReason   *string
	}{
		{
			name:             "U6 - eligible (7 years old)",
			ageGroup:         "U6",
			roleType:         "center",
			matchDate:        "2027-06-15",
			dob:              "2020-01-01",
			certified:        false,
			certExpiry:       nil,
			expectedEligible: true,
			expectedReason:   nil,
		},
		{
			name:             "U6 - not eligible (6 years old)",
			ageGroup:         "U6",
			roleType:         "center",
			matchDate:        "2027-06-15",
			dob:              "2021-07-01",
			certified:        false,
			certExpiry:       nil,
			expectedEligible: false,
			expectedReason:   strPtr("Must be at least 7 years old (currently 5)"),
		},
		{
			name:             "U10 - eligible (11 years old)",
			ageGroup:         "U10",
			roleType:         "assistant_1",
			matchDate:        "2027-06-15",
			dob:              "2016-01-01",
			certified:        false,
			certExpiry:       nil,
			expectedEligible: true,
			expectedReason:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchDate, _ := time.Parse("2006-01-02", tc.matchDate)
			dobStr := tc.dob

			isEligible, reason := CheckEligibility(
				tc.ageGroup,
				tc.roleType,
				matchDate,
				&dobStr,
				tc.certified,
				tc.certExpiry,
			)

			if isEligible != tc.expectedEligible {
				t.Errorf("Expected eligible=%v, got: %v", tc.expectedEligible, isEligible)
			}

			if tc.expectedReason == nil && reason != nil {
				t.Errorf("Expected no reason, got: %v", *reason)
			}

			if tc.expectedReason != nil {
				if reason == nil {
					t.Errorf("Expected reason '%s', got nil", *tc.expectedReason)
				} else if *reason != *tc.expectedReason {
					t.Errorf("Expected reason '%s', got: '%s'", *tc.expectedReason, *reason)
				}
			}
		})
	}
}

func TestCheckEligibility_U12AndOlder_Center(t *testing.T) {
	certExpiry := "2028-12-31"
	expiredCert := "2027-01-01"

	testCases := []struct {
		name             string
		ageGroup         string
		dob              string
		certified        bool
		certExpiry       *string
		expectedEligible bool
		reasonContains   string
	}{
		{
			name:             "U12 center - eligible with valid cert",
			ageGroup:         "U12",
			dob:              "1990-01-01",
			certified:        true,
			certExpiry:       &certExpiry,
			expectedEligible: true,
			reasonContains:   "",
		},
		{
			name:             "U12 center - not eligible without cert",
			ageGroup:         "U12",
			dob:              "1990-01-01",
			certified:        false,
			certExpiry:       nil,
			expectedEligible: false,
			reasonContains:   "Certification required",
		},
		{
			name:             "U12 center - not eligible with expired cert",
			ageGroup:         "U12",
			dob:              "1990-01-01",
			certified:        true,
			certExpiry:       &expiredCert,
			expectedEligible: false,
			reasonContains:   "expires before match date",
		},
		{
			name:             "U12 center - not eligible with missing expiry",
			ageGroup:         "U12",
			dob:              "1990-01-01",
			certified:        true,
			certExpiry:       nil,
			expectedEligible: false,
			reasonContains:   "expiry date missing",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchDate, _ := time.Parse("2006-01-02", "2027-06-15")
			dobStr := tc.dob

			isEligible, reason := CheckEligibility(
				tc.ageGroup,
				"center",
				matchDate,
				&dobStr,
				tc.certified,
				tc.certExpiry,
			)

			if isEligible != tc.expectedEligible {
				t.Errorf("Expected eligible=%v, got: %v", tc.expectedEligible, isEligible)
			}

			if tc.reasonContains != "" {
				if reason == nil {
					t.Errorf("Expected reason containing '%s', got nil", tc.reasonContains)
				} else if !contains(*reason, tc.reasonContains) {
					t.Errorf("Expected reason to contain '%s', got: '%s'", tc.reasonContains, *reason)
				}
			}
		})
	}
}

func TestCheckEligibility_U12AndOlder_Assistant(t *testing.T) {
	testCases := []struct {
		name     string
		roleType string
	}{
		{"assistant_1", "assistant_1"},
		{"assistant_2", "assistant_2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			matchDate, _ := time.Parse("2006-01-02", "2027-06-15")
			dobStr := "1990-01-01"

			// Assistant roles on U12+ have no restrictions
			isEligible, reason := CheckEligibility(
				"U12",
				tc.roleType,
				matchDate,
				&dobStr,
				false, // Not certified
				nil,   // No cert expiry
			)

			if !isEligible {
				t.Errorf("Expected eligible, got: %v", reason)
			}
		})
	}
}

func TestCheckEligibility_InvalidAgeGroup(t *testing.T) {
	matchDate, _ := time.Parse("2006-01-02", "2027-06-15")
	dobStr := "1990-01-01"

	isEligible, reason := CheckEligibility(
		"Invalid",
		"center",
		matchDate,
		&dobStr,
		true,
		nil,
	)

	if isEligible {
		t.Error("Expected not eligible for invalid age group")
	}

	if reason == nil || !contains(*reason, "Invalid age group") {
		t.Errorf("Expected 'Invalid age group' reason, got: %v", reason)
	}
}

func TestCheckEligibility_MissingDOB(t *testing.T) {
	matchDate, _ := time.Parse("2006-01-02", "2027-06-15")

	isEligible, reason := CheckEligibility(
		"U12",
		"center",
		matchDate,
		nil, // No DOB
		true,
		nil,
	)

	if isEligible {
		t.Error("Expected not eligible with missing DOB")
	}

	if reason == nil || !contains(*reason, "Date of birth is required") {
		t.Errorf("Expected 'Date of birth is required' reason, got: %v", reason)
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
