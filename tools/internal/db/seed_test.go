package db

import (
	"database/sql"
	"os"
	"testing"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/cleaner"
	_ "modernc.org/sqlite"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSetupSchema(t *testing.T) {
	db := openTestDB(t)

	if err := setupSchema(db); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='quotes'").Scan(&tableName)
	if err != nil {
		t.Fatalf("failed to query schema: %v", err)
	}
	if tableName != "quotes" {
		t.Errorf("expected table 'quotes', got %q", tableName)
	}

	var indexName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name='idx_quotes_text_unique'").Scan(&indexName)
	if err != nil {
		t.Fatalf("failed to query unique index: %v", err)
	}
	if indexName != "idx_quotes_text_unique" {
		t.Errorf("expected index 'idx_quotes_text_unique', got %q", indexName)
	}
}

func TestInsert(t *testing.T) {
	db := openTestDB(t)
	if err := setupSchema(db); err != nil {
		t.Fatalf("setupSchema failed: %v", err)
	}

	q := &cleaner.CleanQuote{
		Text:    "Test quote text",
		Authors: []cleaner.Author{"Test Author"},
		Source:  "Test Source",
		Tags:    []cleaner.Tag{"tag1", "tag2"},
	}

	id, err := Insert(db, q)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	if id != 1 {
		t.Errorf("expected id 1, got %d", id)
	}

	var text, author, source string
	var wordCount int
	err = db.QueryRow("SELECT text, author, source, word_count FROM quotes WHERE id = ?", id).Scan(&text, &author, &source, &wordCount)
	if err != nil {
		t.Fatalf("failed to query inserted row: %v", err)
	}
	if text != q.Text {
		t.Errorf("expected text %q, got %q", q.Text, text)
	}
	expectedAuthor := `["Test Author"]`
	if author != expectedAuthor {
		t.Errorf("expected author %q, got %q", expectedAuthor, author)
	}
	if source != q.Source {
		t.Errorf("expected source %q, got %q", q.Source, source)
	}
	expectedWordCount := 3
	if wordCount != expectedWordCount {
		t.Errorf("expected word_count %d, got %d", expectedWordCount, wordCount)
	}
}

func TestBatchInsert(t *testing.T) {
	db := openTestDB(t)
	if err := setupSchema(db); err != nil {
		t.Fatalf("setupSchema failed: %v", err)
	}

	quotes := []cleaner.CleanQuote{
		{Text: "First quote", Authors: []cleaner.Author{"Author A"}, Source: "Book A"},
		{Text: "Second quote here", Authors: []cleaner.Author{"Author B"}, Source: "Book B"},
		{Text: "Third quote goes here", Authors: []cleaner.Author{"Author C"}, Source: "Book C"},
	}

	if err := batchInsert(db, quotes); err != nil {
		t.Fatalf("batchInsert failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM quotes").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != len(quotes) {
		t.Errorf("expected %d rows, got %d", len(quotes), count)
	}

	var wordCount int
	err = db.QueryRow("SELECT word_count FROM quotes WHERE text = ?", "Second quote here").Scan(&wordCount)
	if err != nil {
		t.Fatalf("failed to query word count: %v", err)
	}
	if wordCount != 3 {
		t.Errorf("expected word_count 3, got %d", wordCount)
	}
}

func TestBatchInsert_Dedup(t *testing.T) {
	db := openTestDB(t)
	if err := setupSchema(db); err != nil {
		t.Fatalf("setupSchema failed: %v", err)
	}

	quotes := []cleaner.CleanQuote{
		{Text: "Duplicate quote", Authors: []cleaner.Author{"Author A"}, Source: "Book A"},
		{Text: "Duplicate quote", Authors: []cleaner.Author{"Author B"}, Source: "Book B"},
		{Text: "Unique quote", Authors: []cleaner.Author{"Author C"}, Source: "Book C"},
	}

	if err := batchInsert(db, quotes); err != nil {
		t.Fatalf("batchInsert failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM quotes").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 rows after dedup, got %d", count)
	}

	var author string
	err = db.QueryRow("SELECT author FROM quotes WHERE text = ?", "Duplicate quote").Scan(&author)
	if err != nil {
		t.Fatalf("failed to query author: %v", err)
	}
	expectedAuthor := `["Author A"]`
	if author != expectedAuthor {
		t.Errorf("expected first insert's author %q, got %q", expectedAuthor, author)
	}
}

func TestBatchInsert_LengthFilter(t *testing.T) {
	db := openTestDB(t)
	if err := setupSchema(db); err != nil {
		t.Fatalf("setupSchema failed: %v", err)
	}

	longText := make([]byte, 0)
	for i := 0; i < 250; i++ {
		longText = append(longText, "word "...)
	}

	quotes := []cleaner.CleanQuote{
		{Text: "Short quote", Authors: []cleaner.Author{"Author A"}, Source: "Book A"},
		{Text: string(longText), Authors: []cleaner.Author{"Author B"}, Source: "Book B"},
	}

	if err := batchInsert(db, quotes); err != nil {
		t.Fatalf("batchInsert failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM quotes").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 row (long quote filtered), got %d", count)
	}
}

func TestReadFromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_db_read_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	lines := []string{
		`{"Text":"Quote 1","Authors":["Author 1"],"Source":"Source 1","Tags":["tag1"]}`,
		`{"Text":"Quote 2","Authors":["Author 2"],"Source":"Source 2","Tags":["tag2","tag3"]}`,
	}
	for _, line := range lines {
		tmpFile.WriteString(line + "\n")
	}
	tmpFile.Close()

	quotes, err := ReadFromFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("ReadFromFile failed: %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("expected 2 quotes, got %d", len(quotes))
	}
	if quotes[0].Text != "Quote 1" {
		t.Errorf("expected text 'Quote 1', got %q", quotes[0].Text)
	}
	if len(quotes[1].Authors) != 1 || quotes[1].Authors[0] != "Author 2" {
		t.Errorf("expected author 'Author 2', got %v", quotes[1].Authors)
	}
}

func TestReadFromFile_NonexistentFile(t *testing.T) {
	_, err := ReadFromFile("/nonexistent/file.jsonl")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestInsert_NilTags(t *testing.T) {
	db := openTestDB(t)
	if err := setupSchema(db); err != nil {
		t.Fatalf("setupSchema failed: %v", err)
	}

	q := &cleaner.CleanQuote{
		Text:    "No tags quote",
		Authors: []cleaner.Author{"Author"},
		Source:  "Source",
	}

	id, err := Insert(db, q)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	var tagsJSON string
	err = db.QueryRow("SELECT tags FROM quotes WHERE id = ?", id).Scan(&tagsJSON)
	if err != nil {
		t.Fatalf("failed to query tags: %v", err)
	}
	if tagsJSON == "" {
		t.Error("expected non-empty tags JSON")
	}
}

func TestInsert_WithTags(t *testing.T) {
	db := openTestDB(t)
	if err := setupSchema(db); err != nil {
		t.Fatalf("setupSchema failed: %v", err)
	}

	q := &cleaner.CleanQuote{
		Text:    "Tagged quote",
		Authors: []cleaner.Author{"Author"},
		Source:  "Source",
		Tags:    []cleaner.Tag{"philosophy", "life"},
	}

	id, err := Insert(db, q)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	var tagsJSON string
	err = db.QueryRow("SELECT tags FROM quotes WHERE id = ?", id).Scan(&tagsJSON)
	if err != nil {
		t.Fatalf("failed to query tags: %v", err)
	}
	if tagsJSON != `["philosophy","life"]` {
		t.Errorf("expected tags JSON %q, got %q", `["philosophy","life"]`, tagsJSON)
	}
}

func TestSetupSchema_Idempotent(t *testing.T) {
	db := openTestDB(t)

	if err := setupSchema(db); err != nil {
		t.Fatalf("first setupSchema failed: %v", err)
	}
	if err := setupSchema(db); err != nil {
		t.Fatalf("second setupSchema failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='quotes'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query schema: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 quotes table, got %d", count)
	}

	var indexCount int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='index' AND name='idx_quotes_text_unique'").Scan(&indexCount)
	if err != nil {
		t.Fatalf("failed to query index: %v", err)
	}
	if indexCount != 1 {
		t.Errorf("expected 1 unique index, got %d", indexCount)
	}
}
