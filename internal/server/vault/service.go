package vault

import (
	"github.com/google/uuid"
	"golang.org/x/net/context"
)

type Service struct {
	rep *Repository
}

func NewService(rep *Repository) *Service {
	service := Service{rep: rep}

	return &service
}

func (s *Service) Create(ctx context.Context, userId uuid.UUID, vault []byte, s3URL *string) (*VaultModel, error) {
	vaultModel := VaultModel{
		ID:        uuid.New(),
		UserID:    userId,
		Vault:     vault,
		Version:   0,
		IsDeleted: false,
		S3:        s3URL,
	}

	err := s.rep.Create(ctx, &vaultModel)

	if err != nil {
		return nil, err
	}

	return &vaultModel, nil
}

func (s *Service) Update(ctx context.Context, userId, vaultID uuid.UUID, vault []byte, version int) (int, error) {
	newVersion, err := s.rep.UpdateVault(ctx, userId, vaultID, vault, version)

	if err != nil {
		return 0, err
	}

	return newVersion, nil
}

func (s *Service) Delete(ctx context.Context, userID, vaultID uuid.UUID, version int) error {
	err := s.rep.Delete(ctx, userID, vaultID, version)

	return err
}

func (s *Service) LoadUpdated(ctx context.Context, userID uuid.UUID, vaultsVersionsDTO []VaultVersionDTO) ([]VaultModel, error) {
	newVaults, err := s.rep.LoadUpdatedVaults(ctx, userID, vaultsVersionsDTO)

	return newVaults, err
}
