package models

import "time"

type Product struct {
	Url      string  `json:"url"`
	Title    string  `json:"title"`
	Price    string  `json:"price"`
	ImageUrl *string `json:"image_url"`
	InStock  *bool   `json:"in_stock"`
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

type ProductMetadata struct {
	Url              string     `json:"url"`
	S3Key            *string    `json:"s3_key"`
	DetectedAt       *time.Time `json:"detected_at"`
	LastCrawledAt    *time.Time `json:"last_crawled_at"`
	LastMinimizedAt  *time.Time `json:"last_minimized_at"`
	LastClassifiedAt *time.Time `json:"last_classified_at"`
	LastInferredAt   *time.Time `json:"last_inferred_at"`
}

type ProductSourcePages struct {
	CrawledPage   *string `json:"crawled_page"`
	MinimizedPage *string `json:"minimized_page"`
}

type SourcePage struct {
	Url string `json:"url"`
}
