package models

import "time"

type Source struct {
	ID      int64
	Name    string
	FeedURL string
	Created time.Time
}
