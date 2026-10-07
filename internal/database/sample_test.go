package database

import (
	"context"
	"soi-tro/internal/gemini"
	"soi-tro/internal/viewing"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func initSampleDB(t *testing.T) {
	t.Helper()
	mockDBPath(t)
	require.NoError(t, InitDB())
}

func TestSeedSampleRentalsInsertsTaggedRows(t *testing.T) {
	initSampleDB(t)

	n, err := SeedSampleRentals()
	require.NoError(t, err)
	require.Positive(t, n)

	records, err := ListRentals()
	require.NoError(t, err)
	assert.Len(t, records, n)

	var withMissing, withClaim bool
	for _, rec := range records {
		assert.True(t, strings.HasPrefix(rec.Result.AdditionalNotes, SamplePrefix), "row %d is not tagged", rec.ID)
		withMissing = withMissing || len(rec.Result.MissingFields) > 0
		withClaim = withClaim || rec.Result.Price != ""
	}
	assert.True(t, withMissing, "samples need a rental with missing fields")
	assert.True(t, withClaim, "samples need a rental with a listing claim")
}

func TestSeedSampleRentalsTwiceDoesNotDuplicate(t *testing.T) {
	initSampleDB(t)

	first, err := SeedSampleRentals()
	require.NoError(t, err)
	require.Positive(t, first)

	second, err := SeedSampleRentals()
	require.NoError(t, err)
	assert.Zero(t, second)

	records, err := ListRentals()
	require.NoError(t, err)
	assert.Len(t, records, first)
}

func TestDeleteSampleRentalsKeepsRealData(t *testing.T) {
	initSampleDB(t)
	ctx := context.Background()
	store := NewDecisionTrailStore(DB)

	realID, err := SaveRental(&gemini.RentalExtractionResult{Price: "3tr", AdditionalNotes: "real notes"})
	require.NoError(t, err)

	realTrail := viewing.NewTrail(realID)
	require.NoError(t, realTrail.SetAnswer("default:noise", "yên"))
	require.NoError(t, store.Save(ctx, realTrail))

	seeded, err := SeedSampleRentals()
	require.NoError(t, err)
	require.Positive(t, seeded)

	records, err := ListRentals()
	require.NoError(t, err)

	var sampleID int64
	for _, rec := range records {
		if rec.ID != realID {
			sampleID = rec.ID
		}
	}

	sampleTrail := viewing.NewTrail(sampleID)
	require.NoError(t, sampleTrail.SetAnswer("default:noise", "ồn"))
	require.NoError(t, store.Save(ctx, sampleTrail))

	deleted, err := DeleteSampleRentals()
	require.NoError(t, err)
	assert.Equal(t, seeded, deleted)

	records, err = ListRentals()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, realID, records[0].ID)

	got, err := store.Get(ctx, realID)
	require.NoError(t, err)
	assert.Equal(t, "yên", got.Answers["default:noise"])

	gone, err := store.Get(ctx, sampleID)
	require.NoError(t, err)
	assert.Empty(t, gone.Answers, "sample trail should be deleted with its rental")
}

func TestDeleteSampleRentalsWithNoneDeletesNothing(t *testing.T) {
	initSampleDB(t)

	_, err := SaveRental(&gemini.RentalExtractionResult{Price: "3tr", AdditionalNotes: "real notes"})
	require.NoError(t, err)

	deleted, err := DeleteSampleRentals()
	require.NoError(t, err)
	assert.Zero(t, deleted)

	records, err := ListRentals()
	require.NoError(t, err)
	assert.Len(t, records, 1)
}
