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
	p := filepath.Join(cwd, "quotes.jsonl")
	scraper.ScrapeAllPagesAndWriteToFile("https://www.goodreads.com/quotes", p)
}
