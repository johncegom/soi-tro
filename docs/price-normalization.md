# Rental-price normalization

`internal/priceparser.ParseVND` converts the free-text Vietnamese price
Gemini extracts (`RentalExtractionResult.Price`) into an integer VND amount.
`database.SaveRental` calls it and stores the result in the `price_vnd`
column; the original text is always kept in `price` unchanged. When the
price cannot be unambiguously parsed, `price_vnd` is left `NULL` rather than
guessed, and saving the rental still succeeds.

## Supported syntax

Input is case-insensitive; internal and surrounding whitespace is ignored.
A trailing rent period (`/tháng`, `/tuần`, `/ngày`, `/năm`, with or without
diacritics) and a trailing currency marker (`đ`, `vnd`, `vnđ`) are stripped
before parsing.

| Form | Example | Result (VND) |
|---|---|---|
| `<int>tr` / `<int>m` | `4tr` | 4,000,000 |
| `<int>tr<digit>` / `<int>m<digit>` | `4tr5`, `4m5` | 4,500,000 |
| `<int>[.,]<digits>tr` / `...m` | `4.5tr`, `4,5 triệu` | 4,500,000 |
| `<int>k` | `4500k` | 4,500,000 |
| `<int>[.,]<digits>k` | `4.5k` | 4,500 |
| Dot-grouped literal | `4.500.000` | 4,500,000 |
| Bare integer (>= 1,000, or exactly 0) | `5000000` | 5,000,000 |

`tr`/`m` mean *triệu* (million); `k` means *nghìn* (thousand). `triệu`,
`trieu`, `nghìn`, and `nghin` are accepted spellings.

## Rounding and rejection rules

All arithmetic uses integers; no floating-point rounding is applied. A
decimal form (`4.5tr`, `4.5k`, ...) is accepted only when it resolves to a
whole VND amount, and is rejected otherwise.

The parser rejects rather than guesses:

- empty or non-numeric input;
- a bare decimal number with no unit (e.g. `4.5`), since the intended scale
  is unclear;
- a bare integer below 1,000 (other than exactly `0`, accepted as an
  explicit free-rent value), since it is too small to be an unambiguous VND
  amount;
- a negative number, multiple units on one value, or any other form not
  listed above.

## Schema migration

`price_vnd` was added after the initial schema. `InitDB` adds the column to
any existing `rentals` table that predates it, leaving already-saved rows
with `price_vnd = NULL` until they are re-saved.

## Manual test plan

`internal/priceparser/testdata/manual_price_dataset.csv` lists representative
listing strings with their expected outcome (`accept` with the exact VND
value, or `reject`). `TestParseVND_ManualDataset` runs this file
automatically on every test run, so it doubles as a manual reference and a
regression test: to manually verify a real listing string, run it through
`ParseVND` (or save a rental with that price and check `price_vnd` in
history) and compare against the closest row in the dataset.
