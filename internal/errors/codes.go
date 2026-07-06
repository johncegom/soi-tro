package errors

// Error codes for different types of errors
const (
	// Database errors
	ErrCodeDatabase        = "DB001"
	ErrCodeDatabaseConnect = "DB002"
	ErrCodeDatabaseQuery   = "DB003"
	
	// API errors
	ErrCodeAPI         = "API001"
	ErrCodeAPIAuth     = "API002"
	ErrCodeAPIRateLimit = "API003"
	ErrCodeAPITimeout  = "API004"
	
	// Validation errors
	ErrCodeValidation         = "VAL001"
	ErrCodeValidationRequired = "VAL002"
	ErrCodeValidationFormat   = "VAL003"
	
	// File system errors
	ErrCodeFileSystem     = "FS001"
	ErrCodeFileNotFound   = "FS002"
	ErrCodeFilePermission = "FS003"
	
	// Configuration errors
	ErrCodeConfiguration = "CFG001"
	ErrCodeConfigMissing = "CFG002"
	ErrCodeConfigInvalid = "CFG003"
	
	// General errors
	ErrCodeInternal = "INT001"
	ErrCodeNotFound = "NOT001"
)

// Predefined error instances for common errors
var (
	ErrDatabase = New(ErrCodeDatabase, "database operation failed")
	ErrAPI      = New(ErrCodeAPI, "API request failed")
	ErrValidation = New(ErrCodeValidation, "validation failed")
	ErrFileSystem = New(ErrCodeFileSystem, "file system operation failed")
	ErrConfiguration = New(ErrCodeConfiguration, "configuration error")
)
