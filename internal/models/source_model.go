package models

import "time"

type Source struct { // Стркутура Source для источников
	ID      int64
	Name    string
	FeedURL string
	Created time.Time
}
