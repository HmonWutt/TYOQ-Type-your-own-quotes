package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/scraper"
)

func main() {
	fmt.Println("Hello, World!")

	cwd, err := os.Getwd()
	scraper.Check(err)
	p := filepath.Join(cwd, "quotes.jsonl")
	scraper.ScrapeAllPagesAndWriteToFile("https://www.goodreads.com/quotes", p)
}
