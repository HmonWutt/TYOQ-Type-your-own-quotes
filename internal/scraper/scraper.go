package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Scraper struct {
	OutputFile string
	StartIndex int
	Offset     int
	BaseURL    string
	Author     string
	Referer    string
}
type (
	Tag   string
	Quote struct {
		Text   string `json:"text"`
		Author string `json:"author"`
		Source string `json:"source"`
		Tags   []Tag  `json:"tags"`
	}
)

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func makeFullURL(baseURL string, params map[string]string) string {
	u, _ := url.Parse(baseURL)

	q := u.Query()
	q.Set("commit", "Search")
	q.Set("utf8", "✓")
	for key, val := range params {
		q.Set(key, val)
	}

	u.RawQuery = q.Encode()
	// fmt.Println(u.String())
	return u.String()
}

func Scrape(url string, referer string) []Quote {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("referer", referer)
	req.Header.Set("User-Agent", randomUserAgent())
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: 60 * time.Second,
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
		quotes := Scrape(URL, "")
		allQuotes = append(allQuotes, quotes...)
		time.Sleep(2 * time.Second)
	}
	write(path, allQuotes)
}

func MakeQuotes(doc *goquery.Document) []Quote {
	var quotes []Quote
	doc.Find(".quoteDetails").Each(func(i int, quoteDetails *goquery.Selection) {
		var quote Quote
		html, _ := quoteDetails.Html()
		quoteText := extractQuoteText(html)
		if quoteText != "" {
			quote.Text = quoteText
			quoteDetails.Find(".quoteText").Each(func(_ int, s *goquery.Selection) {
				// if doubleQuotesCount%2 == 0 && len(quote.Text) > 0 { // only take if quotes number fo `"` and `'` are matched and text is not empty
				s.Find(".authorOrTitle").Each(func(j int, a *goquery.Selection) {
					text := strings.TrimSpace(a.Text())
					if j == 0 {
						quote.Author = text
					} else {
						quote.Source = text
					}
				})
			})

			var tags []Tag
			quoteDetails.Find(".quoteFooter").Find(".left").Find("a").Each(func(_ int, t *goquery.Selection) {
				tag := Tag(t.Text())
				tags = append(tags, tag)
			})
			quote.Tags = tags
			quotes = append(quotes, quote)
		}
	})
	return quotes
}

func extractQuoteText(html string) string {
	quoteText := extractQuoteDivContent(html)
	if quoteText == "" {
		return quoteText
	}
	quoteText = splitAndJoin(quoteText, "<br/>", " ")
	//	if !strings.Contains(joined, "<") { // discard if it contains formatting like <i> <b> etc
	var parts []string
	if quoteText == "" {
		return quoteText
	}
	parts = strings.Split(quoteText, "―")
	if len(parts) == 0 {
		return ""
	}
	quoteText = strings.TrimSpace(parts[0]) // remove leading and trailing white space
	quoteText = strings.Join(strings.Fields(quoteText), " ")
	return quoteText
}

func extractQuoteDivContent(html string) string {
	re := regexp.MustCompile(`(?s)<div class="quoteText">(.*?)<span class="authorOrTitle">`)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
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

func (s *Scraper) ScrapeAndAppend() error {
	for i := s.StartIndex; i < s.Offset; i++ {
		params := map[string]string{}
		params["q"] = s.Author
		if i > 1 {
			params["page"] = fmt.Sprintf("%d", i)
		}
		fullURL := makeFullURL(s.BaseURL, params)

		fmt.Printf("Scraping: %s\n", fullURL)
		quotes := Scrape(fullURL, s.Referer)
		fmt.Printf("Total quotes: %d\n", len(quotes))
		s.Referer = fullURL
		err := AppendToJSONL(s.OutputFile, quotes)
		if err != nil {
			return err
		}
		randomDelay := time.Duration(rand.Intn(60)) * time.Second
		fmt.Printf("Sleeping for %v \n", randomDelay)
		time.Sleep(randomDelay)
	}
	return nil
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
