package helpers

import "math"

func CentsToFloat(amount int64) float64 {
	return float64(amount) / 100
}

func FloatToCents(floatMoney float64) int64 {
	return int64(math.Round(floatMoney * 100))
}
