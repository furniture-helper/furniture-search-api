package customerrors

import "fmt"

type ProductNotFoundError struct {
	url string
}

func (e ProductNotFoundError) Error() string {
	return fmt.Sprintf("Product with URL \"%s\" not found", e.url)
}

func NewProductNotFoundError(url string) *ProductNotFoundError {
	return &ProductNotFoundError{url: url}
}
