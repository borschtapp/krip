package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFraction(t *testing.T) {
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
		{"1 ⅓", 1.3333333333333333, false},
		{"1 1/3", 1.3333333333333333, false},
		{"5", 5, false},
		{"0", 0, false},
		{"  ½  ", 0.5, false},
		{"½½", 1, false},
		{"½¼", 0.75, false},
		{"1/0", 0, true},
		{"1 2 1/3", 0, true},
		{"1/2/3", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseFraction(tt.give)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatFraction(t *testing.T) {
	tests := []struct {
		args float64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{5, "5"},
		{0.25, "¼"},
		{0.5, "½"},
		{0.75, "¾"},
		{0.125, "⅛"},
		{0.375, "⅜"},
		{1.5, "1 ½"},
		{2.25, "2 ¼"},
		{1.3333333333333333, "1 ⅓"},
		{1.1, "1 ⅒"},
		{1.11, "1.11"}, // no matching fraction symbol
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatFraction(tt.args))
		})
	}
}
