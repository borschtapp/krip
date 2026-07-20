package utils

import (
	"errors"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sosodev/duration"
)

var timeRegex = regexp.MustCompile(`(?i)(\D*(?P<days>\d+)\s*(days|D))?(\D*(?P<hours>[\d.\s/?¼½¾⅓⅔⅕⅖⅗]+)\s*(hours|hour|hrs|hr|h|óra))?(\D*(?P<minutes>[\d.]+)\s*(minutes|minute|mins|min|m|perc))?`)

var (
	idxDays    = timeRegex.SubexpIndex("days")
	idxHours   = timeRegex.SubexpIndex("hours")
	idxMinutes = timeRegex.SubexpIndex("minutes")
)

func ParseDuration(str string) (time.Duration, bool) {
	matches := timeRegex.FindStringSubmatch(str)
	if len(matches) == 0 || (matches[idxDays] == "" && matches[idxHours] == "" && matches[idxMinutes] == "") {
		log.Printf("ParseDuration: unable to parse duration from string: %s", str)
		return 0, false
	}

	var d time.Duration
	if days, err := strconv.ParseFloat(matches[idxDays], 32); err == nil && days > 0 {
		d += time.Duration(days * float64(time.Hour*24))
	}
	if hours, err := ParseFraction(matches[idxHours]); err == nil && hours > 0 {
		d += time.Duration(hours * float64(time.Hour))
	}
	if minutes, err := strconv.ParseFloat(matches[idxMinutes], 32); err == nil && minutes > 0 {
		d += time.Duration(minutes * float64(time.Minute))
	}
	return d, true
}

func Parse8601Duration(str string) time.Duration {
	if d, err := duration.Parse(str); err == nil {
		return d.ToTimeDuration()
	}
	return 0
}

func ParseDate(str string) (*time.Time, bool) {
	for _, l := range []string{time.RFC3339, time.DateOnly, time.DateTime, time.RFC1123, time.RFC1123Z} {
		if t, err := time.Parse(l, str); err == nil {
			return &t, true
		}
	}
	return nil, false
}

func ParseInt(str string) (int, error) {
	str = strings.TrimSpace(str)
	if i, err := strconv.Atoi(str); err == nil {
		return i, nil
	}
	return 0, errors.New("unable to parse int from string: " + str)
}

func ParseFloat(str string) (float64, error) {
	str = strings.TrimSpace(str)
	lastComma := strings.LastIndex(str, ",")
	lastDot := strings.LastIndex(str, ".")
	switch {
	case lastComma != -1 && lastDot != -1 && lastComma > lastDot:
		// comma is the decimal separator, dot is the thousands separator: "1.234,56"
		str = strings.Replace(strings.ReplaceAll(str, ".", ""), ",", ".", 1)
	case lastComma != -1 && lastDot != -1:
		// dot is the decimal separator, comma is the thousands separator: "1,234.56"
		str = strings.ReplaceAll(str, ",", "")
	case lastComma != -1 && strings.Count(str, ",") == 1:
		str = strings.Replace(str, ",", ".", 1)
	}
	if i, err := strconv.ParseFloat(str, 64); err == nil {
		return i, nil
	}
	return 0, errors.New("unable to parse float from string: " + str)
}
