package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDispatch(t *testing.T) {
	tests := []struct {
		choice string
		want   action
	}{
		{"analyze", actionAnalyze},
		{"history", actionHistory},
		{"manage", actionManage},
		{"export", actionExport},
		{"model", actionModel},
		{"exit", actionExit},
		{"", actionAnalyze},
		{"unknown", actionAnalyze},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, dispatch(tt.choice), "choice %q", tt.choice)
	}
}

// BUG-006: "change" must go back to the input form, not fall through.
func TestOnAnalysisError(t *testing.T) {
	tests := []struct {
		choice string
		want   errorStep
	}{
		{"retry", stepRetry},
		{"change", stepChangeInput},
		{"back", stepMenu},
		{"", stepMenu},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, onAnalysisError(tt.choice), "choice %q", tt.choice)
	}
}

// BUG-007: "new" must go back to the input form, not re-analyze the same listing.
func TestOnAnalysisSuccess(t *testing.T) {
	tests := []struct {
		choice string
		want   successStep
	}{
		{"export", stepExport},
		{"new", stepNewInput},
		{"back", stepBackToMenu},
		{"", stepBackToMenu},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, onAnalysisSuccess(tt.choice), "choice %q", tt.choice)
	}
}

func TestMimeTypeFor(t *testing.T) {
	tests := map[string]string{
		"a.png":        "image/png",
		"A.PNG":        "image/png",
		"dir/b.webp":   "image/webp",
		"c.jpg":        "image/jpeg",
		"c.jpeg":       "image/jpeg",
		"noextension":  "image/jpeg",
		"weird.tar.gz": "image/jpeg",
	}
	for path, want := range tests {
		assert.Equal(t, want, mimeTypeFor(path), path)
	}
}
