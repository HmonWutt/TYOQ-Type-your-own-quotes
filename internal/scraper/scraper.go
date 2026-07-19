package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Quote struct {
	Text   string `json:"text"`
	Author string `json:"author"`
	Source string `json:"source"`
	Tags   []Tag  `json:"tags"`
}

type Tag string

func Check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func Scrape(url string) []Quote {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := http.Client{}
	res, err := client.Do(req)
	Check(err)
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)

	Check(err)
	return MakeQuotes(doc)
}

func ScrapeAllPagesAndWriteToFile(baseURL string, path string) {
	totalPages := 100
	var allQuotes []Quote
	for i := 1; i < totalPages+1; i++ {
		query := fmt.Sprintf("?page=%d", i)
		URL := baseURL + query
		quotes := Scrape(URL)
		allQuotes = append(allQuotes, quotes...)
		time.Sleep(1 * time.Second)
	}
	write(path, allQuotes)
}

func MakeQuotes(doc *goquery.Document) []Quote {
	var quotes []Quote

	doc.Find(".quoteDetails").Each(func(i int, qd *goquery.Selection) {
		var quote Quote
		var doubleQuotesCount int
		oldNew := map[string]string{
			"\u2018": `'`, // left single quote
			"\u2019": `'`, // right single quote
			"\u2027": `'`, // apostrophe
			",":      "",  // remove "," from the end. The original looks like this "Suzanne Collins, Hunger games"
		}

		qd.Find(".quoteText").Each(func(_ int, s *goquery.Selection) {
			s.Find(".authorOrTitle").Each(func(j int, a *goquery.Selection) {
				if j == 0 {
					author := strings.TrimSpace(a.Text())

					quote.Author = cleanText(oldNew, author)
				} else {

					source := strings.TrimSpace(a.Text())
					// fmt.Printf("Book: %s\n", source)
					quote.Source = cleanText(oldNew, source)
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
				oldNewQuoteText := map[string]string{
					"\u201C": `"`, // left curly quote
					"\u201D": `"`, // right curly quote
					//"\n", " ", // replace empty new line with " " instead of "" cause "" leaves no space between the two sentences
					"\u2018": `'`, // left single quote
					"\u2019": `'`, // right single quote
					"\u2027": `'`, // apostrophe look alike
					"&#34;":  `"`, // html ""
					"&#39;":  `'`, // html ""
				}
				quoteText = cleanText(oldNewQuoteText, quoteText)
				doubleQuotesCount = strings.Count(quoteText, `"`)
				if doubleQuotesCount == 2 {
					// Only 2 quotes = single quoted statement, safe to trim
					quoteText = strings.Trim(quoteText, `"`) // remove leading and trailing quotations
				}
				// for _, r := range quoteText {
				// 	fmt.Printf("%U %c\n", r, r)
				// }

				quote.Text = quoteText
			}
		})

		if doubleQuotesCount%2 == 0 { // discard the quotes if `"` are unmatched
			// fmt.Printf("Tags: ")
			var tags []Tag
			qd.Find(".quoteFooter").Find(".left").Find("a").Each(func(_ int, t *goquery.Selection) {
				cleanedTag := cleanText(oldNew, t.Text())
				tag := Tag(cleanedTag)

				tags = append(tags, tag)
				// fmt.Printf("%s, ", t.Text())
			})
			quote.Tags = tags
			quotes = append(quotes, quote)
			// fmt.Println("\n-----------------------------------")
		}
	})
	return quotes
}

func cleanText(dict map[string]string, source string) string {
	for old, new := range dict {
		source = strings.ReplaceAll(source, old, new)
	}
	return source
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
