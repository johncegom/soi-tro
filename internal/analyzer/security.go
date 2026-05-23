package analyzer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

const schemaSecretKey = "soi-tro-secure-openapi-schema-protection-key-2026"
const signatureKey = "x_signature"

// ErrSchemaTampered is returned when schema integrity check fails
var ErrSchemaTampered = errors.New("security error: schema.json has been tampered with or modified outside of the application")

// SignSchema bytes and return the signed JSON bytes
func SignSchema(data []byte) ([]byte, error) {
	var unified map[string]any
	if err := json.Unmarshal(data, &unified); err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema for signing: %w", err)
	}

	delete(unified, signatureKey)

	// Marshal to canonical JSON to ensure deterministic sorting of keys
	canonicalBytes, err := json.Marshal(unified)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal canonical schema: %w", err)
	}

	// Compute HMAC-SHA256
	mac := hmac.New(sha256.New, []byte(schemaSecretKey))
	mac.Write(canonicalBytes)
	signature := hex.EncodeToString(mac.Sum(nil))

	unified[signatureKey] = signature

	signedBytes, err := json.MarshalIndent(unified, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signed schema: %w", err)
	}

	return signedBytes, nil
}

// VerifySchema bytes against its embedded signature
func VerifySchema(data []byte) error {
	var unified map[string]any
	if err := json.Unmarshal(data, &unified); err != nil {
		return fmt.Errorf("failed to unmarshal schema for verification: %w", err)
	}

	sigRaw, ok := unified[signatureKey]
	if !ok {
		return ErrSchemaTampered
	}

	sigStr, ok := sigRaw.(string)
	if !ok || sigStr == "" {
		return ErrSchemaTampered
	}

	delete(unified, signatureKey)

	canonicalBytes, err := json.Marshal(unified)
	if err != nil {
		return fmt.Errorf("failed to marshal canonical schema for verification: %w", err)
	}

	mac := hmac.New(sha256.New, []byte(schemaSecretKey))
	mac.Write(canonicalBytes)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(sigStr), []byte(expectedSignature)) {
		return ErrSchemaTampered
	}

	return nil
}
