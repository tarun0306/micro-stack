package account

import (
	"context"

	"github.com/tarun0306/go-grpc-graphql-microservice/account/pb"
	"google.golang.org/grpc"
)

type Client struct {
	conn    *grpc.ClientConn
	service pb.AccountServiceClient
}

func NewClient(url string) (*Client, error) {
	conn, err := grpc.Dial(url, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := pb.NewAccountServiceClient(conn)
	return &Client{conn, c}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) PostAccount(ctx context.Context, name, password string) (*Account, error) {
	response, err := c.service.PostAccount(ctx,
		&pb.PostAccountRequest{
			Name: name, Password: password,
		})
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:   response.Account.Id,
		Name: response.Account.Name,
		Role: response.Account.Role,
	}, nil
}

func (c *Client) Login(ctx context.Context, name, password string) (*Account, error) {
	response, err := c.service.Login(ctx, &pb.LoginRequest{Name: name, Password: password})
	if err != nil {
		return nil, err
	}
	return &Account{ID: response.Account.Id, Name: response.Account.Name, Role: response.Account.Role}, nil
}
func (c *Client) GetAccount(ctx context.Context, id string) (*Account, error) {
	response, err := c.service.GetAccount(ctx, &pb.GetAccountRequest{
		Id: id,
	})
	if err != nil {
		return nil, err
	}
	return &Account{
		ID:   response.Account.Id,
		Name: response.Account.Name,
		Role: response.Account.Role,
	}, nil
}
func (c *Client) GetAccounts(ctx context.Context, skip uint64, take uint64) ([]Account, error) {

	response, err := c.service.GetAccounts(ctx, &pb.GetAccountsRequest{
		Skip: skip,
		Take: take,
	})
	if err != nil {
		return nil, err
	}
	accounts := []Account{}
	for _, a := range response.Accounts {
		accounts = append(accounts, Account{
			ID:   a.Id,
			Name: a.Name,
			Role: a.Role,
		})
	}
	return accounts, nil

}
