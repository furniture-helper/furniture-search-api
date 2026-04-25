package handlers

import (
	"context"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/models"
	"net/http"
)

type ProductService interface {
	GetFromUrl(ctx context.Context, url string) (models.Product, error)
}

type ProductHandler struct {
	service ProductService
}

func NewProductHandler(service ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetFromUrl(r.Context(), "https://ugreen.lk/product/ugreen-24w-dual-usb-car-charger-cd130/")
	if err != nil {
		helpers.LogError("Failed to get product from service", r.Context(), err, nil)
		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, product); err != nil {
		helpers.LogError("Failed to encode product", r.Context(), err, nil)
		return
	}
}
