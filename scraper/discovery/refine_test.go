package discovery

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borschtapp/krip/model"
)

func TestRefineByUrlPattern_FindsPattern(t *testing.T) {
	all := []*model.Recipe{
		{Url: "https://example.com/recipes/pasta"},
		{Url: "https://example.com/recipes/chicken"},
		{Url: "https://example.com/blog/news"},
		{Url: "https://example.com/about"},
	}
	valid := []*model.Recipe{
		{Url: "https://example.com/recipes/pasta"},
		{Url: "https://example.com/recipes/chicken"},
	}

	pattern, matched := RefineByUrlPattern(all, valid)

	assert.NotEmpty(t, pattern)
	assert.Contains(t, pattern, "recipes")
	assert.Equal(t, 2, matched, "only recipe-path entries should match the pattern")
}

func TestRefineByUrlPattern_NoPattern(t *testing.T) {
	// Two valid entries with completely different top-level path segments → no common prefix.
	all := []*model.Recipe{
		{Url: "https://example.com/about"},
		{Url: "https://example.com/contact"},
	}
	valid := []*model.Recipe{
		{Url: "https://example.com/about"},
		{Url: "https://example.com/contact"},
	}

	pattern, matched := RefineByUrlPattern(all, valid)

	assert.Empty(t, pattern, "divergent top-level paths should yield no common prefix")
	assert.Zero(t, matched)
}

func TestRefineByUrlPattern_EmptyValid(t *testing.T) {
	all := []*model.Recipe{{Url: "https://example.com/recipes/pasta"}}

	pattern, matched := RefineByUrlPattern(all, nil)

	assert.Empty(t, pattern)
	assert.Zero(t, matched)
}
