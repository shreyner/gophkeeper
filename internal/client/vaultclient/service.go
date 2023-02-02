package vaultclient

import (
	"errors"
	"fmt"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/shreyner/gophkeeper/internal/client/state"
	"github.com/shreyner/gophkeeper/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

type Service struct {
	appState *state.State

	client   proto.GophkeeperClient
	metadata metadata.MD
}

func New(appState *state.State, client proto.GophkeeperClient) *Service {
	s := Service{
		appState: appState,
		client:   client,
	}

	s.metadata = metadata.New(map[string]string{})

	return &s
}

func (s *Service) Login(ctx context.Context, login, password string) error {
	request := proto.LoginRequest{
		Login:    login,
		Password: password,
	}

	loginResponse, err := s.client.Login(ctx, &request)

	if err != nil {
		return err
	}

	s.metadata.Set("token", loginResponse.AuthToken)
	s.appState.SetUserToken(loginResponse.AuthToken)

	return nil
}

func (s *Service) Check(ctx context.Context) error {
	if s.appState.GetUserToken() == "" {
		return errors.New("Not authorized")
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, s.metadata)

	response, err := s.client.CheckAuth(ctxWithMetadata, &empty.Empty{})

	if err != nil {
		return err
	}

	fmt.Println(response.Message)

	return nil
}
