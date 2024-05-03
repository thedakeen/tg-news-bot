package models

import "time"

type Source struct {
	ID        int64
	Name      string
	FeedURL   string
	CreatedAt time.Time
}
