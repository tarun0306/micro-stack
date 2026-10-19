package main

import (
	"context"
	"fmt"
)

type queryResolver struct {
	server *Server
}

func (r *queryResolver) Accounts(ctx context.Context, pagination *PaginationInput, id *string) ([]*Account, error) {
	claims, err := claimsFromContext(ctx)
	if err != nil || claims.Role != string(AccountRoleOperator) {
		return nil, fmt.Errorf("operator authentication required")
	}
	if id != nil {
		account, err := r.server.accountClient.GetAccount(ctx, *id)
		if err != nil {
			return nil, err
		}

		return []*Account{{ID: account.ID, Name: account.Name, Role: account.Role}}, nil
	}

	skip, take := paginationValues(pagination)
	accounts, err := r.server.accountClient.GetAccounts(ctx, skip, take)
	if err != nil {
		return nil, err
	}

	result := make([]*Account, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, &Account{ID: account.ID, Name: account.Name, Role: account.Role})
	}
	return result, nil
}

func (r *queryResolver) Viewer(ctx context.Context) (*Account, error) {
	claims, err := claimsFromContext(ctx)
	if err != nil {
		return nil, err
	}
	account, err := r.server.accountClient.GetAccount(ctx, claims.AccountID)
	if err != nil {
		return nil, err
	}
	return &Account{ID: account.ID, Name: account.Name, Role: account.Role}, nil
}

func (r *queryResolver) Products(ctx context.Context, pagination *PaginationInput, query *string, id *string) ([]*Product, error) {
	skip, take := paginationValues(pagination)
	ids := []string{}
	if id != nil {
		ids = append(ids, *id)
	}

	search := ""
	if query != nil {
		search = *query
	}
	products, err := r.server.productClient.GetProducts(ctx, skip, take, ids, search)
	if err != nil {
		return nil, err
	}

	result := make([]*Product, 0, len(products))
	for _, product := range products {
		result = append(result, &Product{
			ID:          product.Id,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
		})
	}
	return result, nil
}

func paginationValues(pagination *PaginationInput) (uint64, uint64) {
	if pagination == nil {
		return 0, 0
	}
	var skip, take uint64
	if pagination.Skip != nil && *pagination.Skip > 0 {
		skip = uint64(*pagination.Skip)
	}
	if pagination.Take != nil && *pagination.Take > 0 {
		take = uint64(*pagination.Take)
	}
	return skip, take
}
