package product

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
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

type openSearchRepository struct {
	client *opensearchapi.Client
}
type productDocument struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func NewOpenSearchRepository(url string) (Repository, error) {
	username := os.Getenv("OPENSEARCH_USERNAME")
	password := os.Getenv("OPENSEARCH_PASSWORD")

	if username == "" || password == "" {
		return nil, errors.New("OpenSearch username or password is missing")
	}

	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses:      []string{url},
			Username:       username,
			Password:       password,
			RequestTimeout: 5 * time.Second,
		},
	})
	if err != nil {
		return nil, err
	}

	if _, err := client.Info(context.Background(), nil); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &openSearchRepository{client: client}, nil
}

func (r *openSearchRepository) Close() {
	_ = r.client.Close()
}

func (r *openSearchRepository) PutProduct(ctx context.Context, p Product) error {
	body, err := json.Marshal(productDocument{
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	})
	if err != nil {
		return err
	}

	_, err = r.client.Index(ctx, opensearchapi.IndexReq{
		Index:      "catalog",
		DocumentID: p.Id,
		Body:       bytes.NewReader(body),
	})
	return err

}
func (r *openSearchRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	res, err := r.client.Document.Get(ctx, opensearchapi.DocumentGetReq{
		Index:      "catalog",
		DocumentID: id,
	})
	if err != nil {
		return nil, err
	}
	if !res.Found {
		return nil, ErrNotFound
	}

	p := productDocument{}
	if err := json.Unmarshal(res.Source, &p); err != nil {
		return nil, err
	}
	return &Product{
		Id:          id,
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
	}, nil

}
func (r *openSearchRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	body, err := json.Marshal(map[string]any{
		"from": skip,
		"size": take,
		"query": map[string]any{
			"match_all": map[string]any{},
		},
	})
	if err != nil {
		return nil, err
	}
	res, err := r.client.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{"catalog"},
		Body:    bytes.NewReader(body),
	})
	if err != nil {
		return nil, err
	}
	products := []Product{}

	for _, hit := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(hit.Source, &p); err == nil {
			products = append(products, Product{
				Id:          hit.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil

}
func (r *openSearchRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	body, err := json.Marshal(map[string]any{"ids": ids})
	if err != nil {
		return nil, err
	}
	items, err := r.client.MGet(ctx, opensearchapi.MGetReq{
		Index: "catalog",
		Body:  bytes.NewReader(body),
	})
	if err != nil {
		return nil, err
	}
	products := make([]Product, 0, len(items.Docs))

	for _, item := range items.Docs {
		if !item.Found {
			return nil, ErrNotFound
		}
		productDocument := productDocument{}
		if err := json.Unmarshal(item.Source, &productDocument); err != nil {
			return nil, err
		}
		products = append(products, Product{
			Id:          item.ID,
			Name:        productDocument.Name,
			Description: productDocument.Description,
			Price:       productDocument.Price,
		})
	}
	return products, nil
}
func (r *openSearchRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	body, err := json.Marshal(map[string]any{
		"from": skip,
		"size": take,
		"query": map[string]any{
			"query_string": map[string]string{"query": query},
		},
	})
	if err != nil {
		return nil, err
	}
	res, err := r.client.Search(ctx, &opensearchapi.SearchReq{
		Indices: []string{"catalog"},
		Body:    bytes.NewReader(body),
	})
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, hit := range res.Hits.Hits {
		p := productDocument{}
		if err = json.Unmarshal(hit.Source, &p); err == nil {
			products = append(products, Product{
				Id:          hit.ID,
				Name:        p.Name,
				Description: p.Description,
				Price:       p.Price,
			})
		}
	}
	return products, nil
}
