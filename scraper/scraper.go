package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func Scrape(url string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", "Mozilla/5.0")
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".quoteText").Each(func(i int, s *goquery.Selection) {
		s.Find(".authorOrTitle").Each(func(j int, a *goquery.Selection) {
			if j == 0 {
				fmt.Printf("Author: %s\n", a.Text())
			} else {
				fmt.Printf("Book: %s\n", a.Text())
			}
		})

		parts := strings.Split(s.Text(), "―")
		fmt.Printf("Quote: %s\n", parts[0])
	})
}
