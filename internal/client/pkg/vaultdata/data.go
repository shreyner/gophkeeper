package vaultdata

type VaultSyncVersion struct {
	ID      string
	Version int
}

type VaultSyncData struct {
	ID        string
	Vault     []byte
	Version   int
	IsDeleted bool
	S3URL     string
}

// For Client

type VaultClientSyncResult struct {
	ID      string
	Version int
}
