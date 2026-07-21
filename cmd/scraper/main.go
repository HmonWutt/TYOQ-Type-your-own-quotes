package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/internal/scraper"
)

func main() {
	fmt.Println("Scraping quotes.....")

	cwd, err := os.Getwd()
	fmt.Println(cwd)
	scraper.Check(err)
	p := filepath.Join(cwd, "gaiman.jsonl")
	startIndex := 1
	offset := 11
	scraper.ScrapeAndAppend("Neil Gaiman", "https://www.goodreads.com/quotes/search", p, startIndex, offset)
}
