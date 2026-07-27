package main

import (
	"log"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/cleaner"
)

func main() {
	jsonlPaths := []string{"fforde.jsonl", "gaiman.jsonl", "adams.jsonl"}
	for _, fromPath := range jsonlPaths {
		err := cleaner.CleanQuotes("../data/"+fromPath, "../data/clean.jsonl")
		if err != nil {
			log.Fatal(err)
		}
	}
}
