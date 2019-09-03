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
	"gopkg.in/mgo.v2"
)

////////
const (
	defaultHost = "localhost:27017"
)

const (
	dbName                = "shippy"
	consignmentCollection = "consignments"
)

type Repository interface {
	Create(*pb.Consignment) error
	GetAll() ([]*pb.Consignment, error)
	Close()
}

type ConsignmentRepository struct {
	session *mgo.Session
}

func (repo *ConsignmentRepository) Create(consignment *pb.Consignment) error {
	return repo.collection().Insert(consignment)
}

func (repo *ConsignmentRepository) GetAll() ([]*pb.Consignment, error) {
	var consignments []*pb.Consignment
	// Find normally takes a query, but as we want everything, we can nil this.
	// We then bind our consignments variable by passing it as an argument to .All().
	// That sets consignments to the result of the find query.
	// There's also a `One()` function for single results.
	err := repo.collection().Find(nil).All(&consignments)
	return consignments, err
}

func (repo *ConsignmentRepository) Close() {
	repo.session.Close()
}

func (repo *ConsignmentRepository) collection() *mgo.Collection {
	return repo.session.DB(dbName).C(consignmentCollection)
}

///////
type Service struct {
	session      *mgo.Session
	vesselClient vesselProto.VesselServiceClient
}

func (s *Service) GetRepo() Repository {
	return &ConsignmentRepository{s.session.Clone()}
}

func (s *Service) CreateConsignment(ctx context.Context, req *pb.Consignment, res *pb.Response) error {
	defer s.GetRepo().Close()

	vesselResponse, err := s.vesselClient.FindAvailable(context.Background(), &vesselProto.Specification{
		MaxWeight: req.Weight,
		Capacity:  int32(len(req.Containers)),
	})
	if err != nil {
		return err
	}
	log.Printf("Found vessel: %s \n", vesselResponse.Vessel.Name)

	req.VesselId = vesselResponse.Vessel.Id

	// Save our consignment
	err = s.GetRepo().Create(req)
	if err != nil {
		return err
	}

	res.Created = true
	res.Consignment = req
	return nil
}

func (s *Service) GetConsignments(ctx context.Context, req *pb.GetRequest, res *pb.Response) error {
	defer s.GetRepo().Close()
	consignments, err := s.GetRepo().GetAll()
	if err != nil {
		return err
	}
	res.Consignments = consignments
	return nil
}

////////
func main() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = defaultHost
	}

	session, err := CreateSession(host)
	defer session.Close()
	if err != nil {
		log.Panicf("Could not connect to datastore with host %s - %v", host, err)
	}

	// Create a new service. Optionally include some options here.
	service := grpc.NewService(
		micro.Name("consignment-service"),
	)

	// Init will parse the command line flags.
	service.Init()
	service.Server().Init(server.Address(os.Getenv("MICRO_SERVER_ADDRESS")))

	vesselClient := vesselProto.NewVesselServiceClient("vessel-service", service.Client())

	// Register handler
	pb.RegisterShippingServiceHandler(service.Server(), &Service{session, vesselClient})

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
