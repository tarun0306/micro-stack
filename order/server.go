package order

import (
	"context"
	"fmt"
	"net"

	"github.com/tarun0306/go-grpc-graphql-microservice/account"
	"github.com/tarun0306/go-grpc-graphql-microservice/internal/logging"
	"github.com/tarun0306/go-grpc-graphql-microservice/order/pb"
	"github.com/tarun0306/go-grpc-graphql-microservice/product"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type grpcServer struct {
	pb.UnimplementedOrderServiceServer
	service       Service
	accountClient *account.Client
	productClient *product.Client
}

func ListenGRPC(s Service, accountURL, productURL string, port int) error {
	accountClient, err := account.NewClient(accountURL)
	if err != nil {
		return fmt.Errorf("create account client: %w", err)
	}
	defer accountClient.Close()

	productClient, err := product.NewClient(productURL)
	if err != nil {
		return fmt.Errorf("create product client: %w", err)
	}
	defer productClient.Close()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen on port %d: %w", port, err)
	}
	defer listener.Close()

	server := grpc.NewServer(grpc.UnaryInterceptor(logging.UnaryServerInterceptor("order")))
	pb.RegisterOrderServiceServer(server, &grpcServer{
		service:       s,
		accountClient: accountClient,
		productClient: productClient,
	})
	reflection.Register(server)

	return server.Serve(listener)
}

func (s *grpcServer) PostOrder(ctx context.Context, req *pb.PostOrderRequest) (*pb.PostOrderResponse, error) {
	if _, err := s.accountClient.GetAccount(ctx, req.AccountId); err != nil {
		return nil, status.Error(codes.NotFound, "account not found")
	}

	productIDs := make([]string, 0, len(req.Products))
	for _, requestedProduct := range req.Products {
		productIDs = append(productIDs, requestedProduct.ProductId)
	}

	availableProducts, err := s.productClient.GetProducts(ctx, 0, uint64(len(productIDs)), productIDs, "")
	if err != nil {
		return nil, status.Error(codes.NotFound, "products not found")
	}

	products := make([]OrderedProducts, 0, len(availableProducts))
	for i, availableProduct := range availableProducts {
		if i >= len(req.Products) {
			break
		}
		products = append(products, OrderedProducts{
			ID:          availableProduct.Id,
			Name:        availableProduct.Name,
			Description: availableProduct.Description,
			Price:       availableProduct.Price,
			Quantity:    uint32(req.Products[i].Quantity),
		})
	}

	createdOrder, err := s.service.PostOrder(ctx, req.AccountId, products)
	if err != nil {
		return nil, err
	}

	order, err := toPBOrder(createdOrder)
	if err != nil {
		return nil, err
	}
	return &pb.PostOrderResponse{Order: order}, nil
}

func (s *grpcServer) GetOrder(context.Context, *pb.GetOrderRequest) (*pb.GetOrderResponse, error) {
	return nil, status.Error(codes.Unimplemented, "GetOrder is not implemented by the order service")
}

func (s *grpcServer) GetOrdersForAccount(ctx context.Context, req *pb.GetOrdersForAccountRequest) (*pb.GetOrdersForAccountResponse, error) {
	orders, err := s.service.GetOrdersForAccount(ctx, req.AccountId)
	if err != nil {
		return nil, err
	}

	for i := range orders {
		ids := make([]string, 0, len(orders[i].Products))
		for _, product := range orders[i].Products {
			ids = append(ids, product.ID)
		}
		products, err := s.productClient.GetProducts(ctx, 0, uint64(len(ids)), ids, "")
		if err != nil {
			return nil, status.Error(codes.Internal, "could not load order products")
		}
		productDetails := make(map[string]product.Product, len(products))
		for _, product := range products {
			productDetails[product.Id] = product
		}
		for j := range orders[i].Products {
			detail, ok := productDetails[orders[i].Products[j].ID]
			if !ok {
				return nil, status.Error(codes.NotFound, "order product not found")
			}
			orders[i].Products[j].Name = detail.Name
			orders[i].Products[j].Description = detail.Description
			orders[i].Products[j].Price = detail.Price
		}
	}

	result := make([]*pb.Order, 0, len(orders))
	for i := range orders {
		order, err := toPBOrder(&orders[i])
		if err != nil {
			return nil, err
		}
		result = append(result, order)
	}
	return &pb.GetOrdersForAccountResponse{Orders: result}, nil
}

func toPBOrder(order *Order) (*pb.Order, error) {
	products := make([]*pb.OrderProduct, 0, len(order.Products))
	for _, product := range order.Products {
		products = append(products, &pb.OrderProduct{
			Id:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Quantity:    uint64(product.Quantity),
		})
	}

	createdAt, err := order.CreatedAt.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("marshal order creation time: %w", err)
	}

	return &pb.Order{
		Id:         order.ID,
		CreatedAt:  createdAt,
		AccountId:  order.AccountID,
		TotalPrice: order.TotalPrice,
		Products:   products,
	}, nil
}
