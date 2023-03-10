//go:generate ./bin/mockgen -source=./interfaces.go -destination=./mock/storage.go -package=vaultsync
package vaultsync

//type StorageID uint32
//type VaultID string

type DataSyncer interface {
	GetID() uint32
	GetVaultID() string

	GetVersion() int
	GetS3URL() string
	GetIsNew() bool
	GetIsDelete() bool
	GetIsUpdate() bool
	IsNeedSync() bool
}

type StorageSyncer interface {
	GetKind() string

	LoadForSync() ([]DataSyncer, error)

	SerializeToVault(data interface{}) ([]byte, error)
	DeserializeFromVault([]byte) (interface{}, error)

	// For update
	UpdateAfterSyncByID(data DataSyncer, externalID string, version int) error
	ConfirmDeleteAfterSyncByID(data DataSyncer) error

	// For created
	CreateDataStorage(externalID string, version int, data interface{}, s3URL string) error
	UpdateDataStorage(externalID string, version int, data interface{}) error
	DeleteDataStorage(externalID string, version int) error

	SetConflictFlag(ID uint32) error
}
