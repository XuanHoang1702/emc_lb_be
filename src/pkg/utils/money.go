package utils

import "math"

const moneyEpsilon = 0.01

// MoneyEqual checks if two monetary values are equal within VND precision.
// Handles IEEE 754 floating-point imprecision for currency amounts.
func MoneyEqual(a, b float64) bool {
	return math.Abs(a-b) < moneyEpsilon
}
