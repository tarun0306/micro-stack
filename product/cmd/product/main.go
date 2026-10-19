package main

import (
	"log"
	"os"
	"time"

	"github.com/tarun0306/go-grpc-graphql-microservice/product"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {
	dburl, ifexists := os.LookupEnv("DATABASE_URL")
	if !ifexists {
		log.Fatal("DATABASE_URL is not set")
	}
	var repo product.Repository

	for {
		var err error
		repo, err = product.NewElasticRepository(dburl)
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
