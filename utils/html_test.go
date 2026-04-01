package utils

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

const metaTestHTML = `<html>
	<head>
		<meta property="og:title" content="Test Title">
		<meta name="description" content="Test Description">
		<meta http-equiv="content-language" content="en-US">
	</head>
	<body></body>
</html>`

func getTestHead(t *testing.T) *goquery.Selection {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(metaTestHTML))
	assert.NoError(t, err)
	return doc.Find("head").First()
}

func TestGetMetaContent(t *testing.T) {
	head := getTestHead(t)

	tests := []struct {
		name          string
		key           string
		attrName      string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "property selector",
			key:           "property",
			attrName:      "og:title",
			expectedValue: "Test Title",
			expectedFound: true,
		},
		{
			name:          "name selector",
			key:           "name",
			attrName:      "description",
			expectedValue: "Test Description",
			expectedFound: true,
		},
		{
			name:          "http-equiv selector",
			key:           "http-equiv",
			attrName:      "content-language",
			expectedValue: "en-US",
			expectedFound: true,
		},
		{
			name:          "not found",
			key:           "name",
			attrName:      "not-found",
			expectedValue: "",
			expectedFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := GetMetaContent(head, tt.key, tt.attrName)
			assert.Equal(t, tt.expectedValue, val)
			assert.Equal(t, tt.expectedFound, ok)
		})
	}
}

func TestFirstMetaContent(t *testing.T) {
	head := getTestHead(t)

	tests := []struct {
		name          string
		selectors     []string
		expectedValue string
		expectedFound bool
	}{
		{
			name:          "property selector",
			selectors:     []string{"og:title"},
			expectedValue: "Test Title",
			expectedFound: true,
		},
		{
			name:          "name selector",
			selectors:     []string{"description"},
			expectedValue: "Test Description",
			expectedFound: true,
		},
		{
			name:          "not found",
			selectors:     []string{"not-found"},
			expectedValue: "",
			expectedFound: false,
		},
		{
			name:          "multiple selectors found first",
			selectors:     []string{"og:title", "description"},
			expectedValue: "Test Title",
			expectedFound: true,
		},
		{
			name:          "multiple selectors found second",
			selectors:     []string{"not-found", "description"},
			expectedValue: "Test Description",
			expectedFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := FirstMetaContent(head, tt.selectors...)
			assert.Equal(t, tt.expectedValue, val)
			assert.Equal(t, tt.expectedFound, ok)
		})
	}
}
