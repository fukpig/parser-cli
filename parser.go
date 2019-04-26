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
	return html
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

func getHTML(ctx context.Context, urlsChan chan string, timeout, maxGoroutines int) chan parserStruct {

	htmlChan := make(chan parserStruct)
	client := &http.Client{}

	go func(ctx context.Context, client *http.Client, timeout, maxGoroutines int, urlsChan chan string) {
		var wg sync.WaitGroup
		limiter := make(chan struct{}, maxGoroutines)

		for url := range urlsChan {
			wg.Add(1)
			go func(ctx context.Context, client *http.Client, timeout int, url string) {
				defer wg.Done()
				defer func(ch <-chan struct{}) {
					<-ch
				}(limiter)
				limiter <- struct{}{}
				html := scrap(ctx, client, timeout, url)
				htmlChan <- parserStruct{url: url, html: html}

			}(ctx, client, timeout, url)
		}

		wg.Wait()
		close(htmlChan)

	}(ctx, client, timeout, maxGoroutines, urlsChan)
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
func Parse(ctx context.Context, wg *sync.WaitGroup, urls, searchString string, maxGoroutines, timeout int) {
	defer wg.Done()
	urlsChan := getUrls(urls)
	htmlChan := getHTML(ctx, urlsChan, timeout, maxGoroutines)
	occurrencesChan := parseHTML(htmlChan, searchString, maxGoroutines)
	render(occurrencesChan)
}
