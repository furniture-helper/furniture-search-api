package services

import (
	"context"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
)

type ProductService struct{}

func NewProductService() *ProductService {
	return &ProductService{}
}

func (s *ProductService) GetProductFromUrl(ctx context.Context, url string) (models.Product, error) {
	product := models.Product{
		Url:   url,
		Title: "iPhone",
		Price: "1000.00",
	}
	helpers.LogDebug("Inside Product Service", nil, nil)
	return product, nil
}
