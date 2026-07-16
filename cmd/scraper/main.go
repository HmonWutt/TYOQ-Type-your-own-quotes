package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Scraping quotes.....")

	cwd, _ := os.Getwd()
	fmt.Println(cwd)
	// scraper.Check(err)
	// p := filepath.Join(cwd, "quotes.jsonl")
	// scraper.ScrapeAllPagesAndWriteToFile("https://www.goodreads.com/quotes", p)
}
