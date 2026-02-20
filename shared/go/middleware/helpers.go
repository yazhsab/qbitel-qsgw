package middleware

import "strings"

// ParseAPIKeyEntries parses API key configuration strings in the format "key:subject:role".
// Invalid entries (wrong number of parts) are silently skipped.
func ParseAPIKeyEntries(entries []string) []APIKeyEntry {
	var result []APIKeyEntry
	for _, entry := range entries {
		parts := strings.SplitN(entry, ":", 3)
		if len(parts) != 3 {
			continue
		}
		result = append(result, APIKeyEntry{
			Key:     strings.TrimSpace(parts[0]),
			Subject: strings.TrimSpace(parts[1]),
			Role:    strings.TrimSpace(parts[2]),
		})
	}
	return result
}
