# Test Coverage Analysis for `soi-tro`

This document details the test coverage analysis of the `soi-tro` repository. The analysis was conducted by executing the unit tests with coverage profile generation and using `go tool cover` to dissect coverage on a per-package and per-function basis.

---

## Executive Summary

| Metric | Value | Status |
| :--- | :--- | :--- |
| **Total Statement Coverage** | **21.7%** | ⚠️ Needs Improvement |
| **Tested Packages** | 3 / 5 | 🟢 Exporter, Analyzer, Gemini |
| **Untested Packages** | 2 / 5 | 🔴 Cmd, UI |

### High-Level Summary
- **`internal/exporter`** (**81.2%**): Highly covered. Core business logic regarding file writes, rotation, and custom formatting has rigorous test coverage.
- **`internal/analyzer`** (**60.8%**): Good coverage. Key features like config signatures (`SignSchema` / `VerifySchema`) and `.env` parsing are heavily tested (including a fuzz test for `.env` parsing).
- **`internal/gemini`** (**40.7%**): Moderate coverage. Schema and export configuration loading/saving are well-tested, but functions interacting with the live Google GenAI API are not.
- **`internal/ui`** (**0.0%**): No coverage. Built on interactive terminal UI libraries (`bubbletea`, `huh`) and clipboard access, which are generally omitted from standard unit tests.
- **`cmd`** (**0.0%**): No coverage. Represents the CLI entry point (`main.go`).

---

## Detailed Coverage Breakdown

### 1. Per-Package & Per-Function Analysis

| Package / File | Function | Coverage % | Insights / Reasons |
| :--- | :--- | :--- | :--- |
| **`soi-tro/cmd`** | | **0.0%** | |
| └─ `main.go` | `main` | 0.0% | CLI entry point initiating the app and parsing arguments. |
| **`soi-tro/internal/ui`** | | **0.0%** | |
| ├─ `forms.go` | *All functions* | 0.0% | Manages interactive Bubble Tea / Huh forms. |
| ├─ `renderer.go` | *All functions* | 0.0% | Renders tables to `Stdout` and copies content to system clipboard. |
| └─ `schema_manager.go` | *All functions* | 0.0% | Drives terminal schema management wizard. |
| **`soi-tro/internal/gemini`** | | **40.7%** | |
| └─ `client.go` | `NewClient` | 0.0% | Depends on live `GEMINI_API_KEY` env var. |
| | `ExtractRentalInfo` | 0.0% | Interacts with live Gemini 3.1 Flash-Lite API endpoint. |
| | `LoadSchema` | 66.7% | Tested using mock home directories. |
| | `SaveSchema` | 76.0% | Tested using mock home directories. |
| | `LoadExportConfig` | 64.3% | Tested using mock home directories. |
| | `SaveExportConfig` | 66.7% | Tested using mock home directories. |
| **`soi-tro/internal/analyzer`** | | **60.8%** | |
| ├─ `engine.go` | `LoadConfig` / `SaveConfig` | 77.8% | Verified configuration storage and signature validations. |
| ├─ `env.go` | `ParseEnv` | 100.0% | Comprehensive tests, examples, and active **Fuzz test**. |
| | `LoadEnv` | 0.0% | Omitted from tests; loads `.env` files from disk. |
| ├─ `global_config.go` | `GetSchemaPath` | 75.0% | Tested with mock home folder structure. |
| | `EnsureSchemaFile` | 63.6% | Verified schema copy/migration checks. |
| | `GetGlobalConfigPath` | 75.0% | Tested with mock home folder structure. |
| | `LoadGlobalAPIKey` | 64.3% | Tested using mock configurations. |
| | `SaveGlobalAPIKey` | 75.0% | Tested using mock configurations. |
| | `EnsureGlobalAPIKey` | 0.0% | Prompts user interactively for Gemini API Key if missing. |
| └─ `security.go` | `SignSchema` | 80.0% | Cryptographic integrity validation. |
| | `VerifySchema` | 94.7% | Schema tempering detection checks. |
| **`soi-tro/internal/exporter`** | | **81.2%** | |
| └─ `exporter.go` | `DefaultConfig` | 100.0% | Default config presets checks. |
| | `WriteResult` | 61.1% | Verified results writing to directory. |
| | `ActiveFilePath` | 85.7% | Verified rollover file indexing. |
| | `appendToFile` | 71.4% | Handles low-level writes. |
| | `findActiveFile` | 80.0% | Scans for correct active segment file. |
| | `buildFilePath` | 100.0% | String formatting utility. |
| | `formatResult` | 96.0% | Markdown text structure compiler. |

---

## Observations & Recommendations

> [!NOTE]
> Lower coverage (21.7% overall) is primarily due to the codebase's strong reliance on **Interactive CLI inputs** (via `huh`/`bubbletea`) and **Live Cloud APIs** (via Google GenAI client).

### 🛠️ Strategic Suggestions to Improve Coverage

#### 1. Mocking the Gemini GenAI Client
Currently, `internal/gemini/client.go` directly initializes and uses `*genai.Client`. To test `ExtractRentalInfo`:
- Refactor the code to accept a Go `interface` representing the GenAI operations rather than the concrete struct.
- In unit tests, pass a mock implementation of that interface that returns hardcoded JSON mock responses. This will allow testing `ExtractRentalInfo` logic without triggering network requests or needing a real API Key.

#### 2. Unit Testing UI Elements via CLI Testers
`huh` and `bubbletea` provide mechanics for testing forms programmatically:
- We can pass simulated inputs (keyboard strokes) into `huh.Form` loops and verify the model's outputs.
- Consider moving business/validation logic away from `forms.go` and `schema_manager.go` into pure helper functions that are easily unit testable.

#### 3. Test `LoadEnv`
`LoadEnv` in `internal/analyzer/env.go` has 0% coverage. This can be tested easily by:
- Creating a temporary file in a unit test via `t.TempDir()`.
- Writing dummy variables inside it.
- Calling `LoadEnv` on that temp file path, and checking `os.Getenv` to make sure they were correctly loaded.
