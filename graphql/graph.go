package main

import (
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/tarun0306/go-grpc-graphql-microservice/account"
	"github.com/tarun0306/go-grpc-graphql-microservice/order"
	"github.com/tarun0306/go-grpc-graphql-microservice/product"
)

type Server struct {
	accountClient *account.Client
	productClient *product.Client
	orderClient   *order.Client
	jwtSecret     []byte
}

func NewGraphQLServer(accountUrl, productUrl, orderUrl string, jwtSecret string) (*Server, error) {
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be configured")
	}
	accountClient, err := account.NewClient(accountUrl)
	if err != nil {
		return nil, err
	}

	productClient, err := product.NewClient(productUrl)
	if err != nil {
		accountClient.Close()
		return nil, err
	}

	orderClient, err := order.NewClient(orderUrl)
	if err != nil {
		accountClient.Close()
		productClient.Close()
		return nil, err
	}
	return &Server{
		accountClient,
		productClient,
		orderClient,
		[]byte(jwtSecret),
	}, nil
}

func (s *Server) Mutation() MutationResolver {
	return &mutationResolver{
		server: s,
	}
}
func (s *Server) Query() QueryResolver {
	return &queryResolver{
		server: s,
	}
}
func (s *Server) Account() AccountResolver {
	return &accountResolver{
		server: s,
	}
}

func (s *Server) ToExecutableSchema() graphql.ExecutableSchema {
	return NewExecutableSchema(Config{
		Resolvers: s,
	})
}
