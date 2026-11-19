package repositories

import (
	"context"
	"errors"
	"fmt"
	customerrors "furniture-search-api/internal/errors"
	"furniture-search-api/internal/models"
	"time"

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
		SELECT url, product_title, product_price, product_image_url, in_stock
		FROM page_inferred_labels
		WHERE url = $1
		LIMIT 1
	`

	var product models.Product
	if err := r.pool.QueryRow(ctx, query, url).Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl, &product.InStock); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Product{}, customerrors.NewProductNotFoundError(url)
		}

		return models.Product{}, fmt.Errorf("failed to query product by url: %w", err)
	}

	return product, nil
}

func (r *ProductRepository) SearchByTitle(ctx context.Context, searchQuery string, titleSimilarityThreshold float64, limit int) ([]models.Product, error) {
	const query = `
		WITH scored AS (
			SELECT
				pil.url,
				pil.product_title,
				pil.product_price,
				pil.product_image_url,
				pil.in_stock,
				COALESCE(similarity(lower(pil.product_title), lower($1)), 0) AS title_similarity
			FROM page_inferred_labels pil
			INNER JOIN page_classifications pc ON pc.url = pil.url
			WHERE to_tsvector('simple', pil.product_title) @@ plainto_tsquery('simple', $1)
				AND pil.product_title IS NOT NULL
				AND pil.url NOT LIKE '%bigdeals.lk%'
				AND pc.type = 'product'
				AND pil.last_inferred_at >= CURRENT_TIMESTAMP - INTERVAL '14 days'
		)
		SELECT url, product_title, product_price, product_image_url, in_stock
		FROM scored
		WHERE title_similarity >= $2
		ORDER BY
			CASE WHEN $2 > 0 THEN title_similarity ELSE NULL END DESC,
			product_title <-> $1 ASC
		LIMIT $3;
	`

	products := make([]models.Product, 0)
	rows, err := r.pool.Query(ctx, query, searchQuery, titleSimilarityThreshold, limit)
	if err != nil {
		return []models.Product{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl, &product.InStock); err != nil {
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
		WHERE url = $1 AND price IS NOT NULL AND price > 1
		ORDER BY recorded_at DESC
	`

	priceHistory := make([]models.PriceHistoryEntry, 0)
	rows, err := r.pool.Query(ctx, query, url)
	if err != nil {
		return []models.PriceHistoryEntry{}, fmt.Errorf("failed to query price history: %w", err)
	}
	defer rows.Close()

	var lastKeptDate time.Time
	for rows.Next() {
		var entry models.PriceHistoryEntry
		if err := rows.Scan(&entry.Price, &entry.Timestamp); err != nil {
			return []models.PriceHistoryEntry{}, fmt.Errorf("failed to query price history: %w", err)
		}

		// Rows are newest first, so keep an entry only when it is at least 48 hours
		// away from the newest entry already kept.
		if !lastKeptDate.IsZero() && lastKeptDate.Sub(entry.Timestamp) < 48*time.Hour {
			continue
		}

		priceHistory = append(priceHistory, entry)
		lastKeptDate = entry.Timestamp
	}

	if err := rows.Err(); err != nil {
		return []models.PriceHistoryEntry{}, fmt.Errorf("failed to read price history: %w", err)
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
				pe_finetuned_768_2.embedding AS input_embedding_finetuned_768_2
			FROM page_inferred_labels pil
				 LEFT JOIN products_embeddings_finetuned_768_2 pe_finetuned_768_2 ON pe_finetuned_768_2.url = pil.url
			WHERE pil.url = $1  -- seed URL
		),
			 candidates AS (
				 SELECT
					 pil.url AS candidate_url,
					 pil.product_title AS candidate_title,
					 pil.product_price AS candidate_price,
					 pil.product_image_url AS candidate_image_url,
					 pil.in_stock AS candidate_in_stock,
					 lower(regexp_replace(substring(pil.url FROM '^(?:.*?://)?(?:[^@]+@)?([^:/?#]+)'), '^www\.', '')) AS candidate_domain,
					 pe_finetuned_768_2.embedding AS candidate_embedding_finetuned_768_2
				 FROM page_inferred_labels pil
						  LEFT JOIN products_embeddings_finetuned_768_2 pe_finetuned_768_2 ON pe_finetuned_768_2.url = pil.url
						  CROSS JOIN input i
				 WHERE pil.url <> i.input_url
				   AND pil.product_title IS NOT NULL
				   AND pil.product_price IS NOT NULL
				   AND pil.last_inferred_at >= CURRENT_TIMESTAMP - INTERVAL '14 days'
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
					 c.candidate_in_stock,
					 COALESCE(similarity(lower(i.input_title), lower(c.candidate_title)), 0) AS title_similarity,
					 COALESCE(1 - (c.candidate_embedding_finetuned_768_2 <=> i.input_embedding_finetuned_768_2), 0) AS cosine_similarity_finetuned_768
				 FROM candidates c
						  CROSS JOIN input i
			 )
		SELECT
			candidate_url,
			candidate_title,
			candidate_price,
			candidate_image_url,
            candidate_in_stock,
			title_similarity,
			cosine_similarity_finetuned_768,
			(0.5 * title_similarity + 0.5 * cosine_similarity_finetuned_768) AS combined_score
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
			&similarProduct.Product.InStock,
			&similarProduct.TitleSimilarity,
			&similarProduct.CosineSimilarityFinetuned768,
			&similarProduct.CombinedSimilarity,
		); err != nil {
			return []models.SimilarProduct{}, fmt.Errorf("failed to scan similar product: %w", err)
		}
		similarProducts = append(similarProducts, similarProduct)
	}

	return similarProducts, nil
}

