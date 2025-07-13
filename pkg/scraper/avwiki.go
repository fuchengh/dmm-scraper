package scraper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const avWikiUrl = "https://av-wiki.net/%s/"

func fetchAvWikiActors(movieId string) []string {
	cookie := &http.Cookie{
	} // empty cookie for now

	actors := make([]string, 0) // return value

	searchUrl := fmt.Sprintf(avWikiUrl, movieId)

	resp, err := client.Get(searchUrl, cookie)
	if err != nil {
		log.Errorf("Error sending request: %v", err)
		return actors
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("%v", err)
		}
		doc, err := goquery.NewDocumentFromReader(bytes.NewReader(bodyBytes))
		if err != nil {
			log.Errorf("Error parsing HTML: %v", err)
			return actors
		}
		doc.Find("dl[class=\"dltable\"] dd a").Each(func(i int, s *goquery.Selection) {
			actorName := s.Text()
			if actorName != "" {
				actors = append(actors, actorName)
			}
		}) // find all actor names in the document
	}
	return actors
}