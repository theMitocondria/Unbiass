package utils

import (
	"strings"
)
func CleanString(s string) string {
    // Remove all newlines and spaces
    s = strings.ReplaceAll(s, "\n", "")
    s = strings.ReplaceAll(s, " ", "")
    return strings.TrimSpace(s)
}