func (r *ProductRepository) GetSimilarProductsWithinSet(ctx context.Context, url string, candidateUrls []string, cosineSimilarityThreshold float64) ([]models.Product, error) {
	if len(candidateUrls) == 0 {
		return []models.Product{}, nil
	}

	const query = `
		WITH input AS (
			SELECT pe.embedding AS input_embedding
			FROM products_embeddings_finetuned_768_2 pe
			WHERE pe.url = $1
			LIMIT 1
		),
		candidates AS (
			SELECT
				pil.url,
				pil.product_title,
				pil.product_price,
				pil.product_image_url,
				pil.in_stock,
				pe.embedding AS candidate_embedding
			FROM page_inferred_labels pil
			INNER JOIN products_embeddings_finetuned_768_2 pe ON pe.url = pil.url
			WHERE pil.url = ANY($2::text[])
				AND pil.url <> $1
				AND pil.product_title IS NOT NULL
				AND pil.last_inferred_at >= CURRENT_TIMESTAMP - INTERVAL '14 days'
		)
		SELECT
			c.url,
			c.product_title,
			c.product_price,
			c.product_image_url,
			c.in_stock
		FROM candidates c
		CROSS JOIN input i
		WHERE (1 - (c.candidate_embedding <=> i.input_embedding)) >= $3
		ORDER BY (1 - (c.candidate_embedding <=> i.input_embedding)) DESC;
	`

	products := make([]models.Product, 0)
	rows, err := r.pool.Query(ctx, query, url, candidateUrls, cosineSimilarityThreshold)
	if err != nil {
		return []models.Product{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl, &product.InStock); err != nil {
			return []models.Product{}, fmt.Errorf("failed to query products: %w", err)
		}
		products = append(products, product)
	}

	return products, nil
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

func (r *ProductRepository) GetRandomProduct(ctx context.Context, domain *string) (models.Product, error) {
	var rows pgx.Rows
	var err error
	if domain != nil && *domain != "" {
		const query = `
			SELECT pil.url, pil.product_title, pil.product_price, pil.product_image_url, pil.in_stock
			FROM page_inferred_labels pil
					 INNER JOIN pages p ON pil.url = p.url
					 INNER JOIN page_classifications pc ON pil.url = pc.url
			WHERE p.domain = $1
				AND pil.product_title IS NOT NULL
				AND pil.product_price IS NOT NULL
				AND pc.type = 'product'
				AND pil.last_inferred_at >= CURRENT_TIMESTAMP - INTERVAL '14 days'
			ORDER BY RANDOM()
			LIMIT 1;
		`
		rows, err = r.pool.Query(ctx, query, *domain)
	} else {
		const query = `
			WITH RandomDomain AS (
				SELECT p.domain
				FROM pages p
				INNER JOIN page_inferred_labels pil ON pil.url = p.url
				INNER JOIN page_classifications pc ON pc.url = p.url
				WHERE pil.product_title IS NOT NULL
					AND pil.product_price IS NOT NULL
					AND pc.type = 'product'
					AND pil.last_inferred_at >= CURRENT_TIMESTAMP - INTERVAL '14 days'
				GROUP BY p.domain
				ORDER BY RANDOM()
				LIMIT 1
			)
			SELECT pil.url, pil.product_title, pil.product_price, pil.product_image_url, pil.in_stock
			FROM page_inferred_labels pil
					 INNER JOIN pages p ON pil.url = p.url
					 INNER JOIN page_classifications pc ON pil.url = pc.url
			WHERE p.domain = (SELECT domain FROM RandomDomain)
				AND pil.product_title IS NOT NULL
				AND pil.product_price IS NOT NULL
				AND pc.type = 'product'
				AND pil.last_inferred_at >= CURRENT_TIMESTAMP - INTERVAL '14 days'
			ORDER BY RANDOM()
			LIMIT 1;
		`
		rows, err = r.pool.Query(ctx, query)
	}

	if err != nil {
		return models.Product{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var product models.Product
		if err := rows.Scan(&product.Url, &product.Title, &product.Price, &product.ImageUrl, &product.InStock); err != nil {
			return models.Product{}, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return models.Product{}, fmt.Errorf("failed to scan products: %w", err)
	}

	return products[0], nil
}

func (r *ProductRepository) GetProductMetadata(ctx context.Context, url string) (models.ProductMetadata, error) {
	const query = `
	SELECT
		pages.url,
		pages.s3_key,
		pages.created_at AS detected_at,
		pages.last_crawled_at,
		minimized_pages.last_minimized_at,
		page_classifications.last_classified_at,
		page_inferred_labels.last_inferred_at
	FROM pages
		 LEFT JOIN minimized_pages on minimized_pages.url = pages.url
		 LEFT JOIN page_classifications on page_classifications.url = pages.url
		 LEFT JOIN page_inferred_labels on page_inferred_labels.url = pages.url
	WHERE pages.url = $1
	LIMIT 1;
	`

	rows, err := r.pool.Query(ctx, query, url)
	if err != nil {
		return models.ProductMetadata{}, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var productMetadata models.ProductMetadata
	for rows.Next() {
		if err := rows.Scan(
			&productMetadata.Url,
			&productMetadata.S3Key,
			&productMetadata.DetectedAt,
			&productMetadata.LastCrawledAt,
			&productMetadata.LastMinimizedAt,
			&productMetadata.LastClassifiedAt,
			&productMetadata.LastInferredAt,
		); err != nil {
			return models.ProductMetadata{}, fmt.Errorf("failed to scan product metadata: %w", err)
		}
	}

	return productMetadata, nil
}

func (r *ProductRepository) GetS3Key(ctx context.Context, url string) (string, error) {
	const get_s3_key_query = `
		SELECT s3_key
		FROM pages
		WHERE url = $1
	`

	rows, err := r.pool.Query(ctx, get_s3_key_query, url)
	if err != nil {
		return "", fmt.Errorf("failed to get s3_key: %w", err)
	}
	defer rows.Close()

	var s3_key string
	for rows.Next() {
		if err := rows.Scan(&s3_key); err != nil {
			return "", fmt.Errorf("failed to get s3_key: %w", err)
		}
	}

	return s3_key, nil
}
