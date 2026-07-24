package main

import (
	"database/sql"
	"fmt"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/internal/scraper"
)

const dbPath = "init.db"

var jsonlPath = []string{"fforde.jsonl", "gaiman.jsonl", "adams.jsonl"}

func Insert(db *sql.DB, c *scraper.Quote) (int64, error) {
	sql := `INSERT INTO quotes (name, population, area) 
            VALUES (?, ?, ?);`
	result, err := db.Exec(sql, c.Name, c.Population, c.Area)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func main() {
	// connect to the SQLite database
	db, err := sql.Open("sqlite", "./init.db")
	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

	// create a new quote
	quote := &scraper.Quote{}

	// insert the country
	countryId, err := Insert(db, country)
	if err != nil {
		fmt.Println(err)
		return
	}

	// print the inserted country
	fmt.Printf(
		"The country %s was inserted with ID:%d\n",
		country.Name,
		countryId,
	)
}
