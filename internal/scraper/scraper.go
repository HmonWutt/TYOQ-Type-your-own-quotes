package scraper

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
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
		qd.Find(".quoteText").Each(func(_ int, s *goquery.Selection) {
			s.Find(".authorOrTitle").Each(func(j int, a *goquery.Selection) {
				if j == 0 {
					author := strings.TrimSpace(a.Text())
					fmt.Printf("Author: %s\n", author)
					quote.Author = author
				} else {

					source := strings.TrimSpace(a.Text())
					fmt.Printf("Book: %s\n", source)
					quote.Source = source
				}
			})

			parts := strings.Split(qd.Text(), "―")
			quoteText := strings.TrimSpace(parts[0])
			quoteText = strings.ReplaceAll(quoteText, "\n", "")
			quoteText = strings.Trim(quoteText, "\u201C\u201D")
			// for _, r := range quoteText {
			// 	fmt.Printf("%U %c\n", r, r)
			// }
			fmt.Printf("Quote: %s\n", quoteText)
			quote.Text = quoteText
		})
		fmt.Printf("Tags: ")
		var tags []Tag
		qd.Find(".quoteFooter").Find(".left").Find("a").Each(func(_ int, t *goquery.Selection) {
			tag := Tag(t.Text())
			tags = append(tags, tag)
			fmt.Printf("%s, ", t.Text())
		})
		quote.Tags = tags
		quotes = append(quotes, quote)
		fmt.Println("\n-----------------------------------")
	})
	return quotes
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
