package handlers

import (
	"context"
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
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, product); err != nil {
		helpers.LogError("Failed to encode product", r.Context(), err, nil)
		return
	}
}
