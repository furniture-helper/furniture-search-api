package repositories

import (
	"context"
	"fmt"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"

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
		//if errors.Is(err, pgx.ErrNoRows) {
		//	helpers.LogInfo("Product not found", ctx, map[string]any{"url": url})
		//	return models.Product{}, customerrors.NewProductNotFoundError(url)
		//}

		helpers.LogError("Failed to query product by url", ctx, err, map[string]any{"url": url})
		return models.Product{}, fmt.Errorf("failed to query product by url: %w", err)
	}

	return product, nil
}
