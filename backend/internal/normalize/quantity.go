package normalize

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidQuantity = errors.New("invalid quantity format")
	ErrZeroQuantity    = errors.New("quantity must be greater than zero")
	ErrInvalidBaseUnit = errors.New("unsupported base unit or conversion")
	
	quantityRe = regexp.MustCompile(`(?i)^\s*([\d\.]+)\s*([a-z]+)?\s*$`)
)

func ParseQuantity(raw string) (float64, string, error) {
	if strings.TrimSpace(raw) == "" {
		return 0, "", ErrInvalidQuantity
	}
	
	matches := quantityRe.FindStringSubmatch(raw)
	if len(matches) == 0 {
		return 0, "", ErrInvalidQuantity
	}
	
	valStr := matches[1]
	unitStr := strings.ToLower(matches[2])
	
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %v", ErrInvalidQuantity, err)
	}
	
	if unitStr == "" {
		unitStr = "unit"
	}
	
	return val, unitStr, nil
}

func StandardizeUnit(unit string) string {
	u := strings.ToLower(strings.TrimSpace(unit))
	switch u {
	case "g", "grams", "gram":
		return "g"
	case "kg", "kilograms", "kilogram":
		return "kg"
	case "ml", "milliliters", "millilitre", "mliter":
		return "ml"
	case "l", "liters", "litre", "liter":
		return "l"
	case "lb", "lbs", "pounds", "pound":
		return "lb"
	case "oz", "ounces", "ounce":
		return "oz"
	case "ct", "pk", "pack", "count", "units", "unit":
		return "unit"
	default:
		if u == "" {
			return "unit"
		}
		return u
	}
}

func CalculateUnitPrice(totalPrice float64, quantityValue float64, quantityUnit string, baseUnit string) (float64, error) {
	if quantityValue <= 0 {
		return 0, ErrZeroQuantity
	}

	stdUnit := StandardizeUnit(quantityUnit)
	var convertedQuantity float64

	if stdUnit == baseUnit {
		convertedQuantity = quantityValue
	} else {
		switch baseUnit {
		case "kg":
			switch stdUnit {
			case "g":
				convertedQuantity = quantityValue * 0.001
			case "lb":
				convertedQuantity = quantityValue * 0.453592
			case "oz":
				convertedQuantity = quantityValue * 0.0283495
			default:
				return 0, fmt.Errorf("%w: cannot convert %s to kg", ErrInvalidBaseUnit, stdUnit)
			}
		case "g":
			switch stdUnit {
			case "kg":
				convertedQuantity = quantityValue * 1000.0
			case "lb":
				convertedQuantity = quantityValue * 453.592
			case "oz":
				convertedQuantity = quantityValue * 28.3495
			default:
				return 0, fmt.Errorf("%w: cannot convert %s to g", ErrInvalidBaseUnit, stdUnit)
			}
		case "l":
			switch stdUnit {
			case "ml":
				convertedQuantity = quantityValue * 0.001
			default:
				return 0, fmt.Errorf("%w: cannot convert %s to l", ErrInvalidBaseUnit, stdUnit)
			}
		case "ml":
			switch stdUnit {
			case "l":
				convertedQuantity = quantityValue * 1000.0
			default:
				return 0, fmt.Errorf("%w: cannot convert %s to ml", ErrInvalidBaseUnit, stdUnit)
			}
		default:
			return 0, fmt.Errorf("%w: cannot convert %s to %s", ErrInvalidBaseUnit, stdUnit, baseUnit)
		}
	}

	unitPrice := totalPrice / convertedQuantity
	unitPriceRounded := math.Round(unitPrice*100) / 100
	return unitPriceRounded, nil
}
