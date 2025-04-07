package scraper

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	dmmMonoSearchUrl = "https://www.dmm.co.jp/mono/dvd/-/search/=/searchstr=%s/"
	dmmListSearchUrl = "https://www.dmm.co.jp/search/=/searchstr=%s/"
)

type DMMScraper struct {
	DefaultScraper
}

func (s *DMMScraper) GetType() string {
	return "DMMScraper"
}

func (s *DMMScraper) FetchDocFromListSearch(query string) (err error) {
	s.cookie = &http.Cookie{
		Name:    "age_check_done",
		Value:   "1",
		Path:    "/",
		Domain:  "dmm.co.jp",
		Expires: time.Now().Add(1 * time.Hour),
	}

	// Using list search to try to get the detail page
	err = s.GetDocFromURL(fmt.Sprintf(dmmListSearchUrl, query))

	if err != nil {
		log.Errorf("%s fallback search [movie %s] failed: %v", s.GetType(), query, err)
		return err
	}

	var hrefs []string
	re := regexp.MustCompile(`https://www\.dmm\.co\.jp/digital/videoa/-/detail/=/cid=(?:[^/]+)/`)
	s.doc.Find("script").Each(func(i int, s *goquery.Selection) {
		scriptText := s.Text()
		matches := re.FindAllStringSubmatch(scriptText, -1)
		for _, match := range matches {
			// check if this is a valid http link
			if len(match) > 0 {
				href := match[0]
				if strings.Contains(href, "cid=") && strings.Contains(href, "/digital/videoa/-/detail/") && strings.Contains(href, "https://") {
					hrefs = append(hrefs, href)
				}
			}
		}
	})

	if len(hrefs) == 0 {
		log.Errorf("%s fallback search [movie %s] failed: no movie found", s.GetType(), query)
		return fmt.Errorf("no movie found for query: %s", query)
	}

	var detail string
	for _, href := range hrefs {
		if isURLMatchQuery(href, query) {
			detail = href
		}
	}
	if detail == "" {
		return fmt.Errorf("unable to match in hrefs %v", hrefs)
	}

	return s.GetDocFromURL(detail)
}

// FetchDoc search once or twice to get a detail page
func (s *DMMScraper) FetchDoc(query string) (err error) {
	s.cookie = &http.Cookie{
		Name:    "age_check_done",
		Value:   "1",
		Path:    "/",
		Domain:  "dmm.co.jp",
		Expires: time.Now().Add(1 * time.Hour),
	}
	// dmm 搜索页
	if strings.Contains(query, "-") {
		strs := strings.Split(query, "-")
		if len(strs) == 2 {
			query = strs[0] + fmt.Sprintf("%05s", strs[1])
		}
	}
	err = s.GetDocFromURL(fmt.Sprintf(dmmMonoSearchUrl, query))
	if err != nil {
		return s.FetchDocFromListSearch(query)
	}

	var hrefs []string
	s.doc.Find("#list li").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Find(".tmb a").Attr("href")
		hrefs = append(hrefs, href)
	})

	if len(hrefs) == 0 {
		return s.FetchDocFromListSearch(query)
	}

	var detail string
	for _, href := range hrefs {
		if isURLMatchQuery(href, query) {
			detail = href
		}
	}
	if detail == "" {
		return fmt.Errorf("unable to match in hrefs %v", hrefs)
	}

	return s.GetDocFromURL(detail)
}

func (s *DMMScraper) GetPlot() string {
	if s.doc == nil {
		return ""
	}
	val := s.doc.Find("div[class=\"mg-b20 lh4\"] p[class=\"mg-b20\"]").Text()
	if val == "" {
		// try to get from the div tag if the p tag is empty
		tmp_html, _ := s.doc.Find("div[class=\"mg-b20 lh4\"]").Html()

		// Find the first <a> tag and get everything before it
		if idx := strings.Index(tmp_html, "<a href"); idx != -1 {
			tmp_html = tmp_html[:idx]
		}

		// Replace <br> tags with newlines
		tmp_html = strings.ReplaceAll(tmp_html, "<br>", "\n")

		// Clean up any remaining HTML tags
		re := regexp.MustCompile(`<[^>]*>`)
		val = re.ReplaceAllString(tmp_html, "")

		// Clean up extra whitespace and newlines
		val = strings.TrimSpace(val)
		// Replace multiple newlines with single newline
		re = regexp.MustCompile(`\n\s*\n`)
		val = re.ReplaceAllString(val, "\n")
	}

	if ts.Enable {
		log.Infof("Translating plot for %s...", s.GetFormatNumber())
		res, err := ts.Translate(val)
		if err != nil {
			log.Errorf("Error translating plot: %v", err)
			return val
		}
		return res
	}

	return val
}

func (s *DMMScraper) GetTranslatedTitle() string {
	if s.doc == nil {
		return ""
	}
	val := s.doc.Find("h1#title.item.fn").Text()

	if ts.Enable {
		log.Infof("Translating title for %s...", s.GetFormatNumber())
		res, err := ts.Translate(val)
		if err != nil {
			log.Errorf("Error translating title: %v", err)
			return fmt.Sprintf("%s %s", s.GetFormatNumber(), val)
		}
		return fmt.Sprintf("%s %s", s.GetFormatNumber(), res)
	}

	return fmt.Sprintf("%s %s", s.GetFormatNumber(), val)
}

func (s *DMMScraper) GetTitle() string {
	if s.doc == nil {
		return ""
	}
	val := s.doc.Find("h1#title.item.fn").Text()
	return val
}

func (s *DMMScraper) GetDirector() string {
	if s.doc == nil {
		return ""
	}
	return getDmmTableValue("監督", s.doc)
}

