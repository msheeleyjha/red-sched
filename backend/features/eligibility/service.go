package eligibility

import (
	"context"
	"fmt"
	"time"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles eligibility business logic
type Service struct {
	repo RepositoryInterface
}

// NewService creates a new eligibility service
func NewService(repo RepositoryInterface) *Service {
	return &Service{repo: repo}
}

// ValidRoleTypes are the allowed role type values
var ValidRoleTypes = map[string]bool{
	"center":      true,
	"assistant_1": true,
	"assistant_2": true,
}

// GetEligibleReferees returns all referees with eligibility status for a specific match and role
func (s *Service) GetEligibleReferees(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error) {
	// Validate role type
	if !ValidRoleTypes[roleType] {
		return nil, errors.NewBadRequest("Invalid role type. Must be: center, assistant_1, or assistant_2")
	}

	// Get match data
	matchData, err := s.repo.GetMatchData(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to get match data", err)
	}
	if matchData == nil {
		return nil, errors.NewNotFound("Match")
	}

	// Parse match date
	matchDate, err := time.Parse("2006-01-02", matchData.MatchDate)
	if err != nil {
		return nil, errors.NewInternal("Failed to parse match date", err)
	}

	// Get active referees
	refereesData, err := s.repo.GetActiveReferees(ctx, matchID)
	if err != nil {
		return nil, errors.NewInternal("Failed to get active referees", err)
	}

	// Build eligible referees list with eligibility checks
	result := make([]EligibleReferee, 0, len(refereesData))

	for _, refData := range refereesData {
		ref := EligibleReferee{
			ID:          refData.ID,
			FirstName:   refData.FirstName,
			LastName:    refData.LastName,
			Email:       refData.Email,
			Grade:       refData.Grade,
			DateOfBirth: refData.DateOfBirth,
			Certified:   refData.Certified,
			CertExpiry:  refData.CertExpiry,
			IsAvailable: refData.IsAvailable,
		}

		// Calculate age at match if DOB is available
		if refData.DateOfBirth != nil {
			dob, err := time.Parse("2006-01-02", *refData.DateOfBirth)
			if err == nil {
				age := CalculateAgeAtDate(dob, matchDate)
				ref.AgeAtMatch = &age
			}
		}

		// Check eligibility
		ref.IsEligible, ref.IneligibleReason = CheckEligibility(
			matchData.AgeGroup,
			roleType,
			matchDate,
			refData.DateOfBirth,
			refData.Certified,
			refData.CertExpiry,
		)

		result = append(result, ref)
	}

	return result, nil
}

// CalculateAgeAtDate calculates age in years at a specific date
func CalculateAgeAtDate(birthDate, targetDate time.Time) int {
	age := targetDate.Year() - birthDate.Year()

	// Adjust if birthday hasn't occurred yet this year
	if targetDate.Month() < birthDate.Month() ||
		(targetDate.Month() == birthDate.Month() && targetDate.Day() < birthDate.Day()) {
		age--
	}

	return age
}

// CheckEligibility determines if a referee is eligible for a specific role on a match
// Returns (isEligible, ineligibleReason)
func CheckEligibility(
	ageGroup string,
	roleType string,
	matchDate time.Time,
	dobStr *string,
	certified bool,
	certExpiryStr *string,
) (bool, *string) {
	// Parse age group to get the age number (e.g., "U12" -> 12)
	var ageGroupNum int
	_, err := fmt.Sscanf(ageGroup, "U%d", &ageGroupNum)
	if err != nil {
		reason := "Invalid age group format"
		return false, &reason
	}

	// Parse date of birth
	if dobStr == nil {
		reason := "Date of birth is required"
		return false, &reason
	}

	dob, err := time.Parse("2006-01-02", *dobStr)
	if err != nil {
		reason := "Invalid date of birth format"
		return false, &reason
	}

	age := CalculateAgeAtDate(dob, matchDate)

	// Rule 1: U10 and younger - age-based eligibility for all roles
	if ageGroupNum <= 10 {
		requiredAge := ageGroupNum + 1
		if age < requiredAge {
			reason := fmt.Sprintf("Must be at least %d years old (currently %d)", requiredAge, age)
			return false, &reason
		}
		return true, nil
	}

	// Rule 2: U12 and older - Center Referee requires certification
	if roleType == "center" {
		if !certified {
			reason := "Certification required for center referee role on U12+ matches"
			return false, &reason
		}

		// Check if certification expiry is provided
		if certExpiryStr == nil {
			reason := "Certification expiry date missing"
			return false, &reason
		}

		// Parse certification expiry
		certExpiry, err := time.Parse("2006-01-02", *certExpiryStr)
		if err != nil {
			reason := "Invalid certification expiry format"
			return false, &reason
		}

		// Cert must be valid on match date (expires AFTER match date)
		if !certExpiry.After(matchDate) {
			reason := fmt.Sprintf("Certification expires before match date (%s)", certExpiry.Format("2006-01-02"))
			return false, &reason
		}

		return true, nil
	}

	// Rule 3: U12 and older - Assistant Referee has no restrictions
	if roleType == "assistant_1" || roleType == "assistant_2" {
		return true, nil
	}

	// Should not reach here
	reason := "Unknown role type"
	return false, &reason
}
