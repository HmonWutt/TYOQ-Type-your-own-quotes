package main

import (
	"log"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/cleaner"
	_ "github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/cleaner"
)

func main() {
	err := cleaner.CleanQuotes("../data/adams.jsonl", "../data/cleanadams.jsonl")
	if err != nil {
		log.Fatal(err)
	}
}
