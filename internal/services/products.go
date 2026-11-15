package services

import (
	"context"
	"fmt"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
	"net/url"
	"strings"
)

type ProductStore interface {
	GetByURL(ctx context.Context, url string) (models.Product, error)
	SearchByTitle(ctx context.Context, query string, titleSimilarityThreshold float64, limit int) ([]models.Product, error)
	GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error)
	GetSimilarProducts(ctx context.Context, url string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.SimilarProduct, error)
	GetSimilarProductsWithinSet(ctx context.Context, url string, candidateUrls []string, cosineSimilarityThreshold float64) ([]models.Product, error)
	MarkMatchingProduct(ctx context.Context, url1 string, url2 string, isMatching bool) error
	GetRandomProduct(ctx context.Context, domain *string) (models.Product, error)
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
	products, err := s.repository.SearchByTitle(ctx, searchQuery, 0, 25)
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

func (s *ProductService) GetRandomProduct(ctx context.Context, domain *string) (models.Product, error) {
	product, err := s.repository.GetRandomProduct(ctx, domain)
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

func (s *ProductService) SearchByTitleWithGrouping(ctx context.Context, searchQuery string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.ProductGroup, error) {
	candidateProducts, err := s.repository.SearchByTitle(ctx, searchQuery, titleSimilarityThreshold, 100)
	if err != nil {
		return []models.ProductGroup{}, fmt.Errorf("failed to search products by title: %w", err)
	}

	filteredCandidates := make([]models.Product, 0, len(candidateProducts))
	candidateUrls := make([]string, 0, len(candidateProducts))
	seenCandidateUrls := make(map[string]bool, len(candidateProducts))
	seenDomainTitle := make(map[string]bool, len(candidateProducts))
	for _, product := range candidateProducts {
		if product.Url == "" {
			continue
		}
		nonVariantUrl := helpers.GetNonVariantUrl(product.Url)
		if nonVariantUrl == "" {
			nonVariantUrl = product.Url
		}

		if seenCandidateUrls[nonVariantUrl] {
			continue
		}

		domainTitleKey := buildDomainTitleKey(product.Url, product.Title)
		if domainTitleKey != "" && seenDomainTitle[domainTitleKey] {
			continue
		}

		seenCandidateUrls[nonVariantUrl] = true
		if domainTitleKey != "" {
			seenDomainTitle[domainTitleKey] = true
		}
		filteredCandidates = append(filteredCandidates, product)
		candidateUrls = append(candidateUrls, product.Url)
	}

	clusteredUrls := make(map[string]bool, len(filteredCandidates))
	addedUrls := make(map[string]bool, len(filteredCandidates))
	productGroups := make([]models.ProductGroup, 0)

	for _, product := range filteredCandidates {
		if addedUrls[product.Url] {
			continue
		}

		nonVariantUrl := helpers.GetNonVariantUrl(product.Url)
		if nonVariantUrl == "" {
			nonVariantUrl = product.Url
		}

		if clusteredUrls[nonVariantUrl] {
			continue
		}

		productGroup := models.ProductGroup{
			Products: []models.Product{product},
		}
		addedUrls[product.Url] = true
		clusteredUrls[nonVariantUrl] = true

		similarProducts, err := s.repository.GetSimilarProductsWithinSet(ctx, product.Url, candidateUrls, cosineSimilarityThreshold)
		if err != nil {
			helpers.LogError("Failed to get similar products for URL: "+product.Url, ctx, err, map[string]any{"search_query": searchQuery})
		} else {
			for _, similarProduct := range similarProducts {
				if addedUrls[similarProduct.Url] {
					continue
				}

				nonVariantSimilarUrl := helpers.GetNonVariantUrl(similarProduct.Url)
				if nonVariantSimilarUrl == "" {
					nonVariantSimilarUrl = similarProduct.Url
				}

				if clusteredUrls[nonVariantSimilarUrl] {
					continue
				}

				addedUrls[similarProduct.Url] = true
				clusteredUrls[nonVariantSimilarUrl] = true
				productGroup.Products = append(productGroup.Products, similarProduct)
			}
		}

		productGroups = append(productGroups, productGroup)
	}
	return productGroups, nil
}

func buildDomainTitleKey(rawURL string, title string) string {
	normalizedTitle := strings.ToLower(strings.TrimSpace(title))
	if normalizedTitle == "" {
		return ""
	}

	parsedURL, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}

	domain := strings.ToLower(parsedURL.Hostname())
	domain = strings.TrimPrefix(domain, "www.")
	if domain == "" {
		return ""
	}

	return domain + "|" + normalizedTitle
}
