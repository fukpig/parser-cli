package parser

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestParseUrl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, `<html><head><head><body><a href="123">Hello</a><h1>Test</h1></body></html>`)
	}))
	defer ts.Close()

	var buf bytes.Buffer
	log.SetOutput(&buf)

	urls := ts.URL
	searchURL := "href"
	Parse(urls, searchURL)
	log.SetOutput(os.Stdout)
	fmt.Println("1" + buf.String() + "2")
}
