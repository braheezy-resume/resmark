package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const mdStr = `
# header

Sample text.

[link](http://example.com)
`

func TestFindTitle(t *testing.T) {
	require.Equal(t, "header", findTitle([]byte(mdStr)))
}

func TestWriteMdToHtml(t *testing.T) {
	err := os.WriteFile("test.md", []byte(mdStr), 0644)
	check(err)

	writeMdToHtml("test.md", "test.html", []byte{})

	expectedHtmlStr := `
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>header</title>
        <style>

        </style>
    </head>
    <body>
        <div id="resume">
            <h1 id="header">header</h1>

<p>Sample text.</p>

<p><a href="http://example.com" target="_blank">link</a></p>

        </div>
    </body>
</html>`

	actualHtml, _ := os.ReadFile("test.html")
	// Verify the strings are equal sans whitespace
	require.Equal(t,
		strings.ReplaceAll(expectedHtmlStr, " ", ""),
		strings.ReplaceAll(string(actualHtml), " ", ""),
	)

	os.Remove("test.html")
	os.Remove("test.md")
}
