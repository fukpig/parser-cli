package parsercli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
)

func TestScrappingWithHref(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><head><head><body><a href="123">Hello</a><h1>Test</h1></body></html>`)
	}))
	defer ts.Close()

	findSubstringRegExp := regexp.MustCompile("href")

	if result := scrapCount(ts.Client(), ts.URL, findSubstringRegExp, 2); result != 1 {
		t.Errorf("Expected count hrefs of 1, but it was %d .", result)
	}
}

func TestScrappingWithoutHref(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><head><head><body><a>Hello</a><h1>Test</h1></body></html>`)
	}))
	defer ts.Close()

	findSubstringRegExp := regexp.MustCompile("href")

	if result := scrapCount(ts.Client(), ts.URL, findSubstringRegExp, 2); result != 0 {
		t.Errorf("Expected count hrefs of 0, but it was %d .", result)
	}
}
