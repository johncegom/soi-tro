package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"soi-tro/internal/logger"
	"soi-tro/internal/viewing"
	"time"
)

const createDecisionTrailsTable = `
	CREATE TABLE IF NOT EXISTS decision_trails (
		rental_id INTEGER PRIMARY KEY,
		status TEXT NOT NULL,
		next_action TEXT,
		answers TEXT NOT NULL DEFAULT '{}',
		updated_at TEXT NOT NULL
	);`

// DecisionTrailStore persists viewing decision trails in SQLite. Answers are
// personal notes: nothing here logs them.
type DecisionTrailStore struct {
	db *sql.DB
}

// NewDecisionTrailStore returns a store backed by db.
func NewDecisionTrailStore(db *sql.DB) *DecisionTrailStore {
	return &DecisionTrailStore{db: db}
}

// Get returns the trail for a rental, or an empty Unreviewed trail when none
// has been saved.
func (s *DecisionTrailStore) Get(ctx context.Context, rentalID int64) (trail viewing.DecisionTrail, err error) {
	started := time.Now()
	log := logger.With("operation", "database.get_decision_trail").With("rental_id", rentalID)
	defer func() { logger.LogOperationResult(log, started, "query_trail", err) }()

	var (
		status      string
		nextAction  sql.NullString
		answersJSON string
	)

	err = s.db.QueryRowContext(ctx,
		"SELECT status, next_action, answers FROM decision_trails WHERE rental_id = ?", rentalID,
	).Scan(&status, &nextAction, &answersJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return viewing.NewTrail(rentalID), nil
	}

	if err != nil {
		return viewing.DecisionTrail{}, fmt.Errorf("failed to read decision trail: %w", err)
	}

	trail = viewing.NewTrail(rentalID)
	trail.Status = viewing.DecisionStatus(status)

	if err = json.Unmarshal([]byte(answersJSON), &trail.Answers); err != nil {
		return viewing.DecisionTrail{}, fmt.Errorf("failed to decode decision trail answers: %w", err)
	}

	if trail.Answers == nil {
		trail.Answers = map[viewing.ItemID]string{}
	}

	if nextAction.Valid {
		due, perr := time.Parse(time.RFC3339, nextAction.String)
		if perr != nil {
			err = fmt.Errorf("failed to parse decision trail next action: %w", perr)

			return viewing.DecisionTrail{}, err
		}

		trail.NextAction = &due
	}

	return trail, nil
}

// Save stores a trail, replacing any earlier one for the same rental.
func (s *DecisionTrailStore) Save(ctx context.Context, trail viewing.DecisionTrail) (err error) {
	started := time.Now()
	log := logger.With("operation", "database.save_decision_trail").With("rental_id", trail.RentalID).With("answer_count", len(trail.Answers))
	defer func() { logger.LogOperationResult(log, started, "save_trail", err) }()

	if !trail.Status.Valid() {
		return viewing.ErrInvalidStatus
	}

	answers := trail.Answers
	if answers == nil {
		answers = map[viewing.ItemID]string{}
	}

	answersJSON, err := json.Marshal(answers)
	if err != nil {
		return fmt.Errorf("failed to encode decision trail answers: %w", err)
	}

	var nextAction sql.NullString
	if trail.NextAction != nil {
		nextAction = sql.NullString{String: trail.NextAction.UTC().Format(time.RFC3339), Valid: true}
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO decision_trails (rental_id, status, next_action, answers, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(rental_id) DO UPDATE SET
			status = excluded.status,
			next_action = excluded.next_action,
			answers = excluded.answers,
			updated_at = excluded.updated_at`,
		trail.RentalID, string(trail.Status), nextAction, string(answersJSON), time.Now().UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to save decision trail: %w", err)
	}

	return nil
}

// Delete removes a rental's trail. A missing trail is not an error.
func (s *DecisionTrailStore) Delete(ctx context.Context, rentalID int64) (err error) {
	started := time.Now()
	log := logger.With("operation", "database.delete_decision_trail").With("rental_id", rentalID)
	defer func() { logger.LogOperationResult(log, started, "delete_trail", err) }()

	_, err = s.db.ExecContext(ctx, "DELETE FROM decision_trails WHERE rental_id = ?", rentalID)

	return err
}

// DeleteRentalAndTrail removes a rental and its decision trail in one
// transaction, so a failure never leaves notes without their rental.
func DeleteRentalAndTrail(ctx context.Context, rentalID int64) (err error) {
	started := time.Now()
	log := logger.With("operation", "database.delete_rental_and_trail").With("rental_id", rentalID)
	defer func() { logger.LogOperationResult(log, started, "delete_records", err) }()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() { _ = tx.Rollback() }()

	if _, err = tx.ExecContext(ctx, "DELETE FROM decision_trails WHERE rental_id = ?", rentalID); err != nil {
		return err
	}

	if _, err = tx.ExecContext(ctx, "DELETE FROM rentals WHERE id = ?", rentalID); err != nil {
		return err
	}

	return tx.Commit()
}
