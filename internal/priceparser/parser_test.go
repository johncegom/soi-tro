package priceparser

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"
)

func TestParseVND(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"tr with tenth", "4tr5", 4_500_000, false},
		{"decimal tr", "4.5tr", 4_500_000, false},
		{"k literal", "4500k", 4_500_000, false},
		{"decimal trieu word", "4.5 triệu", 4_500_000, false},
		{"m with tenth", "4m5", 4_500_000, false},
		{"plain tr", "4tr", 4_000_000, false},
		{"trieu with period suffix", "5 triệu/tháng", 5_000_000, false},
		{"uppercase and spacing", "  4.5 TR  ", 4_500_000, false},
		{"comma decimal", "4,5tr", 4_500_000, false},
		{"comma decimal k", "4,5k", 4_500, false},
		{"dot-grouped literal", "4.500.000", 4_500_000, false},
		{"dot-grouped with currency suffix", "4.500.000đ", 4_500_000, false},
		{"dot-grouped with vnd suffix", "4.500.000 vnd", 4_500_000, false},
		{"bare literal at minimum", "1000", 1_000, false},
		{"zero is explicit free rent", "0", 0, false},
		{"period suffix tuan", "500k/tuần", 500_000, false},

		{"empty", "", 0, true},
		{"whitespace only", "   ", 0, true},
		{"bare decimal without unit is ambiguous", "4.5", 0, true},
		{"bare integer below minimum is ambiguous", "999", 0, true},
		{"negative is malformed", "-4tr", 0, true},
		{"non numeric", "thoả thuận", 0, true},
		{"non-whole decimal million", "4.1234567tr", 0, true},
		{"multiple units", "4tr5k", 0, true},
		{"overflow", "99999999999999999999tr", 0, true},
		{"ambiguous partial thousands group", "4.50", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseVND(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseVND(%q) = %d, nil; want error", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseVND(%q) returned unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseVND(%q) = %d; want %d", tt.input, got, tt.want)
			}
		})
	}
}

// TestParseVND_ManualDataset runs the representative listing strings from
// docs/price-normalization.md's manual test plan against the parser, so the
// documented examples stay correct as the grammar evolves.
func TestParseVND_ManualDataset(t *testing.T) {
	f, err := os.Open("testdata/manual_price_dataset.csv")
	if err != nil {
		t.Fatalf("failed to open manual dataset: %v", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("failed to read manual dataset: %v", err)
	}

	for i, row := range rows[1:] {
		line := i + 2 // header is line 1
		input, expectedVND, expectedResult := row[0], row[1], row[2]

		got, err := ParseVND(input)
		switch expectedResult {
		case "accept":
			if err != nil {
				t.Errorf("line %d: ParseVND(%q) returned unexpected error: %v", line, input, err)
				continue
			}
			want, parseErr := strconv.ParseInt(expectedVND, 10, 64)
			if parseErr != nil {
				t.Fatalf("line %d: dataset has invalid expected_vnd %q: %v", line, expectedVND, parseErr)
			}
			if got != want {
				t.Errorf("line %d: ParseVND(%q) = %d; want %d", line, input, got, want)
			}
		case "reject":
			if err == nil {
				t.Errorf("line %d: ParseVND(%q) = %d, nil; want an error", line, input, got)
			}
		default:
			t.Fatalf("line %d: dataset has unknown expected_result %q", line, expectedResult)
		}
	}
}

func FuzzParseVND(f *testing.F) {
	seeds := []string{
		"4tr5", "4.5tr", "4500k", "4.5 triệu", "4m5", "", "-4tr", "4.5",
		"4.500.000đ", "thoả thuận", "99999999999999999999tr", "4tr5k",
	}
	for _, s := range seeds {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The parser must never panic; the returned error (if any) is not
		// otherwise checked since arbitrary fuzz input has no known answer.
		_, _ = ParseVND(input)
	})
}
