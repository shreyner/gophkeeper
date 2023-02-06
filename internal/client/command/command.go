package command

import (
	"github.com/shreyner/gophkeeper/internal/client/pkg/promptcmd"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultclient"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultcrypt"
	"github.com/shreyner/gophkeeper/internal/client/pkg/vaultsync"
	"github.com/shreyner/gophkeeper/internal/client/storage"
)

func NewCommands(
	vclient *vaultclient.Client,
	vaultCrypt *vaultcrypt.VaultCrypt,
	vsync *vaultsync.VaultSync,
	siteLoginStorage *storage.LoginVaultStorage,
) []promptcmd.Command {

	loginCommand := NewLoginCommand(vclient, vaultCrypt, vsync)
	siteLoginCommand := NewSiteLoginCommand(vclient, vaultCrypt, siteLoginStorage)
	syncCommand := NewSyncCommand(vsync)

	return []promptcmd.Command{
		{
			Command:     "login",
			Description: "Authenticate user",
			Auth:        promptcmd.CommandAuthNot,
			Run:         loginCommand.Run,
		},
		{
			Command:     "check",
			Description: "Authenticate user",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         loginCommand.RunCheck,
		},

		// Vault Site Login

		{
			Command:     "site-login",
			Description: "Show all",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         siteLoginCommand.RunView,
		},
		{
			Command:     "site-login-view",
			Description: "Show login password by ID",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         siteLoginCommand.RunViewLogin,
		},
		{
			Command:     "site-login-create",
			Description: "Create login password",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         siteLoginCommand.RunCreate,
		},
		{
			Command:     "site-login-delete",
			Description: "Create by ID",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         siteLoginCommand.RunDelete,
		},
		{
			Command:     "site-login-update",
			Description: "Create update id login password",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         siteLoginCommand.RunUpdate,
		},
		{
			Command:     "sync",
			Description: "Force sync storage",
			Auth:        promptcmd.CommandAuthNeed,
			Run:         syncCommand.Run,
		},
	}

}
