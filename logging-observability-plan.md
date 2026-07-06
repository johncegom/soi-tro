# 🔍 Logging & Observability Implementation Plan - Soi Trọ CLI

Tài liệu này đề xuất kế hoạch chi tiết để tích hợp hệ thống logging và observability mạnh mẽ cho ứng dụng **Soi Trọ CLI**, giúp phát hiện lỗi nhanh chóng và cải thiện quy trình debug cho developer.

> **Note**: Tính năng này được đăng ký như **Task 6** trong `feature-plan.md` - [Xem kế hoạch tổng quan](feature-plan.md)

---

## 📊 Current State Analysis

### Existing Issues
- **Basic Logging**: Sử dụng `log` package cơ bản với cấu trúc kém
- **Inconsistent Error Messages**: Thông báo lỗi không đồng nhất giữa các package
- **No Log Levels**: Không phân loại log (debug, info, warn, error)
- **No Context Propagation**: Không có context để trace luồng hoạt động
- **No Centralized Error Handling**: Xử lý lỗi phân tán, khó quản lý
- **No Log Persistence**: Logs không được lưu vào file
- **Difficult Debugging**: Khó debug khi có lỗi production

### Current Logging Examples
```go
log.Fatalf("❌ Lỗi khởi tạo cơ sở dữ liệu lịch sử: %v", err)
log.Printf("❌ Lỗi nhận dữ liệu đầu vào: %v", err)
fmt.Printf("⚠️  Không thể đọc cấu hình xuất: %v\n", exportErr)
```

---

## 🎯 Proposed Logging & Observability System

### 1. Structured Logging Package (`internal/logger/`)

**Features:**
- Sử dụng Go's `log/slog` cho structured logging
- Hỗ trợ multiple log levels: Debug, Info, Warn, Error
- JSON và text output formats
- File và console output writers
- Context-aware logging với request tracing

**Key Components:**
```go
// Configuration
type Config struct {
    Level      string  // debug, info, warn, error
    Format     string  // json, text
    OutputPath string  // path to log file
    Console    bool    // enable console output
}

// Logger interface
type Logger interface {
    Debug(msg string, args ...any)
    Info(msg string, args ...any)
    Warn(msg string, args ...any)
    Error(msg string, args ...any)
    WithContext(ctx context.Context) Logger
    With(key string, value any) Logger
}
```

### 2. Centralized Error Handling

**Features:**
- Custom error types với error codes
- Error wrapping với context
- Consistent error message formatting
- Error tracking và alerting hooks

**Error Types:**
```go
type AppError struct {
    Code       string
    Message    string
    Cause      error
    Context    map[string]any
    StackTrace string
}

// Common error codes
const (
    ErrCodeDatabase      = "DB001"
    ErrCodeAPI           = "API001"
    ErrCodeValidation    = "VAL001"
    ErrCodeFileSystem    = "FS001"
    ErrCodeConfiguration = "CFG001"
)
```

### 3. Observability Features

**Request Tracing:**
- Request ID tracking cho traceability
- Correlation ID xuyên suốt request lifecycle
- Distributed tracing support cho tương lai

**Performance Monitoring:**
- Operation timing và duration tracking
- Function execution metrics
- Resource usage monitoring

**Alerting:**
- Critical error alerting
- Performance threshold alerts
- Health check integration

### 4. Configuration Management

**Environment Variables:**
```bash
# Log level (debug, info, warn, error)
LOG_LEVEL=info

# Log format (json, text)
LOG_FORMAT=json

# Log file path
LOG_FILE=~/.config/soi-tro/app.log

# Enable console output
LOG_CONSOLE=true

# Log rotation settings
LOG_MAX_SIZE=100MB
LOG_MAX_BACKUPS=5
LOG_MAX_AGE=30days
```

**Configuration File:**
```json
{
  "logging": {
    "level": "info",
    "format": "json",
    "output": {
      "file": "~/.config/soi-tro/app.log",
      "console": true
    },
    "rotation": {
      "max_size_mb": 100,
      "max_backups": 5,
      "max_age_days": 30
    }
  }
}
```

### 5. Integration Points

**Logging Enhancement Areas:**
- Database operations (queries, connections, errors)
- API calls (Gemini requests, responses, latency)
- File I/O operations (read, write, permissions)
- User interactions (form submissions, choices)
- Background processes (async operations)
- Configuration loading (schema, env, settings)

---

## 🚀 Implementation Plan

