package slug

import "strings"

const maxLength = 60

// Make turns text into a URL-friendly slug. Slugs are cosmetic only; lookups
// always use IDs. Non-ASCII text therefore falls back to a neutral value.
func Make(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	pendingDash := false
	for _, r := range strings.ToLower(text) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			if pendingDash && b.Len() > 0 {
				b.WriteByte('-')
			}
			pendingDash = false
			b.WriteRune(r)
			continue
		}
		pendingDash = true
	}

	result := strings.Trim(b.String(), "-")
	if len(result) > maxLength {
		result = strings.TrimRight(result[:maxLength], "-")
	}
	if result == "" {
		return "item"
	}
	return result
}
