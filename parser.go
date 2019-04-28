// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

type htmlParams struct {
	ctx           context.Context
	client        *http.Client
	timeout       int
	maxGoroutines int
}

type resultStruct struct {
	url   string
	count int
}

type parserStruct struct {
	url  string
	html string
}

func scrap(ctx context.Context, client *http.Client, timeout int, targetURL string) string {
	ctxTimeout, timeoutCancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer timeoutCancel()

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	resp, err := client.Do(req.WithContext(ctxTimeout))
	if err != nil {
		fmt.Println(err)
		return ""
	}
	defer resp.Body.Close()
	htmlBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	html := string(htmlBytes)
	return strings.TrimSpace(html)
}

func countMatches(html string, substringRegExp *regexp.Regexp) int {
	matches := substringRegExp.FindAllStringIndex(html, -1)
	matchesCount := len(matches)
	return matchesCount
}

func getUrls(urls string) chan string {
	urlList := strings.Split(urls, ",")
	urlsChan := make(chan string, len(urlList))

	go func(urlsChan chan string, urlList []string) {
		defer close(urlsChan)

		for _, url := range urlList {
			urlsChan <- url
		}
	}(urlsChan, urlList)
	return urlsChan
}

func getHTML(params htmlParams, urlsChan chan string) chan parserStruct {

	htmlChan := make(chan parserStruct)
	client := &http.Client{}
	params.client = client

	go func(params htmlParams) {
		var wg sync.WaitGroup
		limiter := make(chan struct{}, params.maxGoroutines)

		for url := range urlsChan {
			wg.Add(1)
			go func(params htmlParams, limiter chan struct{}, url string) {
				defer wg.Done()
				defer func(ch <-chan struct{}) {
					<-ch
				}(limiter)
				limiter <- struct{}{}
				html := scrap(params.ctx, params.client, params.timeout, url)
				htmlChan <- parserStruct{url: url, html: html}

			}(params, limiter, url)
		}

		wg.Wait()
		close(htmlChan)

	}(params)
	return htmlChan
}

func parseHTML(htmlChan chan parserStruct,
	searchString string,
	maxGoroutines int) chan resultStruct {
	occurrencesChan := make(chan resultStruct)
	findSubstringRegExp := regexp.MustCompile(searchString)

	limiter := make(chan struct{}, maxGoroutines)
	go func() {
		var wg sync.WaitGroup
		for parserInfo := range htmlChan {
			wg.Add(1)
			go func(info parserStruct) {
				defer func(ch <-chan struct{}) {
					<-ch
				}(limiter)
				limiter <- struct{}{}
				count := countMatches(info.html, findSubstringRegExp)
				occurrencesChan <- resultStruct{url: info.url, count: count}
				wg.Done()
			}(parserInfo)
		}
		wg.Wait()
		close(occurrencesChan)
	}()

	return occurrencesChan
}

//render generate output Url - count
func render(occurrencesChan chan resultStruct) {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for resultStruct := range occurrencesChan {
			fmt.Println(resultStruct.url, "-", resultStruct.count)
		}
	}()
	wg.Wait()
}

//Parse get array of urls and parse them to find occurrences of search string
func Parse(ctx context.Context, wg *sync.WaitGroup, urls, searchString string, parsingProcessesCount, countingProcessesCount, timeout int) {

	htmlParams := htmlParams{
		ctx:           ctx,
		timeout:       timeout,
		maxGoroutines: parsingProcessesCount,
	}

	defer wg.Done()
	urlsChan := getUrls(urls)
	htmlChan := getHTML(htmlParams, urlsChan)
	occurrencesChan := parseHTML(htmlChan, searchString, countingProcessesCount)
	render(occurrencesChan)
}

func getUrlsPreload(urlList []string) chan string {
	urlsChan := make(chan string, len(urlList))

	go func(urlsChan chan string, urlList []string) {
		defer close(urlsChan)

		for _, url := range urlList {
			urlsChan <- url
		}
	}(urlsChan, urlList)
	return urlsChan
}

func getHTMLPreload(params htmlParams, urlsChan chan string, urlsCount int) chan parserStruct {

	htmlChan := make(chan parserStruct)
	client := &http.Client{}
	params.client = client
	go func() {
		var wg sync.WaitGroup
		for i := 1; i <= urlsCount; i++ {
			wg.Add(1)
			go func(params htmlParams) {
				defer wg.Done()
				url := <-urlsChan
				html := scrap(params.ctx, params.client, params.timeout, url)
				htmlChan <- parserStruct{url: url, html: html}
			}(params)
		}
		wg.Wait()
		close(htmlChan)
	}()
	return htmlChan
}

func parseHTMLPreload(htmlChan chan parserStruct, searchString string, urlsCount int) chan resultStruct {
	occurrencesChan := make(chan resultStruct)
	findSubstringRegExp := regexp.MustCompile(searchString)

	go func() {
		var wg sync.WaitGroup
		for i := 1; i <= urlsCount; i++ {
			wg.Add(1)
			go func(htmlChan chan parserStruct) {
				defer wg.Done()
				info := <-htmlChan
				count := countMatches(info.html, findSubstringRegExp)
				occurrencesChan <- resultStruct{url: info.url, count: count}
			}(htmlChan)
		}
		wg.Wait()
		close(occurrencesChan)
	}()

	return occurrencesChan
}

//PipelineWithPreloadGoroutines pipeline with preload goroutines
func PipelineWithPreloadGoroutines(ctx context.Context, wg *sync.WaitGroup, urls, searchString string, parsingProcessesCount, countingProcessesCount, timeout int) {
	htmlParams := htmlParams{
		ctx:           ctx,
		timeout:       timeout,
		maxGoroutines: parsingProcessesCount,
	}

	defer wg.Done()

	urlList := strings.Split(urls, ",")
	urlsCount := len(urlList)

	urlsChan := getUrlsPreload(urlList)
	htmlChan := getHTMLPreload(htmlParams, urlsChan, urlsCount)
	occurrencesChan := parseHTMLPreload(htmlChan, searchString, urlsCount)
	render(occurrencesChan)
}
