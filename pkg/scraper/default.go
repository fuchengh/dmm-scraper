package scraper

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"dmm-scraper/pkg/archive"
	myclient "dmm-scraper/pkg/client"

	"github.com/PuerkitoBio/goquery"
	"github.com/imroc/req/v3"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type DefaultScraper struct {
	doc       *goquery.Document
	cookie    *http.Cookie
	isArchive bool
}

func (DefaultScraper) FetchDoc(query string) (err error) {
	return nil
}

func (DefaultScraper) GetPlot() string {
	return ""
}

func (DefaultScraper) GetTitle() string {
	return ""
}

func (DefaultScraper) GetTranslatedTitle() string {
	return ""
}

func (DefaultScraper) GetRating() string {
	return ""
}

func (DefaultScraper) GetDirector() string {
	return ""
}

func (DefaultScraper) GetRuntime() string {
	return ""
}

func (DefaultScraper) GetTags() []string {
	return []string{}
}

func (DefaultScraper) GetMaker() string {
	return ""
}

func (DefaultScraper) GetActors() []string {
	return []string{}
}

func (DefaultScraper) GetLabel() string {
	return ""
}

func (DefaultScraper) GetNumber() string {
	return ""
}

func (DefaultScraper) GetFormatNumber() string {
	return ""
}

func (DefaultScraper) GetCover() string {
	return ""
}

func (DefaultScraper) GetPoster() string {
	return ""
}

func (s *DefaultScraper) GetWebsite() string {
	if s.doc == nil {
		return ""
	}
	return s.doc.Url.String()
}

func (DefaultScraper) GetPremiered() string {
	return ""
}

func (DefaultScraper) GetYear() string {
	return ""
}

func (DefaultScraper) GetSeries() string {
	return ""
}

func (DefaultScraper) GetType() string {
	return ""
}

func (s *DefaultScraper) GetDocFromURL(u string) (err error) {
	log.Debugf("Scraper: fetching doc from url %s", u)
	if s.cookie == nil {
		s.cookie = &http.Cookie{}
	}
	res, err := client.Get(u, s.cookie)
	if err != nil {
		return err
	}

	bodyBytes, _ := io.ReadAll(res.Body)
	defer res.Body.Close()

	r, name, certain, err := myclient.ToUtf8Encoding(ioutil.NopCloser(bytes.NewBuffer(bodyBytes)))
	if err != nil {
		return err
	}
	log.Debugf("detect content %s %v", name, certain)
	switch name {
	case "utf-8":
		s.doc, err = goquery.NewDocumentFromReader(r)
	default:
		reader := transform.NewReader(ioutil.NopCloser(bytes.NewBuffer(bodyBytes)), japanese.EUCJP.NewDecoder())
		s.doc, err = goquery.NewDocumentFromReader(reader) //
	}

	if err != nil {
		return err
	}
	s.doc.Url = res.Request.URL
	return nil
}

// Download ...
func Download(url, filename string, progress func(info req.DownloadInfo)) error {
	return client.Download(url, filename, progress)
}

func (s *DefaultScraper) GetAvailableUrl(orginUrl string) (string, error) {

	resp := &archive.AvailableResp{}
	err := client.GetJSON(fmt.Sprintf(archive.GetAvailableUrl, orginUrl), resp)
	if err != nil {
		return "", err
	}

	return resp.ArchivedSnapshots.Closest.URL, nil
}

// GetOutputPath ...
func GetOutputPath(s Scraper, conf string) string {
	p := strings.Replace(conf, "{year}", s.GetYear(), 1)
	actors := s.GetActors()
	if len(actors) > 0 {
		actorList := make([]string, 0)
		for _, actor := range actors {
			actorList = append(actorList, actor)
			if len(actorList) >= 5 {
				actorList = append(actorList, "…")
				break
			}
		}
		actorDir := strings.Join(actorList, ",")
		p = strings.Replace(p, "{actor}", actorDir, 1)
	} else {
		p = strings.Replace(p, "{actor}", "#未知女優", 1)
	}
	p = strings.Replace(p, "{maker}", s.GetMaker(), 1)
	p = strings.Replace(p, "{num}", s.GetFormatNumber(), 1)

	return strings.Replace(p, "//", "/", -1)
}
