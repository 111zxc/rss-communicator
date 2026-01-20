package rss

import (
	"context"
	"io"
	"net/http"
	"time"
)

type FetchResult struct {
	StatusCode   int
	Body         []byte
	ETag         *string
	LastModified *string
	NotModified  bool
}

type Fetcher struct {
	client *http.Client
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, url string, etag, lastModified *string) (FetchResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return FetchResult{}, err
	}
	req.Header.Set("User-Agent", "rss-communicator/0.1 (+local)")

	if etag != nil && *etag != "" {
		req.Header.Set("If-None-Match", *etag)
	}
	if lastModified != nil && *lastModified != "" {
		req.Header.Set("If-Modified-Since", *lastModified)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return FetchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		return FetchResult{StatusCode: resp.StatusCode, NotModified: true}, nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return FetchResult{}, err
	}

	var outETag *string
	if v := resp.Header.Get("ETag"); v != "" {
		outETag = &v
	}
	var outLM *string
	if v := resp.Header.Get("Last-Modified"); v != "" {
		outLM = &v
	}

	return FetchResult{
		StatusCode:   resp.StatusCode,
		Body:         b,
		ETag:         outETag,
		LastModified: outLM,
	}, nil
}
