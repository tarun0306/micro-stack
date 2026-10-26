package product

import (
	"context"
	"fmt"
	"net"

	"github.com/tarun0306/go-grpc-graphql-microservice/internal/logging"
	"github.com/tarun0306/go-grpc-graphql-microservice/product/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	pb.UnimplementedProductServiceServer
	service Service
}

func ListenGrpc(s Service, port int) error {
	listening, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	server := grpc.NewServer(grpc.UnaryInterceptor(logging.UnaryServerInterceptor("product")))
	pb.RegisterProductServiceServer(server, &grpcServer{
		UnimplementedProductServiceServer: pb.UnimplementedProductServiceServer{},
		service:                           s})
	reflection.Register(server)
	return server.Serve(listening)

}

func (s *grpcServer) PostProduct(ctx context.Context, req *pb.PostProductRequest) (*pb.PostProductResponse, error) {
	prod, err := s.service.PostProduct(ctx, req.Name, req.Description, req.Price)
	if err != nil {
		return nil, err
	}
	return &pb.PostProductResponse{Product: &pb.Product{
		Id:          prod.Id,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
	}}, nil

}
func (s *grpcServer) GetProduct(ctx context.Context, req *pb.GetProductRequest) (*pb.GetProductResponse, error) {
	prod, err := s.service.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &pb.GetProductResponse{Product: &pb.Product{
		Id:          prod.Id,
		Name:        prod.Name,
		Description: prod.Description,
		Price:       prod.Price,
	}}, nil
}
func (s *grpcServer) GetProducts(ctx context.Context, req *pb.GetProductsRequest) (*pb.GetProductsResponse, error) {
	var res []Product
	var err error
	if req.Query != "" {
		res, err = s.service.SearchProducts(ctx, req.Query, req.Skip, req.Take)
	} else if len(req.Ids) != 0 {
		res, err = s.service.GetProductByIDs(ctx, req.Ids)
	} else {
		res, err = s.service.GetProducts(ctx, req.Skip, req.Take)
	}
	if err != nil {
		return nil, err
	}
	products := []*pb.Product{}
	for _, p := range res {
		products = append(products, &pb.Product{
			Id:          p.Id,
			Name:        p.Name,
			Description: p.Description,
			Price:       p.Price,
		},
		)
	}
	return &pb.GetProductsResponse{Products: products}, nil
}
