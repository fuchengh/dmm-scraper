package client

import (
	"net/http"
	"net/url"

	"github.com/imroc/req/v3"
)

// ReqClient ...
type ReqClient struct {
	client *req.Client
}

func (rc *ReqClient) SetProxyUrl(rawurl string) error {
	_, err := url.Parse(rawurl)
	if err != nil {
		return err
	}
	rc.client.SetProxyURL(rawurl)
	return nil
}

// Get ...
func (rc *ReqClient) Get(url string, v interface{}) (*http.Response, error) {
	resp, err := rc.client.R().
		SetCookies(v.(*http.Cookie)).
		Get(url)
	if err != nil {
		return nil, err
	}
	return resp.Response, nil
}

// GetJSON ...
func (rc *ReqClient) GetJSON(url string, v interface{}) error {
	resp, err := rc.client.R().
		SetSuccessResult(v).
		Get(url)
	if err != nil {
		return err
	}
	return resp.Err

}

// Post ...
func (rc *ReqClient) Post(url string, v interface{}) (*http.Response, error) {
	resp, err := rc.client.R().
		SetBody(v).
		Post(url)
	if err != nil {
		return nil, err
	}
	return resp.Response, err
}

// Download ...
func (rc *ReqClient) Download(url, filename string, progress func(current, total int64)) error {
	resp, err := rc.client.R().
		SetOutputFile(filename).
		Get(url)
	if err != nil {
		return err
	}
	return resp.Err
}
