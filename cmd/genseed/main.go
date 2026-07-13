package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/scraper"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	_ "github.com/lib/pq" // To register the driver.
)

func main() {
	quotes, err := readFromFile("quotes.jsonl")
	if err != nil {
		log.Fatal(err)
	}

	err = genSeed(quotes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Wrote to sql successfully.")
}

func trim(s string) string {
	if s == "" {
		return "'', "
	}
	cutset := `'`
	text := strings.TrimLeft(s, cutset)
	text = strings.TrimRight(text, cutset)
	text = strings.ReplaceAll(text, `'`, `''`)
	return "'" + text + "', "
}

func genSeed(quotes []scraper.Quote) error {
	var sb strings.Builder
	sb.WriteString("BEGIN;\n\nINSERT INTO quotes (text, author, source, word_count,tags) VALUES\n")

	for i, quote := range quotes {
		sb.WriteString(" (")
		sb.WriteString(trim(quote.Text))
		sb.WriteString(trim(quote.Author))
		sb.WriteString(trim(quote.Source))

		wordCount := len(strings.Fields(quote.Text))
		sb.WriteString(strconv.Itoa(wordCount) + ", '")
		if len(quote.Tags) > 0 {
			tagsByte, _ := json.Marshal(quote.Tags)
			tags := bytes.ReplaceAll(tagsByte, []byte("'"), []byte("''"))
			sb.Write(tags)
		} else {
			sb.WriteString("[]")
		}
		if i == len(quotes)-1 {
			sb.WriteString("'::jsonb);")
		} else {
			sb.WriteString("'::jsonb),\n")
		}
	}
	sb.WriteString("COMMIT;\n")
	os.WriteFile("init-db/02_seed.sql", []byte(sb.String()), 0o644)
	return nil
}

func readFromFile(filename string) ([]scraper.Quote, error) {
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

func run() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	port, err := strconv.ParseUint(os.Getenv("PORT"), 0, 16)
	if err != nil {
		log.Fatal(err)
	}
	cfg := pq.Config{
		Host:           os.Getenv("HOST"),
		Port:           uint16(port),
		User:           os.Getenv("POSTGRES_USER"),
		Password:       os.Getenv("POSTGRES_PASSWORD"),
		Database:       os.Getenv("POSTGRES_DB"),
		ConnectTimeout: 5 * time.Second,
	}

	c, err := pq.NewConnectorConfig(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Create connection pool.
	db := sql.OpenDB(c)
	defer db.Close()

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	err = writeWithRollback(db)
	if err != nil {
		log.Fatal(err)
	}
}

func writeWithRollback(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	query := "insert into quotes (text,author,source,tags,word_count) values($1,$2,$3,$4::jsonb,$5)"
	quotes, err := readFromFile("quotes.jsonl")
	if err != nil {
		return err
	}
	for _, quote := range quotes {
		words := strings.Fields(quote.Text)
		wordCount := len(words)
		_, err := tx.Exec(query, quote.Text, quote.Author, quote.Source, quote.Tags, wordCount)
		if err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	fmt.Println("Quotes written to postgres successfully.")
	return nil
}
