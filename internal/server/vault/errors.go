package vault

import "errors"

var ErrVaultNotFound = errors.New("vault not found")

var ErrVaultConflict = errors.New("vault conflict")
