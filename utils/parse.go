package utils

import (
	"errors"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sosodev/duration"
)

var timeRegex = regexp.MustCompile(`(?i)(\D*(?P<days>\d+)\s*(days|D))?(\D*(?P<hours>[\d.\s/?¼½¾⅓⅔⅕⅖⅗]+)\s*(hours|hour|hrs|hr|h|óra))?(\D*(?P<minutes>[\d.]+)\s*(minutes|minute|mins|min|m|perc))?`)

func ParseDuration(str string) (time.Duration, bool) {
	matches := timeRegex.FindStringSubmatch(str)
	if len(matches) == 0 {
		log.Println("unable to parse duration from string: " + str)
		return 0, false
	}

	var d time.Duration
	if days, err := strconv.ParseFloat(matches[2], 32); err == nil && days > 0 {
		d += time.Duration(days) * time.Hour * 24
	}
	if hours, err := ParseFraction(matches[5]); err == nil && hours > 0 {
		d += time.Duration(hours) * time.Hour
	}
	if minutes, err := strconv.ParseFloat(matches[8], 32); err == nil && minutes > 0 {
		d += time.Duration(minutes) * time.Minute
	}
	return d, true
}

func Parse8601Duration(str string) time.Duration {
	if d, err := duration.Parse(str); err == nil {
		return d.ToTimeDuration()
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

func ParseToMillis(str string, baseOpt ...uint) uint {
	if len(str) != 0 {
		var base uint
		if len(baseOpt) > 0 {
			base = baseOpt[0]
		}

		if base == 0 {
			base = 1000
		}

		if strings.Contains(str, "mg") || strings.Contains(str, "milligram") {
			base = 1
		}

		val := FindNumber(str)
		return uint(math.Ceil(val * float64(base)))
	}

	return 0
}
