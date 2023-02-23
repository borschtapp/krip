package utils

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
	"golang.org/x/net/html"
	"golang.org/x/net/publicsuffix"
)

var paragraphsRegex = regexp.MustCompile(`\n\r?\s*\n\r?`)

func IsAbsolute(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Malformed url:", err)
		return false
	}

	if u.IsAbs() {
		return true
	}

	return false
}

func ToAbsoluteUrl(base *url.URL, urlStr string) string {
	if len(urlStr) == 0 || base == nil {
		return ""
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		fmt.Println("Error parsing url:", err)
		return ""
	}

	if u.IsAbs() {
		return urlStr
	}

	return base.ResolveReference(u).String()
}

func RemoveTrailingSlash(urlStr string) string {
	return strings.TrimSuffix(urlStr, "/")
}

func CleanupLang(lang string) string {
	lang = strings.ReplaceAll(lang, "_", "-")
	return lang
}

func BaseUrl(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String()
}

func Hostname(urlStr string) string {
	if !strings.Contains(urlStr, "/") {
		return urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := strings.ToLower(u.Hostname())
	host = strings.TrimPrefix(host, "www.")
	return host
}

func DomainZone(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	hostParts := strings.Split(strings.ToLower(u.Hostname()), ".")
	return hostParts[len(hostParts)-1]
}

func HostAlias(urlStr string) string {
	alias := Hostname(urlStr)

	// remove public domain
	suffix, _ := publicsuffix.PublicSuffix(alias)
	alias = strings.TrimSuffix(alias, "."+suffix)

	// remove common prefixes
	alias = strings.TrimPrefix(alias, "api.")

	// replace dots with underscores
	alias = strings.ReplaceAll(alias, ".", "_")
	return alias
}

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
	s = strings.ReplaceAll(s, "­", "-")
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
	splitter := func(r rune) bool {
		return r == '|' || r == '-' || r == '–'
	}
	return strings.FieldsFunc(title, splitter)
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

var timeRegex = regexp.MustCompile(`(?i)(\D*(?P<days>\d+)\s*(days|D))?(\D*(?P<hours>[\d.\s/?¼½¾⅓⅔⅕⅖⅗]+)\s*(hours|hour|hrs|hr|h|óra))?(\D*(?P<minutes>[\d.]+)\s*(minutes|minute|mins|min|m|perc))?`)

func ParseDuration(str string) (time.Duration, bool) {
	matches := timeRegex.FindStringSubmatch(str)
	if len(matches) == 0 {
		log.Println("unable to parse duration from string: " + str)
		return 0, false
	}

	var duration time.Duration
	if days, err := strconv.ParseFloat(matches[2], 32); err == nil && days > 0 {
		duration += time.Duration(days) * time.Hour * 24
	}
	if hours, err := ParseFraction(matches[5]); err == nil && hours > 0 {
		duration += time.Duration(hours) * time.Hour
	}
	if minutes, err := strconv.ParseFloat(matches[8], 32); err == nil && minutes > 0 {
		duration += time.Duration(minutes) * time.Minute
	}
	return duration, true
}

var numberRegex = regexp.MustCompile("\\d+([.,]\\d+)?")

func FindNumber(str string) float64 {
	groups := numberRegex.FindAllString(str, 1)
	for _, group := range groups {
		if i, err := ParseFloat(group); err == nil {
			return i
		}
	}
	return 0
}

func ParseInt(str string) (int, error) {
	str = strings.TrimSpace(str)
	if i, err := strconv.Atoi(str); err == nil {
		return i, nil
	}

	return 0, errors.New("unable to parse int from string: " + str)
}

func ParseFloat(str string) (float64, error) {
	str = strings.Replace(strings.TrimSpace(str), ",", ".", 1)
	if i, err := strconv.ParseFloat(str, 64); err == nil {
		return i, nil
	}

	return 0, errors.New("unable to parse float from string: " + str)
}
