package matches

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	defaultMatchRetentionDays = 730
	matchPurgeBatchSize       = 100
)

// RetentionService manages archived match retention and purging
type RetentionService struct {
	db              *sql.DB
	retentionDays   int
	schedulerTicker *time.Ticker
	stopChan        chan bool
}

// PurgeResult contains statistics about a match purge operation
type PurgeResult struct {
	MatchesDeleted int       `json:"matches_deleted"`
	RolesDeleted   int       `json:"roles_deleted"`
	CutoffDate     time.Time `json:"cutoff_date"`
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    time.Time `json:"completed_at"`
	DurationMs     int64     `json:"duration_ms"`
}

// NewRetentionService creates a new match retention service
func NewRetentionService(db *sql.DB, retentionDays int) *RetentionService {
	service := &RetentionService{
		db:            db,
		retentionDays: retentionDays,
		stopChan:      make(chan bool),
	}

	log.Printf("Match retention service initialized: %d days retention", retentionDays)
	return service
}

// Start begins the monthly purge scheduler
func (s *RetentionService) Start() {
	now := time.Now()
	firstDayNextMonth := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	durationUntilNextMonth := firstDayNextMonth.Sub(now)

	log.Printf("Match retention scheduler starting. First purge in %v at %v", durationUntilNextMonth, firstDayNextMonth)

	go func() {
		time.Sleep(durationUntilNextMonth)

		log.Println("Running scheduled archived match purge")
		result, err := s.PurgeOldMatches()
		if err != nil {
			log.Printf("Error during scheduled match purge: %v", err)
		} else {
			log.Printf("Scheduled match purge completed: deleted %d matches older than %v",
				result.MatchesDeleted, result.CutoffDate.Format("2006-01-02"))
		}

		s.schedulerTicker = time.NewTicker(30 * 24 * time.Hour)
		defer s.schedulerTicker.Stop()

		for {
			select {
			case <-s.schedulerTicker.C:
				log.Println("Running scheduled archived match purge")
				result, err := s.PurgeOldMatches()
				if err != nil {
					log.Printf("Error during scheduled match purge: %v", err)
				} else {
					log.Printf("Scheduled match purge completed: deleted %d matches older than %v",
						result.MatchesDeleted, result.CutoffDate.Format("2006-01-02"))
				}
			case <-s.stopChan:
				log.Println("Match retention scheduler stopped")
				return
			}
		}
	}()
}

// Stop halts the scheduler
func (s *RetentionService) Stop() {
	close(s.stopChan)
	if s.schedulerTicker != nil {
		s.schedulerTicker.Stop()
	}
}

// PurgeOldMatches deletes archived matches older than the retention period
func (s *RetentionService) PurgeOldMatches() (*PurgeResult, error) {
	startTime := time.Now()
	cutoffDate := time.Now().AddDate(0, 0, -s.retentionDays)

	log.Printf("Starting archived match purge for matches archived before %v", cutoffDate.Format("2006-01-02"))

	result := &PurgeResult{
		MatchesDeleted: 0,
		RolesDeleted:   0,
		CutoffDate:     cutoffDate,
		StartedAt:      startTime,
	}

	var totalToDelete int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM matches WHERE archived = TRUE AND archived_at < $1",
		cutoffDate,
	).Scan(&totalToDelete)
	if err != nil {
		return nil, fmt.Errorf("failed to count matches for purge: %w", err)
	}

	if totalToDelete == 0 {
		log.Println("No archived matches to purge")
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	log.Printf("Found %d archived matches to purge", totalToDelete)

	for {
		tx, err := s.db.Begin()
		if err != nil {
			return nil, fmt.Errorf("failed to begin transaction: %w", err)
		}

		rows, err := tx.Query(
			`SELECT id FROM matches
			WHERE archived = TRUE AND archived_at < $1
			ORDER BY archived_at ASC
			LIMIT $2`,
			cutoffDate,
			matchPurgeBatchSize,
		)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to select matches for deletion: %w", err)
		}

		var matchIDs []int64
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				tx.Rollback()
				return nil, fmt.Errorf("failed to scan match ID: %w", err)
			}
			matchIDs = append(matchIDs, id)
		}
		rows.Close()

		if len(matchIDs) == 0 {
			tx.Rollback()
			break
		}

		rolesResult, err := tx.Exec(
			`DELETE FROM assignments WHERE match_id = ANY($1)`,
			convertToPostgresArray(matchIDs),
		)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete match roles: %w", err)
		}

		rolesDeleted, _ := rolesResult.RowsAffected()
		result.RolesDeleted += int(rolesDeleted)

		matchesResult, err := tx.Exec(
			`DELETE FROM matches WHERE id = ANY($1)`,
			convertToPostgresArray(matchIDs),
		)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to delete matches: %w", err)
		}

		matchesDeleted, _ := matchesResult.RowsAffected()
		result.MatchesDeleted += int(matchesDeleted)

		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}

		log.Printf("Purged batch: %d matches, %d roles", matchesDeleted, rolesDeleted)

		if len(matchIDs) < matchPurgeBatchSize {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(startTime).Milliseconds()

	log.Printf("Archived match purge completed: deleted %d matches and %d role assignments in %dms",
		result.MatchesDeleted, result.RolesDeleted, result.DurationMs)

	if result.MatchesDeleted > 0 {
		s.logPurgeOperation(result)
	}

	return result, nil
}

func (s *RetentionService) logPurgeOperation(result *PurgeResult) {
	_, err := s.db.Exec(
		`INSERT INTO audit_logs (user_id, action_type, entity_type, entity_id, old_values, new_values, ip_address, created_at)
		VALUES (NULL, 'delete', 'match_purge', 0, $1, NULL, NULL, CURRENT_TIMESTAMP)`,
		fmt.Sprintf(`{
			"matches_deleted": %d,
			"roles_deleted": %d,
			"cutoff_date": "%s",
			"retention_days": %d,
			"duration_ms": %d,
			"started_at": "%s",
			"completed_at": "%s"
		}`,
			result.MatchesDeleted,
			result.RolesDeleted,
			result.CutoffDate.Format(time.RFC3339),
			s.retentionDays,
			result.DurationMs,
			result.StartedAt.Format(time.RFC3339),
			result.CompletedAt.Format(time.RFC3339),
		),
	)
	if err != nil {
		log.Printf("Warning: Failed to create audit log for match purge operation: %v", err)
	}
}

// GetRetentionDays returns the current retention period in days
func (s *RetentionService) GetRetentionDays() int {
	return s.retentionDays
}

func convertToPostgresArray(ids []int64) interface{} {
	return ids
}
