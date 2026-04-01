package utils

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
)

var paragraphsRegex = regexp.MustCompile(`\n\r?\s*\n\r?`)

func Cleanup(s string) string {
	s = bluemonday.UGCPolicy().Sanitize(s)
	s = cleanupCommon(s)

	var res string
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if len(res) > 0 {
			res += "\n"
		}
		res += line
	}
	return res
}

func CleanupInline(s string) string {
	s = bluemonday.StrictPolicy().Sanitize(s)
	s = cleanupCommon(s)
	s = strings.Trim(s, ",;")
	s = RemoveNewLines(s)
	return s
}

func cleanupCommon(s string) string {
	s = html.UnescapeString(s)
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", " ")
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
		if uint32(r) <= unicode.MaxLatin1 {
			switch r {
			case '\t', '\v', '\f', '\r', ' ', 0x85, 0xA0:
				return true
			}
			return false
		}
		return false
	}), " ")
}

func Unquote(s string) string {
	if strings.HasPrefix(s, "\"") {
		if strings.HasSuffix(s, "\"") {
			if strings.Count(s[1:len(s)-1], "\"")%2 == 0 {
				return s[1 : len(s)-1]
			}
		} else if strings.Count(s[1:], "\"")%2 == 0 {
			return s[1:]
		}
	} else if strings.HasSuffix(s, "\"") && strings.Count(s[:len(s)-1], "\"")%2 == 0 {
		return s[:len(s)-1]
	}
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

func Coalesce(dst, src string) string {
	if len(dst) == 0 {
		return src
	}
	return dst
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
