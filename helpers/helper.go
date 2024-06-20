package helpers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// Result holds the analysis results of a web page.
type Result struct {
	URL               string
	HTMLVersion       string
	Title             string
	Headings          map[int]int
	InternalLinks     int
	ExternalLinks     int
	InaccessibleLinks int
	ContainsLoginForm bool
	Error             string
}

// AnalyzeDocument performs the actual analysis of the HTML document.
func AnalyzeDocument(doc *goquery.Document, logFunc func(string)) *Result {
	result := &Result{
		Headings: make(map[int]int),
	}

	logFunc("Analyzing document...")

	// Get the title of the page.
	result.Title = doc.Find("title").First().Text()
	logFunc("Title: " + result.Title)

	// Determine HTML version.
	if doc.Find("!DOCTYPE").Size() > 0 {
		doctype := doc.Find("!DOCTYPE").First().Nodes[0]
		if doctype.Type == html.DoctypeNode {
			result.HTMLVersion = doctype.Data
		}
	} else {
		result.HTMLVersion = "HTML 4.01"
	}
	logFunc("HTML Version: " + result.HTMLVersion)

	// Count the number of headings at each level.
	for i := 1; i <= 6; i++ {
		tag := fmt.Sprintf("h%d", i) // Construct the heading tag name.
		result.Headings[i] = doc.Find(tag).Length()
		logFunc(fmt.Sprintf("Headings h%d: %d", i, result.Headings[i]))
	}

	// Analyze the links.
	internalLinks := 0
	externalLinks := 0
	inaccessibleLinks := 0
	linkChan := make(chan int, 10)

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, exists := s.Attr("href")
		if !exists {
			return
		}

		logFunc("Found link: " + link)

		go func(link string) {
			defer func() { linkChan <- 1 }()
			u, err := url.Parse(link)
			if err != nil || !u.IsAbs() {
				internalLinks++
				logFunc("Internal link: " + link)
				return
			}

			externalLinks++
			logFunc("External link: " + link)
			resp, err := http.Head(link)
			if err != nil || resp.StatusCode != http.StatusOK {
				inaccessibleLinks++
				logFunc("Inaccessible link: " + link)
			}
		}(link)
	})

	for i := 0; i < doc.Find("a").Length(); i++ {
		<-linkChan
	}

	result.InternalLinks = internalLinks
	result.ExternalLinks = externalLinks
	result.InaccessibleLinks = inaccessibleLinks

	// Check if the page contains a login form.
	result.ContainsLoginForm = doc.Find("input[type=password]").Size() > 0
	logFunc(fmt.Sprintf("Contains login form: %t", result.ContainsLoginForm))

	logFunc("Document analysis completed.")
	return result
}
