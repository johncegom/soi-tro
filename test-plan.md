# Test Plan: Increasing Unit Test Coverage

This document outlines the strategy and specific test cases designed to fill unit test coverage gaps in the `soi-tro` project.

## 1. Objectives
- Maximize test coverage in critical `internal` packages (`analyzer`, `exporter`, `gemini`).
- Cover edge cases, error branches, and key initialization/configuration routines without relying on live external networks or APIs.
- Ensure the codebase behaves robustly during initialization and error scenarios.

## 2. Target Areas & Gaps

| Package | File | Target Functions | Current Coverage | Goal Coverage |
|---|---|---|---|---|
| `internal/analyzer` | `env.go` | `LoadEnv` | 0.0% | ~100.0% |
| `internal/analyzer` | `global_config.go` | `EnsureGlobalAPIKey`, `LoadGlobalAPIKey`, `GetSchemaPath`, `EnsureSchemaFile` | 68.6% (package) | >85.0% |
| `internal/exporter` | `exporter.go` | `WriteResult` (directory creation failure), `appendToFile` (file opening failure) | 81.2% (package) | >90.0% |
| `internal/gemini` | `client.go` | `NewClient`, `LoadSchema`, `LoadExportConfig`, `SaveExportConfig`, `ExtractRentalInfo` (pre-network fails) | 40.7% (package) | >80.0% |

## 3. Detailed Test Cases

### 3.1. Package `internal/analyzer`
1. **`TestLoadEnv`**:
   - Verify that calling `LoadEnv` with a non-existent file returns an error.
   - Verify that loading a valid `.env` file successfully populates the environment variables.
2. **`TestEnsureGlobalAPIKey_EnvExists`**:
   - Verify that if `GEMINI_API_KEY` is already in the environment, `EnsureGlobalAPIKey` returns early and successfully.
3. **`TestEnsureGlobalAPIKey_FromConfig`**:
   - Verify that if the environment key is empty but the key exists in the global config file, it loads the key and sets the environment.
4. **`TestLoadGlobalAPIKey_InvalidJSON`**:
   - Verify that loading a malformed global config file returns a JSON parsing error.
5. **`TestGetSchemaPath_Error`**:
   - Verify that if the home directory lookup fails (by clearing all home directory environment variables), an error is returned.
6. **`TestEnsureSchemaFile_Migration`**:
   - Verify the migration flow where a schema file exists but is not signed. The function should sign the existing schema and write it back.

### 3.2. Package `internal/exporter`
1. **`TestWriteResult_DirCreationFailure`**:
   - Set up an exporter config with a directory path that is nested under a file (making `MkdirAll` fail). Verify that `WriteResult` fails with the expected error.
2. **`TestWriteResult_FileOpenFailure`**:
   - Create a directory with the same name as the active file output path (making `OpenFile` fail with a directory-path error). Verify that `WriteResult` returns the expected error.

### 3.3. Package `internal/gemini`
1. **`TestNewClient_NoAPIKey`**:
   - Verify that `NewClient` fails when `GEMINI_API_KEY` is empty.
2. **`TestNewClient_Success`**:
   - Verify that `NewClient` succeeds when `GEMINI_API_KEY` is set.
3. **`TestLoadSchema_FileNotExist` & `TestLoadSchema_InvalidSignature` & `TestLoadSchema_InvalidJSON`**:
   - Verify all edge-case error branches when reading the OpenAPI 3.0 schema file.
4. **`TestLoadExportConfig_FileNotExist` & `TestLoadExportConfig_InvalidSignature` & `TestLoadExportConfig_InvalidJSON` & `TestLoadExportConfig_InvalidConfigType`**:
   - Verify all edge-case error branches when parsing export configurations from the schema file.
5. **`TestSaveExportConfig_FileNotExist` & `TestSaveExportConfig_Untrusted` & `TestSaveExportConfig_InvalidJSON`**:
   - Verify all edge-case error branches when saving export configuration.
6. **`TestExtractRentalInfo_EmptyInputs` & `TestExtractRentalInfo_SchemaLoadFail` & `TestExtractRentalInfo_GetSchemaPathFail`**:
   - Verify fast-fail logic of the AI extraction engine before any external API request is made.

## 4. Execution Plan
1. Present the detailed test changes to the user and obtain approval.
2. Apply the tests to:
   - `internal/analyzer/env_test.go`
   - `internal/analyzer/global_config_test.go`
   - `internal/exporter/exporter_test.go`
   - `internal/gemini/client_test.go`
3. Execute `go test -coverprofile="coverage.out" ./...`.
4. Analyze the new coverage via `go tool cover -func="coverage.out"`.
