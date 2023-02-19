package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFraction(t *testing.T) {
	tests := []struct {
		args string
		want float64
	}{
		{args: "1 ⅓", want: 1.3333333333333333},
		{args: "1 1/3", want: 1.3333333333333333},
	}
	for _, tt := range tests {
		t.Run(tt.args, func(t *testing.T) {
			got, err := ParseFraction(tt.args)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatFraction(t *testing.T) {
	tests := []struct {
		args float64
		want string
	}{
		{args: 1.3333333333333333, want: "1 ⅓"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			assert.Equal(t, tt.want, FormatFraction(tt.args))
		})
	}
}
