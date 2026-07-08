package render

import "strings"

const labelCharWidth = 7

// ASCII replaces German letters unsupported by the bitmap font (ae/oe/ue/ss).
func ASCII(s string) string {
	return strings.NewReplacer(
		"ä", "ae", "ö", "oe", "ü", "ue",
		"Ä", "Ae", "Ö", "Oe", "Ü", "Ue",
		"ß", "ss",
		"—", "-",
	).Replace(s)
}

// WrapCaption splits caption paragraphs to fit maxWidthPx using the bitmap font width.
func WrapCaption(caption string, maxWidthPx int) []string {
	if caption == "" {
		return nil
	}
	maxChars := maxWidthPx / labelCharWidth
	if maxChars < 12 {
		maxChars = 12
	}
	var out []string
	for _, paragraph := range strings.Split(caption, "\n") {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			continue
		}
		out = append(out, wrapParagraph(paragraph, maxChars)...)
	}
	return out
}

func wrapParagraph(text string, maxChars int) []string {
	text = ASCII(text)
	if len(text) <= maxChars {
		return []string{text}
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, 4)
	var current string
	for _, word := range words {
		for len(word) > maxChars {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			lines = append(lines, word[:maxChars])
			word = word[maxChars:]
		}
		if current == "" {
			current = word
			continue
		}
		candidate := current + " " + word
		if len(candidate) <= maxChars {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = word
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func captionTextWidth(lines []string) int {
	max := 0
	for _, line := range lines {
		if w := len(ASCII(line)) * labelCharWidth; w > max {
			max = w
		}
	}
	return max
}
