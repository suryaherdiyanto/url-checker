package main

import (
	"fmt"
	"net/http"
	"os"

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

func main() {
	url := os.Args[1]

	if url == "" {
		fmt.Println("URL not provided")
		os.Exit(1)
	}

	if len(url) <= 8 {
		fmt.Println("Might not a valid URL")
		os.Exit(1)
	}

	if url[:8] != "https://" && url[:7] != "http://" {
		fmt.Println("URL is invalid")
		os.Exit(1)
	}

	fmt.Println("Fetching: " + url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
			url, _ = GetAttribute(n, "href")
			fmt.Println(url)
		}
	})
}
