package main

import (
	"fmt"
	"log"
	"os"

	pb "github.com/imeraj/go_microservices/shippy/vessel-service/proto/vessel"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"golang.org/x/net/context"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

////////
const (
	defaultHost = "localhost:27017"
)

const (
	dbName           = "shippy"
	vesselCollection = "vessels"
)

type Repository interface {
	FindAvailable(*pb.Specification) (*pb.Vessel, error)
	Create(vessel *pb.Vessel) error
	Close()
}

type VesselRepository struct {
	session *mgo.Session
}

func (repo *VesselRepository) FindAvailable(spec *pb.Specification) (*pb.Vessel, error) {
	var vessel *pb.Vessel

	// Here we define a more complex query than our consignment-service's
	// GetAll function. Here we're asking for a vessel who's max weight and
	// capacity are greater than and equal to the given capacity and weight.
	// We're also using the `One` function here as that's all we want.
	err := repo.collection().Find(bson.M{
		"capacity":  bson.M{"$gte": spec.Capacity},
		"maxweight": bson.M{"$gte": spec.MaxWeight},
	}).One(&vessel)
	if err != nil {
		return nil, err
	}
	return vessel, nil
}

func (repo *VesselRepository) Create(vessel *pb.Vessel) error {
	return repo.collection().Insert(vessel)
}

func (repo *VesselRepository) Close() {
	repo.session.Close()
}

func (repo *VesselRepository) collection() *mgo.Collection {
	return repo.session.DB(dbName).C(vesselCollection)
}

///////
type Service struct {
	session *mgo.Session
}

func (s *Service) GetRepo() Repository {
	return &VesselRepository{s.session.Clone()}
}

func (s *Service) FindAvailable(ctx context.Context, req *pb.Specification, res *pb.Response) error {
	defer s.GetRepo().Close()
	vessel, err := s.GetRepo().FindAvailable(req)
	if err != nil {
		return err
	}

	res.Vessel = vessel
	return nil
}

func (s *Service) Create(ctx context.Context, req *pb.Vessel, res *pb.Response) error {
	defer s.GetRepo().Close()
	if err := s.GetRepo().Create(req); err != nil {
		return err
	}
	res.Vessel = req
	res.Created = true
	return nil
}

////////
func createDummyData(repo Repository) {
	defer repo.Close()
	vessels := []*pb.Vessel{
		{Id: "vessel001", Name: "Kane's Salty Secret", MaxWeight: 200000, Capacity: 500},
	}
	for _, v := range vessels {
		repo.Create(v)
	}
}

func main() {
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = defaultHost
	}

	session, err := CreateSession(host)
	defer session.Close()

	if err != nil {
		log.Fatalf("Error connecting to datastore: %v", err)
	}

	repo := &VesselRepository{session.Copy()}
	createDummyData(repo)

	// Create a new service. Optionally include some options here.
	service := grpc.NewService(
		micro.Name("vessel-service"),
	)

	// Init will parse the command line flags.
	service.Init()
	service.Server().Init(server.Address(os.Getenv("MICRO_SERVER_ADDRESS")))

	// Register handler
	pb.RegisterVesselServiceHandler(service.Server(), &Service{session})

	// Run the server
	if err := service.Run(); err != nil {
		fmt.Println(err)
	}
}
