package vault

import "github.com/google/uuid"

type VaultVersionDTO struct {
	ID      uuid.UUID
	Version int
}

type VaultModel struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Vault     []byte
	Version   int
	IsDeleted bool
	S3        *string
}
