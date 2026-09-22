package main

import (
	"path/filepath"
	"strings"
)

type action int

const (
	actionAnalyze action = iota
	actionHistory
	actionManage
	actionExport
	actionModel
	actionExit
)

// dispatch maps a main-menu value to an action; anything else analyzes.
func dispatch(choice string) action {
	switch choice {
	case "history":
		return actionHistory
	case "manage":
		return actionManage
	case "export":
		return actionExport
	case "model":
		return actionModel
	case "exit":
		return actionExit
	default:
		return actionAnalyze
	}
}

type errorStep int

const (
	stepRetry errorStep = iota
	stepChangeInput
	stepMenu
)

// onAnalysisError maps a PromptErrorRetry value to the next step of the analyze loop.
func onAnalysisError(choice string) errorStep {
	switch choice {
	case "retry":
		return stepRetry
	case "change":
		return stepChangeInput
	default:
		return stepMenu
	}
}

// mimeTypeFor returns the image MIME type for a file path, defaulting to JPEG.
func mimeTypeFor(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg" // .jpg hoặc .jpeg
	}
}
