package vaultcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
)

var ErrNotSetKey = errors.New("don't set key")

type VaultCrypt struct {
	aesGCM cipher.AEAD
	nonce  []byte

	isSetKey bool
}

func New() *VaultCrypt {
	v := VaultCrypt{
		isSetKey: false,
	}

	return &v
}

func (c *VaultCrypt) SetKey(key []byte) error {
	sh := sha256.New()
	sh.Write(key)

	hashKey := sh.Sum(nil)

	aesBlock, err := aes.NewCipher(hashKey)

	if err != nil {
		return err
	}

	aesGCM, err := cipher.NewGCM(aesBlock)

	if err != nil {
		return err
	}

	nonce := hashKey[len(hashKey)-aesGCM.NonceSize():]

	c.aesGCM = aesGCM
	c.nonce = nonce
	c.isSetKey = true

	return nil
}

func (c *VaultCrypt) Encrypt(data []byte) ([]byte, error) {
	if !c.isSetKey {
		return nil, ErrNotSetKey
	}

	return c.aesGCM.Seal(nil, c.nonce, data, nil), nil
}

func (c *VaultCrypt) Decrypt(data []byte) ([]byte, error) {
	if !c.isSetKey {
		return nil, ErrNotSetKey
	}

	decryptedData, err := c.aesGCM.Open(nil, c.nonce, data, nil)

	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
