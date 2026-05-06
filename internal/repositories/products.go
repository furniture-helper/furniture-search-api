package repositories

import (
	"context"
	"errors"
	"fmt"
	customerrors "furniture-search-api/internal/errors"
	"furniture-search-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepository struct {
	pool *pgxpool.Pool
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

func (r *ProductRepository) GetByURL(ctx context.Context, url string) (models.Product, error) {
	const query = `
		SELECT url, product_title, product_price
		FROM page_inferred_labels
		WHERE url = $1
		LIMIT 1
	`

	var product models.Product
	if err := r.pool.QueryRow(ctx, query, url).Scan(&product.Url, &product.Title, &product.Price); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Product{}, customerrors.NewProductNotFoundError(url)
		}

		return models.Product{}, fmt.Errorf("failed to query product by url: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) SearchByTitle(ctx context.Context, searchQuery string) ([]models.Product, error) {
	const query = `
		SELECT url, product_title, product_price 
		FROM page_inferred_labels
		WHERE to_tsvector('simple', product_title) @@ plainto_tsquery('simple', $1)
		ORDER BY product_title <-> $1 ASC
		LIMIT 25;
	`

	products := make([]models.Product, 0)
	rows, err := r.pool.Query(ctx, query, searchQuery)
	if err != nil {
		return []models.Product{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.Url, &product.Title, &product.Price); err != nil {
			return []models.Product{}, fmt.Errorf("failed to query products: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
}

func (r *ProductRepository) GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error) {
	const query = `
		SELECT price, recorded_at 
		FROM product_price_history 
		WHERE url = $1 AND price IS NOT NULL
		ORDER BY recorded_at DESC
	`

	priceHistory := make([]models.PriceHistoryEntry, 0)
	rows, err := r.pool.Query(ctx, query, url)
	if err != nil {
		return []models.PriceHistoryEntry{}, fmt.Errorf("failed to query price history: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var entry models.PriceHistoryEntry
		if err := rows.Scan(&entry.Price, &entry.Timestamp); err != nil {
			return []models.PriceHistoryEntry{}, fmt.Errorf("failed to query price history: %w", err)
		}
		priceHistory = append(priceHistory, entry)
	}

	return priceHistory, nil
}
