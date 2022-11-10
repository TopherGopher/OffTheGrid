package offthegrid

import (
	"log"
	"os"
	"strings"

	"github.com/andybalholm/cascadia"
	"golang.org/x/net/html"
)

// Analyzer is a helper for analyzing HTML file contents
type Analyzer struct {
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

// FindFormFields looks at HTML strings to find form fields
func (a *Analyzer) FindFormFields(htmlBody string) {
	doc, err := html.Parse(strings.NewReader(htmlBody))
	if err != nil {
		log.Fatal(err)
	}

	forms := cascadia.MustCompile("form").MatchAll(doc)
	err = html.Render(os.Stdout, forms[0])
	if err != nil {
		log.Fatal(err)
	}
	inputs := cascadia.MustCompile("input").MatchAll(doc)
	// scripts := cascadia.MustCompile("script").MatchAll(doc)
	// buttons := cascadia.MustCompile("script").MatchAll(doc)
	html.Render(os.Stdout, inputs[0])
}
