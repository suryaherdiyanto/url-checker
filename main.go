package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/html"
)

type LinkResult struct {
	Url      string
	Status   int
	Duration time.Duration
}

func TraverseDescendants(n *html.Node, fn func(*html.Node)) {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		fn(c)
		if n.Type != html.ErrorNode {
			TraverseDescendants(c, fn)
		}
	}
}

func GetAttribute(n *html.Node, name string) (string, bool) {
	for _, attr := range n.Attr {

		if attr.Key == "role" && attr.Val == "button" {
			continue
		}

		if attr.Key == name && len(attr.Val) < 4 {
			return "", false
		}

		if attr.Key == name && (attr.Val[:1] == "/" || attr.Val[:4] == "http" || attr.Val[:5] == "https") {
			return attr.Val, true
		}
	}

	return "", false
}

func fetchInternalLinkWorker(path string, url *url.URL, c chan<- LinkResult) {
	link := path
	var result LinkResult

	if link[:1] == "/" {
		link = url.Scheme + "://" + url.Host + path
	}

	t := time.Now()
	fmt.Println("Fetching: ", link)
	res, err := http.Get(link)
	fmt.Println("")

	if err != nil {
		fmt.Println(err)
		result = LinkResult{Url: link, Status: 0, Duration: time.Since(t)}
	}

	if err == nil {
		result = LinkResult{Url: link, Status: res.StatusCode, Duration: time.Since(t)}
	}

	c <- result
}

var domain string

func main() {
	inputUrl := os.Args[1]

	if inputUrl == "" {
		fmt.Println("URL not provided")
		os.Exit(1)
	}

	if len(inputUrl) <= 8 {
		fmt.Println("Might not a valid URL")
		os.Exit(1)
	}

	if inputUrl[:8] != "https://" && inputUrl[:7] != "http://" {
		fmt.Println("URL is invalid")
		os.Exit(1)
	}

	fmt.Println("Fetching: " + inputUrl)
	fmt.Println("")
	parsedUrl, _ := url.Parse(inputUrl)

	res, err := http.Get(inputUrl)
	if err != nil {
		panic(err)
	}
	fmt.Println(res.Status)

	doc, err := html.Parse(res.Body)
	if err != nil {
		panic(err)
	}

	var linkChan = make(chan LinkResult, 8)
	TraverseDescendants(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			attrUrl, ok := GetAttribute(n, "href")
			if ok {
				go fetchInternalLinkWorker(attrUrl, parsedUrl, linkChan)
			}
		}
	})

	fmt.Println("==================")
	fmt.Println("==================")
	fmt.Println("")

	var isBreak bool
	var brokenCount int
	var okCount int

	fmt.Printf("Results:")
	for {
		if isBreak {
			close(linkChan)
			break
		}

		select {
		case v := <-linkChan:
			if v.Status == http.StatusOK || v.Status == http.StatusPermanentRedirect {
				okCount += 1
			}
			if v.Status > 400 || v.Status > 500 {
				brokenCount += 1
			}

			fmt.Printf("URL: %s, status: %d duration: %v\n", v.Url, v.Status, v.Duration)
		case <-time.After(2 * time.Second):
			fmt.Println("================")
			fmt.Println("Done")
			fmt.Println("Ok URLs: ", okCount)
			fmt.Println("Broken URLs: ", brokenCount)
			isBreak = true
			break
		}
	}

}