### Phase 1: Core Logging Infrastructure
**Timeline:** 2-3 hours

**Tasks:**
1. Create `internal/logger/` package
2. Implement structured logging using `log/slog`
3. Add configuration management
4. Implement file logging with rotation
5. Add console output with color coding
6. Write unit tests for logger package

**Deliverables:**
- `internal/logger/logger.go` - Main logger implementation
- `internal/logger/config.go` - Configuration management
- `internal/logger/writer.go` - Custom writers (file, console)
- `internal/logger/logger_test.go` - Unit tests

### Phase 2: Error Handling Enhancement
**Timeline:** 1-2 hours

**Tasks:**
1. Create custom error types
2. Implement error wrapping utilities
3. Add error context propagation
4. Create error code registry
5. Write unit tests for error handling

**Deliverables:**
- `internal/errors/errors.go` - Custom error types
- `internal/errors/codes.go` - Error code definitions
- `internal/errors/wrap.go` - Error wrapping utilities
- `internal/errors/errors_test.go` - Unit tests

### Phase 3: Application Integration
**Timeline:** 3-4 hours

**Tasks:**
1. Replace existing `log.Printf` calls with structured logging
2. Add request tracing with context
3. Implement performance monitoring
4. Integrate error handling throughout application
5. Add logging to major operations:
   - `cmd/main.go` - Application lifecycle
   - `internal/database/db.go` - Database operations
   - `internal/gemini/client.go` - API calls
   - `internal/exporter/exporter.go` - File operations
   - `internal/ui/` - User interactions

**Deliverables:**
- Updated all Go files with structured logging
- Context propagation in major operations
- Performance monitoring hooks

### Phase 4: Testing & Validation
**Timeline:** 1-2 hours

**Tasks:**
1. Write unit tests for logging package
2. Verify log output formats (JSON/text)
3. Test error scenarios
4. Validate log rotation
5. Performance testing for logging overhead
6. Integration testing

**Deliverables:**
- Comprehensive test coverage
- Log format validation
- Performance benchmarks
- Integration test results

---

## 📁 Project Structure

```
soi-tro/
├── internal/
│   ├── logger/
│   │   ├── logger.go       # Main logger implementation
│   │   ├── config.go       # Configuration management
│   │   ├── writer.go       # Custom writers
│   │   ├── context.go      # Context utilities
│   │   └── logger_test.go  # Unit tests
│   ├── errors/
│   │   ├── errors.go       # Custom error types
│   │   ├── codes.go        # Error code definitions
│   │   ├── wrap.go         # Error wrapping utilities
│   │   └── errors_test.go  # Unit tests
│   ├── analyzer/
│   ├── database/
│   ├── exporter/
│   ├── gemini/
│   └── ui/
├── .config/
│   └── soi-tro/
│       ├── app.log         # Application log file
│       ├── app.log.1       # Rotated log files
│       └── config.json     # Configuration with logging settings
└── logging-observability-plan.md
```

---

## 🔧 Technical Implementation Details

### Structured Logging Example

**Before:**
```go
log.Printf("❌ Lỗi khởi tạo cơ sở dữ liệu lịch sử: %v", err)
```

**After:**
```go
logger.Error("database initialization failed",
    "error", err,
    "path", dbPath,
    "operation", "InitDB",
)
```

**JSON Output:**
```json
{
  "time": "2025-01-15T10:30:45.123Z",
  "level": "ERROR",
  "msg": "database initialization failed",
  "error": "permission denied",
  "path": "/home/user/.config/soi-tro/rentals.db",
  "operation": "InitDB",
  "request_id": "req_abc123"
}
```

### Error Handling Example

**Before:**
```go
return fmt.Errorf("failed to read config file: %w", err)
```

**After:**
```go
return errors.Wrap(err, errors.ErrCodeConfiguration, 
    "failed to read config file",
    "path", filePath,
    "operation", "LoadConfig",
)
```

### Context Propagation Example

```go
func (c *Client) ExtractRentalInfo(ctx context.Context, ...) {
    logger := logger.FromContext(ctx).With(
        "operation", "ExtractRentalInfo",
        "model", modelName,
    )
    
    logger.Info("starting rental extraction")
    
    result, err := c.genaiClient.Models.GenerateContent(...)
    if err != nil {
        logger.Error("gemini API call failed",
            "error", err,
            "latency_ms", time.Since(start).Milliseconds(),
        )
        return nil, errors.Wrap(err, errors.ErrCodeAPI, "gemini API call failed")
    }
    
    logger.Info("rental extraction completed",
        "fields_extracted", len(result.RawFields),
        "latency_ms", time.Since(start).Milliseconds(),
    )
    
    return result, nil
}
```

