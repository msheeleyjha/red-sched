package match_reports

import "context"

// ServiceInterface defines the match reports service contract
type ServiceInterface interface {
	// CreateReport creates a new match report
	CreateReport(ctx context.Context, matchID, userID int64, req *CreateMatchReportRequest) (*MatchReport, error)

	// UpdateReport updates an existing match report
	UpdateReport(ctx context.Context, matchID, userID int64, req *UpdateMatchReportRequest) (*MatchReport, error)

	// GetReportByMatchID retrieves a match report by match ID
	GetReportByMatchID(ctx context.Context, matchID int64) (*MatchReport, error)

	// GetReportsBySubmitter retrieves all match reports submitted by a user
	GetReportsBySubmitter(ctx context.Context, userID int64) ([]MatchReport, error)
}
