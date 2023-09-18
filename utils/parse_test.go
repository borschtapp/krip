package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseToMillis(t *testing.T) {
	tests := []struct {
		give string
		want uint
	}{
		{"5", 5000},
		{"5.3 g", 5300},
		{"0 g", 0},
		{"0.0 g", 0},
		{"8 grams saturated fat", 8000},
		{"6 g", 6000},
		{"19 grams", 19000},
		{"0.0386162663043889 g", 39},
		{"3 mg", 3},
		{"801.5 mg", 802},
		{"802,6 mg", 803},
		{"31 milligrams cholesterol", 31},
		{"milligrams", 0},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := ParseToMillis(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
