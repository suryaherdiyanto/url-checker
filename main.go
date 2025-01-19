package main

import (
	"fmt"
	"net/http"
	"os"

	"net/url"

	"golang.org/x/net/html"
)

type Link struct {
	Url     string
	Content string
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
	parsedUrl, _ := url.Parse(inputUrl)
	domain := parsedUrl.Host
	fmt.Println(domain)

	req, err := http.NewRequest(http.MethodGet, inputUrl, nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	fmt.Println(res.Status)

	doc, err := html.Parse(res.Body)
	if err != nil {
		panic(err)
	}

	TraverseDescendants(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			attr, _ := GetAttribute(n, "href")
			fmt.Println(attr)
		}
	})
}
