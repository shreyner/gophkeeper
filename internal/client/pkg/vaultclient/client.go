package vaultclient

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultdata"
	"github.com/shreyner/gophkeeper/proto"
)

var ServerUploaderHost = "http://localhost:3280"

type Client struct {
	appState vaultdata.State

	client   proto.GophkeeperClient
	metadata metadata.MD
}

func New(appState vaultdata.State, client proto.GophkeeperClient) *Client {
	s := Client{
		appState: appState,
		client:   client,
	}

	s.metadata = metadata.New(map[string]string{})

	return &s
}

func (s *Client) Login(ctx context.Context, login, password string) error {
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

func (s *Client) Check(ctx context.Context) error {
	if s.appState.GetUserToken() == "" {
		return ErrNotAuth
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, s.metadata)

	response, err := s.client.CheckAuth(ctxWithMetadata, &empty.Empty{})

	if err != nil {
		return err
	}

	fmt.Println(response.Message)

	return nil
}

func (s *Client) VaultSync(ctx context.Context, vaultSync []vaultdata.VaultSyncVersion) ([]vaultdata.VaultSyncData, error) {
	if s.appState.GetUserToken() == "" {
		return nil, ErrNotAuth
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, s.metadata)

	arrRequest := make([]*proto.VaultSyncRequest_VaultVersion, 0, len(vaultSync))

	for _, v := range vaultSync {
		dataRequest := proto.VaultSyncRequest_VaultVersion{
			Id:      v.ID,
			Version: int32(v.Version),
		}

		arrRequest = append(arrRequest, &dataRequest)
	}

	request := proto.VaultSyncRequest{
		VaultVersions: arrRequest,
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctxWithMetadata, 30*time.Second)
	defer cancel()

	response, err := s.client.VaultSync(ctxWithTimeout, &request)

	if err != nil {
		return nil, err
	}

	resultArr := make([]vaultdata.VaultSyncData, 0, len(response.UpdatedVaults))

	for _, v := range response.UpdatedVaults {
		var s3URL string

		if v.S3 != nil {
			s3URL = v.S3.Value
		}

		data := vaultdata.VaultSyncData{
			ID:        v.Id,
			Vault:     v.Vault,
			Version:   int(v.Version),
			IsDeleted: v.IsDeleted,
			S3URL:     s3URL,
		}

		resultArr = append(resultArr, data)
	}

	return resultArr, nil
}

func (s *Client) VaultCreate(ctx context.Context, encryptedVault []byte, s3URL string) (*vaultdata.VaultClientSyncResult, error) {
	if s.appState.GetUserToken() == "" {
		return nil, ErrNotAuth
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, s.metadata)

	var s3URLRequest *wrapperspb.StringValue

	if s3URL != "" {
		s3URLRequest = wrapperspb.String(s3URL)
	}

	request := proto.VaultCreateRequest{
		Vault: encryptedVault,
		S3:    s3URLRequest,
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctxWithMetadata, 30*time.Second)
	defer cancel()

	response, err := s.client.VaultCreate(ctxWithTimeout, &request)

	if err != nil {
		return nil, err
	}

	d := vaultdata.VaultClientSyncResult{
		ID:      response.Id,
		Version: int(response.Version),
	}

	return &d, nil
}

func (s *Client) VaultUpdate(ctx context.Context, ID string, version int, encryptedVault []byte) (*vaultdata.VaultClientSyncResult, error) {
	if s.appState.GetUserToken() == "" {
		return nil, ErrNotAuth
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, s.metadata)

	request := proto.VaultUpdateRequest{
		Id:      ID,
		Vault:   encryptedVault,
		Version: int32(version),
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctxWithMetadata, 30*time.Second)
	defer cancel()

	response, err := s.client.VaultUpdate(ctxWithTimeout, &request)

	if err != nil {
		return nil, err
	}

	d := vaultdata.VaultClientSyncResult{
		ID:      ID,
		Version: int(response.Version),
	}

	return &d, nil
}

func (s *Client) VaultDelete(ctx context.Context, ID string, version int) error {
	if s.appState.GetUserToken() == "" {
		return ErrNotAuth
	}

	ctxWithMetadata := metadata.NewOutgoingContext(ctx, s.metadata)

	request := proto.VaultDeleteRequest{
		Id:      ID,
		Version: int32(version),
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctxWithMetadata, 30*time.Second)
	defer cancel()

	_, err := s.client.VaultDelete(ctxWithTimeout, &request)

	if err != nil {
		return err
	}

	return nil
}

func (s *Client) VaultUpload(ctx context.Context, r io.Reader) (string, error) {
	if s.appState.GetUserToken() == "" {
		return "", ErrNotAuth
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	request, err := http.NewRequestWithContext(
		ctxWithTimeout,
		http.MethodPut,
		fmt.Sprintf("%s/upload", ServerUploaderHost),
		r,
	)

	if err != nil {
		return "", err
	}

	request.Header.Set("Authorization", s.appState.GetUserToken())
	request.Header.Set("Content-Type", "application/octet-stream")

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return "", err
	}

	defer response.Body.Close()

	respBytes, err := io.ReadAll(response.Body)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(respBytes)), nil
}

func (s *Client) VaultDownload(ctx context.Context, url string) (io.ReadCloser, error) {
	if s.appState.GetUserToken() == "" {
		return nil, ErrNotAuth
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)

	if err != nil {
		return nil, err
	}

	response, err := http.DefaultClient.Do(request)

	if err != nil {
		return nil, err
	}

	return response.Body, nil
}
