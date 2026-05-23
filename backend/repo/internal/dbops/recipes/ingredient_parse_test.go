package recipes

import (
	"errors"
	"math/big"
	"testing"
)

func TestParseIngredientLine(t *testing.T) {
	t.Parallel()
	cases := []struct {
		line     string
		wantQty  string // rat string
		wantUnit string
		wantName string
	}{
		{"2 cups all-purpose flour", "2", "cups", "all-purpose flour"},
		{"1/2 tsp salt", "1/2", "tsp", "salt"},
		{"1 1/2 tbsp olive oil", "3/2", "tbsp", "olive oil"},
		{"200 g butter", "200", "g", "butter"},
		{"200g butter", "200", "g", "butter"},
		{"salt", "1", "each", "salt"},
		{"0.25 teaspoon almond extract", "1/4", "teaspoon", "almond extract"},
	}
	for _, tc := range cases {
		t.Run(tc.line, func(t *testing.T) {
			t.Parallel()
			got, err := ParseIngredientLine(tc.line)
			if err != nil {
				t.Fatal(err)
			}
			wantRat := new(big.Rat)
			if _, ok := wantRat.SetString(tc.wantQty); !ok {
				t.Fatalf("bad want qty %q", tc.wantQty)
			}
			if got.Quantity.Cmp(wantRat) != 0 {
				t.Fatalf("qty = %v, want %v", got.Quantity, wantRat)
			}
			if got.Unit != tc.wantUnit {
				t.Fatalf("unit = %q, want %q", got.Unit, tc.wantUnit)
			}
			if got.Name != tc.wantName {
				t.Fatalf("name = %q, want %q", got.Name, tc.wantName)
			}
		})
	}
}

func TestParseIngredientLine_errors(t *testing.T) {
	t.Parallel()
	for _, line := range []string{"", "   "} {
		t.Run(line, func(t *testing.T) {
			t.Parallel()
			_, err := ParseIngredientLine(line)
			if !errors.Is(err, ErrParseIngredient) {
				t.Fatalf("err = %v", err)
			}
		})
	}
}

func TestMergeParsedIngredients_orderAndSum(t *testing.T) {
	t.Parallel()
	a, _ := ParseIngredientLine("1 cup flour")
	b, _ := ParseIngredientLine("2 cup flour")
	out, err := MergeParsedIngredients([]ParsedIngredient{a, b})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 1 {
		t.Fatalf("len = %d", len(out))
	}
	if out[0].Name != "flour" {
		t.Fatalf("name = %q", out[0].Name)
	}
	want := big.NewRat(3, 1)
	if out[0].Quantity.Cmp(want) != 0 {
		t.Fatalf("qty = %v want 3", out[0].Quantity)
	}
}

func TestMergeParsedIngredients_conflictingUnits(t *testing.T) {
	t.Parallel()
	a, _ := ParseIngredientLine("1 cup flour")
	b, _ := ParseIngredientLine("200 g flour")
	_, err := MergeParsedIngredients([]ParsedIngredient{a, b})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrParseIngredient) {
		t.Fatalf("err = %v", err)
	}
}

func TestFormatIngredientLine_roundTripish(t *testing.T) {
	t.Parallel()
	p, err := ParseIngredientLine("2 cups sugar")
	if err != nil {
		t.Fatal(err)
	}
	s := FormatIngredientLine(p)
	if s == "" {
		t.Fatal("empty format")
	}
	again, err := ParseIngredientLine(s)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if again.Quantity.Cmp(p.Quantity) != 0 || again.Unit != p.Unit || again.Name != p.Name {
		t.Fatalf("again = %#v orig = %#v", again, p)
	}
}
