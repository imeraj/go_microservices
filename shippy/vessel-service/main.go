package main

import (
	"errors"
	"fmt"
	"os"

	pb "github.com/imeraj/go_microservices/shippy/vessel-service/proto/vessel"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
)

////////
type IRepository interface {
	FindAvailable(*pb.Specification) (*pb.Vessel, error)
}

type Repository struct {
	vessels []*pb.Vessel
}

func (repo *Repository) FindAvailable(spec *pb.Specification) (*pb.Vessel, error) {
	for _, vessel := range repo.vessels {
		if spec.Capacity <= vessel.Capacity && spec.MaxWeight <= vessel.MaxWeight {
			return vessel, nil
		}
	}
	return nil, errors.New("No vessel found by that spec")
}

///////
type Service struct {
	repo IRepository
}

func (s *Service) FindAvailable(ctx context.Context, req *pb.Specification, res *pb.Response) error {
	vessel, err := s.repo.FindAvailable(req)
	if err != nil {
		return err
	}

	res.Vessel = vessel
	return nil
}

////////
func main() {
	vessels := []*pb.Vessel{
		&pb.Vessel{Id: "vessel001", Name: "Boaty McBoatface", MaxWeight: 200000, Capacity: 500},
	}
	repo := &Repository{vessels}

	// Create a new service. Optionally include some options here.
	service := grpc.NewService(
		micro.Name("vessel-service"),
	)

	// Init will parse the command line flags.
	service.Init()
	service.Server().Init(server.Address(os.Getenv("MICRO_SERVER_ADDRESS")))

	// Register handler
	pb.RegisterVesselServiceHandler(service.Server(), &Service{repo})

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
