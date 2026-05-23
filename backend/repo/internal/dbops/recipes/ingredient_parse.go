package recipes

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"sort"
	"strings"
)

// ParsedIngredient is the structured form of one ingredients[] line.
type ParsedIngredient struct {
	Quantity *big.Rat // always non-nil after a successful parse
	Unit     string   // normalized lowercase for storage, except display casing from list
	Name     string   // trimmed display name
}

var (
	reMixedFraction = regexp.MustCompile(`(?i)^\s*(\d+)\s+(\d+)\s*/\s*(\d+)`)
	reSlashFraction = regexp.MustCompile(`(?i)^\s*(\d+)\s*/\s*(\d+)`)
	reDecimal       = regexp.MustCompile(`(?i)^\s*(\d+\.\d*|\d*\.\d+)`)
	reInteger       = regexp.MustCompile(`^\s*(\d+)`)
)

// knownUnits is sorted longest-first in init().
var knownUnits []string

func init() {
	seen := map[string]struct{}{}
	raw := []string{
		"fluid ounces", "fl. oz.", "fl oz",
		"milliliters", "milliliter", "millilitres", "millilitre",
		"tablespoons", "tablespoon", "teaspoons", "teaspoon",
		"kilograms", "kilogram", "milligrams", "milligram",
		"gallons", "gallon", "ounces", "ounce", "pounds", "pound",
		"quarts", "quart", "pints", "pint",
		"cups", "cup", "tbsp", "tbl", "tsp", "lbs", "lb",
		"slices", "slice", "cloves", "clove",
		"pinches", "pinch", "dashes", "dash",
		"grams", "gram",
		"liters", "liter", "litres", "litre",
		"gal", "oz", "kg", "ml", "qt", "pt",
		"T",
		"g", "l", "t",
	}
	byLen := append([]string(nil), raw...)
	sort.Slice(byLen, func(i, j int) bool { return len(byLen[i]) > len(byLen[j]) })
	for _, u := range byLen {
		key := strings.ToLower(strings.TrimSpace(u))
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		knownUnits = append(knownUnits, u)
	}
}

// ParseIngredientLine parses strings like "2 cups flour", "1 1/2 tbsp oil", "200g butter", "salt".
func ParseIngredientLine(line string) (ParsedIngredient, error) {
	raw := strings.TrimSpace(line)
	if raw == "" {
		return ParsedIngredient{}, fmt.Errorf("%w: empty line", ErrParseIngredient)
	}

	qty, rest, hadQty := parseLeadingQuantity(raw)
	if !hadQty {
		name := normalizeIngredientName(raw)
		if name == "" {
			return ParsedIngredient{}, fmt.Errorf("%w: no ingredient name", ErrParseIngredient)
		}
		return ParsedIngredient{Quantity: big.NewRat(1, 1), Unit: "each", Name: name}, nil
	}

	rest = strings.TrimLeft(rest, " \t")
	unit, afterUnit, ok := takeKnownUnit(rest)
	if !ok || unit == "" {
		name := normalizeIngredientName(rest)
		if name == "" {
			return ParsedIngredient{}, fmt.Errorf("%w: remainder empty after quantity in %q", ErrParseIngredient, raw)
		}
		return ParsedIngredient{Quantity: new(big.Rat).Set(qty), Unit: "each", Name: name}, nil
	}

	name := normalizeIngredientName(afterUnit)
	if name == "" {
		return ParsedIngredient{}, fmt.Errorf("%w: missing ingredient name after unit in %q", ErrParseIngredient, raw)
	}
	return ParsedIngredient{Quantity: new(big.Rat).Set(qty), Unit: strings.ToLower(unit), Name: name}, nil
}

func parseLeadingQuantity(s string) (qty *big.Rat, rest string, matched bool) {
	switch {
	case reMixedFraction.MatchString(s):
		m := reMixedFraction.FindStringSubmatchIndex(s)
		whole := mustInt(s[m[2]:m[3]])
		num := mustInt(s[m[4]:m[5]])
		den := mustInt(s[m[6]:m[7]])
		if den == 0 {
			return nil, "", false
		}
		q := fracFromMixed(whole, num, den)
		return q, s[m[1]:], true
	case reSlashFraction.MatchString(s):
		m := reSlashFraction.FindStringSubmatchIndex(s)
		num := mustInt(s[m[2]:m[3]])
		den := mustInt(s[m[4]:m[5]])
		if den == 0 {
			return nil, "", false
		}
		q := big.NewRat(int64(num), int64(den))
		return q, s[m[1]:], true
	case reDecimal.MatchString(s):
		m := reDecimal.FindStringSubmatchIndex(s)
		txt := s[m[2]:m[3]]
		q, err := parseDecimalRat(txt)
		if err != nil {
			return nil, "", false
		}
		return q, s[m[1]:], true
	case reInteger.MatchString(s):
		m := reInteger.FindStringSubmatchIndex(s)
		n := mustInt(s[m[2]:m[3]])
		q := big.NewRat(int64(n), 1)
		return q, s[m[1]:], true
	default:
		return nil, s, false
	}
}