---

## 🎨 Log Format Examples

### Development Mode (Text Format)
```
2025-01-15T10:30:45.123Z [INFO] starting rental extraction
2025-01-15T10:30:45.234Z [INFO] gemini API call completed
2025-01-15T10:30:45.235Z [INFO] rental extraction completed fields_extracted=8 latency_ms=112
```

### Production Mode (JSON Format)
```json
{"time":"2025-01-15T10:30:45.123Z","level":"INFO","msg":"starting rental extraction","operation":"ExtractRentalInfo","model":"gemini-3.1-flash-lite","request_id":"req_abc123"}
{"time":"2025-01-15T10:30:45.234Z","level":"INFO","msg":"gemini API call completed","operation":"ExtractRentalInfo","latency_ms":111,"request_id":"req_abc123"}
{"time":"2025-01-15T10:30:45.235Z","level":"INFO","msg":"rental extraction completed","operation":"ExtractRentalInfo","fields_extracted":8,"latency_ms":112,"request_id":"req_abc123"}
```

---

## ✅ Key Benefits

### Debugging
- ✅ Structured logs giúp dễ dàng filter và phân tích issues
- ✅ Rich context với từng log entry
- ✅ Error codes giúp nhanh chóng xác định loại lỗi

### Monitoring
- ✅ Real-time insight vào application health
- ✅ Performance metrics tracking
- ✅ Operation timing và bottleneck identification

### Traceability
- ✅ Request IDs giúp correlate logs xuyên suốt operations
- ✅ Context propagation theo luồng request
- ✅ Distributed tracing support cho tương lai

### Production Ready
- ✅ JSON logs cho log aggregation systems (ELK, Splunk, etc.)
- ✅ Log rotation tránh disk space issues
- ✅ Configurable log levels cho môi trường khác nhau

### Developer Friendly
- ✅ Human-readable text logs cho local development
- ✅ Color-coded console output
- ✅ Easy integration với existing debugging tools

---

## 🔄 Migration Strategy

### Step-by-Step Migration
1. **Infrastructure Setup**: Deploy logger package
2. **Gradual Replacement**: Replace logging calls package by package
3. **Testing**: Validate logs at each step
4. **Configuration**: Set up production logging configuration
5. **Monitoring**: Integrate với monitoring systems

### Backward Compatibility
- Maintain existing functionality during migration
- Use feature flags để enable/disable new logging
- Gradual rollout với monitoring

---

## 📝 Success Criteria

- [x] All packages sử dụng structured logging
- [x] Log levels configurable qua environment variables
- [x] JSON logs cho production, text logs cho development
- [x] File logging với automatic rotation
- [x] Context propagation xuyên suốt application
- [x] Custom error types với error codes
- [x] Performance monitoring hooks
- [x] Comprehensive unit test coverage
- [x] Zero breaking changes cho existing functionality
- [x] Documentation cho logger usage

---

## 🛠️ Dependencies

**New Dependencies:**
- `golang.org/x/exp/slog` (or standard `log/slog` in Go 1.21+)
- `lumberjack` cho log rotation (if needed)

**Existing Dependencies:**
- `google.golang.org/genai` (already in use)
- `modernc.org/sqlite` (already in use)

---

## 📚 Additional Resources

- [Go Structured Logging with slog](https://go.dev/blog/slog)
- [Twelve-Factor App - Logging](https://12factor.net/logs)
- [Observability Best Practices](https://www.oreilly.com/library/view/observability-engineering/9781492076217/)

---

## 🎯 Next Steps

1. **Review & Approve**: Review và approve implementation plan
2. **Phase 1 Implementation**: Implement core logging infrastructure
3. **Phase 2 Implementation**: Implement error handling enhancement
4. **Phase 3 Implementation**: Integrate logging throughout application
5. **Phase 4 Testing**: Comprehensive testing và validation
6. **Documentation**: Update usage documentation
7. **Deployment**: Deploy vào production với monitoring

---

## 📞 Contact & Support

For questions hoặc issues liên quan đến logging implementation:
- Check unit tests trong `internal/logger/logger_test.go`
- Review error codes trong `internal/errors/codes.go`
- Consult documentation trong `docs/logging.md` (to be created)
