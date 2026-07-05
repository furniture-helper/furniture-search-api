package main

import (
	"context"
	"furniture-search-api/internal/config"
	"furniture-search-api/internal/handlers"
	"furniture-search-api/internal/helpers"
	"furniture-search-api/internal/middleware"
	"furniture-search-api/internal/repositories"
	"furniture-search-api/internal/services"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

func getRouter() *mux.Router {
	r := mux.NewRouter()

	r.Use(middleware.RequestIdMiddleware)
	r.Use(middleware.LoggingMiddleware)

	dbConnectionPool, err := repositories.NewPostgresPool(context.Background(), config.GetDatabaseConfig(context.Background()))
	if err != nil {
		panic(err)
	}

	productStore := repositories.NewProductRepository(dbConnectionPool)
	pageService := services.NewS3Service()

	productService := services.NewProductService(productStore, pageService)
	productHandler := handlers.NewProductHandler(productService)
	r.HandleFunc("/products", productHandler.GetProduct).Methods(http.MethodGet)
	r.HandleFunc("/products/search", productHandler.SearchByTitle).Methods(http.MethodGet)
	r.HandleFunc("/products/price-history", productHandler.GetPriceHistory).Methods(http.MethodGet)
	r.HandleFunc("/products/similar", productHandler.GetSimilarProducts).Methods(http.MethodGet)
	r.HandleFunc("/products/mark-matching", productHandler.MarkMatchingProduct).Methods(http.MethodPost)
	r.HandleFunc("/products/random", productHandler.GetRandomProduct).Methods(http.MethodGet)
	r.HandleFunc("/products/metadata", productHandler.GetProductMetadata).Methods(http.MethodGet)
	r.HandleFunc("/products/source-crawled-page", productHandler.GetSourceCrawledPageUrl).Methods(http.MethodGet)
	r.HandleFunc("/products/source-minimized-page", productHandler.GetSourceMinimizedPageUrl).Methods(http.MethodGet)

	return r
}

func newLambdaHandler(mux http.Handler) func(context.Context, events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	adapter := httpadapter.NewV2(mux)

	return func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return adapter.ProxyWithContext(ctx, req)
	}
}

func main() {
	helpers.InitLogger()
	router := getRouter()

	allowedOrigins := config.GetCorsAllowedOrigins()
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{http.MethodGet, http.MethodPost},
		AllowedHeaders: []string{},
	})

	handler := c.Handler(router)

	if config.IsLocal() {
		err := http.ListenAndServe(":8080", handler)
		if err != nil {
			panic(err)
		}

		return
	}

	lambdaHandler := newLambdaHandler(handler)
	lambda.Start(lambdaHandler)
}
