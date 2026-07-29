package main

import (
	"log"

	db "github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/db"
)

func main() {
	if err := db.Seed(); err != nil {
		log.Fatal(err)
	}
}
