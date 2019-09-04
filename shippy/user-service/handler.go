package main

import (
	pb "github.com/imeraj/go_microservices/shippy/user-service/proto/user"
	"golang.org/x/net/context"
)

type Service struct {
	repo         Repository
	tokenService Authable
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
	_, err := s.repo.GetByEmailAndPassword(req)
	if err != nil {
		return err
	}
	res.Token = "testingabc"
	return nil
}

func (s *Service) Create(ctx context.Context, req *pb.User, res *pb.Response) error {
	if err := s.repo.Create(req); err != nil {
		return err
	}
	res.User = req
	return nil
}

func (s *Service) ValidateToken(ctx context.Context, req *pb.Token, res *pb.Token) error {
	return nil
}
