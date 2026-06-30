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
		SELECT url, product_title, product_price, product_image_url
		FROM page_inferred_labels
		WHERE url = $1
		LIMIT 1
	`

	var product models.Product
	if err := r.pool.QueryRow(ctx, query, url).Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Product{}, customerrors.NewProductNotFoundError(url)
		}

		return models.Product{}, fmt.Errorf("failed to query product by url: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) SearchByTitle(ctx context.Context, searchQuery string) ([]models.Product, error) {
	const query = `
		SELECT url, product_title, product_price, product_image_url
		FROM page_inferred_labels
		WHERE to_tsvector('simple', product_title) @@ plainto_tsquery('simple', $1)
			AND url NOT LIKE '%bigdeals.lk%'
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
		if err := rows.Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl); err != nil {
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

func (r *ProductRepository) GetSimilarProducts(ctx context.Context, url string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.SimilarProduct, error) {
	const query = `
		WITH input AS (
			SELECT
				pil.url AS input_url,
				pil.product_title AS input_title,
				pil.product_price AS input_price,
				lower(regexp_replace(substring(pil.url FROM '^(?:.*?://)?(?:[^@]+@)?([^:/?#]+)'), '^www\.', '')) AS input_domain,
				pe.embedding AS input_embedding,
				pe_finetuned_768.embedding AS input_embedding_finetuned_768
			FROM page_inferred_labels pil
					 LEFT JOIN products_embeddings_768 pe ON pe.url = pil.url
					 LEFT JOIN products_embeddings_finetuned_768 pe_finetuned_768 ON pe_finetuned_768.url = pil.url
			WHERE pil.url = $1  -- seed URL
		),
			 candidates AS (
				 SELECT
					 pil.url AS candidate_url,
					 pil.product_title AS candidate_title,
					 pil.product_price AS candidate_price,
					 pil.product_image_url AS candidate_image_url,	
					 lower(regexp_replace(substring(pil.url FROM '^(?:.*?://)?(?:[^@]+@)?([^:/?#]+)'), '^www\.', '')) AS candidate_domain,
					 pe.embedding AS candidate_embedding,
					 pe_finetuned_768.embedding AS candidate_embedding_finetuned_768
				 FROM page_inferred_labels pil
						  LEFT JOIN products_embeddings_768 pe ON pe.url = pil.url
						  LEFT JOIN products_embeddings_finetuned_768 pe_finetuned_768 ON pe_finetuned_768.url = pil.url
						  CROSS JOIN input i
				 WHERE pil.url <> i.input_url
				   AND pil.product_title IS NOT NULL
				   AND pil.product_price IS NOT NULL
				   AND lower(regexp_replace(substring(pil.url FROM '^(?:.*?://)?(?:[^@]+@)?([^:/?#]+)'), '^www\.', '')) <> i.input_domain
				 -- remove the line above if you want same-domain candidates too
			 ),
			 scored AS (
				 SELECT
					 i.input_url,
					 i.input_domain,
					 i.input_title,
					 i.input_price,
					 c.candidate_url,
					 c.candidate_domain,
					 c.candidate_title,
					 c.candidate_price,
					 c.candidate_image_url,
					 COALESCE(similarity(lower(i.input_title), lower(c.candidate_title)), 0) AS title_similarity,
					 COALESCE(1 - (c.candidate_embedding <=> i.input_embedding), 0) AS cosine_similarity_768,
					 COALESCE(1 - (c.candidate_embedding_finetuned_768 <=> i.input_embedding_finetuned_768), 0) AS cosine_similarity_finetuned_768
				 FROM candidates c
						  CROSS JOIN input i
			 )
		SELECT
			candidate_url,
			candidate_title,
			candidate_price,
			candidate_image_url,
			title_similarity,
			cosine_similarity_768,
			cosine_similarity_finetuned_768,
			(0.5 * title_similarity + 0.5 * cosine_similarity_768) AS combined_score
		FROM scored
		WHERE title_similarity >= $2 AND cosine_similarity_finetuned_768 >= $3
		ORDER BY cosine_similarity_finetuned_768 DESC, combined_score DESC
		LIMIT 100;
	`

	rows, err := r.pool.Query(ctx, query, url, titleSimilarityThreshold, cosineSimilarityThreshold)
	if err != nil {
		return []models.SimilarProduct{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var similarProducts []models.SimilarProduct
	for rows.Next() {
		var similarProduct models.SimilarProduct
		if err := rows.Scan(
			&similarProduct.Product.Url,
			&similarProduct.Product.Title,
			&similarProduct.Product.Price,
			&similarProduct.Product.ImageUrl,
			&similarProduct.TitleSimilarity,
			&similarProduct.CosineSimilarity,
			&similarProduct.CosineSimilarityFinetuned768,
			&similarProduct.CombinedSimilarity,
		); err != nil {
			return []models.SimilarProduct{}, fmt.Errorf("failed to scan similar product: %w", err)
		}
		similarProducts = append(similarProducts, similarProduct)
	}

	return similarProducts, nil
}

func (r *ProductRepository) MarkMatchingProduct(ctx context.Context, url1 string, url2 string, isMatching bool) error {
	const query = `
		INSERT INTO matching_products
		(url_1, url_2, matching)
		VALUES
		($1, $2, $3)
		ON CONFLICT (url_1, url_2) DO UPDATE SET matching = EXCLUDED.matching
	`

	_, err := r.pool.Exec(ctx, query, url1, url2, isMatching)
	if err != nil {
		return fmt.Errorf("failed to mark product: %w", err)
	}
	return nil
}

func (r *ProductRepository) GetRandomProduct(ctx context.Context) (models.Product, error) {
	const query = `
		WITH RandomDomain AS (
			SELECT domain
			FROM pages
			GROUP BY domain
			ORDER BY RANDOM()
			LIMIT 1
		)
		SELECT pil.url, pil.product_title, pil.product_price, pil.product_image_url
		FROM page_inferred_labels pil
				 JOIN pages p ON pil.url = p.url
		WHERE p.domain = (SELECT domain FROM RandomDomain)
			AND pil.product_title IS NOT NULL
			AND pil.product_price IS NOT NULL
		ORDER BY RANDOM()
		LIMIT 1;
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return models.Product{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl); err != nil {
			return models.Product{}, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return models.Product{}, fmt.Errorf("failed to scan products: %w", err)
	}

	return products[0], nil
}
