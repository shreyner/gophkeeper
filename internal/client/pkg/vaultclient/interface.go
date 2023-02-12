//go:generate ./bin/mockgen -source=./interfaces.go -destination=./mock/storage.go -package=vaultclient
package vaultclient

import (
	"context"
	"io"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultdata"
)

type VClient interface {
	Login(ctx context.Context, login, password string) error
	Check(ctx context.Context) error
	VaultSync(ctx context.Context, vaultSync []vaultdata.VaultSyncVersion) ([]vaultdata.VaultSyncData, error)
	VaultCreate(ctx context.Context, encryptedVault []byte, s3URL string) (*vaultdata.VaultClientSyncResult, error)
	VaultUpdate(ctx context.Context, ID string, version int, encryptedVault []byte) (*vaultdata.VaultClientSyncResult, error)
	VaultDelete(ctx context.Context, ID string, version int) error
	VaultUpload(ctx context.Context, r io.Reader) (string, error)
	VaultDownload(ctx context.Context, url string) (io.ReadCloser, error)
}
