package vaultsync

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultdata"
)

type VaultSync struct {
	vcrypt  *vaultcrypt.VaultCrypt
	vclient *vaultclient.Client

	storages map[string]StorageSyncer
}

func New(
	vcrypt *vaultcrypt.VaultCrypt,
	vclient *vaultclient.Client,
	storages []StorageSyncer,
) *VaultSync {
	s := VaultSync{
		vcrypt:  vcrypt,
		vclient: vclient,
	}

	mapStorages := make(map[string]StorageSyncer, len(storages))

	for _, vaultStorage := range storages {
		mapStorages[vaultStorage.GetKind()] = vaultStorage
	}

	s.storages = mapStorages

	return &s
}

type dataSync struct {
	typeVaultStorage string
	vault            DataSyncer
	s3URL            string
}

type vaultSyncData struct {
	TypeVaultStorage string
	Data             []byte
}

func (v *VaultSync) EncryptVault(src vaultSyncData) ([]byte, error) {
	var buffer bytes.Buffer
	err := gob.NewEncoder(&buffer).Encode(src)

	if err != nil {
		return nil, err
	}

	encryptedData, err := v.vcrypt.Encrypt(buffer.Bytes())

	if err != nil {
		return nil, err
	}

	return encryptedData, nil
}

func (v *VaultSync) DecryptVault(encryptedDist []byte) (*vaultSyncData, error) {
	dst, err := v.vcrypt.Decrypt(encryptedDist)

	if err != nil {
		return nil, err
	}

	var src vaultSyncData

	err = gob.NewDecoder(bytes.NewReader(dst)).Decode(&src)

	if err != nil {
		return &src, err
	}

	return &src, nil
}

