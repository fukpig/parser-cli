// Package parser was built for parsing url for search string
package parser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

const maxGoroutines = 2

type resultStruct struct {
	urlsWithCount map[string]int
	locker        sync.RWMutex
}

func readFromLimiter(limiterCh chan struct{}) {
	<-limiterCh
}

//Parse get array of urls and parse them to find occurrences of search string
func Parse(urls, searchURL string) {

	limiter := make(chan struct{}, maxGoroutines)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchURL)

	var wg sync.WaitGroup

	timeout := 5 * time.Second
	client := http.Client{
		Timeout: timeout,
	}

	result := resultStruct{urlsWithCount: make(map[string]int)}
	for _, url := range urlList {
		wg.Add(1)
		go func(targetUrl string, substringRegExp *regexp.Regexp) {
			defer wg.Done()
			defer result.locker.Unlock()
			defer readFromLimiter(limiter)
			limiter <- struct{}{}

			resp, err := client.Get(targetUrl)
			if err != nil {
				fmt.Println(err)
				result.urlsWithCount[targetUrl] = 0
			}
			defer resp.Body.Close()
			html, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				result.urlsWithCount[targetUrl] = 0
			}

			if err == nil {
				html := string(html)
				matchesCount := 0
				if html != "" {
					matches := substringRegExp.FindAllStringIndex(html, -1)
					matchesCount = len(matches)
				}
				result.locker.Lock()

				result.urlsWithCount[targetUrl] = matchesCount
			}
		}(url, findSubstringRegExp)
	}
	wg.Wait()

	for key, value := range result.urlsWithCount {
		fmt.Println(key, " - ", value)
	}
}
