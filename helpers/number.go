package helpers

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
)

// Round rounds a decimal number to a specified number of decimal places.
// Example: Round(12.5678, 2) returns 12.57.
func Round(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}

// IsEven checks whether an integer is even.
// Example: IsEven(4) returns true.
func IsEven(value int) bool {
	return value%2 == 0
}

// IsOdd checks whether an integer is odd.
// Example: IsOdd(5) returns true.
func IsOdd(value int) bool {
	return value%2 != 0
}

// Absolute returns the absolute value of a number.
// Example: Absolute(-10) returns 10.
func Absolute(value float64) float64 {
	return math.Abs(value)
}

// Percentage calculates a percentage of a number.
// Example: Percentage(200, 10) returns 20.
func Percentage(value, percent float64) float64 {
	return value * percent / 100
}

// Min returns the smaller of two numbers.
func Min(a, b float64) float64 {
	return math.Min(a, b)
}

// Max returns the larger of two numbers.
func Max(a, b float64) float64 {
	return math.Max(a, b)
}

// Clamp keeps a number between a minimum and maximum.
func Clamp(value, minimum, maximum float64) float64 {
	return math.Max(minimum, math.Min(value, maximum))
}

// FormatNumber formats a number with comma separators.
// Example: FormatNumber(1500000) returns "1,500,000".
func FormatNumber(value float64) string {
	return fmt.Sprintf("%,.2f", value)
}

// FormatInteger converts an integer to a string.
func FormatInteger(value int) string {
	return strconv.Itoa(value)
}

// RandomInt returns a random integer between minimum and maximum,
// including both endpoints. It assumes minimum <= maximum.
func RandomInt(minimum, maximum int) int {
	if minimum > maximum {
		return 0
	}

	return rand.Intn(maximum-minimum+1) + minimum
}

// 	// Round a decimal number.
// 	(helpers.Round(12.5678, 2)) // 12.57

// 	// Check even and odd numbers.
// 	(helpers.IsEven(4)) // true
// 	(helpers.IsOdd(5))  // true

// 	// Calculate 10% of 200.
// 	(helpers.Percentage(200, 10)) // 20

// 	// Format a number with commas and two decimal places.
// 	(helpers.FormatNumber(1500000)) // 1,500,000.00

// 	// Convert an integer to a string.
// 	(helpers.FormatInteger(250)) // "250"

// 	// Generate a random integer between 1 and 100.
// 	(helpers.RandomInt(1, 100))
