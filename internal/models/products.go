package models

import "time"

type Product struct {
	Url   string `json:"url"`
	Title string `json:"title"`
	Price string `json:"price"`
}

type PriceHistoryEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Price     string    `json:"price"`
}
