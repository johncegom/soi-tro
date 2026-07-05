package database

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"soi-tro/internal/gemini"
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

	err = DeleteRental(id)
	require.NoError(t, err)

	records, err = ListRentals()
	require.NoError(t, err)
	assert.Len(t, records, 0)
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
	err := os.MkdirAll(filepath.Dir(configDir), 0700)
	require.NoError(t, err)
	err = os.WriteFile(configDir, []byte("some-file"), 0600)
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

