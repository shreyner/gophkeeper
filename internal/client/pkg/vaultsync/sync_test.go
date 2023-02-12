package vaultsync

import (
	"testing"

	"github.com/golang/mock/gomock"
	vaultclient "github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient/mock"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
)

func TestVaultSync_Sync(t *testing.T) {
	vcrypto := vaultcrypt.New()
	_ = vcrypto.SetMasterPassword("Alex", "123")

	type fields struct {
		vcrypt *vaultcrypt.VaultCrypt
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Call with storages",
			fields: fields{
				vcrypt: vcrypto,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctl := gomock.NewController(t)
			defer ctl.Finish()

			clientMock := vaultclient.NewMockVClient(ctl)

			s := New(
				tt.fields.vcrypt,
				clientMock,
				[]StorageSyncer{},
			)

			if err := s.Sync(); (err != nil) != tt.wantErr {
				t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
