package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/html"
)

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
	fmt.Println(doc)

}
