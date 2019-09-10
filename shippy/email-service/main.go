package main

import (
	"encoding/json"
	"log"
	"os"

	pb "github.com/imeraj/go_microservices/shippy/email-service/proto/user"
	micro "github.com/micro/go-micro"
	"github.com/micro/go-micro/server"
	"github.com/nats-io/nats.go"
)

const topic = "user.created"

func main() {
	service := micro.NewService(
		micro.Name("email-service"),
	)

	// Init will parse the command line flags.
	service.Init()
	service.Server().Init(server.Address(os.Getenv("MICRO_SERVER_ADDRESS")))

	uri := os.Getenv("NATS_URI")
	nc, err := nats.Connect(uri)
	if err != nil {
		log.Fatal(err)
	}

	nc.Subscribe(topic, func(m *nats.Msg) {
		var user *pb.User
		_ = json.Unmarshal(m.Data, &user)
		go sendEmail(user)
	})

	// Run the server
	if err := service.Run(); err != nil {
		log.Println(err)
	}
}

func sendEmail(user *pb.User) error {
	log.Println("Sending email to:", user.Name)
	return nil
}
