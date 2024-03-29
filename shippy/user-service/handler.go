package main

import (
	"encoding/json"
	"errors"
	"log"

	pb "github.com/imeraj/go_microservices/shippy/user-service/proto/auth"
	"github.com/nats-io/nats.go"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/context"
)

const topic = "user.created"

type Service struct {
	repo         Repository
	tokenService Authable
	nc           *nats.Conn
}

func (s *Service) Get(ctx context.Context, req *pb.User, res *pb.Response) error {
	user, err := s.repo.Get(req.Id)
	if err != nil {
		return err
	}
	res.User = user
	return nil
}

func (s *Service) GetAll(ctx context.Context, req *pb.Request, res *pb.Response) error {
	users, err := s.repo.GetAll()
	if err != nil {
		return err
	}
	res.Users = users
	return nil
}

func (s *Service) Auth(ctx context.Context, req *pb.User, res *pb.Token) error {
	log.Println("Logging in with:", req.Email, req.Password)

	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		return err
	}
	log.Println(user)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return err
	}

	token, err := s.tokenService.Encode(user)
	if err != nil {
		return err
	}

	res.Token = token
	return nil
}

func (s *Service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	req.Password = string(hashedPass)
	if err := s.repo.Create(req); err != nil {
		return err
	}

	res.User = req

	if err := s.publishEvent(req); err != nil {
		return err
	}

	return nil
}

func (s *Service) ValidateToken(ctx context.Context, req *pb.Token, res *pb.Token) error {
	claims, err := s.tokenService.Decode(req.Token)
	if err != nil {
		return err
	}

	if claims.User.Id == "" {
		return errors.New("invalid user")
	}

	res.Valid = true
	return nil
}

func (s *Service) publishEvent(user *pb.User) error {
	body, err := json.Marshal(user)
	if err != nil {
		return err
	}

	s.nc.Publish(topic, body)
	log.Printf("[pub] successful.")
	return nil
}
