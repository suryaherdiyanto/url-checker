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

	fmt.Println("Hello world")
}
