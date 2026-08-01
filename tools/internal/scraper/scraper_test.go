package scraper

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

func TestMakeFullURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		params   map[string]string
		contains []string
	}{
		{
			name:    "no extra params",
			baseURL: "https://www.goodreads.com/quotes/search",
			params:  map[string]string{},
			contains: []string{
				"https://www.goodreads.com/quotes/search",
				"commit=Search",
				"utf8=%E2%9C%93",
			},
		},
		{
			name:    "with author param",
			baseURL: "https://www.goodreads.com/quotes/search",
			params:  map[string]string{"q": "Tom Holt"},
			contains: []string{
				"q=Tom+Holt",
				"commit=Search",
			},
		},
		{
			name:    "with page param",
			baseURL: "https://www.goodreads.com/quotes/search",
			params:  map[string]string{"q": "Author", "page": "3"},
			contains: []string{
				"page=3",
				"q=Author",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := makeFullURL(tc.baseURL, tc.params)
			if err != nil {
				t.Fatalf("makeFullURL failed: %v", err)
			}
			for _, c := range tc.contains {
				if !strings.Contains(result, c) {
					t.Errorf("expected URL to contain %q, got %q", c, result)
				}
			}
		})
	}
}

func TestMakeFullURL_InvalidURL(t *testing.T) {
	_, err := makeFullURL("://invalid-url", nil)
	if err == nil {
		t.Error("expected error for invalid URL, got nil")
	}
}

func TestSplitAndJoin(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		spliton   string
		delimiter string
		expected  string
	}{
		{
			name:      "replace br with space",
			text:      "hello<br/>world",
			spliton:   "<br/>",
			delimiter: " ",
			expected:  "hello world",
		},
		{
			name:      "multiple br tags",
			text:      "a<br/>b<br/>c",
			spliton:   "<br/>",
			delimiter: " ",
			expected:  "a b c",
		},
		{
			name:      "no split match",
			text:      "hello world",
			spliton:   "<br/>",
			delimiter: " ",
			expected:  "hello world",
		},
		{
			name:      "empty string",
			text:      "",
			spliton:   "<br/>",
			delimiter: " ",
			expected:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := splitAndJoin(tc.text, tc.spliton, tc.delimiter)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestExtractQuoteDivContent(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "valid quote text",
			html:     `<div class="quoteText">This is a quote<span class="authorOrTitle">Author</span></div>`,
			expected: "This is a quote",
		},
		{
			name:     "no authorOrTitle span",
			html:     `<div class="quoteText">Just text</div>`,
			expected: "",
		},
		{
			name:     "multiline quote",
			html:     `<div class="quoteText">Line 1<br/>Line 2<span class="authorOrTitle">Author</span></div>`,
			expected: "Line 1<br/>Line 2",
		},
		{
			name:     "no quoteText div",
			html:     `<div>Nothing here</div>`,
			expected: "",
		},
		{
			name:     "empty quoteText",
			html:     `<div class="quoteText"><span class="authorOrTitle">Author</span></div>`,
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractQuoteDivContent(tc.html)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestExtractQuoteText(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name: "simple quote",
			html: `<div class="quoteText">A great quote<span class="authorOrTitle">Author</span></div>`,
			expected: "A great quote",
		},
		{
			name: "quote with br tags",
			html: `<div class="quoteText">Line 1<br/>Line 2<span class="authorOrTitle">Author</span></div>`,
			expected: "Line 1 Line 2",
		},
		{
			name: "quote with em dash author separator",
			html: `<div class="quoteText">The quote text<br/>―<span class="authorOrTitle">Author</span></div>`,
			expected: "The quote text",
		},
		{
			name:     "empty html",
			html:     ``,
			expected: "",
		},
		{
			name: "contains html tags (preserved)",
			html: `<div class="quoteText">Text with <i>italic</i><span class="authorOrTitle">Author</span></div>`,
			expected: "Text with <i>italic</i>",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := extractQuoteText(tc.html)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestAppendToJSONL(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test_quotes_*.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	quotes := []Quote{
		{Text: "Hello world", Author: "Author1", Source: "Book1", Tags: []Tag{"tag1"}},
		{Text: "Another quote", Author: "Author2", Source: "Book2", Tags: []Tag{"tag2", "tag3"}},
	}

	err = AppendToJSONL(tmpFile.Name(), quotes)
	if err != nil {
		t.Fatalf("AppendToJSONL failed: %v", err)
	}

	err = AppendToJSONL(tmpFile.Name(), quotes)
	if err != nil {
		t.Fatalf("second AppendToJSONL failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

func TestMakeQuotes(t *testing.T) {
	html := `<!DOCTYPE html>
<html><body>
<div class="quoteDetails">
  <div class="quoteText">
    "Life is what happens when you're busy making other plans."
    <br/>
    ―
    <span class="authorOrTitle">John Lennon</span>
  </div>
  <div class="quoteFooter">
    <div class="left">
      <a href="/quotes/tag/life">life</a>
      <a href="/quotes/tag/inspirational">inspirational</a>
    </div>
  </div>
</div>
</body></html>`

	doc, err := goqueryFromString(html)
	if err != nil {
		t.Fatalf("failed to parse HTML: %v", err)
	}

	quotes := MakeQuotes(doc)
	if len(quotes) == 0 {
		t.Fatal("expected at least one quote, got none")
	}

	q := quotes[0]
	if q.Author != "John Lennon" {
		t.Errorf("expected author 'John Lennon', got %q", q.Author)
	}
	if len(q.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(q.Tags))
	}
	expectedTags := []Tag{"life", "inspirational"}
	if !reflect.DeepEqual(q.Tags, expectedTags) {
		t.Errorf("expected tags %v, got %v", expectedTags, q.Tags)
	}
	if q.Text == "" {
		t.Error("expected non-empty quote text")
	}
}

func goqueryFromString(html string) (*goquery.Document, error) {
	reader := strings.NewReader(html)
	return goquery.NewDocumentFromReader(reader)
}
