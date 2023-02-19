package utils

import (
	"testing"
	"time"

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

func TestCleanup(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{`remove
new lines`, "remove\nnew lines"},
		{`hello    world `, "hello world"},
		{`hello world `, "hello world"},
		{`hello , world `, "hello, world"},
		{`&quot;Fran &amp; Freddie&#39;s Diner&quot;`, "Fran & Freddie's Diner"},
		{`&quot;Fran &amp; Freddie&#39;s Diner&quot; test`, "\"Fran & Freddie's Diner\" test"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := Cleanup(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCleanupInline(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{"Info\n  \n\n\n6\noz.\nLinguine", "Info 6 oz. Linguine"},
		{"remove\nnew lines", "remove new lines"},
		{`"hello    world "`, "hello world"},
		{`hello world `, "hello world"},
		{`hello, world `, "hello, world"},
		{`hello , world `, "hello, world"},
		{`&quot;Fran &amp; Freddie&#39;s Diner&quot;`, "Fran & Freddie's Diner"},
		{`&quot;Fran &amp; Freddie&#39;s Diner&quot; test`, "\"Fran & Freddie's Diner\" test"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := CleanupInline(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveDoubleSpace(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{`hello world`, `hello world`},
		{`hello   world`, `hello world`},
		{`hello world `, `hello world`},
		{`  hello world `, `hello world`},
		{`  hello	world  `, `hello world`},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := RemoveDoubleSpace(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestUnquote(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{`"hello "test" world"`, `hello "test" world`},
		{`hello "test" world`, `hello "test" world`},
		{`"hello "test" world`, `hello "test" world`},
		{`"hello world`, `hello world`},
		{`hello world"`, `hello world`},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := Unquote(tt.give)
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

func TestFindNumber(t *testing.T) {
	tests := []struct {
		give string
		want float64
	}{
		{"8-10 servings", 8},
		{"123", 123},
		{"123.25", 123.25},
		{"this is decimal 123.25", 123.25},
		{"33 test", 33},
		{"this is 15 inside text", 15},
		{"use only first number 12", 12},
		{"use only 1 st number 12", 1},
		{"hello world", 0},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := FindNumber(tt.give)
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

func TestRemoveNewLines(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{`remove
new lines`, "remove new lines"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := RemoveNewLines(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRemoveSpaces(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{`  test`, "test"},
		{`remove   all spaces   `, "removeallspaces"},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := RemoveSpaces(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSplitParagraphs(t *testing.T) {
	tests := []struct {
		give string
		want []string
	}{
		{`test paragraph

second paragraph`, []string{"test paragraph", "second paragraph"}},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := SplitParagraphs(tt.give)
			assert.Equal(t, tt.want, got)
		})
	}
}
