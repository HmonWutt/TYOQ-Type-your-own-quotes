package main

import (
	"fmt"
	"log"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/cleaner"
)

func main() {
	jsonlPaths := []string{"pratchett.jsonl", "gaiman.jsonl", "adams.jsonl"}
	for _, fromPath := range jsonlPaths {
		fmt.Println(fromPath)
		err := cleaner.CleanQuotes("../data/"+fromPath, "../data/clean.jsonl")
		if err != nil {
			log.Fatal(err)
		}
	}
}
