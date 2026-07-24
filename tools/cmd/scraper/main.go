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
		OutputFile: filepath.Join(cwd, "holt.jsonl"),
		StartIndex: 2,
		Offset:     21,
		BaseURL:    "https://www.goodreads.com/quotes/search",
		Author:     "Tom Holt",
		Referer:    "https://www.goodreads.com/quotes",
	}
	fmt.Println(scraper.Referer)
	err = scraper.ScrapeAndAppend()
	if err != nil {
		log.Fatal(err)
	}
}
