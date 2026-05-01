package eligibility

// EligibleReferee represents a referee with computed eligibility for a match role
type EligibleReferee struct {
	ID               int64   `json:"id"`
	FirstName        string  `json:"first_name"`
	LastName         string  `json:"last_name"`
	Email            string  `json:"email"`
	Grade            *string `json:"grade"` // Junior, Mid, Senior, or null
	DateOfBirth      *string `json:"date_of_birth"`
	Certified        bool    `json:"certified"`
	CertExpiry       *string `json:"cert_expiry"`
	AgeAtMatch       *int    `json:"age_at_match"`        // computed age on match date
	IsEligible       bool    `json:"is_eligible"`         // overall eligibility for this role
	IneligibleReason *string `json:"ineligible_reason"`   // why not eligible, if applicable
	IsAvailable      bool    `json:"is_available"`        // has the referee marked availability for this match
}

// RefereeData represents raw referee data from the database
type RefereeData struct {
	ID          int64
	FirstName   string
	LastName    string
	Email       string
	Grade       *string
	DateOfBirth *string
	Certified   bool
	CertExpiry  *string
	IsAvailable bool
}

// MatchData represents match information needed for eligibility checking
type MatchData struct {
	ID        int64
	AgeGroup  string
	MatchDate string // YYYY-MM-DD format
}
