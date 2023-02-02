package command

import (
	"github.com/shreyner/gophkeeper/internal/client/app"
	"github.com/shreyner/gophkeeper/internal/client/pkg/promptcmd"
	"github.com/shreyner/gophkeeper/internal/client/vaultclient"
)

func NewCommands(app *app.App, vclient *vaultclient.Service) []promptcmd.Command {

	loginCommand := NewLoginCommand(app, vclient)

	return []promptcmd.Command{
		{Command: "login", Description: "Authenticate user", Run: loginCommand.Run},
		{Command: "check", Description: "Authenticate user", Run: loginCommand.RunCheck},
	}

}
