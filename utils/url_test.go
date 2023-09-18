package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseUrl(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"https://example.com/test-page", "https://example.com"},
		{"https://example.org/test-page", "https://example.org"},
		{"https://www.example.com.ua/test-page", "https://www.example.com.ua"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := BaseUrl(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDomainZone(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"https://example.com/test-page", "com"},
		{"https://example.org/test-page", "org"},
		{"https://www.example.com.ua/test-page", "ua"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := DomainZone(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHostAlias(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"https://example.com/test-page", "example"},
		{"https://example.org/test-page", "example"},
		{"https://www.example.com.ua/test-page", "example"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := HostAlias(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHostname(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"example.com", "example.com"},
		{"https://www.example.com/test-page", "example.com"},
		{"https://example.com/", "example.com"},
		{"https://example.com/chicken-broccoli-sweet-potatoes-meal-prep/", "example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := Hostname(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
