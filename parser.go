// Package parsercli was built for parsing url for search string
package parsercli

import (
	"context"
	"errors"
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

func getUrls(ctx context.Context, wg *sync.WaitGroup, urls string) chan string {

	urlList := strings.Split(urls, ",")
	urlsChan := make(chan string, len(urlList))

	wg.Add(1)
	go func(ctx context.Context, urlsChan chan string, urlList []string) {
		defer wg.Done()
		defer close(urlsChan)

		for _, url := range urlList {
			fmt.Println("GET", url)
			urlsChan <- url
		}
	}(ctx, urlsChan, urlList)

	return urlsChan
}

func getHTML(ctx context.Context, wg *sync.WaitGroup, urlsChan chan string, timeout int) chan parserStruct {
	htmlChan := make(chan parserStruct)
	client := &http.Client{}

	go func(ctx context.Context, urlsChan chan string) {
		for url := range urlsChan {
			wg.Add(1)
			go func(ctx context.Context, wg *sync.WaitGroup, client *http.Client, timeout int, url string) {
				defer wg.Done()
				html := scrap(ctx, client, timeout, url)
				htmlChan <- parserStruct{url: url, html: html}
				fmt.Println("SCRAP", url)

			}(ctx, wg, client, timeout, url)
		}
		defer close(htmlChan)
	}(ctx, urlsChan)
	//defer close(htmlChan)

	return htmlChan
}

func parseHTML(ctx context.Context, wg *sync.WaitGroup, htmlChan chan parserStruct, searchString string) chan resultStruct {
	occurrencesChan := make(chan resultStruct)
	findSubstringRegExp := regexp.MustCompile(searchString)

	for parserInfo := range htmlChan {
		//wg.Add(1)
		fmt.Println("PARSE", parserInfo.url)
		count := countMatches(parserInfo.html, findSubstringRegExp)
		occurrencesChan <- resultStruct{url: parserInfo.url, count: count}
		/*go func(ctx context.Context, wg *sync.WaitGroup, parserInfo parserStruct, findSubstringRegExp *regexp.Regexp) {
			defer wg.Done()
			count := countMatches(parserInfo.html, findSubstringRegExp)
			occurrencesChan <- resultStruct{url: parserInfo.url, count: count}
		}(ctx, wg, parserInfo, findSubstringRegExp)*/
	}
	close(occurrencesChan)

	return occurrencesChan
}

//render generate output Url - count
func render(ctx context.Context, wg *sync.WaitGroup, occurrencesChan chan resultStruct) {
	fmt.Println("render")
	for resultStruct := range occurrencesChan {
		fmt.Println("123")
		fmt.Println(resultStruct.url, "-", resultStruct.count)
	}
}

func shutdown(ctx context.Context, wg *sync.WaitGroup, htmlChan chan parserStruct, occurrencesChan chan resultStruct) error {
	ch := make(chan struct{})

	go func() {
		wg.Wait()
		close(htmlChan)
		close(occurrencesChan)
		close(ch)
	}()

	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return errors.New("timeout")
	}
}

//Parse get array of urls and parse them to find occurrences of search string
func Parse(ctx context.Context, urls, searchString string, maxGoroutines, timeout int) {
	var wg sync.WaitGroup

	urlsChan := getUrls(ctx, &wg, urls)
	htmlChan := getHTML(ctx, &wg, urlsChan, timeout)
	occurrencesChan := parseHTML(ctx, &wg, htmlChan, searchString)
	render(ctx, &wg, occurrencesChan)

	shutdown(ctx, &wg, htmlChan, occurrencesChan)

}
