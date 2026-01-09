package botlogic

import (
	"strings"
	"unicode"
)

// ParseArgs разбирает строку аргументов команды.
// Поддерживает кавычки "..." или '...'.
func ParseArgs(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	var out []string
	var cur strings.Builder
	inQuote := rune(0)

	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}

	for _, r := range s {
		if inQuote != 0 {
			if r == inQuote {
				inQuote = 0
				flush()
				continue
			}
			cur.WriteRune(r)
			continue
		}

		// начало кавычек: двойные "..." или одинарные '...'
		if r == '"' || r == '\'' {
			inQuote = r
			continue
		}

		if unicode.IsSpace(r) {
			flush()
			continue
		}
		cur.WriteRune(r)
	}

	flush()
	return out
}
