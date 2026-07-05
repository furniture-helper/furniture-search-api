package services

import (
	"context"
	"fmt"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
)

type ProductStore interface {
	GetByURL(ctx context.Context, url string) (models.Product, error)
	SearchByTitle(ctx context.Context, query string) ([]models.Product, error)
	GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error)
	GetSimilarProducts(ctx context.Context, url string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.SimilarProduct, error)
	MarkMatchingProduct(ctx context.Context, url1 string, url2 string, isMatching bool) error
	GetRandomProduct(ctx context.Context) (models.Product, error)
	GetProductMetadata(ctx context.Context, url string) (models.ProductMetadata, error)
	GetS3Key(ctx context.Context, url string) (string, error)
}

type PageService interface {
	GetCrawledPageUrl(ctx context.Context, url string) (string, error)
	GetMinimizedPageUrl(ctx context.Context, url string) (string, error)
}

type ProductService struct {
	repository  ProductStore
	pageService PageService
}

func NewProductService(repository ProductStore, pageService PageService) *ProductService {
	return &ProductService{
		repository:  repository,
		pageService: pageService,
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

func (s *ProductService) GetProductMetadata(ctx context.Context, url string) (models.ProductMetadata, error) {
	productMetadata, err := s.repository.GetProductMetadata(ctx, url)
	if err != nil {
		return models.ProductMetadata{}, err
	}

	return productMetadata, nil
}

func (s *ProductService) GetSourceCrawledPageUrl(ctx context.Context, url string) (string, error) {
	s3Key, err := s.repository.GetS3Key(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to get s3 key for url %s: %w", url, err)
	}

	if s3Key == "" {
		return "", fmt.Errorf("no s3 key found for url %s", url)
	}

	sourceUrl, err := s.pageService.GetCrawledPageUrl(ctx, s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to get source crawled page url: %w", err)
	}
	return sourceUrl, nil
}

func (s *ProductService) GetSourceMinimizedPageUrl(ctx context.Context, url string) (string, error) {
	s3Key, err := s.repository.GetS3Key(ctx, url)
	helpers.LogInfo(fmt.Sprintf("Retrieved S3 key for url %s: %s", url, s3Key), ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get s3 key for url %s: %w", url, err)
	}

	if s3Key == "" {
		return "", fmt.Errorf("no s3 key found for url %s", url)
	}

	minimizedUrl, err := s.pageService.GetMinimizedPageUrl(ctx, s3Key)
	if err != nil {
		return "", fmt.Errorf("failed to get source minimized page url: %w", err)
	}
	return minimizedUrl, nil
}
