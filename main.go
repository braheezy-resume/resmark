package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"html/template"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
    <head>
        <meta charset="UTF-8">
        <title>{{.Title}}</title>
        <script>
{{.JSContent}}
        </script>
        <style>
{{.CSSContent}}
        </style>
    </head>
    <body>
        <div id="resume">
            {{.Resume}}
        </div>
    </body>
</html>`

//go:embed default.css
var defaultCss []byte

//go:embed default.js
var defaultJss []byte

var cssFile string
var noCSS bool
var showCss bool
var jsFile string
var noJS bool
var showJs bool

func init() {
	flag.StringVar(&cssFile, "cssFile", "default.css", "the CSS file to apply")
	flag.BoolVar(&noCSS, "nocss", false, "if set, don't apply any CSS")
	flag.BoolVar(&showCss, "showcss", false, "if set, print the final CSS to the screen and exit")
	flag.StringVar(&jsFile, "jsFile", "default.js", "the JSS file to apply")
	flag.BoolVar(&noJS, "nojs", false, "if set, don't apply any JS")
	flag.BoolVar(&showJs, "showjs", false, "if set, print the JS to the screen and exit")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writeMdToHtml(markdownFile string, htmlFile string, cssContents []byte, jsContents []byte) {
	htmlFile, err := filepath.Abs(htmlFile)
	check(err)
	markdownFile, err = filepath.Abs(markdownFile)
	check(err)

	mdData, err := os.ReadFile(markdownFile)
	check(err)
	// Force all line endings to Unix line endings
	mdData = markdown.NormalizeNewlines(mdData)
	// Enable extra Markdown features when parsing
	// CommonExtensions: Sane defaults expected in most Markdown docs
	// AutoHeadingIDs: Create the heading ID from the text
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	// Run the parser on the Markdown data, generating the AST,
	// an abstract representation of the information
	doc := p.Parse(mdData)

	// Similar to the parser, setup a renderer with certain features (flags) enabled
	// HrefTargetBlank: Add a blank target
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	// Turn the AST into HTML
	html := markdown.Render(doc, renderer)

	t, err := template.New("webpage").Parse(htmlTemplate)
	check(err)
	data := struct {
		Title      string
		Resume     template.HTML
		CSSContent template.CSS
		JSContent  template.JS
	}{
		Title:      findTitle(mdData),
		Resume:     template.HTML(html),
		CSSContent: template.CSS(cssContents),
		JSContent:  template.JS(jsContents),
	}
	f, err := os.Create(htmlFile)
	check(err)
	defer f.Close()

	err = t.Execute(f, data)
	check(err)
}

func writeHtmlToPdf(htmlFile string, pdfFile string) {
	pdfFile, err := filepath.Abs(pdfFile)
	check(err)
	htmlFile, err = filepath.Abs(htmlFile)
	check(err)

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// capture pdf
	var buf []byte
	if err := chromedp.Run(ctx, printToPDF(fmt.Sprintf("file://%v", htmlFile), &buf)); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(pdfFile, buf, 0644); err != nil {
		log.Fatal(err)
	}
}

// print a specific pdf page.
func printToPDF(url string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(url),
		chromedp.ActionFunc(func(ctx context.Context) error {
			buf, _, err := page.PrintToPDF().WithPrintBackground(false).Do(ctx)
			if err != nil {
				return err
			}
			*res = buf
			return nil
		}),
	}
}

func findTitle(markdownContents []byte) string {
	re, _ := regexp.Compile(`(?m)^#[^#]\s*(.*)`)
	matches := re.FindSubmatch(markdownContents)

	if len(matches) >= 2 {
		return string(matches[1])
	}
	return ""
}

func main() {
	flag.Parse()
	var cssContents []byte
	var err error
	if !noCSS {
		if cssFile == "default.css" {
			cssContents = defaultCss
		} else {
			cssContents, err = os.ReadFile(cssFile)
			check(err)
		}
	}
	if showCss {
		fmt.Printf("Here's the CSS that will be used:\n\n%v\n", string(cssContents))
		os.Exit(0)
	}

	var jsContents []byte
	if !noJS {
		if jsFile == "default.js" {
			jsContents = defaultJss
		} else {
			jsContents, err = os.ReadFile(jsFile)
			check(err)
		}
	}
	if showJs {
		fmt.Printf("Here's the JS that will be used:\n\n%v\n", string(jsContents))
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		flag.Usage()
		fmt.Println("\tpositional arg: <markdownFile>")
		os.Exit(1)
	}
	markdownFilename := flag.Args()[0]
	outputName := strings.TrimSuffix(markdownFilename, filepath.Ext(markdownFilename))

	htmlFilename := fmt.Sprintf("%v.html", outputName)
	pdfFilename := fmt.Sprintf("%v.pdf", outputName)

	writeMdToHtml(markdownFilename, htmlFilename, cssContents, jsContents)
	writeHtmlToPdf(htmlFilename, pdfFilename)
}
