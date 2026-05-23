package recipes

import (
	"math/big"
	"testing"
)

func TestRatToNumericString(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   *big.Rat
		want string
	}{
		{big.NewRat(1, 1), "1"},
		{big.NewRat(3, 2), "1.5"},
		{big.NewRat(1, 100), "0.01"},
		{nil, "0"},
	}
	for _, tc := range cases {
		t.Run(tc.want, func(t *testing.T) {
			t.Parallel()
			if got := ratToNumericString(tc.in); got != tc.want {
				t.Fatalf("ratToNumericString(%v) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
