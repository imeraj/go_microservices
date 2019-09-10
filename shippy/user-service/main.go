package main

import (
	"log"
	"os"

	pb "github.com/imeraj/go_microservices/shippy/user-service/proto/auth"
	"github.com/micro/go-grpc"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/nats-io/nats.go"
)

func main() {
	db, err := CreateConnection()
	defer db.Close()

	if err != nil {
		log.Fatalf("Error connecting to DB: %v", err)
	}

	db.AutoMigrate(&pb.User{})

	repo := &UserRepository{db}
	tokenService := &TokenService{repo}

	// Create a new service. Optionally include some options here.
	service := grpc.NewService(
		micro.Name("user-service"),
	)

	// Init will parse the command line flags.
	service.Init()
	service.Server().Init(server.Address(os.Getenv("MICRO_SERVER_ADDRESS")))

	uri := os.Getenv("NATS_URI")
	nc, err := nats.Connect(uri)
	if err != nil {
		log.Fatal(err)
	}

	// Register handler
	pb.RegisterAuthHandler(service.Server(), &Service{repo, tokenService, nc})

	// Run the server
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
