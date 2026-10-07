package database

import (
	"context"
	"fmt"
	"soi-tro/internal/gemini"
	"soi-tro/internal/logger"
	"time"
)

// SamplePrefix tags the additional notes of every sample rental, so sample rows
// can be told apart from real ones without a schema change.
const SamplePrefix = "[SAMPLE]"

const isSampleRow = "substr(additional_notes, 1, length(?)) = ?"

// sampleRentals returns fake listings for manual testing. Phone numbers are
// obviously fake. Each lists the fields it omits so the viewing checklist has
// something to ask, and states a price and deposit an answer can contradict.
func sampleRentals() []*gemini.RentalExtractionResult {
	return []*gemini.RentalExtractionResult{
		{
			Price: "4tr5/tháng", Deposit: "1 tháng", Floor: "Tầng 2", Electricity: "3.500đ/số", Water: "100k/người",
			PhoneNumber: "0900000001", AdditionalNotes: SamplePrefix + " Đủ thông tin, thiếu phí giữ xe và thú cưng",
			MissingFields: []string{"parking_fee", "pets_allowed"},
			RawFields: map[string]string{
				"price": "4tr5/tháng", "deposit": "1 tháng", "floor": "Tầng 2", "electricity": "3.500đ/số",
				"water": "100k/người", "phone_number": "0900000001",
			},
		},
		{
			Price: "3 triệu", ParkingFee: "100k/tháng", PetsAllowed: "Không",
			PhoneNumber: "0900000002", AdditionalNotes: SamplePrefix + " Thiếu cọc, tầng, điện",
			MissingFields: []string{"deposit", "floor", "electricity"},
			RawFields: map[string]string{
				"price": "3 triệu", "parking_fee": "100k/tháng", "pets_allowed": "Không", "phone_number": "0900000002",
			},
		},
		{
			Price: "5tr", Deposit: "2 tháng", Floor: "Tầng 5, có thang máy", Electricity: "Giá dân", ParkingFee: "Miễn phí", PetsAllowed: "Được nuôi mèo",
			PhoneNumber: "0900000003", AdditionalNotes: SamplePrefix + " Thiếu giá nước",
			MissingFields: []string{"water"},
			RawFields: map[string]string{
				"price": "5tr", "deposit": "2 tháng", "floor": "Tầng 5, có thang máy", "electricity": "Giá dân",
				"parking_fee": "Miễn phí", "pets_allowed": "Được nuôi mèo", "phone_number": "0900000003",
			},
		},
	}
}

// SeedSampleRentals inserts fake rentals for manual testing and returns how many
// it added. It adds none when sample rentals already exist.
func SeedSampleRentals() (n int, err error) {
	started := time.Now()
	log := logger.With("operation", "database.seed_sample_rentals")
	defer func() { logger.LogOperationResult(log, started, "seed_samples", err) }()

	var existing int
	if err = DB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM rentals WHERE "+isSampleRow, SamplePrefix, SamplePrefix).Scan(&existing); err != nil {
		return 0, fmt.Errorf("failed to count sample rentals: %w", err)
	}

	if existing > 0 {
		return 0, nil
	}

	for _, r := range sampleRentals() {
		if _, err = SaveRental(r); err != nil {
			return n, err
		}

		n++
	}

	return n, nil
}

// DeleteSampleRentals removes every sample rental and its decision trail in one
// transaction and returns how many rentals it deleted. Real rentals are untouched.
func DeleteSampleRentals() (n int, err error) {
	started := time.Now()
	log := logger.With("operation", "database.delete_sample_rentals")
	defer func() { logger.LogOperationResult(log, started, "delete_samples", err) }()

	ctx := context.Background()

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	defer func() { _ = tx.Rollback() }()

	if _, err = tx.ExecContext(ctx, "DELETE FROM decision_trails WHERE rental_id IN (SELECT id FROM rentals WHERE "+isSampleRow+")", SamplePrefix, SamplePrefix); err != nil {
		return 0, err
	}

	res, err := tx.ExecContext(ctx, "DELETE FROM rentals WHERE "+isSampleRow, SamplePrefix, SamplePrefix)
	if err != nil {
		return 0, err
	}

	deleted, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return int(deleted), nil
}
