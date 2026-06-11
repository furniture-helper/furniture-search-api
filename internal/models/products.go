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

type SimilarProduct struct {
	Product            Product `json:"product"`
	TitleSimilarity    string  `json:"title_similarity"`
	CosineSimilarity   string  `json:"cosine_similarity"`
	CombinedSimilarity string  `json:"combined_similarity"`
}
