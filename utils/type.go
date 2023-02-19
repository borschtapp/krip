package utils

import (
	"log"
	"strconv"
)

func FindInt(val interface{}) int {
	switch val.(type) {
	case int:
		return val.(int)
	case float64:
		return int(val.(float64))
	case string:
		val, err := strconv.Atoi(val.(string))
		if err == nil {
			return val
		}
	default:
		log.Printf("FindInt: unexpected type %T of val: %v\n", val, val)
	}

	return 0
}

func FindFloat(val interface{}) float64 {
	switch val.(type) {
	case int:
		return float64(val.(int))
	case float32:
		return float64(val.(float32))
	case float64:
		return val.(float64)
	case string:
		val, err := strconv.ParseFloat(val.(string), 64)
		if err == nil {
			return val
		}
	default:
		log.Printf("FindFloat32: unexpected type %T of val: %v\n", val, val)
	}

	return 0
}
