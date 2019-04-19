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
	urlsWithCount map[string]int
	mx            sync.RWMutex
}

//addCountToUrl add count of occurrences to targetURL
func (r *resultStruct) addCountToURL(targetURL string, count int) {
	r.mx.Lock()
	r.urlsWithCount[targetURL] = count
	r.mx.Unlock()
}

//render generate output Url - count
func render(urlsResult map[string]int) {
	for key, value := range urlsResult {
		fmt.Println(key, " - ", value)
	}
}

//scrapCount get html by url and find by regexp count matches
func scrapCount(
	ctx context.Context,
	client *http.Client,
	targetURL string,
	substringRegExp *regexp.Regexp,
	timeout int,
) int {
	ctxTimeout, timeoutCancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer timeoutCancel()

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	resp, err := client.Do(req.WithContext(ctxTimeout))
	if err != nil {
		fmt.Println(err)
		return 0
	}
	defer resp.Body.Close()
	htmlBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return 0
	}

	html := string(htmlBytes)
	matchesCount := 0
	if html != "" {
		matches := substringRegExp.FindAllStringIndex(html, -1)
		matchesCount = len(matches)
	}
	return matchesCount
}

//Parse get array of urls and parse them to find occurrences of search string
func Parse(ctx context.Context, urls, searchURL string, maxGoroutines, timeout int) {
	limiter := make(chan struct{}, maxGoroutines)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchURL)

	var wg sync.WaitGroup

	client := &http.Client{}

	result := resultStruct{urlsWithCount: make(map[string]int, len(urlList))}
	for _, url := range urlList {
		wg.Add(1)
		go func(ctx context.Context, targetURL string, substringRegExp *regexp.Regexp, timeout int) {
			defer wg.Done()
			defer func(ch <-chan struct{}) {
				<-ch
			}(limiter)
			limiter <- struct{}{}

			count := scrapCount(ctx, client, targetURL, substringRegExp, timeout)

			result.addCountToURL(targetURL, count)
		}(ctx, url, findSubstringRegExp, timeout)
	}
	wg.Wait()

	render(result.urlsWithCount)
}
