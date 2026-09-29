package preferences

import (
	"fmt"
	"math"
	"regexp"
	"slices"
	"soi-tro/internal/priceparser"
	"strconv"
	"strings"
)

// RentalFacts is the local, text-only input to profile evaluation.
type RentalFacts struct{ Price, Deposit, Floor, ParkingFee, PetsAllowed, Elevator string }

// Finding is one configured requirement's local assessment.
type Finding struct{ Field, Status, Reason string }

// Result carries the classification and every configured requirement's reason.
type Result struct {
	Classification string
	Findings       []Finding
	UnknownCount   int
	MoveInCashVND  *int64
}

const (
	findingPass    = "pass"
	findingFail    = "fail"
	findingUnknown = "unknown"
)

var (
	commaMoney    = regexp.MustCompile(`^\d{1,3}(,\d{3})+$`)
	depositMonths = regexp.MustCompile(`^(?:(?:cọc|đặt cọc)\s+)?(\d+)\s+tháng$`)
	floorNumber   = regexp.MustCompile(`^(?:(?:tầng|lầu)\s*)?(\d+)$`)
	parkingFee    = regexp.MustCompile(`^\d[\d.,]*\s*(?:vnd|vnđ|đ)\s*/\s*xe(?:\s*/\s*tháng)?$`)
)

// Evaluate applies each configured rule exactly once without external calls.
func Evaluate(p SearchProfile, f RentalFacts) Result {
	r := Result{Classification: "Shortlist"}
	if cash, ok := moveInCash(f.Price, f.Deposit); ok {
		r.MoveInCashVND = &cash
	}
	add := func(field, status, reason string) {
		r.Findings = append(r.Findings, Finding{field, status, reason})
		if status == findingUnknown {
			r.UnknownCount++
		}
		if status == findingFail {
			r.Classification = "Reject"
		}
	}
	if p.MaxRentVND != nil {
		if rent, ok := parseMoney(f.Price); !ok {
			add("price", findingUnknown, "Giá thuê hằng tháng bị thiếu hoặc không rõ.")
		} else if rent > *p.MaxRentVND {
			add("price", findingFail, fmt.Sprintf("Giá thuê %d đồng vượt mức tối đa %d đồng.", rent, *p.MaxRentVND))
		} else {
			add("price", findingPass, fmt.Sprintf("Giá thuê %d đồng nằm trong giới hạn.", rent))
		}
	}
	if p.MaxMoveInCashVND != nil {
		if cash, ok := moveInCash(f.Price, f.Deposit); !ok {
			add("move_in_cash", findingUnknown, "Không thể tính chính xác tiền thuê cộng tiền cọc; chưa tính các khoản trả trước khác.")
		} else if cash > *p.MaxMoveInCashVND {
			add("move_in_cash", findingFail, fmt.Sprintf("Tiền thuê cộng tiền cọc %d đồng vượt mức tối đa %d đồng; chưa tính các khoản trả trước khác.", cash, *p.MaxMoveInCashVND))
		} else {
			add("move_in_cash", findingPass, fmt.Sprintf("Tiền thuê cộng tiền cọc %d đồng nằm trong giới hạn; chưa tính các khoản trả trước khác.", cash))
		}
	}
	if p.MaxDepositMonths != nil {
		if status, reason := checkDepositMonths(f.Price, f.Deposit, *p.MaxDepositMonths); status != "" {
			add("deposit_months", status, reason)
		}
	}
	evaluateOtherRules(p, f, add)
	if p.MaxUnknown != nil && r.UnknownCount > *p.MaxUnknown {
		add("unknown_count", findingFail, fmt.Sprintf("Có %d tiêu chí chưa rõ, vượt giới hạn %d.", r.UnknownCount, *p.MaxUnknown))
	}
	if r.Classification != "Reject" && r.UnknownCount > 0 {
		r.Classification = "Needs checking"
	}
	return r
}

