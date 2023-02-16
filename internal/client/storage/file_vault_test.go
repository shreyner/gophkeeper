package storage

import (
	"os"
	"path"
	"testing"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileVaultStorage_LoadFromLocalFile(t *testing.T) {
	t.Run("Success open file", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		pwd, _ := os.Getwd()
		fileTestDataDB := path.Join(pwd, "testdata", "files-base.db")

		vcrypto := vaultcrypt.New()
		_ = vcrypto.SetMasterPassword("Alex", "123")

		fileStorage := NewFileVaultStorage(vcrypto, nil)

		err := fileStorage.LoadFromLocalFile(fileTestDataDB)
		require.Nil(err, "failed load db file")

		require.Len(fileStorage.storage, 3, "incorrect length storage")

		assert.Equal(fileStorage.storage[1].GetFileName(), "screen.png")
		assert.Equal(fileStorage.storage[2].GetFileName(), "screen.png")
		assert.Equal(fileStorage.storage[3].GetFileName(), "screen.png")
	})

	t.Run("Success create file if not exists", func(t *testing.T) {
		require := require.New(t)
		fileTestDataDB := path.Join(t.TempDir(), "files-new.db")

		vcrypto := vaultcrypt.New()
		_ = vcrypto.SetMasterPassword("Alex", "123")

		fileStorage := NewFileVaultStorage(vcrypto, nil)

		err := fileStorage.LoadFromLocalFile(fileTestDataDB)

		require.Nil(err, "failed load db file")
		require.Len(fileStorage.storage, 0, "incorrect length storage")

		file, err := os.Open(fileTestDataDB)
		require.Nil(err, "cant find file after create")
		defer file.Close()
	})
}

func TestFileVaultStorage_SaveToFile(t *testing.T) {
	t.Run("", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		filesTestDataDB := path.Join(t.TempDir(), "files-new.db")

		vcrypto := vaultcrypt.New()
		_ = vcrypto.SetMasterPassword("Alex", "123")

		fileStorage := NewFileVaultStorage(vcrypto, nil)

		model := FileVaultModel{
			ID:         1,
			ExternalID: "",
			Data:       []byte{},
			MetaData:   map[string]string{},
			S3URL:      "",
		}
		model.SetFileName("screen.png")

		fileStorage.storage[1] = &model

		err := fileStorage.SaveToFile(filesTestDataDB)
		require.Nil(err, "cant save to file")

		fileStorage = NewFileVaultStorage(vcrypto, nil)
		err = fileStorage.LoadFromLocalFile(filesTestDataDB)

		require.Nil(err, "failed load db file")
		require.Len(fileStorage.storage, 1, "incorrect length storage")

		assert.Equal(fileStorage.storage[1].GetID(), uint32(1))
		assert.Equal(fileStorage.storage[1].GetFileName(), "screen.png")
	})
}
