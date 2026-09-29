package preferences

import "soi-tro/internal/gemini"

// FactsFromExtraction gives fresh and saved listings the same local assessment.
func FactsFromExtraction(r *gemini.RentalExtractionResult) RentalFacts {
	if r == nil {
		return RentalFacts{}
	}
	return RentalFacts{
		Price: r.Price, Deposit: r.Deposit, Floor: r.Floor,
		ParkingFee: r.ParkingFee, PetsAllowed: r.PetsAllowed,
		Elevator: r.RawFields["elevator"],
	}
}
