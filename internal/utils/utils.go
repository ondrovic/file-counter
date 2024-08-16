package utils

import (
	"strings"
)

// ToLower converts a string ToLower.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// Contains returns if a string contains and item.
func Contains(s, i string) bool {
	return strings.Contains(s, i)
}
