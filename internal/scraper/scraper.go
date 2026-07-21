package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type (
	Tag   string
	Quote struct {
		Text   string `json:"text"`
		Author string `json:"author"`
		Source string `json:"source"`
		Tags   []Tag  `json:"tags"`
	}
)

type Scraper struct {
	buffer    []Quote
	filename  string
	batchSize int
}

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

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

func makeFullURL(baseURL string, author string) string {
	u, _ := url.Parse(baseURL)

	q := u.Query()
	q.Set("q", author)
	q.Set("commit", "Search")
	u.RawQuery = q.Encode()

	fmt.Println(u.String())
	return u.String()
}

func Scrape(url string) []Quote {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Connection", "keep-alive")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := client.Do(req)
	Check(err)
	defer res.Body.Close()
	if res.StatusCode == 200 || res.StatusCode == 202 {
		doc, err := goquery.NewDocumentFromReader(res.Body)
		Check(err)
		return MakeQuotes(doc)
	}

	log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	return []Quote{}
}

func ScrapeAllPagesAndWriteToFile(baseURL string, path string) {
	totalPages := 100
	var allQuotes []Quote
	for i := 1; i < totalPages+1; i++ {
		query := fmt.Sprintf("?page=%d", i)
		URL := baseURL + query
		quotes := Scrape(URL)
		allQuotes = append(allQuotes, quotes...)
		time.Sleep(2 * time.Second)
	}
	write(path, allQuotes)
}

func MakeQuotes(doc *goquery.Document) []Quote {
	var quotes []Quote
	doc.Find(".quoteDetails").Each(func(i int, qd *goquery.Selection) {
		var quote Quote
		// var doubleQuotesCount int
		// oldNew := map[string]string{
		// 	"\u2018": `'`, // left single quote
		// 	"\u2019": `'`, // right single quote
		// 	"\u2027": `'`, // apostrophe
		// 	",":      "",  // remove "," from the end. The original looks like this "Suzanne Collins, Hunger games"
		// }
		// oldNewQuoteText := map[string]string{
		// 	"\u201C": `"`, // left curly quote
		// 	"\u201D": `"`, // right curly quote
		// 	//"\n", " ", // replace empty new line with " " instead of "" cause "" leaves no space between the two sentences
		// 	"\u2018": `'`, // left single quote
		// 	"\u2019": `'`, // right single quote
		// 	"\u2027": `'`, // apostrophe look alike
		// 	"&#34;":  `"`, // html ""
		// 	"&#39;":  `'`, // html ""
		// }

		qd.Find(".quoteText").Each(func(_ int, s *goquery.Selection) {
			s.Find(".authorOrTitle").Each(func(j int, a *goquery.Selection) {
				text := strings.TrimSpace(a.Text())
				if j == 0 {
					quote.Author = text
				} else {
					quote.Source = text
				}
			})
			html, _ := qd.Html()
			re := regexp.MustCompile(`(?s)<div class="quoteText">(.*?)<span class="authorOrTitle">`)
			matches := re.FindStringSubmatch(html)
			var joined string
			if len(matches) > 1 {
				quoteText := matches[1]

				parts := strings.Split(quoteText, "<br/>")
				joined = strings.Join(parts, " ")
			}

			if !strings.Contains(joined, "<") { // discard if it contains formatting like <i> <b> etc
				parts := strings.Split(joined, "―")
				quoteText := strings.TrimSpace(parts[0]) // remove leading and trailing white space

				quoteText = strings.Join(strings.Fields(quoteText), " ")
				quote.Text = quoteText
			}
		})

		// if doubleQuotesCount%2 == 0 && len(quote.Text) > 0 { // only take if quotes number fo `"` and `'` are matched and text is not empty
		// fmt.Printf("Tags: ")
		var tags []Tag
		qd.Find(".quoteFooter").Find(".left").Find("a").Each(func(_ int, t *goquery.Selection) {
			// cleanedTag := cleanText(oldNew, t.Text())
			// tag := Tag(cleanedTag)
			tag := Tag(t.Text())

			tags = append(tags, tag)
		})
		quote.Tags = tags
		quotes = append(quotes, quote)
	})
	return quotes
}

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

func write(filepath string, quotes []Quote) {
	file, _ := os.Create(filepath)
	defer file.Close()

	encoder := json.NewEncoder(file)
	for _, quote := range quotes {
		err := encoder.Encode(quote)
		if err != nil {
			log.Println("failed to write to file")
		}
	}
	log.Printf("✓ Saved %d quotes\n", len(quotes))
}

func ScrapeAndAppend(author string, baseURL string, path string, startIndex int, offset int) error {
	fullURL := makeFullURL(baseURL, author)
	var allQuotes []Quote
	for i := startIndex; i < offset; i++ {
		fmt.Printf("Scraping page: %d\n", i)
		URL := fmt.Sprintf("%s&page=%d", fullURL, i)
		quotes := Scrape(URL)
		allQuotes = append(allQuotes, quotes...)
		time.Sleep(5 * time.Second)
	}

	return AppendToJSONL(path, allQuotes)
}

func AppendToJSONL(filename string, quotes []Quote) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, quote := range quotes {
		bytes, _ := json.Marshal(quote)
		_, err = file.Write(bytes)
		if err != nil {
			return err
		}
		_, err = file.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
