package match_reports

import "errors"

var (
	// ErrNotFound is returned when a match report is not found
	ErrNotFound = errors.New("match report not found")

	// ErrAlreadyExists is returned when trying to create a report for a match that already has one
	ErrAlreadyExists = errors.New("match report already exists for this match")

	// ErrInvalidScore is returned when score values are invalid
	ErrInvalidScore = errors.New("invalid score: must be non-negative")

	// ErrInvalidCards is returned when card values are invalid
	ErrInvalidCards = errors.New("invalid card count: must be non-negative")

	// ErrUnauthorized is returned when user is not authorized to perform action
	ErrUnauthorized = errors.New("unauthorized: only center referee or assignor can submit/edit reports")
)
