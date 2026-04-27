package users

import (
	"context"
	"database/sql"
	"testing"
	"time"
)

// mockRepository is a mock implementation of Repository for testing
type mockRepository struct {
	findByGoogleIDFunc  func(ctx context.Context, googleID string) (*User, error)
	findByIDFunc        func(ctx context.Context, id int64) (*User, error)
	createFunc          func(ctx context.Context, googleID, email, name string) (*User, error)
	updateProfileFunc   func(ctx context.Context, userID int64, data ProfileUpdateData) (*User, error)
}

func (m *mockRepository) FindByGoogleID(ctx context.Context, googleID string) (*User, error) {
	if m.findByGoogleIDFunc != nil {
		return m.findByGoogleIDFunc(ctx, googleID)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockRepository) Create(ctx context.Context, googleID, email, name string) (*User, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, googleID, email, name)
	}
	return nil, nil
}

func (m *mockRepository) UpdateProfile(ctx context.Context, userID int64, data ProfileUpdateData) (*User, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, data)
	}
	return nil, nil
}

func TestService_FindOrCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("returns existing user when found", func(t *testing.T) {
		existingUser := &User{
			ID:       1,
			GoogleID: "google123",
			Email:    "test@example.com",
			Name:     "Test User",
		}

		mockRepo := &mockRepository{
			findByGoogleIDFunc: func(ctx context.Context, googleID string) (*User, error) {
				if googleID == "google123" {
					return existingUser, nil
				}
				return nil, nil
			},
		}

		service := NewService(mockRepo)
		user, err := service.FindOrCreate(ctx, "google123", "test@example.com", "Test User")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", user.ID)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Expected email test@example.com, got %s", user.Email)
		}
	})

	t.Run("creates new user when not found", func(t *testing.T) {
		newUser := &User{
			ID:       2,
			GoogleID: "google456",
			Email:    "new@example.com",
			Name:     "New User",
		}

		mockRepo := &mockRepository{
			findByGoogleIDFunc: func(ctx context.Context, googleID string) (*User, error) {
				return nil, nil // User not found
			},
			createFunc: func(ctx context.Context, googleID, email, name string) (*User, error) {
				if googleID == "google456" {
					return newUser, nil
				}
				return nil, sql.ErrNoRows
			},
		}

		service := NewService(mockRepo)
		user, err := service.FindOrCreate(ctx, "google456", "new@example.com", "New User")

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.ID != 2 {
			t.Errorf("Expected user ID 2, got %d", user.ID)
		}
		if user.GoogleID != "google456" {
			t.Errorf("Expected Google ID google456, got %s", user.GoogleID)
		}
	})

	t.Run("returns error when find fails", func(t *testing.T) {
		mockRepo := &mockRepository{
			findByGoogleIDFunc: func(ctx context.Context, googleID string) (*User, error) {
				return nil, sql.ErrConnDone
			},
		}

		service := NewService(mockRepo)
		_, err := service.FindOrCreate(ctx, "google123", "test@example.com", "Test User")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("returns error when create fails", func(t *testing.T) {
		mockRepo := &mockRepository{
			findByGoogleIDFunc: func(ctx context.Context, googleID string) (*User, error) {
				return nil, nil // Not found
			},
			createFunc: func(ctx context.Context, googleID, email, name string) (*User, error) {
				return nil, sql.ErrConnDone
			},
		}

		service := NewService(mockRepo)
		_, err := service.FindOrCreate(ctx, "google123", "test@example.com", "Test User")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user when found", func(t *testing.T) {
		existingUser := &User{
			ID:    1,
			Email: "test@example.com",
			Name:  "Test User",
		}

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*User, error) {
				if id == 1 {
					return existingUser, nil
				}
				return nil, nil
			},
		}

		service := NewService(mockRepo)
		user, err := service.GetByID(ctx, 1)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", user.ID)
		}
	})

	t.Run("returns not found error when user doesn't exist", func(t *testing.T) {
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*User, error) {
				return nil, nil // Not found
			},
		}

		service := NewService(mockRepo)
		_, err := service.GetByID(ctx, 999)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		// Check it's the right type of error (NotFound)
		if err.Error() != "User not found" {
			t.Errorf("Expected 'User not found' error, got %v", err)
		}
	})

	t.Run("returns error when database fails", func(t *testing.T) {
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*User, error) {
				return nil, sql.ErrConnDone
			},
		}

		service := NewService(mockRepo)
		_, err := service.GetByID(ctx, 1)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

