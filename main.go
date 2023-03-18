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

var cssFile string
var noCSS bool
var showCss bool

func init() {
	flag.StringVar(&cssFile, "cssFile", "default.css", "the css file to apply")
	flag.BoolVar(&noCSS, "nocss", false, "if set, don't apply any css")
	flag.BoolVar(&showCss, "showcss", false, "if set, print the CSS to the screen and exit")
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func writeMdToHtml(inFile string, outFile string, cssContents []byte) {
	outFile, err := filepath.Abs(outFile)
	check(err)
	inFile, err = filepath.Abs(inFile)
	check(err)

	mdData, err := os.ReadFile(inFile)
	check(err)
	mdData = markdown.NormalizeNewlines(mdData)
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	doc := p.Parse(mdData)

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	html := markdown.Render(doc, renderer)

	t, err := template.New("webpage").Parse(htmlTemplate)
	check(err)
	data := struct {
		Title      string
		Resume     template.HTML
		CSSContent template.CSS
	}{
		Title:      findTitle(mdData),
		Resume:     template.HTML(html),
		CSSContent: template.CSS(cssContents),
	}
	f, err := os.Create(outFile)
	check(err)
	defer f.Close()

	err = t.Execute(f, data)
	check(err)
}

func writeHtmlToPdf(inFile string, outFile string) {
	outFile, err := filepath.Abs(outFile)
	check(err)
	inFile, err = filepath.Abs(inFile)
	check(err)

	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// capture pdf
	var buf []byte
	if err := chromedp.Run(ctx, printToPDF(fmt.Sprintf("file://%v", inFile), &buf)); err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile(outFile, buf, 0644); err != nil {
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

	if flag.NArg() == 0 {
		flag.Usage()
		fmt.Println("\tpositional arg: <markdownFile>")
		os.Exit(1)
	}
	markdownFilename := flag.Args()[0]
	outputName := strings.TrimSuffix(markdownFilename, filepath.Ext(markdownFilename))

	htmlFilename := fmt.Sprintf("%v.html", outputName)
	pdfFilename := fmt.Sprintf("%v.pdf", outputName)

	writeMdToHtml(markdownFilename, htmlFilename, cssContents)
	writeHtmlToPdf(htmlFilename, pdfFilename)
}
