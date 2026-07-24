package seed

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/scraper"
	_ "modernc.org/sqlite"
)

type Row struct {
	text string
	tags string
}
type Result struct {
	Id        int
	Text      string
	Author    string
	Source    string
	Tags      string
	WordCount int
}

const dbPath = "../data/seed.db"

var jsonlPaths = []string{"fforde.jsonl", "gaiman.jsonl", "adams.jsonl"}

func Insert(db *sql.DB, q *scraper.Quote) (int64, error) {
	sql := "INSERT INTO quotes (text, author, source, word_count, tags) VALUES (?, ?, ?, ? ,?);"
	tagsJSON, _ := json.Marshal(q)
	result, err := db.Exec(sql, q.Text, q.Author, q.Source, len(strings.Fields(q.Text)), string(tagsJSON))
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, nil
	}
	fmt.Printf(
		"The quote was inserted with ID:%d\n",
		id,
	)
	return id, nil
}

func Seed() {
	// connect to the SQLite database
	db, err := sql.Open("sqlite", dbPath)
	defer db.Close()
	exitIfError(err)
	// Make sure it works.
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}
	setupSchema(db)
	quotes, err := ReadFromFile("../data/fforde.jsonl")
	if err != nil {
		exitIfError(err)
	}
	batchInsert(db, quotes)
}

func selectQuote() {
	// insert the quote
	// id, err := Insert(db, quote)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// // print the inserted country

	// // var quotes []Row
	// rows := db.QueryRow("select text,tags from quotes where id=1")
	//
	// // defer rows.Close()
	// // for rows.Next() {
	// q := &Row{}
	// err = rows.Scan(&q.text, &q.tags)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	//
	// fmt.Println(quote.Tags)
	// //quotes = append(quotes, *q)
	// //}
}

func batchInsert(db *sql.DB, quotes []scraper.Quote) {
	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	defer tx.Rollback()
	stmt, _ := tx.Prepare(`
        INSERT INTO quotes (text, author, source, word_count,tags)
        VALUES (?, ?, ?, ?, ?)
    `)
	defer stmt.Close()

	for _, q := range quotes {
		tagsJSON, _ := json.Marshal(q.Tags)
		if _, err := stmt.Exec(q.Text, q.Author, q.Source, string(tagsJSON), len(strings.Fields(q.Text))); err != nil {
			exitIfError(err)
		}
	}

	if err = tx.Commit(); err != nil {
		exitIfError(err)
	}
}

func ReadFromFile(filename string) ([]scraper.Quote, error) {
	var quotes []scraper.Quote
	file, err := os.Open(filename)
	if err != nil {
		return []scraper.Quote{}, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var quote scraper.Quote
		err := json.Unmarshal([]byte(line), &quote)
		if err != nil {
			return []scraper.Quote{}, fmt.Errorf("failed to parse data: %v", err)
		}
		quotes = append(quotes, quote)
	}
	return quotes, nil
}

func exitIfError(err error) {
	if err != nil {
		log.Output(2, err.Error())
		os.Exit(1)
	}
}

func setupSchema(db *sql.DB) {
	sql := `CREATE TABLE IF NOT EXISTS quotes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    text       TEXT NOT NULL,
    author     TEXT DEFAULT 'Unknown',
    source     TEXT DEFAULT 'Unknown',
    tags       TEXT DEFAULT '[]',
    word_count INTEGER,
    created_at TEXT DEFAULT (datetime('now'))
	);`
	_, err := db.Exec(sql)
	exitIfError(err)
}
