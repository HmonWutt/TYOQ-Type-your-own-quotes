package db

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/cleaner"
	_ "modernc.org/sqlite"
)

func seedDBPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "seed.db")
}

func Insert(db *sql.DB, q *cleaner.CleanQuote) (int64, error) {
	sql := "INSERT OR IGNORE INTO quotes (text, author, source, word_count, tags) VALUES (?, ?, ?, ? ,?);"
	tagsJSON, _ := json.Marshal(q.Tags)
	authorsJSON, _ := json.Marshal(q.Authors)
	result, err := db.Exec(sql, q.Text, string(authorsJSON), q.Source, len(strings.Fields(q.Text)), string(tagsJSON))
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return id, nil
}

func Seed() error {
	// connect to the SQLite database
	db, err := sql.Open("sqlite", seedDBPath())
	if err != nil {
		return err
	}
	defer db.Close()

	// Make sure it works.
	err = db.Ping()
	if err != nil {
		return err
	}
	err = setupSchema(db)
	if err != nil {
		return err
	}
	quotes, err := ReadFromFile("../data/clean.jsonl")
	if err != nil {
		return err
	}
	err = batchInsert(db, quotes)
	return err
}

func batchInsert(db *sql.DB, quotes []cleaner.CleanQuote) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()
	stmt, err := tx.Prepare(`
        INSERT OR IGNORE INTO quotes (text, author, source, word_count,tags)
        VALUES (?, ?, ?, ?, ?)
    `)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, q := range quotes {
		length := len(strings.Fields(q.Text))
		if length > 200 {
			continue
		}
		authorsJSON, _ := json.Marshal(q.Authors)
		tagsJSON, _ := json.Marshal(q.Tags)
		if _, err := stmt.Exec(q.Text, string(authorsJSON), q.Source, length, string(tagsJSON)); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func ReadFromFile(filename string) ([]cleaner.CleanQuote, error) {
	var quotes []cleaner.CleanQuote
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var quote cleaner.CleanQuote
		err := json.Unmarshal([]byte(line), &quote)
		if err != nil {
			return nil, fmt.Errorf("failed to parse data: %v", err)
		}
		if quote.Text != "" {
			quotes = append(quotes, quote)
		}
	}
	return quotes, nil
}

func setupSchema(db *sql.DB) error {
	sql := `CREATE TABLE IF NOT EXISTS quotes (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    text       TEXT NOT NULL,
    author     TEXT DEFAULT '[]',
    source     TEXT DEFAULT 'Unknown',
    tags       TEXT DEFAULT '[]',
    word_count INTEGER,
    created_at TEXT DEFAULT (datetime('now'))
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_quotes_text_unique ON quotes(text);`
	_, err := db.Exec(sql)
	return err
}
