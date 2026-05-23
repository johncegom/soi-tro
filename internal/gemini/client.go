package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/genai"
	"soi-tro/internal/analyzer"
)

// SampleMessage represents the generated follow-up message in a specific style.
type SampleMessage struct {
	Style   string `json:"style"`
	Content string `json:"content"`
}

// RentalExtractionResult maps to the Gemini API JSON output schema.
type RentalExtractionResult struct {
	Price           string            `json:"price"`
	Deposit         string            `json:"deposit"`
	Floor           string            `json:"floor"`
	Electricity     string            `json:"electricity"`
	Water           string            `json:"water"`
	ParkingFee      string            `json:"parking_fee"`
	PetsAllowed     string            `json:"pets_allowed"`
	PhoneNumber     string            `json:"phone_number"`
	AdditionalNotes string            `json:"additional_notes"`
	MissingFields   []string          `json:"missing_fields"`
	SampleMessages  []SampleMessage   `json:"sample_messages"`
	RawFields       map[string]string `json:"-"`
}

// Client wraps the official GenAI client.
type Client struct {
	genaiClient *genai.Client
}

// NewClient initializes the GenAI Client using standard configuration.
func NewClient(ctx context.Context) (*Client, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	// Initialize the official GenAI Client (picks up GEMINI_API_KEY from environment)
	genaiClient, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create GenAI client: %w", err)
	}

	return &Client{
		genaiClient: genaiClient,
	}, nil
}

// ExtractRentalInfo performs the parsing using the gemini-2.5-flash model.
func (c *Client) ExtractRentalInfo(ctx context.Context, text string, imageBytes []byte, imageMIME string, requiredFields []string) (*RentalExtractionResult, error) {
	var parts []*genai.Part

	if text != "" {
		parts = append(parts, &genai.Part{Text: text})
	}

	if len(imageBytes) > 0 {
		parts = append(parts, &genai.Part{
			InlineData: &genai.Blob{
				Data:     imageBytes,
				MIMEType: imageMIME,
			},
		})
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("either text listing or image file must be provided")
	}

	contents := []*genai.Content{
		{
			Parts: parts,
			Role:  "user",
		},
	}

	// Load response schema dynamically from file
	schemaPath, err := analyzer.GetSchemaPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get schema path: %w", err)
	}
	schema, err := LoadSchema(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load output schema (%s): %w", schemaPath, err)
	}

	// Dynamically align missing_fields schema enum and description with active config
	if missingSchema, ok := schema.Properties["missing_fields"]; ok {
		missingSchema.Items.Enum = requiredFields
		missingSchema.Description = fmt.Sprintf("Danh sách các trường thông tin bắt buộc bị thiếu trong bài đăng, chọn từ các khóa: %v.", requiredFields)
	}

	reqFieldsStr := fmt.Sprintf("%v", requiredFields)
	systemPrompt := fmt.Sprintf(`Trích xuất thông tin thuê phòng từ bài đăng (văn bản/hình ảnh) sang JSON theo schema.

Quy tắc chuẩn hóa:
- Giá/Cọc: "4tr5"/"4.5tr" -> "4,500,000 VND"; "cọc 1t"/"cọc 1 tháng" -> "Cọc 1 tháng"; "free" -> "Miễn phí".
- Điện/Nước: "4k/số" -> "4,000 VND/kWh"; "100k/ng" -> "100,000 VND/người".
- Phí xe: "free xe" -> "Miễn phí giữ xe"; "xe 100k" -> "100,000 VND/xe/tháng".
- SĐT: Chuỗi số liền mạch (ví dụ "0912345678"). Không có ghi "Không đề cập".

Phân tích khoảng trống thông tin (GAP analysis):
- Xác định các trường bắt buộc bị thiếu trong bài đăng: %s.
- Điền các trường thiếu vào 'missing_fields' từ danh sách: %s.

Tin nhắn mẫu (trong 'sample_messages', chính xác 2 tin):
- Địa chỉ: Các tin nhắn mẫu BẮT BUỘC phải tích hợp địa chỉ/vị trí của phòng trọ/căn hộ được đề cập trong bài đăng (ví dụ: tên đường, ngõ, hoặc khu vực cụ thể) để chủ nhà dễ dàng xác định người dùng đang hỏi về phòng/nhà nào (do chủ nhà thường có nhiều phòng cho thuê). Nếu bài đăng hoàn toàn không có thông tin địa chỉ, hãy sử dụng phần giữ chỗ như "[Địa chỉ bài đăng]".
- Xưng hô: Trong cả hai tin nhắn mẫu, luôn luôn sử dụng "mình" để xưng hô cho người hỏi (user) và "b" hoặc "bạn" cho chủ nhà (landlord). Tuyệt đối không sử dụng các từ xưng hô khác như "em", "anh", "chị", "tôi".
1. "Lịch sự": Lịch sự, giới thiệu bản thân xưng "mình", bày tỏ sự quan tâm, đề cập rõ địa chỉ phòng đang hỏi, hỏi chi tiết các thông tin thiếu xưng hô với "b" hoặc "bạn".
2. "Trực diện": Ngắn gọn, trực diện, đi thẳng vào vấn đề (như nhắn Zalo/SMS nhanh), đề cập rõ địa chỉ phòng đang hỏi, xưng "mình" và xưng hô với "b/bạn".`, reqFieldsStr, reqFieldsStr)

	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   schema,
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: systemPrompt},
			},
		},
	}

	resp, err := c.genaiClient.Models.GenerateContent(ctx, "gemini-2.5-flash", contents, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate content from Gemini API: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response candidates returned by Gemini")
	}

	responseText := resp.Candidates[0].Content.Parts[0].Text

	var result RentalExtractionResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response from model (raw response was %q): %w", responseText, err)
	}

	// Dynamic parsing to capture any user-added fields
	var rawMap map[string]json.RawMessage
	if err := json.Unmarshal([]byte(responseText), &rawMap); err == nil {
		result.RawFields = make(map[string]string)
		for k, v := range rawMap {
			if k == "missing_fields" || k == "sample_messages" {
				continue
			}
			var s string
			if err := json.Unmarshal(v, &s); err == nil {
				result.RawFields[k] = s
			} else {
				result.RawFields[k] = string(v)
			}
		}
	}

	return &result, nil
}

