package rpchandlers

import (
	"errors"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/uuid"
	"github.com/shreyner/gophkeeper/internal/server/pgk/stoken"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/shreyner/gophkeeper/internal/server/auth"
	interceptorauth "github.com/shreyner/gophkeeper/internal/server/interceptor/auth"
	"github.com/shreyner/gophkeeper/internal/server/vault"
	pb "github.com/shreyner/gophkeeper/proto"
)

var (
	_ pb.GophkeeperServer = (*GophkeeperServer)(nil)
)

type GophkeeperServer struct {
	pb.UnimplementedGophkeeperServer

	log          *zap.Logger
	authService  *auth.Service
	vaultService *vault.Service
	stoken       *stoken.Service
}

func NewGophkeeperServer(
	log *zap.Logger,
	authService *auth.Service,
	stoken *stoken.Service,
	vaultService *vault.Service,
) *GophkeeperServer {
	return &GophkeeperServer{
		log:          log,
		authService:  authService,
		stoken:       stoken,
		vaultService: vaultService,
	}
}

func (s *GophkeeperServer) Login(ctx context.Context, in *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := s.authService.FindOrCreate(ctx, in.Login, in.Password)

	if err != nil {
		s.log.Error("can't create or find user", zap.Error(err))
		return nil, status.Error(codes.Internal, "error create or find user")
	}

	valid, err := user.VerifyPassword(in.Password)

	if err != nil {
		s.log.Error("can't verification password", zap.Error(err))
		return nil, status.Error(codes.Internal, "error auth user")
	}

	if !valid {
		return nil, status.Error(codes.PermissionDenied, "invalid password")
	}

	tokenData := stoken.Data{ID: user.ID}

	token, err := s.stoken.CreateToken(&tokenData)

	if err != nil {
		s.log.Error("can't create token", zap.Error(err))
		return nil, status.Error(codes.Internal, "error auth user")
	}

	loginResponse := pb.LoginResponse{
		AuthToken: token,
	}

	return &loginResponse, nil
}

func (s *GophkeeperServer) CheckAuth(ctx context.Context, _ *empty.Empty) (*pb.CheckAuthResponse, error) {
	_, ok := interceptorauth.GetTokenDataCtx(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Не авторизован")
	}

	response := pb.CheckAuthResponse{
		Message: "Hello resp",
	}

	return &response, nil
}

func (s *GophkeeperServer) VaultCreate(ctx context.Context, in *pb.VaultCreateRequest) (*pb.VaultCreateResponse, error) {
	tokenData, ok := interceptorauth.GetTokenDataCtx(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Не авторизован")
	}

	userID := tokenData.ID

	var s3ulr *string

	if in.S3 != nil {
		s3ulr = &(in.S3.Value)
	}

	vaultModel, err := s.vaultService.Create(ctx, userID, in.Vault, s3ulr)

	if err != nil {
		s.log.Error("can't create vault", zap.Error(err))
		return nil, status.Error(codes.Internal, "error create vault")
	}

	response := pb.VaultCreateResponse{
		Id:      vaultModel.ID.String(),
		Version: int32(vaultModel.Version),
	}

	return &response, nil
}

func (s *GophkeeperServer) VaultUpdate(ctx context.Context, in *pb.VaultUpdateRequest) (*pb.VaultUpdateResponse, error) {
	tokenData, ok := interceptorauth.GetTokenDataCtx(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Не авторизован")
	}

	userID := tokenData.ID
	vaultID, err := uuid.Parse(in.Id)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vault ID")
	}

	updatedVersion, err := s.vaultService.Update(ctx, userID, vaultID, in.Vault, int(in.Version))

	if errors.Is(err, vault.ErrVaultNotFound) {
		return nil, status.Error(codes.NotFound, "vault not found")
	}

	if errors.Is(err, vault.ErrVaultConflict) {
		return nil, status.Error(codes.AlreadyExists, "vault conflict")
	}

	if err != nil {
		s.log.Error("can't update vault", zap.Error(err))
		return nil, status.Error(codes.Internal, "error update vault")
	}

	response := pb.VaultUpdateResponse{Version: int32(updatedVersion)}

	return &response, nil
}

func (s *GophkeeperServer) VaultDelete(ctx context.Context, in *pb.VaultDeleteRequest) (*empty.Empty, error) {
	tokenData, ok := interceptorauth.GetTokenDataCtx(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Не авторизован")
	}

	userID := tokenData.ID
	vaultID, err := uuid.Parse(in.Id)

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid vault ID")
	}

	err = s.vaultService.Delete(ctx, userID, vaultID, int(in.Version))

	if errors.Is(err, vault.ErrVaultNotFound) {
		return nil, status.Error(codes.NotFound, "vault not found")
	}

	if errors.Is(err, vault.ErrVaultConflict) {
		return nil, status.Error(codes.AlreadyExists, "vault conflict")
	}

	if err != nil {
		s.log.Error("can't delete vault", zap.Error(err))
		return nil, status.Error(codes.Internal, "error delete vault")
	}

	return &empty.Empty{}, nil
}

func (s *GophkeeperServer) VaultSync(ctx context.Context, in *pb.VaultSyncRequest) (*pb.VaultSyncResponse, error) {
	tokenData, ok := interceptorauth.GetTokenDataCtx(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "Не авторизован")
	}

	userID := tokenData.ID

	vaultVersions := make([]vault.VaultVersionDTO, len(in.VaultVersions))

	for i, vaultVersion := range in.VaultVersions {
		ID, err := uuid.Parse(vaultVersion.Id)

		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid vault ID")
		}

		v := vault.VaultVersionDTO{
			ID:      ID,
			Version: int(vaultVersion.Version),
		}

		vaultVersions[i] = v
	}

	newVaults, err := s.vaultService.LoadUpdated(ctx, userID, vaultVersions)

	if err != nil {
		s.log.Error("can't load for sync vault", zap.Error(err))
		return nil, status.Error(codes.Internal, "error load vault")
	}

	responseVaults := make([]*pb.VaultSyncResponse_Vault, 0, len(newVaults))

	for _, newVault := range newVaults {
		var s3Response *wrapperspb.StringValue

		if newVault.S3 != nil {
			s3Response = wrapperspb.String(*newVault.S3)
		}

		v := pb.VaultSyncResponse_Vault{
			Id:        newVault.ID.String(),
			Vault:     newVault.Vault,
			Version:   int32(newVault.Version),
			IsDeleted: newVault.IsDeleted,
			S3:        s3Response,
		}

		responseVaults = append(responseVaults, &v)
	}

	response := pb.VaultSyncResponse{
		UpdatedVaults: responseVaults,
	}

	return &response, nil
}
