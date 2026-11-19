package components

import (
	"strings"
	"unicode"
)

const khayalPreviewLimit = 68

// truncateKhayal keeps a preview on a single line and cuts it at a word
// boundary. The browser refines the preview further to fit the available
// width, but this keeps a sensible fallback when JavaScript is unavailable.
func truncateKhayal(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= khayalPreviewLimit {
		return text
	}

	cut := 0
	for i := khayalPreviewLimit; i > 0; i-- {
		if unicode.IsSpace(runes[i]) {
			cut = i
			break
		}
	}
	if cut == 0 {
		cut = khayalPreviewLimit
	}

	return strings.TrimSpace(string(runes[:cut])) + ".."
}
