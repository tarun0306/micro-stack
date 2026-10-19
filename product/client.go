package product

import (
	"context"

	"github.com/tarun0306/go-grpc-graphql-microservice/product/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.ProductServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	serviceClient := pb.NewProductServiceClient(conn)
	return &Client{conn, serviceClient}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostProduct(ctx context.Context, name, description string, price float64) (*Product, error) {
	response, err := c.service.PostProduct(ctx, &pb.PostProductRequest{
		Name:        name,
		Description: description,
		Price:       price,
	})
	if err != nil {
		return nil, err
	}
	return &Product{
		Id:          response.Product.Id,
		Name:        response.Product.Name,
		Description: response.Product.Description,
		Price:       response.Product.Price,
	}, nil
}

func (c *Client) GetProduct(ctx context.Context, id string) (*Product, error) {
	prodRes, err := c.service.GetProduct(ctx, &pb.GetProductRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return &Product{
		Id:          prodRes.Product.Id,
		Name:        prodRes.Product.Name,
		Description: prodRes.Product.Description,
		Price:       prodRes.Product.Price,
	}, nil

}

func (c *Client) GetProducts(ctx context.Context, skip uint64, take uint64, ids []string, query string) ([]Product, error) {
	res, err := c.service.GetProducts(
		ctx,
		&pb.GetProductsRequest{
			Ids:   ids,
			Skip:  skip,
			Take:  take,
			Query: query,
		},
	)
	if err != nil {
		return nil, err
	}
	products := []Product{}
	for _, p := range res.Products {

		products = append(products, Product{
			Id:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		})

	}
	return products, nil
}
