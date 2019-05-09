package parsercli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUrls(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config := PipelineConfig{
		Ctx:                    ctx,
		ParsingProcessesCount:  2,
		CountingProcessesCount: 2,
		Timeout:                5,
	}
	pipeline := pipelineCommon{config: config}

	urlsList := "https://google.com,https://yandex.kz"

	urlsChan := pipeline.getUrls(urlsList)

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
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config := PipelineConfig{
		Ctx:                    ctx,
		ParsingProcessesCount:  2,
		CountingProcessesCount: 2,
		Timeout:                5,
	}
	pipeline := pipelineCommon{config: config}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintln(w, html)
	}))
	defer ts.Close()

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

	htmlChan := pipeline.getHTML(htmlParams, urlsChan)

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

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	config := PipelineConfig{
		Ctx:                    ctx,
		ParsingProcessesCount:  2,
		CountingProcessesCount: 2,
		Timeout:                5,
	}
	pipeline := pipelineCommon{config: config}

	htmlChan := make(chan parserStruct)

	go func() {
		htmlChan <- parserStruct{url: "http://test.kz", html: html}
		close(htmlChan)
	}()

	occurrencesChan := pipeline.parseHTML(htmlChan, "href", maxGoroutines)

	result := <-occurrencesChan
	if result.url != "http://test.kz" {
		t.Errorf("Expected %s, found %s", "http://test.kz", result.url)
	}

	if result.count != 1 {
		t.Errorf("Expected %d, found %d", 0, result.count)
	}
}
