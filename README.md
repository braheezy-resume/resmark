# resmark

A pure Go project to convert a Resume Markdown file to HTML and PDF. No system dependencies are needed.

## Installation
Use one of the [pre-compiled binaries in the Releases](https://github.com/braheezy-resume/resume/releases).

Or install with `go`:

    go install github.com/braheezy-resume/resmark@latest

Or clone the project and run from source:

    go run main.go

## Usage

    resmark sample.md

This converts the file into `sample.html` and `sample.pdf`.

## CSS
By default, the `default.css` file is applied to the HTML before rendering. The file can be changed with the `cssFile` argument. CSS can be skipped entirely by adding the `-nocss` flag.