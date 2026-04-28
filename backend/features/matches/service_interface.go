package matches

import (
	"context"
	"mime/multipart"
)

// ServiceInterface defines the interface for match business logic
type ServiceInterface interface {
	ParseCSV(ctx context.Context, file multipart.File, filename string) (*ImportPreviewResponse, error)
	ImportMatches(ctx context.Context, req *ImportConfirmRequest, currentUserID int64) (*ImportResult, error)
	CreateRoleSlotsForMatch(ctx context.Context, matchID int64, ageGroup string) error
	ListMatches(ctx context.Context, params *MatchListParams) (*PaginatedMatchesResponse, error)
	ListActiveMatches(ctx context.Context) ([]MatchWithRoles, error)
	ListArchivedMatches(ctx context.Context) ([]MatchWithRoles, error)
	GetMatchWithRoles(ctx context.Context, matchID int64) (*MatchWithRoles, error)
	UpdateMatch(ctx context.Context, matchID int64, req *MatchUpdateRequest, actorID int64) (*MatchWithRoles, error)
	AddRoleSlot(ctx context.Context, matchID int64, roleType string) error
	ArchiveMatch(ctx context.Context, matchID int64, userID int64) error
	UnarchiveMatch(ctx context.Context, matchID int64) error
}
