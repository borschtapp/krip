package utils

import (
	"math"
	"strings"
)

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
