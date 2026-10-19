package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/tarun0306/go-grpc-graphql-microservice/account"
)

type Config struct {
	DatabaseURL string `envconfig:"DATABASE_URL"`
}

func main() {

	dburl, ispresent := os.LookupEnv("DATABASE_URL")
	if !ispresent {
		log.Fatal("url doesnot exists in en")
		return
	}
	var r account.Repository
	for {
		var err error
		r, err = account.NewPostgresRRepository(dburl)
		if err != nil {
			log.Println(err)
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}
	defer r.Close()
	log.Println("Listening on Port 8080...")
	s := account.NewService(r)
	if err := s.BootstrapOperator(context.Background(), os.Getenv("BOOTSTRAP_OPERATOR_NAME"), os.Getenv("BOOTSTRAP_OPERATOR_PASSWORD")); err != nil {
		log.Fatal(err)
	}
	log.Fatal(account.ListenGRPC(s, 8080))

}
