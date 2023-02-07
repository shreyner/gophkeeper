package vaultcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
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

func (c *VaultCrypt) setKey(key []byte) error {
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

func (c *VaultCrypt) EncryptStream(out io.Writer, key []byte) (*cipher.StreamWriter, error) {
	sh := sha256.New()
	sh.Write(key)

	hashKey := sh.Sum(nil)

	aesBlock, err := aes.NewCipher(hashKey)

	if err != nil {
		return nil, err
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(aesBlock, iv[:])

	//writer := &cipher.StreamWriter{
	//	S: stream,
	//	W: out,
	//}

	//_, err = io.Copy(writer, in)
	//
	//if err != nil {
	//	return err
	//}

	//return nil

	return &cipher.StreamWriter{
		S: stream,
		W: out,
	}, nil
}

func (c *VaultCrypt) DecryptStream(in io.Reader, out io.Writer, key []byte) error {
	sh := sha256.New()
	sh.Write(key)

	hashKey := sh.Sum(nil)

	aesBlock, err := aes.NewCipher(hashKey)

	if err != nil {
		return err
	}

	var iv [aes.BlockSize]byte
	stream := cipher.NewOFB(aesBlock, iv[:])

	reader := &cipher.StreamReader{
		S: stream,
		R: in,
	}

	_, err = io.Copy(out, reader)

	if err != nil {
		return err
	}

	return nil
}

func (c *VaultCrypt) SetMasterPassword(login, password string) error {
	key, err := scrypt.Key([]byte(password), []byte(login), 1<<15, 8, 1, 32)

	if err != nil {
		fmt.Println(err)
		return err
	}

	err = c.setKey(key)

	if err != nil {
		return err
	}

	return nil
}
