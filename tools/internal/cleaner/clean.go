package cleaner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/scraper"
	"github.com/pemistahl/lingua-go"
)

const (
	RIGHTDOUBLEQUOTE = `”`
	LEFTDOUBLEQUOTE  = `“`
	NEWLINE          = "\n"
	LEFTSINGLEQUOTE  = "\u2018"
	RIGHTSINGLEQUOTE = "\u2019"
	APOSTROPHE       = `’`
	HTMLSINGLEQUOTE  = "&#39;"
	HTMLDOUBLEQUOTE  = "&#34;"
	HTMLOPENINGTAG   = "\u003c"
)

var (
	zeroWidthChars = []string{"\u200B", "\u200C", "\u200D", "\u200E", "\u200F", "\uFEFF"}
	languages      = []lingua.Language{
		lingua.English,
		lingua.French,
		lingua.German,
		lingua.Portuguese,
		lingua.Croatian,
		lingua.Serbian,
		lingua.Arabic,
		lingua.Turkish,
		lingua.Spanish,
		lingua.Italian,
		lingua.Lithuanian,
		lingua.Bulgarian,
	}
)

type (
	Tag        string
	Author     string
	CleanQuote struct {
		Text    string
		Authors []Author
		Source  string
		Tags    []Tag
	}
)

var detector = lingua.NewLanguageDetectorBuilder().
	FromLanguages(languages...).WithMinimumRelativeDistance(0.3).
	Build()

func isEnglish(detector lingua.LanguageDetector, quote string) bool {
	language, exists := detector.DetectLanguageOf(quote)
	if !exists {
		return true
	}
	return language == lingua.English
}

func readJSONL(path string) ([]scraper.Quote, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var quotes []scraper.Quote
	scanner := bufio.NewScanner(file)

	// buf := make([]byte, 0, 64*1024)
	// scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue // skip blank lines
		}
		var quote scraper.Quote
		if err := json.Unmarshal(line, &quote); err != nil {
			return nil, fmt.Errorf("unmarshal line: %w", err)
		}
		quotes = append(quotes, quote)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return quotes, nil
}

func writeJSONL(path string, quotes []scraper.Quote) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	for _, quote := range quotes {
		if err := encoder.Encode(quote); err != nil {
			return fmt.Errorf("encode quote: %w", err)
		}
	}
	return nil
}

func CleanQuotes(fromPath string, toPath string) error {
	quotes, err := readJSONL(fromPath)
	cleanedQuotes := []CleanQuote{}
	if err != nil {
		return err
	}
	for i := range quotes {
		originalQuote := quotes[i]
		cleanQuote := CleanQuote{}
		cleanedText := cleanText(originalQuote.Text)
		cleanedAuthors := cleanAuthors(originalQuote.Author)
		// fmt.Println("clean:\n" + cleanText)
		if cleanedText != "" {
			cleanQuote.Text = cleanedText
			cleanQuote.Authors = cleanedAuthors
			cleanedQuotes = append(cleanedQuotes, cleanQuote)
		} else {
			// fmt.Println("empty quote: " + original_quote.Text)
		}
	}
	err = scraper.AppendToJSONL(toPath, cleanedQuotes)
	return err
}

func cleanText(source string) string {
	if source == "" {
		return ""
	}
	if strings.Contains(source, HTMLOPENINGTAG) { // discard if the text contains nested tags <i> <b> <a> etc; no much work to clean"
		return ""
	}
	for _, zw := range zeroWidthChars {
		source = strings.ReplaceAll(source, zw, "")
	}

	for _, r := range source {
		if r >= 0x1D00 && r <= 0x1D7F {
			return "" // drop quotes containing small-cap characters
		}
	}
	if !isEnglish(detector, source) {
		return ""
	}
	source = strings.ReplaceAll(source, `’ `, `" `)
	if strings.Contains(source, `' `) {
		source = strings.ReplaceAll(source, `' `, `" `)
	}

	oldNew := map[string]string{
		RIGHTDOUBLEQUOTE: `"`,
		LEFTDOUBLEQUOTE:  `"`,
		HTMLSINGLEQUOTE:  `'`,
		HTMLDOUBLEQUOTE:  `"`,
		APOSTROPHE:       `'`,
		`. . .`:          `...`,
		`…`:              `...`,
		`‘`:              `"`,
	}
	for old, new := range oldNew {
		source = strings.ReplaceAll(source, old, new)
	}
	if !strings.Contains(source, `" `) && len(source) >= 2 && strings.HasPrefix(source, `"`) && strings.HasSuffix(source, `"`) {
		source = source[1 : len(source)-1]
	}
	return source
}

func cleanAuthors(authors string) []Author {
	cleanAuthors := []Author{}
	authors = strings.ReplaceAll(authors, "&", ",")
	authors = strings.ReplaceAll(authors, " and ", ",")
	authorsList := strings.SplitSeq(authors, ",")
	for author := range authorsList {
		if author != "" {
			cleanAuthors = append(cleanAuthors, Author(strings.TrimSpace(author)))
		}
	}
	return cleanAuthors
}
