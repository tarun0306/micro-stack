package order

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	ErrNotFound = errors.New("Entity not Found")
)

type Repository interface {
	Close()

	PutOrder(ctx context.Context, o Order) error

	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return &postgresRepository{db: db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}

func (r *postgresRepository) PutOrder(ctx context.Context, o Order) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders(id,created_at,account_id,total_price) VALUES ($1,$2,$3,$4)",
		o.ID, o.CreatedAt, o.AccountID, o.TotalPrice,
	)
	if err != nil {
		return
	}

	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	if err != nil {
		return
	}
	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return
		}
	}
	_, err = stmt.Exec()
	if err != nil {
		return
	}
	stmt.Close()
	return
}

func (r *postgresRepository) GetOrdersForAccount(ctx context.Context, accountId string) ([]Order, error) {
	// Query all rows for this account.
	// Each row contains one order header and one product line item because the query joins
	// the orders table with the order_products table.
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			o.id,
			o.created_at,
			o.account_id,
			o.total_price::money::numeric::float8,
			op.product_id,
			op.quantity
		FROM orders o
		JOIN order_products op ON o.id = op.order_id
		WHERE o.account_id = $1
		ORDER BY o.id, op.product_id`,
		accountId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// orders will hold the final list of orders to return.
	orders := make([]Order, 0)

	// current keeps track of the order we are building while iterating rows.
	var current *Order

	// currentProducts stores product entries for the current order before it is finalized.
	var currentProducts []OrderedProducts

	for rows.Next() {
		var (
			id        string
			createdAt time.Time
			accountID string
			total     float64
			productID string
			quantity  int
		)

		// Scan one database row into variables.
		if err = rows.Scan(&id, &createdAt, &accountID, &total, &productID, &quantity); err != nil {
			return nil, err
		}

		// If we moved to a new order ID, the previous order is complete.
		// Save the previous order and start a new one.
		if current == nil || current.ID != id {
			if current != nil {
				current.Products = append([]OrderedProducts(nil), currentProducts...)
				orders = append(orders, *current)
			}
			current = &Order{
				ID:         id,
				CreatedAt:  createdAt,
				AccountID:  accountID,
				TotalPrice: total,
			}
			currentProducts = nil
		}

		// Add this product to the current order's product list.
		currentProducts = append(currentProducts, OrderedProducts{
			ID:       productID,
			Quantity: uint32(quantity),
		})
	}

	// Finalize the last order if there was one.
	if current != nil {
		current.Products = append([]OrderedProducts(nil), currentProducts...)
		orders = append(orders, *current)
	}

	// Check for any iterator errors after reading all rows.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
