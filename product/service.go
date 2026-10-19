package product

import (
	"context"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostProduct(ctx context.Context, name, description string, price float64) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	GetProductByIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}
type Product struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}
type productService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &productService{r}
}

func (s *productService) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	prod := &Product{
		Id:          ksuid.New().String(),
		Name:        name,
		Description: description,
		Price:       price,
	}
	if err := s.repository.PutProduct(ctx, *prod); err != nil {
		return nil, err
	}
	return prod, nil

}

func (s *productService) GetProduct(ctx context.Context, id string) (*Product, error) {

	p, err := s.repository.GetProductByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (s *productService) GetProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	listAccnts, err := s.repository.ListProducts(ctx, skip, take)
	if err != nil {
		return nil, err
	}
	return listAccnts, nil
}

func (s *productService) GetProductByIDs(ctx context.Context, ids []string) ([]Product, error) {

	products, err := s.repository.ListProductsWithIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	return products, nil

}

func (s *productService) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	if take > 100 || (skip == 0 && take == 0) {
		take = 100
	}
	return s.repository.SearchProducts(ctx, query, skip, take)
}
