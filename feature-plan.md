# 🚀 Kế hoạch Phát triển Tính năng (Feature Plan) - Soi Trọ CLI

Tài liệu này đề xuất lộ trình và kế hoạch chi tiết để tích hợp thêm các tính năng hữu ích cho ứng dụng dòng lệnh **Soi Trọ** (CLI), giúp hỗ trợ người tìm kiếm và phân tích thông tin phòng trọ hiệu quả hơn.

---

## 📌 Danh sách các Task (Task List)

- [x] [CONFIRMED] **Task 1: Tích Hợp Cơ Sở Dữ Liệu SQLite Cục Bộ (History Database)**
  - *Mô tả*: Tự động lưu trữ kết quả phân tích phòng trọ của mỗi lượt chạy vào cơ sở dữ liệu SQLite cục bộ thay vì chỉ xuất tệp văn bản.
  - *Giải pháp kỹ thuật*: Sử dụng thư viện SQLite viết hoàn toàn bằng Go (pure-Go port như `modernc.org/sqlite`). Cơ sở dữ liệu sẽ tự động được khởi tạo tại `~/.config/soi-tro/rentals.db`.
  - *Trạng thái*: Hoàn thành.
  - *Phụ thuộc*: Không phụ thuộc vào task nào (Prerequisite cho Task 3).

- [x] [CONFIRMED] **Task 3: So Sánh Phòng Trọ Song Song (Side-by-Side Comparison UI)**
  - *Mô tả*: Cho phép người dùng lựa chọn từ 2 đến 3 phòng trọ từ lịch sử lưu trữ để đối chiếu trực quan theo cột.
  - *Giải pháp kỹ thuật*: Thiết kế giao diện so sánh nhiều cột trực quan trên Terminal UI bằng cách mở rộng `tablewriter` hoặc thiết lập bố cục layout tùy biến của Bubble Tea.
  - *Trạng thái*: Hoàn thành.
  - *Phụ thuộc*: **Bị chặn bởi/Phụ thuộc vào Task 1** (cần cơ sở dữ liệu SQLite để lưu trữ và tải danh sách phòng trọ lịch sử).

- [ ] [ASSUMPTION] **Task 2: Quét & Phân Tích Hàng Loạt (Batch Directory Processing)**
  - *Mô tả*: Hỗ trợ quét và xử lý đồng thời một thư mục chứa nhiều tệp tin ảnh chụp màn hình phòng trọ.
  - *Giải pháp kỹ thuật*: Đọc tất cả tệp ảnh (`.jpg`, `.jpeg`, `.png`, `.webp`) trong thư mục do người dùng cung cấp. Sử dụng Go goroutines và worker pool để gửi yêu cầu song song đến Gemini API (giới hạn số luồng để tránh vượt mức Rate Limit).
  - *Trạng thái*: Lập kế hoạch (Đợi xác nhận giới hạn rate limit của API Gemini).
  - *Phụ thuộc*: Không phụ thuộc.

- [ ] [ASSUMPTION] **Task 4: Tự Do Tùy Biến Mẫu Tin Nhắn (Custom TUI Message Templates)**
  - *Mô tả*: Cho phép người dùng tự định nghĩa và chỉnh sửa mẫu tin nhắn hỏi chủ nhà (xưng hô, các câu hỏi thiếu thông tin) trực tiếp trên giao diện TUI.
  - *Giải pháp kỹ thuật*: Lưu trữ các mẫu tin nhắn trong tệp cấu hình toàn cục `~/.config/soi-tro/config.json`. Sử dụng cú pháp mẫu của Go (`text/template`) để thay thế động các trường như `{{.Address}}`, `{{.Price}}`, `{{.MissingFields}}`.
  - *Trạng thái*: Lập kế hoạch.
  - *Phụ thuộc*: Không phụ thuộc.

- [ ] [ASSUMPTION] **Task 5: Bộ Chuẩn Hóa Giá Cục Bộ (Vietnamese Rental Price Normalizer Library)**
  - *Mô tả*: Xây dựng module xử lý chuỗi bằng Go để nhận diện và quy đổi các cách viết giá phòng trọ phổ biến ở Việt Nam.
  - *Giải pháp kỹ thuật*: Tự động chuyển đổi các chuỗi viết tắt/tiếng lóng như: `4tr5`, `4.5tr`, `4500k`, `4.5 triệu`, `4m5` thành giá trị số nguyên (`4500000`).
  - *Trạng thái*: Lập kế hoạch.
  - *Phụ thuộc*: Không phụ thuộc.

- [x] [CONFIRMED] **Task 6: Hệ Thống Logging & Observability (Logging & Observability System)**
  - *Mô tả*: Tích hợp hệ thống logging có cấu trúc và observability để giúp phát hiện lỗi nhanh chóng và cải thiện quy trình debug cho developer.
  - *Giải pháp kỹ thuật*: Sử dụng Go's `log/slog` cho structured logging, custom error types với error codes, context propagation cho request tracing, file logging với rotation, và configurable log levels. Xem chi tiết trong `logging-observability-plan.md`.
  - *Trạng thái*: Hoàn thành (Phase 1 & 2 - Core infrastructure and main.go integration).
  - *Phụ thuộc*: Không phụ thuộc.

---

## 🛠️ Kế Hoạch Triển Khai Tiếp Theo (Next Steps)
1. **Bước 1**: Nhận sự chấp thuận cho cấu trúc tài liệu mới.
2. **Bước 2**: Thực hiện tích hợp tính năng **Task 1 (Lưu trữ SQLite)** làm bước đệm đầu tiên (đang bị chặn do Task 3 phụ thuộc vào Task 1).
3. **Bước 3**: Triển khai tính năng **Task 3 (So sánh song song)** sau khi Task 1 hoàn tất.
4. **Bước 4**: Triển khai tính năng **Task 6 (Logging & Observability)** để cải thiện khả năng debug và monitoring cho toàn bộ ứng dụng.
