package parsercli

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//scrap function make request to url and get html from answer
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

//countMatches count of entrance of substring in html
func countMatches(html string, substringRegExp *regexp.Regexp) int {
	matches := substringRegExp.FindAllStringIndex(html, -1)
	matchesCount := len(matches)
	return matchesCount
}