func TestService_UpdateProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("successfully updates profile with valid data", func(t *testing.T) {
		dob := "1990-01-15"
		certExpiry := "2027-12-31"

		updatedUser := &User{
			ID:        1,
			FirstName: stringPtr("John"),
			LastName:  stringPtr("Doe"),
			Certified: true,
		}

		mockRepo := &mockRepository{
			updateProfileFunc: func(ctx context.Context, userID int64, data ProfileUpdateData) (*User, error) {
				if userID == 1 {
					return updatedUser, nil
				}
				return nil, sql.ErrNoRows
			},
		}

		service := NewService(mockRepo)
		req := ProfileUpdateRequest{
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: &dob,
			Certified:   true,
			CertExpiry:  &certExpiry,
		}

		user, err := service.UpdateProfile(ctx, 1, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", user.ID)
		}
	})

	t.Run("rejects invalid date of birth format", func(t *testing.T) {
		invalidDOB := "01/15/1990" // Wrong format

		service := NewService(&mockRepository{})
		req := ProfileUpdateRequest{
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: &invalidDOB,
			Certified:   false,
		}

		_, err := service.UpdateProfile(ctx, 1, req)

		if err == nil {
			t.Fatal("Expected error for invalid date format, got nil")
		}
		if err.Error() != "Invalid date of birth format. Use YYYY-MM-DD" {
			t.Errorf("Expected date format error, got %v", err)
		}
	})

	t.Run("rejects future date of birth", func(t *testing.T) {
		futureDOB := time.Now().Add(24 * time.Hour).Format("2006-01-02")

		service := NewService(&mockRepository{})
		req := ProfileUpdateRequest{
			FirstName:   "John",
			LastName:    "Doe",
			DateOfBirth: &futureDOB,
			Certified:   false,
		}

		_, err := service.UpdateProfile(ctx, 1, req)

		if err == nil {
			t.Fatal("Expected error for future date, got nil")
		}
		if err.Error() != "Date of birth cannot be in the future" {
			t.Errorf("Expected future date error, got %v", err)
		}
	})

	t.Run("requires cert expiry when certified", func(t *testing.T) {
		service := NewService(&mockRepository{})
		req := ProfileUpdateRequest{
			FirstName:  "John",
			LastName:   "Doe",
			Certified:  true,
			CertExpiry: nil, // Missing!
		}

		_, err := service.UpdateProfile(ctx, 1, req)

		if err == nil {
			t.Fatal("Expected error for missing cert expiry, got nil")
		}
		if err.Error() != "Certification expiry date is required when certified" {
			t.Errorf("Expected cert expiry required error, got %v", err)
		}
	})

	t.Run("rejects past certification expiry", func(t *testing.T) {
		pastExpiry := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

		service := NewService(&mockRepository{})
		req := ProfileUpdateRequest{
			FirstName:  "John",
			LastName:   "Doe",
			Certified:  true,
			CertExpiry: &pastExpiry,
		}

		_, err := service.UpdateProfile(ctx, 1, req)

		if err == nil {
			t.Fatal("Expected error for past cert expiry, got nil")
		}
		if err.Error() != "Certification expiry must be in the future" {
			t.Errorf("Expected future expiry error, got %v", err)
		}
	})

	t.Run("allows uncertified without cert expiry", func(t *testing.T) {
		updatedUser := &User{
			ID:        1,
			Certified: false,
		}

		mockRepo := &mockRepository{
			updateProfileFunc: func(ctx context.Context, userID int64, data ProfileUpdateData) (*User, error) {
				if !data.Certified && data.CertExpiry == nil {
					return updatedUser, nil
				}
				return nil, sql.ErrNoRows
			},
		}

		service := NewService(mockRepo)
		req := ProfileUpdateRequest{
			FirstName:  "John",
			LastName:   "Doe",
			Certified:  false,
			CertExpiry: nil,
		}

		user, err := service.UpdateProfile(ctx, 1, req)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if user.Certified {
			t.Error("Expected user to be uncertified")
		}
	})

	t.Run("returns error when update fails", func(t *testing.T) {
		certExpiry := "2025-12-31"

		mockRepo := &mockRepository{
			updateProfileFunc: func(ctx context.Context, userID int64, data ProfileUpdateData) (*User, error) {
				return nil, sql.ErrConnDone
			},
		}

		service := NewService(mockRepo)
		req := ProfileUpdateRequest{
			FirstName:  "John",
			LastName:   "Doe",
			Certified:  true,
			CertExpiry: &certExpiry,
		}

		_, err := service.UpdateProfile(ctx, 1, req)

		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})
}

func TestService_GetProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("returns user profile", func(t *testing.T) {
		user := &User{
			ID:    1,
			Email: "test@example.com",
		}

		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id int64) (*User, error) {
				if id == 1 {
					return user, nil
				}
				return nil, nil
			},
		}

		service := NewService(mockRepo)
		profile, err := service.GetProfile(ctx, 1)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if profile.ID != 1 {
			t.Errorf("Expected profile ID 1, got %d", profile.ID)
		}
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
