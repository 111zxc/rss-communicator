package rss

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

func ComputeUniqKey(title, link string, published *time.Time) string {
	pub := ""
	if published != nil {
		pub = published.Format(time.RFC3339)
	}
	h := sha256.Sum256([]byte(title + "\n" + link + "\n" + pub))
	return hex.EncodeToString(h[:])
}
