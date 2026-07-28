package cleaner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/HmonWutt/TYOQ-Type-your-own-quotes/tools/internal/scraper"
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

func CleanQuotes(fromPath string, toPath string) error {
	quotes, err := readJSONL(fromPath)
	cleanedQuotes := []scraper.Quote{}
	if err != nil {
		return err
	}
	for i := range quotes {
		original_quote := quotes[i]
		cleanText := cleanText(original_quote.Text)

		// fmt.Println("clean:\n" + cleanText)
		if cleanText != "" {
			original_quote.Text = cleanText
			cleanedQuotes = append(cleanedQuotes, original_quote)
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
	if !strings.Contains(source, `" `) {
		source = source[1 : len(source)-1]
	}

	return source
}
