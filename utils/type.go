package utils

import (
	"log"
	"strconv"
)

func FindInt(val any) int {
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	default:
		log.Printf("FindInt: unexpected type %T of val: %v\n", val, val)
	}

	return 0
}

func FindFloat(val any) float64 {
	switch v := val.(type) {
	case int:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	case string:
		f, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return f
		}
	default:
		log.Printf("FindFloat32: unexpected type %T of val: %v\n", val, val)
	}

	return 0
}
