package product

import (
	"context"
	"encoding/json"
	"errors"

	elastic "github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
)

var (
	ErrNotFound = errors.New("Entity not found")
)

type Repository interface {
	Close()
	PutProduct(ctx context.Context, p Product) error
	GetProductByID(ctx context.Context, id string) (*Product, error)
	ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error)
	ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error)
	SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error)
}

type elasticRepository struct {
	client *elastic.TypedClient
}
type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewElasticRepository(url string) (Repository, error) {
	client, err := elastic.NewTyped(
		elastic.WithAddresses(url),
	)
	if err != nil {
		return nil, err
	}
	return &elasticRepository{client}, nil
}

func (r *elasticRepository) Close() {}

func (r *elasticRepository) PutProduct(ctx context.Context, p Product) error {
	_, err := r.client.
		Index("catalog").
		Id(p.Id).Document(productDocument{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}).Do(ctx)

	return err

}
func (r *elasticRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Get("catalog", id).Do(ctx)
	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, ErrNotFound
	}

	p := productDocument{}
	if err := json.Unmarshal(res.Source_, &p); err != nil {
		return nil, err
	}
	return &Product{
		Id:          id,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, nil

}
func (r *elasticRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		From(int(skip)).
		Size(int(take)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}

	for _, hits := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(hits.Source_, &p); err == nil {
			products = append(products, Product{
				Id:          *hits.Id_,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil

}
func (r *elasticRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	items, err := r.client.Mget().Index("catalog").Ids(ids...).Do(ctx)
	if err != nil {
		return nil, err
	}
	products := make([]Product, 0, len(items.Docs))

	for _, item := range items.Docs {

		getRes, ok := item.(*types.GetResult)
		if !ok {
			return nil, ErrNotFound
		}
		if !getRes.Found {
			return nil, ErrNotFound
		}
		productDocument := productDocument{}
		if err := json.Unmarshal(getRes.Source_, &productDocument); err != nil {
			return nil, err
		}
		products = append(products, Product{
			Id:          getRes.Id_,
			Name:        productDocument.Name,
			Description: productDocument.Description,
			Price:       productDocument.Price,
		})
	}
	return products, nil
}
func (r *elasticRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	res, err := r.client.Search().
		Q(query).
		From(int(skip)).
		Size(int(take)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(hit.Source_, &p); err == nil {
			products = append(products, Product{
				Id:          *hit.Id_,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil
}
