package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	tests := []struct {
		give   string
		want   time.Duration
		wantOk bool
	}{
		{"1 hour 5 minutes", time.Duration(65) * time.Minute, true},
		{"10 minutes", time.Duration(10) * time.Minute, true},
		{"1 minute", time.Duration(1) * time.Minute, true},
		{"5 min", time.Duration(5) * time.Minute, true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, ok := ParseDuration(tt.give)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantOk, ok)
		})
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		give    string
		want    float64
		wantErr bool
	}{
		{"1", 1, false},
		{"1.256", 1.256, false},
		{"1,35", 1.35, false},
		{"0", 0, false},
		{"test", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseFloat(tt.give)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestParseInt(t *testing.T) {
	tests := []struct {
		give string
		want int
	}{
		{"123", 123},
		{"56 ", 56},
		{"15.25", 0},
		{"33 test", 0},
		{"hello world", 0},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, _ := ParseInt(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseFractions(t *testing.T) {
	tests := []struct {
		give    string
		want    float64
		wantErr bool
	}{
		{"3 ¼", 3.25, false},
		{"2 1/2", 2.5, false},
		{"1/2", 0.5, false},
		{"1/3", 0.3333333333333333, false},
		{"¾", 0.75, false},
		{"⅖", 0.4, false},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseFraction(tt.give)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

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
