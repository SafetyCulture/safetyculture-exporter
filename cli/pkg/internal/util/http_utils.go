package util

import (
	"net/http"

	"github.com/dghubble/sling"
)

type DefaultHTTPDoer struct {
	Req        *http.Request
	HttpClient *http.Client
}

func (b *DefaultHTTPDoer) Do() (*http.Response, error) {
	return b.HttpClient.Do(b.Req)
}

func (b *DefaultHTTPDoer) URL() string {
	return b.Req.URL.String()
}

func (b *DefaultHTTPDoer) Error() interface{} {
	return nil
}

type SlingHTTPDoer struct {
	Sl       *sling.Sling
	Req      *http.Request
	SuccessV interface{}
	FailureV interface{}
}

func (b *SlingHTTPDoer) Do() (*http.Response, error) {
	return b.Sl.Do(b.Req, b.SuccessV, b.FailureV)
}

func (b *SlingHTTPDoer) URL() string {
	return b.Req.URL.String()
}

func (b *SlingHTTPDoer) Error() interface{} {
	return b.FailureV
}
