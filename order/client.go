package order

import (
	"context"
	"fmt"
	"time"

	"github.com/tarun0306/go-grpc-graphql-microservice/order/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.OrderServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("dial order service: %w", err)
	}

	return &Client{
		conn:    conn,
		service: pb.NewOrderServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) PostOrder(ctx context.Context, accountID string, products []OrderedProducts) (*Order, error) {
	requestProducts := make([]*pb.PostOrderRequest_OrderProduct, 0, len(products))
	for _, product := range products {
		requestProducts = append(requestProducts, &pb.PostOrderRequest_OrderProduct{
			ProductId: product.ID,
			Quantity:  uint64(product.Quantity),
		})
	}

	response, err := c.service.PostOrder(ctx, &pb.PostOrderRequest{
		AccountId: accountID,
		Products:  requestProducts,
	})
	if err != nil {
		return nil, err
	}
	if response == nil || response.Order == nil {
		return nil, fmt.Errorf("order service returned an empty order")
	}

	return fromPBOrder(response.Order)
}

func (c *Client) GetOrder(ctx context.Context, id string) (*Order, error) {
	response, err := c.service.GetOrder(ctx, &pb.GetOrderRequest{Id: id})
	if err != nil {
		return nil, err
	}
	if response == nil || response.Order == nil {
		return nil, fmt.Errorf("order service returned an empty order")
	}

	return fromPBOrder(response.Order)
}

func (c *Client) GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error) {
	response, err := c.service.GetOrdersForAccount(ctx, &pb.GetOrdersForAccountRequest{
		AccountId: accountID,
	})
	if err != nil {
		return nil, err
	}
	if response == nil {
		return nil, fmt.Errorf("order service returned an empty response")
	}

	orders := make([]Order, 0, len(response.Orders))
	for _, order := range response.Orders {
		if order == nil {
			return nil, fmt.Errorf("order service returned a nil order")
		}
		converted, err := fromPBOrder(order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, *converted)
	}

	return orders, nil
}

func fromPBOrder(order *pb.Order) (*Order, error) {
	createdAt, err := decodeOrderTime(order.CreatedAt)
	if err != nil {
		return nil, err
	}

	products := make([]OrderedProducts, 0, len(order.Products))
	for _, product := range order.Products {
		if product == nil {
			return nil, fmt.Errorf("order service returned a nil product")
		}
		products = append(products, OrderedProducts{
			ID:          product.Id,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Quantity:    uint32(product.Quantity),
		})
	}

	return &Order{
		ID:         order.Id,
		CreatedAt:  createdAt,
		AccountID:  order.AccountId,
		TotalPrice: order.TotalPrice,
		Products:   products,
	}, nil
}

func decodeOrderTime(data []byte) (createdAtTime time.Time, err error) {
	err = createdAtTime.UnmarshalBinary(data)
	if err != nil {
		return time.Time{}, fmt.Errorf("decode order creation time: %w", err)
	}
	return createdAtTime, nil
}
