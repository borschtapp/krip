package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestFindNumber(t *testing.T) {
	tests := []struct {
		give string
		want *float64
	}{
		{"8-10 servings", new(8.0)},
		{"123", new(123.0)},
		{"123.25", new(123.25)},
		{"this is decimal 123.25", new(123.25)},
		{"33 test", new(33.0)},
		{"this is 15 inside text", new(15.0)},
		{"use only first number 12", new(12.0)},
		{"use only 1 st number 12", new(1.0)},
		{"hello world", nil},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := FindNumber(tt.give)
			assert.Equal(t, tt.want, got)
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

func TestSplitTitle(t *testing.T) {
	tests := []struct {
		give string
		want []string
	}{
		{"", []string{}},
		{"Recipe Title | Website", []string{"Recipe Title", "Website"}},
		{"Recipe Title || Website", []string{"Recipe Title", "Website"}},
		{"Recipe Name - Blog Name", []string{"Recipe Name", "Blog Name"}},
		{"Title – Site", []string{"Title", "Site"}},
		{"Just a title", []string{"Just a title"}},
		{"Grillat kött med lime -och chilismör | Recept - Coop", []string{"Grillat kött med lime -och chilismör", "Recept", "Coop"}},
		{"Butter chicken | Oppskrift og video - Matprat", []string{"Butter chicken", "Oppskrift og video", "Matprat"}},
	}
	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			got := SplitTitle(tt.give)
			if len(tt.want) == 0 && len(got) == 0 {
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
