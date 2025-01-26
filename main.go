package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
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
	urls := 0
	TraverseDescendants(doc, func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			attrUrl, ok := GetAttribute(n, "href")
			if ok {
				urls += 0
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

	fmt.Println("Results:")
	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "URL", "Status", "Duration"})
	t.SetOutputMirror(os.Stdout)
	t.SetAutoIndex(true)
	t.SetAllowedRowLength(90)
	t.SetStyle(table.StyleColoredBright)

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

			t.AppendRow(table.Row{"", text.WrapSoft(v.Url, 40), v.Status, v.Duration})
			t.AppendSeparator()
		case <-time.After(2 * time.Second):
			t.AppendFooter(table.Row{"", "Results: ", fmt.Sprintf("Ok URLs: %d", okCount), fmt.Sprintf("Broken URLs: %d", brokenCount)})
			t.Render()
			fmt.Println("Done")
			isBreak = true
			break
		}
	}
}
