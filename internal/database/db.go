// Package database persists analyzed rentals in a local SQLite file.
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
	"soi-tro/internal/priceparser"
	"time"

	_ "modernc.org/sqlite" // registers the "sqlite" driver
)

// DB is the process-wide connection opened by InitDB.
var DB *sql.DB

// RentalRecord is one analyzed rental as stored in SQLite.
type RentalRecord struct {
	ID        int64
	CreatedAt string
	// PriceVND is the normalized VND amount for Result.Price, or nil when
	// the listed price could not be unambiguously parsed.
	PriceVND *int64
	Result   *gemini.RentalExtractionResult
}

// GetDBPath returns the SQLite file path under the user's config dir.
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
		price_vnd INTEGER,
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

	ctx := context.Background()

	failureStage = "create_schema"
	if _, err = db.ExecContext(ctx, query); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to create table: %w", err)
	}

	failureStage = "migrate_schema"
	if err = addColumnIfMissing(ctx, db, "rentals", "price_vnd", "INTEGER"); err != nil {
		_ = db.Close()
		return err
	}

	DB = db
	return nil
}

// addColumnIfMissing adds a nullable column to an existing table when a
// database created before that column existed is opened again.
func addColumnIfMissing(ctx context.Context, db *sql.DB, table, column, sqlType string) error {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return fmt.Errorf("failed to inspect %s schema: %w", table, err)
	}
	defer func() { _ = rows.Close() }()

	var (
		cid        int
		name       string
		colType    string
		notNull    int
		defaultVal sql.NullString
		pk         int
	)

	for rows.Next() {
		if err := rows.Scan(&cid, &name, &colType, &notNull, &defaultVal, &pk); err != nil {
			return fmt.Errorf("failed to read %s schema: %w", table, err)
		}

		if name == column {
			return nil
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("failed to read %s schema: %w", table, err)
	}

	if _, err := db.ExecContext(ctx, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, sqlType)); err != nil {
		return fmt.Errorf("failed to add %s.%s column: %w", table, column, err)
	}

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

	var priceVND sql.NullInt64
	if vnd, perr := priceparser.ParseVND(result.Price); perr == nil {
		priceVND = sql.NullInt64{Int64: vnd, Valid: true}
	}

	failureStage = "insert_record"
	res, err := DB.ExecContext(context.Background(), `
		INSERT INTO rentals (
			price, price_vnd, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.Price, priceVND, result.Deposit, result.Floor, result.Electricity, result.Water, result.ParkingFee, result.PetsAllowed, result.PhoneNumber, result.AdditionalNotes,
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

	rows, err := DB.QueryContext(context.Background(), "SELECT id, datetime(created_at, 'localtime'), price, price_vnd, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages FROM rentals ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rec RentalRecord
		var res gemini.RentalExtractionResult
		var priceVND sql.NullInt64
		var rawFieldsStr, missingFieldsStr, sampleMessagesStr string
		failureStage = "scan_record"
		err = rows.Scan(
			&rec.ID, &rec.CreatedAt, &res.Price, &priceVND, &res.Deposit, &res.Floor, &res.Electricity, &res.Water, &res.ParkingFee, &res.PetsAllowed, &res.PhoneNumber, &res.AdditionalNotes,
			&rawFieldsStr, &missingFieldsStr, &sampleMessagesStr,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal([]byte(rawFieldsStr), &res.RawFields)
		_ = json.Unmarshal([]byte(missingFieldsStr), &res.MissingFields)
		_ = json.Unmarshal([]byte(sampleMessagesStr), &res.SampleMessages)

		if priceVND.Valid {
			rec.PriceVND = &priceVND.Int64
		}
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
