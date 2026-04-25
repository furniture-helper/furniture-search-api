package services

import (
	"context"
	"fmt"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
)

type ProductStore interface {
	GetByURL(ctx context.Context, url string) (models.Product, error)
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
		helpers.LogError("Failed to get product from repository", ctx, err, map[string]any{"url": url})
		return models.Product{}, fmt.Errorf("failed to retrieve product by url: %w", err)
	}
	return product, nil
}
