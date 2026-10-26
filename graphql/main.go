package main

import (
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/tarun0306/go-grpc-graphql-microservice/internal/logging"
)

type AppConfig struct {
	AccountURL string
	CatalogURL string
	OrderURL   string
}

func main() {

	var cfg AppConfig

	cfg.AccountURL = os.Getenv("ACCOUNT_SERVICES_URL")
	cfg.CatalogURL = os.Getenv("PRODUCT_SERVICES_URL")
	cfg.OrderURL = os.Getenv("ORDER_SERVICES_URL")

	s, err := NewGraphQLServer(cfg.AccountURL, cfg.CatalogURL, cfg.OrderURL, os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/graphql", logging.HTTPMiddleware("graphql", authMiddleware(s.jwtSecret, handler.NewDefaultServer(s.ToExecutableSchema()))))
	http.Handle("/playground", logging.HTTPMiddleware("graphql", playground.Handler("tarun", "/graphql")))

	log.Fatal(http.ListenAndServe(":8080", nil))

}