// LoadSchema loads the OpenAPI 3.0 schema from a JSON file.
func LoadSchema(filePath string) (*genai.Schema, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open schema file: %w", err)
	}
	defer file.Close()

	var schema genai.Schema
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&schema); err != nil {
		return nil, fmt.Errorf("failed to decode schema JSON: %w", err)
	}

	return &schema, nil
}

// SaveSchema writes the genai.Schema and custom config back to a JSON file path.
func SaveSchema(filePath string, schema *genai.Schema, xRequiredFields []string) error {
	schemaBytes, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	var unified map[string]any
	if err := json.Unmarshal(schemaBytes, &unified); err != nil {
		return fmt.Errorf("failed to unmarshal schema to map: %w", err)
	}

	unified["x_required_fields"] = xRequiredFields

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open schema file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(unified); err != nil {
		return fmt.Errorf("failed to encode unified JSON: %w", err)
	}

	return nil
}

// exportConfigKey is the top-level key in schema.json that stores export settings.
const exportConfigKey = "x_export_config"

// LoadExportConfig reads x_export_config from a schema JSON file.
// Returns the config, a boolean indicating whether it was previously configured,
// and any I/O or parse error.
func LoadExportConfig(filePath string) (exportCfg struct {
	Dir       string `json:"dir"`
	MaxSizeKB int    `json:"max_size_kb"`
}, configured bool, err error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return exportCfg, false, fmt.Errorf("failed to read schema file: %w", err)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return exportCfg, false, fmt.Errorf("failed to parse schema file: %w", err)
	}

	blob, ok := raw[exportConfigKey]
	if !ok {
		return exportCfg, false, nil
	}

	if err := json.Unmarshal(blob, &exportCfg); err != nil {
		return exportCfg, false, fmt.Errorf("failed to parse export config: %w", err)
	}

	return exportCfg, true, nil
}

// SaveExportConfig writes export settings into the x_export_config key of schema.json,
// preserving all other existing top-level keys.
func SaveExportConfig(filePath string, dir string, maxSizeKB int) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	var unified map[string]any
	if err := json.Unmarshal(data, &unified); err != nil {
		return fmt.Errorf("failed to parse schema file: %w", err)
	}

	unified[exportConfigKey] = map[string]any{
		"dir":         dir,
		"max_size_kb": maxSizeKB,
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("failed to open schema file for writing: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(unified); err != nil {
		return fmt.Errorf("failed to encode schema with export config: %w", err)
	}

	return nil
}

