package database

import (
	"context"
	"os"
	"runtime"
	"soi-tro/internal/gemini"
	"soi-tro/internal/viewing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTrailRepo(t *testing.T) *DecisionTrailStore {
	t.Helper()
	mockDBPath(t)
	require.NoError(t, InitDB())

	return NewDecisionTrailStore(DB)
}

func TestDecisionTrailGetMissingReturnsEmptyTrail(t *testing.T) {
	repo := newTrailRepo(t)

	trail, err := repo.Get(context.Background(), 42)

	require.NoError(t, err)
	assert.Equal(t, viewing.NewTrail(42), trail)
}

func TestDecisionTrailSaveGetRoundTrip(t *testing.T) {
	repo := newTrailRepo(t)
	ctx := context.Background()

	due := time.Date(2026, 10, 12, 9, 30, 0, 0, time.FixedZone("ICT", 7*3600))
	trail := viewing.NewTrail(7)
	require.NoError(t, trail.SetStatus(viewing.StatusViewingPlanned, &due))
	require.NoError(t, trail.SetAnswer("field:deposit", "Cọc 1 tháng, hoàn khi trả phòng"))
	require.NoError(t, trail.SetAnswer("default:noise", "Ồn vào buổi tối"))

	require.NoError(t, repo.Save(ctx, trail))

	got, err := repo.Get(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, viewing.StatusViewingPlanned, got.Status)
	require.NotNil(t, got.NextAction)
	assert.True(t, due.Equal(*got.NextAction), "next action instant must survive a round trip")
	assert.Equal(t, trail.Answers, got.Answers)
}

func TestDecisionTrailSaveReplacesAnswersAndClearsDate(t *testing.T) {
	repo := newTrailRepo(t)
	ctx := context.Background()

	due := time.Date(2026, 10, 12, 0, 0, 0, 0, time.UTC)
	trail := viewing.NewTrail(7)
	require.NoError(t, trail.SetStatus(viewing.StatusConsidering, &due))
	require.NoError(t, trail.SetAnswer("field:deposit", "1 tháng"))
	require.NoError(t, trail.SetAnswer("field:water", "100k/người"))
	require.NoError(t, repo.Save(ctx, trail))

	trail.ClearAnswer("field:water")
	require.NoError(t, trail.SetStatus(viewing.StatusRejected, nil))
	require.NoError(t, repo.Save(ctx, trail))

	got, err := repo.Get(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, viewing.StatusRejected, got.Status)
	assert.Nil(t, got.NextAction)
	assert.Equal(t, map[viewing.ItemID]string{"field:deposit": "1 tháng"}, got.Answers)
}

func TestDecisionTrailSaveRejectsInvalidStatus(t *testing.T) {
	repo := newTrailRepo(t)

	trail := viewing.NewTrail(7)
	trail.Status = "bogus"

	err := repo.Save(context.Background(), trail)

	require.ErrorIs(t, err, viewing.ErrInvalidStatus)
}

func TestDecisionTrailDeleteRemovesOnlyThatTrail(t *testing.T) {
	repo := newTrailRepo(t)
	ctx := context.Background()

	for _, id := range []int64{1, 2} {
		trail := viewing.NewTrail(id)
		require.NoError(t, trail.SetAnswer("field:price", "5tr"))
		require.NoError(t, repo.Save(ctx, trail))
	}

	require.NoError(t, repo.Delete(ctx, 1))
	require.NoError(t, repo.Delete(ctx, 99), "deleting a missing trail is not an error")

	gone, err := repo.Get(ctx, 1)
	require.NoError(t, err)
	assert.Empty(t, gone.Answers)

	kept, err := repo.Get(ctx, 2)
	require.NoError(t, err)
	assert.Equal(t, "5tr", kept.Answers["field:price"])
}

func TestDecisionTrailSurvivesReopen(t *testing.T) {
	repo := newTrailRepo(t)
	ctx := context.Background()

	trail := viewing.NewTrail(3)
	require.NoError(t, trail.SetAnswer("default:noise", "yên tĩnh"))
	require.NoError(t, repo.Save(ctx, trail))

	require.NoError(t, DB.Close())
	DB = nil
	require.NoError(t, InitDB())

	got, err := NewDecisionTrailStore(DB).Get(ctx, 3)
	require.NoError(t, err)
	assert.Equal(t, "yên tĩnh", got.Answers["default:noise"])
}

func TestDeleteRentalAndTrailRemovesBoth(t *testing.T) {
	repo := newTrailRepo(t)
	ctx := context.Background()

	keepID, err := SaveRental(&gemini.RentalExtractionResult{Price: "4tr"})
	require.NoError(t, err)
	delID, err := SaveRental(&gemini.RentalExtractionResult{Price: "5tr"})
	require.NoError(t, err)

	for _, id := range []int64{keepID, delID} {
		trail := viewing.NewTrail(id)
		require.NoError(t, trail.SetAnswer("field:price", "ghi chú riêng"))
		require.NoError(t, repo.Save(ctx, trail))
	}

	require.NoError(t, DeleteRentalAndTrail(ctx, delID))

	records, err := ListRentals()
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, keepID, records[0].ID)

	gone, err := repo.Get(ctx, delID)
	require.NoError(t, err)
	assert.Empty(t, gone.Answers, "no orphaned notes may remain")

	kept, err := repo.Get(ctx, keepID)
	require.NoError(t, err)
	assert.NotEmpty(t, kept.Answers)
}

func TestInitDBMakesDatabaseFileOwnerOnly(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX file modes are not enforced on Windows")
	}

	mockDBPath(t)
	require.NoError(t, InitDB())

	path, err := GetDBPath()
	require.NoError(t, err)

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}