func (v *VaultSync) createVault(arr []dataSync) error {
	ctx := context.Background()

	for _, d := range arr {
		storage := v.storages[d.typeVaultStorage]
		dstVault, err := storage.SerializeToVault(d.vault)

		if err != nil {
			return err
		}

		vsd := vaultSyncData{
			TypeVaultStorage: d.typeVaultStorage,
			Data:             dstVault,
		}

		encrypted, err := v.EncryptVault(vsd)

		if err != nil {
			return err
		}

		createdInfo, err := v.vclient.VaultCreate(ctx, encrypted, d.s3URL)

		if err != nil {
			return err
		}

		err = storage.UpdateAfterSyncByID(d.vault, createdInfo.ID, createdInfo.Version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VaultSync) updateVault(arr []dataSync) error {
	ctx := context.Background()

	for _, d := range arr {
		storage := s.storages[d.typeVaultStorage]
		dstVault, err := storage.SerializeToVault(d.vault)

		if err != nil {
			return err
		}

		vsd := vaultSyncData{
			TypeVaultStorage: d.typeVaultStorage,
			Data:             dstVault,
		}

		encrypted, err := s.EncryptVault(vsd)

		if err != nil {
			return err
		}

		createdInfo, err := s.vclient.VaultUpdate(ctx, d.vault.GetVaultID(), d.vault.GetVersion(), encrypted)

		if err != nil {
			return err
		}

		err = storage.UpdateAfterSyncByID(d.vault, createdInfo.ID, createdInfo.Version)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VaultSync) deleteVault(arr []dataSync) error {
	ctx := context.Background()

	for _, d := range arr {
		storage := s.storages[d.typeVaultStorage]

		err := s.vclient.VaultDelete(ctx, d.vault.GetVaultID(), d.vault.GetVersion())

		if err != nil {
			return err
		}

		err = storage.ConfirmDeleteAfterSyncByID(d.vault)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VaultSync) createVaultStorage(data []vaultdata.VaultSyncData) error {
	for _, datum := range data {
		vsd, err := s.DecryptVault(datum.Vault)

		if err != nil {
			return err
		}

		storage, ok := s.storages[vsd.TypeVaultStorage]

		if !ok {
			continue
		}

		d, err := storage.DeserializeFromVault(vsd.Data)

		if err != nil {
			return err
		}

		err = storage.CreateDataStorage(datum.ID, datum.Version, d, datum.S3URL)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VaultSync) updateVaultStorage(data []vaultdata.VaultSyncData) error {
	for _, datum := range data {
		vsd, err := s.DecryptVault(datum.Vault)

		if err != nil {
			return err
		}

		storage, ok := s.storages[vsd.TypeVaultStorage]

		if !ok {
			continue
		}

		d, err := storage.DeserializeFromVault(vsd.Data)

		if err != nil {
			return err
		}

		err = storage.UpdateDataStorage(datum.ID, datum.Version, d)

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *VaultSync) deleteVaultStorage(data []vaultdata.VaultSyncData) error {
	for _, syncer := range s.storages {
		for _, datum := range data {
			err := syncer.DeleteDataStorage(datum.ID, datum.Version)

			if errors.Is(err, vaultdata.ErrNotFoundVaultInStorage) {
				continue
			}

			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *VaultSync) Sync() error {
	ctx := context.Background()

	// First
	newVaultForStorage := make([]dataSync, 0)
	updateVaultForStorage := make([]dataSync, 0)
	deleteVaultForStorage := make([]dataSync, 0)

	for typeVaultStorage, storage := range s.storages {
		arr, err := storage.LoadForSync()

		if err != nil {
			return err
		}

		for _, v := range arr {
			d := dataSync{
				typeVaultStorage: typeVaultStorage,
				vault:            v,
				s3URL:            v.GetS3URL(),
			}

			if v.GetIsNew() {
				newVaultForStorage = append(newVaultForStorage, d)
				continue
			}

			if v.GetIsDelete() {
				deleteVaultForStorage = append(deleteVaultForStorage, d)
				continue
			}

			if v.GetIsUpdate() {
				updateVaultForStorage = append(updateVaultForStorage, d)
				continue
			}
		}
	}

	err := s.createVault(newVaultForStorage)
	if err != nil {
		return err
	}

	err = s.deleteVault(deleteVaultForStorage)
	if err != nil {
		return err
	}

	err = s.updateVault(updateVaultForStorage)
	if err != nil {
		return err
	}

	// Second

	localVaults := make(map[string]dataSync)

	for typeVaultStorage, storage := range s.storages {
		arr, err := storage.LoadForSync()

		if err != nil {
			return err
		}

		for _, v := range arr {
			localVaults[v.GetVaultID()] = dataSync{
				typeVaultStorage: typeVaultStorage,
				vault:            v,
				s3URL:            v.GetS3URL(),
			}
		}
	}

	vaultsVersionForRequest := make([]vaultdata.VaultSyncVersion, 0, len(localVaults))

	for ID, v := range localVaults {
		d := vaultdata.VaultSyncVersion{
			ID:      ID,
			Version: v.vault.GetVersion(),
		}
		vaultsVersionForRequest = append(vaultsVersionForRequest, d)
	}

	responseVaultSync, err := s.vclient.VaultSync(ctx, vaultsVersionForRequest)

	if err != nil {
		return err
	}

	newVault := make([]vaultdata.VaultSyncData, 0)
	updateVault := make([]vaultdata.VaultSyncData, 0)
	deletedVault := make([]vaultdata.VaultSyncData, 0)

	conflictVault := make([]vaultdata.VaultSyncData, 0)

	for _, responseData := range responseVaultSync {
		vaultCurrent, ok := localVaults[responseData.ID]

		if !ok {
			newVault = append(newVault, responseData)
			continue
		}

		if responseData.Version > vaultCurrent.vault.GetVersion() && vaultCurrent.vault.GetIsUpdate() {
			fmt.Printf("Vault Type: %v, ID: %v conflict merge. Please resolve conflict for vault\n", vaultCurrent.typeVaultStorage, vaultCurrent.vault.GetID())
			conflictVault = append(conflictVault, responseData)
			storage, _ := s.storages[vaultCurrent.typeVaultStorage]
			err = storage.SetConflictFlag(vaultCurrent.vault.GetID())
			if err != nil {
				// TODO: Add Log
				return err
			}

			continue
		}

		if responseData.IsDeleted {
			deletedVault = append(deletedVault, responseData)
			continue
		}

		updateVault = append(updateVault, responseData)
	}

	err = s.deleteVaultStorage(deletedVault)
	if err != nil {
		return err
	}

	err = s.createVaultStorage(newVault)
	if err != nil {
		return err
	}

	err = s.updateVaultStorage(updateVault)
	if err != nil {
		return err
	}

	return nil
}
