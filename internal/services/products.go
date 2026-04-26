package services

import (
	"context"
	"furniture-search-api/internal/models"
)

type ProductStore interface {
	GetByURL(ctx context.Context, url string) (models.Product, error)
	SearchByTitle(ctx context.Context, query string) ([]models.Product, error)
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
