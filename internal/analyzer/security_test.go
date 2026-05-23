package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityIntegrity(t *testing.T) {
	t.Parallel()

	// Initial valid test data representing a typical schema.json format
	baseJSON := `{
		"type": "object",
		"properties": {
			"price": {
				"type": "string",
				"title": "Giá thuê"
			}
		},
		"x_required_fields": ["price"]
	}`

	t.Run("valid signature", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// 1. Sign the base payload
		signedBytes, err := SignSchema([]byte(baseJSON))
		is.NoError(err, "SignSchema should succeed on valid JSON")
		is.NotEmpty(signedBytes, "Signed output should not be empty")

		// 2. Verify it passes
		err = VerifySchema(signedBytes)
		is.NoError(err, "VerifySchema should succeed on valid signed schema")
	})

	t.Run("deterministic output", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Sign twice and ensure signature bytes match exactly (proves stable sorting)
		signedBytes1, err := SignSchema([]byte(baseJSON))
		is.NoError(err)

		signedBytes2, err := SignSchema([]byte(baseJSON))
		is.NoError(err)

		is.Equal(signedBytes1, signedBytes2, "Deterministic marshaling should yield identical outputs")
	})

	t.Run("missing signature", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Verify raw unsigned JSON directly
		err := VerifySchema([]byte(baseJSON))
		is.ErrorIs(err, ErrSchemaTampered, "VerifySchema must fail on unsigned schema")
	})

	t.Run("invalid/altered signature", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// 1. Sign
		signedBytes, err := SignSchema([]byte(baseJSON))
		is.NoError(err)

		// 2. Decode, alter signature, encode back
		var unified map[string]any
		err = json.Unmarshal(signedBytes, &unified)
		is.NoError(err)

		// Alter one byte of the signature string
		sigStr := unified[signatureKey].(string)
		alteredSig := sigStr[:len(sigStr)-1] + "x"
		unified[signatureKey] = alteredSig

		alteredBytes, err := json.Marshal(unified)
		is.NoError(err)

		// 3. Verify must fail
		err = VerifySchema(alteredBytes)
		is.ErrorIs(err, ErrSchemaTampered, "VerifySchema must reject altered signature strings")
	})

	t.Run("tampered payload", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// 1. Sign
		signedBytes, err := SignSchema([]byte(baseJSON))
		is.NoError(err)

		// 2. Decode, modify properties (add property), encode back
		var unified map[string]any
		err = json.Unmarshal(signedBytes, &unified)
		is.NoError(err)

		properties := unified["properties"].(map[string]any)
		properties["deposit"] = map[string]any{
			"type":  "string",
			"title": "Tiền cọc",
		}
		unified["properties"] = properties

		tamperedBytes, err := json.Marshal(unified)
		is.NoError(err)

		// 3. Verify must fail
		err = VerifySchema(tamperedBytes)
		is.ErrorIs(err, ErrSchemaTampered, "VerifySchema must reject a tampered payload even with a valid old signature key")
	})

	t.Run("empty signature", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		var unified map[string]any
		err := json.Unmarshal([]byte(baseJSON), &unified)
		is.NoError(err)

		// Put empty string
		unified[signatureKey] = ""
		emptySigBytes, err := json.Marshal(unified)
		is.NoError(err)

		err = VerifySchema(emptySigBytes)
		is.ErrorIs(err, ErrSchemaTampered, "VerifySchema must reject empty signature keys")
	})

	t.Run("malformed JSON", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Corrupt bytes
		err := VerifySchema([]byte("{corrupted-json...}"))
		is.Error(err, "VerifySchema must return parsing error on malformed JSON")
		is.True(strings.Contains(err.Error(), "unmarshal"), "expected unmarshal parsing error, got %q", err.Error())
	})

	t.Run("format variations (whitespace immune)", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// 1. Sign
		signedBytes, err := SignSchema([]byte(baseJSON))
		is.NoError(err)

		// 2. Unmarshal
		var unified map[string]any
		err = json.Unmarshal(signedBytes, &unified)
		is.NoError(err)

		// 3. Re-encode with extreme whitespaces, tabs, and newlines
		formattedBytes, err := json.MarshalIndent(unified, "\t\t", "    ")
		is.NoError(err)

		// 4. Verify should still succeed since it unmarshals first and marshals back canonically
		err = VerifySchema(formattedBytes)
		is.NoError(err, "VerifySchema should be immune to JSON spacing, tabs, or newlines modifications")
	})

	t.Run("key ordering variations (order immune)", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// 1. Sign
		signedBytes, err := SignSchema([]byte(baseJSON))
		is.NoError(err)

		var unified map[string]any
		err = json.Unmarshal(signedBytes, &unified)
		is.NoError(err)

		// Get the signature value
		sigStr := unified[signatureKey].(string)

		// 2. Create custom JSON manually with completely different key order
		// Here signature Key is placed FIRST, and x_required_fields is LAST.
		customOrderJSON := fmt.Sprintf(`{
			"x_signature": %q,
			"properties": {
				"price": {
					"title": "Giá thuê",
					"type": "string"
				}
			},
			"type": "object",
			"x_required_fields": ["price"]
		}`, sigStr)

		// 3. Verify must succeed since key order is normalized alphabetically before signature calculation
		err = VerifySchema([]byte(customOrderJSON))
		is.NoError(err, "VerifySchema should succeed regardless of JSON key order variations")
	})

	t.Run("signature type mismatch", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Test cases for invalid non-string signature types
		testCases := []struct {
			name  string
			value any
		}{
			{name: "integer signature", value: 123456789},
			{name: "boolean signature", value: true},
			{name: "array signature", value: []any{"sig1", "sig2"}},
			{name: "object signature", value: map[string]any{"val": "sig"}},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				var unified map[string]any
				err := json.Unmarshal([]byte(baseJSON), &unified)
				is.NoError(err)

				unified[signatureKey] = tc.value
				payloadBytes, err := json.Marshal(unified)
				is.NoError(err)

				err = VerifySchema(payloadBytes)
				is.ErrorIs(err, ErrSchemaTampered, "VerifySchema must reject non-string signature types")
			})
		}
	})

	t.Run("empty JSON object", func(t *testing.T) {
		t.Parallel()
		is := assert.New(t)

		// Sign an empty JSON object
		signedBytes, err := SignSchema([]byte("{}"))
		is.NoError(err)

		err = VerifySchema(signedBytes)
		is.NoError(err, "VerifySchema should successfully sign and verify an empty JSON object")
	})
}
