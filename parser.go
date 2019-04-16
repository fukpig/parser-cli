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

const MaxGoroutines = 2

func Parse(urls, searchURL string) {

	limiter := make(chan int, MaxGoroutines)

	result := make(map[string]int)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchURL)

	var wg sync.WaitGroup
	var lock = sync.RWMutex{}

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	for _, url := range urlList {
		wg.Add(1)
		go func(targetUrl string, substringRegExp *regexp.Regexp) {
			defer wg.Done()
			defer lock.Unlock()
			limiter <- 1

			resp, err := client.Get(targetUrl)
			if err != nil {
				fmt.Println(err)
				result[targetUrl] = 0
			}
			defer resp.Body.Close()
			html, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println(err)
				result[targetUrl] = 0
			}

			if err == nil {
				html := string(html)
				matchesCount := 0
				if html != "" {
					matches := substringRegExp.FindAllStringIndex(html, -1)
					matchesCount = len(matches)
				}
				lock.Lock()
				result[targetUrl] = matchesCount
			}
			<-limiter
		}(url, findSubstringRegExp)
	}
	wg.Wait()

	for key, value := range result {
		fmt.Println(key, " - ", value)
	}
}
