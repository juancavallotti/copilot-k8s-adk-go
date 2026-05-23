package recipes

import (
	"math/big"
	"strings"
)

func ratToNumericString(r *big.Rat) string {
	if r == nil {
		return "0"
	}
	s := r.FloatString(12)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimSuffix(s, ".")
	}
	if s == "" {
		return "0"
	}
	return s
}
