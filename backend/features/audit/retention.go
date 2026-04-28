package audit

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	defaultRetentionDays = 730
	batchSize            = 1000
)

// RetentionService manages audit log retention and purging
type RetentionService struct {
	db              *sql.DB
	retentionDays   int
	schedulerTicker *time.Ticker
	stopChan        chan bool
}

// PurgeResult contains statistics about a purge operation
type PurgeResult struct {
	DeletedCount int       `json:"deleted_count"`
	CutoffDate   time.Time `json:"cutoff_date"`
	StartedAt    time.Time `json:"started_at"`
	CompletedAt  time.Time `json:"completed_at"`
	DurationMs   int64     `json:"duration_ms"`
}

// NewRetentionService creates a new retention service
func NewRetentionService(db *sql.DB, retentionDays int) *RetentionService {
	service := &RetentionService{
		db:            db,
		retentionDays: retentionDays,
		stopChan:      make(chan bool),
	}

	log.Printf("Audit retention service initialized: %d days retention", retentionDays)
	return service
}

// Start begins the daily purge scheduler
func (s *RetentionService) Start() {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	durationUntilMidnight := nextMidnight.Sub(now)

	log.Printf("Audit retention scheduler starting. First purge in %v at %v", durationUntilMidnight, nextMidnight)

	go func() {
		time.Sleep(durationUntilMidnight)

		log.Println("Running scheduled audit log purge")
		result, err := s.PurgeOldLogs()
		if err != nil {
			log.Printf("Error during scheduled purge: %v", err)
		} else {
			log.Printf("Scheduled purge completed: deleted %d logs older than %v",
				result.DeletedCount, result.CutoffDate.Format("2006-01-02"))
		}

		s.schedulerTicker = time.NewTicker(24 * time.Hour)
		defer s.schedulerTicker.Stop()

		for {
			select {
			case <-s.schedulerTicker.C:
				log.Println("Running scheduled audit log purge")
				result, err := s.PurgeOldLogs()
				if err != nil {
					log.Printf("Error during scheduled purge: %v", err)
				} else {
					log.Printf("Scheduled purge completed: deleted %d logs older than %v",
						result.DeletedCount, result.CutoffDate.Format("2006-01-02"))
				}
			case <-s.stopChan:
				log.Println("Audit retention scheduler stopped")
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

// PurgeOldLogs deletes audit logs older than the retention period
func (s *RetentionService) PurgeOldLogs() (*PurgeResult, error) {
	startTime := time.Now()
	cutoffDate := time.Now().AddDate(0, 0, -s.retentionDays)

	log.Printf("Starting audit log purge for logs older than %v", cutoffDate.Format("2006-01-02"))

	result := &PurgeResult{
		DeletedCount: 0,
		CutoffDate:   cutoffDate,
		StartedAt:    startTime,
	}

	var totalToDelete int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM audit_logs WHERE created_at < $1",
		cutoffDate,
	).Scan(&totalToDelete)
	if err != nil {
		return nil, fmt.Errorf("failed to count logs for purge: %w", err)
	}

	if totalToDelete == 0 {
		log.Println("No audit logs to purge")
		result.CompletedAt = time.Now()
		result.DurationMs = time.Since(startTime).Milliseconds()
		return result, nil
	}

	log.Printf("Found %d audit logs to purge", totalToDelete)

	for {
		deleteResult, err := s.db.Exec(
			`DELETE FROM audit_logs
			WHERE id IN (
				SELECT id FROM audit_logs
				WHERE created_at < $1
				ORDER BY created_at ASC
				LIMIT $2
			)`,
			cutoffDate,
			batchSize,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to delete audit logs batch: %w", err)
		}

		rowsDeleted, _ := deleteResult.RowsAffected()
		result.DeletedCount += int(rowsDeleted)

		if rowsDeleted < int64(batchSize) {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}

	result.CompletedAt = time.Now()
	result.DurationMs = time.Since(startTime).Milliseconds()

	log.Printf("Audit log purge completed: deleted %d logs in %dms",
		result.DeletedCount, result.DurationMs)

	if result.DeletedCount > 0 {
		s.logPurgeOperation(result)
	}

	return result, nil
}

func (s *RetentionService) logPurgeOperation(result *PurgeResult) {
	_, err := s.db.Exec(
		`INSERT INTO audit_logs (user_id, action_type, entity_type, entity_id, old_values, new_values, ip_address, created_at)
		VALUES (NULL, 'delete', 'audit_log_purge', 0, $1, NULL, NULL, CURRENT_TIMESTAMP)`,
		fmt.Sprintf(`{
			"deleted_count": %d,
			"cutoff_date": "%s",
			"retention_days": %d,
			"duration_ms": %d,
			"started_at": "%s",
			"completed_at": "%s"
		}`,
			result.DeletedCount,
			result.CutoffDate.Format(time.RFC3339),
			s.retentionDays,
			result.DurationMs,
			result.StartedAt.Format(time.RFC3339),
			result.CompletedAt.Format(time.RFC3339),
		),
	)
	if err != nil {
		log.Printf("Warning: Failed to create meta-audit log for purge operation: %v", err)
	}
}

// GetRetentionDays returns the current retention period in days
func (s *RetentionService) GetRetentionDays() int {
	return s.retentionDays
}
