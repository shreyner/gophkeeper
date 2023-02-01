package auth

import (
	"errors"

	"github.com/shreyner/gophkeeper/internal/server/user"
	"golang.org/x/net/context"
)

type Service struct {
	userService *user.Service
}

func NewService(userService *user.Service) *Service {
	service := Service{
		userService: userService,
	}

	return &service
}

func (s *Service) FindOrCreate(ctx context.Context, login, password string) (*user.UserModel, error) {
	userModel, err := s.userService.FindByLogin(ctx, login)

	if err == nil {
		return userModel, nil
	}

	if !errors.Is(err, user.ErrUserNotFound) {
		return nil, err
	}

	return s.userService.Create(ctx, login, password)
}
