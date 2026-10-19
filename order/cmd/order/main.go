package main

import (
	"log"
	"os"
	"time"

	"github.com/tarun0306/go-grpc-graphql-microservice/order"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
	AccountURL  string `envconfig:"ACCOUNT_SERVICES_URL"`
	ProductURL  string `envconfig:"PRODUCT_SERVICES_URL"`
}

func main() {
	dbURL, ok := os.LookupEnv("DATABASE_URL")
	if !ok {
		log.Fatal("DATABASE_URL is not set")
	}
	accountURL, ok := os.LookupEnv("ACCOUNT_SERVICES_URL")
	if !ok {
		log.Fatal("ACCOUNT_SERVICE IS NOT SET")
	}
	productURL, ok := os.LookupEnv("PRODUCT_SERVICES_URL")
	if !ok {
		log.Fatal("PRODUCT_SERVICE IS NOT SET")
	}
	var r order.Repository
	for {
		var err error
		r, err = order.NewPostgresRepository(dbURL)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	defer r.Close()
	log.Println("Listening on Port 8082")
	s := order.NewService(r)
	log.Fatal(order.ListenGRPC(s, accountURL, productURL, 8082))
}
