package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
	"net/http"
)

type ProductService interface {
	GetProductFromUrl(ctx context.Context, url string) (models.Product, error)
}

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetProductFromUrl(r.Context(), "test")
	if err != nil {
		helpers.LogError("Failed to get product", r.Context(), err, nil)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(product); err != nil {
		helpers.LogError("Failed to encode product", r.Context(), err, nil)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
