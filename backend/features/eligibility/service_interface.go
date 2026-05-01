package eligibility

import "context"

// ServiceInterface defines the interface for eligibility business logic
type ServiceInterface interface {
	// GetEligibleReferees returns all referees with eligibility status for a specific match and role
	GetEligibleReferees(ctx context.Context, matchID int64, roleType string) ([]EligibleReferee, error)
}
