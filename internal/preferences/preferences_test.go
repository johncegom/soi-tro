package preferences

import (
	"encoding/json"
	"os"
	"path/filepath"
	"soi-tro/internal/gemini"
	"strings"
	"testing"
)

func ptr64(v int64) *int64 { return &v }
func ptrInt(v int) *int    { return &v }

func TestEvaluateRules(t *testing.T) {
	tests := []struct {
		name    string
		profile SearchProfile
		facts   RentalFacts
		want    string
		field   string
		status  string
	}{
		{"rent equal", SearchProfile{MaxRentVND: ptr64(4_500_000)}, RentalFacts{Price: "4tr5"}, "Shortlist", "price", "pass"},
		{"rent high", SearchProfile{MaxRentVND: ptr64(4_499_999)}, RentalFacts{Price: "4tr5"}, "Reject", "price", "fail"},
		{"rent ambiguous", SearchProfile{MaxRentVND: ptr64(5_000_000)}, RentalFacts{Price: "4-5 triệu"}, "Needs checking", "price", "unknown"},
		{"cash amount equal", SearchProfile{MaxMoveInCashVND: ptr64(9_000_000)}, RentalFacts{Price: "4.5tr", Deposit: "4,500,000 VND"}, "Shortlist", "move_in_cash", "pass"},
		{"cash months high", SearchProfile{MaxMoveInCashVND: ptr64(8_999_999)}, RentalFacts{Price: "4.5tr", Deposit: "Cọc 1 tháng"}, "Reject", "move_in_cash", "fail"},
		{"cash deposit unknown", SearchProfile{MaxMoveInCashVND: ptr64(10_000_000)}, RentalFacts{Price: "4.5tr", Deposit: "Cọc thỏa thuận"}, "Needs checking", "move_in_cash", "unknown"},
		{"pets allowed", SearchProfile{RequirePets: true}, RentalFacts{PetsAllowed: "Được nuôi pet"}, "Shortlist", "pets_allowed", "pass"},
		{"pets prohibited", SearchProfile{RequirePets: true}, RentalFacts{PetsAllowed: "Không cho nuôi pet"}, "Reject", "pets_allowed", "fail"},
		{"pets conditional", SearchProfile{RequirePets: true}, RentalFacts{PetsAllowed: "Chỉ cho nuôi mèo"}, "Needs checking", "pets_allowed", "unknown"},
		{"parking fee", SearchProfile{RequireParking: true}, RentalFacts{ParkingFee: "100,000 VND/xe/tháng"}, "Shortlist", "parking_fee", "pass"},
		{"parking absent", SearchProfile{RequireParking: true}, RentalFacts{ParkingFee: "Không có chỗ để xe"}, "Reject", "parking_fee", "fail"},
		{"parking generic free", SearchProfile{RequireParking: true}, RentalFacts{ParkingFee: "Miễn phí"}, "Needs checking", "parking_fee", "unknown"},
		{"elevator yes", SearchProfile{RequireElevator: true}, RentalFacts{Elevator: "Có"}, "Shortlist", "elevator", "pass"},
		{"elevator no", SearchProfile{RequireElevator: true}, RentalFacts{Elevator: "Không"}, "Reject", "elevator", "fail"},
		{"floor equal", SearchProfile{MaxFloor: ptrInt(3)}, RentalFacts{Floor: "Tầng 3"}, "Shortlist", "floor", "pass"},
		{"floor high", SearchProfile{MaxFloor: ptrInt(2)}, RentalFacts{Floor: "Lầu 3"}, "Reject", "floor", "fail"},
		{"ground floor", SearchProfile{MaxFloor: ptrInt(0)}, RentalFacts{Floor: "Tầng trệt"}, "Shortlist", "floor", "pass"},
		{"floor range", SearchProfile{MaxFloor: ptrInt(3)}, RentalFacts{Floor: "Tầng 2-3"}, "Needs checking", "floor", "unknown"},
		{"numeric floor", SearchProfile{MaxFloor: ptrInt(3)}, RentalFacts{Floor: "3"}, "Shortlist", "floor", "pass"},
		{"deposit bare months", SearchProfile{MaxMoveInCashVND: ptr64(9_000_000)}, RentalFacts{Price: "4.5tr", Deposit: "1 tháng"}, "Shortlist", "move_in_cash", "pass"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(tt.profile, tt.facts)
			if got.Classification != tt.want {
				t.Fatalf("classification = %q, want %q", got.Classification, tt.want)
			}
			if len(got.Findings) != 1 || got.Findings[0].Field != tt.field || got.Findings[0].Status != tt.status {
				t.Fatalf("findings = %+v, want one %s %s", got.Findings, tt.field, tt.status)
			}
		})
	}
}

