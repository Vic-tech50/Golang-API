package helpers

import (
	"strings"
	"unicode"
)

// Capitalize converts the first character to uppercase.
// Example: "hello world" -> "Hello world"
func Capitalize(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return value
	}

	runes := []rune(value)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// TitleCase capitalizes the first letter of each word.
// Example: "hello world" -> "Hello World"
func TitleCase(value string) string {
	return strings.Title(strings.ToLower(value))
}

// Uppercase converts all characters to uppercase.
// Example: "hello" -> "HELLO"
func Uppercase(value string) string {
	return strings.ToUpper(value)
}

// Lowercase converts all characters to lowercase.
// Example: "HELLO" -> "hello"
func Lowercase(value string) string {
	return strings.ToLower(value)
}

// Trim removes spaces from the beginning and end.
// Example: "  hello  " -> "hello"
func Trim(value string) string {
	return strings.TrimSpace(value)
}

// Reverse reverses a string while preserving Unicode characters.
// Example: "hello" -> "olleh"
func Reverse(value string) string {
	runes := []rune(value)

	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}

// Contains checks whether a string contains another string.
// Example: Contains("hello world", "world") -> true
func Contains(value, search string) bool {
	return strings.Contains(value, search)
}

// RemoveSpaces removes all whitespace characters from a string.
// Example: "hello world" -> "helloworld"
func RemoveSpaces(value string) string {
	return strings.Join(strings.Fields(value), "")
}

// Slug converts text into a simple URL-friendly format.
// Example: "My New Product" -> "my-new-product"
func Slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	words := strings.Fields(value)
	return strings.Join(words, "-")
}

//
// 	helpers.Capitalize("hello world"))
// 	helpers.TitleCase("hello world"))
// 	helpers.Uppercase("hello"))
// 	helpers.Lowercase("HELLO"))
// 	helpers.Trim("  hello  "))
// 	helpers.Reverse("hello"))
// 	helpers.Contains("hello world", "world"))
// 	helpers.RemoveSpaces("hello world"))
// 	helpers.Slug("My New Product"))
//
