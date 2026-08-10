package utils

import "slices"

func Deduplicate[E comparable](s []E) []E {
	if len(s) < 2 {
		return slices.Clone(s)
	}

	seen := make(map[E]struct{}, len(s))
	arr := make([]E, 0, len(s))
	for _, item := range s {
		if _, ok := seen[item]; !ok {
			seen[item] = struct{}{}
			arr = append(arr, item)
		}
	}
	return arr
}

func AppendUnique(s []string, v string) []string {
	if !slices.Contains(s, v) {
		s = append(s, v)
	}
	return s
}
