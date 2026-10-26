package order

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/ksuid"
)

type Service interface {
	PostOrder(ctx context.Context, accountID string, products []OrderedProducts) (*Order, error)
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type Order struct {
	ID         string
	CreatedAt  time.Time
	TotalPrice float64
	AccountID  string
	Products   []OrderedProducts
}

type OrderedProducts struct {
	ID          string
	Name        string
	Description string
	Price       float64
	Quantity    uint32
}

type orderService struct {
	repository Repository
}

func NewService(r Repository) Service {
	return &orderService{r}
}

func (s *orderService) PostOrder(ctx context.Context, accountID string, products []OrderedProducts) (*Order, error) {
	log.Printf("service=order operation=PostOrder account_id=%q products=%d", accountID, len(products))
	o := &Order{
		ID:        ksuid.New().String(),
		CreatedAt: time.Now(),
		AccountID: accountID,
		Products:  products,
	}
	o.TotalPrice = 0.0
	for _, p := range products {
		o.TotalPrice += p.Price * float64(p.Quantity)
	}
	err := s.repository.PutOrder(ctx, *o)
	if err != nil {
		return nil, err
	}
	return o, err
}
func (s *orderService) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	log.Printf("service=order operation=GetOrdersForAccount account_id=%q", accountID)

	orders, err := s.repository.GetOrdersForAccount(ctx, accountID)
	if err != nil {
		return nil, err
	}
	return orders, nil
}
