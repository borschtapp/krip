package utils

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

var paragraphsRegex = regexp.MustCompile(`\n\r?\s*\n\r?`)

var (
	ugcPolicy    = bluemonday.UGCPolicy()
	strictPolicy = bluemonday.StrictPolicy()
)

func Cleanup(s string) string {
	s = ugcPolicy.Sanitize(s)
	s = cleanupCommon(s)

	var b strings.Builder
	for line := range strings.SplitSeq(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(line)
	}
	return b.String()
}

func CleanupInline(s string) string {
	s = strictPolicy.Sanitize(s)
	s = cleanupCommon(s)
	s = strings.Trim(s, ",;")
	s = RemoveNewLines(s)
	return s
}

func cleanupCommon(s string) string {
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\u00a0", " ")
	s = strings.ReplaceAll(s, "\u00ad", "-")
	s = Unquote(s)
	s = RemoveDoubleSpace(s)
	s = strings.ReplaceAll(s, " , ", ", ")
	s = strings.ReplaceAll(s, " : ", ": ")
	return s
}

func TrimZeroWidthSpaces(s string) string {
	return strings.TrimFunc(s, func(r rune) bool {
		return r == 0x200B || r == 0xFEFF || r == 0x200C || r == 0x200D
	})
}

func SplitTitle(title string) []string {
	title = strings.ReplaceAll(title, " - ", "|")
	title = strings.ReplaceAll(title, " – ", "|")
	title = strings.ReplaceAll(title, " — ", "|")
	title = strings.ReplaceAll(title, " • ", "|")
	title = strings.ReplaceAll(title, " :: ", "|")
	parts := strings.Split(title, "|")

	var result []string
	for _, p := range parts {
		cleaned := CleanupInline(p)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}
	return result
}

func RemoveDoubleSpace(str string) string {
	return strings.Join(strings.FieldsFunc(str, func(r rune) bool {
		return r != '\n' && unicode.IsSpace(r)
	}), " ")
}

func Unquote(s string) string {
	trimStrict := func(s, open, close string) string {
		if strings.HasPrefix(s, open) && strings.HasSuffix(s, close) &&
			strings.Count(s, open) == 1 && strings.Count(s, close) == 1 {
			return s[len(open) : len(s)-len(close)]
		}
		return s
	}

	trimQuote := func(s, ch string) string {
		hasPrefix := strings.HasPrefix(s, ch)
		hasSuffix := strings.HasSuffix(s, ch)
		if !hasPrefix && !hasSuffix {
			return s
		}

		start, end := 0, len(s)
		if hasPrefix {
			start = 1
		}
		if hasSuffix {
			end = len(s) - 1
		}
		if start > end {
			return s
		}

		if strings.Count(s[start:end], ch)%2 != 0 {
			return s
		}
		return s[start:end]
	}

	s = trimStrict(s, "“", "”")
	s = trimStrict(s, "«", "»")
	s = trimQuote(s, "\"")
	return s
}

func RemoveSpaces(s string) string {
	return strings.Join(strings.Fields(s), "")
}

func SplitParagraphs(s string) []string {
	// TODO: check for colon, add it as a section (bigoven, blueapron)
	split := paragraphsRegex.Split(s, -1)

	var result []string
	for _, p := range split {
		p = CleanupInline(p)
		if len(p) != 0 {
			result = append(result, p)
		}
	}

	return result
}

func RemoveNewLines(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

var numberRegex = regexp.MustCompile(`\d+([.,]\d+)?`)

func FindNumber(str string) *float64 {
	groups := numberRegex.FindAllString(str, 1)
	for _, group := range groups {
		if i, err := ParseFloat(group); err == nil {
			return &i
		}
	}
	return nil
}
