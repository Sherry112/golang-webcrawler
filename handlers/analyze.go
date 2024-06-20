package handlers

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/Sherry112/go-webcrawler/helpers" // Adjust the import path accordingly

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

func AnalyzeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	urlStr := r.FormValue("url")
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		renderError(w, urlStr, "Invalid URL")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(ctx, "GET", parsedURL.String(), nil)
	if err != nil {
		renderError(w, urlStr, err.Error())
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		renderError(w, urlStr, err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		renderError(w, urlStr, http.StatusText(resp.StatusCode))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		renderError(w, urlStr, "Failed to read response body")
		return
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(body))
	if err != nil {
		renderError(w, urlStr, "Failed to parse HTML")
		return
	}

	logFunc := func(message string) {
		log.Println(message)
		SSE.BroadcastMessage(message)
	}

	logFunc("Starting analysis...")

	result := analyzeDocument(ctx, doc, logFunc)
	result.URL = urlStr

	logFunc("Analysis completed.")

	tmpl, _ := template.ParseFiles(filepath.Join("templates", "result.html"))
	tmpl.Execute(w, result)
}

func renderError(w http.ResponseWriter, url, err string) {
	log.Println("Error: " + err)
	result := &helpers.Result{URL: url, Error: err}
	tmpl, _ := template.ParseFiles(filepath.Join("templates", "result.html"))
	tmpl.Execute(w, result)
}

func analyzeDocument(ctx context.Context, doc *goquery.Document, logFunc func(string)) *helpers.Result {
	result := &helpers.Result{
		Headings: make(map[int]int),
	}

	logFunc("Analyzing document...")

	result.Title = doc.Find("title").First().Text()
	logFunc("Title: " + result.Title)

	if doc.Find("!DOCTYPE").Size() > 0 {
		doctype := doc.Find("!DOCTYPE").First().Nodes[0]
		if doctype.Type == html.DoctypeNode {
			result.HTMLVersion = doctype.Data
		}
	} else {
		result.HTMLVersion = "HTML 4.01"
	}
	logFunc("HTML Version: " + result.HTMLVersion)

	for i := 1; i <= 6; i++ {
		tag := fmt.Sprintf("h%d", i)
		result.Headings[i] = doc.Find(tag).Length()
		logFunc(fmt.Sprintf("Headings h%d: %d", i, result.Headings[i]))
	}

	var wg sync.WaitGroup
	internalLinks := 0
	externalLinks := 0
	inaccessibleLinks := 0
	linkCount := doc.Find("a").Length()

	for i := 0; i < linkCount; i++ {
		link, exists := doc.Find("a").Eq(i).Attr("href")
		if !exists {
			continue
		}

		wg.Add(1)
		go func(link string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				logFunc("Found link: " + link)
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
			}
		}(link)
	}

	wg.Wait()

	result.InternalLinks = internalLinks
	result.ExternalLinks = externalLinks
	result.InaccessibleLinks = inaccessibleLinks

	result.ContainsLoginForm = doc.Find("input[type=password]").Size() > 0
	logFunc(fmt.Sprintf("Contains login form: %t", result.ContainsLoginForm))

	logFunc("Document analysis completed.")
	return result
}
