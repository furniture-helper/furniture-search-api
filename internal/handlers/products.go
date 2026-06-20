package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	customerrors "furniture-search-api/internal/errors"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
	"net/http"
	"strconv"
)

type ProductService interface {
	GetFromUrl(ctx context.Context, url string) (models.Product, error)
	SearchByTitle(ctx context.Context, searchQuery string) ([]models.Product, error)
	GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error)
	GetSimilarProducts(ctx context.Context, url string, titleSimilarityThreshold float64, cosineSimilarityThreshold float64) ([]models.SimilarProduct, error)
	MarkMatchingProduct(ctx context.Context, url1 string, url2 string, isMatching bool) error
	GetRandomProduct(ctx context.Context) (models.Product, error)
}

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		helpers.LogInfo("Missing url query parameter", r.Context(), nil)
		helpers.WriteJSONErrorResponse(w, http.StatusBadRequest, "Missing url query parameter")
		return
	}

	product, err := h.service.GetFromUrl(r.Context(), url)
	if err != nil {

		var notFoundErr *customerrors.ProductNotFoundError
		if errors.As(err, &notFoundErr) {
			helpers.LogInfo(fmt.Sprintf("Product with url \"%s\" not found", url), r.Context(), nil)
			helpers.WriteJSONErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Product with url \"%s\" not found", url))
			return
		}

		helpers.LogError("Failed to retrieve product", r.Context(), err, map[string]interface{}{"url": url})
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, product); err != nil {
		helpers.LogError("Failed to encode product", r.Context(), err, nil)
		return
	}
}

func (h *ProductHandler) SearchByTitle(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("query")
	if searchQuery == "" {
		helpers.LogInfo("Missing search query parameter", r.Context(), nil)
	}

	products, err := h.service.SearchByTitle(r.Context(), searchQuery)
	if err != nil {
		helpers.LogError("Failed to retrieve search results", r.Context(), err, map[string]interface{}{"query": searchQuery})
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, products); err != nil {
		helpers.LogError("Failed to encode products", r.Context(), err, nil)
		return
	}
}

func (h *ProductHandler) GetPriceHistory(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		helpers.LogInfo("Missing url query parameter", r.Context(), nil)
	}

	priceHistory, err := h.service.GetPriceHistory(r.Context(), url)
	if err != nil {
		helpers.LogError("Failed to retrieve price history", r.Context(), err, map[string]interface{}{"url": url})
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, priceHistory); err != nil {
		helpers.LogError("Failed to encode products", r.Context(), err, nil)
	}
}

func (h *ProductHandler) GetSimilarProducts(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		helpers.LogInfo("Missing url query parameter", r.Context(), nil)
		helpers.WriteJSONErrorResponse(w, http.StatusBadRequest, "Missing url query parameter")
		return
	}

	titleSimilarityThreshold := r.URL.Query().Get("title_similarity_threshold")
	if titleSimilarityThreshold == "" {
		titleSimilarityThreshold = "0.5"
	}

	cosineSimilarityThreshold := r.URL.Query().Get("cosine_similarity_threshold")
	if cosineSimilarityThreshold == "" {
		cosineSimilarityThreshold = "0.75"
	}

	titleSimilarityThresholdFloat, err := strconv.ParseFloat(titleSimilarityThreshold, 64)
	if err != nil {
		helpers.LogInfo("Invalid title similarity threshold", r.Context(), nil)
		helpers.WriteJSONErrorResponse(w, http.StatusBadRequest, "Invalid title similarity threshold")
		return
	}

	cosineSimilarityThresholdFloat, err := strconv.ParseFloat(cosineSimilarityThreshold, 64)
	if err != nil {
		helpers.LogInfo("Invalid cosine similarity threshold", r.Context(), nil)
		helpers.WriteJSONErrorResponse(w, http.StatusBadRequest, "Invalid cosine similarity threshold")
		return
	}

	similarProducts, err := h.service.GetSimilarProducts(r.Context(), url, titleSimilarityThresholdFloat, cosineSimilarityThresholdFloat)
	if err != nil {
		helpers.LogError("Failed to retrieve similar products", r.Context(), err, map[string]interface{}{"url": url})
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, similarProducts); err != nil {
		helpers.LogError("Failed to encode products", r.Context(), err, nil)
		return
	}
}

func (h *ProductHandler) MarkMatchingProduct(w http.ResponseWriter, r *http.Request) {
	var req models.MarkMatchingProductRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		helpers.LogInfo("Invalid JSON payload", r.Context(), nil)
		helpers.WriteJSONErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err = h.service.MarkMatchingProduct(r.Context(), req.Url1, req.Url2, req.IsMatching)
	if err != nil {
		helpers.LogError("Failed to mark matching product", r.Context(), err, map[string]interface{}{"url1": req.Url1, "url2": req.Url2})
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, nil); err != nil {
		helpers.LogError("Failed to mark matching product", r.Context(), err, nil)
		return
	}
}

func (h *ProductHandler) GetRandomProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetRandomProduct(r.Context())
	if err != nil {
		helpers.LogError("Failed to retrieve random product", r.Context(), err, nil)
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, product); err != nil {
		helpers.LogError("Failed to encode product", r.Context(), err, nil)
		return
	}
}