func (s *DMMScraper) GetRating() string {
	if s.doc == nil {
		return ""
	}
	val := s.doc.Find("p.dcd-review__average strong").Text()
	return val
}

func (s *DMMScraper) GetRuntime() string {
	if s.doc == nil {
		return ""
	}
	time := getDmmTableValue("収録時間", s.doc)
	return strings.Replace(time, "分", "", 1)
}

func (s *DMMScraper) GetTags() (tags []string) {
	if s.doc == nil {
		return
	}
	s.doc.Find("table[class=mg-b20] tr").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			if strings.Contains(s.Text(), "ジャンル") {
				s.Find("td a").Each(func(i int, ss *goquery.Selection) {
					tags = append(tags, ss.Text())
				})
				return false
			}
			return true
		})
	return
}

func (s *DMMScraper) GetMaker() string {
	if s.doc == nil {
		return ""
	}
	return getDmmTableValue("メーカー", s.doc)
}

func (s *DMMScraper) FetchAllActors() (actors []string) {
	actors = make([]string, 0)

	const baseUrl = "https://www.dmm.co.jp/mono/dvd/-/detail/performer/=/cid="
	cid := s.GetNumber()

	searchUrl := fmt.Sprintf("%s%s/", baseUrl, cid)

	// Send the request
	resp, err := client.Get(searchUrl, s.cookie)
	if err != nil {
		log.Errorf("Error sending request: %v", err)
		return
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
			return
		}
		doc.Find("a[href^='/mono/dvd/-/list/=/article=actress/id=']").Each(func(i int, s *goquery.Selection) {
			actorName := s.Text()
			actors = append(actors, actorName)
		})
	}
	return
}

func (s *DMMScraper) GetActors() (actors []string) {
	if s.doc == nil {
		return
	}
	if showMore := s.doc.Find("a[id=\"a_performer\"]:contains(\"▼すべて表示する\")"); showMore.Length() > 0 {
		actors = s.FetchAllActors()
	} else {
		s.doc.Find("#performer a").Each(func(i int, s *goquery.Selection) {
			actors = append(actors, s.Text())
		})
	}
	return
}

func (s *DMMScraper) GetLabel() string {
	if s.doc == nil {
		return ""
	}
	return getDmmTableValue("レーベル", s.doc)
}

func (s *DMMScraper) GetNumber() string {
	if s.doc == nil {
		return ""
	}
	return getDmmTableValue("品番", s.doc)
}

func (s *DMMScraper) GetFormatNumber() string {
	l, i := GetLabelNumber(s.GetNumber())
	if l == "" {
		return fmt.Sprintf("%03d", i)
	}
	return strings.ToUpper(fmt.Sprintf("%s-%03d", l, i))
}

func (s *DMMScraper) GetCover() string {
	if s.doc == nil {
		return ""
	}
	img, _ := s.doc.Find("meta[property=\"og:image\"]").Attr("content")
	return strings.Replace(img, "ps.jpg", "pl.jpg", 1)
}

func ValidPoster(posterUrl string) bool {

	// check if we are being redirected to a different page
	resp, err := client.Get(posterUrl, &http.Cookie{
		Name:    "age_check_done",
		Value:   "1",
		Path:    "/",
		Domain:  "dmm.co.jp",
		Expires: time.Now().Add(1 * time.Hour),
	})

	if err != nil {
		log.Errorf("Error sending request: %v", err)
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return (resp.Request.URL.String() == posterUrl)
	}
	return false
}

func (s *DMMScraper) GetPoster() string {
	if s.doc == nil {
		return ""
	}
	poster := fmt.Sprintf("https://pics.dmm.co.jp/digital/video/%s/%sjp-1.jpg", s.GetNumber(), s.GetNumber())

	// check if we are being redirected to a different page
	if ValidPoster(poster) {
		return poster
	}

	// Try with fallback cid
	l, i := GetLabelNumber(s.GetNumber())
	if l == "" || i == 0 {
		return ""
	}
	cid := fmt.Sprintf("%s%05d", l, i)
	fallback_poster := fmt.Sprintf("https://pics.dmm.co.jp/digital/video/%s/%sjp-1.jpg", cid, cid)
	if ValidPoster(fallback_poster) {
		return fallback_poster
	}
	// Not found, crop from cover
	return ""
}

func (s *DMMScraper) GetPremiered() (rel string) {
	if s.doc == nil {
		return ""
	}
	rel = getDmmTableValue("発売日", s.doc)
	if rel == "" {
		rel = getDmmTableValue("配信開始日", s.doc)
	}
	return strings.Replace(rel, "/", "-", -1)
}

func (s *DMMScraper) GetYear() (rel string) {
	if s.doc == nil {
		return ""
	}
	return regexp.MustCompile(`\d{4}`).FindString(s.GetPremiered())
}

func (s *DMMScraper) GetSeries() string {
	if s.doc == nil {
		return ""
	}
	return getDmmTableValue("シリーズ", s.doc)
}

func getDmmTableValue(key string, doc *goquery.Document) (val string) {
	doc.Find("table[class=mg-b20] tr").EachWithBreak(
		func(i int, s *goquery.Selection) bool {
			if strings.Contains(s.Text(), key) {
				val = s.Find("td a").Text()
				if val == "" {
					val = s.Find("td").Last().Text()
				}
				if val == "----" {
					val = ""
				}
				val = strings.TrimSpace(val)
				return false
			}
			return true
		})
	return
}

func getDmmTableValue2(x int, doc *goquery.Document) (val string) {
	//log.Info(doc.Find("table[class=mg-b20] td[width]").Html())
	return doc.Find("table[class=mg-b20] td").Eq(x - 1).Text()
}
