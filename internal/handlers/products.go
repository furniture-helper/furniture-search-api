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
		helpers.LogError("Missing url query parameter", r.Context(), nil, nil)
		helpers.WriteJSONErrorResponse(w, http.StatusBadRequest, "Missing url query parameter")
		return
	}

	product, err := h.service.GetFromUrl(r.Context(), url)
	if err != nil {

		var notFoundErr *customerrors.ProductNotFoundError
		if errors.As(err, &notFoundErr) {
			helpers.WriteJSONErrorResponse(w, http.StatusNotFound, fmt.Sprintf("Product with url \"%s\" not found", url))
			return
		}

		helpers.WriteJSONErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	if err := helpers.WriteJSONResponse(w, http.StatusOK, product); err != nil {
		helpers.LogError("Failed to encode product", r.Context(), err, nil)
		return
	}
}
