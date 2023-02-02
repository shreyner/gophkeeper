package app

import (
	"fmt"

	"golang.org/x/net/context"
)

type App struct {
}

func New() *App {
	app := App{}

	return &app
}

func (a *App) StopAndClose(ctx context.Context) {

}

func (a *App) Login() {
	fmt.Println("Call login")
}

func (a *App) Register() {
	fmt.Println("Call Register")
}

func (a *App) List() {
	fmt.Println("Call List")
}
