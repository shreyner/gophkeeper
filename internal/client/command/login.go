package command

import (
	"context"
	"fmt"

	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultsync"
)

type LoginCommand struct {
	vclient    *vaultclient.Client
	vaultCrypt *vaultcrypt.VaultCrypt
	vsync      *vaultsync.VaultSync
}

func NewLoginCommand(
	vclient *vaultclient.Client,
	vaultCrypt *vaultcrypt.VaultCrypt,
	vsync *vaultsync.VaultSync,
) *LoginCommand {
	command := LoginCommand{
		vclient:    vclient,
		vaultCrypt: vaultCrypt,
		vsync:      vsync,
	}

	return &command
}

func (c *LoginCommand) Run(ctx context.Context, args []string) {
	if len(args) < 2 {
		fmt.Println("incorrect login and password")
		return
	}

	login, password := args[0], args[1]

	if len(login) < 3 || len(password) < 3 {
		fmt.Println("incorrect login and password")
		return
	}

	err := c.vclient.Login(ctx, login, password)

	if err != nil {
		fmt.Println(err)
		return
	}

	err = c.vaultCrypt.SetMasterPassword(login, password)

	if err != nil {
		fmt.Println(err)
		return
	}

	//err = c.vsync.Sync()
	//
	//if err != nil {
	//	fmt.Println("Error first sync storage: ", err)
	//	return
	//}
}

func (c *LoginCommand) RunCheck(ctx context.Context, _ []string) {
	err := c.vclient.Check(ctx)

	if err != nil {
		fmt.Println(err)
		return
	}
}
