package umeshu

import (
	"strings"

	"github.com/knchan0x/umeshu/log"
)

// cleanPrefix returns clean prefix.
func cleanPrefix(prefix string) string {
	if prefix[0] != '/' {
		prefix = "/" + prefix
	}
	if prefix[len(prefix)-1] == '/' {
		prefix = prefix[:len(prefix)-1]
	}
	return prefix
}

// parsePattern parses path into string slice.
func parsePattern(pattern string) []string {
	s := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, part := range s {
		if part != "" {
			parts = append(parts, part)
			if part[0] == '*' {
				break // only one "*" is allowed
			}
		}
	}
	return parts
}

// asset check guard condition, panic if not true.
func assert(guard bool, errMsg string) {
	if !guard {
		log.Panic(errMsg)
	}
}
