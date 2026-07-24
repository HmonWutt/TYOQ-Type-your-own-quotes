package scraper

import "strings"

const (
	RIGHTDOUBLEQUOTE = "\u201D"
	LEFTDOUBLEQUOTE  = "\u201C"
	NEWLINE          = "\n"
	LEFTSINGLEQUOTE  = "\u2018"
	RIGHTSINGLEQUOTE = "\u2019"
	APOSTROPHE       = "\u2027"
	HTMLSINGLEQUOTE  = "&#34;"
	HTMLDOUBLEQUOTE  = "&#39;"
)

func removeLeadingAndTrailingQuotes(quoteText string) string {
	var cleanText string
	doubleQuotesCount := strings.Count(quoteText, `"`)
	if doubleQuotesCount == 2 {
		// Only 2 quotes = single quoted statement, safe to trim
		cleanText = strings.Trim(quoteText, `"`) // remove leading and trailing quotations
	}
	return cleanText
}

func cleanText(dict map[string]string, source string) string {
	for old, new := range dict {
		source = strings.ReplaceAll(source, old, new)
	}
	return source
}

func splitAndJoin(text string, spliton string, delimiter string) string {
	var joined string
	parts := strings.Split(text, spliton)
	joined = strings.Join(parts, delimiter)
	return joined
}
