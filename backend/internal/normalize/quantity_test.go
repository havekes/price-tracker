package normalize

import (
	"errors"
	"testing"
)

func TestParseQuantity(t *testing.T) {
	tests := []struct {
		name      string
		raw       string
		wantValue float64
		wantUnit  string
		wantErr   error
	}{
		{"Decimal with unit", "1.5 kg", 1.5, "kg", nil},
		{"Integer with unit", "500 g", 500.0, "g", nil},
		{"No spaces", "500g", 500.0, "g", nil},
		{"Unitless integer", "3", 3.0, "unit", nil},
		{"Unitless decimal", "2.5", 2.5, "unit", nil},
		{"Uppercase unit", "10 KG", 10.0, "kg", nil},
		{"Extra spaces", "  20.5   oz  ", 20.5, "oz", nil},
		{"Empty string", "", 0, "", ErrInvalidQuantity},
		{"Whitespace string", "   ", 0, "", ErrInvalidQuantity},
		{"Malformed text", "abc kg", 0, "", ErrInvalidQuantity},
		{"Zero quantity", "0 ml", 0.0, "ml", nil},
		{"Zero quantity decimal", "0.0 l", 0.0, "l", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotValue, gotUnit, err := ParseQuantity(tt.raw)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ParseQuantity() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotValue != tt.wantValue {
				t.Errorf("ParseQuantity() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
			if gotUnit != tt.wantUnit {
				t.Errorf("ParseQuantity() gotUnit = %v, want %v", gotUnit, tt.wantUnit)
			}
		})
	}
}

func TestStandardizeUnit(t *testing.T) {
	tests := []struct {
		name string
		unit string
		want string
	}{
		{"grams alias 1", "g", "g"},
		{"grams alias 2", "grams", "g"},
		{"kilograms alias 1", "kg", "kg"},
		{"kilograms alias 2", "kilograms", "kg"},
		{"liters alias 1", "l", "l"},
		{"liters alias 2", "liters", "l"},
		{"milliliters alias 1", "ml", "ml"},
		{"milliliters alias 2", "milliliters", "ml"},
		{"pounds alias 1", "lb", "lb"},
		{"pounds alias 2", "lbs", "lb"},
		{"ounces alias 1", "oz", "oz"},
		{"count alias 1", "ct", "unit"},
		{"count alias 2", "pack", "unit"},
		{"count alias 3", "unit", "unit"},
		{"empty", "", "unit"},
		{"unknown", "unknown", "unknown"},
		{"uppercase", "KG", "kg"},
		{"whitespace", "  lbs  ", "lb"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StandardizeUnit(tt.unit); got != tt.want {
				t.Errorf("StandardizeUnit() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateUnitPrice(t *testing.T) {
	tests := []struct {
		name          string
		totalPrice    float64
		quantityValue float64
		quantityUnit  string
		baseUnit      string
		wantPrice     float64
		wantErr       error
	}{
		{"g to kg", 10.00, 500.0, "g", "kg", 20.00, nil},
		{"kg to kg", 20.00, 2.0, "kg", "kg", 10.00, nil},
		{"lb to kg", 10.00, 1.0, "lb", "kg", 22.05, nil}, // 1 / 0.453592 = 2.2046... * 10 = 22.046... rounded to 22.05
		{"oz to kg", 5.00, 10.0, "oz", "kg", 17.64, nil},
		{"kg to g", 5.00, 0.5, "kg", "g", 0.01, nil},
		{"ml to l", 2.50, 500.0, "ml", "l", 5.00, nil},
		{"l to ml", 10.00, 2.0, "l", "ml", 0.01, nil},
		{"unit to unit", 15.00, 3.0, "unit", "unit", 5.00, nil},
		{"pack to unit", 15.00, 3.0, "pack", "unit", 5.00, nil},
		{"zero quantity", 10.00, 0.0, "kg", "kg", 0.0, ErrZeroQuantity},
		{"negative quantity", 10.00, -1.0, "kg", "kg", 0.0, ErrZeroQuantity},
		{"invalid conversion", 10.00, 1.0, "l", "kg", 0.0, ErrInvalidBaseUnit},
		{"unsupported base unit", 10.00, 1.0, "kg", "unknown", 0.0, ErrInvalidBaseUnit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPrice, err := CalculateUnitPrice(tt.totalPrice, tt.quantityValue, tt.quantityUnit, tt.baseUnit)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("CalculateUnitPrice() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotPrice != tt.wantPrice {
				t.Errorf("CalculateUnitPrice() gotPrice = %v, want %v", gotPrice, tt.wantPrice)
			}
		})
	}
}
