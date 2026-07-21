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
	p := filepath.Join(cwd, "pratchett.jsonl")
	startIndex := 1
	offset := 101
	scraper.ScrapeAndAppend("Terry Pratchett", "https://www.goodreads.com/quotes/search", p, startIndex, offset)
}
