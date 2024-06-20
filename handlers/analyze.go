package handlers

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/Sherry112/go-webcrawler/helpers"

	"github.com/PuerkitoBio/goquery"
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

	result := helpers.AnalyzeDocument(doc, logFunc)
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
