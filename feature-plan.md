# 🚀 Kế hoạch Phát triển Tính năng (Feature Plan) - Soi Trọ CLI

Tài liệu này đề xuất lộ trình và kế hoạch chi tiết để tích hợp thêm các tính năng hữu ích cho ứng dụng dòng lệnh **Soi Trọ** (CLI), giúp hỗ trợ người tìm kiếm và phân tích thông tin phòng trọ hiệu quả hơn.

---

## 📌 Các Tính Năng Đề Xuất

### 1. Tích Hợp Cơ Sở Dữ Liệu SQLite Cục Cấu Trúc (History Database)
*   **Mô tả**: Tự động lưu trữ kết quả phân tích phòng trọ của mỗi lượt chạy vào cơ sở dữ liệu SQLite cục bộ thay vì chỉ xuất tệp văn bản.
*   **Giải pháp kỹ thuật**:
    *   Sử dụng thư viện SQLite viết hoàn toàn bằng Go (pure-Go port như `modernc.org/sqlite`). Thư viện này sẽ được biên dịch tĩnh trực tiếp vào tệp thực thi `.exe`, đảm bảo tính độc lập và không cần cài đặt bên ngoài.
    *   Cơ sở dữ liệu sẽ tự động được khởi tạo tại `~/.config/soi-tro/rentals.db`.
*   **Trải nghiệm TUI**:
    *   Thêm menu chính: `Xem lịch sử phân tích (History)`.
    *   Hỗ trợ xem lại chi tiết kết quả dưới dạng bảng, lọc tìm kiếm nhanh theo giá, số điện thoại hoặc trạng thái thiếu thông tin, và cho phép xóa bản ghi cũ.

### 2. Quét & Phân Tích Hàng Loạt (Batch Directory Processing)
*   **Mô tả**: Hỗ trợ quét và xử lý đồng thời một thư mục chứa nhiều tệp tin ảnh chụp màn hình phòng trọ.
*   **Giải pháp kỹ thuật**:
    *   Đọc tất cả tệp ảnh (`.jpg`, `.jpeg`, `.png`, `.webp`) trong thư mục do người dùng cung cấp.
    *   Sử dụng Go goroutines và worker pool để gửi yêu cầu song song đến Gemini API (giới hạn số luồng để tránh vượt mức Rate Limit).
*   **Trải nghiệm TUI**:
    *   Hiển thị thanh tiến trình xử lý (Progress Bar) bằng `github.com/charmbracelet/bubbles/progress`.
    *   Xuất ra một tệp báo cáo tổng hợp Markdown/CSV so sánh tất cả các phòng trọ đã quét.

### 3. So Sánh Phòng Trọ Song Song (Side-by-Side Comparison UI)
*   **Mô tả**: Cho phép người dùng lựa chọn từ 2 đến 3 phòng trọ từ lịch sử lưu trữ để đối chiếu trực quan theo cột.
*   **Giải pháp kỹ thuật**:
    *   Thiết kế giao diện so sánh nhiều cột trực quan trên Terminal UI bằng cách mở rộng `tablewriter` hoặc thiết lập bố cục layout tùy biến của Bubble Tea.
*   **Trải nghiệm TUI**:
    *   Đối chiếu trực quan các thông số cốt lõi: Giá thuê, tiền cọc, chi phí điện nước, thang máy, tiện ích đi kèm để người dùng đưa ra quyết định tối ưu.

### 4. Tự Do Tùy Biến Mẫu Tin Nhắn (Custom TUI Message Templates)
*   **Mô tả**: Cho phép người dùng tự định nghĩa và chỉnh sửa mẫu tin nhắn hỏi chủ nhà (xưng hô, các câu hỏi thiếu thông tin) trực tiếp trên giao diện TUI.
*   **Giải pháp kỹ thuật**:
    *   Lưu trữ các mẫu tin nhắn trong tệp cấu hình toàn cục `~/.config/soi-tro/config.json`.
    *   Sử dụng cú pháp mẫu của Go (`text/template`) để thay thế động các trường như `{{.Address}}`, `{{.Price}}`, `{{.MissingFields}}`.
*   **Trải nghiệm TUI**:
    *   Thêm mục chỉnh sửa Template trong phần cài đặt TUI bằng biểu mẫu Huh.

### 5. Bộ Chuẩn Hóa Giá Cục Bộ (Vietnamese Rental Price Normalizer Library)
*   **Mô tả**: Xây dựng module xử lý chuỗi bằng Go để nhận diện và quy đổi các cách viết giá phòng trọ phổ biến ở Việt Nam.
*   **Giải pháp kỹ thuật**:
    *   Tự động chuyển đổi các chuỗi viết tắt/tiếng lóng như: `4tr5`, `4.5tr`, `4500k`, `4.5 triệu`, `4m5` thành giá trị số nguyên (`4500000`).
    *   Giúp tăng độ chính xác 100% khi thực hiện lọc hoặc sắp xếp giá trên SQLite và giảm số lượng token hướng dẫn gửi lên Gemini API.

---

## 🛠️ Kế Hoạch Triển Khai Tiếp Theo (Next Steps)
1.  **Bước 1**: Tạo tệp tài liệu kế hoạch tính năng `feature-plan.md` trong thư mục gốc của dự án.
2.  **Bước 2**: Thực hiện tích hợp tính năng **1 (Lưu trữ SQLite)** làm bước đệm đầu tiên.
3.  **Bước 3**: Triển khai các tính năng nâng cao tiếp theo dựa trên phản hồi của người dùng.
