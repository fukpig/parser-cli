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

func Parse(urls, searchUrl string) {

	result := make(map[string]int)

	urlList := strings.Split(urls, ",")

	findSubstringRegExp := regexp.MustCompile(searchUrl)

	limiter := time.Tick(time.Second * 1)

	var wg sync.WaitGroup

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	for _, url := range urlList {
		wg.Add(1)

		go func(targetUrl string, substringRegExp *regexp.Regexp) {
			<-limiter

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
					matches := findSubstringRegExp.FindAllStringIndex(html, -1)
					matchesCount = len(matches)
				}
				result[targetUrl] = matchesCount
				wg.Done()
			}
		}(url, findSubstringRegExp)
		wg.Wait()
	}

	for key, value := range result {
		fmt.Println(key, " - ", value)
	}
}
