package main

import (
	"fmt"

	pb "github.com/imeraj/go_microservices/shippy/consignment-service/proto/consignment"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"golang.org/x/net/context"
)

////////
type IRepository interface {
	Create(*pb.Consignment) (*pb.Consignment, error)
	GetAll() []*pb.Consignment
}

type Repository struct {
	consignments []*pb.Consignment
}

func (repo *Repository) Create(consignment *pb.Consignment) (*pb.Consignment, error) {
	updated := append(repo.consignments, consignment)
	repo.consignments = updated
	return consignment, nil
}

func (repo *Repository) GetAll() []*pb.Consignment {
	return repo.consignments
}

///////
type Service struct {
	repo IRepository
}

func (s *Service) CreateConsignment(ctx context.Context, req *pb.Consignment, resp *pb.Response) error {
	consignment, err := s.repo.Create(req)
	if err != nil {
		resp = nil
		return err
	}

	resp.Created = true
	resp.Consignment = consignment

	return nil
}

func (s *Service) GetConsignments(ctx context.Context, req *pb.GetRequest, resp *pb.Response) error {
	consignments := s.repo.GetAll()
	resp.Consignments = consignments
	return nil
}

////////
func main() {
	repo := &Repository{}

	// Create a new service. Optionally include some options here.
	service := grpc.NewService(
		micro.Name("consignment-service"),
	)

	// Init will parse the command line flags.
	service.Init()

	// Register handler
	pb.RegisterShippingServiceHandler(service.Server(), &Service{repo})

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
