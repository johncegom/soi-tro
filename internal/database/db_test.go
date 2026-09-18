package database

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"soi-tro/internal/gemini"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockDBPath(t *testing.T) string {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	os.Setenv("HOME", tmpDir)
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
		if DB != nil {
			_ = DB.Close()
			DB = nil
		}
	})

	return tmpDir
}

func TestDBOperations(t *testing.T) {
	_ = mockDBPath(t)

	err := InitDB()
	require.NoError(t, err)
	defer DB.Close()

	rental := &gemini.RentalExtractionResult{
		Price:           "5 triệu/tháng",
		Deposit:         "5 triệu",
		Floor:           "Tầng 3",
		PhoneNumber:     "0911223344",
		AdditionalNotes: "Yên tĩnh",
		MissingFields:   []string{"parking_fee"},
		SampleMessages: []gemini.SampleMessage{
			{Style: "Lịch sự", Content: "Chào chủ nhà..."},
		},
		RawFields: map[string]string{
			"price": "5 triệu/tháng",
		},
	}

	id, err := SaveRental(rental)
	require.NoError(t, err)
	assert.True(t, id > 0)

	records, err := ListRentals()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, id, records[0].ID)
	assert.Equal(t, "5 triệu/tháng", records[0].Result.Price)
	assert.Equal(t, "5 triệu/tháng", records[0].Result.RawFields["price"])
	assert.Equal(t, "0911223344", records[0].Result.PhoneNumber)
	assert.Contains(t, records[0].Result.MissingFields, "parking_fee")
	require.NotNil(t, records[0].PriceVND)
	assert.Equal(t, int64(5_000_000), *records[0].PriceVND)

	err = DeleteRental(id)
	require.NoError(t, err)

	records, err = ListRentals()
	require.NoError(t, err)
	assert.Len(t, records, 0)
}

func TestSaveRental_UnparsablePriceLeavesPriceVNDNil(t *testing.T) {
	_ = mockDBPath(t)
	require.NoError(t, InitDB())
	defer DB.Close()

	id, err := SaveRental(&gemini.RentalExtractionResult{Price: "thoả thuận"})
	require.NoError(t, err)

	records, err := ListRentals()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, id, records[0].ID)
	assert.Nil(t, records[0].PriceVND)
}

func TestInitDB_MigratesLegacySchemaMissingPriceVND(t *testing.T) {
	mockHome := mockDBPath(t)

	dbPath := filepath.Join(mockHome, ".config", "soi-tro", "rentals.db")
	require.NoError(t, os.MkdirAll(filepath.Dir(dbPath), 0o700))

	legacyDB, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)
	_, err = legacyDB.Exec(`
	CREATE TABLE rentals (
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
	);`)
	require.NoError(t, err)
	_, err = legacyDB.Exec(`
	INSERT INTO rentals (
		price, deposit, floor, electricity, water, parking_fee, pets_allowed, phone_number, additional_notes, raw_fields, missing_fields, sample_messages
	) VALUES ('4.5tr', '', '', '', '', '', '', '', '', '{}', '[]', '[]')`)
	require.NoError(t, err)
	require.NoError(t, legacyDB.Close())

	require.NoError(t, InitDB())
	defer DB.Close()

	records, err := ListRentals()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Nil(t, records[0].PriceVND, "legacy row has no price_vnd until re-saved")

	id, err := SaveRental(&gemini.RentalExtractionResult{Price: "4.5tr"})
	require.NoError(t, err)

	records, err = ListRentals()
	require.NoError(t, err)
	require.Len(t, records, 2)
	for _, rec := range records {
		if rec.ID == id {
			require.NotNil(t, rec.PriceVND)
			assert.Equal(t, int64(4_500_000), *rec.PriceVND)
		}
	}
}

func TestInitDB_HomeDirError(t *testing.T) {
	origHome := os.Getenv("HOME")
	origUserProfile := os.Getenv("USERPROFILE")
	origHomeDrive := os.Getenv("HOMEDRIVE")
	origHomePath := os.Getenv("HOMEPATH")

	os.Setenv("HOME", "")
	os.Setenv("USERPROFILE", "")
	os.Setenv("HOMEDRIVE", "")
	os.Setenv("HOMEPATH", "")

	t.Cleanup(func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOMEDRIVE", origHomeDrive)
		os.Setenv("HOMEPATH", origHomePath)
	})

	err := InitDB()
	assert.Error(t, err)
}

func TestInitDB_MkdirAllError(t *testing.T) {
	mockHome := mockDBPath(t)

	configDir := filepath.Join(mockHome, ".config", "soi-tro")
	err := os.MkdirAll(filepath.Dir(configDir), 0o700)
	require.NoError(t, err)
	err = os.WriteFile(configDir, []byte("some-file"), 0o600)
	require.NoError(t, err)

	err = InitDB()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create db directory")
}

func TestSaveRental_DBError(t *testing.T) {
	_ = mockDBPath(t)
	err := InitDB()
	require.NoError(t, err)
	defer DB.Close()

	_, err = DB.Exec("DROP TABLE rentals")
	require.NoError(t, err)

	rental := &gemini.RentalExtractionResult{}
	_, err = SaveRental(rental)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to insert rental")
}

func TestListRentals_DBError(t *testing.T) {
	_ = mockDBPath(t)
	err := InitDB()
	require.NoError(t, err)
	defer DB.Close()

	_, err = DB.Exec("DROP TABLE rentals")
	require.NoError(t, err)

	_, err = ListRentals()
	assert.Error(t, err)
}

//nolint:paralleltest,wsl_v5 // This test replaces process-wide logging and database state.
func TestSaveRental_LogsSafeStructuredFields(t *testing.T) {
	var output bytes.Buffer
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&output, nil)))
	t.Cleanup(func() { slog.SetDefault(originalLogger) })

	_ = mockDBPath(t)
	require.NoError(t, InitDB())
	output.Reset()

	const sensitiveMarker = "0912-SENSITIVE-CONTACT"
	_, err := SaveRental(&gemini.RentalExtractionResult{
		PhoneNumber: sensitiveMarker,
		RawFields:   map[string]string{"listing": sensitiveMarker},
	})
	require.NoError(t, err)

	logs := output.String()
	assert.Contains(t, logs, `"operation":"database.save_rental"`)
	assert.Contains(t, logs, `"duration_ms":`)
	assert.NotContains(t, logs, sensitiveMarker)

	require.NoError(t, func() error {
		_, dropErr := DB.ExecContext(context.Background(), "DROP TABLE rentals")
		return dropErr
	}())
	output.Reset()
	_, err = SaveRental(&gemini.RentalExtractionResult{PhoneNumber: sensitiveMarker})
	require.Error(t, err)
	assert.Contains(t, output.String(), `"error":"insert_record"`)
	assert.NotContains(t, output.String(), sensitiveMarker)
}
