package product

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/lib/pq"
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

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) Close() {
	_ = r.db.Close()
}

func (r *postgresRepository) PutProduct(ctx context.Context, p Product) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO products (id, name, description, price)
		VALUES ($1, $2, $3, $4)`,
		p.Id, p.Name, p.Description, p.Price,
	)
	return err
}

func (r *postgresRepository) GetProductByID(ctx context.Context, id string) (*Product, error) {
	p := &Product{}
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, price
		FROM products
		WHERE id = $1`, id).Scan(&p.Id, &p.Name, &p.Description, &p.Price)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *postgresRepository) ListProducts(ctx context.Context, skip uint64, take uint64) ([]Product, error) {
	return r.queryProducts(ctx, `
		SELECT id, name, description, price
		FROM products
		ORDER BY id DESC
		OFFSET $1 LIMIT $2`, skip, take)
}

func (r *postgresRepository) ListProductsWithIDs(ctx context.Context, ids []string) ([]Product, error) {
	products, err := r.queryProducts(ctx, `
		WITH requested AS (
			SELECT id, ordinality
			FROM unnest($1::varchar[]) WITH ORDINALITY AS input(id, ordinality)
		)
		SELECT id, name, description, price
		FROM requested
		JOIN products USING (id)
		ORDER BY requested.ordinality`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	if len(products) != len(ids) {
		return nil, ErrNotFound
	}
	return products, nil
}

func (r *postgresRepository) SearchProducts(ctx context.Context, query string, skip uint64, take uint64) ([]Product, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return r.ListProducts(ctx, skip, take)
	}

	return r.queryProducts(ctx, `
		SELECT id, name, description, price
		FROM products
		WHERE search_vector @@ websearch_to_tsquery('english', $1)
		   OR name ILIKE '%' || $1 || '%'
		   OR description ILIKE '%' || $1 || '%'
		ORDER BY ts_rank(search_vector, websearch_to_tsquery('english', $1)) DESC, id DESC
		OFFSET $2 LIMIT $3`, query, skip, take)
}

func (r *postgresRepository) queryProducts(ctx context.Context, query string, args ...any) ([]Product, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		p := Product{}
		if err := rows.Scan(&p.Id, &p.Name, &p.Description, &p.Price); err != nil {
			return nil, fmt.Errorf("scan product: %w", err)
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate products: %w", err)
	}
	return products, nil
}
