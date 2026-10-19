package main

import "context"

import (
	"fmt"

	"github.com/tarun0306/go-grpc-graphql-microservice/order"
)

type mutationResolver struct {
	server *Server
}

func (r *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
	account, err := r.server.accountClient.PostAccount(ctx, input.Name, input.Password)
	if err != nil {
		return nil, err
	}

	return &Account{ID: account.ID, Name: account.Name, Role: account.Role}, nil
}

func (r *mutationResolver) Login(ctx context.Context, input LoginInput) (*AuthPayload, error) {
	account, err := r.server.accountClient.Login(ctx, input.Name, input.Password)
	if err != nil {
		return nil, err
	}
	token, err := issueToken(r.server.jwtSecret, &Account{ID: account.ID, Role: account.Role})
	if err != nil {
		return nil, err
	}
	return &AuthPayload{
		Token:   token,
		Account: &Account{ID: account.ID, Name: account.Name, Role: account.Role},
	}, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	claims, err := claimsFromContext(ctx)
	if err != nil || claims.Role != string(AccountRoleOperator) {
		return nil, fmt.Errorf("operator authentication required")
	}
	product, err := r.server.productClient.PostProduct(ctx, input.Name, input.Description, input.Price)
	if err != nil {
		return nil, err
	}
	return &Product{
		ID:          product.Id,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
	}, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	claims, err := claimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != string(AccountRoleOperator) && claims.AccountID != input.AccountID {
		return nil, fmt.Errorf("cannot create an order for another account")
	}
	products := make([]order.OrderedProducts, 0, len(input.Products))
	for _, product := range input.Products {
		products = append(products, order.OrderedProducts{
			ID:       product.ID,
			Quantity: uint32(product.Quantity),
		})
	}

	createdOrder, err := r.server.orderClient.PostOrder(ctx, input.AccountID, products)
	if err != nil {
		return nil, err
	}
	return graphqlOrder(createdOrder), nil
}
