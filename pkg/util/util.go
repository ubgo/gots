// Package util provides string case conversion helpers used across the gots codebase.
//
// These are used for transforming Go naming conventions to TypeScript conventions:
//   - ToCamelCase: PascalCase → camelCase (for TS field names)
//   - ToSnakeCase: PascalCase → snake_case (for JSON field names)
package util

import (
	"strings"
	"unicode"
)

// ToCamelCase converts PascalCase or snake_case to camelCase.
//
// This is a simple lowercase-first-letter conversion. It handles PascalCase input
// (which is the common case for Go exported names).
//
// Examples:
//
//	"MessageRole"  → "messageRole"
//	"ID"           → "iD"
//	"URLParser"    → "uRLParser"
//	""             → ""
//
// Note: does not handle snake_case → camelCase conversion.
// For Go names, PascalCase → camelCase is the typical need.
func ToCamelCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// ToSnakeCase converts PascalCase or camelCase to snake_case.
//
// Inserts an underscore before each uppercase letter (except the first)
// and lowercases everything.
//
// Examples:
//
//	"MessageRole" → "message_role"
//	"UserID"      → "user_i_d"     (naive — doesn't handle acronyms)
//	"createdAt"   → "created_at"
//	""            → ""
//
// Note: acronym handling (UserID → user_id) is not implemented yet.
// This is a known limitation — most struct fields use json tags anyway.
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
