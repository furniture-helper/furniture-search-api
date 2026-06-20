package services

import (
	"context"
	"furniture-search-api/internal/models"
)

type ProductStore interface {
	GetByURL(ctx context.Context, url string) (models.Product, error)
	SearchByTitle(ctx context.Context, query string) ([]models.Product, error)
	GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error)
	GetSimilarProducts(ctx context.Context, url string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.SimilarProduct, error)
	MarkMatchingProduct(ctx context.Context, url1 string, url2 string, isMatching bool) error
	GetRandomProduct(ctx context.Context) (models.Product, error)
}

type ProductService struct {
	repository ProductStore
}

func NewProductService(repository ProductStore) *ProductService {
	return &ProductService{
		repository: repository,
	}
}

func (s *ProductService) GetFromUrl(ctx context.Context, url string) (models.Product, error) {
	product, err := s.repository.GetByURL(ctx, url)
	if err != nil {
		return models.Product{}, err
	}
	return product, nil
}

func (s *ProductService) SearchByTitle(ctx context.Context, searchQuery string) ([]models.Product, error) {
	products, err := s.repository.SearchByTitle(ctx, searchQuery)
	if err != nil {
		return []models.Product{}, err
	}
	return products, nil
}

func (s *ProductService) GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error) {
	priceHistory, err := s.repository.GetPriceHistory(ctx, url)
	if err != nil {
		return []models.PriceHistoryEntry{}, err
	}

	return priceHistory, nil
}

func (s *ProductService) GetSimilarProducts(ctx context.Context, url string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.SimilarProduct, error) {
	similarProducts, err := s.repository.GetSimilarProducts(ctx, url, titleSimilarityThreshold, cosineSimilarityThreshold)
	if err != nil {
		return []models.SimilarProduct{}, err
	}
	return similarProducts, nil
}

func (s *ProductService) MarkMatchingProduct(ctx context.Context, url1 string, url2 string, isMatching bool) error {
	err := s.repository.MarkMatchingProduct(ctx, url1, url2, isMatching)
	if err != nil {
		return err
	}
	return nil
}

func (s *ProductService) GetRandomProduct(ctx context.Context) (models.Product, error) {
	product, err := s.repository.GetRandomProduct(ctx)
	if err != nil {
		return models.Product{}, err
	}

	return product, nil
}
