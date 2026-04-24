package main

import (
	"context"
	"furniture-search-api/internal/handlers"
	"furniture-search-api/internal/services"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

func getMux() *http.ServeMux {
	mux := http.NewServeMux()

	productService := services.NewProductService()
	productHandler := handlers.NewProductHandler(productService)
	mux.HandleFunc("GET /products", productHandler.GetProduct)

	return mux
}

func newLambdaHandler(mux http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	adapter := httpadapter.NewV2(mux)

	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return adapter.ProxyWithContext(ctx, req)
	}
}

func main() {
	mux := getMux()

	if os.Getenv("LOCAL") == "1" {
		err := http.ListenAndServe(":8080", mux)
		if err != nil {
			panic(err)
		}

		return
	}

	lambdaHandler := newLambdaHandler(mux)
	lambda.Start(lambdaHandler)
}
