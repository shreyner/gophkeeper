package command

import (
	"fmt"

	"github.com/shreyner/gophkeeper/internal/client/app"
	"github.com/shreyner/gophkeeper/internal/client/vaultclient"
	"golang.org/x/net/context"
)

type LoginCommand struct {
	app     *app.App
	vclient *vaultclient.Service
}

func NewLoginCommand(app *app.App, vclient *vaultclient.Service) *LoginCommand {
	loginCommand := LoginCommand{
		app:     app,
		vclient: vclient,
	}

	return &loginCommand
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
}

func (c *LoginCommand) RunCheck(ctx context.Context, _ []string) {
	err := c.vclient.Check(ctx)

	if err != nil {
		fmt.Println(err)
		return
	}
}
