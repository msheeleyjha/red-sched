package users

import (
	"context"
	"fmt"
	"time"

	"github.com/msheeley/referee-scheduler/shared/errors"
)

// Service handles user business logic
type Service struct {
	repo *Repository
}

// NewService creates a new user service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// FindOrCreate finds an existing user by Google ID or creates a new one
func (s *Service) FindOrCreate(ctx context.Context, googleID, email, name string) (*User, error) {
	// Try to find existing user
	user, err := s.repo.FindByGoogleID(ctx, googleID)
	if err != nil {
		return nil, errors.NewInternal("failed to find user", err)
	}

	// User exists
	if user != nil {
		return user, nil
	}

	// Create new user
	user, err = s.repo.Create(ctx, googleID, email, name)
	if err != nil {
		return nil, errors.NewInternal("failed to create user", err)
	}

	return user, nil
}

// GetByID retrieves a user by ID
func (s *Service) GetByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.NewInternal("failed to get user", err)
	}

	if user == nil {
		return nil, errors.NewNotFound("User")
	}

	return user, nil
}

// UpdateProfile updates a user's profile with validation
func (s *Service) UpdateProfile(ctx context.Context, userID int64, req ProfileUpdateRequest) (*User, error) {
	// Validate and parse date of birth
	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsedDOB, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, errors.NewBadRequest("Invalid date of birth format. Use YYYY-MM-DD")
		}
		if parsedDOB.After(time.Now()) {
			return nil, errors.NewBadRequest("Date of birth cannot be in the future")
		}
		dob = &parsedDOB
	}

	// Validate certification expiry
	var certExpiry *time.Time
	if req.Certified {
		if req.CertExpiry == nil || *req.CertExpiry == "" {
			return nil, errors.NewBadRequest("Certification expiry date is required when certified")
		}
		parsedExpiry, err := time.Parse("2006-01-02", *req.CertExpiry)
		if err != nil {
			return nil, errors.NewBadRequest("Invalid certification expiry format. Use YYYY-MM-DD")
		}
		if parsedExpiry.Before(time.Now()) {
			return nil, errors.NewBadRequest("Certification expiry must be in the future")
		}
		certExpiry = &parsedExpiry
	}

	// Prepare update data
	data := ProfileUpdateData{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		DOB:        dob,
		Certified:  req.Certified,
		CertExpiry: certExpiry,
	}

	// Update profile
	user, err := s.repo.UpdateProfile(ctx, userID, data)
	if err != nil {
		return nil, errors.NewInternal("failed to update profile", err)
	}

	return user, nil
}

// GetProfile retrieves a user's complete profile
func (s *Service) GetProfile(ctx context.Context, userID int64) (*User, error) {
	return s.GetByID(ctx, userID)
}

// GetByGoogleID retrieves a user by Google ID (for auth)
func (s *Service) GetByGoogleID(ctx context.Context, googleID string) (*User, error) {
	user, err := s.repo.FindByGoogleID(ctx, googleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by google id: %w", err)
	}
	return user, nil
}
