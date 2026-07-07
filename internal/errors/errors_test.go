package errors

import (
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("TEST001", "test error")
	
	if err.Code != "TEST001" {
		t.Errorf("Expected code to be 'TEST001', got '%s'", err.Code)
	}
	
	if err.Message != "test error" {
		t.Errorf("Expected message to be 'test error', got '%s'", err.Message)
	}
	
	if err.Cause != nil {
		t.Error("Expected cause to be nil")
	}
}

func TestWrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "WRAP001", "wrapped error")
	
	if wrappedErr.Code != "WRAP001" {
		t.Errorf("Expected code to be 'WRAP001', got '%s'", wrappedErr.Code)
	}
	
	if wrappedErr.Message != "wrapped error" {
		t.Errorf("Expected message to be 'wrapped error', got '%s'", wrappedErr.Message)
	}
	
	if wrappedErr.Cause == nil {
		t.Error("Expected cause to be set")
	}
	
	if wrappedErr.Cause != originalErr {
		t.Error("Expected cause to be the original error")
	}
}

func TestWrapNil(t *testing.T) {
	wrappedErr := Wrap(nil, "WRAP001", "wrapped error")
	
	if wrappedErr != nil {
		t.Error("Expected wrapped error to be nil when wrapping nil")
	}
}

func TestWith(t *testing.T) {
	err := New("TEST001", "test error")
	err = err.With("key", "value")
	
	if err.Context["key"] != "value" {
		t.Errorf("Expected context key to be 'value', got '%v'", err.Context["key"])
	}
}

func TestError(t *testing.T) {
	err := New("TEST001", "test error")
	errorStr := err.Error()
	
	expected := "[TEST001] test error"
	if errorStr != expected {
		t.Errorf("Expected error string to be '%s', got '%s'", expected, errorStr)
	}
}

func TestErrorWithCause(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "WRAP001", "wrapped error")
	errorStr := wrappedErr.Error()
	
	expected := "[WRAP001] wrapped error: original error"
	if errorStr != expected {
		t.Errorf("Expected error string to be '%s', got '%s'", expected, errorStr)
	}
}

func TestUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, "WRAP001", "wrapped error")
	
	unwrapped := errors.Unwrap(wrappedErr)
	if unwrapped != originalErr {
		t.Error("Expected unwrapped error to be the original error")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		expectedCode string
	}{
		{"ErrDatabase", ErrDatabase, ErrCodeDatabase},
		{"ErrAPI", ErrAPI, ErrCodeAPI},
		{"ErrValidation", ErrValidation, ErrCodeValidation},
		{"ErrFileSystem", ErrFileSystem, ErrCodeFileSystem},
		{"ErrConfiguration", ErrConfiguration, ErrCodeConfiguration},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.expectedCode {
				t.Errorf("Expected error code to be '%s', got '%s'", tt.expectedCode, tt.err.Code)
			}
		})
	}
}
