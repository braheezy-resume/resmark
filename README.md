# resumerk

A pure Go implementation to convert a Resume Markdown file to HTML and PDF.

This project requires no system dependencies.

## Installation
Use one of the [pre-compiled binaries in the Releases](https://github.com/braheezy-resume/resume/releases).

Or install with `go`:

    go install github.com/braheezy-resume/resumerk@latest

Or clone the project and run from source:

    go run main.go

## Usage

    resumerk sample.md

This converts the file into `sample.html` and `sample.pdf`. The output files are named according to the name of the input file.

## CSS and JS
CSS and Javascript can be injected into the generated HTML or completely ignored. Use the `-help` flag to see the options.

## Special Thanks
This project wouldn't exist without this guy's project. Particularly, the CSS that makes the resume actually look like a resume wouldn't exist.

https://github.com/mikepqr/resume.md
