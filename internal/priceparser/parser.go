// Package priceparser normalizes the free-text Vietnamese rental-price
// strings extracted by Gemini into an exact integer VND amount.
//
// Supported syntax (case-insensitive, surrounding and internal whitespace
// ignored, optional trailing "/tháng", "/tuần", "/ngày", "/năm" period and
// "đ"/"vnd"/"vnđ" currency markers are stripped before parsing):
//
//   - "<int>tr<digit>?" or "<int>m<digit>?": triệu (million) shorthand, e.g.
//     "4tr" = 4,000,000; "4tr5"/"4m5" = 4,500,000. The optional trailing
//     digit is tenths of a million.
//   - "<int>[.,]<digits>tr" or "...m": decimal triệu, e.g. "4.5tr",
//     "4,5 triệu" = 4,500,000. "triệu"/"trieu" are accepted spellings of "tr".
//   - "<int>k": nghìn (thousand) shorthand, e.g. "4500k" = 4,500,000.
//     "nghìn"/"nghin" are accepted spellings of "k".
//   - "<int>[.,]<digits>k": decimal nghìn, e.g. "4.5k" = 4,500.
//   - "<digits>.<3 digits>(.<3 digits>)*": dot-grouped literal VND, e.g.
//     "4.500.000" = 4,500,000.
//   - "<digits>": a bare integer of at least 1,000, taken literally as VND.
//     Zero is also accepted as an explicit free-rent value.
//
// A decimal amount that does not resolve to a whole VND value, and any input
// that does not match one of the forms above (including bare decimals
// without a unit, or integers below 1,000), is rejected as ambiguous or
// malformed rather than guessed. All arithmetic is done with integers; no
// rounding is applied.
package priceparser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	million      = 1_000_000
	thousand     = 1_000
	minLiteral   = 1_000
	tenthMillion = million / 10
)

var (
	periodSuffixRe   = regexp.MustCompile(`/\s*(tháng|thang|tuần|tuan|ngày|ngay|năm|nam)\s*$`)
	thousandsGroupRe = regexp.MustCompile(`^\d{1,3}(\.\d{3})+$`)
	plainIntRe       = regexp.MustCompile(`^\d+$`)
	trTenthRe        = regexp.MustCompile(`^(\d+)(?:tr|m)(\d)?$`)
	trDecimalRe      = regexp.MustCompile(`^(\d+)[.,](\d+)(?:tr|m)$`)
	kIntRe           = regexp.MustCompile(`^(\d+)k$`)
	kDecimalRe       = regexp.MustCompile(`^(\d+)[.,](\d+)k$`)
)

// ParseVND normalizes a Vietnamese rental-price string into an integer VND
// amount. It returns an error instead of guessing when the input is
// ambiguous or malformed.
func ParseVND(raw string) (int64, error) {
	work, err := normalize(raw)
	if err != nil {
		return 0, err
	}

	switch {
	case trTenthRe.MatchString(work):
		m := trTenthRe.FindStringSubmatch(work)
		return wholeWithTenth(raw, m[1], m[2], million)
	case trDecimalRe.MatchString(work):
		m := trDecimalRe.FindStringSubmatch(work)
		return decimalTimes(raw, m[1], m[2], million)
	case kDecimalRe.MatchString(work):
		m := kDecimalRe.FindStringSubmatch(work)
		return decimalTimes(raw, m[1], m[2], thousand)
	case kIntRe.MatchString(work):
		m := kIntRe.FindStringSubmatch(work)
		return wholeTimes(raw, m[1], thousand)
	case thousandsGroupRe.MatchString(work):
		return parseThousandsGroup(raw, work)
	case plainIntRe.MatchString(work):
		return parseLiteral(raw, work)
	default:
		return 0, fmt.Errorf("price %q does not match a supported or unambiguous format", raw)
	}
}

// normalize lowercases raw, strips a trailing rent period and currency
// marker, and expands accepted unit spellings, ready for pattern matching.
func normalize(raw string) (string, error) {
	work := strings.ToLower(strings.TrimSpace(raw))
	if work == "" {
		return "", fmt.Errorf("price %q is empty", raw)
	}

	work = periodSuffixRe.ReplaceAllString(work, "")
	work = strings.Join(strings.Fields(work), "")

	for _, suffix := range []string{"vnđ", "vnd", "đ"} {
		if trimmed, ok := strings.CutSuffix(work, suffix); ok {
			work = trimmed
			break
		}
	}

	work = strings.NewReplacer(
		"triệu", "tr",
		"trieu", "tr",
		"nghìn", "k",
		"nghin", "k",
	).Replace(work)

	if work == "" {
		return "", fmt.Errorf("price %q has no numeric content", raw)
	}

	return work, nil
}

func wholeTimes(raw, wholePart string, multiplier int64) (int64, error) {
	base, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("price %q has an invalid whole number: %w", raw, err)
	}

	return base * multiplier, nil
}

func wholeWithTenth(raw, wholePart, tenthPart string, multiplier int64) (int64, error) {
	vnd, err := wholeTimes(raw, wholePart, multiplier)
	if err != nil {
		return 0, err
	}

	if tenthPart != "" {
		tenth, err := strconv.ParseInt(tenthPart, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("price %q has an invalid tenths digit: %w", raw, err)
		}

		vnd += tenth * tenthMillion
	}

	return vnd, nil
}

// decimalTimes computes (wholePart.fracPart) * multiplier using only integer
// arithmetic, rejecting the input if the result is not a whole VND amount.
func decimalTimes(raw, wholePart, fracPart string, multiplier int64) (int64, error) {
	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("price %q has an invalid whole number: %w", raw, err)
	}

	frac, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("price %q has an invalid decimal number: %w", raw, err)
	}

	denom := int64(1)
	for range fracPart {
		denom *= 10
	}

	numerator := whole*denom + frac

	scaled := numerator * multiplier
	if scaled%denom != 0 {
		return 0, fmt.Errorf("price %q does not resolve to a whole VND amount", raw)
	}

	return scaled / denom, nil
}

// parseThousandsGroup parses a dot-grouped literal such as "4.500.000".
func parseThousandsGroup(raw, work string) (int64, error) {
	vnd, err := strconv.ParseInt(strings.ReplaceAll(work, ".", ""), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("price %q is not a valid number: %w", raw, err)
	}

	return vnd, nil
}

// parseLiteral parses a bare integer, rejecting values too small to be an
// unambiguous VND amount.
func parseLiteral(raw, work string) (int64, error) {
	vnd, err := strconv.ParseInt(work, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("price %q is not a valid number: %w", raw, err)
	}

	if vnd != 0 && vnd < minLiteral {
		return 0, fmt.Errorf("price %q is too small to be an unambiguous VND amount", raw)
	}

	return vnd, nil
}
