// Package preferences stores one local search profile and evaluates rental facts.
package preferences

import "errors"

// SearchProfile contains optional hard requirements. A nil limit is unset.
type SearchProfile struct {
	MaxRentVND       *int64 `json:"max_rent_vnd,omitempty"`
	MaxMoveInCashVND *int64 `json:"max_move_in_cash_vnd,omitempty"`
	MaxDepositMonths *int   `json:"max_deposit_months,omitempty"`
	RequirePets      bool   `json:"require_pets,omitempty"`
	RequireParking   bool   `json:"require_parking,omitempty"`
	RequireElevator  bool   `json:"require_elevator,omitempty"`
	MaxFloor         *int   `json:"max_floor,omitempty"`
	MaxUnknown       *int   `json:"max_unknown,omitempty"`
}

// Validate rejects empty profiles and invalid hard limits.
func (p SearchProfile) Validate() error {
	if p.MaxRentVND == nil && p.MaxMoveInCashVND == nil && p.MaxDepositMonths == nil && !p.RequirePets &&
		!p.RequireParking && !p.RequireElevator && p.MaxFloor == nil {
		return errors.New("hãy đặt ít nhất một điều kiện tìm phòng")
	}
	if p.MaxRentVND != nil && *p.MaxRentVND < 0 {
		return errors.New("giá thuê tối đa không được âm")
	}
	if p.MaxMoveInCashVND != nil && *p.MaxMoveInCashVND < 0 {
		return errors.New("tiền thuê cộng tiền cọc tối đa không được âm")
	}
	if p.MaxDepositMonths != nil && *p.MaxDepositMonths < 0 {
		return errors.New("số tháng cọc tối đa không được âm")
	}
	if p.MaxDepositMonths != nil && p.MaxMoveInCashVND != nil {
		return errors.New("chỉ được đặt một trong hai giới hạn cọc theo tháng hoặc tiền thuê cộng tiền cọc cũ")
	}
	if p.MaxFloor != nil && *p.MaxFloor < 0 {
		return errors.New("tầng tối đa không được âm")
	}
	if p.MaxUnknown != nil && *p.MaxUnknown < 0 {
		return errors.New("số tiêu chí thiếu hoặc chưa rõ tối đa không được âm")
	}
	return nil
}
