package vaultsync

import (
	"testing"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
)

func TestVaultSync_Sync(t *testing.T) {
	vcrypto := vaultcrypt.New()

	type fields struct {
		vcrypt   *vaultcrypt.VaultCrypt
		vclient  *vaultclient.Client
		storages map[string]StorageSyncer
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &VaultSync{
				vcrypt:   tt.fields.vcrypt,
				vclient:  tt.fields.vclient,
				storages: tt.fields.storages,
			}
			if err := s.Sync(); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
