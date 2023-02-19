package utils

import "golang.org/x/exp/slices"

func Deduplicate[E comparable](s []E) []E {
	if len(s) < 2 {
		return s
	}

	allKeys := make(map[E]bool)
	var arr []E
	for _, item := range s {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			arr = append(arr, item)
		}
	}
	return arr
}

func AppendUnique[E comparable](s []E, v E) []E {
	if !slices.Contains(s, v) {
		s = append(s, v)
	}
	return s
}
