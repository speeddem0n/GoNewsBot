package models

import "time"

type Item struct { // RSS item
	Title      string
	Categories []string
	Link       string
	Date       time.Time
	Summary    string
	SourceName string
}
