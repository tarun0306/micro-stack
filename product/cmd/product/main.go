package main

import (
	"log"
	"os"
	"time"

	"github.com/tarun0306/go-grpc-graphql-microservice/product"
)

type Config struct {
	OpenSearchURL string `envconfig:"OPENSEARCH_URL"`
}

func main() {
	openSearchURL, ifexists := os.LookupEnv("OPENSEARCH_URL")
	if !ifexists {
		log.Fatal("OPENSEARCH_URL is not set")
	}
	var repo product.Repository

	for {
		var err error
		repo, err = product.NewOpenSearchRepository(openSearchURL)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	log.Println("Listening on Port 8081...")
	s := product.NewService(repo)
	log.Fatal(product.ListenGrpc(s, 8081))

}
