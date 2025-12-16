package packcalc

import (
	"testing"

	"github.com/NikolaNedicVCS/re-order-packs-calculator/internal/app/models"
)

func resetCalculatorToDefault(t *testing.T) {
	t.Helper()
	SetCalculator(defaultCalculator{})
}

func isSortedDesc(alloc []models.PackAllocation) bool {
	for i := 1; i < len(alloc); i++ {
		if alloc[i-1].Size < alloc[i].Size {
			return false
		}
	}
	return true
}

func TestCalculate_InvalidInputs(t *testing.T) {
	resetCalculatorToDefault(t)

	t.Run("quantity <= 0", func(t *testing.T) {
		if _, err := Calculate(0, []models.PackSize{{ID: 1, Size: 1}}); err != ErrInvalidQuantity {
			t.Fatalf("expected ErrInvalidQuantity, got %v", err)
		}
	})

	t.Run("no pack sizes", func(t *testing.T) {
		if _, err := Calculate(1, nil); err != ErrNoPackSizes {
			t.Fatalf("expected ErrNoPackSizes, got %v", err)
		}
	})

	t.Run("no pack sizes (empty slice)", func(t *testing.T) {
		if _, err := Calculate(1, []models.PackSize{}); err != ErrNoPackSizes {
			t.Fatalf("expected ErrNoPackSizes, got %v", err)
		}
	})

	t.Run("invalid pack sizes", func(t *testing.T) {
		if _, err := Calculate(1, []models.PackSize{{ID: 1, Size: 2}, {ID: 2, Size: 0}}); err != ErrInvalidPackSizes {
			t.Fatalf("expected ErrInvalidPackSizes, got %v", err)
		}
	})

	t.Run("quantity too large", func(t *testing.T) {
		if _, err := Calculate(50_000_000, []models.PackSize{{Size: 250}, {Size: 500}}); err != ErrQuantityTooLarge {
			t.Fatalf("expected ErrQuantityTooLarge, got %v", err)
		}
	})
}

func TestCalculate_SpecificCases(t *testing.T) {
	resetCalculatorToDefault(t)

	cases := []struct {
		name     string
		qty      int
		packs    []models.PackSize
		expected []models.PackAllocation
	}{
		{
			name:     "trivial overage",
			qty:      1,
			packs:    []models.PackSize{{Size: 250}, {Size: 500}},
			expected: []models.PackAllocation{{Size: 250, Count: 1}},
		},
		{
			name:     "exact match",
			qty:      500,
			packs:    []models.PackSize{{Size: 250}, {Size: 500}, {Size: 1000}},
			expected: []models.PackAllocation{{Size: 500, Count: 1}},
		},
		{
			name:     "min shipped then min packs tie",
			qty:      11,
			packs:    []models.PackSize{{Size: 4}, {Size: 6}, {Size: 9}},
			expected: []models.PackAllocation{{Size: 6, Count: 2}}, // 12 with 2 packs beats 12 with 3 packs
		},
		{
			name:     "min shipped prefers smaller overage",
			qty:      12,
			packs:    []models.PackSize{{Size: 5}, {Size: 8}},
			expected: []models.PackAllocation{{Size: 8, Count: 1}, {Size: 5, Count: 1}}, // 13 beats 15/16
		},
		{
			name:     "dedupe + unsorted inputs",
			qty:      8,
			packs:    []models.PackSize{{Size: 4}, {Size: 3}, {Size: 4}, {Size: 6}},
			expected: []models.PackAllocation{{Size: 4, Count: 2}},
		},
		{
			name:     "gaps exist",
			qty:      1,
			packs:    []models.PackSize{{Size: 10}, {Size: 6}},
			expected: []models.PackAllocation{{Size: 6, Count: 1}},
		},
		{
			name:  "large hint case",
			qty:   500_000,
			packs: []models.PackSize{{Size: 23}, {Size: 31}, {Size: 53}},
			expected: []models.PackAllocation{
				{Size: 53, Count: 9429},
				{Size: 31, Count: 7},
				{Size: 23, Count: 2},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Calculate(tc.qty, tc.packs)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if len(got) != len(tc.expected) {
				t.Fatalf("unexpected allocation len=%d expected len=%d; got=%+v expected=%+v", len(got), len(tc.expected), got, tc.expected)
			}
			for i := range tc.expected {
				if got[i] != tc.expected[i] {
					t.Fatalf("unexpected allocation at idx=%d; got=%+v expected=%+v; full got=%+v", i, got[i], tc.expected[i], got)
				}
			}
			if !isSortedDesc(got) {
				t.Fatalf("expected size-descending output, got %+v", got)
			}
			for _, a := range got {
				if a.Count <= 0 {
					t.Fatalf("expected all counts > 0, got %+v", got)
				}
			}
		})
	}
}
