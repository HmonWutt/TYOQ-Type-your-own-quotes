package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/internal/scraper"
)

func main() {
	fmt.Println("Scraping quotes.....")

	cwd, err := os.Getwd()
	fmt.Println(cwd)
	scraper := scraper.Scraper{
		OutputFile: filepath.Join(cwd, "gaiman.jsonl"),
		StartIndex: 1,
		Offset:     6,
		BaseURL:    "https://www.goodreads.com/quotes/search",
		Author:     "neil gaiman",
		Referer:    fmt.Sprintf("https://www.goodreads.com/quotes"),
	}
	fmt.Println(scraper.Referer)
	err = scraper.ScrapeAndAppend()
	if err != nil {
		log.Fatal(err)
	}
}
