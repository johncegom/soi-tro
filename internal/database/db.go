package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"

	"soi-tro/internal/gemini"
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

func InitDB() error {
	dbPath, err := GetDBPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("failed to create db directory: %w", err)
	}

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

	if _, err := db.Exec(query); err != nil {
		db.Close()
		return fmt.Errorf("failed to create table: %w", err)
	}

	DB = db
	return nil
}

func SaveRental(result *gemini.RentalExtractionResult) (int64, error) {
	rawFieldsBytes, err := json.Marshal(result.RawFields)
	if err != nil {
		return 0, err
	}
	missingFieldsBytes, err := json.Marshal(result.MissingFields)
	if err != nil {
		return 0, err
	}
	sampleMessagesBytes, err := json.Marshal(result.SampleMessages)
	if err != nil {
		return 0, err
	}

	res, err := DB.Exec(`
		INSERT INTO rentals (
			price, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		result.Price, result.Deposit, result.Floor, result.Electricity, result.Water, result.ParkingFee, result.PetsAllowed, result.PhoneNumber, result.AdditionalNotes,
		string(rawFieldsBytes), string(missingFieldsBytes), string(sampleMessagesBytes),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rental: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func ListRentals() ([]RentalRecord, error) {
	rows, err := DB.Query("SELECT id, datetime(created_at, 'localtime'), price, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages FROM rentals ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []RentalRecord
	for rows.Next() {
		var rec RentalRecord
		var res gemini.RentalExtractionResult
		var rawFieldsStr, missingFieldsStr, sampleMessagesStr string
		err := rows.Scan(
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

	return records, nil
}

func DeleteRental(id int64) error {
	_, err := DB.Exec("DELETE FROM rentals WHERE id = ?", id)
	return err
}
