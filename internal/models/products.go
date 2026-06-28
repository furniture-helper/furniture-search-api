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
	Product                      Product `json:"product"`
	TitleSimilarity              string  `json:"title_similarity"`
	CosineSimilarity             string  `json:"cosine_similarity"`
	CombinedSimilarity           string  `json:"combined_similarity"`
	CosineSimilarityFinetuned768 string  `json:"cosine_similarity_finetuned_768"`
}

type MarkMatchingProductRequest struct {
	Url1       string `json:"url1"`
	Url2       string `json:"url2"`
	IsMatching bool   `json:"is_matching"`
}
