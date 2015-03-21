package expvar

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Fetcher interface {
	Fetch() (Expvars, error)
}

type fetcher struct {
	url    string
	client *http.Client
}

func NewFetcher(url string) *fetcher {
	return &fetcher{
		url:    url,
		client: &http.Client{},
	}
}

func (f *fetcher) Fetch() (Expvars, error) {
	response, err := f.client.Get(f.url)
	if err != nil {
		return Expvars{}, fmt.Errorf("fetching expvar: %s", err)
	}

	var expvars Expvars
	if err = json.NewDecoder(response.Body).Decode(&expvars); err != nil {
		return Expvars{}, fmt.Errorf("decoding expvar: %s", err)
	}

	return expvars, nil
}
