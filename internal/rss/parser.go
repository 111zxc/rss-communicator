package rss

import (
	"bytes"
	"time"

	"github.com/mmcdole/gofeed"
)

type ParsedItem struct {
	ExternalID  *string
	UniqKey     string
	Title       string
	Link        string
	Summary     *string
	Author      *string
	PublishedAt *time.Time
}

func Parse(data []byte) ([]ParsedItem, error) {
	fp := gofeed.NewParser()
	feed, err := fp.Parse(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	out := make([]ParsedItem, 0, len(feed.Items))
	for _, it := range feed.Items {
		var external *string
		if it.GUID != "" {
			v := it.GUID
			external = &v
		}

		title := it.Title
		link := it.Link
		if link == "" && it.Links != nil && len(it.Links) > 0 {
			link = it.Links[0]
		}

		var summary *string
		if it.Description != "" {
			v := it.Description
			summary = &v
		}

		var author *string
		if it.Author != nil && it.Author.Name != "" {
			v := it.Author.Name
			author = &v
		}

		var published *time.Time
		if it.PublishedParsed != nil {
			t := it.PublishedParsed.UTC()
			published = &t
		}

		uniq := ComputeUniqKey(title, link, published)

		out = append(out, ParsedItem{
			ExternalID:  external,
			UniqKey:     uniq,
			Title:       title,
			Link:        link,
			Summary:     summary,
			Author:      author,
			PublishedAt: published,
		})
	}
	return out, nil
}
