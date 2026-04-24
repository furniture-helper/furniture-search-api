package services

import "furniture-search-api/internal/models"

type ProductService struct{}

func NewProductService() *ProductService {
	return &ProductService{}
}

func (s *ProductService) GetProductFromUrl(url string) (models.Product, error) {
	product := models.Product{
		Url:   url,
		Title: "iPhone",
		Price: "1000.00",
	}
	return product, nil
}
