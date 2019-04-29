package parsercli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

const html = "<html><head><head><body><a href=\"123\">Hello</a><h1>Test</h1></body></html>"

func TestScrap(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, html)
	}))
	defer ts.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	timeout := 2

	if result := scrap(ctx, ts.Client(), timeout, ts.URL); result != html {
		t.Errorf("Expected another html, but it was %s .", result)
	}
}

func TestCountMatches(t *testing.T) {
	findSubstringRegExp := regexp.MustCompile("href")

	if resultWithOneMatch := countMatches(html, findSubstringRegExp); resultWithOneMatch != 1 {
		t.Errorf("Expected count hrefs of 1, but it was %d .", resultWithOneMatch)
	}

	htmlWithoutHref := "<html><head><head><body><a>Hello</a><h1>Test</h1></body></html>"
	if resultWithZeroMatch := countMatches(htmlWithoutHref, findSubstringRegExp); resultWithZeroMatch != 0 {
		t.Errorf("Expected count hrefs of 0, but it was %d .", resultWithZeroMatch)
	}
}

func TestGetUrls(t *testing.T) {
	urlsList := "https://google.com,https://yandex.kz"

	urlsChan := getUrls(urlsList)

	firstURL := <-urlsChan
	if firstURL != "https://google.com" {
		t.Errorf("Expected %s, found %s", "https://google.com", firstURL)
	}
	secondURL := <-urlsChan
	if secondURL != "https://yandex.kz" {
		t.Errorf("Expected %s, found %s", "https://yandex.kz", secondURL)
	}
}

func TestGetHTML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, html)
	}))
	defer ts.Close()

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	timeout := 2

	urlsChan := make(chan string)

	go func() {
		urlsChan <- ts.URL
		close(urlsChan)
	}()

	htmlParams := htmlParams{
		ctx:           ctx,
		timeout:       timeout,
		maxGoroutines: 2,
	}

	htmlChan := getHTML(htmlParams, urlsChan)

	result := <-htmlChan
	if result.url != ts.URL {
		t.Errorf("Expected %s, found %s", ts.URL, result.url)
	}

	if result.html != html {
		t.Errorf("Expected %s, found %s", html, result.html)
	}
}

func TestParseHTML(t *testing.T) {
	maxGoroutines := 2

	htmlChan := make(chan parserStruct)

	go func() {
		htmlChan <- parserStruct{url: "http://test.kz", html: html}
		close(htmlChan)
	}()

	occurrencesChan := parseHTML(htmlChan, "href", maxGoroutines)

	result := <-occurrencesChan
	if result.url != "http://test.kz" {
		t.Errorf("Expected %s, found %s", "http://test.kz", result.url)
	}

	if result.count != 1 {
		t.Errorf("Expected %d, found %d", 0, result.count)
	}
}
