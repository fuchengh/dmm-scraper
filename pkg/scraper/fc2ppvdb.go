package scraper

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Fc2PPVDbScraper struct {
	DefaultScraper
}

const fc2ppvdbUrl = "https://fc2ppvdb.com/articles/%s"

func (s *Fc2PPVDbScraper) GetType() string {
	return "Fc2PPVDbScraper"
}

func (s *Fc2PPVDbScraper) FetchDoc(query string) (err error) {
	err = s.GetDocFromURL(fmt.Sprintf(fc2ppvdbUrl, query))
	if err != nil {
		return err
	}
	if strings.Contains(s.doc.Find("title").Text(), "Not Found") {
		return fmt.Errorf("FC2 movie %s not found", query)
	}

	return err
}

func (s *Fc2PPVDbScraper) GetTranslatedTitle() string {
	if s.doc == nil {
		return ""
	}
	rawTitle := strings.Split(s.doc.Find("title").Text(), " ")
	var title string
	// remove movid Id from raw title
	for _, s := range rawTitle[1 : len(rawTitle)-3] {
		title += string(s)
	}

	if ts.Enable {
		log.Infof("Translating title for %s...", s.GetFormatNumber())
		res, err := ts.Translate(title)
		if err != nil {
			log.Errorf("Error translating title: %v", err)
			return fmt.Sprintf("%s %s", s.GetFormatNumber(), title)
		}
		return fmt.Sprintf("%s %s", s.GetFormatNumber(), res)
	}

	return fmt.Sprintf("%s %s", s.GetFormatNumber(), title)
}

func (s *Fc2PPVDbScraper) GetTitle() string {
	if s.doc == nil {
		return ""
	}
	rawTitle := strings.Split(s.doc.Find("title").Text(), " ")
	var title string
	// remove movid Id from raw title
	for _, s := range rawTitle[1 : len(rawTitle)-3] {
		title += string(s)
	}
	return title
}

func (s *Fc2PPVDbScraper) GetPlot() string {
	return ""
}

func (s *Fc2PPVDbScraper) GetDirector() string {
	if s.doc == nil {
		return ""
	}
	director := s.doc.Find("div:contains(\"販売者：\") > span > a").Text()
	if len(director) == 0 {
		return "Unknown"
	}
	return director
}

func (s *Fc2PPVDbScraper) GetRuntime() string {
	if s.doc == nil {
		return ""
	}

	runtime := s.doc.Find("div:contains(\"収録時間：\") > span").Text()

	if runtime != "" {
		t, err := time.Parse("15:04:05", runtime)
		if err != nil {
			t, err = time.Parse("04:05", runtime)
			if err != nil {
				log.Errorf("Error parsing runtime, %s", err)
				return ""
			}
			runtime = strconv.Itoa(t.Minute())
		} else {
			runtime = strconv.Itoa(t.Hour()*60 + t.Minute())
		}
	}
	return runtime
}

func (s *Fc2PPVDbScraper) GetTags() (tags []string) {
	if s.doc == nil {
		return
	}
	tags = make([]string, 0)
	s.doc.Find("div:contains(\"タグ：\") > span > a").Each(func(i int, ss *goquery.Selection) {
		tags = append(tags, strings.TrimSpace(ss.Text()))
	})
	return
}

func (s *Fc2PPVDbScraper) GetMaker() string {
	if s.doc == nil {
		return ""
	}
	return s.GetDirector()
}

func (s *Fc2PPVDbScraper) GetActors() (actors []string) {
	if s.doc == nil {
		return
	}
	actors = make([]string, 0)

	s.doc.Find("div:contains(\"女優：\") > span > a").Each(func(i int, ss *goquery.Selection) {
		actors = append(actors, strings.TrimSpace(ss.Text()))
	})

	return
}

func (s *Fc2PPVDbScraper) GetNumber() string {
	if s.doc == nil {
		return ""
	}
	id := strings.Split(strings.Split(s.doc.Find("title").Text(), " ")[0], "-")[1]
	return id
}

func (s *Fc2PPVDbScraper) GetCover() string {
	if s.doc == nil {
		return ""
	}
	img := s.doc.Find(fmt.Sprintf("img[alt=\"%s\"]", s.GetNumber())).AttrOr("src", "")
	if img != "" {
		log.Infof("Cover image not found for %s, url = %s", s.GetNumber(), img)
		return img
	}
	return ""
}

func (s *Fc2PPVDbScraper) GetPremiered() (rel string) {
	if s.doc == nil {
		return
	}

	date := s.doc.Find("div:contains(\"販売日：\") > span").Text()
	if date == "" {
		return "2000-01-01" // dummy date
	}

	return date
}

func (s *Fc2PPVDbScraper) GetYear() (rel string) {
	if s.doc == nil {
		return ""
	}

	return regexp.MustCompile(`\d{4}`).FindString(s.GetPremiered())
}

func (s *Fc2PPVDbScraper) GetFormatNumber() string {
	return strings.ToUpper(fmt.Sprintf("FC2-%s", s.GetNumber()))
}
