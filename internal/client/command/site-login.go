package command

import (
	"context"
	"fmt"
	"strconv"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/storage"
)

type SiteLoginCommand struct {
	vclient           *vaultclient.Client
	vaultCrypt        *vaultcrypt.VaultCrypt
	loginVaultStorage *storage.LoginVaultStorage
}

func NewSiteLoginCommand(
	vclient *vaultclient.Client,
	vaultCrypt *vaultcrypt.VaultCrypt,
	loginVaultStorage *storage.LoginVaultStorage,
) *SiteLoginCommand {
	command := SiteLoginCommand{
		vclient:           vclient,
		vaultCrypt:        vaultCrypt,
		loginVaultStorage: loginVaultStorage,
	}

	return &command
}

func (c *SiteLoginCommand) RunCreate(ctx context.Context, args []string) {
	if len(args) < 3 {
		fmt.Println("incorrect login and password")
		return
	}

	login, password, siteURL := args[0], args[1], args[2]

	if len(login) < 3 || len(password) < 3 || len(siteURL) < 3 {
		fmt.Println("incorrect login and password")
		return
	}

	siteLoginData := storage.LoginSecreteData{
		Login:    login,
		Password: password,
	}

	err := c.loginVaultStorage.Create(&siteLoginData, siteURL)

	if err != nil {
		fmt.Println(err)
	}

}

func (c *SiteLoginCommand) RunView(_ context.Context, _ []string) {
	arr := c.loginVaultStorage.GetAll()

	for _, model := range arr {
		fmt.Printf(
			"ID: %v, IsNew: %v, IsUpdate: %v, IsDeleted: %v, SiteURL: %v\n",
			model.ID,
			model.IsNew,
			model.IsUpdate,
			model.IsDelete,
			model.GetSite(),
		)
	}

}

func (c *SiteLoginCommand) RunViewLogin(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println("incorrect login and password")
		return
	}

	siteLoginID := args[0]

	ID, err := strconv.ParseUint(siteLoginID, 10, 32)

	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	siteLoginData, err := c.loginVaultStorage.ViewDataByID(uint32(ID))

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("ID: %v, Login: %v, Password: %v\n", ID, siteLoginData.Login, siteLoginData.Password)

}

func (c *SiteLoginCommand) RunDelete(ctx context.Context, args []string) {
	if len(args) < 1 {
		fmt.Println("incorrect login and password")
		return
	}

	siteLoginID := args[0]

	ID, err := strconv.ParseUint(siteLoginID, 10, 32)

	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	err = c.loginVaultStorage.DeleteByID(uint32(ID))

	if err != nil {
		fmt.Println(err)
	}
}

func (c *SiteLoginCommand) RunUpdate(ctx context.Context, args []string) {
	if len(args) < 3 {
		fmt.Println("incorrect login and password")
		return
	}

	siteLoginID, login, password := args[0], args[1], args[2]

	if len(login) < 3 || len(password) < 3 {
		fmt.Println("incorrect login and password")
		return
	}

	ID, err := strconv.ParseUint(siteLoginID, 10, 32)

	if err != nil {
		fmt.Println("Invalid ID")
		return
	}

	err = c.loginVaultStorage.UpdateByID(uint32(ID), login, password)

	if err != nil {
		fmt.Println(err)
		return
	}
}
