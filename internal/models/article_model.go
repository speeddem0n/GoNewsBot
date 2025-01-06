package models

import "time"

type Article struct { // Стркутура Article для статей
	ID        int64
	SourceID  int64
	Title     string
	Link      string
	Summary   string
	Published time.Time
	Posted    time.Time
	Created   time.Time
}
