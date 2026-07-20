package utils

import (
	"context"
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

func TestRemoveTrailingSlash(t *testing.T) {
	assert.Equal(t, "https://Example.com/Path", RemoveTrailingSlash("https://Example.com/Path/"))
	assert.Equal(t, "https://Example.com/Path", RemoveTrailingSlash("https://Example.com/Path"))
}

func TestIsPrivateOrLoopbackHost(t *testing.T) {
	ctx := context.Background()

	// Safe public hosts
	assert.False(t, IsPrivateOrLoopbackHost(ctx, "example.com"))
	assert.False(t, IsPrivateOrLoopbackHost(ctx, "8.8.8.8"))
	assert.False(t, IsPrivateOrLoopbackHost(ctx, "google.com"))

	// Private/loopback IPs
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "127.0.0.1"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "::1"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "10.0.0.1"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "172.16.0.1"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "192.168.1.1"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "169.254.0.1"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "0.0.0.0"))

	// Private/loopback Hostnames
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "localhost"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "test.local"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, "my.internal"))
	assert.True(t, IsPrivateOrLoopbackHost(ctx, ""))
}
