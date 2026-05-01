package handlers

import (
	"context"
	"errors"
	"fmt"
	customerrors "furniture-search-api/internal/errors"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
	"net/http"
)

type ProductService interface {
	GetFromUrl(ctx context.Context, url string) (models.Product, error)
	SearchByTitle(ctx context.Context, searchQuery string) ([]models.Product, error)
	GetPriceHistory(ctx context.Context, url string) ([]models.PriceHistoryEntry, error)
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
