package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"soi-tro/internal/gemini"
	"soi-tro/internal/logger"
	"time"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

type RentalRecord struct {
	ID        int64
	CreatedAt string
	Result    *gemini.RentalExtractionResult
}

func GetDBPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "soi-tro", "rentals.db"), nil
}

// InitDB opens the local history database and ensures its schema exists.
//
//nolint:wsl_v5 // Preserve the existing straight-line database flow around logging stages.
func InitDB() (err error) {
	started := time.Now()
	failureStage := "get_path"
	log := logger.With("operation", "database.init")
	defer func() { logger.LogOperationResult(log, started, failureStage, err) }()

	dbPath, err := GetDBPath()
	if err != nil {
		return err
	}

	failureStage = "create_directory"
	configDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

	failureStage = "open_database"
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	query := `
	CREATE TABLE IF NOT EXISTS rentals (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		price TEXT,
		deposit TEXT,
		floor TEXT,
		electricity TEXT,
		water TEXT,
		parking_fee TEXT,
		pets_allowed TEXT,
		phone_number TEXT,
		additional_notes TEXT,
		raw_fields TEXT,
		missing_fields TEXT,
		sample_messages TEXT
	);`

	failureStage = "create_schema"
	if _, err = db.ExecContext(context.Background(), query); err != nil {
		db.Close()
		return fmt.Errorf("failed to create table: %w", err)
	}

	DB = db
	return nil
}

// SaveRental persists one extracted rental result in local history.
//
//nolint:wsl_v5 // Preserve the existing straight-line database flow around logging stages.
func SaveRental(result *gemini.RentalExtractionResult) (id int64, err error) {
	started := time.Now()
	failureStage := "encode_raw_fields"
	log := logger.With("operation", "database.save_rental")
	defer func() { logger.LogOperationResult(log, started, failureStage, err) }()

	rawFieldsBytes, err := json.Marshal(result.RawFields)
	if err != nil {
		return 0, err
	}
	failureStage = "encode_missing_fields"
	missingFieldsBytes, err := json.Marshal(result.MissingFields)
	if err != nil {
		return 0, err
	}
	failureStage = "encode_sample_messages"
	sampleMessagesBytes, err := json.Marshal(result.SampleMessages)
	if err != nil {
		return 0, err
	}

	failureStage = "insert_record"
	res, err := DB.ExecContext(context.Background(), `
		INSERT INTO rentals (
			price, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.Price, result.Deposit, result.Floor, result.Electricity, result.Water, result.ParkingFee, result.PetsAllowed, result.PhoneNumber, result.AdditionalNotes,
		string(rawFieldsBytes), string(missingFieldsBytes), string(sampleMessagesBytes),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rental: %w", err)
	}

	failureStage = "read_record_id"
	id, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListRentals returns saved rental records in reverse insertion order.
//
//nolint:wsl_v5 // Preserve the existing row-decoding flow around logging stages.
func ListRentals() (records []RentalRecord, err error) {
	started := time.Now()
	failureStage := "query_records"
	log := logger.With("operation", "database.list_rentals")
	defer func() { logger.LogOperationResult(log, started, failureStage, err) }()

	rows, err := DB.QueryContext(context.Background(), "SELECT id, datetime(created_at, 'localtime'), price, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages FROM rentals ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rec RentalRecord
		var res gemini.RentalExtractionResult
		var rawFieldsStr, missingFieldsStr, sampleMessagesStr string
		failureStage = "scan_record"
		err = rows.Scan(
			&rec.ID, &rec.CreatedAt, &res.Price, &res.Deposit, &res.Floor, &res.Electricity, &res.Water, &res.ParkingFee, &res.PetsAllowed, &res.PhoneNumber, &res.AdditionalNotes,
			&rawFieldsStr, &missingFieldsStr, &sampleMessagesStr,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(rawFieldsStr), &res.RawFields)
		_ = json.Unmarshal([]byte(missingFieldsStr), &res.MissingFields)
		_ = json.Unmarshal([]byte(sampleMessagesStr), &res.SampleMessages)

		rec.Result = &res
		records = append(records, rec)
	}

	failureStage = "iterate_records"
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return records, nil
}

// DeleteRental removes one rental record from local history.
//
//nolint:wsl_v5 // Preserve the existing straight-line database flow around logging stages.
func DeleteRental(id int64) (err error) {
	started := time.Now()
	log := logger.With("operation", "database.delete_rental")
	defer func() { logger.LogOperationResult(log, started, "delete_record", err) }()

	_, err = DB.ExecContext(context.Background(), "DELETE FROM rentals WHERE id = ?", id)
	return err
}
