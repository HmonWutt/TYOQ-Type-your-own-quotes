package main

import (
	"fmt"
	"log"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/internal/genseed"
)

func main() {
	quotes, err := genseed.ReadFromFile("quotes.jsonl")
	if err != nil {
		log.Fatal(err)
	}

	err = genseed.GenSeed(quotes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wrote to sql successfully.")
}
