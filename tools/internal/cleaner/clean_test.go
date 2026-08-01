package cleaner

import (
	"os"
	"strings"
	"testing"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/scraper"
)

func TestCleanText_UnicodeQuotes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "smart double quotes become straight quotes then stripped",
			input:    "\u201C" + "hello" + "\u201D",
			expected: "hello",
		},
		{
			name:     "single quotes become double quotes then stripped",
			input:    "\u2018" + "hello" + "\u2019",
			expected: `"hello'`,
		},
		{
			name:     "smart double quotes with internal quote-space preserved",
			input:    "\u201C" + "hello \" world" + "\u201D",
			expected: `"hello " world"`,
		},
		{
			name:     "ellipsis dots become three dots",
			input:    "hello . . . world",
			expected: "hello ... world",
		},
		{
			name:     "unicode ellipsis becomes three dots",
			input:    "hello\u2026world",
			expected: "hello...world",
		},
		{
			name:     "html single quote entity replaced",
			input:    "it&#39;s nice",
			expected: "it's nice",
		},
		{
			name:     "html double quote entity replaced then stripped",
			input:    "&#34;quote&#34;",
			expected: "quote",
		},
		{
			name:     "curly apostrophe becomes straight apostrophe",
			input:    "dont\u2019stop",
			expected: "dont'stop",
		},
		{
			name:     "html tags discarded",
			input:    "\u003chere",
			expected: "",
		},
		{
			name:     "surrounding quotes stripped when no internal space-quote",
			input:    "\u201Chello\u201D",
			expected: "hello",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := cleanText(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestCleanText_PlainText(t *testing.T) {
	input := "plain text unchanged"
	expected := "plain text unchanged"
	result := cleanText(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCleanText_WithQuoteSpace(t *testing.T) {
	input := "\u201Chello \" world\u201D"
	expected := `"hello " world"`
	result := cleanText(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestCleanText_EmptyString(t *testing.T) {
	result := cleanText("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestCleanText_SingleChar(t *testing.T) {
	result := cleanText("x")
	if result != "x" {
		t.Errorf("expected %q for single char, got %q", "x", result)
	}
}

func TestCleanText_TwoChars(t *testing.T) {
	result := cleanText("ab")
	if result != "ab" {
		t.Errorf("expected %q for two chars, got %q", "ab", result)
	}
}

func TestCleanText_OnlyQuotes(t *testing.T) {
	result := cleanText(`""`)
	if result != "" {
		t.Errorf("expected empty string for only quotes, got %q", result)
	}
}

func TestCleanText_PreservesNewlines(t *testing.T) {
	result := cleanText("line1\nline2")
	if !strings.Contains(result, "\n") {
		t.Error("expected newline to be preserved")
	}
}

func TestReadWriteJSONL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_clean_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	quotes := []scraper.Quote{
		{Text: "Hello world", Author: "Author1", Source: "Book1"},
		{Text: "Another quote", Author: "Author2", Source: "Book2"},
	}

	err = writeJSONL(tmpFile.Name(), quotes)
	if err != nil {
		t.Fatalf("writeJSONL failed: %v", err)
	}

	readBack, err := readJSONL(tmpFile.Name())
	if err != nil {
		t.Fatalf("readJSONL failed: %v", err)
	}

	if len(readBack) != len(quotes) {
		t.Errorf("expected %d quotes, got %d", len(quotes), len(readBack))
	}
	for i := range quotes {
		if readBack[i].Text != quotes[i].Text {
			t.Errorf("quote %d: expected text %q, got %q", i, quotes[i].Text, readBack[i].Text)
		}
		if readBack[i].Author != quotes[i].Author {
			t.Errorf("quote %d: expected author %q, got %q", i, quotes[i].Author, readBack[i].Author)
		}
	}
}

func TestCleanQuotes(t *testing.T) {
	tmpIn, err := os.CreateTemp("", "test_clean_in_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpIn.Name())
	tmpIn.Close()

	tmpOut := tmpIn.Name() + ".out.jsonl"
	defer os.Remove(tmpOut)

	quotes := []scraper.Quote{
		{Text: "\u201Chello world\u201D", Author: "Author1", Source: "Book1"},
		{Text: "<i>has html tag</i>", Author: "Author2", Source: "Book2"},
	}

	err = writeJSONL(tmpIn.Name(), quotes)
	if err != nil {
		t.Fatalf("writeJSONL failed: %v", err)
	}

	err = CleanQuotes(tmpIn.Name(), tmpOut)
	if err != nil {
		t.Fatalf("CleanQuotes failed: %v", err)
	}

	cleaned, err := readJSONL(tmpOut)
	if err != nil {
		t.Fatalf("readJSONL failed: %v", err)
	}

	if len(cleaned) != 1 {
		t.Fatalf("expected 1 quote (html-tagged quote dropped), got %d", len(cleaned))
	}

	t.Logf("cleaned[0].Text = %q", cleaned[0].Text)

	if !strings.Contains(cleaned[0].Text, "hello") {
		t.Errorf("expected cleaned text containing 'hello', got %q", cleaned[0].Text)
	}
}

func TestReadJSONL_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_empty_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	quotes, err := readJSONL(tmpFile.Name())
	if err != nil {
		t.Fatalf("readJSONL failed: %v", err)
	}
	if len(quotes) != 0 {
		t.Errorf("expected 0 quotes, got %d", len(quotes))
	}
}

func TestReadJSONL_NonexistentFile(t *testing.T) {
	_, err := readJSONL("/nonexistent/path/file.jsonl")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}
