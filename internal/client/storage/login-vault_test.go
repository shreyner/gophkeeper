package storage

import (
	"os"
	"path"
	"testing"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginVaultStorage_LoadFromLocalFile(t *testing.T) {
	t.Run("Success open file", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)
		pwd, _ := os.Getwd()
		siteloginTestDataDB := path.Join(pwd, "testdata", "site-login-base.db")

		vcrypto := vaultcrypt.New()
		_ = vcrypto.SetMasterPassword("Alex", "123")

		siteLoginStorage := NewLoginVaultStorage(vcrypto)

		err := siteLoginStorage.LoadFromLocalFile(siteloginTestDataDB)

		require.Nil(err, "failed load db file")
		require.Len(siteLoginStorage.storage, 5, "incorrect length storage")

		assert.Equal(siteLoginStorage.storage[1].GetSite(), "vk.vom")
		assert.Equal(siteLoginStorage.storage[3].GetSite(), "vk.vom")
		assert.Equal(siteLoginStorage.storage[5].GetSite(), "vk.vom")

		secret1, err := siteLoginStorage.ViewDataByID(1)
		require.Nil(err, "error encrypted data")
		assert.Equal(secret1.Login, "Alex")
		assert.Equal(secret1.Password, "123")

		secret5, err := siteLoginStorage.ViewDataByID(5)
		require.Nil(err, "error encrypted data")
		assert.Equal(secret5.Login, "Alexx")
		assert.Equal(secret5.Password, "444")
	})

	t.Run("Success create file if not exists", func(t *testing.T) {
		require := require.New(t)
		siteloginTestDataDB := path.Join(t.TempDir(), "site-login-new.db")

		vcrypto := vaultcrypt.New()
		_ = vcrypto.SetMasterPassword("Alex", "123")

		siteLoginStorage := NewLoginVaultStorage(vcrypto)

		err := siteLoginStorage.LoadFromLocalFile(siteloginTestDataDB)

		require.Nil(err, "failed load db file")
		require.Len(siteLoginStorage.storage, 0, "incorrect length storage")

		file, err := os.Open(siteloginTestDataDB)
		require.Nil(err, "cant find file after create")
		defer file.Close()
	})
}

func TestLoginVaultStorage_SaveToFile(t *testing.T) {
	t.Run("Success save folder", func(t *testing.T) {
		assert := assert.New(t)
		require := require.New(t)

		siteloginTestDataDB := path.Join(t.TempDir(), "site-login-base.db")

		vcrypto := vaultcrypt.New()
		_ = vcrypto.SetMasterPassword("Alex", "123")

		siteLoginStorage := NewLoginVaultStorage(vcrypto)

		secretData1 := LoginSecreteData{
			Login:    "alex",
			Password: "123",
		}
		secretData2 := LoginSecreteData{
			Login:    "polly",
			Password: "321",
		}

		err := siteLoginStorage.Create(&secretData1, "vk.com")
		require.Nil(err, "error create data site login")

		err = siteLoginStorage.Create(&secretData2, "yandex.ru")
		require.Nil(err, "error create data site login")

		err = siteLoginStorage.SaveToFile(siteloginTestDataDB)
		require.Nil(err, "cant save to file")

		siteLoginStorage = NewLoginVaultStorage(vcrypto)
		err = siteLoginStorage.LoadFromLocalFile(siteloginTestDataDB)

		require.Nil(err, "failed load db file")
		require.Len(siteLoginStorage.storage, 2, "incorrect length storage")

		assert.Equal(siteLoginStorage.storage[1].GetSite(), "vk.com")
		assert.Equal(siteLoginStorage.storage[2].GetSite(), "yandex.ru")

		secret1, err := siteLoginStorage.ViewDataByID(1)
		require.Nil(err, "error encrypted data")
		assert.Equal(secret1.Login, "alex")
		assert.Equal(secret1.Password, "123")

		secret5, err := siteLoginStorage.ViewDataByID(2)
		require.Nil(err, "error encrypted data")
		assert.Equal(secret5.Login, "polly")
		assert.Equal(secret5.Password, "321")
	})
}
