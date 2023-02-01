package user

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Service struct {
	rep *Repository
}

func NewService(rep *Repository) *Service {
	service := Service{
		rep: rep,
	}

	return &service
}

func (s *Service) Create(ctx context.Context, login, password string) (*UserModel, error) {
	userModel := UserModel{
		ID:    uuid.New(),
		Login: login,
	}

	if err := userModel.SetPassword(password); err != nil {
		return nil, err
	}

	if err := s.rep.Create(ctx, &userModel); err != nil {
		return nil, err
	}

	return &userModel, nil
}

func (s *Service) FindByLogin(ctx context.Context, login string) (*UserModel, error) {
	return s.rep.FindByLogin(ctx, login)
}
