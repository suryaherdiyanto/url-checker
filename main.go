package main

import (
	"fmt"
	"os"
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

	fmt.Println("Hello world")
}