func evaluateOtherRules(p SearchProfile, f RentalFacts, add func(field, status, reason string)) {
	if p.RequirePets {
		status := yesNo(f.PetsAllowed,
			[]string{"có", "được nuôi pet", "cho nuôi pet", "được nuôi thú cưng", "cho nuôi thú cưng"},
			[]string{"không", "không cho nuôi pet", "không cho nuôi thú cưng", "cấm nuôi pet", "cấm nuôi thú cưng"})
		reason := "Quy định nuôi thú cưng chưa rõ hoặc có điều kiện."
		switch status {
		case findingPass:
			reason = "Tin đăng cho phép nuôi thú cưng."
		case findingFail:
			reason = "Tin đăng không cho nuôi thú cưng."
		}
		add("pets_allowed", status, reason)
	}
	if p.RequireParking {
		status := parking(f.ParkingFee)
		reason := "Thông tin chỗ giữ xe bị thiếu hoặc chưa rõ."
		switch status {
		case findingPass:
			reason = "Tin đăng có chỗ giữ xe."
		case findingFail:
			reason = "Tin đăng không có chỗ giữ xe."
		}
		add("parking_fee", status, reason)
	}
	if p.RequireElevator {
		status := yesNo(f.Elevator,
			[]string{"có", "có thang máy", "có sử dụng thang máy"},
			[]string{"không", "không có thang máy", "ko thang máy"})
		reason := "Thang máy chưa được đề cập hoặc thông tin chưa rõ."
		switch status {
		case findingPass:
			reason = "Tin đăng có thang máy."
		case findingFail:
			reason = "Tin đăng không có thang máy."
		}
		add("elevator", status, reason)
	}
	if p.MaxFloor != nil {
		if floor, ok := parseFloor(f.Floor); !ok {
			add("floor", findingUnknown, "Tầng của phòng bị thiếu hoặc không rõ.")
		} else if floor > *p.MaxFloor {
			add("floor", findingFail, fmt.Sprintf("Tầng %d cao hơn tầng tối đa %d.", floor, *p.MaxFloor))
		} else {
			add("floor", findingPass, fmt.Sprintf("Tầng %d nằm trong giới hạn.", floor))
		}
	}
}

func checkDepositMonths(price, deposit string, limit int) (string, string) {
	if m := depositMonths.FindStringSubmatch(strings.ToLower(strings.TrimSpace(deposit))); m != nil {
		months, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil {
			return findingUnknown, "Số tháng cọc trong tin đăng không rõ."
		}
		if months > int64(limit) {
			return findingFail, fmt.Sprintf("Tiền cọc %d tháng vượt giới hạn %d tháng.", months, limit)
		}
		return findingPass, fmt.Sprintf("Tiền cọc %d tháng nằm trong giới hạn %d tháng.", months, limit)
	}
	amount, amountOK := parseMoney(deposit)
	if !amountOK {
		return findingUnknown, "Tiền cọc bị thiếu hoặc không rõ."
	}
	rent, rentOK := parseMoney(price)
	if !rentOK || rent == 0 {
		return findingUnknown, "Không thể so tiền cọc với giá thuê hằng tháng vì giá thuê bị thiếu hoặc không rõ."
	}
	q, remainder := amount/rent, amount%rent
	if q > int64(limit) || (q == int64(limit) && remainder > 0) {
		return findingFail, fmt.Sprintf("Tiền cọc %d đồng vượt giới hạn %d tháng tiền thuê.", amount, limit)
	}
	return findingPass, fmt.Sprintf("Tiền cọc %d đồng nằm trong giới hạn %d tháng tiền thuê.", amount, limit)
}

func parseMoney(raw string) (int64, bool) {
	work := strings.TrimSpace(raw)
	// The price normalizer accepts dot-grouped VND. Gemini also emits comma groups.
	fields := strings.Fields(strings.ToLower(work))
	if len(fields) > 0 && commaMoney.MatchString(fields[0]) {
		work = strings.Replace(work, fields[0], strings.ReplaceAll(fields[0], ",", "."), 1)
	}
	v, err := priceparser.ParseVND(work)
	return v, err == nil && v >= 0
}

func moveInCash(price, deposit string) (int64, bool) {
	rent, ok := parseMoney(price)
	if !ok {
		return 0, false
	}
	var amount int64
	if v, exact := parseMoney(deposit); exact {
		amount = v
	} else if m := depositMonths.FindStringSubmatch(strings.ToLower(strings.TrimSpace(deposit))); m != nil {
		months, err := strconv.ParseInt(m[1], 10, 64)
		if err != nil || (months > 0 && rent > math.MaxInt64/months) {
			return 0, false
		}
		amount = rent * months
	} else {
		return 0, false
	}
	if rent > math.MaxInt64-amount {
		return 0, false
	}
	return rent + amount, true
}

func yesNo(raw string, yes, no []string) string {
	work := strings.ToLower(strings.TrimSpace(raw))
	if slices.Contains(no, work) {
		return findingFail
	}
	if slices.Contains(yes, work) {
		return findingPass
	}
	return findingUnknown
}

func parking(raw string) string {
	work := strings.ToLower(strings.TrimSpace(raw))
	if slices.Contains([]string{"không có chỗ để xe", "không có chỗ đậu xe", "không có chỗ gửi xe", "không cho để xe"}, work) {
		return findingFail
	}
	if slices.Contains([]string{"có chỗ để xe", "có chỗ đậu xe", "có chỗ gửi xe", "giữ xe miễn phí"}, work) {
		return findingPass
	}
	if parkingFee.MatchString(work) {
		return findingPass
	}
	return findingUnknown
}

func parseFloor(raw string) (int, bool) {
	work := strings.ToLower(strings.TrimSpace(raw))
	if work == "tầng trệt" || work == "trệt" {
		return 0, true
	}
	m := floorNumber.FindStringSubmatch(work)
	if m == nil {
		return 0, false
	}
	v, err := strconv.Atoi(m[1])
	return v, err == nil
}
