package handlers

import (
	"bytes"
	"encoding/json"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
	"net/http"
)

type ProductService interface {
	GetProductFromUrl(url string) (models.Product, error)
}

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetProductFromUrl("test")
	if err != nil {
		helpers.LogError("Failed to get product", r, err, nil)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(product); err != nil {
		helpers.LogError("Failed to encode product", r, err, nil)
		helpers.WriteJSONError(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
