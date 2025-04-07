package scraper

import (
	myclient "dmm-scraper/pkg/client"
	"dmm-scraper/pkg/config"
	"dmm-scraper/pkg/logger"
	translateClient "dmm-scraper/pkg/translate"

	"github.com/dmmlabo/dmm-go-sdk/api"
)

// Scraper is interface
type Scraper interface {
	FetchDoc(query string) (err error)
	GetPlot() string
	GetTitle() string
	GetTranslatedTitle() string
	GetRating() string
	GetDirector() string
	GetRuntime() string
	GetTags() []string
	GetMaker() string
	GetActors() []string
	GetLabel() string
	GetNumber() string
	GetFormatNumber() string
	GetCover() string
	GetPoster() string
	GetWebsite() string
	GetPremiered() string
	GetYear() string
	GetSeries() string
	GetType() string
}

var (
	client            myclient.Client
	log               logger.Logger
	dmmProductService *api.ProductService
	ts                *translateClient.Client
)

// Setup ...
func Setup(conf *config.Configs) {
	log = logger.New()
	client = myclient.New()
	if conf.Proxy.Enable {
		err := client.SetProxyUrl(conf.Proxy.Socket)
		if err != nil {
			log.Errorf("Error parse proxy url, %s, proxy disabled", err)
		}
	}
	if conf.DMMApi.ApiId != "" && conf.DMMApi.AffiliateId != "" {
		dmmProductService = api.NewProductService(conf.DMMApi.AffiliateId, conf.DMMApi.ApiId)
	}
	if conf.Translate.Enable {
		ts = translateClient.New()
		err := ts.InitTranslateApi(&conf.Translate)
		if err != nil {
			log.Errorf("Error init translate api, %s", err)
		}
	}
}
