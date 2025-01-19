package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/html"
)

type Link struct {
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

		if attr.Key == name && attr.Val != "#" && attr.Val != "javascript:void(0)" && attr.Val[:1] != "#" {
			return attr.Val, true
		}
	}

	return "", false
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

	var linkChan = make(chan Link, 8)
	TraverseDescendants(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			attrUrl, ok := GetAttribute(n, "href")
			if ok {
				go func(path string, u *url.URL) {
					link := path
					if path[:1] == "/" {
						link = u.Scheme + "://" + u.Host + path
					}

					t := time.Now()
					fmt.Println("Fetching: ", link)
					res, err := http.Get(link)
					fmt.Println("")

					if err != nil {
						panic(err)
					}

					linkChan <- Link{Url: link, Status: res.StatusCode, Duration: time.Since(t)}

				}(attrUrl, parsedUrl)
			}
		}
	})

	fmt.Println("==================")
	fmt.Println("==================")
	fmt.Println("")

	var isBreak bool

	fmt.Printf("Results:")
	for {
		if isBreak {
			close(linkChan)
			break
		}

		select {
		case v := <-linkChan:
			fmt.Printf("URL: %s, status: %d duration: %v\n", v.Url, v.Status, v.Duration)
		case <-time.After(2 * time.Second):
			fmt.Println("================")
			fmt.Println("Done")
			isBreak = true
			break
		}
	}

}
