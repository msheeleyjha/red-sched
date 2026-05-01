package users

import "context"

// ServiceInterface defines the interface for user business logic
type ServiceInterface interface {
	FindOrCreate(ctx context.Context, googleID, email, name string) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	UpdateProfile(ctx context.Context, userID int64, req ProfileUpdateRequest) (*User, error)
	GetProfile(ctx context.Context, userID int64) (*User, error)
	GetByGoogleID(ctx context.Context, googleID string) (*User, error)
}