func TestDepositMonthRule(t *testing.T) {
	var profile SearchProfile
	if err := json.Unmarshal([]byte(`{"max_deposit_months":1}`), &profile); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, price, deposit, classification, status string
	}{
		{"money equal", "5tr", "5tr", "Shortlist", "pass"},
		{"money one dong over", "5tr", "5.000.001 VND", "Reject", "fail"},
		{"money one dong under", "5tr", "4.999.999 VND", "Shortlist", "pass"},
		{"fractional months", "5tr", "7.5tr", "Reject", "fail"},
		{"months high", "", "2 tháng", "Reject", "fail"},
		{"months equal without rent", "", "1 tháng", "Shortlist", "pass"},
		{"money without rent", "", "5tr", "Needs checking", "unknown"},
		{"money with zero rent", "0 VND", "5tr", "Needs checking", "unknown"},
		{"ambiguous deposit", "5tr", "thỏa thuận", "Needs checking", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Evaluate(profile, RentalFacts{Price: tt.price, Deposit: tt.deposit})
			if got.Classification != tt.classification || len(got.Findings) != 1 || got.Findings[0].Field != "deposit_months" || got.Findings[0].Status != tt.status {
				t.Fatalf("result = %+v, want %s deposit_months/%s", got, tt.classification, tt.status)
			}
		})
	}
}

