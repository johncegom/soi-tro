package ui

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// InputType represents the selected user input type.
type InputType int

const (
	// InputTypeImage represents a local image file path.
	InputTypeImage InputType = iota
	// InputTypeText represents raw pasted text listing.
	InputTypeText
)

// InputResult holds the collected inputs.
type InputResult struct {
	Type      InputType
	Text      string
	ImagePath string
}

// ErrGoBack is returned when the user presses the left arrow to go back to the previous menu.
var ErrGoBack = errors.New("quay lại menu trước")

type arrowModel struct {
	form        *huh.Form
	backPressed bool
}

func (m *arrowModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m *arrowModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch k := msg.(type) {
	case tea.KeyMsg:
		switch k.Type {
		case tea.KeyLeft:
			m.backPressed = true
			return m, tea.Quit
		case tea.KeyRight:
			// Translate Right arrow to Enter key
			msg = tea.KeyMsg{Type: tea.KeyEnter}
		}
	}

	newForm, cmd := m.form.Update(msg)
	m.form = newForm.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		return m, tea.Quit
	}
	if m.form.State == huh.StateAborted {
		return m, tea.Quit
	}

	return m, cmd
}

func (m *arrowModel) View() string {
	return m.form.View()
}

// RunFormWithArrows runs a form allowing Left to go back and Right/Enter to submit.
func RunFormWithArrows(form *huh.Form) (bool, error) {
	m := &arrowModel{
		form: form,
	}

	p := tea.NewProgram(m)
	_, err := p.Run()
	if err != nil {
		return false, err
	}

	if m.form.State == huh.StateAborted {
		return false, huh.ErrUserAborted
	}

	return m.backPressed, nil
}

// GetUserInput displays an interactive form in the terminal to capture input details.
func GetUserInput() (*InputResult, error) {
	var inputChoice string
	var textInput string
	var imagePathInput string
	var textFilePathInput string

	// Step 1: Selection Form
	selectForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Chọn phương thức nhập thông tin đăng").
				Options(
					huh.NewOption("Đường dẫn đến file hình ảnh local", "image"),
					huh.NewOption("Dán văn bản thông tin đăng trực tiếp (An toàn hơn cho dán nhiều dòng)", "text"),
					huh.NewOption("Đường dẫn đến file văn bản (.txt) chứa thông tin đăng", "file"),
				).
				Value(&inputChoice),
		),
	)

	backPressed, err := RunFormWithArrows(selectForm)
	if err != nil {
		return nil, fmt.Errorf("lỗi khi chọn phương thức nhập: %w", err)
	}
	if backPressed {
		return nil, ErrGoBack
	}

	res := &InputResult{}

	// Step 2: Capture input based on selection
	if inputChoice == "image" {
		res.Type = InputTypeImage
		imageForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Nhập hoặc kéo thả file hình ảnh (JPEG, PNG, WebP)").
					Placeholder("VD: C:\\Rentals\\listing.jpg").
					Value(&imagePathInput).
					Validate(func(str string) error {
						str = strings.TrimSpace(str)
						// Strip quotes typical of drag-and-drop file paths
						if (strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"")) ||
							(strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'")) {
							str = str[1 : len(str)-1]
						}
						str = strings.TrimSpace(str)
						if str == "" {
							return errors.New("đường dẫn không được để trống")
						}
						info, err := os.Stat(str)
						if err != nil {
							return fmt.Errorf("không tìm thấy file: %w", err)
						}
						if info.IsDir() {
							return errors.New("đây là thư mục, hãy nhập đường dẫn file hình ảnh")
						}
						lower := strings.ToLower(str)
						if !strings.HasSuffix(lower, ".jpg") && !strings.HasSuffix(lower, ".jpeg") &&
							!strings.HasSuffix(lower, ".png") && !strings.HasSuffix(lower, ".webp") {
							return errors.New("định dạng file không được hỗ trợ (chỉ nhận .jpg, .jpeg, .png, .webp)")
						}
						return nil
					}),
			),
		)
		err = imageForm.Run()
		if err != nil {
			return nil, fmt.Errorf("lỗi khi nhập đường dẫn hình ảnh: %w", err)
		}
		path := strings.TrimSpace(imagePathInput)
		if (strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"")) ||
			(strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'")) {
			path = path[1 : len(path)-1]
		}
		res.ImagePath = strings.TrimSpace(path)
	} else if inputChoice == "file" {
		res.Type = InputTypeText
		fileForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Nhập hoặc kéo thả file văn bản (.txt)").
					Placeholder("VD: C:\\Rentals\\listing.txt").
					Value(&textFilePathInput).
					Validate(func(str string) error {
						str = strings.TrimSpace(str)
						// Strip quotes typical of drag-and-drop file paths
						if (strings.HasPrefix(str, "\"") && strings.HasSuffix(str, "\"")) ||
							(strings.HasPrefix(str, "'") && strings.HasSuffix(str, "'")) {
							str = str[1 : len(str)-1]
						}
						str = strings.TrimSpace(str)
						if str == "" {
							return errors.New("đường dẫn không được để trống")
						}
						info, err := os.Stat(str)
						if err != nil {
							return fmt.Errorf("không tìm thấy file: %w", err)
						}
						if info.IsDir() {
							return errors.New("đây là thư mục, hãy nhập đường dẫn file văn bản")
						}
						return nil
					}),
			),
		)
		err = fileForm.Run()
		if err != nil {
			return nil, fmt.Errorf("lỗi khi nhập đường dẫn file: %w", err)
		}
		path := strings.TrimSpace(textFilePathInput)
		if (strings.HasPrefix(path, "\"") && strings.HasSuffix(path, "\"")) ||
			(strings.HasPrefix(path, "'") && strings.HasSuffix(path, "'")) {
			path = path[1 : len(path)-1]
		}
		path = strings.TrimSpace(path)
		// Read text file contents
		bytes, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("lỗi khi đọc file văn bản: %w", err)
		}
		res.Text = string(bytes)
	} else {
		res.Type = InputTypeText
		fmt.Println("\n-------------------------------------------------------------------------")
		fmt.Println("👉 HƯỚNG DẪN DÁN VĂN BẢN:")
		fmt.Println("   1. Dán nội dung tin đăng của bạn xuống dưới dòng này.")
		fmt.Println("   2. Để hoàn tất nhập liệu, xuống dòng mới, nhập từ khóa 'END' rồi nhấn Enter.")
		fmt.Println("-------------------------------------------------------------------------")
		fmt.Print("Nội dung tin đăng:\n")

		scanner := bufio.NewScanner(os.Stdin)
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "END" {
				break
			}
			lines = append(lines, line)
		}
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("lỗi khi đọc dữ liệu nhập: %w", err)
		}

		textInput = strings.Join(lines, "\n")
		if strings.TrimSpace(textInput) == "" {
			return nil, errors.New("nội dung tin đăng không được để trống")
		}
		res.Text = textInput
	}

	return res, nil
}

