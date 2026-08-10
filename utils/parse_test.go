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
		{"1.5 hours", 90 * time.Minute, true},
		{"1 ½ hours", 90 * time.Minute, true},
		{"about an hour", 0, false},
		{"roughly one hour", 0, false},
		{"garbage", 0, false},
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
		{"-1.5", -1.5, false},
		{"  1.5  ", 1.5, false},
		{"test", 0, true},
		{"", 0, true},
		{"1,2,3", 0, true},
		{"1,234.5", 1234.5, false},
		{"1,234", 1234, false},
		{"1.234,5", 1234.5, false},
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
		give    string
		want    int
		wantErr bool
	}{
		{"123", 123, false},
		{"56 ", 56, false},
		{" 5 ", 5, false},
		{"-5", -5, false},
		{"15.25", 0, true},
		{"33 test", 0, true},
		{"hello world", 0, true},
		{"", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got, err := ParseInt(tt.give)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

func TestParseDate(t *testing.T) {
	t1, ok1 := ParseDate("2023-01-05")
	assert.True(t, ok1)
	assert.Equal(t, 2023, t1.Year())

	t2, ok2 := ParseDate("2023-01-05 10:00:00")
	assert.True(t, ok2)
	assert.Equal(t, 10, t2.Hour())

	_, ok3 := ParseDate("invalid date")
	assert.False(t, ok3)
}
