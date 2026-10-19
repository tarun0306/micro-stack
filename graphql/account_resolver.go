package main

import (
	"context"
	"fmt"

	"github.com/tarun0306/go-grpc-graphql-microservice/order"
)

type accountResolver struct {
	server *Server
}

func (r *accountResolver) Role(ctx context.Context, account *Account) (AccountRole, error) {
	switch account.Role {
	case string(AccountRoleCustomer):
		return AccountRoleCustomer, nil
	case string(AccountRoleOperator):
		return AccountRoleOperator, nil
	default:
		return "", fmt.Errorf("unknown account role %q", account.Role)
	}
}

func (r *accountResolver) Orders(ctx context.Context, account *Account) ([]*Order, error) {
	claims, err := claimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if claims.Role != string(AccountRoleOperator) && claims.AccountID != account.ID {
		return nil, fmt.Errorf("cannot access another account's orders")
	}
	orders, err := r.server.orderClient.GetOrdersForAccount(ctx, account.ID)
	if err != nil {
		return nil, err
	}

	result := make([]*Order, 0, len(orders))
	for _, currentOrder := range orders {
		result = append(result, graphqlOrder(&currentOrder))
	}
	return result, nil
}

func graphqlOrder(currentOrder *order.Order) *Order {
	products := make([]*OrderedProduct, 0, len(currentOrder.Products))
	for _, product := range currentOrder.Products {
		products = append(products, &OrderedProduct{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Quantity:    int(product.Quantity),
		})
	}

	return &Order{
		ID:         currentOrder.ID,
		CreatedAt:  currentOrder.CreatedAt,
		TotalPrice: currentOrder.TotalPrice,
		Products:   products,
	}
}