// PromptErrorRetry presents a choice to the user when Gemini or file loading fails.
func PromptErrorRetry(inputType InputType) string {
	var choice string
	var options []huh.Option[string]

	if inputType == InputTypeImage {
		options = []huh.Option[string]{
			huh.NewOption("1. Thử lại phân tích ảnh này (Retry)", "retry"),
			huh.NewOption("2. Chọn ảnh khác (Change)", "change"),
			huh.NewOption("3. Quay lại menu chính (Back)", "back"),
		}
	} else {
		options = []huh.Option[string]{
			huh.NewOption("1. Thử lại phân tích văn bản này (Retry)", "retry"),
			huh.NewOption("2. Nhập văn bản khác (Change)", "change"),
			huh.NewOption("3. Quay lại menu chính (Back)", "back"),
		}
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Đã xảy ra lỗi. Vui lòng chọn hành động tiếp theo:").
				Options(options...).
				Value(&choice),
		),
	)

	backPressed, err := RunFormWithArrows(form)
	if err != nil || backPressed {
		return "back"
	}
	return choice
}

// PromptAfterSuccess presents a choice to the user after a successful analysis.
// exportConfigured controls whether the export option is shown.
// exportDone changes the export label to indicate the result was already exported.
func PromptAfterSuccess(exportConfigured, exportDone bool) string {
	var choice string

	options := []huh.Option[string]{}
	if exportConfigured {
		exportLabel := "2. Xuất kết quả ra file (Export)"
		if exportDone {
			exportLabel = "2. Xuất lại kết quả ra file (Re-export)"
		}
		options = append(options, huh.NewOption(exportLabel, "export"))
	}
	options = append(options,
		huh.NewOption("3. Tiếp tục phân tích tin đăng khác", "new"),
		huh.NewOption("4. Quay lại menu chính", "back"),
	)

	// Re-number the first option correctly when export is not shown.
	if !exportConfigured {
		options[0] = huh.NewOption("1. Tiếp tục phân tích tin đăng khác", "new")
		options[1] = huh.NewOption("2. Quay lại menu chính", "back")
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Phân tích hoàn tất. Bạn muốn làm gì tiếp theo?").
				Options(options...).
				Value(&choice),
		),
	)

	backPressed, err := RunFormWithArrows(form)
	if err != nil || backPressed {
		return "back"
	}
	return choice
}
