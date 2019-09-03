package main

import (
	"fmt"
	"log"
	"os"

	pb "github.com/imeraj/go_microservices/shippy/consignment-service/proto/consignment"
	vesselProto "github.com/imeraj/go_microservices/shippy/vessel-service/proto/vessel"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
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
	repo         IRepository
	vesselClient vesselProto.VesselServiceClient
}

func (s *Service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {
	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})
	if err != nil {
		return err
	}
	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)

	req.VesselId = vesselResponse.Vessel.Id
	consignment, err := s.repo.Create(req)
	if err != nil {
		return err
	}

	res.Created = true
	res.Consignment = consignment
	return nil
}

func (s *Service) GetConsignments(ctx context.Context, req *pb.GetRequest, resp *pb.Response) error {
	consignments := s.repo.GetAll()

	if len(consignments) > 0 {
		resp.Created = true
		resp.Consignment = consignments[0]
	}

	if len(consignments) > 1 {
		resp.Consignments = consignments[1:]
	}

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
	service.Server().Init(server.Address(os.Getenv("MICRO_SERVER_ADDRESS")))

	vesselClient := vesselProto.NewVesselServiceClient("vessel-service", service.Client())

	// Register handler
	pb.RegisterShippingServiceHandler(service.Server(), &Service{repo, vesselClient})

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
