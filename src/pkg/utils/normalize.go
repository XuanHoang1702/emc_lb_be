package utils

import "strings"

func NormalizeStrings(values ...*string) {
	for _, value := range values {
		if value == nil {
			continue
		}

		*value = strings.TrimSpace(*value)
	}
}

func NormalizeEmail(email *string) {
	if email == nil {
		return
	}

	*email = strings.ToLower(strings.TrimSpace(*email))
}
