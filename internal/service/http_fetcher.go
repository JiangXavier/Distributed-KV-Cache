package service

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type httpFetcher struct {
	baseURL string
}

var _ Fetcher = (*httpFetcher)(nil)

// httpFetcher responsible for querying the value of key from the group cache of the specified node through http request
func (h *httpFetcher) Fetch(name string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(name), url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body failed: %v", err)
	}

	return bytes, err
}