func TestDepositMonthsZeroAndLegacyPersistence(t *testing.T) {
	zero := 0
	got := Evaluate(SearchProfile{MaxDepositMonths: &zero}, RentalFacts{Price: "5tr", Deposit: "0 VND"})
	if got.Classification != "Shortlist" || got.Findings[0].Status != "pass" {
		t.Fatalf("zero deposit result = %+v", got)
	}
	got = Evaluate(SearchProfile{MaxDepositMonths: &zero}, RentalFacts{Deposit: "1 tháng"})
	if got.Classification != "Reject" {
		t.Fatalf("month deposit with zero limit = %+v", got)
	}
	store := Store{Path: filepath.Join(t.TempDir(), "profile.json")}
	legacy := []byte(`{"max_move_in_cash_vnd":9000000}` + "\n")
	if err := os.WriteFile(store.Path, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	p, err := store.Load()
	if err != nil || p == nil || p.MaxMoveInCashVND == nil || *p.MaxMoveInCashVND != 9_000_000 {
		t.Fatalf("legacy load = %+v, %v", p, err)
	}
	got = Evaluate(*p, RentalFacts{Price: "5tr", Deposit: "5tr"})
	if got.Classification != "Reject" || got.Findings[0].Field != "move_in_cash" {
		t.Fatalf("legacy rule changed: %+v", got)
	}
	data, err := os.ReadFile(store.Path)
	if err != nil || string(data) != string(legacy) {
		t.Fatalf("load rewrote legacy profile: %q, %v", data, err)
	}
	months := 1
	p.MaxDepositMonths = &months
	if err := p.Validate(); err == nil {
		t.Fatal("profile with both deposit limits should fail")
	}
}

func TestUnknownPrecedence(t *testing.T) {
	facts := RentalFacts{Price: "unknown", PetsAllowed: "Không"}
	profile := SearchProfile{MaxRentVND: ptr64(5_000_000), RequirePets: true, MaxUnknown: ptrInt(0)}
	got := Evaluate(profile, facts)
	if got.Classification != "Reject" || got.UnknownCount != 1 || len(got.Findings) != 3 {
		t.Fatalf("result = %+v, want Reject, one unknown, all reasons", got)
	}
	profile.RequirePets = false
	got = Evaluate(profile, facts)
	if got.Classification != "Reject" || got.UnknownCount != 1 {
		t.Fatalf("unknown limit result = %+v", got)
	}
	profile.MaxUnknown = ptrInt(1)
	got = Evaluate(profile, facts)
	if got.Classification != "Needs checking" {
		t.Fatalf("tolerated unknown classification = %q", got.Classification)
	}
}

func TestFindingReasonsUseVietnameseWithStableCodes(t *testing.T) {
	tests := []struct {
		name, classification, field, status, reason string
		profile                                     SearchProfile
		facts                                       RentalFacts
	}{
		{"rent fail", "Reject", "price", "fail", "Giá thuê", SearchProfile{MaxRentVND: ptr64(4_000_000)}, RentalFacts{Price: "5tr"}},
		{"rent unknown", "Needs checking", "price", "unknown", "thiếu hoặc không rõ", SearchProfile{MaxRentVND: ptr64(4_000_000)}, RentalFacts{}},
		{"cash fail", "Reject", "move_in_cash", "fail", "Tiền thuê cộng tiền cọc", SearchProfile{MaxMoveInCashVND: ptr64(8_000_000)}, RentalFacts{Price: "5tr", Deposit: "5tr"}},
		{"cash unknown", "Needs checking", "move_in_cash", "unknown", "Không thể tính chính xác", SearchProfile{MaxMoveInCashVND: ptr64(8_000_000)}, RentalFacts{Price: "5tr"}},
		{"pets fail", "Reject", "pets_allowed", "fail", "không cho nuôi thú cưng", SearchProfile{RequirePets: true}, RentalFacts{PetsAllowed: "Không"}},
		{"pets unknown", "Needs checking", "pets_allowed", "unknown", "chưa rõ", SearchProfile{RequirePets: true}, RentalFacts{}},
		{"parking fail", "Reject", "parking_fee", "fail", "không có chỗ giữ xe", SearchProfile{RequireParking: true}, RentalFacts{ParkingFee: "Không có chỗ để xe"}},
		{"elevator unknown", "Needs checking", "elevator", "unknown", "Thang máy", SearchProfile{RequireElevator: true}, RentalFacts{}},
		{"floor fail", "Reject", "floor", "fail", "Tầng", SearchProfile{MaxFloor: ptrInt(2)}, RentalFacts{Floor: "3"}},
		{"unknown limit", "Reject", "unknown_count", "fail", "tiêu chí chưa rõ", SearchProfile{RequirePets: true, MaxUnknown: ptrInt(0)}, RentalFacts{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Evaluate(tt.profile, tt.facts)
			if r.Classification != tt.classification {
				t.Fatalf("classification = %q, want %q", r.Classification, tt.classification)
			}
			var finding *Finding
			for i := range r.Findings {
				if r.Findings[i].Field == tt.field {
					finding = &r.Findings[i]
					break
				}
			}
			if finding == nil || finding.Status != tt.status || !strings.Contains(finding.Reason, tt.reason) {
				t.Fatalf("finding = %+v, want %s/%s and Vietnamese reason containing %q", finding, tt.field, tt.status, tt.reason)
			}
		})
	}
}

func TestProfileValidationErrorUsesVietnamese(t *testing.T) {
	if err := (SearchProfile{}).Validate(); err == nil || !strings.Contains(err.Error(), "ít nhất một điều kiện") {
		t.Fatalf("validation error = %v", err)
	}
}

func TestInvalidProfiles(t *testing.T) {
	for _, p := range []SearchProfile{{}, {MaxRentVND: ptr64(-1)}, {MaxMoveInCashVND: ptr64(-1)}, {MaxFloor: ptrInt(-1)}, {MaxUnknown: ptrInt(-1)}} {
		if err := p.Validate(); err == nil {
			t.Fatalf("Validate(%+v) succeeded", p)
		}
	}
}

func TestStoreRoundTripAndClear(t *testing.T) {
	store := Store{Path: filepath.Join(t.TempDir(), "config", "profile.json")}
	profile := SearchProfile{MaxRentVND: ptr64(5_000_000), RequireElevator: true}
	if err := store.Save(profile); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("profile mode = %v", info.Mode().Perm())
	}
	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got.MaxRentVND == nil || *got.MaxRentVND != 5_000_000 || !got.RequireElevator {
		t.Fatalf("loaded profile = %+v", got)
	}
	if err := store.Clear(); err != nil {
		t.Fatal(err)
	}
	got, err = store.Load()
	if err != nil || got != nil {
		t.Fatalf("after clear = %+v, %v", got, err)
	}
}

func TestFactsFromExtraction(t *testing.T) {
	r := &gemini.RentalExtractionResult{Price: "4tr", Deposit: "1 tháng", Floor: "2", RawFields: map[string]string{"elevator": "Có"}}
	f := FactsFromExtraction(r)
	if f.Price != "4tr" || f.Deposit != "1 tháng" || f.Floor != "2" || f.Elevator != "Có" {
		t.Fatalf("facts = %+v", f)
	}
}