func fracFromMixed(whole, num, den int) *big.Rat {
	if den == 0 {
		return big.NewRat(0, 1)
	}
	w := big.NewRat(int64(whole), 1)
	frac := big.NewRat(int64(num), int64(den))
	return new(big.Rat).Add(w, frac)
}

func parseDecimalRat(s string) (*big.Rat, error) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, errors.New("invalid decimal")
	}
	return r, nil
}

func mustInt(s string) int {
	s = strings.TrimSpace(s)
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func takeKnownUnit(s string) (unit string, rest string, ok bool) {
	s = strings.TrimLeft(s, " \t")
	if s == "" {
		return "", "", false
	}
	lower := strings.ToLower(s)
	for _, cand := range knownUnits {
		cl := strings.ToLower(cand)
		if !strings.HasPrefix(lower, cl) {
			continue
		}
		if !unitMatchBoundary(s, len(cand)) {
			continue
		}
		rest := strings.TrimSpace(s[len(cand):])
		return cand, rest, true
	}
	return "", "", false
}

// unitMatchBoundary requires a separator after the unit when more text follows.
// Glued metric forms like "200g butter" still work because the unit letter is followed by a space.
func unitMatchBoundary(original string, runeLen int) bool {
	if len(original) < runeLen {
		return false
	}
	if len(original) == runeLen {
		return true
	}
	next := original[runeLen]
	return next == ' ' || next == '\t' || next == ',' || next == ';' || next == '(' || next == '.'
}

func normalizeIngredientName(s string) string {
	s = strings.Trim(s, " \t,;")
	s = strings.Trim(s, ".")
	return strings.TrimSpace(s)
}

// MergeParsedIngredients merges duplicate ingredient names that use the same unit by summing quantities.
// Order follows the first appearance of each ingredient name (case-insensitive key).
func MergeParsedIngredients(rows []ParsedIngredient) ([]ParsedIngredient, error) {
	byName := map[string][]ParsedIngredient{}
	var order []string
	for _, r := range rows {
		nk := strings.ToLower(strings.TrimSpace(r.Name))
		if nk == "" {
			return nil, fmt.Errorf("%w: empty ingredient name", ErrParseIngredient)
		}
		if _, ok := byName[nk]; !ok {
			order = append(order, nk)
		}
		byName[nk] = append(byName[nk], r)
	}
	out := make([]ParsedIngredient, 0, len(order))
	for _, nk := range order {
		group := byName[nk]
		u := strings.ToLower(strings.TrimSpace(group[0].Unit))
		sum := new(big.Rat).Set(group[0].Quantity)
		name := strings.TrimSpace(group[0].Name)
		for i := 1; i < len(group); i++ {
			u2 := strings.ToLower(strings.TrimSpace(group[i].Unit))
			if u2 != u {
				return nil, fmt.Errorf("%w: duplicate ingredient %q with different units %q vs %q", ErrParseIngredient, name, u, u2)
			}
			sum.Add(sum, group[i].Quantity)
		}
		if sum.Sign() < 0 {
			return nil, fmt.Errorf("%w: negative total quantity for %q", ErrParseIngredient, name)
		}
		out = append(out, ParsedIngredient{Quantity: sum, Unit: u, Name: name})
	}
	return out, nil
}

func formatQtyForLine(r *big.Rat) string {
	if r == nil {
		return "0"
	}
	if r.IsInt() {
		return r.FloatString(0)
	}
	s := r.FloatString(6)
	s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	if s == "" {
		return "0"
	}
	return s
}

// FormatIngredientLine rebuilds a single human-readable line for API responses.
func FormatIngredientLine(p ParsedIngredient) string {
	u := strings.TrimSpace(p.Unit)
	n := strings.TrimSpace(p.Name)
	q := formatQtyForLine(p.Quantity)
	if u == "" || u == "each" {
		if p.Quantity.Cmp(big.NewRat(1, 1)) == 0 {
			return n
		}
		return strings.TrimSpace(q + " " + n)
	}
	return strings.TrimSpace(q + " " + u + " " + n)
}
