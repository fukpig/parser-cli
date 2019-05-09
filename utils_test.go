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
